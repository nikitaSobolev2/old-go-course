package httpfiber // адаптер Fiber к OrderService (application остаётся без Fiber)

import (
	"errors" // errors.Is для ErrNotFound

	"github.com/gofiber/fiber/v2" // контекст запроса, статусы, JSON

	"github.com/example/go-examples/fiber-ms/internal/application" // CreateOrderInput, OrderService
	"github.com/example/go-examples/fiber-ms/internal/domain"      // ErrNotFound
)

// Handlers adapts Fiber to application services (domain/application stay framework-free).
type Handlers struct {
	svc *application.OrderService
}

func NewHandlers(svc *application.OrderService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) Register(app *fiber.App) {
	app.Post("/orders", h.createOrder) // создание заказа из JSON
	app.Get("/orders/:id", h.getOrder) // получение по id из пути
}

func (h *Handlers) createOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest // тело запроса
	if err := c.BodyParser(&req); err != nil { // разбор JSON в структуру
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid json"})
	}
	o, err := h.svc.CreateOrder(c.Context(), application.CreateOrderInput{ // контекст для репозитория
		CustomerID: req.CustomerID,
		Items:      ToDomainItems(req.Items), // DTO → домен
	})
	if err != nil {
		return writeErr(c, err) // маппинг доменных ошибок на HTTP
	}
	return c.Status(fiber.StatusCreated).JSON(ToOrderResponse(o)) // 201 + представление заказа
}

func (h *Handlers) getOrder(c *fiber.Ctx) error {
	id := c.Params("id") // параметр маршрута :id
	o, err := h.svc.GetOrder(c.Context(), id)
	if err != nil {
		return writeErr(c, err)
	}
	return c.JSON(ToOrderResponse(o)) // 200 по умолчанию
}

func writeErr(c *fiber.Ctx, err error) error {
	if errors.Is(err, domain.ErrNotFound) { // отдельный статус для отсутствующей сущности
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "not found"})
	}
	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()}) // прочие — 400 с текстом
}
