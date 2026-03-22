# rabbitmq-redis-ms

DDD example: in-memory orders, **Redis** cache for reads, **RabbitMQ** for `OrderCreated` domain events.

## Stack

- `cmd/api` — HTTP API + publisher
- `cmd/worker-replica` — один consumer на процесс (масштабирование: **несколько инстансов** / несколько терминалов)
- `cmd/worker-pool` — **пул горутин** в одном процессе (параллельная обработка + prefetch + manual Ack)
- Docker Compose — Redis + RabbitMQ (management UI on http://localhost:15672)

### Какой воркер выбрать

| Команда | Идея |
|--------|------|
| `go run ./cmd/worker-replica` | Один поток чтения из AMQP на процесс. Больше пропускной способности → **запустите 2+ процесса** (или реплики в k8s). |
| `go run ./cmd/worker-pool` | Один процесс, **несколько горутин** обрабатывают сообщения; лимит через `WORKER_POOL_SIZE`, prefetch через `WORKER_PREFETCH`. |

## Run dependencies

```bash
docker compose up -d
```

## Run API

```bash
go run ./cmd/api
```

## Run worker (отдельный терминал)

**Вариант A — реплики (простой consumer):**

```bash
go run ./cmd/worker-replica
```

Для сравнения масштабирования откройте второй терминал и снова:

```bash
go run ./cmd/worker-replica
```

Оба подписаны на одну очередь — брокер распределяет сообщения между ними.

**Вариант B — пул горутин в одном процессе:**

```bash
go run ./cmd/worker-pool
```

Опционально:

```bash
set WORKER_POOL_SIZE=8
set WORKER_PREFETCH=16
go run ./cmd/worker-pool
```

Create an order with `POST /orders`; воркер(и) печатают JSON payload.

## Env

See `.env.example`.
