# rabbitmq-redis-ms

DDD example: in-memory orders, **Redis** cache for reads, **RabbitMQ** for `OrderCreated` domain events.

## Stack

- `cmd/api` — HTTP API + publisher
- `cmd/worker` — consumer (logs events; extend with projections)
- Docker Compose — Redis + RabbitMQ (management UI on http://localhost:15672)

## Run dependencies

```bash
docker compose up -d
```

## Run API

```bash
go run ./cmd/api
```

## Run worker (separate terminal)

```bash
go run ./cmd/worker
```

Create an order with `POST /orders`; the worker should print the JSON payload.

## Env

See `.env.example`.
