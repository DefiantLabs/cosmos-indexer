package cmd

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/DefiantLabs/probe/client"
	"github.com/go-co-op/gocron"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/core"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/probe"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/cosmos-indexer/tasks"
	"github.com/spf13/cobra"

	"gorm.io/gorm"
)

type Indexer struct {
	cfg                  *config.IndexConfig
	dryRun               bool
	db                   *gorm.DB
	cl                   *client.ChainClient
	scheduler            *gocron.Scheduler
	blockEnqueueFunction func(chan *core.EnqueueData) error
}

var indexer Indexer

func init() {
	indexer.cfg = &config.IndexConfig{}
	config.SetupLogFlags(&indexer.cfg.Log, indexCmd)
	config.SetupDatabaseFlags(&indexer.cfg.Database, indexCmd)
	config.SetupProbeFlags(&indexer.cfg.Probe, indexCmd)
	config.SetupThrottlingFlag(&indexer.cfg.Base.Throttling, indexCmd)
	config.SetupIndexSpecificFlags(indexer.cfg, indexCmd)

	rootCmd.AddCommand(indexCmd)
}

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Indexes the blockchain according to the configuration defined.",
	Long: `Indexes the Cosmos-based blockchain according to the configurations found on the command line
	or in the specified config file. Indexes taxable events into a database for easy querying. It is
	highly recommended to keep this command running as a background service to keep your index up to date.`,
	PreRunE: setupIndex,
	Run:     index,
}

func setupIndex(cmd *cobra.Command, args []string) error {
	bindFlags(cmd, viperConf)

	err := indexer.cfg.Validate()
	if err != nil {
		return err
	}

	ignoredKeys := config.CheckSuperfluousIndexKeys(viperConf.AllKeys())

	if len(ignoredKeys) > 0 {
		config.Log.Warnf("Warning, the following invalid keys will be ignored: %v", ignoredKeys)
	}

	setupLogger(indexer.cfg.Log.Level, indexer.cfg.Log.Path, indexer.cfg.Log.Pretty)

	// 0 is an invalid starting block, set it to 1
	if indexer.cfg.Base.StartBlock == 0 {
		indexer.cfg.Base.StartBlock = 1
	}

	db, err := connectToDBAndMigrate(indexer.cfg.Database)
	if err != nil {
		config.Log.Fatal("Could not establish connection to the database", err)
	}

	indexer.db = db

	indexer.scheduler = gocron.NewScheduler(time.UTC)

	// We should stop relying on the denom cache now that we are running this as a CLI tool only
	dbTypes.CacheDenoms(db)
	dbTypes.CacheIBCDenoms(db)

	indexer.dryRun = indexer.cfg.Base.Dry

	return nil
}

// The Indexer struct is used to perform index operations

func setupIndexer() *Indexer {
	var err error

	// Setup chain specific stuff
	core.SetupAddressRegex(indexer.cfg.Probe.AccountPrefix + "(valoper)?1[a-z0-9]{38}")
	core.SetupAddressPrefix(indexer.cfg.Probe.AccountPrefix)
	core.ChainSpecificMessageTypeHandlerBootstrap(indexer.cfg.Probe.ChainID)
	core.ChainSpecificBeginBlockerEventTypeHandlerBootstrap(indexer.cfg.Probe.ChainID)
	core.ChainSpecificEndBlockerEventTypeHandlerBootstrap(indexer.cfg.Probe.ChainID)

	config.SetChainConfig(indexer.cfg.Probe.AccountPrefix)

	// Setup scheduler to periodically update denoms
	if indexer.cfg.Base.API != "" {
		_, err = indexer.scheduler.Every(6).Hours().Do(tasks.IBCDenomUpsertTask, indexer.cfg.Base.API, indexer.db)
		if err != nil {
			config.Log.Error("Error scheduling ibc denom upsert task. Err: ", err)
		}

		indexer.scheduler.StartAsync()
	}

	// Some chains do not have the denom metadata URL available on chain, so we do chain specific downloads instead.
	tasks.DoChainSpecificUpsertDenoms(indexer.db, indexer.cfg.Probe.ChainID, indexer.cfg.Base.RequestRetryAttempts, indexer.cfg.Base.RequestRetryMaxWait)
	indexer.cl = probe.GetProbeClient(indexer.cfg.Probe)

	// Depending on the app configuration, wait for the chain to catch up
	chainCatchingUp, err := rpc.IsCatchingUp(indexer.cl)
	for indexer.cfg.Base.WaitForChain && chainCatchingUp && err == nil {
		// Wait between status checks, don't spam the node with requests
		config.Log.Debug("Chain is still catching up, please wait or disable check in config.")
		time.Sleep(time.Second * time.Duration(indexer.cfg.Base.WaitForChainDelay))
		chainCatchingUp, err = rpc.IsCatchingUp(indexer.cl)

		// This EOF error pops up from time to time and is unpredictable
		// It is most likely an error on the node, we would need to see any error logs on the node side
		// Try one more time
		if err != nil && strings.HasSuffix(err.Error(), "EOF") {
			time.Sleep(time.Second * time.Duration(indexer.cfg.Base.WaitForChainDelay))
			chainCatchingUp, err = rpc.IsCatchingUp(indexer.cl)
		}
	}
	if err != nil {
		config.Log.Fatal("Error querying chain status.", err)
	}

	return &indexer
}

