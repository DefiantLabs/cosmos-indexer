package valsetpref

import (
	"fmt"
	"strings"

	parsingTypes "github.com/DefiantLabs/cosmos-indexer/cosmos/modules"
	txModule "github.com/DefiantLabs/cosmos-indexer/cosmos/modules/tx"
	"github.com/DefiantLabs/cosmos-indexer/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankTypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	valsetPrefTypes "github.com/osmosis-labs/osmosis/v19/x/valset-pref/types"
)

const (
	MsgSetValidatorSetPreference  = "/osmosis.valsetpref.v1beta1.MsgSetValidatorSetPreference"
	MsgDelegateToValidatorSet     = "/osmosis.valsetpref.v1beta1.MsgDelegateToValidatorSet"
	MsgUndelegateFromValidatorSet = "/osmosis.valsetpref.v1beta1.MsgUndelegateFromValidatorSet"
	MsgRedelegateValidatorSet     = "/osmosis.valsetpref.v1beta1.MsgRedelegateValidatorSet"
	MsgWithdrawDelegationRewards  = "/osmosis.valsetpref.v1beta1.MsgWithdrawDelegationRewards"
	// linter thinks this is a password
	//nolint:gosec
	MsgDelegateBondedTokens = "/osmosis.valsetpref.v1beta1.MsgDelegateBondedTokens"
)

// Set of common functions shared throughout
func getRewardsReceived(log *txModule.LogMessage, address string) (sdk.Coins, error) {
	receiveEvent := txModule.GetEventsWithType(bankTypes.EventTypeCoinReceived, log)

	var rewardCoins sdk.Coins
	if receiveEvent != nil {
		delegaterCoinsReceivedStrings := txModule.GetCoinsReceived(address, receiveEvent)

		for _, coinString := range delegaterCoinsReceivedStrings {
			coins, err := sdk.ParseCoinsNormalized(coinString)
			if err != nil {
				return nil, err
			}
			rewardCoins = append(rewardCoins, coins...)
		}
	}
	return rewardCoins, nil
}

func getRelevantData(rewards sdk.Coins, address string) []parsingTypes.MessageRelevantInformation {
	relevantData := make([]parsingTypes.MessageRelevantInformation, 0)

	for _, token := range rewards {
		if token.Amount.IsPositive() {
			relevantData = append(relevantData, parsingTypes.MessageRelevantInformation{
				AmountReceived:       token.Amount.BigInt(),
				DenominationReceived: token.Denom,
				SenderAddress:        address,
			})
		}
	}
	return relevantData
}

func getString(messageType string, rewards sdk.Coins, address string) string {
	var tokensSent []string
	if !(len(rewards) == 0) {
		for _, v := range rewards {
			tokensSent = append(tokensSent, v.String())
		}
		return fmt.Sprintf("%s: %s received rewards %s",
			messageType, address, strings.Join(tokensSent, ", "))
	}
	return fmt.Sprintf("%s: %s did not withdraw rewards",
		messageType, address)
}

type WrapperMsgDelegateToValidatorSet struct {
	txModule.Message
	OsmosisMsgDelegateToValidatorSet *valsetPrefTypes.MsgDelegateToValidatorSet
	DelegatorAddress                 string
	RewardsOut                       sdk.Coins
}

func (sf *WrapperMsgDelegateToValidatorSet) String() string {
	return getString("MsggDelegateToValidatorSet", sf.RewardsOut, sf.DelegatorAddress)
}

func (sf *WrapperMsgDelegateToValidatorSet) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgDelegateToValidatorSet = msg.(*valsetPrefTypes.MsgDelegateToValidatorSet)
	sf.DelegatorAddress = sf.OsmosisMsgDelegateToValidatorSet.Delegator

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	coins, err := getRewardsReceived(log, sf.DelegatorAddress)
	if err != nil {
		return err
	}

	sf.RewardsOut = coins

	return nil
}

func (sf *WrapperMsgDelegateToValidatorSet) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return getRelevantData(sf.RewardsOut, sf.DelegatorAddress)
}

type WrapperMsgUndelegateFromValidatorSet struct {
	txModule.Message
	OsmosisMsgUndelegateFromValidatorSet *valsetPrefTypes.MsgUndelegateFromValidatorSet
	DelegatorAddress                     string
	RewardsOut                           sdk.Coins
}

