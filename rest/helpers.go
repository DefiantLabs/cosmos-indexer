package rest

import (
	"fmt"
	"os"
	"strconv"
)

func GetLatestBlockHeight(host string) uint64 {
	resp, err := GetLatestBlock(host)
	if err != nil {
		fmt.Println("Error getting latest block", err)
		os.Exit(1)
	}

	latestBlock, err := strconv.ParseUint(resp.Block.BlockHeader.Height, 10, 64)
	if err != nil {
		fmt.Println("Error getting latest block", err)
		os.Exit(1)
	}

	return latestBlock
}
