package main // точка входа исполняемой программы

import (
	"context" // отмена по таймауту и передача дедлайна вниз по вызовам
	"fmt"     // форматированный вывод в stdout
	"log"     // логирование сообщений (здесь — ожидаемая ошибка валидации)
	"time"    // длительности для WithTimeout

	"github.com/example/go-examples/concurrency-patterns/internal/application" // сценарий: валидация + запуск воркеров
	"github.com/example/go-examples/concurrency-patterns/internal/domain"      // доменные правила и модель задания
	"github.com/example/go-examples/concurrency-patterns/internal/infrastructure/workers"
)

func main() {
	pool := workers.NewPool(4, 2) // пул: 4 горутины-воркера, семафор на 2 одновременных «тяжёлых» задачи
	proc := application.NewProcessor(pool) // use-case: перед запуском проверяет пакет через domain

	jobs := []domain.OrderJob{ // список заказов для обработки (демо-данные)
		{OrderID: "a1", Items: 2},
		{OrderID: "b2", Items: 3},
		{OrderID: "c3", Items: 1},
		{OrderID: "d4", Items: 4},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second) // общий дедлайн 2 с на весь прогон
	defer cancel() // по выходе из main отменяем контекст (освобождаем таймер и сигнализируем горутинам)

	res := proc.Run(ctx, jobs) // валидация batch + параллельная обработка через pool
	fmt.Printf("processed=%d errors=%d\n", res.Processed, res.Errors) // печать итоговых счётчиков

	if err := domain.ValidateBatch([]domain.OrderJob{{OrderID: "x", Items: 99}}); err != nil { // демо: Items > MaxItemsPerBatch
		log.Printf("expected validation error: %v", err) // ожидаемая ошибка — показываем, что правило сработало
	}
}
