package alert_handler

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	handler_helpers "github.com/vekshinnikita/pulse_watch/internal/handlers/web/helpers"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/middleware"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/internal/validators"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

type AlertHandler struct {
	services *service.Service
}

func NewAlertHandler(services *service.Service) *AlertHandler {
	return &AlertHandler{
		services: services,
	}
}

func (h *AlertHandler) InitRoutes(r *gin.RouterGroup) {
	crudRulePermissionMiddleware := middleware.HasPermissionMiddleware(
		h.services,
		constants.CRUDAlertRulePermission,
	)

	// alert rule
	r.POST("/rules", crudRulePermissionMiddleware, h.CreateAlertRule)
	r.GET("/rules/app/:id", crudRulePermissionMiddleware, h.GetAlertRules)
	r.GET("/rules/:id", crudRulePermissionMiddleware, h.GetAlertRule)
	r.PATCH("/rules/:id", crudRulePermissionMiddleware, h.UpdateAlertRule)
	r.DELETE("/rules/:id", crudRulePermissionMiddleware, h.DeleteAlertRule)
}

func (h *AlertHandler) GetAlertRules(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	// Добавление app_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx, slog.Int("app_id", id))

	pagination, err := handler_helpers.ParsePagination(c)
	if err != nil {
		response.NewErrorResponse(
			c,
			http.StatusBadRequest,
			fmt.Sprintf("get app alert rules bad pagination params: %s", err.Error()),
		)
		return
	}

	paginatedRules, err := h.services.Alert.GetAppRulesPaginated(ctx, pagination, id)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("get app rules paginated: %s", err.Error()))
		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, paginatedRules)
}

func (h *AlertHandler) GetAlertRule(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	// Добавление rule_id в контекст для логгера
	ctx = logger.AddLogAttrs(ctx, slog.Int("rule_id", id))

	rule, err := h.services.Alert.GetRule(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("get rule: %s", err.Error()))

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, nfe.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, rule)
}

func (h *AlertHandler) CreateAlertRule(c *gin.Context) {
	var data entities.CreateAlertRule
	ctx := c.Request.Context()

	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("create alert rule bad request: %s", err.Error()))
		return
	}

	// Добавление app_id в контекст для логгера
	if data.AppId != nil {
		ctx = logger.AddLogAttrs(ctx,
			slog.Int("app_id", *data.AppId),
		)
	}

	rule, err := h.services.Alert.CreateAndGetAlertRule(ctx, &data)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("create and get alert rule: %s", err.Error()))

		var ue errs.StructuredError
		if errors.As(err, &ue) {
			response.NewFieldsErrorResponse(c, http.StatusConflict, ue.GetStructuredErrors())
			return
		}

		var cve *errs.CheckViolationError
		if errors.As(err, &cve) {
			response.NewErrorResponse(c, http.StatusBadRequest, cve.Error())
			return
		}

		var de *errs.DuplicateAlertRuleError
		if errors.As(err, &de) {
			response.NewErrorResponse(c, http.StatusConflict, de.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusCreated, rule)
}

func (h *AlertHandler) UpdateAlertRule(c *gin.Context) {
	var data entities.UpdateAlertRule
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	ctx = logger.AddLogAttrs(ctx, slog.Int("rule_id", id))

	if err := validators.ValidateRequestFields(c, &data); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("update alert rule bad request: %s", err.Error()))
		return
	}

	err = h.services.Alert.UpdateAlertRule(ctx, id, &data)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("update alert rule: %s", err.Error()))

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, nil)
}

func (h *AlertHandler) DeleteAlertRule(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.NewErrorResponse(c, http.StatusBadRequest, "id must be valid integer")
		return
	}

	ctx = logger.AddLogAttrs(ctx, slog.Int("rule_id", id))

	err = h.services.Alert.DeleteAlertRule(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("delete alert rule: %s", err.Error()))

		var nfe *errs.NotFoundError
		if errors.As(err, &nfe) {
			response.NewErrorResponse(c, http.StatusNotFound, err.Error())
			return
		}

		response.NewErrorResponse(c, http.StatusInternalServerError, errs.InternalErrorMessage)
		return
	}

	c.JSON(http.StatusOK, nil)
}
