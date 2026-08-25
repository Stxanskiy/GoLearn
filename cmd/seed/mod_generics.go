package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Generics (Go 1.18+)
// После интерфейсов, перед ошибками
// ════════════════════════════════════════════════════════════════

func mod_generics() M {
	return M{
		Slug:          "generics",
		Title:         "Generics — обобщённое программирование",
		Description:   "Type parameters, constraints, когда использовать generics а когда интерфейсы. Go 1.18+.",
		Order:         7, // после интерфейсов
		Track:         "shared",
		Difficulty:    "intermediate",
		Prerequisites: []string{"interfaces"},
		Lessons: []L{
			{
				Slug: "generics-basics", Title: "Зачем нужны generics", Order: 1,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Generics — обобщённое программирование</h1>

<h2>Проблема без generics</h2>
<p>Представь: ты написал функцию поиска максимума для int. Теперь нужен максимум для float64. И для string. Без generics приходилось писать <strong>три разные функции</strong>:</p>

<pre><code>func MaxInt(a, b int) int { if a > b { return a }; return b }
func MaxFloat(a, b float64) float64 { if a > b { return a }; return b }
func MaxString(a, b string) string { if a > b { return a }; return b }
// Один и тот же код, только тип другой!</code></pre>

<h2>Решение: type parameters</h2>
<p>С Go 1.18 можно написать <strong>одну функцию</strong> для любого типа:</p>

<pre><code>// T — параметр типа. comparable означает "можно сравнивать через < > == !="
func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

// Использование — Go выводит тип автоматически
fmt.Println(Max(3, 7))         // 7 (int)
fmt.Println(Max(3.14, 2.71))   // 3.14 (float64)
fmt.Println(Max("abc", "xyz")) // "xyz" (string)</code></pre>

<p><code>[T cmp.Ordered]</code> — <strong>параметр типа с ограничением</strong> (constraint). Говорит: "T может быть любым типом, для которого работают операторы сравнения".</p>

<h2>Синтаксис</h2>
<pre><code>// Обобщённая функция
func Filter[T any](slice []T, pred func(T) bool) []T {
    result := []T{}
    for _, v := range slice {
        if pred(v) {
            result = append(result, v)
        }
    }
    return result
}

// Обобщённая структура
type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(item T) {
    s.items = append(s.items, item)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    item := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return item, true
}</code></pre>

<h2>Constraints — ограничения типов</h2>
<pre><code>import "cmp"
import "golang.org/x/exp/constraints"

// Встроенные:
any            // любой тип (= interface{})
comparable     // типы, которые можно сравнивать через ==
cmp.Ordered    // числа + строки (поддерживают < > <= >=)

// Свой constraint через интерфейс
type Number interface {
    int | int8 | int16 | int32 | int64 |
    float32 | float64
}

func Sum[T Number](nums []T) T {
    var total T
    for _, n := range nums {
        total += n
    }
    return total
}</code></pre>

<h2>Когда использовать generics?</h2>
<ul>
<li><strong>Да:</strong> контейнеры (Stack, Queue, Set), утилиты (Map, Filter, Reduce), алгоритмы (Sort, Search)</li>
<li><strong>Нет:</strong> когда достаточно интерфейса, когда только один тип, когда код становится сложнее чем дублирование</li>
</ul>
<p><strong>Правило:</strong> если ты дублируешь одну и ту же логику для разных типов — это кандидат для generics.</p>`,

				Quiz: []Q{
					{
						Question:    "Что такое T в func Max[T cmp.Ordered](a, b T) T?",
						Options:     []string{"Конкретный тип", "Параметр типа — заменится на конкретный тип при вызове", "Название переменной", "Тип ошибки"},
						Correct:     1,
						Explanation: "T — параметр типа (type parameter). При вызове Max(3, 7) Go подставит T=int. При Max(\"a\", \"b\") — T=string. Одна функция, много типов.",
					},
					{
						Question:    "Что означает constraint cmp.Ordered?",
						Options:     []string{"Тип должен быть отсортирован", "Тип поддерживает операторы сравнения: <, >, <=, >=", "Тип должен быть числом", "Тип является слайсом"},
						Correct:     1,
						Explanation: "cmp.Ordered включает все числовые типы + string — те, для которых определены операторы порядка (<, >, <=, >=).",
					},
					{
						Question:    "Когда НЕ стоит использовать generics?",
						Options:     []string{"Когда достаточно интерфейса или когда только один конкретный тип", "Никогда — всегда используй generics", "Только в тестах", "Только для строк"},
						Correct:     0,
						Explanation: "Generics добавляют сложность. Если интерфейс решает задачу, или код не дублируется — generics не нужны. Простота > универсальность.",
					},
					{
						Question:    "Чем comparable отличается от any в constraints?",
						Options:     []string{"Ничем", "comparable разрешает == и != (нужно для ключей map), any — вообще любой тип без ограничений", "any быстрее", "comparable только для чисел"},
						Correct:     1,
						Explanation: "any (interface{}) — ни одного ограничения. comparable — можно сравнивать через == и !=. Для map[T]V ключ T должен быть comparable. Для Contains(slice, target) нужен == → comparable.",
					},
					{
						Question:    "Как Go выводит тип T при вызове Contains([]int{1,2,3}, 5)?",
						Options:     []string{"Нужно указать явно: Contains[int](...)", "Go автоматически выводит T=int из типа аргументов (type inference)", "Через reflect", "Через interface{}"},
						Correct:     1,
						Explanation: "Type inference: Go смотрит на тип аргументов и выводит параметры типа. Явное указание Contains[int](...) нужно крайне редко — только при неоднозначности.",
					},
				},
				Tasks: []T{
					{
						Title:      "Contains для любого типа",
						Difficulty: "easy",
						Description: `<p>Напиши обобщённую функцию <code>Contains[T comparable](slice []T, target T) bool</code>, которая проверяет есть ли элемент в слайсе.</p>
<p>Ввод:</p>
<pre><code>5
1 2 3 4 5
3</code></pre>
<p>Вывод: <code>true</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "[T comparable]", Definition: "Параметр типа T с ограничением comparable — тип можно сравнивать через ==. Подходят: числа, строки, bool."},
							{Term: "func Name[T constraint](...)", Definition: "Обобщённая функция. T — параметр типа, constraint — какие типы допустимы."},
						},
						TestCases: []TestCase{
							{Input: "5\n1 2 3 4 5\n3", ExpectedOutput: "true"},
							{Input: "5\n1 2 3 4 5\n7", ExpectedOutput: "false"},
						},
						StarterCode: `package main

import "fmt"

func Contains[T comparable](slice []T, target T) bool {
    // Проверь есть ли target в slice
    return false
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }
    var target int
    fmt.Scan(&target)

    fmt.Println(Contains(nums, target))
}`,
						Hints: `<p>Перебери слайс: <code>for _, v := range slice { if v == target { return true } }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func Contains[T comparable](slice []T, target T) bool {
    for _, v := range slice {
        if v == target {
            return true
        }
    }
    return false
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }
    var target int
    fmt.Scan(&target)

    fmt.Println(Contains(nums, target))
}</code></pre>`,
					},
					{
						Title:      "Map и Filter — обобщённые",
						Difficulty: "hard",
						Description: `<p>Напиши две обобщённые функции:</p>
<ul>
<li><code>Map[T, U any](slice []T, fn func(T) U) []U</code> — преобразовать каждый элемент</li>
<li><code>Filter[T any](slice []T, pred func(T) bool) []T</code> — оставить только подходящие</li>
</ul>
<p>Ввод:</p>
<pre><code>5
1 2 3 4 5</code></pre>
<p>Вывод (удвоенные чётные):</p>
<pre><code>4 8</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "[T, U any]", Definition: "Два параметра типа. T — входной тип, U — выходной. any = любой тип без ограничений."},
							{Term: "func(T) U", Definition: "Тип функции: принимает T, возвращает U. Используется как аргумент для Map."},
						},
						TestCases: []TestCase{
							{Input: "5\n1 2 3 4 5", ExpectedOutput: "4 8"},
							{Input: "4\n10 15 20 25", ExpectedOutput: "20 40"},
						},
						StarterCode: `package main

import "fmt"

func Map[T, U any](slice []T, fn func(T) U) []U {
    // Преобразуй каждый элемент через fn
    return nil
}

func Filter[T any](slice []T, pred func(T) bool) []T {
    // Оставь только элементы где pred(v) == true
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    // Сначала отфильтруй чётные, потом удвой
    even := Filter(nums, func(n int) bool { return n%2 == 0 })
    doubled := Map(even, func(n int) int { return n * 2 })

    for i, v := range doubled {
        if i > 0 { fmt.Print(" ") }
        fmt.Print(v)
    }
    fmt.Println()
}`,
						Hints: `<p>Map: создай <code>result := make([]U, len(slice))</code>, в цикле <code>result[i] = fn(v)</code>. Filter: <code>append</code> при <code>pred(v)</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func Map[T, U any](slice []T, fn func(T) U) []U {
    result := make([]U, len(slice))
    for i, v := range slice {
        result[i] = fn(v)
    }
    return result
}

func Filter[T any](slice []T, pred func(T) bool) []T {
    result := []T{}
    for _, v := range slice {
        if pred(v) {
            result = append(result, v)
        }
    }
    return result
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    even := Filter(nums, func(n int) bool { return n%2 == 0 })
    doubled := Map(even, func(n int) int { return n * 2 })

    for i, v := range doubled {
        if i > 0 { fmt.Print(" ") }
        fmt.Print(v)
    }
    fmt.Println()
}</code></pre>`,
					},
					{
						Title:      "Max для любого Ordered",
						Difficulty: "easy",
						Description: `<p>Напиши <code>Max[T cmp.Ordered](a, b T) T</code> и используй для чисел и строк:</p>
<p>Ввод:</p>
<pre><code>3 7
hello world</code></pre>
<p>Вывод:</p>
<pre><code>Max int: 7
Max string: world</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "cmp.Ordered", Definition: "Constraint для типов с операторами <, >, <=, >=. Включает все числа и string."},
						},
						TestCases: []TestCase{
							{Input: "3 7\nhello world", ExpectedOutput: "Max int: 7\nMax string: world"},
							{Input: "10 5\nalpha beta", ExpectedOutput: "Max int: 10\nMax string: beta"},
						},
						StarterCode: `package main

import (
    "bufio"
    "cmp"
    "fmt"
    "os"
)

func Max[T cmp.Ordered](a, b T) T {
    if a > b {
        return a
    }
    return b
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Printf("Max int: %d\n", Max(a, b))

    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    var s1, s2 string
    fmt.Sscan(scanner.Text(), &s1, &s2)
    fmt.Printf("Max string: %s\n", Max(s1, s2))
}`,
						Hints: `<p><code>if a > b { return a }; return b</code> — одна функция для int и string.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "cmp"
    "fmt"
    "os"
)

func Max[T cmp.Ordered](a, b T) T {
    if a > b { return a }
    return b
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Printf("Max int: %d\n", Max(a, b))

    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    var s1, s2 string
    fmt.Sscan(scanner.Text(), &s1, &s2)
    fmt.Printf("Max string: %s\n", Max(s1, s2))
}</code></pre>`,
					},
					{
						Title:      "Generic Stack",
						Difficulty: "medium",
						Description: `<p>Реализуй обобщённый стек <code>Stack[T any]</code> с методами Push, Pop, Len:</p>
<p>Ввод:</p>
<pre><code>6
push 10
push 20
push 30
pop
pop
len</code></pre>
<p>Вывод:</p>
<pre><code>popped: 30
popped: 20
len: 1</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "type Stack[T any] struct", Definition: "Обобщённая структура. T — параметр типа. Stack[int], Stack[string] — конкретные инстанциации."},
							{Term: "var zero T", Definition: "Нулевое значение для параметра типа. Возвращается при ошибке (Pop из пустого стека)."},
						},
						TestCases: []TestCase{
							{Input: "6\npush 10\npush 20\npush 30\npop\npop\nlen", ExpectedOutput: "popped: 30\npopped: 20\nlen: 1"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Stack[T any] struct {
    items []T
}

func (s *Stack[T]) Push(v T) {
    s.items = append(s.items, v)
}

func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 {
        var zero T
        return zero, false
    }
    v := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return v, true
}

func (s *Stack[T]) Len() int {
    return len(s.items)
}

func main() {
    var n int
    fmt.Scan(&n)
    stack := &Stack[int]{}
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        switch parts[0] {
        case "push":
            var v int
            fmt.Sscan(parts[1], &v)
            stack.Push(v)
        case "pop":
            if v, ok := stack.Pop(); ok {
                fmt.Printf("popped: %d\n", v)
            } else {
                fmt.Println("empty")
            }
        case "len":
            fmt.Printf("len: %d\n", stack.Len())
        }
    }
}`,
						Hints: `<p>Pop: сохрани последний элемент, обрежь слайс на 1. var zero T — для возврата при пустом стеке.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type Stack[T any] struct{ items []T }
func (s *Stack[T]) Push(v T)       { s.items = append(s.items, v) }
func (s *Stack[T]) Pop() (T, bool) {
    if len(s.items) == 0 { var zero T; return zero, false }
    v := s.items[len(s.items)-1]
    s.items = s.items[:len(s.items)-1]
    return v, true
}
func (s *Stack[T]) Len() int { return len(s.items) }

func main() {
    var n int
    fmt.Scan(&n)
    stack := &Stack[int]{}
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ {
        scanner.Scan()
        parts := strings.Fields(scanner.Text())
        switch parts[0] {
        case "push":
            var v int; fmt.Sscan(parts[1], &v); stack.Push(v)
        case "pop":
            if v, ok := stack.Pop(); ok { fmt.Printf("popped: %d\n", v) } else { fmt.Println("empty") }
        case "len":
            fmt.Printf("len: %d\n", stack.Len())
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Index — поиск позиции",
						Difficulty: "easy",
						Description: `<p>Напиши <code>Index[T comparable](slice []T, target T) int</code> — возвращает индекс первого вхождения или -1:</p>
<p>Ввод:</p>
<pre><code>5
apple banana cherry date elderberry
cherry</code></pre>
<p>Вывод: <code>2</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "Index[T comparable]", Definition: "Как strings.Index, но для любого comparable типа. Возвращает -1 если не найден."},
						},
						TestCases: []TestCase{
							{Input: "5\napple banana cherry date elderberry\ncherry", ExpectedOutput: "2"},
							{Input: "3\nfoo bar baz\nqux", ExpectedOutput: "-1"},
						},
						StarterCode: `package main

import "fmt"

func Index[T comparable](slice []T, target T) int {
    for i, v := range slice {
        if v == target {
            return i
        }
    }
    return -1
}

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words { fmt.Scan(&words[i]) }
    var target string
    fmt.Scan(&target)
    fmt.Println(Index(words, target))
}`,
						Hints: `<p>Перебери с индексом: <code>for i, v := range slice { if v == target { return i } }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func Index[T comparable](slice []T, target T) int {
    for i, v := range slice {
        if v == target { return i }
    }
    return -1
}

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words { fmt.Scan(&words[i]) }
    var target string
    fmt.Scan(&target)
    fmt.Println(Index(words, target))
}</code></pre>`,
					},
				},
			},
			{
				Slug: "generics-advanced", Title: "Продвинутые generics: constraints и паттерны", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Продвинутые generics</h1>

<h2>Под капотом: как работают generics</h2>
<p>Go использует подход <strong>GC Shape Stenciling</strong> — компилятор создаёт одну реализацию для каждой "формы" типа:</p>
<ul>
<li>Все указатели → одна реализация (одинаковый размер)</li>
<li>Каждый value-тип (int, string, struct) → своя реализация</li>
</ul>
<p>Это компромисс между C++ (полная мономорфизация — быстро, но раздувает бинарник) и Java (стирание типов — компактно, но медленнее из-за boxing).</p>

<h2>Оператор ~ (underlying type)</h2>
<p>Оператор <code>~</code> означает "этот тип И все типы, основанные на нём":</p>
<pre><code>type Celsius float64
type Fahrenheit float64

// Без ~ : Celsius и Fahrenheit НЕ подходят (они не float64)
type StrictFloat interface { float64 }

// С ~ : подходят ВСЕ типы, основанные на float64
type AnyFloat interface { ~float64 }

func Double[T AnyFloat](v T) T {
    return v * 2
}

var temp Celsius = 36.6
fmt.Println(Double(temp)) // 73.2 — работает благодаря ~</code></pre>

<h2>Union types — объединение типов</h2>
<pre><code>// Constraint = объединение конкретных типов
type Integer interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64
}

