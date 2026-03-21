package auth_service

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
	"go.uber.org/mock/gomock"
)

func TestMain(m *testing.M) {
	// подготовка перед тестами
	testutils.ServiceSetup()

	// запуск всех тестов
	code := m.Run()

	os.Exit(code)
}

func TestAuthService_CreateUser(t *testing.T) {
	cfg := config.GetSecurityConfig()
	user := testutils.NewTestUser("test")

	userSignUp := &entities.SignUpUser{
		Name:     user.Name,
		Username: user.Username,
		Password: "Test1234",
		Email:    user.Email,
		TgId:     user.TgId,
	}

	hashedUserSignUp := *userSignUp
	hashedUserSignUp.Password = utils.HashStringWithSalt(hashedUserSignUp.Password, cfg.SecretKey)

	testCases := []struct {
		name          string
		input         entities.SignUpUser
		expected      int
		mockBehavior  mockBehavior
		wantSomeError bool
	}{
		{
			name:     "OK",
			input:    *userSignUp,
			expected: user.Id,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				mock.auth_repo.EXPECT().
					CreateUser(context.Background(), &hashedUserSignUp).
					Return(user.Id, nil)
			},
		},
		{
			name:          "Repo error",
			input:         *userSignUp,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				mock.auth_repo.EXPECT().
					CreateUser(context.Background(), &hashedUserSignUp).
					Return(0, fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockAuthService(t, testCase.mockBehavior)

			userId, err := service.CreateUser(context.Background(), &testCase.input)
			if testCase.wantSomeError {
				assert.Error(t, err)
			}

			assert.Equal(t, testCase.expected, userId)
		})
	}
}

func TestAuthService_GetUserById(t *testing.T) {
	user := testutils.NewTestUser("test")

	testCases := []struct {
		name          string
		input         int
		expected      *models.User
		mockBehavior  mockBehavior
		wantSomeError bool
	}{
		{
			name:     "OK",
			input:    user.Id,
			expected: user,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				mock.auth_repo.EXPECT().
					GetUserById(context.Background(), user.Id).
					Return(user, nil)
			},
		},
		{
			name:          "Repo error",
			input:         user.Id,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				mock.auth_repo.EXPECT().
					GetUserById(context.Background(), user.Id).
					Return(nil, fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockAuthService(t, testCase.mockBehavior)

			userId, err := service.GetUserById(context.Background(), testCase.input)
			if testCase.wantSomeError {
				assert.Error(t, err)
			}

			assert.Equal(t, testCase.expected, userId)
		})
	}
}

func TestAuthService_GenerateTokens(t *testing.T) {
	user := testutils.NewTestUser("test")
	accessClaims, accessToken := getTestAccessToken(user, "access")
	refreshClaims, refreshToken := getTestToken(user, "refresh")

	testCases := []struct {
		name          string
		input         int
		expected      *entities.AuthTokens
		mockBehavior  mockBehavior
		wantSomeError bool
	}{
		{
			name:  "OK",
			input: user.Id,
			expected: &entities.AuthTokens{
				AccessToken:  accessToken,
				RefreshToken: refreshToken,
			},
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()

				mock.tokenMng.EXPECT().
					GenerateAccessToken(ctx, user).
					Return(accessClaims, accessToken, nil)

				mock.tokenMng.EXPECT().
					GenerateRefreshToken(ctx, user).
					Return(refreshClaims, refreshToken, nil)

				refreshExpiresAt := time.Unix(refreshClaims.ExpiresAt, 0)
				mock.auth_repo.EXPECT().
					SaveRefreshToken(ctx, user.Id, refreshClaims.JTI, &refreshExpiresAt).
					Return(1, nil)
			},
		},

		{
			name:          "Generate access token failed",
			input:         user.Id,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()

				mock.tokenMng.EXPECT().
					GenerateAccessToken(ctx, user).
					Return(nil, "", fmt.Errorf("some error"))
			},
		},

		{
			name:          "Generate refresh token failed",
			input:         user.Id,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()

				mock.tokenMng.EXPECT().
					GenerateAccessToken(ctx, user).
					Return(accessClaims, accessToken, nil)

				mock.tokenMng.EXPECT().
					GenerateRefreshToken(ctx, user).
					Return(nil, "", fmt.Errorf("some error"))
			},
		},

		{
			name:          "Saving Refresh token failed",
			input:         user.Id,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()

				mock.tokenMng.EXPECT().
					GenerateAccessToken(ctx, user).
					Return(accessClaims, accessToken, nil)

				mock.tokenMng.EXPECT().
					GenerateRefreshToken(ctx, user).
					Return(refreshClaims, refreshToken, nil)

				refreshExpiresAt := time.Unix(refreshClaims.ExpiresAt, 0)
				mock.auth_repo.EXPECT().
					SaveRefreshToken(ctx, user.Id, refreshClaims.JTI, &refreshExpiresAt).
					Return(0, fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockAuthService(t, testCase.mockBehavior)

			tokens, err := service.GenerateTokens(context.Background(), user)
			if testCase.wantSomeError {
				assert.Error(t, err)
			}

			assert.Equal(t, testCase.expected, tokens)
		})
	}
}

