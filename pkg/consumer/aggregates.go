package consumer

import (
	"context"
	"fmt"
	"github.com/nodersteam/cosmos-indexer/pkg/model"
	"github.com/nodersteam/cosmos-indexer/pkg/repository"
	"github.com/rs/zerolog/log"
	"time"
)

type aggregatesConsumer struct {
	totals repository.TotalsCache
	blocks repository.Blocks
	txs    repository.Txs
}

func NewAggregatesConsumer(totals repository.TotalsCache, blocks repository.Blocks, txs repository.Txs) *aggregatesConsumer {
	return &aggregatesConsumer{totals: totals, blocks: blocks, txs: txs}
}

func (s *aggregatesConsumer) Consume(ctx context.Context) error {
	log.Info().Msg("starting aggregates consumer")
	t := time.NewTicker(5 * time.Second)

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-t.C:
			err := s.storeAggregated(ctx)
			if err != nil {
				log.Error().Err(err).Msg("failed to store aggregated data in consumer")
			}
		}
	}

	return fmt.Errorf("aggregatesConsumer failed")
}

func (s *aggregatesConsumer) storeAggregated(ctx context.Context) error {
	blocksTotal, err := s.blocks.TotalBlocks(ctx, time.Now().UTC())
	if err != nil {
		log.Err(err).Msg("failed to fetch total blocks")
		return err
	}

	var res model.TotalTransactions
	res.Total, res.Total24H, res.Total48H, res.Total30D, err = s.txs.TransactionsPerPeriod(ctx, time.Now().UTC())
	if err != nil {
		log.Err(err).Msg("failed to fetch transactions per period")
		return err
	}

	res.Volume24H, res.Volume30D, err = s.txs.VolumePerPeriod(ctx, time.Now().UTC())
	if err != nil {
		log.Err(err).Msg("failed to fetch transactions volume per period")
		return err
	}

	wallets, err := s.txs.GetWalletsCount(ctx)
	if err != nil {
		log.Err(err).Msg("failed to fetch GetWalletsCount")
		return err
	}

	info := &model.AggregatedInfo{
		UpdatedAt:    time.Now().UTC(),
		Blocks:       *blocksTotal,
		Transactions: res,
		Wallets:      *wallets,
	}

	return s.totals.AddTotals(ctx, info)
}
