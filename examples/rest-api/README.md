# REST API (DDD) — FoodTech orders

## Stack

- Chi router
- In-memory repository (replace with Postgres in production)

## Run

```bash
cd examples/rest-api
go run ./cmd/api
```

## Example

```bash
curl -X POST http://localhost:8080/orders -H "Content-Type: application/json" -d "{\"customer_id\":\"c1\",\"items\":[{\"product_id\":\"p1\",\"name\":\"Pizza\",\"quantity\":2,\"price_cents\":500}]}"
curl http://localhost:8080/orders/<id>
```

## Layout

- `internal/domain` — Order aggregate, repository interface
- `internal/application` — OrderService use cases
- `internal/infrastructure` — HTTP (DTO), persistence