func TestAuthService_SignIn(t *testing.T) {
	securityCfg := config.GetSecurityConfig()

	user := testutils.NewTestUser("test")
	accessClaims, accessToken := getTestAccessToken(user, "access")
	refreshClaims, refreshToken := getTestToken(user, "refresh")

	userSignIn := &entities.SignInUser{
		Username: user.Username,
		Password: "Test1234",
	}
	tokens := &entities.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	testCases := []struct {
		name          string
		input         *entities.SignInUser
		expected      *entities.AuthTokens
		mockBehavior  mockBehavior
		wantSomeError bool
	}{
		{
			name:     "OK",
			input:    userSignIn,
			expected: tokens,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				hashedPassword := utils.HashStringWithSalt(userSignIn.Password, securityCfg.SecretKey)

				mock.auth_repo.EXPECT().
					GetUserByUsernameAndPassword(gomock.Any(), userSignIn.Username, hashedPassword).
					Return(user, nil)

				mock.tokenMng.EXPECT().
					GenerateAccessToken(gomock.Any(), gomock.Any()).
					Return(accessClaims, accessToken, nil)

				mock.tokenMng.EXPECT().
					GenerateRefreshToken(gomock.Any(), gomock.Any()).
					Return(refreshClaims, refreshToken, nil)

				mock.auth_repo.EXPECT().
					SaveRefreshToken(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return(1, nil)
			},
		},

		{
			name:          "Getting user by username and password failed",
			input:         userSignIn,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				hashedPassword := utils.HashStringWithSalt(userSignIn.Password, securityCfg.SecretKey)

				mock.auth_repo.EXPECT().
					GetUserByUsernameAndPassword(gomock.Any(), userSignIn.Username, hashedPassword).
					Return(nil, fmt.Errorf("some error"))
			},
		},

		{
			name:          "Generate tokens failed",
			input:         userSignIn,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				hashedPassword := utils.HashStringWithSalt(userSignIn.Password, securityCfg.SecretKey)

				mock.auth_repo.EXPECT().
					GetUserByUsernameAndPassword(gomock.Any(), userSignIn.Username, hashedPassword).
					Return(user, nil)

				mock.tokenMng.EXPECT().
					GenerateAccessToken(gomock.Any(), gomock.Any()).
					Return(nil, "", fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockAuthService(t, testCase.mockBehavior)

			userId, err := service.SignIn(context.Background(), testCase.input)
			if testCase.wantSomeError {
				assert.Error(t, err)
			}

			assert.Equal(t, testCase.expected, userId)
		})
	}
}

