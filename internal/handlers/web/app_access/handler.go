package app_access_handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/internal/validators"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type AppAccessHandler struct {
	services *service.Service
}

func NewAppAccessHandler(services *service.Service) *AppAccessHandler {
	return &AppAccessHandler{
		services: services,
	}
}

func (h *AppAccessHandler) InitRoutes(r *gin.RouterGroup) {

	// logs
	r.POST("/logs", h.SendLogs)

}

func (h *AppAccessHandler) SendLogs(c *gin.Context) {
	var data entities.SendLogs
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("send logs bad request: %s", err.Error()))
		return
	}

	err := h.services.Logs.SendLogs(ctx, data.Logs)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("failed to create app: %s", err.Error()))
		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			errs.InternalErrorMessage,
		)
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
