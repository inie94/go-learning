package main

import "fmt"


type Calculator interface {
	Add(a,b int) int
	Sub(a,b int) int
}

type BasicCalculator struct {

}

func (c BasicCalculator) Add(a,b int) int {
	return a + b
}

func (c BasicCalculator) Sub(a,b int) int {
	return a - b
}

type SafeCalculator struct {
	Log []string
}

func (c *SafeCalculator) Add(a,b int) int {
	result := a + b
	c.Log = append(c.Log, fmt.Sprintf("Add(%d, %d) = %d", a, b, result))
	return result
}

func (c *SafeCalculator) Sub(a,b int) int {
	result := a - b
	c.Log = append(c.Log, fmt.Sprintf("Sub(%d, %d) = %d", a, b, result))
	return result
}

func ProcessCalculator(c Calculator) {
    // Проверяем, является ли c SafeCalc (указатель)
    if safe, ok := c.(*SafeCalculator); ok {
        fmt.Println("✅ Обнаружен SafeCalc!")
        fmt.Printf("📝 Лог операций: %v\n", safe.Log)
    }
    
    // Проверяем, является ли c BasicCalc (значение)
    if _, ok := c.(BasicCalculator); ok {
        fmt.Println("✅ Обнаружен BasicCalc!")
        fmt.Println("ℹ️  Логирование не поддерживается")
    }
    
    // Альтернативная проверка через switch (более элегантно)
    switch v := c.(type) {
    case *SafeCalculator:
        fmt.Printf("🔒 SafeCalc с логом из %d записей\n", len(v.Log))
    case BasicCalculator:
        fmt.Println("🔓 BasicCalc без логов")
    default:
        fmt.Println("❓ Неизвестный тип")
    }
}

func main() {
    // Создаём экземпляры
    basic := BasicCalculator{}
    safe := &SafeCalculator{Log: []string{}}
    
    // Используем калькуляторы
    fmt.Println("=== BasicCalc ===")
    fmt.Printf("Add(5,3) = %d\n", basic.Add(5, 3))
    fmt.Printf("Sub(10,4) = %d\n", basic.Sub(10, 4))
    
    fmt.Println("\n=== SafeCalc ===")
    fmt.Printf("Add(7,2) = %d\n", safe.Add(7, 2))
    fmt.Printf("Sub(15,6) = %d\n", safe.Sub(15, 6))
    
    fmt.Println("\n=== Проверка типов ===")
    ProcessCalculator(basic)  // передаём BasicCalc
    ProcessCalculator(safe)   // передаём SafeCalc
}