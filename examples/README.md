# Go DDD examples

Ten isolated modules under `examples/`, each with `cmd/`, `internal/domain`, `internal/application`, `internal/infrastructure`.

| Directory | Focus |
|-----------|--------|
| `rest-api` | Chi / `net/http`, FoodTech orders |
| `observability-ms` | Prometheus + OpenTelemetry |
| `k8s-grpc` | gRPC + Dockerfile + Kubernetes manifests |
| `fiber-ms` | Fiber HTTP adapter |
| `rabbitmq-redis-ms` | RabbitMQ events + Redis cache |
| `swagger-openapi` | Swagger / OpenAPI (swag) |
| `mysql-ms` | MySQL + golang-migrate |
| `postgres-ms` | PostgreSQL (pgx) + golang-migrate |
| `concurrency-patterns` | Worker pool, semaphore, atomic |
| `s3-storage` | S3 API (MinIO locally) |

Use the repo root `go.work` to work on all modules together.
