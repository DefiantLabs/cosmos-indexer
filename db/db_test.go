package db_test

import (
	"testing"
	"time"

	dbTypes "github.com/DefiantLabs/cosmos-indexer/db"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	testUtils "github.com/DefiantLabs/cosmos-indexer/test/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// TODO: Optimize tests to use a single database instance, clean database after each test, and teardown database after all tests are done

type DBTestSuite struct {
	suite.Suite
	db    *gorm.DB
	clean func()
}

func (suite *DBTestSuite) SetupTest() {
	conf, err := testUtils.SetupTestDatabase("")
	suite.Require().NoError(err)

	suite.db = conf.GormDB
	suite.clean = conf.Clean
}

func (suite *DBTestSuite) TearDownTest() {
	if suite.clean != nil {
		suite.clean()
	}

	suite.db = nil
	suite.clean = nil
}

func (suite *DBTestSuite) TestMigrateModels() {
	err := dbTypes.MigrateModels(suite.db)
	suite.Require().NoError(err)
}

func (suite *DBTestSuite) TestGetDBChainID() {
	err := dbTypes.MigrateModels(suite.db)
	suite.Require().NoError(err)

	initChain := models.Chain{
		ChainID: "testchain-1",
	}

	err = suite.db.Create(&initChain).Error
	suite.Require().NoError(err)

	chainID, err := dbTypes.GetDBChainID(suite.db, initChain)
	suite.Require().NoError(err)
	suite.Assert().NotZero(chainID)
}

func createMockBlock(mockDb *gorm.DB, chain models.Chain, address models.Address, height int64, txIndexed bool, eventIndexed bool) (models.Block, error) {
	block := models.Block{
		Chain:               chain,
		Height:              height,
		TimeStamp:           time.Now(),
		TxIndexed:           txIndexed,
		BlockEventsIndexed:  eventIndexed,
		ProposerConsAddress: address,
	}

	err := mockDb.Create(&block).Error
	return block, err
}

func (suite *DBTestSuite) TestGetHighestBlockFunctions() {
	err := dbTypes.MigrateModels(suite.db)
	suite.Require().NoError(err)

	initChain := models.Chain{
		ChainID: "testchain-1",
	}

	err = suite.db.Create(&initChain).Error
	suite.Require().NoError(err)

	initConsAddress := models.Address{
		Address: "testchainaddress",
	}

	err = suite.db.Create(&initConsAddress).Error
	suite.Require().NoError(err)

	block1, err := createMockBlock(suite.db, initChain, initConsAddress, 1, true, true)
	suite.Require().NoError(err)

	txBlock := dbTypes.GetHighestIndexedBlock(suite.db, initChain.ID)
	eventBlock, err := dbTypes.GetHighestEventIndexedBlock(suite.db, initChain.ID)
	suite.Require().NoError(err)

	suite.Assert().Equal(block1.Height, txBlock.Height)
	suite.Assert().Equal(block1.Height, eventBlock.Height)

	_, err = createMockBlock(suite.db, initChain, initConsAddress, 2, false, false)
	suite.Require().NoError(err)

	txBlock = dbTypes.GetHighestIndexedBlock(suite.db, initChain.ID)
	eventBlock, err = dbTypes.GetHighestEventIndexedBlock(suite.db, initChain.ID)
	suite.Require().NoError(err)

	suite.Assert().Equal(block1.Height, txBlock.Height)
	suite.Assert().Equal(block1.Height, eventBlock.Height)

	block3, err := createMockBlock(suite.db, initChain, initConsAddress, 3, true, true)
	suite.Require().NoError(err)

	txBlock = dbTypes.GetHighestIndexedBlock(suite.db, initChain.ID)
	eventBlock, err = dbTypes.GetHighestEventIndexedBlock(suite.db, initChain.ID)
	suite.Require().NoError(err)

	suite.Assert().Equal(block3.Height, txBlock.Height)
	suite.Assert().Equal(block3.Height, eventBlock.Height)
}

func TestDBSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}
