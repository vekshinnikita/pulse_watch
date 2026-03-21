package metric_service

import (
	"context"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	metric_usecases "github.com/vekshinnikita/pulse_watch/internal/service/metric/usecases"
)

type MetricUseCases struct {
	incrementMetrics          IncrementMetrics
	aggregateOneMinuteMetrics AggregateOneMinuteMetrics
	subscribeChannel          SubscribeChannel
	sendChannelMessage        SendChannelMessage
	publishToChannelBulk      PublishToChannelBulk
	publishMeticsToChannels   PublishMeticsToChannels
	getMetrics                GetMetrics
}

type MetricService struct {
	trm  repository.TransactionManager
	repo *repository.Repository
	uc   *MetricUseCases
}

func NewDefaultMetricService(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *MetricService {
	return &MetricService{
		repo: repo,
		trm:  trm,
		uc: &MetricUseCases{
			incrementMetrics:          metric_usecases.NewIncrementMetricsUseCase(trm, repo),
			aggregateOneMinuteMetrics: metric_usecases.NewAggregateOneMinuteMetricsUseCase(trm, repo),
			subscribeChannel:          metric_usecases.NewSubscribeChannelUseCase(trm, repo),
			sendChannelMessage:        metric_usecases.NewSendChannelMessageUseCase(trm, repo),
			publishToChannelBulk:      metric_usecases.NewPublishToChannelBulkUseCase(trm, repo),
			publishMeticsToChannels:   metric_usecases.NewPublishMetricsToChannelsUseCase(trm, repo),
			getMetrics:                metric_usecases.NewGetMetricsUseCase(trm, repo),
		},
	}
}

func (s *MetricService) IncrementMetrics(
	ctx context.Context,
	metrics dtos.MetricsMap,
	uniqueMetrics dtos.UniqueMetricsMap,
) error {
	return s.uc.incrementMetrics.Increment(ctx, metrics, uniqueMetrics)
}

func (s *MetricService) GetKeysCursor(
	ctx context.Context,
	cursor uint64,
	pattern string,
	count int,
) ([]string, uint64, error) {
	return s.repo.AnalyticsRedis.GetKeysCursor(ctx, cursor, pattern, count)
}

func (s *MetricService) TransferLiveMetricsToStreams(
	ctx context.Context,
	metricKeys []string,
) ([]dtos.TransferredMetrics, error) {
	return s.repo.AnalyticsRedis.TransferLiveMetricsToStreams(ctx, metricKeys)
}

func (s *MetricService) AggregateOneMinuteMetrics(
	ctx context.Context,
	appIds []int,
	windowStart time.Time,
) error {
	return s.uc.aggregateOneMinuteMetrics.Aggregate(
		ctx,
		appIds,
		windowStart,
	)
}

func (s *MetricService) CreateAppMetrics(
	ctx context.Context,
	data []dtos.CreateAppMetric,
) ([]int, error) {
	return s.repo.Metric.CreateAppMetrics(ctx, data)
}

func (s *MetricService) GetAggregatedAppMetricsByPeriod(
	ctx context.Context,
	appIds []int,
	periodType constants.PeriodType,
	startTime time.Time,
	endTime time.Time,
) ([]dtos.AggregatedAppMetric, error) {
	return s.repo.Metric.GetAggregatedAppMetricsByPeriod(ctx, appIds, periodType, startTime, endTime)
}

func (s *MetricService) GetAggregatedUniqueMetricsByApps(
	ctx context.Context,
	appIds []int,
	periodType constants.PeriodType,
	windowStart time.Time,
) ([]dtos.AggregatedMetric, error) {
	return s.repo.AnalyticsRedis.GetAggregatedUniqueMetricsByApps(ctx, appIds, periodType, windowStart)
}

func (s *MetricService) ClearMetrics(
	ctx context.Context,
	data dtos.ClearMetrics,
) error {
	return s.repo.Metric.ClearMetrics(ctx, data)
}

func (s *MetricService) SubscribeChannel(
	ctx context.Context,
	channelId string,
	handler func(channelId string, message string),
) error {
	return s.uc.subscribeChannel.Subscribe(ctx, channelId, handler)
}

func (s *MetricService) SendChannelMessage(
	client *entities.WSClient,
	channelId string,
	incomingMessage string,
) error {
	return s.uc.sendChannelMessage.Send(client, channelId, incomingMessage)
}

func (s *MetricService) PublishToChannelBulk(ctx context.Context, data []dtos.PublishToChannel) error {
	return s.uc.publishToChannelBulk.Publish(ctx, data)
}

func (s *MetricService) PublishMetricsToChannels(
	ctx context.Context,
	createMetrics []dtos.CreateAppMetric,
) error {
	return s.uc.publishMeticsToChannels.Publish(ctx, createMetrics)
}

func (s *MetricService) GetMetrics(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) (*entities.ResultMetricsGroups, error) {
	return s.uc.getMetrics.Get(ctx, appId, data)
}
