package kafka_repository

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type KafkaAdmin struct {
	client *kafka.AdminClient
}

func NewKafkaAdmin() (*KafkaAdmin, error) {
	kafkaCfg := GetConfig()

	client, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": strings.Join(kafkaCfg.Servers, ","),
	})
	if err != nil {
		return nil, fmt.Errorf("Init KafkaAdmin %w:", err)
	}

	return &KafkaAdmin{
		client: client,
	}, nil
}

func (a *KafkaAdmin) Close() {
	a.client.Close()
}

func (a *KafkaAdmin) CreateTopics(
	specs []kafka.TopicSpecification,
) error {
	if len(specs) == 0 {
		return nil
	}

	results, err := a.client.CreateTopics(context.Background(), specs)
	if err != nil {
		return fmt.Errorf("create topics: %w", err)
	}

	// Проверяем что все топики создались
	for _, res := range results {
		if res.Error.Code() != kafka.ErrNoError &&
			res.Error.Code() != kafka.ErrTopicAlreadyExists {
			return res.Error
		}
	}

	slog.Info("Kafka topics are created")

	return nil
}

func (a *KafkaAdmin) EnsureTopicsExists(
	specs []kafka.TopicSpecification,
) error {
	// Получаем информацию о всех топиках
	metadata, err := a.client.GetMetadata(nil, true, 5_000)
	if err != nil {
		return fmt.Errorf("ensure topics get metadata: %w", err)
	}

	// Проверяем какие топики не созданы
	createTopics := make([]kafka.TopicSpecification, 0)
	for _, spec := range specs {
		meta, ok := metadata.Topics[spec.Topic]
		if !ok || meta.Error.Code() == kafka.ErrUnknownTopicOrPart {
			createTopics = append(createTopics, spec)
		}
	}

	// Создаем нужные топики
	err = a.CreateTopics(createTopics)
	if err != nil {
		return fmt.Errorf("ensure topics create topics: %w", err)
	}

	return nil
}
