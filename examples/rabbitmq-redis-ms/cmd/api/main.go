package main // HTTP API: создание заказа → память + Redis кэш + событие в RabbitMQ

import (
	"context"   // Ping Redis, shutdown HTTP
	"log"       // ошибки подключений
	"net/http"  // Server
	"os"        // env, сигналы
	"os/signal" // graceful stop
	"syscall"   // SIGTERM
	"time"      // таймауты сервера

	amqp "github.com/rabbitmq/amqp091-go" // клиент AMQP для publisher
	"github.com/redis/go-redis/v9"        // клиент Redis для кэша заказов

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/application"                     // OrderService с cache+events
	rediscache "github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/cache" // адаптер Redis
	httpapi "github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/http"     // REST
	"github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/messaging"        // publisher RabbitMQ
	"github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/persistence"      // источник истины в памяти
)

func main() {
	redisAddr := getenv("REDIS_ADDR", "localhost:6379") // хост:порт Redis
	amqpURL := getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/") // строка подключения к брокеру

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr}) // один клиент на процесс
	defer func() { _ = rdb.Close() }()
	if err := rdb.Ping(context.Background()).Err(); err != nil { // проверка до приёма HTTP
		log.Fatalf("redis: %v", err)
	}

	conn, err := amqp.Dial(amqpURL) // TCP + AMQP handshake
	if err != nil {
		log.Fatalf("amqp dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	ch, err := conn.Channel() // канал для declare/publish (в демо один на всё)
	if err != nil {
		log.Fatalf("amqp channel: %v", err)
	}
	defer func() { _ = ch.Close() }()

	pub, err := messaging.NewPublisher(ch) // объявляет exchange topic «orders»
	if err != nil {
		log.Fatalf("publisher: %v", err)
	}

	repo := persistence.NewMemoryOrderRepository() // система записи заказов
	cache := rediscache.NewOrderCache(rdb) // read-through кэш JSON
	svc := application.NewOrderService(repo, pub, cache) // три порта
	h := httpapi.NewHandlers(svc)

	addr := getenv("HTTP_ADDR", ":8080")
	srv := &http.Server{
		Addr:         addr,
		Handler:      h.Routes(),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		log.Printf("api listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
