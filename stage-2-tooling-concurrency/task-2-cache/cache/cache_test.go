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
	c.Set("a", 1, 20*time.Millisecond)
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
