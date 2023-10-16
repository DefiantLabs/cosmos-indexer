package rpc

import (
	"time"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/DefiantLabs/cosmos-indexer/config"
	lensClient "github.com/DefiantLabs/lens/client"
	lensQuery "github.com/DefiantLabs/lens/client/query"
	lensEpochsTypes "github.com/DefiantLabs/lens/osmosis/x/epochs/types"
	lensProtorevTypes "github.com/DefiantLabs/lens/osmosis/x/protorev/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	txTypes "github.com/cosmos/cosmos-sdk/types/tx"
)

var apiEndpoints = map[string]string{
	"blocks_endpoint":              "/cosmos/base/tendermint/v1beta1/blocks/%d",
	"latest_block_endpoint":        "/blocks/latest",
	"txs_by_block_height_endpoint": "/cosmos/tx/v1beta1/txs?events=tx.height=%d&pagination.limit=100&order_by=ORDER_BY_UNSPECIFIED",
	"denoms_metadata":              "/cosmos/bank/v1beta1/denoms_metadata",
}

func GetEndpoint(key string) string {
	return apiEndpoints[key]
}

// GetBlockByHeight makes a request to the Cosmos RPC API and returns all the transactions for a specific block
func GetBlockByHeight(cl *lensClient.ChainClient, height int64) (*coretypes.ResultBlockResults, error) {
	options := lensQuery.QueryOptions{Height: height}
	query := lensQuery.Query{Client: cl, Options: &options}
	resp, err := query.BlockResults()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetBlockTimestamp
func GetBlock(cl *lensClient.ChainClient, height int64) (*coretypes.ResultBlock, error) {
	options := lensQuery.QueryOptions{Height: height}
	query := lensQuery.Query{Client: cl, Options: &options}
	resp, err := query.Block()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// GetTxsByBlockHeight makes a request to the Cosmos RPC API and returns all the transactions for a specific block
func GetTxsByBlockHeight(cl *lensClient.ChainClient, height int64) (*txTypes.GetTxsEventResponse, error) {
	pg := query.PageRequest{Limit: 100}
	options := lensQuery.QueryOptions{Height: height, Pagination: &pg}
	query := lensQuery.Query{Client: cl, Options: &options}
	resp, err := query.TxByHeight(cl.Codec)
	if err != nil {
		return nil, err
	}

	// handle pagination if needed
	if resp != nil && resp.Pagination != nil {
		// if there are more total objects than we have so far, keep going
		for resp.Pagination.Total > uint64(len(resp.Txs)) {
			query.Options.Pagination.Offset = uint64(len(resp.Txs))
			chunkResp, err := query.TxByHeight(cl.Codec)
			if err != nil {
				return nil, err
			}
			resp.Txs = append(resp.Txs, chunkResp.Txs...)
			resp.TxResponses = append(resp.TxResponses, chunkResp.TxResponses...)
		}
	}

	return resp, nil
}

// IsCatchingUp true if the node is catching up to the chain, false otherwise
func IsCatchingUp(cl *lensClient.ChainClient) (bool, error) {
	query := lensQuery.Query{Client: cl, Options: &lensQuery.QueryOptions{}}
	ctx, cancel := query.GetQueryContext()
	defer cancel()

	resStatus, err := query.Client.RPCClient.Status(ctx)
	if err != nil {
		return false, err
	}
	return resStatus.SyncInfo.CatchingUp, nil
}

func GetLatestBlockHeight(cl *lensClient.ChainClient) (int64, error) {
	query := lensQuery.Query{Client: cl, Options: &lensQuery.QueryOptions{}}
	ctx, cancel := query.GetQueryContext()
	defer cancel()

	resStatus, err := query.Client.RPCClient.Status(ctx)
	if err != nil {
		return 0, err
	}
	return resStatus.SyncInfo.LatestBlockHeight, nil
}

func GetLatestBlockHeightWithRetry(cl *lensClient.ChainClient, retryMaxAttempts int64, retryMaxWaitSeconds uint64) (int64, error) {
	if retryMaxAttempts == 0 {
		return GetLatestBlockHeight(cl)
	}

	if retryMaxWaitSeconds < 2 {
		retryMaxWaitSeconds = 2
	}

	var attempts int64
	maxRetryTime := time.Duration(retryMaxWaitSeconds) * time.Second
	if maxRetryTime < 0 {
		config.Log.Warn("Detected maxRetryTime overflow, setting time to sane maximum of 30s")
		maxRetryTime = 30 * time.Second
	}

	currentBackoffDuration, maxReached := GetBackoffDurationForAttempts(attempts, maxRetryTime)

	for {
		resp, err := GetLatestBlockHeight(cl)
		attempts++
		if err != nil && (retryMaxAttempts < 0 || (attempts <= retryMaxAttempts)) {
			config.Log.Error("Error getting RPC response, backing off and trying again", err)
			config.Log.Debugf("Attempt %d with wait time %+v", attempts, currentBackoffDuration)
			time.Sleep(currentBackoffDuration)

			// guard against overflow
			if !maxReached {
				currentBackoffDuration, maxReached = GetBackoffDurationForAttempts(attempts, maxRetryTime)
			}

		} else {
			if err != nil {
				config.Log.Error("Error getting RPC response, reached max retry attempts")
			}
			return resp, err
		}
	}
}

func GetEarliestAndLatestBlockHeights(cl *lensClient.ChainClient) (int64, int64, error) {
	query := lensQuery.Query{Client: cl, Options: &lensQuery.QueryOptions{}}
	ctx, cancel := query.GetQueryContext()
	defer cancel()

	resStatus, err := query.Client.RPCClient.Status(ctx)
	if err != nil {
		return 0, 0, err
	}
	return resStatus.SyncInfo.EarliestBlockHeight, resStatus.SyncInfo.LatestBlockHeight, nil
}

// GetEpochsAtHeight makes a request to the Cosmos RPC API and returns the Epoch at a specific height
func GetEpochsAtHeight(cl *lensClient.ChainClient, height int64) (*lensEpochsTypes.QueryEpochsInfoResponse, error) {
	options := lensQuery.QueryOptions{}
	query := lensQuery.Query{Client: cl, Options: &options}
	resp, err := query.EpochsAtHeight(height)
	return resp, err
}

// GetEpochsAtHeight makes a request to the Cosmos RPC API and returns the Epoch at a specific height
func GetProtorevDeveloperAccount(cl *lensClient.ChainClient) (*lensProtorevTypes.QueryGetProtoRevDeveloperAccountResponse, error) {
	options := lensQuery.QueryOptions{}
	query := lensQuery.Query{Client: cl, Options: &options}
	resp, err := query.ProtorevDeveloperAccount()
	return resp, err
}
