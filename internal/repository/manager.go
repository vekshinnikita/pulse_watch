package repository

import "context"

//go:generate mockgen -destination=mocks/mock_transaction_manager.go -package=mock_repository . TransactionManager

// TransactionManager — интерфейс для работы с транзакциями
type TransactionManager interface {
	// Do запускает функцию fn в транзакции
	// fn получает контекст и возвращает ошибку
	Do(ctx context.Context, fn func(ctx context.Context) error) error
}
