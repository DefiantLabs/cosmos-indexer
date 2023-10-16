package epochs

import (
	eventTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	incentivesEventTypes "github.com/DefiantLabs/cosmos-indexer/osmosis/epochs/incentives"
	protorevEventTypes "github.com/DefiantLabs/cosmos-indexer/osmosis/epochs/protorev"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/events"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/epochs"
)

var dayBeginBlockEventTypesToHandlers = map[string][]func() eventTypes.CosmosEvent{
	events.BlockEventDistribution: {func() eventTypes.CosmosEvent { return &incentivesEventTypes.WrapperBlockDistribution{} }},
}

var weekBeginBlockEventTypesToHandlers = map[string][]func() eventTypes.CosmosEvent{
	events.BlockEventCoinReceived: {func() eventTypes.CosmosEvent { return &protorevEventTypes.WrapperBlockCoinReceived{} }},
}

var dayEventTypeHandlers = map[string]map[string][]func() eventTypes.CosmosEvent{
	"begin_block": dayBeginBlockEventTypesToHandlers,
	"end_block":   nil,
}

var weekEventTypeHandlers = map[string]map[string][]func() eventTypes.CosmosEvent{
	"begin_block": weekBeginBlockEventTypesToHandlers,
	"end_block":   nil,
}

// EpochIdentifierBlockEventHandlers is a mapping of epoch identifiers to event types and their associated event handlers
// It is used to get a list of event handlers for:
// 1. A particular Epoch Identifier
// 2. CosmosHub begin blocker or end blocker events for
// 3. A particular Event Type
var EpochIdentifierBlockEventHandlers = map[string]map[string]map[string][]func() eventTypes.CosmosEvent{
	epochs.DayEpochIdentifier:  dayEventTypeHandlers,
	epochs.WeekEpochIdentifier: weekEventTypeHandlers,
}
