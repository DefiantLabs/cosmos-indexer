package bank

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	MsgSendV0 = "MsgSend"
	MsgSend   = "/cosmos.bank.v1beta1.MsgSend"

	MsgMultiSendV0 = "MsgMultiSend"
	MsgMultiSend   = "/cosmos.bank.v1beta1.MsgMultiSend"
)

// HandleMsg: Unmarshal JSON for MsgSend.
// Note that MsgSend ignores the LogMessage because it isn't needed.
func (sf *WrapperMsgSend) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgSend = msg.(*bankTypes.MsgSend)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator withdrawal address and amount received
	receivedCoinsEvt := txModule.GetEventWithType(bankTypes.EventTypeTransfer, log)
	if receivedCoinsEvt == nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	receiverAddress, err := txModule.GetValueForAttribute(bankTypes.AttributeKeyRecipient, receivedCoinsEvt)
	if err != nil {
		return err
	}
	// coins_received := txModule.GetValueForAttribute("amount", receivedCoinsEvt)

	if !strings.EqualFold(sf.CosmosMsgSend.ToAddress, receiverAddress) {
		return fmt.Errorf("transaction receiver address %s does not match log event '%s' receiver address %s",
			sf.CosmosMsgSend.ToAddress, bankTypes.EventTypeCoinReceived, receiverAddress)
	}

	return nil
}

func (sf *WrapperMsgMultiSend) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgMultiSend = msg.(*bankTypes.MsgMultiSend)

	// Make sure the that the total amount sent matches the total amount received for each coin
	// sum up input coins
	sentMap := make(map[string]sdk.Int)
	for _, input := range sf.CosmosMsgMultiSend.Inputs {
		for _, coin := range input.Coins {
			if currentTotal, ok := sentMap[coin.Denom]; ok {
				sentMap[coin.Denom] = currentTotal.Add(coin.Amount)
			} else {
				sentMap[coin.Denom] = coin.Amount
			}
		}
	}

	// sum up output coins
	recievedMap := make(map[string]sdk.Int)
	for _, output := range sf.CosmosMsgMultiSend.Outputs {
		for _, coin := range output.Coins {
			if currentTotal, ok := recievedMap[coin.Denom]; ok {
				recievedMap[coin.Denom] = currentTotal.Add(coin.Amount)
			} else {
				recievedMap[coin.Denom] = coin.Amount
			}
		}
	}

	// compare
	for coin, amountSent := range sentMap {
		if amountReceived, coinWasReceived := recievedMap[coin]; coinWasReceived {
			if !amountReceived.Equal(amountSent) {
				return fmt.Errorf("error processing MultiSend, inputs and outputs mismatch for denom %v, send %s != received %s", coin, amountSent, amountReceived)
			}
		} else {
			return fmt.Errorf("error processing MultiSend, sent denom %v does not appear in received coins", coin)
		}
	}

	return nil
}

func (sf *WrapperMsgSend) String() string {
	return fmt.Sprintf("MsgSend: Address %s received %s from %s",
		sf.CosmosMsgSend.ToAddress, sf.CosmosMsgSend.Amount, sf.CosmosMsgSend.FromAddress)
}

func (sf *WrapperMsgMultiSend) String() string {
	var sendsAndReceives []string
	for _, v := range sf.SenderReceiverAmounts {
		sendsAndReceives = append(sendsAndReceives, fmt.Sprintf("%s %s -> %s", v.Amount, v.Sender, v.Receiver))
	}
	return fmt.Sprintf("MsgMultiSend: %s", strings.Join(sendsAndReceives, ", "))
}

func (sf *WrapperMsgSend) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.CosmosMsgSend.Amount))

	for i, v := range sf.CosmosMsgSend.Amount {
		var currRelevantData parsingTypes.MessageRelevantInformation
		currRelevantData.SenderAddress = sf.CosmosMsgSend.FromAddress
		currRelevantData.ReceiverAddress = sf.CosmosMsgSend.ToAddress
		// Amount always seems to be an integer, float may be an extra uneeded step
		currRelevantData.AmountSent = v.Amount.BigInt()
		currRelevantData.DenominationSent = v.Denom

		// This is required since we do CSV parsing on the receiver here too
		currRelevantData.AmountReceived = v.Amount.BigInt()
		currRelevantData.DenominationReceived = v.Denom

		relevantData[i] = currRelevantData
	}

	return relevantData
}

func (sf *WrapperMsgMultiSend) ParseRelevantData() (relevantData []parsingTypes.MessageRelevantInformation) {
	for _, senderReceiverAmount := range sf.SenderReceiverAmounts {
		var currRelevantData parsingTypes.MessageRelevantInformation
		currRelevantData.SenderAddress = senderReceiverAmount.Sender
		currRelevantData.ReceiverAddress = senderReceiverAmount.Receiver

		currRelevantData.AmountSent = senderReceiverAmount.Amount.Amount.BigInt()
		currRelevantData.DenominationSent = senderReceiverAmount.Amount.Denom

		currRelevantData.AmountReceived = senderReceiverAmount.Amount.Amount.BigInt()
		currRelevantData.DenominationReceived = senderReceiverAmount.Amount.Denom

		relevantData = append(relevantData, currRelevantData)
	}

	return relevantData
}

type WrapperMsgSend struct {
	txModule.Message
	CosmosMsgSend *bankTypes.MsgSend
}

type WrapperMsgMultiSend struct {
	txModule.Message
	CosmosMsgMultiSend    *bankTypes.MsgMultiSend
	SenderReceiverAmounts []SenderReceiverAmount
}

type SenderReceiverAmount struct {
	Sender   string
	Receiver string
	Amount   sdk.Coin
}
