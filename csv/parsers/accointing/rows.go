package accointing

import (
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"
)

func (row Row) GetRowForCsv() []string {
	return []string{
		row.TransactionType.String(),
		row.Date,
		row.InBuyAmount,
		row.InBuyAsset,
		row.OutSellAmount,
		row.OutSellAsset,
		row.FeeAmount,
		row.FeeAsset,
		row.Classification.String(),
		row.OperationID,
		row.Comments,
	}
}

func (row Row) GetDate() string {
	return row.Date
}

func (row *Row) EventParseBasic(event db.TaxableEvent) error {
	conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.Amount), event.Denomination)
	if err == nil {
		row.InBuyAmount = conversionAmount.Text('f', -1)
		row.InBuyAsset = conversionSymbol
	} else {
		row.InBuyAmount = util.NumericToString(event.Amount)
		row.InBuyAsset = event.Denomination.Base
	}
	row.TransactionType = Deposit
	row.Date = event.Block.TimeStamp.Format(TimeLayout)
	row.Classification = LiquidityPool

	return nil
}

// ParseBasic: Handles the fields that are shared between most types.
func (row *Row) ParseBasic(address string, event db.TaxableTransaction) error {
	row.Date = event.Message.Tx.Block.TimeStamp.Format(TimeLayout)
	row.OperationID = event.Message.Tx.Hash

	// deposit
	if event.ReceiverAddress.Address == address {
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: deposit)", row.OperationID)
		}
		row.InBuyAmount = conversionAmount.Text('f', -1)
		row.InBuyAsset = conversionSymbol
		row.TransactionType = Deposit
	} else if event.SenderAddress.Address == address { // withdrawal
		conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
		if err != nil {
			return fmt.Errorf("cannot parse denom units for TX %s (classification: withdrawal)", row.OperationID)
		}
		row.OutSellAmount = conversionAmount.Text('f', -1)
		row.OutSellAsset = conversionSymbol
		row.TransactionType = Withdraw
	}

	return nil
}

func (row *Row) ParseSwap(event db.TaxableTransaction) error {
	row.Date = event.Message.Tx.Block.TimeStamp.Format(TimeLayout)
	row.OperationID = event.Message.Tx.Hash
	row.TransactionType = Order

	recievedConversionAmount, recievedConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountReceived), event.DenominationReceived)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap received)", row.OperationID)
	}

	row.InBuyAmount = recievedConversionAmount.Text('f', -1)
	row.InBuyAsset = recievedConversionSymbol

	sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(event.AmountSent), event.DenominationSent)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.OperationID)
	}

	row.OutSellAmount = sentConversionAmount.Text('f', -1)
	row.OutSellAsset = sentConversionSymbol

	return nil
}

func (row *Row) ParseFee(tx db.Tx, fee db.Fee) error {
	row.Date = tx.Block.TimeStamp.Format(TimeLayout)
	row.OperationID = tx.Hash
	row.TransactionType = Withdraw
	row.Classification = Fee

	sentConversionAmount, sentConversionSymbol, err := db.ConvertUnits(util.FromNumeric(fee.Amount), fee.Denomination)
	if err != nil {
		return fmt.Errorf("cannot parse denom units for TX %s (classification: swap sent)", row.OperationID)
	}

	row.OutSellAmount = sentConversionAmount.Text('f', -1)
	row.OutSellAsset = sentConversionSymbol

	return nil
}
