package queue_handlers

import (
	alert_queue_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/queue/alert"
	log_queue_handler "github.com/vekshinnikita/pulse_watch/internal/handlers/queue/log"
	kafka_repository "github.com/vekshinnikita/pulse_watch/internal/repository/kafka"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type HandlerConstructor func(s *service.Service) kafka_repository.QueueHandler

var handlerConstructorsMap map[string]HandlerConstructor = map[string]HandlerConstructor{
	"logs_handler": log_queue_handler.NewLogsQueueHandler,
	"send_alerts":  alert_queue_handler.NewAlertsQueueHandler,
}
