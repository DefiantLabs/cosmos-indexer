package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type BlockInfo struct {
	BlockHeight              int64
	ProposedValidatorAddress string
	GenerationTime           time.Time
	TimeElapsed              int64
	TotalFees                decimal.Decimal
	BlockHash                string
}

type Validators struct {
	Address string
}
