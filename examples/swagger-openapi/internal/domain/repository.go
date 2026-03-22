package domain

import "context"

type ProductRepository interface {
	GetByID(ctx context.Context, id string) (*Product, error) // порт чтения каталога
}
