package alert_usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

var MetricTypeByAlertLogLevel = map[constants.LogLevel]constants.MetricType{
	constants.CriticalAlertRuleLevel: constants.CriticalMetricType,
	constants.ErrorAlertRuleLevel:    constants.ErrorsMetricType,
	constants.WarningAlertRuleLevel:  constants.WarningsMetricType,
}

var AlertLogLevelByMetricType = utils.SwapMap(MetricTypeByAlertLogLevel)

type GetMetricsByRulesUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetMetricsByRulesUseCase(
	trm repository.TransactionManager,
	repo *repository.Repository,
) *GetMetricsByRulesUseCase {
	return &GetMetricsByRulesUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (uc *GetMetricsByRulesUseCase) convertLevelToMetricType(
	level constants.LogLevel,
) (constants.MetricType, bool) {
	metricType, ok := MetricTypeByAlertLogLevel[level]
	if !ok {
		return "", false
	}

	return metricType, true
}

func (uc *GetMetricsByRulesUseCase) convertMetricTypeToLevel(
	metricType constants.MetricType,
) (constants.LogLevel, bool) {
	level, ok := AlertLogLevelByMetricType[metricType]
	if !ok {
		return "", false
	}

	return level, true
}

func (uc *GetMetricsByRulesUseCase) makeAggregateMetricsForAlert(
	rules []models.AlertRule,
) ([]dtos.AggregateMetricsForAlert, error) {
	data := make([]dtos.AggregateMetricsForAlert, 0)
	now := time.Now()

	for _, rule := range rules {
		metricType, ok := uc.convertLevelToMetricType(rule.Level)
		if !ok {
			continue
		}

		data = append(data, dtos.AggregateMetricsForAlert{
			MetricType:  metricType,
			PeriodStart: now.Add(-time.Duration(rule.Interval) * time.Minute),
			PeriodEnd:   now,
		})
	}

	return data, nil
}

func (uc *GetMetricsByRulesUseCase) makeRulesByLevelMap(
	rules []models.AlertRule,
) map[constants.LogLevel][]models.AlertRule {
	rulesByLevelMap := make(map[constants.LogLevel][]models.AlertRule)
	for _, rule := range rules {
		_, ok := rulesByLevelMap[rule.Level]
		if !ok {
			rulesByLevelMap[rule.Level] = make([]models.AlertRule, 0)
		}

		rulesByLevelMap[rule.Level] = append(rulesByLevelMap[rule.Level], rule)
	}

	return rulesByLevelMap
}

func (uc *GetMetricsByRulesUseCase) getStoredMetrics(
	ctx context.Context,
	rules []models.AlertRule,
) ([]dtos.AggregatedAppMetricForAlert, error) {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("can't get app_id")
	}

	data, err := uc.makeAggregateMetricsForAlert(rules)
	if err != nil {
		return nil, fmt.Errorf("make aggregate metrics for alert: %w", err)
	}

	aggregatedMetrics, err := uc.repo.Metric.GetAggregatedMetricsForAlert(ctx, appId, data)
	if err != nil {
		return nil, fmt.Errorf("get aggregated metrics for alert: %w", err)
	}

	return aggregatedMetrics, nil
}

func (uc *GetMetricsByRulesUseCase) addStoredMetrics(
	ctx context.Context,
	rules []models.AlertRule,
	rulesByLevelMap map[constants.LogLevel][]models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) (*time.Time, error) {
	var lastPeriodStart *time.Time

	aggregatedMetrics, err := uc.getStoredMetrics(ctx, rules)
	if err != nil {
		return lastPeriodStart, fmt.Errorf("get stored metrics: %w", err)
	}

	for _, aggregatedMetric := range aggregatedMetrics {
		level, ok := uc.convertMetricTypeToLevel(aggregatedMetric.MetricType)
		if !ok {
			continue
		}

		// Добавляем значение ко всем подходящим значениям правил
		levelRules := rulesByLevelMap[level]
		for _, rule := range levelRules {
			_, ok := metricsByRuleIds[rule.Id]
			if !ok {
				metricsByRuleIds[rule.Id] = 0
			}

			metricsByRuleIds[rule.Id] += aggregatedMetric.Sum
		}
	}

	if len(aggregatedMetrics) != 0 {
		lastPeriodStart = &aggregatedMetrics[0].MaxPeriodStart
	}

	return lastPeriodStart, nil
}

func (uc *GetMetricsByRulesUseCase) getLiveMetrics(
	ctx context.Context,
	lastStoredPeriod *time.Time,
) ([]dtos.AggregatedMetric, error) {
	appId, ok := ctx.Value(constants.AppIdCtxKey).(int)
	if !ok {
		return nil, fmt.Errorf("can't get app_id")
	}

	now := time.Now()
	windowStartSec := (now.Add(-time.Minute).Unix() / 60) * 60

	end := now
	start := time.Unix(windowStartSec, 0)
	if lastStoredPeriod != nil {
		start = lastStoredPeriod.Add(time.Minute)
	}

	aggregatedMetrics, err := uc.repo.AnalyticsRedis.GetAggregatedLiveMetrics(ctx, appId, start, end)
	if err != nil {
		return nil, fmt.Errorf("get aggregated live metrics: %w", err)
	}

	return aggregatedMetrics, nil
}

func (uc *GetMetricsByRulesUseCase) addLiveMetrics(
	ctx context.Context,
	lastStoredPeriod *time.Time,
	rules []models.AlertRule,
	rulesByLevelMap map[constants.LogLevel][]models.AlertRule,
	metricsByRuleIds dtos.MetricsByRuleIds,
) error {
	aggregatedMetrics, err := uc.getLiveMetrics(ctx, lastStoredPeriod)
	if err != nil {
		return fmt.Errorf("get live metrics: %w", err)
	}

	for _, aggregatedMetric := range aggregatedMetrics {
		level, ok := uc.convertMetricTypeToLevel(
			constants.MetricType(aggregatedMetric.MetricName),
		)
		if !ok {
			continue
		}

		// Добавляем значение ко всем подходящим значениям правил
		levelRules := rulesByLevelMap[level]
		for _, rule := range levelRules {
			_, ok := metricsByRuleIds[rule.Id]
			if !ok {
				metricsByRuleIds[rule.Id] = 0
			}

			metricsByRuleIds[rule.Id] += aggregatedMetric.Value
		}
	}

	return nil
}

func (uc *GetMetricsByRulesUseCase) Get(
	ctx context.Context,
	rules []models.AlertRule,
) (dtos.MetricsByRuleIds, error) {
	rulesByLevelMap := uc.makeRulesByLevelMap(rules)

	metricsByRuleIds := make(dtos.MetricsByRuleIds)

	lastStoredPeriod, err := uc.addStoredMetrics(ctx, rules, rulesByLevelMap, metricsByRuleIds)
	if err != nil {
		return nil, fmt.Errorf("add stored metrics: %w", err)
	}

	err = uc.addLiveMetrics(ctx, lastStoredPeriod, rules, rulesByLevelMap, metricsByRuleIds)
	if err != nil {
		return nil, fmt.Errorf("add stored metrics: %w", err)
	}

	return metricsByRuleIds, nil
}
