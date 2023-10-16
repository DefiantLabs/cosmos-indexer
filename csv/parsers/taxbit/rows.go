package taxbit

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"
)

func (row Row) GetRowForCsv() []string {
	// Add source and dest as needed
	if row.SentCurrency != "" {
		row.SendingSource = fmt.Sprintf("%s Wallet", row.SentCurrency)
	}
	if row.ReceivedCurrency != "" {
		row.ReceivingDestination = fmt.Sprintf("%s Wallet", row.ReceivedCurrency)
	}
	return []string{
		row.Date,
		row.TransactionType.String(),
		row.SentAmount,
		row.SentCurrency,
		row.SendingSource,
		row.ReceivedAmount,
		row.ReceivedCurrency,
		row.ReceivingDestination,
		row.FeeAmount,
		row.FeeCurrency,
		row.ExchangeTransactionID,
		row.TxHash,
	}
}

func (row Row) GetDate() string {
	return row.Date
}

// EventParseBasic handles the deposit os osmos rewards
func (row *Row) EventParseBasic(event db.TaxableEvent) error {
	row.Date = event.Block.TimeStamp.Format(TimeLayout)

	conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.Amount), event.Denomination)
	if err == nil {
		row.ReceivedAmount = conversionAmount.Text('f', -1)
		row.ReceivedCurrency = conversionSymbol
	} else {
		row.ReceivedAmount = util.NumericToString(event.Amount)
		row.ReceivedCurrency = event.Denomination.Base
	}
	// row.Label = Reward
	return nil
}

// ParseBasic: Handles the fields that are shared between most types.
func (row *Row) ParseBasic(address string, event db.TaxableTransaction) error {
	row.Date = event.Message.Tx.Block.TimeStamp.Format(TimeLayout)
	row.TxHash = event.Message.Tx.Hash

	// deposit
	if event.ReceiverAddress.Address == address {
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: deposit)", row.TxHash)
		}
		row.ReceivedAmount = conversionAmount.Text('f', -1)
		row.ReceivedCurrency = conversionSymbol
		row.TransactionType = Sale
	} else if event.SenderAddress.Address == address { // withdrawal
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: withdrawal)", row.TxHash)
		}
		row.SentAmount = conversionAmount.Text('f', -1)
		row.SentCurrency = conversionSymbol
		row.TransactionType = Buy
	}

	// Once we support indexing across multiple chains, we can look if the transaction is from one of the user's
	// wallets to another one of their wallets, if this is the case this is a "Transfer" "In" or "Out".

	return nil
}

func (row *Row) ParseSwap(event db.TaxableTransaction) error {
	row.Date = event.Message.Tx.Block.TimeStamp.Format(TimeLayout)
	row.TxHash = event.Message.Tx.Hash
	row.TransactionType = Trade

	recievedConversionAmount, recievedConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap received)", row.TxHash)
	}

	row.ReceivedAmount = recievedConversionAmount.Text('f', -1)
	row.ReceivedCurrency = recievedConversionSymbol

	sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.TxHash)
	}

	row.SentAmount = sentConversionAmount.Text('f', -1)
	row.SentCurrency = sentConversionSymbol

	return nil
}

func (row *Row) ParseFee(tx db.Tx, fee db.Fee) error {
	row.Date = tx.Block.TimeStamp.Format(TimeLayout)
	row.TxHash = tx.Hash
	row.TransactionType = Expense

	sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(fee.Amount), fee.Denomination)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.TxHash)
	}

	row.SentAmount = sentConversionAmount.Text('f', -1)
	row.SentCurrency = sentConversionSymbol

	return nil
}
