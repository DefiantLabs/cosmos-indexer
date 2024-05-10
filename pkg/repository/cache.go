package repository

import (
	"context"
	"encoding/json"
	"github.com/DefiantLabs/cosmos-indexer/db/models"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/redis/go-redis/v9"
)

const (
	blocksChannel            = "pub/blocks"
	maxTransactionsCacheSize = 50
	maxBlocksCacheSize       = 50
	transactionsKey          = "c/latest_transactions"
	blocksKey                = "c/latest_blocks"
)

type TransactionsCache interface {
	AddTransaction(ctx context.Context, transaction *models.Tx) error
	GetTransactions(ctx context.Context, start, stop int64) ([]*models.Tx, error)
}

type BlocksCache interface {
	AddBlock(ctx context.Context, info *model.BlockInfo) error
	GetBlocks(ctx context.Context, start, stop int64) ([]*model.BlockInfo, error)
	PublishBlock(ctx context.Context, info *model.BlockInfo) error
}

type Cache struct {
	rdb *redis.Client
}

func NewCache(rdb *redis.Client) *Cache {
	return &Cache{
		rdb: rdb,
	}
}

func (s *Cache) AddTransaction(ctx context.Context, transaction *models.Tx) error {
	res, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	if err := s.rdb.LPush(ctx, transactionsKey, string(res)).Err(); err != nil {
		return err
	}

	if err := s.rdb.LTrim(ctx, transactionsKey, 0, maxTransactionsCacheSize).Err(); err != nil {
		return err
	}

	return nil
}

func (s *Cache) GetTransactions(ctx context.Context, start, stop int64) ([]*models.Tx, error) {
	if stop > maxTransactionsCacheSize {
		stop = maxTransactionsCacheSize
	}
	res, err := s.rdb.LRange(ctx, transactionsKey, start, stop).Result()
	if err != nil {
		return nil, err
	}

	var transactions []*models.Tx
	for _, r := range res {
		var tx models.Tx
		if err := json.Unmarshal([]byte(r), &tx); err != nil {
			return nil, err
		}
		transactions = append(transactions, &tx)
	}

	return transactions, nil
}

func (s *Cache) PublishBlock(ctx context.Context, info *model.BlockInfo) error {
	res, err := json.Marshal(&info)
	if err != nil {
		return err
	}

	return s.rdb.Publish(ctx, blocksChannel, res).Err()
}

func (s *Cache) AddBlock(ctx context.Context, info *model.BlockInfo) error {
	res, err := json.Marshal(info)
	if err != nil {
		return err
	}

	if err := s.rdb.LPush(ctx, blocksKey, string(res)).Err(); err != nil {
		return err
	}

	if err := s.rdb.LTrim(ctx, blocksKey, 0, maxBlocksCacheSize).Err(); err != nil {
		return err
	}

	return nil
}

func (s *Cache) GetBlocks(ctx context.Context, start, stop int64) ([]*model.BlockInfo, error) {
	if stop > maxBlocksCacheSize {
		stop = maxBlocksCacheSize
	}

	res, err := s.rdb.LRange(ctx, blocksKey, start, stop).Result()
	if err != nil {
		return nil, err
	}

	var blcs []*model.BlockInfo
	for _, r := range res {
		var tx model.BlockInfo
		if err := json.Unmarshal([]byte(r), &tx); err != nil {
			return nil, err
		}
		blcs = append(blcs, &tx)
	}

	return blcs, nil
}
