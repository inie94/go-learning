# План изучения Go

> Данный проект разработан для глубокого погружения в Go


## 🗺️ Дорожная карта (Roadmap)

| Этап | Название | Уровень | Статус |
|:----:|:---------|:-------:|:------:|
| 1 | Основы языка | Junior | ⬜️ |
| 2 | Инструментарий и конкурентность | Middle | ⬜️ |
| 3 | Внутреннее устройство и архитектура | Middle+ / Senior | ⬜️ |
| 4 | Продакшен-эксплуатация и навыки Senior | Senior | ⬜️ |

---

## Этап 1: Основы языка (Junior)

Цель — научиться писать идиоматичный, работающий код.

### Базовый синтаксис и типы
- Структура программы, пакеты, импорты
- Переменные, константы, базовые типы (`int`, `float`, `string`, `bool`)
- Zero values (нулевые значения по умолчанию)
- Указатели: разница между value и pointer semantics

### Составные типы (Composite Types)
- Массивы (Arrays) и Срезы (Slices): внутреннее устройство (`len`, `cap`, underlying array)
- Карты (Maps): особенности работы, map header, коллизии
- Структуры (Structs): теги (tags), публичные/приватные поля, embedding (композиция вместо наследования)

### Функции
- Сигнатуры, вариативные параметры (`...`)
- Множественный возврат значений, именованные возвращаемые параметры
- Функции как значения первого класса (first-class citizens), замыкания
- `defer`: порядок выполнения (LIFO), аргументы вычисляются сразу, а не при вызове

### Интерфейсы (Критически важная тема)
- Неявная имплементация (structural typing)
- Пустой интерфейс `interface{}` / `any`
- Interface values под капотом: указатель на тип и указатель на данные (`itab` / dynamic type)
- Type assertions и type switches

### Обработка ошибок (Error Handling)
- Философия Go: ошибки — это значения
- Сентинел-ошибки (Sentinel Errors)
- Создание кастомных ошибок: `errors.New`, `fmt.Errorf` с `%w`
- Wrapping, Unwrapping, `errors.Is` и `errors.As`

---

## Этап 2: Инструментарий и продвинутый синтаксис (Middle)

Изучаем экосистему и начинаем писать безопасный конкурентный код.

### Конкурентность (Concurrency)
- **Горутины (Goroutines):** легковесные потоки, модель многопоточности
- **Каналы (Channels):** буферизированные и небуферизированные, закрытие, nil-каналы
- **Паттерн `select`:** мультиплексирование, таймауты, неблокирующие операции, приоритетное чтение
- **Примитивы синхронизации (`sync`):**
  - `Mutex`, `RWMutex`, `WaitGroup`, `Once`, `Cond`
  - Пакет `sync/atomic` для lock-free операций
- **Пакет `context`:** отмена операций, дедлайны, передача сквозных значений

### Среда разработки и тулинг
- Go Modules (`go mod tidy`, `vendor`, версионирование, `go.sum`)
- Линтеры и форматирование: `go fmt`, `goimports`, `golangci-lint`
- Компиляция и кросс-компиляция (`GOOS`, `GOARCH`)
- Race Detector (`go run -race`)

### Работа с ОС и вводом-выводом
- Чтение/запись `os.Stdin/Stdout`, работа с ФС (`os`, `io`, `io/fs`)
- Интерфейсы `io.Reader` и `io.Writer` — основа всего ввода-вывода в Go

### Тестирование
- Unit-тесты (пакет `testing`)
- Table-driven tests (идиоматичный подход)
- Моки: ручные через интерфейсы, `gomock`, `mockery`
- Фаззинг (`go test -fuzz`)

---

## Этап 3: Внутреннее устройство и архитектура (Middle+ / Senior)

Здесь начинается понимание, «почему оно так быстро работает и где ломается».

### Рантайм Go
- **Планировщик (G-M-P Model):** Горутины (G), Машины/Потоки ОС (M), Процессоры (P)
- **Сборка мусора (GC):** Трехцветный алгоритм (Tri-color), барьеры записи (Write barriers), STW-паузы (Stop The World)
- **Escape Analysis:** где выделяется память — на стеке или в куче. Анализ через `go build -gcflags="-m"`
- Стек горутины: как он растет и копируется (continuous stack)

### Профилирование и оптимизация
- **Benchmarking:** `go test -bench`
- **Профилировщики (pprof):** CPU, Memory (Heap), Block, Mutex. Чтение flame graphs
- **Трейсинг (Execution Tracer):** `runtime/trace`

### Архитектурные паттерны
- Микросервисы и монолит
- Domain-Driven Design (DDD) в Go: Transport → Service → Repository
- Чистая архитектура / Гексагональная (Ports & Adapters)
- Паттерны конкурентности: Pipeline, Fan-in / Fan-out, Worker Pool

### Продвинутые возможности
- **Дженерики (Generics):** обобщённые функции и типы, type constraints
- **Рефлексия (Reflection):** использовать осознанно, с пониманием цены производительности

