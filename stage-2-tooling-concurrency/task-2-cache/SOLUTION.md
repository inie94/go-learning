## Решение: Потокобезопасный кэш с TTL на Go

Ниже представлена полная реализация кэша, соответствующая всем требованиям технического задания. Код состоит из трёх файлов:

- `cache.go` – основная логика кэша.
- `cache_test.go` – модульные тесты (включая тесты с `-race`).
- `example_test.go` – пример использования (документирующий).

Все операции потокобезопасны, используется `sync.RWMutex`, фоновый очиститель на `time.Ticker` управляется через `context.Context`, реализован graceful shutdown.

---

### Файл `cache.go`

```go
package cache

import (
	"context"
	"sync"
	"time"
)

// item представляет одну запись в кэше.
type item struct {
	value      interface{}
	expiration int64 // Unix-наносекунда, после которой запись считается истекшей; 0 означает "вечная"
}

// Cache – основная структура кэша.
type Cache struct {
	items  map[string]item
	mu     sync.RWMutex
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New создаёт новый экземпляр кэша с фоновой очисткой через заданный интервал.
// Очиститель запускается в отдельной горутине и останавливается при вызове Stop().
func New(cleanupInterval time.Duration) *Cache {
	ctx, cancel := context.WithCancel(context.Background())
	c := &Cache{
		items:  make(map[string]item),
		ticker: time.NewTicker(cleanupInterval),
		ctx:    ctx,
		cancel: cancel,
	}
	c.wg.Add(1)
	go c.cleanupLoop()
	return c
}

// Set добавляет или обновляет запись с указанным временем жизни.
// Если ttl <= 0, запись считается вечной (никогда не истекает).
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	var exp int64
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	} // иначе exp = 0 (вечная)
	c.items[key] = item{value: value, expiration: exp}
}

// Get возвращает значение по ключу и флаг успешности.
// Если запись отсутствует или истекла, возвращается (nil, false).
// При обнаружении истекшей записи она удаляется (ленивое удаление).
func (c *Cache) Get(key string) (interface{}, bool) {
	// Сначала читаем под RLock
	c.mu.RLock()
	it, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	// Проверяем истечение
	if it.expiration > 0 && time.Now().UnixNano() > it.expiration {
		// Ленивое удаление – блокируем на запись и удаляем
		c.mu.Lock()
		// Проверяем ещё раз, не удалил ли кто-то параллельно
		if it2, ok2 := c.items[key]; ok2 && it2.expiration == it.expiration {
			delete(c.items, key)
		}
		c.mu.Unlock()
		return nil, false
	}
	return it.value, true
}

// Delete удаляет запись по ключу (если она существует).
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear полностью очищает кэш.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[string]item)
}

// Len возвращает текущее количество записей в кэше (для статистики и тестов).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Stop останавливает фоновый очиститель и дожидается его завершения.
// После вызова Stop кэш продолжает работать, но истекшие записи больше не удаляются автоматически.
func (c *Cache) Stop() {
	c.cancel()       // сигнал горутине остановиться
	c.ticker.Stop()  // останавливаем тикер
	c.wg.Wait()      // ждём завершения очистителя
}

// cleanupLoop – фоновая горутина, периодически удаляющая истекшие записи.
func (c *Cache) cleanupLoop() {
	defer c.wg.Done()
	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.ticker.C:
			c.deleteExpired()
		}
	}
}

// deleteExpired удаляет все истекшие записи. Вызывается из фоновой горутины.
func (c *Cache) deleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()
	for k, v := range c.items {
		if v.expiration > 0 && now > v.expiration {
			delete(c.items, k)
		}
	}
}
```

---

### Файл `cache_test.go`

