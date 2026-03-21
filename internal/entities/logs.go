package entities

import (
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

type AppLogMeta map[string]any

type AppLog struct {
	Type      constants.LogType  `json:"type" binding:"required,oneof=LOG REQUEST_LOG"`
	Level     constants.LogLevel `json:"level" binding:"required,oneof=DEBUG INFO WARNING ERROR CRITICAL"`
	Timestamp time.Time          `json:"timestamp" binding:"required"`
	Message   string             `json:"message" binding:"required"`
	Meta      AppLogMeta         `json:"meta" binding:"omitempty"`
}

type EnrichedAppLog struct {
	AppLog
	AppId int `json:"app_id"`
}

type AppLogs []AppLog

type SendLogs struct {
	Logs AppLogs `json:"logs" binding:"required,dive"`
}

type LogMetaVarResult struct {
	Name string `db:"name" json:"name" binding:"required"`
	Code string `db:"code" json:"code" binding:"required"`
	Type string `db:"type" json:"type" binding:"required"`
}
