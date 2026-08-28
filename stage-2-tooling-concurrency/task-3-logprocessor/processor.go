package main

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
)

type TimeFilter struct {
	From   time.Time
	To     time.Time
	Format string
}

type LocalResult struct {
	TotalLines int
	ErrorLines int
	IPCounts   map[string]int
}

func runProcessing(ctx context.Context, readerAt io.ReaderAt, chunks []Chunk, workers int, errorKw []string, TimeFilter *TimeFilter, verbose bool) *Result {
	chunkChan := make(chan Chunk, len(chunks))
	for _, ch := range chunks {
		chunkChan <- ch
	}
	close(chunkChan)

	resultChan := make(chan *LocalResult, workers)
	var wg sync.WaitGroup

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case ch, ok := <-chunkChan:
					if !ok {
						return
					}
					local := processChunk(readerAt, ch, errorKw, TimeFilter)
					select {
					case resultChan <- local:
					case <-ctx.Done():
						return
					}
					if verbose {
						fmt.Printf("Обработан чанк [%d, %d]\n", ch.Start, ch.Start+ch.Size)
					}
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	finalResult := &Result{
		TotalLines: 0,
		ErrorLines: 0,
		IPCounts:   make(map[string]int),
	}
	var mu sync.Mutex

	for local := range resultChan {
		mu.Lock()
		finalResult.TotalLines += local.TotalLines
		finalResult.ErrorLines += local.ErrorLines
		for ip, count := range local.IPCounts {
			finalResult.IPCounts[ip] += count
		}
		mu.Unlock()
	}
	return finalResult
}

func processChunk(readerAt io.ReaderAt, ch Chunk, errorKw []string, timeFilter *TimeFilter) *LocalResult {
	buf := make([]byte, ch.Size)
	_, err := readerAt.ReadAt(buf, ch.Start)
	if err != nil && err != io.EOF {
		// Логируем ошибку, но продолжаем
		fmt.Fprintf(os.Stderr, "Ошибка чтения чанка [%d, %d]: %v\n", ch.Start, ch.Start+ch.Size, err)
		return &LocalResult{IPCounts: make(map[string]int)}
	}
	scanner := bufio.NewScanner(bytes.NewReader(buf))
	local := &LocalResult{IPCounts: make(map[string]int)}
	for scanner.Scan() {
		line := scanner.Text()
		local.TotalLines++
		// Фильтр по времени
		if timeFilter != nil {
			ok := checkTime(line, timeFilter)
			if ok {
				continue
			}
		}
		// Поиск ошибки
		if containsError(line, errorKw) {
			local.ErrorLines++
		}
		// Извлечение IP
		ips := extractIPs(line)
		for _, ip := range ips {
			local.IPCounts[ip]++
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка сканирования чанка: %v\n", err)
	}
	return local
}

func containsError(line string, keywords []string) bool {
	lineLower := strings.ToLower(line)
	for _, kw := range keywords {
		if strings.Contains(lineLower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}

func extractIPs(line string) []string {
    // Находит все подстроки, состоящие из допустимых символов IP
    re := regexp.MustCompile(`[0-9a-fA-F:.]+`)
    matches := re.FindAllString(line, -1)
    var ips []string
    for _, m := range matches {
        if ip := net.ParseIP(m); ip != nil {
            ips = append(ips, m)
        }
    }
    return ips
}

func checkTime(line string, tf *TimeFilter) bool {
	// Ищем временную метрку в строке.
	// Но проще всего попытаться распарсить всю строку или искать подстроку с форматом.
	// Для простоты ищем первую подстроку, которая удовлетворяет формату.
	// Это неэффективно, но для демонстрации.
	// Можно использовать регулярное выражение, но оставим упрощённо.
	for i := 0; i < len(line)-len(tf.Format); i++ {
		sub := line[i : i+len(tf.Format)]
		t, err := time.Parse(tf.Format, sub)
		if err == nil {
			return (t.Equal(tf.From) || t.After(tf.From)) && (t.Equal(tf.To) || t.Before(tf.To))
		}
	}
	// Если не нашли метку - пропускаем строку (или считаем, что она не попадает в диапазон)
	// По условию, если метка отсутствует, мы не знаем, что делать. Лучше пропустить.
	return false
}
