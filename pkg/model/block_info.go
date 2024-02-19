package model

import (
	"time"

	"github.com/shopspring/decimal"
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
