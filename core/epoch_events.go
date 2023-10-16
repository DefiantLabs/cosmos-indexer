package core

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/config"
	eventTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	osmosisTypes "github.com/DefiantLabs/cosmos-indexer/osmosis"
	osmosisEpochTypes "github.com/DefiantLabs/cosmos-indexer/osmosis/epochs"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
)

var epochIdentifierEventTypeHandlers = map[string]map[string]map[string][]func() eventTypes.CosmosEvent{}

func ChainSpecificEpochIdentifierEventTypeHandlersBootstrap(chainID string) {
	if chainID == osmosisTypes.ChainID {
		// This is overwriting the entire map, but we only have one epoch module to worry about for now
		epochIdentifierEventTypeHandlers = osmosisEpochTypes.EpochIdentifierBlockEventHandlers
	}
}

func ProcessRPCEpochEvents(blockResults *ctypes.ResultBlockResults, epochIdentifier string) ([]eventTypes.EventRelevantInformation, error) {
	var taxableEvents []eventTypes.EventRelevantInformation

	if handlers, ok := epochIdentifierEventTypeHandlers[epochIdentifier]; ok {
		beginBlockHandlers, beginHandlersExist := handlers["begin_block"]
		endBlockHandlers, endHandlersExist := handlers["end_block"]

		if beginHandlersExist && len(beginBlockHandlers) != 0 {
			for _, event := range blockResults.BeginBlockEvents {
				handlers, handlersFound := beginBlockHandlers[event.Type]
				if !handlersFound {
					continue
				}
				var err error
				for _, handler := range handlers {
					cosmosEventHandler := handler()
					err = cosmosEventHandler.HandleEvent(event.Type, event)
					if err != nil {
						config.Log.Debug(fmt.Sprintf("[Block: %v] Cosmos Block BeginBlocker event of known type: %s. Handler failed", blockResults.Height, event.Type), err)
						continue
					}
					relevantData := cosmosEventHandler.ParseRelevantData()

					taxableEvents = append(taxableEvents, relevantData...)

					config.Log.Debug(fmt.Sprintf("[Block: %v] Cosmos Block BeginBlocker event of known type: %s: %s", blockResults.Height, event.Type, cosmosEventHandler))
					break
				}

				// If err is not nil here, all handlers failed
				if err != nil {
					return nil, fmt.Errorf("could not handle event type %s, all handlers failed", event.Type)
				}
			}
		}

		if endHandlersExist && len(endBlockHandlers) != 0 {
			for _, event := range blockResults.EndBlockEvents {
				handlers, handlersFound := endBlockHandlers[event.Type]
				if !handlersFound {
					continue
				}
				var err error
				for _, handler := range handlers {
					cosmosEventHandler := handler()
					err = cosmosEventHandler.HandleEvent(event.Type, event)
					if err != nil {
						config.Log.Debug(fmt.Sprintf("[Block: %v] Cosmos Block EndBlocker event of known type: %s. Handler failed", blockResults.Height, event.Type), err)
						continue
					}
					relevantData := cosmosEventHandler.ParseRelevantData()

					taxableEvents = append(taxableEvents, relevantData...)

					config.Log.Debug(fmt.Sprintf("[Block: %v] Cosmos Block EndBlocker event of known type: %s: %s", blockResults.Height, event.Type, cosmosEventHandler))
					break
				}

				// If err is not nil here, all handlers failed
				if err != nil {
					return nil, fmt.Errorf("could not handle event type %s, all handlers failed", event.Type)
				}
			}
		}

	}

	return taxableEvents, nil
}
