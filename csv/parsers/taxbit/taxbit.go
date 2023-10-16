package taxbit

import (
	"sort"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/bank"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/distribution"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/gov"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/ibc"
	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/staking"
	"github.com/DefiantLabs/cosmos-indexer/csv/parsers"
	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/gamm"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/poolmanager"
)

func (p *Parser) TimeLayout() string {
	return TimeLayout
}

func (p *Parser) ProcessTaxableTx(address string, taxableTxs []db.TaxableTransaction) error {
	// Build a map, so we know which TX go with which messages
	txMap := parsers.MakeTXMap(taxableTxs)

	// Pull messages out of txMap that must be grouped together
	parsers.SeparateParsingGroups(txMap, p.ParsingGroups)

	// Parse all the potentially taxable events (one transaction group at a time)
	for _, txGroup := range txMap {
		// All messages have been removed into a parsing group
		if len(txGroup) != 0 {
			// For the current transaction group, generate the rows for the CSV.
			// Usually (but not always) a transaction will only have a single row in the CSV.
			txRows, err := ParseTx(address, txGroup)
			if err != nil {
				return err
			}
			for _, v := range txRows {
				p.Rows = append(p.Rows, v.(Row))
			}
		}
	}

	// Parse all the TXs found in the Parsing Groups
	for _, txParsingGroup := range p.ParsingGroups {
		err := txParsingGroup.ParseGroup(ParseGroup)
		if err != nil {
			return err
		}
	}

	// Handle fees on all taxableTxs at once, we don't do this in the regular parser or in the parsing groups
	// This requires HandleFees to process the fees into unique mappings of tx -> fees (since we gather Taxable Messages in the taxableTxs)
	// If we move it into the ParseTx function or into the ParseGroup function, we may be able to reduce the logic in the HandleFees func
	feeRows, err := HandleFees(address, taxableTxs)
	if err != nil {
		return err
	}

	p.Rows = append(p.Rows, feeRows...)

	return nil
}

func (p *Parser) ProcessTaxableEvent(taxableEvents []db.TaxableEvent) error {
	// Parse all the potentially taxable events
	for _, event := range taxableEvents {
		// generate the rows for the CSV.
		rows, err := ParseEvent(event)
		if err != nil {
			return err
		}
		p.Rows = append(p.Rows, rows...)
	}

	return nil
}

func (p *Parser) InitializeParsingGroups() {
	p.ParsingGroups = append(p.ParsingGroups, parsers.GetOsmosisTxParsingGroups()...)
}

func (p *Parser) GetRows(address string, startDate, endDate *time.Time) ([]parsers.CsvRow, error) {
	// Combine all normal rows and parser group rows into 1
	taxbitRows := p.Rows // contains TX rows and fees as well as taxable events
	for _, v := range p.ParsingGroups {
		for _, row := range v.GetRowsForParsingGroup() {
			taxbitRows = append(taxbitRows, row.(Row))
		}
	}

	// Sort by date
	sort.Slice(taxbitRows, func(i int, j int) bool {
		leftDate, err := time.Parse(TimeLayout, taxbitRows[i].Date)
		if err != nil {
			config.Log.Error("Error sorting left date.", err)
			return false
		}
		rightDate, err := time.Parse(TimeLayout, taxbitRows[j].Date)
		if err != nil {
			config.Log.Error("Error sorting right date.", err)
			return false
		}
		return leftDate.Before(rightDate)
	})

	// Now that we are sorted, if we have a start date, drop everything from before it, if end date is set, drop everything after it
	var firstToKeep *int
	var lastToKeep *int
	for i := range taxbitRows {
		if startDate != nil && firstToKeep == nil {
			rowDate, err := time.Parse(TimeLayout, taxbitRows[i].Date)
			if err != nil {
				config.Log.Error("Error parsing row date.", err)
				return nil, err
			}
			if rowDate.Before(*startDate) {
				continue
			}
			startIdx := i
			firstToKeep = &startIdx
		} else if endDate != nil && lastToKeep == nil {
			rowDate, err := time.Parse(TimeLayout, taxbitRows[i].Date)
			if err != nil {
				config.Log.Error("Error parsing row date.", err)
				return nil, err
			}
			if rowDate.Before(*endDate) {
				continue
			} else if i > 0 {
				endIdx := i - 1
				lastToKeep = &endIdx
				break
			}
		}
	}
	if firstToKeep != nil && lastToKeep != nil { // nolint:gocritic
		taxbitRows = taxbitRows[*firstToKeep:*lastToKeep]
	} else if firstToKeep != nil {
		taxbitRows = taxbitRows[*firstToKeep:]
	} else if lastToKeep != nil {
		taxbitRows = taxbitRows[:*lastToKeep]
	}

	// Copy cointrackerRows into csvRows for return val
	csvRows := make([]parsers.CsvRow, len(taxbitRows))
	for i, v := range taxbitRows {
		csvRows[i] = v
	}

	return csvRows, nil
}

