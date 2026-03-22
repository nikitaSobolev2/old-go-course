package persistence // источник истины для заказов в RAM (без БД в этом примере)

import (
	"context"
	"sync"

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/domain"
)

type MemoryOrderRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{data: make(map[string]*domain.Order)}
}

func (r *MemoryOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[order.ID()] = order // после Save API кладёт снимок в Redis
	return nil
}

func (r *MemoryOrderRepository) GetByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	o, ok := r.data[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return o, nil
}
