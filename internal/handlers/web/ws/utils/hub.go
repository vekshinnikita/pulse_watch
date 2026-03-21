package ws_utils

import (
	"context"
	"sync"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
)

type clientSendFunc func(
	client *entities.WSClient,
	roomId string,
	message string,
) error

type roomSubscribeFunc func(
	ctx context.Context,
	roomId string,
	handler func(channelId string, message string),
) error

type Room struct {
	clients map[*entities.WSClient]struct{}
	cancel  context.CancelFunc
}

type hub struct {
	clientRooms map[*entities.WSClient]map[string]struct{}
	rooms       map[string]*Room
	mu          sync.RWMutex

	clientSend    clientSendFunc
	roomSubscribe roomSubscribeFunc
}

func NewHub(roomSubscribe roomSubscribeFunc, clientSend clientSendFunc) Hub {
	return &hub{
		rooms:       make(map[string]*Room),
		clientRooms: make(map[*entities.WSClient]map[string]struct{}),

		clientSend:    clientSend,
		roomSubscribe: roomSubscribe,
	}
}

func (h *hub) subscribeToRoom(ctx context.Context, roomID string) {
	h.roomSubscribe(ctx, roomID, func(channelId string, msg string) {
		h.BroadcastByRoom(roomID, msg)
	})
}

func (h *hub) BroadcastByRoom(roomId string, message string) {
	h.mu.RLock()
	room, ok := h.rooms[roomId]
	h.mu.RUnlock()

	if !ok {
		return
	}

	for client := range room.clients {
		h.clientSend(client, roomId, message)
	}
}

func (h *hub) JoinRoom(roomId string, client *entities.WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.rooms[roomId] == nil {
		ctx, cancel := context.WithCancel(context.Background())

		h.rooms[roomId] = &Room{
			clients: make(map[*entities.WSClient]struct{}),
			cancel:  cancel,
		}

		// Подписываемся на комнату
		h.subscribeToRoom(ctx, roomId)
	}

	h.rooms[roomId].clients[client] = struct{}{}

	if h.clientRooms[client] == nil {
		h.clientRooms[client] = make(map[string]struct{})
	}

	h.clientRooms[client][roomId] = struct{}{}
}

func (h *hub) LeaveRoom(roomId string, client *entities.WSClient) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[roomId]; ok {
		delete(room.clients, client)

		// Если подключенных клиентов к комнате не осталось
		// отписываемся от комнаты
		if len(room.clients) == 0 {
			room.cancel()
			delete(h.rooms, roomId)
		}
	}

	// Удаляем комнату из списка у клиента
	if rooms, ok := h.clientRooms[client]; ok {
		delete(rooms, roomId)
	}
}

func (h *hub) LeaveAllRooms(client *entities.WSClient) {
	h.mu.Lock()
	rooms, ok := h.clientRooms[client]
	h.mu.Unlock()

	if !ok {
		return
	}

	for roomId := range rooms {
		h.LeaveRoom(roomId, client)
	}
}
