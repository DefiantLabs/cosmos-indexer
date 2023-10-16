package parsers

import (
	"time"

	"github.com/DefiantLabs/cosmos-indexer/db"
)

type Parser interface {
	InitializeParsingGroups()
	ProcessTaxableTx(address string, taxableTxs []db.TaxableTransaction) error
	ProcessTaxableEvent(taxableEvents []db.TaxableEvent) error
	GetHeaders() []string
	GetRows(address string, startDate, endDate *time.Time) ([]CsvRow, error)
	TimeLayout() string
}

type ParsingGroup interface {
	BelongsToGroup(db.TaxableTransaction) bool
	String() string
	AddTxToGroup(db.TaxableTransaction)
	GetGroupedTxes() map[uint][]db.TaxableTransaction
	ParseGroup(parsingFunc func(sf *WrapperLpTxGroup) error) error
	GetRowsForParsingGroup() []CsvRow
}

type CsvRow interface {
	GetRowForCsv() []string
	GetDate() string
}
