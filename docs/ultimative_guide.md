# Ультимативный гайд по Go: стиль, идиомы и практики (с примерами)

Этот гайд объединяет **Effective Go** и **Google Go Style Best Practices**. Дубликаты убраны, все примеры сохранены. Цель — помочь писать идиоматичный, читаемый и поддерживаемый код на Go.

---

## 1. Философия

Go — не C++ и не Java. Прямой перевод кода из этих языков редко даёт хороший результат. Чтобы писать на Go хорошо, нужно думать в терминах его идиом: простоты, композиции, явной обработки ошибок, конкурентности через каналы.

Исходники стандартной библиотеки — лучший источник примеров. Многие пакеты содержат исполняемые примеры.

---

## 2. Форматирование

- Используйте `gofmt` (или `go fmt`). Он приводит код к стандартному стилю.
- Отступы — табы. Пробелы — только при необходимости.
- Нет ограничения на длину строки. Если строка длинная — перенесите с дополнительным табом.
- Go требует меньше скобок: `if`, `for`, `switch` не имеют круглых скобок.
- Все стандартные пакеты отформатированы `gofmt`.

**Пример выравнивания комментариев:**

```go
type T struct {
    name string // name of the object
    value int // its value
}
```

`gofmt` выровняет столбцы:

```go
type T struct {
    name    string // name of the object
    value   int    // its value
}
```

**Приоритеты операторов:**

```
x<<8 + y<<16
```

означает то, что подразумевает пробел, в отличие от других языков.

---

## 3. Комментарии и документация

- `//` — строчные комментарии, норма. `/* */` — блочные, для пакетов или отключения кода.
- Doc-комментарии ставятся перед объявлением без пустой строки. Начинаются с имени сущности.
- Комментарий пакета описывает назначение пакета. Обычно в `doc.go`.
- Внутри функций комментируйте **почему**, а не **что**.
- Все экспортируемые сущности должны иметь doc-комментарий.

```go
// Good:
// Config содержит параметры конфигурации для клиента.
type Config struct {
    // ...
}

// NewConfig создаёт новый Config с разумными значениями по умолчанию.
func NewConfig() *Config {
    // ...
}
```

```go
// Package creditcard предоставляет операции с кредитными картами
// через внешних поставщиков платежей.
package creditcard
```

```go
// Плохо:
// Увеличиваем i на 1.
i++

// Хорошо:
// Компенсируем смещение, введённое предыдущей итерацией.
i++
```

---

## 4. Именование

### 4.1. Общие правила

- Заглавная первая буква — экспорт.
- Используйте `MixedCaps` или `mixedCaps`, а не подчёркивания.
- Имена должны быть короткими, лаконичными, выразительными.

### 4.2. Пакеты

- Имя пакета — нижний регистр, одно слово, без подчёркиваний и mixedCaps.
- Имя пакета становится аксессором: `bytes.Buffer`, `ring.New`.
- Избегайте имён `util`, `helper`, `common`.
- Имя пакета должно быть связано с тем, что он предоставляет.

```go
// Хорошо:
db := spannertest.NewDatabaseFromFile(...)
_, err := f.Seek(0, io.SeekStart)
b := elliptic.Marshal(curve, x, y)
```

```go
// Плохо:
db := test.NewDatabaseFromFile(...)
_, err := f.Seek(0, common.SeekStart)
b := helper.Marshal(curve, x, y)
```

Пример с пакетом `bufio`: тип называется `Reader`, а не `BufReader`, потому что пользователи видят `bufio.Reader`. Функция `NewRing` в пакете `ring` вызывается как `ring.New`. `once.Do` из `sync`, не `once.DoOrWaitUntilDone`.

### 4.3. Функции и методы

- Избегайте повторений: не повторяйте имя пакета, тип получателя, типы параметров, типы возвратов.
- Функции, возвращающие значение, называются существительными.
- Геттеры не используют `Get`: `Owner()`, а не `GetOwner()`. Сеттер — `SetOwner`.
- Функции-действия называются глаголами.

```go
// Плохо:
package yamlconfig

func ParseYAMLConfig(input string) (*Config, error)

// Хорошо:
package yamlconfig

func Parse(input string) (*Config, error)
```

```go
// Плохо:
func (c *Config) WriteConfigTo(w io.Writer) (int64, error)

// Хорошо:
func (c *Config) WriteTo(w io.Writer) (int64, error)
```

```go
// Плохо:
func OverrideFirstWithSecond(dest, source *Config) error

// Хорошо:
func Override(dest, source *Config) error
```

```go
// Плохо:
func TransformToJSON(input *Config) *jsonconfig.Config

// Хорошо:
func Transform(input *Config) *jsonconfig.Config
```

Когда нужно устранить неоднозначность:

```go
// Хорошо:
func (c *Config) WriteTextTo(w io.Writer) (int64, error)
func (c *Config) WriteBinaryTo(w io.Writer) (int64, error)
```

```go
// Хорошо:
func (c *Config) JobName(key string) (value string, ok bool)
```

```go
// Плохо:
func (c *Config) GetJobName(key string) (value string, ok bool)
```

```go
// Хорошо:
func (c *Config) WriteDetail(w io.Writer) (int64, error)
```

```go
// Хорошо:
func ParseInt(input string) (int, error)
func ParseInt64(input string) (int64, error)
func AppendInt(buf []byte, value int) []byte
func AppendInt64(buf []byte, value int64) []byte
```

```go
// Хорошо:
func (c *Config) Marshal() ([]byte, error)
func (c *Config) MarshalText() (string, error)
```

Геттеры в Go:

```go
owner := obj.Owner()
if owner != user {
    obj.SetOwner(user)
}
```

### 4.4. Интерфейсы

