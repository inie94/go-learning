package main

import (
	"compress/gzip"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"time"
	"strconv"
)

var (
	filePath      string
	configPath    string
	chunkSizeStr  string
	workers       int
	outputFile    string
	outputFormat  string
	errorKeywords string
	fromTimeStr   string
	toTimeStr     string
	timeFormat    string
	verbose       bool
)

func main() {
	parseFlags()
	cfg := loadConfig(configPath)
	applyConfig(cfg)

	// Открываем файл
	f, err := os.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка открытия файла: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	var readerAt io.ReaderAt = f
	var fileSize int64
	if strings.HasSuffix(filePath, ".gz") {
		// Для gzip используем обертку, читающую весь файл в память (упрощение)
		gzReader, err := gzip.NewReader(f)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка распаковки gzip: %v\n", err)
			os.Exit(1)
		}
		defer gzReader.Close()
		data, err := io.ReadAll(gzReader)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения gzip: %v\n", err)
			os.Exit(1)
		}
		readerAt = &bytesReaderAt{data: data}
		fileSize = int64(len(data))
	} else {
		stat, err := f.Stat()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка получения размера файла: %v\n", err)
			os.Exit(1)
		}
		fileSize = stat.Size()
	}

	chunkSize := parseSize(chunkSizeStr)
	if chunkSize <= 0 {
		chunkSize = 1 << 20 // 1MB
	}
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	errorKwList := strings.Split(errorKeywords, ",")
	for i := range errorKwList {
		errorKwList[i] = strings.TrimSpace(errorKwList[i])
	}

	var timeFilter *TimeFilter
	if fromTimeStr != "" && toTimeStr != "" {
		from, err := time.Parse(timeFormat, fromTimeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка парсинга fromTime: %v\n", err)
			os.Exit(1)
		}
		to, err := time.Parse(timeFormat, toTimeStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка парсинга toTime: %v\n", err)
			os.Exit(1)
		}
		timeFilter = &TimeFilter{From: from, To: to, Format: timeFormat}
	}

	// Создаём чанки
	chunks, err := createChunks(readerAt, fileSize, chunkSize)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка создания чанков: %v\n", err)
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("Создано %d чанков, размер файла %d байт\n", len(chunks), fileSize)
	}

	// Контекст для graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, os.Kill)
	go func() {
		<-sigChan
		fmt.Println("\nПолучен сигнал прерывания, завершаем обработку...")
		cancel()
	}()

	// Запуск
	startTime := time.Now()
	result := runProcessing(ctx, readerAt, chunks, workers, errorKwList, timeFilter, verbose)
	elapsed := time.Since(startTime)

	// Вывод метрик
	printMetrics(elapsed)

	// Вывод результатов
	outWriter := os.Stdout
	if outputFile != "" {
		outWriter, err = os.Create(outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка создания выходного файла: %v\n", err)
			os.Exit(1)
		}
		defer outWriter.Close()
	}

	err = writeOutput(outWriter, result)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка вывода: %v\n", err)
		os.Exit(1)
	}
}

func parseFlags() {
	flag.StringVar(&filePath, "file", "", "путь к файлу логов")
	flag.StringVar(&configPath, "config", "", "путь к JSON конфигурации")
	flag.StringVar(&chunkSizeStr, "chunk-size", "1MB", "размер чанка (например 1MB, 512KB)")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "количество воркеров")
	flag.StringVar(&outputFile, "output", "", "файл для сохранения результата")
	flag.StringVar(&outputFormat, "format", "text", "формат вывода: text, json, csv")
	flag.StringVar(&errorKeywords, "error-keywords", "error,ERROR,panic", "ключевые слова ошибок через запятую")
	flag.StringVar(&fromTimeStr, "from", "", "начало временного диапазона (ISO 8601)")
	flag.StringVar(&toTimeStr, "to", "", "конец временного диапазона (ISO 8601)")
	flag.StringVar(&timeFormat, "time-format", "2006-01-02T15:04:05", "формат временной метки в логе")
	flag.BoolVar(&verbose, "verbose", false, "выводить прогресс обработки")
	flag.Parse()

	if filePath == "" {
		fmt.Fprintln(os.Stderr, "Необходимо указать -file")
		os.Exit(1)
	}
}

func applyConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	if cfg.ChunkSize != "" {
		chunkSizeStr = cfg.ChunkSize
	}
	if cfg.Workers > 0 {
		workers = cfg.Workers
	}
	if cfg.ErrorKeywords != "" {
		errorKeywords = cfg.ErrorKeywords
	}
	if cfg.Output != "" {
		outputFile = cfg.Output
	}
	if cfg.OutputFormat != "" {
		outputFormat = cfg.OutputFormat
	}
	if cfg.Verbose {
		verbose = true
	}
	if cfg.FromTime != "" {
		fromTimeStr = cfg.FromTime
	}
	if cfg.ToTime != "" {
		toTimeStr = cfg.ToTime
	}
	if cfg.TimeFormat != "" {
		timeFormat = cfg.TimeFormat
	}
}

type bytesReaderAt struct {
	data []byte
}

func (b *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= int64(len(b.data)) {
		return 0, io.EOF
	}
	n = copy(p, b.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}

func parseSize(s string) int64 {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return 0
	}
	var multiplier int64 = 1
	if strings.HasSuffix(s, "TB") {
		multiplier = 1 << 40
		s = strings.TrimSuffix(s, "TB")
	} else if strings.HasSuffix(s, "GB") {
		multiplier = 1 << 30
		s = strings.TrimSuffix(s, "GB")
	} else if strings.HasSuffix(s, "MB") {
		multiplier = 1 << 20
		s = strings.TrimSuffix(s, "MB")
	} else if strings.HasSuffix(s, "KB") {
		multiplier = 1 << 10
		s = strings.TrimSuffix(s, "KB")
	} else if strings.HasSuffix(s, "B") {
		multiplier = 1
		s = strings.TrimSuffix(s, "B")
	}
	val, err := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	if err != nil {
		return 0
	}
	return val * multiplier
}