func index(cmd *cobra.Command, args []string) {
	// Setup the indexer with config, db, and cl
	idxr := setupIndexer()
	dbConn, err := idxr.db.DB()
	if err != nil {
		config.Log.Fatal("Failed to connect to DB", err)
	}
	defer dbConn.Close()

	// blockChans are just the block heights; limit max jobs in the queue, otherwise this queue would contain one
	// item (block height) for every block on the entire blockchain we're indexing. Furthermore, once the queue
	// is close to empty, we will spin up a new thread to fill it up with new jobs.
	blockEnqueueChan := make(chan *core.EnqueueData, 10000)

	// This channel represents query job results for the RPC queries to Cosmos Nodes. Every time an RPC query
	// completes, the query result will be sent to this channel (for later processing by a different thread).
	// Realistically, I expect that RPC queries will be slower than our relational DB on the local network.
	// If RPC queries are faster than DB inserts this buffer will fill up.
	// We will periodically check the buffer size to monitor performance so we can optimize later.
	rpcQueryThreads := int(idxr.cfg.Base.RPCWorkers)
	if rpcQueryThreads == 0 {
		rpcQueryThreads = 4
	} else if rpcQueryThreads > 64 {
		rpcQueryThreads = 64
	}

	var wg sync.WaitGroup // This group is to ensure we are done processing transactions and events before returning

	chain := models.Chain{
		ChainID: idxr.cfg.Probe.ChainID,
		Name:    idxr.cfg.Probe.ChainName,
	}

	dbChainID, err := dbTypes.GetDBChainID(idxr.db, chain)
	if err != nil {
		config.Log.Fatal("Failed to add/create chain in DB", err)
	}

	// This block consolidates all base RPC requests into one worker.
	// Workers read from the enqueued blocks and query blockchain data from the RPC server.
	var blockRPCWaitGroup sync.WaitGroup
	blockRPCWorkerDataChan := make(chan core.IndexerBlockEventData, 10)
	for i := 0; i < rpcQueryThreads; i++ {
		blockRPCWaitGroup.Add(1)
		go core.BlockRPCWorker(&blockRPCWaitGroup, blockEnqueueChan, dbChainID, idxr.cfg.Probe.ChainID, idxr.cfg, idxr.cl, idxr.db, blockRPCWorkerDataChan)
	}

	go func() {
		blockRPCWaitGroup.Wait()
		close(blockRPCWorkerDataChan)
	}()

	// Block BeginBlocker and EndBlocker indexing requirements. Indexes block events that took place in the BeginBlock and EndBlock state transitions
	blockEventsDataChan := make(chan *blockEventsDBData, 4*rpcQueryThreads)
	txDataChan := make(chan *dbData, 4*rpcQueryThreads)

	wg.Add(1)
	go idxr.processBlocks(&wg, core.HandleFailedBlock, blockRPCWorkerDataChan, blockEventsDataChan, txDataChan, dbChainID)

	wg.Add(1)
	go idxr.doDBUpdates(&wg, txDataChan, blockEventsDataChan, dbChainID)

	switch {
	// If block enqueue function has been explicitly set, use that
	case idxr.blockEnqueueFunction != nil:
	// Default block enqueue functions based on config values
	case idxr.cfg.Base.ReindexMessageType != "":
		idxr.blockEnqueueFunction, err = core.GenerateMsgTypeEnqueueFunction(idxr.db, *idxr.cfg, dbChainID, idxr.cfg.Base.ReindexMessageType)
		if err != nil {
			config.Log.Fatal("Failed to generate block enqueue function", err)
		}
	case idxr.cfg.Base.BlockInputFile != "":
		idxr.blockEnqueueFunction, err = core.GenerateBlockFileEnqueueFunction(idxr.db, *idxr.cfg, idxr.cl, dbChainID, idxr.cfg.Base.BlockInputFile)
		if err != nil {
			config.Log.Fatal("Failed to generate block enqueue function", err)
		}
	default:
		idxr.blockEnqueueFunction, err = core.GenerateDefaultEnqueueFunction(idxr.db, *idxr.cfg, idxr.cl, dbChainID)
		if err != nil {
			config.Log.Fatal("Failed to generate block enqueue function", err)
		}
	}

	err = idxr.blockEnqueueFunction(blockEnqueueChan)
	if err != nil {
		config.Log.Fatal("Block enqueue failed", err)
	}

	close(blockEnqueueChan)

	// If we error out in the main loop, this will block. Meaning we may not know of an error for 6 hours until last scheduled task stops
	idxr.scheduler.Stop()
	wg.Wait()
}

