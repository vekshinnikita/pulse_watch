package postgres_repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
)

func ApplyPagination(
	builder sq.SelectBuilder,
	pagination *entities.PaginationData,
) sq.SelectBuilder {
	return builder.
		Limit(uint64(pagination.Limit())).
		Offset(uint64(pagination.Offset()))
}

func FormatPaginatedResult[T any](
	pagination *entities.PaginationData,
	total int,
	result []T,
) *entities.PaginationResult[T] {
	return &entities.PaginationResult[T]{
		Page:  pagination.Page,
		Total: total,
		Items: result,
	}
}

func RunWithPagination(
	executor trmsqlx.Tr,
	ctx context.Context,
	builder sq.SelectBuilder,
	pagination *entities.PaginationData,
	dest any,
) (int, error) {
	paginatedBuilder := ApplyPagination(builder, pagination)

	query, args, err := paginatedBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	sqlStr, argsCount, err := builder.ToSql()
	if err != nil {
		return 0, err
	}
	queryCount := "SELECT COUNT(*) FROM (" + sqlStr + ") AS subquery"

	err = executor.SelectContext(ctx, dest, query, args...)
	if err != nil {
		return 0, err
	}

	var count int
	err = executor.GetContext(ctx, &count, queryCount, argsCount...)
	if err != nil {
		return 0, err
	}

	return count, nil
}
