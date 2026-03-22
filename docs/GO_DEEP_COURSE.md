# Глубокий курс Go: от основ до коммерческой разработки

Подробный курс для разработчиков с опытом в PHP, Python, Laravel, FastAPI, Django. Покрывает все аспекты языка, веб-фреймворки, инфраструктурные технологии и требования к коммерческой разработке (3+ года опыта).

---

## Содержание

1. [Основы языка Go](#часть-1-основы-языка-go)
2. [Стандартная библиотека](#часть-2-стандартная-библиотека)
3. [Конкурентность](#часть-3-конкурентность-в-go)
4. [HTTP-роутеры и веб-фреймворки](#часть-4-http-роутеры-и-веб-фреймворки)
5. [DDD](#часть-5-ddd-domain-driven-design)
6. [Паттерны проектирования](#часть-6-паттерны-проектирования)
7. [Юнит-тестирование](#часть-7-юнит-тестирование-в-go)
8. [Инфраструктурные технологии](#часть-8-инфраструктурные-технологии)
9. [Коммерческая разработка](#часть-9-требования-к-коммерческой-разработке) (в т.ч. [безопасность API](#97-безопасность-api-jwt-секреты-owasp-ориентир))
10. [Структура проекта](#часть-10-рекомендуемая-структура-проекта)
11. [VSCode / Cursor для Go](#часть-11-vscode--cursor-для-go)

---

## Часть 1: Основы языка Go

### 1.1 Синтаксис и идиоматика

#### Типы данных

**Примитивы — как, зачем, когда:**


| Тип                      | Размер                               | Зачем               | Когда использовать                               | Альтернативы                                              |
| ------------------------ | ------------------------------------ | ------------------- | ------------------------------------------------ | --------------------------------------------------------- |
| `int`                    | Зависит от платформы (32 или 64 бит) | Универсальное целое | Индексы, счётчики, когда размер не критичен      | `int32`/`int64` для сериализации, бинарных протоколов     |
| `int8`–`int64`           | Фиксированный                        | Явный размер        | Протоколы, БД, когда нужна предсказуемость       | `int` для локальных вычислений                            |
| `uint`                   | Как int                              | Беззнаковое         | Битовые операции, когда отрицательные невозможны | Осторожно: `uint` для вычитания может дать огромное число |
| `uint8` (byte)           | 1 байт                               | Байт, символ ASCII  | Бинарные данные, `[]byte`                        | `rune` для Unicode                                        |
| `float32`                | 4 байта                              | Экономия памяти     | Машинное обучение, большие массивы               | `float64` для точности                                    |
| `float64`                | 8 байт                               | Высокая точность    | Финансы, научные расчёты                         | Стандарт в Go для float                                   |
| `bool`                   | 1 байт                               | Логика              | Флаги, условия                                   | —                                                         |
| `string`                 | Неизменяемый                         | Текст               | Везде, где текст                                 | `[]byte` для бинарных данных                              |
| `complex64`/`complex128` | 8/16 байт                            | Комплексные числа   | DSP, FFT, научные задачи                         | Редко в веб-разработке                                    |


```go
// FinTech: баланс счёта — float64 для точности
var balance float64 = 1234.56

// FoodTech: количество порций — int
var portions int = 4

// FoodTech: цена в копейках — int64 для сериализации в БД
var priceCents int64 = 1999
```

**Структуры (struct):**

- **Как:** объявление полей, встраивание через анонимные поля, теги в обратных кавычках
- **Зачем:** группировка данных, композиция вместо наследования
- **Когда:** модели данных, конфигурация, DTO
- **Альтернативы:** `map[string]interface{}` — когда структура динамическая, но теряется типобезопасность

```go
// FoodTech: заказ с тегами для JSON
type Order struct {
    ID        string    `json:"id"`
    Items     []OrderItem `json:"items"`
    TotalCents int64     `json:"total_cents"`
}

// FinTech: встраивание — счёт расширяет базовую сущность
type Account struct {
    Entity
    Balance   float64 `json:"balance"`
    Currency  string  `json:"currency"`
}
```

**Интерфейсы:**

- **Как:** объявление набора методов, неявная реализация (не нужен `implements`)
- **Зачем:** полиморфизм, тестируемость (подмена реализаций), dependency injection
- **Когда:** везде, где нужна абстракция — репозитории, сервисы, клиенты
- **Альтернативы:** конкретные типы — когда полиморфизм не нужен
- **Идиома:** маленькие интерфейсы (1–3 метода) — `io.Reader`, `io.Writer`

```go
// FinTech: маленький интерфейс для платёжного провайдера
type PaymentProcessor interface {
    Charge(ctx context.Context, amount Money, cardToken string) error
}

// FoodTech: интерфейс репозитория заказов
type OrderRepository interface {
    GetByID(ctx context.Context, id string) (*Order, error)
    Save(ctx context.Context, order *Order) error
}
```

**Слайсы (slices):**

- **Как:** `make([]T, len, cap)`, `append()`, срезы `s[low:high]`, `copy(dst, src)`
- **Зачем:** динамические массивы, ссылка на underlying array — эффективно по памяти
- **Когда:** списки данных переменной длины
- **Важно:** слайс — это дескриптор (ptr, len, cap); два слайса могут указывать на один массив
- **Альтернативы:** массив `[n]T` — фиксированная длина; `container/list` — двусвязный список

```go
// FoodTech: позиции в заказе
items := make([]OrderItem, 0, 8)
items = append(items, OrderItem{ProductID: "pizza", Qty: 2})
items = append(items, OrderItem{ProductID: "cola", Qty: 1})

// FinTech: последние 10 транзакций
txns := allTxns[len(allTxns)-10:]
```

**Мапы (maps):**

- **Как:** `make(map[K]V)`, `v, ok := m[key]`, `delete(m, key)`
- **Зачем:** ассоциативные массивы, кэши, множества (map[T]struct{})
- **Когда:** поиск по ключу O(1), группировка
- **Важно:** мапы не thread-safe — используй `sync.RWMutex` или `sync.Map`
- **Альтернативы:** `sync.Map` — специфичные кейсы; слайс структур — когда порядок важен

```go
// FoodTech: кэш меню по ID
menuCache := make(map[string]MenuItem)
if item, ok := menuCache["pizza"]; ok {
    return item
}

// FinTech: множество активных сессий
activeSessions := make(map[string]struct{})
activeSessions[sessionID] = struct{}{}
if _, ok := activeSessions[sessionID]; ok { /* авторизован */ }
```

#### Указатели

- **Как:** `*T` — тип указателя, `&x` — взять адрес, `*p` — разыменовать
- **Зачем:** передача по ссылке (избежать копирования), мутация, optional (nil)
- **Когда:** большие структуры, необходимость мутации, nil как "отсутствие значения"
- **Отличие от C/PHP:** GC управляет памятью, нет `free`; нет арифметики указателей
- **Опасность:** разыменование nil → panic
- **Альтернативы:** передача по значению — для маленьких структур (до ~3 полей) часто быстрее из-за escape analysis

```go
// FoodTech: мутация заказа через указатель
func (o *Order) AddItem(item OrderItem) {
    o.Items = append(o.Items, item)
    o.recalculateTotal()
}

// FinTech: optional — nil = счёт не найден
func FindAccount(id string) (*Account, error) {
    acc, err := repo.GetByID(id)
    if err != nil {
        return nil, err  // возвращаем nil, error
    }
    return acc, nil
}
```

#### Получатели методов: `func (s *OrderService)` и символы `*` и `&`

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

#### Обработка ошибок

**Паттерн `if err != nil`:**

- **Как:** функция возвращает `(value, error)`, error — последний; каждый вызов проверяется
- **Зачем:** явная обработка, нет исключений — предсказуемый flow
- **Когда:** всегда при вызове функций, возвращающих error
- **Альтернативы:** `panic`/`recover` — только для фатальных ошибок (например, инициализация)

```go
// FoodTech: создание заказа с проверкой каждой ошибки
func (s *OrderService) CreateOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    menu, err := s.menuRepo.GetByID(ctx, req.MenuID)
    if err != nil {
        return nil, fmt.Errorf("get menu: %w", err)
    }
    order, err := NewOrder(req.CustomerID, menu)
    if err != nil {
        return nil, fmt.Errorf("new order: %w", err)
    }
    if err := s.orderRepo.Save(ctx, order); err != nil {
        return nil, fmt.Errorf("save order: %w", err)
    }
    return order, nil
}
```

**Кастомные ошибки:**

- **Как:** `errors.New("msg")`, `fmt.Errorf("context: %w", err)` с `%w` для обёртки
- **Зачем:** типизированные ошибки, цепочки для контекста
- **Проверка:** `errors.Is(err, target)` — сравнение по значению; `errors.As(err, &target)` — извлечение типа (**`target` — указатель на переменную**, куда запишется первое совпадение в цепочке, например `var e *MyError; errors.As(err, &e)`)
- **Когда использовать `%w`:** когда добавляешь контекст и хочешь сохранить возможность `errors.Is`/`errors.As`
- **Альтернативы:** пакет `github.com/pkg/errors` (устарел в пользу stdlib)

```go
// FinTech: sentinel и typed errors
var ErrInsufficientFunds = errors.New("insufficient funds")

type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

// Проверка
if errors.Is(err, ErrInsufficientFunds) {
    return http.StatusPaymentRequired, nil
}
var valErr *ValidationError
if errors.As(err, &valErr) {
    return http.StatusBadRequest, valErr.Message
}
```

#### Именование

- **Экспорт:** заглавная буква — публично, строчная — приватно (в рамках пакета)
- **Короткие имена:** `i`, `err`, `ctx` — для локальных переменных с малой областью видимости
- **Акронимы:** `ID`, `URL`, `HTTP` — все заглавные
- **Зачем:** читаемость, консистентность с экосистемой Go

#### Идиомы

- **Zero value:** переменные инициализируются нулём (0, "", nil) — не нужен конструктор для "пустого" состояния
- **Композиция вместо наследования:** встраивание структур, интерфейсы
- **"Accept interfaces, return structs":** функции принимают интерфейсы (гибкость), возвращают конкретные типы (ясность)

```go
// Zero value: OrderStatus = "" (пустая строка) — валидное начальное состояние
var status OrderStatus  // status == ""

// Accept interfaces, return structs
func ProcessOrder(ctx context.Context, repo OrderRepository) (*Order, error) {
    order, err := repo.GetByID(ctx, id)
    return order, err  // возвращаем *Order, не интерфейс
}

// Композиция: Order содержит []OrderItem
type Order struct {
    ID    string
    Items []OrderItem  // встраивание слайса
}
```

### 1.2 Управление памятью и производительность

**Стек vs куча:**

- **Как:** компилятор решает через escape analysis
- **Зачем:** стек — быстрее (LIFO), куча — для данных, переживающих функцию
- **Когда что:** локальные переменные, не покидающие функцию → стек; указатели, возвращаемые наружу, попадают в кучу
- **Анализ:** `go build -gcflags="-m"` — вывод escape analysis

**Профилирование (pprof):**

- **Как:** `import _ "net/http/pprof"`, эндпоинты `/debug/pprof/profile`, `/debug/pprof/heap`, `/debug/pprof/goroutine`
- **Зачем:** поиск узких мест по CPU, памяти, утечкам горутин
- **Когда:** при проблемах с производительностью, периодически в production (осторожно с CPU profile)
- **Инструменты:** `go tool pprof`, `go tool trace`

---

## Часть 2: Стандартная библиотека

### 2.1 Пакет `net/http`

#### Сервер — как, зачем, когда

**Интерфейс Handler:**

- **Как:** `ServeHTTP(ResponseWriter, *Request)` — один метод
- **Зачем:** единый контракт для всех обработчиков, композиция (middleware)
- **Когда:** любой HTTP-обработчик

```go
// FoodTech: handler для получения заказа
func (h *OrderHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    order, err := h.svc.GetByID(r.Context(), id)
    if err != nil {
        http.Error(w, err.Error(), http.StatusNotFound)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(order)
}
```

**http.HandleFunc vs http.Handle:**

- **HandleFunc** — для функций `func(w, r)` — удобно для простых handlers
- **Handle** — для объектов, реализующих Handler — когда нужна структура с состоянием
- **Когда что:** HandleFunc для лёгких handlers; Handle для middleware, обёрток

**http.ResponseWriter:**

- **WriteHeader(statusCode)** — вызвать до первого Write; иначе подставляется 200
- **Header().Set()** — заголовки до WriteHeader
- **Зачем порядок:** заголовки отправляются при первом Write или WriteHeader

**http.Request:**

- **Method, URL.Path, URL.Query(), Header, Body** — основные поля
- **Context()** — контекст запроса; отменяется при разрыве соединения
- **Зачем Context:** таймауты, отмена, передача request-scoped данных

**http.Server:**

- **ReadTimeout** — макс. время на чтение запроса (включая body)
- **WriteTimeout** — макс. время на запись ответа
- **IdleTimeout** — keep-alive соединения
- **Зачем:** защита от медленных клиентов (Slowloris), утечки соединений
- **Shutdown(ctx)** — graceful shutdown, ожидание завершения активных запросов

```go
// FinTech: сервер с таймаутами
server := &http.Server{
    Addr:         ":8080",
    Handler:      router,
    ReadTimeout:  10 * time.Second,
    WriteTimeout: 10 * time.Second,
    IdleTimeout:  120 * time.Second,
}
go server.ListenAndServe()
<-ctx.Done()
server.Shutdown(context.Background())
```

**JSON:**

- **json.NewEncoder(w).Encode(v)** — сериализация в ResponseWriter
- **json.NewDecoder(r.Body).Decode(&v)** — десериализация
- **Теги:** `json:"name"`, `json:"-"`, `json:"name,omitempty"`
- **Альтернативы:** `encoding/json` — стандарт; `jsoniter` — быстрее; `easyjson` — кодогенерация

```go
// FoodTech: DTO для API
type CreateOrderRequest struct {
    CustomerID string      `json:"customer_id"`
    Items      []OrderItem `json:"items"`
    Address    string      `json:"address,omitempty"`
}

// Сериализация ответа
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(OrderResponse{ID: order.ID, Total: order.Total})

// Десериализация запроса
var req CreateOrderRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    http.Error(w, "invalid json", http.StatusBadRequest)
    return
}
defer r.Body.Close()
```

**`defer` подробнее:**

- **Порядок:** при выходе из функции отложенные вызовы выполняются в порядке **LIFO** (последний `defer` — первым).
- **Когда срабатывает:** аргументы `defer` вычисляются **сразу** при объявлении `defer`, а вызов — при `return`/панике в конце функции.
- **Типичная ошибка в цикле:** `for _, x := range items { ctx, cancel := context.WithTimeout(ctx, time.Second); defer cancel() }` — накопятся лишние `cancel` и таймауты. Правильно: оборачивать итерацию в функцию `func() { ... defer cancel() }()` или вызывать `cancel()` без `defer` в конце итерации.
- **Сервер vs клиент:** для **входящего запроса** закрывайте `r.Body`, если прочитали не весь body (или для симметрии и освобождения соединения в keep-alive). Для **исходящего** HTTP-клиента после `client.Do` обязательно **`defer resp.Body.Close()`** — иначе утечка TCP-соединений из пула.

#### Клиент

- **http.Get/Post** — простые запросы
- **http.Client** — кастомный клиент: Timeout, Transport, CheckRedirect
- **http.NewRequest** — полный контроль: метод, заголовки, body
- **Зачем Timeout:** без него запрос может висеть бесконечно
- **Всегда:** `defer resp.Body.Close()` — иначе утечка соединений

**Go 1.22+ ServeMux:** path parameters `/users/{id}`, method-aware routing — можно обойтись без сторонних роутеров.

### 2.2 Пакет `database/sql`

**Что это:** универсальный интерфейс для SQL-БД (драйверы: PostgreSQL, MySQL, SQLite)

**Как:** `sql.Open("postgres", "connection string")` — возвращает `*sql.DB` (пул соединений)

**Основные операции:**

- `db.Query(ctx, "SELECT ...")` — несколько строк, возвращает `*sql.Rows`
- `db.QueryRow(ctx, "SELECT ...")` — одна строка, возвращает `*sql.Row`
- `db.Exec(ctx, "INSERT ...")` — без возврата строк
- `db.Prepare(ctx, "SELECT ...")` — подготовленный запрос (осторожно с пулом)
- `db.BeginTx(ctx, nil)` — транзакция

**Важно:**

- Всегда `defer rows.Close()` после Query
- `db.SetMaxOpenConns(25)`, `db.SetMaxIdleConns(5)` — настройка пула
- `QueryContext`, `ExecContext` — передача context для отмены

**Когда использовать:** любой доступ к SQL-БД

**Альтернативы:** GORM, sqlx — ORM и расширения; sqlc — кодогенерация из SQL

```go
// FinTech: получение транзакций по счёту
rows, err := db.QueryContext(ctx,
    "SELECT id, amount_cents, created_at FROM transactions WHERE account_id = $1",
    accountID,
)
if err != nil {
    return nil, err
}
defer rows.Close()

var txns []Transaction
for rows.Next() {
    var t Transaction
    if err := rows.Scan(&t.ID, &t.AmountCents, &t.CreatedAt); err != nil {
        return nil, err
    }
    txns = append(txns, t)
}
return txns, rows.Err()
```

### 2.3 Пакет `context`

**context.Background vs context.TODO:**

- **Background** — корневой контекст (main, init, тесты)
- **TODO** — заглушка, когда контекст пока не определён (избегай в production)
- **Когда:** Background — точка входа; TODO — временно при рефакторинге

**WithCancel, WithTimeout, WithDeadline:**

- **WithCancel** — ручная отмена через `cancel()`
- **WithTimeout** — автоотмена через N секунд
- **WithDeadline** — отмена в конкретный момент
- **Зачем:** распространение отмены по цепочке вызовов
- **Важно:** всегда вызывай `cancel()` (через defer), иначе утечка ресурсов

**WithValue:**

- **Как:** `context.WithValue(ctx, key, value)` — только request-scoped данные
- **Зачем:** передача user ID, request ID, trace ID без изменения сигнатур
- **Когда:** данные, специфичные для одного запроса
- **Осторожно:** не использовать для опциональных параметров; ключи — свой тип (не string), чтобы избежать коллизий

```go
// FoodTech: middleware добавляет user ID в context
type contextKey string
const userIDKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        userID := extractUserFromJWT(r)
        ctx := context.WithValue(r.Context(), userIDKey, userID)
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// В handler
userID := r.Context().Value(userIDKey).(string)
```

**Интеграция:**

- `r.Context()` в HTTP — отмена при разрыве
- `db.QueryContext(ctx, ...)` — отмена долгих запросов
- `signal.NotifyContext(ctx, os.Interrupt)` — graceful shutdown

### 2.4 Пакет `sync`

**sync.Mutex:**

- **Как:** `Lock()`, `Unlock()`, лучше с `defer mu.Unlock()`
- **Зачем:** эксклюзивный доступ к shared state
- **Когда:** любая запись в общие данные из нескольких горутин
- **Альтернативы:** каналы — когда передаётся владение; RWMutex — когда читателей много

```go
// FinTech: защита баланса при конкурентных списаниях
type Account struct {
    mu      sync.Mutex
    balance int64
}

func (a *Account) Withdraw(amount int64) error {
    a.mu.Lock()
    defer a.mu.Unlock()
    if a.balance < amount {
        return ErrInsufficientFunds
    }
    a.balance -= amount
    return nil
}
```

**sync.RWMutex:**

- **Как:** `RLock()/RUnlock()` — чтение; `Lock()/Unlock()` — запись
- **Зачем:** множественное чтение параллельно, запись эксклюзивно
- **Когда:** read-heavy workload (кэши, конфиги)
- **Осторожно:** RLock не даёт приоритет писателям — при постоянной записи может быть starvation

```go
// FoodTech: кэш меню — много читателей, редко обновляется
type MenuCache struct {
    mu    sync.RWMutex
    items map[string]MenuItem
}

func (c *MenuCache) Get(id string) (MenuItem, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    item, ok := c.items[id]
    return item, ok
}

func (c *MenuCache) Update(items map[string]MenuItem) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.items = items
}
```

#### `sync` в процессе vs блокировки в БД (`SELECT FOR UPDATE`)

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

**sync.WaitGroup:**

- **Как:** `Add(n)`, `Done()` (эквивалент `Add(-1)`), `Wait()` (блокировка до нуля счётчика)
- **Зачем:** барьер — дождаться завершения набора горутин
- **Когда:** fan-out/fan-in, параллельные подзадачи с общим «все готово»
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

**sync.Once:**

- **Как:** `once.Do(func() { ... })` — выполнится ровно один раз
- **Зачем:** singleton, lazy init
- **Когда:** инициализация при первом использовании

```go
// FinTech: ленивая инициализация платёжного клиента
var (
    paymentClient PaymentProcessor
    initOnce      sync.Once
)

func GetPaymentClient() PaymentProcessor {
    initOnce.Do(func() {
        paymentClient = stripe.NewClient(os.Getenv("STRIPE_KEY"))
    })
    return paymentClient
}
```

**sync.Map:**

- **Как:** `Store`, `Load`, `LoadOrStore`, `Delete`, `Range`
- **Зачем:** конкурентная мапа без явного Mutex
- **Когда:** ключи в основном только добавляются; много горутин; разные ключи разными горутинами
- **Когда НЕ использовать:** общий случай — `map + Mutex` часто быстрее

### 2.5 Пакет `sync/atomic`

**Атомарные операции:**

- **Как:** `AddInt64`, `LoadInt64`, `StoreInt64`, `CompareAndSwapInt64`
- **Зачем:** lock-free обновления счётчиков, флагов
- **Когда:** счётчики, флаги, простые структуры данных
- **Альтернативы:** Mutex — когда операция сложная (не одна инструкция)

**atomic.Value:**

- **Как:** `Store(x)`, `Load()` — любой тип
- **Зачем:** атомарная замена целиком (например, конфиг)
- **Когда:** hot-reload конфига, кэша
- **Осторожно:** первый Store должен быть до любого Load

**Типизированные атомики (Go 1.19+):**

- `atomic.Int64`, `atomic.Bool`, `atomic.Pointer[T]`
- **Зачем:** удобнее, типобезопаснее
- **Когда:** новые проекты на Go 1.19+

```go
// FinTech: счётчик активных запросов к платёжному API (rate limit)
var activePayments atomic.Int64

func (p *PaymentService) Process(ctx context.Context, amount Money) error {
    if activePayments.Add(1) > 100 {
        activePayments.Add(-1)
        return ErrRateLimited
    }
    defer activePayments.Add(-1)
    return p.process(ctx, amount)
}
```

---

## Часть 3: Конкурентность в Go

### 3.1 Горутины

- **Как:** `go func() { ... }()`
- **Зачем:** параллелизм, асинхронность
- **Когда:** I/O-bound задачи, параллельная обработка
- **Реализация:** M:N модель, планировщик GMP, ~2KB стек на горутину
- **Захват переменных:** `for i := range items { i := i; go func() { use(i) }() }` — копия для замыкания

**Планировщик и потоки ОС:** горутина — это лёгкая «задача» в рантайме Go; планировщик **GMP** мультиплексирует тысячи горутин на **несколько** потоков ОС (`M`). Одна горутина **не равна** одному системному потоку: при блокировке на I/O поток может обслужить другие горутины.

**Стоимость:** старт горутины дешевле потока ОС (килобайты стека, растущий при необходимости), но бесконтрольный `go` на CPU-bound задачах создаёт лишние переключения — нужны `GOMAXPROCS`, worker pool или ограничение параллелизма.

```go
// FoodTech: параллельная отправка уведомлений в кухню и курьеру
go func() {
    kitchen.NotifyNewOrder(ctx, order)
}()
go func() {
    courier.AssignDelivery(ctx, order.ID, order.Address)
}()

// Правильный захват в цикле
for i, item := range order.Items {
    i, item := i, item  // копия для замыкания
    go func() {
        processItem(ctx, item)
    }()
}
```

### 3.2 Каналы

**Операторы:**

- **`ch <- v`** — отправить значение `v` в канал `ch` (если небуферизованный — блок до получателя).
- **`v := <-ch`** или **`v, ok := <-ch`** — получить из канала; `ok == false` если канал закрыт и данных больше нет.
- **`orders <- newOrder`** — то же, что отправка: канал слева от `<-`, справа выражение (здесь в канал `orders` кладётся `newOrder`).
- **`<-done`** или **`case <-ctx.Done():`** — **чтение** из канала без сохранения в переменную: ждём события закрытия/отмены. `ctx.Done()` возвращает `<-chan struct{}`, по закрытию которого контекст отменён.

**Небуферизованный vs буферизованный:**

- **Небуферизованный** — отправка блокируется до получения (синхронизация)
- **Буферизованный** — N слотов; блокировка при переполнении/пустоте
- **Когда что:** небуферизованный — handshake, синхронизация; буферизованный — сглаживание пиков, worker pool

**Закрытие:**

- Только отправитель закрывает
- `close(ch)` — получатели получают zero value и ok=false
- **Идиома:** `for v := range ch` — выход при close

**Select:**

- **Как:** `select { case x := <-chA: ... case chB <- v: ... }` — блокируется, пока **один** из готовых `case` не может выполниться; если несколько готовы — выбирается псевдослучайно.
- **`default`:** если ни один case не готов — выполняется `default` сразу (**неблокирующий** try-send/try-recv).
- **Таймаут:** `case <-time.After(5 * time.Second):` — отдельная горутина-таймер; для production чаще `context.WithTimeout`.
- **Отмена:** `case <-ctx.Done(): return ctx.Err()` — предпочтительнее, чем ручной `done`, когда контекст уже есть в API.
- **Когда:** мультиплексирование каналов, отмена, таймауты, неблокирующая отправка через `default`.

**Паттерны:**

- **Pipeline:** цепочка горутин: `ch1 → process → ch2 → process → ch3`
- **Fan-out:** одна горутина отправляет в N воркеров (или N воркеров читают из одного канала)
- **Fan-in:** N горутин отправляют в один канал
- **Worker pool:** буферизованный канал задач; N воркеров читают и обрабатывают
- **Done channel:** `done := make(chan struct{}); close(done)` — сигнал отмены; `select { case <-done: return }`

```go
// FoodTech: worker pool — N поваров обрабатывают заказы
orders := make(chan *Order, 100)
for i := 0; i < 5; i++ {
    go func() {
        for order := range orders {
            kitchen.Cook(ctx, order)
        }
    }()
}
orders <- newOrder

// FinTech: done channel для отмены
done := make(chan struct{})
go func() {
    processPayments(ctx)
    close(done)
}()
select {
case <-done:
    return nil
case <-ctx.Done():
    return ctx.Err()
}
```

### 3.3 Синхронизация

**Правило:** "Don't communicate by sharing memory; share memory by communicating."

**Каналы vs Mutex:**

- Каналы — координация, передача данных
- Mutex — защита shared state
- **Когда что:** каналы для pipeline, worker pool; Mutex для кэша, счётчика

**Race detector:** `go run -race`, `go test -race` — обязательно в CI.

### 3.4 Практические паттерны

**Context:**

- **Как:** передавать `ctx` первым аргументом; проверять `select { case <-ctx.Done(): return ctx.Err() }`
- **Зачем:** отмена длительных операций, таймауты
- **Когда:** любой блокирующий вызов (HTTP, DB, sleep)

**errgroup:**

- **Как:** `g, ctx := errgroup.WithContext(ctx); g.Go(func() error { ... }); g.Wait()`
- **Зачем:** параллельные задачи с общей отменой при первой ошибке
- **Когда:** несколько независимых операций (параллельные HTTP-запросы)

**Semaphore:**

- **Как:** `sem := make(chan struct{}, N); sem <- struct{}{} /* acquire */; defer func() { <-sem }() /* release */`
- **Зачем:** ограничение параллелизма (не более N одновременных операций)
- **Когда:** ограничение нагрузки на внешний API, БД

```go
// FinTech: не более 10 одновременных запросов к платёжному шлюзу
var paymentSem = make(chan struct{}, 10)

func (p *PaymentGateway) Charge(ctx context.Context, amount Money) error {
    select {
    case paymentSem <- struct{}{}:
        defer func() { <-paymentSem }()
    case <-ctx.Done():
        return ctx.Err()
    }
    return p.doCharge(ctx, amount)
}
```

---

## Часть 4: HTTP-роутеры и веб-фреймворки

### 4.1 Сравнительная таблица


| Фреймворк       | Тип        | Основа     | Особенности                     | Когда использовать                 |
| --------------- | ---------- | ---------- | ------------------------------- | ---------------------------------- |
| **net/http**    | Stdlib     | —          | Минимум, Go 1.22+ path params   | Простые API, минимум зависимостей  |
| **Chi**         | Роутер     | net/http   | Минимализм, context, middleware | Микросервисы, гибкость             |
| **Gorilla Mux** | Роутер     | net/http   | Path vars, regex, matchers      | Legacy, maintenance mode           |
| **Gin**         | Фреймворк  | HttpRouter | Популярный, валидация, binding  | Быстрый старт, enterprise          |
| **Echo**        | Фреймворк  | net/http   | HTTP/2, middleware, WebSocket   | Баланс возможностей                |
| **Fiber**       | Фреймворк  | Fasthttp   | Express-подобный, быстрый       | Макс. скорость, Node.js background |
| **Beego**       | Full-stack | net/http   | MVC, ORM, админка               | Крупные монолиты                   |
| **Buffalo**     | Full-stack | —          | Генераторы, asset pipeline      | Быстрая разработка веб-приложений  |


### 4.2 net/http (стандартная библиотека)

**Как:** `http.ServeMux`, `Handle`, `HandleFunc`; Go 1.22+ — `mux.HandleFunc("GET /users/{id}", handler)`

**Зачем:** нулевые зависимости, полный контроль

**Когда:** простые REST API, микросервисы, когда важна стабильность

**Go 1.22+ ServeMux:** `mux.HandleFunc("GET /users/{id}", handler)` — `mux.HandlerFunc` для path params; `mux.HandlerFunc` для извлечения `{id}` из `r.PathValue("id")`

**Альтернативы:** любой роутер/фреймворк при росте требований

### 4.2.1 Middleware (паттерн)

**Как:** `func(next http.Handler) http.Handler { return http.HandlerFunc(func(w, r) { /* до */; next.ServeHTTP(w, r); /* после */ }) }`

**Зачем:** логирование, auth, recovery, CORS — без дублирования в каждом handler

**Когда:** любой cross-cutting concern

**Примеры:** логирование, проверка JWT, rate limiting, добавление request ID

**Альтернативы:** встроить в handler — когда middleware только для одного маршрута

```go
// FoodTech: middleware — логирование и recovery
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        defer func() {
            if err := recover(); err != nil {
                log.Printf("panic: %v", err)
                http.Error(w, "internal error", 500)
            }
        }()
        next.ServeHTTP(w, r)
        log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
    })
}

// FinTech: rate limiting по user ID
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow(r.Context(), userID) {
                http.Error(w, "rate limited", 429)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}
```

### 4.3 Chi

**Как:** `chi.NewRouter()`, `r.Get("/path", handler)`, `r.Route("/api", func(r chi.Router) { ... })`, `chi.URLParam(r, "id")`

**Зачем:** 100% совместимость с net/http, context для path params, composable middleware

**Когда:** микросервисы, API, когда нужна гибкость без "магии"

**Особенности:**

- Subrouters с middleware: `r.Route("/api", func(r chi.Router) { r.Use(middleware); r.Get("/users", handler) })`
- `r.Use()` — глобальный middleware для всех маршрутов
- `r.With(middleware).Get(...)` — middleware только для конкретного маршрута
- `chi.URLParam(r, "id")` — извлечение path параметров из context
- `chi.URLParamFromCtx(ctx, "id")` — из context напрямую
- Поддержка regex: `r.Get("/users/{userId:[0-9]+}", handler)`

**Работа:**

- Chi реализует `http.Handler` — можно вставить в любой `http.Server`
- Middleware — `func(next http.Handler) http.Handler` — стандартная сигнатура
- Context передаётся через `r.Context()` — Chi добавляет в него path params

**Альтернативы:** Gorilla Mux (похожий API), стандартный ServeMux (проще)

```go
// FinTech: Chi — API переводов
r := chi.NewRouter()
r.Route("/api/v1", func(r chi.Router) {
    r.Use(authMiddleware)
    r.Get("/accounts/{accountID}/transactions", listTransactionsHandler)
    r.Post("/accounts/{accountID}/transfer", transferHandler)
})

// Извлечение path param
accountID := chi.URLParam(r, "accountID")
```

### 4.4 Gorilla Mux

**Как:** `mux.NewRouter()`, `r.HandleFunc("/products/{key}", handler)`, `mux.Vars(r)["key"]`, `{id:[0-9]+}`

**Зачем:** path variables, regex, matchers (Methods, Host, Headers, Queries)

**Когда:** существующие проекты; для новых — Chi предпочтительнее (Gorilla в maintenance)

**Subrouter:** `r.PathPrefix("/api").Subrouter()` — группировка

### 4.5 Gin

**Как:** `gin.Default()`, `r.GET("/path", func(c *gin.Context) { c.JSON(200, data) })`, `c.Param("id")`, `c.ShouldBindJSON(&obj)`

**Зачем:** скорость, binding, валидация, middleware из коробки

**Когда:** быстрый старт, enterprise, когда нужна валидация и удобный API

**Особенности:**

- `gin.Context` — обёртка над Request/Response; хранит Keys (для передачи данных между middleware и handler)
- `c.JSON(200, data)`, `c.XML()`, `c.HTML()` — удобные методы ответа
- `c.ShouldBindJSON(&obj)`, `c.ShouldBindQuery()` — автоматический парсинг и валидация
- Валидация через теги: `binding:"required"`, `binding:"email"`
- `gin.Default()` — уже включает Logger и Recovery middleware
- `gin.New()` — без middleware, полный контроль
- Группировка: `v1 := r.Group("/v1"); v1.GET("/users", handler)`

**Работа:**

- Использует HttpRouter внутри — radix tree для быстрого маршрутизации
- Context pool — переиспользование gin.Context для снижения аллокаций
- Не совместим с net/http Handler напрямую — нужен `gin.WrapH(http.Handler)`

**Альтернативы:** Echo (похожий уровень), Chi (минимализм)

### 4.6 Echo

**Как:** `e := echo.New()`, `e.GET("/path", handler)`, `c.Param("id")`, `c.Bind(&obj)`

**Зачем:** HTTP/2, автоматический TLS (Let's Encrypt), middleware, WebSocket

**Когда:** enterprise, нужны продвинутые HTTP-возможности

**Особенности:**

- `echo.Context` — обёртка с методами `JSON`, `Bind`, `Param`, `QueryParam`
- HTTP/2 из коробки
- `e.AutoTLSManager` — автоматический TLS через Let's Encrypt
- Централизованная обработка ошибок: `e.HTTPErrorHandler = customHandler`
- Группировка: `g := e.Group("/api"); g.Use(middleware)`
- Встроенная валидация через теги
- WebSocket: `e.GET("/ws", func(c echo.Context) error { ... })`

**Работа:** использует net/http — полная совместимость

### 4.7 Fiber

**Как:** `app := fiber.New()`, `app.Get("/path", func(c *fiber.Ctx) error { return c.JSON(data) })`

**Зачем:** максимальная скорость (Fasthttp), Express-подобный API

**Когда:** высоконагруженные API, переход с Node.js

**Особенности:**

- `fiber.Ctx` — аналогично Express `req`/`res`: `c.Params()`, `c.Query()`, `c.Body()`
- `c.JSON()`, `c.Send()` — ответы
- Middleware: `app.Use()`, `app.Group("/api", middleware)`
- Встроенные: compress, logger, recover, cors, limiter
- WebSocket: `app.Get("/ws", websocket.New(handler))`
- Статика: `app.Static("/", "./public")`

**Работа:**

- Fasthttp — низкоуровневый HTTP, не совместим с net/http
- Object pooling — минимизация аллокаций
- Нельзя использовать: `net/http` middleware, `httptest` напрямую — нужны адаптеры

**Важно:** использует Fasthttp, не net/http — не все библиотеки совместимы

**Альтернативы:** Gin, Echo — если важна совместимость с net/http

```go
// FoodTech: Gin handler с binding
r.POST("/orders", func(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"error": err.Error()})
        return
    }
    order, err := orderService.Create(c.Request.Context(), req)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(201, order)
})
```

### 4.8 Beego

**Как:** MVC, роутеры по контроллерам, встроенный ORM, админка

**Зачем:** full-stack, всё из коробки

**Когда:** крупные монолиты, нужны ORM и админка

**Особенности:**

- Контроллеры: `beego.Router("/user", &UserController{})` — методы `Get`, `Post` вызываются по HTTP-методу
- ORM: встроенный, похож на Django ORM
- Админка: автоматическая генерация CRUD по моделям
- Конфиг: `beego.AppConfig.String("key")`
- Сессии, логирование, статика — из коробки

**Работа:** использует net/http

**Альтернативы:** Buffalo, ручная сборка с Chi + GORM

### 4.9 Buffalo

**Как:** генераторы, asset pipeline, hot reload

**Зачем:** быстрая разработка веб-приложений (не только API)

**Когда:** full-stack приложения, нужны генераторы

**Особенности:**

- `buffalo new app` — генерация проекта
- `buffalo generate resource` — CRUD, миграции, тесты
- Webpack/asset pipeline для фронтенда
- Hot reload при разработке
- Плагины: auth, goth (OAuth), pop (ORM)

---

## Часть 5: DDD (Domain-Driven Design)

### 5.1 Bounded Contexts

**Как:** отдельные пакеты/модули для каждого контекста (Ordering, Shipping, Billing)

**Зачем:** изоляция моделей, разная терминология в разных контекстах

**Когда:** сложная предметная область, несколько поддоменов

**Context Map — типы связей:**

- **Shared Kernel** — общие модели (осторожно с изменениями)
- **Customer-Supplier** — один контекст зависит от другого по контракту
- **Conformist** — зависимый принимает модель без изменений
- **ACL (Anti-Corruption Layer)** — адаптер для перевода внешней модели во внутреннюю

**Альтернативы:** один монолитный контекст — для простых приложений

```text
FoodTech: границы контекстов
internal/
  ordering/     # Заказы, меню, корзина
  delivery/     # Курьеры, маршруты, доставка
  billing/      # Оплата, счета, комиссии

Связь: ordering --[ACL]--> billing (заказ → платёж через адаптер)
```

### 5.2 Entities

**Как:** структура с полем ID, методы-поведение (AddItem, Cancel, Confirm)

**Зачем:** объекты с идентичностью, переживающие изменения во времени

**Когда:** User, Order, Product — то, что мы различаем по ID

**Правила:** сравнение по ID; инварианты сохраняются в методах

**Альтернативы:** Value Object — когда идентичность не важна

```go
// FoodTech: Entity Order
type Order struct {
    ID         string
    CustomerID string
    Status     OrderStatus
    Items      []OrderItem
    TotalCents int64
}

func (o *Order) AddItem(item OrderItem) error {
    if o.Status != OrderStatusDraft {
        return ErrOrderNotEditable
    }
    o.Items = append(o.Items, item)
    o.recalculateTotal()
    return nil
}

func (o *Order) Confirm() error {
    if len(o.Items) == 0 {
        return ErrEmptyOrder
    }
    o.Status = OrderStatusConfirmed
    return nil
}
```

### 5.3 Value Objects

**Как:** неэкспортируемые поля, конструктор `NewMoney(amount, currency)`, геттеры; неизменяемость (возвращать новый объект при изменении)

**Зачем:** объекты без идентичности, сравнение по значению

**Когда:** Money, Address, Email — значения, не сущности

**В Go:** структура по значению (не указатель), все поля lowercase

**Альтернативы:** примитивы — когда валидация не нужна

```go
// FinTech: Value Object Money
type Money struct {
    amountCents int64
    currency    string
}

func NewMoney(amountCents int64, currency string) (Money, error) {
    if currency == "" || len(currency) != 3 {
        return Money{}, ErrInvalidCurrency
    }
    return Money{amountCents: amountCents, currency: currency}, nil
}

func (m Money) AmountCents() int64 { return m.amountCents }
func (m Money) Currency() string   { return m.currency }

// FoodTech: Value Object Address
type Address struct {
    street, city, zip string
}

func NewAddress(street, city, zip string) (Address, error) {
    if street == "" || city == "" {
        return Address{}, ErrInvalidAddress
    }
    return Address{street: street, city: city, zip: zip}, nil
}
```

### 5.4 Repositories

**Как:** интерфейс в domain (`GetByID`, `Save`, `Delete`), реализация в infrastructure (`PostgresOrderRepository`)

**Зачем:** абстракция персистентности, тестируемость (mock репозитория)

**Когда:** доступ к агрегатам

**Правила:** репозиторий работает с агрегатом целиком; один репозиторий на агрегат

**Альтернативы:** активная запись (Active Record) — проще, но смешивает домен и персистентность

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

### 5.5 Агрегаты

**Как:** кластер сущностей с корнем (Aggregate Root); все изменения только через корень

**Зачем:** инварианты, границы транзакций

**Когда:** группа связанных сущностей с общими правилами

**Правила:** внешние ссылки только на корень; загрузка агрегата целиком

Пример **FoodTech:** заказ — корень; позиции (`OrderLine`) нельзя добавлять «снаружи» мимо правил; слайс позиций инкапсулирован.

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

---

## Часть 6: Паттерны проектирования

### 6.1 CQRS

**Как:** разделение команд (меняют состояние) и запросов (только читают)

**Зачем:** независимая оптимизация чтения/записи, масштабирование

**Когда:** read-heavy, разные модели для чтения и записи

**Commands:** императивные имена (CreateOrder, CancelOrder), возвращают error

**Queries:** возвращают данные, не меняют состояние

**Command Bus / Query Bus:** диспетчеризация по типу команды/запроса на соответствующий handler

**Event Sourcing (опционально):** хранение событий вместо состояния; replay для аудита

**Альтернативы:** CRUD — когда модель простая

```go
// FoodTech: Command
type CreateOrderCommand struct {
    CustomerID string
    Items      []OrderItem
    Address    Address
}

func (h *CreateOrderHandler) Handle(ctx context.Context, cmd CreateOrderCommand) error {
    order := NewOrder(cmd.CustomerID, cmd.Items, cmd.Address)
    return h.repo.Save(ctx, order)
}

// FoodTech: Query
type GetOrderQuery struct {
    OrderID string
}

func (h *GetOrderHandler) Handle(ctx context.Context, q GetOrderQuery) (*OrderDTO, error) {
    order, err := h.repo.GetByID(ctx, q.OrderID)
    if err != nil {
        return nil, err
    }
    return toOrderDTO(order), nil
}
```

### 6.2 Repository Pattern

См. полный разбор и код в разделе **5.4 Repositories** выше: интерфейс в domain, реализация в `infrastructure`, загрузка/сохранение агрегата.

**Альтернативы:** Active Record, прямой SQL в handler (антипаттерн для сложного домена).

### 6.3 DTO (Data Transfer Objects)

**Как:** структуры для передачи между слоями (API request/response, между сервисами)

**Зачем:** отделить внешний контракт от domain-модели; API может быть проще domain

**Когда:** HTTP API, gRPC, сообщения между сервисами

**Маппинг:** domain entity → DTO при отдаче; DTO → command при приёме

**Альтернативы:** отдавать domain entity напрямую — когда контракт совпадает

```go
// FoodTech: DTO для API (отдельно от domain)
type OrderResponse struct {
    ID         string       `json:"id"`
    Status     string       `json:"status"`
    Items      []ItemDTO    `json:"items"`
    TotalCents int64        `json:"total_cents"`
}

func toOrderDTO(o *Order) OrderResponse {
    return OrderResponse{
        ID:         o.ID,
        Status:     string(o.Status),
        Items:      toItemDTOs(o.Items),
        TotalCents: o.TotalCents,
    }
}

// Request DTO → Command
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateOrderRequest
    json.NewDecoder(r.Body).Decode(&req)
    cmd := CreateOrderCommand{CustomerID: req.CustomerID, Items: req.Items}
    h.handler.Handle(r.Context(), cmd)
}
```

### 6.4 Service Layer

**Как:** слой оркестрации — вызывает репозитории, доменные сервисы, внешние API

**Зачем:** транзакции, координация, применение use case

**Когда:** сложная бизнес-логика, несколько репозиториев

**В Go:** часто один сервис на bounded context или use case

```go
// FoodTech: Service Layer — оркестрация use case
type OrderService struct {
    orderRepo   OrderRepository
    menuRepo    MenuRepository
    paymentSvc  PaymentProcessor
    notifySvc   NotificationService
}

func (s *OrderService) CreateAndPay(ctx context.Context, req CreateOrderRequest) (*Order, error) {
    menu, _ := s.menuRepo.GetByID(ctx, req.MenuID)
    order := NewOrder(req.CustomerID, req.Items)
    if err := s.orderRepo.Save(ctx, order); err != nil {
        return nil, err
    }
    if err := s.paymentSvc.Charge(ctx, order.Total(), req.CardToken); err != nil {
        return nil, err
    }
    go s.notifySvc.OrderCreated(ctx, order.ID)
    return order, nil
}
```

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

**Зачем:** вынести взаимозаменяемые алгоритмы (скидки, налоги) за интерфейс.

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

**Зачем:** привести внешний API (HTTP, gRPC стороннего банка) к вашему доменному интерфейсу.

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

**Зачем:** добавить сквозное поведение (логирование, метрики, auth) без ломки handler-ов.

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

## Часть 7: Юнит-тестирование в Go

### 7.1 Пакет testing

**Как:** файлы `*_test.go`, функции `func TestXxx(t *testing.T)`

**t.Error vs t.Fatal:**

- **t.Error** — фиксирует ошибку, тест продолжается; накапливает ошибки
- **t.Fatal** — ошибка и немедленный выход; последующий код не выполнится
- **Когда что:** Error — когда хочешь увидеть все провалы в тесте; Fatal — когда продолжение бессмысленно (nil pointer)

**t.Run(name, fn):**

- Подтесты — каждый кейс как отдельный тест
- Вывод: `go test -v` показывает имена подтестов
- **t.Parallel()** — подтест может выполняться параллельно; осторожно с shared state

**t.Log, t.Logf:** вывод только при `-v`; не мешают при обычном запуске

### 7.2 Table-Driven Tests

**Как:** слайс структур с полями (name, input, want, wantErr), цикл `for _, tt := range tests` с `t.Run(tt.name, ...)`

**Зачем:** один тест — много кейсов; легко добавлять; читаемый вывод

**Когда:** функция с несколькими ветками, граничные случаи

**Пример структуры:**

```go
// FoodTech: table-driven тест для расчёта стоимости заказа
func TestOrder_TotalCents(t *testing.T) {
    tests := []struct {
        name    string
        items   []OrderItem
        want    int64
        wantErr error
    }{
        {"empty", nil, 0, ErrEmptyOrder},
        {"single", []OrderItem{{PriceCents: 100, Qty: 2}}, 200, nil},
        {"multiple", []OrderItem{{PriceCents: 50, Qty: 1}, {PriceCents: 100, Qty: 2}}, 250, nil},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            order := &Order{Items: tt.items}
            got, err := order.TotalCents()
            if !errors.Is(err, tt.wantErr) {
                t.Errorf("err = %v, want %v", err, tt.wantErr)
            }
            if got != tt.want {
                t.Errorf("TotalCents() = %d, want %d", got, tt.want)
            }
        })
    }
}
```

**Альтернативы:** отдельные тесты для каждого кейса — когда кейсы сильно различаются по setup

### 7.3 Тестирование HTTP (httptest)

**Как:**

- `req := httptest.NewRequest("GET", "/path", nil)` — создать запрос
- `rec := httptest.NewRecorder()` — mock ResponseWriter
- `handler.ServeHTTP(rec, req)` — вызвать handler
- Проверка: `rec.Code`, `rec.Body.String()`, `rec.Header().Get("Content-Type")`

**С роутером:** для path params нужен роутер: `router.ServeHTTP(rec, req)` — роутер заполнит context

**С Query:** `req := httptest.NewRequest("GET", "/path?key=value", nil)`

**С Body:** `body := strings.NewReader(`{"key":"value"}`); httptest.NewRequest("POST", "/path", body)`

**Когда:** тестирование handlers, middleware

```go
// FoodTech: тест HTTP handler получения заказа
func TestOrderHandler_Get(t *testing.T) {
    mockRepo := &MockOrderRepo{orders: map[string]*Order{"o1": {ID: "o1", TotalCents: 1500}}}
    handler := NewOrderHandler(NewOrderService(mockRepo))

    req := httptest.NewRequest("GET", "/orders/o1", nil)
    rec := httptest.NewRecorder()

    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Errorf("status = %d, want 200", rec.Code)
    }
    var resp OrderResponse
    json.NewDecoder(rec.Body).Decode(&resp)
    if resp.TotalCents != 1500 {
        t.Errorf("total = %d, want 1500", resp.TotalCents)
    }
}
```

### 7.4 Моки и интерфейсы

**Как:** интерфейс в коде → структура с методами, возвращающими заданные значения

**Ручной мок:** структура с полями для настроек; методы возвращают эти значения или вызывают функцию

**mockgen:** `go generate` с `//go:generate mockgen -source=repo.go -destination=mock_repo.go` — генерирует мок из интерфейса

**testify/mock:** `mock.Mock`, `mock.On("Method", args).Return(results)` — проверка вызовов

**Когда что:** ручной мок — простые случаи; mockgen — много методов; testify — когда нужна проверка вызовов

```go
// FinTech: ручной мок PaymentProcessor
type MockPaymentProcessor struct {
    ChargeFunc func(ctx context.Context, amount Money, token string) error
}

func (m *MockPaymentProcessor) Charge(ctx context.Context, amount Money, token string) error {
    if m.ChargeFunc != nil {
        return m.ChargeFunc(ctx, amount, token)
    }
    return nil
}

// В тесте
mockPay := &MockPaymentProcessor{
    ChargeFunc: func(ctx context.Context, amount Money, token string) error {
        return ErrInsufficientFunds  // симулируем отказ
    },
}
svc := NewOrderService(repo, mockPay)
```

### 7.5 Покрытие и бенчмарки

**Покрытие:**

- `go test -cover` — процент покрытия
- `go test -coverprofile=coverage.out` — детальный отчёт
- `go tool cover -html=coverage.out` — визуализация
- `-coverpkg=./...` — покрытие всех пакетов

**Бенчмарки:**

- `func BenchmarkXxx(b *testing.B)` — функция бенчмарка
- `b.N` — количество итераций (автоматически подбирается)
- `go test -bench=. -benchmem` — память на операцию
- `b.ResetTimer()` — сброс таймера после setup

**t.Parallel():** осторожно — shared state (глобальные переменные, БД) может дать flaky тесты

---

## Часть 8: Инфраструктурные технологии

### 8.1 Redis

**Что это:** in-memory key-value store, структуры данных (strings, hashes, lists, sets, sorted sets)

**Зачем в Go:** кэш, сессии, rate limiting, очереди, pub/sub

**Основная библиотека:** `github.com/redis/go-redis/v9`

**Как подключаться:**

- `redis.NewClient(&redis.Options{Addr: "localhost:6379", ...})`
- Production: PoolSize (10–100), MinIdleConns (5), DialTimeout (5s), ReadTimeout (3s), WriteTimeout (3s), MaxRetries (3)
- Cluster: `redis.NewClusterClient()` — для Redis Cluster
- Sentinel: `redis.NewFailoverClient()` — для High Availability

**Основные операции:**

- Strings: `Set(ctx, key, value, ttl)`, `Get(ctx, key)`, `SetNX(ctx, key, value, ttl)` — если не существует
- Hashes: `HSet(ctx, key, field, value)`, `HGet(ctx, key, field)`, `HGetAll(ctx, key)` — для объектов
- Lists: `LPush`, `RPush`, `LPop`, `RPop` — очереди
- Sets: `SAdd`, `SMembers`, `SIsMember` — множества
- Sorted Sets: `ZAdd`, `ZRangeByScore` — рейтинги, временные ряды
- Expiration: `Expire(ctx, key, ttl)`, `Set` с `Expiration`
- Pipeline: `pipe := rdb.Pipeline(); pipe.Set(...); pipe.Get(...); pipe.Exec(ctx)` — батч в один round-trip
- Transactions: `TxPipeline()` — MULTI/EXEC, атомарность

**Паттерны:**

- **Cache-aside:** проверить Redis → если нет, загрузить из БД → записать в Redis
- **Rate limiting:** `INCR key` + `EXPIRE key 60` (скользящее окно сложнее)
- **Session:** `SET session:{id} {data} EX 3600`
- **Distributed lock:** `SET lock:key NX EX 30` — осторожно с таймаутами

```go
// FoodTech: cache-aside — меню из Redis
func (s *MenuService) GetByID(ctx context.Context, id string) (*Menu, error) {
    cached, err := s.redis.Get(ctx, "menu:"+id).Result()
    if err == nil {
        var menu Menu
        json.Unmarshal([]byte(cached), &menu)
        return &menu, nil
    }
    if err != redis.Nil {
        return nil, err
    }
    menu, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }
    data, _ := json.Marshal(menu)
    s.redis.Set(ctx, "menu:"+id, data, 5*time.Minute)
    return menu, nil
}

// FinTech: rate limiting по user
func (l *RateLimiter) Allow(ctx context.Context, userID string) bool {
    key := "ratelimit:" + userID
    n, err := l.redis.Incr(ctx, key).Result()
    if err != nil {
        return false
    }
    if n == 1 {
        l.redis.Expire(ctx, key, time.Minute)
    }
    return n <= 100
}
```

**Обработка ошибок:**

- `redis.Nil` — ключ не существует (не всегда ошибка, для Get — нормально)
- `context.DeadlineExceeded` — таймаут
- Retry с exponential backoff — для временных сбоев

**Когда использовать:**

- Кэш (результаты запросов, тяжёлые вычисления)
- Сессии (session store)
- Rate limiting (INCR + EXPIRE)
- Очереди (LPUSH/RPOP, Streams)
- Pub/Sub (real-time уведомления)

**Альтернативы:** Memcached (проще, только strings); KeyDB (Redis-совместимый, multi-threaded)

#### Redis Pub/Sub и real-time уведомления

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

### 8.2 RabbitMQ

**Что это:** message broker, AMQP 0.9.1, очереди сообщений

**Зачем в Go:** асинхронная обработка, decoupling сервисов, гарантированная доставка

**Основная библиотека:** `github.com/rabbitmq/amqp091-go`

**Как устроено:**

- **Connection** — TCP-соединение (один на приложение)
- **Channel** — виртуальное соединение (много каналов на connection; не thread-safe — один канал на горутину)
- **Exchange** — получает сообщения, маршрутизирует в очереди (direct, fanout, topic, headers)
- **Queue** — хранит сообщения
- **Binding** — связь exchange → queue с routing key

**Producer:**

- `conn, _ := amqp.Dial("amqp://guest:guest@localhost:5672/")`
- `ch, _ := conn.Channel()`
- `ch.ExchangeDeclare("logs", "fanout", true, false, false, false, nil)`
- `ch.Publish("logs", "", false, false, amqp.Publishing{Body: []byte("msg")})`
- Confirm mode: `ch.Confirm(false)` — для гарантии доставки

**Consumer:**

- Declare queue: `ch.QueueDeclare("tasks", true, false, false, false, nil)`
- Bind: `ch.QueueBind("tasks", "routing_key", "exchange", false, nil)`
- `msgs, _ := ch.Consume("tasks", "", false, false, false, false, nil)` — false = manual ack
- `for d := range msgs { process(d); d.Ack(false) }` — Ack обязателен при manual ack

**Типы exchange:**

- **Direct** — routing key = имя очереди
- **Fanout** — broadcast во все привязанные очереди (логирование)
- **Topic** — pattern matching: `logs.*.error`, `user.#` (# — любой суффикс)
- **Headers** — маршрутизация по заголовкам

**Паттерны:**

- **Work queue** — одна очередь, несколько consumers (round-robin)
- **Pub/Sub** — fanout exchange, каждый consumer — своя очередь
- **Routing** — direct exchange, routing key
- **RPC** — reply_to, correlation_id

**Важно:** библиотека не предоставляет auto-reconnect — нужно реализовать переподключение при `conn.NotifyClose`

**Когда использовать:**

- Асинхронная обработка задач (отправка email, генерация отчётов)
- Decoupling микросервисов
- Гарантированная доставка, retry при сбоях

**Альтернативы:** Kafka (высокая пропускная способность, event log); NATS (простота, низкая задержка)

```go
// FoodTech: producer — публикация нового заказа в очередь
ch.Publish("", "orders.new", false, false, amqp.Publishing{
    ContentType: "application/json",
    Body:        []byte(`{"order_id":"o123","total":1500}`),
})

// FoodTech: consumer — кухня получает заказы
msgs, _ := ch.Consume("orders.new", "", false, false, false, false, nil)
for d := range msgs {
    var order OrderEvent
    json.Unmarshal(d.Body, &order)
    kitchen.ProcessOrder(ctx, order)
    d.Ack(false)
}
```

### 8.3 ClickHouse

**Что это:** колоночная OLAP СУБД для аналитики

**Зачем в Go:** логи, метрики, аналитика, большие объёмы данных

**Основная библиотека:** `github.com/ClickHouse/clickhouse-go/v2`

**Как подключаться:**

- Native protocol (порт 9000) — быстрее, нативный драйвер
- HTTP (порт 8123) — через `database/sql` + драйвер
- `conn, _ := clickhouse.Open(&clickhouse.Options{Addr: []string{"127.0.0.1:9000"}, Auth: clickhouse.Auth{Database: "default"}})`
- Connection pooling, failover — встроены

**Особенности:**

- Колоночное хранение — эффективные агрегации (COUNT, SUM, AVG по колонкам)
- Движки таблиц: MergeTree (основной), ReplacingMergeTree (дедупликация), SummingMergeTree (агрегация при merge)
- Партиционирование по дате: `PARTITION BY toYYYYMM(date)`
- Bulk insert: `batch, _ := conn.PrepareBatch(ctx, "INSERT INTO events"); batch.Append(...); batch.Send()`
- Материализованные представления — автоматическая агрегация при вставке
- TTL — автоматическое удаление старых данных

**Паттерны:**

- **Логи:** `INSERT INTO logs (timestamp, level, message) VALUES (?, ?, ?)` — batch по 1000–10000 строк
- **События:** `INSERT INTO events` — append-only
- **Агрегации:** материализованные представления для pre-aggregated данных
- **Запросы:** `SELECT toDate(timestamp) as day, count() FROM events GROUP BY day`

**Когда использовать:**

- Логи приложений
- Метрики, события
- Аналитические отчёты, дашборды
- Большие объёмы append-only данных

**Когда НЕ использовать:** OLTP, транзакции, частые UPDATE/DELETE — ClickHouse оптимизирован для INSERT и SELECT

**Альтернативы:** PostgreSQL (если нужны транзакции); TimescaleDB (временные ряды); Apache Druid

```go
// FoodTech: bulk insert событий заказов в ClickHouse
batch, _ := conn.PrepareBatch(ctx, "INSERT INTO order_events")
for _, e := range events {
    batch.Append(e.OrderID, e.EventType, e.Timestamp, e.Payload)
}
batch.Send()

// FinTech: аналитический запрос — объём платежей по дням
rows, _ := conn.Query(ctx, `
    SELECT toDate(created_at) as day, sum(amount_cents)
    FROM transactions
    WHERE account_id = $1 AND created_at >= $2
    GROUP BY day
`, accountID, since)
```

### 8.4 Prometheus

**Что это:** система мониторинга, сбор метрик, pull-модель

**Зачем в Go:** метрики приложения (счётчики, гистограммы), алертинг

**Основная библиотека:** `github.com/prometheus/client_golang`

**Как экспонировать метрики:**

- `http.Handle("/metrics", promhttp.Handler())` — эндпоинт для Prometheus
- Регистрация: `prometheus.MustRegister(counter)` или `promauto.NewCounter(...)` — автоматическая регистрация
- Labels: `prometheus.NewCounterVec(prometheus.CounterOpts{Name: "http_requests_total"}, []string{"method", "path"})`

**Типы метрик:**

- **Counter** — монотонно растёт: `counter.Inc()`, `counter.Add(5)` — запросы, ошибки
- **Gauge** — текущее значение: `gauge.Set(100)`, `gauge.Inc()` — активные соединения, размер очереди
- **Histogram** — распределение: `histogram.Observe(0.5)` — латентность; создаёт _bucket, _sum, _count
- **Summary** — квантили на стороне приложения; для latency — чаще Histogram (агрегация при scrape)

**Паттерны:**

- **HTTP middleware:** `promhttp.InstrumentHandlerCounter(counter, handler)`, `InstrumentHandlerDuration`
- **Бизнес-метрики:** `ordersTotal.Inc()` при создании заказа
- **Размер очереди:** `queueSize.Set(float64(len(queue)))` — Gauge

**Pull vs Push:** Prometheus сам тянет метрики с `/metrics` — не нужен Push в приложении; для short-lived jobs — Pushgateway

```go
// FoodTech: метрики заказов
var (
    ordersTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "orders_total"},
        []string{"status"},
    )
    orderLatency = prometheus.NewHistogram(
        prometheus.HistogramOpts{Name: "order_creation_seconds", Buckets: []float64{.1, .5, 1}},
    )
)

func init() {
    prometheus.MustRegister(ordersTotal, orderLatency)
}

func (s *OrderService) Create(ctx context.Context, req CreateOrderRequest) error {
    start := time.Now()
    defer func() { orderLatency.Observe(time.Since(start).Seconds()) }()
    // ...
    ordersTotal.WithLabelValues("created").Inc()
    return nil
}
```

**Когда использовать:**

- Метрики HTTP (количество запросов, латентность)
- Бизнес-метрики (заказы, конверсии)
- Системные метрики (goroutines, память) — `go prometheus.NewGoCollector()` уже есть

**Альтернативы:** OpenTelemetry (универсальная телеметрия); StatsD (push-модель)

### 8.5 Kafka

**Что это:** распределённый event streaming, log-based storage

**Зачем в Go:** event-driven архитектура, высокая пропускная способность, replay событий

**Основная библиотека:** `github.com/IBM/sarama`

**Как устроено:**

- **Topic** — поток сообщений
- **Partition** — sharding топика; порядок гарантирован в рамках партиции
- **Producer** — отправка в топик (с указанием партиции или ключа)
- **Consumer Group** — группа потребителей; каждая партиция — одному потребителю в группе
- **Offset** — позиция в партиции; consumer group хранит offset в `__consumer_offsets`

**Producer:**

- **SyncProducer** — `SendMessage()` блокируется до ack; надёжнее
- **AsyncProducer** — `Input()` не блокируется; `Successes()`/`Errors()` каналы; выше throughput
- Config: `Producer.RequiredAcks = WaitForAll` — все реплики; `Producer.Retry.Max = 5`
- Key: одинаковый key → одна партиция (гарантия порядка)
- Idempotence: `Producer.Idempotent = true` — exactly-once семантика

**Consumer:**

- **Consumer** — низкоуровневый: `consumer.ConsumePartition(topic, partition, offset)`
- **ConsumerGroup** — автоматическое распределение партиций, rebalance: `consumerGroup.Consume(ctx, topics, handler)`
- Handler: `Setup(session)` — при назначении партиций; `ConsumeClaim(session, claim)` — обработка сообщений; `Cleanup(session)` — при уходе
- Offset: `session.MarkMessage(msg, "")` — commit offset
- `Consumer.Offsets.Initial = OffsetNewest` или `OffsetOldest`

**Паттерны:**

- **Event sourcing:** все события в топик; состояние — replay
- **CQRS:** write топик для команд; read model — consumer
- **Outbox:** запись в БД + в outbox таблицу; отдельный процесс публикует в Kafka

```go
// FinTech: producer — событие списания
msg := &sarama.ProducerMessage{
    Topic: "transactions",
    Key:   sarama.StringEncoder(accountID),
    Value: sarama.ByteEncoder(mustMarshal(TransactionEvent{
        AccountID: accountID,
        Amount:    -1000,
        Type:      "debit",
    })),
}
producer.SendMessage(msg)

// FoodTech: consumer group — обработка событий заказов
handler := &OrderEventHandler{svc: orderService}
consumerGroup.Consume(ctx, []string{"order.events"}, handler)
```

**Когда использовать:**

- Event sourcing, event-driven архитектура
- Высокая пропускная способность (миллионы сообщений/сек)
- Replay событий, аудит
- Связь микросервисов через события

**Альтернативы:** RabbitMQ (проще, меньше throughput); NATS JetStream; Apache Pulsar

### 8.6 Шардирование баз данных

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

## Часть 9: Требования к коммерческой разработке (3+ года)

### 9.1 Типичные требования

Опыт 5–6+ лет backend, 3+ года на Go; REST API, микросервисы, PostgreSQL, Redis; Docker, Kubernetes, CI/CD; unit и integration тесты; Agile, code review.

### 9.2 Навыки

**Обязательно:** полный цикл разработки, database/sql, миграции, structured logging, конфигурация через env, graceful shutdown.

**Желательно:** gRPC, Kafka/RabbitMQ, Prometheus/OpenTelemetry, JWT, rate limiting.

### 9.3 Graceful Shutdown

**Как:**

1. `ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)`
2. Запуск сервера в горутине
3. `<-ctx.Done()` — ожидание сигнала
4. `server.Shutdown(shutdownCtx)` — с таймаутом (например, 10 секунд)

**Зачем:** завершение активных запросов, освобождение ресурсов

**Когда:** любой долгоживущий сервис

### 9.4 Конфигурация (12-factor)

**Как:** переменные окружения (`os.Getenv`), конфиг-файлы (YAML/JSON), флаги

**Зачем:** разные настройки для dev/staging/prod без изменения кода

**Библиотеки:** `github.com/kelseyhightower/envconfig`, `github.com/spf13/viper`

### 9.5 Structured Logging

**Библиотеки:** `github.com/rs/zerolog`, `go.uber.org/zap`

**Зачем:** структурированные логи (JSON), уровни, контекст (request ID, user ID)

**zerolog:** `log.Info().Str("user", id).Msg("request")` — fluent API, zero alloc

**zap:** `zap.L().Info("message", zap.String("key", "value"))` — быстрый, sugared/unsugared

**Когда:** production-приложения

### 9.6 Проекты для портфолио

REST API с DDD + CQRS; микросервис с gRPC, Kafka, PostgreSQL; worker с конкурентной обработкой; CLI с cobra.

### 9.7 Безопасность API (JWT, секреты, OWASP-ориентир)

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

## Часть 10: Рекомендуемая структура проекта

```
cmd/
  api/          # main.go для HTTP API
  worker/       # main.go для воркеров
internal/
  domain/       # entities, value objects, repository interfaces
  application/  # commands, queries, services (use cases)
  infrastructure/
    http/       # handlers, middleware, routing
    persistence/ # repository implementations (Postgres, Redis)
pkg/            # переиспользуемый код (если нужен внешним пакетам)
```

**Правила:**

- `internal/` — недоступен извне модуля (компилятор Go)
- `cmd/` — точка входа; минимальная логика
- domain не зависит от infrastructure
- application зависит от domain (интерфейсы)

---

## Часть 11: VSCode / Cursor для Go

Рекомендуемые расширения для комфортной разработки на Go в VSCode и Cursor.

### Обязательные


| Расширение | ID          | Зачем                                                                                                                                                                                                                                       |
| ---------- | ----------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Go**     | `golang.Go` | Официальное расширение от Go Team (Google). IntelliSense, навигация, форматирование, тесты, отладка (Delve). Использует gopls. Требует Go 1.21+, VSCode 1.90+. [Marketplace](https://marketplace.visualstudio.com/items?itemName=golang.Go) |


**Установка инструментов:** после установки Go extension выполни `Go: Install/Update Tools` (Ctrl+Shift+P) — установит gopls, staticcheck, gotests и др.

### Рекомендуемые


| Расширение           | ID                           | Зачем                                                                                                         |
| -------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------- |
| **Error Lens**       | `usernamehw.errorlens`       | Показывает ошибки и предупреждения прямо в коде (inline). Работает с диагностикой gopls.                      |
| **Go Test Explorer** | `premparihar.gotestexplorer` | Панель для запуска тестов по пакетам и функциям (опционально — в Go extension есть встроенный запуск тестов). |


**Примечание:** Go extension включает Delve для отладки — дополнительный отладчик не обязателен.

### Настройка golangci-lint

В `settings.json`:

```json
"go.lintTool": "golangci-lint",
"go.lintFlags": ["--fast"]
```

Флаг `--fast` обязателен — без него редактор может подвисать. Конфиг `.golangci.yml` в корне проекта.

### Настройка gopls (опционально)

В `settings.json`:

```json
"gopls": {
  "ui.semanticTokens": true,
  "analyses": {
    "unusedparams": true,
    "shadow": true
  }
}
```

### Полезные команды (Ctrl+Shift+P)

- `Go: Install/Update Tools` — установка/обновление gopls, staticcheck, gotests
- `Go: Add Import` — добавление импорта
- `Go: Generate Unit Tests` — генерация тестов
- `Go: Run test at cursor` — запуск теста под курсором

### Ссылки

- [Go extension (marketplace)](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [gopls](https://pkg.go.dev/golang.org/x/tools/gopls)
- [golangci-lint](https://golangci-lint.run/)
- [Error Lens](https://marketplace.visualstudio.com/items?itemName=usernamehw.errorlens)

---

## Рекомендуемый порядок изучения

1. Недели 1–2: основы
2. Недели 3–4: stdlib, конкурентность
3. Неделя 5: роутер/фреймворк
4. Неделя 6: тестирование
5. Недели 7–8: DDD
6. Недели 9–10: CQRS, паттерны
7. Недели 11–12: Redis, RabbitMQ/Kafka, Prometheus, полный проект

---

## Шпаргалка: ключевые команды и паттерны

Используй эти команды в CI и локально перед ревью. Подробнее см. правило `.cursor/rules/go-cli-commands.mdc`.

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

## Полезные ресурсы

- [Go Tour](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
- [Three Dots Labs: DDD, CQRS](https://threedots.tech/post/ddd-cqrs-clean-architecture-combined/)
- [Standard library](https://pkg.go.dev/std)
- [Go Memory Model](https://go.dev/ref/mem)

