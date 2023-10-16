package staking

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"

	stdTypes "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakeTypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

const (
	MsgDelegate        = "/cosmos.staking.v1beta1.MsgDelegate"
	MsgUndelegate      = "/cosmos.staking.v1beta1.MsgUndelegate"
	MsgBeginRedelegate = "/cosmos.staking.v1beta1.MsgBeginRedelegate"
	MsgCreateValidator = "/cosmos.staking.v1beta1.MsgCreateValidator" // An explicitly ignored msg for tx parsing purposes
	MsgEditValidator   = "/cosmos.staking.v1beta1.MsgEditValidator"   // An explicitly ignored msg for tx parsing purposes
)

type WrapperMsgDelegate struct {
	txModule.Message
	CosmosMsgDelegate     *stakeTypes.MsgDelegate
	DelegatorAddress      string
	AutoWithdrawalReward  *stdTypes.Coin
	AutoWithdrawalRewards stdTypes.Coins
}

type WrapperMsgUndelegate struct {
	txModule.Message
	CosmosMsgUndelegate   *stakeTypes.MsgUndelegate
	DelegatorAddress      string
	AutoWithdrawalReward  *stdTypes.Coin
	AutoWithdrawalRewards stdTypes.Coins
}

type WrapperMsgBeginRedelegate struct {
	txModule.Message
	CosmosMsgBeginRedelegate *stakeTypes.MsgBeginRedelegate
	DelegatorAddress         string
	AutoWithdrawalRewards    stdTypes.Coins
}

// HandleMsg: Handle type checking for MsgFundCommunityPool
func (sf *WrapperMsgDelegate) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgDelegate = msg.(*stakeTypes.MsgDelegate)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	delegatorReceivedCoinsEvt := txModule.GetEventWithType(bankTypes.EventTypeTransfer, log)
	if delegatorReceivedCoinsEvt == nil {
		sf.AutoWithdrawalReward = nil
		sf.DelegatorAddress = sf.CosmosMsgDelegate.DelegatorAddress
	} else {
		delegatorAddress, err := txModule.GetValueForAttribute("recipient", delegatorReceivedCoinsEvt)
		if err != nil {
			return err
		}

		sf.DelegatorAddress = delegatorAddress

		coinsReceived, err := txModule.GetValueForAttribute("amount", delegatorReceivedCoinsEvt)
		if err != nil {
			return err
		}

		coin, err := stdTypes.ParseCoinNormalized(coinsReceived)
		if err != nil {
			sf.AutoWithdrawalRewards, err = stdTypes.ParseCoinsNormalized(coinsReceived)
			if err != nil {
				fmt.Println("Error parsing coins normalized")
				fmt.Println(err)
				return err
			}
			return nil
		}
		sf.AutoWithdrawalReward = &coin
	}

	return nil
}

func (sf *WrapperMsgUndelegate) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgUndelegate = msg.(*stakeTypes.MsgUndelegate)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	sf.DelegatorAddress = sf.CosmosMsgUndelegate.DelegatorAddress
	delegatorReceivedCoinsEvt := txModule.GetEventWithType(bankTypes.EventTypeCoinReceived, log)
	if delegatorReceivedCoinsEvt == nil {
		sf.AutoWithdrawalReward = nil
		sf.DelegatorAddress = sf.CosmosMsgUndelegate.DelegatorAddress
	} else {
		var receivers []string
		var amounts []string

		// Pair off amounts and receivers in order
		for _, v := range delegatorReceivedCoinsEvt.Attributes {
			if v.Key == "receiver" {
				receivers = append(receivers, v.Value)
			} else if v.Key == "amount" {
				amounts = append(amounts, v.Value)
			}
		}

		// Find delegator address in receivers if its there, find its paired amount and set as the withdrawn rewards
		for i, v := range receivers {
			if v == sf.DelegatorAddress {
				coin, err := stdTypes.ParseCoinNormalized(amounts[i])
				if err != nil {
					var coins stdTypes.Coins
					coins, err = stdTypes.ParseCoinsNormalized(amounts[i])
					if err != nil {
						fmt.Println("Error parsing coins normalized")
						fmt.Println(err)
						return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
					}
					sf.AutoWithdrawalRewards = append(sf.AutoWithdrawalRewards, coins...)
					continue
				}
				sf.AutoWithdrawalReward = &coin
			}
		}
	}

	return nil
}

// HandleMsg: Handle type checking for MsgFundCommunityPool
func (sf *WrapperMsgBeginRedelegate) HandleMsg(msgType string, msg stdTypes.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.CosmosMsgBeginRedelegate = msg.(*stakeTypes.MsgBeginRedelegate)

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	delegatorReceivedCoinsEvt := txModule.GetEventWithType(bankTypes.EventTypeCoinReceived, log)
	sf.DelegatorAddress = sf.CosmosMsgBeginRedelegate.DelegatorAddress
	if delegatorReceivedCoinsEvt == nil {
		sf.AutoWithdrawalRewards = make(stdTypes.Coins, 0)
	} else {
		var receivers []string
		var amounts []string

		// Pair off amounts and receivers in order
		for _, v := range delegatorReceivedCoinsEvt.Attributes {
			if v.Key == "receiver" {
				receivers = append(receivers, v.Value)
			} else if v.Key == "amount" {
				amounts = append(amounts, v.Value)
			}
		}

		// Find delegator address in receivers if its there, find its paired amount and set as the withdrawn rewards
		// We use a cosmos.Coins array type for redelegations as redelegating could force withdrawal from both validators
		for i, v := range receivers {
			if v == sf.DelegatorAddress {
				coin, err := stdTypes.ParseCoinNormalized(amounts[i])
				if err != nil {
					var coins stdTypes.Coins
					coins, err = stdTypes.ParseCoinsNormalized(amounts[i])
					if err != nil {
						fmt.Println("Error parsing coins normalized")
						fmt.Println(err)
						return &txModule.MessageLogFormatError{MessageType: msgType, Log: fmt.Sprintf("%+v", log)}
					}
					sf.AutoWithdrawalRewards = append(sf.AutoWithdrawalRewards, coins...)
					continue
				}
				sf.AutoWithdrawalRewards = append(sf.AutoWithdrawalRewards, coin)
			}
		}
	}

	return nil
}

