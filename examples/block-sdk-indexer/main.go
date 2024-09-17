package main

import (
	"log"

	blockSDKModules "github.com/DefiantLabs/cosmos-indexer-modules/block-sdk"
	"github.com/DefiantLabs/cosmos-indexer/cmd"
)

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.RegisterCustomMsgTypesByTypeURLs(blockSDKModules.GetBlockSDKTypeMap())

	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
