package application // сценарий каталога: получить продукт по id

import (
	"context" // проброс в репозиторий

	"github.com/example/go-examples/swagger-openapi/internal/domain" // Product, ProductRepository
)

type CatalogService struct {
	repo domain.ProductRepository
}

func NewCatalogService(repo domain.ProductRepository) *CatalogService {
	return &CatalogService{repo: repo} // DI порта чтения
}

func (s *CatalogService) GetProduct(ctx context.Context, id string) (*domain.Product, error) {
	return s.repo.GetByID(ctx, id) // делегируем хранилищу
}
