package main

import (
	"fmt"
	"runtime"
	"time"
)

func printMetrics(elapsed time.Duration) {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fmt.Printf("Время выполнения: %v\n", elapsed)
	fmt.Printf("Пиковое использование памяти: %.2f MB\n", float64(memStats.Alloc)/1024/1024)
}