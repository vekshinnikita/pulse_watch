package logs_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type SendLogsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewSendLogsUseCase(trm repository.TransactionManager, repo *repository.Repository) *SendLogsUseCase {
	return &SendLogsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *SendLogsUseCase) Send(ctx context.Context, logs entities.AppLogs) error {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return fmt.Errorf("send logs unable to get app_id")
	}

	options := &dtos.MessageOptions{
		Headers: []kafka.Header{
			{
				Key:   "app_id",
				Value: []byte(strconv.Itoa(appId)),
			},
		},
		Topic: constants.LogsTopic,
		Value: logs,
	}

	err := s.repo.Producer.Publish(ctx, options)
	if err != nil {
		return fmt.Errorf("send log publish: %w", err)
	}

	slog.InfoContext(ctx, "logs sent to the queue")

	return nil
}
