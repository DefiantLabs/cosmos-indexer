package config

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

type ConfigTestSuite struct {
	suite.Suite
}

func (suite *ConfigTestSuite) TestValidateDatabaseConf() {
	conf := Database{
		Host:     "",
		Port:     "",
		Database: "",
		User:     "",
		Password: "",
	}

	err := validateDatabaseConf(conf)
	suite.Require().Error(err)
	conf.Host = "fake-host"

	err = validateDatabaseConf(conf)
	suite.Require().Error(err)

	conf.Port = "5432"
	err = validateDatabaseConf(conf)
	suite.Require().Error(err)

	conf.Database = "fake-database"
	err = validateDatabaseConf(conf)
	suite.Require().Error(err)

	conf.User = "fake-user"
	err = validateDatabaseConf(conf)
	suite.Require().Error(err)

	conf.Password = "fake-password"
	err = validateDatabaseConf(conf)
	suite.Require().NoError(err)
}

func (suite *ConfigTestSuite) TestValidateProbeConf() {
	conf := Probe{
		RPC:           "",
		AccountPrefix: "",
		ChainID:       "",
		ChainName:     "",
	}

	_, err := validateProbeConf(conf)
	suite.Require().Error(err)

	conf.RPC = "fake-rpc"
	_, err = validateProbeConf(conf)
	suite.Require().Error(err)

	conf.AccountPrefix = "fake-account-prefix"
	_, err = validateProbeConf(conf)
	suite.Require().Error(err)

	conf.ChainID = "fake-chain-id"
	_, err = validateProbeConf(conf)
	suite.Require().Error(err)

	conf.ChainName = "fake-chain-name"
	_, err = validateProbeConf(conf)
	suite.Require().NoError(err)
}

func (suite *ConfigTestSuite) TestValidateThrottlingConf() {
	conf := throttlingBase{
		Throttling: -1,
	}

	err := validateThrottlingConf(conf)
	suite.Require().Error(err)

	conf.Throttling = 0.5
	err = validateThrottlingConf(conf)
	suite.Require().NoError(err)
}

func TestConfigSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
