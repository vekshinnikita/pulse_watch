package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetAppUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetAppUseCase(trm repository.TransactionManager, repo *repository.Repository) *GetAppUseCase {
	return &GetAppUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetAppUseCase) Get(ctx context.Context, appId int) (*models.App, error) {
	return uc.repo.App.GetApp(ctx, appId)
}
