package accointing

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
			row.TransactionType = Order
			row.OperationID = message.Message.Tx.Hash
			row.Date = message.Message.Tx.Block.TimeStamp.Format(TimeLayout)

			denomRecieved := message.DenominationReceived
			valueRecieved := message.AmountReceived
			conversionAmount, conversionSymbol, err := db.ConvertUnits(util.FromNumeric(valueRecieved), denomRecieved)
			if err != nil {
				row.InBuyAmount = util.NumericToString(valueRecieved)
				row.InBuyAsset = denomRecieved.Base
			} else {
				row.InBuyAmount = conversionAmount.Text('f', -1)
				row.InBuyAsset = conversionSymbol
			}

			denomSent := message.DenominationSent
			valueSent := message.AmountSent
			conversionAmount, conversionSymbol, err = db.ConvertUnits(util.FromNumeric(valueSent), denomSent)
			if err != nil {
				row.OutSellAmount = util.NumericToString(valueSent)
				row.OutSellAsset = denomSent.Base
			} else {
				row.OutSellAmount = conversionAmount.Text('f', -1)
				row.OutSellAsset = conversionSymbol
			}

			// We deliberately exclude the GAMM tokens from OutSell/InBuy for Exits/Joins respectively
			// Accointing has no way of using the GAMM token to determine LP cost basis etc...
			if _, ok := parsers.IsOsmosisExit[message.Message.MessageType.MessageType]; ok {
				// add the value of gam tokens
				price, err := parsers.GetRate(cbClient, message.DenominationReceived.Symbol, message.Message.Tx.Block.TimeStamp)
				if err != nil {
					row.Comments = fmt.Sprintf("could not lookup value of %v %v. It will be equivalent to %v %v at %v.", row.OutSellAmount, row.OutSellAsset, row.InBuyAmount, row.InBuyAsset, row.Date)
				} else {
					receivedAmount, err := strconv.ParseFloat(row.InBuyAmount, 64)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Could not parse amount %v", row.InBuyAmount), err)
					}
					gamValue := receivedAmount * price
					row.Comments = fmt.Sprintf("%v %v on %v was $%v USD", row.OutSellAmount, row.OutSellAsset, row.Date, gamValue)
				}
			} else if _, ok := parsers.IsOsmosisJoin[message.Message.MessageType.MessageType]; ok {
				// add the value of gam tokens
				price, err := parsers.GetRate(cbClient, message.DenominationSent.Symbol, message.Message.Tx.Block.TimeStamp)
				if err != nil {
					row.Comments = fmt.Sprintf("could not lookup value of %v %v. It will be equivalent to %v %v at %v.", row.InBuyAmount, row.InBuyAsset, row.OutSellAmount, row.OutSellAsset, row.Date)
				} else {
					sentAmount, err := strconv.ParseFloat(row.OutSellAmount, 64)
					if err != nil {
						config.Log.Fatal(fmt.Sprintf("Could not parse amount %v", row.OutSellAmount), err)
					}
					gamValue := sentAmount * price
					row.Comments = fmt.Sprintf("%v %v on %v was $%v USD", row.InBuyAmount, row.InBuyAsset, row.Date, gamValue)
				}
			}
			sf.Rows = append(sf.Rows, row)
		}
	}
	return nil
}
