package service

import (
	"context"
	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
	pb "github.com/DefiantLabs/cosmos-indexer/proto"
	"time"
)

type Txs interface {
	ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error)
	GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error)
}

type txs struct {
	txRepo repository.Txs
}

func NewTxs(txRepo repository.Txs) *txs {
	return &txs{txRepo: txRepo}
}

func (s *txs) ChartTxByDay(ctx context.Context, from time.Time, to time.Time) ([]*model.TxsByDay, error) {
	return s.txRepo.ChartTxByDay(ctx, from, to)
}

func (s *txs) GetTxByHash(ctx context.Context, txHash string) (*pb.TxByHashResponse, error) {
	return nil, nil
}
