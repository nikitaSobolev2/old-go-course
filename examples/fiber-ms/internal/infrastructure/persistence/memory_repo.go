package persistence // in-memory реализация OrderRepository для демо и тестов

import (
	"context" // интерфейс требует ctx; здесь не используется для I/O
	"sync"    // RWMutex для конкурентного доступа к map

	"github.com/example/go-examples/fiber-ms/internal/domain" // Order, ErrNotFound
)

// MemoryOrderRepository is an in-memory implementation for demos/tests.
type MemoryOrderRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Order
}

func NewMemoryOrderRepository() *MemoryOrderRepository {
	return &MemoryOrderRepository{data: make(map[string]*domain.Order)} // пустая map заказов по id
}

func (r *MemoryOrderRepository) Save(_ context.Context, order *domain.Order) error {
	r.mu.Lock() // эксклюзивная блокировка на запись
	defer r.mu.Unlock()
	r.data[order.ID()] = order // сохраняем указатель на агрегат (в демо без копирования)
	return nil
}

func (r *MemoryOrderRepository) GetByID(_ context.Context, id string) (*domain.Order, error) {
	r.mu.RLock() // чтение может идти параллельно
	defer r.mu.RUnlock()
	o, ok := r.data[id] // поиск по первичному ключу
	if !ok {
		return nil, domain.ErrNotFound // единый сентинел для HTTP 404
	}
	return o, nil
}
