package analytics_ws_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	ws_utils "github.com/vekshinnikita/pulse_watch/internal/handlers/web/ws/utils"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type Handler struct {
	services *service.Service

	upgrader *websocket.Upgrader
	router   WSRouter
	hub      ws_utils.Hub
}

func NewAnalyticsHandler(services *service.Service) *Handler {
	hub := ws_utils.NewHub(
		services.Metric.SubscribeChannel,
		services.Metric.SendChannelMessage,
	)

	return &Handler{
		services: services,

		upgrader: &websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		router: newAnalyticsWSRouter(services, hub),
		hub:    hub,
	}
}

func (h *Handler) Handle(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("upgrade connection: %s", err.Error()))
		return
	}

	client := &entities.WSClient{
		Id:     uuid.NewString(),
		Conn:   conn,
		SendCh: make(chan []byte, 256),
		Ctx:    context.Background(),
	}

	go h.writePump(client)
	h.readPump(client)
}

func (h *Handler) HandleLive(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.Error(fmt.Sprintf("upgrade connection: %s", err.Error()))
		return
	}

	client := &entities.WSClient{
		Id:     uuid.NewString(),
		Conn:   conn,
		SendCh: make(chan []byte, 256),
		Ctx:    context.Background(),
	}

	h.sendLive(client)

	go h.writePump(client)
	h.readPump(client)
}

func (h *Handler) sendLive(client *entities.WSClient) {

	message := &entities.WSMessage{
		Type: "subscribe",
		Payload: map[string]any{
			"period_type": "live",
			"app_id":      float64(6),
		},
	}
	err := h.router.Process(client, message)
	if err != nil {
		slog.ErrorContext(client.Ctx, fmt.Sprintf("process message with type %s: %s", message.Type, err.Error()))
		return
	}
}

func (h *Handler) readPump(client *entities.WSClient) {
	defer func() {
		h.hub.LeaveAllRooms(client)
		client.Conn.Close()
	}()

	for {
		_, messageBytes, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}

		var message *entities.WSMessage
		err = json.Unmarshal(messageBytes, &message)
		if err != nil {
			slog.ErrorContext(client.Ctx, fmt.Sprintf("unmarshal ws client message: %s", err.Error()))
			break
		}

		err = h.router.Process(client, message)
		if err != nil {
			slog.ErrorContext(client.Ctx, fmt.Sprintf("process message with type %s: %s", message.Type, err.Error()))
			break
		}
	}
}

func (h *Handler) writePump(c *entities.WSClient) {
	for msg := range c.SendCh {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}