type Numeric interface {
    Integer | ~float32 | ~float64
}

func Abs[T Numeric](v T) T {
    if v < 0 {
        return -v
    }
    return v
}

fmt.Println(Abs(-42))    // 42
fmt.Println(Abs(-3.14))  // 3.14</code></pre>

<h2>Constraint с методами</h2>
<pre><code>// Constraint может требовать и типы, и методы
type Stringer interface {
    ~int | ~string
    String() string  // должен иметь метод String()
}

// Практичнее — только метод:
type Formatter interface {
    Format() string
}

func PrintAll[T Formatter](items []T) {
    for _, item := range items {
        fmt.Println(item.Format())
    }
}</code></pre>

<h2>Обобщённые структуры данных</h2>
<pre><code>// Set — множество уникальных элементов
type Set[T comparable] struct {
    m map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
    return &Set[T]{m: make(map[T]struct{})}
}

func (s *Set[T]) Add(v T)         { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool    { _, ok := s.m[v]; return ok }
func (s *Set[T]) Remove(v T)      { delete(s.m, v) }
func (s *Set[T]) Len() int        { return len(s.m) }

// Использование:
names := NewSet[string]()
names.Add("Alice")
names.Add("Bob")
names.Add("Alice") // дубликат — не добавится
fmt.Println(names.Len()) // 2</code></pre>

<h2>Стандартная библиотека generics (Go 1.21+)</h2>
<pre><code>import (
    "slices"
    "maps"
    "cmp"
)

// slices — обобщённые операции над слайсами
slices.Contains([]int{1,2,3}, 2)    // true
slices.Sort([]int{3,1,2})           // [1,2,3]
slices.Index([]string{"a","b"}, "b") // 1
slices.Compact([]int{1,1,2,2,3})    // [1,2,3]

// maps — обобщённые операции над map
keys := maps.Keys(myMap)      // []K
vals := maps.Values(myMap)    // []V
maps.Equal(m1, m2)            // глубокое сравнение

// cmp — сравнение
cmp.Or(a, b, c)               // первое ненулевое значение
cmp.Compare(a, b)              // -1, 0, 1</code></pre>

<h2>Когда generics vs интерфейсы?</h2>
<table>
<tr><th>Generics</th><th>Интерфейсы</th></tr>
<tr><td>Один алгоритм, разные типы</td><td>Разное поведение, одна сигнатура</td></tr>
<tr><td>Тип известен при компиляции</td><td>Тип определяется в runtime</td></tr>
<tr><td>Нет boxing, zero allocation</td><td>Может быть boxing (interface{})</td></tr>
<tr><td>Контейнеры: Set, Stack, Queue</td><td>Подменяемость: Repository, Logger</td></tr>
</table>`,

				Quiz: []Q{
					{
						Question:    "Что означает ~int в constraint?",
						Options:     []string{"Только тип int", "Тип int И все типы, основанные на int (type MyInt int)", "Приблизительно int", "Указатель на int"},
						Correct:     1,
						Explanation: "~ означает underlying type. ~int включает int, type UserID int, type Age int и любой другой тип с underlying type int.",
					},
					{
						Question:    "Какой пакет стандартной библиотеки содержит обобщённые операции над слайсами?",
						Options:     []string{"generic", "slices — Contains, Sort, Index, Compact и другие", "reflect", "collections"},
						Correct:     1,
						Explanation: "Пакет slices (Go 1.21+) заменил кучу ручных циклов. slices.Contains, slices.Sort, slices.Index — обобщённые и безопасные.",
					},
					{
						Question:    "Когда лучше интерфейс, а не generics?",
						Options:     []string{"Всегда", "Когда нужно разное поведение при одной сигнатуре (dependency injection, подмена в тестах)", "Никогда", "Только для строк"},
						Correct:     1,
						Explanation: "Интерфейсы — для полиморфизма поведения (Repository может быть PostgresRepo или MockRepo). Generics — для полиморфизма типов (одна логика, разные данные).",
					},
					{
						Question:    "Что делает slices.Contains([]int{1,2,3}, 2)?",
						Options:     []string{"Удаляет 2", "Возвращает true — проверяет наличие элемента в слайсе (generic функция из stdlib)", "Добавляет 2", "Сортирует"},
						Correct:     1,
						Explanation: "Пакет slices (Go 1.21+) содержит обобщённые утилиты: Contains, Sort, Index, Compact, Reverse. Не нужно писать свои циклы — стандартная библиотека покрывает 90% случаев.",
					},
					{
						Question:    "Зачем struct{} в map[T]struct{} для Set?",
						Options:     []string{"Для красоты", "struct{} занимает 0 байт — экономим память по сравнению с map[T]bool", "Это обязательно", "Быстрее чем bool"},
						Correct:     1,
						Explanation: "map[T]bool тратит 1 байт на каждое значение. map[T]struct{} — 0 байт. При миллионах элементов разница существенна. s.m[v] = struct{}{} — идиома Go для множеств.",
					},
				},
				Tasks: []T{
					{
						Title:      "Обобщённый Set",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "map[T]struct{}", Definition: "Идиома для множества в Go. struct{} занимает 0 байт. Ключи map — элементы множества."},
							{Term: "[T comparable]", Definition: "Constraint comparable нужен для ключей map — они должны поддерживать ==."},
						},
						Description: `<p>Реализуй обобщённое множество с операциями Add, Has, Len.</p>
<p>Ввод:</p>
<pre><code>add 5
add 3
add 5
has 5
has 7
len</code></pre>
<p>Вывод:</p>
<pre><code>true
false
2</code></pre>`,
						Hints: `<p>Используй <code>map[T]struct{}</code>. Add: <code>s.m[v] = struct{}{}</code>. Has: <code>_, ok := s.m[v]</code>.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{m: make(map[T]struct{})}
}

