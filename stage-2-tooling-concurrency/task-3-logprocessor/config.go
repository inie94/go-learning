package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ChunkSize     string `json:"chunkSize"`
	Workers       int    `json:"workers"`
	ErrorKeywords string `json:"errorKeywords"`
	Output        string `json:"output"`
	OutputFormat  string `json:"outputFormat"`
	Verbose       bool   `json:"verbose`
	FromTime      string `json:"fromTime"`
	ToTime        string `json:"toTime"`
	TimeFormat    string `json:"timeFormat"`
}

func loadConfig(path string) *Config {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения конфига: %v\n", err)
		os.Exit(1)
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка парсинга конфига %v\n", err)
		os.Exit(1)
	}
	return &cfg
}
