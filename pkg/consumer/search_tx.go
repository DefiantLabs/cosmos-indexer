package consumer

import (
	"context"
	"encoding/json"
	"github.com/DefiantLabs/cosmos-indexer/db/models"
	"github.com/DefiantLabs/cosmos-indexer/pkg/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

type SearchTxConsumer interface {
	Consume(ctx context.Context) error
}

type searchTxPublisher struct {
	rdb   *redis.Client
	topic string
	repo  repository.Search
}

func NewSearchTxConsumer(rdb *redis.Client,
	blocksTopic string, repo repository.Search) *searchTxPublisher {
	return &searchTxPublisher{rdb: rdb, repo: repo, topic: blocksTopic}
}

func (s *searchTxPublisher) Consume(ctx context.Context) error {
	log.Info().Msgf("SearchTxConsumer started.")
	subscriber := s.rdb.Subscribe(ctx, s.topic)
	defer func() {
		err := subscriber.Unsubscribe(ctx, s.topic)
		if err != nil {
			log.Error().Err(err).Msg("unsubscribe pubsub")
		}
		err = subscriber.Close()
		if err != nil {
			log.Error().Err(err).Msg("close pubsub")
		}
	}()

	innerReceiver := make(chan models.Tx)
	defer close(innerReceiver)

	go func(inner chan models.Tx) {
		for {
			msg, err := subscriber.ReceiveMessage(ctx)
			if err != nil {
				log.Err(err).Msgf("error in subscriber.ReceiveMessage")
				break
			}

			var in models.Tx
			if err = json.Unmarshal([]byte(msg.Payload), &in); err != nil {
				log.Err(err).Msgf("error unmarshalling message")
				continue
			}

			inner <- in
		}
	}(innerReceiver)

	for {
		select {
		case <-ctx.Done():
			log.Debug().Msgf("breaking the worker loop.")
			return nil
		case newRecord := <-innerReceiver:
			s.repo.AddHash(ctx, newRecord.Hash, "transaction")
		}
	}
}
