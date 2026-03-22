# concurrency-patterns — goroutines, channels, atomic

## What lives where

- **`internal/domain`** — rules only (`ValidateBatch`, `MaxItemsPerBatch`). No goroutines.
- **`internal/application`** — `Processor` orchestrates validation then delegates to `JobRunner`.
- **`internal/infrastructure/workers`** — worker pool, **semaphore** (`chan struct{}` buffer), **fan-in** via `WaitGroup`, **atomic** counters for processed/error totals.

## Run

```bash
go run ./cmd/demo
```

Tune worker count and semaphore size in `cmd/demo/main.go`.
