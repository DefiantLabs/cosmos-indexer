package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/nodersteam/cosmos-indexer/pkg/consumer"
	"github.com/nodersteam/cosmos-indexer/pkg/model"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nodersteam/cosmos-indexer/pkg/repository"
	"github.com/nodersteam/cosmos-indexer/pkg/server"
	"github.com/nodersteam/cosmos-indexer/pkg/service"
	blocks "github.com/nodersteam/cosmos-indexer/proto"
	"google.golang.org/grpc"

	"github.com/DefiantLabs/probe/client"

	"github.com/nodersteam/cosmos-indexer/config"
	"github.com/nodersteam/cosmos-indexer/core"
	dbTypes "github.com/nodersteam/cosmos-indexer/db"
	"github.com/nodersteam/cosmos-indexer/db/models"
	"github.com/nodersteam/cosmos-indexer/filter"
	"github.com/nodersteam/cosmos-indexer/parsers"
	"github.com/nodersteam/cosmos-indexer/probe"
	"github.com/nodersteam/cosmos-indexer/rpc"
	"github.com/spf13/cobra"

	migrate "github.com/xakep666/mongo-migrate"
	"gorm.io/gorm"
)

type Indexer struct {
	cfg                                 *config.IndexConfig
	dryRun                              bool
	db                                  *gorm.DB
	cl                                  *client.ChainClient
	blockEnqueueFunction                func(chan *core.EnqueueData) error
	blockEventFilterRegistries          blockEventFilterRegistries
	messageTypeFilters                  []filter.MessageTypeFilter
	customBeginBlockEventParserRegistry map[string][]parsers.BlockEventParser // Used for associating parsers to block event types in BeginBlock events
	customEndBlockEventParserRegistry   map[string][]parsers.BlockEventParser // Used for associating parsers to block event types in EndBlock events
	customBeginBlockParserTrackers      map[string]models.BlockEventParser    // Used for tracking block event parsers in the database
	customEndBlockParserTrackers        map[string]models.BlockEventParser    // Used for tracking block event parsers in the database
	customModels                        []any
}

type blockEventFilterRegistries struct {
	beginBlockEventFilterRegistry *filter.StaticBlockEventFilterRegistry
	endBlockEventFilterRegistry   *filter.StaticBlockEventFilterRegistry
}

var indexer Indexer

