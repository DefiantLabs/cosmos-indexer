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

type BlockDBWrapper struct {
	Block                         *models.Block
	BeginBlockEvents              []BlockEventDBWrapper
	EndBlockEvents                []BlockEventDBWrapper
	UniqueBlockEventTypes         map[string]models.BlockEventType
	UniqueBlockEventAttributeKeys map[string]models.BlockEventAttributeKey
}

type BlockEventDBWrapper struct {
	BlockEvent models.BlockEvent
	Attributes []models.BlockEventAttribute
}

// Store transactions with their messages for easy database creation
type TxDBWrapper struct {
	Tx                         models.Tx
	SignerAddress              models.Address
	Messages                   []MessageDBWrapper
	UniqueMessageEventTypes    map[string]models.MessageEventType
	UniqueMessageAttributeKeys map[string]models.MessageEventAttributeKey
}

type MessageDBWrapper struct {
	Message       models.Message
	MessageEvents []MessageEventDBWrapper
}

type MessageEventDBWrapper struct {
	MessageEvent models.MessageEvent
	Attributes   []models.MessageEventAttribute
}

type DenomDBWrapper struct {
	Denom      models.Denom
	DenomUnits []DenomUnitDBWrapper
}

type DenomUnitDBWrapper struct {
	DenomUnit models.DenomUnit
}
