package constants

type PeriodType string
type MetricType string

const (
	AppIdCtxKey = "AppId"
)

const (
	LivePeriodType PeriodType = "live"

	MinutePeriodType     PeriodType = "minute"
	TenMinutesPeriodType PeriodType = "ten_minutes"
	HourPeriodType       PeriodType = "hour"
	DayPeriodType        PeriodType = "day"
)

var PeriodTypeList = []PeriodType{
	LivePeriodType,
	MinutePeriodType,
	TenMinutesPeriodType,
	HourPeriodType,
	DayPeriodType,
}

const (
	RequestsMetricType    MetricType = "total_requests"
	CriticalMetricType    MetricType = "critical"
	ErrorsMetricType      MetricType = "errors"
	WarningsMetricType    MetricType = "warnings"
	InfoMetricType        MetricType = "info"
	StatusMetricType      MetricType = "status"
	UniqueUsersMetricType MetricType = "unique_users"
)

var MetricTypeList = []MetricType{
	RequestsMetricType,
	CriticalMetricType,
	ErrorsMetricType,
	WarningsMetricType,
	InfoMetricType,
	StatusMetricType,
	UniqueUsersMetricType,
}
