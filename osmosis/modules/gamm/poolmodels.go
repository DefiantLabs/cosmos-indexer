package gamm

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gammBalancerPoolModelsTypes "github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/balancer"
	gammStableswapPoolModelsTypes "github.com/osmosis-labs/osmosis/v19/x/gamm/pool-models/stableswap"
)

const (
	PoolModelsMsgCreateBalancerPool   = "/osmosis.gamm.poolmodels.balancer.v1beta1.MsgCreateBalancerPool"
	PoolModelsMsgCreateStableswapPool = "/osmosis.gamm.poolmodels.stableswap.v1beta1.MsgCreateStableswapPool"
)

type WrapperPoolModelsMsgCreateBalancerPool struct {
	txModule.Message
	OsmosisMsgCreateBalancerPool *gammBalancerPoolModelsTypes.MsgCreateBalancerPool
	CoinsSpent                   []sdk.Coin
	GammCoinsReceived            sdk.Coin
	OtherCoinsReceived           []coinReceived // e.g. from claims module (airdrops)
}

func (sf *WrapperPoolModelsMsgCreateBalancerPool) String() string {
	var tokensIn []string
	if !(len(sf.OsmosisMsgCreateBalancerPool.PoolAssets) == 0) {
		for _, v := range sf.OsmosisMsgCreateBalancerPool.PoolAssets {
			tokensIn = append(tokensIn, v.Token.String())
		}
	}
	return fmt.Sprintf("MsgCreateBalancerPool: %s created pool with %s",
		sf.OsmosisMsgCreateBalancerPool.Sender, strings.Join(tokensIn, ", "))
}

func (sf *WrapperPoolModelsMsgCreateBalancerPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreateBalancerPool = msg.(*gammBalancerPoolModelsTypes.MsgCreateBalancerPool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinSpentEvents := txModule.GetEventsWithType(bankTypes.EventTypeCoinSpent, log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsSpent := txModule.GetCoinsSpent(sf.OsmosisMsgCreateBalancerPool.Sender, coinSpentEvents)

	if len(coinsSpent) < 2 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("invalid number of coins spent: %+v", log)}
	}

	sf.CoinsSpent = []sdk.Coin{}
	for _, coin := range coinsSpent {
		t, err := sdk.ParseCoinNormalized(coin)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		sf.CoinsSpent = append(sf.CoinsSpent, t)
	}

	coinReceivedEvents := txModule.GetEventsWithType(bankTypes.EventTypeCoinReceived, log)
	if len(coinReceivedEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsReceived := txModule.GetCoinsReceived(sf.OsmosisMsgCreateBalancerPool.Sender, coinReceivedEvents)

	gammCoinsReceived := []string{}

	for _, coin := range coinsReceived {
		if strings.Contains(coin, "gamm/pool") {
			gammCoinsReceived = append(gammCoinsReceived, coin)
		} else {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("unexpected non-gamm/pool coin received: %+v", log)}
		}
	}

	if len(gammCoinsReceived) != 1 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("invalid number of coins received: %+v", log)}
	}

	var err error
	sf.GammCoinsReceived, err = sdk.ParseCoinNormalized(gammCoinsReceived[0])
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return nil
}

func (sf *WrapperPoolModelsMsgCreateBalancerPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// need to make a relevant data block for all Tokens sent to the pool on creation
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.CoinsSpent)+len(sf.OtherCoinsReceived))

	// figure out how many gams per token
	nthGamms, remainderGamms := calcNthGams(sf.GammCoinsReceived.Amount.BigInt(), len(sf.CoinsSpent))
	for i, v := range sf.CoinsSpent {
		// split received tokens across entry so we receive GAMM tokens for both exchanges
		// each swap will get 1 nth of the gams until the last one which will get the remainder
		if i != len(sf.CoinsSpent)-1 {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       nthGamms,
				DenominationReceived: sf.GammCoinsReceived.Denom,
				SenderAddress:        sf.OsmosisMsgCreateBalancerPool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreateBalancerPool.Sender,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       remainderGamms,
				DenominationReceived: sf.GammCoinsReceived.Denom,
				SenderAddress:        sf.OsmosisMsgCreateBalancerPool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreateBalancerPool.Sender,
			}
		}
	}

	i := len(sf.CoinsSpent)
	for _, c := range sf.OtherCoinsReceived {
		relevantData[i] = parsingTypes.MessageRelevantInformation{
			AmountSent:           c.coinReceived.Amount.BigInt(),
			DenominationSent:     c.coinReceived.Denom,
			AmountReceived:       c.coinReceived.Amount.BigInt(),
			DenominationReceived: c.coinReceived.Denom,
			SenderAddress:        c.sender,
			ReceiverAddress:      sf.OsmosisMsgCreateBalancerPool.Sender,
		}
		i++
	}
	return relevantData
}