func GetBlockEventsStartIndexHeight(db *gorm.DB, chainID uint) int64 {
	block, err := dbTypes.GetHighestEventIndexedBlock(db, chainID)
	if err != nil && err.Error() != "record not found" {
		log.Fatalf("Cannot retrieve highest indexed block event. Err: %v", err)
	}

	return block.Height
}

// GetIndexerStartingHeight will determine which block to start at
// if start block is set to -1, it will start at the highest block indexed
// otherwise, it will start at the first missing block between the start and end height
func (idxr *Indexer) GetIndexerStartingHeight(chainID uint) int64 {
	// If the start height is set to -1, resume from the highest block already indexed
	if idxr.cfg.Base.StartBlock == -1 {
		latestBlock, err := rpc.GetLatestBlockHeight(idxr.cl)
		if err != nil {
			log.Fatalf("Error getting blockchain latest height. Err: %v", err)
		}

		fmt.Println("Found latest block", latestBlock)
		highestIndexedBlock := dbTypes.GetHighestIndexedBlock(idxr.db, chainID)
		if highestIndexedBlock.Height < latestBlock {
			return highestIndexedBlock.Height + 1
		}
	}

	// if we are re-indexing, just start at the configured start block
	if idxr.cfg.Base.ReIndex {
		return idxr.cfg.Base.StartBlock
	}

	maxStart := idxr.cfg.Base.EndBlock
	if maxStart == -1 {
		heighestBlock := dbTypes.GetHighestIndexedBlock(idxr.db, chainID)
		maxStart = heighestBlock.Height
	}

	// Otherwise, start at the first block after the configured start block that we have not yet indexed.
	return dbTypes.GetFirstMissingBlockInRange(idxr.db, idxr.cfg.Base.StartBlock, maxStart, chainID)
}

type dbData struct {
	txDBWrappers []dbTypes.TxDBWrapper
	blockTime    time.Time
	blockHeight  int64
}

type blockEventsDBData struct {
	blockDBWrapper *dbTypes.BlockDBWrapper
	blockTime      time.Time
	blockHeight    int64
}

// This function is responsible for processing raw RPC data into app-usable types. It handles both block events and transactions.
// It parses each dataset according to the application configuration requirements and passes the data to the channels that handle the parsed data.
func (idxr *Indexer) processBlocks(wg *sync.WaitGroup, failedBlockHandler core.FailedBlockHandler, blockRPCWorkerChan chan core.IndexerBlockEventData, blockEventsDataChan chan *blockEventsDBData, txDataChan chan *dbData, chainID uint) {
	defer close(blockEventsDataChan)
	defer close(txDataChan)
	defer wg.Done()

	for blockData := range blockRPCWorkerChan {
		currentHeight := blockData.BlockData.Block.Height
		config.Log.Infof("Parsing data for block %d", currentHeight)

		if blockData.IndexBlockEvents && !blockData.BlockEventRequestsFailed {
			config.Log.Info("Parsing block events")
			blockDBWrapper, err := core.ProcessRPCBlockResults(blockData.BlockResultsData)
			if err != nil {
				config.Log.Errorf("Failed to process block events during block %d event processing, adding to failed block events table", currentHeight)
				failedBlockHandler(currentHeight, core.FailedBlockEventHandling, err)
				err := dbTypes.UpsertFailedEventBlock(idxr.db, currentHeight, idxr.cfg.Probe.ChainID, idxr.cfg.Probe.ChainName)
				if err != nil {
					config.Log.Fatal("Failed to insert failed block event", err)
				}
			} else {
				config.Log.Infof("Finished parsing block event data for block %d", currentHeight)

				blockEventsDataChan <- &blockEventsDBData{
					blockHeight:    currentHeight,
					blockTime:      blockData.BlockData.Block.Time,
					blockDBWrapper: blockDBWrapper,
				}
			}
		}

		if blockData.IndexTransactions && !blockData.TxRequestsFailed {
			config.Log.Info("Parsing transactions")
			var txDBWrappers []dbTypes.TxDBWrapper
			var err error

			if blockData.GetTxsResponse != nil {
				txDBWrappers, _, err = core.ProcessRPCTXs(idxr.db, blockData.GetTxsResponse)
			} else if blockData.BlockResultsData != nil {
				txDBWrappers, _, err = core.ProcessRPCBlockByHeightTXs(idxr.db, idxr.cl, blockData.BlockData, blockData.BlockResultsData)
			}

			if err != nil {
				config.Log.Error("ProcessRpcTxs: unhandled error", err)
				failedBlockHandler(currentHeight, core.UnprocessableTxError, err)
				err := dbTypes.UpsertFailedBlock(idxr.db, currentHeight, idxr.cfg.Probe.ChainID, idxr.cfg.Probe.ChainName)
				if err != nil {
					config.Log.Fatal("Failed to insert failed block", err)
				}
			} else {
				txDataChan <- &dbData{
					txDBWrappers: txDBWrappers,
					blockTime:    blockData.BlockData.Block.Time,
					blockHeight:  currentHeight,
				}
			}

		}
	}
}

