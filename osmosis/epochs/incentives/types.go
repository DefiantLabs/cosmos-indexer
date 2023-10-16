package incentives

import (
	"errors"
	"fmt"

	"github.com/DefiantLabs/cosmos-indexer/cosmos/events"
	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	osmosisEvents "github.com/DefiantLabs/cosmos-indexer/osmosis/events"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abciTypes "github.com/tendermint/tendermint/abci/types"
)

type WrapperBlockDistribution struct {
	Event           abciTypes.Event
	RewardsReceived sdk.Coins
	ReceiverAddress string
}

func (sf *WrapperBlockDistribution) GetType() string {
	return osmosisEvents.BlockEventDistribution
}

func (sf *WrapperBlockDistribution) HandleEvent(eventType string, event abciTypes.Event) error {
	var receiverAddr string
	var receiverAmount string

	for _, attr := range event.Attributes {
		if string(attr.Key) == "receiver" {
			receiverAddr = string(attr.Value)
		}
		if string(attr.Key) == "amount" {
			receiverAmount = string(attr.Value)
		}
	}

	if receiverAddr != "" && receiverAmount != "" {
		coins, err := sdk.ParseCoinsNormalized(receiverAmount)
		if err != nil {
			return err
		}
		sf.ReceiverAddress = receiverAddr
		sf.RewardsReceived = coins
	} else {
		return errors.New("rewards received or address were not present")
	}

	return nil
}

func (sf *WrapperBlockDistribution) ParseRelevantData() []events.EventRelevantInformation {
	relevantData := make([]events.EventRelevantInformation, len(sf.RewardsReceived))

	for i, coin := range sf.RewardsReceived {
		relevantData[i] = events.EventRelevantInformation{
			Address:      sf.ReceiverAddress,
			Amount:       coin.Amount.BigInt(),
			Denomination: coin.Denom,
			EventSource:  dbTypes.OsmosisRewardDistribution,
		}
	}

	return relevantData
}

func (sf *WrapperBlockDistribution) String() string {
	return fmt.Sprintf("Osmosis Incentives event %s: Address %s received %s rewards.", sf.GetType(), sf.ReceiverAddress, sf.RewardsReceived)
}
