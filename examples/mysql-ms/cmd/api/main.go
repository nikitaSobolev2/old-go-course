package main // REST заказов с персистенцией в MySQL и миграциями golang-migrate

import (
	"context"       // shutdown HTTP
	"database/sql"  // пул соединений с БД
	"log"           // фатальные ошибки и адрес
	"net/http"      // Server
	"os"            // env и сигналы
	"os/signal"     // NotifyContext
	"path/filepath" // путь к каталогу миграций
	"syscall"       // SIGTERM
	"time"          // таймауты

	_ "github.com/go-sql-driver/mysql" // регистрация драйвера "mysql" для database/sql

	"github.com/example/go-examples/mysql-ms/internal/application"                      // OrderService
	httpapi "github.com/example/go-examples/mysql-ms/internal/infrastructure/http"      // chi + JSON
	"github.com/example/go-examples/mysql-ms/internal/infrastructure/persistence/mysql" // репозиторий и миграции
)

func main() {
	dsn := getenv("MYSQL_DSN", "root:secret@tcp(127.0.0.1:3306)/orders?parseTime=true&multiStatements=true") // DSN для локального docker-compose
	db, err := sql.Open("mysql", dsn) // ленивый пул; соединение по Ping
	if err != nil {
		log.Fatalf("sql open: %v", err)
	}
	defer func() { _ = db.Close() }() // освобождаем ресурсы при выходе
	db.SetMaxOpenConns(10) // ограничение одновременных соединений с MySQL
	db.SetMaxIdleConns(2) // простаивающие соединения в пуле
	if err := db.Ping(); err != nil { // проверка доступности сервера до миграций
		log.Fatalf("ping: %v", err)
	}

	migDir := getenv("MIGRATIONS_DIR", "") // можно переопределить путь к SQL-файлам
	if migDir == "" {
		migDir = filepath.Join("migrations") // по умолчанию ./migrations от cwd
	}
	if err := mysql.RunMigrations(db, migDir); err != nil { // применяем up до последней версии
		log.Fatalf("migrate: %v", err)
	}

	repo := mysql.NewOrderRepository(db) // репозиторий на том же *sql.DB
	svc := application.NewOrderService(repo) // use case
	h := httpapi.NewHandlers(svc) // HTTP-слой

	srv := &http.Server{
		Addr:         getenv("HTTP_ADDR", ":8080"),
		Handler:      h.Routes(),
		ReadTimeout:  15 * time.Second, // запросы могут включать тело заказа
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("mysql-ms listening on %s", srv.Addr)
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
