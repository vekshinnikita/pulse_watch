package metric_service

import (
	"context"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
)

type IncrementMetrics interface {
	Increment(
		ctx context.Context,
		metics dtos.MetricsMap,
		uniqueMetrics dtos.UniqueMetricsMap,
	) error
}

type AggregateOneMinuteMetrics interface {
	Aggregate(
		ctx context.Context,
		appIds []int,
		windowStart time.Time,
	) error
}

type SubscribeChannel interface {
	Subscribe(
		ctx context.Context,
		channelId string,
		handler func(channelId string, message string),
	) error
}

type SendChannelMessage interface {
	Send(
		client *entities.WSClient,
		channelId string,
		incomingMessage string,
	) error
}

type PublishToChannelBulk interface {
	Publish(
		ctx context.Context,
		data []dtos.PublishToChannel,
	) error
}

type PublishMeticsToChannels interface {
	Publish(
		ctx context.Context,
		createMetrics []dtos.CreateAppMetric,
	) error
}

type GetMetrics interface {
	Get(
		ctx context.Context,
		appId int,
		data *entities.GetMetricsData,
	) (*entities.ResultMetricsGroups, error)
}
