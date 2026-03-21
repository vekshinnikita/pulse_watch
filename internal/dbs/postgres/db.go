package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func NewDB() (*sqlx.DB, error) {
	config := GetConfig()

	db, err := sqlx.Open(
		"postgres",
		fmt.Sprintf("host=%s port=%v user=%s dbname=%s password=%s sslmode=%s",
			config.Host,
			config.Port,
			config.Username,
			config.DBName,
			config.Password,
			config.SSLMode,
		),
	)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	return db, nil
}