func (p Parser) GetHeaders() []string {
	return []string{
		"Date and Time", "Transaction Type", "Sent Quantity", "Sent Currency", "Sending Source",
		"Received Quantity", "Received Currency", "Receiving Destination", "Fee", "Fee Currency", "Exchange Transaction ID",
		"Blockchain Transaction Hash",
	}
}

// HandleFees:
// If the transaction lists the same amount of fees as there are rows in the CSV,
// then we spread the fees out one per row. Otherwise we add a line for the fees,
// where each fee has a separate line.
func HandleFees(address string, events []db.TaxableTransaction) (rows []Row, err error) {
	// No events -- This address didn't pay any fees
	if len(events) == 0 {
		return rows, nil
	}

	// We need to gather all unique fees, but we are receiving Messages not Txes
	// Make a map from TX hash to fees array to keep unique
	txToFeesMap := make(map[uint][]db.Fee)
	txIdsToTx := make(map[uint]db.Tx)
	for _, event := range events {
		txID := event.Message.Tx.ID
		feeStore := event.Message.Tx.Fees
		txToFeesMap[txID] = feeStore
		txIdsToTx[txID] = event.Message.Tx
	}

	for id, txFees := range txToFeesMap {
		for _, fee := range txFees {
			if fee.PayerAddress.Address == address {
				newRow := Row{}
				err = newRow.ParseFee(txIdsToTx[id], fee)
				if err != nil {
					return nil, err
				}
				rows = append(rows, newRow)
			}
		}
	}

	return rows, nil
}

// ParseEvent: Parse the potentially taxable event
func ParseEvent(event db.TaxableEvent) (rows []Row, err error) {
	if event.Source == db.OsmosisRewardDistribution {
		row, err := ParseOsmosisReward(event)
		if err != nil {
			config.Log.Error("error parsing row. Should be impossible to reach this condition, ideally (once all bugs worked out)", err)
			return nil, err
		}
		rows = append(rows, row)
	}

	return rows, nil
}

