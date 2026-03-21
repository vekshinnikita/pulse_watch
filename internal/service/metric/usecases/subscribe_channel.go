package metric_usecases

import (
	"context"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type SubscribeChannelUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewSubscribeChannelUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *SubscribeChannelUseCase {
	return &SubscribeChannelUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *SubscribeChannelUseCase) Subscribe(
	ctx context.Context,
	channelId string,
	handler func(channelId string, message string),
) error {

	pubsub, err := s.repo.AnalyticsRedis.SubscribePubSub(ctx, channelId)
	if err != nil {
		return fmt.Errorf("subscribe pubsub: %w", err)
	}

	ch := pubsub.Channel()
	go func() {
		defer pubsub.Close()

		for {
			select {
			case <-ctx.Done():
				return
			case m := <-ch:
				handler(m.Channel, m.Payload)
			}
		}
	}()

	return nil
}
