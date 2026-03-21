package elastic_utils

import (
	"time"

	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/vekshinnikita/pulse_watch/pkg/utils"
)

func AddTextFilter[T ~string](q *types.Query, field string, value *T) {

	if value != nil {
		q.Bool.Must = append(q.Bool.Must, types.Query{
			Match: map[string]types.MatchQuery{
				field: {Query: string(*value)},
			},
		})
	}
}

func AddTermFilter[T any](q *types.Query, field string, value *T) {
	if value != nil {
		q.Bool.Filter = append(q.Bool.Filter, types.Query{
			Term: map[string]types.TermQuery{
				field: {Value: value},
			},
		})
	}
}

func AddDateRangeFilter(q *types.Query, field string, start, end *time.Time) {
	if start != nil || end != nil {
		dateRange := types.DateRangeQuery{}
		if start != nil {
			dateRange.Gte = utils.ToPtr(start.Format(time.RFC3339))
		}
		if end != nil {
			dateRange.Lte = utils.ToPtr(end.Format(time.RFC3339))
		}

		q.Bool.Filter = append(q.Bool.Filter, types.Query{
			Range: map[string]types.RangeQuery{field: dateRange},
		})
	}
}
