// file: workerpool.go
package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// Task представляет функцию, выполняемую воркером.
type Task func(ctx context.Context) error

// WorkerPool управляет пулом горутин для обработки задач.
type WorkerPool struct {
	workersCount int
	taskQueue    chan Task
	errorCh      chan error
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup

	mu      sync.Mutex
	started bool // true, если пул запущен и активен

	closeTaskQueueOnce sync.Once
	closeErrorChOnce   sync.Once
}

// NewWorkerPool создаёт новый пул с заданным числом воркеров и размером очереди.
func NewWorkerPool(workersCount int, queueSize int) *WorkerPool {
	return &WorkerPool{
		workersCount: workersCount,
		taskQueue:    make(chan Task, queueSize),
		errorCh:      make(chan error, queueSize),
	}
}

// Start запускает воркеры. Контекст используется для внешней отмены.
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	if wp.started {
		return
	}
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.started = true

	for i := 0; i < wp.workersCount; i++ {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// AddTask добавляет задачу в очередь. Возвращает ошибку, если пул остановлен или очередь полна.
func (wp *WorkerPool) AddTask(task Task) error {
    wp.mu.Lock()
    started := wp.started
    wp.mu.Unlock()

    if !started {
        return errors.New("pool not started or already stopped")
    }

    select {
    case <-wp.ctx.Done():
        return errors.New("pool context cancelled")
    case wp.taskQueue <- task:
        return nil
    default:
        return errors.New("task queue is full")
    }
}

// Stop инициирует graceful shutdown: закрывает канал задач и запрещает добавление новых задач.
func (wp *WorkerPool) Stop() {
    wp.mu.Lock()
    defer wp.mu.Unlock()
    if !wp.started {
        return
    }
    wp.closeTaskQueueOnce.Do(func() {
        close(wp.taskQueue)
    })
    // Устанавливаем флаг, чтобы AddTask возвращал ошибку, а не пытался писать в закрытый канал.
    wp.started = false
}

// StopAndWait останавливает пул, ожидает завершения всех воркеров и закрывает канал ошибок.
func (wp *WorkerPool) StopAndWait() {
    wp.Stop()          // закрываем taskQueue и устанавливаем started=false
    wp.wg.Wait()       // ждём, пока все воркеры завершатся
    wp.closeErrorChOnce.Do(func() {
        close(wp.errorCh) // закрываем канал ошибок
    })
}

// Errors возвращает канал для чтения ошибок.
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errorCh
}

// worker – горутина-воркер.
func (wp *WorkerPool) worker(id int) {
	defer wp.wg.Done()

	// Защита от паники в самом воркере (на случай ошибок отправки в errorCh).
	defer func() {
		if r := recover(); r != nil {
			select {
			case wp.errorCh <- fmt.Errorf("worker %d panic: %v", id, r):
			default:
			}
		}
	}()

	for {
		select {
		case <-wp.ctx.Done():
			// Контекст отменён извне – выходим, не дожидаясь задач.
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				// Канал задач закрыт – больше задач не будет.
				return
			}
			// Выполняем задачу с защитой от паники.
			func() {
				defer func() {
					if r := recover(); r != nil {
						select {
						case wp.errorCh <- fmt.Errorf("task panic in worker %d: %v", id, r):
						default:
						}
					}
				}()
				if err := task(wp.ctx); err != nil {
					select {
					case wp.errorCh <- fmt.Errorf("task error in worker %d: %w", id, err):
					default:
					}
				}
			}()
		}
	}
}