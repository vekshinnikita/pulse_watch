package postgres_repository

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/entities/dtos"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type LogPostgres struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewLogPostgres(db *sqlx.DB, c *trmsqlx.CtxGetter) *LogPostgres {
	return &LogPostgres{
		db:      db,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *LogPostgres) GetAppLogMetaVarsByCodes(
	ctx context.Context,
	appId int,
	varCodes []string,
) ([]models.LogMetaVar, error) {
	queryBuilder := r.builder.
		Select(
			"id as id",
			"name as name",
			"code as code",
			"app_id as app_id",
			"type as type",
			"deleted as deleted",
		).
		From("log_meta_var").
		Where(sq.Eq{
			"app_id":  appId,
			"code":    varCodes,
			"deleted": false,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	logMetaVars := make([]models.LogMetaVar, 0)
	err = executor.SelectContext(ctx, &logMetaVars, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return logMetaVars, nil
}

func (r *LogPostgres) CreateMetaVars(ctx context.Context, data []dtos.CreateMetaVar) ([]int, error) {
	queryBuilder := r.builder.
		Insert("log_meta_var").
		Columns("app_id", "name", "code", "type").
		Suffix("RETURNING id")
	for _, item := range data {
		queryBuilder = queryBuilder.Values(
			item.AppId,
			item.Name,
			item.Code,
			item.Type,
		)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var ids []int
	err = executor.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	return ids, nil
}

func (r *LogPostgres) GetMetaVarsPaginated(
	ctx context.Context,
	pagination *entities.PaginationData,
	appId *int,
) (*entities.PaginationResult[entities.LogMetaVarResult], error) {
	queryBuilder := r.builder.
		Select(
			"name",
			"code",
			"type",
		).
		Distinct().
		From("log_meta_var").
		Where(sq.Eq{
			"deleted": false,
		})

	if appId != nil {
		queryBuilder = queryBuilder.Where(sq.Eq{
			"app_id": appId,
		})
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	metaVars := make([]entities.LogMetaVarResult, 0)
	total, err := RunWithPagination(executor, ctx, queryBuilder, pagination, &metaVars)
	if err != nil {
		return nil, fmt.Errorf("get paginated alert rules: %w", err)
	}

	return FormatPaginatedResult(pagination, total, metaVars), nil
}
