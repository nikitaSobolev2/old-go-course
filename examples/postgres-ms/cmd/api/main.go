package main // REST заказов с PostgreSQL через драйвер pgx и миграциями golang-migrate

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // регистрация драйвера "pgx" для database/sql

	"github.com/example/go-examples/postgres-ms/internal/application"
	httpapi "github.com/example/go-examples/postgres-ms/internal/infrastructure/http"
	"github.com/example/go-examples/postgres-ms/internal/infrastructure/persistence/postgres"
)

func main() {
	dsn := getenv("POSTGRES_DSN", "postgres://postgres:secret@127.0.0.1:5432/orders?sslmode=disable") // URI PostgreSQL
	db, err := sql.Open("pgx", dsn) // имя драйвера из jackc/pgx/stdlib
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	defer func() { _ = db.Close() }()
	db.SetMaxOpenConns(10) // ограничение соединений к Postgres
	if err := db.Ping(); err != nil {
		log.Fatalf("ping: %v", err)
	}

	migDir := getenv("MIGRATIONS_DIR", "migrations") // каталог относительно cwd
	if err := postgres.RunMigrations(db, migDir); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := postgres.NewOrderRepository(db)
	svc := application.NewOrderService(repo)
	h := httpapi.NewHandlers(svc)

	srv := &http.Server{
		Addr:         getenv("HTTP_ADDR", ":8080"),
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("postgres-ms listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
