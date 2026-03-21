package auth_handler

import (
	"testing"

	"github.com/vekshinnikita/pulse_watch/internal/service"
	mock_service "github.com/vekshinnikita/pulse_watch/internal/service/mocks"
	gomock "go.uber.org/mock/gomock"
)

type AuthMockBehavior func(s *mock_service.MockAuthService)

func NewAuthMockHandler(t *testing.T, mockBehavior AuthMockBehavior) *AuthHandler {
	c := gomock.NewController(t)
	t.Cleanup(func() { c.Finish() })

	auth := mock_service.NewMockAuthService(c)
	mockBehavior(auth)

	services := &service.Service{Auth: auth}
	return NewAuthHandler(services)
}
