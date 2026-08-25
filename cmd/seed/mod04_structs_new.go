package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 4: Структуры и методы
// ════════════════════════════════════════════════════════════════

func mod04_structs_new() M {
	return M{
		Slug:          "structs",
		Title:         "Структуры и методы",
		Description:   "Свои типы данных: как объединить связанные поля в одну сущность и добавить поведение через методы.",
		Order:         4,
		Track:         "shared",
		Difficulty:    "beginner",
		Prerequisites: []string{"functions"},
		Lessons: []L{
			{
				Slug: "struct-basics", Title: "Структуры — свои типы данных", Order: 1,
				Difficulty: "beginner", Track: "shared",
				Content: `<h1>Структуры — свои типы данных</h1>

<h2>Зачем нужны структуры?</h2>
<p>Представь: для каждого пользователя нужно хранить имя, email, возраст. Без структур пришлось бы жонглировать отдельными переменными. <strong>Структура</strong> (struct) объединяет связанные данные в один тип.</p>

<pre><code>// Определение структуры
type User struct {
    Name  string
    Email string
    Age   int
}

// Создание
u1 := User{Name: "Alice", Email: "alice@mail.com", Age: 25}
u2 := User{"Bob", "bob@mail.com", 30}  // по порядку полей (не рекомендуется)

// Доступ к полям через точку
fmt.Println(u1.Name)   // "Alice"
u1.Age = 26            // изменение поля

// Нулевая структура — все поля zero value
var empty User  // Name="", Email="", Age=0</code></pre>

<h2>Методы — поведение структуры</h2>
<p><strong>Метод</strong> — это функция, привязанная к типу. Вместо <code>greet(user)</code> пишем <code>user.Greet()</code>:</p>

<pre><code>// Метод с value receiver (получает КОПИЮ)
func (u User) Greeting() string {
    return fmt.Sprintf("Hi, I'm %s (%d)", u.Name, u.Age)
}

// Метод с pointer receiver (получает ССЫЛКУ — может изменять)
func (u *User) Birthday() {
    u.Age++  // изменяет оригинал!
}

alice := User{Name: "Alice", Age: 25}
fmt.Println(alice.Greeting())  // "Hi, I'm Alice (25)"
alice.Birthday()
fmt.Println(alice.Age)         // 26</code></pre>

<p><strong>Когда что использовать:</strong></p>
<ul>
<li><code>(u User)</code> — value receiver: только чтение, не меняет структуру</li>
<li><code>(u *User)</code> — pointer receiver: может изменять поля структуры</li>
<li>Правило: если хотя бы один метод с pointer receiver, делай все с pointer receiver</li>
</ul>

<h2>Конструктор — функция New</h2>
<pre><code>func NewUser(name, email string, age int) *User {
    return &User{
        Name:  name,
        Email: email,
        Age:   age,
    }
}

user := NewUser("Alice", "alice@mail.com", 25)</code></pre>

<h2>Вложенные структуры</h2>
<pre><code>type Address struct {
    City    string
    Country string
}

type User struct {
    Name    string
    Address Address  // вложенная структура
}

u := User{
    Name: "Alice",
    Address: Address{City: "Moscow", Country: "Russia"},
}
fmt.Println(u.Address.City)  // "Moscow"</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Чем value receiver отличается от pointer receiver?",
						Options:     []string{"Ничем", "Value получает копию (нельзя менять оригинал), pointer получает ссылку (можно менять)", "Value быстрее", "Pointer только для чтения"},
						Correct:     1,
						Explanation: "(u User) — копия, изменения не влияют на оригинал. (u *User) — указатель на оригинал, изменения сохраняются.",
					},
					{
						Question:    "Почему в Go нет ключевого слова class?",
						Options:     []string{"Забыли добавить", "Go использует struct + методы вместо классов. Нет наследования — используется композиция", "class запланирован в будущем", "Go не поддерживает ООП"},
						Correct:     1,
						Explanation: "Go сознательно отказался от классов и наследования. Вместо этого: struct для данных, методы для поведения, интерфейсы для абстракций, композиция вместо наследования.",
					},
					{
						Question:    "Какое нулевое значение у struct, если не инициализировать?",
						Options:     []string{"nil", "Ошибка компиляции", "Все поля получают нулевые значения своих типов (0, \"\", false, nil)", "Случайные значения"},
						Correct:     2,
						Explanation: "var u User создаёт структуру где int=0, string=\"\", bool=false, pointer=nil. Это предсказуемо и безопасно — можно использовать без явной инициализации.",
					},
					{
						Question:    "Почему стиль User{\"Bob\", \"bob@mail.com\", 30} (по порядку) не рекомендуется?",
						Options:     []string{"Это медленнее", "Если добавить или переставить поля в struct — код сломается без ошибки компиляции", "Нельзя использовать в Go 1.18+", "Только для экспортированных полей"},
						Correct:     1,
						Explanation: "Позиционная инициализация хрупкая: добавил поле в середину struct — все литералы без имён молча станут неправильными. Всегда используй именованные поля: User{Name: \"Bob\", ...}.",
					},
					{
						Question:    "Зачем нужна функция-конструктор NewUser(...) *User?",
						Options:     []string{"Обязательно в Go", "Чтобы скрыть поля", "Для валидации, установки дефолтов и гарантии корректного состояния при создании", "Конструктор работает быстрее литерала"},
						Correct:     2,
						Explanation: "Конструктор — соглашение, не требование языка. Он позволяет: проверить входные данные, установить defaults, вернуть ошибку если данные некорректные. В стандартной библиотеке: http.NewRequest, os.NewFile, etc.",
					},
				},
				Tasks: []T{
					{
						Title:      "Структура Video",
						Difficulty: "easy",
						Description: `<p>Создай структуру <code>Video</code> для WatchTogether и методы для неё:</p>
<p>Ввод:</p>
<pre><code>Matrix 1999 136</code></pre>
<p>Вывод:</p>
<pre><code>Matrix (1999) - 2h 16m</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "type Name struct { ... }", Definition: "Определение нового типа-структуры. Поля перечисляются внутри фигурных скобок: имя, потом тип."},
							{Term: "func (v Video) Method()", Definition: "Метод структуры. v — receiver (получатель), через него доступны поля: v.Title, v.Year и т.д."},
						},
						TestCases: []TestCase{
							{Input: "Matrix 1999 136", ExpectedOutput: "Matrix (1999) - 2h 16m"},
							{Input: "Inception 2010 148", ExpectedOutput: "Inception (2010) - 2h 28m"},
						},
						StarterCode: `package main

import "fmt"

type Video struct {
    Title    string
    Year     int
    Duration int // minutes
}

func (v Video) FormatDuration() string {
    // Верни строку "Xh Ym"
    return ""
}

func (v Video) String() string {
    // Верни "Title (Year) - Xh Ym"
    return ""
}

func main() {
    var v Video
    fmt.Scan(&v.Title, &v.Year, &v.Duration)
    fmt.Println(v.String())
}`,
						Hints: `<p>Часы: <code>v.Duration / 60</code>. Минуты: <code>v.Duration % 60</code>. Формат: <code>fmt.Sprintf("%dh %dm", h, m)</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type Video struct {
    Title    string
    Year     int
    Duration int
}

func (v Video) FormatDuration() string {
    return fmt.Sprintf("%dh %dm", v.Duration/60, v.Duration%60)
}

func (v Video) String() string {
    return fmt.Sprintf("%s (%d) - %s", v.Title, v.Year, v.FormatDuration())
}

func main() {
    var v Video
    fmt.Scan(&v.Title, &v.Year, &v.Duration)
    fmt.Println(v.String())
}</code></pre>`,
					},
					{
						Title:      "Счётчик просмотров",
						Difficulty: "medium",
						Description: `<p>Создай структуру <code>Counter</code> с методами <code>Inc()</code>, <code>Dec()</code>, <code>Value() int</code>. Используй pointer receiver для изменения.</p>
<p>Ввод (серия команд):</p>
<pre><code>5
inc inc inc dec value</code></pre>
<p>Вывод:</p>
<pre><code>2</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "(c *Counter)", Definition: "Pointer receiver — метод получает указатель на структуру и может изменять её поля. Без * метод получит копию."},
						},
						TestCases: []TestCase{
							{Input: "5\ninc inc inc dec value", ExpectedOutput: "2"},
							{Input: "3\ninc inc value", ExpectedOutput: "2"},
						},
						StarterCode: `package main

import "fmt"

type Counter struct {
    count int
}

// Напиши методы Inc, Dec, Value с правильным receiver

func main() {
    var n int
    fmt.Scan(&n)

    c := &Counter{}
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "inc":
            c.Inc()
        case "dec":
            c.Dec()
        case "value":
            fmt.Println(c.Value())
        }
    }
}`,
						Hints: `<p><code>func (c *Counter) Inc() { c.count++ }</code>. Value может быть value receiver: <code>func (c Counter) Value() int { return c.count }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type Counter struct {
    count int
}

