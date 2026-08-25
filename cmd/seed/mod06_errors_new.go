package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 6: Обработка ошибок
// ════════════════════════════════════════════════════════════════

func mod06_errors_new() M {
	return M{
		Slug:          "errors",
		Title:         "Обработка ошибок",
		Description:   "Как Go обрабатывает ошибки без исключений. Паттерн val, err, создание своих ошибок, errors.Is/As.",
		Order:         6,
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"interfaces"},
		Lessons: []L{
			{
				Slug: "error-basics", Title: "Ошибки в Go — нет исключений!", Order: 1,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Ошибки в Go — нет исключений!</h1>

<h2>Философия ошибок в Go</h2>
<p>В большинстве языков (Python, Java, JavaScript) ошибки обрабатываются через <strong>исключения</strong> (try-catch). Go пошёл другим путём: ошибка — это <strong>обычное значение</strong>, которое возвращается из функции.</p>

<pre><code>// В Python:
// try:
//     result = dangerous_operation()
// except ValueError as e:
//     print(f"Error: {e}")

// В Go:
result, err := dangerousOperation()
if err != nil {
    fmt.Println("Error:", err)
    return
}
// Продолжаем работать с result</code></pre>

<h2>Паттерн val, err</h2>
<p>Это самый частый паттерн в Go. Функция возвращает результат И ошибку:</p>

<pre><code>import "strconv"

// strconv.Atoi возвращает (int, error)
num, err := strconv.Atoi("42")
if err != nil {
    fmt.Println("Not a number:", err)
    return
}
fmt.Println("Number:", num)  // 42

// Если строка не число:
num, err = strconv.Atoi("hello")
// num = 0, err = &NumError{...}
if err != nil {
    fmt.Println("Error:", err)  // Error: strconv.Atoi: parsing "hello": invalid syntax
}</code></pre>

<h2>Что такое error?</h2>
<pre><code>// error — это интерфейс с одним методом:
type error interface {
    Error() string
}

// Создать простую ошибку:
err := fmt.Errorf("user %s not found", username)
err := errors.New("something went wrong")</code></pre>

<h2>Свои ошибки</h2>
<pre><code>// Через fmt.Errorf с контекстом
func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("cannot divide %f by zero", a)
    }
    return a / b, nil
}

// Свой тип ошибки
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation failed: %s - %s", e.Field, e.Message)
}</code></pre>

<h2>Оборачивание ошибок (wrapping)</h2>
<pre><code>import "errors"

// %w оборачивает ошибку — сохраняет оригинал внутри
func getUser(id int) (*User, error) {
    user, err := db.FindByID(id)
    if err != nil {
        return nil, fmt.Errorf("getUser(%d): %w", id, err)
    }
    return user, nil
}

// Проверка типа ошибки
if errors.Is(err, sql.ErrNoRows) {
    fmt.Println("User not found")
}

// Извлечение конкретного типа
var valErr *ValidationError
if errors.As(err, &valErr) {
    fmt.Printf("Invalid field: %s\n", valErr.Field)
}</code></pre>

