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
	// TODO: Should block event indexing be split out or rolled up?
	BlockEventsIndexed bool
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

type MessageDBWrapper struct {
	Message Message
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
