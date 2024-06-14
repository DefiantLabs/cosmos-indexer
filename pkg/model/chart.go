package model

import (
	"github.com/shopspring/decimal"
	"time"
)

type TxsByDay struct {
	TxNum int32
	Day   time.Time
}

type TxsByHour struct {
	TxNum int32
	Hour  time.Time
}

type TxByHourWithCount struct {
	Points   []*TxsByHour
	Total24H int64
	Total48H int64
}

type TxVolumeByHour struct {
	TxVolume decimal.Decimal
	Hour     time.Time
}
