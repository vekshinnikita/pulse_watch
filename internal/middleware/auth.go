package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

const (
	authorizationHeader = "Authorization"
)

func AuthUserMiddleware(services *service.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		header := c.GetHeader(authorizationHeader)

		if header == "" {
			response.NewErrorResponse(c, http.StatusUnauthorized, errs.UnauthorizedErrorMessage)
			return
		}

		headerParts := strings.Split(header, " ")
		if len(headerParts) != 2 {
			response.NewErrorResponse(c, http.StatusUnauthorized, errs.InvalidAuthHeaderErrorMessage)
			return
		}

		if headerParts[0] != "Bearer" {
			response.NewErrorResponse(c, http.StatusUnauthorized, errs.InvalidAuthHeaderErrorMessage)
			return
		}

		tokenClaims, err := services.Auth.ParseAccessToken(ctx, headerParts[1])
		if err != nil {
			var te *errs.ExpiredOrRevokedTokenError
			if errors.As(err, &te) {
				// Токен истек
				response.NewErrorResponse(c, http.StatusUnauthorized, te.Error())
				return
			}

			slog.ErrorContext(ctx, fmt.Sprintf("auth user middleware parse token: %s", err.Error()))
			response.NewErrorResponse(c, http.StatusUnauthorized, errs.InvalidTokenErrorMessage)
			return
		}

		ctx = context.WithValue(ctx, constants.UserRoleCtxKey, tokenClaims.Role)
		ctx = context.WithValue(ctx, constants.UserIdCtxKey, tokenClaims.UserId)
		// Добавляем user_id в контекст логгера
		ctx = logger.AddLogAttrs(ctx,
			slog.Int("user_id", tokenClaims.UserId),
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

func HasRoleMiddleware(services *service.Service, roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		role, err := services.Auth.GetCurrentUserRole(ctx)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("get current user role: %s", err.Error()))

			response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
			return
		}

		if !utils.SliceContains(roleCodes, role.Code) {
			// нет в списке ролей
			response.NewErrorResponse(c, http.StatusForbidden, errs.ForbiddenErrorMessage)
			return

		}

		c.Next()
	}
}

func HasNoRoleMiddleware(services *service.Service, roleCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		role, err := services.Auth.GetCurrentUserRole(ctx)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("get current user role: %s", err.Error()))

			response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
			return
		}

		if utils.SliceContains(roleCodes, role.Code) {
			// есть списке ролей
			response.NewErrorResponse(c, http.StatusForbidden, errs.ForbiddenErrorMessage)
			return

		}

		c.Next()
	}
}

func HasPermissionMiddleware(services *service.Service, permissionCodes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		ok, err := services.Auth.CheckCurrentUserAnyPermission(ctx, permissionCodes)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("check current user any permission: %s", err.Error()))

			response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
			return
		}

		if !ok {
			response.NewErrorResponse(c, http.StatusForbidden, errs.ForbiddenErrorMessage)
			return
		}

		c.Next()
	}
}