- Одно-методные интерфейсы: имя метода + `-er`: `Reader`, `Writer`, `Formatter`.
- Не используйте канонические имена (`Read`, `Write`, `Close`, `String`) иначе.
- Если ваш тип реализует `String`, называйте метод `String`, а не `ToString`.

### 4.5. Тестовые двойники

- Пакет: `creditcardtest`.
- Для одного типа — краткое имя: `Stub`.
- Для нескольких поведений — по поведению: `AlwaysCharges`, `AlwaysDeclines`.
- Для нескольких типов — явное имя: `StubService`, `StubStoredValue`.

```go
// Хорошо:
package creditcardtest
```

```go
// Хорошо:
import (
    "path/to/creditcard"
    "path/to/money"
)

// Stub заглушает creditcard.Service и не предоставляет собственного поведения.
type Stub struct{}

func (Stub) Charge(*creditcard.Card, money.Money) error { return nil }
```

```python
# Хорошо:
go_library(
    name = "creditcardtest",
    srcs = ["creditcardtest.go"],
    deps = [
        ":creditcard",
        ":money",
    ],
    testonly = True,
)
```

```go
// Хорошо:
// AlwaysCharges заглушает creditcard.Service и имитирует успех.
type AlwaysCharges struct{}

func (AlwaysCharges) Charge(*creditcard.Card, money.Money) error { return nil }

// AlwaysDeclines заглушает creditcard.Service и имитирует отклонённые
// списания.
type AlwaysDeclines struct{}

func (AlwaysDeclines) Charge(*creditcard.Card, money.Money) error {
    return creditcard.ErrDeclined
}
```

```go
// Хорошо:
type StubService struct{}

func (StubService) Charge(*creditcard.Card, money.Money) error { return nil }

type StubStoredValue struct{}

func (StubStoredValue) Credit(*creditcard.Card, money.Money) error { return nil }
```

```go
// Хорошо:
package payment

import "path/to/creditcardtest"

func TestProcessor(t *testing.T) {
    var spyCC creditcardtest.Spy
    proc := &Processor{CC: spyCC}

    // объявления опущены: card и amount
    if err := proc.Process(card, amount); err != nil {
        t.Errorf("proc.Process(card, amount) = %v, want nil", err)
    }

    charges := []creditcardtest.Charge{
        {Card: card, Amount: amount},
    }

    if got, want := spyCC.Charges, charges; !cmp.Equal(got, want) {
        t.Errorf("spyCC.Charges = %v, want %v", got, want)
    }
}
```

```go
// Плохо:
func TestProcessor(t *testing.T) {
    var cc creditcardtest.Spy
    proc := &Processor{CC: cc}

    if err := proc.Process(card, amount); err != nil {
        t.Errorf("proc.Process(card, amount) = %v, want nil", err)
    }

    charges := []creditcardtest.Charge{
        {Card: card, Amount: amount},
    }

    if got, want := cc.Charges, charges; !cmp.Equal(got, want) {
        t.Errorf("cc.Charges = %v, want %v", got, want)
    }
}
```

### 4.6. Затенение (shadowing)

```go
// Хорошо:
func abs(i int) int {
    if i < 0 {
        i *= -1
    }
    return i
}
```

```go
// Хорошо:
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    // Безусловно ограничиваем дедлайн для этой части обработки запроса.
    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()
    ctxlog.Info(ctx, "Capped deadline in inner request")
    // ...
}
```

```go
// Плохо:
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
        defer cancel()
        ctxlog.Info(ctx, "Capped deadline in inner request")
    }
    // ОШИБКА: здесь "ctx" снова означает контекст, предоставленный вызывающей стороной.
    // ...
}
```

```go
// Хорошо:
func (s *Server) innerHandler(ctx context.Context, req *pb.MyRequest) *pb.MyResponse {
    if *shortenDeadlines {
        var cancel func()
        // Обратите внимание на использование простого присваивания, =, а не :=.
        ctx, cancel = context.WithTimeout(ctx, 3*time.Second)
        defer cancel()
        ctxlog.Info(ctx, "Capped deadline in inner request")
    }
    // ...
}
```

```go
// Плохо:
func LongFunction() {
    url := "https://example.com/"
    // Ой, теперь мы не можем использовать net/url в коде ниже.
}
```

---

## 5. Управляющие структуры

### 5.1. Точки с запятой

Go вставляет точки с запятой автоматически. Правило: если последний токен перед новой строкой — идентификатор, литерал или `break continue fallthrough return ++ -- ) }`, вставляется `;`.

```go
    go func() { for { dst <- <-src } }()
```

Правильно:

```go
if i < f() {
    g()
}
```

Неправильно:

```go
if i < f()  // wrong!
{           // wrong!
    g()
}
```

### 5.2. If

```go
if x > 0 {
    return y
}
```

```go
if err := file.Chmod(0664); err != nil {
    log.Print(err)
    return err
}
```

```go
f, err := os.Open(name)
if err != nil {
    return err
}
codeUsing(f)
```

```go
f, err := os.Open(name)
if err != nil {
    return err
}
d, err := f.Stat()
if err != nil {
    f.Close()
    return err
}
codeUsing(f, d)
```

### 5.3. Переобъявление и переназначение

```go
f, err := os.Open(name)
d, err := f.Stat()
```

`err` переприсваивается, а не объявляется заново, потому что `d` — новая переменная.

### 5.4. For

```go
// Like a C for
for init; condition; post { }

// Like a C while
for condition { }

// Like a C for(;;)
for { }
```

```go
sum := 0
for i := 0; i < 10; i++ {
    sum += i
}
```

```go
for key, value := range oldMap {
    newMap[key] = value
}
```

```go
for key := range m {
    if key.expired() {
        delete(m, key)
    }
}
```

