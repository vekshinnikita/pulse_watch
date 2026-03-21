package auth_handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	entities "github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/internal/validators"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type AuthHandler struct {
	services *service.Service
}

func NewAuthHandler(services *service.Service) *AuthHandler {
	return &AuthHandler{
		services: services,
	}
}

func (h *AuthHandler) InitRoutes(r *gin.RouterGroup) {
	r.POST("/refresh_tokens", h.RefreshTokens)

	sign := r.Group("/sign")
	{
		sign.POST("/in", h.SignIn)
		sign.POST("/up", h.SignUp)
	}

}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var createUser entities.SignUpUser
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &createUser); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("create user bad request: %s", err.Error()))
		return
	}

	user, err := h.services.Auth.CreateAndGetUser(ctx, &createUser)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to create user: %s", err.Error()))

		var se errs.StructuredError
		if errors.As(err, &se) {
			response.NewFieldsErrorResponse(c, http.StatusConflict, se.GetStructuredErrors())
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusCreated, user)

}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var signInUser entities.SignInUser
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &signInUser); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("sign in user bad request: %s", err.Error()))
		return
	}

	logger.AddLogAttrs(ctx,
		slog.String("username", signInUser.Username),
	)

	tokens, err := h.services.Auth.SignIn(ctx, &signInUser)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("user sign in failed: %s", err.Error()))

		var ue *errs.UserNotFoundError
		if errors.As(err, &ue) {
			response.NewErrorResponse(c, http.StatusBadRequest, errs.InvalidLoginOrPasswordErrorMessage)
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusOK, tokens)
}

func (h *AuthHandler) RefreshTokens(c *gin.Context) {
	var input entities.RefreshToken
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &input); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("refresh tokens user bad request: %s", err.Error()))
		return
	}

	tokens, err := h.services.Auth.RefreshTokens(ctx, input.RefreshToken)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("user refresh tokens failed: %s", err.Error()))

		var te *errs.ExpiredOrRevokedTokenError
		if errors.As(err, &te) {
			response.NewErrorResponse(c, http.StatusUnauthorized, te.Error())
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusOK, tokens)
}
