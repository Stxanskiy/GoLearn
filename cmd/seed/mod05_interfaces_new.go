package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 5: Интерфейсы
// ════════════════════════════════════════════════════════════════

func mod05_interfaces_new() M {
	return M{
		Slug:          "interfaces",
		Title:         "Интерфейсы",
		Description:   "Контракты поведения: как писать гибкий код, который работает с разными типами данных через единый интерфейс.",
		Order:         5,
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"structs"},
		Lessons: []L{
			{
				Slug: "interface-basics", Title: "Интерфейсы — контракты поведения", Order: 1,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Интерфейсы — контракты поведения</h1>

<h2>Что такое интерфейс?</h2>
<p><strong>Интерфейс</strong> — это набор методов, которые тип должен реализовать. Интерфейс описывает <em>что</em> тип умеет делать, но не <em>как</em>.</p>

<p>Аналогия: розетка — это интерфейс. Любое устройство с вилкой подходит. Розетке не важно, телевизор это или чайник — главное, что вилка подходит.</p>

<pre><code>// Определение интерфейса
type Speaker interface {
    Speak() string
}

// Любой тип с методом Speak() string автоматически реализует Speaker
type Dog struct { Name string }
func (d Dog) Speak() string { return d.Name + ": Woof!" }

type Cat struct { Name string }
func (c Cat) Speak() string { return c.Name + ": Meow!" }

// Функция принимает интерфейс — работает с ЛЮБЫМ Speaker
func MakeNoise(s Speaker) {
    fmt.Println(s.Speak())
}

MakeNoise(Dog{Name: "Rex"})    // "Rex: Woof!"
MakeNoise(Cat{Name: "Whiskers"}) // "Whiskers: Meow!"</code></pre>

<p><strong>Ключевое отличие Go:</strong> типы реализуют интерфейсы <strong>неявно</strong>. Не нужно писать <code>implements</code>. Если у типа есть нужные методы — он автоматически подходит.</p>

<h2>Встроенные интерфейсы</h2>
<pre><code>// fmt.Stringer — если реализовать, fmt.Println выведет твой формат
type Stringer interface {
    String() string
}

type User struct { Name string; Age int }
func (u User) String() string {
    return fmt.Sprintf("%s (age %d)", u.Name, u.Age)
}
fmt.Println(User{"Alice", 25})  // "Alice (age 25)"

// error — самый важный интерфейс в Go
type error interface {
    Error() string
}</code></pre>

<h2>Пустой интерфейс — any</h2>
<pre><code>// interface{} (или any в Go 1.18+) принимает ЛЮБОЙ тип
func printAnything(v any) {
    fmt.Printf("Type: %T, Value: %v\n", v, v)
}

printAnything(42)        // Type: int, Value: 42
printAnything("hello")   // Type: string, Value: hello
printAnything(true)      // Type: bool, Value: true</code></pre>

<h2>Type assertion — извлечение конкретного типа</h2>
<pre><code>var s Speaker = Dog{Name: "Rex"}

// Утверждение типа
dog, ok := s.(Dog)
if ok {
    fmt.Println("It's a dog:", dog.Name)
}

// Type switch
switch v := s.(type) {
case Dog:
    fmt.Println("Dog:", v.Name)
case Cat:
    fmt.Println("Cat:", v.Name)
}</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Как тип реализует интерфейс в Go?",
						Options:     []string{"Через ключевое слово implements", "Автоматически — если у типа есть все методы интерфейса", "Через наследование", "Нужна регистрация"},
						Correct:     1,
						Explanation: "Неявная реализация — ключевая идея Go. Если тип Dog имеет метод Speak() string, он автоматически является Speaker. Без объявления.",
					},
					{
						Question:    "Что такое any (interface{})?",
						Options:     []string{"Тип строки", "Пустой интерфейс — принимает значение ЛЮБОГО типа", "Указатель", "Массив"},
						Correct:     1,
						Explanation: "any = interface{} — интерфейс без методов. Любой тип его реализует (у любого типа есть 0 или более методов). Используй осторожно — теряется информация о типе.",
					},
					{
						Question:    "Может ли один тип реализовывать несколько интерфейсов?",
						Options:     []string{"Нет — только один", "Да — если у типа есть все методы нескольких интерфейсов, он реализует все", "Только через embedding", "Только если интерфейсы из одного пакета"},
						Correct:     1,
						Explanation: "os.File реализует io.Reader, io.Writer, io.Closer, io.ReadWriter, io.Seeker — все сразу. Нет ограничений на количество интерфейсов для одного типа.",
					},
					{
						Question:    "Что произойдёт если присвоить типу интерфейс, но не реализовать все методы?",
						Options:     []string{"Ошибка в рантайме", "Ошибка компиляции — Go проверяет в момент присваивания", "Nil будет возвращён", "Ничего — методы просто пустые"},
						Correct:     1,
						Explanation: "Go проверяет совместимость с интерфейсом во время компиляции. var _ Speaker = Dog{} — статическая проверка. Если Dog не реализует Speaker — код не скомпилируется.",
					},
					{
						Question:    "Что значит 'duck typing' в Go?",
						Options:     []string{"Утиная охота в коде", "Если тип имеет нужные методы — он 'утка', не нужно объявлять что он реализует интерфейс", "Специальный тип данных", "Только для тестов"},
						Correct:     1,
						Explanation: "'If it walks like a duck and quacks like a duck, it is a duck.' В Go не нужно писать implements Duck. Достаточно иметь методы Walk() и Quack() — тип автоматически Duck.",
					},
				},
				Tasks: []T{
					{
						Title:      "Фигуры — площадь и периметр",
						Difficulty: "medium",
						Description: `<p>Создай интерфейс <code>Shape</code> с методами <code>Area() float64</code> и <code>Perimeter() float64</code>. Реализуй для Circle и Rectangle.</p>
<p>Ввод: <code>circle 5</code> или <code>rect 4 6</code></p>
<p>Вывод для <code>circle 5</code>:</p>
<pre><code>Area: 78.5
Perimeter: 31.4</code></pre>
<p>Вывод для <code>rect 4 6</code>:</p>
<pre><code>Area: 24.0
Perimeter: 20.0</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "type Shape interface { ... }", Definition: "Определение интерфейса. Перечисляет методы, которые должен реализовать тип."},
							{Term: "math.Pi", Definition: "Константа числа Пи (3.14159...) из пакета math. Нужен import \"math\"."},
							{Term: "неявная реализация", Definition: "Тип реализует интерфейс автоматически, если имеет все нужные методы. Не нужно писать implements."},
						},
						TestCases: []TestCase{
							{Input: "circle 5", ExpectedOutput: "Area: 78.5\nPerimeter: 31.4"},
							{Input: "rect 4 6", ExpectedOutput: "Area: 24.0\nPerimeter: 20.0"},
						},
						StarterCode: `package main

import (
    "fmt"
    "math"
)

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct { Radius float64 }
type Rectangle struct { Width, Height float64 }

// Реализуй методы Area и Perimeter для Circle и Rectangle

func printShape(s Shape) {
    fmt.Printf("Area: %.1f\n", s.Area())
    fmt.Printf("Perimeter: %.1f\n", s.Perimeter())
}

func main() {
    var kind string
    fmt.Scan(&kind)

    _ = math.Pi // убери когда используешь

    switch kind {
    case "circle":
        var r float64
        fmt.Scan(&r)
        printShape(Circle{Radius: r})
    case "rect":
        var w, h float64
        fmt.Scan(&w, &h)
        printShape(Rectangle{Width: w, Height: h})
    }
}`,
						Hints: `<p>Circle: Area = Pi*r², Perimeter = 2*Pi*r. Rectangle: Area = w*h, Perimeter = 2*(w+h).</p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "math"
)

type Shape interface {
    Area() float64
    Perimeter() float64
}

type Circle struct{ Radius float64 }
type Rectangle struct{ Width, Height float64 }

func (c Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c Circle) Perimeter() float64  { return 2 * math.Pi * c.Radius }
func (r Rectangle) Area() float64    { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func printShape(s Shape) {
    fmt.Printf("Area: %.1f\n", s.Area())
    fmt.Printf("Perimeter: %.1f\n", s.Perimeter())
}

func main() {
    var kind string
    fmt.Scan(&kind)

    switch kind {
    case "circle":
        var r float64
        fmt.Scan(&r)
        printShape(Circle{Radius: r})
    case "rect":
        var w, h float64
        fmt.Scan(&w, &h)
        printShape(Rectangle{Width: w, Height: h})
    }
}</code></pre>`,
					},
					{
						Title:      "Stringer — свой формат вывода",
						Difficulty: "easy",
						Description: `<p>Реализуй интерфейс <code>fmt.Stringer</code> для структуры <code>User</code>, чтобы <code>fmt.Println(user)</code> выводил красивый формат.</p>
<p>Ввод: <code>Alice 25</code></p>
<p>Вывод: <code>User(Alice, age 25)</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "fmt.Stringer", Definition: "Интерфейс с методом String() string. Если тип реализует его — fmt.Println выводит результат String()."},
						},
						TestCases: []TestCase{
							{Input: "Alice 25", ExpectedOutput: "User(Alice, age 25)"},
							{Input: "Bob 30", ExpectedOutput: "User(Bob, age 30)"},
						},
						StarterCode: `package main

import "fmt"

type User struct {
    Name string
    Age  int
}

// Реализуй метод String() string для User

func main() {
    var u User
    fmt.Scan(&u.Name, &u.Age)
    fmt.Println(u)
}`,
						Hints: `<p><code>func (u User) String() string { return fmt.Sprintf("User(%s, age %d)", u.Name, u.Age) }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type User struct {
    Name string
    Age  int
}

func (u User) String() string {
    return fmt.Sprintf("User(%s, age %d)", u.Name, u.Age)
}

func main() {
    var u User
    fmt.Scan(&u.Name, &u.Age)
    fmt.Println(u)
}</code></pre>`,
					},
					{
						Title:      "Универсальный сортировщик",
						Difficulty: "hard",
						Description: `<p>Напиши функцию <code>sortByLength</code>, которая сортирует слайс строк по длине (от короткой к длинной). При равной длине — по алфавиту.</p>
<p>Ввод:</p>
<pre><code>4
banana fig apple kiwi</code></pre>
<p>Вывод:</p>
<pre><code>fig kiwi apple banana</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "sort.Slice(s, less)", Definition: "Сортирует слайс по произвольному правилу. less(i, j) — функция сравнения: true если s[i] должен быть раньше s[j]."},
						},
						TestCases: []TestCase{
							{Input: "4\nbanana fig apple kiwi", ExpectedOutput: "fig kiwi apple banana"},
							{Input: "3\nc aa b", ExpectedOutput: "b c aa"},
						},
						StarterCode: `package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words {
        fmt.Scan(&words[i])
    }

    // Отсортируй по длине, при равной — по алфавиту
    _ = sort.Slice

    for i, w := range words {
        if i > 0 { fmt.Print(" ") }
        fmt.Print(w)
    }
    fmt.Println()
}`,
						Hints: `<p><code>sort.Slice(words, func(i, j int) bool { if len(words[i]) == len(words[j]) { return words[i] < words[j] } return len(words[i]) < len(words[j]) })</code></p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words {
        fmt.Scan(&words[i])
    }

    sort.Slice(words, func(i, j int) bool {
        if len(words[i]) == len(words[j]) {
            return words[i] < words[j]
        }
        return len(words[i]) < len(words[j])
    })

    for i, w := range words {
        if i > 0 { fmt.Print(" ") }
        fmt.Print(w)
    }
    fmt.Println()
}</code></pre>`,
					},
					{
						Title:      "Метод оплаты — интерфейс",
						Difficulty: "easy",
						Description: `<p>Создай интерфейс <code>PaymentMethod</code> с методом <code>Pay(amount int) string</code>. Реализуй <code>Bank</code> и <code>Crypto</code>:</p>