```go
sum := 0
for _, value := range array {
    sum += value
}
```

```go
for pos, char := range "日本\x80語" { // \x80 is an illegal UTF-8 encoding
    fmt.Printf("character %#U starts at byte position %d\n", char, pos)
}
```

Вывод:

```
character U+65E5 '日' starts at byte position 0
character U+672C '本' starts at byte position 3
character U+FFFD '�' starts at byte position 6
character U+8A9E '語' starts at byte position 7
```

```go
// Reverse a
for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
    a[i], a[j] = a[j], a[i]
}
```

### 5.5. Switch

```go
func unhex(c byte) byte {
    switch {
    case '0' <= c && c <= '9':
        return c - '0'
    case 'a' <= c && c <= 'f':
        return c - 'a' + 10
    case 'A' <= c && c <= 'F':
        return c - 'A' + 10
    }
    return 0
}
```

```go
func shouldEscape(c byte) bool {
    switch c {
    case ' ', '?', '&', '=', '#', '+', '%':
        return true
    }
    return false
}
```

```go
Loop:
    for n := 0; n < len(src); n += size {
        switch {
        case src[n] < sizeOne:
            if validateOnly {
                break
            }
            size = 1
            update(src[n])
        case src[n] < sizeTwo:
            if n+1 >= len(src) {
                err = errShortInput
                break Loop
            }
            if validateOnly {
                break
            }
            size = 2
            update(src[n] + src[n+1]<<shift)
        }
    }
```

```go
// Compare returns an integer comparing the two byte slices,
// lexicographically.
// The result will be 0 if a == b, -1 if a < b, and +1 if a > b
func Compare(a, b []byte) int {
    for i := 0; i < len(a) && i < len(b); i++ {
        switch {
        case a[i] > b[i]:
            return 1
        case a[i] < b[i]:
            return -1
        }
    }
    switch {
    case len(a) > len(b):
        return 1
    case len(a) < len(b):
        return -1
    }
    return 0
}
```

### 5.6. Type switch

```go
var t interface{}
t = functionOfSomeType()
switch t := t.(type) {
default:
    fmt.Printf("unexpected type %T\n", t)     // %T prints whatever type t has
case bool:
    fmt.Printf("boolean %t\n", t)             // t has type bool
case int:
    fmt.Printf("integer %d\n", t)             // t has type int
case *bool:
    fmt.Printf("pointer to boolean %t\n", *t) // t has type *bool
case *int:
    fmt.Printf("pointer to integer %d\n", *t) // t has type *int
}
```

---

## 6. Функции

### 6.1. Множественные возвращаемые значения

```go
func (file *File) Write(b []byte) (n int, err error)
```

```go
func nextInt(b []byte, i int) (int, int) {
    for ; i < len(b) && !isDigit(b[i]); i++ {
    }
    x := 0
    for ; i < len(b) && isDigit(b[i]); i++ {
        x = x*10 + int(b[i]) - '0'
    }
    return x, i
}
```

```go
for i := 0; i < len(b); {
    x, i = nextInt(b, i)
    fmt.Println(x)
}
```

### 6.2. Именованные параметры результата

```go
func nextInt(b []byte, pos int) (value, nextPos int) {
```

```go
func ReadFull(r Reader, buf []byte) (n int, err error) {
    for len(buf) > 0 && err == nil {
        var nr int
        nr, err = r.Read(buf)
        n += nr
        buf = buf[nr:]
    }
    return
}
```

### 6.3. Defer

```go
// Contents returns the file's contents as a string.
func Contents(filename string) (string, error) {
    f, err := os.Open(filename)
    if err != nil {
        return "", err
    }
    defer f.Close()  // f.Close will run when we're finished.

    var result []byte
    buf := make([]byte, 100)
    for {
        n, err := f.Read(buf[0:])
        result = append(result, buf[0:n]...) // append is discussed later.
        if err != nil {
            if err == io.EOF {
                break
            }
            return "", err  // f will be closed if we return here.
        }
    }
    return string(result), nil // f will be closed if we return here.
}
```

```go
for i := 0; i < 5; i++ {
    defer fmt.Printf("%d ", i)
}
```

Выводит `4 3 2 1 0`.

```go
func trace(s string)   { fmt.Println("entering:", s) }
func untrace(s string) { fmt.Println("leaving:", s) }

// Use them like this:
func a() {
    trace("a")
    defer untrace("a")
    // do something....
}
```

```go
func trace(s string) string {
    fmt.Println("entering:", s)
    return s
}

func un(s string) {
    fmt.Println("leaving:", s)
}

func a() {
    defer un(trace("a"))
    fmt.Println("in a")
}

func b() {
    defer un(trace("b"))
    fmt.Println("in b")
    a()
}

func main() {
    b()
}
```

Вывод:

```
entering: b
in b
entering: a
in a
leaving: a
leaving: b
```

### 6.4. Вариативные функции

```go
func Printf(format string, v ...interface{}) (n int, err error) {
```

```go
// Println prints to the standard logger in the manner of fmt.Println.
func Println(v ...interface{}) {
    std.Output(2, fmt.Sprintln(v...))  // Output takes parameters (int, string)
}
```

```go
// Min returns the smallest of the integers in its argument list.
func Min(a ...int) int {
    min := int(^uint(0) >> 1)  // largest int
    for _, i := range a {
        if i < min {
            min = i
        }
    }
    return min
}
```

### 6.5. Append

```go
func append(slice []T, elements ...T) []T
```

```go
x := []int{1,2,3}
x = append(x, 4, 5, 6)
fmt.Println(x)
```

Выводит `[1 2 3 4 5 6]`.

