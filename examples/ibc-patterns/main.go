package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/DefiantLabs/cosmos-indexer/cmd"
	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/filter"
	"github.com/DefiantLabs/cosmos-indexer/parsers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	indexerTxTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	stdTypes "github.com/cosmos/cosmos-sdk/types"
	chanTypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
)

type MsgType int

const (
	MsgRecvPacket MsgType = iota
	MsgAcknowledgement
)

type TransactionType string

const (
	ValidatorUpdatesTransactionType TransactionType = "ValidatorUpdates"
	TokenTransferTransactionType    TransactionType = "TokenTransfer"
	CCVSlashTransactionType         TransactionType = "CCVSlash"
	CCVVSCMaturedTransactionType    TransactionType = "CCVVSCMatured"
)

type IBCTransactionType struct {
	ID              uint `gorm:"primaryKey"`
	TransactionType TransactionType
}

type IBCTransaction struct {
	ID                   uint `gorm:"primaryKey"`
	MessageID            uint `gorm:"uniqueIndex"`
	Message              models.Message
	IBCMsgType           MsgType
	IBCTransactionTypeID uint
	IBCTransactionType   IBCTransactionType
	ChainIBCPathID       uint
	ChainIBCPath         ChainIBCPath
}

type ChainIBCPath struct {
	ID              uint `gorm:"primaryKey"`
	ChainID         uint `gorm:"uniqueIndex:chain_ibc_path,priority:1"`
	Chain           models.Chain
	ChainChannel    string `gorm:"uniqueIndex:chain_ibc_path,priority:2"`
	ChainPort       string `gorm:"uniqueIndex:chain_ibc_path,priority:3"`
	OffchainChannel string `gorm:"uniqueIndex:chain_ibc_path,priority:4"`
	OffchainPort    string `gorm:"uniqueIndex:chain_ibc_path,priority:5"`
}

type IBCTransactionParser struct {
	UniqueID string
}

func (c *IBCTransactionParser) Identifier() string {
	return c.UniqueID
}

func (c *IBCTransactionParser) ParseMessage(cosmosMsg stdTypes.Msg, log *indexerTxTypes.LogMessage, cfg config.IndexConfig) (*any, error) {
	// Check if this is a MsgAcknowledgement
	msgAck, okMsgAck := cosmosMsg.(*chanTypes.MsgAcknowledgement)

	if okMsgAck {
		parsedMsgAck, err := parseMsgAcknowledgement(msgAck)
		if err != nil {
			return nil, fmt.Errorf("error parsing MsgAcknowledgement: %w", err)
		}

		anyCast := any(parsedMsgAck)

		return &anyCast, nil
	}

	// Check if this is a MsgRecvPacket
	msgRecvPacket, okMsgRecvPacket := cosmosMsg.(*chanTypes.MsgRecvPacket)

	if okMsgRecvPacket {
		parsedMsgRecvPacket, err := parseMsgRecvPacket(msgRecvPacket)
		if err != nil {
			return nil, fmt.Errorf("error parsing MsgAcknowledgement: %w", err)
		}

		anyCast := any(parsedMsgRecvPacket)

		return &anyCast, nil
	}

	return nil, fmt.Errorf("unsupported message type passed to parser")
}

type IBCTransactionParsedData struct {
	ParsedIBCMessage   *IBCTransaction
	IBCTransactionType *IBCTransactionType
	ChainIBCPath       *ChainIBCPath
}

func parseMsgAcknowledgement(msg *chanTypes.MsgAcknowledgement) (IBCTransactionParsedData, error) {
	parsedData := IBCTransactionParsedData{
		ParsedIBCMessage: &IBCTransaction{
			IBCMsgType: MsgAcknowledgement,
		},
	}

	ibcTransactionType, err := determinePacketType(msg.Packet.Data)
	if err != nil {
		return parsedData, err
	}

	parsedData.IBCTransactionType = &ibcTransactionType

	parsedData.ChainIBCPath = &ChainIBCPath{
		ChainChannel:    msg.Packet.GetSourceChannel(),
		ChainPort:       msg.Packet.GetSourcePort(),
		OffchainChannel: msg.Packet.GetDestChannel(),
		OffchainPort:    msg.Packet.GetDestPort(),
	}

	return parsedData, nil
}

func parseMsgRecvPacket(msg *chanTypes.MsgRecvPacket) (IBCTransactionParsedData, error) {
	parsedData := IBCTransactionParsedData{
		ParsedIBCMessage: &IBCTransaction{
			IBCMsgType: MsgRecvPacket,
		},
	}

	ibcTransactionType, err := determinePacketType(msg.Packet.Data)
	if err != nil {
		return parsedData, err
	}

	if ibcTransactionType.TransactionType == "" {
		return parsedData, fmt.Errorf("unsupported packet data")
	}

	parsedData.IBCTransactionType = &ibcTransactionType

	parsedData.ChainIBCPath = &ChainIBCPath{
		ChainChannel:    msg.Packet.GetDestChannel(),
		ChainPort:       msg.Packet.GetDestPort(),
		OffchainChannel: msg.Packet.GetSourceChannel(),
		OffchainPort:    msg.Packet.GetSourcePort(),
	}

	return parsedData, nil
}

type ValidatorUpdates struct {
	ValidatorUpdates *[]json.RawMessage `json:"validator_updates"`
	ValsetUpdateID   *string            `json:"valset_update_id"`
}

