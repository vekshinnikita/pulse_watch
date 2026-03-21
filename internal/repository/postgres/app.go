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

type AppPostgres struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewAppPostgres(db *sqlx.DB, c *trmsqlx.CtxGetter) *AppPostgres {
	return &AppPostgres{
		db:      db,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *AppPostgres) CreateApp(
	ctx context.Context,
	data *entities.CreateAppData,
) (int, error) {
	queryBuilder := r.builder.
		Insert("app").
		Columns("name", "code", "description").
		Values(data.Name, data.Code, data.Description).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("create app build query: %w", err)
	}

	var id int
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.GetContext(ctx, &id, query, args...)
	if err != nil {
		// Проверка на ошибку уникальности
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == postgres.UniqueViolationCode {
			switch pqErr.Constraint {

			case "app_code_key":
				return 0, &errs.UniqueFieldError{Field: "code"}
			}
		}

		return 0, fmt.Errorf("create app run query: %w", err)
	}

	return id, nil
}

func (r *AppPostgres) GetApp(
	ctx context.Context,
	appId int,
) (*models.App, error) {
	queryBuilder := r.builder.
		Select(
			"id as id",
			"name as name",
			"code as code",
			"description as description",
			"created_at as created_at",
			"deleted as deleted",
		).
		From("app").
		Where(sq.Eq{
			"id": appId,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Сервис не найден
			return nil, &errs.NotFoundError{Message: errs.AppNotFoundErrorMessage}
		}

		return nil, fmt.Errorf("get app build query: %w", err)
	}

	var app models.App
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.GetContext(ctx, &app, query, args...)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			// Приложение не найдено
			return nil, &errs.NotFoundError{Message: errs.AppNotFoundErrorMessage}
		}

		return nil, fmt.Errorf("get app run query: %w", err)
	}

	return &app, nil
}

func (r *AppPostgres) GetAppIds(
	ctx context.Context,
) ([]int, error) {
	queryBuilder := r.builder.
		Select("id").
		From("app").
		Where(sq.Eq{
			"deleted": false,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("get app build query: %w", err)
	}

	var ids []int
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &ids, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get app run query: %w", err)
	}

	return ids, nil
}

func (r *AppPostgres) DeleteApp(ctx context.Context, appId int) error {
	queryBuilder := r.builder.
		Update("app").
		Set("deleted", true).
		Where(sq.Eq{
			"id": appId,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("delete app build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("delete app run query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete app rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &errs.NotFoundError{Message: errs.AppNotFoundErrorMessage}
	}

	return nil
}

func (r *AppPostgres) CreateApiKey(ctx context.Context, data *dtos.CreateApiKeyData) (int, error) {
	queryBuilder := r.builder.
		Insert("api_key").
		Columns("app_id", "name", "expires_at", "created_at", "key_hash").
		Values(data.AppId, data.Name, data.ExpiresAt, data.CreatedAt, data.KeyHash).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("create api key build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	var id int
	err = executor.GetContext(ctx, &id, query, args...)
	if err != nil {

		// Проверка на ошибку constraint
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == postgres.ForeignKeyViolationCode {
			switch pqErr.Constraint {

			case "api_key_app_id_fkey":
				return 0, &errs.NotFoundFieldError{Field: "app_id", Message: errs.AppNotFoundErrorMessage}
			}
		}
		return 0, fmt.Errorf("create api key run query: %w", err)
	}

	return id, nil
}

func (r *AppPostgres) GetAppApiKeysPaginated(
	ctx context.Context,
	p *entities.PaginationData,
	appId int,
) (*entities.PaginationResult[models.ApiKey], error) {
	queryBuilder := r.builder.
		Select("id", "name", "created_at", "expires_at", "revoked").
		From("api_key").
		Where(sq.Eq{
			"app_id": appId,
		}).
		OrderBy("created_at ASC")

	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	keys := make([]models.ApiKey, 0)
	total, err := RunWithPagination(executor, ctx, queryBuilder, p, &keys)
	if err != nil {
		return nil, fmt.Errorf("get app api keys: %w", err)
	}

	return FormatPaginatedResult(p, total, keys), nil
}

func (r *AppPostgres) RevokeApiKey(ctx context.Context, apiKeyId int) error {
	queryBuilder := r.builder.
		Update("api_key").
		Set("revoked", true).
		Where(sq.Eq{
			"id": apiKeyId,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("revoke api key build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	result, err := executor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("revoke api key run query: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("revoke api key rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return &errs.NotFoundError{Message: errs.ApiKeyNotFoundErrorMessage}
	}

	return nil
}

func (r *AppPostgres) CheckApiKey(ctx context.Context, keyHash string) (int, error) {
	existsQuery := r.builder.
		Select("app_id", "expires_at", "revoked").
		From("api_key").
		Where(sq.Eq{
			"key_hash": keyHash,
		})

	query, args, err := existsQuery.ToSql()
	if err != nil {
		return 0, fmt.Errorf("check api key build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var result struct {
		AppId     int        `db:"app_id"`
		ExpiresAt *time.Time `db:"expires_at"`
		Revoked   bool       `db:"revoked"`
	}
	err = executor.GetContext(ctx, &result, query, args...)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			// API ключ не найдено
			return 0, &errs.InvalidApiKeyError{Message: errs.ApiKeyNotFoundErrorMessage}
		}

		return 0, fmt.Errorf("check api key run query: %w", err)
	}

	e := &errs.ExpiredOrRevokedApiKeyError{Message: errs.ExpiredOrRevokedApiKeyErrorMessage}

	// API ключ аннулирован
	if result.Revoked == true {
		return 0, e
	}

	// API ключ истек
	if result.ExpiresAt != nil && time.Now().After(*result.ExpiresAt) {
		return 0, e
	}

	return result.AppId, err
}
