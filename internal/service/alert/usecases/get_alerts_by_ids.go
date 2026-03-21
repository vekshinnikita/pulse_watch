package alert_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetFullAlertsByIdsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetFullAlertsByIdsUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetFullAlertsByIdsUseCase {
	return &GetFullAlertsByIdsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetFullAlertsByIdsUseCase) Get(ctx context.Context, alertIds []int) ([]models.AlertFull, error) {
	return uc.repo.Alert.GetFullAlertsByIds(ctx, alertIds)
}
