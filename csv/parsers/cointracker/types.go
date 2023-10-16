package cointracker

import "github.com/DefiantLabs/cosmos-indexer/csv/parsers"

const (
	// ParserKey is the key used to identify this parser
	ParserKey = "cointracker"

	// TimeLayout is the golang time format string for this parser
	TimeLayout = "01/02/2006 15:04:05"
)

type Parser struct {
	Rows          []Row
	ParsingGroups []parsers.ParsingGroup
}

type Row struct {
	Date             string
	ReceivedAmount   string
	ReceivedCurrency string
	SentAmount       string
	SentCurrency     string
	FeeAmount        string
	FeeCurrency      string
	Tag              Tag
}

type Tag int

const (
	// send transactions
	None Tag = iota
	Gift
	Lost
	Donation

	// receive transactions
	Fork
	Airdrop
	Mined
	Payment
	Staked

	// Trades & Transfers cannot have tags
)

func (at Tag) String() string {
	return [...]string{"", "gift", "lost", "donation", "fork", "airdrop", "mined", "payment", "staked"}[at]
}
