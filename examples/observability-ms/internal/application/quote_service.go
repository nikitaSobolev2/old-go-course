package application // сценарий получения котировки (изолирован от HTTP и OTel)

import (
	"context" // контекст передаётся для будущих реальных вызовов (сейчас не используется)

	"github.com/example/go-examples/observability-ms/internal/domain" // тип Quote
)

// QuoteService returns mock quotes (business logic isolated from HTTP).
type QuoteService struct{}

func NewQuoteService() *QuoteService { return &QuoteService{} } // конструктор без зависимостей

func (s *QuoteService) GetQuote(_ context.Context, symbol string) domain.Quote {
	return domain.Quote{Symbol: symbol, Price: 100.5} // фиксированная цена для демо трассировки
}
