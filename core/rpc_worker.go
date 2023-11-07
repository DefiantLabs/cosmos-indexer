package core

import (
	"net/http"
	"sync"

	"github.com/DefiantLabs/cosmos-indexer/config"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/probe/client"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
	"gorm.io/gorm"
)

// Wrapper types for gathering full dataset.
type IndexerBlockEventData struct {
	BlockData                *ctypes.ResultBlock
	BlockResultsData         *ctypes.ResultBlockResults
	BlockEventRequestsFailed bool
	GetTxsResponse           *txTypes.GetTxsEventResponse
	TxRequestsFailed         bool
	IndexBlockEvents         bool
	IndexTransactions        bool
}

// This function is responsible for making all RPC requests to the chain needed for later processing.
// The indexer relies on a number of RPC endpoints for full block data, including block event and transaction searches.
func BlockRPCWorker(wg *sync.WaitGroup, blockEnqueueChan chan *EnqueueData, chainID uint, chainStringID string, cfg *config.IndexConfig, chainClient *client.ChainClient, db *gorm.DB, outputChannel chan IndexerBlockEventData) {
	defer wg.Done()
	rpcClient := rpc.URIClient{
		Address: chainClient.Config.RPCAddr,
		Client:  &http.Client{},
	}

	for {
		// Get the next block to process
		block, open := <-blockEnqueueChan
		if !open {
			config.Log.Debugf("Block enqueue channel closed. Exiting RPC worker.")
			break
		}

		currentHeightIndexerData := IndexerBlockEventData{
			BlockEventRequestsFailed: false,
			TxRequestsFailed:         false,
			IndexBlockEvents:         block.IndexBlockEvents,
			IndexTransactions:        block.IndexTransactions,
		}

		// Get the block from the RPC
		blockData, err := rpc.GetBlock(chainClient, block.Height)
		if err != nil {
			// This is the only response we continue on. If we can't get the block, we can't index anything.
			config.Log.Errorf("Error getting block %v from RPC. Err: %v", block, err)
			err := dbTypes.UpsertFailedEventBlock(db, block.Height, chainStringID, cfg.Probe.ChainName)
			if err != nil {
				config.Log.Fatal("Failed to insert failed block event", err)
			}
			err = dbTypes.UpsertFailedBlock(db, block.Height, chainStringID, cfg.Probe.ChainName)
			if err != nil {
				config.Log.Fatal("Failed to insert failed block", err)
			}
			continue
		}

		currentHeightIndexerData.BlockData = blockData

		if block.IndexBlockEvents {
			bresults, err := rpc.GetBlockResultWithRetry(rpcClient, block.Height, cfg.Base.RequestRetryAttempts, cfg.Base.RequestRetryMaxWait)

			if err != nil {
				config.Log.Errorf("Error getting block results for block %v from RPC. Err: %v", block, err)
				err := dbTypes.UpsertFailedEventBlock(db, block.Height, chainStringID, cfg.Probe.ChainName)
				if err != nil {
					config.Log.Fatal("Failed to insert failed block event", err)
				}
				currentHeightIndexerData.BlockResultsData = nil
				currentHeightIndexerData.BlockEventRequestsFailed = true
			} else {
				currentHeightIndexerData.BlockResultsData = bresults
			}
		}

		if block.IndexTransactions {
			txsEventResp, err := rpc.GetTxsByBlockHeight(chainClient, block.Height)

			if err != nil {
				// Attempt to get block results to attempt an in-app codec decode of transactions.
				if currentHeightIndexerData.BlockResultsData == nil {

					bresults, err := rpc.GetBlockResultWithRetry(rpcClient, block.Height, cfg.Base.RequestRetryAttempts, cfg.Base.RequestRetryMaxWait)

					if err != nil {
						config.Log.Errorf("Error getting txs for block %v from RPC. Err: %v", block, err)
						err := dbTypes.UpsertFailedBlock(db, block.Height, chainStringID, cfg.Probe.ChainName)
						if err != nil {
							config.Log.Fatal("Failed to insert failed block", err)
						}
						currentHeightIndexerData.GetTxsResponse = nil
						currentHeightIndexerData.BlockResultsData = nil
						// Only set failed when we can't get the block results either.
						currentHeightIndexerData.TxRequestsFailed = true
					} else {
						currentHeightIndexerData.BlockResultsData = bresults
					}

				}
			} else {
				currentHeightIndexerData.GetTxsResponse = txsEventResp
			}
		}

		outputChannel <- currentHeightIndexerData
	}
}
