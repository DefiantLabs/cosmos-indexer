// nolint:unused
package csv

import (
	"crypto/rand"
	"fmt"
	"testing"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/modules/ibc"
	"github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/osmosis/modules/gamm"

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

func getTestSwapTXs(t *testing.T, targetAddress db.Address, targetChain db.Chain) []db.TaxableTransaction {
	randoAddress := mkAddress(t, 2)

	// BlockTimes
	oneYearAgo := time.Now().Add(-1 * time.Hour * 24 * 365)
	sixMonthAgo := time.Now().Add(-1 * time.Hour * 24 * 182)

	// create some blocks to put the transactions in
	block1 := mkBlk(1, 1, oneYearAgo, targetChain)
	block2 := mkBlk(2, 2, sixMonthAgo, targetChain)

	// create the swap messages
	// joins
	joinSwapExternAmountIn := mkMsgType(1, gamm.MsgJoinSwapExternAmountIn)
	joinSwapShareAmountOut := mkMsgType(2, gamm.MsgJoinSwapShareAmountOut)
	joinPool := mkMsgType(3, gamm.MsgJoinPool)
	// exits
	exitSwapShareAmountIn := mkMsgType(4, gamm.MsgExitSwapShareAmountIn)
	exitSwapExternAmountOut := mkMsgType(5, gamm.MsgExitSwapExternAmountOut)
	exitPool := mkMsgType(6, gamm.MsgExitPool)

	// create TXs
	joinPoolTX1 := mkTx(1, "somehash1", 0, block1, randoAddress, nil)
	joinPoolTX2 := mkTx(2, "somehash2", 0, block1, randoAddress, nil)

	leavePoolTX1 := mkTx(3, "somehash4", 0, block2, randoAddress, nil)
	leavePoolTX2 := mkTx(4, "somehash5", 0, block2, randoAddress, nil)

	// create Msgs
	joinSwapExternAmountInMsg := mkMsg(1, joinPoolTX1, joinSwapExternAmountIn, 0)
	joinSwapShareAmountOutMsg := mkMsg(2, joinPoolTX1, joinSwapShareAmountOut, 1)
	joinPoolMsg := mkMsg(3, joinPoolTX2, joinPool, 0)

	exitSwapShareAmountInMsg := mkMsg(4, leavePoolTX1, exitSwapShareAmountIn, 0)
	exitSwapExternAmountOutMsg := mkMsg(5, leavePoolTX1, exitSwapExternAmountOut, 1)
	exitPoolMsg := mkMsg(6, leavePoolTX2, exitPool, 2)

	// create denoms
	coin1, coin1DenomUnit := mkDenom(1, "coin1", "Some Coin", "SC1")
	coin2, coin2DenomUnit := mkDenom(2, "coin2", "Another Coin", "AC2")
	gamm1, gamm1DenomUnit := mkDenom(3, "gamm/pool/1", "UNKNOWN", "UNKNOWN")

	// populate denom cache
	db.CachedDenomUnits = []db.DenomUnit{coin1DenomUnit, coin2DenomUnit, gamm1DenomUnit}

	// create taxable transactions
	// joins
	taxableTX1 := mkTaxableTransaction(1, joinSwapExternAmountInMsg, decimal.NewFromInt(12000), decimal.NewFromInt(24000038), coin1, gamm1, targetAddress, targetAddress)
	taxableTX2 := mkTaxableTransaction(2, joinSwapShareAmountOutMsg, decimal.NewFromInt(11999), decimal.NewFromInt(24000000), coin1, gamm1, targetAddress, targetAddress)
	taxableTX3 := mkTaxableTransaction(3, joinPoolMsg, decimal.NewFromInt(6000), decimal.NewFromFloat(12000019), coin1, gamm1, targetAddress, targetAddress)
	taxableTX4 := mkTaxableTransaction(4, joinPoolMsg, decimal.NewFromInt(3000), decimal.NewFromFloat(12000019), coin2, gamm1, targetAddress, targetAddress)
	// exits
	taxableTX5 := mkTaxableTransaction(5, exitSwapShareAmountInMsg, decimal.NewFromInt(24000038), decimal.NewFromInt(12438), gamm1, coin1, targetAddress, targetAddress)
	taxableTX6 := mkTaxableTransaction(6, exitSwapExternAmountOutMsg, decimal.NewFromInt(23152955), decimal.NewFromInt(11999), gamm1, coin1, targetAddress, targetAddress)
	taxableTX7 := mkTaxableTransaction(7, exitPoolMsg, decimal.NewFromInt(12000019), decimal.NewFromInt(6219), gamm1, coin1, targetAddress, targetAddress)
	taxableTX8 := mkTaxableTransaction(8, exitPoolMsg, decimal.NewFromInt(12000019), decimal.NewFromInt(2853), gamm1, coin2, targetAddress, targetAddress)

	return []db.TaxableTransaction{taxableTX1, taxableTX2, taxableTX3, taxableTX4, taxableTX5, taxableTX6, taxableTX7, taxableTX8}
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
