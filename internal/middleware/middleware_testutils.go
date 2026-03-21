package middleware

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	mock_service "github.com/vekshinnikita/pulse_watch/internal/service/mocks"
	"go.uber.org/mock/gomock"
)

type AuthMockBehavior func(s *mock_service.MockAuthService)

func NewAuthMockMiddleware(t *testing.T, mockBehavior AuthMockBehavior) gin.HandlerFunc {
	c := gomock.NewController(t)
	t.Cleanup(func() { c.Finish() })

	auth := mock_service.NewMockAuthService(c)
	mockBehavior(auth)

	services := &service.Service{Auth: auth}
	return AuthUserMiddleware(services)
}
