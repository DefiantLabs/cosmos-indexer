package test

import (
	"math/big"
	"strings"
	"testing"

	"github.com/DefiantLabs/cosmos-indexer/osmosis"

	dbUtils "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"github.com/stretchr/testify/assert"
)

func TestGetRewardsForAddress(t *testing.T) {
	// Working SQL query shown below. SQL query included so that test results can be verified through the DB.
	// The SQL query shown selects the top earnings from the DB, but can easily be adjusted to only search a certain block height.
	// To do so you'd simply add another INNER JOIN following the syntax below, and search for a given block height.

	/*
		SELECT a.id, a.amount, a.address_id, c.address, d.height, sd.denom
		FROM taxable_event a
		INNER JOIN (
			SELECT id, block_id, MAX(amount) amount
			FROM taxable_event
			GROUP BY id
			ORDER BY amount desc
		) b ON a.id = b.id AND a.amount = b.amount
		INNER JOIN addresses as c ON c.id = a.address_id
		INNER JOIN blocks as d ON d.id = a.block_id
		INNER JOIN simple_denoms as sd ON sd.id = a.denomination_id
		order by a.amount desc limit 5
	*/

	addressRegex := OsmosisAddressRegex
	addressPrefix := OsmosisAddressPrefix
	gorm, err := dbSetup(addressRegex, addressPrefix)
	if err != nil {
		t.Fail()
	}

	addr := "osmo1g9tdk9kcmptq033t7tp2yglpfx5kztuc279hzq"
	taxableEvts, err := dbUtils.GetTaxableEvents(addr, gorm)
	if err != nil {
		t.Fail()
	}
	foundBlockEvent := false

	for _, evt := range taxableEvts {
		assert.Equal(t, evt.EventAddress.Address, addr)
		assert.Greater(t, evt.Amount, 0.0)
		assert.Greater(t, evt.Block.Height, int64(0))
		assert.Contains(t, strings.ToLower(evt.Block.Chain.Name), osmosis.Name)
		assert.Contains(t, strings.ToLower(evt.Block.Chain.ChainID), osmosis.ChainID)

		if evt.Block.Height == 4823317 && evt.EventAddress.Address == addr && util.FromNumeric(evt.Amount).Cmp(big.NewInt(3632580308)) == 0 {
			foundBlockEvent = true
		}
	}

	// We know the above address earned this much
	assert.Equal(t, foundBlockEvent, true)
}

func TestGetOsmosisRewardIndex(t *testing.T) {
	addressRegex := OsmosisAddressRegex
	addressPrefix := OsmosisAddressPrefix
	gorm, err := dbSetup(addressRegex, addressPrefix)
	if err != nil {
		t.Fail()
	}

	setupOsmosisTestModels(gorm)
	createOsmosisTaxableEvent(gorm, 100)

	block, err := dbUtils.GetHighestTaxableEventBlock(gorm, osmosis.ChainID)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, block.Height, int64(100))
}

func TestInsertOsmosisRewards(t *testing.T) {
	addressRegex := OsmosisAddressRegex
	addressPrefix := OsmosisAddressPrefix
	gorm, err := dbSetup(addressRegex, addressPrefix)
	if err != nil {
		t.Fail()
	}

	setupOsmosisTestModels(gorm)
	createOsmosisTaxableEvent(gorm, 1111111111)

	block, err := dbUtils.GetHighestTaxableEventBlock(gorm, osmosis.ChainID)
	if err != nil {
		t.Fail()
	}

	assert.Equal(t, block.Height, int64(100))
}
