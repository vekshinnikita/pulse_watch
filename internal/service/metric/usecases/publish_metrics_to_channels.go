package metric_usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type publishMetricsUseCases struct {
	publishToChannelBulk *PublishToChannelBulkUseCase
}

type PublishMetricsToChannelsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
	uc   *publishMetricsUseCases
}

func NewPublishMetricsToChannelsUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *PublishMetricsToChannelsUseCase {
	return &PublishMetricsToChannelsUseCase{
		repo: repo,
		trm:  trm,
		uc: &publishMetricsUseCases{
			publishToChannelBulk: NewPublishToChannelBulkUseCase(trm, repo),
		},
	}
}

func (u *PublishMetricsToChannelsUseCase) makeDataAndChannelId(
	ctx context.Context,
	createAppMetrics []dtos.CreateAppMetric,
) (string, []byte, error) {

	metrics := make([]entities.AggregatedMetric, 0)
	for _, metric := range createAppMetrics {

		var params map[string]any
		if metric.Params != nil {
			params = *metric.Params
		}

		metrics = append(metrics, entities.AggregatedMetric{
			MetricType: metric.MetricType,
			Params:     params,
			Value:      metric.Value,
		})
	}

	var periodStart time.Time
	var periodType constants.PeriodType
	var appId int
	if len(createAppMetrics) > 0 {
		periodStart = createAppMetrics[0].PeriodStart
		appId = createAppMetrics[0].AppId
		periodType = createAppMetrics[0].PeriodType
	}

	message := &entities.AggregatedMetricsMessage{
		PeriodStart: int(periodStart.Unix()),
		Metrics:     metrics,
	}

	data, err := json.Marshal(message)
	if err != nil {
		return "", nil, fmt.Errorf("marshal message: %w", err)
	}

	channelId := fmt.Sprintf("metric:channel:%s:%d", periodType, appId)

	return channelId, data, nil
}

func (u *PublishMetricsToChannelsUseCase) groupMetricsByAppId(
	ctx context.Context,
	createMetrics []dtos.CreateAppMetric,
) map[int][]dtos.CreateAppMetric {
	groupedByAppId := make(map[int][]dtos.CreateAppMetric)
	for _, createMetric := range createMetrics {
		_, ok := groupedByAppId[createMetric.AppId]
		if !ok {
			groupedByAppId[createMetric.AppId] = make([]dtos.CreateAppMetric, 0)
		}

		groupedByAppId[createMetric.AppId] = append(
			groupedByAppId[createMetric.AppId],
			createMetric,
		)
	}

	return groupedByAppId
}

func (u *PublishMetricsToChannelsUseCase) Publish(
	ctx context.Context,
	createMetrics []dtos.CreateAppMetric,
) error {
	// Группируем по AppId
	groupedByAppId := u.groupMetricsByAppId(ctx, createMetrics)

	// Собираем в формат для отправки в канал
	data := make([]dtos.PublishToChannel, 0)
	for _, appMetrics := range groupedByAppId {

		channelId, message, err := u.makeDataAndChannelId(
			ctx,
			appMetrics,
		)
		if err != nil {
			return fmt.Errorf("make data and channel id: %w", err)
		}

		data = append(data, dtos.PublishToChannel{
			ChannelId: channelId,
			Data:      message,
		})
	}

	// Отправка в каналы
	err := u.uc.publishToChannelBulk.Publish(ctx, data)
	if err != nil {
		return fmt.Errorf("publish to channel bulk: %w", err)
	}

	return nil
}
