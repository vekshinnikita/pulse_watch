package asynq_server

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	"github.com/hibiken/asynq"
)

type AsynqServer struct {
	scheduler *asynq.Scheduler
	server    *asynq.Server
	client    *asynq.Client
}

func NewServer(clientOpt *asynq.RedisClientOpt, client *asynq.Client) *AsynqServer {
	srv := &AsynqServer{
		client:    client,
		scheduler: asynq.NewScheduler(clientOpt, &asynq.SchedulerOpts{}),
		server: asynq.NewServer(
			clientOpt,
			asynq.Config{
				Concurrency: 5,
				Logger:      NewLogger(),
			},
		),
	}
	return srv
}

func (s *AsynqServer) registerPeriodicTasks() {
	backgroundTasksCfg := GetBackgroundTasks()
	registerTasks := make([]string, 0)

	for _, t := range backgroundTasksCfg {
		_, err := s.scheduler.Register(t.Cron, asynq.NewTask(t.Task, nil))
		if err == nil {
			registerTasks = append(registerTasks, t.Task)
		}
	}

	slog.Info(fmt.Sprintf("Scheduler registered tasks: %s", strings.Join(registerTasks, ", ")))
}

func (s *AsynqServer) Run(mux *asynq.ServeMux) error {
	s.registerPeriodicTasks()

	// Запуск планировщика
	go func() {
		if err := s.scheduler.Run(); err != nil {
			slog.Error(fmt.Sprintf("run asynq scheduler: %s", err.Error()))
			log.Fatal(err)
		}
	}()

	return s.server.Run(mux)
}

func (s *AsynqServer) Stop() {
	s.server.Stop()
}
