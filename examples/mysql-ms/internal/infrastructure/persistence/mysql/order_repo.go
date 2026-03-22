package mysql // реализация domain.OrderRepository: таблица orders, позиции в JSON

import (
	"context"       // отмена/таймаут SQL
	"database/sql"  // QueryRowContext, ExecContext
	"encoding/json" // сериализация слайса позиций
	"errors"        // sql.ErrNoRows
	"time"          // сканирование created_at

	"github.com/example/go-examples/mysql-ms/internal/domain"
)

type OrderRepository struct {
	db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
	return &OrderRepository{db: db} // пул переиспользуется между запросами
}

func (r *OrderRepository) Save(ctx context.Context, o *domain.Order) error {
	items, err := json.Marshal(o.Items()) // JSON массив OrderItem для колонки items
	if err != nil {
		return err
	}
	// INSERT новой строки заказа (демо без upsert); плейсхолдеры MySQL
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, items, created_at)
		VALUES (?, ?, ?, ?, ?)
	`, o.ID(), o.CustomerID(), string(o.Status()), items, o.CreatedAt())
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
	var (
		customerID string
		status     string
		itemsRaw   []byte // сырой JSON из БД
		created    time.Time
	)
	// SELECT одной строки по id; Scan заполняет локальные переменные
	err := r.db.QueryRowContext(ctx, `
		SELECT customer_id, status, items, created_at FROM orders WHERE id = ?
	`, id).Scan(&customerID, &status, &itemsRaw, &created)
	if errors.Is(err, sql.ErrNoRows) { // нет строки — не найдено
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	var items []domain.OrderItem
	if err := json.Unmarshal(itemsRaw, &items); err != nil { // восстанавливаем слайс позиций
		return nil, err
	}
	return domain.RehydrateOrder(id, customerID, items, domain.OrderStatus(status), created.UTC()), nil // доменный объект из БД
}
