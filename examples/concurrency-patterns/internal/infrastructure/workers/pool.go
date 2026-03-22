package workers // реализация параллельного выполнения заданий (воркер-пул + семафор)

import (
	"context"     // отмена: воркеры и отправка заданий реагируют на ctx.Done()
	"sync"        // WaitGroup: дождаться завершения всех воркеров
	"sync/atomic" // атомарные счётчики без mutex для processed/errors
	"time"        // имитация задержки обработки

	"github.com/example/go-examples/concurrency-patterns/internal/application" // тип результата ProcessResult
	"github.com/example/go-examples/concurrency-patterns/internal/domain"      // тип задания OrderJob
)

// Pool is a worker pool with:
//   - bounded concurrency (goroutines)
//   - semaphore (buffered channel) limiting in-flight work
//   - fan-in results via WaitGroup + atomic counters
type Pool struct {
	workers int
	sem     chan struct{}
}

func NewPool(workers int, semaphore int) *Pool {
	if workers < 1 { // защита от некорректной конфигурации
		workers = 1 // минимум одна горутина-воркер
	}
	if semaphore < 1 { // семафор не может быть нулевым/отрицательным
		semaphore = 1 // минимум одно «разрешение» на параллельную работу
	}
	return &Pool{workers: workers, sem: make(chan struct{}, semaphore)} // буферизованный канал = счётчик слотов
}

// Run implements application.JobRunner.
func (p *Pool) Run(ctx context.Context, jobs []domain.OrderJob) application.ProcessResult {
	var processed, errors int64 // счётчики успехов/условных ошибок (атомарно наращиваются в воркерах)
	jobsCh := make(chan domain.OrderJob) // очередь заданий: продюсер — цикл ниже, потребители — воркеры
	var wg sync.WaitGroup // ждём завершения всех горутин worker после close(jobsCh)

	worker := func() { // замыкание: читает из jobsCh и обрабатывает с учётом семафора и ctx
		defer wg.Done() // уменьшаем счётчик WaitGroup при выходе воркера (после закрытия канала и drain)
		for j := range jobsCh { // читаем пока канал открыт; после close — цикл завершится
			select { // либо отмена контекста, либо захват слота семафора
			case <-ctx.Done(): // родитель отменил контекст — прекращаем обработку
				return // выходим из воркера; defer wg.Done выполнится
			case p.sem <- struct{}{}: // отправка в канал-семафор = занять один слот «in-flight»
			}
			// simulate work; release semaphore when done
			func(job domain.OrderJob) { // IIFE: отдельная область для defer освобождения семафора на каждую job
				defer func() { <-p.sem }() // по завершении «работы» освобождаем слот семафора
				time.Sleep(5 * time.Millisecond) // имитация I/O или вычислений
				if len(job.OrderID)%2 == 0 { // чётная длина ID — считаем успех (демо-логика)
					atomic.AddInt64(&processed, 1) // безопасно из нескольких горутин
				} else { // нечётная длина — условная «ошибка»
					atomic.AddInt64(&errors, 1)
				}
			}(j) // передаём копию j в горутину-обработчик (здесь синхронно, но defer отработает после sleep)
		}
	}

	wg.Add(p.workers) // ожидаем завершения ровно стольких воркеров
	for i := 0; i < p.workers; i++ { // запускаем фиксированное число горутин
		go worker() // каждая крутит свой цикл for range jobsCh
	}

	for _, j := range jobs { // продюсер: скармливаем задания каналу или выходим по отмене
		select {
		case <-ctx.Done(): // таймаут/отмена во время отправки
			close(jobsCh) // закрываем канал — воркеры дочитают остаток и завершатся
			wg.Wait() // ждём всех воркеров
			return application.ProcessResult{Processed: atomic.LoadInt64(&processed), Errors: atomic.LoadInt64(&errors)} // частичный результат
		case jobsCh <- j: // задание поставлено в очередь
		}
	}
	close(jobsCh) // все задания отправлены — сигнал воркерам, что больше не будет
	wg.Wait() // дожидаемся, пока все воркеры обработают хвост и выйдут из for range
	return application.ProcessResult{Processed: atomic.LoadInt64(&processed), Errors: atomic.LoadInt64(&errors)} // финальные счётчики
}
