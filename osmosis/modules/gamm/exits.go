package gamm

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gammTypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"
)

const (
	MsgExitSwapShareAmountIn   = "/osmosis.gamm.v1beta1.MsgExitSwapShareAmountIn"
	MsgExitSwapExternAmountOut = "/osmosis.gamm.v1beta1.MsgExitSwapExternAmountOut"
	MsgExitPool                = "/osmosis.gamm.v1beta1.MsgExitPool"
)

type WrapperMsgExitPool2 struct {
	WrapperMsgExitPool
}

type WrapperMsgExitSwapShareAmountIn struct {
	txModule.Message
	OsmosisMsgExitSwapShareAmountIn *gammTypes.MsgExitSwapShareAmountIn
	Address                         string
	TokenOut                        sdk.Coin
	TokenIn                         sdk.Coin
}

// Same as WrapperMsgExitSwapShareAmountIn but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgExitSwapShareAmountIn2 struct {
	txModule.Message
	OsmosisMsgExitSwapShareAmountIn *gammTypes.MsgExitSwapShareAmountIn
	Address                         string
	TokensOut                       sdk.Coins
	TokenSwaps                      []tokenSwap
	TokenIn                         sdk.Coin
}

type WrapperMsgExitSwapExternAmountOut struct {
	txModule.Message
	OsmosisMsgExitSwapExternAmountOut *gammTypes.MsgExitSwapExternAmountOut
	Address                           string
	TokenOut                          sdk.Coin
	TokenIn                           sdk.Coin
}

type WrapperMsgExitPool struct {
	txModule.Message
	OsmosisMsgExitPool *gammTypes.MsgExitPool
	Address            string
	TokensOutOfPool    []sdk.Coin // exits can received multiple tokens out
	TokenIntoPool      sdk.Coin
}

