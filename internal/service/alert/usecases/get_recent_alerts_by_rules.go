package alert_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetRecentAlertsByRulesUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetRecentAlertsByRulesUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetRecentAlertsByRulesUseCase {
	return &GetRecentAlertsByRulesUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetRecentAlertsByRulesUseCase) Get(
	ctx context.Context,
	data *dtos.GetRecentAlerts,
) ([]models.Alert, error) {
	return uc.repo.Alert.GetRecentAlertsByRules(ctx, data)
}
