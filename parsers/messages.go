package parsers

import (
	"github.com/DefiantLabs/cosmos-indexer/config"
	txtypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"gorm.io/gorm"
)

// Intermediate type for the database inserted message datasets
// Is there a way to remove this? It may require a one-to-many mapping of the message events + attributes instead of the belongs-to
type MessageEventWithAttributes struct {
	Event      models.MessageEvent
	Attributes []models.MessageEventAttribute
}

type MessageParser interface {
	Identifier() string
	ParseMessage(sdkTypes.Msg, *txtypes.LogMessage, config.IndexConfig) (*any, error)
	IndexMessage(*any, *gorm.DB, models.Message, []MessageEventWithAttributes, config.IndexConfig) error
}

type MessageParsedData struct {
	Data   *any
	Error  error
	Parser *MessageParser
}
