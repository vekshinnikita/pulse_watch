package apps_usecases

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type CreateApiKeyUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateApiKeyUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *CreateApiKeyUseCase {
	return &CreateApiKeyUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CreateApiKeyUseCase) generateKey() (string, error) {
	b := make([]byte, 32) // 256 бит
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}

func (uc *CreateApiKeyUseCase) Create(
	ctx context.Context,
	data *entities.CreateApiKeyData,
) (*models.ApiKeyWithKey, error) {
	securityConfig := config.GetSecurityConfig()

	key, err := uc.generateKey()
	if err != nil {
		return nil, fmt.Errorf("create api key usecase generate key: %w", err)
	}

	dto := &dtos.CreateApiKeyData{
		Name:      data.Name,
		AppId:     data.AppId,
		ExpiresAt: data.ExpiresAt,
		CreatedAt: time.Now(),
		KeyHash:   utils.HashStringWithSalt(key, securityConfig.TokenSecretKey),
	}
	id, err := uc.repo.App.CreateApiKey(ctx, dto)
	if err != nil {
		return nil, fmt.Errorf("create api key usecase create api key: %w", err)
	}

	return &models.ApiKeyWithKey{
		Id:        id,
		Name:      dto.Name,
		Key:       key,
		ExpiresAt: dto.ExpiresAt,
		CreatedAt: dto.CreatedAt,
		Revoked:   false,
	}, nil
}