<h2>Главные правила</h2>
<ul>
<li>Всегда проверяй err != nil</li>
<li>Не игнорируй ошибки: <code>result, _ := f()</code> — только если точно знаешь что делаешь</li>
<li>Добавляй контекст: <code>fmt.Errorf("открытие файла %s: %w", name, err)</code></li>
<li>Обрабатывай ошибку один раз — либо логируй, либо возвращай, не оба сразу</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Как Go обрабатывает ошибки?",
						Options:     []string{"Через try-catch", "Ошибка — обычное значение, возвращается из функции как (result, error)", "Через panic/recover", "Игнорирует ошибки"},
						Correct:     1,
						Explanation: "Go использует явную обработку ошибок через возвращаемые значения. Каждая функция, которая может завершиться с ошибкой, возвращает error как последний результат.",
					},
					{
						Question:    "Что делает %w в fmt.Errorf?",
						Options:     []string{"Форматирует как строку", "Оборачивает ошибку — сохраняет оригинал для проверки через errors.Is/As", "Печатает предупреждение", "Создаёт панику"},
						Correct:     1,
						Explanation: "%w (wrap) оборачивает ошибку в новую с контекстом, но сохраняет оригинальную ошибку внутри. Потом можно проверить errors.Is(err, originalErr).",
					},
					{
						Question:    "Почему val, _ := f() — плохая практика?",
						Options:     []string{"Go не позволяет это", "Ошибка игнорируется — баг может остаться незамеченным. Допустимо только если точно знаешь что ошибки не будет", "_ медленнее", "Нужно писать val, err = f()"},
						Correct:     1,
						Explanation: "_ = ошибка = 'мне всё равно'. Если функция неожиданно вернёт ошибку — программа продолжит с нулевым значением val, что может привести к panic или скрытым багам.",
					},
					{
						Question:    "Что лучше: логировать ошибку или оборачивать и возвращать?",
						Options:     []string{"Всегда логировать", "Всегда возвращать", "Обрабатывать один раз: либо логируем и обрабатываем, либо оборачиваем и возвращаем. Не оба сразу", "Не важно"},
						Correct:     2,
						Explanation: "Двойная обработка (log + return) приводит к дублированию логов. Правило: на каждом уровне либо обработай полностью (лог + fallback), либо оберни контекстом и передай выше.",
					},
					{
						Question:    "errors.New(\"msg\") vs fmt.Errorf(\"msg: %w\", err) — когда что?",
						Options:     []string{"Нет разницы", "errors.New — новая ошибка. fmt.Errorf с %w — обёртка существующей ошибки с сохранением оригинала", "fmt.Errorf медленнее", "errors.New устарел"},
						Correct:     1,
						Explanation: "errors.New создаёт самостоятельную ошибку (sentinel). fmt.Errorf(\"%w\") оборачивает — добавляет контекст и сохраняет цепочку для errors.Is/As.",
					},
				},
				Tasks: []T{
					{
						Title:      "Безопасный калькулятор",
						Difficulty: "medium",
						Description: `<p>Напиши калькулятор, который обрабатывает ошибки:</p>
<p>Ввод: <code>10 / 0</code> → Вывод: <code>Error: division by zero</code></p>
<p>Ввод: <code>10 + 5</code> → Вывод: <code>Result: 15</code></p>
<p>Ввод: <code>10 ^ 5</code> → Вывод: <code>Error: unknown operator: ^</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "error", Definition: "Интерфейс Go для ошибок. Имеет один метод Error() string. nil означает 'нет ошибки'."},
							{Term: "fmt.Errorf(\"...\", args)", Definition: "Создаёт новую ошибку с форматированным сообщением. Как Printf, но возвращает error."},
							{Term: "nil", Definition: "Нулевое значение для указателей, интерфейсов, map, slice, channel. Для error: nil = нет ошибки."},
						},
						TestCases: []TestCase{
							{Input: "10 / 0", ExpectedOutput: "Error: division by zero"},
							{Input: "10 + 5", ExpectedOutput: "Result: 15"},
							{Input: "10 ^ 5", ExpectedOutput: "Error: unknown operator: ^"},
							{Input: "20 - 8", ExpectedOutput: "Result: 12"},
							{Input: "6 * 7", ExpectedOutput: "Result: 42"},
						},
						StarterCode: `package main

import "fmt"

func calc(a int, op string, b int) (int, error) {
    // Реализуй: +, -, *, /
    // Деление на 0 → ошибка
    // Неизвестный оператор → ошибка
    return 0, nil
}

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &op, &b)

    result, err := calc(a, op, b)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("Result: %d\n", result)
    }
}`,
						Hints: `<p>switch op: case "+": return a+b, nil. case "/": if b == 0 { return 0, fmt.Errorf("division by zero") }. default: return 0, fmt.Errorf("unknown operator: %s", op).</p>`,
						Solution: `<pre><code>package main

import "fmt"

func calc(a int, op string, b int) (int, error) {
    switch op {
    case "+":
        return a + b, nil
    case "-":
        return a - b, nil
    case "*":
        return a * b, nil
    case "/":
        if b == 0 {
            return 0, fmt.Errorf("division by zero")
        }
        return a / b, nil
    default:
        return 0, fmt.Errorf("unknown operator: %s", op)
    }
}

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &op, &b)

    result, err := calc(a, op, b)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("Result: %d\n", result)
    }
}</code></pre>`,
					},
					{
						Title:      "Валидация пользователя",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>validateUser(name string, age int) error</code>:</p>
<ul>
<li>Имя пустое → <code>Error: name is required</code></li>
<li>Возраст < 0 или > 150 → <code>Error: invalid age</code></li>
<li>Всё ок → <code>Valid user: Name, age Age</code></li>
</ul>
<p>Ввод: <code>Alice 25</code> → Вывод: <code>Valid user: Alice, age 25</code></p>
<p>Ввод: <code>"" -5</code> → Вывод: <code>Error: name is required</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "errors.New(msg)", Definition: "Создаёт простую ошибку с текстовым сообщением. import \"errors\"."},
							{Term: "if err != nil { return err }", Definition: "Стандартный паттерн: проверить ошибку и вернуть выше. Не игнорируй ошибки!"},
						},
						TestCases: []TestCase{
							{Input: "Alice 25", ExpectedOutput: "Valid user: Alice, age 25"},
							{Input: "Bob -5", ExpectedOutput: "Error: invalid age"},
						},
						StarterCode: `package main

