package e2e_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/DefiantLabs/cosmos-indexer/config"
	testUtils "github.com/DefiantLabs/cosmos-indexer/test/utils"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
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
	// BUG: The ports here are currently the mapped ports. Since the indexer is running in-network, it needs to use the host + non-mapped port
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
	err := uploadConfigFileToContainer(baseConfig, suite.testIndexerConf.DockerPool, suite.testIndexerConf.DockerResourceName, "/go/src/app/")
	suite.Require().NoError(err)

	fmt.Println(baseConfig)
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
			RPC:           suite.suiteChain.GetRPCAddress(),
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

func uploadConfigFileToContainer(conf config.IndexConfig, pool *dockertest.Pool, dockerResourceName string, path string) error {
	gzipFile, err := encodeConfigTomlToTarGz(conf)
	if err != nil {
		return err
	}

	r, err := gzip.NewReader(&gzipFile)
	if err != nil {
		return err
	}

	uploadOptions := docker.UploadToContainerOptions{
		InputStream: r,
		Path:        path,
		Context:     context.Background(),
	}

	err = pool.Client.UploadToContainer(dockerResourceName, uploadOptions)

	return err
}

// Required by dockertest's UploadToContainer function which uses the docker archive upload API
func encodeConfigTomlToTarGz(conf config.IndexConfig) (bytes.Buffer, error) {

	var b bytes.Buffer
	if err := toml.NewEncoder(&b).Encode(conf); err != nil {
		// failed to encode
		return b, err
	}

	var tarB bytes.Buffer
	tw := tar.NewWriter(&tarB)
	tarHeader := &tar.Header{
		Name: "config.toml",
		Size: int64(b.Len()),
	}

	if err := tw.WriteHeader(tarHeader); err != nil {
		// failed to write header
		return b, err
	}

	if _, err := tw.Write(b.Bytes()); err != nil {
		// failed to write
		return b, err
	}

	tw.Close()

	var gzipB bytes.Buffer
	gz := gzip.NewWriter(&gzipB)
	if _, err := gz.Write(tarB.Bytes()); err != nil {
		// failed to write
		return b, err
	}

	gz.Close()

	return gzipB, nil
}

func TestE2ETest(t *testing.T) {
	suite.Run(t, new(E2ETest))
}
