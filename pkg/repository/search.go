package repository

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const searchCollection = "search"

type Search interface {
	AddHash(ctx context.Context, hash string, hashType string, blockHeight int64) error
	HashByText(ctx context.Context, text string) ([]model.SearchResult, error)
	BlockByHeight(ctx context.Context, blockHeight int64) ([]model.SearchResult, error)
}

type search struct {
	pool *mongo.Database
}

func NewSearch(pool *mongo.Database) *search {
	return &search{pool: pool}
}

func (a *search) AddHash(ctx context.Context, hash string, hashType string, blockHeight int64) error {
	searchResult := model.SearchResult{TxHash: hash, Type: hashType, BlockHeight: fmt.Sprintf("%d", blockHeight)}

	filter := bson.D{primitive.E{Key: "tx_hash", Value: hash}, primitive.E{Key: "type", Value: hashType}}
	update := bson.D{{"$set",
		searchResult,
	}}

	upsert := true
	opts := options.UpdateOptions{Upsert: &upsert}

	res, err := a.pool.Collection(searchCollection).UpdateOne(ctx,
		filter, update, &opts)
	if err != nil {
		return err
	}
	log.Debug().Msgf("inserted new collection with ID %d", res)
	return nil
}

func (a *search) HashByText(ctx context.Context, text string) ([]model.SearchResult, error) {
	filter := bson.D{{"tx_hash", primitive.Regex{Pattern: text, Options: "i"}}}
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

func (a *search) BlockByHeight(ctx context.Context, blockHeight int64) ([]model.SearchResult, error) {
	filter := bson.D{
		{"block_height", primitive.Regex{Pattern: fmt.Sprintf("%d", blockHeight), Options: "i"}},
	}
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
