package logs_service

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type SendLogsUseCase interface {
	Send(ctx context.Context, logs entities.AppLogs) error
}

type SaveLogsUseCase interface {
	Save(ctx context.Context, logs []entities.EnrichedAppLog) error
}

type GetExistingMetaVars interface {
	Get(ctx context.Context, appId int, varCodes []string) ([]models.LogMetaVar, error)
}

type CreateMetaVars interface {
	Create(ctx context.Context, data []dtos.CreateMetaVar) ([]int, error)
}
