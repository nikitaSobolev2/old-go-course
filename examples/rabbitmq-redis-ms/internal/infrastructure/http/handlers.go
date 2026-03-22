package httpapi // REST как в rest-api; за кулисами — кэш и RabbitMQ

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/application"
	"github.com/example/go-examples/rabbitmq-redis-ms/internal/domain"
)

type Handlers struct {
	svc *application.OrderService
}

func NewHandlers(svc *application.OrderService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/orders", h.createOrder) // триггерит save + cache + publish
	r.Get("/orders/{id}", h.getOrder) // сначала Redis, иначе память
	return r
}

func (h *Handlers) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
		return
	}
	o, err := h.svc.CreateOrder(r.Context(), application.CreateOrderInput{
		CustomerID: req.CustomerID,
		Items:      ToDomainItems(req.Items),
	})
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(ToOrderResponse(o))
}

func (h *Handlers) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	o, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ToOrderResponse(o))
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json")
	if errors.Is(err, domain.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
