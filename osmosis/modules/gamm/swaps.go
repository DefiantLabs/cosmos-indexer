package gamm

import (
	"errors"
	"fmt"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gammTypes "github.com/osmosis-labs/osmosis/v19/x/gamm/types"
)

const (
	MsgSwapExactAmountIn  = "/osmosis.gamm.v1beta1.MsgSwapExactAmountIn"
	MsgSwapExactAmountOut = "/osmosis.gamm.v1beta1.MsgSwapExactAmountOut"
)

type WrapperMsgSwapExactAmountIn struct {
	txModule.Message
	OsmosisMsgSwapExactAmountIn *gammTypes.MsgSwapExactAmountIn
	Address                     string
	TokenOut                    sdk.Coin
	TokenIn                     sdk.Coin
}

// Same as WrapperMsgSwapExactAmountIn but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgSwapExactAmountIn2 struct {
	WrapperMsgSwapExactAmountIn
}

// Same as WrapperMsgSwapExactAmountIn but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgSwapExactAmountIn3 struct {
	WrapperMsgSwapExactAmountIn
}

// Same as WrapperMsgSwapExactAmountIn but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgSwapExactAmountIn4 struct {
	WrapperMsgSwapExactAmountIn
}

// Same as WrapperMsgSwapExactAmountIn but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).
type WrapperMsgSwapExactAmountIn5 struct {
	WrapperMsgSwapExactAmountIn
}

