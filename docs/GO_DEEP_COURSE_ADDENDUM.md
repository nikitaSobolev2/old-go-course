# Приложение к курсу: дополнения к GO_DEEP_COURSE.md

Этот файл — **отдельная копия материала**, который был дописан в основной документ [GO_DEEP_COURSE.md](GO_DEEP_COURSE.md) (указатели, получатели методов, `defer`, `sync` vs БД, каналы, DDD-примеры, паттерны, Pub/Sub, шардирование, безопасность, шпаргалка CLI). Полный курс по-прежнему в `GO_DEEP_COURSE.md`.

---

## 1. Получатели методов: `func (s *OrderService)` и символы `*` и `&`

- **`func (s OrderService)` — value receiver:** в метод попадает **копия** структуры. Изменения полей **не** видны снаружи (если только не мутируете слайсы/мапы внутри копии — см. слайсы как ссылочные дескрипторы). Подходит для маленьких неизменяемых типов и иммутабельных API.
- **`func (s *OrderService)` — pointer receiver:** `s` имеет тип `*OrderService` (указатель). **Нет лишнего копирования** всей структуры; мутации полей — это изменения **одного** объекта в памяти. Обычно так объявляют методы сервисов и агрегатов, которым нужно менять состояние или держать зависимости (репозитории).
- **Символ `*` в `*OrderService`:** это **тип** «указатель на `OrderService`», не операция разыменования. Читается: «метод с получателем типа указатель на OrderService».
- **`&` при передаче аргументов:** `json.NewDecoder(r.Body).Decode(&req)` — **`&req`** передаёт **адрес** переменной `req`, чтобы декодер **заполнил** её поля. Без `&` передали бы копию, и заполнить исходную переменную было бы нельзя.
- **`&` в `errors.As(err, &valErr)`:** второй аргумент — **указатель на переменную** нужного типа (здесь `var valErr *ValidationError`). `errors.As` **присваивает** этой переменной извлечённое значение; поэтому нужен именно адрес (`&valErr`), а не «значение по значению».

```go
// Тот же FinTech-пример: valErr объявлен как *ValidationError, в As передаём &valErr
var valErr *ValidationError
if errors.As(err, &valErr) {
    // valErr теперь указывает на конкретную *ValidationError из цепочки обёрток
    _ = valErr.Field
}
```

**Уточнение к разделу про ошибки:** `errors.As(err, &target)` — **`target` — указатель на переменную**, куда запишется первое совпадение в цепочке, например `var e *MyError; errors.As(err, &e)`.

---

## 2. `defer` подробнее

- **Порядок:** при выходе из функции отложенные вызовы выполняются в порядке **LIFO** (последний `defer` — первым).
- **Когда срабатывает:** аргументы `defer` вычисляются **сразу** при объявлении `defer`, а вызов — при `return`/панике в конце функции.
- **Типичная ошибка в цикле:** `for _, x := range items { ctx, cancel := context.WithTimeout(ctx, time.Second); defer cancel() }` — накопятся лишние `cancel` и таймауты. Правильно: оборачивать итерацию в функцию `func() { ... defer cancel() }()` или вызывать `cancel()` без `defer` в конце итерации.
- **Сервер vs клиент:** для **входящего запроса** закрывайте `r.Body`, если прочитали не весь body (или для симметрии и освобождения соединения в keep-alive). Для **исходящего** HTTP-клиента после `client.Do` обязательно **`defer resp.Body.Close()`** — иначе утечка TCP-соединений из пула.

---

## 3. `sync` в процессе vs блокировки в БД (`SELECT FOR UPDATE`)

| Механизм | Где действует | Задача |
| -------- | ------------- | ------ |
| `sync.Mutex` / `sync.RWMutex` | Один **процесс** Go (все горутины этого бинарника) | Защита **общей памяти** в адресном пространстве |
| Транзакция + `SELECT ... FOR UPDATE` | **Сервер БД** | Согласованное чтение/запись **строк** между **несколькими инстансами** сервиса и параллельными HTTP-запросами |

**Зачем не путать:** mutex **не** видит другой pod/сервер и **не** заменяет транзакционную изоляцию в PostgreSQL/MySQL. Два реплики сервиса могут одновременно списать баланс, если полагаться только на `sync.Mutex` в каждом процессе.

