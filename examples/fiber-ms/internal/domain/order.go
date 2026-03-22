package domain // агрегат Order (FoodTech): черновик, позиции, подтверждение

import (
	"errors" // ErrEmptyOrder, ErrInvalidItemQty
	"time"   // время создания и rehydrate

	"github.com/google/uuid" // генерация id заказа
)

var (
	ErrEmptyOrder     = errors.New("order must have at least one item") // Confirm без строк
	ErrInvalidItemQty = errors.New("item quantity must be positive") // AddItem с qty <= 0
)

// Order is the aggregate root for the ordering bounded context (FoodTech).
type Order struct {
	id         string
	customerID string
	items      []OrderItem
	status     OrderStatus
	createdAt  time.Time
}

type OrderStatus string

const (
	OrderStatusDraft     OrderStatus = "draft" // можно добавлять позиции
	OrderStatusConfirmed OrderStatus = "confirmed" // финальный статус после Confirm
)

type OrderItem struct {
	ProductID  string
	Name       string
	Quantity   int
	PriceCents int64
}

func NewOrder(customerID string) *Order {
	return &Order{ // новый заказ в статусе draft
		id:         uuid.NewString(), // уникальный идентификатор
		customerID: customerID,
		items:      nil, // позиции добавляются через AddItem
		status:     OrderStatusDraft,
		createdAt:  time.Now().UTC(), // фиксируем время в UTC
	}
}

func (o *Order) ID() string              { return o.id }
func (o *Order) CustomerID() string     { return o.customerID }
func (o *Order) Items() []OrderItem       { return append([]OrderItem(nil), o.items...) } // копия слайса наружу
func (o *Order) Status() OrderStatus     { return o.status }
func (o *Order) CreatedAt() time.Time     { return o.createdAt }

func (o *Order) AddItem(item OrderItem) error {
	if item.Quantity <= 0 { // инвариант количества
		return ErrInvalidItemQty
	}
	o.items = append(o.items, item) // мутация агрегата
	return nil
}

func (o *Order) Confirm() error {
	if len(o.items) == 0 { // нельзя подтвердить пустой заказ
		return ErrEmptyOrder
	}
	o.status = OrderStatusConfirmed // переход состояния
	return nil
}

func (o *Order) TotalCents() int64 {
	var t int64 // сумма в минимальных единицах валюты
	for _, it := range o.items {
		t += it.PriceCents * int64(it.Quantity) // цена × количество по каждой строке
	}
	return t
}

// Rehydrate reconstructs order from persistence (infrastructure only).
func RehydrateOrder(id, customerID string, items []OrderItem, status OrderStatus, createdAt time.Time) *Order {
	return &Order{ // восстановление из строки БД без бизнес-методов NewOrder
		id:         id,
		customerID: customerID,
		items:      append([]OrderItem(nil), items...), // копия слайса
		status:     status,
		createdAt:  createdAt,
	}
}
