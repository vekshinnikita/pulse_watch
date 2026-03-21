package auth_handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	mock_service "github.com/vekshinnikita/pulse_watch/internal/service/mocks"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type testCase struct {
	name                string
	inputBody           string
	inputUser           entities.SignUpUser
	mockBehavior        AuthMockBehavior
	expectedStatusCode  int
	expectedRequestBody string
}

func TestMain(m *testing.M) {
	// подготовка перед тестами
	testutils.HandlerSetup()

	// запуск всех тестов
	code := m.Run()

	os.Exit(code)
}

func TestHandler_SignUp(t *testing.T) {
	user := testutils.NewTestUser("test")

	// OK test
	inputUser := &entities.SignUpUser{
		Name:     user.Name,
		Username: user.Username,
		Password: "Test1234",
		Email:    user.Email,
		TgId:     user.TgId,
	}
	inputUserJson := testutils.MarshalTestJson(t, inputUser)

	testCases := []testCase{
		{
			name:      "OK",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().CreateAndGetUser(context.Background(), inputUser).Return(user, nil)
			},
			expectedStatusCode:  http.StatusCreated,
			expectedRequestBody: testutils.MarshalTestJson(t, user),
		},

		{
			name:               "No required fields",
			inputBody:          `{"username":"test","password":"Test1234"}`,
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedRequestBody: testutils.MarshalTestJson(t, response.FieldErrorsResponse{
				Message: errs.InvalidInputErrorMessage,
				Errors: errs.StructuredFieldErrors{
					errs.StructuredFieldError{
						Field:   "name",
						Type:    errs.RequiredFieldErrorType,
						Message: errs.RequiredFieldErrorMessage,
					},
				},
			}),
		},

		{
			name:      "Unique field",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					CreateAndGetUser(context.Background(), inputUser).
					Return(nil, fmt.Errorf("service error: %w", &errs.UniqueFieldError{Field: "username"}))
			},
			expectedStatusCode: http.StatusConflict,
			expectedRequestBody: testutils.MarshalTestJson(t, response.FieldErrorsResponse{
				Message: errs.InvalidInputErrorMessage,
				Errors: errs.StructuredFieldErrors{
					errs.StructuredFieldError{
						Field:   "username",
						Type:    errs.UniqueFieldErrorType,
						Message: fmt.Sprintf(errs.UniqueFieldErrorMessage, "username"),
					},
				},
			}),
		},

		{
			name:      "Service failure",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					CreateAndGetUser(context.Background(), inputUser).
					Return(nil, errors.New("service failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InternalErrorMessage,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewAuthMockHandler(t, testCase.mockBehavior)

			r := testutils.MakeTestRequest(&testutils.MakeRequestOptions{
				Handlers: []gin.HandlerFunc{handler.SignUp},
				Body:     testCase.inputBody,
			})

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, r.Code)
			assert.Equal(t, testCase.expectedRequestBody, r.Body.String())
		})
	}
}

func TestHandler_SignIn(t *testing.T) {
	tokens := &entities.AuthTokens{
		AccessToken:  "AccessToken",
		RefreshToken: "RefreshToken",
	}

	// OK test
	inputUser := &entities.SignInUser{
		Username: "test",
		Password: "Test1234",
	}
	inputUserJson := testutils.MarshalTestJson(t, inputUser)

	testCases := []testCase{
		{
			name:      "OK",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().SignIn(context.Background(), inputUser).Return(tokens, nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: testutils.MarshalTestJson(t, tokens),
		},

		{
			name:               "No required fields",
			inputBody:          `{"username":"test"}`,
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedRequestBody: testutils.MarshalTestJson(t, response.FieldErrorsResponse{
				Message: errs.InvalidInputErrorMessage,
				Errors: errs.StructuredFieldErrors{
					errs.StructuredFieldError{
						Field:   "password",
						Type:    errs.RequiredFieldErrorType,
						Message: errs.RequiredFieldErrorMessage,
					},
				},
			}),
		},

		{
			name:      "User not found",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					SignIn(context.Background(), inputUser).
					Return(nil, fmt.Errorf("service error: %w", &errs.UserNotFoundError{}))
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InvalidLoginOrPasswordErrorMessage,
			}),
		},

		{
			name:      "Service failure",
			inputBody: inputUserJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					SignIn(context.Background(), inputUser).
					Return(nil, errors.New("service failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InternalErrorMessage,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewAuthMockHandler(t, testCase.mockBehavior)

			r := testutils.MakeTestRequest(&testutils.MakeRequestOptions{
				Handlers: []gin.HandlerFunc{handler.SignIn},
				Body:     testCase.inputBody,
			})

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, r.Code)
			assert.Equal(t, testCase.expectedRequestBody, r.Body.String())
		})
	}
}

func TestHandler_RefreshTokens(t *testing.T) {
	newTokens := &entities.AuthTokens{
		AccessToken:  "AccessToken",
		RefreshToken: "RefreshToken",
	}

	// OK test
	input := &entities.RefreshToken{
		RefreshToken: "RefreshToken",
	}
	inputJson := testutils.MarshalTestJson(t, input)

	testCases := []testCase{
		{
			name:      "OK",
			inputBody: inputJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().RefreshTokens(context.Background(), input.RefreshToken).Return(newTokens, nil)
			},
			expectedStatusCode:  http.StatusOK,
			expectedRequestBody: testutils.MarshalTestJson(t, newTokens),
		},

		{
			name:               "No required fields",
			inputBody:          `{}`,
			mockBehavior:       func(s *mock_service.MockAuthService) {},
			expectedStatusCode: http.StatusBadRequest,
			expectedRequestBody: testutils.MarshalTestJson(t, response.FieldErrorsResponse{
				Message: errs.InvalidInputErrorMessage,
				Errors: errs.StructuredFieldErrors{
					errs.StructuredFieldError{
						Field:   "refresh_token",
						Type:    errs.RequiredFieldErrorType,
						Message: errs.RequiredFieldErrorMessage,
					},
				},
			}),
		},

		{
			name:      "Expired token",
			inputBody: inputJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					RefreshTokens(context.Background(), input.RefreshToken).
					Return(
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

		{
			name:      "Service failure",
			inputBody: inputJson,
			mockBehavior: func(s *mock_service.MockAuthService) {
				s.EXPECT().
					RefreshTokens(context.Background(), input.RefreshToken).
					Return(nil, errors.New("service failed"))
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedRequestBody: testutils.MarshalTestJson(t, response.ErrorResponse{
				Message: errs.InternalErrorMessage,
			}),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			handler := NewAuthMockHandler(t, testCase.mockBehavior)

			r := testutils.MakeTestRequest(&testutils.MakeRequestOptions{
				Handlers: []gin.HandlerFunc{handler.RefreshTokens},
				Body:     testCase.inputBody,
			})

			// Assert
			assert.Equal(t, testCase.expectedStatusCode, r.Code)
			assert.Equal(t, testCase.expectedRequestBody, r.Body.String())
		})
	}
}
