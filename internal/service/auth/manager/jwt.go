package jwt_manager

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type TokenManagerOptions struct {
	SecretKey  string
	AccessTtl  time.Duration
	RefreshTtl time.Duration
}

type JWTTokenManager struct {
	secretKey  string
	accessTtl  time.Duration
	refreshTtl time.Duration
}

func NewDefaultJWTTokenManager() *JWTTokenManager {
	securityCfg := config.GetSecurityConfig()
	authCfg := GetConfig()

	return &JWTTokenManager{
		secretKey:  securityCfg.TokenSecretKey,
		accessTtl:  time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
		refreshTtl: 24 * time.Hour * time.Duration(authCfg.RefreshTokenTtlDays),
	}
}

func NewJWTTokenManager(options *TokenManagerOptions) *JWTTokenManager {
	return &JWTTokenManager{
		secretKey:  options.SecretKey,
		accessTtl:  options.AccessTtl,
		refreshTtl: options.RefreshTtl,
	}
}

func (m *JWTTokenManager) GenerateAccessToken(
	ctx context.Context,
	user *models.User,
) (*entities.AccessTokenClaims, string, error) {
	now := time.Now()
	claims := &entities.AccessTokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(m.accessTtl).Unix(),
			IssuedAt:  now.Unix(),
		},
		UserId: user.Id,
		Role:   &user.Role,
		JTI:    uuid.New().String(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(m.secretKey))

	if err != nil {
		return nil, "", fmt.Errorf("generate access token: %w", err)
	}

	return claims, token, err
}

func (m *JWTTokenManager) GenerateRefreshToken(
	ctx context.Context,
	user *models.User,
) (*entities.TokenClaims, string, error) {
	now := time.Now()
	claims := &entities.TokenClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(m.refreshTtl).Unix(),
			IssuedAt:  now.Unix(),
		},
		UserId: user.Id,
		JTI:    uuid.New().String(),
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).
		SignedString([]byte(m.secretKey))

	if err != nil {
		return nil, "", fmt.Errorf("generate refresh token: %w", err)
	}

	return claims, token, err
}

func (m *JWTTokenManager) ParseToken(ctx context.Context, inputToken string) (*entities.TokenClaims, error) {
	token, err := jwt.ParseWithClaims(inputToken, &entities.TokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing token method")
		}

		return []byte(m.secretKey), nil
	})
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok {
			// Токен истек
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, &errs.ExpiredOrRevokedTokenError{Message: errs.TokenExpiredErrorMessage}
			}

			return nil, &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage}
		}

		return nil, err
	}

	claims, ok := token.Claims.(*entities.TokenClaims)
	if !ok {

		return nil, errors.New("token claims are not of type *entities.UserClaims")
	}

	return claims, nil
}

func (m *JWTTokenManager) ParseAccessToken(ctx context.Context, inputToken string) (*entities.AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(inputToken, &entities.AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing token method")
		}

		return []byte(m.secretKey), nil
	})
	if err != nil {
		ve, ok := err.(*jwt.ValidationError)
		if ok {
			// Токен истек
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, &errs.ExpiredOrRevokedTokenError{Message: errs.TokenExpiredErrorMessage}
			}

			return nil, &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage}
		}

		return nil, err
	}

	claims, ok := token.Claims.(*entities.AccessTokenClaims)
	if !ok {

		return nil, errors.New("token claims are not of type *entities.UserClaims")
	}

	return claims, nil
}
