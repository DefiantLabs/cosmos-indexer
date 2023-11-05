package core

import (
	"net/http"
	"sync"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	"github.com/DefiantLabs/probe/client"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	"gorm.io/gorm"
)

type IndexerBlockEventData struct {
	BlockResultsData *ctypes.ResultBlockResults
	BlockData        *ctypes.ResultBlock
}

func BlockEventRPCWorker(wg *sync.WaitGroup, blockEnqueueChan chan int64, chainID uint, cfg *config.IndexConfig, chainClient *client.ChainClient, db *gorm.DB, transactionIndexingEnabled bool, blockEventIndexingEnabled bool, outputChannel chan IndexerBlockEventData) {
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

		currentHeightIndexerData := IndexerBlockEventData{}

		// Get the block from the RPC
		blockData, err := rpc.GetBlock(chainClient, block)
		if err != nil {
			config.Log.Errorf("Error getting block %v from RPC. Err: %v", block, err)
			continue
		}

		currentHeightIndexerData.BlockData = blockData

		bresults, err := rpc.GetBlockResultWithRetry(rpcClient, block, cfg.Base.RequestRetryAttempts, cfg.Base.RequestRetryMaxWait)

		if err != nil {
			config.Log.Errorf("Error getting block results for block %v from RPC. Err: %v", block, err)
			currentHeightIndexerData.BlockResultsData = nil
		} else {
			currentHeightIndexerData.BlockResultsData = bresults
		}

		outputChannel <- currentHeightIndexerData
	}
}
