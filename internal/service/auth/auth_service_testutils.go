package auth_service

import (
	"fmt"
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	mock_repository "github.com/vekshinnikita/pulse_watch/internal/repository/mocks"
	mock_auth_service "github.com/vekshinnikita/pulse_watch/internal/service/auth/mocks"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
	"go.uber.org/mock/gomock"
)

type mockAuthServiceDeps struct {
	trm       *mock_repository.MockTransactionManager
	auth_repo *mock_repository.MockAuthRepository
	tokenMng  *mock_auth_service.MockTokenManager
}

type mockBehavior func(s *mockAuthServiceDeps)

func NewMockAuthService(t *testing.T, mockBehavior mockBehavior) *AuthService {
	c := gomock.NewController(t)
	t.Cleanup(func() { c.Finish() })

	auth_repo := mock_repository.NewMockAuthRepository(c)
	auth_redis_repo := mock_repository.NewMockAuthRedisRepository(c)

	tokenMng := mock_auth_service.NewMockTokenManager(c)

	trManager := mock_repository.NewMockTransactionManager(c)

	mockBehavior(&mockAuthServiceDeps{
		auth_repo: auth_repo,
		trm:       trManager,
		tokenMng:  tokenMng,
	})
	return NewAuthService(&AuthServiceDeps{
		repo: &repository.Repository{
			Auth:      auth_repo,
			AuthRedis: auth_redis_repo,
		},
		trm:      trManager,
		tokenMng: tokenMng,
	})
}

func getTestToken(user *models.User, name string) (*entities.TokenClaims, string) {
	claims := &entities.TokenClaims{
		UserId: user.Id,
		JTI:    uuid.New().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: testutils.FixedTime.Unix(),
		},
	}

	token := fmt.Sprintf("%sToken", name)

	return claims, token
}

func getTestAccessToken(user *models.User, name string) (*entities.AccessTokenClaims, string) {
	claims := &entities.AccessTokenClaims{
		UserId: user.Id,
		Role:   &user.Role,
		JTI:    uuid.New().String(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: testutils.FixedTime.Unix(),
		},
	}

	token := fmt.Sprintf("%sToken", name)

	return claims, token
}
