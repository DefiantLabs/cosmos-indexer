package parsers

import "github.com/DefiantLabs/cosmos-indexer/db"

// Parsers should be used to check in your parsers.
var Parsers map[string]bool

func init() {
	Parsers = make(map[string]bool)
}

func RegisterParsers(keys []string) {
	for _, key := range keys {
		Parsers[key] = true
	}
}

func GetParserKeys() []string {
	var parserKeys []string

	for i := range Parsers {
		parserKeys = append(parserKeys, i)
	}

	return parserKeys
}

// MakeTXMap will make a map of transaction ID to list of taxable transactions
func MakeTXMap(taxableTXs []db.TaxableTransaction) map[uint][]db.TaxableTransaction {
	// process taxableTx into Rows
	txMap := map[uint][]db.TaxableTransaction{} // Map transaction ID to List of taxable transactions

	// Build a map, so we know which TX go with which messages
	for _, taxableTx := range taxableTXs {
		if list, ok := txMap[taxableTx.Message.Tx.ID]; ok {
			list = append(list, taxableTx)
			txMap[taxableTx.Message.Tx.ID] = list
		} else {
			txMap[taxableTx.Message.Tx.ID] = []db.TaxableTransaction{taxableTx}
		}
	}
	return txMap
}

// SeparateParsingGroups will pull messages out of the txMap that need to be grouped together for tax purposes.
func SeparateParsingGroups(txMap map[uint][]db.TaxableTransaction, parsingGroups []ParsingGroup) {
	// The basic idea is we want to do the following:
	// 1. Loop through each message for each transaction
	// 2. Check if it belongs in a group by message type
	// 3. If so, add them to that group
	// 4. If not, keep them on that tx
	// We update the tx msgs to the new list to ensure that the message will not be parsed twice
	for txIdx, txMsgs := range txMap {
		var remainingTxMsgs []db.TaxableTransaction
		// Loop through the transactions
		for _, message := range txMsgs {
			// if the msg in this tx belongs to the group
			var txInGroup bool
			for _, txGroup := range parsingGroups {
				if txGroup.BelongsToGroup(message) {
					// add to the group list
					txGroup.AddTxToGroup(message)
					txInGroup = true
				}
			}
			if !txInGroup {
				remainingTxMsgs = append(remainingTxMsgs, message)
			}
		}
		txMap[txIdx] = remainingTxMsgs
	}
}