<p>Ввод:</p>
<pre><code>2
bank 100
crypto 50</code></pre>
<p>Вывод:</p>
<pre><code>Bank payment: $100
Crypto payment: 50 USDT</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "PaymentMethod interface", Definition: "Классический пример: один интерфейс, несколько реализаций (Bank, Crypto, PayPal). Код принимает интерфейс — работает с любой реализацией."},
						},
						TestCases: []TestCase{
							{Input: "2\nbank 100\ncrypto 50", ExpectedOutput: "Bank payment: $100\nCrypto payment: 50 USDT"},
						},
						StarterCode: `package main

import "fmt"

type PaymentMethod interface {
    Pay(amount int) string
}

type Bank struct{}
type Crypto struct{}

func (b Bank) Pay(amount int) string {
    return fmt.Sprintf("Bank payment: $%d", amount)
}

func (c Crypto) Pay(amount int) string {
    // "Crypto payment: N USDT"
    return ""
}

func process(method PaymentMethod, amount int) {
    fmt.Println(method.Pay(amount))
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind string
        var amount int
        fmt.Scan(&kind, &amount)
        switch kind {
        case "bank":   process(Bank{}, amount)
        case "crypto": process(Crypto{}, amount)
        }
    }
}`,
						Hints: `<p><code>fmt.Sprintf("Crypto payment: %d USDT", amount)</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type PaymentMethod interface {
    Pay(amount int) string
}

type Bank struct{}
type Crypto struct{}

func (b Bank) Pay(amount int) string   { return fmt.Sprintf("Bank payment: $%d", amount) }
func (c Crypto) Pay(amount int) string { return fmt.Sprintf("Crypto payment: %d USDT", amount) }

func process(method PaymentMethod, amount int) {
    fmt.Println(method.Pay(amount))
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind string; var amount int
        fmt.Scan(&kind, &amount)
        switch kind {
        case "bank":   process(Bank{}, amount)
        case "crypto": process(Crypto{}, amount)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Уведомления через интерфейс",
						Difficulty: "medium",
						Description: `<p>Создай интерфейс <code>Notifier</code> с методом <code>Send(to, msg string) string</code>. Реализуй <code>EmailNotifier</code> и <code>SMSNotifier</code>. Функция <code>notify</code> принимает интерфейс:</p>
