# Этап 1: Основы языка Go (Junior)

> **Цель этапа** — научиться писать идиоматичный, работающий код на Go, понимая не только синтаксис, но и философию языка.

---

## 📚 Содержание

- [Структура программы и базовый синтаксис](#структура-программы-и-базовый-синтаксис)
- [Составные типы](#составные-типы)
- [Функции](#функции)
- [Интерфейсы](#интерфейсы-критически-важная-тема)
- [Обработка ошибок](#обработка-ошибок)
- [Практические задания](#практические-задания)
- [Контрольные вопросы](#контрольные-вопросы)
- [Рекомендуемые ресурсы](#рекомендуемые-ресурсы)

---

## Структура программы и базовый синтаксис

### Структура программы

Каждая программа на Go состоит из пакетов. Точка входа — пакет `main` с функцией `main()`.

```go
package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}
```

#### Ключевые моменты:

* Пакет — это способ организации кода. Каждый файл принадлежит одному пакету.
* Импорты группируются в блоки, можно задавать алиасы.
* Экспортируемые (публичные) идентификаторы начинаются с заглавной буквы.
* Неэкспортируемые (приватные) — со строчной.

### Переменные, константы, базовые типы

```go
// Объявление переменных
var name string = "Go"
var age int = 15
var version = "1.21" // вывод типа

// Короткая форма (только внутри функций)
lang := "Go"

// Константы
const Pi = 3.14159

// Базовые типы
var integer int = 42
var float float64 = 3.14
var boolean bool = true
var str string = "Hello"
```

#### Zero values (нулевые значения):

int → 0
float → 0.0
bool → false
string → ""
Указатели, слайсы, карты, каналы, функции, интерфейсы → nil
### Указатели: value vs pointer semantics

Go поддерживает указатели, но не арифметику указателей.

```go
func main() {
    x := 42
    p := &x       // указатель на x
    fmt.Println(*p) // разыменование: 42
    *p = 100      // изменение значения через указатель
    fmt.Println(x) // 100
}
```
#### Когда использовать:

* Value semantics — передача по значению (копирование) — безопасно, иммутабельно.
* Pointer semantics — передача по указателю — когда нужно изменить оригинал или большие структуры.

> c 📌 Идиома Go: предпочитайте value semantics, если нет явной причины использовать pointer.

---

## Составные типы

### Массивы (Arrays)

Массивы имеют фиксированную длину, которая является частью типа.

```go
var arr [5]int           // массив из 5 int, все 0
arr2 := [3]int{1, 2, 3}  // литерал
arr3 := [...]int{1, 2, 3} // длина выводится автоматически
```
### Срезы (Slices) — динамические массивы

Срезы — это гибкие, динамические структуры, ссылающиеся на underlying array.

```go
slice := []int{1, 2, 3}        // объявление
slice = append(slice, 4)       // добавление элементов
slice2 := make([]int, 5, 10)   // длина 5, емкость 10
``` 
Внутреннее устройство слайса:

```go
type slice struct {
    array unsafe.Pointer // указатель на underlying array
    len   int            // длина
    cap   int            // емкость
}
```
⚠️ Важно: при append, если cap недостаточен, создаётся новый массив, и ссылка меняется.
### Карты (Maps)

```go
m := make(map[string]int)
m["key"] = 42

// Литерал
m2 := map[string]int{
    "one": 1,
    "two": 2,
}

// Проверка существования ключа
val, ok := m2["three"]
if !ok {
    fmt.Println("key not found")
}
```
#### Особенности:

* Карты ссылочные, передаются по указателю неявно.
*Не потокобезопасны (используйте sync.Mutex или sync.Map для конкурентного доступа).
*Итерация по карте не гарантирует порядок.
### Структуры (Structs)

```go
type Person struct {
    Name string
    Age  int
    email string // приватное поле
}

// Теги (tags) — используются для сериализации
type User struct {
    ID   int    `json:"id"`
    Name string `json:"name,omitempty"`
}
```
#### Embedding (композиция вместо наследования):

```go
type Employee struct {
    Person      // встраивание (embedding)
    Company string
}

func main() {
    e := Employee{
        Person: Person{Name: "Alice", Age: 30},
        Company: "Google",
    }
    fmt.Println(e.Name) // доступ к полям Person напрямую
}
```

---

## Функции

### Сигнатуры и множественный возврат

```go
func add(a, b int) int {
    return a + b
}

// Множественный возврат
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}

// Именованные возвращаемые параметры
func getCoords() (x, y int) {
    x = 10
    y = 20
    return // naked return
}
```
### Вариативные параметры

```go
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

// sum(1, 2, 3) → 6
```
### Функции как значения. Замыкания.

```go
func multiplier(factor int) func(int) int {
    return func(x int) int {
        return x * factor
    }
}

func main() {
    double := multiplier(2)
    fmt.Println(double(5)) // 10
}
```
### Defer

defer откладывает выполнение функции до выхода из текущей.

```go
func main() {
    defer fmt.Println("world") // выполнится последним
    fmt.Println("hello")
}
// Вывод:
// hello
// world

// Аргументы вычисляются сразу
func main() {
    i := 1
    defer fmt.Println(i) // 1
    i = 2
    // Вывод: 1 (не 2)
}
```
> 📌 Порядок выполнения: LIFO (последний defer — первый выполняется).

---

## Интерфейсы (Критически важная тема)

### Неявная имплементация

В Go интерфейсы реализуются неявно — достаточно просто реализовать все методы.

```go
type Writer interface {
    Write([]byte) (int, error)
}

type File struct {
    name string
}

func (f File) Write(data []byte) (int, error) {
    // реализация
    return len(data), nil
}

// File реализует Writer автоматически
var w Writer = File{name: "file.txt"}
```
### Пустой интерфейс (interface{} / any)

Пустой интерфейс принимает значения любого типа.

```go
var any interface{}
any = 42
any = "hello"
any = struct{}{}
```
⚠️ Используйте осознанно: потеря безопасности типов.
### Внутреннее устройство interface values

Интерфейс в Go под капотом — это структура из двух указателей:

```text
type iface struct {
    tab  *itab          // информация о типе и таблице методов
    data unsafe.Pointer // указатель на данные
}
```

### Type assertions и Type switches

```go
var i interface{} = "hello"

// Type assertion
s, ok := i.(string)
if ok {
    fmt.Println(s) // "hello"
}

// Type switch
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("unknown type")
}
```

---

## Обработка ошибок

### Философия: ошибки — это значения

В Go нет исключений. Ошибки — это обычные значения, которые возвращаются явно.

```go
f, err := os.Open("file.txt")
if err != nil {
    // обработать ошибку
    return err
}
defer f.Close()
```

### Создание ошибок

```go
import "errors"

// Sentinel error (предопределённая ошибка)
var ErrNotFound = errors.New("not found")

// Кастомная ошибка с дополнительной информацией
type MyError struct {
    Code int
    Msg  string
}

func (e MyError) Error() string {
    return fmt.Sprintf("code %d: %s", e.Code, e.Msg)
}

// fmt.Errorf с %w для обёртывания
err := fmt.Errorf("failed to process: %w", originalErr)
```

### Проверка ошибок: errors.Is и errors.As

```go
// errors.Is — проверка на sentinel ошибку
if errors.Is(err, ErrNotFound) {
    // обработать
}

// errors.As — получение кастомной ошибки
var myErr *MyError
if errors.As(err, &myErr) {
    fmt.Println("Code:", myErr.Code)
}
```

---

## Практические задания

### 🧪 Задание 1: Управление пользователями

Реализуйте структуру User с полями ID, Name, Email. Напишите функции:

NewUser(name, email string) *User — конструктор с валидацией email.
Метод UpdateEmail(newEmail string) error — обновление с проверкой формата.
Функция GetUserByID(users []User, id int) (*User, error) — поиск, возвращает ошибку ErrNotFound при отсутствии.
### 🧪 Задание 2: Парсер конфигурации

Напишите функцию ParseConfig(data []byte) (map[string]string, error), которая парсит данные вида:

```text
key1=value1
key2=value2
# комментарий
```
И возвращает карту. Обработайте ошибки: пустые строки, отсутствие =, дубликаты ключей.

### 🧪 Задание 3: Калькулятор с интерфейсами

Определите интерфейс Calculator с методами Add(a, b int) int, Sub(a, b int) int. Реализуйте две структуры:

BasicCalc — обычный калькулятор.
SafeCalc — калькулятор, логирующий операции в слайс строк.
Используйте type assertion для проверки типа.

---

## Контрольные вопросы

* Чем отличаются массивы от срезов в Go?
* Что такое zero value и почему это важно?
* Когда использовать указатели, а когда — значения?
* Как работает встраивание (embedding) в структурах?
* В чём отличие interface{} от интерфейса с методами?
* Какая структура у интерфейсного значения под капотом?
* Как defer обрабатывает аргументы?
* Какие способы создания ошибок вы знаете?
* В чём разница между errors.Is и errors.As?
* Почему в Go нет исключений?

---

## Рекомендуемые ресурсы

### 📖 Обязательно к прочтению

A Tour of Go — https://go.dev/tour/
Effective Go — https://go.dev/doc/effective_go
Go by Example — https://gobyexample.com/
### 📚 Книги

«The Go Programming Language» (Donovan, Kernighan) — главы 1-5.
«Go 100 Mistakes and How to Avoid Them» (Harsanyi) — ошибки #1-#30.
### 🛠️ Инструменты

go fmt — автоматическое форматирование.
go vet — статический анализ кода.
golangci-lint — комплексный линтер.
### 🎯 Практика

Gophercises — https://gophercises.com/
Exercism Go Track — https://exercism.org/tracks/go
---
🎯 Результат этапа: вы пишете идиоматичный код на Go, понимаете структуру интерфейсов, работаете с ошибками как с обычными значениями и используете слайсы и карты осознанно.