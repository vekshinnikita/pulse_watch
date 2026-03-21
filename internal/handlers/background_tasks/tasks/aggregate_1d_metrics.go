package background_tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/hibiken/asynq"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type AggregateDayMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewAggregateDayMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *AggregateDayMetricsTask {
	return &AggregateDayMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *AggregateDayMetricsTask) aggregateUniqueMetrics(
	ctx context.Context,
	appIds []int,
	periodStart time.Time,
) ([]dtos.CreateAppMetric, error) {
	uniqueMetrics, err := t.services.Metric.GetAggregatedUniqueMetricsByApps(
		ctx,
		appIds,
		constants.DayPeriodType,
		periodStart,
	)
	if err != nil {
		return nil, fmt.Errorf("get aggregated unique metrics by apps: %w", err)
	}

	data := make([]dtos.CreateAppMetric, 0)
	for _, item := range uniqueMetrics {
		if item.Value == 0 {
			continue
		}

		data = append(data, dtos.CreateAppMetric{
			AppId:       item.AppId,
			PeriodStart: periodStart,
			PeriodType:  constants.DayPeriodType,
			MetricType:  constants.UniqueUsersMetricType,
			IsUnique:    true,
			Value:       item.Value,
		})
	}

	return data, nil
}

func (t *AggregateDayMetricsTask) aggregateMetrics(
	ctx context.Context,
	appIds []int,
	periodStart time.Time,
	periodEnd time.Time,
) ([]dtos.CreateAppMetric, error) {
	aggregated, err := t.services.Metric.GetAggregatedAppMetricsByPeriod(
		ctx,
		appIds,
		constants.HourPeriodType,
		periodStart,
		periodEnd,
	)
	if err != nil {
		return nil, fmt.Errorf("get aggregated app metrics by period: %w", err)
	}

	data := make([]dtos.CreateAppMetric, 0)
	for _, item := range aggregated {
		if item.Sum == 0 {
			continue
		}

		data = append(data, dtos.CreateAppMetric{
			AppId:       item.AppId,
			PeriodStart: periodStart,
			PeriodType:  constants.DayPeriodType,
			MetricType:  item.MetricType,
			IsUnique:    false,
			Params:      item.Params,
			Value:       item.Sum,
		})
	}

	return data, nil
}

func (t *AggregateDayMetricsTask) aggregateByApps(
	ctx context.Context,
	payload *dtos.WindowStartPayload,
	appIds []int,
) error {
	periodStart := time.Unix(int64(payload.WindowStart), 0)
	periodEnd := periodStart.Add(24 * time.Hour)

	err := utils.ProcessBatches(appIds, 100, func(batchIds []int) error {

		createDataMetrics, err := t.aggregateMetrics(
			ctx,
			batchIds,
			periodStart,
			periodEnd,
		)
		if err != nil {
			return fmt.Errorf("aggregate metrics: %w", err)
		}

		createDataUniqueMetrics, err := t.aggregateUniqueMetrics(
			ctx,
			batchIds,
			periodStart,
		)
		if err != nil {
			return fmt.Errorf("aggregate metrics: %w", err)
		}

		data := append(createDataMetrics, createDataUniqueMetrics...)
		if len(data) == 0 {
			return nil
		}

		_, err = t.services.Metric.CreateAppMetrics(ctx, data)
		if err != nil {
			return fmt.Errorf("create app metrics: %w", err)
		}

		// Отправка метрик в каналы
		err = t.services.Metric.PublishMetricsToChannels(ctx, data)
		if err != nil {
			return fmt.Errorf("publish metrics to channels: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("process batches: %w", err)
	}

	return nil
}

func (t *AggregateDayMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload *dtos.WindowStartPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		slog.Error(fmt.Sprintf("aggregate day: unmarshal payload: %s", err.Error()))
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	ids, err := t.services.App.GetAppIds(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate day: get app ids: %s", err.Error()))
		return fmt.Errorf("get app ids: %w", err)
	}

	err = t.aggregateByApps(ctx, payload, ids)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate day: aggregate by apps: %s", err.Error()))
		return fmt.Errorf("aggregate by apps: %w", err)
	}

	return nil
}
