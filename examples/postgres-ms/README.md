# postgres-ms — PostgreSQL + migrations (DDD)

- Access: `database/sql` with **`pgx`** driver (`github.com/jackc/pgx/v5/stdlib`, driver name `pgx`)
- Migrations: `golang-migrate` + SQL in `migrations/`
- Repositories: `internal/infrastructure/persistence/postgres`

## Run Postgres

```bash
docker compose up -d
```

```bash
cd examples/postgres-ms
go run ./cmd/api
```

## Env

See `.env.example`.