func TestAuthService_RefreshTokens(t *testing.T) {
	user := testutils.NewTestUser("test")
	accessClaims, accessToken := getTestAccessToken(user, "access")
	refreshClaims, refreshToken := getTestToken(user, "refresh")

	tokens := &entities.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	testCases := []struct {
		name          string
		input         string
		expected      *entities.AuthTokens
		mockBehavior  mockBehavior
		wantSomeError bool
		wantErr       error
	}{
		{
			name:     "OK",
			input:    refreshToken,
			expected: tokens,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()
				mock.tokenMng.EXPECT().
					ParseToken(ctx, refreshToken).
					Return(refreshClaims, nil)

				mock.auth_repo.EXPECT().
					IsRefreshTokenValid(ctx, refreshClaims.UserId, refreshClaims.JTI).
					Return(true, nil)

				mock.auth_repo.EXPECT().
					GetUserById(ctx, refreshClaims.UserId).
					Return(user, nil)

				mock.trm.EXPECT().Do(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						mock.tokenMng.EXPECT().
							GenerateAccessToken(gomock.Any(), user).
							Return(accessClaims, accessToken, nil)

						mock.tokenMng.EXPECT().
							GenerateRefreshToken(gomock.Any(), user).
							Return(refreshClaims, refreshToken, nil)

						expiresAt := time.Unix(refreshClaims.ExpiresAt, 0)
						mock.auth_repo.EXPECT().
							SaveRefreshToken(
								gomock.Any(),
								user.Id,
								refreshClaims.JTI,
								&expiresAt,
							).
							Return(1, nil)

						mock.auth_repo.EXPECT().
							RevokeRefreshToken(gomock.Any(), refreshClaims.JTI).
							Return(nil)

						err := fn(ctx) // вызываем реальную функцию внутри Do

						return err
					})
			},
		},

		{
			name:          "Parse token failed",
			input:         refreshToken,
			expected:      tokens,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()
				mock.tokenMng.EXPECT().
					ParseToken(ctx, refreshToken).
					Return(nil, fmt.Errorf("some error"))
			},
		},

		{
			name:          "Is refresh token valid failed",
			input:         refreshToken,
			expected:      tokens,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()
				mock.tokenMng.EXPECT().
					ParseToken(ctx, refreshToken).
					Return(refreshClaims, nil)

				mock.auth_repo.EXPECT().
					IsRefreshTokenValid(ctx, refreshClaims.UserId, refreshClaims.JTI).
					Return(false, fmt.Errorf("some error"))
			},
		},

		{
			name:     "Refresh token is expired",
			input:    refreshToken,
			expected: tokens,
			wantErr:  &errs.ExpiredOrRevokedTokenError{Message: errs.TokenRevokedErrorMessage},
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()
				mock.tokenMng.EXPECT().
					ParseToken(ctx, refreshToken).
					Return(refreshClaims, nil)

				mock.auth_repo.EXPECT().
					IsRefreshTokenValid(ctx, refreshClaims.UserId, refreshClaims.JTI).
					Return(false, nil)
			},
		},

		{
			name:          "Transaction failed",
			input:         refreshToken,
			expected:      tokens,
			wantSomeError: true,
			mockBehavior: func(mock *mockAuthServiceDeps) {
				ctx := context.Background()
				mock.tokenMng.EXPECT().
					ParseToken(ctx, refreshToken).
					Return(refreshClaims, nil)

				mock.auth_repo.EXPECT().
					IsRefreshTokenValid(ctx, refreshClaims.UserId, refreshClaims.JTI).
					Return(true, nil)

				mock.auth_repo.EXPECT().
					GetUserById(ctx, refreshClaims.UserId).
					Return(user, nil)

				mock.trm.EXPECT().Do(ctx, gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(ctx context.Context) error) error {
						mock.tokenMng.EXPECT().
							GenerateAccessToken(gomock.Any(), user).
							Return(nil, "", fmt.Errorf("some error"))

						return fn(ctx)
					})
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			service := NewMockAuthService(t, testCase.mockBehavior)

			userId, err := service.RefreshTokens(context.Background(), testCase.input)
			if testCase.wantSomeError {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.Equal(t, testCase.expected, userId)
		})
	}
}
