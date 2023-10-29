package models

import "github.com/shopspring/decimal"

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
