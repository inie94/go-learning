```go
// file: go.mod
module workerpool

go 1.21
```

```go
// file: workerpool.go
// Пакет workerpool реализует паттерн "пул воркеров" для конкурентной обработки задач.
package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Task представляет собой функцию, которая выполняется в воркере.
// Она принимает контекст и возвращает ошибку в случае сбоя.
type Task func(ctx context.Context) error

// WorkerPool управляет пулом горутин (воркеров), которые обрабатывают задачи из очереди.
type WorkerPool struct {
	workersCount int           // количество воркеров
	taskQueue    chan Task     // канал для постановки задач
	errorCh      chan error    // канал для ошибок от задач
	ctx          context.Context // контекст для управления жизненным циклом пула
	cancel       context.CancelFunc // функция отмены контекста
	wg           sync.WaitGroup // ожидание завершения всех воркеров

	// Защита от повторного закрытия каналов.
	closeTaskQueueOnce sync.Once
	closeErrorChOnce   sync.Once

	// Защита состояния "запущен".
	mu      sync.Mutex
	started bool // true, если Start был вызван
}

// NewWorkerPool создаёт новый пул воркеров с заданным количеством воркеров и размером очереди задач.
// Размер очереди также используется для буферизации канала ошибок.
func NewWorkerPool(workersCount int, queueSize int) *WorkerPool {
	// Создаём каналы с буфером указанного размера.
	return &WorkerPool{
		workersCount: workersCount,
		taskQueue:    make(chan Task, queueSize), // буферизированный канал задач
		errorCh:      make(chan error, queueSize), // буферизированный канал ошибок
	}
}

// Start запускает указанное количество воркеров.
// Переданный контекст используется для отмены работы пула извне.
// Внутри создаётся дочерний контекст, который отменяется при вызове Stop.
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Если уже запущен, ничего не делаем (можно также вернуть ошибку, но по спецификации — просто игнорируем).
	if wp.started {
		return
	}

	// Создаём дочерний контекст от переданного, чтобы иметь собственный сигнал отмены.
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.started = true

	// Запускаем горутины-воркеры.
	for i := 0; i < wp.workersCount; i++ {
		wp.wg.Add(1) // увеличиваем счётчик ожидания
		go wp.worker(i) // запускаем воркера с его идентификатором
	}
}

// AddTask добавляет новую задачу в очередь.
// Если пул остановлен или очередь переполнена, возвращается ошибка.
// Отправка в канал неблокирующая (используется select с default) для немедленного возврата при переполнении.
func (wp *WorkerPool) AddTask(task Task) error {
	wp.mu.Lock()
	started := wp.started
	wp.mu.Unlock()

	if !started {
		// Если Start ещё не был вызван, нельзя добавлять задачи.
		return errors.New("pool not started")
	}

	// Проверяем, не отменён ли контекст (пул остановлен).
	// Если контекст ещё не инициализирован (например, Start не завершён), но started=true — такое возможно?
	// В любом случае проверяем на nil.
	if wp.ctx == nil {
		return errors.New("pool not properly initialized")
	}

	select {
	case <-wp.ctx.Done():
		// Контекст отменён — пул в процессе остановки или уже остановлен.
		return errors.New("pool is stopped")
	case wp.taskQueue <- task:
		// Задача успешно помещена в очередь.
		return nil
	default:
		// Очередь задач переполнена — возвращаем ошибку.
		return errors.New("task queue is full")
	}
}

// Stop инициирует graceful shutdown пула:
// - отменяется контекст, чтобы новые задачи не принимались и текущие могли завершиться;
// - закрывается канал задач, чтобы воркеры вышли из цикла после обработки текущих.
// Stop неблокирующий.
func (wp *WorkerPool) Stop() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	if !wp.started {
		// Если пул не запущен, ничего не делаем.
		return
	}

	// Отменяем контекст — это заставит воркеры завершиться, когда они проверят ctx.Done().
	if wp.cancel != nil {
		wp.cancel()
	}

	// Закрываем канал задач — воркеры, ожидающие чтения, получат (ok == false) и выйдут.
	// Используем sync.Once, чтобы избежать паники при повторном закрытии.
	wp.closeTaskQueueOnce.Do(func() {
		close(wp.taskQueue)
	})
}

// StopAndWait останавливает пул и ждёт завершения всех воркеров.
// После этого закрывается канал ошибок.
func (wp *WorkerPool) StopAndWait() {
	// Останавливаем пул (отменяем контекст и закрываем taskQueue).
	wp.Stop()

	// Ожидаем, пока все воркеры завершат выполнение.
	wp.wg.Wait()

	// Закрываем канал ошибок — после завершения всех воркеров отправок больше не будет.
	wp.closeErrorChOnce.Do(func() {
		close(wp.errorCh)
	})
}

// Errors возвращает канал только для чтения, из которого можно получать ошибки, возникшие при выполнении задач.
// Канал закрывается после вызова StopAndWait.
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errorCh
}

// worker — внутренняя функция, выполняющаяся в отдельной горутине.
// Она читает задачи из taskQueue и выполняет их, обрабатывая паники и отправляя ошибки в errorCh.
func (wp *WorkerPool) worker(id int) {
	// Уменьшаем счётчик ожидания при выходе из горутины.
	defer wp.wg.Done()

	// Перехватываем панику, если она произойдёт в самом воркере (например, при отправке в закрытый канал).
	// Это защита от неожиданных паник, но при правильной работе их быть не должно.
	defer func() {
		if r := recover(); r != nil {
			// Попытка отправить ошибку о панике воркера.
			// Если канал ошибок закрыт, это вызовет панику, но мы закрываем его только после wg.Wait(),
			// значит, к этому моменту все воркеры уже завершены, так что отправка не произойдёт.
			// Тем не менее, добавляем обработку на случай, если канал закроется раньше.
			select {
			case wp.errorCh <- fmt.Errorf("worker %d panic: %v", id, r):
			default:
				// Если не удалось отправить (канал закрыт или переполнен), логируем или игнорируем.
				// В реальном проекте здесь можно использовать log.
			}
		}
	}()

	// Основной цикл обработки задач.
	for {
		select {
		case <-wp.ctx.Done():
			// Контекст отменён — выходим из воркера.
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				// Канал задач закрыт — больше задач не будет, выходим.
				return
			}
			// Выполняем задачу в отдельной анонимной функции для изоляции паники.
			func() {
				// Перехватываем панику внутри задачи.
				defer func() {
					if r := recover(); r != nil {
						// Отправляем ошибку о панике в канал ошибок.
						// Используем select с default, чтобы не блокироваться, если канал заполнен или закрыт.
						select {
						case wp.errorCh <- fmt.Errorf("task panic in worker %d: %v", id, r):
						default:
							// Если не удалось отправить (канал закрыт или переполнен), игнорируем.
							// В реальном проекте можно логировать.
						}
					}
				}()
				// Выполняем задачу с контекстом пула.
				if err := task(wp.ctx); err != nil {
					// Отправляем ошибку выполнения задачи.
					select {
					case wp.errorCh <- fmt.Errorf("task error in worker %d: %w", id, err):
					default:
						// Если канал ошибок переполнен или закрыт, пропускаем.
					}
				}
			}()
		}
	}
}
```

