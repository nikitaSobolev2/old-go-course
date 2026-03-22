package application // слой use case: связывает domain и способ выполнения (интерфейс JobRunner)

import (
	"context" // отмена и таймауты при пробросе в runner

	"github.com/example/go-examples/concurrency-patterns/internal/domain" // валидация пакета перед запуском
)

// ProcessResult aggregates outcomes from concurrent workers.
type ProcessResult struct {
	Processed int64
	Errors    int64
}

// JobRunner executes validated jobs concurrently (implemented in infrastructure).
type JobRunner interface {
	Run(ctx context.Context, jobs []domain.OrderJob) ProcessResult
}

// Processor wires domain validation to concurrent execution.
type Processor struct {
	runner JobRunner
}

func NewProcessor(runner JobRunner) *Processor {
	return &Processor{runner: runner} // сохраняем реализацию (пул воркеров) за интерфейсом
}

func (p *Processor) Run(ctx context.Context, jobs []domain.OrderJob) ProcessResult {
	if err := domain.ValidateBatch(jobs); err != nil { // сначала доменные правила — без воркеров
		return ProcessResult{} // при ошибке валидации возвращаем нулевой результат (без счётчиков)
	}
	return p.runner.Run(ctx, jobs) // пакет валиден — делегируем инфраструктуре (конкурентный запуск)
}
