package httpapi // HTTP: health, JSON API котировок, Prometheus /metrics

import (
	"encoding/json" // JSON-ответы для health и quotes
	"net/http"      // Handler, ResponseWriter

	"github.com/example/go-examples/observability-ms/internal/application" // QuoteService
	"github.com/go-chi/chi/v5"                                             // лёгкий роутер
	"github.com/prometheus/client_golang/prometheus/promhttp"              // стандартный handler метрик
)

func NewRouter(svc *application.QuoteService) http.Handler {
	r := chi.NewRouter() // корневой mux
	r.Handle("/metrics", promhttp.Handler()) // endpoint для scrape Prometheus
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) { // liveness для k8s/load balancer
		w.Header().Set("Content-Type", "application/json") // явный тип тела
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"}) // минимальный JSON
	})
	r.Get("/v1/quotes/{symbol}", func(w http.ResponseWriter, r *http.Request) { // котировка по тикеру
		sym := chi.URLParam(r, "symbol") // извлекаем символ из пути
		q := svc.GetQuote(r.Context(), sym) // прикладная логика (мок)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"symbol": q.Symbol, "price": q.Price}) // отдаём цену
	})
	return r
}
