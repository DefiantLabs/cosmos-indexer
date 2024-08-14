package main

import (
	"log"

	terraModules "github.com/DefiantLabs/cosmos-indexer-modules/terra-classic"
	"github.com/DefiantLabs/cosmos-indexer/cmd"
	"github.com/DefiantLabs/cosmos-indexer/indexer"
)

func includeTerraModules(setupData indexer.PostSetupCustomDataset) error {
	terraModules.IncludeMarketImplementations(setupData.ChainClient)
	terraModules.IncludeWASMImplementations(setupData.ChainClient)
	terraModules.IncludeOracleImplementations(setupData.ChainClient)

	return nil
}

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.PostSetupCustomFunction = includeTerraModules
	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
