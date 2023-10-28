package db

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	"gorm.io/gorm"
)

func IndexBlockEvents(db *gorm.DB, dryRun bool, blockHeight int64, blockTime time.Time, blockEvents []events.EventRelevantInformation, dbChainID string, dbChainName string, identifierLoggingString string) error {
	//TODO: Stub for when indexing block events is generalized
	return nil
}