func (s *Set[T]) Add(v T) { s.m[v] = struct{}{} }
func (s *Set[T]) Has(v T) bool { _, ok := s.m[v]; return ok }
func (s *Set[T]) Len() int { return len(s.m) }

func main() {
	set := NewSet[int]()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "add":
			var v int
			fmt.Sscan(parts[1], &v)
			set.Add(v)
		case "has":
			var v int
			fmt.Sscan(parts[1], &v)
			fmt.Println(set.Has(v))
		case "len":
			fmt.Println(set.Len())
		}
	},
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Set[T comparable] struct {
	m map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{m: make(map[T]struct{})}
}

// TODO: реализуй методы
func (s *Set[T]) Add(v T) { /* s.m[v] = struct{}{} */ }
func (s *Set[T]) Has(v T) bool { return false }
func (s *Set[T]) Len() int { return 0 }

func main() {
	set := NewSet[int]()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "add":
			var v int
			fmt.Sscan(parts[1], &v)
			set.Add(v)
		case "has":
			var v int
			fmt.Sscan(parts[1], &v)
			fmt.Println(set.Has(v))
		case "len":
			fmt.Println(set.Len())
		}
	},
}`,
						TestCases: []TestCase{
							{Input: "add 5\nadd 3\nadd 5\nhas 5\nhas 7\nlen", ExpectedOutput: "true\nfalse\n2"},
							{Input: "add 1\nadd 2\nadd 3\nlen\nhas 2\nhas 4", ExpectedOutput: "3\ntrue\nfalse"},
						},
					},
					{
						Title:      "Reduce — обобщённая свёртка",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "Reduce[T, U any]", Definition: "Свёртка: проход по слайсу с накоплением результата. Два параметра типа: T — элемент, U — аккумулятор."},
							{Term: "func(U, T) U", Definition: "Функция-аккумулятор: принимает текущий результат и следующий элемент, возвращает новый результат."},
						},
						Description: `<p>Реализуй обобщённую функцию <code>Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U</code>.</p>
