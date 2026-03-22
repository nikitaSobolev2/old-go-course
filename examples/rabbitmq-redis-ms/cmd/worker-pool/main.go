// Режим «пул горутин»: один процесс, несколько параллельных обработчиков сообщений.
// QoS (prefetch) ограничивает число неподтверждённых сообщений у этого consumer’а; manual Ack после обработки.
// Масштабирование по CPU внутри машины; для нескольких машин дополнительно используйте worker-replica.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/infrastructure/messaging"
)

func main() {
	workers := getenvInt("WORKER_POOL_SIZE", 4)
	if workers < 1 {
		workers = 1
	}
	prefetch := getenvInt("WORKER_PREFETCH", workers*2)
	if prefetch < 1 {
		prefetch = workers
	}

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

	// Сколько сообщений брокер может отправить без Ack (распределение нагрузки между горутинами).
	if err := ch.Qos(prefetch, 0, false); err != nil {
		log.Fatalf("qos: %v", err)
	}

	qName, err := messaging.SetupWorkerQueue(ch)
	if err != nil {
		log.Fatalf("queue setup: %v", err)
	}

	// autoAck=false — подтверждаем после успешной обработки в воркере.
	msgs, err := ch.Consume(qName, "", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("consume: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Семафор: не больше `workers` одновременных обработок.
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup

	log.Printf("[pool] queue=%s workers=%d prefetch=%d", qName, workers, prefetch)

loop:
	for {
		select {
		case <-ctx.Done():
			break loop
		case d, ok := <-msgs:
			if !ok {
				break loop
			}
			wg.Add(1)
			sem <- struct{}{}
			go func(d amqp.Delivery) {
				defer wg.Done()
				defer func() { <-sem }()
				// Демо: «обработка» — лог; в проде здесь идемпотентная бизнес-логика.
				log.Printf("OrderCreated event: %s", string(d.Body))
				if err := d.Ack(false); err != nil {
					log.Printf("ack: %v", err)
				}
			}(d)
		}
	}

	log.Println("waiting in-flight handlers...")
	wg.Wait()
	log.Println("shutdown complete")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvInt(k string, def int) int {
	s := os.Getenv(k)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}
