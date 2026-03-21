package alert_queue_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	telegram_integration "github.com/vekshinnikita/pulse_watch/internal/integretions/telegram"
	"github.com/vekshinnikita/pulse_watch/internal/logger"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type AlertsQueueHandler struct {
	services *service.Service
}

func NewAlertsQueueHandler(services *service.Service) kafka_repository.QueueHandler {
	return &AlertsQueueHandler{
		services: services,
	}
}

func (h *AlertsQueueHandler) addVarsToCtx(ctx context.Context, message *kafka.Message) context.Context {
	appId, ok := kafka_repository.GetHeaderInt(message.Headers, "app_id")
	if ok {
		ctx = context.WithValue(ctx, constants.AppIdCtxKey, appId)
		ctx = logger.AddLogAttrs(ctx, slog.Int("app_id", appId))
	}

	return ctx
}

func (h *AlertsQueueHandler) parseMessage(
	ctx context.Context,
	message *kafka.Message,
) (*entities.SendAlertsMessage, error) {
	var value entities.SendAlertsMessage
	err := json.Unmarshal(message.Value, &value)
	if err != nil {
		return nil, err
	}

	return &value, nil
}

func (h *AlertsQueueHandler) sendAlerts(
	ctx context.Context,
	alerts []models.AlertFull,
) error {
	client, err := telegram_integration.NewDefaultClient()
	if err != nil {
		return fmt.Errorf("new default telegram client: %w", err)
	}

	resolvedAlertIds := make([]int, 0)
	for _, alert := range alerts {
		if alert.Rule.User.TgId == nil {
			continue
		}

		err = client.SendMessage(*alert.Rule.User.TgId, alert.Message)
		if err != nil {
			return fmt.Errorf("send message: %w", err)
		}

		resolvedAlertIds = append(resolvedAlertIds, alert.Id)
	}

	if len(resolvedAlertIds) > 0 {
		err = h.services.Alert.SetResolvedAlerts(ctx, resolvedAlertIds)
		if err != nil {
			return fmt.Errorf("set resolved alerts: %w", err)
		}
	}

	slog.InfoContext(ctx, fmt.Sprintf("%v alerts are resolved", len(resolvedAlertIds)))
	return nil
}

func (h *AlertsQueueHandler) Process(
	ctx context.Context,
	message *kafka.Message,
) (context.Context, error) {
	ctx = h.addVarsToCtx(ctx, message)

	sendAlerts, err := h.parseMessage(ctx, message)
	if err != nil {
		return ctx, fmt.Errorf("logs handler parse message: %w", err)
	}

	alerts, err := h.services.Alert.GetFullAlertsByIds(ctx, sendAlerts.AlertIds)
	if err != nil {
		return ctx, fmt.Errorf("get alerts by ids: %w", err)
	}

	err = h.sendAlerts(ctx, alerts)
	if err != nil {
		return ctx, fmt.Errorf("send alerts: %w", err)
	}

	return ctx, nil
}
