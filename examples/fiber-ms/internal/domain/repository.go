package domain // порт сохранения агрегата Order

import "context" // контекст для отмены/таймаута при Save/Get

// OrderRepository abstracts persistence for Order aggregate.
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
}
