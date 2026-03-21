package postgres_repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/vekshinnikita/pulse_watch/internal/dbs/postgres"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type AlertPostgres struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewAlertPostgres(db *sqlx.DB, c *trmsqlx.CtxGetter) *AlertPostgres {
	return &AlertPostgres{
		db:      db,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *AlertPostgres) GetAppRules(
	ctx context.Context,
	appId int,
) ([]models.AlertRule, error) {
	queryBuilder := r.builder.
		Select(
			"id",
			"app_id",
			"user_id",
			"name",
			"level",
			"threshold",
			"message",
			"interval",
			"deleted",
		).
		From("alert_rule").
		Where(sq.And{
			sq.Or{
				sq.Eq{"app_id": appId},
				sq.Expr("app_id IS NULL"),
			},
			sq.Eq{
				"deleted": false,
			},
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	rules := make([]models.AlertRule, 0)
	err = executor.SelectContext(ctx, &rules, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return rules, nil
}

func (r *AlertPostgres) GetAppRulesPaginated(
	ctx context.Context,
	p *entities.PaginationData,
	appId int,
) (*entities.PaginationResult[models.AlertRule], error) {

	queryBuilder := r.builder.
		Select(
			"id",
			"app_id",
			"user_id",
			"name",
			"level",
			"threshold",
			"message",
			"interval",
			"deleted",
		).
		From("alert_rule").
		Where(sq.And{
			sq.Or{
				sq.Eq{"app_id": appId},
				sq.Expr("app_id IS NULL"),
			},
			sq.Eq{
				"deleted": false,
			},
		})

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	rules := make([]models.AlertRule, 0)
	total, err := RunWithPagination(executor, ctx, queryBuilder, p, &rules)
	if err != nil {
		return nil, fmt.Errorf("get paginated alert rules: %w", err)
	}

	return FormatPaginatedResult(p, total, rules), nil
}

func (r *AlertPostgres) GetRule(
	ctx context.Context,
	ruleId int,
) (*models.AlertRule, error) {
	queryBuilder := r.builder.
		Select(
			"id",
			"app_id",
			"user_id",
			"name",
			"level",
			"threshold",
			"message",
			"interval",
			"deleted",
		).
		From("alert_rule").
		Where(sq.Eq{
			"id": ruleId,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var rule models.AlertRule
	err = executor.GetContext(ctx, &rule, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Правило не найдено
			return nil, &errs.NotFoundError{Message: errs.AlertRuleNotFoundErrorMessage}
		}

		return nil, fmt.Errorf("run query: %w", err)
	}

	return &rule, nil
}

func (r *AlertPostgres) GetFullAlertsByIds(
	ctx context.Context,
	alertIds []int,
) ([]models.AlertFull, error) {

	queryBuilder := r.builder.
		Select(
			//alert
			"a.id as id",
			"a.message as message",
			"a.status as status",
			"a.created_at as created_at",
			"a.resolved_at as resolved_at",

			//alert app
			jsonObject("aa.id", map[string]string{
				"id":          "aa.id",
				"name":        "aa.name",
				"code":        "aa.code",
				"description": "aa.description",
				"created_at":  "aa.created_at",
				"deleted":     "aa.deleted",
			})+" AS app",

			// rule
			jsonObject("ar.id", map[string]string{
				"id":        "ar.id",
				"name":      "ar.name",
				"level":     "ar.level",
				"threshold": "ar.threshold",
				"message":   "ar.message",
				"interval":  "ar.interval",
				"deleted":   "ar.deleted",

				"app": jsonObject("ra.id", map[string]string{
					"id":          "ra.id",
					"name":        "ra.name",
					"code":        "ra.code",
					"description": "ra.description",
					"created_at":  "ra.created_at",
					"deleted":     "ra.deleted",
				}),

				"user": jsonObject("ru.id", map[string]string{
					"id":         "ru.id",
					"name":       "ru.name",
					"username":   "ru.username",
					"email":      "ru.email",
					"tg_id":      "ru.tg_id",
					"created_at": "ru.created_at",
					"updated_at": "ru.updated_at",
					"deleted":    "ru.deleted",

					"role": jsonObject("ur.id", map[string]string{
						"id":   "ur.id",
						"code": "ur.code",
						"name": "ur.name",
					}),
				}),
			})+" AS rule",
		).
		From("alert a").
		Join("alert_rule ar on ar.id = a.rule_id").
		Join("app_user ru on ru.id = ar.user_id").
		Join("role ur on ur.id = ru.role_id").
		LeftJoin("app ra on ra.id = ar.app_id").
		Join("app aa on aa.id = a.app_id").
		Where(sq.Eq{
			"a.id": alertIds,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}
	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	alerts := make([]models.AlertFull, 0)
	err = executor.SelectContext(ctx, &alerts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return alerts, nil
}

func (r *AlertPostgres) GetRecentAlertsByRules(
	ctx context.Context,
	data *dtos.GetRecentAlerts,
) ([]models.Alert, error) {
	intervalStr := fmt.Sprintf("%d minutes", int(data.Period.Minutes()))
	queryBuilder := r.builder.
		Select(
			"id",
			"app_id",
			"rule_id",
			"message",
			"status",
			"created_at",
			"resolved_at",
		).
		Options("DISTINCT ON (rule_id)").
		From("alert").
		Where(sq.Eq{
			"rule_id": data.RuleIds,
			"app_id":  data.AppId,
		}).
		// Фильтруем по последнему период
		Where("created_at >= NOW() - ?::interval", intervalStr).
		OrderBy("rule_id", "created_at DESC")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	alerts := make([]models.Alert, 0)
	err = executor.SelectContext(ctx, &alerts, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return alerts, nil
}

func (r *AlertPostgres) CreateAlerts(
	ctx context.Context,
	data []dtos.CreateAlert,
) ([]int, error) {
	queryBuilder := r.builder.
		Insert("alert").
		Columns("app_id", "rule_id", "message").
		Suffix("RETURNING id")

	for _, item := range data {
		queryBuilder = queryBuilder.Values(item.AppId, item.RuleId, item.Message)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	ids := make([]int, 0)
	err = executor.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return ids, nil
}

func (r *AlertPostgres) CreateAlertRule(
	ctx context.Context,
	data *entities.CreateAlertRule,
) (int, error) {
	queryBuilder := r.builder.
		Insert("alert_rule").
		Columns(
			"app_id",
			"user_id",
			"name",
			"level",
			"threshold",
			"interval",
			"message",
		).
		Values(
			data.AppId,
			data.UserId,
			data.Name,
			data.Level,
			data.Threshold,
			data.Interval,
			data.Message,
		).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var id int
	err = executor.GetContext(ctx, &id, query, args...)
	if err != nil {
		// Проверка на ошибку constraint
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == postgres.ForeignKeyViolationCode {
			switch pqErr.Constraint {

			case "alert_rule_user_id_fkey":
				return 0, &errs.NotFoundFieldError{Field: "user_id", Message: errs.UserNotFoundErrorMessage}

			case "alert_rule_app_id_fkey":
				return 0, &errs.NotFoundFieldError{Field: "app_id", Message: errs.AppNotFoundErrorMessage}

			}
		}

		if ok && pqErr.Code == postgres.UniqueViolationCode {
			switch pqErr.Constraint {

			case "alert_rule_app_id_level_threshold_interval_user_id_key":
				return 0, &errs.DuplicateAlertRuleError{Message: errs.AlertRuleDuplicateErrorMessage}

			}
		}

		return 0, fmt.Errorf("run query: %w", err)
	}

	return id, nil
}

func (r *AlertPostgres) SetResolvedAlerts(
	ctx context.Context,
	alertIds []int,
) error {
	queryBuilder := r.builder.
		Update("alert").
		Set("resolved_at", time.Now()).
		Where(sq.Eq{
			"id": alertIds,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err = executor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	return nil
}

func (r *AlertPostgres) UpdateAlertRule(
	ctx context.Context,
	ruleId int,
	data *entities.UpdateAlertRule,
) error {
	queryBuilder := r.builder.
		Update("alert_rule").
		Where(sq.Eq{
			"id": ruleId,
		}).
		Suffix("RETURNING id")

	if data.UserId != nil {
		queryBuilder = queryBuilder.Set("user_id", data.UserId)
	}
	if data.Name != nil {
		queryBuilder = queryBuilder.Set("name", data.Name)
	}
	if data.Level != nil {
		queryBuilder = queryBuilder.Set("level", data.Level)
	}
	if data.Threshold != nil {
		queryBuilder = queryBuilder.Set("threshold", data.Threshold)
	}
	if data.Interval != nil {
		queryBuilder = queryBuilder.Set("interval", data.Interval)
	}
	if data.Message != nil {
		queryBuilder = queryBuilder.Set("message", data.Message)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		// Проверка на ошибку constraint
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == postgres.ForeignKeyViolationCode {
			switch pqErr.Constraint {

			case "alert_rule_user_id_fkey":
				return &errs.NotFoundFieldError{Field: "user_id", Message: errs.UserNotFoundErrorMessage}
			}
		}

		return fmt.Errorf("run query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update alert rule rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &errs.NotFoundError{Message: errs.AlertRuleNotFoundErrorMessage}
	}

	return nil
}

func (r *AlertPostgres) DeleteAlertRule(
	ctx context.Context,
	ruleId int,
) error {
	queryBuilder := r.builder.
		Update("alert_rule").
		Set("deleted", true).
		Where(sq.Eq{
			"id": ruleId,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("run query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete alert rule rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &errs.NotFoundError{Message: errs.AlertRuleNotFoundErrorMessage}
	}

	return nil
}
