package models

type Denom struct {
	ID   uint
	Base string `gorm:"uniqueIndex"`
}
