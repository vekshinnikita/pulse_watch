package analytics_ws_handler

import "github.com/vekshinnikita/pulse_watch/internal/entities"

type WSRouter interface {
	Process(client *entities.WSClient, message *entities.WSMessage) error
}
