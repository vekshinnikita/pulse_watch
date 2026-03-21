package models

import (
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type App struct {
	Id          int        `db:"id" json:"id" binding:"required"`
	Name        string     `db:"name" json:"name" binding:"required"`
	Code        string     `db:"code" json:"code" binding:"required"`
	Description *string    `db:"description" json:"description"`
	CreatedAt   *time.Time `db:"created_at" json:"created_at"`
	Deleted     *bool      `db:"deleted" json:"deleted"`
}

type AppJSON struct {
	App
}

func (s *AppJSON) Scan(value any) error {
	return utils.JSONScan(value, s)
}

type ApiKey struct {
	Id        int        `db:"id" json:"id" binding:"required"`
	Name      string     `db:"name" json:"name" binding:"required"`
	ExpiresAt *time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	Revoked   bool       `db:"revoked" json:"revoked"`
}

type ApiKeyWithKey struct {
	Id        int        `json:"id"`
	Name      string     `json:"name"`
	Key       string     `json:"key"`
	ExpiresAt *time.Time `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	Revoked   bool       `json:"revoked"`
}

type AppMetric struct {
	Id          int                  `json:"id" db:"id"`
	AppId       int                  `json:"app_id" db:"app_id"`
	PeriodStart time.Time            `json:"period_start" db:"period_start"`
	PeriodType  constants.PeriodType `json:"period_type" db:"period_type"`
	MetricType  constants.MetricType `json:"type" db:"type"`
	IsUnique    bool                 `json:"is_unique" db:"is_unique"`
	Params      *JSONB               `json:"params" db:"params"`
	Value       int                  `json:"value" db:"value"`
	CreatedAt   time.Time            `json:"created_at" db:"created_at"`
}
