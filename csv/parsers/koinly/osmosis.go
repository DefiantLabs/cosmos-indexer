package koinly

import (
	"fmt"
	"strconv"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/csv/parsers"
	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"

	"github.com/preichenberger/go-coinbasepro/v2"
)

func ParseGroup(sf *parsers.WrapperLpTxGroup) error {
	cbClient := coinbasepro.NewClient()
	for _, txMessages := range sf.GroupedTxes {
		for _, message := range txMessages {
			row := Row{}
			row.TxHash = message.Message.Tx.Hash
			row.Date = message.Message.Tx.Block.TimeStamp.Format(TimeLayout)

			denomRecieved := message.DenominationReceived
			valueRecieved := message.AmountReceived
			conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(valueRecieved), denomRecieved)
			if err != nil {
				row.ReceivedAmount = util.NumericToString(valueRecieved)
				row.ReceivedCurrency = denomRecieved.Base
			} else {
				row.ReceivedAmount = conversionAmount.Text('f', -1)
				row.ReceivedCurrency = conversionSymbol
			}

			denomSent := message.DenominationSent
			valueSent := message.AmountSent
			conversionAmount, conversionSymbol, err = db.ConvertUnits(util.FromNumeric(valueSent), denomSent)
			if err != nil {
				row.SentAmount = util.NumericToString(valueSent)
				row.SentCurrency = denomSent.Base
			} else {
				row.SentAmount = conversionAmount.Text('f', -1)
				row.SentCurrency = conversionSymbol
			}

			// We deliberately exclude the GAMM tokens from OutSell/InBuy for Exits/Joins respectively
			// Accointing has no way of using the GAMM token to determine LP cost basis etc...
			if _, ok := parsers.IsOsmosisExit[message.Message.MessageType.MessageType]; ok {
				row.Label = LiquidityOut
				// add the value of gam tokens
				price, err := parsers.GetRate(cbClient, message.DenominationReceived.Symbol, message.Message.Tx.Block.TimeStamp)
				if err != nil {
					row.Description = fmt.Sprintf("could not lookup value of %v %v. It will be equivalent to %v %v at %v.", row.SentAmount, row.SentCurrency, row.ReceivedAmount, row.ReceivedCurrency, row.Date)
				} else {
					receivedAmount, err := strconv.ParseFloat(row.ReceivedAmount, 64)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Could not parse amount %v", row.ReceivedAmount), err)
					}
					gamValue := receivedAmount * price
					row.Description = fmt.Sprintf("%v %v on %v was $%v USD", row.SentAmount, row.SentCurrency, row.Date, gamValue)
				}
			} else if _, ok := parsers.IsOsmosisJoin[message.Message.MessageType.MessageType]; ok {
				row.Label = LiquidityIn
				// add the value of gam tokens
				price, err := parsers.GetRate(cbClient, message.DenominationSent.Symbol, message.Message.Tx.Block.TimeStamp)
				if err != nil {
					row.Description = fmt.Sprintf("could not lookup value of %v %v. It will be equivalent to %v %v at %v.", row.ReceivedAmount, row.ReceivedCurrency, row.SentAmount, row.SentCurrency, row.Date)
				} else {
					sentAmount, err := strconv.ParseFloat(row.SentAmount, 64)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Could not parse amount %v", row.SentAmount), err)
					}
					gamValue := sentAmount * price
					row.Description = fmt.Sprintf("%v %v on %v was $%v USD", row.ReceivedAmount, row.ReceivedCurrency, row.Date, gamValue)
				}
			}
			sf.Rows = append(sf.Rows, row)
		}
	}
	return nil
}
