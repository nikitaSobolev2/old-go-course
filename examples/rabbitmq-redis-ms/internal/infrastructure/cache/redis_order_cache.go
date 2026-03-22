package cache // Redis: ключи order:{id}, значение — JSON снимка заказа

import (
	"context" // для команд go-redis v9
	"errors"  // redis.Nil как «ключ не найден»

	"github.com/redis/go-redis/v9" // клиент с поддержкой context
)

// OrderCache stores serialized order JSON in Redis.
type OrderCache struct {
	rdb *redis.Client
}

func NewOrderCache(rdb *redis.Client) *OrderCache {
	return &OrderCache{rdb: rdb} // разделяем клиент с main
}

func keyOrder(id string) string { return "order:" + id } // префикс пространства ключей

func (c *OrderCache) SetOrderJSON(ctx context.Context, id string, json []byte) error {
	return c.rdb.Set(ctx, keyOrder(id), json, 0).Err() // TTL 0 — без автоистечения
}

func (c *OrderCache) GetOrderJSON(ctx context.Context, id string) ([]byte, error) {
	s, err := c.rdb.Get(ctx, keyOrder(id)).Bytes() // []byte напрямую
	if errors.Is(err, redis.Nil) { // отсутствие ключа не считаем ошибкой сервиса
		return nil, nil // пустой слайс — сигнал cache miss в application
	}
	return s, err
}

func (c *OrderCache) DeleteOrder(ctx context.Context, id string) error {
	return c.rdb.Del(ctx, keyOrder(id)).Err() // инвалидация при будущих UPDATE (в демо не вызывается)
}
