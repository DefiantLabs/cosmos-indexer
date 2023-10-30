package cmd

import (
	"encoding/json"
	"math"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/cosmos-indexer/util"
)

// enqueueBlocksToProcessByMsgType will pass the blocks containing the specified msg type to the indexer
func (idxr *Indexer) enqueueBlocksToProcessByMsgType(blockChan chan int64, chainID uint, msgType string) {
	// get the block range
	startBlock := idxr.cfg.Base.StartBlock
	endBlock := idxr.cfg.Base.EndBlock
	if endBlock == -1 {
		heighestBlock := dbTypes.GetHighestIndexedBlock(idxr.db, chainID)
		endBlock = heighestBlock.Height
	}

	rows, err := idxr.db.Raw(`SELECT height FROM blocks
							JOIN txes ON txes.block_id = blocks.id
							JOIN messages ON messages.tx_id = txes.id
							JOIN message_types ON message_types.id = messages.message_type_id
							AND message_types.message_type = ?
							WHERE height >= ? AND height <= ? AND chain_id = ?::int;
							`, msgType, startBlock, endBlock, chainID).Rows()
	if err != nil {
		config.Log.Fatalf("Error checking DB for blocks to reindex. Err: %v", err)
	}
	defer rows.Close()
	for rows.Next() {
		var block int64
		err = idxr.db.ScanRows(rows, &block)
		if err != nil {
			config.Log.Fatal("Error getting block height. Err: %v", err)
		}
		config.Log.Debugf("Sending block %v to be re-indexed.", block)

		if idxr.cfg.Base.Throttling != 0 {
			time.Sleep(time.Second * time.Duration(idxr.cfg.Base.Throttling))
		}

		// Add the new block to the queue
		blockChan <- block
	}
}

func (idxr *Indexer) enqueueFailedBlocks(blockChan chan int64, chainID uint) {
	// Get all failed blocks
	failedBlocks := dbTypes.GetFailedBlocks(idxr.db, chainID)
	if len(failedBlocks) == 0 {
		return
	}
	for _, block := range failedBlocks {
		if idxr.cfg.Base.Throttling != 0 {
			time.Sleep(time.Second * time.Duration(idxr.cfg.Base.Throttling))
		}
		config.Log.Infof("Will re-attempt failed block: %v", block.Height)
		blockChan <- block.Height
	}
	config.Log.Info("All failed blocks have been re-enqueued for processing")
}

func (idxr *Indexer) enqueueBlocksToProcessFromBlockInputFile(blockChan chan int64, blockInputFile string) {
	plan, err := os.ReadFile(blockInputFile)
	if err != nil {
		config.Log.Fatalf("Error reading block input file. Err: %v", err)
	}
	var blocksToIndex []uint64
	err = json.Unmarshal(plan, &blocksToIndex)

	if err != nil {
		errString := err.Error()

		switch {
		case errString == "json: cannot unmarshal string into Go value of type int":
			config.Log.Fatalf("Error parsing block input file. Err: Found non-integer value in block array")
		case errString == "cannot unmarshal object into Go value of type []uint64":
			config.Log.Fatalf("Error parsing block input file. Err: Found object that could not be parsed into an array of integers")
		case strings.Contains(errString, "cannot unmarshal number"):
			config.Log.Fatalf("Error parsing block input file. Err: Found number that could not be parsed into Go unsigned integer")
		default:
			config.Log.Fatalf("Error parsing block input file. Err: %v", err)
		}
	}

	// sort the data array
	blocksToIndex = util.RemoveDuplicatesFromUint64Slice(blocksToIndex)
	sort.Slice(blocksToIndex, func(i, j int) bool { return blocksToIndex[i] < blocksToIndex[j] })

	// Get latest block height and check to see if we are trying to index blocks outside range
	earliestBlock, latestBlock, err := rpc.GetEarliestAndLatestBlockHeights(idxr.cl)
	if err != nil {
		config.Log.Fatal("Error getting blockchain latest height. Err: %v", err)
	}

	unindexableBlockHeights := []uint64{}
	blockInRange := []uint64{}
	for _, block := range blocksToIndex {
		if block > uint64(latestBlock) || block < uint64(earliestBlock) {
			unindexableBlockHeights = append(unindexableBlockHeights, block)
		} else {
			blockInRange = append(blockInRange, block)
		}
	}

	if len(unindexableBlockHeights) != 0 {
		config.Log.Warnf("The following blocks are past the blockchain earliest height (%d) and latest height (%d) and will be skipped: %v", earliestBlock, latestBlock, unindexableBlockHeights)
	}

	if len(blockInRange) == 0 {
		config.Log.Infof("No blocks to index within blockchain earliest height (%d) and latest height (%d), exiting", earliestBlock, latestBlock)
		return
	}

	// Add jobs to the queue to be processed
	for _, height := range blockInRange {
		if idxr.cfg.Base.Throttling != 0 {
			time.Sleep(time.Second * time.Duration(idxr.cfg.Base.Throttling))
		}
		config.Log.Debugf("Sending block %v to be indexed.", height)
		// Add the new block to the queue
		blockChan <- int64(height)
	}
}

