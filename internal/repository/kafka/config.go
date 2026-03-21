package kafka_repository

import (
	"log/slog"
	"sync"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/spf13/viper"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

var (
	onceKafkaConfig sync.Once
	kafkaCfg        *KafkaConfig
)

var topicSpecs = map[constants.KafkaTopic]kafka.TopicSpecification{
	constants.LogsTopic: {
		Topic:             string(constants.LogsTopic),
		NumPartitions:     3,
		ReplicationFactor: 1,
	},
	constants.LogsDLQTopic: {
		Topic:             string(constants.LogsDLQTopic),
		NumPartitions:     1,
		ReplicationFactor: 1,
	},
	constants.AlertsSendTopic: {
		Topic:             string(constants.AlertsSendTopic),
		NumPartitions:     1,
		ReplicationFactor: 1,
	},
}

type KafkaConfig struct {
	ProducerName string
	Servers      []string
	TopicSpecs   map[constants.KafkaTopic]kafka.TopicSpecification
}

func GetConfig() *KafkaConfig {
	onceKafkaConfig.Do(func() {
		kafkaCfg = &KafkaConfig{
			ProducerName: viper.GetString("kafka.producer_name"),
			Servers:      viper.GetStringSlice("kafka.servers"),
			TopicSpecs:   topicSpecs,
		}
	})

	if kafkaCfg == nil {
		slog.Error("common config not loaded")
	}

	return kafkaCfg
}
