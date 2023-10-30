package core

import (
	"encoding/base64"

	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

// TODO: This is a stub, for use when we have begin blocker events in generic manner
// var (
// 	beginBlockerEventTypeHandlers = map[string][]func() eventTypes.CosmosEvent{}
// 	endBlockerEventTypeHandlers   = map[string][]func() eventTypes.CosmosEvent{}
// )

func ChainSpecificEndBlockerEventTypeHandlerBootstrap(chainID string) {
	// Stub, for use when we have begin blocker events
}

func ChainSpecificBeginBlockerEventTypeHandlerBootstrap(chainID string) {
	// Stub, for use when we have begin blocker events
}

func ProcessRPCBlockResults(blockResults *ctypes.ResultBlockResults) (*db.BlockDBWrapper, error) {
	var blockDBWrapper db.BlockDBWrapper

	blockDBWrapper.Block = models.Block{
		Height: blockResults.Height,
	}
	var err error
	blockDBWrapper.BeginBlockEvents, err = ProcessRPCBlockEvents(blockDBWrapper.Block, blockResults.BeginBlockEvents, models.BeginBlockEvent)

	if err != nil {
		return nil, err
	}

	blockDBWrapper.EndBlockEvents, err = ProcessRPCBlockEvents(blockDBWrapper.Block, blockResults.EndBlockEvents, models.EndBlockEvent)

	if err != nil {
		return nil, err
	}

	return &blockDBWrapper, nil
}

func ProcessRPCBlockEvents(block models.Block, blockEvents []abci.Event, blockLifecyclePosition models.BlockLifecyclePosition) ([]db.BlockEventDBWrapper, error) {
	beginBlockEvents := make([]db.BlockEventDBWrapper, len(blockEvents))
	for index, event := range blockEvents {

		beginBlockEvents[index].BlockEvent = models.BlockEvent{
			Index:             uint64(index),
			LifecyclePosition: blockLifecyclePosition,
			Block:             block,
			BlockEventType: models.BlockEventType{
				Type: event.Type,
			},
		}

		beginBlockEvents[index].Attributes = make([]models.BlockEventAttribute, len(event.Attributes))

		for attrIndex, attribute := range event.Attributes {

			// Should we even be decoding these from base64? What are the implications?
			valueBytes, err := base64.StdEncoding.DecodeString(attribute.Value)
			if err != nil {
				return nil, err
			}

			keyBytes, err := base64.StdEncoding.DecodeString(attribute.Key)
			if err != nil {
				return nil, err
			}

			beginBlockEvents[index].Attributes[attrIndex] = models.BlockEventAttribute{
				Value: string(valueBytes),
				BlockEventAttributeKey: models.BlockEventAttributeKey{
					Key: string(keyBytes),
				},
			}
		}

	}

	return beginBlockEvents, nil
}
