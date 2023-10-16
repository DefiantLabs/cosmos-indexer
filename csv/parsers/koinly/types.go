package koinly

import "github.com/DefiantLabs/cosmos-indexer/csv/parsers"

const (
	// ParserKey is the key used to identify this parser
	ParserKey = "koinly"

	// TimeLayout is the golang time format string for this parser
	TimeLayout = "2006-01-02 15:04:05"
)

type Parser struct {
	Rows          []Row
	ParsingGroups []parsers.ParsingGroup
}

type Row struct {
	Date             string
	SentAmount       string
	SentCurrency     string
	ReceivedAmount   string
	ReceivedCurrency string
	FeeAmount        string
	FeeCurrency      string
	NetWorthAmount   string // not going to use this for now?
	NetWorthCurrency string // not going to use this for now?
	Label            Label
	Description      string
	TxHash           string
}

type Label int

const (
	// outgoing transactions
	None Label = iota
	Gift
	Lost
	Cost
	MarginFee
	RealizedGain
	Stake

	// incoming transactions
	Airdrop
	Fork
	Mining
	Reward
	Income
	LoanInterest
	// RealizedGain this is duplicated in their docs
	Unstake

	// Trades
	Swap
	LiquidityIn
	LiquidityOut
)

func (at Label) String() string {
	return [...]string{
		"", "gift", "lost", "cost", "margin fee", "realized gain", "stake",
		"airdrop", "fork", "mining", "reward", "income", "loan interest", "unstake",
		"swap", "liquidity in", "liquidity out",
	}[at]
}
