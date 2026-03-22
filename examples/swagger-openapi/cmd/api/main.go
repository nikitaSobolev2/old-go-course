package main // HTTP-каталог с Swagger UI; аннотации swag для генерации OpenAPI

import (
	"context"   // shutdown сервера
	"log"       // адрес и Swagger UI URL
	"net/http"  // Server
	"os"        // сигналы
	"os/signal" // NotifyContext
	"syscall"   // SIGTERM
	"time"      // таймауты

	"github.com/go-chi/chi/v5"                   // корневой роутер: swagger + API
	httpSwagger "github.com/swaggo/http-swagger" // отдаёт Swagger UI по /swagger/*

	"github.com/example/go-examples/swagger-openapi/internal/application"                 // CatalogService
	httpapi "github.com/example/go-examples/swagger-openapi/internal/infrastructure/http" // хендлеры с godoc-аннотациями
	"github.com/example/go-examples/swagger-openapi/internal/infrastructure/persistence"  // фикстуры продуктов

	_ "github.com/example/go-examples/swagger-openapi/docs" // side-effect: регистрация сгенерированной спецификации swag
)

// @title           Swagger/OpenAPI catalog example
// @version         1.0
// @description     DDD example with **Swagger/OpenAPI** (not "twigger"). Generated via swaggo/swag.
// @BasePath        /

func main() {
	repo := persistence.NewMemoryProductRepository() // предзаполненный каталог (демо)
	svc := application.NewCatalogService(repo) // чтение товара по id
	h := httpapi.NewHandlers(svc) // хендлер с swag-аннотациями на getProduct

	r := chi.NewRouter() // единая точка маршрутизации
	r.Get("/swagger/*", httpSwagger.WrapHandler) // UI и статика OpenAPI из пакета docs
	r.Mount("/", h.Routes()) // API под корнем: GET /v1/products/{id}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r, // chi на верхнем уровне
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("swagger-openapi listening on :8080 — Swagger UI at /swagger/index.html") // подсказка для браузера
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
