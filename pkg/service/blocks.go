package service

import (
	"context"
	"time"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
)

type Blocks interface {
	BlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error)
	BlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error)
	TotalBlocks(ctx context.Context, to time.Time) (*model.TotalBlocks, error)
	Blocks(ctx context.Context, limit int64, offset int64) ([]*model.BlockInfo, int64, error)
	BlockSignatures(ctx context.Context, height int64, limit int64, offset int64) ([]*model.BlockSigners, int64, error)
}

type blocks struct {
	blocksRepo repository.Blocks
}

func NewBlocks(blocksRepo repository.Blocks) *blocks {
	return &blocks{blocksRepo: blocksRepo}
}

func (s *blocks) BlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error) {
	return s.blocksRepo.GetBlockInfo(ctx, block, chainID)
}

func (s *blocks) BlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error) {
	return s.blocksRepo.GetBlockValidators(ctx, block, chainID)
}

func (s *blocks) TotalBlocks(ctx context.Context, to time.Time) (*model.TotalBlocks, error) {
	return s.blocksRepo.TotalBlocks(ctx, to)
}

func (s *blocks) Blocks(ctx context.Context, limit int64, offset int64) ([]*model.BlockInfo, int64, error) {
	return s.blocksRepo.Blocks(ctx, limit, offset)
}

func (s *blocks) BlockSignatures(ctx context.Context, height int64, limit int64, offset int64) ([]*model.BlockSigners, int64, error) {
	return s.blocksRepo.BlockSignatures(ctx, height, limit, offset)
}
