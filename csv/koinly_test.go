// nolint:unused
package csv

import (
	"testing"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/csv/parsers/koinly"
	"github.com/DefiantLabs/cosmos-indexer/osmosis"
	"github.com/stretchr/testify/assert"
)

func TestKoinlyOsmoLPParsing(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(koinly.ParserKey)
	parser.InitializeParsingGroups()

	// setup user and chain
	targetAddress := mkAddress(t, 1)
	chain := mkChain(1, osmosis.ChainID, osmosis.Name)

	// make transactions for this user entering and leaving LPs
	transferTxs := getTestSwapTXs(t, targetAddress, chain)

	// attempt to parse
	err := parser.ProcessTaxableTx(targetAddress.Address, transferTxs)
	assert.Nil(t, err, "should not get error from parsing these transactions")

	// validate output
	rows, err := parser.GetRows(targetAddress.Address, nil, nil)
	assert.Nil(t, err, "should not get error from getting rows")
	assert.Equalf(t, len(transferTxs), len(rows), "you should have one row for each transfer transaction ")

	// all transactions should be orders classified as liquidity_pool
	for i, row := range rows {
		cols := row.GetRowForCsv()
		// first column should parse as a time and not be zero-time
		time, err := time.Parse(koinly.TimeLayout, cols[0])
		assert.Nil(t, err, "time should parse properly")
		assert.NotEqual(t, time, zeroTime)

		// make sure transactions are properly labeled
		if i < 4 {
			assert.Equal(t, cols[9], koinly.LiquidityIn.String())
		} else {
			assert.Equal(t, cols[9], koinly.LiquidityOut.String())
		}
	}
}

func TestKoinlyOsmoRewardParsing(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(koinly.ParserKey)
	parser.InitializeParsingGroups()

	// setup user and chain
	targetAddress := mkAddress(t, 1)
	chain := mkChain(1, osmosis.ChainID, osmosis.Name)

	// make transactions for this user entering and leaving LPs
	taxableEvents := getTestTaxableEvents(t, targetAddress, chain)

	// attempt to parse
	err := parser.ProcessTaxableEvent(taxableEvents)
	assert.Nil(t, err, "should not get error from parsing these transactions")

	// validate output
	rows, err := parser.GetRows(targetAddress.Address, nil, nil)
	assert.Nil(t, err, "should not get error from getting rows")
	assert.Equalf(t, len(taxableEvents), len(rows), "you should have one row for each transfer transaction ")

	// all transactions should be orders classified as liquidity_pool
	for _, row := range rows {
		cols := row.GetRowForCsv()
		// first column should parse as a time
		_, err := time.Parse(koinly.TimeLayout, cols[0])
		assert.Nil(t, err, "time should parse properly")

		// make sure transactions are properly labeled
		assert.Equal(t, cols[9], "reward")
	}
}
