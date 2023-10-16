package cosmwasmpool

import (
	"errors"
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	cosmwasmPoolModelTypes "github.com/osmosis-labs/osmosis/v19/x/cosmwasmpool/model"
)

const (
	MsgCreateCosmWasmPool = "/osmosis.cosmwasmpool.v1beta1.MsgCreateCosmWasmPool"
)

type WrapperMsgCreateCosmWasmPool struct {
	txModule.Message
	OsmosisMsgCreateCosmWasmPool *cosmwasmPoolModelTypes.MsgCreateCosmWasmPool
	Address                      string
	TokensSpent                  sdk.Coins
}

func (sf *WrapperMsgCreateCosmWasmPool) String() string {
	var tokensSpent []string
	if !(len(sf.TokensSpent) == 0) {
		for _, v := range sf.TokensSpent {
			tokensSpent = append(tokensSpent, v.String())
		}
	}
	return fmt.Sprintf("MsgCreateCosmWasmPool: %s created pool and spent %s",
		sf.Address, strings.Join(tokensSpent, ", "))
}

func (sf *WrapperMsgCreateCosmWasmPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreateCosmWasmPool = msg.(*cosmwasmPoolModelTypes.MsgCreateCosmWasmPool)

	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinSpentEvents := txModule.GetEventsWithType("coin_spent", log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	senderCoinsSpentStrings := txModule.GetCoinsSpent(sf.OsmosisMsgCreateCosmWasmPool.Sender, coinSpentEvents)

	for _, coinSpentString := range senderCoinsSpentStrings {
		coinsSpent, err := sdk.ParseCoinsNormalized(coinSpentString)
		if err != nil {
			return errors.New("error parsing coins spent from event")
		}

		sf.TokensSpent = append(sf.TokensSpent, coinsSpent...)
	}

	sf.Address = sf.OsmosisMsgCreateCosmWasmPool.Sender

	return nil
}

func (sf *WrapperMsgCreateCosmWasmPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)
	for _, token := range sf.TokensSpent {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountSent:       token.Amount.BigInt(),
				DenominationSent: token.Denom,
				SenderAddress:    sf.Address,
			})
		}
	}
	return relevantData
}
