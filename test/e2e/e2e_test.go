package e2e_test

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/DefiantLabs/cosmos-indexer/config"
	testUtils "github.com/DefiantLabs/cosmos-indexer/test/utils"
	"github.com/strangelove-ventures/interchaintest/v8"
	"github.com/strangelove-ventures/interchaintest/v8/ibc"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type E2ETest struct {
	suite.Suite
	testDB     *gorm.DB
	testDBHost string
	testDBPort string
	testDBName string
	testDBUser string
	testDBPass string
	cleanDB    func()
	suiteChain ibc.Chain
	interchain *interchaintest.Interchain
}

func (suite *E2ETest) SetupSuite1() {

}

func (suite *E2ETest) SetupSuite() {
	// Setup the test database
	dbConf, err := testUtils.SetupTestDatabase()
	suite.Require().NoError(err)

	suite.testDB = dbConf.GormDB
	suite.testDBHost = dbConf.Host
	suite.testDBPort = dbConf.Port
	suite.testDBName = dbConf.Database
	suite.testDBUser = dbConf.User
	suite.testDBPass = dbConf.Password

	suite.cleanDB = dbConf.Clean

	var numVals int = 1
	var numNodes int = 0

	// Setup the test chain
	chainSpec := &interchaintest.ChainSpec{
		Name:          "gaia",
		ChainName:     "cosmoshub",
		Version:       "v14.1.0",
		NumValidators: &numVals,
		NumFullNodes:  &numNodes,
	}

	zapConfig := zap.Config{
		OutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "message",
			LevelKey:    "level",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
		},
		Encoding: "json",
		Level:    zap.NewAtomicLevelAt(zap.ErrorLevel),
	}

	zapLogger, err := zapConfig.Build()
	suite.Require().NoError(err)

	chainFactory := interchaintest.NewBuiltinChainFactory(zapLogger, []*interchaintest.ChainSpec{chainSpec})

	chains, err := chainFactory.Chains(suite.T().Name())
	suite.Require().NoError(err)

	suite.suiteChain = chains[0]

	suite.interchain = interchaintest.NewInterchain().AddChain(suite.suiteChain)

	dockerClient, dockerNetwork := interchaintest.DockerSetup(suite.T())
	dbDir := interchaintest.TempDir(suite.T())
	dbPath := filepath.Join(dbDir, "blocks.db")

	err = suite.interchain.Build(context.Background(), nil, interchaintest.InterchainBuildOptions{
		TestName:          suite.T().Name(),
		Client:            dockerClient,
		NetworkID:         dockerNetwork,
		SkipPathCreation:  false,
		BlockDatabaseFile: dbPath,
	})

	suite.Require().NoError(err)
}

func (suite *E2ETest) TearDownSuite() {
	// Setup the test database
	suite.cleanDB()
}

func (suite *E2ETest) TestE2E() {
	baseConfig := suite.getSuiteBaseConf()

	err := baseConfig.Validate()
	suite.Require().NoError(err)
	createAndStoreConfigToml(baseConfig, "./config.toml")
}

func (suite *E2ETest) getSuiteBaseConf() config.IndexConfig {
	return config.IndexConfig{
		Database: config.Database{
			Host:     suite.testDBHost,
			Port:     suite.testDBPort,
			Database: suite.testDBName,
			User:     suite.testDBUser,
			Password: suite.testDBPass,
			LogLevel: "silent",
		},
		Probe: config.Probe{
			RPC:           suite.suiteChain.GetHostRPCAddress(),
			AccountPrefix: suite.suiteChain.Config().Bech32Prefix,
			ChainID:       suite.suiteChain.Config().ChainID,
			ChainName:     suite.suiteChain.Config().Name,
		},
	}
}

func createAndStoreConfigToml(conf config.IndexConfig, path string) error {

	f, err := os.Create(path)
	if err != nil {
		// failed to create/open the file
		log.Fatal(err)
	}
	if err := toml.NewEncoder(f).Encode(conf); err != nil {
		// failed to encode
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		// failed to close the file
		log.Fatal(err)

	}
	return nil
}

func TestE2ETest(t *testing.T) {
	suite.Run(t, new(E2ETest))
}
