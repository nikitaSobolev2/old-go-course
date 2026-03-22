// Режим «несколько инстансов»: один consumer на процесс, без горутин-пула.
// Масштабирование — запуск нескольких процессов (несколько терминалов, несколько реплик в k8s).
// RabbitMQ сам round-robin между consumer’ами одной очереди; одно сообщение — одному процессу.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/messaging"
)

func main() {
	amqpURL := getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
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

	qName, err := messaging.SetupWorkerQueue(ch)
	if err != nil {
		log.Fatalf("queue setup: %v", err)
	}

	// Один consumer на процесс. Несколько процессов = несколько consumer’ов на ту же очередь.
	msgs, err := ch.Consume(qName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("[replica] single consumer on queue %s (scale: run more processes)", qName)
	for {
		select {
		case <-ctx.Done():
			log.Println("shutdown")
			return
		case d, ok := <-msgs:
			if !ok {
				return
			}
			log.Printf("OrderCreated event: %s", string(d.Body))
		}
	}
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
