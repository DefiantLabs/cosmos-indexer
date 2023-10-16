package db

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"gorm.io/gorm"
)

func IndexBlockEvents(db *gorm.DB, dryRun bool, blockHeight int64, blockTime time.Time, blockEvents []events.EventRelevantInformation, dbChainID string, dbChainName string, identifierLoggingString string) error {
	dbEvents := []TaxableEvent{}

	for _, blockEvent := range blockEvents {

		denom, err := GetDenomForBase(blockEvent.Denomination)
		if err != nil {
			// attempt to add missing denoms to the database
			config.Log.Warnf("Denom lookup failed. Will be inserted as UNKNOWN. Denom Received: %v. Err: %v", blockEvent.Denomination, err)
			denom, err = AddUnknownDenom(db, blockEvent.Denomination)
			if err != nil {
				config.Log.Error(fmt.Sprintf("There was an error adding a missing denom. Denom Received: %v", blockEvent.Denomination), err)
				return err
			}
		}

		// Create unique hash for each event to ensure idempotency
		hash := sha256.New()
		// WARN: The space in the amount/denom hash part is deliberate, it matches an old version of the hash to maintain backwards
		// compatibility with an old version of the indexer and old indexed data
		hashParts := fmt.Sprint(blockEvent.Address, blockHeight, fmt.Sprintf(" %v%s", blockEvent.Amount, blockEvent.Denomination))
		hash.Write([]byte(hashParts))

		evt := TaxableEvent{
			Source:       blockEvent.EventSource,
			Amount:       util.ToNumeric(blockEvent.Amount),
			EventHash:    fmt.Sprintf("%x", hash.Sum(nil)),
			Denomination: denom,
			Block:        Block{Height: blockHeight, TimeStamp: blockTime, Chain: Chain{ChainID: dbChainID, Name: dbChainName}},
			EventAddress: Address{Address: blockEvent.Address},
		}
		dbEvents = append(dbEvents, evt)

	}

	// sort by hash
	sort.SliceStable(dbEvents, func(i, j int) bool {
		return dbEvents[i].EventHash < dbEvents[j].EventHash
	})

	// insert events into DB in batches of batchSize
	batchSize := 10000
	numItems := len(dbEvents)
	currentIter := 1
	numIters := 1

	if batchSize < numItems {
		numIters = (numItems / batchSize) + 1
	}

	for i := 0; i < len(dbEvents); i += batchSize {
		batchEnd := i + batchSize
		if batchEnd > len(dbEvents) {
			batchEnd = len(dbEvents)
		}

		awaitingInsert := dbEvents[i:batchEnd]

		// Only way this can happen is if i == batchEnd
		if len(awaitingInsert) == 0 {
			awaitingInsert = []TaxableEvent{dbEvents[i]}
		}

		if !dryRun {
			config.Log.Infof("Sending %d block events to DB for %s %d/%d", len(awaitingInsert), identifierLoggingString, currentIter, numIters)
			err := createTaxableEvents(db, awaitingInsert)
			if err != nil {
				config.Log.Error("Error storing DB events.", err)
				return err
			}
		}
		currentIter++
	}

	return nil
}

func UpdateEpochIndexingStatus(db *gorm.DB, dryRun bool, epochNumber uint, epochIdentifier string, dbChainID string, dbChainName string) error {
	if !dryRun {
		epochToUpdate := Epoch{
			EpochNumber: epochNumber,
			Chain:       Chain{ChainID: dbChainID, Name: dbChainName},
			Identifier:  epochIdentifier,
		}

		return db.Model(&Epoch{}).Where(&epochToUpdate).Update("indexed", true).Error
	}
	return nil
}

func createTaxableEvents(db *gorm.DB, events []TaxableEvent) error {
	// Ordering matters due to foreign key constraints. Call Create() first to get right foreign key ID
	return db.Transaction(func(dbTransaction *gorm.DB) error {
		if len(events) == 0 {
			return errors.New("no events to insert")
		}

		var chainPrev Chain
		var blockPrev Block

		for _, event := range events {
			if chainPrev.ChainID != event.Block.Chain.ChainID || event.Block.Chain.Name != chainPrev.Name {
				if chainErr := dbTransaction.Where("chain_id = ?", event.Block.Chain.ChainID).FirstOrCreate(&event.Block.Chain).Error; chainErr != nil {
					fmt.Printf("Error %s creating chain DB object.\n", chainErr)
					return chainErr
				}

				chainPrev = event.Block.Chain
			}

			event.Block.Chain = chainPrev

			if blockPrev.Height != event.Block.Height {
				whereCond := Block{Chain: event.Block.Chain, Height: event.Block.Height}

				if blockErr := dbTransaction.Where(whereCond).FirstOrCreate(&event.Block).Error; blockErr != nil {
					fmt.Printf("Error %s creating block DB object.\n", blockErr)
					return blockErr
				}

				blockPrev = event.Block
			}

			event.Block = blockPrev

			if event.EventAddress.Address != "" {
				// viewing gorm logs shows this gets translated into a single ON CONFLICT DO NOTHING RETURNING "id"
				if err := dbTransaction.Where(&event.EventAddress).FirstOrCreate(&event.EventAddress).Error; err != nil {
					fmt.Printf("Error %s creating address for TaxableEvent.\n", err)
					return err
				}
			}

			if event.Denomination.Base == "" || event.Denomination.Symbol == "" {
				return fmt.Errorf("denom not cached for base %s and symbol %s", event.Denomination.Base, event.Denomination.Symbol)
			}

			thisEvent := event // This is redundant but required for the picky gosec linter
			if err := dbTransaction.Where(TaxableEvent{EventHash: event.EventHash}).FirstOrCreate(&thisEvent).Error; err != nil {
				fmt.Printf("Error %s creating tx.\n", err)
				return err
			}
		}

		return nil
	})
}

func GetHighestTaxableEventBlock(db *gorm.DB, chainID string) (Block, error) {
	var block Block

	result := db.Joins("JOIN taxable_event ON blocks.id = taxable_event.block_id").
		Joins("JOIN chains ON blocks.blockchain_id = chains.id AND chains.chain_id = ?", chainID).Order("height desc").First(&block)

	return block, result.Error
}
