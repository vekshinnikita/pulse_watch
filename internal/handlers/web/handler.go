package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	alert_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/alert"
	analytics_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/analytics"
	app_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/app"
	app_access_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/app_access"
	auth_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/auth"
	ws_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/web/ws"
	"github.com/vekshinnikita/pulse_watch/internal/middleware"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type Handler struct {
	services  *service.Service
	Auth      SubHandler
	App       SubHandler
	AppAccess SubHandler
	Alert     SubHandler
	Analytics SubHandler

	Ws SubHandler
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		services:  services,
		Auth:      auth_handler.NewAuthHandler(services),
		App:       app_handler.NewAppHandler(services),
		AppAccess: app_access_handler.NewAppAccessHandler(services),
		Alert:     alert_handler.NewAlertHandler(services),
		Analytics: analytics_handler.NewAnalyticsHandler(services),

		Ws: ws_handler.NewHandler(services),
	}
}

func initMiddleware(r *gin.Engine) {
	// логгер запросов
	r.Use(middleware.RequestLoggerMiddleware())

	// recovery
	r.Use(middleware.RecoveryWithLogging())

	// добавление requestId в контекст
	r.Use(middleware.RequestIDMiddleware())
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()
	initMiddleware(router)

	authorizedGroup := router.Group(
		"",
		middleware.AuthUserMiddleware(h.services),
		middleware.HasNoRoleMiddleware(h.services, constants.GuestRoleCode),
	)
	authorizedAppGroup := router.Group(
		"",
		middleware.AuthAppMiddleware(h.services),
	)

	h.Auth.InitRoutes(router.Group("/auth"))
	h.Ws.InitRoutes(router.Group("/ws"))

	h.AppAccess.InitRoutes(authorizedAppGroup.Group("/access/apps"))
	h.InitUserRoutes(authorizedGroup.Group("/user"))
	h.App.InitRoutes(authorizedGroup.Group("/apps"))
	h.Alert.InitRoutes(authorizedGroup.Group("/alerts"))
	h.Analytics.InitRoutes(authorizedGroup.Group("/analytics"))

	return router
}
