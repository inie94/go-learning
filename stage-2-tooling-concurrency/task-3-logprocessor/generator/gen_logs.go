package generator

import (
    "bufio"
    "fmt"
    "math/rand"
    "os"
    "time"
)

const (
    targetSize = 100 * 1024 * 1024 // 100 MB
    ips        = 50                // количество уникальных IP
)

var (
    ipPool []string
    statusCodes = []int{200, 201, 301, 400, 404, 500, 502, 503}
    errorKeywords = []string{"ERROR", "error", "panic", "fatal", "timeout"}
    methods = []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
    paths   = []string{"/", "/api", "/users", "/login", "/dashboard", "/profile", "/images", "/styles", "/js", "/health"}
)

func init() {
    rand.Seed(time.Now().UnixNano())
    // Генерируем пул IP-адресов
    for i := 0; i < ips; i++ {
        ipPool = append(ipPool, fmt.Sprintf("%d.%d.%d.%d",
            rand.Intn(255), rand.Intn(255), rand.Intn(255), rand.Intn(255)))
    }
}

func randomIP() string {
    return ipPool[rand.Intn(len(ipPool))]
}

func randomStatus() int {
    return statusCodes[rand.Intn(len(statusCodes))]
}

func randomMethod() string {
    return methods[rand.Intn(len(methods))]
}

func randomPath() string {
    return paths[rand.Intn(len(paths))]
}

func randomDate() string {
    t := time.Now().Add(-time.Duration(rand.Intn(30*24)) * time.Hour)
    return t.Format("02/Jan/2006:15:04:05 -0700")
}

func main() {
    file, err := os.Create("big_test.log")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    writer := bufio.NewWriterSize(file, 1024*1024) // 1MB буфер
    var written int64

    for written < targetSize {
        ip := randomIP()
        date := randomDate()
        method := randomMethod()
        path := randomPath()
        status := randomStatus()

        line := fmt.Sprintf("%s - - [%s] \"%s %s HTTP/1.1\" %d %d",
            ip, date, method, path, status, rand.Intn(4096)+64)

        // В 20% строк добавляем ошибку
        if rand.Float32() < 0.2 {
            errKeyword := errorKeywords[rand.Intn(len(errorKeywords))]
            line += fmt.Sprintf(" - %s: %s", errKeyword, randomErrorMessage())
        }

        line += "\n"
        n, _ := writer.WriteString(line)
        written += int64(n)

        // Периодически сбрасываем буфер для реалистичности
        if rand.Intn(100) == 0 {
            writer.Flush()
        }
    }
    writer.Flush()
    fmt.Printf("Создан файл big_test.log размером ~%d байт\n", written)
}

func randomErrorMessage() string {
    msgs := []string{
        "database connection lost",
        "timeout while reading",
        "invalid input",
        "permission denied",
        "resource exhausted",
        "unexpected EOF",
        "divide by zero",
        "nil pointer dereference",
    }
    return msgs[rand.Intn(len(msgs))]
}