package constants

import (
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

type AlertStatus string

const (
	CriticalAlertRuleLevel LogLevel = "CRITICAL"
	ErrorAlertRuleLevel    LogLevel = "ERROR"
	WarningAlertRuleLevel  LogLevel = "WARNING"
)

var AlertRulesLevelSet = utils.Set(
	CriticalAlertRuleLevel,
	ErrorAlertRuleLevel,
	WarningAlertRuleLevel,
)

const (
	SentAlertStatus AlertStatus = "sent"
	NewAlertStatus  AlertStatus = "new"
)
