package models

type Address struct {
	ID      uint
	Address string `gorm:"uniqueIndex"`
}