type WrapperPoolModelsMsgCreateStableswapPool struct {
	txModule.Message
	OsmosisMsgCreateStableswapPool *gammStableswapPoolModelsTypes.MsgCreateStableswapPool
	CoinsSpent                     []sdk.Coin
	GammCoinsReceived              sdk.Coin
	OtherCoinsReceived             []coinReceived // e.g. from claims module (airdrops)
}

func (sf *WrapperPoolModelsMsgCreateStableswapPool) String() string {
	var tokensIn []string
	if !(len(sf.CoinsSpent) == 0) {
		for _, v := range sf.CoinsSpent {
			tokensIn = append(tokensIn, v.String())
		}
	}
	return fmt.Sprintf("MsgCreateStableswapPool: %s created pool with %s",
		sf.OsmosisMsgCreateStableswapPool.Sender, strings.Join(tokensIn, ", "))
}

func (sf *WrapperPoolModelsMsgCreateStableswapPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreateStableswapPool = msg.(*gammStableswapPoolModelsTypes.MsgCreateStableswapPool)
	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinSpentEvents := txModule.GetEventsWithType(bankTypes.EventTypeCoinSpent, log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsSpent := txModule.GetCoinsSpent(sf.OsmosisMsgCreateStableswapPool.Sender, coinSpentEvents)

	if len(coinsSpent) < 2 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("invalid number of coins spent: %+v", log)}
	}

	sf.CoinsSpent = []sdk.Coin{}
	for _, coin := range coinsSpent {
		t, err := sdk.ParseCoinNormalized(coin)
		if err != nil {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
		}

		sf.CoinsSpent = append(sf.CoinsSpent, t)
	}

	coinReceivedEvents := txModule.GetEventsWithType(bankTypes.EventTypeCoinReceived, log)
	if len(coinReceivedEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsReceived := txModule.GetCoinsReceived(sf.OsmosisMsgCreateStableswapPool.Sender, coinReceivedEvents)

	gammCoinsReceived := []string{}

	for _, coin := range coinsReceived {
		if strings.Contains(coin, "gamm/pool") {
			gammCoinsReceived = append(gammCoinsReceived, coin)
		} else {
			return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("unexpected non-gamm/pool coin received: %+v", log)}
		}
	}

	if len(gammCoinsReceived) != 1 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("invalid number of coins received: %+v", log)}
	}

	var err error
	sf.GammCoinsReceived, err = sdk.ParseCoinNormalized(gammCoinsReceived[0])
	if err != nil {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	return nil
}

func (sf *WrapperPoolModelsMsgCreateStableswapPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	// need to make a relevant data block for all Tokens sent to the pool on creation
	relevantData := make([]parsingTypes.MessageRelevantInformation, len(sf.CoinsSpent)+len(sf.OtherCoinsReceived))

	// figure out how many gams per token
	nthGamms, remainderGamms := calcNthGams(sf.GammCoinsReceived.Amount.BigInt(), len(sf.CoinsSpent))
	for i, v := range sf.CoinsSpent {
		// split received tokens across entry so we receive GAMM tokens for both exchanges
		// each swap will get 1 nth of the gams until the last one which will get the remainder
		if i != len(sf.CoinsSpent)-1 {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       nthGamms,
				DenominationReceived: sf.GammCoinsReceived.Denom,
				SenderAddress:        sf.OsmosisMsgCreateStableswapPool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreateStableswapPool.Sender,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       remainderGamms,
				DenominationReceived: sf.GammCoinsReceived.Denom,
				SenderAddress:        sf.OsmosisMsgCreateStableswapPool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreateStableswapPool.Sender,
			}
		}
	}

	i := len(sf.CoinsSpent)
	for _, c := range sf.OtherCoinsReceived {
		relevantData[i] = parsingTypes.MessageRelevantInformation{
			AmountSent:           c.coinReceived.Amount.BigInt(),
			DenominationSent:     c.coinReceived.Denom,
			AmountReceived:       c.coinReceived.Amount.BigInt(),
			DenominationReceived: c.coinReceived.Denom,
			SenderAddress:        c.sender,
			ReceiverAddress:      sf.OsmosisMsgCreateStableswapPool.Sender,
		}
		i++
	}
	return relevantData
}
