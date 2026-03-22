package httpfiber // JSON DTO и маппинг в/из доменных типов

import "github.com/example/go-examples/fiber-ms/internal/domain"

type CreateOrderRequest struct {
	CustomerID string            `json:"customer_id"`
	Items      []CreateOrderItem `json:"items"`
}

type CreateOrderItem struct {
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

type OrderResponse struct {
	ID         string         `json:"id"`
	CustomerID string         `json:"customer_id"`
	Status     string         `json:"status"`
	Items      []OrderItemDTO `json:"items"`
	TotalCents int64          `json:"total_cents"`
	CreatedAt  string         `json:"created_at"`
}

type OrderItemDTO struct {
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

func ToOrderResponse(o *domain.Order) OrderResponse {
	items := o.Items() // снимок позиций из агрегата
	dto := make([]OrderItemDTO, 0, len(items)) // слайс ответа с известной ёмкостью
	for _, it := range items {
		dto = append(dto, OrderItemDTO{ // копируем поля в DTO
			ProductID:  it.ProductID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			PriceCents: it.PriceCents,
		})
	}
	return OrderResponse{ // итоговый JSON для клиента
		ID:         o.ID(),
		CustomerID: o.CustomerID(),
		Status:     string(o.Status()), // enum как строка
		Items:      dto,
		TotalCents: o.TotalCents(), // доменный расчёт суммы
		CreatedAt:  o.CreatedAt().Format("2006-01-02T15:04:05Z07:00"), // RFC3339-подобная строка
	}
}

func ToDomainItems(items []CreateOrderItem) []domain.OrderItem {
	out := make([]domain.OrderItem, 0, len(items)) // доменный слайм той же длины
	for _, it := range items {
		out = append(out, domain.OrderItem{ // поля совпадают с CreateOrderItem
			ProductID:  it.ProductID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			PriceCents: it.PriceCents,
		})
	}
	return out
}
