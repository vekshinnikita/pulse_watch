package constants

type EnumSet map[any]struct{}

func enumListToSet[T any](list []T) EnumSet {
	set := make(EnumSet)

	for _, v := range list {
		set[v] = struct{}{}
	}

	return set
}

var EnumSetsMap = map[string]EnumSet{
	"PeriodType":  enumListToSet(PeriodTypeList),
	"MetricType":  enumListToSet(MetricTypeList),
	"MetaVarType": enumListToSet(MetaVarTypeList),
	"LogLevel":    enumListToSet(LogLevelList),
	"LogType":     enumListToSet(LogTypeList),
	"FilterType":  enumListToSet(FilterTypeList),
}
