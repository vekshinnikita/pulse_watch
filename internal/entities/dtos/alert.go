package dtos

import "time"

type CreateAlertRuleMetric struct {
	RuleId      int
	AppId       int
	Count       int
	StartPeriod time.Time
	EndPeriod   time.Time
}

type GetTotalAlertRulesMetric struct {
	RuleId      int
	StartPeriod time.Time
	EndPeriod   time.Time
}

type TotalAlertRulesMetric struct {
	RuleId int `db:"rule_id"`
	Total  int `db:"total"`
}

type GetRecentAlerts struct {
	AppId   int
	RuleIds []int
	Period  time.Duration
}

type CreateAlert struct {
	AppId   int
	RuleId  int
	Message string
}

type AlertRuleChatId struct {
	RuleId int
	ChatId *int
}

type MetricsByRuleIds map[int]int