func (c *Counter) Inc() { c.count++ }
func (c *Counter) Dec() { c.count-- }
func (c Counter) Value() int { return c.count }

func main() {
    var n int
    fmt.Scan(&n)

    c := &Counter{}
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "inc":
            c.Inc()
        case "dec":
            c.Dec()
        case "value":
            fmt.Println(c.Value())
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Банковский счёт (реальный кейс)",
						Difficulty: "hard",
						Description: `<p>Реализуй структуру <code>Account</code> для банковского счёта — это типичная задача на собеседованиях и в реальных проектах:</p>
<ul>
<li><code>Deposit(amount float64) error</code> — пополнение (ошибка если amount <= 0)</li>
<li><code>Withdraw(amount float64) error</code> — снятие (ошибка если недостаточно средств)</li>
<li><code>Balance() float64</code> — текущий баланс</li>
</ul>
<p>Ввод (серия операций):</p>
<pre><code>4
deposit 100.50
deposit 50.00
withdraw 30.25
balance</code></pre>
<p>Вывод:</p>
<pre><code>Deposited: 100.50
Deposited: 50.00
Withdrawn: 30.25
Balance: 120.25</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "fmt.Errorf()", Definition: "Создаёт ошибку с форматированным сообщением."},
							{Term: "%.2f", Definition: "Printf-формат: дробное число с двумя знаками после точки."},
						},
						TestCases: []TestCase{
							{Input: "4\ndeposit 100.50\ndeposit 50.00\nwithdraw 30.25\nbalance", ExpectedOutput: "Deposited: 100.50\nDeposited: 50.00\nWithdrawn: 30.25\nBalance: 120.25"},
							{Input: "2\nwithdraw 50.00\nbalance", ExpectedOutput: "Error: insufficient funds\nBalance: 0.00"},
						},
						StarterCode: `package main

import "fmt"

type Account struct {
    balance float64
}

func (a *Account) Deposit(amount float64) error {
    // Проверь amount > 0, увеличь баланс
    return nil
}

func (a *Account) Withdraw(amount float64) error {
    // Проверь достаточно ли средств
    return nil
}

func (a Account) Balance() float64 {
    return a.balance
}

func main() {
    var n int
    fmt.Scan(&n)
    acc := &Account{}

    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "deposit":
            var amount float64
            fmt.Scan(&amount)
            if err := acc.Deposit(amount); err != nil {
                fmt.Printf("Error: %s\n", err)
            } else {
                fmt.Printf("Deposited: %.2f\n", amount)
            }
        case "withdraw":
            var amount float64
            fmt.Scan(&amount)
            if err := acc.Withdraw(amount); err != nil {
                fmt.Printf("Error: %s\n", err)
            } else {
                fmt.Printf("Withdrawn: %.2f\n", amount)
            }
        case "balance":
            fmt.Printf("Balance: %.2f\n", acc.Balance())
        }
    }
}`,
						Hints: `<p>Withdraw: <code>if amount > a.balance { return fmt.Errorf("insufficient funds") }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type Account struct {
    balance float64
}

func (a *Account) Deposit(amount float64) error {
    if amount <= 0 {
        return fmt.Errorf("invalid amount")
    }
    a.balance += amount
    return nil
}

func (a *Account) Withdraw(amount float64) error {
    if amount > a.balance {
        return fmt.Errorf("insufficient funds")
    }
    a.balance -= amount
    return nil
}

func (a Account) Balance() float64 { return a.balance }

