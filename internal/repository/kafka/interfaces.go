package kafka_repository

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
)

type QueueHandler interface {
	Process(ctx context.Context, message *kafka.Message) (context.Context, error)
}

type MessageProducer interface {
	Publish(
		ctx context.Context,
		options *dtos.MessageOptions,
	) error

	PublishRaw(
		ctx context.Context,
		message *kafka.Message,
	) error

	Close()
}
