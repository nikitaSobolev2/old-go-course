package main // классический net/http + chi: REST заказов с in-memory хранилищем

import (
	"context"   // shutdown сервера
	"log"       // адрес и фатальные ошибки
	"net/http"  // Server, ErrServerClosed
	"os"        // сигналы
	"os/signal" // NotifyContext
	"syscall"   // SIGTERM
	"time"      // таймауты сервера и graceful shutdown

	"github.com/example/go-examples/rest-api/internal/application"                 // OrderService
	httpapi "github.com/example/go-examples/rest-api/internal/infrastructure/http" // chi handlers
	"github.com/example/go-examples/rest-api/internal/infrastructure/persistence"  // memory repo
)

func main() {
	repo := persistence.NewMemoryOrderRepository() // потокобезопасная map заказов
	svc := application.NewOrderService(repo) // сценарии создания и чтения
	h := httpapi.NewHandlers(svc) // регистрация маршрутов

	srv := &http.Server{ // стандартный сервер Go
		Addr:         ":8080", // фиксированный порт демо
		Handler:      h.Routes(), // chi.Router реализует http.Handler
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second, // keep-alive
	}

	go func() { // ListenAndServe блокирует
		log.Printf("rest-api listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { // Shutdown завершает без ошибки
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // ждём сигнал остановки
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // время на завершение запросов
	defer cancel()
	_ = srv.Shutdown(shutdownCtx) // закрываем listener и ждём обработчики
}
