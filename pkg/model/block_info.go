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
	TotalTx                  int64
	BlockHash                string
	GasUsed                  decimal.Decimal
	GasWanted                decimal.Decimal
	BlockRewards             decimal.Decimal
}

type Validators struct {
	Address string
}

type TotalBlocks struct {
	BlockHeight int64           `json:"block_height"`
	Count24H    int64           `json:"count_24h"`
	Count48H    int64           `json:"count_48h"`
	BlockTime   int64           `json:"block_time"`
	TotalFee24H decimal.Decimal `json:"total_fee_24h"`
}

type BlockSigners struct {
	BlockHeight int64
	Validator   string
	Time        time.Time
	Rank        int64
}

type AggregatedInfo struct {
	UpdatedAt    time.Time         `json:"updated_at"`
	Blocks       TotalBlocks       `json:"blocks"`
	Transactions TotalTransactions `json:"transactions"`
	Wallets      TotalWallets      `json:"wallets"`
}

type TotalWallets struct {
	Total    int64 `json:"total"`
	Count24H int64 `json:"count_24h"`
	Count48H int64 `json:"count_48h"`
}