// enqueueBlocksToProcess will pass the blocks that need to be processed to the blockchannel
func (idxr *Indexer) enqueueBlocksToProcess(blockChan chan int64, chainID uint) {
	// Unless explicitly prevented, lets attempt to enqueue any failed blocks
	if idxr.cfg.Base.ReattemptFailedBlocks {
		idxr.enqueueFailedBlocks(blockChan, chainID)
	}

	// Start at the last indexed block height (or the block height in the config, if set)
	currBlock := idxr.GetIndexerStartingHeight(chainID)
	// Don't index past this block no matter what
	lastBlock := idxr.cfg.Base.EndBlock
	var latestBlock int64 = math.MaxInt64

	// Add jobs to the queue to be processed
	for {
		// The program is configured to stop running after a set block height.
		// Generally this will only be done while debugging or if a particular block was incorrectly processed.
		if lastBlock != -1 && currBlock > lastBlock {
			config.Log.Info("Hit the last block we're allowed to index, exiting enqueue func.")
			return
		} else if idxr.cfg.Base.ExitWhenCaughtUp && currBlock > latestBlock {
			config.Log.Info("Hit the last block we're allowed to index, exiting enqueue func.")
			return
		}

		// The job queue is running out of jobs to process, see if the blockchain has produced any new blocks we haven't indexed yet.
		if len(blockChan) <= cap(blockChan)/4 {
			// This is the latest block height available on the Node.
			var err error
			latestBlock, err = rpc.GetLatestBlockHeightWithRetry(idxr.cl, idxr.cfg.Base.RequestRetryAttempts, idxr.cfg.Base.RequestRetryMaxWait)
			if err != nil {
				config.Log.Fatal("Error getting blockchain latest height. Err: %v", err)
			}

			// Throttling in case of hitting public APIs
			if idxr.cfg.Base.Throttling != 0 {
				time.Sleep(time.Second * time.Duration(idxr.cfg.Base.Throttling))
			}

			// Already at the latest block, wait for the next block to be available.
			for currBlock < latestBlock && (currBlock <= lastBlock || lastBlock == -1) && len(blockChan) != cap(blockChan) {
				// if we are not re-indexing, skip curr block if already indexed
				if !idxr.cfg.Base.ReIndex && blockAlreadyIndexed(currBlock, chainID, idxr.db) {
					currBlock++
					continue
				}

				if idxr.cfg.Base.Throttling != 0 {
					time.Sleep(time.Second * time.Duration(idxr.cfg.Base.Throttling))
				}

				// Add the new block to the queue
				blockChan <- currBlock
				currBlock++
			}
		}
	}
}
