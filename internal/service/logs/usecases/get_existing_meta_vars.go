package logs_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetExistingMetaVarsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetExistingMetaVarsUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetExistingMetaVarsUseCase {
	return &GetExistingMetaVarsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *GetExistingMetaVarsUseCase) Get(
	ctx context.Context,
	appId int,
	varCodes []string,
) ([]models.LogMetaVar, error) {
	return s.repo.Log.GetAppLogMetaVarsByCodes(ctx, appId, varCodes)
}
