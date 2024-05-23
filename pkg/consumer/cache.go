package consumer

import (
	"context"

	"github.com/nodersteam/cosmos-indexer/db/models"

	"github.com/nodersteam/cosmos-indexer/pkg/model"
	"github.com/nodersteam/cosmos-indexer/pkg/repository"
	"github.com/rs/zerolog/log"
)

type cacheConsumer struct {
	blocksCh chan *model.BlockInfo
	txCh     chan *models.Tx
	blocks   repository.BlocksCache
	txs      repository.TransactionsCache
}

func NewCacheConsumer(blocks repository.BlocksCache, blocksCh chan *model.BlockInfo,
	txCh chan *models.Tx, txs repository.TransactionsCache) *cacheConsumer {
	return &cacheConsumer{blocks: blocks, blocksCh: blocksCh, txCh: txCh, txs: txs}
}

func (s *cacheConsumer) RunBlocks(ctx context.Context) error {
	log.Info().Msgf("Starting cache consumer: RunBlocks")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("breaking the worker loop.")
			break
		case newBlock, _ := <-s.blocksCh:
			err := s.blocks.AddBlock(ctx, newBlock)
			if err != nil {
				log.Err(err).Msgf("Error publishing block")
			}
		}
	}
}

func (s *cacheConsumer) RunTransactions(ctx context.Context) error {
	log.Info().Msgf("Starting cache consumer: RunTransactions")
	for {
		select {
		case <-ctx.Done():
			log.Info().Msgf("breaking the worker loop.")
			break
		case newTx, _ := <-s.txCh:
			err := s.txs.AddTransaction(ctx, newTx)
			if err != nil {
				log.Err(err).Msgf("Error publishing block")
			}
		}
	}
}
