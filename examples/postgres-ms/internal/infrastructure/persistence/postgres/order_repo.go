package postgres // OrderRepository: плейсхолдеры $1..$5 (стиль PostgreSQL)

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/example/go-examples/postgres-ms/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Save(ctx context.Context, o *domain.Order) error {
	items, err := json.Marshal(o.Items()) // JSONB-совместимый массив в Go
	if err != nil {
		return err
	}
	// INSERT с нумерованными параметрами Postgres
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, items, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, o.ID(), o.CustomerID(), string(o.Status()), items, o.CreatedAt())
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var (
		customerID string
		status     string
		itemsRaw   []byte
		created    time.Time
	)
	err := r.db.QueryRowContext(ctx, `
		SELECT customer_id, status, items, created_at FROM orders WHERE id = $1
	`, id).Scan(&customerID, &status, &itemsRaw, &created)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var items []domain.OrderItem
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		return nil, err
	}
	return domain.RehydrateOrder(id, customerID, items, domain.OrderStatus(status), created.UTC()), nil
}