func main() {
    var n int
    fmt.Scan(&n)
    acc := &Account{}

    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        switch cmd {
        case "deposit":
            var amount float64
            fmt.Scan(&amount)
            if err := acc.Deposit(amount); err != nil {
                fmt.Printf("Error: %s\n", err)
            } else {
                fmt.Printf("Deposited: %.2f\n", amount)
            }
        case "withdraw":
            var amount float64
            fmt.Scan(&amount)
            if err := acc.Withdraw(amount); err != nil {
                fmt.Printf("Error: %s\n", err)
            } else {
                fmt.Printf("Withdrawn: %.2f\n", amount)
            }
        case "balance":
            fmt.Printf("Balance: %.2f\n", acc.Balance())
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Точка на плоскости",
						Difficulty: "easy",
						Description: `<p>Реализуй структуру <code>Point</code> с координатами X, Y и методом <code>Distance() float64</code> — расстояние от начала координат.</p>
<p>Ввод: <code>3 4</code></p>
<p>Вывод: <code>Distance: 5.00</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "math.Sqrt", Definition: "Квадратный корень. Импортируй math: math.Sqrt(x)."},
							{Term: "Теорема Пифагора", Definition: "Расстояние = sqrt(x² + y²). В Go: math.Sqrt(p.X*p.X + p.Y*p.Y)."},
						},
						TestCases: []TestCase{
							{Input: "3 4", ExpectedOutput: "Distance: 5.00"},
							{Input: "0 0", ExpectedOutput: "Distance: 0.00"},
							{Input: "1 1", ExpectedOutput: "Distance: 1.41"},
						},
						StarterCode: `package main

import (
    "fmt"
    "math"
)

type Point struct {
    X, Y float64
}

func (p Point) Distance() float64 {
    // Расстояние от (0,0) до (X,Y)
    return 0
}

func main() {
    var p Point
    fmt.Scan(&p.X, &p.Y)
    fmt.Printf("Distance: %.2f\n", p.Distance())
}`,
						Hints: `<p><code>math.Sqrt(p.X*p.X + p.Y*p.Y)</code> — теорема Пифагора.</p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "math"
)

type Point struct {
    X, Y float64
}

func (p Point) Distance() float64 {
    return math.Sqrt(p.X*p.X + p.Y*p.Y)
}

func main() {
    var p Point
    fmt.Scan(&p.X, &p.Y)
    fmt.Printf("Distance: %.2f\n", p.Distance())
}</code></pre>`,
					},
					{
						Title:      "Конструктор с валидацией",
						Difficulty: "medium",
						Description: `<p>Реализуй функцию <code>NewProduct(name string, price float64) (*Product, error)</code>:</p>
<ul>
<li>Если name пустое → error "name required"</li>
<li>Если price &lt;= 0 → error "price must be positive"</li>
<li>Иначе создать продукт</li>
</ul>
<p>Ввод:</p>
<pre><code>3
Widget 9.99
 -5.00
Gadget 0</code></pre>
<p>Вывод:</p>
<pre><code>Created: Widget $9.99
Error: name required
Error: price must be positive</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "strings.TrimSpace", Definition: "Убирает пробелы по краям строки. Используй для проверки что name не пустой."},
							{Term: "(*Product, error)", Definition: "Идиома Go: возвращать объект + ошибку. При ошибке — nil, error. При успехе — *Product, nil."},
						},
						TestCases: []TestCase{
							{Input: "3\nWidget 9.99\n  -5.00\nGadget 0", ExpectedOutput: "Created: Widget $9.99\nError: name required\nError: price must be positive"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Product struct {
    Name  string
    Price float64
}

func NewProduct(name string, price float64) (*Product, error) {
    // Валидация + создание
    return nil, nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        line := strings.TrimSpace(scanner.Text())
        parts := strings.Fields(line)
        var name string
        var price float64
        if len(parts) >= 2 {
            name = parts[0]
            fmt.Sscanf(parts[1], "%f", &price)
        }
        p, err := NewProduct(name, price)
        if err != nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            fmt.Printf("Created: %s $%.2f\n", p.Name, p.Price)
        }
    }
}`,
						Hints: `<p>Проверяй <code>strings.TrimSpace(name) == ""</code> для пустого имени. Затем <code>price &lt;= 0</code>.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Product struct {
    Name  string
    Price float64
}

func NewProduct(name string, price float64) (*Product, error) {
    if strings.TrimSpace(name) == "" {
        return nil, fmt.Errorf("name required")
    }
    if price <= 0 {
        return nil, fmt.Errorf("price must be positive")
    }
    return &Product{Name: name, Price: price}, nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        line := strings.TrimSpace(scanner.Text())
        parts := strings.Fields(line)
        var name string
        var price float64
        if len(parts) >= 2 {
            name = parts[0]
            fmt.Sscanf(parts[1], "%f", &price)
        }
        p, err := NewProduct(name, price)
        if err != nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            fmt.Printf("Created: %s $%.2f\n", p.Name, p.Price)
        }
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug: "composition-embedding", Title: "Композиция и встраивание", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Композиция и встраивание (embedding)</h1>

<h2>Нет наследования — есть композиция</h2>
<p>В Go нет классов и наследования (extends/inherits). Вместо этого используется <strong>композиция</strong> — struct содержит другой struct как поле:</p>

<pre><code>// Композиция — struct как поле
type Address struct {
    City    string
    Country string
}

type User struct {
    Name    string
    Age     int
    Address Address  // вложенная структура
}

u := User{
    Name: "Alice",
    Age:  25,
    Address: Address{City: "Moscow", Country: "Russia"},
}
fmt.Println(u.Address.City)  // "Moscow"</code></pre>

<h2>Embedding — встраивание</h2>
<p><strong>Embedding</strong> — когда struct включается без имени поля. Методы и поля "поднимаются" на уровень выше:</p>

<pre><code>type Animal struct {
    Name string
}

func (a Animal) Speak() string {
    return a.Name + " makes a sound"
}

type Dog struct {
    Animal    // embedding — без имени поля!
    Breed string
}

d := Dog{
    Animal: Animal{Name: "Rex"},
    Breed:  "Labrador",
}

// Поля и методы Animal доступны напрямую:
fmt.Println(d.Name)     // "Rex" — не d.Animal.Name
fmt.Println(d.Speak())  // "Rex makes a sound"
fmt.Println(d.Breed)    // "Labrador"</code></pre>

<h2>Переопределение методов</h2>
<pre><code>type Dog struct {
    Animal
    Breed string
}

// Dog может "переопределить" метод Animal
func (d Dog) Speak() string {
    return d.Name + " says Woof!"
}

d := Dog{Animal: Animal{Name: "Rex"}, Breed: "Lab"}
fmt.Println(d.Speak())         // "Rex says Woof!" — метод Dog
fmt.Println(d.Animal.Speak())  // "Rex makes a sound" — метод Animal</code></pre>

<h2>Практический пример: базовая модель</h2>
<pre><code>// Общие поля для всех моделей БД
type BaseModel struct {
    ID        int
    CreatedAt time.Time
    UpdatedAt time.Time
}

type User struct {
    BaseModel            // embedding
    Name  string
    Email string
}