import (
    "fmt"
    "errors"
)

func validateUser(name string, age int) error {
    // Проверь: имя не пустое, возраст от 0 до 150
    _ = errors.New // убери когда используешь
    return nil
}

func main() {
    var name string
    var age int
    fmt.Scan(&name, &age)

    if err := validateUser(name, age); err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("Valid user: %s, age %d\n", name, age)
    }
}`,
						Hints: `<p><code>if name == "" { return errors.New("name is required") }</code></p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "errors"
)

func validateUser(name string, age int) error {
    if name == "" {
        return errors.New("name is required")
    }
    if age < 0 || age > 150 {
        return errors.New("invalid age")
    }
    return nil
}

func main() {
    var name string
    var age int
    fmt.Scan(&name, &age)

    if err := validateUser(name, age); err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("Valid user: %s, age %d\n", name, age)
    }
}</code></pre>`,
					},
					{
						Title:      "Цепочка преобразований",
						Difficulty: "hard",
						Description: `<p>Напиши программу-конвертер, которая читает строку-число и единицу, преобразует:</p>
<ul>
<li><code>km</code> → метры (× 1000)</li>
<li><code>mi</code> → км (× 1.60934)</li>
<li>Неизвестная единица → ошибка</li>
<li>Невалидное число → ошибка</li>
</ul>
<p>Ввод: <code>5 km</code> → Вывод: <code>5000.00 meters</code></p>
<p>Ввод: <code>abc km</code> → Вывод: <code>Error: invalid number: abc</code></p>
<p>Ввод: <code>5 ft</code> → Вывод: <code>Error: unknown unit: ft</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "fmt.Errorf(\"...: %w\", err)", Definition: "%w оборачивает ошибку. Сохраняет оригинал для проверки через errors.Is/As."},
							{Term: "strconv.ParseFloat(s, 64)", Definition: "Парсит строку в float64. Возвращает (float64, error). 64 — битность."},
						},
						TestCases: []TestCase{
							{Input: "5 km", ExpectedOutput: "5000.00 meters"},
							{Input: "10 mi", ExpectedOutput: "16093.40 meters"},
							{Input: "abc km", ExpectedOutput: "Error: invalid number: abc"},
							{Input: "5 ft", ExpectedOutput: "Error: unknown unit: ft"},
						},
						StarterCode: `package main

import (
    "fmt"
    "strconv"
)

func convert(valueStr string, unit string) (float64, error) {
    // 1. Парси число через strconv.ParseFloat
    // 2. Конвертируй по unit
    // 3. Верни ошибку с контекстом при проблемах
    _ = strconv.ParseFloat
    return 0, nil
}

func main() {
    var valueStr, unit string
    fmt.Scan(&valueStr, &unit)

    result, err := convert(valueStr, unit)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("%.2f meters\n", result)
    }
}`,
						Hints: `<p>ParseFloat для парсинга. При ошибке: <code>fmt.Errorf("invalid number: %s", valueStr)</code>. switch unit для конвертации.</p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "strconv"
)

func convert(valueStr string, unit string) (float64, error) {
    val, err := strconv.ParseFloat(valueStr, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid number: %s", valueStr)
    }

    switch unit {
    case "km":
        return val * 1000, nil
    case "mi":
        return val * 1609.34, nil
    default:
        return 0, fmt.Errorf("unknown unit: %s", unit)
    }
}

func main() {
    var valueStr, unit string
    fmt.Scan(&valueStr, &unit)

    result, err := convert(valueStr, unit)
    if err != nil {
        fmt.Printf("Error: %s\n", err)
    } else {
        fmt.Printf("%.2f meters\n", result)
    }
}</code></pre>`,
					},
					{
						Title:      "Чтение конфига с ошибками",
						Difficulty: "easy",
						Description: `<p>Функция <code>parseConfig(line string) (key, value string, err error)</code> парсит строку "key=value". Ошибки:</p>
<ul>
<li>Пустая строка → "empty line"</li>
<li>Нет символа = → "missing separator"</li>
</ul>
<p>Ввод:</p>
<pre><code>3
host=localhost
port
</code></pre>
<p>Вывод:</p>
<pre><code>host = localhost
Error: missing separator
Error: empty line</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "strings.SplitN(s, sep, n)", Definition: "Разбивает строку на n частей по разделителю. SplitN(s, \"=\", 2) — первое = и остаток."},
						},
						TestCases: []TestCase{
							{Input: "3\nhost=localhost\nport\n", ExpectedOutput: "host = localhost\nError: missing separator\nError: empty line"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func parseConfig(line string) (string, string, error) {
    if strings.TrimSpace(line) == "" {
        return "", "", fmt.Errorf("empty line")
    }
    parts := strings.SplitN(line, "=", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("missing separator")
    }
    return parts[0], parts[1], nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        key, val, err := parseConfig(scanner.Text())
        if err != nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            fmt.Printf("%s = %s\n", key, val)
        }
    }
}`,
						Hints: `<p><code>strings.SplitN(line, "=", 2)</code> и проверяй <code>len(parts) != 2</code>.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func parseConfig(line string) (string, string, error) {
    if strings.TrimSpace(line) == "" {
        return "", "", fmt.Errorf("empty line")
    }
    parts := strings.SplitN(line, "=", 2)
    if len(parts) != 2 {
        return "", "", fmt.Errorf("missing separator")
    }
    return parts[0], parts[1], nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        key, val, err := parseConfig(scanner.Text())
        if err != nil {
            fmt.Printf("Error: %s\n", err)
        } else {
            fmt.Printf("%s = %s\n", key, val)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Цепочка с оборачиванием (%w)",
						Difficulty: "medium",
						Description: `<p>Реализуй цепочку вызовов: <code>getUser</code> → <code>findInDB</code>. Каждый уровень оборачивает ошибку с контекстом через <code>%w</code>:</p>
<p>Ввод: <code>42</code> → Вывод: <code>getUser(42): findInDB: user not found</code></p>
<p>Ввод: <code>1</code> → Вывод: <code>Found: Alice</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "%w (wrap)", Definition: "fmt.Errorf(\"context: %w\", err) — оборачивает ошибку. Оригинал доступен через errors.Is/As."},
							{Term: "errors.Is(err, target)", Definition: "Проверяет всю цепочку обёрток. Даже если ошибка обёрнута 5 раз — найдёт оригинал."},
						},
						TestCases: []TestCase{
							{Input: "42", ExpectedOutput: "getUser(42): findInDB: user not found"},
							{Input: "1", ExpectedOutput: "Found: Alice"},
						},
						StarterCode: `package main

import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("user not found")

func findInDB(id int) (string, error) {
    if id == 1 { return "Alice", nil }
    return "", fmt.Errorf("findInDB: %w", ErrNotFound)
}

func getUser(id int) (string, error) {
    name, err := findInDB(id)
    if err != nil {
        return "", fmt.Errorf("getUser(%d): %w", id, err)
    }
    return name, nil
}

func main() {
    var id int
    fmt.Scan(&id)
    name, err := getUser(id)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("Found: %s\n", name)
    }
}`,
						Hints: `<p>Каждый уровень: <code>fmt.Errorf("context: %w", err)</code>. Ошибки вкладываются как матрёшки.</p>`,
						Solution: `<pre><code>package main

import (
    "errors"
    "fmt"
)

var ErrNotFound = errors.New("user not found")

func findInDB(id int) (string, error) {
    if id == 1 { return "Alice", nil }
    return "", fmt.Errorf("findInDB: %w", ErrNotFound)
}

func getUser(id int) (string, error) {
    name, err := findInDB(id)
    if err != nil {
        return "", fmt.Errorf("getUser(%d): %w", id, err)
    }
    return name, nil
}

func main() {
    var id int
    fmt.Scan(&id)
    name, err := getUser(id)
    if err != nil {
        fmt.Println(err)
    } else {
        fmt.Printf("Found: %s\n", name)
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug: "custom-errors", Title: "Свои типы ошибок и panic/recover", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Свои типы ошибок и panic/recover</h1>

<h2>Свой тип ошибки</h2>
<p>Иногда строки недостаточно — нужна ошибка с дополнительными данными:</p>

<pre><code>type NotFoundError struct {
    Entity string
    ID     int
}

// Реализуем интерфейс error
func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with id %d not found", e.Entity, e.ID)
}

func GetUser(id int) (*User, error) {
    // ... поиск в БД
    return nil, &NotFoundError{Entity: "user", ID: id}
}

// Проверка типа ошибки
err := GetUser(42)
var nfe *NotFoundError
if errors.As(err, &nfe) {
    fmt.Printf("Not found: %s #%d\n", nfe.Entity, nfe.ID)
    // Можно вернуть HTTP 404
}</code></pre>

<h2>errors.Is vs errors.As</h2>
<pre><code>import "errors"

// errors.Is — проверка на конкретное значение ошибки
var ErrNotFound = errors.New("not found")
var ErrForbidden = errors.New("forbidden")

func GetVideo(id int) (*Video, error) {
    return nil, fmt.Errorf("GetVideo(%d): %w", id, ErrNotFound)
}

err := GetVideo(42)
if errors.Is(err, ErrNotFound) {
    // true! errors.Is разворачивает цепочку %w
    fmt.Println("Video not found")
}

// errors.As — проверка на тип ошибки (извлечь данные)
var nfe *NotFoundError
if errors.As(err, &nfe) {
    fmt.Println("Entity:", nfe.Entity)
}</code></pre>

<h2>panic и recover</h2>
<p><strong>panic</strong> — аварийная остановка. Используй только для "невозможных" ситуаций:</p>

<pre><code>// panic — программа падает
func mustParse(s string) int {
    n, err := strconv.Atoi(s)
    if err != nil {
        panic("invalid number: " + s)  // крайняя мера!
    }
    return n
}

// recover — перехват panic (только внутри defer)
func safeDiv(a, b int) (result int, err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    return a / b, nil  // b=0 → panic → recover → err
}</code></pre>

<h2>Когда panic, когда error?</h2>
<ul>
<li><strong>error</strong> — 99% случаев. Файл не найден, сеть упала, невалидный ввод</li>
<li><strong>panic</strong> — баг в программе: nil pointer, index out of range, невозможное состояние</li>
<li><strong>Никогда:</strong> panic как замена error для обычных ошибок</li>
</ul>

<h2>Best practices</h2>
<pre><code>// 1. Всегда добавляй контекст
return fmt.Errorf("repository.GetUser(%d): %w", id, err)

// 2. Sentinel errors для известных состояний
var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exists")

// 3. Custom errors для данных
type ValidationError struct {
    Field   string
    Message string
}

// 4. Обрабатывай ошибку один раз
// ПЛОХО:
log.Println("error:", err)
return err  // логируем И возвращаем — двойная обработка

// ХОРОШО:
return fmt.Errorf("doing X: %w", err)  // только возвращаем с контекстом</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Чем errors.Is отличается от errors.As?",
						Options:     []string{"Ничем", "Is проверяет на конкретное значение ошибки, As — на тип ошибки и извлекает данные", "Is быстрее", "As устарел"},
						Correct:     1,
						Explanation: "errors.Is(err, ErrNotFound) — 'это ошибка not found?'. errors.As(err, &nfe) — 'это ошибка типа *NotFoundError? Если да, дай мне данные.'",
					},
					{
						Question:    "Когда уместно использовать panic?",
						Options:     []string{"При любой ошибке", "Только при багах — невозможных состояниях, которые означают ошибку в коде", "Вместо error", "Для валидации ввода"},
						Correct:     1,
						Explanation: "panic = 'в программе баг'. Пример: nil map, нарушенный инвариант. Для штатных ошибок (файл не найден, сеть упала) — всегда error.",
					},
					{
						Question:    "Что делает recover() и где его можно использовать?",
						Options:     []string{"Перезапускает программу", "Перехватывает panic, но только внутри defer-функции", "Останавливает горутину", "Ловит error"},
						Correct:     1,
						Explanation: "recover() перехватывает panic и возвращает значение, переданное в panic(). Работает ТОЛЬКО внутри defer. Без defer — возвращает nil и ничего не делает.",
					},
					{
						Question:    "Что такое sentinel error?",
						Options:     []string{"Ошибка в часовом поясе", "Глобальная переменная-ошибка для сравнения: var ErrNotFound = errors.New(\"not found\")", "Ошибка из пакета sentinel", "Ошибка в цикле"},
						Correct:     1,
						Explanation: "Sentinel (часовой) — известная ошибка для сравнения. io.EOF, sql.ErrNoRows, os.ErrNotExist. Проверяются через errors.Is(err, ErrNotFound). Пишутся с Err-префиксом.",
					},
					{
						Question:    "Почему свой тип ошибки реализует Error() на pointer receiver (*T)?",
						Options:     []string{"Это обязательно", "Чтобы errors.As мог заполнить target-переменную через указатель, и чтобы два разных экземпляра не считались 'одной ошибкой'", "Pointer быстрее", "Это стиль Google"},
						Correct:     1,
						Explanation: "errors.As(err, &target) проверяет *T. Если Error() на T — errors.As с **T ломается. Также *T гарантирует что каждый &MyError{} уникален для errors.Is.",
					},
				},
				Tasks: []T{
					{
						Title:      "HTTP-ошибки с типами",
						Difficulty: "hard",
						Description: `<p>Создай систему ошибок для HTTP API:</p>