func (sf *WrapperMsgExitSwapShareAmountIn) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("MsgMsgExitSwapShareAmountIn: %s exited with %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgExitSwapShareAmountIn2) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string

	var postExitTokenSwaps []string
	var postExitTokenSwapsRepr string

	if !sf.TokensOut.Empty() {
		tokenSwappedOut = sf.TokensOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}

	if !(len(sf.TokenSwaps) == 0) {
		for _, swap := range sf.TokenSwaps {
			postExitTokenSwaps = append(postExitTokenSwaps, fmt.Sprintf("%s for %s", swap.TokenSwappedIn.String(), swap.TokenSwappedOut.String()))
		}

		postExitTokenSwapsRepr = strings.Join(postExitTokenSwaps, ", ")
	}

	return fmt.Sprintf("MsgMsgExitSwapShareAmountIn: %s exited with %s and received %s, then swapped %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut, postExitTokenSwapsRepr)
}

func (sf *WrapperMsgExitSwapExternAmountOut) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("WrapperMsgExitSwapExternAmountOut: %s exited with %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgExitPool) String() string {
	var tokensOut []string
	var tokenIn string
	if !(len(sf.TokensOutOfPool) == 0) {
		for _, v := range sf.TokensOutOfPool {
			tokensOut = append(tokensOut, v.String())
		}
	}
	if !sf.TokenIntoPool.IsNil() {
		tokenIn = sf.TokenIntoPool.String()
	}
	return fmt.Sprintf("MsgExitPool: %s exited pool with %s and received %s",
		sf.Address, tokenIn, strings.Join(tokensOut, ", "))
}

func (sf *WrapperMsgExitPool2) String() string {
	return sf.WrapperMsgExitPool.String()
}

func (sf *WrapperMsgExitSwapShareAmountIn) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgExitSwapShareAmountIn = msg.(*gammTypes.MsgExitSwapShareAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the burned GAMM tokens sent to the pool
	burnEvt := txModule.GetEventWithType("burn", log)
	if burnEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM exited with
	gammTokenInStr, err := txModule.GetValueForAttribute(EventAttributeAmount, burnEvt)
	if err != nil {
		return err
	}

	if !strings.Contains(gammTokenInStr, "gamm") {
		fmt.Println("Gamm token in string must contain gamm")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	gammTokenIn, err := sdk.ParseCoinNormalized(gammTokenInStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn = gammTokenIn

	// Address of whoever initiated the exit
	poolExitedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolExited, log)
	if poolExitedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the exit.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolExitedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	tokenOut, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensOut, poolExitedEvent)
	if err != nil {
		return err
	}

	if tokenOut == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut, err = sdk.ParseCoinNormalized(tokenOut)

	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgExitSwapShareAmountIn2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgExitSwapShareAmountIn = msg.(*gammTypes.MsgExitSwapShareAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the burned GAMM tokens sent to the pool
	burnEvt := txModule.GetEventWithType("burn", log)
	if burnEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM exited with
	gammTokenInStr, err := txModule.GetValueForAttribute(EventAttributeAmount, burnEvt)
	if err != nil {
		return err
	}

	if !strings.Contains(gammTokenInStr, "gamm") {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	gammTokenIn, err := sdk.ParseCoinNormalized(gammTokenInStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn = gammTokenIn

	// Address of whoever initiated the exit
	poolExitedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolExited, log)
	if poolExitedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the exit.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolExitedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	tokensOut, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensOut, poolExitedEvent)
	if err != nil {
		return err
	}

	if tokensOut == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	multiTokensOut, err := sdk.ParseCoinsNormalized(tokensOut)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	sf.TokensOut = multiTokensOut

	// The token swapped events contain the final amount of tokens out in this tx
	tokenSwappedEvents := txModule.GetAllEventsWithType(gammTypes.TypeEvtTokenSwapped, log)

	// This is to handle multi-token pool exit swaps
	for i := range tokenSwappedEvents {
		tokenSwappedIn, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, &tokenSwappedEvents[i])
		if err != nil {
			return err
		}

		tokenSwappedOut, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensOut, &tokenSwappedEvents[i])
		if err != nil {
			return err
		}

		parsedTokensSwappedIn, err := sdk.ParseCoinNormalized(tokenSwappedIn)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		parsedTokensSwappedOut, err := sdk.ParseCoinNormalized(tokenSwappedOut)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		tokenSwap := tokenSwap{TokenSwappedIn: parsedTokensSwappedIn, TokenSwappedOut: parsedTokensSwappedOut}

		sf.TokenSwaps = append(sf.TokenSwaps, tokenSwap)
	}

	return err
}

