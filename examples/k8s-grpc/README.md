# k8s-grpc — gRPC + Kubernetes (DDD)

**Note:** In the Go ecosystem the RPC stack used here is **gRPC** (Protocol Buffers + `google.golang.org/grpc`), not tRPC.

## Layout

- `cmd/api` — wiring, graceful shutdown
- `internal/domain` — `Quote`, domain errors
- `internal/application` — `QuoteService` use case
- `internal/infrastructure/grpc` — gRPC handlers
- `proto/quote/v1/quote.proto` — service definition
- `gen/quote/v1` — generated code (`make proto` after installing `protoc` + plugins)

## Regenerate protobuf

Requires `protoc`, `protoc-gen-go`, and `protoc-gen-go-grpc` on `PATH`:

```bash
make proto
```

## Run locally

```bash
go run ./cmd/api
```

Default listen address: `:50051` (override with `GRPC_ADDR`).

## grpcurl (server reflection enabled)

List services:

```bash
grpcurl -plaintext localhost:50051 list
```

Call `GetQuote`:

```bash
grpcurl -plaintext -d '{"symbol":"AAPL"}' localhost:50051 quote.v1.QuoteService/GetQuote
```

## Docker

```bash
docker build -t k8s-grpc:latest .
```

## Kubernetes

Apply manifests (build/load image into your cluster first, e.g. `kind load docker-image k8s-grpc:latest`):

```bash
kubectl apply -f deploy/k8s/
```

Port-forward:

```bash
kubectl port-forward svc/k8s-grpc 50051:50051
```

## Optional: Ingress

For gRPC behind Ingress you typically need an Ingress controller that supports gRPC (e.g. NGINX with `grpc` annotations). This example uses a ClusterIP `Service` only.