// doDBUpdates will read the data out of the db data chan that had been processed by the workers
// if this is a dry run, we will simply empty the channel and track progress
// otherwise we will index the data in the DB.
// it will also read rewars data and index that.
func (idxr *Indexer) doDBUpdates(wg *sync.WaitGroup, txDataChan chan *dbData, blockEventsDataChan chan *blockEventsDBData, dbChainID uint) {
	blocksProcessed := 0
	dbWrites := 0
	dbReattempts := 0
	timeStart := time.Now()
	defer wg.Done()

	for {
		// break out of loop once all channels are fully consumed
		if txDataChan == nil && blockEventsDataChan == nil {
			config.Log.Info("DB updates complete")
			break
		}

		select {
		// read tx data from the data chan
		case data, ok := <-txDataChan:
			if !ok {
				txDataChan = nil
				continue
			}
			dbWrites++
			// While debugging we'll sometimes want to turn off INSERTS to the DB
			// Note that this does not turn off certain reads or DB connections.
			if !idxr.dryRun {
				config.Log.Info(fmt.Sprintf("Indexing %v TXs from block %d", len(data.txDBWrappers), data.blockHeight))
				err := dbTypes.IndexNewBlock(idxr.db, data.blockHeight, data.blockTime, data.txDBWrappers, dbChainID)
				if err != nil {
					// Do a single reattempt on failure
					dbReattempts++
					err = dbTypes.IndexNewBlock(idxr.db, data.blockHeight, data.blockTime, data.txDBWrappers, dbChainID)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Error indexing block %v.", data.blockHeight), err)
					}
				}
			} else {
				config.Log.Info(fmt.Sprintf("Processing block %d (dry run, block data will not be stored in DB).", data.blockHeight))
			}

			// Just measuring how many blocks/second we can process
			if idxr.cfg.Base.BlockTimer > 0 {
				blocksProcessed++
				if blocksProcessed%int(idxr.cfg.Base.BlockTimer) == 0 {
					totalTime := time.Since(timeStart)
					config.Log.Info(fmt.Sprintf("Processing %d blocks took %f seconds. %d total blocks have been processed.\n", idxr.cfg.Base.BlockTimer, totalTime.Seconds(), blocksProcessed))
					timeStart = time.Now()
				}
				if float64(dbReattempts)/float64(dbWrites) > .1 {
					config.Log.Fatalf("More than 10%% of the last %v DB writes have failed.", dbWrites)
				}
			}
		case eventData, ok := <-blockEventsDataChan:
			if !ok {
				blockEventsDataChan = nil
				continue
			}
			dbWrites++
			numEvents := len(eventData.blockDBWrapper.BeginBlockEvents) + len(eventData.blockDBWrapper.EndBlockEvents)
			config.Log.Info(fmt.Sprintf("Indexing %v Block Events from block %d", numEvents, eventData.blockHeight))
			identifierLoggingString := fmt.Sprintf("block %d", eventData.blockHeight)

			err := dbTypes.IndexBlockEvents(idxr.db, idxr.dryRun, eventData.blockHeight, eventData.blockTime, eventData.blockDBWrapper, dbChainID, idxr.cfg.Probe.ChainName, identifierLoggingString)
			if err != nil {
				// TODO: Should we reattempt here still?
				// Do a single reattempt on failure
				// dbReattempts++
				// err = dbTypes.IndexBlockEvents(idxr.db, idxr.dryRun, eventData.blockHeight, eventData.blockTime, eventData.blockDBWrapper, dbChainID, idxr.cfg.Probe.ChainName, identifierLoggingString)
				// if err != nil {
				config.Log.Fatal(fmt.Sprintf("Error indexing block events for %s.", identifierLoggingString), err)
				// }
			}
			config.Log.Info(fmt.Sprintf("Finished indexing %v Block Events from block %d", numEvents, eventData.blockHeight))
		}
	}
}
