package osmosis

import (
	txTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/concentratedliquidity"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/cosmwasmpool"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/gamm"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/poolmanager"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/valsetpref"
)

// MessageTypeHandler is used to unmarshal JSON to a particular type.
var MessageTypeHandler = map[string][]func() txTypes.CosmosMessage{
	gamm.MsgSwapExactAmountIn:                       {func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountIn{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountIn2{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountIn3{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountIn4{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountIn5{} }},
	gamm.MsgSwapExactAmountOut:                      {func() txTypes.CosmosMessage { return &gamm.WrapperMsgSwapExactAmountOut{} }},
	gamm.MsgJoinSwapExternAmountIn:                  {func() txTypes.CosmosMessage { return &gamm.WrapperMsgJoinSwapExternAmountIn{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgJoinSwapExternAmountIn2{} }},
	gamm.MsgJoinSwapShareAmountOut:                  {func() txTypes.CosmosMessage { return &gamm.WrapperMsgJoinSwapShareAmountOut{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgJoinSwapShareAmountOut2{} }},
	gamm.MsgJoinPool:                                {func() txTypes.CosmosMessage { return &gamm.WrapperMsgJoinPool{} }},
	gamm.MsgExitSwapShareAmountIn:                   {func() txTypes.CosmosMessage { return &gamm.WrapperMsgExitSwapShareAmountIn{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgExitSwapShareAmountIn2{} }},
	gamm.MsgExitSwapExternAmountOut:                 {func() txTypes.CosmosMessage { return &gamm.WrapperMsgExitSwapExternAmountOut{} }},
	gamm.MsgExitPool:                                {func() txTypes.CosmosMessage { return &gamm.WrapperMsgExitPool{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgExitPool2{} }},
	gamm.MsgCreatePool:                              {func() txTypes.CosmosMessage { return &gamm.WrapperMsgCreatePool{} }, func() txTypes.CosmosMessage { return &gamm.WrapperMsgCreatePool2{} }},
	gamm.MsgCreateBalancerPool:                      {func() txTypes.CosmosMessage { return &gamm.WrapperMsgCreateBalancerPool{} }},
	gamm.PoolModelsMsgCreateBalancerPool:            {func() txTypes.CosmosMessage { return &gamm.WrapperPoolModelsMsgCreateBalancerPool{} }},
	gamm.PoolModelsMsgCreateStableswapPool:          {func() txTypes.CosmosMessage { return &gamm.WrapperPoolModelsMsgCreateStableswapPool{} }},
	poolmanager.MsgSwapExactAmountIn:                {func() txTypes.CosmosMessage { return &poolmanager.WrapperMsgSwapExactAmountIn{} }},
	poolmanager.MsgSwapExactAmountOut:               {func() txTypes.CosmosMessage { return &poolmanager.WrapperMsgSwapExactAmountOut{} }},
	poolmanager.MsgSplitRouteSwapExactAmountIn:      {func() txTypes.CosmosMessage { return &poolmanager.WrapperMsgSplitRouteSwapExactAmountIn{} }},
	concentratedliquidity.MsgCreatePosition:         {func() txTypes.CosmosMessage { return &concentratedliquidity.WrapperMsgCreatePosition{} }},
	concentratedliquidity.MsgWithdrawPosition:       {func() txTypes.CosmosMessage { return &concentratedliquidity.WrapperMsgWithdrawPosition{} }},
	concentratedliquidity.MsgCollectSpreadRewards:   {func() txTypes.CosmosMessage { return &concentratedliquidity.WrapperMsgCollectSpreadRewards{} }},
	concentratedliquidity.MsgCreateConcentratedPool: {func() txTypes.CosmosMessage { return &concentratedliquidity.WrappeMsgCreateConcentratedPool{} }},
	concentratedliquidity.MsgCollectIncentives:      {func() txTypes.CosmosMessage { return &concentratedliquidity.WrapperMsgCollectIncentives{} }},
	concentratedliquidity.MsgAddToPosition:          {func() txTypes.CosmosMessage { return &concentratedliquidity.WrapperMsgAddToPosition{} }},
	cosmwasmpool.MsgCreateCosmWasmPool:              {func() txTypes.CosmosMessage { return &cosmwasmpool.WrapperMsgCreateCosmWasmPool{} }},
	valsetpref.MsgDelegateToValidatorSet:            {func() txTypes.CosmosMessage { return &valsetpref.WrapperMsgDelegateToValidatorSet{} }},
	valsetpref.MsgUndelegateFromValidatorSet:        {func() txTypes.CosmosMessage { return &valsetpref.WrapperMsgUndelegateFromValidatorSet{} }},
	valsetpref.MsgWithdrawDelegationRewards:         {func() txTypes.CosmosMessage { return &valsetpref.WrapperMsgWithdrawDelegationRewards{} }},
}