```go
x := []int{1,2,3}
y := []int{4,5,6}
x = append(x, y...)
fmt.Println(x)
```

---

## 7. Данные

### 7.1. new и make

```go
type SyncedBuffer struct {
    lock    sync.Mutex
    buffer  bytes.Buffer
}
```

```go
p := new(SyncedBuffer)  // type *SyncedBuffer
var v SyncedBuffer      // type  SyncedBuffer
```

```go
make([]int, 10, 100)
```

```go
var p *[]int = new([]int)       // allocates slice structure; *p == nil; rarely useful
var v  []int = make([]int, 100) // the slice v now refers to a new array of 100 ints

// Unnecessarily complex:
var p *[]int = new([]int)
*p = make([]int, 100, 100)

// Idiomatic:
v := make([]int, 100)
```

### 7.2. Конструкторы и составные литералы

```go
func NewFile(fd int, name string) *File {
    if fd < 0 {
        return nil
    }
    f := new(File)
    f.fd = fd
    f.name = name
    f.dirinfo = nil
    f.nepipe = 0
    return f
}
```

```go
func NewFile(fd int, name string) *File {
    if fd < 0 {
        return nil
    }
    f := File{fd, name, nil, 0}
    return &f
}
```

```go
    return &File{fd, name, nil, 0}
```

```go
    return &File{fd: fd, name: name}
```

```go
a := [...]string   {Enone: "no error", Eio: "Eio", Einval: "invalid argument"}
s := []string      {Enone: "no error", Eio: "Eio", Einval: "invalid argument"}
m := map[int]string{Enone: "no error", Eio: "Eio", Einval: "invalid argument"}
```

### 7.3. Массивы

```go
func Sum(a *[3]float64) (sum float64) {
    for _, v := range *a {
        sum += v
    }
    return
}

array := [...]float64{7.0, 8.5, 9.1}
x := Sum(&array)  // Note the explicit address-of operator
```

### 7.4. Срезы

```go
func Append(slice, data []byte) []byte {
    l := len(slice)
    if l + len(data) > cap(slice) {  // reallocate
        // Allocate double what's needed, for future growth.
        newSlice := make([]byte, (l+len(data))*2)
        // The copy function is predeclared and works for any slice type.
        copy(newSlice, slice)
        slice = newSlice
    }
    slice = slice[0:l+len(data)]
    copy(slice[l:], data)
    return slice
}
```

### 7.5. Двумерные срезы

```go
type Transform [3][3]float64  // A 3x3 array, really an array of arrays.
type LinesOfText [][]byte     // A slice of byte slices.
```

```go
text := LinesOfText{
    []byte("Now is the time"),
    []byte("for all good gophers"),
    []byte("to bring some fun to the party."),
}
```

```go
// Allocate the top-level slice.
picture := make([][]uint8, YSize) // One row per unit of y.
// Loop over the rows, allocating the slice for each row.
for i := range picture {
    picture[i] = make([]uint8, XSize)
}
```

```go
// Allocate the top-level slice, the same as before.
picture := make([][]uint8, YSize) // One row per unit of y.
// Allocate one large slice to hold all the pixels.
pixels := make([]uint8, XSize*YSize) // Has type []uint8 even though picture is [][]uint8.
// Loop over the rows, slicing each row from the front of the remaining pixels slice.
for i := range picture {
    picture[i], pixels = pixels[:XSize], pixels[XSize:]
}
```

### 7.6. Карты

```go
var timeZone = map[string]int{
    "UTC":  0*60*60,
    "EST": -5*60*60,
    "CST": -6*60*60,
    "MST": -7*60*60,
    "PST": -8*60*60,
}
```

```go
offset := timeZone["EST"]
```

```go
attended := map[string]bool{
    "Ann": true,
    "Joe": true,
    ...
}

if attended[person] { // will be false if person is not in the map
    fmt.Println(person, "was at the meeting")
}
```

```go
var seconds int
var ok bool
seconds, ok = timeZone[tz]
```

```go
func offset(tz string) int {
    if seconds, ok := timeZone[tz]; ok {
        return seconds
    }
    log.Println("unknown time zone:", tz)
    return 0
}
```

```go
_, present := timeZone[tz]
```

```go
delete(timeZone, "PST")
```

### 7.7. Печать

```go
fmt.Printf("Hello %d\n", 23)
fmt.Fprint(os.Stdout, "Hello ", 23, "\n")
fmt.Println("Hello", 23)
fmt.Println(fmt.Sprint("Hello ", 23))
```

```go
var x uint64 = 1<<64 - 1
fmt.Printf("%d %x; %d %x\n", x, x, int64(x), int64(x))
```

Вывод:

```
18446744073709551615 ffffffffffffffff; -1 -1
```

```go
fmt.Printf("%v\n", timeZone)  // or just fmt.Println(timeZone)
```

Вывод:

```
map[CST:-21600 EST:-18000 MST:-25200 PST:-28800 UTC:0]
```

```go
type T struct {
    a int
    b float64
    c string
}
t := &T{ 7, -2.35, "abc\tdef" }
fmt.Printf("%v\n", t)
fmt.Printf("%+v\n", t)
fmt.Printf("%#v\n", t)
fmt.Printf("%#v\n", timeZone)
```

Вывод:

```
&{7 -2.35 abc   def}
&{a:7 b:-2.35 c:abc     def}
&main.T{a:7, b:-2.35, c:"abc\tdef"}
map[string]int{"CST":-21600, "EST":-18000, "MST":-25200, "PST":-28800, "UTC":0}
```

```go
func (t *T) String() string {
    return fmt.Sprintf("%d/%g/%q", t.a, t.b, t.c)
}
fmt.Printf("%v\n", t)
```

Вывод:

```
7/-2.35/"abc\tdef"
```

