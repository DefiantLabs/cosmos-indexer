package parsers

import (
	"github.com/DefiantLabs/cosmos-indexer/config"
	txtypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	sdkTypes "github.com/cosmos/cosmos-sdk/types"
	"gorm.io/gorm"
)

type MessageParser interface {
	Identifier() string
	ParseMessage(sdkTypes.Msg, *txtypes.LogMessage, config.IndexConfig) (*any, error)
	IndexMessage(*any, *gorm.DB, models.Message, []models.MessageEvent, []models.MessageEventAttribute, config.IndexConfig) error
}

type MessageParsedData struct {
	Data   *any
	Error  error
	Parser *MessageParser
}
