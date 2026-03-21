package models

import (
	"time"

	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type AlertRule struct {
	Id        int                `db:"id" json:"id"`
	AppId     *int               `db:"app_id" json:"app_id"`
	UserId    int                `db:"user_id" json:"user_id"`
	Name      string             `db:"name" json:"name"`
	Level     constants.LogLevel `db:"level" json:"level"`
	Threshold int                `db:"threshold" json:"threshold"`
	Message   string             `db:"message" json:"message"`
	Interval  int                `db:"interval" json:"interval"`
	Deleted   bool               `db:"deleted" json:"deleted"`
}

type AlertRuleJSON struct {
	AlertRule
	App  *App `db:"app" json:"app"`
	User User `db:"user" json:"user"`
}

func (s *AlertRuleJSON) Scan(value any) error {
	return utils.JSONScan(value, s)
}

type Alert struct {
	Id         int                   `db:"id" json:"id"`
	AppId      int                   `db:"app_id" json:"app_id"`
	RuleId     int                   `db:"rule_id" json:"rule_id"`
	Message    string                `db:"message" json:"message"`
	Status     constants.AlertStatus `db:"status" json:"status"`
	CreatedAt  time.Time             `db:"created_at" json:"created_at"`
	ResolvedAt *time.Time            `db:"resolved_at" json:"resolved_at"`
}

type AlertFull struct {
	Id         int                   `db:"id" json:"id"`
	App        AppJSON               `db:"app" json:"app"`
	Rule       AlertRuleJSON         `db:"rule" json:"rule"`
	Message    string                `db:"message" json:"message"`
	Status     constants.AlertStatus `db:"status" json:"status"`
	CreatedAt  time.Time             `db:"created_at" json:"created_at"`
	ResolvedAt *time.Time            `db:"resolved_at" json:"resolved_at"`
}
