package apps_service

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type CreateApp interface {
	Create(ctx context.Context, data *entities.CreateAppData) (int, error)
}

type GetApp interface {
	Get(ctx context.Context, appId int) (*models.App, error)
}

type DeleteApp interface {
	Delete(ctx context.Context, userId int) error
}

type CreateApiKey interface {
	Create(ctx context.Context, data *entities.CreateApiKeyData) (*models.ApiKeyWithKey, error)
}

type RevokeApiKey interface {
	Revoke(ctx context.Context, apiKeyId int) error
}

type CheckApiKey interface {
	Check(ctx context.Context, keyHash string) (int, error)
}

type GetAppApiKeys interface {
	GetPaginated(
		ctx context.Context,
		p *entities.PaginationData,
		appId int,
	) (*entities.PaginationResult[models.ApiKey], error)
}
