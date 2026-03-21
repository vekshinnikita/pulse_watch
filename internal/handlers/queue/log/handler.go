package log_queue_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	log_queue_processors "github.com/vekshinnikita/pulse_watch/internal/handlers/queue/log/processors"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type handlerProcessors struct {
	newMetaVars Processor
	metrics     Processor
	storeLogs   Processor
	alert       Processor
}

type LogsQueueHandler struct {
	services   *service.Service
	processors *handlerProcessors
}

func NewLogsQueueHandler(services *service.Service) kafka_repository.QueueHandler {
	return &LogsQueueHandler{
		services: services,
		processors: &handlerProcessors{
			newMetaVars: log_queue_processors.NewMetaVarsProcessor(services),
			storeLogs:   log_queue_processors.NewStoreLogsProcessor(services),
			alert:       log_queue_processors.NewAlertProcessor(services),
			metrics:     log_queue_processors.NewMetricsProcessor(services),
		},
	}
}

func (h *LogsQueueHandler) addVarsToCtx(ctx context.Context, message *kafka.Message) context.Context {
	appId, ok := kafka_repository.GetHeaderInt(message.Headers, "app_id")
	if ok {
		ctx = context.WithValue(ctx, constants.AppIdCtxKey, appId)
		ctx = logger.AddLogAttrs(ctx, slog.Int("app_id", appId))
	}

	return ctx
}

func (h *LogsQueueHandler) parseMessage(
	ctx context.Context,
	message *kafka.Message,
) (entities.AppLogs, error) {
	var value entities.AppLogs
	err := json.Unmarshal(message.Value, &value)
	if err != nil {
		return nil, err
	}

	return value, nil
}

func (h *LogsQueueHandler) enrichLogs(
	ctx context.Context,
	logs entities.AppLogs,
) ([]entities.EnrichedAppLog, error) {
	enrichedLogs := make([]entities.EnrichedAppLog, 0)

	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("unable to get app id")
	}

	for _, log := range logs {
		enrichedLogs = append(enrichedLogs, entities.EnrichedAppLog{
			AppLog: log,
			AppId:  appId,
		})
	}

	return enrichedLogs, nil
}

func (h *LogsQueueHandler) getBehavior() *Behavior[[]entities.EnrichedAppLog] {
	behavior := NewDefaultBehavior[[]entities.EnrichedAppLog]()

	// Обработчик сохранения логов
	behavior.AddAsync(h.processors.storeLogs)

	// Обработчик новых мета переменных
	behavior.AddAsync(h.processors.newMetaVars)

	// Обработчик подсчета метрик
	behavior.AddSync(h.processors.metrics)

	// Обработчик алертов
	// Должен запускать после подсчета метрик
	behavior.AddSync(h.processors.alert)

	return behavior
}

func (h *LogsQueueHandler) Process(ctx context.Context, message *kafka.Message) (context.Context, error) {
	ctx = h.addVarsToCtx(ctx, message)

	logs, err := h.parseMessage(ctx, message)
	if err != nil {
		return ctx, fmt.Errorf("logs handler parse message: %w", err)
	}

	// Обогащение логов
	enrichedLogs, err := h.enrichLogs(ctx, logs)
	if err != nil {
		return ctx, fmt.Errorf("logs handler enrich logs: %w", err)
	}

	// Запуск обработчиков
	behavior := h.getBehavior()
	err = behavior.Execute(ctx, enrichedLogs)
	if err != nil {
		return ctx, fmt.Errorf("log behavior execute: %w", err)
	}

	return ctx, nil
}