func init() {
	indexer.cfg = &config.IndexConfig{}
	config.SetupLogFlags(&indexer.cfg.Log, indexCmd)
	config.SetupDatabaseFlags(&indexer.cfg.Database, indexCmd)
	config.SetupProbeFlags(&indexer.cfg.Probe, indexCmd)
	config.SetupServerFlags(&indexer.cfg.Server, indexCmd)
	config.SetupThrottlingFlag(&indexer.cfg.Base.Throttling, indexCmd)
	config.SetupIndexSpecificFlags(indexer.cfg, indexCmd)
	config.SetupRedisFlags(&indexer.cfg.RedisConf, indexCmd)
	config.SetupMongoDBFlags(&indexer.cfg.MongoConf, indexCmd)

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

func RegisterCustomBeginBlockEventParser(eventKey string, parser parsers.BlockEventParser) {
	var err error
	indexer.customBeginBlockEventParserRegistry, indexer.customBeginBlockParserTrackers, err = customBlockEventRegistration(
		indexer.customBeginBlockEventParserRegistry,
		indexer.customBeginBlockParserTrackers,
		eventKey,
		parser,
		models.BeginBlockEvent,
	)

	if err != nil {
		config.Log.Fatal("Error registering BeginBlock custom parser", err)
	}
}

func RegisterCustomEndBlockEventParser(eventKey string, parser parsers.BlockEventParser) {
	var err error
	indexer.customEndBlockEventParserRegistry, indexer.customEndBlockParserTrackers, err = customBlockEventRegistration(
		indexer.customEndBlockEventParserRegistry,
		indexer.customEndBlockParserTrackers,
		eventKey,
		parser,
		models.EndBlockEvent,
	)

	if err != nil {
		config.Log.Fatal("Error registering EndBlock custom parser", err)
	}
}

func customBlockEventRegistration(registry map[string][]parsers.BlockEventParser, tracker map[string]models.BlockEventParser, eventKey string, parser parsers.BlockEventParser, lifecycleValue models.BlockLifecyclePosition) (map[string][]parsers.BlockEventParser, map[string]models.BlockEventParser, error) {
	if registry == nil {
		registry = make(map[string][]parsers.BlockEventParser)
	}

	if tracker == nil {
		tracker = make(map[string]models.BlockEventParser)
	}

	registry[eventKey] = append(registry[eventKey], parser)

	if _, ok := tracker[parser.Identifier()]; ok {
		return registry, tracker, fmt.Errorf("found duplicate block event parser with identifier \"%s\", parsers must be uniquely identified", parser.Identifier())
	}

	tracker[parser.Identifier()] = models.BlockEventParser{
		Identifier:             parser.Identifier(),
		BlockLifecyclePosition: lifecycleValue,
	}
	return registry, tracker, nil
}

func RegisterCustomModels(models []any) {
	indexer.customModels = models
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

	indexer.dryRun = indexer.cfg.Base.Dry

	indexer.blockEventFilterRegistries = blockEventFilterRegistries{
		beginBlockEventFilterRegistry: &filter.StaticBlockEventFilterRegistry{},
		endBlockEventFilterRegistry:   &filter.StaticBlockEventFilterRegistry{},
	}

	if indexer.cfg.Base.FilterFile != "" {
		f, err := os.Open(indexer.cfg.Base.FilterFile)
		if err != nil {
			config.Log.Fatalf("Failed to open block event filter file %s: %s", indexer.cfg.Base.FilterFile, err)
		}

		b, err := io.ReadAll(f)
		if err != nil {
			config.Log.Fatal("Failed to parse block event filter config", err)
		}

		indexer.blockEventFilterRegistries.beginBlockEventFilterRegistry.BlockEventFilters,
			indexer.blockEventFilterRegistries.beginBlockEventFilterRegistry.RollingWindowEventFilters,
			indexer.blockEventFilterRegistries.endBlockEventFilterRegistry.BlockEventFilters,
			indexer.blockEventFilterRegistries.endBlockEventFilterRegistry.RollingWindowEventFilters,
			indexer.messageTypeFilters,
			err = config.ParseJSONFilterConfig(b)

		if err != nil {
			config.Log.Fatal("Failed to parse block event filter config", err)
		}

	}

	if len(indexer.customModels) != 0 {
		err = dbTypes.MigrateInterfaces(indexer.db, indexer.customModels)
		if err != nil {
			config.Log.Fatal("Failed to migrate custom models", err)
		}
	}

	if len(indexer.customBeginBlockParserTrackers) != 0 {
		err = dbTypes.FindOrCreateCustomParsers(indexer.db, indexer.customBeginBlockParserTrackers)
		if err != nil {
			config.Log.Fatal("Failed to migrate custom block event parsers", err)
		}
	}

	if len(indexer.customEndBlockParserTrackers) != 0 {
		err = dbTypes.FindOrCreateCustomParsers(indexer.db, indexer.customEndBlockParserTrackers)
		if err != nil {
			config.Log.Fatal("Failed to migrate custom block event parsers", err)
		}
	}
	return nil
}

// The Indexer struct is used to perform index operations

func setupIndexer() *Indexer {
	var err error

	// Setup chain specific stuff
	core.ChainSpecificMessageTypeHandlerBootstrap(indexer.cfg.Probe.ChainID)

	config.SetChainConfig(indexer.cfg.Probe.AccountPrefix)

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

	// inbound grpc server
	ctx := context.Background()

	grpcServUrl := fmt.Sprintf(":%d", idxr.cfg.Server.Port)
	listener, err := net.Listen("tcp", grpcServUrl)
	if err != nil {
		config.Log.Fatal("Unable to run listener", err)
	}

	dbDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		idxr.cfg.Database.User, idxr.cfg.Database.Password, idxr.cfg.Database.Host, idxr.cfg.Database.Port, idxr.cfg.Database.Database)
	dbConnRepo, err := connectPgxPool(ctx, dbDSN)
	if err != nil {
		config.Log.Fatal("Error connecting DB", err)
	}
	repoBlocks := repository.NewBlocks(dbConnRepo)
	srvBlocks := service.NewBlocks(repoBlocks)

	repoTxs := repository.NewTxs(dbConnRepo)
	srvTxs := service.NewTxs(repoTxs)

	// setup mongoDB
	mongoClient, err := mongo.Connect(ctx, options.Client().ApplyURI(idxr.cfg.MongoConf.MongoAddr))
	if err != nil {
		panic(err)
	}
	defer func() {
		if err = mongoClient.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()
	err = mongoClient.Ping(ctx, &readpref.ReadPref{})
	if err != nil {
		panic(err)
	}

	db := mongoClient.Database(idxr.cfg.MongoConf.MongoDB)
	searchRepo := repository.NewSearch(db)
	srvSearch := service.NewSearch(searchRepo)

	// setup cache
	rdb := redis.NewClient(&redis.Options{
		Addr:     idxr.cfg.RedisConf.RedisAddr,
		Password: idxr.cfg.RedisConf.RedisPsw,
		DB:       0, // use default DB
	})

	ctxPing, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err = rdb.Ping(ctxPing).Err(); err != nil {
		panic(err)
	}
	cache := repository.NewCache(rdb)

	blocksServer := server.NewBlocksServer(srvBlocks, srvTxs, srvSearch, *cache)
	size := 1024 * 1024 * 50
	grpcServer := grpc.NewServer(
		grpc.MaxSendMsgSize(size),
		grpc.MaxRecvMsgSize(size))
	blocks.RegisterBlocksServiceServer(grpcServer, blocksServer)
	go func() {
		log.Println("blocks server started: " + grpcServUrl)
		if err = grpcServer.Serve(listener); err != nil {
			log.Fatal(err)
			grpcServer.GracefulStop()
			return
		}
	}()

	chBlocks := make(chan *model.BlockInfo, 1000)
	defer close(chBlocks)
	chTxs := make(chan *models.Tx, 1000)
	defer close(chTxs)

	cacheConsumer := consumer.NewCacheConsumer(cache, chBlocks, chTxs, cache)
	go cacheConsumer.RunBlocks(ctx)
	go cacheConsumer.RunTransactions(ctx)
	defer ctx.Done()

	aggregatesConsumer := consumer.NewAggregatesConsumer(cache, repoBlocks, repoTxs)
	go aggregatesConsumer.Consume(ctx)

	wg.Add(1)
	go idxr.processBlocks(&wg, core.HandleFailedBlock,
		blockRPCWorkerDataChan,
		blockEventsDataChan,
		txDataChan,
		dbChainID,
		indexer.blockEventFilterRegistries,
		chBlocks,
		*cache)

	wg.Add(1)
	go idxr.doDBUpdates(&wg, txDataChan, blockEventsDataChan, dbChainID, chTxs, repoTxs, cache)

	// search index
	txSearchConsumer := consumer.NewSearchTxConsumer(rdb, "pub/txs", searchRepo) // TODO
	go txSearchConsumer.Consume(ctx)

	blSearchConsumer := consumer.NewSearchBlocksConsumer(rdb, "pub/blocks", searchRepo) // TODO
	go blSearchConsumer.Consume(ctx)

	// migration
	go func() {
		config.Log.Info("Starting migration")
		db, err = mongoDBMigrate(ctx, db, dbConnRepo, searchRepo)
		if err != nil {
			panic(err)
		}
		config.Log.Info("Migration complete")
	}()

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

	wg.Wait()
}

func mongoDBMigrate(ctx context.Context,
	db *mongo.Database,
	pg *pgxpool.Pool, search repository.Search) (*mongo.Database, error) {
	m := migrate.NewMigrate(db, migrate.Migration{
		Version:     1,
		Description: "add unique index idx_txhash_type",
		Up: func(ctx context.Context, db *mongo.Database) error {
			opt := options.Index().SetName("idx_txhash_type").SetUnique(true)
			keys := bson.D{{"tx_hash", 1}, {"type", 1}}
			model := mongo.IndexModel{Keys: keys, Options: opt}
			_, err := db.Collection("search").Indexes().CreateOne(ctx, model)
			if err != nil {
				return err
			}

			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			_, err := db.Collection("search").Indexes().DropOne(ctx, "idx_txhash_type")
			if err != nil {
				return err
			}
			return nil
		},
	}, migrate.Migration{ // TODO not the best place to migrate data
		Version:     2,
		Description: "migrate existing hashes",
		Up: func(ctx context.Context, db *mongo.Database) error {
			config.Log.Info("starting txs v2 migration")
			rows, err := pg.Query(ctx, `select distinct hash from txes`)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			} else {
				for rows.Next() {
					var txHash string
					if err = rows.Scan(&txHash); err != nil {
						return err
					}
					if err = search.AddHash(context.Background(), txHash, "transaction", 0); err != nil {
						log.Println(err)
					}
				}
			}

			config.Log.Info("starting blocks migration")
			rows, err = pg.Query(ctx, `select distinct block_hash from blocks`)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			} else {
				for rows.Next() {
					var txHash string
					if err = rows.Scan(&txHash); err != nil {
						return err
					}
					if err = search.AddHash(context.Background(), txHash, "block", 0); err != nil {
						log.Println(err)
					}
				}
			}

			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			// ignoring, what's done is done.
			return nil
		},
	}, migrate.Migration{ // TODO not the best place to migrate data
		Version:     3,
		Description: "migrate existing hashes with block height",
		Up: func(ctx context.Context, db *mongo.Database) error {
			config.Log.Info("starting txs v3 migration")
			db.Collection("search").Drop(ctx)

			rows, err := pg.Query(ctx, `select distinct hash from txes`)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			} else {
				for rows.Next() {
					var txHash string
					if err = rows.Scan(&txHash); err != nil {
						return err
					}
					if err = search.AddHash(context.Background(), txHash, "transaction", 0); err != nil {
						log.Println(err)
					}
				}
			}

			config.Log.Info("starting blocks migration")
			rows, err = pg.Query(ctx, `select distinct block_hash,height from blocks`)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return err
			} else {
				for rows.Next() {
					var txHash string
					var blockHeight int64
					if err = rows.Scan(&txHash, &blockHeight); err != nil {
						return err
					}
					if err = search.AddHash(context.Background(), txHash, "block", blockHeight); err != nil {
						log.Println(err)
					}
				}
			}

			return nil
		},
		Down: func(ctx context.Context, db *mongo.Database) error {
			// ignoring, what's done is done.
			return nil
		}})
	if err := m.Up(ctx, migrate.AllAvailable); err != nil {
		return nil, err
	}

	return db, nil
}

