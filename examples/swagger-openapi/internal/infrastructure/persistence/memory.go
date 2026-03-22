package persistence // in-memory каталог для демо Swagger

import (
	"context" // в сигнатуре GetByID

	"github.com/example/go-examples/swagger-openapi/internal/domain"
)

type MemoryProductRepository struct {
	data map[string]*domain.Product
}

func NewMemoryProductRepository() *MemoryProductRepository {
	return &MemoryProductRepository{
		data: map[string]*domain.Product{ // стартовые данные для проверки API/Swagger
			"p1": {ID: "p1", Name: "Espresso", Price: 2.5}, // пример id из README
		},
	}
}

func (r *MemoryProductRepository) GetByID(_ context.Context, id string) (*domain.Product, error) {
	p, ok := r.data[id] // линейный поиск в map
	if !ok {
		return nil, domain.ErrNotFound
	}
	return p, nil // возвращаем указатель на запись в map (для демо допустимо)
}
