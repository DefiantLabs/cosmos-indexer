package filter

import (
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/cosmos/cosmos-sdk/types"
)

type MessageFilter interface {
	ShouldIndex(types.Msg, tx.LogMessage) bool
}
