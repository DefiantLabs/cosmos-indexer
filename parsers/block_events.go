package parsers

import (
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/nodersteam/cosmos-indexer/config"
	"github.com/nodersteam/cosmos-indexer/db/models"
	"gorm.io/gorm"
)

type BlockEventParser interface {
	Identifier() string
	ParseBlockEvent(abci.Event, config.IndexConfig) (*any, error)
	IndexBlockEvent(*any, *gorm.DB, models.Block, models.BlockEvent, []models.BlockEventAttribute, config.IndexConfig) error
}

type BlockEventParsedData struct {
	Data   *any
	Error  error
	Parser *BlockEventParser
}
