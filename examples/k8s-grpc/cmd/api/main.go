package main // gRPC-сервер котировок с health и reflection (удобно для k8s и grpcurl)

import (
	"context"   // таймаут принудительной остановки
	"log"       // лог адреса и ошибок listen/serve/shutdown
	"net"       // Listen TCP для gRPC
	"os"        // сигналы и getenv
	"os/signal" // Notify
	"syscall"   // SIGINT/SIGTERM
	"time"      // таймаут graceful stop

	"google.golang.org/grpc"                                // создание gRPC сервера
	"google.golang.org/grpc/health"                         // встроенный health service
	healthpb "google.golang.org/grpc/health/grpc_health_v1" // регистрация health
	"google.golang.org/grpc/reflection"                     // ServerReflection для grpcurl/postman

	quotev1 "github.com/example/go-examples/k8s-grpc/gen/quote/v1"                   // сгенерированные регистрация и типы
	"github.com/example/go-examples/k8s-grpc/internal/application"                   // use case котировок
	grpcinfra "github.com/example/go-examples/k8s-grpc/internal/infrastructure/grpc" // адаптер gRPC
)

func main() {
	addr := getenv("GRPC_ADDR", ":50051") // адрес из env (в k8s — Service port)
	lis, err := net.Listen("tcp", addr) // слушаем TCP до Accept
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	app := application.NewQuoteService() // прикладная логика без транспорта
	srv := grpc.NewServer() // сервер с дефолтными опциями
	quotev1.RegisterQuoteServiceServer(srv, grpcinfra.NewQuoteServer(app)) // связываем сервис с реализацией

	hs := health.NewServer() // состояние health checks
	hs.SetServingStatus("", healthpb.HealthCheckResponse_SERVING) // сервис готов принимать трафик
	healthpb.RegisterHealthServer(srv, hs) // регистрируем стандартный health endpoint
	reflection.Register(srv) // включаем reflection API

	go func() { // Serve блокирует — в горутине
		log.Printf("gRPC listening on %s", addr)
		if err := srv.Serve(lis); err != nil { // обработка соединений до Shutdown
			log.Fatalf("serve: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1) // буфер 1 — не пропустить сигнал при занятости
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM) // подписка на остановку
	<-stop // ждём сигнал

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // дедлайн на graceful
	defer cancel()
	done := make(chan struct{}) // сигнал завершения GracefulStop
	go func() {
		srv.GracefulStop() // перестаём принимать; ждём активные RPC
		close(done)
	}()
	select {
	case <-done: // graceful завершился вовремя
	case <-ctx.Done(): // таймаут — принудительно рвём соединения
		srv.Stop()
	}
	log.Println("shutdown complete")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