<p>Ввод:</p>
<pre><code>5
1 2 3 4 5</code></pre>
<p>Вывод (сумма и произведение через Reduce):</p>
<pre><code>Sum: 15
Product: 120</code></pre>`,
						Hints: `<p><code>acc := initial; for _, v := range slice { acc = fn(acc, v) }; return acc</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	acc := initial
	for _, v := range slice {
		acc = fn(acc, v)
	},
	return acc
}

func main() {
	var n int
	fmt.Scan(&n)
	nums := make([]int, n)
	for i := range nums {
		fmt.Scan(&nums[i])
	},

	sum := Reduce(nums, 0, func(acc, v int) int { return acc + v })
	product := Reduce(nums, 1, func(acc, v int) int { return acc * v })

	fmt.Printf("Sum: %d\nProduct: %d\n", sum, product)
}</code></pre>`,
						StarterCode: `package main

import "fmt"

func Reduce[T, U any](slice []T, initial U, fn func(U, T) U) U {
	// TODO: пройди по slice, применяя fn к аккумулятору
	// acc := initial
	// for _, v := range slice { acc = fn(acc, v) }
	return initial
}

func main() {
	var n int
	fmt.Scan(&n)
	nums := make([]int, n)
	for i := range nums {
		fmt.Scan(&nums[i])
	},

	sum := Reduce(nums, 0, func(acc, v int) int { return acc + v })
	product := Reduce(nums, 1, func(acc, v int) int { return acc * v })

	fmt.Printf("Sum: %d\nProduct: %d\n", sum, product)
}`,
						TestCases: []TestCase{
							{Input: "5\n1 2 3 4 5", ExpectedOutput: "Sum: 15\nProduct: 120"},
							{Input: "3\n10 20 30", ExpectedOutput: "Sum: 60\nProduct: 6000"},
						},
					},
					{
						Title:      "Min с custom constraint",
						Difficulty: "easy",
						Description: `<p>Создай constraint <code>Number</code> с <code>~int | ~float64</code> и функцию <code>Min[T Number](a, b T) T</code>:</p>
