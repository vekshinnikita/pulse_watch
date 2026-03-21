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

type AggregateHourMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewAggregateHourMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *AggregateHourMetricsTask {
	return &AggregateHourMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *AggregateHourMetricsTask) aggregateUniqueMetrics(
	ctx context.Context,
	appIds []int,
	periodStart time.Time,
) ([]dtos.CreateAppMetric, error) {
	uniqueMetrics, err := t.services.Metric.GetAggregatedUniqueMetricsByApps(
		ctx,
		appIds,
		constants.HourPeriodType,
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
			PeriodType:  constants.HourPeriodType,
			MetricType:  constants.UniqueUsersMetricType,
			IsUnique:    true,
			Value:       item.Value,
		})
	}

	return data, nil
}

func (t *AggregateHourMetricsTask) aggregateMetrics(
	ctx context.Context,
	appIds []int,
	periodStart time.Time,
	periodEnd time.Time,
) ([]dtos.CreateAppMetric, error) {
	aggregated, err := t.services.Metric.GetAggregatedAppMetricsByPeriod(
		ctx,
		appIds,
		constants.TenMinutesPeriodType,
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
			PeriodType:  constants.HourPeriodType,
			MetricType:  item.MetricType,
			IsUnique:    false,
			Params:      item.Params,
			Value:       item.Sum,
		})
	}

	return data, nil
}

func (t *AggregateHourMetricsTask) aggregateByApps(
	ctx context.Context,
	payload *dtos.WindowStartPayload,
	appIds []int,
) error {
	periodStart := time.Unix(int64(payload.WindowStart), 0)
	periodEnd := periodStart.Add(time.Hour)

	err := utils.ProcessBatches(appIds, 100, func(batchIds []int) error {
		// Агрегация метрик
		createDataMetrics, err := t.aggregateMetrics(
			ctx,
			batchIds,
			periodStart,
			periodEnd,
		)
		if err != nil {
			return fmt.Errorf("aggregate metrics: %w", err)
		}

		// Агрегация уникальных метрик
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

		// Сохранение метрик
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

func (t *AggregateHourMetricsTask) sendOtherMetricsTasks(
	ctx context.Context,
	p *dtos.WindowStartPayload,
) error {
	windowStartUnix := p.WindowStart + 60*60

	daySec := 24 * 60 * 60
	// Если не начало суток то пропускаем
	if windowStartUnix%daySec != 0 {
		return nil
	}

	payload := &dtos.WindowStartPayload{
		WindowStart: windowStartUnix - daySec,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Отправка задачи на агрегацию метрик за день
	_, err = t.asynqClient.Enqueue(
		asynq.NewTask(constants.AggregateDayMetricsType, payloadBytes),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("send aggregate ten minutes metrics task: %w", err)
	}

	return nil
}

func (t *AggregateHourMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload *dtos.WindowStartPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		slog.Error(fmt.Sprintf("aggregate hour: unmarshal payload: %s", err.Error()))
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	ids, err := t.services.App.GetAppIds(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate hour: get app ids: %s", err.Error()))
		return fmt.Errorf("get app ids: %w", err)
	}

	err = t.aggregateByApps(ctx, payload, ids)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate hour: aggregate by apps: %s", err.Error()))
		return fmt.Errorf("aggregate by apps: %w", err)
	}

	// Отправка задачи на агрегацию метрик
	err = t.sendOtherMetricsTasks(ctx, payload)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate hour: send aggregate alert metrics task: %s", err.Error()))
		return fmt.Errorf("send aggregate alert metrics task: %w", err)
	}

	return nil
}
