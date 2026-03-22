package domain // доменная котировка и ошибка валидации символа

import "errors"

// ErrInvalidSymbol is returned when a quote symbol is empty or invalid.
var ErrInvalidSymbol = errors.New("invalid symbol") // пустой символ после trim

// Quote is a market quote for a symbol (FinTech bounded context).
type Quote struct {
	Symbol string
	Price  float64
}
