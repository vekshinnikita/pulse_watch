package kafka_repository

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type KafkaConsumer struct {
	ctx      context.Context
	GroupId  string
	consumer *kafka.Consumer
	handler  QueueHandler
	producer MessageProducer
	dlqTopic *string
	stopped  bool
}

type ConsumerParams struct {
	GroupId  string
	Handler  QueueHandler
	Producer MessageProducer
	DLQTopic *string
	Topics   []constants.KafkaTopic
}

func NewKafkaConsumer(params *ConsumerParams) (*KafkaConsumer, error) {
	kafkaCfg := GetConfig()

	consumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(kafkaCfg.Servers, ","),
		"group.id":                 params.GroupId,
		"session.timeout.ms":       60 * 1000,      // 1 мин
		"heartbeat.interval.ms":    1000,           // 1 сек
		"max.poll.interval.ms":     10 * 60 * 1000, // 10 мин
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5 * 1000, // 5 сек
		"auto.offset.reset":        "earliest",
	})
	if err != nil {
		return nil, fmt.Errorf("init kafka consumer: %w", err)
	}

	ctx := context.WithValue(context.Background(), constants.GroupIdCtxKey, params.GroupId)
	logger.AddLogAttrs(ctx, slog.String("consumer_group", params.GroupId))
	c := &KafkaConsumer{
		consumer: consumer,
		GroupId:  params.GroupId,
		handler:  params.Handler,
		producer: params.Producer,
		dlqTopic: params.DLQTopic,
		ctx:      ctx,
		stopped:  false,
	}

	topicsStr := utils.Map(params.Topics, func(topic constants.KafkaTopic) string { return string(topic) })
	err = c.consumer.SubscribeTopics(topicsStr, nil)
	if err != nil {
		return nil, fmt.Errorf("init kafka consumer subscribe topics: %w", err)
	}

	return c, nil
}

func (c *KafkaConsumer) sendToDLQ(ctx context.Context, message *kafka.Message) {
	if c.dlqTopic == nil {
		return
	}

	dlqMessage := &kafka.Message{
		Key:     message.Key,
		Value:   message.Value,
		Headers: message.Headers,
		TopicPartition: kafka.TopicPartition{
			Topic: c.dlqTopic,
		},
	}
	err := c.producer.PublishRaw(ctx, dlqMessage)
	if err != nil {
		slog.ErrorContext(ctx, "unable to publish message to dlq")
	} else {
		slog.InfoContext(ctx, "message send to dlq",
			slog.String("dlq_topic", *c.dlqTopic),
		)
	}

}

func (c *KafkaConsumer) makeMessageContext(message *kafka.Message) context.Context {
	ctx := c.ctx

	// Добавляем request_id в логи если он есть
	requestId := GetHeaderString(message.Headers, "request_id")
	if requestId != "" {
		ctx = context.WithValue(ctx, constants.RequestIDKey, requestId)

		// Добавляем в контекст логов
		ctx = logger.AddLogAttrs(ctx, slog.String("request_id", requestId))
	}

	return ctx
}

func (c *KafkaConsumer) processMessage(ctx context.Context, message *kafka.Message) {
	startTime := time.Now()

	args := []any{
		slog.String("topic", *message.TopicPartition.Topic),
		slog.Int("partition", int(message.TopicPartition.Partition)),
	}

	// Запускаем обработку сообщения
	ctx, err := c.handler.Process(ctx, message)
	args = append(args, slog.Int("duration_ms", int(time.Since(startTime).Milliseconds())))
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("message proceeding failed: %s", err.Error()), args...)

		// Отправляем в dlq
		c.sendToDLQ(ctx, message)
		return
	}

	c.commitOffset(message)

	// Лог об успешной обработке сообщения
	slog.InfoContext(ctx, "message successfully processed", args...)
}

func (c *KafkaConsumer) commitOffset(message *kafka.Message) {
	_, err := c.consumer.StoreMessage(message)
	if err != nil {
		slog.ErrorContext(c.ctx, fmt.Sprintf("storing kafka message: %s", err.Error()))
	}
}

func (c *KafkaConsumer) Start() {
	for {
		if c.stopped {
			break
		}
		message, err := c.consumer.ReadMessage(-1) // без таймаута
		if err != nil {
			slog.ErrorContext(c.ctx, fmt.Sprintf("reading kafka message: %s", err.Error()))
			continue
		}

		ctx := c.makeMessageContext(message)

		defer func() {
			if r := recover(); r != nil {
				slog.Error(fmt.Sprintf(
					"[PANIC] consumer %s crashed, restarting: %v\n%s",
					c.GroupId,
					r,
					debug.Stack(),
				))
				c.sendToDLQ(ctx, message)
				c.commitOffset(message)
				time.Sleep(time.Second)
			}
		}()

		c.processMessage(ctx, message)
	}
}

func (c *KafkaConsumer) Close() error {
	c.stopped = true

	// Коммит локальных offset
	_, err := c.consumer.Commit()
	if err != nil {
		slog.Warn(fmt.Sprintf("consumer commit offsets: %s", err.Error()))
	}

	return c.consumer.Close()
}
