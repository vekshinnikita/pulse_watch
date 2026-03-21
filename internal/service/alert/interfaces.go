package alert_service

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type GetAppRules interface {
	Get(ctx context.Context, appId int) ([]models.AlertRule, error)
}

type GetRecentAlertsByRules interface {
	Get(
		ctx context.Context,
		data *dtos.GetRecentAlerts,
	) ([]models.Alert, error)
}

type CreateAlertsTask interface {
	Create(
		ctx context.Context,
		data []dtos.CreateAlert,
	) error
}

type GetFullAlertsByIds interface {
	Get(
		ctx context.Context,
		alertIds []int,
	) ([]models.AlertFull, error)
}

type CreateAndGetAlertRule interface {
	Create(
		ctx context.Context,
		data *entities.CreateAlertRule,
	) (*models.AlertRule, error)
}

type GetMetricsByRules interface {
	Get(
		ctx context.Context,
		rules []models.AlertRule,
	) (dtos.MetricsByRuleIds, error)
}
