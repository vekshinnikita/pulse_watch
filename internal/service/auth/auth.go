package auth_service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	entities "github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	jwt_manager "github.com/vekshinnikita/pulse_watch/internal/service/auth/manager"
	auth_usecases "github.com/vekshinnikita/pulse_watch/internal/service/auth/usecases"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type authUseCases struct {
	checkPermission CheckPermissionUseCase
}

type AuthService struct {
	trm      repository.TransactionManager
	repo     *repository.Repository
	tokenMng TokenManager
	uc       *authUseCases
}

type AuthServiceDeps struct {
	repo     *repository.Repository
	trm      repository.TransactionManager
	tokenMng TokenManager
}

func NewAuthService(deps *AuthServiceDeps) *AuthService {
	return &AuthService{
		trm:      deps.trm,
		repo:     deps.repo,
		tokenMng: deps.tokenMng,
		uc: &authUseCases{
			checkPermission: auth_usecases.NewCheckPermissionUseCase(deps.trm, deps.repo),
		},
	}
}

func NewDefaultAuthService(trm repository.TransactionManager, repo *repository.Repository) *AuthService {
	return &AuthService{
		trm:      trm,
		repo:     repo,
		tokenMng: jwt_manager.NewDefaultJWTTokenManager(),
		uc: &authUseCases{
			checkPermission: auth_usecases.NewCheckPermissionUseCase(trm, repo),
		},
	}
}

func (s *AuthService) CreateUser(ctx context.Context, createUser *entities.SignUpUser) (int, error) {
	cfg := config.GetSecurityConfig()

	createUser.Password = utils.HashStringWithSalt(createUser.Password, cfg.SecretKey)

	userId, err := s.repo.Auth.CreateUser(ctx, createUser)
	if err == nil {
		slog.InfoContext(ctx, "user was created in db",
			slog.Int("user_id", userId),
		)
	}

	return userId, err
}

func (s *AuthService) GetUserById(ctx context.Context, userId int) (*models.User, error) {
	return s.repo.Auth.GetUserById(ctx, userId)
}

func (s *AuthService) CreateAndGetUser(
	ctx context.Context,
	createUser *entities.SignUpUser,
) (*models.User, error) {
	userId, err := s.CreateUser(ctx, createUser)
	if err != nil {
		return nil, fmt.Errorf("service create and get user: %w", err)
	}

	return s.GetUserById(ctx, userId)
}

func (s *AuthService) GenerateTokens(
	ctx context.Context,
	user *models.User,
) (*entities.AuthTokens, error) {
	_, accessToken, err := s.tokenMng.GenerateAccessToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service generate tokens generate access token: %w", err)
	}

	refreshClaims, refreshToken, err := s.tokenMng.GenerateRefreshToken(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("service generate tokens generate refresh token: %w", err)
	}

	// Сохраняем refresh token в БД
	refreshExpiresAt := time.Unix(refreshClaims.ExpiresAt, 0)
	_, err = s.repo.Auth.SaveRefreshToken(ctx, user.Id, refreshClaims.JTI, &refreshExpiresAt)
	if err != nil {
		return nil, fmt.Errorf("service generate tokens save refresh token: %w", err)
	}

	return &entities.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) SignIn(
	ctx context.Context,
	signInUser *entities.SignInUser,
) (*entities.AuthTokens, error) {
	securityCfg := config.GetSecurityConfig()

	passwordHash := utils.HashStringWithSalt(signInUser.Password, securityCfg.SecretKey)

	user, err := s.repo.Auth.GetUserByUsernameAndPassword(ctx, signInUser.Username, passwordHash)
	if err != nil {
		return nil, fmt.Errorf("service generate tokens get user by username and password: %w", err)
	}

	return s.GenerateTokens(ctx, user)
}

func (s *AuthService) ParseToken(
	ctx context.Context,
	accessToken string,
) (*entities.TokenClaims, error) {
	return s.tokenMng.ParseToken(ctx, accessToken)
}

func (s *AuthService) ParseAccessToken(
	ctx context.Context,
	accessToken string,
) (*entities.AccessTokenClaims, error) {
	return s.tokenMng.ParseAccessToken(ctx, accessToken)
}

func (s *AuthService) RefreshTokens(
	ctx context.Context,
	refreshToken string,
) (*entities.AuthTokens, error) {
	claims, err := s.tokenMng.ParseToken(ctx, refreshToken)
	if err != nil {
		return nil, err
	}

	// Проверяем наличие токена в БД
	exists, err := s.repo.Auth.IsRefreshTokenValid(ctx, claims.UserId, claims.JTI)
	if err != nil {
		return nil, fmt.Errorf("service refresh tokens check token: %w", err)
	}
	if !exists {
		return nil, &errs.ExpiredOrRevokedTokenError{Message: errs.TokenRevokedErrorMessage}
	}

	// Получение пользователя
	user, err := s.repo.Auth.GetUserById(ctx, claims.UserId)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	// Делаем обновление токена в транзакции
	var newTokens *entities.AuthTokens
	err = s.trm.Do(ctx, func(ctx context.Context) error {
		newTokens, err = s.GenerateTokens(ctx, user)
		if err != nil {
			return err
		}

		err = s.repo.Auth.RevokeRefreshToken(ctx, claims.JTI)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("service refresh tokens generate new tokens: %s", err)
	}

	return newTokens, nil
}

func (s *AuthService) GetCurrentUser(ctx context.Context) (*models.User, error) {
	userId, ok := ctx.Value(constants.UserIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("there isn't key '%s' in context", constants.UserIdCtxKey)
	}

	user, err := s.repo.Auth.GetUserById(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *AuthService) GetCurrentUserRole(ctx context.Context) (*models.Role, error) {
	role, ok := ctx.Value(constants.UserRoleCtxKey).(*models.Role)
	if !ok {
		return nil, fmt.Errorf("can't get value by key '%s' in context", constants.UserIdCtxKey)
	}

	return role, nil
}

func (s *AuthService) CheckRolePermission(
	ctx context.Context,
	roleCode string,
	permissionCode string,
) (bool, error) {
	return s.uc.checkPermission.Check(ctx, roleCode, permissionCode)
}

func (s *AuthService) CheckRoleAnyPermission(
	ctx context.Context,
	roleCode string,
	permissionCodes []string,
) (bool, error) {
	return s.uc.checkPermission.CheckAny(ctx, roleCode, permissionCodes)
}

func (s *AuthService) CheckCurrentUserPermission(
	ctx context.Context,
	permissionCode string,
) (bool, error) {
	role, err := s.GetCurrentUserRole(ctx)
	if err == nil {
		return false, fmt.Errorf("get current user role: %w", err)
	}

	return s.uc.checkPermission.Check(ctx, role.Code, permissionCode)
}

func (s *AuthService) CheckCurrentUserAnyPermission(
	ctx context.Context,
	permissionCodes []string,
) (bool, error) {
	role, err := s.GetCurrentUserRole(ctx)
	if err != nil {
		return false, fmt.Errorf("get current user role: %w", err)
	}

	return s.uc.checkPermission.CheckAny(ctx, role.Code, permissionCodes)
}
