package ws_utils

import "github.com/vekshinnikita/pulse_watch/internal/entities"

type Router interface {
	Register(messageType string, handler RouteHandler)
	Process(client *entities.WSClient, message *entities.WSMessage) error
}

type Hub interface {
	BroadcastByRoom(roomId string, message string)
	JoinRoom(roomId string, client *entities.WSClient)
	LeaveRoom(roomId string, client *entities.WSClient)
	LeaveAllRooms(client *entities.WSClient)
}
