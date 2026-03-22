# fiber-ms — Fiber HTTP adapter + DDD

Fiber is used only in `internal/infrastructure/http`. **Domain** and **application** layers do not import Fiber.

## Run

```bash
go run ./cmd/api
```

`HTTP_ADDR` defaults to `:8080`.

## API

- `POST /orders` — create order (same JSON shape as `examples/rest-api`)
- `GET /orders/:id` — fetch order

## Example

```bash
curl -s -X POST localhost:8080/orders -H "Content-Type: application/json" -d "{\"customer_id\":\"c1\",\"items\":[{\"product_id\":\"p1\",\"name\":\"Coffee\",\"quantity\":2,\"price_cents\":250}]}"
```