<p>Ввод:</p>
<pre><code>3
email alice@mail.com Welcome
sms +79001234567 Your code: 1234
email bob@mail.com Goodbye</code></pre>
<p>Вывод:</p>
<pre><code>Email to alice@mail.com: Welcome
SMS to +79001234567: Your code: 1234
Email to bob@mail.com: Goodbye</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Dependency injection", Definition: "Функция получает интерфейс, не конкретный тип. Легко добавить новый канал (Telegram, Push) не меняя функцию notify."},
						},
						TestCases: []TestCase{
							{Input: "3\nemail alice@mail.com Welcome\nsms +79001234567 Your code: 1234\nemail bob@mail.com Goodbye", ExpectedOutput: "Email to alice@mail.com: Welcome\nSMS to +79001234567: Your code: 1234\nEmail to bob@mail.com: Goodbye"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Notifier interface {
    Send(to, msg string) string
}

type EmailNotifier struct{}
type SMSNotifier struct{}

func (e EmailNotifier) Send(to, msg string) string {
    return fmt.Sprintf("Email to %s: %s", to, msg)
}

func (s SMSNotifier) Send(to, msg string) string {
    // "SMS to TO: MSG"
    return ""
}

func notify(n Notifier, to, msg string) {
    fmt.Println(n.Send(to, msg))
}

func main() {
    var count int
    fmt.Scan(&count)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < count; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 3)
        if len(parts) < 3 { continue }
        kind, to, msg := parts[0], parts[1], parts[2]
        switch kind {
        case "email": notify(EmailNotifier{}, to, msg)
        case "sms":   notify(SMSNotifier{}, to, msg)
        }
    }
}`,
						Hints: `<p><code>fmt.Sprintf("SMS to %s: %s", to, msg)</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Notifier interface {
    Send(to, msg string) string
}

type EmailNotifier struct{}
type SMSNotifier struct{}

func (e EmailNotifier) Send(to, msg string) string { return fmt.Sprintf("Email to %s: %s", to, msg) }
func (s SMSNotifier) Send(to, msg string) string   { return fmt.Sprintf("SMS to %s: %s", to, msg) }

func notify(n Notifier, to, msg string) {
    fmt.Println(n.Send(to, msg))
}

func main() {
    var count int
    fmt.Scan(&count)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < count; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 3)
        if len(parts) < 3 { continue }
        kind, to, msg := parts[0], parts[1], parts[2]
        switch kind {
        case "email": notify(EmailNotifier{}, to, msg)
        case "sms":   notify(SMSNotifier{}, to, msg)
        }
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug: "io-interfaces", Title: "Стандартные интерфейсы: io.Reader, io.Writer", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Стандартные интерфейсы Go</h1>

<h2>io.Reader и io.Writer — основа всего I/O</h2>
<p>Два самых важных интерфейса в Go:</p>

<pre><code>type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// Что реализует Reader?
// - os.File        — файлы
// - http.Request.Body — тело HTTP-запроса
// - strings.Reader — строка как поток
// - bytes.Buffer   — буфер в памяти
// - gzip.Reader    — сжатые данные

// Одна функция работает со ВСЕМИ:
func countBytes(r io.Reader) (int, error) {
    buf := make([]byte, 1024)
    total := 0
    for {
        n, err := r.Read(buf)
        total += n
        if err == io.EOF { return total, nil }
        if err != nil { return total, err }
    }
}

// Работает с файлом:
f, _ := os.Open("data.txt")
n, _ := countBytes(f)

// С HTTP-телом:
n, _ = countBytes(req.Body)

// Со строкой:
n, _ = countBytes(strings.NewReader("hello world"))</code></pre>

<h2>Маленькие интерфейсы — сила Go</h2>
<pre><code>// Stringer — для красивого вывода
type Stringer interface {
    String() string
}

// error — для ошибок
type error interface {
    Error() string
}

// sort.Interface — для сортировки
type Interface interface {
    Len() int
    Less(i, j int) bool
    Swap(i, j int)
}

// http.Handler — для HTTP
type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}</code></pre>

<p><strong>Правило Go:</strong> чем меньше интерфейс, тем полезнее. 1-2 метода — идеал. Большие интерфейсы труднее реализовать и мокать.</p>

