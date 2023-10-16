package concentratedliquidity

import (
	"errors"
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	clPoolTypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/model"
	clTypes "github.com/osmosis-labs/osmosis/v19/x/concentrated-liquidity/types"
)

const (
	MsgCreatePosition         = "/osmosis.concentratedliquidity.v1beta1.MsgCreatePosition"
	MsgWithdrawPosition       = "/osmosis.concentratedliquidity.v1beta1.MsgWithdrawPosition"
	MsgCollectSpreadRewards   = "/osmosis.concentratedliquidity.v1beta1.MsgCollectSpreadRewards"
	MsgCreateConcentratedPool = "/osmosis.concentratedliquidity.poolmodel.concentrated.v1beta1.MsgCreateConcentratedPool"
	MsgCollectIncentives      = "/osmosis.concentratedliquidity.v1beta1.MsgCollectIncentives"
	MsgAddToPosition          = "/osmosis.concentratedliquidity.v1beta1.MsgAddToPosition"
	tokensOutEvent            = "tokens_out"
)

type WrapperMsgCreatePosition struct {
	txModule.Message
	OsmosisMsgCreatePosition *clTypes.MsgCreatePosition
	TokensSent               sdk.Coins
	Address                  string
}

func (sf *WrapperMsgCreatePosition) String() string {
	var tokensSent []string
	if !(len(sf.TokensSent) == 0) {
		for _, v := range sf.TokensSent {
			tokensSent = append(tokensSent, v.String())
		}
	}
	return fmt.Sprintf("MsgCreatePosition: %s created position by sending %s",
		sf.Address, strings.Join(tokensSent, ", "))
}

func (sf *WrapperMsgCreatePosition) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreatePosition = msg.(*clTypes.MsgCreatePosition)

	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinSpentEvents := txModule.GetEventsWithType("coin_spent", log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	senderCoinsSpentStrings := txModule.GetCoinsSpent(sf.OsmosisMsgCreatePosition.Sender, coinSpentEvents)

	for _, coinReceivedString := range senderCoinsSpentStrings {
		coinsReceived, err := sdk.ParseCoinsNormalized(coinReceivedString)
		if err != nil {
			return errors.New("error parsing coins received from event")
		}

		sf.TokensSent = append(sf.TokensSent, coinsReceived...)
	}

	sf.Address = sf.OsmosisMsgCreatePosition.Sender

	return nil
}

func (sf *WrapperMsgCreatePosition) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)

	for _, token := range sf.TokensSent {
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

type WrapperMsgWithdrawPosition struct {
	txModule.Message
	OsmosisMsgWithdrawPosition *clTypes.MsgWithdrawPosition
	TokensRecieved             sdk.Coins
	Address                    string
}

func (sf *WrapperMsgWithdrawPosition) String() string {
	var tokensRecv []string
	if !(len(sf.TokensRecieved) == 0) {
		for _, v := range sf.TokensRecieved {
			tokensRecv = append(tokensRecv, v.String())
		}
	}
	return fmt.Sprintf("MsgWithdrawPosition: %s withdrew position by receiving %s",
		sf.Address, strings.Join(tokensRecv, ", "))
}

func (sf *WrapperMsgWithdrawPosition) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgWithdrawPosition = msg.(*clTypes.MsgWithdrawPosition)

	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coinReceivedEvents := txModule.GetEventsWithType("coin_received", log)
	if len(coinReceivedEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	senderCoinsReceivedStrings := txModule.GetCoinsReceived(sf.OsmosisMsgWithdrawPosition.Sender, coinReceivedEvents)

	for _, coinReceivedString := range senderCoinsReceivedStrings {
		coinsReceived, err := sdk.ParseCoinsNormalized(coinReceivedString)
		if err != nil {
			return errors.New("error parsing coins received from event")
		}

		sf.TokensRecieved = append(sf.TokensRecieved, coinsReceived...)
	}

	sf.Address = sf.OsmosisMsgWithdrawPosition.Sender

	return nil
}

func (sf *WrapperMsgWithdrawPosition) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)
	for _, token := range sf.TokensRecieved {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountReceived:       token.Amount.BigInt(),
				DenominationReceived: token.Denom,
				SenderAddress:        sf.Address,
			})
		}
	}
	return relevantData
}

