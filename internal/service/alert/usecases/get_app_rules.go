package alert_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetAppRulesUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetAppRulesUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetAppRulesUseCase {
	return &GetAppRulesUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetAppRulesUseCase) Get(ctx context.Context, appId int) ([]models.AlertRule, error) {
	return uc.repo.Alert.GetAppRules(ctx, appId)
}
