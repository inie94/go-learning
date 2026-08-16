// file: workerpool_test.go
package workerpool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// --- Базовые тесты ---

func TestNewWorkerPool(t *testing.T) {
	wp := NewWorkerPool(3, 5)
	if wp.targetWorkers != 3 {
		t.Errorf("expected targetWorkers 3, got %d", wp.targetWorkers)
	}
	if cap(wp.highTaskQueue) != 5 {
		t.Errorf("expected highTaskQueue cap 5, got %d", cap(wp.highTaskQueue))
	}
	if cap(wp.mediumTaskQueue) != 5 {
		t.Errorf("expected mediumTaskQueue cap 5, got %d", cap(wp.mediumTaskQueue))
	}
	if cap(wp.lowTaskQueue) != 5 {
		t.Errorf("expected lowTaskQueue cap 5, got %d", cap(wp.lowTaskQueue))
	}
	if cap(wp.errorCh) != 5 {
		t.Errorf("expected errorCh cap 5, got %d", cap(wp.errorCh))
	}
	if wp.started {
		t.Error("new pool should not be started")
	}
}

func TestStartAndStop(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())
	if !wp.started {
		t.Error("pool should be started after Start")
	}
	wp.StopAndWait()
	if wp.started {
		t.Error("pool should be stopped after StopAndWait")
	}
	_, ok := <-wp.Errors()
	if ok {
		t.Error("errorCh should be closed")
	}
}

func TestBasic(t *testing.T) {
	wp := NewWorkerPool(3, 10)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var completed int32
	const n = 10
	for i := 0; i < n; i++ {
		err := wp.AddTask(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	}
	time.Sleep(200 * time.Millisecond)
	if atomic.LoadInt32(&completed) != n {
		t.Errorf("expected %d, got %d", n, completed)
	}
	select {
	case err := <-wp.Errors():
		t.Errorf("unexpected error: %v", err)
	default:
	}
}

func TestGracefulShutdown(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())

	var wg sync.WaitGroup
	var completed int32
	const longTasks = 3
	wg.Add(longTasks)

	for i := 0; i < longTasks; i++ {
		err := wp.AddTask(func(ctx context.Context) error {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	wp.Stop()

	err := wp.AddTask(func(ctx context.Context) error { return nil })
	if err == nil {
		t.Error("expected error when adding task after Stop")
	}

	wg.Wait()
	if atomic.LoadInt32(&completed) != longTasks {
		t.Errorf("expected %d tasks completed, got %d", longTasks, completed)
	}

	wp.StopAndWait()
	_, ok := <-wp.Errors()
	if ok {
		t.Error("errorCh should be closed")
	}
}

func TestPanicHandling(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var completed int32
	panicReceived := make(chan bool, 1)
	go func() {
		for range wp.Errors() {
			panicReceived <- true
			return
		}
	}()

	err := wp.AddTask(func(ctx context.Context) error {
		panic("test panic")
	})
	if err != nil {
		t.Fatalf("failed to add panic task: %v", err)
	}

	err = wp.AddTask(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&completed, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to add normal task: %v", err)
	}

	select {
	case <-panicReceived:
	case <-time.After(time.Second):
		t.Error("timeout waiting for panic error")
	}
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&completed) != 1 {
		t.Errorf("expected 1 normal task, got %d", completed)
	}
}

func TestContextCancellation(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	ctx, cancel := context.WithCancel(context.Background())
	wp.Start(ctx)

	var cancelled int32
	var completed int32

	err := wp.AddTask(func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			atomic.AddInt32(&cancelled, 1)
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			atomic.AddInt32(&completed, 1)
			return nil
		}
	})
	if err != nil {
		t.Fatalf("failed to add task: %v", err)
	}

	time.Sleep(20 * time.Millisecond)
	cancel()

	wp.StopAndWait()
	if atomic.LoadInt32(&cancelled) != 1 {
		t.Errorf("expected cancelled 1, got %d", cancelled)
	}
	if atomic.LoadInt32(&completed) != 0 {
		t.Errorf("expected completed 0, got %d", completed)
	}
}

func TestQueueFull(t *testing.T) {
	wp := NewWorkerPool(1, 2)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	block := func(ctx context.Context) error {
		time.Sleep(100 * time.Millisecond)
		return nil
	}
	for i := 0; i < 3; i++ {
		err := wp.AddTask(block)
		if i < 2 {
			if err != nil {
				t.Fatalf("unexpected error at %d: %v", i, err)
			}
		} else {
			if err == nil {
				t.Error("expected error on full queue")
			}
		}
	}
}

