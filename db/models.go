package db

import "github.com/DefiantLabs/cosmos-indexer/db/models"

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

// Store transactions with their messages for easy database creation
type TxDBWrapper struct {
	Tx            models.Tx
	SignerAddress models.Address
	Messages      []MessageDBWrapper
}

type MessageDBWrapper struct {
	Message models.Message
}

type DenomDBWrapper struct {
	Denom      models.Denom
	DenomUnits []DenomUnitDBWrapper
}

type DenomUnitDBWrapper struct {
	DenomUnit models.DenomUnit
}
