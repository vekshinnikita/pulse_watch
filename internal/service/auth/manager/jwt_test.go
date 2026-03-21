package jwt_manager

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vekshinnikita/pulse_watch/internal/config"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
)

func TestJWTTokenManager_GenerateAccessToken(t *testing.T) {
	authCfg := GetConfig()
	user := testutils.NewTestUser("test")

	mng := NewDefaultJWTTokenManager()
	ctx := context.Background()

	testCases := []struct {
		name          string
		generateToken func(ctx context.Context, user *models.User) (*entities.AccessTokenClaims, string, error)
		tokenTtl      time.Duration
	}{
		{
			name:          "Access token",
			generateToken: mng.GenerateAccessToken,
			tokenTtl:      time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claims, token, err := testCase.generateToken(ctx, user)
			assert.NoError(t, err)

			assert.NotEmpty(t, token, "Token is empty")
			issuedAtTime := time.Unix(claims.IssuedAt, 0)
			assert.NotEmpty(t, claims.JTI, "Token JTI is empty")
			assert.Greater(t, claims.ExpiresAt, claims.IssuedAt, "ExpiresAt must be greater then IssuedAt")
			assert.Equal(t, user.Id, claims.UserId, "Claims userId and input userId must be equal")
			assert.Equal(t, claims.ExpiresAt, issuedAtTime.Add(testCase.tokenTtl).Unix(), "ExpiresAt must be IssuedAt + ttl")
		})
	}

}

func TestJWTTokenManager_GenerateRefreshToken(t *testing.T) {
	authCfg := GetConfig()
	user := testutils.NewTestUser("test")

	mng := NewDefaultJWTTokenManager()
	ctx := context.Background()

	testCases := []struct {
		name          string
		generateToken func(ctx context.Context, user *models.User) (*entities.TokenClaims, string, error)
		tokenTtl      time.Duration
	}{
		{
			name:          "Refresh token",
			generateToken: mng.GenerateRefreshToken,
			tokenTtl:      24 * time.Hour * time.Duration(authCfg.RefreshTokenTtlDays),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			claims, token, err := testCase.generateToken(ctx, user)
			assert.NoError(t, err)

			assert.NotEmpty(t, token, "Token is empty")
			issuedAtTime := time.Unix(claims.IssuedAt, 0)
			assert.NotEmpty(t, claims.JTI, "Token JTI is empty")
			assert.Greater(t, claims.ExpiresAt, claims.IssuedAt, "ExpiresAt must be greater then IssuedAt")
			assert.Equal(t, user.Id, claims.UserId, "Claims userId and input userId must be equal")
			assert.Equal(t, claims.ExpiresAt, issuedAtTime.Add(testCase.tokenTtl).Unix(), "ExpiresAt must be IssuedAt + ttl")
		})
	}

}

func TestJWTTokenManager_ParseToken(t *testing.T) {
	securityCfg := config.GetSecurityConfig()
	authCfg := GetConfig()
	user := testutils.NewTestUser("test")

	testCases := []struct {
		name          string
		generateToken func() (string, error)
		wantError     error
		wantSomeError bool
	}{
		{
			name: "OK",
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey:  securityCfg.TokenSecretKey,
					RefreshTtl: time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
				})

				_, token, err := generateMng.GenerateRefreshToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Token expired",
			wantError: &errs.ExpiredOrRevokedTokenError{Message: errs.TokenExpiredErrorMessage},
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey:  securityCfg.TokenSecretKey,
					RefreshTtl: -time.Hour,
				})

				_, token, err := generateMng.GenerateRefreshToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Invalid signature",
			wantError: &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage},
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey:  "sometoken",
					RefreshTtl: time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
				})

				_, token, err := generateMng.GenerateRefreshToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Invalid token",
			wantError: &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage},
			generateToken: func() (string, error) {
				return "token", nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mng := NewDefaultJWTTokenManager()

			token, err := testCase.generateToken()
			assert.NoError(t, err)

			claims, err := mng.ParseToken(context.Background(), token)
			if testCase.wantSomeError {
				assert.Error(t, err)
				return
			}

			if testCase.wantError != nil {
				assert.Equal(t, testCase.wantError, err)
				return
			}

			assert.NoError(t, err)

			assert.NotEmpty(t, claims.JTI, "Token JTI is empty")
			assert.Greater(t, claims.ExpiresAt, claims.IssuedAt, "ExpiresAt must be greater then IssuedAt")
			assert.Equal(t, user.Id, claims.UserId, "Claims userId and input userId must be equal")
		})
	}

}

func TestJWTTokenManager_ParseAccessToken(t *testing.T) {
	securityCfg := config.GetSecurityConfig()
	authCfg := GetConfig()
	user := testutils.NewTestUser("test")

	testCases := []struct {
		name          string
		generateToken func() (string, error)
		wantError     error
		wantSomeError bool
	}{
		{
			name: "OK",
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey: securityCfg.TokenSecretKey,
					AccessTtl: time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
				})

				_, token, err := generateMng.GenerateAccessToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Token expired",
			wantError: &errs.ExpiredOrRevokedTokenError{Message: errs.TokenExpiredErrorMessage},
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey: securityCfg.TokenSecretKey,
					AccessTtl: -time.Hour,
				})

				_, token, err := generateMng.GenerateAccessToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Invalid signature",
			wantError: &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage},
			generateToken: func() (string, error) {
				generateMng := NewJWTTokenManager(&TokenManagerOptions{
					SecretKey: "sometoken",
					AccessTtl: time.Minute * time.Duration(authCfg.AccessTokenTtlMinutes),
				})

				_, token, err := generateMng.GenerateAccessToken(context.Background(), user)

				return token, err
			},
		},

		{
			name:      "Invalid token",
			wantError: &errs.InvalidTokenError{Message: errs.InvalidTokenErrorMessage},
			generateToken: func() (string, error) {
				return "token", nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mng := NewDefaultJWTTokenManager()

			token, err := testCase.generateToken()
			assert.NoError(t, err)

			claims, err := mng.ParseAccessToken(context.Background(), token)
			if testCase.wantSomeError {
				assert.Error(t, err)
				return
			}

			if testCase.wantError != nil {
				assert.Equal(t, testCase.wantError, err)
				return
			}

			assert.NoError(t, err)

			assert.NotEmpty(t, claims.JTI, "Token JTI is empty")
			assert.Greater(t, claims.ExpiresAt, claims.IssuedAt, "ExpiresAt must be greater then IssuedAt")
			assert.Equal(t, user.Id, claims.UserId, "Claims userId and input userId must be equal")
		})
	}

}
