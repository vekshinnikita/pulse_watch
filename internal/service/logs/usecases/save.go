package logs_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type SaveLogsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewSaveLogsUseCase(trm repository.TransactionManager, repo *repository.Repository) *SaveLogsUseCase {
	return &SaveLogsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *SaveLogsUseCase) Save(ctx context.Context, logs []entities.EnrichedAppLog) error {
	return s.repo.LogsES.BulkSave(ctx, logs)
}
