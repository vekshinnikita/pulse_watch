package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type CheckApiKeyUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCheckApiKeyUseCase(trm repository.TransactionManager, repo *repository.Repository) *CheckApiKeyUseCase {
	return &CheckApiKeyUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CheckApiKeyUseCase) Check(ctx context.Context, key string) (int, error) {
	securityConfig := config.GetSecurityConfig()
	keyHash := utils.HashStringWithSalt(key, securityConfig.TokenSecretKey)

	return uc.repo.App.CheckApiKey(ctx, keyHash)
}
