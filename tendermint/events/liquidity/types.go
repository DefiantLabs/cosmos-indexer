package liquidity

import (
	"fmt"
	"strings"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	tendermintEvents "github.com/DefiantLabs/cosmos-indexer/tendermint/events"
	abciTypes "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventAttributePoolID  = "pool_id"
	EventAttributeSuccess = "success"
)

type WrapperBlockEventDepositToPool struct {
	Event            abciTypes.Event
	Address          string
	AcceptedCoins    sdk.Coins
	PoolID           string
	Success          string
	PoolCoinReceived sdk.Coin
}

type WrapperBlockEventSwapTransacted struct {
	Event          abciTypes.Event
	CoinSwappedIn  sdk.Coin
	CoinSwappedOut sdk.Coin
	Fees           sdk.Coins
	Address        string
	PoolID         string
	Success        string
}

type WrapperBlockWithdrawFromPool struct {
	Event         abciTypes.Event
	Address       string
	PoolCoinSent  sdk.Coin
	WithdrawCoins sdk.Coins
	WithdrawFees  sdk.Coins
	PoolID        string
	Success       string
}

func (sf *WrapperBlockEventDepositToPool) GetType() string {
	return tendermintEvents.BlockEventDepositToPool
}

func (sf *WrapperBlockEventSwapTransacted) GetType() string {
	return tendermintEvents.BlockEventSwapTransacted
}

func (sf *WrapperBlockWithdrawFromPool) GetType() string {
	return tendermintEvents.BlockEventWithdrawFromPool
}

func (sf *WrapperBlockEventDepositToPool) HandleEvent(_ string, event abciTypes.Event) error {
	sf.Event = event
	var poolCoinAmount string
	var poolCoinDenom string
	for _, attribute := range event.Attributes {
		switch string(attribute.Key) {
		case "depositor":
			sf.Address = string(attribute.Value)
		case "accepted_coins":
			acceptedCoins, err := sdk.ParseCoinsNormalized(string(attribute.Value))
			if err != nil {
				return err
			}
			sf.AcceptedCoins = acceptedCoins
		case EventAttributeSuccess:
			sf.Success = string(attribute.Value)
		case EventAttributePoolID:
			sf.PoolID = string(attribute.Value)
		case "pool_coin_amount":
			poolCoinAmount = string(attribute.Value)
		case "pool_coin_denom":
			poolCoinDenom = string(attribute.Value)
		}
	}

	poolCoinReceived, err := sdk.ParseCoinNormalized(poolCoinAmount + poolCoinDenom)
	if err != nil {
		return err
	}

	sf.PoolCoinReceived = poolCoinReceived

	return nil
}

