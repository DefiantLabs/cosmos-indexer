package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type IndexConfigTestSuite struct {
	suite.Suite
}

func (suite *IndexConfigTestSuite) TestIndexConfig() {
	conf := IndexConfig{
		// Setup valid configs for everything but base, these are tested elsewhere
		Database: Database{
			Host:     "fake-host",
			Port:     "5432",
			Database: "fake-database",
			User:     "fake-user",
			Password: "fake-password",
			LogLevel: "info",
		},
		Log: log{
			Level:  "info",
			Path:   "",
			Pretty: false,
		},
		Probe: Probe{
			RPC:           "fake-rpc",
			AccountPrefix: "cosmos",
			ChainID:       "fake-chain-id",
			ChainName:     "fake-chain-name",
		},
		Flags: flags{
			IndexTxMessageRaw: false,
		},
	}

	err := conf.Validate()
	suite.Require().Error(err)

	conf.Base.TransactionIndexingEnabled = true

	err = conf.Validate()
	suite.Require().Error(err)

	conf.Base.StartBlock = 1
	err = conf.Validate()
	suite.Require().Error(err)

	conf.Base.EndBlock = 2
	err = conf.Validate()
	suite.Require().NoError(err)
}

func (suite *IndexConfigTestSuite) TestCheckSuperfluousIndexKeys() {
	keys := []string{
		"fake-key",
	}
	validKeys := CheckSuperfluousIndexKeys(keys)
	suite.Require().Len(validKeys, 1)

	keys = append(keys, "base.start-block")

	validKeys = CheckSuperfluousIndexKeys(keys)
	suite.Require().Len(validKeys, 1)
}

func TestIndexConfig(t *testing.T) {
	suite.Run(t, new(IndexConfigTestSuite))
}
