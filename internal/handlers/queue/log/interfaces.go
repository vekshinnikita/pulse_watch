package log_queue_handler

import (
	"context"

	"github.com/vekshinnikita/pulse_watch/internal/entities"
)

type Processor interface {
	Process(ctx context.Context, logs []entities.EnrichedAppLog) error
}