func (sf *WrapperBlockEventSwapTransacted) HandleEvent(eventType string, event abciTypes.Event) error {
	sf.Event = event

	// Swap transaction storage
	var offerCoinAmount string
	var offerCoinDenom string
	var demandCoinAmount string
	var demandCoinDenom string

	// Fee storage
	var offerCoinFeeAmount string
	var demandCoinFeeAmount string

	for _, attribute := range event.Attributes {
		switch string(attribute.Key) {
		case "swap_requester":
			sf.Address = string(attribute.Value)
		case "exchanged_offer_coin_amount":
			offerCoinAmount = string(attribute.Value)
		case "offer_coin_denom":
			offerCoinDenom = string(attribute.Value)
		case "exchanged_demand_coin_amount":
			demandCoinAmount = string(attribute.Value)
		case "demand_coin_denom":
			demandCoinDenom = string(attribute.Value)
		case "offer_coin_fee_amount":
			offerCoinFeeAmount = string(attribute.Value)
		case "exchanged_coin_fee_amount":
			demandCoinFeeAmount = string(attribute.Value)
		case EventAttributeSuccess:
			sf.Success = string(attribute.Value)
		case EventAttributePoolID:
			sf.PoolID = string(attribute.Value)
		}
	}

	offerAmount, ok := sdk.NewIntFromString(offerCoinAmount)
	if !ok {
		return fmt.Errorf("error parsing coin amount for offerCoinAmount %s", offerCoinAmount)
	}
	sf.CoinSwappedIn = sdk.NewCoin(offerCoinDenom, offerAmount)

	demandAmount, ok := sdk.NewIntFromString(demandCoinAmount)
	if !ok {
		return fmt.Errorf("error parsing coin amount for demandCoinAmount %s", demandCoinAmount)
	}
	sf.CoinSwappedOut = sdk.NewCoin(demandCoinDenom, demandAmount)

	// Here we are throwing out the decimal value. Why does it have a decimal in the first place and should we care?
	if strings.Contains(offerCoinFeeAmount, ".") {
		offerCoinFeeAmount = strings.Split(offerCoinFeeAmount, ".")[0]
	}
	offerFeeAmount, ok := sdk.NewIntFromString(offerCoinFeeAmount)
	if !ok {
		return fmt.Errorf("error parsing coin amount for offerCoinFeeAmount %s", offerCoinFeeAmount)
	}
	firstFee := sdk.NewCoin(offerCoinDenom, offerFeeAmount)

	// Here we are throwing out the decimal value. Why does it have a decimal in the first place and should we care?
	if strings.Contains(demandCoinFeeAmount, ".") {
		demandCoinFeeAmount = strings.Split(demandCoinFeeAmount, ".")[0]
	}
	demandFeeAmount, ok := sdk.NewIntFromString(demandCoinFeeAmount)
	if !ok {
		return fmt.Errorf("error parsing coin amount for demandCoinFeeAmount %s", demandCoinFeeAmount)
	}
	secondFee := sdk.NewCoin(demandCoinDenom, demandFeeAmount)

	sf.Fees = sdk.NewCoins(firstFee, secondFee)

	return nil
}

func (sf *WrapperBlockWithdrawFromPool) HandleEvent(eventType string, event abciTypes.Event) error {
	sf.Event = event

	var poolCoinAmount string
	var poolCoinDenom string
	var withdrawCoinsString string
	var withdrawFeesString string

	for _, attribute := range event.Attributes {
		switch string(attribute.Key) {
		case "withdrawer":
			sf.Address = string(attribute.Value)
		case "pool_coin_amount":
			poolCoinAmount = string(attribute.Value)
		case "pool_coin_denom":
			poolCoinDenom = string(attribute.Value)
		case "withdraw_coins":
			withdrawCoinsString = string(attribute.Value)
		case "withdraw_fee_coins":
			withdrawFeesString = string(attribute.Value)
		case EventAttributePoolID:
			sf.PoolID = string(attribute.Value)
		case EventAttributeSuccess:
			sf.Success = string(attribute.Value)
		}
	}

	poolCoin, err := sdk.ParseCoinNormalized(poolCoinAmount + poolCoinDenom)
	if err != nil {
		return err
	}
	sf.PoolCoinSent = poolCoin

	withdrawCoins, err := sdk.ParseCoinsNormalized(withdrawCoinsString)
	if err != nil {
		return err
	}
	sf.WithdrawCoins = withdrawCoins

	withdrawFees, err := sdk.ParseCoinsNormalized(withdrawFeesString)
	if err != nil {
		return err
	}
	sf.WithdrawFees = withdrawFees

	return nil
}

func (sf *WrapperBlockEventDepositToPool) ParseRelevantData() []events.EventRelevantInformation {
	relevantData := make([]events.EventRelevantInformation, len(sf.AcceptedCoins)+1)

	for i, coin := range sf.AcceptedCoins {
		relevantData[i] = events.EventRelevantInformation{
			EventSource:  dbTypes.TendermintLiquidityDepositCoinsToPool,
			Amount:       coin.Amount.BigInt(),
			Denomination: coin.Denom,
			Address:      sf.Address,
		}
	}

	relevantData[len(relevantData)-1] = events.EventRelevantInformation{
		EventSource:  dbTypes.TendermintLiquidityDepositPoolCoinReceived,
		Amount:       sf.PoolCoinReceived.Amount.BigInt(),
		Denomination: sf.PoolCoinReceived.Denom,
		Address:      sf.Address,
	}
	return relevantData
}

