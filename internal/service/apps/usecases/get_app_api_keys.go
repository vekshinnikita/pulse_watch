package apps_usecases

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type GetAppApiKeysUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetAppApiKeysUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetAppApiKeysUseCase {
	return &GetAppApiKeysUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetAppApiKeysUseCase) GetPaginated(
	ctx context.Context,
	p *entities.PaginationData,
	appId int,
) (*entities.PaginationResult[models.ApiKey], error) {
	return uc.repo.App.GetAppApiKeysPaginated(ctx, p, appId)
}
