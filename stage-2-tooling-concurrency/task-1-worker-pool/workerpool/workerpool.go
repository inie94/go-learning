// file: workerpool.go
package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Task представляет функцию, выполняемую воркером.
type Task func(ctx context.Context) error

// WorkerPool управляет пулом горутин с дополнительными улучшениями.
type WorkerPool struct {
	taskQueue		chan Task
	errorCh         chan error

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	targetWorkers int
	workerCount   int
	workerMu      sync.Mutex
	started       bool

	// Метрики (используем atomic для потокобезопасности)
	totalTasks     atomic.Int64
	completedTasks atomic.Int64
	failedTasks    atomic.Int64
	totalDuration  atomic.Int64 // наносекунды

	closeOnce sync.Once
}

// NewWorkerPool создаёт новый пул.
func NewWorkerPool(workersCount int, queueSize int) *WorkerPool {
	if workersCount < 0 {
		workersCount = 0
	}
	if queueSize < 0 {
		queueSize = 0
	}
	return &WorkerPool{
		taskQueue: make(chan Task, queueSize),
		errorCh:         make(chan error, queueSize),
		targetWorkers:   workersCount,
	}
}

// Start запускает пул.
func (wp *WorkerPool) Start(ctx context.Context) {
	wp.workerMu.Lock()
	defer wp.workerMu.Unlock()
	if wp.started {
		return
	}
	wp.ctx, wp.cancel = context.WithCancel(ctx)
	wp.started = true
	wp.workerCount = 0
	for i := 0; i < wp.targetWorkers; i++ {
		wp.startWorker()
	}
}

// startWorker запускает одного воркера (вызывается с заблокированной workerMu).
func (wp *WorkerPool) startWorker() {
	id := wp.workerCount
	wp.workerCount++
	wp.wg.Add(1)
	go wp.worker(id, wp.ctx)
}

// AddTask добавляет задачу со средним приоритетом.
func (wp *WorkerPool) AddTask(task Task) error {
wp.workerMu.Lock()
	started := wp.started
	wp.workerMu.Unlock()
	if !started {
		return errors.New("Pool not started")
	}
	select {
	case <-wp.ctx.Done():
		return errors.New("Pool is stopped")
	default:
	}

	select {
	case wp.taskQueue <- task:
		wp.totalTasks.Add(1)
		return nil
	default:
		return fmt.Errorf("Task queue is full")
	}
}

// AddWorker увеличивает число воркеров.
func (wp *WorkerPool) AddWorker() error {
	wp.workerMu.Lock()
	defer wp.workerMu.Unlock()
	if !wp.started {
		return errors.New("pool not started")
	}
	wp.targetWorkers++
	wp.startWorker()
	return nil
}

// RemoveWorker уменьшает число воркеров (до минимум 1).
func (wp *WorkerPool) RemoveWorker() error {
	wp.workerMu.Lock()
	defer wp.workerMu.Unlock()
	if !wp.started {
		return errors.New("pool not started")
	}
	if wp.targetWorkers <= 1 {
		return errors.New("cannot remove worker: at least one required")
	}
	wp.targetWorkers--
	return nil
}

// Stop инициирует graceful shutdown: закрывает очереди задач и запрещает добавление новых.
// Воркеры продолжают обрабатывать оставшиеся задачи, контекст НЕ отменяется.
func (wp *WorkerPool) Stop() {
	wp.workerMu.Lock()
	defer wp.workerMu.Unlock()
	if !wp.started {
		return
	}
	// Закрываем очереди задач – воркеры выйдут после их опустошения.
	close(wp.taskQueue)
	// Запрещаем добавление новых задач.
	wp.started = false
}

// StopAndWait останавливает пул, ждёт завершения всех воркеров и закрывает канал ошибок.
func (wp *WorkerPool) StopAndWait() {
	wp.Stop()          // закрывает очереди, started=false
	wp.wg.Wait()       // ждём завершения воркеров (после обработки всех задач)
	wp.closeOnce.Do(func() {
		close(wp.errorCh)
	})
	// Не отменяем контекст – он может быть отменён только извне.
}

// Errors возвращает канал ошибок.
func (wp *WorkerPool) Errors() <-chan error {
	return wp.errorCh
}

// --- Воркер ---

// worker – горутина, обрабатывающая задачи с приоритетом.
func (wp *WorkerPool) worker(id int, ctx context.Context) {
	defer wp.wg.Done()
	defer func() {
		if r := recover(); r != nil {
			select {
			case wp.errorCh <- fmt.Errorf("worker %d panic: %v", id, r):
			default:
			}
		}
	}()

	for {
		// Проверка: нужно ли удалить воркера (при динамическом изменении количества)
		wp.workerMu.Lock()
		target := wp.targetWorkers
		wp.workerMu.Unlock()
		if id >= target {
			return
		}

		// Если все очереди закрыты, выходим
		if wp.taskQueue == nil {
			return
		}

		// Все очереди пусты, но не закрыты – ждём появления данных в любой из них
		select {
		case <-ctx.Done():
			return
		case task, ok :=  <-wp.taskQueue:
			if !ok {
				wp.taskQueue = nil
				continue
			}
			wp.executeTask(task, id)
		}
	}
}

// executeTask выполняет задачу и обновляет метрики.
func (wp *WorkerPool) executeTask(task Task, workerID int) {
	start := time.Now()

	// Убедимся, что метрика времени обновляется даже при панике.
	defer func() {
		// Используем elapsed и передаём в atomic.AddInt64
		elapsed := time.Since(start).Nanoseconds()
		wp.totalDuration.Add(elapsed)
	}()

	// Защита от паники в задаче.
	defer func() {
		if r := recover(); r != nil {
			wp.failedTasks.Add(1)
			select {
			case wp.errorCh <- fmt.Errorf("task panic in worker %d: %v", workerID, r):
			default:
			}
		}
	}()

	if err := task(wp.ctx); err != nil {
		wp.failedTasks.Add(1)
		select {
		case wp.errorCh <- fmt.Errorf("task error in worker %d: %w", workerID, err):
		default:
		}
	} else {
		wp.completedTasks.Add(1)
	}
}

// --- Метрики ---

func (wp *WorkerPool) TotalTasks() int64 {
	return wp.totalTasks.Load()
}

func (wp *WorkerPool) CompletedTasks() int64 {
	return wp.completedTasks.Load()
}

func (wp *WorkerPool) FailedTasks() int64 {
	return wp.failedTasks.Load()
}

func (wp *WorkerPool) AverageDuration() float64 {
	completed := wp.completedTasks.Load()
	if completed == 0 {
		return 0
	}
	total := wp.totalDuration.Load()
	return float64(total) / float64(completed)
}

// --- Вспомогательные обёртки ---

// TaskWithTimeout добавляет ограничение по времени.
func TaskWithTimeout(task Task, timeout time.Duration) Task {
	return func(ctx context.Context) error {
		ctxTimeout, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		err := task(ctxTimeout)
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("task timeout: %w", err)
		}
		return err
	}
}

// TaskWithRetry добавляет повторные попытки.
func TaskWithRetry(task Task, maxRetries int, initialDelay time.Duration) Task {
	if maxRetries < 1 {
		maxRetries = 1
	}
	return func(ctx context.Context) error {
		var lastErr error
		delay := initialDelay
		for attempt := 0; attempt < maxRetries; attempt++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			if attempt > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
				}
				delay *= 2
			}
			lastErr = task(ctx)
			if lastErr == nil {
				return nil
			}
		}
		return fmt.Errorf("retry failed after %d attempts: %w", maxRetries, lastErr)
	}
}