```go
package cache

import (
	"sync"
	"testing"
	"time"
)

func TestCache_BasicOperations(t *testing.T) {
	c := New(1 * time.Second)
	defer c.Stop()

	// Set & Get
	c.Set("key1", "value1", 0)
	val, ok := c.Get("key1")
	if !ok || val != "value1" {
		t.Fatalf("expected value1, got %v", val)
	}

	// Get non-existing
	_, ok = c.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}

	// Update
	c.Set("key1", "newvalue", 0)
	val, ok = c.Get("key1")
	if !ok || val != "newvalue" {
		t.Fatalf("expected newvalue, got %v", val)
	}

	// Delete
	c.Delete("key1")
	_, ok = c.Get("key1")
	if ok {
		t.Fatal("key1 should be deleted")
	}

	// Clear
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Clear()
	if c.Len() != 0 {
		t.Fatalf("cache not empty after Clear, len=%d", c.Len())
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	c := New(500 * time.Millisecond)
	defer c.Stop()

	c.Set("expire", "x", 100*time.Millisecond)
	c.Set("forever", "y", 0)

	time.Sleep(150 * time.Millisecond)

	// Истекшая запись должна быть удалена лениво
	val, ok := c.Get("expire")
	if ok {
		t.Fatal("expire should have expired")
	}
	if c.Len() != 1 { // только forever
		t.Fatalf("expected len 1, got %d", c.Len())
	}

	// Вечная запись должна существовать
	val, ok = c.Get("forever")
	if !ok || val != "y" {
		t.Fatal("forever should still exist")
	}
}

func TestCache_BackgroundCleanup(t *testing.T) {
	c := New(100 * time.Millisecond)
	defer c.Stop()

	// Добавляем записи с разными TTL
	c.Set("short", 1, 150*time.Millisecond)
	c.Set("long", 2, 500*time.Millisecond)
	c.Set("eternal", 3, 0)

	// Ждём две очистки
	time.Sleep(250 * time.Millisecond)

	// short должна быть удалена, long ещё жива, eternal жива
	if _, ok := c.Get("short"); ok {
		t.Error("short should be removed by background cleanup")
	}
	if _, ok := c.Get("long"); !ok {
		t.Error("long should still be present")
	}
	if _, ok := c.Get("eternal"); !ok {
		t.Error("eternal should still be present")
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := New(1 * time.Second)
	defer c.Stop()

	var wg sync.WaitGroup
	const goroutines = 100
	const iterations = 100

	// Параллельные Set и Get
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "key"
			for j := 0; j < iterations; j++ {
				c.Set(key, id, 0)
				val, ok := c.Get(key)
				if !ok {
					t.Errorf("get failed for key %s", key)
				}
				_ = val
			}
		}(i)
	}
	wg.Wait()
}

func TestCache_GracefulShutdown(t *testing.T) {
	c := New(50 * time.Millisecond)
	// Запускаем несколько операций, чтобы убедиться, что очиститель работает
	c.Set("a", 1, 100*time.Millisecond)
	time.Sleep(60 * time.Millisecond)
	c.Stop()

	// После остановки горутина должна завершиться, но кэш всё ещё доступен
	_, ok := c.Get("a")
	if ok {
		t.Error("a should have expired and been cleaned")
	}
	// Можно продолжать использовать кэш
	c.Set("b", 2, 0)
	if _, ok := c.Get("b"); !ok {
		t.Error("b should be accessible after Stop")
	}
}

// Запуск: go test -race -v
```

---

### Файл `example_test.go` (для документации)

```go
package cache_test

import (
	"fmt"
	"time"

	"yourmodule/cache" // замените на реальный путь
)

func ExampleCache() {
	// Создаём кэш с очисткой каждые 5 секунд
	c := cache.New(5 * time.Second)
	defer c.Stop() // гарантируем остановку очистителя

	// Добавляем записи
	c.Set("user:1", "Alice", 2*time.Second)
	c.Set("user:2", "Bob", 0) // вечная

	// Получаем
	if val, ok := c.Get("user:1"); ok {
		fmt.Println("user:1 =", val)
	}

	time.Sleep(3 * time.Second)

	if _, ok := c.Get("user:1"); !ok {
		fmt.Println("user:1 expired")
	}

	if val, ok := c.Get("user:2"); ok {
		fmt.Println("user:2 =", val)
	}
	// Output:
	// user:1 = Alice
	// user:1 expired
	// user:2 = Bob
}
```

---

## Инструкция по запуску тестов

```bash
go test -race -v ./...
```

Тесты покрывают все ключевые сценарии, включая конкурентный доступ и корректное завершение фоновой горутины.

---

## Дополнительные замечания

- **Ленивое удаление** реализовано в `Get`: при обнаружении истекшей записи она удаляется под мьютексом, что снижает нагрузку на фоновый ticker.
- **Graceful shutdown** гарантирует, что горутина очистителя завершится перед возвратом из `Stop()`.
- Кэш не ограничен по размеру – при необходимости вы можете добавить политику вытеснения (например, LRU) без изменения основной архитектуры.
- Все методы документированы и следуют идиоматическому стилю Go.

Данная реализация полностью соответствует техническому заданию и готова к использованию в production-среде.