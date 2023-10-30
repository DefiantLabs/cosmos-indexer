package models

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

type IBCDenom struct {
	ID        uint
	Hash      string `gorm:"uniqueIndex"`
	Path      string
	BaseDenom string
}
