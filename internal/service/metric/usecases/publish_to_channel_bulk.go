package metric_usecases

import (
	"context"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type PublishToChannelBulkUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewPublishToChannelBulkUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *PublishToChannelBulkUseCase {
	return &PublishToChannelBulkUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *PublishToChannelBulkUseCase) Publish(
	ctx context.Context,
	data []dtos.PublishToChannel,
) error {

	err := uc.trm.Do(ctx, func(ctx context.Context) error {
		for _, item := range data {
			err := uc.repo.AnalyticsRedis.PublishToChannel(ctx, item.ChannelId, item.Data)
			if err != nil {
				return fmt.Errorf("publish to channel: %w", err)
			}
		}

		return nil
	})

	return err
}