func (sf *WrapperBlockEventSwapTransacted) ParseRelevantData() []events.EventRelevantInformation {
	relevantData := make([]events.EventRelevantInformation, len(sf.Fees)+2)

	for i, coin := range sf.Fees {
		relevantData[i] = events.EventRelevantInformation{
			EventSource:  dbTypes.TendermintLiquiditySwapTransactedFee,
			Amount:       coin.Amount.BigInt(),
			Denomination: coin.Denom,
			Address:      sf.Address,
		}
	}

	relevantData[len(relevantData)-2] = events.EventRelevantInformation{
		EventSource:  dbTypes.TendermintLiquiditySwapTransactedCoinIn,
		Amount:       sf.CoinSwappedIn.Amount.BigInt(),
		Denomination: sf.CoinSwappedIn.Denom,
		Address:      sf.Address,
	}

	relevantData[len(relevantData)-1] = events.EventRelevantInformation{
		EventSource:  dbTypes.TendermintLiquiditySwapTransactedCoinOut,
		Amount:       sf.CoinSwappedOut.Amount.BigInt(),
		Denomination: sf.CoinSwappedOut.Denom,
		Address:      sf.Address,
	}

	return relevantData
}

func (sf *WrapperBlockWithdrawFromPool) ParseRelevantData() []events.EventRelevantInformation {
	relevantData := make([]events.EventRelevantInformation, len(sf.WithdrawFees)+len(sf.WithdrawCoins)+1)

	for i, coin := range sf.WithdrawFees {
		relevantData[i] = events.EventRelevantInformation{
			EventSource:  dbTypes.TendermintLiquidityWithdrawFee,
			Amount:       coin.Amount.BigInt(),
			Denomination: coin.Denom,
			Address:      sf.Address,
		}
	}

	i := len(sf.WithdrawFees)

	for _, coin := range sf.WithdrawCoins {
		relevantData[i] = events.EventRelevantInformation{
			EventSource:  dbTypes.TendermintLiquidityWithdrawCoinReceived,
			Amount:       coin.Amount.BigInt(),
			Denomination: coin.Denom,
			Address:      sf.Address,
		}
		i++
	}

	relevantData[len(relevantData)-1] = events.EventRelevantInformation{
		EventSource:  dbTypes.TendermintLiquidityWithdrawPoolCoinSent,
		Amount:       sf.PoolCoinSent.Amount.BigInt(),
		Denomination: sf.PoolCoinSent.Denom,
		Address:      sf.Address,
	}

	return relevantData
}

func (sf *WrapperBlockEventDepositToPool) String() string {
	return fmt.Sprintf("Tendermint Liquidity event %s: Address %s deposited %s into pool %s and received %s with status %s", sf.GetType(), sf.Address, sf.AcceptedCoins, sf.PoolID, sf.PoolCoinReceived, sf.Success)
}

func (sf *WrapperBlockEventSwapTransacted) String() string {
	return fmt.Sprintf("Tendermint Liquidity event %s: Address %s swapped %s into pool %s and received %s with status %s. Fees paid were %s", sf.GetType(), sf.Address, sf.CoinSwappedIn, sf.PoolID, sf.CoinSwappedOut, sf.Success, sf.Fees)
}

func (sf *WrapperBlockWithdrawFromPool) String() string {
	feesPaidString := "No fees were paid"
	if len(sf.WithdrawFees) != 0 {
		feesPaidString = fmt.Sprintf("Fees paid were %s", sf.WithdrawFees)
	}
	return fmt.Sprintf("Tendermint Liquidity event %s: Address %s sent %s into pool %s and received %s with status %s. %s.", sf.GetType(), sf.Address, sf.PoolCoinSent, sf.PoolID, sf.WithdrawCoins, sf.Success, feesPaidString)
}