<ul>
<li>Тип <code>APIError</code> с полями Status (int) и Message (string)</li>
<li>Функция <code>handleRequest(path string) error</code> возвращает разные ошибки</li>
<li>Вызывающий код через errors.As определяет HTTP-код</li>
</ul>
<p>Ввод: <code>/users/999</code> → Вывод: <code>404: user not found</code></p>
<p>Ввод: <code>/admin</code> → Вывод: <code>403: forbidden</code></p>
<p>Ввод: <code>/</code> → Вывод: <code>200: ok</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "errors.As(err, &target)", Definition: "Проверяет: является ли err (или обёрнутая ошибка) типом target. Если да — заполняет target данными."},
							{Term: "func (e *T) Error() string", Definition: "Реализация интерфейса error для своего типа. Позволяет использовать тип как ошибку."},
						},
						TestCases: []TestCase{
							{Input: "/users/999", ExpectedOutput: "404: user not found"},
							{Input: "/admin", ExpectedOutput: "403: forbidden"},
							{Input: "/", ExpectedOutput: "200: ok"},
						},
						StarterCode: `package main

import (
    "errors"
    "fmt"
)

type APIError struct {
    Status  int
    Message string
}

func (e *APIError) Error() string {
    return e.Message
}

func handleRequest(path string) error {
    // /users/999 → APIError{404, "user not found"}
    // /admin → APIError{403, "forbidden"}
    // / → nil (нет ошибки)
    return nil
}

func main() {
    var path string
    fmt.Scan(&path)

    err := handleRequest(path)
    if err != nil {
        var apiErr *APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("%d: %s\n", apiErr.Status, apiErr.Message)
        } else {
            fmt.Printf("500: %s\n", err)
        }
    } else {
        fmt.Println("200: ok")
    }
}`,
						Hints: `<p>switch path: case "/users/999": return &APIError{404, "user not found"}. case "/admin": return &APIError{403, "forbidden"}.</p>`,
						Solution: `<pre><code>package main

import (
    "errors"
    "fmt"
)

type APIError struct {
    Status  int
    Message string
}

func (e *APIError) Error() string { return e.Message }

func handleRequest(path string) error {
    switch path {
    case "/users/999":
        return &APIError{404, "user not found"}
    case "/admin":
        return &APIError{403, "forbidden"}
    default:
        return nil
    }
}

func main() {
    var path string
    fmt.Scan(&path)

    err := handleRequest(path)
    if err != nil {
        var apiErr *APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("%d: %s\n", apiErr.Status, apiErr.Message)
        } else {
            fmt.Printf("500: %s\n", err)
        }
    } else {
        fmt.Println("200: ok")
    }
}</code></pre>`,
					},
					{
						Title:      "Sentinel errors и errors.Is",
						Difficulty: "easy",
						Description: `<p>Создай sentinel errors и проверяй через errors.Is:</p>
<p>Ввод:</p>
<pre><code>3
hello

aaaaabbbbbcccccdddddeeeee12345</code></pre>
<p>Вывод:</p>
<pre><code>OK: hello
Error: empty input
Error: too long (max 20)</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "var ErrX = errors.New(...)", Definition: "Sentinel error — глобальная переменная для сравнения через errors.Is."},
						},
						TestCases: []TestCase{
							{Input: "3\nhello\n \naaaaabbbbbcccccdddddeeeee12345", ExpectedOutput: "OK: hello\nError: empty input\nError: too long (max 20)"},
						},
						StarterCode: `package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"
)