Рекурсивная ошибка:

```go
type MyString string
func (m MyString) String() string {
    return fmt.Sprintf("MyString=%s", m) // Error: will recur forever.
}
```

Исправление:

```go
type MyString string
func (m MyString) String() string {
    return fmt.Sprintf("MyString=%s", string(m)) // OK: note conversion.
}
```

---

## 8. Инициализация

### 8.1. Константы

```go
type ByteSize float64

const (
    _           = iota // ignore first value by assigning to blank identifier
    KB ByteSize = 1 << (10 * iota)
    MB
    GB
    TB
    PB
    EB
    ZB
    YB
)
```

```go
func (b ByteSize) String() string {
    switch {
    case b >= YB:
        return fmt.Sprintf("%.2fYB", b/YB)
    case b >= ZB:
        return fmt.Sprintf("%.2fZB", b/ZB)
    case b >= EB:
        return fmt.Sprintf("%.2fEB", b/EB)
    case b >= PB:
        return fmt.Sprintf("%.2fPB", b/PB)
    case b >= TB:
        return fmt.Sprintf("%.2fTB", b/TB)
    case b >= GB:
        return fmt.Sprintf("%.2fGB", b/GB)
    case b >= MB:
        return fmt.Sprintf("%.2fMB", b/MB)
    case b >= KB:
        return fmt.Sprintf("%.2fKB", b/KB)
    }
    return fmt.Sprintf("%.2fB", b)
}
```

`YB` печатает как `1.00YB`, `ByteSize(1e13)` — как `9.09TB`.

### 8.2. Переменные

```go
var (
    home   = os.Getenv("HOME")
    user   = os.Getenv("USER")
    gopath = os.Getenv("GOPATH")
)
```

### 8.3. Функция init

```go
func init() {
    if user == "" {
        log.Fatal("$USER not set")
    }
    if home == "" {
        home = "/home/" + user
    }
    if gopath == "" {
        gopath = home + "/go"
    }
    // gopath may be overridden by --gopath flag on command line.
    flag.StringVar(&gopath, "gopath", gopath, "override default GOPATH")
}
```

---

## 9. Методы

### 9.1. Указатели против значений

```go
type ByteSlice []byte

func (slice ByteSlice) Append(data []byte) []byte {
    // Body exactly the same as the Append function defined above.
}
```

```go
func (p *ByteSlice) Append(data []byte) {
    slice := *p
    // Body as above, without the return.
    *p = slice
}
```

```go
func (p *ByteSlice) Write(data []byte) (n int, err error) {
    slice := *p
    // Again as above.
    *p = slice
    return len(data), nil
}
```

```go
    var b ByteSlice
    fmt.Fprintf(&b, "This hour has %d days\n", 7)
```

Правило: методы значений могут вызываться как для указателей, так и для значений; методы указателей — только для указателей. Для адресуемых значений Go автоматически берёт адрес: `b.Write` → `(&b).Write`.

---

## 10. Интерфейсы и встраивание

### 10.1. Интерфейсы

```go
type Sequence []int

// Methods required by sort.Interface.
func (s Sequence) Len() int {
    return len(s)
}
func (s Sequence) Less(i, j int) bool {
    return s[i] < s[j]
}
func (s Sequence) Swap(i, j int) {
    s[i], s[j] = s[j], s[i]
}

// Copy returns a copy of the Sequence.
func (s Sequence) Copy() Sequence {
    copy := make(Sequence, 0, len(s))
    return append(copy, s...)
}

// Method for printing - sorts the elements before printing.
func (s Sequence) String() string {
    s = s.Copy() // Make a copy; don't overwrite argument.
    sort.Sort(s)
    str := "["
    for i, elem := range s { // Loop is O(N²); will fix that in next example.
        if i > 0 {
            str += " "
        }
        str += fmt.Sprint(elem)
    }
    return str + "]"
}
```

```go
func (s Sequence) String() string {
    s = s.Copy()
    sort.Sort(s)
    return fmt.Sprint([]int(s))
}
```

```go
type Sequence []int

// Method for printing - sorts the elements before printing
func (s Sequence) String() string {
    s = s.Copy()
    sort.IntSlice(s).Sort()
    return fmt.Sprint([]int(s))
}
```

### 10.2. Преобразования интерфейсов и утверждения типа

```go
type Stringer interface {
    String() string
}

var value interface{} // Value provided by caller.
switch str := value.(type) {
case string:
    return str
case Stringer:
    return str.String()
}
```

```go
value.(typeName)
```

```go
str := value.(string)
```

```go
str, ok := value.(string)
if ok {
    fmt.Printf("string value is: %q\n", str)
} else {
    fmt.Printf("value is not a string\n")
}
```

```go
if str, ok := value.(string); ok {
    return str
} else if str, ok := value.(Stringer); ok {
    return str.String()
}
```

### 10.3. Общность

```go
type Block interface {
    BlockSize() int
    Encrypt(dst, src []byte)
    Decrypt(dst, src []byte)
}

type Stream interface {
    XORKeyStream(dst, src []byte)
}
```

```go
// NewCTR returns a Stream that encrypts/decrypts using the given Block in
// counter mode. The length of iv must be the same as the Block's block size.
func NewCTR(block Block, iv []byte) Stream
```

### 10.4. Встраивание

```go
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}
```

```go
// ReadWriter is the interface that combines the Reader and Writer interfaces.
type ReadWriter interface {
    Reader
    Writer
}
```

```go
// ReadWriter stores pointers to a Reader and a Writer.
// It implements io.ReadWriter.
type ReadWriter struct {
    *Reader  // *bufio.Reader
    *Writer  // *bufio.Writer
}
```

```go
type ReadWriter struct {
    reader *Reader
    writer *Writer
}
```

