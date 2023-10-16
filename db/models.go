package db

import (
	"time"

	"github.com/shopspring/decimal"
)

type Block struct {
	ID           uint
	TimeStamp    time.Time
	Height       int64 `gorm:"uniqueIndex:chainheight"`
	BlockchainID uint  `gorm:"uniqueIndex:chainheight"`
	Chain        Chain `gorm:"foreignKey:BlockchainID"`
	Indexed      bool
}

type FailedBlock struct {
	ID           uint
	Height       int64 `gorm:"uniqueIndex:failedchainheight"`
	BlockchainID uint  `gorm:"uniqueIndex:failedchainheight"`
	Chain        Chain `gorm:"foreignKey:BlockchainID"`
}

type FailedEventBlock struct {
	ID           uint
	Height       int64 `gorm:"uniqueIndex:failedchaineventheight"`
	BlockchainID uint  `gorm:"uniqueIndex:failedchaineventheight"`
	Chain        Chain `gorm:"foreignKey:BlockchainID"`
}

type Chain struct {
	ID      uint   `gorm:"primaryKey"`
	ChainID string `gorm:"uniqueIndex"` // e.g. osmosis-1
	Name    string // e.g. Osmosis
}

type Tx struct {
	ID              uint
	Hash            string `gorm:"uniqueIndex"`
	Code            uint32
	BlockID         uint
	Block           Block
	SignerAddressID *uint // *int allows foreign key to be null
	SignerAddress   Address
	Fees            []Fee
}

type Fee struct {
	ID             uint            `gorm:"primaryKey"`
	TxID           uint            `gorm:"uniqueIndex:txDenomFee"`
	Amount         decimal.Decimal `gorm:"type:decimal(78,0);"`
	DenominationID uint            `gorm:"uniqueIndex:txDenomFee"`
	Denomination   Denom           `gorm:"foreignKey:DenominationID"`
	PayerAddressID uint            `gorm:"index:idx_payer_addr"`
	PayerAddress   Address         `gorm:"foreignKey:PayerAddressID"`
}

// dbTypes.Address{Address: currTx.FeePayer().String()}

type Address struct {
	ID      uint
	Address string `gorm:"uniqueIndex"`
}

type MessageType struct {
	ID          uint   `gorm:"primaryKey"`
	MessageType string `gorm:"uniqueIndex;not null"`
}

type Message struct {
	ID            uint
	TxID          uint `gorm:"index:idx_txid_typeid"`
	Tx            Tx
	MessageTypeID uint `gorm:"foreignKey:MessageTypeID,index:idx_txid_typeid"`
	MessageType   MessageType
	MessageIndex  int
}

const (
	OsmosisRewardDistribution uint = iota
	TendermintLiquidityDepositCoinsToPool
	TendermintLiquidityDepositPoolCoinReceived
	TendermintLiquiditySwapTransactedCoinIn
	TendermintLiquiditySwapTransactedCoinOut
	TendermintLiquiditySwapTransactedFee
	TendermintLiquidityWithdrawPoolCoinSent
	TendermintLiquidityWithdrawCoinReceived
	TendermintLiquidityWithdrawFee
	OsmosisProtorevDeveloperRewardDistribution
)

// An event does not necessarily need to be part of a Transaction. For example, Osmosis rewards.
// Events can happen on chain and generate tendermint ABCI events that do not show up in transactions.
type TaxableEvent struct {
	ID             uint
	Source         uint            // This will indicate what type of event occurred on chain. Currently, only used for Osmosis rewards.
	Amount         decimal.Decimal `gorm:"type:decimal(78,0);"` // 2^256 or 78 digits, cosmos Int can be up to this length
	DenominationID uint
	Denomination   Denom   `gorm:"foreignKey:DenominationID"`
	AddressID      uint    `gorm:"index:idx_addr"`
	EventAddress   Address `gorm:"foreignKey:AddressID"`
	EventHash      string  `gorm:"uniqueIndex:idx_teevthash"`
	BlockID        uint    `gorm:"index:idx_teblkid"`
	Block          Block   `gorm:"foreignKey:BlockID"`
}

// type SimpleDenom struct {
// 	ID     uint
// 	Denom  string `gorm:"uniqueIndex:denom_idx"`
// 	Symbol string `gorm:"uniqueIndex:denom_idx"`
// }

func (TaxableEvent) TableName() string {
	return "taxable_event"
}

type TaxableTransaction struct {
	ID                     uint
	MessageID              uint            `gorm:"index:idx_msg"`
	Message                Message         `gorm:"foreignKey:MessageID"`
	AmountSent             decimal.Decimal `gorm:"type:decimal(78,0);"`
	AmountReceived         decimal.Decimal `gorm:"type:decimal(78,0);"`
	DenominationSentID     *uint
	DenominationSent       Denom `gorm:"foreignKey:DenominationSentID"`
	DenominationReceivedID *uint
	DenominationReceived   Denom `gorm:"foreignKey:DenominationReceivedID"`
	SenderAddressID        *uint `gorm:"index:idx_sender"`
	SenderAddress          Address
	ReceiverAddressID      *uint `gorm:"index:idx_receiver"`
	ReceiverAddress        Address
}

func (TaxableTransaction) TableName() string {
	return "taxable_tx" // Legacy
}

type Denom struct {
	ID     uint
	Base   string `gorm:"uniqueIndex"`
	Name   string
	Symbol string
}

type DenomUnit struct {
	ID       uint
	DenomID  uint `gorm:"uniqueIndex:,composite:denom_id_name"`
	Denom    Denom
	Exponent uint
	Name     string `gorm:"uniqueIndex:,composite:denom_id_name"`
}

// Store transactions with their messages for easy database creation
type TxDBWrapper struct {
	Tx            Tx
	SignerAddress Address
	Messages      []MessageDBWrapper
}

// Store messages with their taxable events for easy database creation
type MessageDBWrapper struct {
	Message    Message
	TaxableTxs []TaxableTxDBWrapper
}

// Store taxable tx with their sender/receiver address for easy database creation
type TaxableTxDBWrapper struct {
	TaxableTx       TaxableTransaction
	SenderAddress   Address
	ReceiverAddress Address
}

type DenomDBWrapper struct {
	Denom      Denom
	DenomUnits []DenomUnitDBWrapper
}

type DenomUnitDBWrapper struct {
	DenomUnit DenomUnit
}

type IBCDenom struct {
	ID        uint
	Hash      string `gorm:"uniqueIndex"`
	Path      string
	BaseDenom string
}

type Epoch struct {
	ID           uint
	BlockchainID uint   `gorm:"uniqueIndex:chainepochidentifierheight"`
	Chain        Chain  `gorm:"foreignKey:BlockchainID"`
	StartHeight  uint   `gorm:"uniqueIndex:chainepochidentifierheight"`
	Identifier   string `gorm:"uniqueIndex:chainepochidentifierheight"`
	EpochNumber  uint
	Indexed      bool `gorm:"default:false"`
}