// connectPgxPool establishes a connection to a PostgreSQL database.
func connectPgxPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect: %v", err)
	}

	if err = conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping: %v", err)
	}

	return conn, nil
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
	block        models.Block
}

type blockEventsDBData struct {
	blockDBWrapper *dbTypes.BlockDBWrapper
}

// This function is responsible for processing raw RPC data into app-usable types. It handles both block events and transactions.
// It parses each dataset according to the application configuration requirements and passes the data to the channels that handle the parsed data.
func (idxr *Indexer) processBlocks(wg *sync.WaitGroup,
	failedBlockHandler core.FailedBlockHandler,
	blockRPCWorkerChan chan core.IndexerBlockEventData,
	blockEventsDataChan chan *blockEventsDBData,
	txDataChan chan *dbData,
	chainID uint,
	blockEventFilterRegistry blockEventFilterRegistries,
	blocksCh chan *model.BlockInfo,
	cache repository.Cache) {

	defer close(blockEventsDataChan)
	defer close(txDataChan)
	defer wg.Done()

	for blockData := range blockRPCWorkerChan {
		currentHeight := blockData.BlockData.Block.Height
		config.Log.Infof("Parsing data for block %d", currentHeight)

		block, err := core.ProcessBlock(blockData.BlockData, blockData.BlockResultsData, chainID)
		if err != nil {
			config.Log.Error("ProcessBlock: unhandled error", err)
			failedBlockHandler(currentHeight, core.UnprocessableTxError, err)
			err := dbTypes.UpsertFailedBlock(idxr.db, currentHeight, idxr.cfg.Probe.ChainID, idxr.cfg.Probe.ChainName)
			if err != nil {
				config.Log.Fatal("Failed to insert failed block", err)
			}
			continue
		}

		if blockData.IndexBlockEvents && !blockData.BlockEventRequestsFailed {
			config.Log.Info("Parsing block events")
			blockDBWrapper, err := core.ProcessRPCBlockResults(*indexer.cfg, block, blockData.BlockResultsData, indexer.customBeginBlockEventParserRegistry, indexer.customEndBlockEventParserRegistry)
			if err != nil {
				config.Log.Errorf("Failed to process block events during block %d event processing, adding to failed block events table", currentHeight)
				failedBlockHandler(currentHeight, core.FailedBlockEventHandling, err)
				err := dbTypes.UpsertFailedEventBlock(idxr.db, currentHeight, idxr.cfg.Probe.ChainID, idxr.cfg.Probe.ChainName)
				if err != nil {
					config.Log.Fatal("Failed to insert failed block event", err)
				}
			} else {
				config.Log.Infof("Finished parsing block event data for block %d", currentHeight)

				var beginBlockFilterError error
				var endBlockFilterError error
				if blockEventFilterRegistry.beginBlockEventFilterRegistry != nil && blockEventFilterRegistry.beginBlockEventFilterRegistry.NumFilters() > 0 {
					blockDBWrapper.BeginBlockEvents, beginBlockFilterError = core.FilterRPCBlockEvents(blockDBWrapper.BeginBlockEvents, *blockEventFilterRegistry.beginBlockEventFilterRegistry)
				}

				if blockEventFilterRegistry.endBlockEventFilterRegistry != nil && blockEventFilterRegistry.endBlockEventFilterRegistry.NumFilters() > 0 {
					blockDBWrapper.EndBlockEvents, endBlockFilterError = core.FilterRPCBlockEvents(blockDBWrapper.EndBlockEvents, *blockEventFilterRegistry.endBlockEventFilterRegistry)
				}

				if beginBlockFilterError == nil && endBlockFilterError == nil {
					blockEventsDataChan <- &blockEventsDBData{
						blockDBWrapper: blockDBWrapper,
					}
				} else {
					config.Log.Errorf("Failed to filter block events during block %d event processing, adding to failed block events table. Begin blocker filter error %s. End blocker filter error %s", currentHeight, beginBlockFilterError, endBlockFilterError)
					failedBlockHandler(currentHeight, core.FailedBlockEventHandling, err)
					err := dbTypes.UpsertFailedEventBlock(idxr.db, currentHeight, idxr.cfg.Probe.ChainID, idxr.cfg.Probe.ChainName)
					if err != nil {
						config.Log.Fatal("Failed to insert failed block event", err)
					}
				}
			}
		}

		if blockData.IndexTransactions && !blockData.TxRequestsFailed {
			config.Log.Info("Parsing transactions")
			var txDBWrappers []dbTypes.TxDBWrapper
			var err error

			if blockData.GetTxsResponse != nil {
				config.Log.Debug("Processing TXs from RPC TX Search response")
				txDBWrappers, _, err = core.ProcessRPCTXs(idxr.db, idxr.cl, idxr.messageTypeFilters, blockData.GetTxsResponse)
			} else if blockData.BlockResultsData != nil {
				config.Log.Debug("Processing TXs from BlockResults search response")
				txDBWrappers, _, err = core.ProcessRPCBlockByHeightTXs(idxr.db, idxr.cl, idxr.messageTypeFilters, blockData.BlockData, blockData.BlockResultsData)
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
					block:        block,
				}
			}
		}
		blocksCh <- idxr.toBlockInfo(block)
		if err := cache.PublishBlock(context.Background(), &block); err != nil {
			config.Log.Error("Failed to publish block info", err)
		}
	}
}