```go
func (rw *ReadWriter) Read(p []byte) (n int, err error) {
    return rw.reader.Read(p)
}
```

```go
type Job struct {
    Command string
    *log.Logger
}
```

```go
job.Println("starting now...")
```

```go
func NewJob(command string, logger *log.Logger) *Job {
    return &Job{command, logger}
}
```

```go
job := &Job{command, log.New(os.Stderr, "Job: ", log.Ldate)}
```

```go
func (job *Job) Printf(format string, args ...interface{}) {
    job.Logger.Printf("%q: %s", job.Command, fmt.Sprintf(format, args...))
}
```

---

## 11. Обработка ошибок

### 11.1. Sentinel-ошибки

```go
type Animal string

var (
    // ErrDuplicate возникает, если это животное уже было замечено.
    ErrDuplicate = errors.New("duplicate")

    // ErrMarsupial возникает, потому что у нас аллергия на сумчатых за пределами Австралии.
    // Извините.
    ErrMarsupial = errors.New("marsupials are not supported")
)

func process(animal Animal) error {
    switch {
    case seen[animal]:
        return ErrDuplicate
    case marsupial(animal):
        return ErrMarsupial
    }
    seen[animal] = true
    // ...
    return nil
}
```

```go
// Хорошо:
func handlePet(...) {
    switch err := process(an); err {
    case ErrDuplicate:
        return fmt.Errorf("feed %q: %v", an, err)
    case ErrMarsupial:
        // Пытаемся восстановиться с помощью друга вместо этого.
        alternate = an.BackupAnimal()
        return handlePet(..., alternate, ...)
    }
}
```

```go
// Хорошо:
func handlePet(...) {
    switch err := process(an); {
    case errors.Is(err, ErrDuplicate):
        return fmt.Errorf("feed %q: %v", an, err)
    case errors.Is(err, ErrMarsupial):
        // ...
    }
}
```

```go
// Плохо:
func handlePet(...) {
    err := process(an)
    if regexp.MatchString(`duplicate`, err.Error()) {...}
    if regexp.MatchString(`marsupial`, err.Error()) {...}
}
```

### 11.2. Добавление информации к ошибкам

```go
// Хорошо:
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("launch codes unavailable: %v", err)
}

// Вывод:
//
// launch codes unavailable: open settings.txt: no such file or directory
```

```go
// Плохо:
if err := os.Open("settings.txt"); err != nil {
    return fmt.Errorf("could not open settings.txt: %v", err)
}

// Вывод:
//
// could not open settings.txt: open settings.txt: no such file or directory
```

```go
// Плохо:
return fmt.Errorf("failed: %v", err)

// просто верните err вместо этого
```

### 11.3. %v vs %w

```go
// Хорошо:
func (*FortuneTeller) SuggestFortune(context.Context, *pb.SuggestionRequest) (*pb.SuggestionResponse, error) {
    // ...
    if err != nil {
        return nil, fmt.Errorf("couldn't find fortune database: %v", err)
    }
}
```

```go
// Хорошо:
import (
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func (*FortuneTeller) SuggestFortune(context.Context, *pb.SuggestionRequest) (*pb.SuggestionResponse, error) {
    // ...
    if err != nil {
        // Или используйте fmt.Errorf с глаголом %w, если намеренно обёртываете
        // ошибку, которую вызывающая сторона должна развернуть.
        return nil, status.Errorf(codes.Internal, "couldn't find fortune database", status.ErrInternal)
    }
}
```

```go
// Хорошо:
func (s *Server) internalFunction(ctx context.Context) error {
    // ...
    if err != nil {
        return fmt.Errorf("couldn't find remote file: %w", err)
    }
}
```

### 11.4. Размещение %w

```go
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("err2: %w", err1)
err3 := fmt.Errorf("err3: %w", err2)
```

```go
// Хорошо:
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("err2: %w", err1)
err3 := fmt.Errorf("err3: %w", err2)
fmt.Println(err3) // err3: err2: err1
// err3 — это цепочка ошибок от новой к старой, которая печатает от новой к старой.
```

```go
// Плохо:
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("%w: err2", err1)
err3 := fmt.Errorf("%w: err3", err2)
fmt.Println(err3) // err1: err2: err3
// err3 — это цепочка ошибок от новой к старой, которая печатает от старой к новой.
```

```go
// Плохо:
err1 := fmt.Errorf("err1")
err2 := fmt.Errorf("err2-1 %w err2-2", err1)
err3 := fmt.Errorf("err3-1 %w err3-2", err2)
fmt.Println(err3) // err3-1 err2-1 err1 err2-2 err3-2
// err3 — это цепочка ошибок от новой к старой, которая не печатает ни от новой к старой,
// ни от старой к новой.
```

### 11.5. Размещение sentinel-ошибок

```go
// Хорошо:
package parser

var ErrParse = fmt.Errorf("parse error")

// Это ещё одна ошибка пакета, которая может быть возвращена.
var ErrParseInvalidHeader = fmt.Errorf("%w: invalid header", ErrParse)

func parseHeader() error {
    err := checkHeader()
    return fmt.Errorf("%w: invalid character in header: %v", ErrParseInvalidHeader, err)
}

err := fmt.Errorf("%w: couldn't find fortune database: %v", ErrInternal, err)
```

```go
// Плохо:
package parser

var ErrParse = fmt.Errorf("parse error")

// Это ещё одна ошибка пакета, которая может быть возвращена.
var ErrParseInvalidHeader = fmt.Errorf("%w: invalid header", ErrParse)

func parseHeader() error {
    err := checkHeader()
    return fmt.Errorf("invalid character in header: %v: %w", err, ErrParseInvalidHeader)
}

var ErrInternal = status.Error(codes.Internal, "internal")
err2 := fmt.Errorf("couldn't find fortune database: %v: %w", err, ErrInternal)
```

