package util

import (
	"math/big"

	"github.com/shopspring/decimal"
)

func ToNumeric(i *big.Int) decimal.Decimal {
	num := decimal.NewFromBigInt(i, 0)
	return num
}

// StrNotSet will return true if the string value provided is empty
func StrNotSet(value string) bool {
	return len(value) == 0
}

func RemoveDuplicatesFromUint64Slice(sliceList []uint64) []uint64 {
	allKeys := make(map[uint64]bool)
	list := []uint64{}
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
