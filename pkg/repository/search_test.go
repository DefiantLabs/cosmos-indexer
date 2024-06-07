package repository

import (
	"context"
	"fmt"
	testdb "github.com/nodersteam/cosmos-indexer/pkg/repository/test_db"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"testing"
	"time"
)

type SearchRepositorySuite struct {
	suite.Suite
	repository *search
	db         *mongo.Database
	now        time.Time
	cleanups   []func()
}

func (suite *SearchRepositorySuite) SetupSuite() {
	db, cancel, err := testdb.NewMongoDB()
	if err != nil {
		suite.Failf("Error creating ClickHouse Client", fmt.Sprintf("%v", err))
	}
	suite.db = db
	suite.cleanups = append(suite.cleanups, cancel)
	repo := NewSearch(suite.db)
	suite.repository = repo
}

func (suite *SearchRepositorySuite) TestCreateBlock() {
	ctx := context.Background()

	var err error

	err = suite.repository.AddHash(ctx, "Qd74FlZyrasFfT97l3KLoEiKAu4bb7zPwQ30N+ZrYbI=",
		"block", 778391)
	suite.Assert().NoError(err)

	result, err := suite.repository.BlockByHeight(ctx, 91)
	suite.Assert().NoError(err)
	suite.Assert().Len(result, 1)
	suite.Assert().Equal(result[0].Type, "block")
	suite.Assert().Equal(result[0].TxHash, "Qd74FlZyrasFfT97l3KLoEiKAu4bb7zPwQ30N+ZrYbI=")
	suite.Assert().Equal(result[0].BlockHeight, "778391")

	err = suite.repository.AddHash(ctx, "4581DF321FB0CABF43E876032061058F497C4B5F81FDACC1F851768C35294E7B+ZrYbI=",
		"transaction", 778392)
	suite.Assert().NoError(err)

	result, err = suite.repository.HashByText(ctx, "Qd74Fl")
	suite.Assert().NoError(err)
	suite.Assert().Len(result, 1)

	suite.Equal(result[0].TxHash, "Qd74FlZyrasFfT97l3KLoEiKAu4bb7zPwQ30N+ZrYbI=")
	suite.Equal(result[0].Type, "block")

	// lower case
	result, err = suite.repository.HashByText(ctx, "qd74fl")
	suite.Assert().Len(result, 1)
	suite.Assert().NoError(err)
	suite.Equal(result[0].TxHash, "Qd74FlZyrasFfT97l3KLoEiKAu4bb7zPwQ30N+ZrYbI=")
	suite.Equal(result[0].Type, "block")

	result, err = suite.repository.HashByText(ctx, "4581DF32")
	suite.Assert().Len(result, 1)
	suite.Assert().NoError(err)
	suite.Equal(result[0].TxHash, "4581DF321FB0CABF43E876032061058F497C4B5F81FDACC1F851768C35294E7B+ZrYbI=")
	suite.Equal(result[0].Type, "transaction")

	result, err = suite.repository.HashByText(ctx, "123")
	suite.Assert().Len(result, 0)

	// multiple results
	err = suite.repository.AddHash(ctx, "Qd74FlZyrasFfT97l3KLoEiKAu4bb7zPwQ30N+ZrYbI=", "block", 778393)
	suite.Assert().NoError(err)

	err = suite.repository.AddHash(ctx, "4581DF321FB0CQd74FlZyrABF43E876032061058F497C4B5F81FDA", "transaction", 778392)
	suite.Assert().NoError(err)
	result, err = suite.repository.HashByText(ctx, "qd74fl")
	suite.Assert().Len(result, 2)
	suite.Assert().NoError(err)
}

func (suite *SearchRepositorySuite) TearDownSuite() {
	for i := range suite.cleanups {
		suite.cleanups[i]()
	}
}

func TestSearchRepositorySuite(t *testing.T) {
	suite.Run(t, new(SearchRepositorySuite))
}