type TokenTransfer struct {
	Denom    *string `json:"denom"`
	Amount   *string `json:"amount"`
	Sender   *string `json:"sender"`
	Receiver *string `json:"receiver"`
}

type VSCMatured struct {
	VscMaturedPacketData *json.RawMessage `json:"vscMaturedPacketData"`
}

type Slash struct {
	Validator      *json.RawMessage `json:"validator"`
	ValsetUpdateID *string          `json:"valset_update_id"`
}

func determinePacketType(packetData []byte) (IBCTransactionType, error) {
	// determine the type of packet by attempting to cast the data to our supported JSON structs
	var ibcTransactionType IBCTransactionType

	var validatorUpdates ValidatorUpdates
	err := json.Unmarshal(packetData, &validatorUpdates)
	if err != nil {
		return ibcTransactionType, err
	}

	var tokenTransfer TokenTransfer
	err = json.Unmarshal(packetData, &tokenTransfer)

	if err != nil {
		return ibcTransactionType, err
	}

	var vscMatured VSCMatured
	err = json.Unmarshal(packetData, &vscMatured)

	if err != nil {
		return ibcTransactionType, err
	}

	var slash Slash
	err = json.Unmarshal(packetData, &slash)

	if err != nil {
		return ibcTransactionType, err
	}

	switch {
	case validatorUpdates.ValidatorUpdates != nil && len(*validatorUpdates.ValidatorUpdates) > 0 && validatorUpdates.ValsetUpdateID != nil:
		ibcTransactionType.TransactionType = ValidatorUpdatesTransactionType
	case tokenTransfer.Denom != nil && tokenTransfer.Amount != nil && tokenTransfer.Sender != nil && tokenTransfer.Receiver != nil:
		ibcTransactionType.TransactionType = TokenTransferTransactionType
	case vscMatured.VscMaturedPacketData != nil:
		ibcTransactionType.TransactionType = CCVVSCMaturedTransactionType
	case slash.Validator != nil && slash.ValsetUpdateID != nil:
		ibcTransactionType.TransactionType = CCVSlashTransactionType
	default:
		return ibcTransactionType, fmt.Errorf("unsupported packet data")
	}

	return ibcTransactionType, nil
}

func (c *IBCTransactionParser) IndexMessage(dataset *any, db *gorm.DB, message models.Message, messageEvents []parsers.MessageEventWithAttributes, cfg config.IndexConfig) error {
	ibcTransaction, ok := (*dataset).(IBCTransactionParsedData)

	if !ok {
		return fmt.Errorf("invalid IBC transaction type passed to parser index message function")
	}

	ibcTransactionType := ibcTransaction.IBCTransactionType
	parsedIBCMessage := ibcTransaction.ParsedIBCMessage

	// Create or update the IBC transaction type
	err := db.Where(&ibcTransactionType).FirstOrCreate(&ibcTransactionType).Error
	if err != nil {
		return err
	}

	// Set the IBC transaction type ID on the IBC transaction and link it to the default message model
	parsedIBCMessage.IBCTransactionTypeID = ibcTransactionType.ID
	parsedIBCMessage.IBCTransactionType = *ibcTransactionType
	parsedIBCMessage.Message = message
	parsedIBCMessage.MessageID = message.ID

	ibcTransaction.ChainIBCPath.ChainID = message.Tx.Block.ChainID

	// attempt to find relation for chain path

	chainPath := ChainIBCPath{
		ChainID:         message.Tx.Block.ChainID,
		ChainChannel:    ibcTransaction.ChainIBCPath.ChainChannel,
		ChainPort:       ibcTransaction.ChainIBCPath.ChainPort,
		OffchainChannel: ibcTransaction.ChainIBCPath.OffchainChannel,
		OffchainPort:    ibcTransaction.ChainIBCPath.OffchainPort,
	}

	err = db.Where(&chainPath).FirstOrCreate(&chainPath).Error

	if err != nil {
		return err
	}

	parsedIBCMessage.ChainIBCPathID = chainPath.ID
	parsedIBCMessage.ChainIBCPath = chainPath

	// Create or update the IBC transaction
	if err := db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "message_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"ibc_msg_type", "ibc_transaction_type_id", "chain_ibc_path_id"}),
	}).Create(parsedIBCMessage).Error; err != nil {
		return err
	}

	return nil
}

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.RegisterCustomModels([]any{IBCTransactionType{}, IBCTransaction{}, ChainIBCPath{}})

	// This indexer is only concerned with MsgRecvPacket and MsgAcknowledgement messages, so we create regex filters to only index those messages.
	// This significantly reduces the size of the indexed dataset, saving space and processing time.
	// We use a regex because the message type so we can match both message types in one filter.
	ibcRegexMessageTypeFilter, err := filter.NewRegexMessageTypeFilter("^/ibc.core.channel.v1.Msg(RecvPacket|Acknowledgement)$")
	if err != nil {
		log.Fatalf("Failed to create regex message type filter. Err: %v", err)
	}

	indexer.RegisterMessageTypeFilter(ibcRegexMessageTypeFilter)

	ibcRecvParser := &IBCTransactionParser{
		UniqueID: "ibc-recv-parser",
	}

	ibcAckParser := &IBCTransactionParser{
		UniqueID: "ibc-ack-parser",
	}

	indexer.RegisterCustomMessageParser("/ibc.core.channel.v1.MsgRecvPacket", ibcRecvParser)
	indexer.RegisterCustomMessageParser("/ibc.core.channel.v1.MsgAcknowledgement", ibcAckParser)

	err = cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
