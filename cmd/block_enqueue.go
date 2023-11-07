package cmd

import (
	"math"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/core"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
)

func (idxr *Indexer) enqueueFailedBlocks(blockChan chan *core.EnqueueData, chainID uint) {
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
		blockChan <- &core.EnqueueData{
			Height:            block.Height,
			IndexBlockEvents:  idxr.cfg.Base.BlockEventIndexingEnabled,
			IndexTransactions: idxr.cfg.Base.TransactionIndexingEnabled,
		}
	}
	config.Log.Info("All failed blocks have been re-enqueued for processing")
}

// enqueueBlocksToProcess will pass the blocks that need to be processed to the blockchannel
func (idxr *Indexer) enqueueBlocksToProcess(blockChan chan *core.EnqueueData, chainID uint) {
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
				blockChan <- &core.EnqueueData{
					Height:            currBlock,
					IndexBlockEvents:  idxr.cfg.Base.BlockEventIndexingEnabled,
					IndexTransactions: idxr.cfg.Base.TransactionIndexingEnabled,
				}
				currBlock++
			}
		}
	}
}
