package analytics_ws_handler

import (
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	ws_utils "github.com/vekshinnikita/pulse_watch/internal/handlers/web/ws/utils"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type AnalyticsRouter struct {
	services *service.Service
	router   ws_utils.Router
	hub      ws_utils.Hub
}

func newAnalyticsWSRouter(services *service.Service, hub ws_utils.Hub) WSRouter {
	router := &AnalyticsRouter{
		services: services,
		router:   ws_utils.NewRouter(),
		hub:      hub,
	}
	router.initRoutes()

	return router
}

func (r *AnalyticsRouter) initRoutes() {
	r.router.Register("subscribe", r.SubscribeRoute)
}

func (r *AnalyticsRouter) Process(client *entities.WSClient, message *entities.WSMessage) error {
	return r.router.Process(client, message)
}

func (r *AnalyticsRouter) SubscribeRoute(client *entities.WSClient, p entities.WSMessagePayload) error {
	payload, ok := p.(map[string]any)
	if !ok {
		return fmt.Errorf("can't get payload")
	}

	periodType, ok := payload["period_type"].(string)
	if !ok {
		return fmt.Errorf("can't get subscribe type")
	}

	appId, ok := payload["app_id"].(float64)
	if !ok {
		return fmt.Errorf("can't get app id")
	}

	channelId := fmt.Sprintf("metric:channel:%s:%d", periodType, int(appId))

	// Выходим отписываемся от всех каналов
	r.hub.LeaveAllRooms(client)

	r.hub.JoinRoom(channelId, client)
	return nil
}
