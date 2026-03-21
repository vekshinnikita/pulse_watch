package metric_usecases

import (
	"context"
	"fmt"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/repository"
)

type IncrementMetricsUseCase struct {
	repo *repository.Repository
	trm  repository.TransactionManager
}

func NewIncrementMetricsUseCase(trm repository.TransactionManager, repo *repository.Repository) *IncrementMetricsUseCase {
	return &IncrementMetricsUseCase{
		repo: repo,
		trm:  trm,
	}
}

func (s *IncrementMetricsUseCase) addMetrics(
	ctx context.Context,
	meticsMap dtos.MetricsMap,
) error {
	for name, metrics := range meticsMap {
		for key, value := range metrics {
			err := s.repo.AnalyticsRedis.AddMetric(
				ctx,
				name,
				key,
				value,
			)
			if err != nil {
				return fmt.Errorf("add metric: %w", err)
			}
		}

		err := s.repo.AnalyticsRedis.AddExpire(ctx, name, constants.MetricExpire)
		if err != nil {
			return fmt.Errorf("add metric: %w", err)
		}
	}

	return nil
}

func (s *IncrementMetricsUseCase) addUniqueMetrics(
	ctx context.Context,
	uniqueMetrics dtos.UniqueMetricsMap,
) error {
	for name, metric := range uniqueMetrics {
		err := s.repo.AnalyticsRedis.AddUniqueMetric(
			ctx,
			name,
			metric,
		)
		if err != nil {
			return fmt.Errorf("add unique metric: %w", err)
		}
	}

	return nil
}

func (s *IncrementMetricsUseCase) Increment(
	ctx context.Context,
	metics dtos.MetricsMap,
	uniqueMetrics dtos.UniqueMetricsMap,
) error {

	// Делаем в транзакции чтобы данные вставились батчем
	return s.trm.Do(ctx, func(с context.Context) error {
		err := s.addMetrics(с, metics)
		if err != nil {
			return fmt.Errorf("add metrics: %w", err)
		}

		err = s.addUniqueMetrics(с, uniqueMetrics)
		if err != nil {
			return fmt.Errorf("add unique metrics: %w", err)
		}

		return nil
	})
}
