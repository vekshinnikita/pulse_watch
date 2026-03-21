package constants

type MetaVarType string

type LogLevel string

type LogType string

type FilterType string

const (
	MetaVarTypeNumber   MetaVarType = "number"
	MetaVarTypeString   MetaVarType = "string"
	MetaVarTypeDate     MetaVarType = "date"
	MetaVarTypeDatetime MetaVarType = "datetime"
)

var MetaVarTypeList = []MetaVarType{
	MetaVarTypeNumber,
	MetaVarTypeString,
	MetaVarTypeDate,
	MetaVarTypeDatetime,
}

const (
	CriticalLogLevel LogLevel = "CRITICAL"
	ErrorLogLevel    LogLevel = "ERROR"
	WarningLogLevel  LogLevel = "WARNING"
	InfoLogLevel     LogLevel = "INFO"
	DebugLogLevel    LogLevel = "DEBUG"
)

var LogLevelList = []LogLevel{
	CriticalLogLevel,
	ErrorLogLevel,
	WarningLogLevel,
	InfoLogLevel,
	DebugLogLevel,
}

const (
	LogLogType     LogType = "LOG"
	RequestLogType LogType = "REQUEST_LOG"
)

var LogTypeList = []LogType{
	LogLogType,
	RequestLogType,
}

const (
	LTFilterType        FilterType = "lt"
	GTFilterType        FilterType = "gt"
	GTEFilterType       FilterType = "gte"
	LTEFilterType       FilterType = "lte"
	NEQFilterType       FilterType = "neq"
	ExistsFilterType    FilterType = "exists"
	NotExistsFilterType FilterType = "not_exists"
	EQFilterType        FilterType = "eq"
	MatchFilterType     FilterType = "match"
)

var FilterTypeList = []FilterType{
	EQFilterType,
	MatchFilterType,
	NEQFilterType,
	ExistsFilterType,
	NotExistsFilterType,
}
