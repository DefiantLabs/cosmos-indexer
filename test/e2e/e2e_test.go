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
	setupDone       bool
	testDBConf      *testUtils.TestDockerDBConfig
	testIndexerConf *testUtils.TestDockerIndexerConfig
	testDB          *gorm.DB
	testDBHost      string
	testDBPort      string
	testDBName      string
	testDBUser      string
	testDBPass      string
	cleanDB         func()
	cleanIndexer    func()
	suiteChain      ibc.Chain
	interchain      *interchaintest.Interchain
}

func (suite *E2ETest) SetupSuite1() {

}

func (suite *E2ETest) SetupSuite() {

	// TearDown is never called if the setup func fails, which can be triggered through suite.Require().NoError(err)
	// So we defer the teardown here to ensure it is called but only if the bool for setup is false
	defer func() {
		if !suite.setupDone {
			suite.TearDownSuite()
		}
	}()

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

	// Setup the test database
	dbConf, err := testUtils.SetupTestDatabase(dockerNetwork)
	suite.Require().NoError(err)

	suite.testDBConf = dbConf
	suite.testDB = dbConf.GormDB
	suite.testDBHost = dbConf.DockerResourceName // we attach the indexer to the docker network and use the resource name as the host
	suite.testDBPort = dbConf.Port
	suite.testDBName = dbConf.Database
	suite.testDBUser = dbConf.User
	suite.testDBPass = dbConf.Password

	suite.cleanDB = dbConf.Clean

	// Setup the test indexer
	indexerConf, err := testUtils.SetupTestIndexer(dockerNetwork)
	suite.Require().NoError(err)

	suite.testIndexerConf = indexerConf
	suite.cleanIndexer = indexerConf.Clean

	// KEEP THIS UNDER ALL ERROR CHECKS
	suite.setupDone = true

}

func (suite *E2ETest) TearDownSuite() {
	// Setup the test database
	if suite.cleanDB != nil {
		suite.cleanDB()
	}
	if suite.cleanIndexer != nil {
		suite.cleanIndexer()
	}
}

func (suite *E2ETest) TestE2E() {
	baseConfig := suite.getSuiteBaseConf()

	// err := baseConfig.Validate()
	// suite.Require().NoError(err)

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
		Base: config.IndexBase{
			ThrottlingBase: config.ThrottlingBase{
				Throttling: 0,
			},
			RetryBase: config.RetryBase{
				RequestRetryAttempts: 0,
				RequestRetryMaxWait:  0,
			},
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
