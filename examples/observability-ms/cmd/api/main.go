package main // демо HTTP с OpenTelemetry tracing и Prometheus metrics

import (
	"context"   // shutdown трейсера и HTTP-сервера
	"log"       // фатальные ошибки и адрес сервера
	"net/http"  // Server, ErrServerClosed
	"os"        // сигналы ОС
	"os/signal" // NotifyContext
	"syscall"   // SIGTERM
	"time"      // таймауты HTTP

	"github.com/example/go-examples/observability-ms/internal/application"                  // мок-сервис котировок
	httpapi "github.com/example/go-examples/observability-ms/internal/infrastructure/http"  // chi + /metrics
	"github.com/example/go-examples/observability-ms/internal/infrastructure/observability" // инициализация OTel
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"                         // обёртка handler для авто-спанов
)

func main() {
	shutdown, err := observability.InitTracer() // поднимаем TracerProvider и stdout exporter
	if err != nil {
		log.Fatalf("otel: %v", err) // без трейсера можно продолжить, но демо падает явно
	}
	defer func() { _ = shutdown(context.Background()) }() // при выходе сбрасываем буфер экспортера

	svc := application.NewQuoteService() // доменный use case (мок цены)
	h := httpapi.NewRouter(svc) // маршруты /metrics, /health, /v1/quotes/{symbol}
	wrapped := otelhttp.NewHandler(h, "observability-ms") // имя сервиса в спанах HTTP

	srv := &http.Server{
		Addr:         ":8080", // фиксированный порт демо
		Handler:      wrapped, // все запросы проходят через OTel middleware
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() { // ListenAndServe не блокирует main
		log.Println("observability-ms :8080 (metrics /metrics)") // подсказка про Prometheus endpoint
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { // Shutdown даёт ErrServerClosed
			log.Fatal(err)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // graceful stop по сигналу
	defer stop()
	<-sigCtx.Done() // ждём Ctrl+C или SIGTERM
	_ = srv.Shutdown(context.Background()) // останавливаем приём новых соединений
}