### 11.6. Логирование ошибок

```go
// Плохо:
func (s *Server) handleRequest(req *pb.Request) (*pb.Response, error) {
    resp, err := s.doWork(req)
    if err != nil {
        log.Printf("handleRequest: %v", err) // Логируем здесь...
        return nil, err
    }
    return resp, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ...
    resp, err := s.handleRequest(req)
    if err != nil {
        log.Printf("ServeHTTP: %v", err) // ...и снова здесь.
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

```go
// Хорошо:
func (s *Server) handleRequest(req *pb.Request) (*pb.Response, error) {
    resp, err := s.doWork(req)
    if err != nil {
        // Просто возвращаем ошибку с контекстом.
        return nil, fmt.Errorf("do work: %w", err)
    }
    return resp, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // ...
    resp, err := s.handleRequest(req)
    if err != nil {
        log.Printf("handleRequest: %v", err) // Единственное место логирования.
        http.Error(w, "internal error", http.StatusInternalServerError)
        return
    }
    // ...
}
```

### 11.7. Обработка ошибки один раз

```go
// Плохо:
if err != nil {
    log.Printf("could not open file: %v", err)
    return err
}
```

```go
// Хорошо (вариант 1 — поглощение):
if err != nil {
    log.Printf("could not open file: %v", err)
    // не возвращаем; обрабатываем здесь
}
```

```go
// Хорошо (вариант 2 — распространение):
if err != nil {
    return fmt.Errorf("open config: %w", err)
}
```

### 11.8. Имена и строки ошибок

```go
// Хорошо:
var (
    ErrBrokenLink   = errors.New("link is broken")
    ErrCouldNotOpen = errors.New("could not open")
)

// Тип ошибки, а не значение, обычно получает суффикс Error.
type NotFoundError struct {
    // ...
}
```

```go
// Хорошо:
var ErrBrokenLink = errors.New("link is broken")

// Плохо:
var ErrBrokenLink = errors.New("Link is broken.")
var ErrBrokenLink = errors.New("Link is broken")
```

### 11.9. Type assertions

```go
// Плохо:
func process(v interface{}) {
    s := v.(string) // паникует, если v не строка
    // ...
}
```

```go
// Хорошо:
func process(v interface{}) {
    s, ok := v.(string)
    if !ok {
        // обработка ошибки
        return
    }
    // ...
}
```

### 11.10. Не паниковать

```go
// Плохо:
func divide(a, b int) int {
    if b == 0 {
        panic("division by zero")
    }
    return a / b
}
```

```go
// Хорошо:
func divide(a, b int) (int, error) {
    if b == 0 {
        return 0, errors.New("division by zero")
    }
    return a / b, nil
}
```

### 11.11. Must-функции

```go
// Хорошо (инициализация на уровне пакета):
var validID = regexp.MustCompile(`^[a-z]+\[[0-9]+\]$`)

// Хорошо (тест):
func TestSomething(t *testing.T) {
    tmpl := template.Must(template.New("test").Parse("Hello, {{.}}"))
    // ...
}
```

```go
// Плохо:
func parseUser(input string) *User {
    // Паникует при плохом вводе.
    return template.Must(...) 
}
```

### 11.12. Паника и восстановление

```go
panic("something wrong")
```

```go
func protect(g func()) {
    defer func() {
        log.Println("done")  // Println executes normally even if there is a panic
        if x := recover(); x != nil {
            log.Printf("run time panic: %v", x)
        }
    }()
    log.Println("start")
    g()
}
```

---

## 12. Конкурентность

### 12.1. Горутины

```go
go list.Sort()  // run list.Sort concurrently; don't wait for it.
```

```go
func Announce(message string, delay time.Duration) {
    go func() {
        time.Sleep(delay)
        fmt.Println(message)
    }()  // Note the parentheses - must call the function.
}
```

```go
// Хорошо:
func (s *Server) startWorker(ctx context.Context) {
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            case work := <-s.work:
                s.process(work)
            }
        }
    }()
}
```

```go
// Плохо:
func doSomething() {
    go func() {
        for {
            // Никогда не завершается — это утечка.
            doWork()
        }
    }()
}
```

### 12.2. Каналы

```go
ci := make(chan int)            // unbuffered channel of integers
cj := make(chan int, 0)         // unbuffered channel of integers
cs := make(chan *os.File, 100)  // buffered channel of pointers to Files
```

```go
c := make(chan int)  // Allocate a channel.
// Start the sort in a goroutine; when it completes, signal on the channel.
go func() {
    list.Sort()
    c <- 1  // Send a signal; value does not matter.
}()
doSomethingForAWhile()
<-c   // Wait for sort to finish; discard sent value.
```

### 12.3. Семафор

```go
var sem = make(chan int, MaxOutstanding)

func handle(r *Request) {
    sem <- 1    // Wait for active queue to drain.
    process(r)  // May take a long time.
    <-sem       // Done; enable next request to run.
}

func Serve(queue chan *Request) {
    for {
        req := <-queue
        go handle(req)  // Don't wait for handle to finish.
    }
}
```

### 12.4. Ошибка с переменной цикла

```go
func Serve(queue chan *Request) {
    for req := range queue {
        sem <- 1
        go func() {
            process(req) // Buggy; see explanation below.
            <-sem
        }()
    }
}
```

Исправление — передать аргумент:

```go
func Serve(queue chan *Request) {
    for req := range queue {
        sem <- 1
        go func(req *Request) {
            process(req)
            <-sem
        }(req)
    }
}
```

Или создать новую переменную:

```go
func Serve(queue chan *Request) {
    for req := range queue {
        req := req // Create new instance of req for the goroutine.
        sem <- 1
        go func() {
            process(req)
            <-sem
        }()
    }
}
```

### 12.5. Фиксированное количество горутин

```go
func handle(queue chan *Request) {
    for r := range queue {
        process(r)
    }
}

