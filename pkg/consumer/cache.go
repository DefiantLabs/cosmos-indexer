package consumer

import (
	"context"

	"github.com/DefiantLabs/cosmos-indexer/pkg/model"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
	"github.com/rs/zerolog/log"
)

type cacheConsumer struct {
	blocksCh chan *model.BlockInfo
	blocks   repository.BlocksCache
}

func NewCacheConsumer(blocks repository.BlocksCache, blocksCh chan *model.BlockInfo) *cacheConsumer {
	return &cacheConsumer{blocks: blocks, blocksCh: blocksCh}
}

func (s *cacheConsumer) RunBlocks(ctx context.Context) error {
	log.Info().Msgf("Starting cache consumer: RunBlocks")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("breaking the worker loop.")
			break
		case newBlock, _ := <-s.blocksCh:
			err := s.blocks.PublishBlock(ctx, newBlock)
			if err != nil {
				log.Err(err).Msgf("Error publishing block")
			}
		}
	}
}
