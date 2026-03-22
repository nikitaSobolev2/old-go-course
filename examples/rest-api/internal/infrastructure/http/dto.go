package httpapi // JSON модели запроса/ответа

import "github.com/example/go-examples/rest-api/internal/domain"

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
	ID         string            `json:"id"`
	CustomerID string            `json:"customer_id"`
	Status     string            `json:"status"`
	Items      []OrderItemDTO    `json:"items"`
	TotalCents int64             `json:"total_cents"`
	CreatedAt  string            `json:"created_at"`
}

type OrderItemDTO struct {
	ProductID  string `json:"product_id"`
	Name       string `json:"name"`
	Quantity   int    `json:"quantity"`
	PriceCents int64  `json:"price_cents"`
}

func ToOrderResponse(o *domain.Order) OrderResponse {
	items := o.Items()
	dto := make([]OrderItemDTO, 0, len(items))
	for _, it := range items {
		dto = append(dto, OrderItemDTO{
			ProductID:  it.ProductID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			PriceCents: it.PriceCents,
		})
	}
	return OrderResponse{
		ID:         o.ID(),
		CustomerID: o.CustomerID(),
		Status:     string(o.Status()),
		Items:      dto,
		TotalCents: o.TotalCents(),
		CreatedAt:  o.CreatedAt().Format("2006-01-02T15:04:05Z07:00"),
	}
}

func ToDomainItems(items []CreateOrderItem) []domain.OrderItem {
	out := make([]domain.OrderItem, 0, len(items))
	for _, it := range items {
		out = append(out, domain.OrderItem{
			ProductID:  it.ProductID,
			Name:       it.Name,
			Quantity:   it.Quantity,
			PriceCents: it.PriceCents,
		})
	}
	return out
}
