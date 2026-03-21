package background_tasks_handler

import (
	"github.com/hibiken/asynq"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	background_tasks "github.com/vekshinnikita/pulse_watch/internal/handlers/background_tasks/tasks"
	"github.com/vekshinnikita/pulse_watch/internal/service"
)

type tasks struct {
	processLiveMetrics         Task
	aggregateOneMinuteMetrics  Task
	aggregateTenMinutesMetrics Task
	aggregateHourMetrics       Task
	aggregateDayMetrics        Task

	clearMetrics Task
}

type Handler struct {
	services *service.Service
	tasks    *tasks
}

func NewHandler(
	asynqClient *asynq.Client,
	services *service.Service,
) *Handler {
	return &Handler{
		services: services,
		tasks: &tasks{
			processLiveMetrics:         background_tasks.NewProcessLiveMetricsTask(asynqClient, services),
			aggregateOneMinuteMetrics:  background_tasks.NewAggregateOneMinuteMetricsTask(asynqClient, services),
			aggregateTenMinutesMetrics: background_tasks.NewAggregateTenMinutesMetricsTask(asynqClient, services),
			aggregateHourMetrics:       background_tasks.NewAggregateHourMetricsTask(asynqClient, services),
			aggregateDayMetrics:        background_tasks.NewAggregateDayMetricsTask(asynqClient, services),
			clearMetrics:               background_tasks.NewClearMetricsTask(asynqClient, services),
		},
	}
}

func (h *Handler) InitTasks() *asynq.ServeMux {
	mux := asynq.NewServeMux()

	mux.HandleFunc(constants.ProcessLiveMetricsType, h.tasks.processLiveMetrics.Process)
	mux.HandleFunc(constants.AggregateOneMinuteMetricsType, h.tasks.aggregateOneMinuteMetrics.Process)
	mux.HandleFunc(constants.AggregateTenMinutesMetricsType, h.tasks.aggregateTenMinutesMetrics.Process)
	mux.HandleFunc(constants.AggregateHourMetricsType, h.tasks.aggregateHourMetrics.Process)
	mux.HandleFunc(constants.AggregateDayMetricsType, h.tasks.aggregateDayMetrics.Process)

	mux.HandleFunc(constants.ClearMetricsMetricsType, h.tasks.clearMetrics.Process)

	return mux
}
