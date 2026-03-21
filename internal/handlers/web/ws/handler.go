package ws_handler

import (
	"github.com/gin-gonic/gin"
	analytics_ws_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/ws/handlers/analytics"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type Handler struct {
	services *service.Service

	analyticsHandler WSSubHandler
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services: services,

		analyticsHandler: analytics_ws_handler.NewAnalyticsHandler(services),
	}
}

func (h *Handler) InitRoutes(r *gin.RouterGroup) {
	r.GET("/analytics", h.analyticsHandler.Handle)
	r.GET("/analytics/live", h.analyticsHandler.HandleLive)
}
