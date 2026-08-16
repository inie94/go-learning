package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"strings"
)

// ParseConfig парсит конфигурационный файл в формате key=value
func ParseConfig(data []byte) (map[string]string, error) {
	m := make(map[string]string)
	
	if len(data) == 0 {
		return m, nil // Пустые данные - не ошибка
	}
	
	reader := bytes.NewReader(data)
	scanner := bufio.NewScanner(reader)
	lineNum := 0
	
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		
		// Удаляем комментарии
		if idx := strings.Index(line, "#"); idx != -1 {
			line = line[:idx]
		}
		
		// Удаляем пробелы по краям
		line = strings.TrimSpace(line)
		
		// Пропускаем пустые строки
		if line == "" {
			continue
		}
		
		// Ищем первый знак "="
		idx := strings.Index(line, "=")
		if idx == -1 {
			return nil, fmt.Errorf("line %d: missing '=' separator", lineNum)
		}
		
		// Извлекаем ключ и значение с удалением пробелов
		key := strings.TrimSpace(line[:idx])
		value := strings.TrimSpace(line[idx+1:])
		
		// Проверяем, что ключ не пустой
		if key == "" {
			return nil, fmt.Errorf("line %d: empty key", lineNum)
		}
		
		// Проверяем дубликаты
		if _, exists := m[key]; exists {
			return nil, fmt.Errorf("line %d: duplicate key '%s'", lineNum, key)
		}
		
		m[key] = value
	}
	
	// Проверяем ошибки сканера
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan error: %w", err)
	}
	
	return m, nil
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: program <config_file>")
	}
	
	data, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}
	
	config, err := ParseConfig(data)
	if err != nil {
		log.Fatalf("Failed to parse config: %v", err)
	}
	
	log.Printf("Parsed config: %+v", config)
}