<h2>Интерфейс как контракт для тестирования</h2>
<pre><code>// Реальный код зависит от интерфейса, а не реализации:
type UserStore interface {
    GetByID(ctx context.Context, id int) (*User, error)
    Create(ctx context.Context, u *User) error
}

type UserService struct {
    store UserStore  // интерфейс, не конкретный тип!
}

// В продакшене:
service := &UserService{store: &PostgresUserStore{pool: pool}}

// В тестах:
service := &UserService{store: &MockUserStore{users: testData}}</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Почему io.Reader имеет только один метод?",
						Options:     []string{"Забыли добавить другие", "Маленький интерфейс легче реализовать — больше типов подходят, код более переиспользуемый", "Один метод быстрее", "Это временное решение"},
						Correct:     1,
						Explanation: "Чем меньше интерфейс, тем больше типов его реализуют. io.Reader с одним методом подходит для файлов, сети, строк, буферов, сжатия — всё это потоки байтов.",
					},
					{
						Question:    "Зачем зависеть от интерфейса, а не от конкретного типа?",
						Options:     []string{"Для красоты", "Для тестируемости: в продакшене — реальная БД, в тестах — мок", "Для скорости", "Go требует"},
						Correct:     1,
						Explanation: "Dependency Inversion: зависимость от абстракции, не реализации. Можно подменить PostgresRepo на MockRepo без изменения бизнес-логики.",
					},
					{
						Question:    "Что означает io.EOF при чтении?",
						Options:     []string{"Ошибка чтения", "Конец данных — не настоящая ошибка, сигнал о завершении потока", "Файл повреждён", "Недостаточно памяти"},
						Correct:     1,
						Explanation: "io.EOF — специальная ошибка-сентинел: var EOF = errors.New(\"EOF\"). Это не ошибка, а сигнал 'данные закончились'. В цикле чтения: if err == io.EOF { break }. Остальные ошибки — настоящие проблемы.",
					},
					{
						Question:    "Что делает strings.NewReader(s)?",
						Options:     []string{"Открывает файл с именем s", "Создаёт io.Reader из строки — позволяет обработать строку как поток байтов", "Читает строку из stdin", "Создаёт Scanner"},
						Correct:     1,
						Explanation: "strings.NewReader возвращает *strings.Reader, который реализует io.Reader. Удобно для тестов и когда нужно передать строку в функцию, ожидающую io.Reader.",
					},
					{
						Question:    "io.Copy(dst, src) — что делает?",
						Options:     []string{"Копирует файл на диске", "Читает из src (io.Reader) и пишет в dst (io.Writer) до io.EOF", "Только для сетевых соединений", "Требует одинаковые типы"},
						Correct:     1,
						Explanation: "io.Copy — универсальная функция: читает из любого Reader и пишет в любой Writer. Используется для копирования файлов, передачи HTTP-тела, стриминга данных.",
					},
				},
				Tasks: []T{
					{
						Title:      "Подсчёт слов из любого Reader",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>wordCount(r io.Reader) int</code>, которая считает слова из любого источника. Используй <code>bufio.Scanner</code>.</p>
<p>Ввод: <code>hello world foo bar</code></p>
<p>Вывод: <code>4</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "io.Reader", Definition: "Интерфейс потока чтения. Один метод Read([]byte). Реализуют: файлы, строки, сеть, буферы."},
							{Term: "bufio.NewScanner(r)", Definition: "Обёртка над io.Reader для удобного чтения. scanner.Split(bufio.ScanWords) — читать по словам."},
							{Term: "strings.NewReader(s)", Definition: "Превращает строку в io.Reader. Полезно для тестов и обработки строк как потоков."},
						},
						TestCases: []TestCase{
							{Input: "hello world foo bar", ExpectedOutput: "4"},
							{Input: "one", ExpectedOutput: "1"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
)

func wordCount(r io.Reader) int {
    // Используй bufio.Scanner с ScanWords
    _ = bufio.ScanWords
    return 0
}

func main() {
    fmt.Println(wordCount(os.Stdin))
}`,
						Hints: `<p><code>scanner := bufio.NewScanner(r); scanner.Split(bufio.ScanWords); count := 0; for scanner.Scan() { count++ }</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "io"
    "os"
)

func wordCount(r io.Reader) int {
    scanner := bufio.NewScanner(r)
    scanner.Split(bufio.ScanWords)
    count := 0
    for scanner.Scan() {
        count++
    }
    return count
}

func main() {
    fmt.Println(wordCount(os.Stdin))
}</code></pre>`,
					},
					{
						Title:      "Свой io.Writer — подсчёт байт",
						Difficulty: "easy",
						Description: `<p>Реализуй структуру <code>ByteCounter</code>, которая реализует <code>io.Writer</code> и считает сколько байт было записано:</p>
<p>Ввод: <code>Hello World</code></p>
<p>Вывод: <code>Written: 11 bytes</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "io.Writer", Definition: "Интерфейс: Write(p []byte) (n int, err error). Реализуется любым типом, который принимает байты — файл, буфер, сеть."},
							{Term: "fmt.Fprintf(w, ...)", Definition: "Как fmt.Printf но пишет в io.Writer. Позволяет писать в любой Writer."},
						},
						TestCases: []TestCase{
							{Input: "Hello World", ExpectedOutput: "Written: 11 bytes"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
)

type ByteCounter struct {
    count int
}

func (bc *ByteCounter) Write(p []byte) (int, error) {
    // Посчитай количество байт, верни len(p), nil
    return 0, nil
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    counter := &ByteCounter{}
    fmt.Fprint(counter, line)
    fmt.Printf("Written: %d bytes\n", counter.count)
}`,
						Hints: `<p><code>bc.count += len(p); return len(p), nil</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
)

type ByteCounter struct {
    count int
}

func (bc *ByteCounter) Write(p []byte) (int, error) {
    bc.count += len(p)
    return len(p), nil
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    counter := &ByteCounter{}
    fmt.Fprint(counter, line)
    fmt.Printf("Written: %d bytes\n", counter.count)
}</code></pre>`,
					},
					{
						Title:      "io.Copy между Reader и Writer",
						Difficulty: "easy",
						Description: `<p>Используй <code>io.Copy</code> чтобы скопировать содержимое строки в <code>bytes.Buffer</code>, потом выведи результат:</p>
