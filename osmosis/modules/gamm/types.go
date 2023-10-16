package gamm

import (
	"math/big"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	EventTypeTransfer    = "transfer"
	EventTypeClaim       = "claim"
	EventAttributeAmount = "amount"
)

// Same as WrapperMsgExitPool but with different handlers.
// This is due to the Osmosis SDK emitting different Events (chain upgrades).

type tokenSwap struct {
	TokenSwappedIn  sdk.Coin
	TokenSwappedOut sdk.Coin
}

type coinReceived struct {
	sender       string
	coinReceived sdk.Coin
}

type ArbitrageTx struct {
	TokenIn   sdk.Coin
	TokenOut  sdk.Coin
	BlockTime time.Time
}

func calcNthGams(totalGamms *big.Int, numSwaps int) (*big.Int, *big.Int) {
	// figure out how many gamms per token
	var nthGamms big.Int
	nthGamms.Div(totalGamms, big.NewInt(int64(numSwaps)))

	// figure out how many gamms will remain for the last swap
	var remainderGamms big.Int
	remainderGamms.Mod(totalGamms, &nthGamms)
	remainderGamms.Add(&nthGamms, &remainderGamms)
	return &nthGamms, &remainderGamms
}
