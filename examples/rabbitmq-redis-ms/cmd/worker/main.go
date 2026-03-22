package main // отдельный процесс: читает очередь order_created и логирует тело сообщения

import (
	"context" // отмена по сигналу
	"log"     // диагностика и события
	"os"      // сигналы
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go" // consumer API

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/messaging" // объявление очереди и bind
)

func main() {
	amqpURL := getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/") // тот же брокер, что у API
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		log.Fatalf("amqp dial: %v", err)
	}
	defer func() { _ = conn.Close() }()
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("amqp channel: %v", err)
	}
	defer func() { _ = ch.Close() }()

	qName, err := messaging.SetupWorkerQueue(ch) // очередь durable + bind к exchange orders
	if err != nil {
		log.Fatalf("queue setup: %v", err)
	}

	msgs, err := ch.Consume(qName, "", true, false, false, false, nil) // auto-ack; consumer tag пустой
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // стоп по Ctrl+C
	defer stop()

	log.Printf("worker consuming queue %s", qName)
	for { // цикл до отмены контекста или закрытия канала сообщений
		select {
		case <-ctx.Done():
			log.Println("shutdown")
			return
		case d, ok := <-msgs:
			if !ok { // канал закрыт брокером
				return
			}
			log.Printf("OrderCreated event: %s", string(d.Body)) // демо-обработка: только лог JSON
		}
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
