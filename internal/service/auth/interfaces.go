package auth_service

import (
	"context"

	"github.com/dgrijalva/jwt-go"
	entities "github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

//go:generate mockgen -destination=mocks/mock_token_manager.go -package=mock_auth_service . TokenManager
//go:generate mockgen -destination=mocks/mock_token_provider.go -package=mock_auth_service . TokenProvider
//go:generate mockgen -destination=mocks/mock_usecases.go -package=mock_auth_service . CheckPermissionUseCase

type TokenManager interface {
	GenerateAccessToken(
		ctx context.Context,
		user *models.User,
	) (*entities.AccessTokenClaims, string, error)

	GenerateRefreshToken(
		ctx context.Context,
		user *models.User,
	) (*entities.TokenClaims, string, error)

	ParseToken(ctx context.Context, accessToken string) (*entities.TokenClaims, error)
	ParseAccessToken(ctx context.Context, accessToken string) (*entities.AccessTokenClaims, error)
}

type TokenProvider interface {
	NewWithClaims(method jwt.SigningMethod, claims jwt.Claims) *jwt.Token
	ParseWithClaims(tokenString string, claims jwt.Claims, keyFunc jwt.Keyfunc) (*jwt.Token, error)
}

type CheckPermissionUseCase interface {
	Check(ctx context.Context, roleCode, permissionCode string) (bool, error)
	CheckAny(ctx context.Context, roleCode string, permissionCodes []string) (bool, error)
}
