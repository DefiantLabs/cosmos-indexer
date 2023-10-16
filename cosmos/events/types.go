package events

import (
	"github.com/tendermint/tendermint/abci/types"
)

type CosmosEvent interface {
	HandleEvent(string, types.Event) error
	ParseRelevantData() []EventRelevantInformation
	GetType() string
	String() string
}
