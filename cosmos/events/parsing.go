package events

import "math/big"

type EventRelevantInformation struct {
	Address      string
	Amount       *big.Int
	Denomination string
	EventSource  uint
}
