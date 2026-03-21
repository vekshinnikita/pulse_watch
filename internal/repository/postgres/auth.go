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
	entities "github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
)

type AuthPostgres struct {
	db      *sqlx.DB
	getter  *trmsqlx.CtxGetter
	builder sq.StatementBuilderType
}

func NewAuthPostgres(db *sqlx.DB, c *trmsqlx.CtxGetter) *AuthPostgres {
	return &AuthPostgres{
		db:      db,
		getter:  c,
		builder: sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *AuthPostgres) CreateUser(
	ctx context.Context,
	user *entities.SignUpUser,
) (int, error) {

	queryBuilder := r.builder.
		Insert("app_user").
		Columns("name", "username", "password_hash", "email", "tg_id").
		Values(user.Name, user.Username, user.Password, user.Email, user.TgId).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("create user build query: %w", err)
	}

	var id int
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.GetContext(ctx, &id, query, args...)
	if err != nil {
		// Проверка на ошибку уникальности
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == postgres.UniqueViolationCode {
			switch pqErr.Constraint {

			case "app_user_username_key":
				return 0, &errs.UniqueFieldError{Field: "username"}

			case "app_user_email_key":
				return 0, &errs.UniqueFieldError{Field: "email"}

			case "app_user_tg_id_key":
				return 0, &errs.UniqueFieldError{Field: "tg_id"}
			}
		}

		return 0, fmt.Errorf("create user run query: %w", err)
	}

	return id, nil
}

func (r *AuthPostgres) GetUserById(
	ctx context.Context,
	userId int,
) (*models.User, error) {
	queryBuilder := r.builder.
		Select(
			"u.id as id",
			"u.name as name",
			"u.username as username",
			"u.email as email",
			"u.tg_id as tg_id",
			"u.created_at as created_at",
			"u.updated_at as updated_at",
			"u.deleted as deleted",
			"u.deleted as deleted",
			`r.id as "role.id"`,
			`r.code as "role.code"`,
			`r.name as "role.name"`,
		).
		From("app_user u").
		LeftJoin("role r on r.id = u.role_id").
		Where("u.id = ?", userId)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by id=%d build query: %w", userId, err)
	}

	var user models.User
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.GetContext(ctx, &user, query, args...)
	if err != nil {

		if errors.Is(err, sql.ErrNoRows) {
			// Пользователь не найден
			return nil, &errs.UserNotFoundError{Message: errs.UserNotFoundErrorMessage}
		}

		return nil, fmt.Errorf("get user by id=%d run query: %w", userId, err)
	}

	return &user, nil
}

func (r *AuthPostgres) GetRolePermissionsByCode(
	ctx context.Context,
	roleCode string,
) ([]models.Permission, error) {
	queryBuilder := r.builder.
		Select(
			`p.id as "id"`,
			`p.code as "code"`,
			`p.name as "name"`,
		).
		From("role_permission rp").
		Join("role r on r.id = rp.role_id").
		Join("permission p on p.id = rp.permission_id and p.deleted=false").
		Where(sq.Eq{
			"r.code": roleCode,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("get role permission build query: %w", err)
	}

	var permissions []models.Permission
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.SelectContext(ctx, &permissions, query, args...)
	if err != nil {
		return nil, fmt.Errorf("get role permission run query: %w", err)
	}

	return permissions, nil
}

func (r *AuthPostgres) GetUserByUsernameAndPassword(
	ctx context.Context,
	username string,
	passwordHash string,
) (*models.User, error) {
	queryBuilder := r.builder.
		Select(
			"u.id as id",
			"u.name as name",
			"u.username as username",
			"u.email as email",
			"u.tg_id as tg_id",
			"u.created_at as created_at",
			"u.updated_at as updated_at",
			"u.deleted as deleted",
			"u.deleted as deleted",
			`r.id as "role.id"`,
			`r.code as "role.code"`,
			`r.name as "role.name"`,
		).
		From("app_user u").
		LeftJoin("role r on r.id = u.role_id").
		Where("u.username = ? and u.password_hash = ?", username, passwordHash)

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("get user by username and password build query: %w", err)
	}

	var user models.User
	executor := r.getter.DefaultTrOrDB(ctx, r.db)
	err = executor.GetContext(ctx, &user, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Пользователь не найден
			return nil, &errs.UserNotFoundError{Message: errs.UserNotFoundErrorMessage}
		}

		return nil, fmt.Errorf("get user by username and password run query: %w", err)
	}

	return &user, nil
}

func (r *AuthPostgres) IsRefreshTokenValid(ctx context.Context, userId int, jti string) (bool, error) {
	existsQuery := r.builder.
		Select("1").
		From("refresh_token").
		Where(sq.Eq{
			"user_id": userId,
			"jti":     jti,
			"revoked": false,
		}).
		Where("expires_at > NOW()")

	subQuerySQL, args, err := existsQuery.ToSql()
	if err != nil {
		return false, fmt.Errorf("is refresh token valid build query: %w", err)
	}

	// Формируем основной запрос как строку
	query := fmt.Sprintf("SELECT EXISTS (%s)", subQuerySQL)

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var exists bool
	err = executor.GetContext(ctx, &exists, query, args...)
	if err != nil {
		return false, fmt.Errorf("is refresh token valid run query: %w", err)
	}

	return exists, err
}

func (r *AuthPostgres) SaveRefreshToken(
	ctx context.Context,
	userId int,
	jti string,
	expiresAt *time.Time,
) (int, error) {
	queryBuilder := r.builder.
		Insert("refresh_token").
		Columns("jti", "user_id", "expires_at").
		Values(jti, userId, expiresAt).
		Suffix("RETURNING id")

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, fmt.Errorf("save refresh token build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	var id int
	err = executor.GetContext(ctx, &id, query, args...)
	if err != nil {
		return 0, fmt.Errorf("save refresh token run query: %w", err)
	}

	return id, nil
}

func (r *AuthPostgres) RevokeRefreshToken(
	ctx context.Context,
	jti string,
) error {
	queryBuilder := r.builder.
		Update("refresh_token").
		Set("revoked", true).
		Where(sq.Eq{
			"jti": jti,
		})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("revoke refresh token build query: %w", err)
	}

	executor := r.getter.DefaultTrOrDB(ctx, r.db)

	_, err = executor.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("revoke refresh token run query: %w", err)
	}

	return nil
}