<p>Ввод: <code>Hello from io.Copy!</code></p>
<p>Вывод: <code>Hello from io.Copy!</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "bytes.Buffer", Definition: "Буфер в памяти, реализует и io.Reader и io.Writer. buf.String() → содержимое как строка."},
							{Term: "io.Copy(dst, src)", Definition: "Копирует из src (Reader) в dst (Writer) до EOF. Возвращает (bytes_copied, error)."},
						},
						TestCases: []TestCase{
							{Input: "Hello from io.Copy!", ExpectedOutput: "Hello from io.Copy!"},
						},
						StarterCode: `package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    src := strings.NewReader(line)
    var dst bytes.Buffer

    _, err := io.Copy(&dst, src)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Print(dst.String())
}`,
						Hints: `<p><code>io.Copy(&dst, src)</code> — копирует всё из Reader в Writer. <code>dst.String()</code> — содержимое буфера.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    src := strings.NewReader(line)
    var dst bytes.Buffer
    io.Copy(&dst, src)
    fmt.Print(dst.String())
}</code></pre>`,
					},
					{
						Title:      "Uppercase Reader",
						Difficulty: "medium",
						Description: `<p>Реализуй <code>UpperReader</code> — обёртку над io.Reader, которая переводит все буквы в верхний регистр при чтении:</p>
<p>Ввод: <code>hello world</code></p>
<p>Вывод: <code>HELLO WORLD</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "Декоратор Reader", Definition: "Структура содержит io.Reader, метод Read читает данные и трансформирует. Классический паттерн в Go I/O."},
							{Term: "bytes.ToUpper(p)", Definition: "Преобразует байты в верхний регистр. Или: for i, b := range p { p[i] = byte(unicode.ToUpper(rune(b))) }"},
						},
						TestCases: []TestCase{
							{Input: "hello world", ExpectedOutput: "HELLO WORLD"},
						},
						StarterCode: `package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
)

type UpperReader struct {
    r io.Reader
}

func (u UpperReader) Read(p []byte) (int, error) {
    n, err := u.r.Read(p)
    // Преобразуй прочитанные байты в верхний регистр
    return n, err
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    upper := UpperReader{r: strings.NewReader(line)}
    var buf bytes.Buffer
    io.Copy(&buf, upper)
    fmt.Print(buf.String())
}`,
						Hints: `<p>После <code>n, err := u.r.Read(p)</code>: <code>copy(p[:n], bytes.ToUpper(p[:n]))</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "bytes"
    "fmt"
    "io"
    "os"
    "strings"
)

type UpperReader struct {
    r io.Reader
}

func (u UpperReader) Read(p []byte) (int, error) {
    n, err := u.r.Read(p)
    copy(p[:n], bytes.ToUpper(p[:n]))
    return n, err
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    upper := UpperReader{r: strings.NewReader(line)}
    var buf bytes.Buffer
    io.Copy(&buf, upper)
    fmt.Print(buf.String())
}</code></pre>`,
					},
					{
						Title:      "UserStore — интерфейс для хранилища",
						Difficulty: "hard",
						Description: `<p>Реализуй dependency injection паттерн: интерфейс <code>UserStore</code> с двумя реализациями — <code>MemoryStore</code> (in-memory) и вызов через единый сервис:</p>
