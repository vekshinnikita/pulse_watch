package alert_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CreateAlertsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateAlertsUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *CreateAlertsUseCase {
	return &CreateAlertsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CreateAlertsUseCase) Create(
	ctx context.Context,
	data []dtos.CreateAlert,
) ([]int, error) {
	return uc.repo.Alert.CreateAlerts(ctx, data)
}
