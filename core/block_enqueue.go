package core

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/probe/client"
	"gorm.io/gorm"
)

var EnqueueFunctions = map[string]func(chan int64) error{}

// Generates a closure that will enqueue blocks to be processed by the indexer based on the passed in configuration.
// This closure is oriented to a configuration that is not reindexing old blocks. It will start at the last indexed event block and skip already indexed blocks.
func GenerateBlockEventEnqueueFunctionNoReindex(db *gorm.DB, cfg config.IndexConfig, client *client.ChainClient, chainID uint) (func(chan int64) error, error) {
	startHeight := cfg.Base.BlockEventsStartBlock
	endHeight := cfg.Base.BlockEventsEndBlock

	dbLastIndexedBlockEvent, err := dbTypes.GetHighestEventIndexedBlock(db, chainID)
	if err != nil {
		return nil, err
	}
	if dbLastIndexedBlockEvent.Height > 0 {
		startHeight = dbLastIndexedBlockEvent.Height + 1
	}

	// 0 isn't a valid starting block
	if startHeight <= 0 {
		startHeight = 1
	}

	lastKnownBlockHeight, errBh := rpc.GetLatestBlockHeight(client)
	if errBh != nil {
		config.Log.Fatal("Error getting blockchain latest height in block event indexer enqueue builder.", errBh)
	}

	currentHeight := startHeight

	throttling := cfg.Base.Throttling

	// Generate closure that works on the above configured dataset
	return func(blockChan chan int64) error {
		for endHeight == -1 || currentHeight <= endHeight {
			// OPTIMIZE: We should come up with a query to skip blocks in a range that have already been indexed to avoid iterating through
			alreadyIndexed, err := dbTypes.BlockEventsAlreadyIndexed(currentHeight, chainID, db)
			if err != nil {
				return err
			}

			if !alreadyIndexed {
				blockChan <- currentHeight
			} else {
				config.Log.Debugf("Block %d already indexed, skipping", currentHeight)
			}

			currentHeight++
			if currentHeight > lastKnownBlockHeight {
				config.Log.Infof("Block %d has passed lastKnownBlockHeight, checking again", currentHeight)
				// For loop catches both of the following
				// whether we are going too fast and need to do multiple sleeps
				// whether the lastKnownHeight was set a long time ago (as in at app start) and we just need to reset the value
				for {
					lastKnownBlockHeight, err := rpc.GetLatestBlockHeight(client)
					if err != nil {
						return err
					}

					if currentHeight > lastKnownBlockHeight {
						config.Log.Infof("Sleeping...")
						time.Sleep(time.Second * 20)
					} else {
						config.Log.Infof("Continuing until block %d", lastKnownBlockHeight)
						break
					}
				}
			}

			time.Sleep(time.Second * time.Duration(throttling))

		}
		return nil
	}, nil
}
