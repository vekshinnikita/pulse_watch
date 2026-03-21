package metric_usecases

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type SendChannelMessageUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewSendChannelMessageUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *SendChannelMessageUseCase {
	return &SendChannelMessageUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *SendChannelMessageUseCase) makeMetricsPayload(
	client *entities.WSClient,
	channelId string,
	incomingMessage string,
) *entities.MetricsPayload {
	// Канал вида metric:channel:ten_minutes:6
	channelParts := strings.Split(channelId, ":")
	if len(channelParts) != 4 {
		slog.WarnContext(client.Ctx, fmt.Sprintf("unexpected channel id: %s", channelId))
		return nil
	}

	app_id, err := strconv.Atoi(channelParts[3])
	if err != nil {
		slog.WarnContext(client.Ctx, fmt.Sprintf("can't convert app_id to int: %s", channelParts[2]))
		return nil
	}

	periodType := constants.PeriodType(channelParts[2])

	var message entities.AggregatedMetricsMessage
	err = json.Unmarshal([]byte(incomingMessage), &message)
	if err != nil {
		slog.WarnContext(client.Ctx, fmt.Sprintf("can't unmarshal message: %s", err.Error()))
		return nil
	}

	return &entities.MetricsPayload{
		AppId:       app_id,
		PeriodStart: time.Unix(int64(message.PeriodStart), 0),
		PeriodType:  periodType,
		Metrics:     message.Metrics,
	}
}
func (uc *SendChannelMessageUseCase) Send(
	client *entities.WSClient,
	channelId string,
	incomingMessage string,
) error {

	payload := uc.makeMetricsPayload(client, channelId, incomingMessage)
	if payload != nil {
		client.SendMessage(&entities.WSMessage{
			Type:    "metrics",
			Payload: payload,
		})
	}

	return nil
}