type Video struct {
    BaseModel            // те же поля ID, CreatedAt, UpdatedAt
    Title    string
    Duration int
}

u := User{
    BaseModel: BaseModel{ID: 1, CreatedAt: time.Now()},
    Name: "Alice",
}
fmt.Println(u.ID)         // 1 — через embedding
fmt.Println(u.CreatedAt)  // time.Time</code></pre>

<h2>Когда embedding, когда обычное поле?</h2>
<ul>
<li><strong>Embedding:</strong> когда внутренний тип — это "часть" внешнего (User IS-A BaseModel)</li>
<li><strong>Обычное поле:</strong> когда это "содержит" (User HAS-A Address)</li>
<li><strong>Правило:</strong> embedding реже чем кажется. Композиция через поля — чаще правильный выбор</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Чем embedding отличается от обычного поля?",
						Options:     []string{"Ничем", "При embedding методы и поля встроенного типа доступны напрямую без указания имени поля", "Embedding быстрее", "Обычное поле нельзя использовать"},
						Correct:     1,
						Explanation: "Embedding 'поднимает' поля и методы. d.Name вместо d.Animal.Name. Это не наследование — это синтаксический сахар для композиции.",
					},
					{
						Question:    "Зачем Go отказался от наследования?",
						Options:     []string{"Не успели добавить", "Наследование создаёт хрупкие иерархии. Композиция гибче — можно комбинировать поведения без жёсткой иерархии", "Наследование медленнее", "Go не поддерживает ООП"},
						Correct:     1,
						Explanation: "Проблема наследования: глубокие иерархии, fragile base class, diamond problem. Композиция: плоская структура, явные зависимости, легко тестировать.",
					},
					{
						Question:    "Что такое 'продвинутые поля' (promoted fields) при embedding?",
						Options:     []string{"Поля с большой буквы", "Поля встроенного типа, доступные напрямую через внешнюю структуру", "Поля с тегами json", "Экспортированные поля"},
						Correct:     1,
						Explanation: "Когда Dog встраивает Animal, все поля и методы Animal 'продвигаются' в Dog. d.Name работает как d.Animal.Name. Но при конфликте имён нужно явно указывать: d.Animal.Name.",
					},
					{
						Question:    "Можно ли встраивать несколько типов в одну struct?",
						Options:     []string{"Нет, только один", "Да — Go поддерживает множественное embedding", "Только интерфейсы", "Только если типы из одного пакета"},
						Correct:     1,
						Explanation: "type ReadWriter struct { io.Reader; io.Writer } — классический пример из стандартной библиотеки. Методы обоих типов продвигаются в ReadWriter.",
					},
					{
						Question:    "Когда использовать embedding вместо обычного поля?",
						Options:     []string{"Всегда — это быстрее", "Когда внутренний тип — это часть внешнего (IS-A), а не то что он содержит (HAS-A)", "Только для интерфейсов", "Когда нужно скрыть поля"},
						Correct:     1,
						Explanation: "User HAS-A Address → обычное поле address Address. User IS-A BaseModel (с общими ID/timestamps) → embedding BaseModel. Ошибочный embedding создаёт запутанные API и утечку деталей реализации.",
					},
				},
				Tasks: []T{
					{
						Title:      "Модель данных WatchTogether",
						Difficulty: "medium",
						Description: `<p>Создай структуры с embedding для WatchTogether:</p>
<ul>
<li><code>BaseModel</code> с ID и CreatedAt</li>
<li><code>Video</code> с Title, Duration</li>
<li><code>Room</code> с Name, VideoID</li>
</ul>
<p>Ввод: <code>1 Matrix 136</code></p>
<p>Вывод: <code>Video #1: Matrix (2h16m)</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "BaseModel (embedding)", Definition: "Общие поля (ID, CreatedAt) для всех моделей. Вставляется без имени — поля доступны напрямую."},
						},
						TestCases: []TestCase{
							{Input: "1 Matrix 136", ExpectedOutput: "Video #1: Matrix (2h16m)"},
							{Input: "5 Inception 148", ExpectedOutput: "Video #5: Inception (2h28m)"},
						},
						StarterCode: `package main

import "fmt"

type BaseModel struct {
    ID int
}

type Video struct {
    BaseModel
    Title    string
    Duration int
}

func (v Video) String() string {
    // Верни "Video #ID: Title (XhYm)"
    return ""
}

func main() {
    var v Video
    fmt.Scan(&v.ID, &v.Title, &v.Duration)
    fmt.Println(v.String())
}`,
						Hints: `<p><code>fmt.Sprintf("Video #%d: %s (%dh%02dm)", v.ID, v.Title, v.Duration/60, v.Duration%60)</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type BaseModel struct { ID int }
type Video struct {
    BaseModel
    Title    string
    Duration int
}

func (v Video) String() string {
    return fmt.Sprintf("Video #%d: %s (%dh%02dm)", v.ID, v.Title, v.Duration/60, v.Duration%60)
}

func main() {
    var v Video
    fmt.Scan(&v.ID, &v.Title, &v.Duration)
    fmt.Println(v.String())
}</code></pre>`,
					},
					{
						Title:      "Logger с prefixом через embedding",
						Difficulty: "easy",
						Description: `<p>Реализуй структуру <code>Logger</code> с полем Prefix и методами <code>Info(msg string)</code> и <code>Error(msg string)</code>:</p>
<p>Ввод:</p>
<pre><code>3
info Server started
error Connection failed
info Done</code></pre>
<p>Вывод:</p>
<pre><code>[INFO] app: Server started
[ERROR] app: Connection failed
[INFO] app: Done</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "struct field", Definition: "Поле структуры. Logger{Prefix: \"app\"} — создание с именованным полем."},
						},
						TestCases: []TestCase{
							{Input: "3\ninfo Server started\nerror Connection failed\ninfo Done", ExpectedOutput: "[INFO] app: Server started\n[ERROR] app: Connection failed\n[INFO] app: Done"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Logger struct {
    Prefix string
}

func (l Logger) Info(msg string) {
    fmt.Printf("[INFO] %s: %s\n", l.Prefix, msg)
}

func (l Logger) Error(msg string) {
    // Выведи [ERROR] prefix: msg
}

func main() {
    log := Logger{Prefix: "app"}
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        if len(parts) < 2 { continue }
        switch parts[0] {
        case "info":  log.Info(parts[1])
        case "error": log.Error(parts[1])
        }
    }
}`,
						Hints: `<p><code>fmt.Printf("[ERROR] %s: %s\n", l.Prefix, msg)</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Logger struct {
    Prefix string
}

func (l Logger) Info(msg string)  { fmt.Printf("[INFO] %s: %s\n", l.Prefix, msg) }
func (l Logger) Error(msg string) { fmt.Printf("[ERROR] %s: %s\n", l.Prefix, msg) }

func main() {
    log := Logger{Prefix: "app"}
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        if len(parts) < 2 { continue }
        switch parts[0] {
        case "info":  log.Info(parts[1])
        case "error": log.Error(parts[1])
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Переопределение метода",
						Difficulty: "easy",
						Description: `<p>Создай иерархию через embedding: Animal со Speak(), Dog и Cat переопределяют Speak():</p>
<p>Ввод:</p>
<pre><code>3
dog Rex
cat Luna
animal Generic</code></pre>
<p>Вывод:</p>
<pre><code>Rex says Woof!
Luna says Meow!
Generic makes a sound</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "method override via embedding", Definition: "Когда Dog определяет свой Speak(), он перекрывает Animal.Speak(). Для явного вызова: d.Animal.Speak()."},
						},
						TestCases: []TestCase{
							{Input: "3\ndog Rex\ncat Luna\nanimal Generic", ExpectedOutput: "Rex says Woof!\nLuna says Meow!\nGeneric makes a sound"},
						},
						StarterCode: `package main

import "fmt"

type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct{ Animal }
func (d Dog) Speak() string {
    return "" // "Name says Woof!"
}

type Cat struct{ Animal }
func (c Cat) Speak() string {
    return "" // "Name says Meow!"
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind, name string
        fmt.Scan(&kind, &name)
        switch kind {
        case "dog":    fmt.Println(Dog{Animal{name}}.Speak())
        case "cat":    fmt.Println(Cat{Animal{name}}.Speak())
        case "animal": fmt.Println(Animal{name}.Speak())
        }
    }
}`,
						Hints: `<p><code>d.Name + " says Woof!"</code> — поле Name доступно через embedding.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type Animal struct{ Name string }
func (a Animal) Speak() string { return a.Name + " makes a sound" }

type Dog struct{ Animal }
func (d Dog) Speak() string { return d.Name + " says Woof!" }

type Cat struct{ Animal }
func (c Cat) Speak() string { return c.Name + " says Meow!" }

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind, name string
        fmt.Scan(&kind, &name)
        switch kind {
        case "dog":    fmt.Println(Dog{Animal{name}}.Speak())
        case "cat":    fmt.Println(Cat{Animal{name}}.Speak())
        case "animal": fmt.Println(Animal{name}.Speak())
        }
    }
}</code></pre>`,
					},
					{
						Title:      "BaseModel для сущностей",
						Difficulty: "medium",
						Description: `<p>Реализуй <code>BaseModel</code> с ID и методом <code>IsNew() bool</code> (true если ID == 0). Создай <code>User</code> и <code>Post</code> с embedding. Выведи статус:</p>
