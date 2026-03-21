package log_queue_processors

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/service"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type uniqueMetricSet struct {
	Period time.Duration
	Values map[any]struct{}
}

type uniqueMetricsMapSet = map[string]uniqueMetricSet

type MetricsProcessor struct {
	services *service.Service
}

func NewMetricsProcessor(services *service.Service) *MetricsProcessor {
	return &MetricsProcessor{
		services: services,
	}
}

func getCustomWindowTime(t time.Time, d time.Duration) time.Time {
	seconds := int64(d.Seconds())
	return time.Unix((t.Unix()/seconds)*seconds, 0)
}

func getWindowTime(t time.Time) time.Time {
	return time.Unix((t.Unix()/constants.BaseMetricWindowTime)*constants.BaseMetricWindowTime, 0)
}

func (p *MetricsProcessor) calcMetricLevels(
	ctx context.Context,
	metrics dtos.MetricsMap,
	log entities.EnrichedAppLog,
) {
	key := ""
	switch log.Level {
	case constants.CriticalLogLevel:
		key = string(constants.CriticalMetricType)
	case constants.ErrorLogLevel:
		key = string(constants.ErrorsMetricType)
	case constants.WarningLogLevel:
		key = string(constants.WarningsMetricType)
	case constants.InfoLogLevel:
		key = string(constants.InfoMetricType)
	default:
		return
	}

	appId := log.AppId

	metricName := constants.FormatLiveMetricName(appId, getWindowTime(log.Timestamp))
	p.addMetric(ctx, metrics, metricName, key, 1)
}

func (p *MetricsProcessor) calcMetricStatusCodes(
	ctx context.Context,
	metrics dtos.MetricsMap,
	log entities.EnrichedAppLog,
) {
	if log.Type != constants.RequestLogType {
		return
	}

	status, ok := log.Meta["status"]
	if !ok {
		return
	}

	key := fmt.Sprintf("%s:%v", constants.StatusMetricType, status)
	metricName := constants.FormatLiveMetricName(log.AppId, getWindowTime(log.Timestamp))
	p.addMetric(ctx, metrics, metricName, key, 1)
}

func (p *MetricsProcessor) calcMetricRequests(
	ctx context.Context,
	metrics dtos.MetricsMap,
	log entities.EnrichedAppLog,
) {
	if log.Type != constants.RequestLogType {
		return
	}

	key := string(constants.RequestsMetricType)
	metricName := constants.FormatLiveMetricName(log.AppId, getWindowTime(log.Timestamp))

	p.addMetric(ctx, metrics, metricName, key, 1)
}

func (p *MetricsProcessor) addMetric(
	ctx context.Context,
	metrics dtos.MetricsMap,
	metricName string,
	key string,
	value int,
) {
	_, ok := metrics[metricName]
	if !ok {
		metrics[metricName] = make(map[string]int)
	}

	_, ok = metrics[metricName][key]
	if !ok {
		metrics[metricName][key] = 0
	}

	metrics[metricName][key] += value
}

func (p *MetricsProcessor) addUniqueMetrics(
	ctx context.Context,
	uniqueMetrics uniqueMetricsMapSet,
	name string,
	appId int,
	t time.Time,
	value any,
) {
	periods := map[string]time.Duration{
		"hour": time.Hour,
		"day":  time.Hour * 24,
	}

	for period, periodDuration := range periods {
		metricName := constants.FormatUniqueMetricName(
			name,
			period,
			appId,
			getCustomWindowTime(t, periodDuration),
		)

		_, ok := uniqueMetrics[metricName]
		if !ok {
			uniqueMetrics[metricName] = uniqueMetricSet{
				Period: periodDuration,
				Values: make(map[any]struct{}),
			}
		}

		uniqueMetrics[metricName].Values[value] = struct{}{}
	}
}

func (p *MetricsProcessor) calcMetricUniqueUsers(
	ctx context.Context,
	uniqueMetrics uniqueMetricsMapSet,
	log entities.EnrichedAppLog,
) {
	userId, ok := log.Meta["user_id"]
	if !ok {
		return
	}

	key := string(constants.UniqueUsersMetricType)

	p.addUniqueMetrics(
		ctx,
		uniqueMetrics,
		key,
		log.AppId,
		log.Timestamp,
		userId,
	)
}

func (p *MetricsProcessor) calcMetrics(
	ctx context.Context,
	metrics dtos.MetricsMap,
	log entities.EnrichedAppLog,
) {
	p.calcMetricLevels(ctx, metrics, log)
	p.calcMetricStatusCodes(ctx, metrics, log)
	p.calcMetricRequests(ctx, metrics, log)
}

func (p *MetricsProcessor) calcUniqueMetrics(
	ctx context.Context,
	uniqueMetrics uniqueMetricsMapSet,
	log entities.EnrichedAppLog,
) {
	p.calcMetricUniqueUsers(ctx, uniqueMetrics, log)
}

func (p *MetricsProcessor) makeUniqueMetrics(
	ctx context.Context,
	uniqueMetricsMapSet uniqueMetricsMapSet,
) dtos.UniqueMetricsMap {
	uniqueMetrics := make(dtos.UniqueMetricsMap)

	for metricName, metric := range uniqueMetricsMapSet {
		uniqueMetrics[metricName] = &dtos.UniqueMetric{
			Period: metric.Period,
			Values: utils.MapKeys(metric.Values),
		}
	}

	return uniqueMetrics
}

func (p *MetricsProcessor) calcAllMetrics(
	ctx context.Context,
	logs []entities.EnrichedAppLog,
) (dtos.MetricsMap, dtos.UniqueMetricsMap) {
	metrics := make(dtos.MetricsMap)
	uniqueMetricsSet := make(uniqueMetricsMapSet)

	for _, log := range logs {
		p.calcMetrics(ctx, metrics, log)

		p.calcUniqueMetrics(ctx, uniqueMetricsSet, log)
	}

	uniqueMetrics := p.makeUniqueMetrics(ctx, uniqueMetricsSet)

	return metrics, uniqueMetrics
}

// Process подсчет метрик.
func (p *MetricsProcessor) Process(ctx context.Context, logs []entities.EnrichedAppLog) error {
	if len(logs) == 0 {
		return nil
	}

	// Подсчет метрик
	metrics, uniqueMetrics := p.calcAllMetrics(ctx, logs)

	// Сохранение метрик
	err := p.services.Metric.IncrementMetrics(ctx, metrics, uniqueMetrics)
	if err != nil {
		return fmt.Errorf("increment metrics: %w", err)
	}

	slog.InfoContext(ctx, "metrics are incremented in a store")
	return err
}