```go
// file: main.go (пример использования)
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"workerpool" // импортируем наш пакет
)

func main() {
	// Создаём контекст с возможностью отмены (можно использовать и без неё).
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // гарантируем отмену при выходе

	// Создаём пул с 5 воркерами и очередью на 10 задач.
	pool := workerpool.NewWorkerPool(5, 10)

	// Запускаем пул.
	pool.Start(ctx)

	// Отдельная горутина для чтения ошибок из канала.
	go func() {
		for err := range pool.Errors() {
			log.Printf("Получена ошибка: %v", err)
		}
		log.Println("Канал ошибок закрыт, чтение завершено")
	}()

	// Добавляем 100 задач.
	for i := 0; i < 100; i++ {
		// Захватываем i для замыкания.
		taskID := i
		err := pool.AddTask(func(ctx context.Context) error {
			// Имитируем работу.
			select {
			case <-ctx.Done():
				// Если контекст отменён, выходим с ошибкой.
				return fmt.Errorf("task %d cancelled", taskID)
			case <-time.After(time.Duration(50+taskID%50) * time.Millisecond):
				// Если задача с номером, кратным 10, генерируем ошибку.
				if taskID%10 == 0 && taskID != 0 {
					return fmt.Errorf("simulated error in task %d", taskID)
				}
				// Если задача с номером, кратным 7, паникуем.
				if taskID%7 == 0 {
					panic(fmt.Sprintf("panic in task %d", taskID))
				}
				// Успешное завершение.
				fmt.Printf("Task %d completed\n", taskID)
				return nil
			}
		})
		if err != nil {
			log.Printf("Не удалось добавить задачу %d: %v", taskID, err)
		}
	}

	// Даём время на выполнение задач (в реальном коде можно использовать StopAndWait).
	time.Sleep(2 * time.Second)

	// Останавливаем пул с ожиданием завершения всех задач.
	pool.StopAndWait()
	fmt.Println("Все задачи обработаны, пул остановлен")
}
```