func (idxr *Indexer) toBlockInfo(in models.Block) *model.BlockInfo {
	return &model.BlockInfo{
		BlockHeight:              in.Height,
		ProposedValidatorAddress: in.ProposerConsAddress.Address,
		TotalTx:                  int64(in.TotalTxs),
		GenerationTime:           in.TimeStamp,
		BlockHash:                in.BlockHash,
	}
}

// doDBUpdates will read the data out of the db data chan that had been processed by the workers
// if this is a dry run, we will simply empty the channel and track progress
// otherwise we will index the data in the DB.
// it will also read rewars data and index that.
func (idxr *Indexer) doDBUpdates(wg *sync.WaitGroup,
	txDataChan chan *dbData,
	blockEventsDataChan chan *blockEventsDBData,
	dbChainID uint,
	txsCh chan *models.Tx,
	txRepo repository.Txs,
	cache repository.PubSubCache) {

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
				config.Log.Info(fmt.Sprintf("Indexing %v TXs from block %d", len(data.txDBWrappers), data.block.Height))
				_, _, err := dbTypes.IndexNewBlock(idxr.db, data.block, data.txDBWrappers, *idxr.cfg)
				if err != nil {
					// Do a single reattempt on failure
					dbReattempts++
					_, _, err = dbTypes.IndexNewBlock(idxr.db, data.block, data.txDBWrappers, *idxr.cfg)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Error indexing block %v.", data.block.Height), err)
					}
				}
			} else {
				config.Log.Info(fmt.Sprintf("Processing block %d (dry run, block data will not be stored in DB).", data.block.Height))
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

			for _, tx := range data.txDBWrappers {
				transaction := tx.Tx

				transaction.Block = data.block
				res, err := txRepo.GetSenderAndReceiver(context.Background(), transaction.Hash)
				if err != nil {
					config.Log.Error("unable to find sender and receiver", err)
				}
				transaction.SenderReceiver = res
				// TODO decomposite everything
				txsCh <- &transaction
				if err := cache.PublishTx(context.Background(), &transaction); err != nil {
					config.Log.Error(err.Error())
				}
			}

		case eventData, ok := <-blockEventsDataChan:
			if !ok {
				blockEventsDataChan = nil
				continue
			}
			dbWrites++
			numEvents := len(eventData.blockDBWrapper.BeginBlockEvents) + len(eventData.blockDBWrapper.EndBlockEvents)
			config.Log.Info(fmt.Sprintf("Indexing %v Block Events from block %d", numEvents, eventData.blockDBWrapper.Block.Height))
			identifierLoggingString := fmt.Sprintf("block %d", eventData.blockDBWrapper.Block.Height)

			indexedDataset, err := dbTypes.IndexBlockEvents(idxr.db, idxr.dryRun, eventData.blockDBWrapper, identifierLoggingString)
			if err != nil {
				config.Log.Fatal(fmt.Sprintf("Error indexing block events for %s.", identifierLoggingString), err)
			}

			err = dbTypes.IndexCustomBlockEvents(*idxr.cfg, idxr.db, idxr.dryRun, indexedDataset, identifierLoggingString, idxr.customBeginBlockParserTrackers, idxr.customEndBlockParserTrackers)

			if err != nil {
				config.Log.Fatal(fmt.Sprintf("Error indexing custom block events for %s.", identifierLoggingString), err)
			}

			config.Log.Info(fmt.Sprintf("Finished indexing %v Block Events from block %d", numEvents, eventData.blockDBWrapper.Block.Height))
		}
	}
}
