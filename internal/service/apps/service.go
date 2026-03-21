package apps_service

import (
	"context"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	apps_usecases "github.com/vekshinnikita/pulse_watch/internal/service/apps/usecases"
)

type AppsServiceUseCases struct {
	createApp     CreateApp
	getApp        GetApp
	deleteApp     DeleteApp
	createApiKey  CreateApiKey
	revokeApiKey  RevokeApiKey
	checkApiKey   CheckApiKey
	getAppApiKeys GetAppApiKeys
}

type AppsService struct {
	trm  repository.TransactionManager
	repo *repository.Repository
	uc   *AppsServiceUseCases
}

func NewDefaultAppsService(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *AppsService {
	return &AppsService{
		repo: repo,
		trm:  trm,
		uc: &AppsServiceUseCases{
			createApp:     apps_usecases.NewCreateAppUseCase(trm, repo),
			getApp:        apps_usecases.NewGetAppUseCase(trm, repo),
			deleteApp:     apps_usecases.NewDeleteAppUseCase(trm, repo),
			createApiKey:  apps_usecases.NewCreateApiKeyUseCase(trm, repo),
			revokeApiKey:  apps_usecases.NewRevokeApiKeyUseCase(trm, repo),
			checkApiKey:   apps_usecases.NewCheckApiKeyUseCase(trm, repo),
			getAppApiKeys: apps_usecases.NewGetAppApiKeysUseCase(trm, repo),
		},
	}
}

func (s *AppsService) CreateApp(ctx context.Context, data *entities.CreateAppData) (int, error) {
	return s.uc.createApp.Create(ctx, data)
}

func (s *AppsService) GetApp(ctx context.Context, appId int) (*models.App, error) {
	return s.uc.getApp.Get(ctx, appId)
}

func (s *AppsService) GetAppIds(
	ctx context.Context,
) ([]int, error) {
	return s.repo.App.GetAppIds(ctx)
}

func (s *AppsService) CreateAndGetApp(ctx context.Context, data *entities.CreateAppData) (*models.App, error) {

	var app *models.App
	err := s.trm.Do(ctx, func(ctx context.Context) error {
		appId, err := s.uc.createApp.Create(ctx, data)
		if err != nil {
			return fmt.Errorf("create and get app: %w", err)
		}

		app, err = s.uc.getApp.Get(ctx, appId)
		return err
	})
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (s *AppsService) DeleteApp(ctx context.Context, appId int) error {
	return s.uc.deleteApp.Delete(ctx, appId)
}

func (s *AppsService) CreateApiKey(
	ctx context.Context,
	data *entities.CreateApiKeyData,
) (*models.ApiKeyWithKey, error) {
	return s.uc.createApiKey.Create(ctx, data)
}

func (s *AppsService) GetAppApiKeysPaginated(
	ctx context.Context,
	p *entities.PaginationData,
	appId int,
) (*entities.PaginationResult[models.ApiKey], error) {
	return s.uc.getAppApiKeys.GetPaginated(ctx, p, appId)
}

func (s *AppsService) RevokeApiKey(ctx context.Context, apiKeyId int) error {
	return s.uc.revokeApiKey.Revoke(ctx, apiKeyId)
}

func (s *AppsService) CheckApiKey(ctx context.Context, apiKey string) (int, error) {
	return s.uc.checkApiKey.Check(ctx, apiKey)
}
