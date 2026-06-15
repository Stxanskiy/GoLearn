package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 3: Функции
// ════════════════════════════════════════════════════════════════

func mod03_functions_new() M {
	return M{
		Slug:        "functions",
		Title:       "Функции",
		Description: "Как разбивать код на переиспользуемые блоки. Параметры, возвращаемые значения, замыкания.",
		Order:       3,
		Track:       "shared",
		Difficulty:  "beginner",
		Prerequisites: []string{"collections"},
		Lessons: []L{
			{
				Slug: "func-basics", Title: "Функции — основы", Order: 1,
				Difficulty: "beginner", Track: "shared",
				Content: `<h1>Функции — переиспользуемые блоки кода</h1>

<h2>Зачем нужны функции?</h2>
<p>Без функций код превращается в длинную простыню, где одна и та же логика копируется снова и снова. <strong>Функция</strong> — это именованный блок кода, который можно вызывать многократно.</p>

<p>Аналогия: функция — как рецепт. Ты описываешь рецепт один раз, а потом "вызываешь" его каждый раз, когда хочешь приготовить блюдо.</p>

<h2>Синтаксис функции</h2>
<pre><code>// Функция без параметров и без возвращаемого значения
func sayHello() {
    fmt.Println("Hello!")
}

// Функция с параметрами
func greet(name string) {
    fmt.Printf("Hello, %s!\n", name)
}

// Функция с возвращаемым значением
func add(a, b int) int {
    return a + b
}

// Вызов функций
sayHello()          // Hello!
greet("Alice")      // Hello, Alice!
result := add(3, 5) // 8</code></pre>

<p>Формула: <code>func имя(параметры) тип_возврата { тело }</code></p>

<h2>Несколько возвращаемых значений</h2>
<p>Go позволяет возвращать <strong>несколько значений</strong> — это уникальная особенность языка:</p>
<pre><code>func divide(a, b float64) (float64, error) {
    if b == 0 {
        return 0, fmt.Errorf("division by zero")
    }
    return a / b, nil
}

result, err := divide(10, 3)
if err != nil {
    fmt.Println("Error:", err)
} else {
    fmt.Println("Result:", result)
}</code></pre>

<h2>Именованные возвращаемые значения</h2>
<pre><code>func minMax(nums []int) (min, max int) {
    min = nums[0]
    max = nums[0]
    for _, v := range nums {
        if v < min { min = v }
        if v > max { max = v }
    }
    return // "голый" return — возвращает min и max
}</code></pre>

<h2>Вариативные функции (variadic)</h2>
<pre><code>// ... означает "любое количество аргументов"
func sum(nums ...int) int {
    total := 0
    for _, n := range nums {
        total += n
    }
    return total
}

fmt.Println(sum(1, 2, 3))     // 6
fmt.Println(sum(10, 20))      // 30

// Передать слайс — через ...
numbers := []int{1, 2, 3, 4, 5}
fmt.Println(sum(numbers...))   // 15</code></pre>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/3.1.php" target="_blank">Metanit: Функции в Go</a> — синтаксис, параметры, возвращаемые значения</li>
<li><a href="https://golangify.com/functions" target="_blank">Golangify: Функции</a> — variadic, именованные возвраты, типичные паттерны</li>
<li><a href="https://habr.com/ru/articles/197578/" target="_blank">Хабр: Функции в Go</a> — подробный разбор с примерами</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Сколько значений может вернуть функция в Go?",
						Options:     []string{"Только одно", "Любое количество — это особенность Go", "Максимум два", "Зависит от типа"},
						Correct:     1,
						Explanation: "Go поддерживает множественные возвращаемые значения. Чаще всего используют два: результат + ошибка (val, err).",
					},
					{
						Question:    "Что означает ... в параметрах функции?",
						Options:     []string{"Комментарий", "Функция принимает любое количество аргументов этого типа", "Необязательный параметр", "Многоточие"},
						Correct:     1,
						Explanation: "... делает параметр вариативным. func sum(nums ...int) принимает 0 или более int. Внутри функции nums — обычный []int.",
					},
					{
						Question:    "Как передать слайс в вариативную функцию?",
						Options:     []string{"func(slice)", "func(slice...)", "func(&slice)", "func(*slice)"},
						Correct:     1,
						Explanation: "Для передачи слайса в variadic-функцию используй ...: sum(nums...). Это 'разворачивает' слайс в список аргументов.",
					},
					{
						Question:    "Что такое именованные возвращаемые значения?",
						Options:     []string{"Алиасы для типов", "Переменные объявленные прямо в сигнатуре функции — можно вернуть голым return", "Обязательные параметры", "Глобальные переменные"},
						Correct:     1,
						Explanation: "func f() (result int, err error) объявляет result и err как локальные переменные. 'Голый' return вернёт их текущие значения. Удобно но снижает читаемость — используй осознанно.",
					},
					{
						Question:    "Что происходит когда функция рекурсивно вызывает себя без условия выхода?",
						Options:     []string{"Бесконечный цикл", "Stack overflow: panic: runtime: goroutine stack exceeds limit", "Программа просто остановится", "Go оптимизирует хвостовую рекурсию"},
						Correct:     1,
						Explanation: "Каждый вызов создаёт фрейм стека. Без условия выхода стек переполняется → panic. Go НЕ оптимизирует хвостовую рекурсию (в отличие от Haskell/Erlang).",
					},
				},
				Tasks: []T{
					{
						Title:      "Математические функции",
						Difficulty: "easy",
						Description: `<p>Создай три функции и используй их:</p>
<ul>
<li><code>max(a, b int) int</code> — возвращает большее из двух чисел</li>
<li><code>abs(n int) int</code> — возвращает модуль числа (без знака)</li>
<li><code>isEven(n int) bool</code> — возвращает true если число чётное</li>
</ul>
<p>Ввод: <code>-7 3</code></p>
<p>Вывод:</p>
<pre><code>Max: 3
Abs(-7): 7
IsEven(3): false</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "func name(params) returnType", Definition: "Объявление функции. params — входные данные, returnType — что функция вернёт."},
							{Term: "return value", Definition: "Возвращает значение из функции в место вызова. Функция завершается после return."},
						},
						TestCases: []TestCase{
							{Input: "-7 3", ExpectedOutput: "Max: 3\nAbs(-7): 7\nIsEven(3): false"},
							{Input: "4 2", ExpectedOutput: "Max: 4\nAbs(4): 4\nIsEven(2): true"},
						},
						StarterCode: `package main

import "fmt"

// Напиши функции max, abs, isEven здесь

func main() {
    var a, b int
    fmt.Scan(&a, &b)

    fmt.Printf("Max: %d\n", max(a, b))
    fmt.Printf("Abs(%d): %d\n", a, abs(a))
    fmt.Printf("IsEven(%d): %v\n", b, isEven(b))
}`,
						Hints:    `<p>max: <code>if a > b { return a } return b</code>. abs: <code>if n < 0 { return -n } return n</code>. isEven: <code>return n%2 == 0</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func max(a, b int) int {
    if a > b {
        return a
    }
    return b
}

func abs(n int) int {
    if n < 0 {
        return -n
    }
    return n
}

func isEven(n int) bool {
    return n%2 == 0
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)

    fmt.Printf("Max: %d\n", max(a, b))
    fmt.Printf("Abs(%d): %d\n", a, abs(a))
    fmt.Printf("IsEven(%d): %v\n", b, isEven(b))
}</code></pre>`,
					},
					{
						Title:      "Статистика слайса",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>stats(nums []int) (min, max, sum int)</code> с именованными возвращаемыми значениями.</p>
<p>Ввод:</p>
<pre><code>5
3 7 1 9 4</code></pre>
<p>Вывод:</p>
<pre><code>Min: 1, Max: 9, Sum: 24</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "именованные возвраты", Definition: "func f() (min, max int) — min и max объявлены в сигнатуре. return без аргументов вернёт их текущие значения."},
						},
						TestCases: []TestCase{
							{Input: "5\n3 7 1 9 4", ExpectedOutput: "Min: 1, Max: 9, Sum: 24"},
							{Input: "3\n10 20 30", ExpectedOutput: "Min: 10, Max: 30, Sum: 60"},
						},
						StarterCode: `package main

import "fmt"

func stats(nums []int) (min, max, sum int) {
    // Реализуй: найди min, max, sum
    return
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    min, max, sum := stats(nums)
    fmt.Printf("Min: %d, Max: %d, Sum: %d\n", min, max, sum)
}`,
						Hints:    `<p>Начни: <code>min = nums[0]; max = nums[0]</code>. В цикле обновляй min, max, sum.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func stats(nums []int) (min, max, sum int) {
    min = nums[0]
    max = nums[0]
    for _, v := range nums {
        if v < min { min = v }
        if v > max { max = v }
        sum += v
    }
    return
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    min, max, sum := stats(nums)
    fmt.Printf("Min: %d, Max: %d, Sum: %d\n", min, max, sum)
}</code></pre>`,
					},
					{
						Title:      "Рекурсия: факториал",
						Difficulty: "easy",
						Description: `<p>Напиши рекурсивную функцию <code>factorial(n int) int</code> и вычисли факториал числа:</p>
<p>Ввод: <code>5</code> → Вывод: <code>120</code></p>
<p>Факториал: 5! = 5 × 4 × 3 × 2 × 1 = 120. Базовый случай: 0! = 1.</p>`,
						Glossary: []GlossaryItem{
							{Term: "рекурсия", Definition: "Функция, которая вызывает сама себя. Обязательно нужен базовый случай (условие выхода), иначе — бесконечный вызов и stack overflow."},
							{Term: "базовый случай", Definition: "Условие при котором рекурсия останавливается. Для факториала: if n == 0 { return 1 }."},
						},
						TestCases: []TestCase{
							{Input: "5", ExpectedOutput: "120"},
							{Input: "0", ExpectedOutput: "1"},
							{Input: "7", ExpectedOutput: "5040"},
						},
						StarterCode: `package main

import "fmt"

func factorial(n int) int {
    // Базовый случай: если n == 0, вернуть 1
    // Рекурсивный случай: n * factorial(n-1)
    return 0
}

func main() {
    var n int
    fmt.Scan(&n)
    fmt.Println(factorial(n))
}`,
						Hints:    `<p>Условие выхода: <code>if n == 0 { return 1 }</code>. Рекурсия: <code>return n * factorial(n-1)</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func factorial(n int) int {
    if n == 0 {
        return 1
    }
    return n * factorial(n-1)
}

func main() {
    var n int
    fmt.Scan(&n)
    fmt.Println(factorial(n))
}</code></pre>`,
					},
					{
						Title:      "Функция-трансформер",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>mapSlice(nums []int, fn func(int) int) []int</code>, которая применяет fn к каждому элементу и возвращает новый слайс. Использует её для возведения в квадрат:</p>
<p>Ввод:</p>
<pre><code>4
1 2 3 4</code></pre>
<p>Вывод: <code>1 4 9 16</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "func(int) int", Definition: "Тип функции: принимает int, возвращает int. Позволяет передавать трансформацию как аргумент."},
							{Term: "map (функциональный)", Definition: "Классическая операция: применить функцию к каждому элементу коллекции. В Go реализуется вручную."},
						},
						TestCases: []TestCase{
							{Input: "4\n1 2 3 4", ExpectedOutput: "1 4 9 16"},
							{Input: "3\n2 3 4", ExpectedOutput: "4 9 16"},
						},
						StarterCode: `package main

import "fmt"

func mapSlice(nums []int, fn func(int) int) []int {
    // Создай новый слайс и примени fn к каждому элементу
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    square := func(x int) int { return x * x }
    result := mapSlice(nums, square)

    for i, v := range result {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}`,
						Hints:    `<p>Создай <code>result := make([]int, len(nums))</code>. В цикле: <code>result[i] = fn(nums[i])</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func mapSlice(nums []int, fn func(int) int) []int {
    result := make([]int, len(nums))
    for i, v := range nums {
        result[i] = fn(v)
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

    square := func(x int) int { return x * x }
    result := mapSlice(nums, square)

    for i, v := range result {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}</code></pre>`,
					},
					{
						Title:      "Рекурсия: числа Фибоначчи",
						Difficulty: "hard",
						Description: `<p>Напиши функцию <code>fib(n int) int</code> которая возвращает N-ое число Фибоначчи. Но наивная рекурсия будет очень медленной — используй мемоизацию (кэш результатов в map):</p>
<p>Ввод: <code>10</code> → Вывод: <code>55</code></p>
<p>Фибоначчи: 0, 1, 1, 2, 3, 5, 8, 13, 21, 34, <strong>55</strong>...</p>`,
						Glossary: []GlossaryItem{
							{Term: "мемоизация", Definition: "Кэширование результатов функции. Если fib(n) уже вычислен — вернуть из кэша вместо повторного вычисления. Снижает O(2^n) до O(n)."},
							{Term: "cache[n]", Definition: "map[int]int для хранения уже вычисленных значений. Передаётся по ссылке в рекурсивные вызовы."},
						},
						TestCases: []TestCase{
							{Input: "10", ExpectedOutput: "55"},
							{Input: "0", ExpectedOutput: "0"},
							{Input: "20", ExpectedOutput: "6765"},
						},
						StarterCode: `package main

import "fmt"

func fibMemo(n int, cache map[int]int) int {
    if n <= 1 {
        return n
    }
    if val, ok := cache[n]; ok {
        return val
    }
    result := fibMemo(n-1, cache) + fibMemo(n-2, cache)
    cache[n] = result
    return result
}

func main() {
    var n int
    fmt.Scan(&n)
    cache := make(map[int]int)
    fmt.Println(fibMemo(n, cache))
}`,
						Hints:    `<p>Перед рекурсией: <code>if val, ok := cache[n]; ok { return val }</code>. После вычисления: <code>cache[n] = result</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func fibMemo(n int, cache map[int]int) int {
    if n <= 1 {
        return n
    }
    if val, ok := cache[n]; ok {
        return val
    }
    result := fibMemo(n-1, cache) + fibMemo(n-2, cache)
    cache[n] = result
    return result
}

func main() {
    var n int
    fmt.Scan(&n)
    cache := make(map[int]int)
    fmt.Println(fibMemo(n, cache))
}</code></pre>`,
					},
				},
			},
			{
				Slug: "closures", Title: "Замыкания и функции как значения", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Замыкания и функции как значения</h1>

<h2>Функция — это значение</h2>
<p>В Go функцию можно присвоить переменной, передать как аргумент или вернуть из другой функции:</p>

<pre><code>// Присвоить функцию переменной
add := func(a, b int) int { return a + b }
fmt.Println(add(3, 5))  // 8

// Передать функцию как аргумент
func apply(nums []int, fn func(int) int) []int {
    result := make([]int, len(nums))
    for i, v := range nums {
        result[i] = fn(v)
    }
    return result
}

doubled := apply([]int{1,2,3}, func(n int) int { return n * 2 })
// [2, 4, 6]</code></pre>

<h2>Замыкание (closure)</h2>
<p><strong>Замыкание</strong> — функция, которая "помнит" переменные из окружающего scope:</p>

<pre><code>func counter() func() int {
    count := 0
    return func() int {
        count++     // функция "захватила" count
        return count
    }
}

c := counter()
fmt.Println(c())  // 1
fmt.Println(c())  // 2
fmt.Println(c())  // 3
// count живёт внутри замыкания!</code></pre>

<h2>defer — отложенный вызов</h2>
<pre><code>func readFile() {
    f, _ := os.Open("data.txt")
    defer f.Close()  // закроет файл при ВЫХОДЕ из функции

    // ... работаем с файлом ...
    // f.Close() вызовется автоматически
}

// defer выполняется в порядке LIFO (последний добавленный — первый)
defer fmt.Println("1")
defer fmt.Println("2")
defer fmt.Println("3")
// Выведет: 3, 2, 1</code></pre>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/3.7.php" target="_blank">Metanit: Замыкания в Go</a> — что такое closure и как работает захват переменных</li>
<li><a href="https://habr.com/ru/articles/419371/" target="_blank">Хабр: defer, panic, recover в Go</a> — подробный разбор отложенных вызовов</li>
<li><a href="https://golangify.com/closures" target="_blank">Golangify: Замыкания</a> — практические примеры: счётчики, мидлвары, генераторы</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Что такое замыкание?",
						Options:     []string{"Функция без имени", "Функция, которая захватывает и помнит переменные из окружающего scope", "Функция с ошибкой", "Бесконечный цикл"},
						Correct:     1,
						Explanation: "Замыкание — функция, привязанная к своему лексическому окружению. Она 'помнит' переменные, даже когда внешняя функция завершилась.",
					},
					{
						Question:    "Когда выполняется defer?",
						Options:     []string{"Сразу", "При выходе из функции (даже при panic)", "Через 1 секунду", "Никогда"},
						Correct:     1,
						Explanation: "defer откладывает вызов до возврата из функции. Идеально для освобождения ресурсов: Close(), Unlock(), и т.д.",
					},
					{
						Question:    "Что выведет: x := 1; defer fmt.Println(x); x = 99?",
						Options:     []string{"99", "1 — аргументы defer вычисляются в момент вызова defer", "Ошибка", "Ничего"},
						Correct:     1,
						Explanation: "Аргументы defer вычисляются СРАЗУ при вызове defer. x=1 запоминается в момент defer fmt.Println(x). Изменение x=99 не влияет на defer. Для доступа к текущему значению — используй замыкание: defer func() { fmt.Println(x) }().",
					},
					{
						Question:    "Что такое функция первого класса (first-class function)?",
						Options:     []string{"Функция с одним параметром", "Функция которая может быть присвоена переменной, передана и возвращена как значение", "Главная функция main", "Функция без ошибок"},
						Correct:     1,
						Explanation: "В Go функции — first-class values: их можно присвоить переменной (f := func(){}), передать аргументом, вернуть из другой функции. Это основа замыканий, middleware и функционального стиля.",
					},
					{
						Question:    "Как работает мемоизация через замыкание?",
						Options:     []string{"Никак, замыкания не могут хранить данные", "Замыкание захватывает map-кэш и проверяет его перед вычислением", "Только через глобальные переменные", "Только с рекурсивными функциями"},
						Correct:     1,
						Explanation: "Замыкание захватывает cache := map[int]int{}. При каждом вызове сначала проверяет кэш. Если результат есть — возвращает его. Иначе — вычисляет, сохраняет в cache, возвращает. Кэш живёт пока живёт замыкание.",
					},
				},
				Tasks: []T{
					{
						Title:      "Счётчик вызовов",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>makeCounter() func() int</code>, которая возвращает функцию-счётчик. Каждый вызов увеличивает счётчик на 1 и возвращает текущее значение:</p>
<pre><code>1
2
3</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "захват переменной", Definition: "Замыкание 'видит' переменные из внешней функции. count := 0 объявлен в makeCounter, но живёт пока живёт возвращённая функция."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "1\n2\n3"},
						},
						StarterCode: `package main

import "fmt"

func makeCounter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    c := makeCounter()
    fmt.Println(c())
    fmt.Println(c())
    fmt.Println(c())
}`,
						Hints:    `<p>Код уже написан. Запусти и убедись что каждый вызов возвращает следующее число.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func makeCounter() func() int {
    count := 0
    return func() int {
        count++
        return count
    }
}

func main() {
    c := makeCounter()
    fmt.Println(c())
    fmt.Println(c())
    fmt.Println(c())
}</code></pre>`,
					},
					{
						Title:      "Функция-фильтр",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>filter</code>, которая принимает слайс int и функцию-предикат, и возвращает отфильтрованный слайс.</p>
<p>Ввод:</p>
<pre><code>6
1 2 3 4 5 6</code></pre>
<p>Вывод (только чётные):</p>
<pre><code>Filtered: 2 4 6</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "func(int) bool", Definition: "Тип функции — принимает int, возвращает bool. Используется как предикат (условие для фильтрации)."},
							{Term: "предикат (predicate)", Definition: "Функция, которая возвращает true/false. Используется для фильтрации: 'оставить этот элемент?'"},
						},
						TestCases: []TestCase{
							{Input: "6\n1 2 3 4 5 6", ExpectedOutput: "Filtered: 2 4 6"},
							{Input: "4\n10 15 20 25", ExpectedOutput: "Filtered: 10 20"},
						},
						StarterCode: `package main

import "fmt"

func filter(nums []int, pred func(int) bool) []int {
    // Верни новый слайс с элементами, для которых pred вернул true
    return nil
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums {
        fmt.Scan(&nums[i])
    }

    isEven := func(n int) bool { return n%2 == 0 }
    result := filter(nums, isEven)

    fmt.Print("Filtered:")
    for _, v := range result {
        fmt.Printf(" %d", v)
    }
    fmt.Println()
}`,
						Hints:    `<p>Создай <code>result := []int{}</code>, в цикле: <code>if pred(v) { result = append(result, v) }</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func filter(nums []int, pred func(int) bool) []int {
    result := []int{}
    for _, v := range nums {
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

    isEven := func(n int) bool { return n%2 == 0 }
    result := filter(nums, isEven)

    fmt.Print("Filtered:")
    for _, v := range result {
        fmt.Printf(" %d", v)
    }
    fmt.Println()
}</code></pre>`,
					},
					{
						Title:      "Rate limiter (реальный кейс)",
						Difficulty: "hard",
						Description: `<p>Создай rate limiter через замыкание — это реальная задача в backend-разработке. Функция <code>newLimiter(maxCalls int)</code> возвращает функцию, которая:</p>
<ul>
<li>Первые maxCalls вызовов → <code>allowed</code></li>
<li>После maxCalls → <code>blocked</code></li>
</ul>
<p>Ввод:</p>
<pre><code>3 5</code></pre>
<p>(лимит = 3, попыток = 5)</p>
<p>Вывод:</p>
<pre><code>Call 1: allowed
Call 2: allowed
Call 3: allowed
Call 4: blocked
Call 5: blocked</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "func() T (замыкание)", Definition: "Функция без параметров, которая помнит переменные из внешней функции. count сохраняется между вызовами."},
						},
						TestCases: []TestCase{
							{Input: "3 5", ExpectedOutput: "Call 1: allowed\nCall 2: allowed\nCall 3: allowed\nCall 4: blocked\nCall 5: blocked"},
							{Input: "1 3", ExpectedOutput: "Call 1: allowed\nCall 2: blocked\nCall 3: blocked"},
						},
						StarterCode: `package main

import "fmt"

func newLimiter(maxCalls int) func() bool {
    // Верни функцию-замыкание, которая помнит счётчик вызовов
    return nil
}

func main() {
    var limit, attempts int
    fmt.Scan(&limit, &attempts)

    check := newLimiter(limit)
    for i := 1; i <= attempts; i++ {
        if check() {
            fmt.Printf("Call %d: allowed\n", i)
        } else {
            fmt.Printf("Call %d: blocked\n", i)
        }
    }
}`,
						Hints:    `<p>Внутри newLimiter создай <code>count := 0</code>. Верни <code>func() bool { count++; return count <= maxCalls }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func newLimiter(maxCalls int) func() bool {
    count := 0
    return func() bool {
        count++
        return count <= maxCalls
    }
}

func main() {
    var limit, attempts int
    fmt.Scan(&limit, &attempts)

    check := newLimiter(limit)
    for i := 1; i <= attempts; i++ {
        if check() {
            fmt.Printf("Call %d: allowed\n", i)
        } else {
            fmt.Printf("Call %d: blocked\n", i)
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Генератор ID",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>makeIDGenerator(prefix string) func() string</code>, которая возвращает генератор уникальных ID:</p>
<p>Вывод:</p>
<pre><code>user-1
user-2
user-3</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "fmt.Sprintf", Definition: "Форматирует строку и возвращает её (в отличие от Printf который печатает). Используй для создания строк с числами."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "user-1\nuser-2\nuser-3"},
						},
						StarterCode: `package main

import "fmt"

func makeIDGenerator(prefix string) func() string {
    id := 0
    return func() string {
        id++
        return fmt.Sprintf("%s-%d", prefix, id)
    }
}

func main() {
    nextID := makeIDGenerator("user")
    fmt.Println(nextID())
    fmt.Println(nextID())
    fmt.Println(nextID())
}`,
						Hints:    `<p>Замыкание захватывает <code>id := 0</code>. При каждом вызове: <code>id++; return fmt.Sprintf("%s-%d", prefix, id)</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func makeIDGenerator(prefix string) func() string {
    id := 0
    return func() string {
        id++
        return fmt.Sprintf("%s-%d", prefix, id)
    }
}

func main() {
    nextID := makeIDGenerator("user")
    fmt.Println(nextID())
    fmt.Println(nextID())
    fmt.Println(nextID())
}</code></pre>`,
					},
					{
						Title:      "Мемоизация через замыкание",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>memoize(fn func(int) int) func(int) int</code>, которая кэширует результаты любой функции:</p>
<p>Ввод: <code>5</code></p>
<p>Вывод:</p>
<pre><code>computing 5
25
25</code></pre>
<p>Второй вызов с тем же аргументом возвращает результат из кэша без "computing".</p>`,
						Glossary: []GlossaryItem{
							{Term: "map[int]int как кэш", Definition: "Замыкание захватывает cache. При первом вызове с аргументом — вычисляет и сохраняет. При повторном — возвращает из кэша."},
						},
						TestCases: []TestCase{
							{Input: "5", ExpectedOutput: "computing 5\n25\n25"},
						},
						StarterCode: `package main

import "fmt"

func memoize(fn func(int) int) func(int) int {
    cache := map[int]int{}
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val
        }
        fmt.Printf("computing %d\n", n)
        result := fn(n)
        cache[n] = result
        return result
    }
}