func (sf *WrapperMsgUndelegateFromValidatorSet) String() string {
	return getString("MsgUndelegateFromValidatorSet", sf.RewardsOut, sf.DelegatorAddress)
}

func (sf *WrapperMsgUndelegateFromValidatorSet) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgUndelegateFromValidatorSet = msg.(*valsetPrefTypes.MsgUndelegateFromValidatorSet)
	sf.DelegatorAddress = sf.OsmosisMsgUndelegateFromValidatorSet.Delegator

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	coins, err := getRewardsReceived(log, sf.DelegatorAddress)
	if err != nil {
		return err
	}

	sf.RewardsOut = coins

	return nil
}

func (sf *WrapperMsgUndelegateFromValidatorSet) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return getRelevantData(sf.RewardsOut, sf.DelegatorAddress)
}

type WrapperMsgRedelegateValidatorSet struct {
	txModule.Message
	OsmosisMsgRedelegateValidatorSet *valsetPrefTypes.MsgRedelegateValidatorSet
	DelegatorAddress                 string
	RewardsOut                       sdk.Coins
}

func (sf *WrapperMsgRedelegateValidatorSet) String() string {
	return getString("MsgRedelegateValidatorSet", sf.RewardsOut, sf.DelegatorAddress)
}

func (sf *WrapperMsgRedelegateValidatorSet) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgRedelegateValidatorSet = msg.(*valsetPrefTypes.MsgRedelegateValidatorSet)
	sf.DelegatorAddress = sf.OsmosisMsgRedelegateValidatorSet.Delegator

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	coins, err := getRewardsReceived(log, sf.DelegatorAddress)
	if err != nil {
		return err
	}

	sf.RewardsOut = coins

	return nil
}

func (sf *WrapperMsgRedelegateValidatorSet) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return getRelevantData(sf.RewardsOut, sf.DelegatorAddress)
}

type WrapperMsgWithdrawDelegationRewards struct {
	txModule.Message
	OsmosisMsgWithdrawDelegationRewards *valsetPrefTypes.MsgWithdrawDelegationRewards
	DelegatorAddress                    string
	RewardsOut                          sdk.Coins
}

func (sf *WrapperMsgWithdrawDelegationRewards) String() string {
	return getString("MsgWithdrawDelegationRewards", sf.RewardsOut, sf.DelegatorAddress)
}

func (sf *WrapperMsgWithdrawDelegationRewards) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgWithdrawDelegationRewards = msg.(*valsetPrefTypes.MsgWithdrawDelegationRewards)
	sf.DelegatorAddress = sf.OsmosisMsgWithdrawDelegationRewards.Delegator

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	coins, err := getRewardsReceived(log, sf.DelegatorAddress)
	if err != nil {
		return err
	}

	sf.RewardsOut = coins

	return nil
}

func (sf *WrapperMsgWithdrawDelegationRewards) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return getRelevantData(sf.RewardsOut, sf.DelegatorAddress)
}

type WrapperMsgDelegateBondedTokens struct {
	txModule.Message
	OsmosisMsgDelegateBondedTokens *valsetPrefTypes.MsgDelegateBondedTokens
	DelegatorAddress               string
	RewardsOut                     sdk.Coins
}

func (sf *WrapperMsgDelegateBondedTokens) String() string {
	return getString("MsgDelegateBondedTokens", sf.RewardsOut, sf.DelegatorAddress)
}

func (sf *WrapperMsgDelegateBondedTokens) HandleMsg(msgType string, msg sdk.Msg, log *txModule.LogMessage) error {
	sf.Type = msgType
	sf.OsmosisMsgDelegateBondedTokens = msg.(*valsetPrefTypes.MsgDelegateBondedTokens)
	sf.DelegatorAddress = sf.OsmosisMsgDelegateBondedTokens.Delegator

	// Confirm that the action listed in the message log matches the Message type
	validLog := txModule.IsMessageActionEquals(sf.GetType(), log)
	if !validLog {
		return util.ReturnInvalidLog(msgType, log)
	}

	// The attribute in the log message that shows you the delegator rewards auto-received
	coins, err := getRewardsReceived(log, sf.DelegatorAddress)
	if err != nil {
		return err
	}

	sf.RewardsOut = coins

	return nil
}

func (sf *WrapperMsgDelegateBondedTokens) ParseRelevantData() []parsingTypes.MessageRelevantInformation {
	return getRelevantData(sf.RewardsOut, sf.DelegatorAddress)
}
