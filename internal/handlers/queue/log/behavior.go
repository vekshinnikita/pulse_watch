package log_queue_handler

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

type BehaviorProcessor[T any] interface {
	Process(ctx context.Context, data T) error
}

type Behavior[T any] struct {
	timeout         time.Duration
	syncProcessors  []BehaviorProcessor[T]
	asyncProcessors []BehaviorProcessor[T]
}

func NewBehavior[T any](timeout time.Duration) *Behavior[T] {
	return &Behavior[T]{
		timeout:         timeout,
		syncProcessors:  make([]BehaviorProcessor[T], 0),
		asyncProcessors: make([]BehaviorProcessor[T], 0),
	}
}

func NewDefaultBehavior[T any]() *Behavior[T] {
	return &Behavior[T]{
		timeout:         5 * time.Second,
		syncProcessors:  make([]BehaviorProcessor[T], 0),
		asyncProcessors: make([]BehaviorProcessor[T], 0),
	}
}

func (b *Behavior[T]) AddSync(processor BehaviorProcessor[T]) {
	b.syncProcessors = append(b.syncProcessors, processor)
}

func (b *Behavior[T]) AddAsync(processor BehaviorProcessor[T]) {
	b.asyncProcessors = append(b.asyncProcessors, processor)
}

func (b *Behavior[T]) executeSync(
	ctx context.Context,
	wg *sync.WaitGroup,
	errCh chan error,
	data T,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		for idx, processor := range b.syncProcessors {
			select {
			case <-ctx.Done():
				return
			default:
			}

			start := time.Now()
			err := processor.Process(ctx, data)
			slog.DebugContext(ctx, fmt.Sprintf("processing by %d sync processor is completed in %d ms", idx, time.Since(start).Milliseconds()))

			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errCh <- fmt.Errorf("run sync %v processor: %w", idx, err):
					return
				}
			}
		}
	}()
}

func (b *Behavior[T]) executeAsync(
	ctx context.Context,
	wg *sync.WaitGroup,
	errCh chan error,
	data T,
) {
	for idx, processor := range b.asyncProcessors {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				return
			default:
			}

			start := time.Now()
			err := processor.Process(ctx, data)

			slog.DebugContext(ctx, fmt.Sprintf("processing by %d async processor is completed in %d ms", i, time.Since(start).Milliseconds()))
			if err != nil {
				select {
				case <-ctx.Done():
					return
				case errCh <- fmt.Errorf("run async %v processor: %w", i, err):
					return
				}
			}
		}(idx)
	}
}

// При вызове Execute запускаются синхронные и асинхронные обработчики параллельно
func (b *Behavior[T]) Execute(ctx context.Context, data T) error {
	wg := &sync.WaitGroup{}
	ctxWithTimeout, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	// навсякий случае канал буферизированный
	// чтобы горутины не блокировались
	errCh := make(chan error, len(b.asyncProcessors)+len(b.syncProcessors))
	defer close(errCh)

	// Запускаем синхронные обработчики
	b.executeSync(ctxWithTimeout, wg, errCh, data)

	// Запускаем асинхронные обработчики
	b.executeAsync(ctxWithTimeout, wg, errCh, data)

	// Ждем когда все закончат работу
	wg.Wait()

	select {
	case <-ctxWithTimeout.Done():
		return fmt.Errorf("timeout error")

	case err := <-errCh:
		// Если пришла ошибка то возвращаем
		return err

	default:
	}

	return nil
}
