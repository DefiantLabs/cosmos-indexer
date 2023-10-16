package poolmanager

import (
	"errors"
	"fmt"
	"strconv"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	poolManagerTypes "github.com/osmosis-labs/osmosis/v19/x/poolmanager/types"
)

const (
	MsgSwapExactAmountIn           = "/osmosis.poolmanager.v1beta1.MsgSwapExactAmountIn"
	MsgSwapExactAmountOut          = "/osmosis.poolmanager.v1beta1.MsgSwapExactAmountOut"
	MsgSplitRouteSwapExactAmountIn = "/osmosis.poolmanager.v1beta1.MsgSplitRouteSwapExactAmountIn"
)

type WrapperMsgSwapExactAmountIn struct {
	txModule.Message
	OsmosisMsgSwapExactAmountIn *poolManagerTypes.MsgSwapExactAmountIn
	Address                     string
	TokenOut                    sdk.Coin
	TokenIn                     sdk.Coin
}

type WrapperMsgSwapExactAmountOut struct {
	txModule.Message
	OsmosisMsgSwapExactAmountOut *poolManagerTypes.MsgSwapExactAmountOut
	Address                      string
	TokenOut                     sdk.Coin
	TokenIn                      sdk.Coin
}

type WrapperMsgSplitRouteSwapExactAmountIn struct {
	txModule.Message
	OsmosisMsgSplitRouteSwapExactAmountIn *poolManagerTypes.MsgSplitRouteSwapExactAmountIn
	Address                               string
	TokenOut                              sdk.Coin
	TokenIn                               sdk.Coin
}

func (sf *WrapperMsgSwapExactAmountIn) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}

	return fmt.Sprintf("MsgSwapExactAmountIn (pool-manager): %s swapped in %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgSwapExactAmountOut) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("MsgSwapExactAmountOut (pool-manager): %s swapped in %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgSplitRouteSwapExactAmountIn) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("MsgSplitRouteSwapExactAmountIn (pool-manager): %s swapped in %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgSwapExactAmountIn) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*poolManagerTypes.MsgSwapExactAmountIn)
	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	sf.TokenIn = sf.OsmosisMsgSwapExactAmountIn.TokenIn
	sf.Address = sf.OsmosisMsgSwapExactAmountIn.Sender

	// The attribute in the log message that shows you the tokens swapped
	tokensSwappedEvt := txModule.GetEventWithType("token_swapped", log)
	transferEvt := txModule.GetEventWithType("transfer", log)

	// We prefer the tokensSwappedEvt if it exists, but it is prone to error
	// If it does exist, attempt a parse. If parsing fails, try other methods.
	// If it does not exist, we will use the transfer event.

	parsed := false

	if tokensSwappedEvt != nil {

		// The last route in the hops gives the token out denom and pool ID for the final output
		lastRoute := sf.OsmosisMsgSwapExactAmountIn.Routes[len(sf.OsmosisMsgSwapExactAmountIn.Routes)-1]
		lastRouteDenom := lastRoute.TokenOutDenom
		lastRoutePoolID := lastRoute.PoolId

		tokenOutStr := txModule.GetLastValueForAttribute("tokens_out", tokensSwappedEvt)
		tokenOutPoolID := txModule.GetLastValueForAttribute("pool_id", tokensSwappedEvt)

		tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
		// Sanity check last route swap
		if err == nil && tokenOut.Denom == lastRouteDenom && strconv.FormatUint(lastRoutePoolID, 10) == tokenOutPoolID {
			sf.TokenOut = tokenOut
			parsed = true
		}
	}

	if !parsed && transferEvt != nil {
		transferEvts, err := txModule.ParseTransferEvent(*transferEvt)
		if err != nil {
			return err
		}

		// The last transfer event should contain the final transfer to the sender
		lastTransferEvt := transferEvts[len(transferEvts)-1]

		if lastTransferEvt.Recipient != sf.Address {
			return errors.New("transfer event recipient does not match message sender")
		}

		tokenOut, err := sdk.ParseCoinNormalized(lastTransferEvt.Amount)
		if err != nil {
			return err
		}

		// The last route in the hops gives the token out denom and pool ID for the final output
		lastRoute := sf.OsmosisMsgSwapExactAmountIn.Routes[len(sf.OsmosisMsgSwapExactAmountIn.Routes)-1]
		lastRouteDenom := lastRoute.TokenOutDenom

		if tokenOut.Denom != lastRouteDenom {
			return errors.New("final transfer denom does not match last route denom")
		}

		sf.TokenOut = tokenOut

		parsed = true
	}

	if !parsed {
		return errors.New("no processable events for poolmanager MsgSwapExactAmountIn")
	}

	return nil
}

