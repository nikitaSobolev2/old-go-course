package persistence // in-memory OrderRepository

import (
	"context" // контекст в сигнатуре порта; I/O нет
	"sync"    // RWMutex для map

	"github.com/example/go-examples/rest-api/internal/domain"
)

// MemoryOrderRepository is an in-memory implementation for demos/tests.
type MemoryOrderRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{data: make(map[string]*domain.Order)} // пустая хэш-таблица заказов
}

func (r *MemoryOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock() // эксклюзивная блокировка на запись в map
	defer r.mu.Unlock()
	r.data[order.ID()] = order // ключ — id агрегата
	return nil
}

func (r *MemoryOrderRepository) GetByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock() // параллельное чтение
	defer r.mu.RUnlock()
	o, ok := r.data[id] // поиск по id
	if !ok {
		return nil, domain.ErrNotFound
	}
	return o, nil
}
