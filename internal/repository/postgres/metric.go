package postgres_repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type MetricPostgres struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewMetricPostgres(db *sqlx.DB, c *trmsqlx.CtxGetter) *MetricPostgres {
	return &MetricPostgres{
		db:      db,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *MetricPostgres) CreateAppMetrics(
	ctx context.Context,
	data []dtos.CreateAppMetric,
) ([]int, error) {
	queryBuilder := r.builder.
		Insert("app_metric").
		Columns("app_id", "period_start", "period_type", "type", "is_unique", "params", "value").
		Suffix("RETURNING id")

	for _, item := range data {
		paramsBytes, err := json.Marshal(item.Params)
		if err != nil {
			return nil, fmt.Errorf("marshal params: %w", err)
		}

		queryBuilder = queryBuilder.Values(
			item.AppId,
			item.PeriodStart,
			item.PeriodType,
			item.MetricType,
			item.IsUnique,
			paramsBytes,
			item.Value,
		)

	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var ids []int
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return ids, nil
}

func (r *MetricPostgres) GetAggregatedAppMetricsByPeriod(
	ctx context.Context,
	appIds []int,
	periodType constants.PeriodType,
	startTime time.Time,
	endTime time.Time,
) ([]dtos.AggregatedAppMetric, error) {
	queryBuilder := r.builder.
		Select(
			"app_id",
			"type",
			"params",
			"sum(value) as sum",
		).
		From("app_metric").
		Where(sq.And{
			sq.Eq{
				"app_id":      appIds,
				"period_type": periodType,
				"is_unique":   false,
			},
			sq.GtOrEq{"period_start": startTime},
			sq.Lt{"period_start": endTime},
		}).
		GroupBy(
			"app_id",
			"type",
			"params",
		)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var aggregated []dtos.AggregatedAppMetric
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &aggregated, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return aggregated, nil
}

func (r *MetricPostgres) GetAggregatedMetricsForAlert(
	ctx context.Context,
	appId int,
	data []dtos.AggregateMetricsForAlert,
) ([]dtos.AggregatedAppMetricForAlert, error) {
	orConditions := make([]sq.Sqlizer, 0)
	for _, item := range data {
		orConditions = append(orConditions, sq.And{
			sq.Eq{"type": item.MetricType},
			sq.Expr("period_start BETWEEN ? AND ?", item.PeriodStart, item.PeriodEnd),
		})
	}

	queryBuilder := r.builder.
		Select(
			"app_id",
			"type",
			"params",
			"sum(value) as sum",
			"max(period_start) as max_period_start",
		).
		From("app_metric").
		Where(sq.And{
			sq.Eq{
				"app_id":      appId,
				"period_type": constants.MinutePeriodType,
				"is_unique":   false,
			},
			sq.Or(orConditions),
		}).
		GroupBy(
			"app_id",
			"type",
			"params",
		)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var aggregated []dtos.AggregatedAppMetricForAlert
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &aggregated, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return aggregated, nil
}

func (r *MetricPostgres) GetMetrics(
	ctx context.Context,
	appId int,
	data *entities.GetMetricsData,
) ([]models.AppMetric, error) {
	queryBuilder := r.builder.
		Select(
			"id",
			"app_id",
			"period_start",
			"period_type",
			"type",
			"is_unique",
			"params",
			"value",
			"created_at",
		).
		From("app_metric").
		Where(sq.Eq{
			"app_id":      appId,
			"period_type": data.PeriodType,
		})

	if data.MetricType != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{
			"metric_type": data.MetricType,
		})
	}

	if data.PeriodStart != nil {
		queryBuilder = queryBuilder.Where(sq.GtOrEq{
			"period_start": data.PeriodStart,
		})
	}

	if data.PeriodEnd != nil {
		queryBuilder = queryBuilder.Where(sq.LtOrEq{
			"period_start": data.PeriodEnd,
		})
	}

	if data.Params != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{
			"params": data.Params,
		})
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	var metrics []models.AppMetric
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &metrics, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return metrics, nil
}

func (r *MetricPostgres) ClearMetrics(
	ctx context.Context,
	data dtos.ClearMetrics,
) error {
	orConditions := make([]sq.Sqlizer, 0)
	for _, item := range data {
		orConditions = append(orConditions, sq.And{
			sq.LtOrEq{"period_start": item.LessPeriodStart},
			sq.Eq{"period_type": item.PeriodType},
		})
	}

	queryBuilder := r.builder.
		Delete("app_metric").
		Where(sq.Or(orConditions))

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	var ids []int
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	return nil
}
