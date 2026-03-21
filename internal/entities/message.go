package entities

import (
	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

type SendAlertsMessage struct {
	AlertIds []int `json:"alert_ids"`
}

type AggregatedMetric struct {
	MetricType constants.MetricType `json:"metric_type"`
	Params     map[string]any       `json:"params"`
	Value      int                  `json:"value"`
}

type AggregatedMetricsMessage struct {
	PeriodStart int                `json:"period_start"`
	Metrics     []AggregatedMetric `json:"metrics"`
}
