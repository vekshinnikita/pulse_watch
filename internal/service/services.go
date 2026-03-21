package service

import (
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	alert_service "github.com/vekshinnikita/pulse_watch/internal/service/alert"
	apps_service "github.com/vekshinnikita/pulse_watch/internal/service/apps"
	auth_service "github.com/vekshinnikita/pulse_watch/internal/service/auth"
	logs_service "github.com/vekshinnikita/pulse_watch/internal/service/logs"
	metric_service "github.com/vekshinnikita/pulse_watch/internal/service/metric"
)

type Service struct {
	Auth   AuthService
	App    AppService
	Logs   LogsService
	Alert  AlertService
	Metric MetricService
}

func NewServices(repo *repository.Repository, trm repository.TransactionManager) *Service {
	return &Service{
		Auth:   auth_service.NewDefaultAuthService(trm, repo),
		App:    apps_service.NewDefaultAppsService(trm, repo),
		Logs:   logs_service.NewDefaultLogsService(trm, repo),
		Alert:  alert_service.NewDefaultAlertService(trm, repo),
		Metric: metric_service.NewDefaultMetricService(trm, repo),
	}
}