func (sf *WrapperMsgExitSwapExternAmountOut) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgExitSwapExternAmountOut = msg.(*gammTypes.MsgExitSwapExternAmountOut)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the burned GAMM tokens sent to the pool
	burnEvt := txModule.GetEventWithType("burn", log)
	if burnEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM exited with
	gammTokenInStr, err := txModule.GetValueForAttribute(EventAttributeAmount, burnEvt)
	if err != nil {
		return err
	}

	if !strings.Contains(gammTokenInStr, "gamm") {
		fmt.Println("Gamm token in string must contain gamm")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	gammTokenIn, err := sdk.ParseCoinNormalized(gammTokenInStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn = gammTokenIn

	// Address of whoever initiated the exit
	poolExitedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolExited, log)
	if poolExitedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the exit.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolExitedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	tokenOut, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensOut, poolExitedEvent)
	if err != nil {
		return err
	}

	if tokenOut == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut, err = sdk.ParseCoinNormalized(tokenOut)

	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgExitPool2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgExitPool = msg.(*gammTypes.MsgExitPool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the sent GAMM tokens during the exit
	transferEvt := txModule.GetEventWithType(bankTypes.EventTypeTransfer, log)
	if transferEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens sent
	gammTokenOutStr := txModule.GetLastValueForAttribute(EventAttributeAmount, transferEvt)
	if !strings.Contains(gammTokenOutStr, "gamm") {
		fmt.Println("Gamm token out string must contain gamm")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	gammTokenOut, err := sdk.ParseCoinNormalized(gammTokenOutStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIntoPool = gammTokenOut

	if sf.OsmosisMsgExitPool.Sender == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = sf.OsmosisMsgExitPool.Sender

	// The first attribute in the event should have a key 'recipient', and a value with the Msg sender's address (whoever is exiting the pool)
	senderAddr := txModule.GetNthValueForAttribute("recipient", 1, transferEvt)
	if senderAddr != sf.Address {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v, senderAddr != sf.Address", log)}
	}

	// String value for the tokens out, which can be multiple
	tokensOutString := txModule.GetNthValueForAttribute(EventAttributeAmount, 1, transferEvt)
	if tokensOutString == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	sf.TokensOutOfPool, err = sdk.ParseCoinsNormalized(tokensOutString)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgExitPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgExitPool = msg.(*gammTypes.MsgExitPool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the sent GAMM tokens during the exit
	transverEvt := txModule.GetEventWithType(bankTypes.EventTypeTransfer, log)
	if transverEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens sent
	gammTokenOutStr := txModule.GetLastValueForAttribute(EventAttributeAmount, transverEvt)
	if !strings.Contains(gammTokenOutStr, "gamm") {
		fmt.Println("Gamm token out string must contain gamm")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	gammTokenOut, err := sdk.ParseCoinNormalized(gammTokenOutStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIntoPool = gammTokenOut

	// Address of whoever initiated the exit
	poolExitedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolExited, log)
	if poolExitedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the exit.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolExitedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	// String value for the tokens in, which can be multiple
	tokensOutString, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensOut, poolExitedEvent)
	if err != nil {
		return err
	}

	if tokensOutString == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	sf.TokensOutOfPool, err = sdk.ParseCoinsNormalized(tokensOutString)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgExitSwapShareAmountIn) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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

func (sf *WrapperMsgExitSwapShareAmountIn2) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.TokensOut))

	// figure out how many gams per token
	nthGamms, remainderGamms := calcNthGams(sf.TokenIn.Amount.BigInt(), len(sf.TokensOut))

	// Handle the pool exit
	for i, v := range sf.TokensOut {
		// split received tokens across entry so we receive GAMM tokens for both exchanges
		// each swap will get 1 nth of the gams until the last one which will get the remainder
		if i != len(sf.TokensOut)-1 {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           nthGamms,
				DenominationSent:     sf.TokenIn.Denom,
				AmountReceived:       v.Amount.BigInt(),
				DenominationReceived: v.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           remainderGamms,
				DenominationSent:     sf.TokenIn.Denom,
				AmountReceived:       v.Amount.BigInt(),
				DenominationReceived: v.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		}
	}

	// Handle the post exit swap event
	for _, tokensSwapped := range sf.TokenSwaps {
		relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
			AmountSent:           tokensSwapped.TokenSwappedIn.Amount.BigInt(),
			DenominationSent:     tokensSwapped.TokenSwappedIn.Denom,
			AmountReceived:       tokensSwapped.TokenSwappedOut.Amount.BigInt(),
			DenominationReceived: tokensSwapped.TokenSwappedOut.Denom,
			SenderAddress:        sf.Address,
			ReceiverAddress:      sf.Address,
		})
	}

	return relevantData
}

func (sf *WrapperMsgExitSwapExternAmountOut) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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

func (sf *WrapperMsgExitPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// need to make a relevant data block for all Tokens received from the pool since ExitPool can receive 1 or both tokens used in the pool
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.TokensOutOfPool))

	// figure out how many gams per token
	nthGamms, remainderGamms := calcNthGams(sf.TokenIntoPool.Amount.BigInt(), len(sf.TokensOutOfPool))
	for i, v := range sf.TokensOutOfPool {
		// only add received tokens to the first entry so we dont duplicate received GAMM tokens
		if i != len(sf.TokensOutOfPool)-1 {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           nthGamms,
				DenominationSent:     sf.TokenIntoPool.Denom,
				AmountReceived:       v.Amount.BigInt(),
				DenominationReceived: v.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           remainderGamms,
				DenominationSent:     sf.TokenIntoPool.Denom,
				AmountReceived:       v.Amount.BigInt(),
				DenominationReceived: v.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		}
	}

	return relevantData
}

func (sf *WrapperMsgExitPool2) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return sf.WrapperMsgExitPool.ParseRelevantData()
}
