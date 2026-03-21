package app_handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	entities "github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	handler_helpers "github.com/vekshinnikita/pulse_watch/internal/handlers/web/helpers"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/middleware"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/internal/validators"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type AppHandler struct {
	services *service.Service
}

func NewAppHandler(services *service.Service) *AppHandler {
	return &AppHandler{
		services: services,
	}
}

func (h *AppHandler) InitRoutes(r *gin.RouterGroup) {
	adminRoleRequiredMiddleware := middleware.HasRoleMiddleware(h.services, "admin")

	// app
	r.POST("", adminRoleRequiredMiddleware, h.CreateApp)
	r.GET("/:id", h.GetApp)
	r.DELETE("/:id", adminRoleRequiredMiddleware, h.DeleteApp)

	// api key
	r.POST("/api_keys", adminRoleRequiredMiddleware, h.CreateApiKey)
	r.POST("/api_keys/:id/revoke", adminRoleRequiredMiddleware, h.RevokeApiKey)
	r.GET("/:id/api_keys", adminRoleRequiredMiddleware, h.GetAppApiKeys)
}

func (h *AppHandler) CreateApp(c *gin.Context) {
	var data entities.CreateAppData
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("create app bad request: %s", err.Error()))
		return
	}

	app, err := h.services.App.CreateAndGetApp(ctx, &data)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to create app: %s", err.Error()))

		var ue errs.StructuredError
		if errors.As(err, &ue) {
			response.NewFieldsErrorResponse(c, http.StatusConflict, ue.GetStructuredErrors())
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusCreated, app)

}

func (h *AppHandler) GetApp(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("app_id", id),
	)

	app, err := h.services.App.GetApp(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, nfe.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, app)
}

func (h *AppHandler) DeleteApp(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("app_id", id),
	)

	err = h.services.App.DeleteApp(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, nfe.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

func (h *AppHandler) CreateApiKey(c *gin.Context) {
	var data entities.CreateApiKeyData
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("create api key bad request: %s", err.Error()))
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("app_id", data.AppId),
	)

	apiKey, err := h.services.App.CreateApiKey(ctx, &data)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to create api: %s", err.Error()))

		var ue errs.StructuredError
		if errors.As(err, &ue) {
			response.NewFieldsErrorResponse(c, http.StatusConflict, ue.GetStructuredErrors())
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusCreated, apiKey)

}

func (h *AppHandler) GetAppApiKeys(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "app_id must be valid integer")
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("app_id", id),
	)

	pagination, err := handler_helpers.ParsePagination(c)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("get app api keys bad pagination params: %s", err.Error()))
		return
	}

	result, err := h.services.App.GetAppApiKeysPaginated(ctx, pagination, id)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())
		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AppHandler) RevokeApiKey(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	// Добавление api_key_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("api_key_id", id),
	)

	err = h.services.App.RevokeApiKey(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, nfe.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
