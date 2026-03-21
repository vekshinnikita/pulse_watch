package kafka_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type KafkaProducer struct {
	producer *kafka.Producer
}

func NewKafkaProducer() (*KafkaProducer, error) {
	kafkaCfg := GetConfig()

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers":  strings.Join(kafkaCfg.Servers, ","),
		"client.id":          kafkaCfg.ProducerName,
		"acks":               "all",
		"enable.idempotence": true,
		"retries":            3,
		"linger.ms":          10,
	})
	if err != nil {
		return nil, fmt.Errorf("kafka producer init: %w", err)
	}

	p := &KafkaProducer{
		producer: producer,
	}

	err = p.prepare()
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (p *KafkaProducer) prepare() error {
	cfg := GetConfig()

	admin, err := NewKafkaAdmin()
	if err != nil {
		return fmt.Errorf("init kafka admin: %w", err)
	}
	defer admin.Close()

	// Проверяем что все нужные топики созданы
	err = admin.EnsureTopicsExists(utils.MapValues(cfg.TopicSpecs))
	if err != nil {
		return fmt.Errorf("ensure topics exists: %w", err)
	}

	// Добавляем обработку ошибок при отправке сообщений
	p.setErrorHandler()

	return nil
}

func (p *KafkaProducer) setErrorHandler() {
	go func() {
		for e := range p.producer.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					// Логируем ошибку
					args := []any{
						slog.String("topic", *ev.TopicPartition.Topic),
						slog.Int("partition", int(ev.TopicPartition.Partition)),
						slog.String("error", ev.TopicPartition.Error.Error()),
					}

					// Добавляем request_id в лог если он есть
					requestId := GetHeaderString(ev.Headers, "request_id")
					if requestId != "" {
						args = append(args, slog.String("request_id", requestId))
					}

					slog.Error("kafka delivery failed", args...)
				}
			}
		}
	}()
}

func (p *KafkaProducer) Close() {
	slog.Info("Flushing messages to Kafka...")
	p.producer.Flush(5 * 1000)

	slog.Info("Closing kafka producer...")
	p.producer.Close()
}

func (p *KafkaProducer) Publish(
	ctx context.Context,
	options *dtos.MessageOptions,
) error {
	payload, err := json.Marshal(options.Value)
	if err != nil {
		return fmt.Errorf("kafka producer publish marshal value: %w", err)
	}

	// Добавляем дополнительные заголовки
	requestId, ok := ctx.Value(constants.RequestIDKey).(string)
	if ok {
		options.Headers = append(options.Headers, kafka.Header{
			Key:   "request_id",
			Value: []byte(requestId),
		})
	}

	partition := options.Partition
	if partition == nil {
		// если партиция не задана то ставим опцию 'любая партиция'
		partition = utils.ToPtr(int(kafka.PartitionAny))
	}

	strTopic := string(options.Topic)
	return p.producer.Produce(&kafka.Message{
		Headers: options.Headers,
		TopicPartition: kafka.TopicPartition{
			Topic:     &strTopic,
			Partition: int32(*partition),
		},
		Key:   []byte(options.Key),
		Value: payload,
	}, nil)
}

func (p *KafkaProducer) PublishRaw(
	ctx context.Context,
	message *kafka.Message,
) error {
	return p.producer.Produce(message, nil)
}
