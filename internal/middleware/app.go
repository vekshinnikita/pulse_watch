package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

const (
	apiKeyHeader = "X-API-Key"
)

func AuthAppMiddleware(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		apiKey := c.GetHeader(apiKeyHeader)

		if apiKey == "" {
			response.NewErrorResponse(c, http.StatusUnauthorized, errs.UnauthorizedErrorMessage)
			return
		}

		appId, err := services.App.CheckApiKey(ctx, apiKey)
		if err != nil {
			var te *errs.ExpiredOrRevokedApiKeyError
			if errors.As(err, &te) {
				// Токен истек
				response.NewErrorResponse(c, http.StatusUnauthorized, te.Error())
				return
			}

			var ie *errs.InvalidApiKeyError
			if errors.As(err, &ie) {
				// Токен истек
				response.NewErrorResponse(c, http.StatusUnauthorized, te.Error())
				return
			}

			slog.ErrorContext(ctx, fmt.Sprintf("auth app middleware check api key: %s", err.Error()))
			response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
			return
		}

		ctx = context.WithValue(ctx, constants.AppIdCtxKey, appId)
		// Добавляем user_id в контекст логгера
		ctx = logger.AddLogAttrs(ctx,
			slog.Int("app_id", appId),
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}
