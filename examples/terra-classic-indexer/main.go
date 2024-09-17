package main

import (
	"log"

	terraModules "github.com/DefiantLabs/cosmos-indexer-modules/terra-classic"
	"github.com/DefiantLabs/cosmos-indexer/cmd"
)

func main() {
	indexer := cmd.GetBuiltinIndexer()

	indexer.RegisterCustomMsgTypesByTypeURLs(terraModules.GetTerraClassicTypeMap())

	err := cmd.Execute()
	if err != nil {
		log.Fatalf("Failed to execute. Err: %v", err)
	}
}
