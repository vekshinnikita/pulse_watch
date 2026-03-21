package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type DeleteAppUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewDeleteAppUseCase(trm repository.TransactionManager, repo *repository.Repository) *DeleteAppUseCase {
	return &DeleteAppUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *DeleteAppUseCase) Delete(ctx context.Context, appId int) error {
	return uc.repo.App.DeleteApp(ctx, appId)
}
