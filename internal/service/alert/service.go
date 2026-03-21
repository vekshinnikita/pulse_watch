package alert_service

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	alert_usecases "github.com/vekshinnikita/pulse_watch/internal/service/alert/usecases"
)

type AlertServiceUseCases struct {
	getAppRules            GetAppRules
	getRecentAlertsByRules GetRecentAlertsByRules
	createAlertsTask       CreateAlertsTask
	getFullAlertsByIds     GetFullAlertsByIds
	createAndGetAlertRule  CreateAndGetAlertRule
	getMetricsByRules      GetMetricsByRules
}

type AlertService struct {
	trm  repository.TransactionManager
	repo *repository.Repository
	uc   *AlertServiceUseCases
}

func NewDefaultAlertService(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *AlertService {
	return &AlertService{
		repo: repo,
		trm:  trm,
		uc: &AlertServiceUseCases{
			getAppRules:            alert_usecases.NewGetAppRulesUseCase(trm, repo),
			getRecentAlertsByRules: alert_usecases.NewGetRecentAlertsByRulesUseCase(trm, repo),
			createAlertsTask:       alert_usecases.NewCreateAlertsTaskUseCase(trm, repo),
			getFullAlertsByIds:     alert_usecases.NewGetFullAlertsByIdsUseCase(trm, repo),
			createAndGetAlertRule:  alert_usecases.NewCreateAndGetRuleUseCase(trm, repo),
			getMetricsByRules:      alert_usecases.NewGetMetricsByRulesUseCase(trm, repo),
		},
	}
}

func (s *AlertService) GetRecentAlertsByRules(
	ctx context.Context,
	data *dtos.GetRecentAlerts,
) ([]models.Alert, error) {
	return s.uc.getRecentAlertsByRules.Get(ctx, data)
}

func (s *AlertService) CreateAlertsTask(
	ctx context.Context,
	data []dtos.CreateAlert,
) error {
	return s.uc.createAlertsTask.Create(ctx, data)
}

func (s *AlertService) GetFullAlertsByIds(
	ctx context.Context,
	alertIds []int,
) ([]models.AlertFull, error) {
	return s.uc.getFullAlertsByIds.Get(ctx, alertIds)
}

func (s *AlertService) SetResolvedAlerts(
	ctx context.Context,
	alertIds []int,
) error {
	return s.repo.Alert.SetResolvedAlerts(ctx, alertIds)
}

func (s *AlertService) GetAppRules(ctx context.Context, appId int) ([]models.AlertRule, error) {
	return s.uc.getAppRules.Get(ctx, appId)
}

func (s *AlertService) GetAppRulesPaginated(
	ctx context.Context,
	p *entities.PaginationData,
	appId int,
) (*entities.PaginationResult[models.AlertRule], error) {
	return s.repo.Alert.GetAppRulesPaginated(ctx, p, appId)
}

func (s *AlertService) GetRule(
	ctx context.Context,
	ruleId int,
) (*models.AlertRule, error) {
	return s.repo.Alert.GetRule(ctx, ruleId)
}

func (s *AlertService) CreateAndGetAlertRule(
	ctx context.Context,
	data *entities.CreateAlertRule,
) (*models.AlertRule, error) {
	return s.uc.createAndGetAlertRule.Create(ctx, data)
}

func (s *AlertService) CreateAlertRule(
	ctx context.Context,
	data *entities.CreateAlertRule,
) (int, error) {
	return s.repo.Alert.CreateAlertRule(ctx, data)
}

func (s *AlertService) UpdateAlertRule(
	ctx context.Context,
	ruleId int,
	data *entities.UpdateAlertRule,
) error {
	return s.repo.Alert.UpdateAlertRule(ctx, ruleId, data)
}

func (s *AlertService) DeleteAlertRule(
	ctx context.Context,
	ruleId int,
) error {
	return s.repo.Alert.DeleteAlertRule(ctx, ruleId)
}

func (s *AlertService) GetMetricsByRules(
	ctx context.Context,
	rules []models.AlertRule,
) (dtos.MetricsByRuleIds, error) {
	return s.uc.getMetricsByRules.Get(ctx, rules)
}