var (
    ErrEmpty   = errors.New("empty input")
    ErrTooLong = errors.New("too long (max 20)")
)

func validate(s string) error {
    if strings.TrimSpace(s) == "" { return ErrEmpty }
    if len(s) > 20 { return ErrTooLong }
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        err := validate(scanner.Text())
        if errors.Is(err, ErrEmpty) {
            fmt.Println("Error: empty input")
        } else if errors.Is(err, ErrTooLong) {
            fmt.Println("Error: too long (max 20)")
        } else {
            fmt.Printf("OK: %s\n", scanner.Text())
        }
    }
}`,
						Hints: `<p>errors.Is проверяет всю цепочку. Найдёт sentinel даже если обёрнут через %w.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"
)

var (
    ErrEmpty   = errors.New("empty input")
    ErrTooLong = errors.New("too long (max 20)")
)

func validate(s string) error {
    if strings.TrimSpace(s) == "" { return ErrEmpty }
    if len(s) > 20 { return ErrTooLong }
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        err := validate(scanner.Text())
        if errors.Is(err, ErrEmpty) {
            fmt.Println("Error: empty input")
        } else if errors.Is(err, ErrTooLong) {
            fmt.Println("Error: too long (max 20)")
        } else {
            fmt.Printf("OK: %s\n", scanner.Text())
        }
    }
}</code></pre>`,
					},
					{
						Title:      "errors.As — извлечение данных",
						Difficulty: "medium",
						Description: `<p>Создай <code>FieldError{Field, Reason}</code>. Проверяй через errors.As и выводи детали:</p>
