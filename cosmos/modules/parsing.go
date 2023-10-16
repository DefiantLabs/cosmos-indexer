package parsing

import "math/big"

type MessageRelevantInformation struct {
	SenderAddress        string
	ReceiverAddress      string
	AmountSent           *big.Int
	AmountReceived       *big.Int
	DenominationSent     string
	DenominationReceived string
}
