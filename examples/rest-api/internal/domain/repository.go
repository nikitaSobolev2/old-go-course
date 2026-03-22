package domain

import "context"

// OrderRepository abstracts persistence for Order aggregate.
type OrderRepository interface {
	Save(ctx context.Context, order *Order) error
	GetByID(ctx context.Context, id string) (*Order, error)
}
