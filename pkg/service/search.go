package service

import (
	"context"
	"github.com/nodersteam/cosmos-indexer/pkg/model"
	"github.com/nodersteam/cosmos-indexer/pkg/repository"
)

type Search interface {
	SearchByText(ctx context.Context, text string) ([]model.SearchResult, error)
	SearchByBlock(ctx context.Context, height int64) ([]model.SearchResult, error)
}

type search struct {
	repo repository.Search
}

func NewSearch(repo repository.Search) *search {
	return &search{repo: repo}
}

func (s *search) SearchByText(ctx context.Context, text string) ([]model.SearchResult, error) {
	return s.repo.HashByText(ctx, text)
}

func (s *search) SearchByBlock(ctx context.Context, height int64) ([]model.SearchResult, error) {
	return s.repo.BlockByHeight(ctx, height)
}
