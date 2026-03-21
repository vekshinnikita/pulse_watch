package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type RevokeApiKeyUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewRevokeApiKeyUseCase(trm repository.TransactionManager, repo *repository.Repository) *RevokeApiKeyUseCase {
	return &RevokeApiKeyUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *RevokeApiKeyUseCase) Revoke(ctx context.Context, apiKeyId int) error {
	return uc.repo.App.RevokeApiKey(ctx, apiKeyId)
}
