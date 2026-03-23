package background_tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/hibiken/asynq"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type ProcessLiveMetricsTask struct {
	services    *service.Service
	asynqClient *asynq.Client
}

func NewProcessLiveMetricsTask(
	asynqClient *asynq.Client,
	services *service.Service,
) *ProcessLiveMetricsTask {
	return &ProcessLiveMetricsTask{
		asynqClient: asynqClient,
		services:    services,
	}
}

func (t *ProcessLiveMetricsTask) sendOneMinuteMetricsTasks(
	ctx context.Context,
	windowStart time.Time,
) error {
	windowStartUnix := int(windowStart.Unix())
	// Если не начало минуты то пропускаем
	if windowStartUnix%60 != 0 {
		return nil
	}

	// Будем агрегировать за прошедшую минуту
	payload := &dtos.WindowStartPayload{
		WindowStart: windowStartUnix - 60,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Отправка задачи на агрегацию метрик за 1 минуту
	_, err = t.asynqClient.Enqueue(
		asynq.NewTask(constants.AggregateOneMinuteMetricsType, payloadBytes),
		asynq.MaxRetry(3),
	)
	if err != nil {
		return fmt.Errorf("send aggregate one minute metrics task: %w", err)
	}

	return nil
}

func (t *ProcessLiveMetricsTask) sendOtherMetricsTasks(
	ctx context.Context,
	windowStart time.Time,
) error {
	err := t.sendOneMinuteMetricsTasks(ctx, windowStart)
	if err != nil {
		return fmt.Errorf("send one minute metrics tasks:%w", err)
	}

	return nil
}

func (s *ProcessLiveMetricsTask) makeMetricParams(metricNameParts []string) map[string]any {
	switch metricNameParts[0] {
	case "status":
		code, _ := strconv.Atoi(metricNameParts[1])
		return map[string]any{
			"code": code,
		}
	}

	return nil
}

func (t *ProcessLiveMetricsTask) makeDataAndChannelId(
	ctx context.Context,
	windowStart time.Time,
	transferred dtos.TransferredMetrics,
) (string, []byte, error) {
	// Ключ вида metric:live:{app_id}:{time_sec}
	channelParts := strings.Split(transferred.Key, ":")
	if len(channelParts) != 4 {
		return "", nil, fmt.Errorf("unexpected key: %s", transferred.Key)
	}

	appId, err := strconv.Atoi(channelParts[2])
	if err != nil {
		return "", nil, fmt.Errorf("can't convert app_id to int: %s", channelParts[2])
	}

	metrics := make([]entities.AggregatedMetric, 0)
	for _, metric := range transferred.Metrics {
		metricNameParts := strings.Split(metric.Name, ":")

		metrics = append(metrics, entities.AggregatedMetric{
			MetricType: constants.MetricType(metricNameParts[0]),
			Params:     t.makeMetricParams(metricNameParts),
			Value:      metric.Value,
		})
	}

	message := &entities.AggregatedMetricsMessage{
		PeriodStart: int(windowStart.Unix()),
		Metrics:     metrics,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return "", nil, fmt.Errorf("marshal message: %w", err)
	}

	channelId := fmt.Sprintf("metric:channel:live:%d", appId)

	return channelId, data, nil
}

func (t *ProcessLiveMetricsTask) sendMetricsToChannel(
	ctx context.Context,
	windowStart time.Time,
	transferredMetrics []dtos.TransferredMetrics,
) error {
	data := make([]dtos.PublishToChannel, 0)
	for _, transferredMetric := range transferredMetrics {
		channelId, message, err := t.makeDataAndChannelId(ctx, windowStart, transferredMetric)
		if err != nil {
			return fmt.Errorf("make data and channel id: %w", err)
		}

		data = append(data, dtos.PublishToChannel{
			ChannelId: channelId,
			Data:      message,
		})
	}

	err := t.services.Metric.PublishToChannelBulk(ctx, data)
	if err != nil {
		return fmt.Errorf("publish to channel bulk: %w", err)
	}

	return nil
}

func (t *ProcessLiveMetricsTask) processLiveMetrics(
	ctx context.Context,
	windowStart time.Time,
) error {
	pattern := fmt.Sprintf("metric:live:*:%d", windowStart.Unix())

	var cursor uint64
	for {
		// Получение батча ключей для переброса метрик
		keys, nextCursor, err := t.services.Metric.GetKeysCursor(
			ctx, cursor, pattern, 100,
		)
		if err != nil {
			return fmt.Errorf("get keys cursor: %w", err)
		}

		if len(keys) == 0 {
			break
		}

		// Переброс метрик в streams
		transferredMetrics, err := t.services.Metric.TransferLiveMetricsToStreams(ctx, keys)
		if err != nil {
			return fmt.Errorf("transfer live metrics to streams: %w", err)
		}

		if len(transferredMetrics) > 0 {
			// Отправка метрик в канал
			err = t.sendMetricsToChannel(ctx, windowStart, transferredMetrics)
			if err != nil {
				return fmt.Errorf("transfer live metrics to streams: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (t *ProcessLiveMetricsTask) getWindowStart() time.Time {
	currentWindowStart := (time.Now().Unix() / constants.BaseMetricWindowTime) * constants.BaseMetricWindowTime

	windowStartWithDelay := currentWindowStart - constants.BaseMetricCommitDelay
	return time.Unix(int64(windowStartWithDelay), 0)
}

func (t *ProcessLiveMetricsTask) Process(
	ctx context.Context,
	task *asynq.Task,
) error {
	windowStart := t.getWindowStart()

	err := t.processLiveMetrics(ctx, windowStart)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate live: process live metrics: %s", err.Error()))
		return fmt.Errorf("process live metrics: %w", err)
	}

	// Отправка задачи на агрегацию метрик
	err = t.sendOtherMetricsTasks(ctx, windowStart)
	if err != nil {
		slog.Error(fmt.Sprintf("aggregate live: send aggregate alert metrics task: %s", err.Error()))
		return fmt.Errorf("send aggregate alert metrics task: %w", err)
	}

	return nil
}
