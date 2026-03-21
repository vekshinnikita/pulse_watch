package entities

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

type WSMessagePayload any

type WSMessage struct {
	Type    string           `json:"type"`
	Payload WSMessagePayload `json:"payload"`
}

type MetricsPayload struct {
	AppId       int                  `json:"app_id"`
	PeriodStart time.Time            `json:"period_start"`
	PeriodType  constants.PeriodType `json:"period_type"`
	Metrics     []AggregatedMetric   `json:"metrics"`
}

type WSClient struct {
	Id     string
	Conn   *websocket.Conn
	SendCh chan []byte
	Ctx    context.Context
}

func (c *WSClient) SendMessage(msg any) {
	var b []byte

	switch v := msg.(type) {
	case string:
		b = []byte(v)
	case []byte:
		b = v
	default:
		data, err := json.Marshal(v)
		if err != nil {
			return
		}
		b = data
	}

	select {
	case c.SendCh <- b:
	case <-c.Ctx.Done():
	}
}
