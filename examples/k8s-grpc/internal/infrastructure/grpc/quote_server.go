package grpcinfra // gRPC-адаптер: protobuf ↔ domain/application

import (
	"context" // контекст RPC от клиента

	"google.golang.org/grpc/codes"  // коды ошибок gRPC
	"google.golang.org/grpc/status" // преобразование Go error → status.Error

	quotev1 "github.com/example/go-examples/k8s-grpc/gen/quote/v1" // сгенерированный сервис и сообщения
	"github.com/example/go-examples/k8s-grpc/internal/application" // QuoteService
	"github.com/example/go-examples/k8s-grpc/internal/domain"      // ErrInvalidSymbol
)

// QuoteServer implements quotev1.QuoteServiceServer (gRPC adapter).
type QuoteServer struct {
	quotev1.UnimplementedQuoteServiceServer // встраиваем для forward-compat API
	app *application.QuoteService
}

func NewQuoteServer(app *application.QuoteService) *QuoteServer {
	return &QuoteServer{app: app} // DI use case
}

func (s *QuoteServer) GetQuote(ctx context.Context, req *quotev1.GetQuoteRequest) (*quotev1.GetQuoteResponse, error) {
	q, err := s.app.GetQuote(ctx, req.GetSymbol()) // извлекаем символ из protobuf-запроса
	if err != nil {
		if err == domain.ErrInvalidSymbol { // явное сравнение с сентинелом
			return nil, status.Error(codes.InvalidArgument, err.Error()) // 400-уровень в gRPC
		}
		return nil, status.Error(codes.Internal, err.Error()) // прочие — внутренняя ошибка
	}
	return &quotev1.GetQuoteResponse{Symbol: q.Symbol, Price: q.Price}, nil // успешный ответ protobuf
}