<p>Ввод: <code>15 7</code></p>
<p>Вывод: <code>Min: 7</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "~int", Definition: "Underlying type. Включает int и type Age int."},
						},
						TestCases: []TestCase{
							{Input: "15 7", ExpectedOutput: "Min: 7"},
							{Input: "3 8", ExpectedOutput: "Min: 3"},
						},
						StarterCode: `package main

import "fmt"

type Number interface {
    ~int | ~float64
}

func Min[T Number](a, b T) T {
    if a < b { return a }
    return b
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Printf("Min: %d
", Min(a, b))
}`,
						Hints: `<p><code>if a &lt; b { return a }; return b</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

type Number interface { ~int | ~float64 }

func Min[T Number](a, b T) T {
    if a < b { return a }
    return b
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Printf("Min: %d
", Min(a, b))
}</code></pre>`,
					},
					{
						Title:      "GroupBy — группировка",
						Difficulty: "medium",
						Description: `<p>Напиши <code>GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T</code>:</p>
<p>Ввод:</p>
<pre><code>6
apple banana avocado blueberry cherry cranberry</code></pre>
<p>Вывод:</p>
<pre><code>a: [apple avocado]
b: [banana blueberry]
c: [cherry cranberry]</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "GroupBy[T, K]", Definition: "Два параметра типа: T — элемент, K comparable — ключ map."},
						},
						TestCases: []TestCase{
							{Input: "6\napple banana avocado blueberry cherry cranberry", ExpectedOutput: "a: [apple avocado]\nb: [banana blueberry]\nc: [cherry cranberry]"},
						},
						StarterCode: `package main

import (
    "fmt"
    "sort"
)

func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, item := range items {
        key := keyFn(item)
        result[key] = append(result[key], item)
    }
    return result
}

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words { fmt.Scan(&words[i]) }
    grouped := GroupBy(words, func(s string) string { return string(s[0]) })
    keys := make([]string, 0, len(grouped))
    for k := range grouped { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys { fmt.Printf("%s: %v
", k, grouped[k]) }
}`,
						Hints: `<p><code>result[key] = append(result[key], item)</code></p>`,
						Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
    result := make(map[K][]T)
    for _, item := range items {
        result[keyFn(item)] = append(result[keyFn(item)], item)
    }
    return result
}

func main() {
    var n int
    fmt.Scan(&n)
    words := make([]string, n)
    for i := range words { fmt.Scan(&words[i]) }
    grouped := GroupBy(words, func(s string) string { return string(s[0]) })
    keys := make([]string, 0, len(grouped))
    for k := range grouped { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys { fmt.Printf("%s: %v
", k, grouped[k]) }
}</code></pre>`,
					},
					{
						Title:      "Generic Pair",
						Difficulty: "easy",
						Description: `<p>Создай обобщённую структуру <code>Pair[A, B any]</code> с методом <code>Swap() Pair[B, A]</code>:</p>
<p>Ввод: <code>hello 42</code></p>
<p>Вывод:</p>
<pre><code>Original: {hello 42}
Swapped: {42 hello}</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Pair[A, B any]", Definition: "Обобщённая пара с двумя разными типами. Swap возвращает Pair[B, A]."},
						},
						TestCases: []TestCase{
							{Input: "hello 42", ExpectedOutput: "Original: {hello 42}\nSwapped: {42 hello}"},
						},
						StarterCode: `package main

import "fmt"

type Pair[A, B any] struct {
    First  A
    Second B
}

func (p Pair[A, B]) Swap() Pair[B, A] {
    return Pair[B, A]{First: p.Second, Second: p.First}
}

func main() {
    var s string
    var n int
    fmt.Scan(&s, &n)
    p := Pair[string, int]{First: s, Second: n}
    fmt.Printf("Original: {%v %v}\n", p.First, p.Second)
    swapped := p.Swap()
    fmt.Printf("Swapped: {%v %v}\n", swapped.First, swapped.Second)
}`,
						Hints: `<p>Swap возвращает Pair[B, A]{First: p.Second, Second: p.First}</p>`,
						Solution: `<pre><code>package main

import "fmt"

type Pair[A, B any] struct {
    First  A
    Second B
}

func (p Pair[A, B]) Swap() Pair[B, A] {
    return Pair[B, A]{First: p.Second, Second: p.First}
}

func main() {
    var s string
    var n int
    fmt.Scan(&s, &n)
    p := Pair[string, int]{s, n}
    fmt.Printf("Original: {%v %v}\n", p.First, p.Second)
    swapped := p.Swap()
    fmt.Printf("Swapped: {%v %v}\n", swapped.First, swapped.Second)
}</code></pre>`,
					},
				},
			},
		},
	}
}