---

## Этап 4: Производственная эксплуатация и навыки Senior

То, что отличает просто сильного разработчика от Senior.

### Сетевое программирование и API
- Протоколы: HTTP/1.1, HTTP/2, gRPC, WebSocket
- Настройка HTTP-сервера: таймауты (Read, Write, Idle, `max_header_bytes`)
- Управление соединениями, graceful shutdown
- REST vs RPC, Swagger / OpenAPI

### Базы данных и хранилища
- SQL нативно: `database/sql` + драйвер (pgx)
- Миграции (golang-migrate, goose)
- Очереди сообщений (Kafka, NATS, RabbitMQ)
- Key-Value хранилища (Redis, etcd)

### Безопасность
- OWASP Top 10 применительно к Go
- Race conditions, deadlocks, утечки горутин (goroutine leaks)
- Валидация входных данных, предотвращение инъекций
- Управление секретами

### CI/CD и DevOps
- Написание Dockerfile (мультистейдж-билды): минимальный образ (Distroless/Scratch)
- Пайплайны деплоя (GitLab CI, GitHub Actions)
- **Observability (Наблюдаемость):**
  - Логирование: structured logging (`slog`, `zerolog`)
  - Трейсинг: OpenTelemetry, Jaeger
  - Метрики: Prometheus, Grafana

### Мышление Senior-а
- **Менторство:** умение проводить Code Review с фокусом на архитектуру и безопасность, а не на нейминг
- **Принятие решений (Trade-offs):** выбор между перформансом и читаемостью, монолитом и микросервисами
- **Стандарты:** умение писать RFC / ADR (Architecture Decision Records)

---

## Рекомендуемые ресурсы для старта

### Интерактивные учебники и документация
- **A Tour of Go** — официальный интерактивный тур по языку: [https://go.dev/tour/](https://go.dev/tour/)
- **Effective Go** — обязательный к прочтению свод лучших практик: [https://go.dev/doc/effective_go](https://go.dev/doc/effective_go)
- **Go by Example** — изучение концепций через примеры кода: [https://gobyexample.com/](https://gobyexample.com/)
- **Спецификация языка Go** — первоисточник по устройству языка: [https://go.dev/ref/spec](https://go.dev/ref/spec)

### Книги (обязательны к изучению)
- **«Go 100 Mistakes and How to Avoid Them»** (Teiva Harsanyi) — настольная книга, чтобы не наступать на чужие грабли: [https://100go.co/](https://100go.co/)
- **«Concurrency in Go»** (Katherine Cox-Buday) — глубокое погружение в модель конкурентности: [O'Reilly](https://www.oreilly.com/library/view/concurrency-in-go/9781491941294/)
- **«Let's Go» и «Let's Go Further»** (Alex Edwards) — создание веб-приложений с нуля до продакшена: [https://lets-go.alexedwards.net/](https://lets-go.alexedwards.net/)
- **«The Go Programming Language»** (Alan Donovan, Brian Kernighan) — классический учебник от авторов языка: [https://www.gopl.io/](https://www.gopl.io/)

### Продвинутые курсы и блоги
- **Ardan Labs Ultimate Go** — продвинутое обучение внутреннему устройству Go (от Билла Кеннеди): [https://www.ardanlabs.com/training/ultimate-go/](https://www.ardanlabs.com/training/ultimate-go/)
- **Gophercises** — практические упражнения для отработки навыков: [https://gophercises.com/](https://gophercises.com/)
- **JustForFunc** — разборы идиоматичного кода и продвинутых тем на YouTube: [https://www.youtube.com/c/JustForFunc](https://www.youtube.com/c/JustForFunc)

### Исходный код для изучения
- **Стандартная библиотека Go:** [https://cs.opensource.google/go/go](https://cs.opensource.google/go/go)
  - Приоритетные пакеты для чтения: `net/http`, `sync`, `context`, `io`, `fmt`, `encoding/json`
- **Awesome Go** — курируемый список библиотек и инструментов экосистемы: [https://github.com/avelino/awesome-go](https://github.com/avelino/awesome-go)

### Профилирование и внутреннее устройство
- **Официальный блог Go** — статьи о рантайме, GC и планировщике: [https://go.dev/blog/](https://go.dev/blog/)
- **«Go Memory Management»** (A Journey in Go) — визуальное объяснение аллокатора и GC: [https://medium.com/a-journey-with-go](https://medium.com/a-journey-with-go)
- **Pprof в деталях** — руководство от создателей Go: [https://go.dev/blog/pprof](https://go.dev/blog/pprof)

### Сетевое взаимодействие и gRPC
- **gRPC Go Quickstart** — официальное руководство: [https://grpc.io/docs/languages/go/quickstart/](https://grpc.io/docs/languages/go/quickstart/)
- **Protocol Buffers (protobuf)** — язык описания данных для gRPC: [https://protobuf.dev/](https://protobuf.dev/)