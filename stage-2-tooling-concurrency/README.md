# Этап 2: Инструментарий и продвинутый синтаксис (Middle)

> **Цель этапа** — освоить экосистему Go, научиться писать безопасный конкурентный код и использовать инструменты разработчика на профессиональном уровне.

---

## 📚 Содержание

- [Конкурентность (Concurrency)](#конкурентность-concurrency)
- [Среда разработки и тулинг](#среда-разработки-и-тулинг)
- [Работа с ОС и вводом-выводом](#работа-с-ос-и-вводом-выводом)
- [Тестирование](#тестирование)
- [Практические задания](#практические-задания)
- [Контрольные вопросы](#контрольные-вопросы)
- [Рекомендуемые ресурсы](#рекомендуемые-ресурсы)

---

## Конкурентность (Concurrency)

### Горутины (Goroutines)

Горутины — это легковесные потоки, управляемые рантаймом Go. Они дешевле потоков ОС и могут масштабироваться до миллионов.

```go
func main() {
    // Запуск горутины
    go func() {
        fmt.Println("Hello from goroutine")
    }()
    
    // Основная горутина продолжает работу
    time.Sleep(time.Millisecond) // даём время на выполнение
    fmt.Println("Main done")
}
```
> ⚠️ Важно: горутина завершается, когда завершается основная программа.
### Каналы (Channels)

Каналы — это основной способ коммуникации между горутинами. Они обеспечивают синхронизацию и передачу данных.

```go
// Небуферизированный канал (блокирующий)
ch := make(chan int)

// Буферизированный канал
buffered := make(chan string, 3) // размер буфера 3

// Отправка и получение
ch <- 42      // отправка
value := <-ch // получение

// Закрытие канала
close(ch)

// Проверка закрытия
value, ok := <-ch
if !ok {
    fmt.Println("Channel closed")
}
```

#### Пример использования:

```go
func worker(id int, jobs <-chan int, results chan<- int) {
    for job := range jobs { // цикл до закрытия канала
        results <- job * 2
    }
}

func main() {
    jobs := make(chan int, 5)
    results := make(chan int, 5)
    
    // Запуск воркеров
    for w := 1; w <= 3; w++ {
        go worker(w, jobs, results)
    }
    
    // Отправка задач
    for j := 1; j <= 5; j++ {
        jobs <- j
    }
    close(jobs)
    
    // Получение результатов
    for r := 1; r <= 5; r++ {
        <-results
    }
}
```
### Паттерн `select`

`select` позволяет мультиплексировать операции с каналами.

```go
func main() {
    ch1 := make(chan string)
    ch2 := make(chan string)
    
    go func() {
        time.Sleep(1 * time.Second)
        ch1 <- "from ch1"
    }()
    
    go func() {
        time.Sleep(2 * time.Second)
        ch2 <- "from ch2"
    }()
    
    select {
    case msg1 := <-ch1:
        fmt.Println(msg1)
    case msg2 := <-ch2:
        fmt.Println(msg2)
    case <-time.After(3 * time.Second):
        fmt.Println("timeout")
    default:
        fmt.Println("no messages ready")
    }
}
```
#### Особенности:

* Неблокирующие операции через `default`
* Таймауты через `time.After`
* Приоритетное чтение: если несколько каналов готовы, выбирается случайный

### Примитивы синхронизации (`sync`)

#### `Mutex` и `RWMutex`

`sync.Mutex` -  это примитив синхронизации, который гарантирует, что только одна горутина одновременно может выполнять критическую секцию кода.  
`sync.RWMutex` - это примитив синхронизации, который разделяет доступ на чтение и запись:
* Множество читателей могут захватывать блокировку одновременно.
* Только один писатель может захватить блокировку.
* Писатель ждёт, пока все читатели освободят блокировку.

```go
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Inc() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

func (c *Counter) Value() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}

// RWMutex — для read-heavy сценариев
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]interface{}
}

func (s *SafeMap) Get(key string) interface{} {
    s.mu.RLock()
    defer s.mu.RUnlock()
    return s.m[key]
}
```

#### `WaitGroup`

`sync.WaitGroup` — это примитив синхронизации, который позволяет дождаться завершения группы горутин.

```go
func main() {
    var wg sync.WaitGroup
    
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go func(id int) {
            defer wg.Done()
            fmt.Printf("Worker %d done\n", id)
        }(i)
    }
    
    wg.Wait() // блокируется до завершения всех горутин
    fmt.Println("All workers finished")
}
```

#### `Once`

`sync.Once` — это примитив синхронизации, который гарантирует, что переданная функция будет выполнена ровно один раз, даже если Do() вызывается из множества горутин одновременно

```go
var once sync.Once
var config *Config

func GetConfig() *Config {
    once.Do(func() {
        config = loadConfig()
    })
    return config
}
```

#### `Cond`

`sync.Cond` — это примитив синхронизации, который реализует условную переменную (condition variable). Он позволяет горутинам ожидать наступления определённого условия и пробуждаться при его изменении

```go
type Queue struct {
    mu    sync.Mutex
    cond  *sync.Cond
    items []int
}

func NewQueue() *Queue {
    q := &Queue{items: make([]int, 0)}
    q.cond = sync.NewCond(&q.mu)
    return q
}

func (q *Queue) Push(item int) {
    q.mu.Lock()
    defer q.mu.Unlock()
    q.items = append(q.items, item)
    q.cond.Signal() // пробуждаем одну горутину
}

func (q *Queue) Pop() int {
    q.mu.Lock()
    defer q.mu.Unlock()
    for len(q.items) == 0 {
        q.cond.Wait() // ждём сигнала
    }
    item := q.items[0]
    q.items = q.items[1:]
    return item
}
```

### Пакет `sync/atomic` — lock-free операции

`sync/atomic` — это пакет стандартной библиотеки Go, предоставляющий атомарные операции над памятью. Эти операции гарантируют, что чтение и запись переменной выполняются как единое неделимое действие, даже при конкурентном доступе из нескольких горутин.

```go
import "sync/atomic"

var counter int64

func increment() {
    atomic.AddInt64(&counter, 1)
}

func get() int64 {
    return atomic.LoadInt64(&counter)
}

// Compare and Swap (CAS)
func casExample() {
    var value int64 = 0
    atomic.CompareAndSwapInt64(&value, 0, 42)
}
```

### Пакет `context` — управление жизненным циклом

```go
func main() {
    // Создание контекста с отменой
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // С таймаутом
    ctxTimeout, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    // С дедлайном
    ctxDeadline, cancel := context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
    defer cancel()
    
    // Передача значений
    ctxWithVal := context.WithValue(context.Background(), "userID", 123)
    userID := ctxWithVal.Value("userID").(int)
    
    // Использование в горутинах
    go worker(ctx)
}

func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            fmt.Println("Cancelled:", ctx.Err())
            return
        default:
            // работа
            time.Sleep(100 * time.Millisecond)
        }
    }
}
```
## Среда разработки и тулинг

### Go Modules

```bash
# Инициализация модуля
go mod init github.com/user/project

# Обновление зависимостей
go mod tidy           # добавляет недостающие, удаляет неиспользуемые
go mod verify         # проверяет целостность зависимостей

# Вендорирование (для оффлайн-сборки)
go mod vendor

# Обновление зависимостей
go get -u ./...       # обновить все пакеты
go get github.com/pkg/errors@v0.9.1  # конкретная версия

# Файлы:
# go.mod — описание модуля и зависимостей
# go.sum — контрольные суммы для верификации
```

### Форматирование и линтинг

``` bash
# Форматирование кода
go fmt ./...

# Импорты и форматирование
goimports -w .

# Статический анализ
go vet ./...

# Комплексный линтер (установите один раз)
golangci-lint run
```

### Компиляция и кросс-компиляция

```bash
# Обычная компиляция
go build -o app main.go

# Кросс-компиляция (сборка для других ОС/архитектур)
GOOS=linux GOARCH=amd64 go build -o app-linux-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o app.exe main.go
GOOS=darwin GOARCH=arm64 go build -o app-macos-arm64 main.go

# Доступные платформы
go tool dist list

# Оптимизация размера бинарника
go build -ldflags="-s -w" -o app main.go
```

### Race Detector

```bash
# Обнаружение гонок данных
go run -race main.go
go test -race ./...
go build -race -o app main.go
```

> ⚠️ Важно: Race Detector находит только гонки, которые реально произошли во время выполнения. Тестируйте с покрытием!

## Работа с ОС и вводом-выводом

### Файловая система

```go
import (
    "os"
    "io"
    "io/fs"
)

// Чтение всего файла
data, err := os.ReadFile("file.txt")

// Запись в файл
err := os.WriteFile("file.txt", []byte("content"), 0644)

// Открытие файла
f, err := os.Open("file.txt")
if err != nil {
    return err
}
defer f.Close()

// Чтение по частям
buf := make([]byte, 1024)
n, err := f.Read(buf)

// Обход директории
err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
    if err != nil {
        return err
    }
    if !d.IsDir() {
        fmt.Println(path)
    }
    return nil
})

// Проверка существования файла
_, err := os.Stat("file.txt")
if os.IsNotExist(err) {
    // файл не существует
}
```

### Интерфейсы io.Reader и io.Writer

Эти интерфейсы — основа всей системы ввода-вывода в Go.

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Пример: чтение из HTTP-ответа
resp, _ := http.Get("https://example.com")
defer resp.Body.Close()

// Копирование из reader в writer
_, err := io.Copy(os.Stdout, resp.Body)

// Чтение из reader в строку
body, err := io.ReadAll(resp.Body)

// Композиция reader-ов
r := io.LimitReader(file, 1024) // ограничение чтения
r2 := io.TeeReader(file, os.Stdout) // читает и пишет в другой writer

// Буферизация
bufReader := bufio.NewReader(file)
line, err := bufReader.ReadString('\n')
```

## Тестирование

### Unit-тесты (пакет `testing`)

```go
// файл: math_test.go
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    expected := 5
    if result != expected {
        t.Errorf("Add(2,3) = %d; expected %d", result, expected)
    }
}

// Subtests
func TestAddSubtest(t *testing.T) {
    t.Run("positive numbers", func(t *testing.T) {
        if Add(2, 3) != 5 {
            t.Error("failed")
        }
    })
    t.Run("negative numbers", func(t *testing.T) {
        if Add(-2, -3) != -5 {
            t.Error("failed")
        }
    })
}
```

### Table-driven tests (идиоматичный подход)

```go
func TestAddTableDriven(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"negative", -2, -3, -5},
        {"zero", 0, 5, 5},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Add(tt.a, tt.b)
            if result != tt.expected {
                t.Errorf("Add(%d,%d) = %d; expected %d", tt.a, tt.b, result, tt.expected)
            }
        })
    }
}
```

### Моки через интерфейсы

```go
// Интерфейс для тестирования
type UserRepository interface {
    GetUser(id int) (*User, error)
}

// Реализация для продакшена
type DBUserRepo struct { /* ... */ }

// Мок для тестов
type MockUserRepo struct {
    users map[int]*User
    err   error
}

func (m *MockUserRepo) GetUser(id int) (*User, error) {
    if m.err != nil {
        return nil, m.err
    }
    return m.users[id], nil
}

// Тест с моком
func TestUserService(t *testing.T) {
    mock := &MockUserRepo{
        users: map[int]*User{1: {ID: 1, Name: "Alice"}},
    }
    service := NewUserService(mock)
    
    user, err := service.GetUser(1)
    if err != nil {
        t.Error("unexpected error")
    }
    if user.Name != "Alice" {
        t.Errorf("expected Alice, got %s", user.Name)
    }
}
```

### Фаззинг (Fuzzing)

```go
// файл: fuzz_test.go
func FuzzParseJSON(f *testing.F) {
    // Добавление seed-данных
    f.Add(`{"name": "test", "value": 123}`)
    f.Add(`{}`)
    
    f.Fuzz(func(t *testing.T, data string) {
        var result map[string]interface{}
        err := json.Unmarshal([]byte(data), &result)
        if err != nil {
            // fuzzing находит криминальные случаи
            // проверяем, что ошибка ожидаема
            return
        }
        // Проверяем, что результат не nil
        if result == nil {
            t.Errorf("result is nil for valid JSON: %s", data)
        }
    })
}
```
```bash
# Запуск фаззинга
go test -fuzz=FuzzParseJSON -fuzztime=30s
```

## Практические задания

### 🧪 Задание 1: Пул воркеров

Реализуйте пул воркеров, который обрабатывает задачи из канала. Каждая задача — это функция, принимающая `context.Context`. Реализуйте:

* Graceful shutdown при отмене контекста.
* Обработка паник в воркерах.
* Возврат ошибок через канал ошибок.

### 🧪 Задание 2: Кэш с TTL

Реализуйте потокобезопасный кэш с временем жизни (TTL) записей. Используйте:

* `sync.RWMutex` для защиты.
* `time.Ticker` для периодической очистки истекших записей.
* `context.Context` для graceful shutdown очистителя.

### 🧪 Задание 3: Парсер логов с конкурентностью

Напишите программу, которая читает большой файл логов (например, 100MB) и:

* Разбивает его на чанки.
* Параллельно обрабатывает каждый чанк (ищет ошибки, агрегирует IP-адреса).
* Использует `sync.WaitGroup` для ожидания завершения.
* Собирает результат в общую структуру с защитой от гонок.

### 🧪 Задание 4: Тестирование API

Напишите тесты для HTTP-сервера с использованием:

* Table-driven tests.
* Mock-объектов для внешних зависимостей.
* Subtests для группировки тестов.

## Контрольные вопросы

1. В чём отличие горутин от потоков ОС?
2. Когда использовать буферизированные vs небуферизированные каналы?
3. Как работает `select` с `default`?
4. Какие проблемы решает `sync.WaitGroup`?
5. Чем отличается `sync.Mutex` от `sync.RWMutex`?
6. Какие сценарии использования `sync.Once`?
7. Когда использовать `sync/atomic` вместо `sync.Mutex`?
8. Как `context.Context` помогает при отмене операций
9. Зачем нужны `go mod tidy` и `go mod vendor`?
10. Что такое `go test -race` и когда его использовать?
11. Какие преимущества у Table-driven tests?
12. Как работает фаззинг и для чего он нужен?

## Рекомендуемые ресурсы

### 📖 Обязательно к прочтению

* Go by Example: Concurrency — https://gobyexample.com/goroutines
* Go Concurrency Patterns — https://go.dev/blog/pipelines
Understanding Context in Go — https://go.dev/blog/context

### 📚 Книги

* «Concurrency in Go» (Katherine Cox-Buday) — полное погружение.
* «Go 100 Mistakes» (Harsanyi) — разделы про конкурентность (#58-#70).
### 🎥 Видео

* JustForFunc: Concurrency Patterns — YouTube
* GopherCon Talks — https://www.gophercon.com/

### 🛠️ Инструменты

`go test -cover` — покрытие кода тестами.  
`go test -bench` — бенчмарки (будет на Этапе 3).  
`golangci-lint` — комплексный анализ.  
`staticcheck` — дополнительный линтер.  

### 📄 Документация

* Go Modules Reference — https://go.dev/ref/mod
* Testing Package — https://pkg.go.dev/testing
* Sync Package — https://pkg.go.dev/sync

> 🎯 Результат этапа: вы пишете безопасный конкурентный код, используете весь инструментарий Go, тестируете код системно и понимаете, как управлять зависимостями и сборкой проектов.