type WrapperMsgSwapExactAmountOut struct {
	txModule.Message
	OsmosisMsgSwapExactAmountOut *gammTypes.MsgSwapExactAmountOut
	Address                      string
	TokenOut                     sdk.Coin
	TokenIn                      sdk.Coin
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

	return fmt.Sprintf("MsgSwapExactAmountIn: %s swapped in %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgSwapExactAmountIn2) String() string {
	return sf.WrapperMsgSwapExactAmountIn.String()
}

func (sf *WrapperMsgSwapExactAmountIn3) String() string {
	return sf.WrapperMsgSwapExactAmountIn.String()
}

func (sf *WrapperMsgSwapExactAmountIn4) String() string {
	return sf.WrapperMsgSwapExactAmountIn.String()
}

func (sf *WrapperMsgSwapExactAmountIn5) String() string {
	return sf.WrapperMsgSwapExactAmountIn.String()
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
	return fmt.Sprintf("MsgSwapExactAmountOut: %s swapped in %s and received %s",
		sf.Address, tokenSwappedIn, tokenSwappedOut)
}

func (sf *WrapperMsgSwapExactAmountIn) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*gammTypes.MsgSwapExactAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens swapped
	tokensSwappedEvt := txModule.GetEventWithType(gammTypes.TypeEvtTokenSwapped, log)
	if tokensSwappedEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the swap. Will be both sender/receiver.
	senderReceiver, err := txModule.GetValueForAttribute("sender", tokensSwappedEvt)
	if err != nil {
		return err
	}

	if senderReceiver == "" {
		fmt.Println("Error getting sender.")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderReceiver

	// This gets the first token swapped in (if there are multiple pools we do not care about intermediates)
	tokenInStr, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, tokensSwappedEvt)
	if err != nil {
		return err
	}

	tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
	if err != nil {
		fmt.Println("Error parsing coins in. Err: ", err)
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenIn = tokenIn

	// This gets the last token swapped out (if there are multiple pools we do not care about intermediates)
	tokenOutStr := txModule.GetLastValueForAttribute(gammTypes.AttributeKeyTokensOut, tokensSwappedEvt)
	tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
	if err != nil {
		fmt.Println("Error parsing coins out. Err: ", err)
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut = tokenOut

	return err
}

// Handles an OLDER (now defunct) swap on Osmosis mainnet (osmosis-1).
// Example TX hash: EA5C6AB8E3084D933F3E005A952A362DFD13DC79003DC2BC9E247920FCDFDD34
func (sf *WrapperMsgSwapExactAmountIn2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*gammTypes.MsgSwapExactAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens swapped
	tokensSwappedEvt := txModule.GetEventWithType(EventTypeClaim, log)
	if tokensSwappedEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// Address of whoever initiated the swap. Will be both sender/receiver.
	senderReceiver, err := txModule.GetValueForAttribute("sender", tokensSwappedEvt)
	if err != nil {
		return err
	}

	if senderReceiver == "" {
		fmt.Println("Error getting sender.")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.Address = senderReceiver

	// First token swapped in (if there are multiple pools we do not care about intermediates)
	sf.TokenIn = sf.OsmosisMsgSwapExactAmountIn.TokenIn

	// This gets the last token swapped out (if there are multiple pools we do not care about intermediates)
	tokenOutStr := txModule.GetLastValueForAttribute(EventAttributeAmount, tokensSwappedEvt)
	tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
	if err != nil {
		fmt.Println("Error parsing coins out. Err: ", err)
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}
	sf.TokenOut = tokenOut

	return err
}

// Handles an OLDER (now defunct) swap on Osmosis mainnet (osmosis-1).
// Example TX hash: BC8384F767F48EDDF65646EC136518DE00B59A8E2793AABFE7563C62B39A59AE
func (sf *WrapperMsgSwapExactAmountIn3) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*gammTypes.MsgSwapExactAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens transferred
	tokensTransferredEvt := txModule.GetEventWithType(EventTypeTransfer, log)
	if tokensTransferredEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	msgSender := sf.OsmosisMsgSwapExactAmountIn.Sender
	msgTokensIn := sf.OsmosisMsgSwapExactAmountIn.TokenIn

	// First sender should be the address that conducted the swap
	firstSender := txModule.GetNthValueForAttribute("sender", 1, tokensTransferredEvt)
	firstAmount := txModule.GetNthValueForAttribute(EventAttributeAmount, 1, tokensTransferredEvt)

	if firstSender != msgSender {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	} else if firstAmount != msgTokensIn.String() {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	sf.Address = msgSender

	secondReceiver := txModule.GetNthValueForAttribute("recipient", 2, tokensTransferredEvt)
	secondAmount := txModule.GetNthValueForAttribute(EventAttributeAmount, 2, tokensTransferredEvt)

	if secondReceiver != msgSender {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	amountReceived, err := sdk.ParseCoinNormalized(secondAmount)
	if err != nil {
		return err
	}

	outDenom := sf.OsmosisMsgSwapExactAmountIn.Routes[len(sf.OsmosisMsgSwapExactAmountIn.Routes)-1].TokenOutDenom
	if amountReceived.Denom != outDenom {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("amountReceived.Denom != outDenom. Log: %+v", log)}
	}

	// Address of whoever initiated the swap. Will be both sender/receiver.
	senderReceiver, err := txModule.GetValueForAttribute("sender", tokensTransferredEvt)
	if err != nil {
		return err
	}

	if senderReceiver == "" {
		fmt.Println("Error getting sender.")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// First token swapped in (if there are multiple pools we do not care about intermediates)
	sf.TokenIn = sf.OsmosisMsgSwapExactAmountIn.TokenIn
	sf.TokenOut = amountReceived

	return err
}

// Handles an OLDER (now defunct) swap on Osmosis mainnet (osmosis-1).
// Example TX hash: BB954377AB50F8EF204123DC8B101B7CB597153C0B8372166BC28ABDAA262516
func (sf *WrapperMsgSwapExactAmountIn4) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*gammTypes.MsgSwapExactAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens transferred
	tokensTransferredEvt := txModule.GetEventWithType(EventTypeTransfer, log)
	if tokensTransferredEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	msgSender := sf.OsmosisMsgSwapExactAmountIn.Sender
	msgTokensIn := sf.OsmosisMsgSwapExactAmountIn.TokenIn

	// First sender should be the address that conducted the swap
	firstSender := txModule.GetNthValueForAttribute("sender", 1, tokensTransferredEvt)
	firstAmount := txModule.GetNthValueForAttribute(EventAttributeAmount, 1, tokensTransferredEvt)

	if firstSender != msgSender {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	} else if firstAmount != msgTokensIn.String() {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	sf.Address = msgSender

	lastReceiver := txModule.GetLastValueForAttribute("recipient", tokensTransferredEvt)
	lastAmount := txModule.GetLastValueForAttribute(EventAttributeAmount, tokensTransferredEvt)

	if lastReceiver != msgSender {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	amountReceived, err := sdk.ParseCoinNormalized(lastAmount)
	if err != nil {
		return err
	}

	outDenom := sf.OsmosisMsgSwapExactAmountIn.Routes[len(sf.OsmosisMsgSwapExactAmountIn.Routes)-1].TokenOutDenom
	if amountReceived.Denom != outDenom {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("amountReceived.Denom != outDenom. Log: %+v", log)}
	}

	// Address of whoever initiated the swap. Will be both sender/receiver.
	senderReceiver, err := txModule.GetValueForAttribute("sender", tokensTransferredEvt)
	if err != nil {
		return err
	}

	if senderReceiver == "" {
		fmt.Println("Error getting sender.")
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	// First token swapped in (if there are multiple pools we do not care about intermediates)
	sf.TokenIn = sf.OsmosisMsgSwapExactAmountIn.TokenIn
	sf.TokenOut = amountReceived

	return err
}

func (sf *WrapperMsgSwapExactAmountIn5) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountIn = msg.(*gammTypes.MsgSwapExactAmountIn)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens transferred
	tokensTransferredEvt := txModule.GetEventWithType(EventTypeTransfer, log)
	if tokensTransferredEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	transferEvents, err := txModule.ParseTransferEvent(*tokensTransferredEvt)
	if err != nil {
		return fmt.Errorf("error parsing transfer event: %w", err)
	}

	// pull out transfer events that do not have empty amounts
	var properTransferEvents []txModule.TransferEvent
	for _, transferEvent := range transferEvents {
		if transferEvent.Amount != "" {
			properTransferEvents = append(properTransferEvents, transferEvent)
		}
	}

	firstTransfer := properTransferEvents[0]

	// Sanity check transfer events
	if firstTransfer.Amount != sf.OsmosisMsgSwapExactAmountIn.TokenIn.String() {
		return errors.New("first transfer amount does not match token in")
	} else if firstTransfer.Sender != sf.OsmosisMsgSwapExactAmountIn.Sender {
		return errors.New("first transfer sender does not match sender")
	}

	// last transfer contains the final amount out

	lastTransfer := properTransferEvents[len(properTransferEvents)-1]

	lastAmountReceived, err := sdk.ParseCoinNormalized(lastTransfer.Amount)
	if err != nil {
		return err
	}

	outDenom := sf.OsmosisMsgSwapExactAmountIn.Routes[len(sf.OsmosisMsgSwapExactAmountIn.Routes)-1].TokenOutDenom

	if lastAmountReceived.Denom != outDenom {
		return fmt.Errorf("amount received denom %s is not equal to last route denom %s", lastAmountReceived.Denom, outDenom)
	}

	sf.TokenOut = lastAmountReceived

	// First token swapped in (if there are multiple pools we do not care about intermediates)
	sf.TokenIn = sf.OsmosisMsgSwapExactAmountIn.TokenIn
	sf.Address = sf.OsmosisMsgSwapExactAmountIn.Sender

	return nil
}

func (sf *WrapperMsgSwapExactAmountOut) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgSwapExactAmountOut = msg.(*gammTypes.MsgSwapExactAmountOut)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the tokens swapped
	tokensSwappedEvt := txModule.GetEventWithType(gammTypes.TypeEvtTokenSwapped, log)

	if tokensSwappedEvt == nil {
		transferEvt := txModule.GetEventWithType("transfer", log)

		tokenInDenom := sf.OsmosisMsgSwapExactAmountOut.TokenInDenom()

		for _, evt := range transferEvt.Attributes {
			// Get the first amount that matches the token in denom
			if evt.Key == EventAttributeAmount {
				tokenIn, err := sdk.ParseCoinNormalized(evt.Value)
				if err != nil {
					return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
				}

				if tokenIn.Denom == tokenInDenom {
					sf.TokenIn = tokenIn
					break
				}

			}
		}

		senderReceiver, err := txModule.GetValueForAttribute("sender", transferEvt)
		if err != nil {
			return err
		}

		if senderReceiver == "" {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		sf.Address = senderReceiver

		tokenOutStr := txModule.GetLastValueForAttribute(EventAttributeAmount, transferEvt)
		tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		sf.TokenOut = tokenOut

		if sf.TokenIn.IsNil() {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
	} else {
		// This gets the first token swapped in (if there are multiple pools we do not care about intermediates)
		tokenInStr, err := txModule.GetValueForAttribute(gammTypes.AttributeKeyTokensIn, tokensSwappedEvt)
		if err != nil {
			return err
		}

		tokenIn, err := sdk.ParseCoinNormalized(tokenInStr)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		sf.TokenIn = tokenIn

		// Address of whoever initiated the swap. Will be both sender/receiver.
		senderReceiver, err := txModule.GetValueForAttribute("sender", tokensSwappedEvt)
		if err != nil {
			return err
		}

		if senderReceiver == "" {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		sf.Address = senderReceiver

		// This gets the last token swapped out (if there are multiple pools we do not care about intermediates)
		tokenOutStr := txModule.GetLastValueForAttribute(gammTypes.AttributeKeyTokensOut, tokensSwappedEvt)
		tokenOut, err := sdk.ParseCoinNormalized(tokenOutStr)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}
		sf.TokenOut = tokenOut

	}

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

func (sf *WrapperMsgSwapExactAmountIn2) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return sf.WrapperMsgSwapExactAmountIn.ParseRelevantData()
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
