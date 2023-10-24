package ibc

import (
	"fmt"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	chantypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

const (
	MsgRecvPacket      = "/ibc.core.channel.v1.MsgRecvPacket"
	MsgAcknowledgement = "/ibc.core.channel.v1.MsgAcknowledgement"

	// Explicitly ignored messages for tx parsing purposes
	MsgTransfer           = "/ibc.applications.transfer.v1.MsgTransfer"
	MsgChannelOpenTry     = "/ibc.core.channel.v1.MsgChannelOpenTry"
	MsgChannelOpenConfirm = "/ibc.core.channel.v1.MsgChannelOpenConfirm"
	MsgChannelOpenInit    = "/ibc.core.channel.v1.MsgChannelOpenInit"
	MsgChannelOpenAck     = "/ibc.core.channel.v1.MsgChannelOpenAck"

	MsgTimeout        = "/ibc.core.channel.v1.MsgTimeout"
	MsgTimeoutOnClose = "/ibc.core.channel.v1.MsgTimeoutOnClose"

	MsgConnectionOpenTry     = "/ibc.core.connection.v1.MsgConnectionOpenTry"
	MsgConnectionOpenConfirm = "/ibc.core.connection.v1.MsgConnectionOpenConfirm"
	MsgConnectionOpenInit    = "/ibc.core.connection.v1.MsgConnectionOpenInit"
	MsgConnectionOpenAck     = "/ibc.core.connection.v1.MsgConnectionOpenAck"

	MsgChannelCloseConfirm = "/ibc.core.channel.v1.MsgChannelCloseConfirm"
	MsgChannelCloseInit    = "/ibc.core.channel.v1.MsgChannelCloseInit"

	MsgCreateClient = "/ibc.core.client.v1.MsgCreateClient"
	MsgUpdateClient = "/ibc.core.client.v1.MsgUpdateClient"

	// Consts used for classifying Ack messages
	// We may need to keep extending these consts for other types
	AckFungibleTokenTransfer    = 0
	AckNotFungibleTokenTransfer = 1

	// Same as above, we may want to to extend these to track other results
	AckSuccess = 0
	AckFailure = 1

	AlternateMsgAcknowledgementLogAction = "acknowledge_packet"
	AlternateMsgRcvLogAction             = "recv_packet"
)

type WrapperMsgRecvPacket struct {
	txModule.Message
	MsgRecvPacket   *chantypes.MsgRecvPacket
	Sequence        uint64
	SenderAddress   string
	ReceiverAddress string
	Amount          stdTypes.Int
	Denom           string
}

func (w *WrapperMsgRecvPacket) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	w.Type = msgType
	w.MsgRecvPacket = msg.(*chantypes.MsgRecvPacket)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(w.GetType(), log)
	alternateValidLog := txModule.IsMessageActionEquals(AlternateMsgRcvLogAction, log)

	if !validLog && !alternateValidLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// Unmarshal the json encoded packet data so we can access sender, receiver and denom info
	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(w.MsgRecvPacket.Packet.GetData(), &data); err != nil {
		// If there was a failure then this recv was not for a token transfer packet,
		// currently we only consider successful token transfers taxable events.
		return nil
	}

	w.SenderAddress = data.Sender
	w.ReceiverAddress = data.Receiver
	w.Sequence = w.MsgRecvPacket.Packet.Sequence

	amount, ok := stdTypes.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("failed to convert denom amount to sdk.Int, got(%s)", data.Amount)
	}

	w.Amount = amount
	w.Denom = data.Denom

	return nil
}

func (w *WrapperMsgRecvPacket) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// This prevents the item from being indexed
	if w.Amount.IsNil() {
		return nil
	}

	// MsgRecvPacket indicates a user has received assets on this chain so amount sent will always be 0
	amountSent := stdTypes.NewInt(0)

	return []parsingTypes.MessageRelevantInformation{{
		SenderAddress:        w.SenderAddress,
		ReceiverAddress:      w.ReceiverAddress,
		AmountSent:           amountSent.BigInt(),
		AmountReceived:       w.Amount.BigInt(),
		DenominationSent:     "",
		DenominationReceived: w.Denom,
	}}
}

