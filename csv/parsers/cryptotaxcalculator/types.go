package cryptotaxcalculator

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/csv/parsers"
)

const (
	// ParserKey is the key used to identify this parser
	ParserKey = "cryptotaxcalculator"

	// TimeLayout is the golang time format string for this parser
	TimeLayout = "2006-01-02 15:04:05"
)

type Parser struct {
	Rows          []Row
	ParsingGroups []parsers.ParsingGroup
}

type Row struct {
	Date                   time.Time
	Type                   string
	BaseCurrency           string
	BaseAmount             string
	QuoteCurrency          string
	QuoteAmount            string
	FeeCurrency            string
	FeeAmount              string
	From                   string
	To                     string
	Blockchain             string
	ID                     string
	Description            string
	ReferencePricePerUnit  string
	ReferencePriceCurrency string
}

const (
	AirDrop        = "airdrop"
	Buy            = "buy"
	FlatDeposit    = "flat-deposit"
	FlatWithdrawal = "flat-withdrawal"
	Receive        = "receive"
	Sell           = "sell"
)
