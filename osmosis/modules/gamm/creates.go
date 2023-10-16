package gamm

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	osmosisOldTypes "github.com/DefiantLabs/lens/extra-codecs/osmosis/gamm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

const (
	MsgCreatePool         = "/osmosis.gamm.v1beta1.MsgCreatePool"
	MsgCreateBalancerPool = "/osmosis.gamm.v1beta1.MsgCreateBalancerPool"
)

type WrapperMsgCreatePool struct {
	txModule.Message
	OsmosisMsgCreatePool *osmosisOldTypes.MsgCreatePool
	CoinsSpent           []sdk.Coin
	GammCoinsReceived    sdk.Coin
	OtherCoinsReceived   []coinReceived // e.g. from claims module (airdrops)
}

type WrapperMsgCreateBalancerPool struct {
	txModule.Message
	OsmosisMsgCreateBalancerPool *osmosisOldTypes.MsgCreateBalancerPool
	CoinsSpent                   []sdk.Coin
	GammCoinsReceived            sdk.Coin
	OtherCoinsReceived           []coinReceived // e.g. from claims module (airdrops)
}

type WrapperMsgCreatePool2 struct {
	WrapperMsgCreatePool
}

func (sf *WrapperMsgCreatePool) String() string {
	var tokensIn []string
	if !(len(sf.OsmosisMsgCreatePool.PoolAssets) == 0) {
		for _, v := range sf.OsmosisMsgCreatePool.PoolAssets {
			tokensIn = append(tokensIn, v.Token.String())
		}
	}
	return fmt.Sprintf("MsgCreatePool: %s created pool with %s",
		sf.OsmosisMsgCreatePool.Sender, strings.Join(tokensIn, ", "))
}

func (sf *WrapperMsgCreatePool2) String() string {
	return sf.WrapperMsgCreatePool.String()
}

func (sf *WrapperMsgCreateBalancerPool) String() string {
	var tokensIn []string
	if !(len(sf.OsmosisMsgCreateBalancerPool.PoolAssets) == 0) {
		for _, v := range sf.OsmosisMsgCreateBalancerPool.PoolAssets {
			tokensIn = append(tokensIn, v.Token.String())
		}
	}
	return fmt.Sprintf("MsgCreateBalancerPool: %s created pool with %s",
		sf.OsmosisMsgCreateBalancerPool.Sender, strings.Join(tokensIn, ", "))
}

func (sf *WrapperMsgCreatePool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreatePool = msg.(*osmosisOldTypes.MsgCreatePool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinSpentEvents := txModule.GetEventsWithType(bankTypes.EventTypeCoinSpent, log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsSpent := txModule.GetCoinsSpent(sf.OsmosisMsgCreatePool.Sender, coinSpentEvents)

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

	coinsReceived := txModule.GetCoinsReceived(sf.OsmosisMsgCreatePool.Sender, coinReceivedEvents)

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

	return err
}

func (sf *WrapperMsgCreatePool2) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreatePool = msg.(*osmosisOldTypes.MsgCreatePool)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	transferEvents := txModule.GetEventsWithType(bankTypes.EventTypeTransfer, log)
	if len(transferEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	coinsSpent := []string{}
	coinsReceived := []coinReceived{}

	for _, transfer := range transferEvents {
		parsedTransfer, err := txModule.ParseTransferEvent(transfer)
		if err != nil {
			return err
		}

		for _, curr := range parsedTransfer {
			coins := strings.Split(curr.Amount, ",")
			if curr.Recipient == sf.OsmosisMsgCreatePool.Sender {
				for _, coin := range coins {
					t, err := sdk.ParseCoinNormalized(coin)
					if err != nil {
						return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
					}

					coinsReceived = append(coinsReceived, coinReceived{
						sender:       curr.Sender,
						coinReceived: t,
					})
				}
			} else if curr.Sender == sf.OsmosisMsgCreatePool.Sender {
				coinsSpent = append(coinsSpent, coins...)
			}
		}
	}

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

	gammCoinsReceived := []sdk.Coin{}
	sf.OtherCoinsReceived = []coinReceived{}

	for _, coin := range coinsReceived {
		if strings.Contains(coin.coinReceived.Denom, "gamm/pool") {
			gammCoinsReceived = append(gammCoinsReceived, coin.coinReceived)
		} else {
			sf.OtherCoinsReceived = append(sf.OtherCoinsReceived, coin)
		}
	}

	if len(gammCoinsReceived) != 1 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("invalid number of coins received: %+v", log)}
	}

	sf.GammCoinsReceived = gammCoinsReceived[0]

	return nil
}

func (sf *WrapperMsgCreateBalancerPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreateBalancerPool = msg.(*osmosisOldTypes.MsgCreateBalancerPool)

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

	return err
}

func (sf *WrapperMsgCreatePool2) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return sf.WrapperMsgCreatePool.ParseRelevantData()
}

func (sf *WrapperMsgCreatePool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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
				SenderAddress:        sf.OsmosisMsgCreatePool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreatePool.Sender,
			}
		} else {
			relevantData[i] = parsingTypes.MessageRelevantInformation{
				AmountSent:           v.Amount.BigInt(),
				DenominationSent:     v.Denom,
				AmountReceived:       remainderGamms,
				DenominationReceived: sf.GammCoinsReceived.Denom,
				SenderAddress:        sf.OsmosisMsgCreatePool.Sender,
				ReceiverAddress:      sf.OsmosisMsgCreatePool.Sender,
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
			ReceiverAddress:      sf.OsmosisMsgCreatePool.Sender,
		}
		i++
	}

	return relevantData
}

func (sf *WrapperMsgCreateBalancerPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
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
