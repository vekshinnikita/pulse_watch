package metric_usecases

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type aggregateOneMinuteMetricsUseCases struct {
	publishMetricsToChannels *PublishMetricsToChannelsUseCase
}

type AggregateOneMinuteMetricsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
	uc   *aggregateOneMinuteMetricsUseCases
}

func NewAggregateOneMinuteMetricsUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *AggregateOneMinuteMetricsUseCase {
	return &AggregateOneMinuteMetricsUseCase{
		repo: repo,
		trm:  trm,
		uc: &aggregateOneMinuteMetricsUseCases{
			publishMetricsToChannels: NewPublishMetricsToChannelsUseCase(trm, repo),
		},
	}
}

func (s *AggregateOneMinuteMetricsUseCase) getMetricTypeAndParams(
	aggregatedMetric dtos.AggregatedMetric,
) (constants.MetricType, *models.JSONB) {
	var params *models.JSONB
	var metricType constants.MetricType

	metricNameParts := strings.Split(aggregatedMetric.MetricName, ":")
	if len(metricNameParts) == 0 {
		return metricType, params
	}

	switch constants.MetricType(metricNameParts[0]) {
	case constants.ErrorsMetricType:
		metricType = constants.ErrorsMetricType

	case constants.WarningsMetricType:
		metricType = constants.WarningsMetricType

	case constants.InfoMetricType:
		metricType = constants.InfoMetricType

	case constants.StatusMetricType:
		metricType = constants.StatusMetricType

		code, _ := strconv.Atoi(metricNameParts[1])
		params = &models.JSONB{
			"code": code,
		}

	case constants.RequestsMetricType:
		metricType = constants.RequestsMetricType

	}

	return metricType, params
}

func (s *AggregateOneMinuteMetricsUseCase) makeCreateData(
	aggregatedMetrics []dtos.AggregatedMetric,
	windowStart time.Time,
) []dtos.CreateAppMetric {
	data := make([]dtos.CreateAppMetric, 0)

	for _, aggregatedMetric := range aggregatedMetrics {
		if aggregatedMetric.Value == 0 {
			continue
		}

		metricType, params := s.getMetricTypeAndParams(aggregatedMetric)
		data = append(data, dtos.CreateAppMetric{
			AppId:       aggregatedMetric.AppId,
			PeriodStart: windowStart,
			PeriodType:  constants.MinutePeriodType,
			MetricType:  metricType,
			IsUnique:    false,
			Params:      params,
			Value:       aggregatedMetric.Value,
		})
	}

	return data
}

func (s *AggregateOneMinuteMetricsUseCase) Aggregate(
	ctx context.Context,
	appIds []int,
	windowStart time.Time,
) error {

	aggregatedMetrics, err := s.repo.AnalyticsRedis.GetAggregatedMetricsByApps(
		ctx,
		appIds,
		windowStart,
		windowStart.Add(59*time.Second),
	)
	if err != nil {
		return fmt.Errorf("get aggregated metrics by apps: %w", err)
	}

	createData := s.makeCreateData(aggregatedMetrics, windowStart)
	if len(createData) == 0 {
		return nil
	}

	_, err = s.repo.Metric.CreateAppMetrics(ctx, createData)
	if err != nil {
		return fmt.Errorf("create app metrics: %w", err)
	}

	// Отправка метрик в каналы
	err = s.uc.publishMetricsToChannels.Publish(ctx, createData)
	if err != nil {
		return fmt.Errorf("publish metrics to channels: %w", err)
	}

	return nil
}
