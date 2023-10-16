package cryptotaxcalculator

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"
)

func (row Row) GetRowForCsv() []string {
	return []string{
		row.Date.Format(TimeLayout),
		row.Type,
		row.BaseCurrency,
		row.BaseAmount,
		row.QuoteCurrency,
		row.QuoteAmount,
		row.FeeCurrency,
		row.FeeAmount,
		row.From,
		row.To,
		row.Blockchain,
		row.ID,
		row.Description,
		row.ReferencePricePerUnit,
		row.ReferencePriceCurrency,
	}
}

func (row Row) GetDate() string {
	return row.Date.Format(TimeLayout)
}

func (row *Row) EventParseBasic(event db.TaxableEvent) error {
	row.Date = event.Block.TimeStamp

	conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.Amount), event.Denomination)
	if err == nil {
		row.BaseAmount = conversionAmount.Text('f', -1)
		row.BaseCurrency = conversionSymbol
	} else {
		row.BaseAmount = util.NumericToString(event.Amount)
		row.BaseCurrency = event.Denomination.Base
	}

	return nil
}

// ParseBasic: Handles the fields that are shared between most types.
func (row *Row) ParseBasic(address string, event db.TaxableTransaction) error {
	row.Date = event.Message.Tx.Block.TimeStamp
	row.ID = event.Message.Tx.Hash

	// deposit
	if event.ReceiverAddress.Address == address {
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: deposit)", row.ID)
		}
		row.BaseAmount = conversionAmount.Text('f', -1)
		row.BaseCurrency = conversionSymbol
		row.Type = FlatDeposit
	} else if event.SenderAddress.Address == address { // withdrawal
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: withdrawal)", row.ID)
		}
		row.BaseAmount = conversionAmount.Text('f', -1)
		row.BaseCurrency = conversionSymbol
		row.Type = FlatWithdrawal
	}

	row.From = event.SenderAddress.Address
	row.To = event.ReceiverAddress.Address
	for _, fee := range event.Message.Tx.Fees {
		if fee.PayerAddress.Address == address {
			sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(fee.Amount), fee.Denomination)
			if err != nil {
				return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.ID)
			}

			row.FeeAmount = sentConversionAmount.Text('f', -1)
			row.FeeCurrency = sentConversionSymbol
		}
	}

	return nil
}

func (row *Row) ParseSwap(event db.TaxableTransaction, address, eventType string) error {
	row.Date = event.Message.Tx.Block.TimeStamp
	row.ID = event.Message.Tx.Hash
	row.Type = eventType

	recievedConversionAmount, recievedConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap received)", row.ID)
	}

	row.BaseAmount = recievedConversionAmount.Text('f', -1)
	row.BaseCurrency = recievedConversionSymbol

	sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.ID)
	}

	row.QuoteAmount = sentConversionAmount.Text('f', -1)
	row.QuoteCurrency = sentConversionSymbol

	for _, fee := range event.Message.Tx.Fees {
		if fee.PayerAddress.Address == address {
			sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(fee.Amount), fee.Denomination)
			if err != nil {
				return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.ID)
			}

			row.FeeAmount = sentConversionAmount.Text('f', -1)
			row.FeeCurrency = sentConversionSymbol
		}
	}

	return nil
}
