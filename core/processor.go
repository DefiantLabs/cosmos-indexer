package core

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/rpc"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
)

type BlockProcessingFailure int

const (
	NodeMissingBlockTxs BlockProcessingFailure = iota
	BlockQueryError
	UnprocessableTxError
	OsmosisNodeRewardLookupError
	OsmosisNodeRewardIndexError
	NodeMissingHistoryForBlock
	FailedBlockEventHandling
)

type FailedBlockHandler func(height int64, code BlockProcessingFailure, err error)

// Process RPC Block data into the model object used by the application.
func ProcessBlock(blockData *ctypes.ResultBlock, blockResultsData *rpc.CustomBlockResults, chainID uint) (models.Block, error) {
	block := models.Block{
		Height:  blockData.Block.Height,
		ChainID: chainID,
	}

	propAddressFromHex, err := sdkTypes.ConsAddressFromHex(blockData.Block.ProposerAddress.String())
	if err != nil {
		return block, err
	}

	block.ProposerConsAddress = models.Address{Address: propAddressFromHex.String()}
	block.TimeStamp = blockData.Block.Time

	return block, nil
}

// Log error to stdout. Not much else we can do to handle right now.
func HandleFailedBlock(height int64, code BlockProcessingFailure, err error) {
	reason := "{unknown error}"
	switch code {
	case NodeMissingBlockTxs:
		reason = "node has no TX history for block"
	case BlockQueryError:
		reason = "failed to query block result for block"
	case OsmosisNodeRewardLookupError:
		reason = "Failed Osmosis rewards lookup for block"
	case OsmosisNodeRewardIndexError:
		reason = "Failed Osmosis rewards indexing for block"
	case NodeMissingHistoryForBlock:
		reason = "Node has no TX history for block"
	case FailedBlockEventHandling:
		reason = "Failed to process block event"
	}

	config.Log.Error(fmt.Sprintf("Block %v failed. Reason: %v", height, reason), err)
}
