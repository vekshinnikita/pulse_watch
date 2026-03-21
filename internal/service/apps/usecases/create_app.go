package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CreateAppUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateAppUseCase(trm repository.TransactionManager, repo *repository.Repository) *CreateAppUseCase {
	return &CreateAppUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CreateAppUseCase) Create(ctx context.Context, data *entities.CreateAppData) (int, error) {
	return uc.repo.App.CreateApp(ctx, data)
}
