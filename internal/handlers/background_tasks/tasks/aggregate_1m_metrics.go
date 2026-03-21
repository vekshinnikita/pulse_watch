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

type AggregateOneMinuteMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewAggregateOneMinuteMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *AggregateOneMinuteMetricsTask {
	return &AggregateOneMinuteMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *AggregateOneMinuteMetricsTask) aggregateByApps(
	ctx context.Context,
	payload *dtos.WindowStartPayload,
	appIds []int,
) error {
	windowStart := time.Unix(int64(payload.WindowStart), 0)

	err := utils.ProcessBatches(appIds, 100, func(batchIds []int) error {
		err := t.services.Metric.AggregateOneMinuteMetrics(
			ctx,
			batchIds,
			windowStart,
		)
		if err != nil {
			return fmt.Errorf("aggregate metrics by apps: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("process batches: %w", err)
	}

	return nil
}

func (t *AggregateOneMinuteMetricsTask) sendOtherMetricsTasks(
	ctx context.Context,
	p *dtos.WindowStartPayload,
) error {
	windowStartUnix := p.WindowStart + 60

	tenMinutesSec := 10 * 60
	// Если не начало 10-ой минуты то пропускаем
	if windowStartUnix%tenMinutesSec != 0 {
		return nil
	}

	payload := &dtos.WindowStartPayload{
		WindowStart: windowStartUnix - tenMinutesSec,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Отправка задачи на агрегацию метрик за 10 минуту
	_, err = t.asynqClient.Enqueue(
		asynq.NewTask(constants.AggregateTenMinutesMetricsType, payloadBytes),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("send aggregate ten minutes metrics task: %w", err)
	}

	return nil
}

func (t *AggregateOneMinuteMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	var payload *dtos.WindowStartPayload
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		slog.Error(fmt.Sprintf("unmarshal payload: %s", err.Error()))
		return fmt.Errorf("unmarshal payload: %w", err)
	}

	ids, err := t.services.App.GetAppIds(ctx)
	if err != nil {
		slog.Error(fmt.Sprintf("get app ids: %s", err.Error()))
		return fmt.Errorf("get app ids: %w", err)
	}

	err = t.aggregateByApps(ctx, payload, ids)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate by apps: %s", err.Error()))
		return fmt.Errorf("aggregate by apps: %w", err)
	}

	// Отправка задачи на агрегацию метрик
	err = t.sendOtherMetricsTasks(ctx, payload)
	if err != nil {
		slog.Error(fmt.Sprintf("send aggregate alert metrics task: %s", err.Error()))
		return fmt.Errorf("send aggregate alert metrics task: %w", err)
	}

	return nil
}
