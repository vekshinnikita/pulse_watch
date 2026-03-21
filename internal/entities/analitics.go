package entities

import (
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

type GetMetricsData struct {
	PeriodType  constants.PeriodType  `json:"period_type" binding:"required,enum=PeriodType"`
	MetricType  *constants.MetricType `json:"metric_type" binding:"omitempty,enum=MetricType"`
	Params      map[string]any        `json:"params"`
	PeriodStart *time.Time            `json:"period_start"`
	PeriodEnd   *time.Time            `json:"period_end"`
}

type SearchLogMeta struct {
	Type       constants.MetaVarType `json:"type" binding:"enum=MetaVarType"`
	FilterType constants.FilterType  `json:"filter_type" binding:"enum=FilterType"`
	Name       string                `json:"name" binding:"required"`
	Value      any                   `json:"value" binding:"omitempty"`
}

type SearchLogData struct {
	Query *string             `json:"query" binding:"omitempty"`
	AppId *int                `json:"app_id" binding:"omitempty"`
	Start *time.Time          `json:"start" binding:"omitempty"`
	End   *time.Time          `json:"end" binding:"omitempty"`
	Type  *constants.LogType  `json:"type" binding:"omitempty,enum=LogType"`
	Level *constants.LogLevel `json:"level" binding:"omitempty,enum=LogLevel"`
	Meta  []SearchLogMeta     `json:"meta" binding:"omitempty"`

	PaginationData
}

type ResultMetrics struct {
	PeriodStart time.Time `json:"period_start"`
	Value       int       `json:"value"`
}

type ResultMetricsGroup struct {
	MetricType constants.MetricType `json:"metric_type" binding:"enum=MetricType"`
	Params     map[string]any       `json:"params"`
	Metrics    []ResultMetrics      `json:"metrics"`
}

type ResultMetricsGroups struct {
	PeriodType constants.PeriodType `json:"period_type" binding:"required,enum=PeriodType"`
	Groups     []ResultMetricsGroup `json:"groups"`
}
