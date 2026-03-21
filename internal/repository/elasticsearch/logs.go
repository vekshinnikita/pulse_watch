package elasticsearch_repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/elastic/go-elasticsearch/v9"
	"github.com/elastic/go-elasticsearch/v9/esutil"
	"github.com/vekshinnikita/pulse_watch/internal/constants"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	elastic_searchers "github.com/vekshinnikita/pulse_watch/internal/repository/elasticsearch/searchers"
)

type LogsESSearchers struct {
	logs *elastic_searchers.LogSearcher
}

type LogsES struct {
	client *elasticsearch.TypedClient

	searchers *LogsESSearchers
}

func NewLogsES(client *elasticsearch.TypedClient) *LogsES {
	return &LogsES{
		client: client,

		searchers: &LogsESSearchers{
			logs: elastic_searchers.NewLogSearcher(client),
		},
	}
}

func (es *LogsES) BulkSave(ctx context.Context, logs []entities.EnrichedAppLog) error {
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Client:     es.client,
		Index:      constants.LogsESIndex,
		FlushBytes: constants.ESFlushBytes,
		NumWorkers: constants.ESBulkWorkers,
		OnError: func(c context.Context, err error) {
			slog.Error(fmt.Sprintf("bulk save: %s", err.Error()))
		},
	})
	if err != nil {
		return err
	}

	for _, log := range logs {
		data, err := json.Marshal(log)
		if err != nil {
			return fmt.Errorf("marshal data: %w", err)
		}

		err = bi.Add(ctx, esutil.BulkIndexerItem{
			Action: "index",
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			return fmt.Errorf("add to bulk index: %w", err)
		}
	}

	if err := bi.Close(ctx); err != nil {
		return fmt.Errorf("close bulk indexer: %w", err)
	}

	return nil
}

func (es *LogsES) SearchPaginated(
	ctx context.Context,
	data *entities.SearchLogData,
) (*entities.PaginationResult[entities.EnrichedAppLog], error) {
	return es.searchers.logs.SearchPaginated(ctx, data)
}
