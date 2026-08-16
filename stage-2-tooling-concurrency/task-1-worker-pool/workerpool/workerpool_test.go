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

func TestNewWorkerPool(t *testing.T) {
	wp := NewWorkerPool(3, 5)
	if wp.workersCount != 3 {
		t.Errorf("expected 3, got %d", wp.workersCount)
	}
	if cap(wp.taskQueue) != 5 {
		t.Errorf("expected cap 5, got %d", cap(wp.taskQueue))
	}
	if cap(wp.errorCh) != 5 {
		t.Errorf("expected cap 5, got %d", cap(wp.errorCh))
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
	// канал ошибок должен быть закрыт
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
			// Задача не проверяет ctx.Done() до завершения работы,
			// поэтому она всегда выполнится полностью.
			time.Sleep(100 * time.Millisecond)
			atomic.AddInt32(&completed, 1)
			return nil
		})
		if err != nil {
			t.Fatalf("failed to add task: %v", err)
		}
	}

	time.Sleep(20 * time.Millisecond)
	wp.Stop() // теперь Stop не отменяет контекст

	// Пытаемся добавить новую задачу – должно быть ошибкой.
	err := wp.AddTask(func(ctx context.Context) error { return nil })
	if err == nil {
		t.Error("expected error when adding task after Stop")
	}

	wg.Wait()
	if atomic.LoadInt32(&completed) != longTasks {
		t.Errorf("expected %d tasks completed, got %d", longTasks, completed)
	}

	wp.StopAndWait()
	// Проверяем, что канал ошибок закрыт.
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
	cancel() // отменяем контекст извне

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
	wp.StopAndWait() // повторный вызов не должен паниковать
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
	wp.Start(context.Background()) // второй вызов игнорируется
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