type WrapperMsgCollectSpreadRewards struct {
	txModule.Message
	OsmosisMsgCollectSpreadRewards *clTypes.MsgCollectSpreadRewards
	TokensRecieved                 sdk.Coins
	Address                        string
}

func (sf *WrapperMsgCollectSpreadRewards) String() string {
	var tokensRecv []string
	if !(len(sf.TokensRecieved) == 0) {
		for _, v := range sf.TokensRecieved {
			tokensRecv = append(tokensRecv, v.String())
		}
	}
	return fmt.Sprintf("MsgCollectSpreadRewards: %s received rewards of amount %s",
		sf.Address, strings.Join(tokensRecv, ", "))
}

func (sf *WrapperMsgCollectSpreadRewards) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCollectSpreadRewards = msg.(*clTypes.MsgCollectSpreadRewards)

	coinReceivedEvents := txModule.GetEventsWithType("coin_received", log)
	if len(coinReceivedEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	senderCoinsReceivedStrings := txModule.GetCoinsReceived(sf.OsmosisMsgCollectSpreadRewards.Sender, coinReceivedEvents)

	for _, coinReceivedString := range senderCoinsReceivedStrings {
		coinsReceived, err := sdk.ParseCoinsNormalized(coinReceivedString)
		if err != nil {
			return errors.New("error parsing coins received from event")
		}

		sf.TokensRecieved = append(sf.TokensRecieved, coinsReceived...)
	}

	sf.Address = sf.OsmosisMsgCollectSpreadRewards.Sender

	return nil
}

func (sf *WrapperMsgCollectSpreadRewards) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)
	for _, token := range sf.TokensRecieved {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountReceived:       token.Amount.BigInt(),
				DenominationReceived: token.Denom,
				SenderAddress:        sf.Address,
			})
		}
	}
	return relevantData
}

type WrappeMsgCreateConcentratedPool struct {
	txModule.Message
	OsmosisMsgCreateConcentratedPool *clPoolTypes.MsgCreateConcentratedPool
	TokensSent                       sdk.Coins
	Address                          string
}

func (sf *WrappeMsgCreateConcentratedPool) String() string {
	var tokensSent []string
	if !(len(sf.TokensSent) == 0) {
		for _, v := range sf.TokensSent {
			tokensSent = append(tokensSent, v.String())
		}
	}
	return fmt.Sprintf("MsgCreateConcentratedPool: %s created pool and spent %s",
		sf.Address, strings.Join(tokensSent, ", "))
}

func (sf *WrappeMsgCreateConcentratedPool) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCreateConcentratedPool = msg.(*clPoolTypes.MsgCreateConcentratedPool)

	coinSpentEvents := txModule.GetEventsWithType("coin_spent", log)
	if len(coinSpentEvents) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	senderCoinsSpentStrings := txModule.GetCoinsSpent(sf.OsmosisMsgCreateConcentratedPool.Sender, coinSpentEvents)

	for _, coinSpentString := range senderCoinsSpentStrings {
		if coinSpentString != "" {
			coinsSpent, err := sdk.ParseCoinsNormalized(coinSpentString)
			if err != nil {
				return errors.New("error parsing coins received from event")
			}

			sf.TokensSent = append(sf.TokensSent, coinsSpent...)
		}
	}

	sf.Address = sf.OsmosisMsgCreateConcentratedPool.Sender

	return nil
}

func (sf *WrappeMsgCreateConcentratedPool) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)

	for _, token := range sf.TokensSent {
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

type WrapperMsgCollectIncentives struct {
	txModule.Message
	OsmosisMsgCollectIncentives *clTypes.MsgCollectIncentives
	TokensRecv                  sdk.Coins
	Address                     string
}

func (sf *WrapperMsgCollectIncentives) String() string {
	var tokensRecv []string
	if !(len(sf.TokensRecv) == 0) {
		for _, v := range sf.TokensRecv {
			tokensRecv = append(tokensRecv, v.String())
		}
	}
	return fmt.Sprintf("MsgCollectIncentives: %s collected %s",
		sf.Address, strings.Join(tokensRecv, ", "))
}

func (sf *WrapperMsgCollectIncentives) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgCollectIncentives = msg.(*clTypes.MsgCollectIncentives)

	totalCollectIncentivesEvent := txModule.GetEventsWithType("total_collect_incentives", log)
	if len(totalCollectIncentivesEvent) == 0 {
		return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
	}

	for _, collectIncentivesEvent := range totalCollectIncentivesEvent {
		for _, attribute := range collectIncentivesEvent.Attributes {
			if attribute.Key == tokensOutEvent {
				coinsReceived, err := sdk.ParseCoinsNormalized(attribute.Value)
				if err != nil {
					return errors.New("error parsing coins received from incentives event")
				}

				sf.TokensRecv = append(sf.TokensRecv, coinsReceived...)
			}
		}
	}

	sf.Address = sf.OsmosisMsgCollectIncentives.Sender

	return nil
}

