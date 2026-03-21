package logs_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CreateMetaVarsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateMetaVarsUseCase(trm repository.TransactionManager, repo *repository.Repository) *CreateMetaVarsUseCase {
	return &CreateMetaVarsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *CreateMetaVarsUseCase) Create(ctx context.Context, data []dtos.CreateMetaVar) ([]int, error) {
	return s.repo.Log.CreateMetaVars(ctx, data)
}
