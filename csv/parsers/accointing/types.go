package accointing

import "github.com/DefiantLabs/cosmos-indexer/csv/parsers"

const (
	// ParserKey is the key used to identify this parser
	ParserKey = "accointing"

	// timeLayout is the golang time format string for this parser
	TimeLayout = "01/02/2006 15:04:05"
)

type Parser struct {
	Rows          []Row
	ParsingGroups []parsers.ParsingGroup
}

type Row struct {
	Date            string
	InBuyAmount     string
	InBuyAsset      string
	OutSellAmount   string
	OutSellAsset    string
	FeeAmount       string
	FeeAsset        string
	Classification  Classification
	TransactionType Transaction
	OperationID     string
	Comments        string
}

type Transaction int

const (
	Deposit Transaction = iota
	Withdraw
	Order
)

func (at Transaction) String() string {
	return [...]string{"deposit", "withdraw", "order"}[at]
}

type Classification int

const (
	None Classification = iota
	Staked
	Airdrop
	Payment
	Fee
	LiquidityPool
	RemoveFunds // Used for GAMM module exits, is this correct?
	Ignored
)

func (ac Classification) String() string {
	// Note that "None" returns empty string since we're using this for CSV parsing.
	// Accointing considers 'Classification' an optional field, so empty is a valid value.
	return [...]string{"", "staked", "airdrop", "payment", "fee", "liquidity_pool", "remove_funds", "ignored"}[ac]
}