func (sf *WrapperMsgDelegate) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	var relevantData []parsingTypes.MessageRelevantInformation
	if sf.AutoWithdrawalReward != nil {
		data := parsingTypes.MessageRelevantInformation{}
		data.AmountReceived = sf.AutoWithdrawalReward.Amount.BigInt()
		data.DenominationReceived = sf.AutoWithdrawalReward.Denom
		data.ReceiverAddress = sf.DelegatorAddress
		relevantData = append(relevantData, data)
	} else if len(sf.AutoWithdrawalRewards) > 0 {
		for _, coin := range sf.AutoWithdrawalRewards {
			data := parsingTypes.MessageRelevantInformation{}
			data.AmountReceived = coin.Amount.BigInt()
			data.DenominationReceived = coin.Denom
			data.ReceiverAddress = sf.DelegatorAddress
			relevantData = append(relevantData, data)
		}
	}
	return relevantData
}

func (sf *WrapperMsgUndelegate) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	var relevantData []parsingTypes.MessageRelevantInformation
	if sf.AutoWithdrawalReward != nil {
		data := parsingTypes.MessageRelevantInformation{}
		data.AmountReceived = sf.AutoWithdrawalReward.Amount.BigInt()
		data.DenominationReceived = sf.AutoWithdrawalReward.Denom
		data.ReceiverAddress = sf.DelegatorAddress
		relevantData = append(relevantData, data)
	} else if len(sf.AutoWithdrawalRewards) > 0 {
		for _, coin := range sf.AutoWithdrawalRewards {
			data := parsingTypes.MessageRelevantInformation{}
			data.AmountReceived = coin.Amount.BigInt()
			data.DenominationReceived = coin.Denom
			data.ReceiverAddress = sf.DelegatorAddress
			relevantData = append(relevantData, data)
		}
	}
	return relevantData
}

func (sf *WrapperMsgBeginRedelegate) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	var relevantData []parsingTypes.MessageRelevantInformation
	for _, coin := range sf.AutoWithdrawalRewards {
		data := parsingTypes.MessageRelevantInformation{}
		data.AmountReceived = coin.Amount.BigInt()
		data.DenominationReceived = coin.Denom
		data.ReceiverAddress = sf.DelegatorAddress
		relevantData = append(relevantData, data)
	}
	return relevantData
}

func (sf *WrapperMsgDelegate) String() string {
	if sf.AutoWithdrawalReward != nil {
		return fmt.Sprintf("MsgDelegate: Delegator %s auto-withdrew %s", sf.DelegatorAddress, sf.AutoWithdrawalReward)
	}
	if len(sf.AutoWithdrawalRewards) > 0 {
		var coinsRecievedStrings []string
		for _, coin := range sf.AutoWithdrawalRewards {
			coinsRecievedStrings = append(coinsRecievedStrings, coin.String())
			return fmt.Sprintf("MsgDelegate: Delegator %s auto-withdrew %s", sf.DelegatorAddress, strings.Join(coinsRecievedStrings, ", "))
		}
	}
	return fmt.Sprintf("MsgDelegate: Delegator %s did not auto-withdrawal rewards", sf.DelegatorAddress)
}

func (sf *WrapperMsgUndelegate) String() string {
	if sf.AutoWithdrawalReward != nil {
		return fmt.Sprintf("MsgUndelegate: Delegator %s auto-withdrew %s", sf.DelegatorAddress, sf.AutoWithdrawalReward)
	}
	if len(sf.AutoWithdrawalRewards) > 0 {
		var coinsRecievedStrings []string
		for _, coin := range sf.AutoWithdrawalRewards {
			coinsRecievedStrings = append(coinsRecievedStrings, coin.String())
			return fmt.Sprintf("MsgUndelegate: Delegator %s auto-withdrew %s", sf.DelegatorAddress, strings.Join(coinsRecievedStrings, ", "))
		}
	}
	return fmt.Sprintf("MsgUndelegate: Delegator %s did not auto-withdrawal rewards", sf.DelegatorAddress)
}

func (sf *WrapperMsgBeginRedelegate) String() string {
	var coinsRecievedStrings []string
	for _, coin := range sf.AutoWithdrawalRewards {
		coinsRecievedStrings = append(coinsRecievedStrings, coin.String())
	}

	if len(coinsRecievedStrings) > 0 {
		return fmt.Sprintf("MsgBeginRedelegate: Delegator %s auto-withdrew %s", sf.DelegatorAddress, strings.Join(coinsRecievedStrings, ", "))
	}
	return fmt.Sprintf("MsgBeginRedelegate: Delegator %s did not auto-withdrawal rewards", sf.DelegatorAddress)
}
