package db

import (
	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func IndexBlockEvents(db *gorm.DB, dryRun bool, blockDBWrapper *BlockDBWrapper, identifierLoggingString string) (*BlockDBWrapper, error) {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		if err := dbTransaction.
			Exec("DELETE FROM failed_event_blocks WHERE height = ? AND blockchain_id = ?", blockDBWrapper.Block.Height, blockDBWrapper.Block.ChainID).
			Error; err != nil {
			config.Log.Error("Error updating failed block.", err)
			return err
		}

		consAddress := blockDBWrapper.Block.ProposerConsAddress

		// create cons address if it doesn't exist
		if err := dbTransaction.Where(&consAddress).FirstOrCreate(&consAddress).Error; err != nil {
			config.Log.Error("Error getting/creating cons address DB object.", err)
			return err
		}

		// create block if it doesn't exist
		blockDBWrapper.Block.ProposerConsAddressID = consAddress.ID
		blockDBWrapper.Block.ProposerConsAddress = consAddress

		// create block if it doesn't exist
		blockDBWrapper.Block.BlockEventsIndexed = true

		if err := dbTransaction.
			Where(models.Block{Height: blockDBWrapper.Block.Height, ChainID: blockDBWrapper.Block.ChainID}).
			Assign(models.Block{BlockEventsIndexed: true, TimeStamp: blockDBWrapper.Block.TimeStamp, ProposerConsAddress: blockDBWrapper.Block.ProposerConsAddress}).
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
			return err
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
			return err
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
		allBlockEvents := make([]*models.BlockEvent, len(beginBlockEvents)+len(endBlockEvents))
		copy(allBlockEvents, beginBlockEvents)
		copy(allBlockEvents[len(beginBlockEvents):], endBlockEvents)

		// TODO: Should consider the on conflict values here, do we want to provide the user with some control over the behavior here?
		// Something similar to our reindex flag may be appropriate, unless we just want to have that pre-check the block has already been indexed.
		if len(allBlockEvents) != 0 {
			// This clause forces a return of ID for all items even on conflict
			// We need this so that we can then create the proper associations with the attributes below
			if err := dbTransaction.Clauses(
				clause.OnConflict{
					Columns: []clause.Column{{Name: "index"}, {Name: "lifecycle_position"}, {Name: "block_id"}},
					// Force update of block event type ID
					DoUpdates: clause.AssignmentColumns([]string{"block_event_type_id"}),
				},
			).Create(&allBlockEvents).Error; err != nil {
				config.Log.Error("Error creating begin block events.", err)
				return err
			}

			var allAttributes []*models.BlockEventAttribute
			for index := range blockDBWrapper.BeginBlockEvents {
				currAttributes := blockDBWrapper.BeginBlockEvents[index].Attributes
				for attrIndex := range currAttributes {
					currAttributes[attrIndex].BlockEventID = blockDBWrapper.BeginBlockEvents[index].BlockEvent.ID
					currAttributes[attrIndex].BlockEvent = blockDBWrapper.BeginBlockEvents[index].BlockEvent
					currAttributes[attrIndex].BlockEventAttributeKey = blockDBWrapper.UniqueBlockEventAttributeKeys[currAttributes[attrIndex].BlockEventAttributeKey.Key]
				}
				for ii := range currAttributes {
					allAttributes = append(allAttributes, &currAttributes[ii])
				}
			}

			for index := range blockDBWrapper.EndBlockEvents {
				currAttributes := blockDBWrapper.EndBlockEvents[index].Attributes
				for attrIndex := range currAttributes {
					currAttributes[attrIndex].BlockEventID = blockDBWrapper.EndBlockEvents[index].BlockEvent.ID
					currAttributes[attrIndex].BlockEvent = blockDBWrapper.EndBlockEvents[index].BlockEvent
					currAttributes[attrIndex].BlockEventAttributeKey = blockDBWrapper.UniqueBlockEventAttributeKeys[currAttributes[attrIndex].BlockEventAttributeKey.Key]
				}
				for ii := range currAttributes {
					allAttributes = append(allAttributes, &currAttributes[ii])
				}
			}

			if len(allAttributes) != 0 {
				if err := dbTransaction.Clauses(clause.OnConflict{
					Columns: []clause.Column{{Name: "block_event_id"}, {Name: "index"}},
					// Force update of value
					DoUpdates: clause.AssignmentColumns([]string{"value"}),
				}).Create(&allAttributes).Error; err != nil {
					config.Log.Error("Error creating begin block event attributes.", err)
					return err
				}
			}
		}

		return nil
	})

	// Contract: ensure that wrapper has been loaded with all data before returning
	return blockDBWrapper, err
}
