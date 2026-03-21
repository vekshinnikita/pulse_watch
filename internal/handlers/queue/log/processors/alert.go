package log_queue_processors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type currentLogMetrics struct {
	StartPeriod time.Time
	EndPeriod   time.Time
	RulesMetric map[int]int
}

type logMetrics struct {
	Current          *currentLogMetrics
	TotalRuleMetrics map[int]int
}

type AlertProcessor struct {
	services *service.Service
}

func NewAlertProcessor(services *service.Service) *AlertProcessor {
	return &AlertProcessor{
		services: services,
	}
}

// checkMetrics Сверяет метрики с правилами
// Возвращает активированные правила
func (p *AlertProcessor) checkMetrics(
	ctx context.Context,
	rules []models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) []models.AlertRule {
	activatedRules := make([]models.AlertRule, 0)

	for _, rule := range rules {
		metric, ok := metricsByRuleIds[rule.Id]
		if !ok {
			continue
		}

		if rule.Threshold <= metric {
			activatedRules = append(activatedRules, rule)
		}
	}

	return activatedRules
}

// getRulesToSendAlerts решает по каким правилам нужно отправлять алерты,
// потому что не по всем правилам, которые активировались нужно отправлять
// алерты
func (p *AlertProcessor) getRulesToSendAlerts(
	ctx context.Context,
	activatedRules []models.AlertRule,
) ([]models.AlertRule, error) {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("cannot get %s from context", constants.AppIdCtxKey)
	}

	// Получаем недавно отправленные алерты
	data := &dtos.GetRecentAlerts{
		AppId: appId,
		RuleIds: utils.Map(
			activatedRules,
			func(r models.AlertRule) int { return r.Id },
		),
		Period: time.Hour,
	}
	recentAlertsByRules, err := p.services.Alert.GetRecentAlertsByRules(ctx, data)
	if err != nil {
		return nil, fmt.Errorf("get recent alerts by rules: %w", err)
	}

	// Формирует множество недавно по активированным правилам
	recentActivatedRuleIds := utils.SetFromSlice(
		recentAlertsByRules,
		func(a models.Alert) int { return a.RuleId },
	)

	// Формируем слайс правил, которые нужно отправить
	rulesToSend := make([]models.AlertRule, 0)
	for _, rule := range activatedRules {
		_, ok := recentActivatedRuleIds[rule.Id]
		if !ok {
			rulesToSend = append(rulesToSend, rule)
		}
	}

	return rulesToSend, nil
}

func (p *AlertProcessor) makeCreateAlert(
	app *models.App,
	rule *models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) (*dtos.CreateAlert, error) {
	metric, ok := metricsByRuleIds[rule.Id]
	if !ok {
		return nil, fmt.Errorf("unable to get rule metric")
	}

	message := fmt.Sprintf(
		"ALERT RULE %s HAS BEEN ACTIVATED!!!\n"+
			"App: %s\n"+
			"Current metric: %v\n"+
			rule.Message,

		rule.Name,
		app.Name,
		metric,
	)

	return &dtos.CreateAlert{
		AppId:   app.Id,
		RuleId:  rule.Id,
		Message: message,
	}, nil
}

func (p *AlertProcessor) makeCreateAlerts(
	ctx context.Context,
	rules []models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) ([]dtos.CreateAlert, error) {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("cannot get %s from context", constants.AppIdCtxKey)
	}

	app, err := p.services.App.GetApp(ctx, appId)
	if err != nil {
		return nil, fmt.Errorf("get app: %w", err)
	}

	data := make([]dtos.CreateAlert, 0)
	for _, rule := range rules {
		createAlert, err := p.makeCreateAlert(app, &rule, metricsByRuleIds)
		if err != nil {
			return nil, fmt.Errorf("make create alert: %w", err)
		}
		data = append(data, *createAlert)
	}

	return data, nil
}

func (p *AlertProcessor) sendAlerts(
	ctx context.Context,
	rules []models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) error {
	if len(rules) == 0 {
		return nil
	}

	// Формируем данные для создания алертов
	data, err := p.makeCreateAlerts(ctx, rules, metricsByRuleIds)
	if err != nil {
		return fmt.Errorf("make create alerts: %w", err)
	}

	// Создании задачи на отправку алертов
	err = p.services.Alert.CreateAlertsTask(ctx, data)
	if err != nil {
		return fmt.Errorf("create send alerts task: %w", err)
	}

	return nil
}

func (p *AlertProcessor) processActivatedAlerts(
	ctx context.Context,
	activatedRules []models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) error {
	rulesToSendAlerts, err := p.getRulesToSendAlerts(ctx, activatedRules)
	if err != nil {
		return fmt.Errorf("get rules to send alerts: %w", err)
	}

	slog.InfoContext(ctx, fmt.Sprintf("%v alerts has to be send", len(rulesToSendAlerts)))

	err = p.sendAlerts(ctx, rulesToSendAlerts, metricsByRuleIds)
	if err != nil {
		return fmt.Errorf("send alerts: %w", err)
	}

	return nil
}

func (p *AlertProcessor) Process(ctx context.Context, logs []entities.EnrichedAppLog) error {
	if len(logs) == 0 {
		return nil
	}

	appId := logs[0].AppId
	rules, err := p.services.Alert.GetAppRules(ctx, appId)
	if err != nil {
		return fmt.Errorf("get app alert rules: %w", err)
	}

	slog.InfoContext(ctx, fmt.Sprintf("app has %v alert rules", len(rules)))

	// Получение метрик
	metricsByRuleIds, err := p.services.Alert.GetMetricsByRules(ctx, rules)
	if err != nil {
		return fmt.Errorf("get metrics by rules: %w", err)
	}

	// Проверка активации правил
	activatedRules := p.checkMetrics(ctx, rules, metricsByRuleIds)
	slog.InfoContext(ctx, fmt.Sprintf("%v rules are activated", len(activatedRules)))
	if len(activatedRules) > 0 {

		// Обработка активированных правил
		err = p.processActivatedAlerts(ctx, activatedRules, metricsByRuleIds)
		if err != nil {
			return fmt.Errorf("process activated alerts: %w", err)
		}
	}

	return nil
}