<p>Ввод:</p>
<pre><code>3
email missing@
age -5
name Alice</code></pre>
<p>Вывод:</p>
<pre><code>Field 'email' invalid: bad format
Field 'age' invalid: must be positive
OK: name=Alice</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "errors.As(err, &target)", Definition: "Проверяет тип и заполняет target данными ошибки."},
						},
						TestCases: []TestCase{
							{Input: "3\nemail missing@\nage -5\nname Alice", ExpectedOutput: "Field 'email' invalid: bad format\nField 'age' invalid: must be positive\nOK: name=Alice"},
						},
						StarterCode: `package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"
)

type FieldError struct {
    Field  string
    Reason string
}
func (e *FieldError) Error() string { return e.Field + ": " + e.Reason }

func validateField(field, value string) error {
    switch field {
    case "email":
        if strings.HasPrefix(value, "missing") { return &FieldError{field, "bad format"} }
    case "age":
        if strings.HasPrefix(value, "-") { return &FieldError{field, "must be positive"} }
    }
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        field, value := parts[0], parts[1]
        err := validateField(field, value)
        var fe *FieldError
        if errors.As(err, &fe) {
            fmt.Printf("Field '%s' invalid: %s\n", fe.Field, fe.Reason)
        } else {
            fmt.Printf("OK: %s=%s\n", field, value)
        }
    }
}`,
						Hints: `<p><code>var fe *FieldError; if errors.As(err, &fe) { fe.Field, fe.Reason }</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "errors"
    "fmt"
    "os"
    "strings"
)

type FieldError struct{ Field, Reason string }
func (e *FieldError) Error() string { return e.Field + ": " + e.Reason }

func validateField(field, value string) error {
    switch field {
    case "email":
        if strings.HasPrefix(value, "missing") { return &FieldError{field, "bad format"} }
    case "age":
        if strings.HasPrefix(value, "-") { return &FieldError{field, "must be positive"} }
    }
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.SplitN(scanner.Text(), " ", 2)
        field, value := parts[0], parts[1]
        err := validateField(field, value)
        var fe *FieldError
        if errors.As(err, &fe) {
            fmt.Printf("Field '%s' invalid: %s\n", fe.Field, fe.Reason)
        } else {
            fmt.Printf("OK: %s=%s\n", field, value)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "panic/recover — безопасный вызов",
						Difficulty: "medium",
						Description: `<p>Напиши <code>safeCall(fn func()) error</code> — перехватывает panic через recover:</p>
