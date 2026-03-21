package dtos

import (
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type Metrics map[string]int

type UniqueMetric struct {
	Period time.Duration
	Values []any
}

type MetricsMap = map[string]Metrics
type UniqueMetricsMap = map[string]*UniqueMetric

type WindowStartPayload struct {
	WindowStart int `json:"window_start"`
}

type AggregatedMetric struct {
	AppId      int
	MetricName string
	Value      int
}

type TransferredMetric struct {
	Name  string
	Value int
}

type TransferredMetrics struct {
	Key     string
	Metrics []TransferredMetric
}

type CreateAppMetric struct {
	AppId       int
	PeriodStart time.Time
	PeriodType  constants.PeriodType
	MetricType  constants.MetricType
	IsUnique    bool
	Params      *models.JSONB
	Value       int
}

type ClearMetrics []struct {
	LessPeriodStart time.Time
	PeriodType      constants.PeriodType
}

type AggregateMetricsForAlert struct {
	MetricType  constants.MetricType
	PeriodStart time.Time
	PeriodEnd   time.Time
}

type AggregatedAppMetric struct {
	AppId      int                  `db:"app_id"`
	MetricType constants.MetricType `db:"type"`
	Params     *models.JSONB        `db:"params"`
	Sum        int                  `db:"sum"`
}

type AggregatedAppMetricForAlert struct {
	AggregatedAppMetric
	MaxPeriodStart time.Time `db:"max_period_start"`
}

type PublishToChannel struct {
	ChannelId string
	Data      any
}

type GetLiveMetricsData struct {
	AppId      int
	MetricName string
	Start      time.Time
	End        time.Time
}

type LiveMetric struct {
	PeriodStart time.Time
	Value       int
}

type LiveMetrics struct {
	AppId      int
	MetricName string
	Metrics    []LiveMetric
}
