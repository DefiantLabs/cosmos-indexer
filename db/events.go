package db

import (
	"time"

	"gorm.io/gorm"
)

func IndexBlockEvents(db *gorm.DB, dryRun bool, blockHeight int64, blockTime time.Time, blockDBWrapper *BlockDBWrapper, dbChainID string, dbChainName string, identifierLoggingString string) error {
	// TODO: Stub for when indexing block events is generalized
	return nil
}
