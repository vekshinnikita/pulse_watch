package dtos

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
)

type MessageOptions struct {
	Headers   []kafka.Header
	Partition *int
	Topic     constants.KafkaTopic
	Key       string
	Value     any
}
