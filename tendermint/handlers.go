package tendermint

import (
	eventTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	"github.com/DefiantLabs/cosmos-indexer/tendermint/events"
	liquidityEventTypes "github.com/DefiantLabs/cosmos-indexer/tendermint/events/liquidity"
)

var EndBlockerEventTypeHandlers = map[string][]func() eventTypes.CosmosEvent{
	events.BlockEventDepositToPool:    {func() eventTypes.CosmosEvent { return &liquidityEventTypes.WrapperBlockEventDepositToPool{} }},
	events.BlockEventSwapTransacted:   {func() eventTypes.CosmosEvent { return &liquidityEventTypes.WrapperBlockEventSwapTransacted{} }},
	events.BlockEventWithdrawFromPool: {func() eventTypes.CosmosEvent { return &liquidityEventTypes.WrapperBlockWithdrawFromPool{} }},
}
