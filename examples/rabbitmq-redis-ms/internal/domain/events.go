package domain // доменное событие после успешного сохранения заказа

// OrderCreated is raised after an order is persisted (domain event).
type OrderCreated struct {
	OrderID    string
	CustomerID string
}
