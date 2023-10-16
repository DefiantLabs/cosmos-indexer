package test

import (
	"os"
	"testing"

	"github.com/DefiantLabs/cosmos-indexer/csv"
	"github.com/DefiantLabs/cosmos-indexer/db"
)

const (
	OsmosisAddressRegex  = "osmo(valoper)?1[a-z0-9]{38}"
	OsmosisAddressPrefix = "osmo"
)

// Example DB query to get TXs for address:
/*
select * from taxable_tx tx
INNER JOIN addresses as addr ON addr.id = tx.sender_address_id OR addr.id = tx.receiver_address_id
where addr.address = 'osmo...'
*/
func TestOsmosisCsvForAddress(t *testing.T) {
	addressRegex := OsmosisAddressRegex
	addressPrefix := OsmosisAddressPrefix
	gorm, _ := dbSetup(addressRegex, addressPrefix)
	address := "osmo14mmus5h7m6vkp0pteks8wawaj4wf3sx7fy3s2r" // local test key address
	csvRows, headers, err := csv.ParseForAddress([]string{address}, nil, nil, gorm, "accointing")
	if err != nil || len(csvRows) == 0 {
		t.Fatal("Failed to lookup taxable events")
	}

	buffer, err := csv.ToCsv(csvRows, headers)
	if err != nil {
		t.Fatal("CSV writing should not result in error", err)
	}

	if len(buffer.Bytes()) == 0 {
		t.Fatal("CSV length should never be 0, there are always headers!")
	}

	err = os.WriteFile("accointing.csv", buffer.Bytes(), 0o600)
	if err != nil {
		t.Fatal("Failed to write CSV to disk")
	}
}

func TestCsvForAddress(t *testing.T) {
	addressRegex := "juno(valoper)?1[a-z0-9]{38}"
	addressPrefix := "juno"
	gorm, _ := dbSetup(addressRegex, addressPrefix)
	// address := "juno1mt72y3jny20456k247tc5gf2dnat76l4ynvqwl"
	// address := "juno130mdu9a0etmeuw52qfxk73pn0ga6gawk4k539x" // strangelove's delegator
	address := "juno1m2hg5t7n8f6kzh8kmh98phenk8a4xp5wyuz34y" // local test key address
	csvRows, headers, err := csv.ParseForAddress([]string{address}, nil, nil, gorm, "accointing")
	if err != nil || len(csvRows) == 0 {
		t.Fatal("Failed to lookup taxable events")
	}

	buffer, err := csv.ToCsv(csvRows, headers)
	if err != nil {
		t.Fatal("CSV writing should not result in error", err)
	}
	if len(buffer.Bytes()) == 0 {
		t.Fatal("CSV length should never be 0, there are always headers!")
	}

	err = os.WriteFile("accointing.csv", buffer.Bytes(), 0o600)
	if err != nil {
		t.Fatal("Failed to write CSV to disk")
	}
}

func TestLookupTxForAddresses(t *testing.T) {
	addressRegex := "juno(valoper)?1[a-z0-9]{38}"
	addressPrefix := "juno"
	gorm, _ := dbSetup(addressRegex, addressPrefix)
	// "juno1txpxafd7q96nkj5jxnt7qnqy4l0rrjyuv6dgte"
	// juno1mt72y3jny20456k247tc5gf2dnat76l4ynvqwl
	taxableEvts, err := db.GetTaxableTransactions("juno1txpxafd7q96nkj5jxnt7qnqy4l0rrjyuv6dgte", gorm)
	if err != nil || len(taxableEvts) == 0 {
		t.Fatal("Failed to lookup taxable events")
	}
}