```go
// FinTech (идея): в одной транзакции заблокировать строку счёта, прочитать баланс, списать, закоммитить.
// Псевдокод SQL — блокировка строки до конца транзакции:
// BEGIN;
// SELECT balance_cents FROM accounts WHERE id = $1 FOR UPDATE;
// UPDATE accounts SET balance_cents = balance_cents - $2 WHERE id = $1;
// COMMIT;
```

### sync.WaitGroup (расширение)

- **Как:** `Add(n)`, `Done()` (эквивалент `Add(-1)`), `Wait()` (блокировка до нуля счётчика)
- **Важно:** `Add` планировать **до** старта горутины (или заранее суммарно); вызывать `Add` изнутри новой горутины без доп. синхронизации — гонка с `Wait`

```go
// FoodTech: параллельная проверка наличия всех позиций на складе
var wg sync.WaitGroup
results := make([]bool, len(items))
for i, item := range items {
    wg.Add(1)
    go func(i int, item OrderItem) {
        defer wg.Done()
        results[i] = stockRepo.HasStock(ctx, item.ProductID, item.Qty)
    }(i, item)
}
wg.Wait()
```

**Типичная ошибка:** `go func() { wg.Add(1); defer wg.Done(); ... }()` — `Wait` может выполниться раньше `Add`. Правильно: `wg.Add(1)` перед `go`, либо один `wg.Add(N)` до цикла.

---

## 4. Часть 3: горутины и каналы (дополнения)

### 3.1 Горутины

**Планировщик и потоки ОС:** горутина — это лёгкая «задача» в рантайме Go; планировщик **GMP** мультиплексирует тысячи горутин на **несколько** потоков ОС (`M`). Одна горутина **не равна** одному системному потоку: при блокировке на I/O поток может обслужить другие горутины.

**Стоимость:** старт горутины дешевле потока ОС (килобайты стека, растущий при необходимости), но бесконтрольный `go` на CPU-bound задачах создаёт лишние переключения — нужны `GOMAXPROCS`, worker pool или ограничение параллелизма.

### 3.2 Каналы: операторы

- **`ch <- v`** — отправить значение `v` в канал `ch` (если небуферизованный — блок до получателя).
- **`v := <-ch`** или **`v, ok := <-ch`** — получить из канала; `ok == false` если канал закрыт и данных больше нет.
- **`orders <- newOrder`** — то же, что отправка: канал слева от `<-`, справа выражение (здесь в канал `orders` кладётся `newOrder`).
- **`<-done`** или **`case <-ctx.Done():`** — **чтение** из канала без сохранения в переменную: ждём события закрытия/отмены. `ctx.Done()` возвращает `<-chan struct{}`, по закрытию которого контекст отменён.

### select

- **Как:** `select { case x := <-chA: ... case chB <- v: ... }` — блокируется, пока **один** из готовых `case` не может выполниться; если несколько готовы — выбирается псевдослучайно.
- **`default`:** если ни один case не готов — выполняется `default` сразу (**неблокирующий** try-send/try-recv).
- **Таймаут:** `case <-time.After(5 * time.Second):` — отдельная горутина-таймер; для production чаще `context.WithTimeout`.
- **Отмена:** `case <-ctx.Done(): return ctx.Err()` — предпочтительнее, чем ручной `done`, когда контекст уже есть в API.

---

## 5. DDD: 5.4 Repositories (полный пример)

Ниже — **сквозной пример** FoodTech: доменный агрегат `Order`, интерфейс репозитория и PostgreSQL-реализация с реальным SQL (позиции заказа хранятся как `JSONB`).

```go
// --- internal/domain/order.go (фрагмент: модель и порт) ---
package domain

import (
    "context"
    "errors"
    "time"
)

var ErrNotFound = errors.New("not found")

type OrderStatus string

const (
    OrderStatusDraft     OrderStatus = "draft"
    OrderStatusConfirmed OrderStatus = "confirmed"
)

type OrderItem struct {
    ProductID  string `json:"product_id"`
    Name       string `json:"name"`
    Qty        int    `json:"qty"`
    PriceCents int64  `json:"price_cents"`
}

// Order — корень агрегата; конструкторы/фабрики скрывают невалидные состояния.
type Order struct {
    id         string
    customerID string
    status     OrderStatus
    items      []OrderItem
    createdAt  time.Time
}

func (o *Order) ID() string { return o.id }
func (o *Order) CustomerID() string      { return o.customerID }
func (o *Order) Status() OrderStatus    { return o.status }
func (o *Order) Items() []OrderItem     { return append([]OrderItem(nil), o.items...) }
func (o *Order) CreatedAt() time.Time   { return o.createdAt }

// RehydrateOrder — только для слоя persistence (восстановление из БД).
func RehydrateOrder(id, customerID string, items []OrderItem, status OrderStatus, createdAt time.Time) *Order {
    return &Order{
        id: id, customerID: customerID, status: status,
        items: append([]OrderItem(nil), items...), createdAt: createdAt,
    }
}

type OrderRepository interface {
    GetByID(ctx context.Context, id string) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

```go
// --- internal/infrastructure/persistence/postgres/order_repository.go ---
package postgres

