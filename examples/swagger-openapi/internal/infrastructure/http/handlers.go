package httpapi // REST + JSON; swag-аннотации на хендлере для генерации OpenAPI

import (
	"encoding/json" // ответы и тело ошибки 404
	"errors"        // errors.Is
	"net/http"      // стандартный обработчик

	"github.com/go-chi/chi/v5" // параметр пути {id}

	"github.com/example/go-examples/swagger-openapi/internal/application" // CatalogService
	"github.com/example/go-examples/swagger-openapi/internal/domain"      // ErrNotFound
)

type Handlers struct {
	svc *application.CatalogService
}

func NewHandlers(svc *application.CatalogService) *Handlers {
	return &Handlers{svc: svc}
}

func (h *Handlers) Routes() chi.Router {
	r := chi.NewRouter()
	r.Get("/v1/products/{id}", h.getProduct) // маршрут, описанный в swag @Router
	return r
}

// getProduct godoc
// @Summary      Get product by ID
// @Description  Returns a catalog product
// @Tags         catalog
// @Produce      json
// @Param        id   path      string  true  "Product ID"
// @Success      200  {object}  ProductResponse
// @Failure      404  {object}  ErrorResponse
// @Router       /v1/products/{id} [get]
func (h *Handlers) getProduct(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id") // извлекаем id из шаблона
	p, err := h.svc.GetProduct(r.Context(), id) // домен + persistence
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) { // маппинг на 404 JSON
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(ErrorResponse{Error: "not found"})
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError) // прочие ошибки — plain text 500
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ProductResponse{ID: p.ID, Name: p.Name, Price: p.Price}) // успех 200
}

// ProductResponse is a public DTO (mirrors OpenAPI schema).
type ProductResponse struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
