package parsers

import (
	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/gamm"
)

var IsOsmosisJoin = map[string]bool{
	gamm.MsgJoinSwapExternAmountIn: true,
	gamm.MsgJoinSwapShareAmountOut: true,
	gamm.MsgJoinPool:               true,
}

var IsOsmosisExit = map[string]bool{
	gamm.MsgExitSwapShareAmountIn:   true,
	gamm.MsgExitSwapExternAmountOut: true,
	gamm.MsgExitPool:                true,
}

// IsOsmosisLpTxGroup is used as a guard for adding messages to the group.
var IsOsmosisLpTxGroup = make(map[string]bool)

func init() {
	for messageType := range IsOsmosisJoin {
		IsOsmosisLpTxGroup[messageType] = true
	}

	for messageType := range IsOsmosisExit {
		IsOsmosisLpTxGroup[messageType] = true
	}
}

type WrapperLpTxGroup struct {
	GroupedTxes map[uint][]db.TaxableTransaction // TX db ID to its messages
	Rows        []CsvRow
}

func (sf *WrapperLpTxGroup) ParseGroup(parsingFunc func(sf *WrapperLpTxGroup) error) error {
	return parsingFunc(sf)
}

func (sf *WrapperLpTxGroup) GetRowsForParsingGroup() []CsvRow {
	return sf.Rows
}

func (sf *WrapperLpTxGroup) BelongsToGroup(message db.TaxableTransaction) bool {
	_, isInGroup := IsOsmosisLpTxGroup[message.Message.MessageType.MessageType]
	return isInGroup
}

func (sf *WrapperLpTxGroup) String() string {
	return "OsmosisLpTxGroup"
}

func (sf *WrapperLpTxGroup) GetGroupedTxes() map[uint][]db.TaxableTransaction {
	return sf.GroupedTxes
}

func (sf *WrapperLpTxGroup) AddTxToGroup(tx db.TaxableTransaction) {
	// Add tx to group using the TX ID as key and appending to array
	if _, ok := sf.GroupedTxes[tx.Message.Tx.ID]; ok {
		sf.GroupedTxes[tx.Message.Tx.ID] = append(sf.GroupedTxes[tx.Message.Tx.ID], tx)
	} else {
		var txGrouping []db.TaxableTransaction
		txGrouping = append(txGrouping, tx)
		sf.GroupedTxes[tx.Message.Tx.ID] = txGrouping
	}
}

func GetOsmosisTxParsingGroups() []ParsingGroup {
	var messageGroups []ParsingGroup

	// This appending of parsing groups establishes a precedence
	// There is a break statement in the loop doing grouping
	// Which means parsers further up the array will be preferred
	LpTxGroup := WrapperLpTxGroup{}
	LpTxGroup.GroupedTxes = make(map[uint][]db.TaxableTransaction)
	messageGroups = append(messageGroups, &LpTxGroup)

	return messageGroups
}
