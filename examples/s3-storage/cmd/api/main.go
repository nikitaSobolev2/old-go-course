package main // точка входа HTTP-сервиса загрузки/скачивания в S3-совместимое хранилище

import (
	"context"   // корневой контекст и отмена при shutdown
	"log"       // логирование ошибок и адреса сервера
	"net/http"  // тип Server, ListenAndServe, ErrServerClosed
	"os"        // переменные окружения и сигналы
	"os/signal" // подписка на SIGINT/SIGTERM
	"syscall"   // константа SIGTERM
	"time"      // таймауты HTTP-сервера и graceful shutdown

	"github.com/aws/aws-sdk-go-v2/aws"         // вспомогательные указатели для API AWS SDK
	"github.com/aws/aws-sdk-go-v2/config"      // загрузка конфигурации клиента AWS
	"github.com/aws/aws-sdk-go-v2/credentials" // статические ключи для MinIO/S3
	"github.com/aws/aws-sdk-go-v2/service/s3"  // клиент операций с бакетом и объектами

	"github.com/example/go-examples/s3-storage/internal/application"                 // сервис файлов поверх порта ObjectStorage
	httpapi "github.com/example/go-examples/s3-storage/internal/infrastructure/http" // chi-роутер и хендлеры
	"github.com/example/go-examples/s3-storage/internal/infrastructure/storage"      // реализация S3
)

func main() {
	ctx := context.Background() // контекст для инициализации SDK (без отмены до shutdown HTTP)
	endpoint := getenv("S3_ENDPOINT", "http://127.0.0.1:9000") // URL MinIO или совместимого S3
	bucket := getenv("S3_BUCKET", "demo") // имя бакета для объектов
	region := getenv("AWS_REGION", "us-east-1") // регион в конфиге SDK (для MinIO часто формальный)
	access := getenv("S3_ACCESS_KEY", "minioadmin") // access key
	secret := getenv("S3_SECRET_KEY", "minioadmin") // secret key

	cfg, err := config.LoadDefaultConfig(ctx, // собираем конфигурацию AWS SDK v2
		config.WithRegion(region), // задаём регион
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(access, secret, "")), // явные ключи
	)
	if err != nil {
		log.Fatalf("aws config: %v", err) // без конфига клиент S3 не создать
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) { // создаём клиент S3 с кастомизацией
		o.BaseEndpoint = aws.String(endpoint) // перенаправляем на MinIO/localstack, не на AWS
		o.UsePathStyle = true // path-style URL: /bucket/key (удобно для MinIO)
	})

	if _, err := client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: aws.String(bucket)}); err != nil { // проверяем существование бакета
		if _, err := client.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)}); err != nil { // при отсутствии — создаём
			log.Fatalf("create bucket %q: %v", bucket, err) // демо ожидает рабочий бакет
		}
	}

	store := storage.NewS3Storage(client, bucket) // адаптер доменного порта к S3 API
	svc := application.NewFileService(store) // прикладной слой: валидация ключа и вызовы store
	h := httpapi.NewHandlers(svc) // HTTP-слой: маршруты PUT/GET /objects/{key}

	srv := &http.Server{ // настраиваем сервер с таймаутами на чтение/запись тел
		Addr:         getenv("HTTP_ADDR", ":8080"), // адрес прослушивания
		Handler:      h.Routes(), // chi-роутер как корневой handler
		ReadTimeout:  60 * time.Second, // лимит чтения запроса (большие загрузки)
		WriteTimeout: 60 * time.Second, // лимит ответа
	}

	go func() { // ListenAndServe блокирует — запускаем в горутине
		log.Printf("s3-storage listening on %s (bucket=%s endpoint=%s)", srv.Addr, bucket, endpoint) // диагностика
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed { // нормальное завершение Shutdown — не ошибка
			log.Fatal(err)
		}
	}()

	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // отмена по Ctrl+C или SIGTERM
	defer stop() // отписаться от сигналов при выходе
	<-sigCtx.Done() // блокируемся до получения сигнала

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // даём серверу до 10 с на завершение
	defer cancel() // освобождаем таймер
	_ = srv.Shutdown(shutdownCtx) // мягко закрываем listener и ждём активные запросы
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" { // если переменная задана и не пустая
		return v
	}
	return def // иначе значение по умолчанию для локального демо
}
