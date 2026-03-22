package application // сценарии: создание заказа с подтверждением и чтение по id

import (
	"context" // проброс в репозиторий
	"fmt"     // обёртка ошибок с %w

	"github.com/example/go-examples/fiber-ms/internal/domain" // Order, OrderItem, OrderRepository
)

// OrderService orchestrates use cases for orders.
type OrderService struct {
	repo domain.OrderRepository
}

func NewOrderService(repo domain.OrderRepository) *OrderService {
	return &OrderService{repo: repo} // DI порта persistence
}

type CreateOrderInput struct {
	CustomerID string
	Items      []domain.OrderItem
}

func (s *OrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	o := domain.NewOrder(in.CustomerID) // новый агрегат
	for _, it := range in.Items { // добавляем все позиции из запроса
		if err := o.AddItem(it); err != nil { // доменная валидация qty
			return nil, fmt.Errorf("add item: %w", err)
		}
	}
	if err := o.Confirm(); err != nil { // переводим в confirmed только если есть строки
		return nil, fmt.Errorf("confirm: %w", err)
	}
	if err := s.repo.Save(ctx, o); err != nil { // персистим после успешных инвариантов
		return nil, fmt.Errorf("save: %w", err)
	}
	return o, nil // возвращаем созданный заказ вызывающему (HTTP слой)
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	return s.repo.GetByID(ctx, id) // прямое чтение из репозитория
}
