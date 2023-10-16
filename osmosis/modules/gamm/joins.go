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
	MsgJoinSwapExternAmountIn = "/osmosis.gamm.v1beta1.MsgJoinSwapExternAmountIn"
	MsgJoinSwapShareAmountOut = "/osmosis.gamm.v1beta1.MsgJoinSwapShareAmountOut"
	MsgJoinPool               = "/osmosis.gamm.v1beta1.MsgJoinPool"
)

type WrapperMsgJoinSwapExternAmountIn struct {
	txModule.Message
	OsmosisMsgJoinSwapExternAmountIn *gammTypes.MsgJoinSwapExternAmountIn
	Address                          string
	TokenOut                         sdk.Coin
	TokenIn                          sdk.Coin
}

type WrapperMsgJoinSwapExternAmountIn2 struct {
	WrapperMsgJoinSwapExternAmountIn
}

type WrapperMsgJoinSwapShareAmountOut struct {
	txModule.Message
	OsmosisMsgJoinSwapShareAmountOut *gammTypes.MsgJoinSwapShareAmountOut
	Address                          string
	TokenOut                         sdk.Coin
	TokenIn                          sdk.Coin
}

// Same as WrapperMsgJoinSwapShareAmountOut but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgJoinSwapShareAmountOut2 struct {
	WrapperMsgJoinSwapShareAmountOut
}

type WrapperMsgJoinPool struct {
	txModule.Message
	OsmosisMsgJoinPool *gammTypes.MsgJoinPool
	Address            string
	TokenOut           sdk.Coin
	TokensIn           []sdk.Coin // joins can be done with multiple tokens in
	Claim              *sdk.Coin  // option claim
}

func (sf *WrapperMsgJoinSwapExternAmountIn) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("MsgJoinSwapExternAmountIn: %s joined with %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgJoinSwapExternAmountIn2) String() string {
	return sf.WrapperMsgJoinSwapExternAmountIn.String()
}

func (sf *WrapperMsgJoinSwapShareAmountOut) String() string {
	var tokenSwappedOut string
	var tokenSwappedIn string
	if !sf.TokenOut.IsNil() {
		tokenSwappedOut = sf.TokenOut.String()
	}
	if !sf.TokenIn.IsNil() {
		tokenSwappedIn = sf.TokenIn.String()
	}
	return fmt.Sprintf("MsgJoinSwapShareAmountOut: %s joined with %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgJoinPool) String() string {
	var tokenOut string
	var tokensIn []string
	if !(len(sf.TokensIn) == 0) {
		for _, v := range sf.TokensIn {
			tokensIn = append(tokensIn, v.String())
		}
	}
	if !sf.TokenOut.IsNil() {
		tokenOut = sf.TokenOut.String()
	}
	return fmt.Sprintf("MsgJoinPool: %s joined pool with %s and received %s",
		sf.Address, strings.Join(tokensIn, ", "), tokenOut)
}

func (sf *WrapperMsgJoinSwapExternAmountIn) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgJoinSwapExternAmountIn = msg.(*gammTypes.MsgJoinSwapExternAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the received GAMM tokens from the pool
	coinbaseEvt := txModule.GetEventWithType("coinbase", log)
	if coinbaseEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens received
	gammTokenInStr, err := txModule.GetValueForAttribute(EventAttributeAmount, coinbaseEvt)
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
	sf.TokenOut = gammTokenIn

	// we can pull the token in directly from the Osmosis Message
	sf.TokenIn = sf.OsmosisMsgJoinSwapExternAmountIn.TokenIn

	// Address of whoever initiated the join
	poolJoinedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolJoined, log)
	if poolJoinedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the join.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolJoinedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	return err
}