import (
    "context"
    "database/sql"
    "encoding/json"
    "errors"
    "time"

    "example.com/foodtech/internal/domain"
)

type OrderRepository struct {
    db *sql.DB
}

func NewOrderRepository(db *sql.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*domain.Order, error) {
    var (
        customerID string
        status     string
        rawItems   []byte
        createdAt  time.Time
    )
    err := r.db.QueryRowContext(ctx, `
		SELECT customer_id, status, items, created_at
		FROM orders WHERE id = $1
	`, id).Scan(&customerID, &status, &rawItems, &createdAt)
    if errors.Is(err, sql.ErrNoRows) {
        return nil, domain.ErrNotFound
    }
    if err != nil {
        return nil, err
    }
    var items []domain.OrderItem
    if err := json.Unmarshal(rawItems, &items); err != nil {
        return nil, err
    }
    return domain.RehydrateOrder(id, customerID, items, domain.OrderStatus(status), createdAt.UTC()), nil
}

func (r *OrderRepository) Save(ctx context.Context, order *domain.Order) error {
    items, err := json.Marshal(order.Items()) // предполагается метод Items() []OrderItem
    if err != nil {
        return err
    }
    _, err = r.db.ExecContext(ctx, `
		INSERT INTO orders (id, customer_id, status, items, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, order.ID(), order.CustomerID(), string(order.Status()), items, order.CreatedAt())
    return err
}
```

*Замечание:* в реальном коде добавьте миграцию `CREATE TABLE orders (... items JSONB NOT NULL, ...)`, транзакции и обработку `ON CONFLICT` при upsert.

---

## 6. DDD: 5.5 Агрегаты (полный пример FoodTech)

```go
package domain

import (
    "errors"
    "time"

    "github.com/google/uuid"
)

var (
    ErrEmptyOrder   = errors.New("order has no lines")
    ErrTooManyLines = errors.New("too many order lines")
)

const maxOrderLines = 50

// OrderLine — сущность внутри агрегата; не экспортируем слайс lines наружу.
type OrderLine struct {
    ProductID  string
    Name       string
    Qty        int
    PriceCents int64
}

type Order struct {
    id         string
    customerID string
    lines      []OrderLine
    status     string // draft | confirmed
    createdAt  time.Time
}

func NewOrder(customerID string) *Order {
    return &Order{
        id:         uuid.NewString(),
        customerID: customerID,
        status:     "draft",
        createdAt:  time.Now().UTC(),
    }
}

// AddLine — единственный способ добавить позицию; проверяем инварианты здесь.
func (o *Order) AddLine(line OrderLine) error {
    if line.Qty <= 0 || line.PriceCents < 0 {
        return errors.New("invalid line")
    }
    if len(o.lines) >= maxOrderLines {
        return ErrTooManyLines
    }
    o.lines = append(o.lines, line)
    return nil
}

func (o *Order) TotalCents() int64 {
    var t int64
    for _, l := range o.lines {
        t += l.PriceCents * int64(l.Qty)
    }
    return t
}

// Confirm переводит заказ в подтверждённый статус — инвариант «не пустой заказ».
func (o *Order) Confirm() error {
    if len(o.lines) == 0 {
        return ErrEmptyOrder
    }
    o.status = "confirmed"
    return nil
}

func (o *Order) Lines() []OrderLine {
    return append([]OrderLine(nil), o.lines...)
}
```

### 6.2 Repository Pattern (кратко)

См. полный разбор и код в разделе **5.4 Repositories** в основном курсе: интерфейс в domain, реализация в `infrastructure`, загрузка/сохранение агрегата.

**Альтернативы:** Active Record, прямой SQL в handler (антипаттерн для сложного домена).

---

## 7. Паттерны 6.5–6.8

### 6.5 Factory (фабрика)

**Зачем:** скрыть выбор конкретной реализации по конфигурации (какой платёжный провайдер, какой драйвер).

```go
// FinTech: фабрика платёжного провайдера
type PaymentProcessor interface {
    Charge(ctx context.Context, amount int64, token string) error
}

func NewPaymentProcessor(cfg Config) (PaymentProcessor, error) {
    switch cfg.Provider {
    case "stripe":
        return stripe.New(cfg.StripeSecret), nil
    case "mock":
        return &mockProcessor{}, nil
    default:
        return nil, fmt.Errorf("unknown provider %q", cfg.Provider)
    }
}
```

### 6.6 Strategy (стратегия)

```go
// FoodTech: разные стратегии скидки
type DiscountStrategy interface {
    Apply(subtotalCents int64) int64
}

type NoDiscount struct{}
func (NoDiscount) Apply(c int64) int64 { return c }

type PercentOff struct{ Percent int }
func (p PercentOff) Apply(c int64) int64 { return c - (c * int64(p.Percent) / 100) }

func Checkout(subtotal int64, d DiscountStrategy) int64 {
    return d.Apply(subtotal)
}
```

### 6.7 Adapter (адаптер)

```go
// FinTech: доменный порт
type FXRates interface {
    GetRate(ctx context.Context, from, to string) (float64, error)
}

// infrastructure: обёртка над HTTP API ЦБ / провайдера
type HTTPFXAdapter struct {
    client *http.Client
    baseURL string
}

func (a *HTTPFXAdapter) GetRate(ctx context.Context, from, to string) (float64, error) {
    // GET, decode JSON, вернуть курс
    return 1.12, nil
}
```

### 6.8 Decorator / HTTP middleware

```go
// FoodTech: обёртка вокруг http.Handler
func WithRequestID(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        id := uuid.NewString()
        w.Header().Set("X-Request-ID", id)
        ctx := context.WithValue(r.Context(), ctxKey("request_id"), id)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

---

## 8. Redis Pub/Sub и real-time уведомления

**Как работает:** издатель `PUBLISH channel payload`; подписчики `SUBSCRIBE channel` получают сообщения **только пока подключены**. Сообщения **не** персистятся в Redis — если никто не слушал, сообщение потеряно (fire-and-forget).

**Зачем:** лёгкие broadcast-события (статус заказа, чат), сигнализация между сервисами без очереди.

**Связка с клиентом:** браузер не подключается к Redis напрямую. Типичная схема: **WebSocket** или **SSE** в вашем API-сервере; горутина в Go подписана на Redis Pub/Sub и пересылает события в открытые WS-сессии. Альтернатива: **RabbitMQ fanout** — если нужна персистентность и ack.

```go
// Псевдокод go-redis: подписка на канал "order:events"
pubsub := rdb.Subscribe(ctx, "order:events")
defer pubsub.Close()
ch := pubsub.Channel()
for msg := range ch {
    // msg.Channel, msg.Payload — передать подписчикам WebSocket/SSE
    _ = msg.Payload
}
```

---

## 9. Шардирование баз данных (раздел 8.6)

**Что это:** **горизонтальное партиционирование** данных по нескольким узлам БД (шардам). Каждый шард хранит **подмножество** строк (например, по диапазону `user_id` или по хэшу ключа).

**Зачем:** масштабировать запись и объём данных, когда одна машина не тянет.

**Не путать с:**

- **Репликация** — копии одних и тех же данных для чтения и отказоустойчивости; не делит данные по разным наборам.
- **Партиционирование в одной БД** — таблицы/партиции внутри одного кластера PostgreSQL (проще операционно).

**Маршрутизация в приложении (упрощённо):**

```go
import "hash/fnv"

func shardIndex(userID string, numShards int) int {
    h := fnv.New32a()
    _, _ = h.Write([]byte(userID))
    return int(h.Sum32()) % numShards
}
```

**Сложности:** кросс-шардные транзакции и JOIN дороги; ребалансировка шардов; консистентное хэширование при добавлении узлов.

---

## 10. Безопасность API (раздел 9.7)

**JWT (JSON Web Token):**

- **Access token** — короткоживущий (минуты), для API-запросов.
- **Refresh token** — длиннее, хранить **httpOnly** cookie или secure storage; обмен на новый access.
- В Go: `github.com/golang-jwt/jwt/v5` — парсинг и валидация подписи (HS256/RS256); **секрет** только из env/secret manager, не в коде.

```go
// Пример claims (без секретов в репозитории)
type Claims struct {
    UserID string `json:"sub"`
    jwt.RegisteredClaims
}
```

**Секреты:** `API_KEY`, `JWT_SECRET`, пароли БД — переменные окружения в dev; в prod — Vault, AWS Secrets Manager, Kubernetes Secrets + ограничение доступа.

**OWASP-ориентир для API (кратко):**

- Валидация и лимиты на входе (размер body, типы полей).
- **HTTPS** везде в production; HSTS на edge.
- **Rate limiting** (Redis, gateway) против brute force и DoS.
- Идемпотентность критичных операций (**FinTech**): `Idempotency-Key` для платежей и переводов.
- Заголовки безопасности (через reverse proxy или middleware): `Content-Security-Policy` где уместно.

---

## 11. Шпаргалка: ключевые команды и паттерны

Используй эти команды в CI и локально перед ревью. Подробнее см. правило [`.cursor/rules/go-cli-commands.mdc`](../.cursor/rules/go-cli-commands.mdc).

| Задача | Команда / паттерн | Когда |
| ------ | ----------------- | ----- |
| Запуск тестов | `go test ./...` | После изменений в пакетах; в CI на каждый push |
| Тесты с race | `go test -race ./...` | Код с горутинами, map, счётчиками; периодически в CI |
| Покрытие | `go test -coverprofile=coverage.out ./...` затем `go tool cover -html=coverage.out` | Поиск непокрытых веток |
| Бенчмарки | `go test -bench=. -benchmem ./...` | Оптимизация горячих путей |
| Статический анализ | `go vet ./...` | Быстрая проверка подозрительных конструкций |
| Линтеры | `staticcheck ./...`, `golangci-lint run` | Перед merge; конфиг `.golangci.yml` |
| Escape analysis | `go build -gcflags="-m"` | Понять уход значений в кучу |
| Модули | `go mod init`, `go mod tidy`, `go mod verify` | Новый проект; после правок зависимостей |
| Генерация | `go generate ./...` | После добавления `//go:generate` (mockgen, swagger, stringer) |
| Ошибка с контекстом | `fmt.Errorf("op: %w", err)` | Оборачивать ошибки для цепочки |
| Проверка ошибки | `errors.Is(err, sentinel)`, `errors.As(err, &target)` | Различение sentinel/типов |
| Context timeout | `ctx, cancel := context.WithTimeout(parent, 5*time.Second); defer cancel()` | Внешние вызовы и БД |
| Graceful shutdown | `signal.NotifyContext` + `server.Shutdown` | Долгоживущие HTTP/gRPC сервисы |

---

## 12. Правило Cursor: полный дубликат `go-cli-commands.mdc`

Исходный файл: [`.cursor/rules/go-cli-commands.mdc`](../.cursor/rules/go-cli-commands.mdc).

Ниже — тот же текст (frontmatter + содержимое), чтобы приложение было самодостаточным.

```yaml
---
description: Использовать CLI-команды Go для проверки, тестов и модулей (см. шпаргалку в GO_DEEP_COURSE.md)
alwaysApply: true
---
```

### Команды Go для операций

При изменении кода на Go **предлагай или выполняй** соответствующие команды из шпаргалки в [docs/GO_DEEP_COURSE.md](GO_DEEP_COURSE.md) (раздел «Шпаргалка: ключевые команды и паттерны»).

#### Обязательные проверки

- После правок в пакетах: `go test ./...` (или сузить путь до затронутых пакетов).
- При конкурентном коде / общих структурах: `go test -race ./...` где уместно.
- После изменения `go.mod` / импортов: `go mod tidy`.
- Не утверждать «всё собирается», без `go build ./...` или `go test` там, где это применимо.

#### По задачам

| Задача | Команда |
|--------|---------|
| Тесты | `go test ./...` |
| Гонки | `go test -race ./...` |
| Покрытие | `go test -coverprofile=coverage.out ./...` |
| Бенчмарки | `go test -bench=. -benchmem ./...` |
| Vet | `go vet ./...` |
| Линтер | `golangci-lint run` или `staticcheck ./...` |
| Модули | `go mod tidy`, `go mod verify` |
| Генерация | `go generate ./...` |

#### Ошибки и сервер

- Обёртка ошибок: `fmt.Errorf("context: %w", err)`.
- Проверка: `errors.Is` / `errors.As` (см. курс).
- Долгоживущие сервисы: graceful shutdown через `signal.NotifyContext` и `Shutdown` — не подменяй без причины.

Это правило **дополняет** DDD и best practices; фокус здесь — **операционные команды** и проверка перед выводом о готовности кода.
