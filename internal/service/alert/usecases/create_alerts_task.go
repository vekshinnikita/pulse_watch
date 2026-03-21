package alert_usecases

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type CreateAlertsTaskUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewCreateAlertsTaskUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *CreateAlertsTaskUseCase {
	return &CreateAlertsTaskUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *CreateAlertsTaskUseCase) createAlertsTask(
	ctx context.Context,
	alertIds []int,
) error {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return fmt.Errorf("cannot get %s from context", constants.AppIdCtxKey)
	}

	messageOption := &dtos.MessageOptions{
		Headers: []kafka.Header{
			{
				Key:   "app_id",
				Value: []byte(strconv.Itoa(appId)),
			},
		},
		Topic: constants.AlertsSendTopic,
		Value: entities.SendAlertsMessage{
			AlertIds: alertIds,
		},
	}

	err := uc.repo.Producer.Publish(ctx, messageOption)
	if err != nil {
		return fmt.Errorf("publish to the queue: %w", err)
	}
	return nil
}

func (uc *CreateAlertsTaskUseCase) Create(
	ctx context.Context,
	data []dtos.CreateAlert,
) error {

	// Делаем это в транзакции
	err := uc.trm.Do(ctx, func(с context.Context) error {
		// Создание алертов в хранилище
		alertIds, err := uc.repo.Alert.CreateAlerts(с, data)
		if err != nil {
			return fmt.Errorf("create alerts in storage: %w", err)
		}

		// Отправка задачи на отправку алертов
		err = uc.createAlertsTask(ctx, alertIds)
		if err != nil {
			return fmt.Errorf("create alerts task: %w", err)
		}

		return nil
	})

	slog.InfoContext(ctx, "task for sending alerts are created")
	return err
}