func (sf *WrapperMsgJoinSwapExternAmountIn2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgJoinSwapExternAmountIn = msg.(*gammTypes.MsgJoinSwapExternAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// we can pull the token and sender in directly from the Osmosis Message
	sf.TokenIn = sf.OsmosisMsgJoinSwapExternAmountIn.TokenIn
	sf.Address = sf.OsmosisMsgJoinSwapExternAmountIn.Sender

	transferEvt := txModule.GetEventWithType("transfer", log)
	gammOutString := ""
	// Loop backwards to find the GAMM out string
	for i := len(transferEvt.Attributes) - 1; i >= 0; i-- {
		attr := transferEvt.Attributes[i]
		if attr.Key == EventAttributeAmount && strings.Contains(attr.Value, "gamm") && strings.HasSuffix(attr.Value, fmt.Sprintf("/%d", sf.OsmosisMsgJoinSwapExternAmountIn.PoolId)) {
			// Verify the recipient of the gamm output is the sender of the message
			if i-2 > -1 && transferEvt.Attributes[i-2].Key == "recipient" && transferEvt.Attributes[i-2].Value == sf.Address {
				gammOutString = attr.Value
			}
		}
	}

	if gammOutString == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	gammTokenOut, err := sdk.ParseCoinNormalized(gammOutString)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut = gammTokenOut

	return err
}

func (sf *WrapperMsgJoinSwapShareAmountOut) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgJoinSwapShareAmountOut = msg.(*gammTypes.MsgJoinSwapShareAmountOut)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the received GAMM tokens from the pool
	coinbaseEvt := txModule.GetEventWithType("coinbase", log)
	if coinbaseEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens received
	gammTokenInStr, err := txModule.GetValueForAttribute(EventAttributeAmount, coinbaseEvt)
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
	sf.TokenOut = gammTokenIn

	// Address of whoever initiated the join
	poolJoinedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolJoined, log)
	if poolJoinedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the join.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolJoinedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	tokenIn, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, poolJoinedEvent)
	if err != nil {
		return err
	}

	if tokenIn == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn, err = sdk.ParseCoinNormalized(tokenIn)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgJoinSwapShareAmountOut2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgJoinSwapShareAmountOut = msg.(*gammTypes.MsgJoinSwapShareAmountOut)
	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the received GAMM tokens from the pool
	transferEvt := txModule.GetEventWithType("transfer", log)
	if transferEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	gammOutString := ""
	// Loop backwards to find the GAMM out string
	for i := len(transferEvt.Attributes) - 1; i >= 0; i-- {
		attr := transferEvt.Attributes[i]
		if attr.Key == EventAttributeAmount && strings.Contains(attr.Value, "gamm") && strings.HasSuffix(attr.Value, fmt.Sprintf("/%d", sf.OsmosisMsgJoinSwapShareAmountOut.PoolId)) {
			// Verify the recipient of the gamm output is the sender of the message
			if i-2 > -1 && transferEvt.Attributes[i-2].Key == "recipient" && transferEvt.Attributes[i-2].Value == sf.OsmosisMsgJoinSwapShareAmountOut.Sender {
				gammOutString = attr.Value
			}
		}
	}

	if gammOutString == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens received
	gammTokenOut, err := sdk.ParseCoinNormalized(gammOutString)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut = gammTokenOut

	// Address of whoever initiated the join
	poolJoinedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolJoined, log)
	if poolJoinedEvent == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the join.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolJoinedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	tokenIn, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, poolJoinedEvent)
	if err != nil {
		return err
	}

	if tokenIn == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn, err = sdk.ParseCoinNormalized(tokenIn)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgJoinPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgJoinPool = msg.(*gammTypes.MsgJoinPool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the received GAMM tokens from the pool
	transferEvt := txModule.GetEventWithType(bankTypes.EventTypeTransfer, log)
	if transferEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// This gets the amount of GAMM tokens received and claim (if needed)
	var gammTokenOutStr string
	if strings.Contains(fmt.Sprint(log), "claim") {
		// This gets the amount of the claim
		claimStr := txModule.GetLastValueForAttribute(EventAttributeAmount, transferEvt)
		claimTokenOut, err := sdk.ParseCoinNormalized(claimStr)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		sf.Claim = &claimTokenOut

		gammTokenOutStr = txModule.GetNthValueForAttribute(EventAttributeAmount, 2, transferEvt)
	} else {
		gammTokenOutStr = txModule.GetLastValueForAttribute(EventAttributeAmount, transferEvt)
	}
	if !strings.Contains(gammTokenOutStr, "gamm") {
		fmt.Println(gammTokenOutStr)
		fmt.Println("Gamm token out string must contain gamm")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	gammTokenOut, err := sdk.ParseCoinNormalized(gammTokenOutStr)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut = gammTokenOut

	// Address of whoever initiated the join
	poolJoinedEvent := txModule.GetEventWithType(gammTypes.TypeEvtPoolJoined, log)
	if poolJoinedEvent == nil {
		// If the message doesn't have the pool_joined event, we can parse the transaction event to extract the
		// amounts of the tokens they transferred in.

		// find the multi-coin amount that also is not gamms... those must be the coins transferred in
		var tokensIn string
		var sender string
		for i, attr := range transferEvt.Attributes {
			if attr.Key == EventAttributeAmount && strings.Contains(attr.Value, ",") {
				tokensIn = attr.Value
				// If we haven't found the sender yet, it will be the address that sent this non-gamm token
				if i > 0 && transferEvt.Attributes[i-1].Key == "sender" && sf.OsmosisMsgJoinPool.Sender == transferEvt.Attributes[i-1].Value {
					sender = transferEvt.Attributes[i-1].Value
				}
				break
			}
		}
		// if either of these methods failed to get info, give up and return an error
		if len(tokensIn) == 0 || sender == "" {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		// if we got what we needed, set the correct values and return successfully
		sf.Address = sender
		sf.TokensIn, err = sdk.ParseCoinsNormalized(tokensIn)
		if err != nil {
			fmt.Println("Error parsing coins")
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		return nil
	}

	// Address of whoever initiated the join.
	senderAddress, err := txModule.GetValueForAttribute("sender", poolJoinedEvent)
	if err != nil {
		return err
	}

	if senderAddress == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderAddress

	// String value for the tokens in, which can be multiple
	tokensInString, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, poolJoinedEvent)
	if err != nil {
		return err
	}

	if tokensInString == "" {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokensIn, err = sdk.ParseCoinsNormalized(tokensInString)
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return err
}

func (sf *WrapperMsgJoinSwapExternAmountIn) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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

func (sf *WrapperMsgJoinSwapShareAmountOut) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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

func (sf *WrapperMsgJoinPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// need to make a relevant data block for all Tokens sent to the pool since JoinPool can use 1 or both tokens used in the pool
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.TokensIn))

	// figure out how many gams per token
	nthGamms, remainderGamms := calcNthGams(sf.TokenOut.Amount.BigInt(), len(sf.TokensIn))
	for i, v := range sf.TokensIn {
		// split received tokens across entry so we receive GAMM tokens for both exchanges
		// each swap will get 1 nth of the gams until the last one which will get the remainder
		if i != len(sf.TokensIn)-1 {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       nthGamms,
				DenominationReceived: sf.TokenOut.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       remainderGamms,
				DenominationReceived: sf.TokenOut.Denom,
				SenderAddress:        sf.Address,
				ReceiverAddress:      sf.Address,
			}
		}
	}

	// handle claim if there is one
	if sf.Claim != nil {
		relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
			ReceiverAddress:      sf.Address,
			AmountReceived:       sf.Claim.Amount.BigInt(),
			DenominationReceived: sf.Claim.Denom,
		})
	}

	return relevantData
}
