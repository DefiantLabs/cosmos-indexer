// nolint:unused
package csv

import (
	"strings"
	"testing"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/ibc"
	"github.com/DefiantLabs/cosmos-indexer/csv/parsers/accointing"
	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/osmosis"

	"github.com/stretchr/testify/assert"
)

func TestAccointingIbcMsgTransferSelf(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(accointing.ParserKey)
	parser.InitializeParsingGroups()

	me := "osmo18zljeu4lg4jppkz75en82qr3zymfcnchwvsqgu"
	alsoMe := "juno18zljeu4lg4jppkz75en82qr3zymfcnchs9qtej"

	// setup user and chain
	sourceAddress := db.Address{
		ID:      0,
		Address: me,
	}
	destAddress := db.Address{
		ID:      1,
		Address: alsoMe,
	}

	sourceChain := mkChain(1, osmosis.ChainID, osmosis.Name)
	targetChain := mkChain(2, "juno-1", "juno")

	// make transactions for this user entering and leaving LPs
	transferTxs := getTestIbcTransferTXs(t, sourceAddress, destAddress, sourceChain, targetChain)

	// attempt to parse
	err := parser.ProcessTaxableTx(sourceAddress.Address, transferTxs)
	assert.Nil(t, err, "should not get error from parsing these transactions")

	// validate output
	rows, err := parser.GetRows(sourceAddress.Address, nil, nil)
	assert.Nil(t, err, "should not get error from getting rows")
	assert.Equalf(t, len(transferTxs), len(rows), "you should have one row for each transfer transaction ")
	assert.Equal(t, transferTxs[0].Message.MessageType.MessageType, ibc.MsgTransfer, "message type should be an IBC transfer")

	// all transactions should be orders classified as MsgTransfer
	for _, row := range rows {
		cols := row.GetRowForCsv()
		// transfer message should be a 'withdraw' from the sender's perspective
		assert.Equal(t, "withdraw", cols[0], "transaction type should be a withdrawal")
		assert.Equal(t, "ignored", cols[8], "transaction should not have a classification")
	}
}

// Test behavior of an IBC message transfer to someone else's address (e.g. NOT a self transfer)
func TestAccointingIbcMsgTransferExternal(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(accointing.ParserKey)
	parser.InitializeParsingGroups()

	me := "osmo14mmus5h7m6vkp0pteks8wawaj4wf3sx7fy3s2r"
	someoneElse := "juno18zljeu4lg4jppkz75en82qr3zymfcnchs9qtej"

	// setup user and chain
	sourceAddress := db.Address{
		ID:      0,
		Address: me,
	}
	destAddress := db.Address{
		ID:      1,
		Address: someoneElse,
	}

	sourceChain := mkChain(1, osmosis.ChainID, osmosis.Name)
	targetChain := mkChain(2, "juno-1", "juno")

	// make transactions for this user entering and leaving LPs
	transferTxs := getTestIbcTransferTXs(t, sourceAddress, destAddress, sourceChain, targetChain)

	// attempt to parse
	err := parser.ProcessTaxableTx(sourceAddress.Address, transferTxs)
	assert.Nil(t, err, "should not get error from parsing these transactions")

	// validate output
	rows, err := parser.GetRows(sourceAddress.Address, nil, nil)
	assert.Nil(t, err, "should not get error from getting rows")
	assert.Equalf(t, len(transferTxs), len(rows), "you should have one row for each transfer transaction ")
	assert.Equal(t, transferTxs[0].Message.MessageType.MessageType, ibc.MsgTransfer, "message type should be an IBC transfer")

	// all transactions should be orders classified as MsgTransfer
	for _, row := range rows {
		cols := row.GetRowForCsv()
		// transfer message should be a 'withdraw' from the sender's perspective
		assert.Equal(t, cols[0], "withdraw", "transaction type should be a withdrawal")
		assert.Equal(t, cols[8], "", "transaction should not have a classification")
	}
}

func TestAccointingOsmoLPParsing(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(accointing.ParserKey)
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
	for _, row := range rows {
		cols := row.GetRowForCsv()
		// assert on gamms being present
		assert.Equal(t, cols[0], "order", "transaction type should be an order")
		assert.Equal(t, cols[8], "", "transaction should not have a classification")
		// should either contain gamm value or a message about how to find it
		if !strings.Contains(cols[10], "USD") && !strings.Contains(cols[10], "") {
			t.Log("comment should say value of gamm")
			t.Fail()
		}
	}
}

func TestAccointingOsmoRewardParsing(t *testing.T) {
	cfg := config.IndexConfig{}
	cfg.Lens.ChainID = osmosis.ChainID
	parser := GetParser(accointing.ParserKey)
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
		// assert on gamms being present
		assert.Equal(t, cols[0], "deposit", "transaction type should be a deposit")
		assert.Equal(t, cols[8], "liquidity_pool", "transaction should be classified as liquidity_pool")
	}
}
