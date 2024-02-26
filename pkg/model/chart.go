package model

import "time"

type TxsByDay struct {
	TxNum int32
	Day   time.Time
}
