package models

type Chain struct {
	ID      uint   `gorm:"primaryKey"`
	ChainID string `gorm:"uniqueIndex"` // e.g. osmosis-1
	Name    string // e.g. Osmosis
}
