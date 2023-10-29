package models

import (
	"time"
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

type BlockEventType int

const (
	BeginBlockEvent BlockEventType = iota
	EndBlockEvent
)

type BlockEvent struct {
	ID    uint
	Key   string
	Value string
	Index uint64
	Type  BlockEventType
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
