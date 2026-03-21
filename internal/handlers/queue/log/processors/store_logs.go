package log_queue_processors

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type StoreLogsProcessor struct {
	services *service.Service
}

func NewStoreLogsProcessor(services *service.Service) *StoreLogsProcessor {
	return &StoreLogsProcessor{
		services: services,
	}
}

func (uc *StoreLogsProcessor) Process(ctx context.Context, logs []entities.EnrichedAppLog) error {
	err := uc.services.Logs.SaveLogs(ctx, logs)
	if err != nil {
		return fmt.Errorf("save logs: %w", err)
	}

	slog.InfoContext(ctx, "logs are saved in the storage")
	return nil
}
