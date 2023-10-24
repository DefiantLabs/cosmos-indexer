// nolint:unused
package csv

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/ibc"
	"github.com/DefiantLabs/cosmos-indexer/db"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

var zeroTime time.Time

func getTestTaxableEvents(t *testing.T, targetAddress db.Address, targetChain db.Chain) []db.TaxableEvent {
	// BlockTimes
	oneYearAgo := time.Now().Add(-1 * time.Hour * 24 * 365)
	sixMonthAgo := time.Now().Add(-1 * time.Hour * 24 * 182)

	// create some blocks to put the transactions in
	block1 := mkBlk(1, 1, oneYearAgo, targetChain)
	block2 := mkBlk(2, 2, sixMonthAgo, targetChain)

	// create denom
	coinDenom, _ := mkDenom(1, "coin", "Some Coin", "SC")

	event1 := mkTaxableEvent(1, decimal.NewFromInt(10), coinDenom, targetAddress, block1)
	event2 := mkTaxableEvent(2, decimal.NewFromInt(10), coinDenom, targetAddress, block2)

	return []db.TaxableEvent{event1, event2}
}

func mkTaxableEvent(id uint, amount decimal.Decimal, denom db.Denom, address db.Address, block db.Block) db.TaxableEvent {
	return db.TaxableEvent{
		ID:             id,
		Source:         db.OsmosisRewardDistribution,
		Amount:         amount,
		DenominationID: denom.ID,
		Denomination:   denom,
		AddressID:      address.ID,
		EventAddress:   address,
		BlockID:        block.ID,
		Block:          block,
	}
}

func getTestIbcTransferTXs(t *testing.T, sourceAddress db.Address, targetAddress db.Address, sourceChain db.Chain, targetChain db.Chain) []db.TaxableTransaction {
	// BlockTimes
	oneYearAgo := time.Now().Add(-1 * time.Hour * 24 * 365)

	// the transfer on the source chain
	block1 := mkBlk(1, 1, oneYearAgo, sourceChain)

	// create the transfer msg
	msgTypeTransfer := mkMsgType(1, ibc.MsgTransfer)

	// create TXs
	transferTX1 := mkTx(1, "somehash1", 0, block1, sourceAddress, nil)

	// create Msgs
	transferMsg := mkMsg(1, transferTX1, msgTypeTransfer, 0)

	// create denoms
	coin1, coin1DenomUnit := mkDenom(1, "ujuno", "Juno", "JUNO")

	// populate denom cache
	db.CachedDenomUnits = []db.DenomUnit{coin1DenomUnit}

	// create taxable transactions
	// joins
	taxableTX1 := mkTaxableTransaction(1, transferMsg, decimal.NewFromInt(1000000), decimal.NewFromInt(1000000), coin1, coin1, sourceAddress, targetAddress)

	return []db.TaxableTransaction{taxableTX1}
}

func mkTaxableTransaction(id uint, msg db.Message, amntSent, amntReceived decimal.Decimal, denomSent db.Denom, denomReceived db.Denom, senderAddr db.Address, receiverAddr db.Address) db.TaxableTransaction {
	return db.TaxableTransaction{
		ID:                     id,
		MessageID:              msg.ID,
		Message:                msg,
		AmountSent:             amntSent,
		AmountReceived:         amntReceived,
		DenominationSentID:     &denomSent.ID,
		DenominationSent:       denomSent,
		DenominationReceivedID: &denomReceived.ID,
		DenominationReceived:   denomReceived,
		SenderAddressID:        &senderAddr.ID,
		SenderAddress:          senderAddr,
		ReceiverAddressID:      &receiverAddr.ID,
		ReceiverAddress:        receiverAddr,
	}
}

func mkDenom(id uint, base, name, symbol string) (denom db.Denom, denomUnit db.DenomUnit) {
	denom = db.Denom{
		ID:     id,
		Base:   base,
		Name:   name,
		Symbol: symbol,
	}

	denomUnit = db.DenomUnit{
		ID:       id,
		DenomID:  id,
		Denom:    denom,
		Exponent: 0,
		Name:     denom.Base,
	}

	return
}

func mkMsg(id uint, tx db.Tx, msgType db.MessageType, msgIdx int) db.Message {
	return db.Message{
		ID:            id,
		TxID:          tx.ID,
		Tx:            tx,
		MessageTypeID: msgType.ID,
		MessageType:   msgType,
		MessageIndex:  msgIdx,
	}
}

func mkMsgType(id uint, msgType string) db.MessageType {
	return db.MessageType{
		ID:          id,
		MessageType: msgType,
	}
}

func mkTx(id uint, hash string, code uint32, block db.Block, signerAddr db.Address, fees []db.Fee) db.Tx {
	return db.Tx{
		ID:              id,
		Hash:            hash,
		Code:            code,
		BlockID:         block.ID,
		Block:           block,
		SignerAddressID: &signerAddr.ID,
		SignerAddress:   signerAddr,
		Fees:            fees,
	}
}

func mkBlk(id uint, height int64, timestamp time.Time, chain db.Chain) db.Block {
	return db.Block{
		ID:           id,
		Height:       height,
		TimeStamp:    timestamp,
		BlockchainID: chain.ID,
		Chain:        chain,
	}
}

func mkChain(id uint, chainID, chainName string) db.Chain {
	return db.Chain{
		ID:      id,
		ChainID: chainID,
		Name:    chainName,
	}
}

func mkAddress(t *testing.T, id uint) db.Address {
	rand32 := make([]byte, 32)
	_, err := rand.Read(rand32)
	assert.Nil(t, err)
	address := fmt.Sprintf("osmo%v", rand32)
	return db.Address{
		ID:      id,
		Address: address,
	}
}
