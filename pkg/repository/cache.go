package repository

import (
	"context"
	"encoding/json"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/redis/go-redis/v9"
)

const (
	blocksChannel            = "pub/blocks"
	maxTransactionsCacheSize = 50
	transactionsKey          = "latest_transactions"
)

type TransactionsCache interface {
	AddTransaction(ctx context.Context, transaction *model.Tx) error
	GetTransactions(ctx context.Context) ([]*model.Tx, error)
}

type BlocksCache interface {
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

func (s *Cache) AddTransaction(ctx context.Context, transaction *model.Tx) error {
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

func (s *Cache) GetTransactions(ctx context.Context) ([]*model.Tx, error) {
	res, err := s.rdb.LRange(ctx, transactionsKey, 0, 50).Result()
	if err != nil {
		return nil, err
	}

	var transactions []*model.Tx
	for _, r := range res {
		var tx model.Tx
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
