package metric_usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type GetMetricsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewGetMetricsUseCase(trm repository.TransactionManager, repo *repository.Repository) *GetMetricsUseCase {
	return &GetMetricsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *GetMetricsUseCase) getGetMetricsDataStorage(
	ctx context.Context,
	data *entities.GetMetricsData,
) *entities.GetMetricsData {
	dataCopy := utils.CopyPtr(data)

	now := time.Now()
	defaultEnd := &now

	var defaultStartSec int
	var secDivider int
	timeWithDelaySec := int(now.Add(-constants.BaseMetricCommitDelay).Unix())
	switch dataCopy.PeriodType {
	case constants.MinutePeriodType:
		secDivider = 60
		// Будем отображать за последние 10 минут
		defaultStartSec = timeWithDelaySec - (10 * secDivider)

	case constants.TenMinutesPeriodType:
		secDivider = 10 * 60
		// Будем отображать за последние час
		defaultStartSec = timeWithDelaySec - (6 * secDivider)

	case constants.HourPeriodType:
		secDivider = 60 * 60
		// Будем отображать за последние 12 часов
		defaultStartSec = timeWithDelaySec - (24 * secDivider)

	case constants.DayPeriodType:
		secDivider = 24 * 60 * 60
		// Будем отображать за последние 10 дней
		defaultStartSec = timeWithDelaySec - (10 * secDivider)
	}

	defaultStart := utils.ToPtr(time.Unix(int64((defaultStartSec/secDivider)*secDivider), 0))

	if dataCopy.PeriodStart == nil {
		dataCopy.PeriodStart = defaultStart
	}

	if dataCopy.PeriodEnd == nil {
		dataCopy.PeriodEnd = defaultEnd
	}

	return dataCopy
}

func (s *GetMetricsUseCase) getAppMetricsGroupKey(
	metricType constants.MetricType,
	params *models.JSONB,
) (string, error) {
	var keyBuilder strings.Builder
	keyBuilder.WriteString(string(metricType))

	if params != nil {
		paramsJsonBytes, err := json.Marshal(params)
		if err != nil {
			return "", fmt.Errorf("marshal: %w", err)
		}

		keyBuilder.WriteString(string(paramsJsonBytes))
	}

	return keyBuilder.String(), nil
}

func (s *GetMetricsUseCase) groupAppMetric(
	ctx context.Context,
	appMetrics []models.AppMetric,
) (map[string][]models.AppMetric, error) {
	groped := make(map[string][]models.AppMetric, 0)
	for _, appMetric := range appMetrics {
		key, err := s.getAppMetricsGroupKey(appMetric.MetricType, appMetric.Params)
		if err != nil {
			return nil, fmt.Errorf("get app metrics group key: %w", err)
		}

		_, ok := groped[key]
		if !ok {
			groped[key] = make([]models.AppMetric, 0)
		}
		groped[key] = append(groped[key], appMetric)
	}

	return groped, nil
}

func (s *GetMetricsUseCase) makeResultFromAppMetrics(
	ctx context.Context,
	periodType constants.PeriodType,
	appMetrics []models.AppMetric,
) (*entities.ResultMetricsGroups, error) {
	gropedMetrics, err := s.groupAppMetric(ctx, appMetrics)
	if err != nil {
		return nil, fmt.Errorf("group app metric: %w", err)
	}

	groups := make([]entities.ResultMetricsGroup, 0)
	for _, groupMetrics := range gropedMetrics {
		metric := groupMetrics[0]
		group := entities.ResultMetricsGroup{
			MetricType: metric.MetricType,
			Params:     *metric.Params,
			Metrics:    make([]entities.ResultMetrics, 0),
		}

		for _, metric := range groupMetrics {
			group.Metrics = append(group.Metrics, entities.ResultMetrics{
				PeriodStart: metric.PeriodStart,
				Value:       metric.Value,
			})
		}

		groups = append(groups, group)
	}

	return &entities.ResultMetricsGroups{
		PeriodType: periodType,
		Groups:     groups,
	}, nil
}

