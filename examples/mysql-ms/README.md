# mysql-ms — MySQL + migrations (DDD)

- Driver: `database/sql` + `github.com/go-sql-driver/mysql`
- Migrations: `golang-migrate` (`migrations/*.sql`)
- Persistence: `internal/infrastructure/persistence/mysql`

## Run MySQL

```bash
docker compose up -d
```

Wait until healthy, then:

```bash
cd examples/mysql-ms
go run ./cmd/api
```

## Env

See `.env.example`. Default DSN matches `docker-compose.yml` (`root` / `secret`, database `orders`).

## Migrations

Add new pairs `000002_name.up.sql` / `000002_name.down.sql`, then restart the app (migrations run on startup).
