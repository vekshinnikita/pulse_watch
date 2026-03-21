package background_tasks_handler

import (
	"context"

	"github.com/hibiken/asynq"
)

type Task interface {
	Process(ctx context.Context, task *asynq.Task) error
}
