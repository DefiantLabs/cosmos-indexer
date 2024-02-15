package db

import (
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"gorm.io/gorm"
)

func FindOrCreateCustomBlockEventParsers(db *gorm.DB, parsers map[string]models.BlockEventParser) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		for key := range parsers {
			currParser := parsers[key]
			res := db.FirstOrCreate(&currParser, &currParser)

			if res.Error != nil {
				return res.Error
			}
			parsers[key] = currParser
		}
		return nil
	})
	return err
}

func FindOrCreateCustomMessageParsers(db *gorm.DB, parsers map[string]models.MessageParser) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		for key := range parsers {
			currParser := parsers[key]
			res := db.FirstOrCreate(&currParser, &currParser)

			if res.Error != nil {
				return res.Error
			}
			parsers[key] = currParser
		}
		return nil
	})
	return err
}

func CreateBlockEventParserError(db *gorm.DB, blockEvent models.BlockEvent, parser models.BlockEventParser, parserError error) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		res := db.Create(&models.BlockEventParserError{
			BlockEventParserID: parser.ID,
			BlockEventID:       blockEvent.ID,
			Error:              parserError.Error(),
		})
		return res.Error
	})
	return err
}

func DeleteCustomBlockEventParserError(db *gorm.DB, blockEvent models.BlockEvent, parser models.BlockEventParser) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		parserError := models.BlockEventParserError{
			BlockEventParserID: parser.ID,
			BlockEventID:       blockEvent.ID,
		}
		res := db.Where(&parserError).Delete(&parserError)
		return res.Error
	})
	return err
}

func CreateMessageParserError(db *gorm.DB, message models.Message, parser models.MessageParser, parserError error) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		res := db.Create(&models.MessageParserError{
			Error:           parserError.Error(),
			MessageParserID: parser.ID,
			MessageID:       message.ID,
		})
		return res.Error
	})
	return err
}

func DeleteCustomMessageParserError(db *gorm.DB, message models.Message, parser models.MessageParser) error {
	err := db.Transaction(func(dbTransaction *gorm.DB) error {
		parserError := models.MessageParserError{
			MessageParserID: parser.ID,
			MessageID:       message.ID,
		}
		res := db.Where(&parserError).Delete(&parserError)
		return res.Error
	})
	return err
}
