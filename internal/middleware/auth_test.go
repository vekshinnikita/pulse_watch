package middleware

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	mock_service "github.com/vekshinnikita/pulse_watch/internal/service/mocks"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type testCase struct {
	name                string
	header              testutils.RequestHeader
	mockBehavior        AuthMockBehavior
	expectedStatusCode  int
	expectedRequestBody string
}

func TestMiddleware_AuthUserMiddleware(t *testing.T) {
	user := testutils.NewTestUser("test")
	token := "token"
	testCases := []testCase{
		{
			name: "OK",
			header: testutils.RequestHeader{
				Key:   "Authorization",
				Value: fmt.Sprintf("Bearer %s", token),
			},
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().ParseAccessToken(context.Background(), token).Return(
					&entities.AccessTokenClaims{
						UserId: user.Id,
						Role:   &user.Role,
					},
					nil,
				)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: "1",
		},

		{
			name:               "No header",
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.UnauthorizedErrorMessage,
			}),
		},

		{
			name: "Invalid Bearer",
			header: testutils.RequestHeader{
				Key:   "Authorization",
				Value: fmt.Sprintf("Berer %s", token),
			},
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InvalidAuthHeaderErrorMessage,
			}),
		},

		{
			name: "No token",
			header: testutils.RequestHeader{
				Key:   "Authorization",
				Value: "Bearer",
			},
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusUnauthorized,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InvalidAuthHeaderErrorMessage,
			}),
		},

		{
			name: "Expired token",
			header: testutils.RequestHeader{
				Key:   "Authorization",
				Value: fmt.Sprintf("Bearer %s", token),
			},
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().ParseAccessToken(context.Background(), token).Return(
					nil,
					fmt.Errorf(
						"service error: %w",
						&errs.ExpiredOrRevokedTokenError{Message: errs.TokenExpiredErrorMessage},
					),
				)
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.TokenExpiredErrorMessage,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			auth_middleware := NewAuthMockMiddleware(t, testCase.mockBehavior)

			handlerFunc := func(c *gin.Context) {
				ctx := c.Request.Context()

				userId, ok := ctx.Value(constants.UserIdCtxKey).(int)
				if !ok {
					response.NewErrorResponse(
						c,
						http.StatusInternalServerError,
						fmt.Sprintf("%v in the request context has to be int", constants.UserIdCtxKey),
					)
					return
				}

				c.String(http.StatusOK, fmt.Sprintf("%v", userId))
			}

			r := testutils.MakeTestRequest(&testutils.MakeRequestOptions{
				Headers:  []testutils.RequestHeader{testCase.header},
				Handlers: []gin.HandlerFunc{auth_middleware, handlerFunc},
			})

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, r.Code)
			assert.Equal(t, testCase.expectedRequestBody, r.Body.String())
		})
	}
}
