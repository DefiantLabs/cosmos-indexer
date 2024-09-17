package main

import (
	"log"

	blockSDKModules "github.com/DefiantLabs/cosmos-indexer-modules/block-sdk"
	"github.com/DefiantLabs/cosmos-indexer/cmd"
	"github.com/DefiantLabs/cosmos-indexer/indexer"
)

func includeBlockSDK(setupData indexer.PostSetupCustomDataset) error {
	blockSDKModules.IncludeAuctionImplementations(setupData.ChainClient)

	return nil
}

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.PostSetupCustomFunction = includeBlockSDK
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