<p>Ввод:</p>
<pre><code>3
ok
slice
divide</code></pre>
<p>Вывод:</p>
<pre><code>ok: success
slice: recovered: runtime error: index out of range [10] with length 3
divide: recovered: runtime error: integer divide by zero</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "defer+recover", Definition: "Единственный способ перехватить panic. Named return позволяет изменить err из defer."},
						},
						TestCases: []TestCase{
							{Input: "3\nok\nslice\ndivide", ExpectedOutput: "ok: success\nslice: recovered: runtime error: index out of range [10] with length 3\ndivide: recovered: runtime error: integer divide by zero"},
						},
						StarterCode: `package main

import "fmt"

func safeCall(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("recovered: %v", r)
        }
    }()
    fn()
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        var fn func()
        switch cmd {
        case "ok":     fn = func() {}
        case "slice":  fn = func() { s := []int{1, 2, 3}; _ = s[10] }
        case "divide": fn = func() { a, b := 10, 0; _ = a / b }
        }
        err := safeCall(fn)
        if err != nil {
            fmt.Printf("%s: %s\n", cmd, err)
        } else {
            fmt.Printf("%s: success\n", cmd)
        }
    }
}`,
						Hints: `<p>Named return (err error) + defer { recover() → err = fmt.Errorf(...) }</p>`,
						Solution: `<pre><code>package main

import "fmt"

func safeCall(fn func()) (err error) {
    defer func() {
        if r := recover(); r != nil { err = fmt.Errorf("recovered: %v", r) }
    }()
    fn()
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        var fn func()
        switch cmd {
        case "ok":     fn = func() {}
        case "slice":  fn = func() { s := []int{1,2,3}; _ = s[10] }
        case "divide": fn = func() { a, b := 10, 0; _ = a / b }
        }
        if err := safeCall(fn); err != nil {
            fmt.Printf("%s: %s\n", cmd, err)
        } else {
            fmt.Printf("%s: success\n", cmd)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Множественная валидация (MultiError)",
						Difficulty: "hard",
						Description: `<p>Реализуй <code>MultiError</code> — собирает ВСЕ ошибки валидации за один проход:</p>
