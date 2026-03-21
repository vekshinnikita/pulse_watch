package postgres_repository

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	trmsqlx "github.com/avito-tech/go-transaction-manager/drivers/sqlx/v2"
	"github.com/jmoiron/sqlx"
)

type mockBehavior func(mock sqlmock.Sqlmock)

func NewMockAuthPostgres(t *testing.T) (sqlmock.Sqlmock, *AuthPostgres) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create mock db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	return mock, NewAuthPostgres(sqlx.NewDb(db, "postgres"), trmsqlx.DefaultCtxGetter)
}
