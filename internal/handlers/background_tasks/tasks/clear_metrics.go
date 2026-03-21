package background_tasks

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type ClearMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewClearMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *ClearMetricsTask {
	return &ClearMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *ClearMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	now := time.Now()
	dayDuration := 24 * time.Hour
	monthDuration := 30 * dayDuration

	clearMetricsData := dtos.ClearMetrics{
		{
			LessPeriodStart: now.Add(-dayDuration),
			PeriodType:      constants.MinutePeriodType,
		},
		{
			LessPeriodStart: now.Add(-7 * dayDuration),
			PeriodType:      constants.TenMinutesPeriodType,
		},
		{
			LessPeriodStart: now.Add(-monthDuration),
			PeriodType:      constants.HourPeriodType,
		},
		{
			LessPeriodStart: now.Add(-6 * monthDuration),
			PeriodType:      constants.DayPeriodType,
		},
	}

	err := t.services.Metric.ClearMetrics(ctx, clearMetricsData)
	if err != nil {
		slog.Error(fmt.Sprintf("clear metrics: %s", err.Error()))
		return fmt.Errorf("clear metrics: %w", err)
	}

	return nil
}