// ParseTx: Parse the potentially taxable TX and Messages
// This function is used for parsing a single TX that will not need to relate to any others
// Use TX Parsing Groups to parse txes as a group
func ParseTx(address string, events []db.TaxableTransaction) (rows []parsers.CsvRow, err error) {
	for _, event := range events {
		var newRow Row
		var err error
		switch event.Message.MessageType.MessageType {
		case bank.MsgSendV0:
			newRow, err = ParseMsgSend(address, event)
		case bank.MsgSend:
			newRow, err = ParseMsgSend(address, event)
		case bank.MsgMultiSendV0:
			newRow, err = ParseMsgMultiSend(address, event)
		case bank.MsgMultiSend:
			newRow, err = ParseMsgMultiSend(address, event)
		case distribution.MsgFundCommunityPool:
			newRow, err = ParseMsgFundCommunityPool(address, event)
		case distribution.MsgWithdrawValidatorCommission:
			newRow, err = ParseMsgWithdrawValidatorCommission(address, event)
		case distribution.MsgWithdrawRewards:
			newRow, err = ParseMsgWithdrawDelegatorReward(address, event)
		case distribution.MsgWithdrawDelegatorReward:
			newRow, err = ParseMsgWithdrawDelegatorReward(address, event)
		case staking.MsgDelegate:
			newRow, err = ParseMsgWithdrawDelegatorReward(address, event)
		case staking.MsgUndelegate:
			newRow, err = ParseMsgWithdrawDelegatorReward(address, event)
		case staking.MsgBeginRedelegate:
			newRow, err = ParseMsgWithdrawDelegatorReward(address, event)
		case gov.MsgSubmitProposal:
			newRow, err = ParseMsgSubmitProposal(address, event)
		case gamm.MsgSwapExactAmountIn:
			newRow, err = ParseMsgSwapExactAmountIn(event)
		case gamm.MsgSwapExactAmountOut:
			newRow, err = ParseMsgSwapExactAmountOut(event)
		case gov.MsgDeposit:
			newRow, err = ParseMsgDeposit(address, event)
		case ibc.MsgTransfer:
			newRow, err = ParseMsgTransfer(address, event)
		case ibc.MsgAcknowledgement:
			newRow, err = ParseMsgTransfer(address, event)
		case ibc.MsgRecvPacket:
			newRow, err = ParseMsgTransfer(address, event)
		case poolmanager.MsgSplitRouteSwapExactAmountIn, poolmanager.MsgSwapExactAmountIn, poolmanager.MsgSwapExactAmountOut:
			newRow, err = ParsePoolManagerSwap(event)
		default:
			config.Log.Errorf("no parser for message type '%v'", event.Message.MessageType.MessageType)
			continue
		}

		if err != nil {
			config.Log.Errorf("error parsing message type '%v': %v", event.Message.MessageType.MessageType, err)
			continue
		}

		rows = append(rows, newRow)
	}
	return rows, nil
}

// ParseMsgValidatorWithdraw:
// This transaction is always a withdrawal.
func ParseMsgWithdrawValidatorCommission(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgWithdrawValidatorCommission.", err)
	}
	// row.Label = Unstake
	return *row, err
}

// ParseMsgValidatorWithdraw:
// This transaction is always a withdrawal.
func ParseMsgWithdrawDelegatorReward(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgWithdrawDelegatorReward.", err)
	}
	// row.Label = Unstake
	return *row, err
}

// ParseMsgSend:
// If the address we searched is the receiver, then this transaction is a deposit.
// If the address we searched is the sender, then this transaction is a withdrawal.
func ParseMsgSend(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgSend.", err)
	}
	return *row, err
}

func ParseMsgMultiSend(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgMultiSend.", err)
	}
	return *row, err
}

func ParseMsgFundCommunityPool(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgFundCommunityPool.", err)
	}
	return *row, err
}

func ParseMsgSwapExactAmountIn(event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseSwap(event)
	if err != nil {
		config.Log.Error("Error with ParseMsgSwapExactAmountIn.", err)
	}
	return *row, err
}

func ParseMsgSwapExactAmountOut(event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseSwap(event)
	if err != nil {
		config.Log.Error("Error with ParseMsgSwapExactAmountOut.", err)
	}
	return *row, err
}

func ParseMsgTransfer(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgTransfer.", err)
	}
	return *row, err
}

func ParseMsgSubmitProposal(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgSubmitProposal.", err)
	}
	return *row, err
}

func ParseMsgDeposit(address string, event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseBasic(address, event)
	if err != nil {
		config.Log.Error("Error with ParseMsgDeposit.", err)
	}
	return *row, err
}

func ParseOsmosisReward(event db.TaxableEvent) (Row, error) {
	row := &Row{}
	err := row.EventParseBasic(event)
	if err != nil {
		config.Log.Error("Error with ParseOsmosisReward.", err)
	}
	return *row, err
}

func ParsePoolManagerSwap(event db.TaxableTransaction) (Row, error) {
	row := &Row{}
	err := row.ParseSwap(event)
	if err != nil {
		config.Log.Error("Error with ParseMsgSwapExactAmountOut.", err)
	}
	return *row, err
}
