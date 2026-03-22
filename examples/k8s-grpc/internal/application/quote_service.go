package application // сценарий получения котировки (независим от gRPC)

import (
	"context" // для сигнатуры; логика пока не использует отмену
	"strings" // нормализация символа

	"github.com/example/go-examples/k8s-grpc/internal/domain" // Quote, ErrInvalidSymbol
)

// QuoteService resolves quotes (use case layer; transport-agnostic).
type QuoteService struct{}

func NewQuoteService() *QuoteService { return &QuoteService{} } // без зависимостей

// GetQuote returns a deterministic mock price for the symbol.
func (s *QuoteService) GetQuote(_ context.Context, symbol string) (domain.Quote, error) {
	sym := strings.TrimSpace(symbol) // убираем пробелы
	if sym == "" { // бизнес-правило: символ обязателен
		return domain.Quote{}, domain.ErrInvalidSymbol
	}
	return domain.Quote{Symbol: strings.ToUpper(sym), Price: 100.5}, nil // демо-цена и upper-case тикер
}
