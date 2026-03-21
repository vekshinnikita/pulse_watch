package ws_utils

import (
	"fmt"
	"log/slog"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
)

type RouteHandler func(client *entities.WSClient, payload entities.WSMessagePayload) error

type router struct {
	routes map[string]RouteHandler
}

func NewRouter() Router {
	return &router{
		routes: make(map[string]RouteHandler),
	}
}

func (r *router) Register(messageType string, handler RouteHandler) {
	r.routes[messageType] = handler
}

func (r *router) Process(client *entities.WSClient, message *entities.WSMessage) error {
	handler, ok := r.routes[message.Type]
	if !ok {
		slog.WarnContext(client.Ctx, fmt.Sprintf("There is no handler for message type '%s'", message.Type))
		return nil
	}

	err := handler(client, message.Payload)
	if err != nil {
		return fmt.Errorf("process handler: %w", err)
	}

	return nil
}