func TestStopIdempotent(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())
	wp.Stop()
	wp.Stop()
	wp.StopAndWait()
	wp.StopAndWait()
}

func TestAddTaskAfterStop(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())
	wp.Stop()
	err := wp.AddTask(func(ctx context.Context) error { return nil })
	if err == nil {
		t.Error("expected error after Stop")
	}
	wp.StopAndWait()
}

func TestErrorsChannel(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	var errCount int32
	go func() {
		defer wg.Done()
		for range wp.Errors() {
			atomic.AddInt32(&errCount, 1)
		}
	}()

	err := wp.AddTask(func(ctx context.Context) error {
		return errors.New("simulated error")
	})
	if err != nil {
		t.Fatalf("failed to add error task: %v", err)
	}
	err = wp.AddTask(func(ctx context.Context) error { return nil })
	if err != nil {
		t.Fatalf("failed to add success task: %v", err)
	}

	wp.StopAndWait()
	wg.Wait()
	if atomic.LoadInt32(&errCount) != 1 {
		t.Errorf("expected 1 error, got %d", errCount)
	}
}

func TestWorkerRecovery(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var completed int32
	err := wp.AddTask(func(ctx context.Context) error {
		panic("panic")
	})
	if err != nil {
		t.Fatalf("failed to add panic task: %v", err)
	}
	err = wp.AddTask(func(ctx context.Context) error {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&completed, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to add second task: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&completed) != 1 {
		t.Errorf("expected second task completed, got %d", completed)
	}
}

func TestStartTwice(t *testing.T) {
	wp := NewWorkerPool(2, 5)
	wp.Start(context.Background())
	wp.Start(context.Background())
	if !wp.started {
		t.Error("pool should remain started")
	}
	wp.StopAndWait()
}

func TestZeroWorkers(t *testing.T) {
	wp := NewWorkerPool(0, 5)
	wp.Start(context.Background())
	err := wp.AddTask(func(ctx context.Context) error { return nil })
	if err != nil {
		t.Errorf("unexpected error with 0 workers: %v", err)
	}
	wp.StopAndWait()
	_, ok := <-wp.Errors()
	if ok {
		t.Error("errorCh should be closed")
	}
}

// --- Новые тесты для улучшений ---

func TestPriority(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var order []int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 3; i++ {
		wg.Add(1)
		err := wp.AddTaskWithPriority(HighPriority, func(ctx context.Context) error {
			defer wg.Done()
			mu.Lock()
			order = append(order, 0)
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add high task: %v", err)
		}

		wg.Add(1)
		err = wp.AddTaskWithPriority(MediumPriority, func(ctx context.Context) error {
			defer wg.Done()
			mu.Lock()
			order = append(order, 1)
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add medium task: %v", err)
		}

		wg.Add(1)
		err = wp.AddTaskWithPriority(LowPriority, func(ctx context.Context) error {
			defer wg.Done()
			mu.Lock()
			order = append(order, 2)
			mu.Unlock()
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add low task: %v", err)
		}
	}

	wg.Wait()

	if len(order) != 9 {
		t.Fatalf("expected 9 tasks, got %d", len(order))
	}
	for i := 0; i < 3; i++ {
		if order[i] != 0 {
			t.Errorf("position %d: expected 0 (High), got %d", i, order[i])
		}
	}
	for i := 3; i < 6; i++ {
		if order[i] != 1 {
			t.Errorf("position %d: expected 1 (Medium), got %d", i, order[i])
		}
	}
	for i := 6; i < 9; i++ {
		if order[i] != 2 {
			t.Errorf("position %d: expected 2 (Low), got %d", i, order[i])
		}
	}
}

func TestAddRemoveWorker(t *testing.T) {
	// Увеличиваем очередь до 20, чтобы все 10 задач поместились
	wp := NewWorkerPool(2, 20)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	start := time.Now()
	for i := 0; i < 10; i++ {
		err := wp.AddTask(func(ctx context.Context) error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	}
	time.Sleep(100 * time.Millisecond)
	err := wp.AddWorker()
	if err != nil {
		t.Fatalf("failed to add worker: %v", err)
	}
	time.Sleep(200 * time.Millisecond)
	elapsed := time.Since(start)

	// Удалим воркера
	err = wp.RemoveWorker()
	if err != nil {
		t.Fatalf("failed to remove worker: %v", err)
	}
	var done int32
	err = wp.AddTask(func(ctx context.Context) error {
		atomic.AddInt32(&done, 1)
		return nil
	})
	if err != nil {
		t.Fatalf("failed to add task after removal: %v", err)
	}
	time.Sleep(50 * time.Millisecond)
	if atomic.LoadInt32(&done) != 1 {
		t.Error("task after removal not executed")
	}

	// Проверка удаления последнего воркера
	err = wp.RemoveWorker()
	if err != nil {
		// сейчас targetWorkers = 2 (было 3, удалили один) -> можно
	}
	err = wp.RemoveWorker()
	if err == nil {
		t.Error("expected error when removing last worker")
	}
	_ = elapsed // используем, чтобы избежать unused
}

func TestMetrics(t *testing.T) {
	// Увеличиваем очередь до 20
	wp := NewWorkerPool(2, 20)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	for i := 0; i < 5; i++ {
		err := wp.AddTask(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add success task: %v", err)
		}
	}
	for i := 0; i < 3; i++ {
		err := wp.AddTask(func(ctx context.Context) error {
			time.Sleep(10 * time.Millisecond)
			return errors.New("simulated error")
		})
		if err != nil {
			t.Fatalf("failed to add error task: %v", err)
		}
	}
	err := wp.AddTask(func(ctx context.Context) error {
		panic("panic for metrics")
	})
	if err != nil {
		t.Fatalf("failed to add panic task: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	wp.StopAndWait()

	total := wp.TotalTasks()
	completed := wp.CompletedTasks()
	failed := wp.FailedTasks()
	avg := wp.AverageDuration()

	if total != 9 {
		t.Errorf("expected TotalTasks 9, got %d", total)
	}
	if completed != 5 {
		t.Errorf("expected CompletedTasks 5, got %d", completed)
	}
	if failed != 4 {
		t.Errorf("expected FailedTasks 4, got %d", failed)
	}
	if avg < 10_000_000 {
		t.Errorf("AverageDuration too low: %f ns", avg)
	}
}

func TestTaskWithTimeout(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	longTask := func(ctx context.Context) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
			return nil
		}
	}
	timedTask := TaskWithTimeout(longTask, 100*time.Millisecond)

	err := wp.AddTask(timedTask)
	if err != nil {
		t.Fatalf("failed to add timed task: %v", err)
	}

	time.Sleep(150 * time.Millisecond)
	select {
	case err := <-wp.Errors():
		if err == nil {
			t.Error("expected error from timed task, got nil")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for error")
	}
}

func TestTaskWithRetry(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var attemptCount int32
	failingTask := func(ctx context.Context) error {
		curr := atomic.AddInt32(&attemptCount, 1)
		if curr <= 2 {
			return errors.New("temporary error")
		}
		return nil
	}
	retryTask := TaskWithRetry(failingTask, 3, 10*time.Millisecond)

	err := wp.AddTask(retryTask)
	if err != nil {
		t.Fatalf("failed to add retry task: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
	select {
	case err := <-wp.Errors():
		t.Errorf("unexpected error: %v", err)
	default:
	}
}

func TestTaskWithRetryExhausted(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var attemptCount int32
	alwaysFail := func(ctx context.Context) error {
		atomic.AddInt32(&attemptCount, 1)
		return errors.New("always fail")
	}
	retryTask := TaskWithRetry(alwaysFail, 3, 10*time.Millisecond)

	err := wp.AddTask(retryTask)
	if err != nil {
		t.Fatalf("failed to add retry task: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
	select {
	case err := <-wp.Errors():
		if err == nil {
			t.Error("expected error from exhausted retry")
		}
	case <-time.After(50 * time.Millisecond):
		t.Error("timeout waiting for error")
	}
}

func TestCombinedTimeoutAndRetry(t *testing.T) {
	wp := NewWorkerPool(1, 5)
	wp.Start(context.Background())
	defer wp.StopAndWait()

	var attemptCount int32
	slowTask := func(ctx context.Context) error {
		atomic.AddInt32(&attemptCount, 1)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
			return nil
		}
	}
	timedTask := TaskWithTimeout(slowTask, 50*time.Millisecond)
	retryTask := TaskWithRetry(timedTask, 3, 10*time.Millisecond)

	err := wp.AddTask(retryTask)
	if err != nil {
		t.Fatalf("failed to add combined task: %v", err)
	}

	time.Sleep(300 * time.Millisecond)
	select {
	case err := <-wp.Errors():
		if err == nil {
			t.Error("expected timeout error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timeout waiting for error")
	}
	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}