func main() {
    var n int
    fmt.Scan(&n)

    square := func(x int) int { return x * x }
    cachedSquare := memoize(square)

    fmt.Println(cachedSquare(n)) // вычисляет
    fmt.Println(cachedSquare(n)) // из кэша
}`,
						Hints:    `<p>Код уже почти готов в StarterCode — изучи его и запусти. Задача помочь понять как работает мемоизация на практике.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func memoize(fn func(int) int) func(int) int {
    cache := map[int]int{}
    return func(n int) int {
        if val, ok := cache[n]; ok {
            return val
        }
        fmt.Printf("computing %d\n", n)
        result := fn(n)
        cache[n] = result
        return result
    }
}

func main() {
    var n int
    fmt.Scan(&n)

    square := func(x int) int { return x * x }
    cachedSquare := memoize(square)

    fmt.Println(cachedSquare(n))
    fmt.Println(cachedSquare(n))
}</code></pre>`,
					},
				},
			},
			{
				Slug: "func-patterns", Title: "Под капотом: ловушки и паттерны", Order: 3,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Функции — под капотом и ловушки</h1>

<h2>Под капотом: стек вызовов</h2>
<p>Каждый вызов функции создаёт <strong>фрейм стека</strong> (stack frame) — область памяти для локальных переменных и адреса возврата:</p>
<pre><code>func main() {          // фрейм main
    result := calc(5)  // создаёт фрейм calc
}

func calc(n int) int { // фрейм calc: n=5, total=?
    total := n * 2     // total=10
    return total       // фрейм уничтожается, 10 возвращается в main
}</code></pre>

<p><strong>Стек vs Куча:</strong></p>
<ul>
<li><strong>Стек</strong> — быстро (просто сдвигает указатель). Локальные переменные живут тут.</li>
<li><strong>Куча</strong> — медленнее (нужен сборщик мусора). Если переменная "убегает" из функции — Go переносит её на кучу.</li>
</ul>

<pre><code>// Escape analysis — компилятор решает где хранить переменную
// Проверь: go build -gcflags="-m" main.go

func createOnStack() int {
    x := 42    // живёт на стеке — быстро
    return x   // копия значения
}

func createOnHeap() *int {
    x := 42    // "убегает" через указатель → Go перемещает на кучу
    return &x  // указатель валиден после возврата!
}</code></pre>

<h2>⚠️ Ловушка #1: замыкание в цикле</h2>
<p>Это самый частый баг с замыканиями. Замыкание захватывает <strong>переменную, а не значение</strong>:</p>

<pre><code>// БАГИ: все горутины/функции видят ПОСЛЕДНЕЕ значение i
funcs := []func(){}
for i := 0; i < 3; i++ {
    funcs = append(funcs, func() {
        fmt.Println(i) // i — та же переменная!
    })
}
for _, f := range funcs {
    f() // Выведет: 3, 3, 3  (а не 0, 1, 2!)
}

// ИСПРАВЛЕНИЕ 1: передать как аргумент (копия)
for i := 0; i < 3; i++ {
    i := i  // shadowing — создаёт новую переменную i
    funcs = append(funcs, func() {
        fmt.Println(i) // теперь у каждой своя копия
    })
}
// Выведет: 0, 1, 2 ✅</code></pre>

<p><strong>Правило:</strong> если замыкание используется в цикле — всегда делай локальную копию переменной цикла.</p>

<h2>defer — глубоко</h2>
<p><strong>LIFO (Last In, First Out)</strong> — defer выполняются в обратном порядке:</p>

<pre><code>func cleanup() {
    fmt.Println("Start")
    defer fmt.Println("First defer")   // последний
    defer fmt.Println("Second defer")  // второй
    defer fmt.Println("Third defer")   // первый!
    fmt.Println("End")
}
// Start → End → Third defer → Second defer → First defer</code></pre>

<p><strong>defer вычисляет аргументы сразу:</strong></p>
<pre><code>x := 10
defer fmt.Println(x) // запомнит 10, а не текущее значение x
x = 20
// Выведет: 10 (а не 20!)

// Но замыкание в defer видит текущее значение:
defer func() { fmt.Println(x) }()  // Выведет: 20</code></pre>

<h2>Паттерн: функция-обёртка (middleware)</h2>
<p>В backend-разработке функции-обёртки — основа middleware:</p>

<pre><code>// Обёртка: замеряет время выполнения любой функции
func measure(name string, fn func()) {
    start := time.Now()
    fn()
    fmt.Printf("%s took %v\n", name, time.Since(start))
}

measure("sort", func() {
    sort.Ints(bigSlice)
})
// sort took 1.234ms</code></pre>

<h2>Паттерн: функция-конструктор опций</h2>
<pre><code>type Config struct {
    Port    int
    Verbose bool
}

type Option func(*Config)

func WithPort(p int) Option     { return func(c *Config) { c.Port = p } }
func WithVerbose() Option       { return func(c *Config) { c.Verbose = true } }

func NewServer(opts ...Option) *Config {
    cfg := &Config{Port: 8080} // defaults
    for _, opt := range opts {
        opt(cfg) // применяем каждую опцию
    }
    return cfg
}

srv := NewServer(WithPort(3000), WithVerbose())
// {Port: 3000, Verbose: true}</code></pre>
<p>Этот паттерн (functional options) используется в chi, slog, pgx и десятках других Go-библиотек.</p>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://habr.com/ru/articles/477634/" target="_blank">Хабр: Functional Options паттерн в Go</a> — реальный паттерн из крупных проектов</li>
<li><a href="https://habr.com/ru/articles/348674/" target="_blank">Хабр: Escape analysis и оптимизация Go</a> — стек vs куча, как компилятор принимает решения</li>
<li><a href="https://golangify.com/defer" target="_blank">Golangify: defer подробно</a> — LIFO, аргументы, ловушки в цикле</li>
</ul>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: defer в цикле — ресурсы не освобождаются до конца функции
for _, name := range files {
    f, _ := os.Open(name)
    defer f.Close()  // ВСЕ файлы закроются только при выходе из main!
}

// ПРАВИЛЬНО: вынести в отдельную функцию
func processFile(name string) {
    f, _ := os.Open(name)
    defer f.Close()  // закроется при выходе из processFile
    // ...
}
for _, name := range files {
    processFile(name)
}</code></pre>`,

				Quiz: []Q{
					{
						Question:    "Что выведет замыкание в цикле: for i := 0; i < 3; i++ { funcs = append(funcs, func() { fmt.Println(i) }) }?",
						Options:     []string{"0, 1, 2", "3, 3, 3 — замыкание захватывает переменную, а не значение", "0, 0, 0", "Ошибка компиляции"},
						Correct:     1,
						Explanation: "Все замыкания ссылаются на одну и ту же переменную i. К моменту вызова i == 3 (после цикла). Исправление: i := i внутри цикла.",
					},
					{
						Question:    "В каком порядке выполняются defer?",
						Options:     []string{"В порядке добавления (FIFO)", "В обратном порядке (LIFO) — последний defer выполняется первым", "Случайно", "Параллельно"},
						Correct:     1,
						Explanation: "defer работает как стек: Last In, First Out. Это логично для cleanup: открыл A, потом B → закрой сначала B, потом A.",
					},
					{
						Question:    "Что делает go build -gcflags=\"-m\"?",
						Options:     []string{"Ускоряет сборку", "Показывает escape analysis — какие переменные уходят на кучу", "Включает дебаг", "Форматирует код"},
						Correct:     1,
						Explanation: "Escape analysis показывает, где компилятор решил хранить переменную: стек (быстро) или куча (нужен GC). Помогает оптимизировать код.",
					},
					{
						Question:    "Почему defer в цикле опасен?",
						Options:     []string{"Он не работает в циклах", "Все defer накапливаются и вызываются только при выходе из функции — ресурсы не освобождаются между итерациями", "Цикл с defer — ошибка компиляции", "defer в цикле выполняется быстрее"},
						Correct:     1,
						Explanation: "Если открываешь 1000 файлов в цикле с defer f.Close() — все 1000 закроются только при выходе из функции. Исправление: вынести в отдельную функцию, где defer сработает после каждой итерации.",
					},
					{
						Question:    "Что такое паттерн Functional Options?",
						Options:     []string{"Необязательные параметры через указатели", "Функции-опции типа func(*Config) для гибкой настройки структур без сотен перегрузок", "Глобальные переменные конфигурации", "Конфиг через environment variables"},
						Correct:     1,
						Explanation: "WithPort(3000), WithTimeout(5s) — каждая опция это функция func(*Config). NewServer(opts ...Option) применяет их по одной. Используется в chi, pgx, slog, grpc-go. Позволяет добавлять опции без изменения сигнатуры функции.",
					},
				},
				Tasks: []T{
					{
						Title:      "Порядок defer (LIFO)",
						Difficulty: "easy",
						Glossary: []GlossaryItem{
							{Term: "defer f()", Definition: "Откладывает вызов f() до выхода из текущей функции. Множественные defer выполняются в обратном порядке (LIFO)."},
						},
						Description: `<p>На вход число N. Создай N defer-вызовов <code>fmt.Println(i)</code> в цикле от 1 до N. В конце выведи "done".</p>
<p>Ввод: <code>4</code></p>
<p>Вывод:</p>
<pre><code>done
4
3
2
1</code></pre>`,
						Hints: `<p>Цикл <code>for i := 1; i <= n; i++ { defer fmt.Println(i) }</code>. Аргументы defer вычисляются сразу.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)
	for i := 1; i <= n; i++ {
		defer fmt.Println(i)
	}
	fmt.Println("done")
}</code></pre>`,
						StarterCode: `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	// TODO: создай defer fmt.Println(i) для каждого i от 1 до n
	// Помни: defer выполняется в обратном порядке (LIFO)

	fmt.Println("done")
}`,
						TestCases: []TestCase{
							{Input: "4", ExpectedOutput: "done\n4\n3\n2\n1"},
							{Input: "2", ExpectedOutput: "done\n2\n1"},
						},
					},
					{
						Title:      "Ловушка замыкания в цикле",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "i := i", Definition: "Shadowing — создаёт новую локальную переменную с тем же именем. Каждая итерация цикла получает свою копию."},
							{Term: "func() { capture }()", Definition: "Немедленно вызываемое замыкание. Без () в конце — просто создаёт функцию, не вызывает."},
						},
						Description: `<p>На вход число N. Создай N функций в цикле, каждая должна запоминать своё значение i (от 0 до N-1). Потом вызови их все.</p>
<p>Ввод: <code>4</code></p>
<p>Вывод:</p>
<pre><code>0
1
2
3</code></pre>
<p><strong>Внимание:</strong> без правильного захвата переменной все функции выведут одно и то же число!</p>`,
						Hints: `<p>Внутри цикла добавь <code>i := i</code> перед созданием функции. Это создаёт копию переменной для каждого замыкания.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	funcs := make([]func(), n)
	for i := 0; i < n; i++ {
		i := i // копия! без этого все выведут n
		funcs[i] = func() {
			fmt.Println(i)
		}
	}
	for _, f := range funcs {
		f()
	}
}</code></pre>`,
						StarterCode: `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	funcs := make([]func(), n)
	for i := 0; i < n; i++ {
		// TODO: исправь ловушку замыкания!
		// Без исправления все функции выведут n (последнее значение i)
		// Подсказка: i := i создаёт локальную копию
		funcs[i] = func() {
			fmt.Println(i)
		}
	}
	for _, f := range funcs {
		f()
	}
}`,
						TestCases: []TestCase{
							{Input: "4", ExpectedOutput: "0\n1\n2\n3"},
							{Input: "3", ExpectedOutput: "0\n1\n2"},
						},
					},
					{
						Title:      "Middleware-обёртка",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>withLogging(name string, fn func()) func()</code>, которая оборачивает fn и выводит сообщения до и после выполнения:</p>
<p>Вывод:</p>
<pre><code>START: processData
working...
END: processData</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "обёртка (wrapper)", Definition: "Функция которая вызывает другую функцию, добавляя поведение до и/или после. Основа middleware-паттерна."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "START: processData\nworking...\nEND: processData"},
						},
						StarterCode: `package main

import "fmt"

func withLogging(name string, fn func()) func() {
    return func() {
        fmt.Printf("START: %s\n", name)
        fn()
        fmt.Printf("END: %s\n", name)
    }
}

func main() {
    work := func() {
        fmt.Println("working...")
    }

    logged := withLogging("processData", work)
    logged()
}`,
						Hints:    `<p>Возвращаемая функция вызывает <code>fn()</code> между двумя fmt.Printf. Код уже написан — изучи и запусти.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func withLogging(name string, fn func()) func() {
    return func() {
        fmt.Printf("START: %s\n", name)
        fn()
        fmt.Printf("END: %s\n", name)
    }
}

func main() {
    work := func() {
        fmt.Println("working...")
    }

    logged := withLogging("processData", work)
    logged()
}</code></pre>`,
					},
					{
						Title:      "Functional Options",
						Difficulty: "medium",
						Description: `<p>Реализуй паттерн Functional Options для конфигурации сервера:</p>
<pre><code>Port: 3000
Host: localhost
Verbose: true</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "type Option func(*Config)", Definition: "Тип Option — функция которая модифицирует Config. Передаётся как аргумент в конструктор."},
							{Term: "opts ...Option", Definition: "Вариативный параметр типа Option. Позволяет передать любое количество опций в конструктор."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "Port: 3000\nHost: localhost\nVerbose: true"},
						},
						StarterCode: `package main

import "fmt"

type Config struct {
    Port    int
    Host    string
    Verbose bool
}

type Option func(*Config)

func WithPort(p int) Option {
    return func(c *Config) { c.Port = p }
}

func WithHost(h string) Option {
    return func(c *Config) { c.Host = h }
}

func WithVerbose() Option {
    return func(c *Config) { c.Verbose = true }
}

func NewConfig(opts ...Option) *Config {
    cfg := &Config{Port: 8080, Host: "0.0.0.0"}
    for _, opt := range opts {
        opt(cfg)
    }
    return cfg
}

func main() {
    cfg := NewConfig(
        WithPort(3000),
        WithHost("localhost"),
        WithVerbose(),
    )
    fmt.Printf("Port: %d\n", cfg.Port)
    fmt.Printf("Host: %s\n", cfg.Host)
    fmt.Printf("Verbose: %v\n", cfg.Verbose)
}`,
						Hints:    `<p>Код полностью написан — изучи структуру. Каждая WithX функция возвращает func(*Config) которая меняет одно поле. NewConfig применяет их все.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type Config struct {
    Port    int
    Host    string
    Verbose bool
}

type Option func(*Config)

func WithPort(p int) Option    { return func(c *Config) { c.Port = p } }
func WithHost(h string) Option { return func(c *Config) { c.Host = h } }
func WithVerbose() Option      { return func(c *Config) { c.Verbose = true } }

func NewConfig(opts ...Option) *Config {
    cfg := &Config{Port: 8080, Host: "0.0.0.0"}
    for _, opt := range opts {
        opt(cfg)
    }
    return cfg
}

func main() {
    cfg := NewConfig(WithPort(3000), WithHost("localhost"), WithVerbose())
    fmt.Printf("Port: %d\n", cfg.Port)
    fmt.Printf("Host: %s\n", cfg.Host)
    fmt.Printf("Verbose: %v\n", cfg.Verbose)
}</code></pre>`,
					},
					{
						Title:      "Конвейер функций (pipeline)",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "...func(int) int", Definition: "Вариативный параметр: принимает любое количество функций-трансформеров."},
							{Term: "pipeline", Definition: "Цепочка преобразований: результат одной функции → вход следующей. f3(f2(f1(x)))."},
						},
						Description: `<p>Реализуй функцию <code>pipeline(value int, fns ...func(int) int) int</code> — применяет функции последовательно:</p>
<pre><code>pipeline(5, double, addTen, negate)
// double(5)=10 → addTen(10)=20 → negate(20)=-20</code></pre>
<p>Ввод: <code>5</code></p>
<p>Вывод: <code>-20</code></p>`,
						Hints: `<p>Цикл по fns: <code>for _, fn := range fns { value = fn(value) }</code>. Функция просто прогоняет значение через каждую функцию.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func pipeline(value int, fns ...func(int) int) int {
	for _, fn := range fns {
		value = fn(value)
	}
	return value
}

func main() {
	var n int
	fmt.Scan(&n)

	double := func(x int) int { return x * 2 }
	addTen := func(x int) int { return x + 10 }
	negate := func(x int) int { return -x }

	result := pipeline(n, double, addTen, negate)
	fmt.Println(result)
}</code></pre>`,
						StarterCode: `package main

import "fmt"

func pipeline(value int, fns ...func(int) int) int {
	// TODO: применяй функции последовательно
	// for _, fn := range fns { value = fn(value) }
	return value
}

func main() {
	var n int
	fmt.Scan(&n)

	double := func(x int) int { return x * 2 }
	addTen := func(x int) int { return x + 10 }
	negate := func(x int) int { return -x }

	result := pipeline(n, double, addTen, negate)
	fmt.Println(result)
}`,
						TestCases: []TestCase{
							{Input: "5", ExpectedOutput: "-20"},
							{Input: "0", ExpectedOutput: "-10"},
							{Input: "100", ExpectedOutput: "-210"},
						},
					},
				},
			},
		},
	}
}
