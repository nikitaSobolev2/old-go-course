package main // HTTP API заказов на Fiber с in-memory репозиторием

import (
	"context"   // таймаут graceful shutdown
	"log"       // лог адреса и фатальных ошибок
	"os"        // переменные окружения и сигналы
	"os/signal" // NotifyContext
	"syscall"   // SIGTERM
	"time"      // таймауты Fiber и shutdown

	"github.com/gofiber/fiber/v2" // веб-фреймворк

	"github.com/example/go-examples/fiber-ms/internal/application"                   // сценарии Create/Get order
	httpfiber "github.com/example/go-examples/fiber-ms/internal/infrastructure/http" // хендлеры Fiber
	"github.com/example/go-examples/fiber-ms/internal/infrastructure/persistence"    // память вместо БД
)

func main() {
	repo := persistence.NewMemoryOrderRepository() // потокобезопасное хранилище в процессе
	svc := application.NewOrderService(repo) // use case с зависимостью от порта OrderRepository
	h := httpfiber.NewHandlers(svc) // регистрация маршрутов на сервис

	app := fiber.New(fiber.Config{ // создаём приложение с таймаутами соединений
		ReadTimeout:  10 * time.Second, // чтение запроса
		WriteTimeout: 10 * time.Second, // запись ответа
		IdleTimeout:  120 * time.Second, // keep-alive
	})
	h.Register(app) // POST /orders, GET /orders/:id

	addr := getenv("HTTP_ADDR", ":8080") // адрес из env или демо по умолчанию
	go func() { // Listen блокирует — в отдельной горутине
		log.Printf("fiber-ms listening on %s", addr)
		if err := app.Listen(addr); err != nil { // старт TCP listener
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // ожидание сигнала остановки
	defer stop() // освободить ресурсы подписки
	<-ctx.Done() // блок до SIGINT/SIGTERM

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // лимит на завершение активных запросов
	defer cancel()
	_ = app.ShutdownWithContext(shutdownCtx) // мягкая остановка Fiber
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { // непустая переменная
		return v
	}
	return def
}
