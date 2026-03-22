package application // сценарии заказа: память, кэш Redis, событие OrderCreated в RabbitMQ

import (
	"context"       // все операции с таймаутами/отменой запроса
	"encoding/json" // сериализация снимка заказа для Redis
	"fmt"           // обёртка ошибок
	"time"          // парсинг CreatedAt при восстановлении из кэша

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/domain" // Order, OrderCreated
)

// EventPublisher publishes domain events (implemented in infrastructure).
type EventPublisher interface {
	PublishOrderCreated(ctx context.Context, e domain.OrderCreated) error
}

// OrderService coordinates order use cases.
type OrderService struct {
	repo   domain.OrderRepository
	events EventPublisher
	cache  OrderCache
}

// OrderCache abstracts Redis-backed cache for order reads.
type OrderCache interface {
	SetOrderJSON(ctx context.Context, id string, json []byte) error
	GetOrderJSON(ctx context.Context, id string) ([]byte, error)
	DeleteOrder(ctx context.Context, id string) error
}

func NewOrderService(repo domain.OrderRepository, events EventPublisher, cache OrderCache) *OrderService {
	return &OrderService{repo: repo, events: events, cache: cache} // композиция трёх портов
}

type CreateOrderInput struct {
	CustomerID string
	Items      []domain.OrderItem
}

func (s *OrderService) CreateOrder(ctx context.Context, in CreateOrderInput) (*domain.Order, error) {
	o := domain.NewOrder(in.CustomerID) // новый агрегат
	for _, it := range in.Items {
		if err := o.AddItem(it); err != nil {
			return nil, fmt.Errorf("add item: %w", err)
		}
	}
	if err := o.Confirm(); err != nil {
		return nil, fmt.Errorf("confirm: %w", err)
	}
	if err := s.repo.Save(ctx, o); err != nil { // сначала источник истины
		return nil, fmt.Errorf("save: %w", err)
	}
	payload, err := json.Marshal(orderToSnapshot(o)) // стабильный JSON для ключа order:{id}
	if err != nil {
		return nil, err
	}
	if err := s.cache.SetOrderJSON(ctx, o.ID(), payload); err != nil { // прогреваем кэш чтения
		return nil, fmt.Errorf("cache set: %w", err)
	}
	if err := s.events.PublishOrderCreated(ctx, domain.OrderCreated{OrderID: o.ID(), CustomerID: o.CustomerID()}); err != nil { // асинхронные подписчики
		return nil, fmt.Errorf("publish: %w", err)
	}
	return o, nil
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00" // формат времени в snapshot

type orderSnapshot struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customer_id"`
	Status     string     `json:"status"`
	Items      []itemSnap `json:"items"`
	CreatedAt  string     `json:"created_at"`
}

type itemSnap struct {
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

func orderToSnapshot(o *domain.Order) orderSnapshot {
	items := o.Items()
	s := make([]itemSnap, 0, len(items))
	for _, it := range items {
		s = append(s, itemSnap{
			ProductID: it.ProductID, Name: it.Name, Quantity: it.Quantity, PriceCents: it.PriceCents,
		})
	}
	return orderSnapshot{
		ID: o.ID(), CustomerID: o.CustomerID(), Status: string(o.Status()),
		Items: s, CreatedAt: o.CreatedAt().Format(timeRFC3339),
	}
}

func snapshotToOrder(s orderSnapshot) *domain.Order {
	items := make([]domain.OrderItem, 0, len(s.Items))
	for _, it := range s.Items {
		items = append(items, domain.OrderItem{
			ProductID: it.ProductID, Name: it.Name, Quantity: it.Quantity, PriceCents: it.PriceCents,
		})
	}
	createdAt, _ := time.Parse(time.RFC3339, s.CreatedAt) // при битой дате — zero value (демо)
	return domain.RehydrateOrder(s.ID, s.CustomerID, items, domain.OrderStatus(s.Status), createdAt)
}

func (s *OrderService) GetOrder(ctx context.Context, id string) (*domain.Order, error) {
	if b, err := s.cache.GetOrderJSON(ctx, id); err == nil && len(b) > 0 { // cache hit: быстрый путь
		var snap orderSnapshot
		if err := json.Unmarshal(b, &snap); err == nil && snap.ID != "" { // валидный снимок
			return snapshotToOrder(snap), nil // без обращения к repo
		}
	}
	o, err := s.repo.GetByID(ctx, id) // cache miss или битые данные — читаем память
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(orderToSnapshot(o)) // заполняем кэш для следующих GET
	if err != nil {
		return o, nil // отдаём заказ даже если сериализация для кэша не удалась
	}
	_ = s.cache.SetOrderJSON(ctx, o.ID(), raw) // best-effort прогрев
	return o, nil
}
