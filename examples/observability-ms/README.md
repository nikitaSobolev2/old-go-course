# Observability micro-service (Prometheus + OpenTelemetry)

- **Domain:** `Quote` (FinTech mock)
- **Application:** `QuoteService`
- **Infrastructure:** Chi, Prometheus `/metrics`, OTel HTTP middleware + stdout traces

## Run

```bash
go run ./cmd/api
curl http://localhost:8080/v1/quotes/AAPL
curl http://localhost:8080/metrics
```

## Prometheus (Docker)

```bash
docker compose up -d
# Scrape target: app must be reachable from container (adjust host in deploy/prometheus.yml)
```

Traces print to stdout (stdout exporter). For OTLP/Jaeger, replace `internal/infrastructure/observability/otel.go` with OTLP exporter.