<p>Ввод:</p>
<pre><code>3
user 0 Alice
post 5 Hello
user 3 Bob</code></pre>
<p>Вывод:</p>
<pre><code>User Alice: new
Post Hello: saved (id=5)
User Bob: saved (id=3)</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "IsNew() bool", Definition: "Идиома: метод проверяет нулевое состояние. ID==0 означает что запись не сохранена в БД."},
						},
						TestCases: []TestCase{
							{Input: "3\nuser 0 Alice\npost 5 Hello\nuser 3 Bob", ExpectedOutput: "User Alice: new\nPost Hello: saved (id=5)\nUser Bob: saved (id=3)"},
						},
						StarterCode: `package main

import "fmt"

type BaseModel struct {
    ID int
}

func (b BaseModel) IsNew() bool {
    return b.ID == 0
}

type User struct {
    BaseModel
    Name string
}

type Post struct {
    BaseModel
    Title string
}

func printStatus(id int, isNew bool, label, name string) {
    if isNew {
        fmt.Printf("%s %s: new\n", label, name)
    } else {
        fmt.Printf("%s %s: saved (id=%d)\n", label, name, id)
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind string
        var id int
        var name string
        fmt.Scan(&kind, &id, &name)
        switch kind {
        case "user":
            u := User{BaseModel{id}, name}
            printStatus(u.ID, u.IsNew(), "User", u.Name)
        case "post":
            p := Post{BaseModel{id}, name}
            printStatus(p.ID, p.IsNew(), "Post", p.Title)
        }
    }
}`,
						Hints: `<p><code>u.IsNew()</code> — метод доступен через embedding. <code>u.ID</code> — поле тоже.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type BaseModel struct{ ID int }
func (b BaseModel) IsNew() bool { return b.ID == 0 }

type User struct { BaseModel; Name string }
type Post struct { BaseModel; Title string }

func printStatus(id int, isNew bool, label, name string) {
    if isNew {
        fmt.Printf("%s %s: new\n", label, name)
    } else {
        fmt.Printf("%s %s: saved (id=%d)\n", label, name, id)
    }
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var kind string; var id int; var name string
        fmt.Scan(&kind, &id, &name)
        switch kind {
        case "user":
            u := User{BaseModel{id}, name}
            printStatus(u.ID, u.IsNew(), "User", u.Name)
        case "post":
            p := Post{BaseModel{id}, name}
            printStatus(p.ID, p.IsNew(), "Post", p.Title)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "HTTP-конфигурация через builder pattern",
						Difficulty: "hard",
						Description: `<p>Реализуй конфигурацию через embedding и builder pattern (методы возвращают копию):</p>
