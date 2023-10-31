package db

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func IndexBlockEvents(db *gorm.DB, dryRun bool, blockHeight int64, blockTime time.Time, blockDBWrapper *BlockDBWrapper, dbChainID uint, dbChainName string, identifierLoggingString string) error {
	return db.Transaction(func(dbTransaction *gorm.DB) error {
		// TODO: Delete from FailedEventBlock table

		// create block if it doesn't exist
		blockDBWrapper.Block.ChainID = dbChainID
		blockDBWrapper.Block.TimeStamp = blockTime
		blockDBWrapper.Block.BlockEventsIndexed = true

		if err := dbTransaction.
			Where(models.Block{Height: blockHeight, ChainID: dbChainID}).
			Assign(models.Block{BlockEventsIndexed: true, TimeStamp: blockTime}).
			FirstOrCreate(&blockDBWrapper.Block).Error; err != nil {
			config.Log.Error("Error getting/creating block DB object.", err)
			return err
		}

		var uniqueBlockEventTypes []models.BlockEventType

		for _, value := range blockDBWrapper.UniqueBlockEventTypes {
			uniqueBlockEventTypes = append(uniqueBlockEventTypes, value)
		}

		// Bulk find or create on unique event types
		if err := dbTransaction.Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"}, {Name: "type"},
				},
			},
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "type"}},
				DoUpdates: clause.AssignmentColumns([]string{"type"}),
			},
		).Create(&uniqueBlockEventTypes).Error; err != nil {
			config.Log.Error("Error creating begin block event types.", err)
			return nil
		}

		for _, value := range uniqueBlockEventTypes {
			blockDBWrapper.UniqueBlockEventTypes[value.Type] = value
		}

		var uniqueBlockEventAttributeKeys []models.BlockEventAttributeKey

		for _, value := range blockDBWrapper.UniqueBlockEventAttributeKeys {
			uniqueBlockEventAttributeKeys = append(uniqueBlockEventAttributeKeys, value)
		}

		if err := dbTransaction.Clauses(
			clause.Returning{
				Columns: []clause.Column{
					{Name: "id"}, {Name: "key"},
				},
			},
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "key"}},
				DoUpdates: clause.AssignmentColumns([]string{"key"}),
			},
		).Create(&uniqueBlockEventAttributeKeys).Error; err != nil {
			config.Log.Error("Error creating begin block event attribute keys.", err)
			return nil
		}

		for _, value := range uniqueBlockEventAttributeKeys {
			blockDBWrapper.UniqueBlockEventAttributeKeys[value.Key] = value
		}

		// Loop through begin and end block arrays and apply the block ID and event type ID
		beginBlockEvents := make([]*models.BlockEvent, len(blockDBWrapper.BeginBlockEvents))
		for index := range blockDBWrapper.BeginBlockEvents {
			blockDBWrapper.BeginBlockEvents[index].BlockEvent.Block = *blockDBWrapper.Block
			blockDBWrapper.BeginBlockEvents[index].BlockEvent.BlockID = blockDBWrapper.Block.ID
			blockDBWrapper.BeginBlockEvents[index].BlockEvent.BlockEventType = blockDBWrapper.UniqueBlockEventTypes[blockDBWrapper.BeginBlockEvents[index].BlockEvent.BlockEventType.Type]
			beginBlockEvents[index] = &blockDBWrapper.BeginBlockEvents[index].BlockEvent
		}

		endBlockEvents := make([]*models.BlockEvent, len(blockDBWrapper.EndBlockEvents))
		for index := range blockDBWrapper.EndBlockEvents {
			blockDBWrapper.EndBlockEvents[index].BlockEvent.Block = *blockDBWrapper.Block
			blockDBWrapper.EndBlockEvents[index].BlockEvent.BlockID = blockDBWrapper.Block.ID
			blockDBWrapper.EndBlockEvents[index].BlockEvent.BlockEventType = blockDBWrapper.UniqueBlockEventTypes[blockDBWrapper.EndBlockEvents[index].BlockEvent.BlockEventType.Type]
			endBlockEvents[index] = &blockDBWrapper.EndBlockEvents[index].BlockEvent
		}

		// Bulk insert the block events
		var allBlockEvents []*models.BlockEvent = append(beginBlockEvents, endBlockEvents...)
		if len(allBlockEvents) != 0 {
			if err := dbTransaction.Clauses(clause.OnConflict{DoNothing: true}).Create(&allBlockEvents).Error; err != nil {
				config.Log.Error("Error creating begin block events.", err)
				return nil
			}

			var allAttributes []models.BlockEventAttribute
			for index := range blockDBWrapper.BeginBlockEvents {
				currAttributes := blockDBWrapper.BeginBlockEvents[index].Attributes
				for attrIndex := range currAttributes {
					currAttributes[attrIndex].BlockEventID = blockDBWrapper.BeginBlockEvents[index].BlockEvent.ID
					currAttributes[attrIndex].BlockEvent = blockDBWrapper.BeginBlockEvents[index].BlockEvent
					currAttributes[attrIndex].BlockEventAttributeKey = blockDBWrapper.UniqueBlockEventAttributeKeys[currAttributes[attrIndex].BlockEventAttributeKey.Key]
				}
				allAttributes = append(allAttributes, currAttributes...)
			}

			for index := range blockDBWrapper.EndBlockEvents {
				currAttributes := blockDBWrapper.EndBlockEvents[index].Attributes
				for attrIndex := range currAttributes {
					currAttributes[attrIndex].BlockEventID = blockDBWrapper.EndBlockEvents[index].BlockEvent.ID
					currAttributes[attrIndex].BlockEvent = blockDBWrapper.EndBlockEvents[index].BlockEvent
					currAttributes[attrIndex].BlockEventAttributeKey = blockDBWrapper.UniqueBlockEventAttributeKeys[currAttributes[attrIndex].BlockEventAttributeKey.Key]
				}
				allAttributes = append(allAttributes, currAttributes...)
			}

			if len(allAttributes) != 0 {
				if err := dbTransaction.Clauses(clause.OnConflict{DoNothing: true}).Create(&allAttributes).Error; err != nil {
					config.Log.Error("Error creating begin block event attributes.", err)
					return nil
				}
			}
		}

		return nil
	})
}