<p>Ввод:</p>
<pre><code>4
create 1 Alice
create 2 Bob
get 1
get 99</code></pre>
<p>Вывод:</p>
<pre><code>Created: Alice
Created: Bob
Found: Alice
Not found: 99</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "UserStore interface", Definition: "Контракт для хранилища. MemoryStore — для тестов/разработки. В реальном коде подменяется PostgresStore без изменения сервиса."},
							{Term: "Dependency Injection", Definition: "UserService получает UserStore через конструктор. Не создаёт хранилище сам — зависимость 'вводится' снаружи."},
						},
						TestCases: []TestCase{
							{Input: "4\ncreate 1 Alice\ncreate 2 Bob\nget 1\nget 99", ExpectedOutput: "Created: Alice\nCreated: Bob\nFound: Alice\nNot found: 99"},
						},
						StarterCode: `package main

import "fmt"

type User struct {
    ID   int
    Name string
}

type UserStore interface {
    Create(u User)
    GetByID(id int) (User, bool)
}

type MemoryStore struct {
    users map[int]User
}

func NewMemoryStore() *MemoryStore {
    return &MemoryStore{users: make(map[int]User)}
}

func (m *MemoryStore) Create(u User) {
    m.users[u.ID] = u
}

func (m *MemoryStore) GetByID(id int) (User, bool) {
    u, ok := m.users[id]
    return u, ok
}

type UserService struct {
    store UserStore
}

func NewUserService(store UserStore) *UserService {
    return &UserService{store: store}
}

func (s *UserService) Create(id int, name string) {
    s.store.Create(User{ID: id, Name: name})
    fmt.Printf("Created: %s\n", name)
}

func (s *UserService) Get(id int) {
    if u, ok := s.store.GetByID(id); ok {
        fmt.Printf("Found: %s\n", u.Name)
    } else {
        fmt.Printf("Not found: %d\n", id)
    }
}

func main() {
    store := NewMemoryStore()
    svc := NewUserService(store)

    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "create":
            var id int; var name string
            fmt.Scan(&id, &name)
            svc.Create(id, name)
        case "get":
            var id int
            fmt.Scan(&id)
            svc.Get(id)
        }
    }
}`,
						Hints: `<p>UserService не знает о MemoryStore — только об интерфейсе UserStore. Это позволяет подменить хранилище на PostgresStore.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type User struct{ ID int; Name string }

type UserStore interface {
    Create(u User)
    GetByID(id int) (User, bool)
}

type MemoryStore struct{ users map[int]User }

func NewMemoryStore() *MemoryStore { return &MemoryStore{users: make(map[int]User)} }
func (m *MemoryStore) Create(u User) { m.users[u.ID] = u }
func (m *MemoryStore) GetByID(id int) (User, bool) { u, ok := m.users[id]; return u, ok }

type UserService struct{ store UserStore }
func NewUserService(s UserStore) *UserService { return &UserService{store: s} }

func (s *UserService) Create(id int, name string) {
    s.store.Create(User{id, name})
    fmt.Printf("Created: %s\n", name)
}

func (s *UserService) Get(id int) {
    if u, ok := s.store.GetByID(id); ok {
        fmt.Printf("Found: %s\n", u.Name)
    } else {
        fmt.Printf("Not found: %d\n", id)
    }
}

func main() {
    svc := NewUserService(NewMemoryStore())
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "create":
            var id int; var name string
            fmt.Scan(&id, &name)
            svc.Create(id, name)
        case "get":
            var id int
            fmt.Scan(&id)
            svc.Get(id)
        }
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug: "type-assertions", Title: "Type assertions и type switch", Order: 3,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Type assertions и type switch</h1>

<h2>Под капотом: как устроен interface</h2>
<p>Интерфейсная переменная внутри — это пара: <strong>(тип, значение)</strong>:</p>
<pre><code>var w io.Writer     // (nil, nil) — пустой интерфейс
w = os.Stdout       // (*os.File, &{stdout})
w = &bytes.Buffer{} // (*bytes.Buffer, &{[]})

// Интерфейс "забывает" конкретный тип
// Но можно "вспомнить" через type assertion!</code></pre>

<h2>Type assertion — извлечение конкретного типа</h2>
<pre><code>var val any = "hello"

// Опасная форма: паникует если тип не совпал
s := val.(string)  // s = "hello"

// Безопасная форма: возвращает ok
s, ok := val.(string)   // s = "hello", ok = true
n, ok := val.(int)       // n = 0, ok = false — без паники!

// ВСЕГДА используй безопасную форму:
if s, ok := val.(string); ok {
    fmt.Println("это строка:", s)
} else {
    fmt.Println("не строка")
}</code></pre>

<h2>Type switch — проверка нескольких типов</h2>
<pre><code>func describe(val any) string {
    switch v := val.(type) {  // v — уже конкретный тип!
    case int:
        return fmt.Sprintf("int: %d", v)
    case string:
        return fmt.Sprintf("string: %q (len=%d)", v, len(v))
    case bool:
        return fmt.Sprintf("bool: %v", v)
    case []int:
        return fmt.Sprintf("slice of %d ints", len(v))
    case nil:
        return "nil"
    default:
        return fmt.Sprintf("unknown: %T", v)
    }
}</code></pre>

<h2>Реальный пример: обработка ошибок</h2>
<pre><code>type ValidationError struct {
    Field   string
    Message string
}
func (e *ValidationError) Error() string {
    return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func handleError(err error) {
    switch e := err.(type) {
    case *ValidationError:
        fmt.Printf("Ошибка валидации поля %s: %s\n", e.Field, e.Message)
    case *os.PathError:
        fmt.Printf("Ошибка файла %s: %s\n", e.Path, e.Err)
    default:
        fmt.Println("Неизвестная ошибка:", err)
    }
}</code></pre>

<h2>any vs interface{}</h2>
<pre><code>// any — это просто алиас для interface{} (с Go 1.18)
type any = interface{}

// Используй any — читаемее
func process(data any) { ... }   // ✅
func process(data interface{}) { ... }  // ❌ устарело</code></pre>

<h2>Embedding интерфейсов</h2>
<pre><code>type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
type Closer interface { Close() error }

// Комбинирование через embedding:
type ReadWriter interface {
    Reader  // все методы Reader
    Writer  // + все методы Writer
}

type ReadWriteCloser interface {
    Reader
    Writer
    Closer
}

// os.File реализует ReadWriteCloser — все 3 метода</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: type assertion без ok
val := myInterface.(string) // ПАНИКА если не string!

// ОШИБКА: nil interface vs nil pointer
var p *bytes.Buffer   // nil pointer
var w io.Writer = p   // w = (*bytes.Buffer, nil) — NOT nil!
fmt.Println(w == nil) // false! Тип есть, значения нет

// Правильно: проверяй через reflect или type switch</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Что вернёт v, ok := val.(int) если val содержит string?",
						Options:     []string{"Паника", "v = 0, ok = false — безопасная форма не паникует", "v = nil, ok = true", "Ошибка компиляции"},
						Correct:     1,
						Explanation: "Двойная форма type assertion v, ok := val.(T) никогда не паникует. Если тип не совпал: v = zero value, ok = false.",
					},
					{
						Question:    "Что такое any в Go?",
						Options:     []string{"Специальный тип", "Алиас для interface{} — означает 'любой тип'", "Ключевое слово generics", "Тип из пакета reflect"},
						Correct:     1,
						Explanation: "any = interface{} (с Go 1.18). Используй any вместо interface{} — короче и читаемее.",
					},
					{
						Question:    "Чем опасен var w io.Writer = (*bytes.Buffer)(nil)?",
						Options:     []string{"Ничем", "w != nil, хотя значение nil. Интерфейс хранит пару (тип, значение) и тип не nil", "Ошибка компиляции", "Утечка памяти"},
						Correct:     1,
						Explanation: "Интерфейс = (тип, значение). Даже если значение nil, тип (*bytes.Buffer) есть → интерфейс не nil. Классическая ловушка Go.",
					},
					{
						Question:    "Как можно объединить несколько интерфейсов в один?",
						Options:     []string{"Нельзя", "Через embedding: type ReadWriter interface { Reader; Writer }", "Только через наследование", "Через + оператор"},
						Correct:     1,
						Explanation: "Интерфейсы поддерживают embedding: type ReadWriteCloser interface { Reader; Writer; Closer }. Тип должен реализовать все методы всех встроенных интерфейсов.",
					},
					{
						Question:    "val.(string) vs val, ok := val.(string) — в чём разница?",
						Options:     []string{"Нет разницы", "Первая форма паникует если тип не совпал. Вторая форма безопасна — возвращает ok=false", "Вторая медленнее", "Первая только для строк"},
						Correct:     1,
						Explanation: "Одиночная форма val.(T) паникует при несоответствии типа — используй только если 100% уверен. Двойная форма val, ok := val.(T) никогда не паникует — всегда предпочтительна.",
					},
				},
				Tasks: []T{
					{
						Title:      "Describe — type switch",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "switch v := val.(type)", Definition: "Type switch — проверяет конкретный тип интерфейсной переменной. v уже нужного типа в каждом case."},
							{Term: "any", Definition: "Алиас для interface{}. Принимает значение любого типа."},
						},
						Description: `<p>Реализуй функцию <code>describe(val any) string</code> через type switch:</p>
