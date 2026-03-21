package alert_usecases

import (
	"context"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CreateAndGetRuleUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateAndGetRuleUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *CreateAndGetRuleUseCase {
	return &CreateAndGetRuleUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CreateAndGetRuleUseCase) Create(
	ctx context.Context,
	data *entities.CreateAlertRule,
) (*models.AlertRule, error) {
	id, err := uc.repo.Alert.CreateAlertRule(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("create alert rule: %w", err)
	}

	rule, err := uc.repo.Alert.GetRule(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get alert rule: %w", err)
	}

	return rule, nil
}