<ul>
<li><code>BaseConfig</code>: Timeout int, Retries int</li>
<li><code>HTTPConfig</code>: BaseConfig + Host, Port</li>
<li><code>WithTimeout(t int) HTTPConfig</code> — копия с новым Timeout</li>
<li><code>Address() string</code> — "host:port"</li>
<li><code>String() string</code> — полное описание</li>
</ul>
<p>Ввод: <code>api.example.com 8080 30 3</code></p>
<p>Вывод:</p>
<pre><code>Address: api.example.com:8080
Config: api.example.com:8080 timeout=30s retries=3
With 60s timeout: api.example.com:8080 timeout=60s retries=3</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Builder pattern", Definition: "Метод с value receiver возвращает изменённую копию. Позволяет цепочки: cfg.WithTimeout(30).WithRetries(5)."},
						},
						TestCases: []TestCase{
							{Input: "api.example.com 8080 30 3", ExpectedOutput: "Address: api.example.com:8080\nConfig: api.example.com:8080 timeout=30s retries=3\nWith 60s timeout: api.example.com:8080 timeout=60s retries=3"},
						},
						StarterCode: `package main

import "fmt"

type BaseConfig struct {
    Timeout int
    Retries int
}

type HTTPConfig struct {
    BaseConfig
    Host string
    Port int
}

func (c HTTPConfig) Address() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c HTTPConfig) String() string {
    return fmt.Sprintf("%s timeout=%ds retries=%d", c.Address(), c.Timeout, c.Retries)
}

func (c HTTPConfig) WithTimeout(t int) HTTPConfig {
    c.Timeout = t
    return c
}

func main() {
    var cfg HTTPConfig
    fmt.Scan(&cfg.Host, &cfg.Port, &cfg.Timeout, &cfg.Retries)

    fmt.Println("Address:", cfg.Address())
    fmt.Println("Config:", cfg.String())
    fmt.Println("With 60s timeout:", cfg.WithTimeout(60).String())
}`,
						Hints: `<p>Value receiver создаёт копию автоматически. <code>c.Timeout = t; return c</code> — изменяем копию, возвращаем.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type BaseConfig struct {
    Timeout int
    Retries int
}

type HTTPConfig struct {
    BaseConfig
    Host string
    Port int
}

func (c HTTPConfig) Address() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c HTTPConfig) String() string {
    return fmt.Sprintf("%s timeout=%ds retries=%d", c.Address(), c.Timeout, c.Retries)
}

func (c HTTPConfig) WithTimeout(t int) HTTPConfig {
    c.Timeout = t
    return c
}

func main() {
    var cfg HTTPConfig
    fmt.Scan(&cfg.Host, &cfg.Port, &cfg.Timeout, &cfg.Retries)
    fmt.Println("Address:", cfg.Address())
    fmt.Println("Config:", cfg.String())
    fmt.Println("With 60s timeout:", cfg.WithTimeout(60).String())
}</code></pre>`,
					},
				},
			},
			{
				Slug: "struct-tags-json", Title: "Теги структур и JSON", Order: 3,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Теги структур и JSON</h1>

<h2>Что такое теги?</h2>
<p>Тег — строковые метаданные поля, записанные в обратных кавычках. Используются пакетами через рефлексию (<code>reflect</code>). Самый частый — <code>json:</code>:</p>
<pre><code>type User struct {
    ID       int    ` + "`json:\"id\"`" + `
    Name     string ` + "`json:\"name\"`" + `
    Password string ` + "`json:\"-\"`" + `         // никогда не сериализуется!
    Email    string ` + "`json:\"email,omitempty\"`" + ` // пропустить если пустое
}</code></pre>

<h2>json.Marshal — структура → JSON</h2>
<pre><code>import "encoding/json"

user := User{ID: 1, Name: "Alice", Email: "alice@mail.com"}
data, err := json.Marshal(user)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
// {"id":1,"name":"Alice","email":"alice@mail.com"}
// Password не попал — тег json:"-"</code></pre>

<h2>json.Unmarshal — JSON → структура</h2>
<pre><code>jsonStr := ` + "`" + `{"id":2,"name":"Bob","email":"bob@mail.com"}` + "`" + `

var user User
err := json.Unmarshal([]byte(jsonStr), &user)
if err != nil {
    log.Fatal(err)
}
fmt.Println(user.Name)   // "Bob"
fmt.Println(user.ID)     // 2</code></pre>

<h2>Опции тегов json</h2>
<pre><code>type Article struct {
    ID        int    ` + "`json:\"id\"`" + `              // переименовать в JSON
    Title     string ` + "`json:\"title\"`" + `
    Body      string ` + "`json:\"body,omitempty\"`" + `  // пропустить если ""
    Internal  string ` + "`json:\"-\"`" + `              // никогда не включать
    CreatedAt string ` + "`json:\"created_at\"`" + `     // snake_case в JSON
}
// Без тегов Go использует имя поля как есть: "ID", "Title"</code></pre>

<h2>Вложенный JSON</h2>
<pre><code>type Address struct {
    City    string ` + "`json:\"city\"`" + `
    Country string ` + "`json:\"country\"`" + `
}

type User struct {
    Name    string  ` + "`json:\"name\"`" + `
    Address Address ` + "`json:\"address\"`" + `
}

u := User{Name: "Alice", Address: Address{City: "Moscow", Country: "Russia"}}
data, _ := json.Marshal(u)
// {"name":"Alice","address":{"city":"Moscow","country":"Russia"}}</code></pre>

<h2>Encoder/Decoder для потоков</h2>
<pre><code>// Для HTTP-ответов — удобнее энкодер (нет лишней аллокации)
func handler(w http.ResponseWriter, r *http.Request) {
    user := User{ID: 1, Name: "Alice"}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

// Для чтения из запроса:
var user User
json.NewDecoder(r.Body).Decode(&user)</code></pre>

<h2>Типичные ошибки</h2>
<pre><code>// ОШИБКА: неэкспортированное поле — json не видит через reflect
type Bad struct {
    name string ` + "`json:\"name\"`" + ` // НЕ РАБОТАЕТ!
}

// ПРАВИЛЬНО:
type Good struct {
    Name string ` + "`json:\"name\"`" + ` // OK — заглавная буква

// ОШИБКА: игнорировать ошибку
json.Unmarshal(data, &user)

// ПРАВИЛЬНО:
if err := json.Unmarshal(data, &user); err != nil {
    return fmt.Errorf("invalid json: %w", err)
}</code></pre>

<h2>Читать глубже</h2>
<ul>
<li><a href="https://metanit.com/go/golang/10.1.php" target="_blank">Metanit: JSON в Go</a> — подробно про Marshal/Unmarshal</li>
<li><a href="https://habr.com/ru/articles/502822/" target="_blank">Хабр: Работа с JSON в Go</a> — продвинутые техники</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Что означает тег `json:\"-\"`?",
						Options:     []string{"Поле будет первым в JSON", "Поле никогда не включается в JSON (ни Marshal, ни Unmarshal)", "Поле станет числом", "Ошибка синтаксиса"},
						Correct:     1,
						Explanation: "json:\"-\" полностью исключает поле. Используется для паролей, токенов, internal-данных. Даже если JSON содержит это поле — Unmarshal проигнорирует.",
					},
					{
						Question:    "Что делает опция omitempty в теге `json:\"email,omitempty\"`?",
						Options:     []string{"Требует обязательного заполнения", "Пропускает поле если значение нулевое (\"\", 0, false, nil)", "Делает поле nullable", "Только для строк"},
						Correct:     1,
						Explanation: "omitempty пропускает поле при Marshal если значение — zero value. Удобно чтобы не отправлять лишние поля в API-ответах. Для указателей: nil = omit, &value = include.",
					},
					{
						Question:    "Почему json не сериализует поле с маленькой буквы?",
						Options:     []string{"Это баг Go", "Неэкспортированные поля недоступны вне пакета — reflect тоже не может их прочитать", "json не поддерживает такие поля", "Нужен тег json:\"export\""},
						Correct:     1,
						Explanation: "reflect.Value.Field() возвращает только экспортированные поля. encoding/json использует reflect — поля должны начинаться с заглавной буквы. Тег json:\"name\" только переименовывает в JSON.",
					},
					{
						Question:    "json.NewEncoder(w).Encode(v) vs json.Marshal(v) — в чём разница?",
						Options:     []string{"Нет разницы", "Marshal создаёт []byte. Encoder пишет напрямую в io.Writer без промежуточного буфера", "Encoder медленнее", "Marshal для struct, Encoder для slice"},
						Correct:     1,
						Explanation: "Encoder эффективнее для HTTP — нет аллокации []byte. Marshal удобен когда нужен именно []byte (логирование, хранение). Encoder добавляет \\n после JSON.",
					},
					{
						Question:    "Если в Go поле называется CreatedAt, как его назвать в JSON по соглашению?",
						Options:     []string{"CreatedAt", "createdat", "created_at — через тег `json:\"created_at\"`", "createdAt"},
						Correct:     2,
						Explanation: "Соглашение: Go — PascalCase (CreatedAt), JSON — snake_case (created_at). Тег json:\"created_at\" делает переименование. Это стандарт при разработке REST API на Go.",
					},
				},
				Tasks: []T{
					{
						Title:      "Сериализация пользователя",
						Difficulty: "easy",
						Description: `<p>Создай структуру <code>User</code> с тегами json и выведи JSON. Password должен быть скрыт:</p>
<p>Ввод: <code>1 Alice alice@mail.com secret123</code></p>
<p>Вывод: <code>{"id":1,"name":"Alice","email":"alice@mail.com"}</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "json.Marshal", Definition: "Сериализует Go-значение в JSON. Возвращает []byte и error."},
							{Term: "json:\"-\"", Definition: "Исключает поле из JSON. Password никогда не попадёт в API-ответ."},
						},
						TestCases: []TestCase{
							{Input: "1 Alice alice@mail.com secret123", ExpectedOutput: `{"id":1,"name":"Alice","email":"alice@mail.com"}`},
						},
						StarterCode: `package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID       int    ` + "`json:\"id\"`" + `
    Name     string ` + "`json:\"name\"`" + `
    Email    string ` + "`json:\"email\"`" + `
    Password string ` + "`json:\"-\"`" + `
}

func main() {
    var u User
    fmt.Scan(&u.ID, &u.Name, &u.Email, &u.Password)
    data, err := json.Marshal(u)
    if err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(string(data))
}`,
						Hints: `<p>Password с тегом <code>json:"-"</code> не попадёт в вывод. <code>string(data)</code> — байты в строку.</p>`,
						Solution: `<pre><code>package main

import (
    "encoding/json"
    "fmt"
)

type User struct {
    ID       int    ` + "`json:\"id\"`" + `
    Name     string ` + "`json:\"name\"`" + `
    Email    string ` + "`json:\"email\"`" + `
    Password string ` + "`json:\"-\"`" + `
}

func main() {
    var u User
    fmt.Scan(&u.ID, &u.Name, &u.Email, &u.Password)
    data, _ := json.Marshal(u)
    fmt.Println(string(data))
}</code></pre>`,
					},
					{
						Title:      "Десериализация JSON",
						Difficulty: "easy",
						Description: `<p>Прочитай JSON-строку и выведи данные:</p>
<p>Ввод: <code>{"id":42,"name":"Bob","email":"bob@mail.com"}</code></p>
<p>Вывод:</p>
<pre><code>ID: 42
Name: Bob
Email: bob@mail.com</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "json.Unmarshal", Definition: "Десериализует JSON в Go-структуру. Принимает []byte и указатель."},
							{Term: "[]byte(str)", Definition: "Преобразовать строку в байты. json.Unmarshal требует []byte."},
						},
						TestCases: []TestCase{
							{Input: `{"id":42,"name":"Bob","email":"bob@mail.com"}`, ExpectedOutput: "ID: 42\nName: Bob\nEmail: bob@mail.com"},
						},
						StarterCode: `package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type User struct {
    ID    int    ` + "`json:\"id\"`" + `
    Name  string ` + "`json:\"name\"`" + `
    Email string ` + "`json:\"email\"`" + `
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()

    var u User
    if err := json.Unmarshal([]byte(scanner.Text()), &u); err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("ID: %d\nName: %s\nEmail: %s\n", u.ID, u.Name, u.Email)
}`,
						Hints: `<p><code>json.Unmarshal([]byte(line), &u)</code> — передавай указатель на структуру.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type User struct {
    ID    int    ` + "`json:\"id\"`" + `
    Name  string ` + "`json:\"name\"`" + `
    Email string ` + "`json:\"email\"`" + `
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    var u User
    if err := json.Unmarshal([]byte(scanner.Text()), &u); err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Printf("ID: %d\nName: %s\nEmail: %s\n", u.ID, u.Name, u.Email)
}</code></pre>`,
					},
					{
						Title:      "omitempty в API-ответе",
						Difficulty: "medium",
						Description: `<p>Реализуй <code>APIResponse</code> с omitempty — пустые поля не должны попадать в JSON:</p>
<p>Ввод:</p>
<pre><code>2
ok 200
error 0 User not found</code></pre>
<p>Вывод:</p>
<pre><code>{"status":"ok","code":200}
{"status":"error","message":"User not found"}</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "omitempty", Definition: "Пропускает поле если zero value. int 0 → пропустить, string \"\" → пропустить."},
						},
						TestCases: []TestCase{
							{Input: "2\nok 200\nerror 0 User not found", ExpectedOutput: `{"status":"ok","code":200}` + "\n" + `{"status":"error","message":"User not found"}`},
						},
						StarterCode: `package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type APIResponse struct {
    Status  string ` + "`json:\"status\"`" + `
    Code    int    ` + "`json:\"code,omitempty\"`" + `
    Message string ` + "`json:\"message,omitempty\"`" + `
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        if len(parts) < 2 { continue }
        r := APIResponse{Status: parts[0]}
        fmt.Sscanf(parts[1], "%d", &r.Code)
        if len(parts) > 2 {
            r.Message = strings.Join(parts[2:], " ")
        }
        data, _ := json.Marshal(r)
        fmt.Println(string(data))
    }
}`,
						Hints: `<p>Code=0 с omitempty → пропускается. Message="" с omitempty → пропускается.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type APIResponse struct {
    Status  string ` + "`json:\"status\"`" + `
    Code    int    ` + "`json:\"code,omitempty\"`" + `
    Message string ` + "`json:\"message,omitempty\"`" + `
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        if len(parts) < 2 { continue }
        r := APIResponse{Status: parts[0]}
        fmt.Sscanf(parts[1], "%d", &r.Code)
        if len(parts) > 2 { r.Message = strings.Join(parts[2:], " ") }
        data, _ := json.Marshal(r)
        fmt.Println(string(data))
    }
}</code></pre>`,
					},
					{
						Title:      "Конфиг из JSON",
						Difficulty: "medium",
						Description: `<p>Прочитай конфиг из JSON и выведи строку подключения к БД:</p>
<p>Ввод: <code>{"host":"localhost","port":5432,"name":"mydb","user":"admin","password":"secret"}</code></p>
<p>Вывод: <code>postgres://admin:secret@localhost:5432/mydb</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "DSN", Definition: "Data Source Name — строка подключения. PostgreSQL: postgres://user:pass@host:port/dbname."},
						},
						TestCases: []TestCase{
							{Input: `{"host":"localhost","port":5432,"name":"mydb","user":"admin","password":"secret"}`, ExpectedOutput: "postgres://admin:secret@localhost:5432/mydb"},
						},
						StarterCode: `package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type DBConfig struct {
    Host     string ` + "`json:\"host\"`" + `
    Port     int    ` + "`json:\"port\"`" + `
    Name     string ` + "`json:\"name\"`" + `
    User     string ` + "`json:\"user\"`" + `
    Password string ` + "`json:\"password\"`" + `
}

func (c DBConfig) DSN() string {
    return ""
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    var cfg DBConfig
    if err := json.Unmarshal([]byte(scanner.Text()), &cfg); err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(cfg.DSN())
}`,
						Hints: `<p><code>fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Name)</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
)

type DBConfig struct {
    Host     string ` + "`json:\"host\"`" + `
    Port     int    ` + "`json:\"port\"`" + `
    Name     string ` + "`json:\"name\"`" + `
    User     string ` + "`json:\"user\"`" + `
    Password string ` + "`json:\"password\"`" + `
}

func (c DBConfig) DSN() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s", c.User, c.Password, c.Host, c.Port, c.Name)
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    var cfg DBConfig
    if err := json.Unmarshal([]byte(scanner.Text()), &cfg); err != nil {
        fmt.Println("Error:", err)
        return
    }
    fmt.Println(cfg.DSN())
}</code></pre>`,
					},
					{
						Title:      "REST API-ответ с вложенным JSON",
						Difficulty: "hard",
						Description: `<p>Реализуй сериализацию вложенного API-ответа — успешный с данными и ошибочный:</p>
<p>Ввод:</p>
<pre><code>3
success 1 Alice alice@mail.com
success 2 Bob bob@mail.com
error not_found User does not exist</code></pre>
<p>Вывод:</p>
<pre><code>{"ok":true,"data":{"id":1,"name":"Alice","email":"alice@mail.com"}}
{"ok":true,"data":{"id":2,"name":"Bob","email":"bob@mail.com"}}
{"ok":false,"error":{"code":"not_found","message":"User does not exist"}}</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Вложенный JSON", Definition: "Struct внутри struct с тегами json. Go рекурсивно сериализует все уровни вложенности."},
						},
						TestCases: []TestCase{
							{Input: "3\nsuccess 1 Alice alice@mail.com\nsuccess 2 Bob bob@mail.com\nerror not_found User does not exist",
								ExpectedOutput: `{"ok":true,"data":{"id":1,"name":"Alice","email":"alice@mail.com"}}` + "\n" +
									`{"ok":true,"data":{"id":2,"name":"Bob","email":"bob@mail.com"}}` + "\n" +
									`{"ok":false,"error":{"code":"not_found","message":"User does not exist"}}`},
						},
						StarterCode: `package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type UserData struct {
    ID    int    ` + "`json:\"id\"`" + `
    Name  string ` + "`json:\"name\"`" + `
    Email string ` + "`json:\"email\"`" + `
}

type ErrorData struct {
    Code    string ` + "`json:\"code\"`" + `
    Message string ` + "`json:\"message\"`" + `
}

type SuccessResponse struct {
    OK   bool     ` + "`json:\"ok\"`" + `
    Data UserData ` + "`json:\"data\"`" + `
}

type ErrorResponse struct {
    OK    bool      ` + "`json:\"ok\"`" + `
    Error ErrorData ` + "`json:\"error\"`" + `
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        if parts[0] == "success" {
            var id int
            fmt.Sscanf(parts[1], "%d", &id)
            r := SuccessResponse{OK: true, Data: UserData{id, parts[2], parts[3]}}
            data, _ := json.Marshal(r)
            fmt.Println(string(data))
        } else {
            r := ErrorResponse{OK: false, Error: ErrorData{parts[1], strings.Join(parts[2:], " ")}}
            data, _ := json.Marshal(r)
            fmt.Println(string(data))
        }
    }
}`,
						Hints: `<p>Два разных struct для успеха и ошибки. Go сериализует вложенные struct рекурсивно.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "os"
    "strings"
)

type UserData struct {
    ID    int    ` + "`json:\"id\"`" + `
    Name  string ` + "`json:\"name\"`" + `
    Email string ` + "`json:\"email\"`" + `
}

type ErrorData struct {
    Code    string ` + "`json:\"code\"`" + `
    Message string ` + "`json:\"message\"`" + `
}

type SuccessResponse struct {
    OK   bool     ` + "`json:\"ok\"`" + `
    Data UserData ` + "`json:\"data\"`" + `
}

type ErrorResponse struct {
    OK    bool      ` + "`json:\"ok\"`" + `
    Error ErrorData ` + "`json:\"error\"`" + `
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        if parts[0] == "success" {
            var id int
            fmt.Sscanf(parts[1], "%d", &id)
            r := SuccessResponse{OK: true, Data: UserData{id, parts[2], parts[3]}}
            data, _ := json.Marshal(r)
            fmt.Println(string(data))
        } else {
            r := ErrorResponse{OK: false, Error: ErrorData{parts[1], strings.Join(parts[2:], " ")}}
            data, _ := json.Marshal(r)
            fmt.Println(string(data))
        }
    }
}</code></pre>`,
					},
				},
			},
		},
	}
}
