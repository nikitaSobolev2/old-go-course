package application

import (
	"context"
	"fmt"

	"github.com/example/go-examples/postgres-ms/internal/domain"
)

type OrderService struct {
	repo domain.OrderRepository
}

func NewOrderService(repo domain.OrderRepository) *OrderService {
	return &OrderService{repo: repo}
}

type CreateOrderInput struct {
	CustomerID string
	Items      []domain.OrderItem
}

func (s *OrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	o := domain.NewOrder(in.CustomerID)
	for _, it := range in.Items {
		if err := o.AddItem(it); err != nil {
			return nil, fmt.Errorf("add item: %w", err)
		}
	}
	if err := o.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm: %w", err)
	}
	if err := s.repo.Save(ctx, o); err != nil { // INSERT в Postgres
		return nil, fmt.Errorf("save: %w", err)
	}
	return o, nil
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id)
}
