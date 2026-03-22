package domain // агрегат заказа; позиции сериализуются в JSON в колонке MySQL

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrEmptyOrder     = errors.New("order must have at least one item")
	ErrInvalidItemQty = errors.New("item quantity must be positive")
)

type Order struct {
	id         string
	customerID string
	items      []OrderItem
	status     OrderStatus
	createdAt  time.Time
}

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft"
	OrderStatusConfirmed OrderStatus = "confirmed"
)

type OrderItem struct {
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

func NewOrder(customerID string) *Order {
	return &Order{
		id:         uuid.NewString(),
		customerID: customerID,
		status:     OrderStatusDraft,
		createdAt:  time.Now().UTC(),
	}
}

func (o *Order) ID() string              { return o.id }
func (o *Order) CustomerID() string      { return o.customerID }
func (o *Order) Items() []OrderItem      { return append([]OrderItem(nil), o.items...) }
func (o *Order) Status() OrderStatus     { return o.status }
func (o *Order) CreatedAt() time.Time    { return o.createdAt }

func (o *Order) AddItem(item OrderItem) error {
	if item.Quantity <= 0 {
		return ErrInvalidItemQty
	}
	o.items = append(o.items, item)
	return nil
}

func (o *Order) Confirm() error {
	if len(o.items) == 0 {
		return ErrEmptyOrder
	}
	o.status = OrderStatusConfirmed
	return nil
}

func (o *Order) TotalCents() int64 {
	var t int64
	for _, it := range o.items {
		t += it.PriceCents * int64(it.Quantity)
	}
	return t
}

func RehydrateOrder(id, customerID string, items []OrderItem, status OrderStatus, createdAt time.Time) *Order {
	return &Order{
		id: id, customerID: customerID,
		items: append([]OrderItem(nil), items...), status: status, createdAt: createdAt,
	}
}