func Serve(clientRequests chan *Request, quit chan bool) {
    // Start handlers
    for i := 0; i < MaxOutstanding; i++ {
        go handle(clientRequests)
    }
    <-quit  // Wait to be told to exit.
}
```

### 12.6. Каналы каналов

```go
type Request struct {
    args        []int
    f           func([]int) int
    resultChan  chan int
}
```

```go
func sum(a []int) (s int) {
    for _, v := range a {
        s += v
    }
    return
}

request := &Request{[]int{3, 4, 5}, sum, make(chan int)}
// Send request
clientRequests <- request
// Wait for response.
fmt.Printf("answer: %d\n", <-request.resultChan)
```

```go
func handle(queue chan *Request) {
    for req := range queue {
        req.resultChan <- req.f(req.args)
    }
}
```

### 12.7. Параллелизация

```go
type Vector []float64

// Apply the operation to v[i], v[i+1] ... up to v[n-1].
func (v Vector) DoSome(i, n int, u Vector, c chan int) {
    for ; i < n; i++ {
        v[i] += u.Op(v[i])
    }
    c <- 1    // signal that this piece is done
}
```

```go
const numCPU = 4 // number of CPU cores

func (v Vector) DoAll(u Vector) {
    c := make(chan int, numCPU)  // Buffering optional but sensible.
    for i := 0; i < numCPU; i++ {
        go v.DoSome(i*len(v)/numCPU, (i+1)*len(v)/numCPU, u, c)
    }
    // Drain the channel.
    for i := 0; i < numCPU; i++ {
        <-c    // wait for one task to complete
    }
    // All done.
}
```

```go
var numCPU = runtime.NumCPU()
```

```go
var numCPU = runtime.GOMAXPROCS(0)
```

### 12.8. Select

```go
select {
case <-ch1:
    // ...
case x := <-ch2:
    // ...use x...
case ch3 <- y:
    // ...
default:
    // ...
}
```

```go
select {
case <-ch:
    // ...
default:
    // ...
}
```

```go
select {
case <-ch:
    // ...
case <-time.After(1 * time.Second):
    // timed out
}
```

```go
func main() {
    c := make(chan int)
    quit := make(chan int)
    go func() {
        for i := 0; i < 10; i++ {
            fmt.Println(<-c)
        }
        quit <- 0
    }()
    fibonacci(c, quit)
}

func fibonacci(c, quit chan int) {
    x, y := 0, 1
    for {
        select {
        case c <- x:
            x, y = y, x+y
        case <-quit:
            fmt.Println("quit")
            return
        }
    }
}
```

```go
package main

import (
    "fmt"
    "time"
)

func main() {
    c := make(chan int)
    o := make(chan bool)
    go func() {
        for {
            select {
            case v := <-c:
                fmt.Println(v)
            case <-time.After(5 * time.Second):
                fmt.Println("timeout")
                o <- true
                break
            }
        }
    }()
    c <- 1
    c <- 2
    <-o
}
```

---

## 13. Пакеты и импорты

- Имя пакета важнее пути импорта.
- Избегайте `util`, `helper`, `common`.
- Размер: не один огромный пакет, но и не много крошечных.
- Файлы: сфокусированные, не тысячи строк. `doc.go` для документации пакета.
- Proto-импорты: суффиксы `pb`, `grpc`. Описательные имена.

```go
// Хорошо:
import (
    foopb "path/to/package/foo_service_go_proto"
    foogrpc "path/to/package/foo_service_go_grpc"
)
```

```go
// Хорошо:
import (
    pushqueueservicepb "path/to/package/push_queue_service_go_proto"
)
```

Примеры структуры пакетов:

- **маленькие**: `package csv` (`reader.go`, `writer.go`), `package expvar` (`expvar.go`).
- **средние**: `package flag` (`flag.go`).
- **большие**: `package http` (`client.go`, `server.go`, `cookie.go`), `package os` (`exec.go`, `file.go`, `tempfile.go`).

---

## 14. Generics

```go
// Хорошо:
func Map[T, U any](s []T, f func(T) U) []U {
    r := make([]U, len(s))
    for i, v := range s {
        r[i] = f(v)
    }
    return r
}
```

```go
// Плохо (излишнее использование):
func Max[T constraints.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

// Если вы всегда вызываете Max с int, возможно, проще:
func MaxInt(a, b int) int {
    if a > b {
        return a
    }
    return b
}
```

---

## 15. Ключевые принципы

1. **Форматирование** — `gofmt`.
2. **Именование** — коротко, лаконично, без повторений, MixedCaps.
3. **Управляющие структуры** — `if` с инициализацией, `range`, `switch` без fallthrough, type switch.
4. **Функции** — множественные возвраты, именованные результаты, `defer`.
5. **Данные** — `new` для обнуления, `make` для инициализации; срезы — ссылки; карты с comma ok.
6. **Методы и интерфейсы** — указатель vs значение; интерфейсы у потребителя; встраивание для композиции.
7. **Ошибки** — явно, структурно, `%w` для обёртывания, логировать один раз.
8. **Конкурентность** — «разделяйте память, общаясь»; горутины, каналы, `select`; следите за утечками.
9. **Паника** — только для исключительных ситуаций.
10. **Документация** — для экспортируемых сущностей, «почему» внутри.

Для углубления читайте спецификацию языка, Tour of Go, How to Write Go Code, документацию стандартной библиотеки и Go Code Review Comments.