package service

import (
	"context"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
)

type Blocks interface {
	BlockInfo(ctx context.Context, block int32, chainID int32) (*model.BlockInfo, error)
	BlockValidators(ctx context.Context, block int32, chainID int32) ([]string, error)
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
	//return s.blocksRepo.GetBlockInfo(ctx, block, chainID)
	return []string{}, nil
}
