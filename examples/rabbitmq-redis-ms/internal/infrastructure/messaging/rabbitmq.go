package messaging // RabbitMQ: topic exchange «orders», публикация order.created, очередь для воркера

import (
	"context"       // PublishWithContext для отмены/дедлайна
	"encoding/json" // тело сообщения JSON
	"fmt"           // обёртка ошибки ExchangeDeclare

	amqp "github.com/rabbitmq/amqp091-go" // низкоуровневый клиент

	"github.com/example/go-examples/rabbitmq-redis-ms/internal/domain" // OrderCreated
)

const (
	exchangeName = "orders" // topic exchange
	routingKey   = "order.created" // ключ маршрутизации для подписчиков
)

// Publisher publishes OrderCreated events to RabbitMQ.
type Publisher struct {
	ch *amqp.Channel // один канал на соединение (демо)
}

func NewPublisher(ch *amqp.Channel) (*Publisher, error) {
	if err := ch.ExchangeDeclare(exchangeName, "topic", true, false, false, false, nil); err != nil { // durable exchange
		return nil, fmt.Errorf("exchange: %w", err)
	}
	return &Publisher{ch: ch}, nil
}

func (p *Publisher) PublishOrderCreated(ctx context.Context, e domain.OrderCreated) error {
	body, err := json.Marshal(struct { // анонимная структура — контракт JSON для воркера
		OrderID    string `json:"order_id"`
		CustomerID string `json:"customer_id"`
	}{OrderID: e.OrderID, CustomerID: e.CustomerID})
	if err != nil {
		return err
	}
	return p.ch.PublishWithContext(ctx, exchangeName, routingKey, false, false, amqp.Publishing{ // mandatory/immediate false
		ContentType: "application/json",
		Body:        body,
	})
}

// SetupWorkerQueue declares a queue bound to the exchange for consumers.
func SetupWorkerQueue(ch *amqp.Channel) (queue string, err error) {
	q, err := ch.QueueDeclare("order_created", true, false, false, false, nil) // durable очередь
	if err != nil {
		return "", err
	}
	if err := ch.QueueBind(q.Name, routingKey, exchangeName, false, nil); err != nil { // bind по routing key
		return "", err
	}
	return q.Name, nil // имя очереди для Consume
}