func (s *GetMetricsUseCase) getStorageMetrics(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) (*entities.ResultMetricsGroups, error) {
	getDataMetrics := s.getGetMetricsDataStorage(ctx, data)
	appMetrics, err := s.repo.Metric.GetMetrics(ctx, appId, getDataMetrics)
	if err != nil {
		return nil, fmt.Errorf("get metrics: %w", err)
	}

	result, err := s.makeResultFromAppMetrics(ctx, data.PeriodType, appMetrics)
	if err != nil {
		return nil, fmt.Errorf("make result from app metrics: %w", err)
	}

	return result, nil
}

func (s *GetMetricsUseCase) getLivePeriodStartDefault() time.Time {
	minutesAgoWithDelay := time.Now().Add(-time.Minute - constants.BaseMetricCommitDelay)
	windowStart := (minutesAgoWithDelay.Unix() / constants.BaseMetricWindowTime) * constants.BaseMetricWindowTime

	return time.Unix(int64(windowStart), 0)
}

func (s *GetMetricsUseCase) getLiveMetricName(
	metricType *constants.MetricType,
	params map[string]any,
) string {
	if metricType == nil {
		return "*"
	}

	switch *metricType {

	case constants.StatusMetricType:
		code, ok := params["code"]
		if !ok {
			return string(*metricType)
		}

		return fmt.Sprintf("%s:%s", *metricType, code)

	default:
		return string(*metricType)

	}

}

func (s *GetMetricsUseCase) makeGetLiveMetricsData(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) *dtos.GetLiveMetricsData {
	var start time.Time
	var end time.Time

	if data.PeriodStart == nil {
		start = s.getLivePeriodStartDefault()
	} else {
		start = *data.PeriodStart
	}

	if data.PeriodEnd == nil {
		end = time.Now()
	} else {
		end = *data.PeriodEnd
	}

	metricName := s.getLiveMetricName(data.MetricType, data.Params)

	return &dtos.GetLiveMetricsData{
		AppId:      appId,
		MetricName: metricName,
		Start:      start,
		End:        end,
	}
}

func (s *GetMetricsUseCase) getTypeAndParamsFromMetricName(
	metricName string,
) (constants.MetricType, map[string]any) {
	parts := strings.Split(metricName, ":")
	if len(parts) == 0 {
		slog.Warn(fmt.Sprintf("unexpected metric name: %s", metricName))
		return "", nil
	}

	params := make(map[string]any)
	metricType := constants.MetricType(parts[0])
	switch metricType {
	case constants.StatusMetricType:
		params["code"] = parts[1]

		return metricType, params

	default:
		return metricType, nil
	}
}

func (s *GetMetricsUseCase) makeResultFromLiveMetrics(
	ctx context.Context,
	liveMetrics []dtos.LiveMetrics,
) *entities.ResultMetricsGroups {
	groups := make([]entities.ResultMetricsGroup, 0)
	for _, liveMetrics := range liveMetrics {
		metricType, params := s.getTypeAndParamsFromMetricName(liveMetrics.MetricName)
		group := entities.ResultMetricsGroup{
			MetricType: metricType,
			Params:     params,
			Metrics:    make([]entities.ResultMetrics, 0),
		}

		for _, metric := range liveMetrics.Metrics {
			group.Metrics = append(group.Metrics, entities.ResultMetrics{
				PeriodStart: metric.PeriodStart,
				Value:       metric.Value,
			})
		}

		groups = append(groups, group)
	}

	return &entities.ResultMetricsGroups{
		PeriodType: constants.LivePeriodType,
		Groups:     groups,
	}
}

func (s *GetMetricsUseCase) getLiveMetrics(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) (*entities.ResultMetricsGroups, error) {
	getMetricsData := s.makeGetLiveMetricsData(ctx, appId, data)
	liveMetrics, err := s.repo.AnalyticsRedis.GetLiveMetrics(ctx, getMetricsData)
	if err != nil {
		return nil, fmt.Errorf("get live metrics: %w", err)
	}

	result := s.makeResultFromLiveMetrics(ctx, liveMetrics)
	return result, nil
}

func (s *GetMetricsUseCase) Get(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) (*entities.ResultMetricsGroups, error) {
	if data.PeriodType == constants.LivePeriodType {
		result, err := s.getLiveMetrics(ctx, appId, data)
		if err != nil {
			return nil, fmt.Errorf("get live metrics:%w", err)
		}

		return result, nil
	}

	result, err := s.getStorageMetrics(ctx, appId, data)
	if err != nil {
		return nil, fmt.Errorf("get storage metrics:%w", err)
	}

	return result, nil
}