<ul>
<li><code>int</code> → <code>"int:N"</code></li>
<li><code>string</code> → <code>"string:S(N)"</code> где N — длина</li>
<li><code>bool</code> → <code>"bool:true/false"</code></li>
<li>другое → <code>"unknown"</code></li>
</ul>
<p>Ввод: <code>42 hello true 3.14</code> (каждое на новой строке с типом)</p>
<p>Вывод:</p>
<pre><code>int:42
string:hello(5)
bool:true
unknown</code></pre>`,
						Hints: `<p>switch v := val.(type) { case int: ... case string: ... }</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func describe(val any) string {
	switch v := val.(type) {
	case int:
		return fmt.Sprintf("int:%d", v)
	case string:
		return fmt.Sprintf("string:%s(%d)", v, len(v))
	case bool:
		return fmt.Sprintf("bool:%v", v)
	default:
		return "unknown"
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		typ, val := parts[0], parts[1]
		switch typ {
		case "int":
			n, _ := strconv.Atoi(val)
			fmt.Println(describe(n))
		case "string":
			fmt.Println(describe(val))
		case "bool":
			b, _ := strconv.ParseBool(val)
			fmt.Println(describe(b))
		default:
			f, _ := strconv.ParseFloat(val, 64)
			fmt.Println(describe(f))
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func describe(val any) string {
	// TODO: type switch
	// switch v := val.(type) {
	// case int: return fmt.Sprintf("int:%d", v)
	// case string: ...
	// case bool: ...
	// default: "unknown"
	// }
	return "unknown"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		typ, val := parts[0], parts[1]
		switch typ {
		case "int":
			n, _ := strconv.Atoi(val)
			fmt.Println(describe(n))
		case "string":
			fmt.Println(describe(val))
		case "bool":
			b, _ := strconv.ParseBool(val)
			fmt.Println(describe(b))
		default:
			f, _ := strconv.ParseFloat(val, 64)
			fmt.Println(describe(f))
		}
	}
}`,
						TestCases: []TestCase{
							{Input: "int 42\nstring hello\nbool true\nfloat 3.14", ExpectedOutput: "int:42\nstring:hello(5)\nbool:true\nunknown"},
							{Input: "string world\nint -7\nbool false", ExpectedOutput: "string:world(5)\nint:-7\nbool:false"},
						},
					},
					{
						Title:      "Безопасное извлечение типа",
						Difficulty: "easy",
						Description: `<p>Напиши <code>toString(v any) string</code>: string → вернуть; int → в строку; остальное → "unknown":</p>
<p>Ввод:</p>
<pre><code>3
string hello
int 42
float 3.14</code></pre>
<p>Вывод:</p>
<pre><code>hello
42
unknown</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "v, ok := val.(string)", Definition: "Безопасная форма — не паникует, возвращает ok=false при несоответствии."},
						},
						TestCases: []TestCase{
							{Input: "3\nstring hello\nint 42\nfloat 3.14", ExpectedOutput: "hello\n42\nunknown"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func toString(v any) string {
    switch val := v.(type) {
    case string: return val
    case int:    return strconv.Itoa(val)
    default:     return "unknown"
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        if len(parts) < 2 { continue }
        switch parts[0] {
        case "string": fmt.Println(toString(parts[1]))
        case "int":
            num, _ := strconv.Atoi(parts[1])
            fmt.Println(toString(num))
        default:
            f, _ := strconv.ParseFloat(parts[1], 64)
            fmt.Println(toString(f))
        }
    }
}`,
						Hints: `<p>switch val := v.(type) { case string: return val; case int: return strconv.Itoa(val) }</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func toString(v any) string {
    switch val := v.(type) {
    case string: return val
    case int:    return strconv.Itoa(val)
    default:     return "unknown"
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        if len(parts) < 2 { continue }
        switch parts[0] {
        case "string": fmt.Println(toString(parts[1]))
        case "int":
            num, _ := strconv.Atoi(parts[1])
            fmt.Println(toString(num))
        default:
            f, _ := strconv.ParseFloat(parts[1], 64)
            fmt.Println(toString(f))
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Сумма int из []any",
						Difficulty: "easy",
						Description: `<p>Функция <code>sumInts(vals []any) int</code> — суммирует только int-элементы, остальные пропускает:</p>
<p>Ввод: <code>5 int:10 string:foo int:20 bool:true int:5</code></p>
<p>Вывод: <code>35</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "type assertion в range", Definition: "if n, ok := v.(int); ok { sum += n } — безопасно извлекает int из []any."},
						},
						TestCases: []TestCase{
							{Input: "5 int:10 string:foo int:20 bool:true int:5", ExpectedOutput: "35"},
							{Input: "3 string:a string:b string:c", ExpectedOutput: "0"},
						},
						StarterCode: `package main

import (
    "fmt"
    "strconv"
    "strings"
)

func sumInts(vals []any) int {
    sum := 0
    for _, v := range vals {
        if n, ok := v.(int); ok {
            sum += n
        }
    }
    return sum
}

func main() {
    var n int
    fmt.Scan(&n)
    vals := make([]any, n)
    for i := 0; i < n; i++ {
        var token string
        fmt.Scan(&token)
        parts := strings.SplitN(token, ":", 2)
        switch parts[0] {
        case "int":
            num, _ := strconv.Atoi(parts[1])
            vals[i] = num
        case "string":
            vals[i] = parts[1]
        case "bool":
            b, _ := strconv.ParseBool(parts[1])
            vals[i] = b
        }
    }
    fmt.Println(sumInts(vals))
}`,
						Hints: `<p><code>if n, ok := v.(int); ok { sum += n }</code></p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "strconv"
    "strings"
)

func sumInts(vals []any) int {
    sum := 0
    for _, v := range vals {
        if n, ok := v.(int); ok { sum += n }
    }
    return sum
}

func main() {
    var n int
    fmt.Scan(&n)
    vals := make([]any, n)
    for i := 0; i < n; i++ {
        var token string
        fmt.Scan(&token)
        parts := strings.SplitN(token, ":", 2)
        switch parts[0] {
        case "int":    num, _ := strconv.Atoi(parts[1]); vals[i] = num
        case "string": vals[i] = parts[1]
        case "bool":   b, _ := strconv.ParseBool(parts[1]); vals[i] = b
        }
    }
    fmt.Println(sumInts(vals))
}</code></pre>`,
					},
					{
						Title:      "Кастомные ошибки с type switch",
						Difficulty: "medium",
						Description: `<p>Реализуй <code>ValidationError</code> и <code>NotFoundError</code>. Функция <code>handleError</code> различает их через type switch:</p>
<p>Ввод:</p>
<pre><code>3
validation email invalid format
notfound user 42
other something wrong</code></pre>
<p>Вывод:</p>
<pre><code>Validation error on 'email': invalid format
Not found: user with id=42
Unknown error: something wrong</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "switch e := err.(type)", Definition: "Type switch для ошибок — стандартный Go-паттерн для разных типов ошибок."},
						},
						TestCases: []TestCase{
							{Input: "3\nvalidation email invalid format\nnotfound user 42\nother something wrong", ExpectedOutput: "Validation error on 'email': invalid format\nNot found: user with id=42\nUnknown error: something wrong"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type ValidationError struct{ Field, Message string }
func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

type NotFoundError struct{ Resource, ID string }
func (e *NotFoundError) Error() string { return e.Resource + " " + e.ID }

type genericError struct{ msg string }
func (e genericError) Error() string { return e.msg }

func handleError(err error) {
    // Используй type switch: case *ValidationError, case *NotFoundError, default
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        switch parts[0] {
        case "validation":
            handleError(&ValidationError{parts[1], strings.Join(parts[2:], " ")})
        case "notfound":
            handleError(&NotFoundError{parts[1], parts[2]})
        default:
            handleError(genericError{strings.Join(parts[1:], " ")})
        }
    }
}`,
						Hints: `<p>switch e := err.(type) { case *ValidationError: fmt.Printf(..., e.Field, e.Message) }</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type ValidationError struct{ Field, Message string }
func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

type NotFoundError struct{ Resource, ID string }
func (e *NotFoundError) Error() string { return e.Resource + " " + e.ID }

type genericError struct{ msg string }
func (e genericError) Error() string { return e.msg }

func handleError(err error) {
    switch e := err.(type) {
    case *ValidationError:
        fmt.Printf("Validation error on '%s': %s\n", e.Field, e.Message)
    case *NotFoundError:
        fmt.Printf("Not found: %s with id=%s\n", e.Resource, e.ID)
    default:
        fmt.Printf("Unknown error: %s\n", err)
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        switch parts[0] {
        case "validation":
            handleError(&ValidationError{parts[1], strings.Join(parts[2:], " ")})
        case "notfound":
            handleError(&NotFoundError{parts[1], parts[2]})
        default:
            handleError(genericError{strings.Join(parts[1:], " ")})
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Модуль оплат — Strategy pattern",
						Difficulty: "hard",
						Description: `<p>Реализуй систему оплат на интерфейсах (вдохновлено реальным Go-проектом). PaymentModule хранит текущий метод оплаты и выдаёт ID операциям:</p>
<p>Ввод:</p>
<pre><code>5
bank
pay Burger 5
pay Phone 500
crypto
pay Game 20</code></pre>
<p>Вывод:</p>
<pre><code>Bank: paid $5 for Burger (id=1)
Bank: paid $500 for Phone (id=2)
Crypto: paid $20 for Game (id=3)</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Strategy pattern", Definition: "PaymentModule хранит PaymentMethod как интерфейс. Смена стратегии — просто pm.method = Bank{} или Crypto{}."},
						},
						TestCases: []TestCase{
							{Input: "5\nbank\npay Burger 5\npay Phone 500\ncrypto\npay Game 20", ExpectedOutput: "Bank: paid $5 for Burger (id=1)\nBank: paid $500 for Phone (id=2)\nCrypto: paid $20 for Game (id=3)"},
						},
						StarterCode: `package main

import "fmt"

type PaymentMethod interface {
    Pay(desc string, amount int) string
}

type Bank struct{}
func (b Bank) Pay(desc string, amount int) string {
    return fmt.Sprintf("Bank: paid $%d for %s", amount, desc)
}

type Crypto struct{}
func (c Crypto) Pay(desc string, amount int) string {
    // "Crypto: paid $N for Desc"
    return ""
}

type PaymentModule struct {
    method PaymentMethod
    nextID int
}

func (p *PaymentModule) Pay(desc string, amount int) {
    p.nextID++
    fmt.Printf("%s (id=%d)\n", p.method.Pay(desc, amount), p.nextID)
}

func main() {
    var n int
    fmt.Scan(&n)
    pm := &PaymentModule{}
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "bank":   pm.method = Bank{}
        case "crypto": pm.method = Crypto{}
        case "pay":
            var desc string
            var amount int
            fmt.Scan(&desc, &amount)
            pm.Pay(desc, amount)
        }
    }
}`,
						Hints: `<p>Crypto.Pay: <code>fmt.Sprintf("Crypto: paid $%d for %s", amount, desc)</code>. pm.method — интерфейс, подменяется на лету.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type PaymentMethod interface {
    Pay(desc string, amount int) string
}

type Bank struct{}
func (b Bank) Pay(desc string, amount int) string {
    return fmt.Sprintf("Bank: paid $%d for %s", amount, desc)
}

type Crypto struct{}
func (c Crypto) Pay(desc string, amount int) string {
    return fmt.Sprintf("Crypto: paid $%d for %s", amount, desc)
}

type PaymentModule struct {
    method PaymentMethod
    nextID int
}

func (p *PaymentModule) Pay(desc string, amount int) {
    p.nextID++
    fmt.Printf("%s (id=%d)\n", p.method.Pay(desc, amount), p.nextID)
}

func main() {
    var n int
    fmt.Scan(&n)
    pm := &PaymentModule{}
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "bank":   pm.method = Bank{}
        case "crypto": pm.method = Crypto{}
        case "pay":
            var desc string; var amount int
            fmt.Scan(&desc, &amount)
            pm.Pay(desc, amount)
        }
    }
}</code></pre>`,
					},
				},
			},
		},
	}
}