func (sf *WrapperMsgCollectIncentives) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)

	for _, token := range sf.TokensRecv {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountReceived:       token.Amount.BigInt(),
				DenominationReceived: token.Denom,
				SenderAddress:        sf.Address,
			})
		}
	}

	return relevantData
}

type WrapperMsgAddToPosition struct {
	txModule.Message
	OsmosisMsgAddToPosition *clTypes.MsgAddToPosition
	TokensRecv              sdk.Coins
	TokensSent              sdk.Coins
	Address                 string
}

func (sf *WrapperMsgAddToPosition) String() string {
	var tokensRecv []string
	var tokensRecvString string

	if !(len(sf.TokensRecv) == 0) {
		for _, v := range sf.TokensRecv {
			tokensRecv = append(tokensRecv, v.String())
		}

		tokensRecvString = strings.Join(tokensRecv, ", ") + " in rewards"
	} else {
		tokensRecvString = "no rewards"
	}

	var tokensSent []string
	if !(len(sf.TokensSent) == 0) {
		for _, v := range sf.TokensSent {
			tokensSent = append(tokensSent, v.String())
		}
	}
	return fmt.Sprintf("MsgAddToPosition: %s collected %s and spent (%s)",
		sf.Address, tokensRecvString, strings.Join(tokensSent, ", "))
}

func (sf *WrapperMsgAddToPosition) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgAddToPosition = msg.(*clTypes.MsgAddToPosition)

	// Collect spread rewards
	spreadRewardsEvents := txModule.GetEventsWithType("collect_spread_rewards", log)

	for _, spreadRewardsEvent := range spreadRewardsEvents {
		for _, attribute := range spreadRewardsEvent.Attributes {
			if attribute.Key == tokensOutEvent {
				coinsReceived, err := sdk.ParseCoinsNormalized(attribute.Value)
				if err != nil {
					return errors.New("error parsing coins received from spread rewards event")
				}
				sf.TokensRecv = append(sf.TokensRecv, coinsReceived...)
			}
		}
	}

	// Collect incentives
	incentivesEvents := txModule.GetEventsWithType("collect_incentives", log)

	for _, incentivesEvent := range incentivesEvents {
		for _, attribute := range incentivesEvent.Attributes {
			if attribute.Key == tokensOutEvent {
				coinsReceived, err := sdk.ParseCoinsNormalized(attribute.Value)
				if err != nil {
					return errors.New("error parsing coins received from incentives event")
				}
				sf.TokensRecv = append(sf.TokensRecv, coinsReceived...)
			}
		}
	}

	// Collect coins spent
	coinSpentEvents := txModule.GetEventsWithType("coin_spent", log)
	senderCoinsSpentStrings := txModule.GetCoinsSpent(sf.OsmosisMsgAddToPosition.Sender, coinSpentEvents)

	for _, coinSpentString := range senderCoinsSpentStrings {
		if coinSpentString != "" {
			coinsSpent, err := sdk.ParseCoinsNormalized(coinSpentString)
			if err != nil {
				return errors.New("error parsing coins spent from event")
			}
			sf.TokensSent = append(sf.TokensSent, coinsSpent...)
		}
	}

	sf.Address = sf.OsmosisMsgAddToPosition.Sender

	return nil
}

func (sf *WrapperMsgAddToPosition) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)

	for _, token := range sf.TokensRecv {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountReceived:       token.Amount.BigInt(),
				DenominationReceived: token.Denom,
				SenderAddress:        sf.Address,
			})
		}
	}

	for _, token := range sf.TokensSent {
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
