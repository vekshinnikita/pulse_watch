package analytics_handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	handler_helpers "github.com/vekshinnikita/pulse_watch/internal/handlers/web/helpers"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/internal/validators"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type AnalyticsHandler struct {
	services *service.Service
}

func NewAnalyticsHandler(services *service.Service) *AnalyticsHandler {
	return &AnalyticsHandler{
		services: services,
	}
}

func (h *AnalyticsHandler) InitRoutes(r *gin.RouterGroup) {
	appsGroup := r.Group("/apps")

	// metrics
	appsGroup.POST("/:app_id/metrics", h.GetMetrics)

	// logs
	r.POST("/logs/search", h.SearchLogs)
	r.POST("/logs/meta_vars", h.GetMetaVars)

}

func (h *AnalyticsHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	appId, err := strconv.Atoi(c.Param("app_id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "app_id must be valid integer")
		return
	}

	var data *entities.GetMetricsData
	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("get metrics request: %s", err.Error()))
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx,
		slog.Int("app_id", appId),
	)

	result, err := h.services.Metric.GetMetrics(ctx, appId, data)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) SearchLogs(c *gin.Context) {
	ctx := c.Request.Context()

	var data *entities.SearchLogData
	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("search logs bad request: %s", err.Error()))
		return
	}

	if data.Page == 0 {
		data.Page = 1
	}

	if data.PageSize == 0 {
		data.PageSize = 20
	}

	result, err := h.services.Logs.SearchPaginatedLogs(ctx, data)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("search paginated logs: %s", err.Error()))

		var ue *errs.TypeFieldError
		if errors.As(err, &ue) {
			response.NewFieldsErrorResponse(c, http.StatusBadRequest, ue.GetStructuredErrors())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *AnalyticsHandler) GetMetaVars(c *gin.Context) {
	ctx := c.Request.Context()

	var data entities.PaginationDataWithAppId
	if err := handler_helpers.ParseParamsWithPagination(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("get meta var request: %s", err.Error()))
		response.NewErrorResponse(c, http.StatusBadRequest, "invalid query params")
		return
	}

	result, err := h.services.Logs.GetMetaVarsPaginated(ctx, &data.PaginationData, data.AppId)
	if err != nil {
		slog.ErrorContext(ctx, err.Error())

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, result)
}