// This code is currently untested since I cannot find a TX execution for this
// It should be fine for the time being since it is following the same pattern established for GAMM SwapExactAmountOut, which the poolmanager will call
func (sf *WrapperMsgSwapExactAmountOut) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountOut = msg.(*poolManagerTypes.MsgSwapExactAmountOut)

	// The attribute in the log message that shows you the tokens swapped
	tokensSwappedEvt := txModule.GetEventWithType("token_swapped", log)
	if tokensSwappedEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the first token swapped in (if there are multiple pools we do not care about intermediates)
	tokenInStr, err := txModule.GetValueForAttribute("tokens_in", tokensSwappedEvt)
	if err != nil {
		return err
	}

	tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn = tokenIn

	sf.Address = sf.OsmosisMsgSwapExactAmountOut.Sender
	sf.TokenOut = sf.OsmosisMsgSwapExactAmountOut.TokenOut
	return err
}

// This message behaves like the following:
// 1. Given a token in denom and a set of routes that end in the same denom
// 2. Swap the token in denom for the amount specified for each route
// 3. Get the single denom out that is at the end of each route
// The message errors out if the ending route denoms do not match for every route
// https://github.com/osmosis-labs/osmosis/blob/feaa5ef7d01dc3d082b9d4e7d1dd846d2b54cf6d/x/poolmanager/router.go#L130
func (sf *WrapperMsgSplitRouteSwapExactAmountIn) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSplitRouteSwapExactAmountIn = msg.(*poolManagerTypes.MsgSplitRouteSwapExactAmountIn)

	denomIn := sf.OsmosisMsgSplitRouteSwapExactAmountIn.TokenInDenom

	if len(sf.OsmosisMsgSplitRouteSwapExactAmountIn.Routes) == 0 {
		return nil
	}

	found := false
	denomOut := ""

	totalIn := sdk.NewCoin(denomIn, sdk.ZeroInt())

	// Determine the token out denom from the first route that has pools and its final entry - guaranteed to be the same for every route based on the spec
	// Also determine the amount of token in based on the routes provided in the message
	for _, route := range sf.OsmosisMsgSplitRouteSwapExactAmountIn.Routes {
		if len(route.Pools) > 0 && !found {
			found = true
			denomOut = route.Pools[len(route.Pools)-1].TokenOutDenom
		}

		tokenInCoins := sdk.NewCoin(denomIn, route.TokenInAmount)
		totalIn = totalIn.Add(tokenInCoins)
	}

	if !found || denomOut == "" {
		return nil
	}

	// Get the final out amount from the split_route_swap_exact_amount_in event
	splitRouteFinalEvent := txModule.GetEventWithType("split_route_swap_exact_amount_in", log)

	if splitRouteFinalEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	tokensOutString, err := txModule.GetValueForAttribute("tokens_out", splitRouteFinalEvent)
	if err != nil {
		return err
	}

	tokenOutAmount, ok := sdk.NewIntFromString(tokensOutString)
	if !ok {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	finalTokensOut := sdk.NewCoin(denomOut, tokenOutAmount)

	sf.TokenOut = finalTokensOut
	sf.TokenIn = totalIn
	sf.Address = sf.OsmosisMsgSplitRouteSwapExactAmountIn.Sender

	return nil
}

func (sf *WrapperMsgSwapExactAmountIn) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 1)
	relevantData[0] = parsingTypes.MessageRelevantInformation{
		AmountSent:           sf.TokenIn.Amount.BigInt(),
		DenominationSent:     sf.TokenIn.Denom,
		AmountReceived:       sf.TokenOut.Amount.BigInt(),
		DenominationReceived: sf.TokenOut.Denom,
		SenderAddress:        sf.Address,
		ReceiverAddress:      sf.Address,
	}
	return relevantData
}

func (sf *WrapperMsgSwapExactAmountOut) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 1)
	relevantData[0] = parsingTypes.MessageRelevantInformation{
		AmountSent:           sf.TokenIn.Amount.BigInt(),
		DenominationSent:     sf.TokenIn.Denom,
		AmountReceived:       sf.TokenOut.Amount.BigInt(),
		DenominationReceived: sf.TokenOut.Denom,
		SenderAddress:        sf.Address,
		ReceiverAddress:      sf.Address,
	}
	return relevantData
}

func (sf *WrapperMsgSplitRouteSwapExactAmountIn) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 1)
	relevantData[0] = parsingTypes.MessageRelevantInformation{
		AmountSent:           sf.TokenIn.Amount.BigInt(),
		DenominationSent:     sf.TokenIn.Denom,
		AmountReceived:       sf.TokenOut.Amount.BigInt(),
		DenominationReceived: sf.TokenOut.Denom,
		SenderAddress:        sf.Address,
		ReceiverAddress:      sf.Address,
	}
	return relevantData
}
