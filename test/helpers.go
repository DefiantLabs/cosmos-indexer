package test

import (
	"fmt"
	"math/big"

	"github.com/BurntSushi/toml"
	"github.com/DefiantLabs/cosmos-indexer/config"
	"github.com/DefiantLabs/cosmos-indexer/osmosis"

	"github.com/DefiantLabs/cosmos-indexer/core"
	dbUtils "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/util"
	"gorm.io/gorm"
)

func createOsmosisTaxableEvent(db *gorm.DB, blockHeight int64) {
	addr := ensureTestAddress(db)
	simpleDenom := ensureTestDenom(db)
	chain := ensureTestChain(db, osmosis.ChainID, osmosis.Name)
	block := ensureTestBlock(db, chain, blockHeight)
	ensureOsmosisRewardsTaxableEvent(db, simpleDenom, addr, block, big.NewInt(420))
}

func setupOsmosisTestModels(db *gorm.DB) {
	addr := ensureTestAddress(db)
	simpleDenom := ensureTestDenom(db)
	chain := ensureTestChain(db, osmosis.ChainID, osmosis.Name)
	block := ensureTestBlock(db, chain, 1)

	ensureOsmosisRewardsTaxableEvent(db, simpleDenom, addr, block, big.NewInt(100))
}

func ensureOsmosisRewardsTaxableEvent(db *gorm.DB, denom dbUtils.Denom, addr dbUtils.Address, block dbUtils.Block, amount *big.Int) dbUtils.TaxableEvent {
	taxEvt := dbUtils.TaxableEvent{Source: dbUtils.OsmosisRewardDistribution, Amount: util.ToNumeric(amount), Denomination: denom, EventAddress: addr, Block: block}
	db.FirstOrCreate(&taxEvt, &taxEvt)
	return taxEvt
}

func ensureTestChain(db *gorm.DB, chainID string, name string) dbUtils.Chain {
	chain := dbUtils.Chain{ChainID: chainID, Name: name}
	db.FirstOrCreate(&chain, &chain)
	return chain
}

func ensureTestBlock(db *gorm.DB, chain dbUtils.Chain, height int64) dbUtils.Block {
	block := dbUtils.Block{Height: height, Chain: chain}
	db.FirstOrCreate(&block, &block)
	return block
}

func ensureTestDenom(db *gorm.DB) dbUtils.Denom {
	denom := "uosmo"
	simpleDenom := dbUtils.Denom{Base: denom, Symbol: denom}
	db.FirstOrCreate(&simpleDenom)
	return simpleDenom
}

func ensureTestAddress(db *gorm.DB) dbUtils.Address {
	address := "test1m2hg5t7n8f6kzh8kmh98phenk8a4xp5wyuz34y"
	addr := dbUtils.Address{Address: address}
	db.FirstOrCreate(&addr)
	return addr
}

func getConfig(configFileLocation string) (config.IndexConfig, error) {
	var conf config.IndexConfig
	_, err := toml.DecodeFile(configFileLocation, &conf)
	return conf, err
}

// setup does pre-run setup configurations.
//   - Loads the application config from config.tml, cli args and parses/merges
//   - Connects to the database and returns the db object
//   - Returns various values used throughout the application
func dbSetup(addressRegex string, addressPrefix string) (*gorm.DB, error) {
	config, err := getConfig("../config.toml")
	if err != nil {
		fmt.Println("Error opening configuration file", err)
		return nil, err
	}

	db, err := dbUtils.PostgresDbConnectLogInfo(config.Database.Host, config.Database.Port, config.Database.Database, config.Database.User, config.Database.Password)
	if err != nil {
		fmt.Println("Could not establish connection to the database", err)
		return nil, err
	}

	// Could potentially check Node info at startup and pass in ourselves?
	core.SetupAddressRegex(addressRegex)
	core.SetupAddressPrefix(addressPrefix)

	// run database migrations at every runtime
	err = dbUtils.MigrateModels(db)
	if err != nil {
		fmt.Println("Error running database migrations: ", err)
		return nil, err
	}

	dbUtils.CacheDenoms(db) // Have to cache denoms to get translations from e.g. ujuno to Juno
	dbUtils.CacheIBCDenoms(db)

	return db, nil
}
