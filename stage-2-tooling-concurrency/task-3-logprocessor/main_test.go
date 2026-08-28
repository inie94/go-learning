package main

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestExtractIPs(t *testing.T) {
	line := `192.168.1.1 - - [01/Jan/2026:12:00:00] "GET /"`
	ips := extractIPs(line)
	if len(ips) != 1 || ips[0] != "192.168.1.1" {
		t.Errorf("expected [192.168.1.1], got %v", ips)
	}
	line2 := `IPv6: 2001:db8::1`
	ips2 := extractIPs(line2)
	if len(ips2) != 1 || ips2[0] != "2001:db8::1" {
		t.Errorf("expected [2001:db8::1], got %v", ips2)
	}
}

func TestContainsError(t *testing.T) {
	kw := []string{"error", "ERROR"}
	if !containsError("this is an error message", kw) {
		t.Error("expected true")
	}
	if containsError("ok message", kw) {
		t.Error("expected false")
	}
	if !containsError("ERROR", kw) {
		t.Error("expected true")
	}
}

func TestCheckTime(t *testing.T) {
	tf := &TimeFilter{
		From:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		To:     time.Date(2026, 1, 31, 23, 59, 59, 0, time.UTC),
		Format: "2006-01-02T15:04:05",
	}
	line := "2026-01-15T10:00:00 log message"
	if !checkTime(line, tf) {
		t.Error("expected true")
	}
	line2 := "2025-12-31T23:00:00 old"
	if checkTime(line2, tf) {
		t.Error("expected false")
	}
}

func TestIntegration(t *testing.T) {
	content := `192.168.1.1 - - [2026-01-01T10:00:00] "GET /" error
10.0.0.1 - - [2026-01-02T10:00:00] "POST /" ok
192.168.1.1 - - [2026-01-03T10:00:00] "GET /" panic`
	tmpFile, err := os.CreateTemp("", "logtest*.log")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatal(err)
	}
	tmpFile.Close()

	// Открываем файл для чтения
	f, err := os.Open(tmpFile.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	stat, _ := f.Stat()
	chunks, _ := createChunks(f, stat.Size(), 1024)
	kw := []string{"error", "panic"}
	res := runProcessing(context.Background(), f, chunks, 2, kw, nil, false)
	if res.TotalLines != 3 {
		t.Errorf("total lines expected 3, got %d", res.TotalLines)
	}
	if res.ErrorLines != 2 {
		t.Errorf("error lines expected 2, got %d", res.ErrorLines)
	}
	if res.IPCounts["192.168.1.1"] != 2 {
		t.Errorf("IP count for 192.168.1.1 expected 2, got %d", res.IPCounts["192.168.1.1"])
	}
	if res.IPCounts["10.0.0.1"] != 1 {
		t.Errorf("IP count for 10.0.0.1 expected 1, got %d", res.IPCounts["10.0.0.1"])
	}
}
