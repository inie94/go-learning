package cache

import (
	"context"
	"sync"
	"time"
)

// item представляет одну запись в кэше.
type item struct {
	value      interface{}
	expiration int64 // Unix-наносекундаб после которой запись считается истекшей; 0 означает "вечная"
}

// Cache - основная структура кэша.
type Cache struct {
	items  map[string]item
	mu     sync.RWMutex
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New создает новый экземпляр кэша с фоновой очисткой через заданный интервал.
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
// Если ttl <= 0, запись считаетсся вечной (никогда не истекает).
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
		// Ленивое удаление - блокируем на запись и удаляем
		c.mu.Lock()
		// Проверяем еще раз, не удалил ли кто-то параллельно
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
// После вызова Stop кэш продолжает работать, но истекшие записи больше не даляются автоматически.
func (c *Cache) Stop() {
	c.cancel()      // сигнал горутине остановиться
	c.ticker.Stop() // останавливаем тикер
	c.wg.Wait()     // ждём завершения очистителя

}

// cleanupLoop - фоновая горутина, периодически удаляющая истекшие записи.
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

// deleteExpired - удаляет все истекшие записи. Вызывается из фоновой горутины.
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
