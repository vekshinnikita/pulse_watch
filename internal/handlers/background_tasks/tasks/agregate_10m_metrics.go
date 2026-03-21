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

type AggregateTenMinutesMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewAggregateTenMinutesMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *AggregateTenMinutesMetricsTask {
	return &AggregateTenMinutesMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *AggregateTenMinutesMetricsTask) makeCreateData(
	periodStart time.Time,
	aggregated []dtos.AggregatedAppMetric,
) []dtos.CreateAppMetric {
	data := make([]dtos.CreateAppMetric, 0)

	for _, item := range aggregated {
		if item.Sum == 0 {
			continue
		}

		data = append(data, dtos.CreateAppMetric{
			AppId:       item.AppId,
			PeriodStart: periodStart,
			PeriodType:  constants.TenMinutesPeriodType,
			MetricType:  item.MetricType,
			IsUnique:    false,
			Params:      item.Params,
			Value:       item.Sum,
		})
	}

	return data
}

func (t *AggregateTenMinutesMetricsTask) aggregateByApps(
	ctx context.Context,
	payload *dtos.WindowStartPayload,
	appIds []int,
) error {
	periodStart := time.Unix(int64(payload.WindowStart), 0)
	periodEnd := periodStart.Add(10 * time.Minute)

	err := utils.ProcessBatches(appIds, 100, func(batchIds []int) error {
		aggregated, err := t.services.Metric.GetAggregatedAppMetricsByPeriod(
			ctx,
			batchIds,
			constants.MinutePeriodType,
			periodStart,
			periodEnd,
		)
		if err != nil {
			return fmt.Errorf("get aggregated app metrics by period: %w", err)
		}

		data := t.makeCreateData(periodStart, aggregated)
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

func (t *AggregateTenMinutesMetricsTask) sendOtherMetricsTasks(
	ctx context.Context,
	p *dtos.WindowStartPayload,
) error {
	windowStartUnix := p.WindowStart + 10*60

	hourSec := 60 * 60
	// Если не начало часа то пропускаем
	if windowStartUnix%hourSec != 0 {
		return nil
	}

	payload := &dtos.WindowStartPayload{
		WindowStart: windowStartUnix - hourSec,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Отправка задачи на агрегацию метрик за час
	_, err = t.asynqClient.Enqueue(
		asynq.NewTask(constants.AggregateHourMetricsType, payloadBytes),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("send aggregate hour metrics task: %w", err)
	}

	return nil
}

func (t *AggregateTenMinutesMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload *dtos.WindowStartPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		slog.Error(fmt.Sprintf("aggregate ten minutes: unmarshal payload: %s", err.Error()))
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	ids, err := t.services.App.GetAppIds(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate ten minutes: get app ids: %s", err.Error()))
		return fmt.Errorf("get app ids: %w", err)
	}

	err = t.aggregateByApps(ctx, payload, ids)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate ten minutes: aggregate by apps: %s", err.Error()))
		return fmt.Errorf("aggregate by apps: %w", err)
	}

	// Отправка задачи на агрегацию метрик
	err = t.sendOtherMetricsTasks(ctx, payload)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate ten minutes: send aggregate alert metrics task: %s", err.Error()))
		return fmt.Errorf("send aggregate alert metrics task: %w", err)
	}

	return nil
}
