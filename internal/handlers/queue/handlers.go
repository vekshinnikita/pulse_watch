package queue_handlers

import (
	"fmt"
	"log/slog"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type HandlerOptions struct {
	Handler       kafka_repository.QueueHandler
	WorkersCount  int
	DLQTopic      *string
	Topics        []constants.KafkaTopic
	ConsumerGroup string
}

type QueueHandlers struct {
	producer        repository.MessageProducer
	handlersOptions []HandlerOptions
	consumers       []kafka_repository.KafkaConsumer
}

func getHandlersOptions(services *service.Service) []HandlerOptions {
	cfg := GetConfig()
	options := make([]HandlerOptions, 0)

	for _, item := range cfg {
		handlerConstructor, ok := handlerConstructorsMap[item.Handler]
		if !ok {
			continue
		}

		options = append(options, HandlerOptions{
			Handler:      handlerConstructor(services),
			WorkersCount: item.WorkersCount,
			DLQTopic:     item.DLQTopic,
			Topics: utils.Map(item.Topics, func(v string) constants.KafkaTopic {
				return constants.KafkaTopic(v)
			}),
			ConsumerGroup: item.ConsumerGroup,
		})
	}

	return options
}

func NewQueueHandlers(
	services *service.Service,
	producer repository.MessageProducer,
) *QueueHandlers {
	return &QueueHandlers{
		producer:        producer,
		handlersOptions: getHandlersOptions(services),
	}
}

func (h *QueueHandlers) StopConsumers() {
	slog.Info("Stopping kafka consumers...")

	for _, consumer := range h.consumers {
		err := consumer.Close()
		if err != nil {
			slog.Error(err.Error())
		}
	}
}

func (h *QueueHandlers) RunConsumers() error {
	// Инициализируем consumers
	for _, options := range h.handlersOptions {

		for i := 0; i < options.WorkersCount; i++ {
			consumer, err := kafka_repository.NewKafkaConsumer(&kafka_repository.ConsumerParams{
				Producer: h.producer,
				GroupId:  options.ConsumerGroup,
				Handler:  options.Handler,
				Topics:   options.Topics,
				DLQTopic: options.DLQTopic,
			})
			if err != nil {
				return fmt.Errorf("run consumers: %w", err)
			}

			h.consumers = append(h.consumers, *consumer)
		}
	}

	// Запускаем обработчики
	for _, consumer := range h.consumers {
		go consumer.Start()
	}

	return nil
}
