package httpapi // chi + encoding/json: REST поверх OrderService

import (
	"encoding/json" // Decode тела, Encode ответов и ошибок
	"errors"        // errors.Is
	"net/http"      // ResponseWriter, Request, статусы

	"github.com/example/go-examples/rest-api/internal/application"
	"github.com/example/go-examples/rest-api/internal/domain"
	"github.com/go-chi/chi/v5" // URL-параметр {id}
)

type Handlers struct {
	svc *application.OrderService
}

func NewHandlers(svc *application.OrderService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/orders", h.createOrder) // JSON в теле
	r.Get("/orders/{id}", h.getOrder) // id в пути
	return r
}

func (h *Handlers) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // разбор JSON без Fiber
		http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest) // простой ответ (не chi helper)
		return
	}
	o, err := h.svc.CreateOrder(r.Context(), application.CreateOrderInput{
		CustomerID: req.CustomerID,
		Items:      ToDomainItems(req.Items),
	})
	if err != nil {
		writeErr(w, err) // единый формат JSON-ошибок
		return
	}
	w.Header().Set("Content-Type", "application/json") // явный Content-Type успеха
	w.WriteHeader(http.StatusCreated) // 201 Created
	_ = json.NewEncoder(w).Encode(ToOrderResponse(o))
}

func (h *Handlers) getOrder(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // извлечение из шаблона маршрута
	o, err := h.svc.GetOrder(r.Context(), id)
	if err != nil {
		writeErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ToOrderResponse(o)) // 200 по умолчанию
}

func writeErr(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "application/json") // ошибки тоже JSON
	if errors.Is(err, domain.ErrNotFound) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		return
	}
	w.WriteHeader(http.StatusBadRequest) // остальные доменные ошибки — 400
	_ = json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
