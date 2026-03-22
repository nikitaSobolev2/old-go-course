package domain // пакет доменных правил без зависимостей от инфраструктуры

import "errors" // конструктор стандартных ошибок

// ErrInvalidBatch is returned when a batch exceeds configured limits.
var ErrInvalidBatch = errors.New("invalid batch") // сентинел-ошибка: пакет заданий не прошёл проверку

// OrderJob represents work to process in the bounded context (rules only).
type OrderJob struct {
	OrderID string
	Items   int
}

// MaxItemsPerBatch is a domain rule (example).
const MaxItemsPerBatch = 5 // верхняя граница количества позиций в одном задании (пример бизнес-правила)

// ValidateBatch checks business rules before infrastructure runs workers.
func ValidateBatch(jobs []OrderJob) error {
	if len(jobs) == 0 { // пустой список заданий запрещён
		return ErrInvalidBatch // возвращаем доменную ошибку без деталей
	}
	for _, j := range jobs { // проверяем каждое задание в пакете
		if j.Items <= 0 || j.Items > MaxItemsPerBatch { // количество позиций должно быть в допустимом диапазоне
			return ErrInvalidBatch // нарушение правила — тот же сентинел
		}
	}
	return nil // все задания удовлетворяют правилам
}