func (w *WrapperMsgRecvPacket) String() string {
	if w.Amount.IsNil() {
		return "MsgRecvPacket: IBC transfer was not a FungibleTokenTransfer"
	}
	return fmt.Sprintf("MsgRecvPacket: IBC transfer of %s%s from %s to %s", w.Amount, w.Denom, w.SenderAddress, w.ReceiverAddress)
}

type WrapperMsgAcknowledgement struct {
	txModule.Message
	MsgAcknowledgement *chantypes.MsgAcknowledgement
	Sequence           uint64
	SenderAddress      string
	ReceiverAddress    string
	Amount             stdTypes.Int
	Denom              string
	AckType            int
	AckResult          int
}

func (w *WrapperMsgAcknowledgement) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	w.Type = msgType
	w.MsgAcknowledgement = msg.(*chantypes.MsgAcknowledgement)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(w.GetType(), log)
	alternateValidLog := txModule.IsMessageActionEquals(AlternateMsgAcknowledgementLogAction, log)

	if !validLog && !alternateValidLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// Unmarshal the json encoded packet data so we can access sender, receiver and denom info
	var data types.FungibleTokenPacketData
	if err := types.ModuleCdc.UnmarshalJSON(w.MsgAcknowledgement.Packet.GetData(), &data); err != nil {
		// If there was a failure then this ack was not for a token transfer packet,
		// currently we only consider successful token transfers taxable events.
		w.AckType = AckNotFungibleTokenTransfer
		return nil
	}

	w.AckType = AckFungibleTokenTransfer

	w.SenderAddress = data.Sender
	w.ReceiverAddress = data.Receiver
	w.Sequence = w.MsgAcknowledgement.Packet.Sequence

	amount, ok := stdTypes.NewIntFromString(data.Amount)
	if !ok {
		return fmt.Errorf("failed to convert denom amount to sdk.Int, got(%s)", data.Amount)
	}

	// Acknowledgements can contain an error & we only want to index successful acks,
	// so we need to check the ack bytes to determine if it was a result or an error.
	var ack chantypes.Acknowledgement
	if err := types.ModuleCdc.UnmarshalJSON(w.MsgAcknowledgement.Acknowledgement, &ack); err != nil {
		return fmt.Errorf("cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	switch ack.Response.(type) {
	case *chantypes.Acknowledgement_Error:
		// We index nothing on Acknowledgement errors
		w.AckResult = AckFailure
		return nil
	default:
		// the acknowledgement succeeded on the receiving chain
		w.AckResult = AckSuccess
		w.Amount = amount
		w.Denom = data.Denom
		return nil
	}
}

func (w *WrapperMsgAcknowledgement) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// This prevents the item from being indexed
	if w.Amount.IsNil() || w.AckType == AckNotFungibleTokenTransfer || w.AckResult == AckFailure {
		return nil
	}

	// MsgAcknowledgement indicates a user has successfully sent a packet
	// so the received amount will always be zero
	amountReceived := stdTypes.NewInt(0)

	return []parsingTypes.MessageRelevantInformation{{
		SenderAddress:        w.SenderAddress,
		ReceiverAddress:      w.ReceiverAddress,
		AmountSent:           w.Amount.BigInt(),
		AmountReceived:       amountReceived.BigInt(),
		DenominationSent:     w.Denom,
		DenominationReceived: "",
	}}
}

func (w *WrapperMsgAcknowledgement) String() string {
	if w.AckType == AckNotFungibleTokenTransfer {
		return "MsgAcknowledgement: IBC transfer was not a FungibleTokenTransfer"
	}

	if w.AckType == AckFungibleTokenTransfer && w.AckResult == AckFailure {
		return "MsgAcknowledgement: IBC transfer was not successful"
	}

	if w.Amount.IsNil() {
		return "MsgAcknowledgement: IBC transfer was not a FungibleTokenTransfer"
	}

	return fmt.Sprintf("MsgAcknowledgement: IBC transfer of %s%s from %s to %s\n", w.Amount, w.Denom, w.SenderAddress, w.ReceiverAddress)
}
