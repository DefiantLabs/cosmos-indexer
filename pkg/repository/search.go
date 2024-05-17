package repository

import (
	"context"
	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const searchCollection = "search"

type Search interface {
	AddHash(ctx context.Context, hash string, hashType string) error
	HashByText(ctx context.Context, text string) ([]model.SearchResult, error)
}

type search struct {
	pool *mongo.Database
}

func NewSearch(pool *mongo.Database) *search {
	return &search{pool: pool}
}

func (a *search) AddHash(ctx context.Context, hash string, hashType string) error {
	res, err := a.pool.Collection(searchCollection).InsertOne(ctx,
		model.SearchResult{TxHash: hash, Type: hashType})
	if err != nil {
		return err
	}
	log.Debug().Msgf("inserted new collection with ID %d", res)
	return nil
}

func (a *search) HashByText(ctx context.Context, text string) ([]model.SearchResult, error) {
	regex := bson.M{"$regex": primitive.Regex{Pattern: "^" + text, Options: "i"}}
	filter := bson.M{"hash": regex}
	cursor, err := a.pool.Collection(searchCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	var dbResult []model.SearchResult
	if err = cursor.All(ctx, &dbResult); err != nil {
		return nil, err
	}
	return dbResult, nil
}
