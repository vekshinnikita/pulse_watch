package handlers

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/pkg/response"
)

func (h *Handler) InitUserRoutes(r *gin.RouterGroup) {
	r.GET("/me", h.getCurrentUser)
}

func (h *Handler) getCurrentUser(c *gin.Context) {
	ctx := c.Request.Context()

	userId, ok := ctx.Value(constants.UserIdCtxKey).(int)
	if !ok {
		slog.ErrorContext(ctx, "couldn't get user id")
	}

	user, err := h.services.Auth.GetUserById(ctx, userId)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("user sign in failed: %s", err.Error()))

		var ue *errs.UserNotFoundError
		if errors.As(err, &ue) {
			response.NewErrorResponse(c, http.StatusBadRequest, "User not found")
			return
		}

		response.NewErrorResponse(
			c,
			http.StatusInternalServerError,
			"Internal Server error",
		)
		return
	}

	c.JSON(http.StatusOK, user)
}
