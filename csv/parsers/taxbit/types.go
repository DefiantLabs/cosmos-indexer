package taxbit

import "github.com/DefiantLabs/cosmos-indexer/csv/parsers"

const (
	// ParserKey is the key used to identify this parser
	ParserKey = "taxbit"

	// TimeLayout is the golang time format string for this parser
	TimeLayout = "2006-01-02T15:04:05Z"
)

type Parser struct {
	Rows          []Row
	ParsingGroups []parsers.ParsingGroup
}

type Row struct {
	Date                  string
	TransactionType       TransactionType
	SentAmount            string
	SentCurrency          string
	SendingSource         string
	ReceivedAmount        string
	ReceivedCurrency      string
	ReceivingDestination  string
	FeeAmount             string
	FeeCurrency           string
	ExchangeTransactionID string
	TxHash                string
}

type TransactionType int

const (
	None TransactionType = iota

	Buy
	// Example: Purchase of crypto with fiat (i.e; Paid $5,000 for 1 BTC)
	// When recording buy transactions the following is required: Date and Time, Transaction Type, Sent Quantity, Sent Currency, Received Quantity, and Received Currency.

	Sale
	// Example: Sale of crypto for fiat (i.e; Sold 1 BTC for $5,000)
	// When recording sale transactions the following is required: Date and Time, Transaction Type, Sent Quantity, Sent Currency, Received Quantity, and Received Currency.

	Trade
	// Example: Trading one crypto for another crypto (i.e; Trade 1 BTC for 10 ETH)
	// When recording trade transactions the following is required: Date and Time, Transaction Type, Sent Quantity, Sent Currency, Received Quantity, and Received Currency.

	TransfersIn
	TransfersOut
	// Example: Transfer crypto to an exchange or wallet in your possession (i.e; Moved 1 BTC from Coinbase to a hardware wallet)
	// Transfer transactions can be recorded two different ways in the transaction type column:
	// Transfer in - when you are only recording the receiving of a asset (for example, a deposit on an exchange)
	// Transfer out - when you are only recording the sending of an asset (for example, a withdrawal on an exchange)
	// When recording transfer in transactions, the following is required: Date and Time, Transaction Type, Received Quantity, and Received Currency.
	// When recording transfer out transactions, the following is required: Date and Time, Transaction Type, Sent Quantity, and Sent Currency.

	Income
	// Example: Received crypto in return for goods or services (i.e; Received 1 BTC in return for mining)
	// When recording income transactions the following columns are required: Date and Time, Transaction Type, Received Quantity, and Received Currency.

	Expense
	// Example: Sent crypto in return for goods or services (i.e; Sent 1 BTC as payment for a car)
	// When recording Expense transactions the following is required: Date and Time, Transaction Type, Sent Quantity, and Sent Currency.

	Gifts
	// Example: Receiving .1 ETH from a family member from their wallet to yours.
	// When recording gift transactions the following is required: Date and Time, Transaction Type, Sent Quantity, Sent Currency, and Sending Source or Received Quantity, Received Currency, and Receiving Destination.
)

func (at TransactionType) String() string {
	return [...]string{"", "Buy", "Sale", "Trade", "Transfer In", "Transfer Out", "Income", "Expense", "Gift"}[at]
}
