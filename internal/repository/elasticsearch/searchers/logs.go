package elastic_searchers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types"
	"github.com/elastic/go-elasticsearch/v9/typedapi/types/enums/sortorder"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	elastic_utils "github.com/vekshinnikita/pulse_watch/internal/repository/elasticsearch/utils"
)

type LogSearcher struct {
	client *elasticsearch.TypedClient
}

func NewLogSearcher(client *elasticsearch.TypedClient) *LogSearcher {
	return &LogSearcher{
		client: client,
	}
}

func (ls *LogSearcher) buildMetaFilter(idx int, item entities.SearchLogMeta) (types.Query, error) {
	fieldName := fmt.Sprintf("meta.%s", item.Name)
	var q types.Query

	switch item.FilterType {
	case constants.EQFilterType, constants.NEQFilterType:
		q = types.Query{
			Term: map[string]types.TermQuery{
				fieldName: {Value: item.Value},
			},
		}

	case constants.ExistsFilterType, constants.NotExistsFilterType:
		q = types.Query{
			Exists: &types.ExistsQuery{
				Field: fieldName,
			},
		}

	case constants.MatchFilterType:
		strVal, ok := item.Value.(string)
		if !ok {
			return q, &errs.TypeFieldError{
				Message: fmt.Sprintf(errs.TypeFieldErrorMessage, "string"),
				Field:   fieldName,
			}
		}
		q = types.Query{
			Match: map[string]types.MatchQuery{
				fieldName: {Query: strVal},
			},
		}
	default:
		return q, &errs.TypeFieldError{
			Message: fmt.Sprintf("unsupported meta filter type %s. Supports only 'eq' and 'match'", item.FilterType),
			Field:   fmt.Sprintf("meta.%d.filter_type", idx),
		}
	}

	return q, nil
}

func (ls *LogSearcher) buildQuery(data *entities.SearchLogData) (*types.Query, error) {
	query := &types.Query{
		Bool: &types.BoolQuery{
			Must:   []types.Query{},
			Filter: []types.Query{},
		},
	}

	// Основные фильтры
	elastic_utils.AddTextFilter(query, "message", data.Query)
	elastic_utils.AddTermFilter(query, "app_id", data.AppId)
	elastic_utils.AddTermFilter(query, "type", data.Type)
	elastic_utils.AddTermFilter(query, "level", data.Level)
	elastic_utils.AddDateRangeFilter(query, "timestamp", data.Start, data.End)

	// Фильтры по meta
	for idx, item := range data.Meta {
		metaQuery, err := ls.buildMetaFilter(idx, item)
		if err != nil {
			return nil, fmt.Errorf("build meta filter: %w", err)
		}

		switch item.FilterType {
		case constants.MatchFilterType:
			query.Bool.Must = append(query.Bool.Must, metaQuery)

		case constants.NEQFilterType, constants.NotExistsFilterType:
			query.Bool.MustNot = append(query.Bool.Must, metaQuery)

		default:
			query.Bool.Filter = append(query.Bool.Filter, metaQuery)
		}
		if item.FilterType == constants.MatchFilterType {

		} else if item.FilterType == constants.NEQFilterType {

		} else {

		}
	}

	return query, nil
}

func (ls *LogSearcher) makeResult(
	ctx context.Context,
	p *entities.PaginationData,
	resp *search.Response,
) (*entities.PaginationResult[entities.EnrichedAppLog], error) {
	result := &entities.PaginationResult[entities.EnrichedAppLog]{
		Items: make([]entities.EnrichedAppLog, 0, len(resp.Hits.Hits)),
		Total: int(resp.Hits.Total.Value),
		Page:  p.Page,
	}

	for _, hit := range resp.Hits.Hits {
		var log entities.EnrichedAppLog
		logJson, err := hit.Source_.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("get source json: %w", err)
		}

		err = json.Unmarshal(logJson, &log)
		if err != nil {
			return nil, fmt.Errorf("unmarshal json: %s", err)
		}

		result.Items = append(result.Items, log)
	}

	return result, nil
}

func (ls *LogSearcher) SearchPaginated(
	ctx context.Context,
	data *entities.SearchLogData,
) (*entities.PaginationResult[entities.EnrichedAppLog], error) {
	query, err := ls.buildQuery(data)
	if err != nil {
		return nil, fmt.Errorf("build query: %w", err)
	}

	sort := &types.SortOptions{
		SortOptions: map[string]types.FieldSort{
			"timestamp": {Order: &sortorder.Desc},
		},
	}

	b, _ := query.MarshalJSON()
	fmt.Println(string(b))

	resp, err := ls.client.Search().
		Index(constants.LogsESIndex).
		Query(query).
		Sort(sort).
		Size(data.PaginationData.Limit()).
		From(data.PaginationData.Offset()).
		Do(ctx)

	if err != nil {
		return nil, fmt.Errorf("run query: %w", err)
	}

	result, err := ls.makeResult(ctx, &data.PaginationData, resp)
	if err != nil {
		return nil, fmt.Errorf("make result: %w", err)
	}

	return result, nil
}
