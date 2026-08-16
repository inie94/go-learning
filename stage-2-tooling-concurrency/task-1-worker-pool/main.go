// file: main.go (пример использования)
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"local/workerpool" // импортируем наш пакет
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
			log.Printf("An error occurred: %v", err)
		}
		log.Println("Error channel closed, reading complete.")
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
				log.Printf("Task %d completed\n", taskID)
				return nil
			}
		})
		if err != nil {
			log.Printf("Failed to add the task %d: %v", taskID, err)
		}
	}

	// Даём время на выполнение задач (в реальном коде можно использовать StopAndWait).
	time.Sleep(2 * time.Second)

	// Останавливаем пул с ожиданием завершения всех задач.
	pool.StopAndWait()
	log.Println("All task is completed, pool is stoped")
}
