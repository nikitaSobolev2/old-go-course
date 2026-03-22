# swagger-openapi — Swagger / OpenAPI (DDD)

This example uses **Swagger/OpenAPI** via **swaggo/swag** + **http-swagger** (Chi for HTTP). It is **not** “twigger” — the plan explicitly maps that to Swagger/OpenAPI.

## Regenerate API docs

After changing annotations in `cmd/api` or `internal/infrastructure/http`:

```bash
go install github.com/swaggo/swag/cmd/swag@v1.16.3
make swagger
```

## Run

```bash
go run ./cmd/api
```

- API: `GET /v1/products/{id}` (try `p1`)
- Swagger UI: http://localhost:8080/swagger/index.html

## Layout

- `docs/` — generated `docs.go`, `swagger.json`, `swagger.yaml` (re-run `make swagger` when handlers change; **`docs.go` is not hand-commented** — it is produced by swag)
- `internal/domain` — `Product`, repository interface
- `internal/application` — `CatalogService`
- `internal/infrastructure` — HTTP + in-memory persistence