<p>Ввод: <code>  200 short</code> (пустое имя, возраст>150, пароль&lt;8)</p>
<p>Вывод:</p>
<pre><code>Validation failed (3 errors):
- name: required
- age: must be 0-150
- password: min 8 chars</code></pre>
<p>Ввод: <code>Alice 25 securepass123</code> → <code>Valid!</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "MultiError", Definition: "Собирает все проблемы вместо остановки на первой. Популярен в API-валидации."},
						},
						TestCases: []TestCase{
							{Input: "  200 short", ExpectedOutput: "Validation failed (3 errors):\n- name: required\n- age: must be 0-150\n- password: min 8 chars"},
							{Input: "Alice 25 securepass123", ExpectedOutput: "Valid!"},
						},
						StarterCode: `package main

import "fmt"

type MultiError struct {
    Errors []string
}
func (m *MultiError) Error() string { return fmt.Sprintf("%d errors", len(m.Errors)) }
func (m *MultiError) Add(msg string) { m.Errors = append(m.Errors, msg) }
func (m *MultiError) HasErrors() bool { return len(m.Errors) > 0 }

func validateUser(name string, age int, password string) error {
    me := &MultiError{}
    if name == "" { me.Add("name: required") }
    if age < 0 || age > 150 { me.Add("age: must be 0-150") }
    if len(password) < 8 { me.Add("password: min 8 chars") }
    if me.HasErrors() { return me }
    return nil
}

func main() {
    var name string
    var age int
    var password string
    fmt.Scan(&name, &age, &password)

    err := validateUser(name, age, password)
    if err != nil {
        me := err.(*MultiError)
        fmt.Printf("Validation failed (%d errors):\n", len(me.Errors))
        for _, e := range me.Errors {
            fmt.Printf("- %s\n", e)
        }
    } else {
        fmt.Println("Valid!")
    }
}`,
						Hints: `<p>Не возвращай рано. Проверяй ВСЕ условия, потом HasErrors().</p>`,
						Solution: `<pre><code>package main

import "fmt"

type MultiError struct{ Errors []string }
func (m *MultiError) Error() string { return fmt.Sprintf("%d errors", len(m.Errors)) }
func (m *MultiError) Add(msg string) { m.Errors = append(m.Errors, msg) }
func (m *MultiError) HasErrors() bool { return len(m.Errors) > 0 }

func validateUser(name string, age int, password string) error {
    me := &MultiError{}
    if name == "" { me.Add("name: required") }
    if age < 0 || age > 150 { me.Add("age: must be 0-150") }
    if len(password) < 8 { me.Add("password: min 8 chars") }
    if me.HasErrors() { return me }
    return nil
}

func main() {
    var name string; var age int; var password string
    fmt.Scan(&name, &age, &password)
    err := validateUser(name, age, password)
    if err != nil {
        me := err.(*MultiError)
        fmt.Printf("Validation failed (%d errors):\n", len(me.Errors))
        for _, e := range me.Errors { fmt.Printf("- %s\n", e) }
    } else {
        fmt.Println("Valid!")
    }
}</code></pre>`,
					},
				},
			},
		},
	}
}
