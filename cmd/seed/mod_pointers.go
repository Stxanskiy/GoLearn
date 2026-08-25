package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Указатели и память
// Вставляется между Функциями и Структурами
// ════════════════════════════════════════════════════════════════

func mod_pointers() M {
	return M{
		Slug:          "pointers",
		Title:         "Указатели и память",
		Description:   "Как Go хранит данные в памяти. Указатели, & и *, стек vs куча, передача по значению vs по ссылке.",
		Order:         4, // между функциями (3) и структурами (5)
		Track:         "shared",
		Difficulty:    "beginner",
		Prerequisites: []string{"functions"},
		Lessons: []L{
			{
				Slug: "what-is-pointer", Title: "Что такое указатель", Order: 1,
				Difficulty: "beginner", Track: "shared",
				Content: `<h1>Что такое указатель</h1>

<h2>Аналогия: адрес дома</h2>
<p>Представь: у тебя есть друг Вася. Ты можешь либо <strong>позвать его к себе</strong> (скопировать), либо <strong>записать его адрес</strong> (указатель). Если ты изменишь копию Васи — оригинальный Вася не узнает. Если пойдёшь по адресу и изменишь что-то там — изменения останутся.</p>

<p><strong>Указатель</strong> — это переменная, которая хранит <strong>адрес</strong> другой переменной в памяти, а не само значение.</p>

<h2>Зачем нужны указатели?</h2>
<ol>
<li><strong>Изменять данные в функции</strong> — без указателя функция получает копию и не может изменить оригинал</li>
<li><strong>Экономить память</strong> — вместо копирования большой структуры передаём маленький адрес</li>
<li><strong>Разделять данные</strong> — несколько переменных могут ссылаться на одни и те же данные</li>
</ol>

<h2>Операторы & и *</h2>
<pre><code>x := 42

p := &x    // & — взять адрес x. p теперь указатель на x
fmt.Println(p)   // 0xc0000b4008 (адрес в памяти)
fmt.Println(*p)  // 42 — * разыменование: получить значение по адресу

*p = 100   // изменить значение по адресу
fmt.Println(x)   // 100 — x изменился!</code></pre>

<p>Два оператора:</p>
<ul>
<li><code>&x</code> — "дай мне адрес x" (взятие адреса)</li>
<li><code>*p</code> — "дай мне значение по адресу p" (разыменование)</li>
</ul>

<h2>Тип указателя</h2>
<pre><code>var x int = 42
var p *int = &x   // *int — "указатель на int"

// Нулевой указатель
var q *int        // nil — не указывает никуда
fmt.Println(q)    // &lt;nil&gt;
// *q             // PANIC! Нельзя разыменовать nil</code></pre>

<p><code>*int</code> читается как "указатель на int". <code>*string</code> — "указатель на string". И так далее.</p>

<h2>Передача по значению vs по указателю</h2>
<pre><code>// По значению — функция получает КОПИЮ
func double(n int) {
    n = n * 2   // изменяем копию
}
x := 5
double(x)
fmt.Println(x)  // 5 — не изменился!

// По указателю — функция получает АДРЕС
func doublePtr(n *int) {
    *n = *n * 2   // изменяем по адресу
}
y := 5
doublePtr(&y)
fmt.Println(y)  // 10 — изменился!</code></pre>

<h2>new() — создание указателя</h2>
<pre><code>// new(T) создаёт переменную типа T и возвращает указатель на неё
p := new(int)     // *int, значение = 0
*p = 42
fmt.Println(*p)   // 42

// Эквивалентно:
x := 0
p := &x</code></pre>

<h2>Стек vs Куча (stack vs heap)</h2>
<p>Go автоматически решает где хранить данные:</p>
<ul>
<li><strong>Стек</strong> — быстро, автоматически очищается при выходе из функции</li>
<li><strong>Куча</strong> — медленнее, очищается сборщиком мусора (GC)</li>
</ul>
<pre><code>func createUser() *User {
    u := User{Name: "Alice"}  // Go видит что &u уходит из функции
    return &u                  // → помещает в кучу (heap escape)
}
// u не исчезнет после return — Go позаботится</code></pre>
<p>Тебе не нужно вручную управлять памятью — Go делает это сам. Но понимание помогает писать быстрый код.</p>

<h2>Типичные ошибки с указателями</h2>
<p>Источник: опыт Go-разработчиков с Хабра и реальных код-ревью</p>

<pre><code>// ОШИБКА 1: возврат указателя на локальную переменную внутри цикла
ptrs := []*int{}
for i := 0; i < 3; i++ {
    ptrs = append(ptrs, &i) // все указывают на ОДНУ переменную i!
}
fmt.Println(*ptrs[0], *ptrs[1], *ptrs[2]) // 3 3 3

// ИСПРАВЛЕНИЕ: создать копию
for i := 0; i < 3; i++ {
    v := i
    ptrs = append(ptrs, &v) // каждый раз новая переменная
}

// ОШИБКА 2: разыменование без проверки на nil
var p *int
fmt.Println(*p)  // panic! Всегда проверяй if p != nil

// ОШИБКА 3: думать что изменение поля struct через value receiver сохранится
type Counter struct{ n int }
func (c Counter) Inc() { c.n++ }  // копия! Оригинал не меняется
// Правильно:
func (c *Counter) Inc() { c.n++ } // указатель — меняется оригинал</code></pre>

<h2>Читать глубже (русскоязычные источники)</h2>
<ul>
<li><a href="https://metanit.com/go/golang/5.1.php" target="_blank">Metanit: Указатели в Go</a></li>
<li><a href="https://habr.com/ru/articles/460325/" target="_blank">Хабр: Указатели в Go — всё что нужно знать</a></li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Что хранит указатель?",
						Options:     []string{"Значение переменной", "Адрес переменной в памяти", "Тип переменной", "Имя переменной"},
						Correct:     1,
						Explanation: "Указатель хранит адрес — место в памяти, где находится значение. Через этот адрес можно прочитать или изменить оригинальное значение.",
					},
					{
						Question:    "Что делает оператор &?",
						Options:     []string{"Логическое И", "Берёт адрес переменной (создаёт указатель)", "Умножает на 2", "Сравнивает значения"},
						Correct:     1,
						Explanation: "&x возвращает адрес переменной x. Результат — указатель (*T), который можно передать в функцию или сохранить.",
					},
					{
						Question:    "Что произойдёт при *nil?",
						Options:     []string{"Вернёт 0", "Паника (panic) — нельзя разыменовать nil-указатель", "Ошибка компиляции", "Вернёт пустую строку"},
						Correct:     1,
						Explanation: "Разыменование nil-указателя = panic. Всегда проверяй указатель на nil перед использованием: if p != nil { *p }",
					},
					{
						Question:    "Зачем передавать указатель в функцию?",
						Options:     []string{"Для красоты", "Чтобы функция могла изменить оригинальные данные, а не копию", "Указатели быстрее чисел", "Go требует"},
						Correct:     1,
						Explanation: "Без указателя функция получает копию — изменения не видны снаружи. С указателем функция работает с оригиналом через адрес.",
					},
					{
						Question:    "Что вернёт new(int)?",
						Options:     []string{"0", "Указатель на int со значением 0 (*int)", "nil", "Ошибку"},
						Correct:     1,
						Explanation: "new(T) выделяет память для T, инициализирует нулевым значением и возвращает указатель *T. new(int) == &0. На практике чаще используют x := 0; p := &x.",
					},
				},
				Tasks: []T{
					{
						Title:      "Обмен через указатели",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>swap(a, b *int)</code>, которая меняет значения двух переменных местами через указатели.</p>
<p>Ввод: <code>10 20</code></p>
<p>Вывод:</p>
<pre><code>Before: a=10, b=20
After: a=20, b=10</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "*int", Definition: "Тип «указатель на int». Переменная этого типа хранит адрес какого-то int."},
							{Term: "&x", Definition: "Оператор взятия адреса. Возвращает указатель на переменную x."},
							{Term: "*p", Definition: "Оператор разыменования. Возвращает (или изменяет) значение, на которое указывает p."},
						},
						TestCases: []TestCase{
							{Input: "10 20", ExpectedOutput: "Before: a=10, b=20\nAfter: a=20, b=10"},
							{Input: "1 2", ExpectedOutput: "Before: a=1, b=2\nAfter: a=2, b=1"},
						},
						StarterCode: `package main

import "fmt"

func swap(a, b *int) {
    // Поменяй значения через указатели
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)

    fmt.Printf("Before: a=%d, b=%d\n", a, b)
    swap(&a, &b)
    fmt.Printf("After: a=%d, b=%d\n", a, b)
}`,
						Hints: `<p><code>*a, *b = *b, *a</code> — множественное присваивание работает и через указатели.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func swap(a, b *int) {
    *a, *b = *b, *a
}

func main() {
    var a, b int
    fmt.Scan(&a, &b)

    fmt.Printf("Before: a=%d, b=%d\n", a, b)
    swap(&a, &b)
    fmt.Printf("After: a=%d, b=%d\n", a, b)
}</code></pre>`,
					},
					{
						Title:      "Функция-модификатор",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>clamp(val *int, min, max int)</code>, которая ограничивает значение в диапазоне [min, max]:</p>
<ul>
<li>Если *val &lt; min → установить в min</li>
<li>Если *val &gt; max → установить в max</li>
<li>Иначе — не менять</li>
</ul>
<p>Ввод: <code>150 0 100</code></p>
<p>Вывод: <code>100</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "func f(p *int)", Definition: "Функция принимает указатель на int. Внутри можно читать (*p) и изменять (*p = x) оригинал."},
						},
						TestCases: []TestCase{
							{Input: "150 0 100", ExpectedOutput: "100"},
							{Input: "-5 0 100", ExpectedOutput: "0"},
							{Input: "50 0 100", ExpectedOutput: "50"},
						},
						StarterCode: `package main

import "fmt"

func clamp(val *int, min, max int) {
    // Ограничь *val в диапазоне [min, max]
}

func main() {
    var val, min, max int
    fmt.Scan(&val, &min, &max)
    clamp(&val, min, max)
    fmt.Println(val)
}`,
						Hints: `<p><code>if *val < min { *val = min } else if *val > max { *val = max }</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func clamp(val *int, min, max int) {
    if *val < min {
        *val = min
    } else if *val > max {
        *val = max
    }
}

func main() {
    var val, min, max int
    fmt.Scan(&val, &min, &max)
    clamp(&val, min, max)
    fmt.Println(val)
}</code></pre>`,
					},
					{
						Title:      "Счётчик через указатель",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>increment(n *int)</code> которая увеличивает значение по указателю на 1. Вызови её 5 раз:</p>
<p>Вывод: <code>5</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "*n++", Definition: "Разыменовать n и увеличить на 1."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "5"},
						},
						StarterCode: `package main

import "fmt"

func increment(n *int) {
    // Увеличь значение по указателю на 1
}

func main() {
    count := 0
    increment(&count)
    increment(&count)
    increment(&count)
    increment(&count)
    increment(&count)
    fmt.Println(count)
}`,
						Hints: `<p><code>*n++</code> или <code>*n = *n + 1</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func increment(n *int) {
    *n++
}

func main() {
    count := 0
    for i := 0; i < 5; i++ {
        increment(&count)
    }
    fmt.Println(count)
}</code></pre>`,
					},
					{
						Title:      "Безопасное разыменование",
						Difficulty: "medium",
						Description: `<p>Напиши функцию <code>safeDeref(p *int, defaultVal int) int</code> — возвращает значение или defaultVal если nil:</p>
<p>Ввод: <code>42</code></p>
<p>Вывод:</p>
<pre><code>Value: 42
Default: 0</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "nil check", Definition: "if p == nil { return defaultVal } — стандартный паттерн безопасного разыменования."},
						},
						TestCases: []TestCase{
							{Input: "42", ExpectedOutput: "Value: 42\nDefault: 0"},
						},
						StarterCode: `package main

import "fmt"

func safeDeref(p *int, defaultVal int) int {
    if p == nil {
        return defaultVal
    }
    return *p
}

func main() {
    var n int
    fmt.Scan(&n)
    fmt.Printf("Value: %d\n", safeDeref(&n, 0))
    fmt.Printf("Default: %d\n", safeDeref(nil, 0))
}`,
						Hints: `<p>Проверь <code>if p == nil</code> перед разыменованием.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func safeDeref(p *int, defaultVal int) int {
    if p == nil {
        return defaultVal
    }
    return *p
}

func main() {
    var n int
    fmt.Scan(&n)
    fmt.Printf("Value: %d\n", safeDeref(&n, 0))
    fmt.Printf("Default: %d\n", safeDeref(nil, 0))
}</code></pre>`,
					},
					{
						Title:      "Стек на указателях",
						Difficulty: "hard",
						Description: `<p>Реализуй стек (LIFO) на связанном списке. Команды: <code>push N</code> и <code>pop</code>:</p>
<p>Ввод:</p>
<pre><code>4
push 1
push 2
push 3
pop</code></pre>
<p>Вывод:</p>
<pre><code>pushed 1
pushed 2
pushed 3
popped 3</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "**Node", Definition: "Указатель на указатель. Позволяет функции изменять сам указатель top (переставлять вершину стека)."},
						},
						TestCases: []TestCase{
							{Input: "4\npush 1\npush 2\npush 3\npop", ExpectedOutput: "pushed 1\npushed 2\npushed 3\npopped 3"},
						},
						StarterCode: `package main

import "fmt"

type Node struct {
    val  int
    next *Node
}

func push(top **Node, val int) {
    *top = &Node{val: val, next: *top}
    fmt.Printf("pushed %d\n", val)
}

func pop(top **Node) {
    if *top == nil { fmt.Println("empty"); return }
    fmt.Printf("popped %d\n", (*top).val)
    *top = (*top).next
}

func main() {
    var n int
    fmt.Scan(&n)
    var top *Node
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        if cmd == "push" {
            var val int
            fmt.Scan(&val)
            push(&top, val)
        } else {
            pop(&top)
        }
    }
}`,
						Hints: `<p>push: <code>*top = &Node{val: val, next: *top}</code>. pop: <code>*top = (*top).next</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

type Node struct {
    val  int
    next *Node
}

func push(top **Node, val int) {
    *top = &Node{val: val, next: *top}
    fmt.Printf("pushed %d\n", val)
}

func pop(top **Node) {
    if *top == nil { fmt.Println("empty"); return }
    fmt.Printf("popped %d\n", (*top).val)
    *top = (*top).next
}

func main() {
    var n int
    fmt.Scan(&n)
    var top *Node
    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)
        if cmd == "push" {
            var val int
            fmt.Scan(&val)
            push(&top, val)
        } else {
            pop(&top)
        }
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug: "pointers-slices-maps", Title: "Указатели под капотом: слайсы и map", Order: 2,
				Difficulty: "intermediate", Track: "shared",
				Content: `<h1>Указатели под капотом: слайсы и map</h1>

<h2>Почему слайс ведёт себя как ссылка?</h2>
<p>Слайс — это структура из трёх полей:</p>
<pre><code>type slice struct {
    ptr *array   // указатель на массив в памяти
    len int      // текущая длина
    cap int      // ёмкость
}

// Когда ты передаёшь слайс в функцию, копируется структура (ptr, len, cap)
// НО ptr указывает на тот же массив → изменения элементов видны!
func modify(s []int) {
    s[0] = 999  // ИЗМЕНЯЕТ оригинал — тот же массив
}

// Но append может НЕ изменить оригинал!
func addElem(s []int) {
    s = append(s, 42)  // s — копия заголовка, оригинал не знает о новом элементе!
}

nums := []int{1, 2, 3}
modify(nums)
fmt.Println(nums[0])  // 999 — изменился!

addElem(nums)
fmt.Println(len(nums))  // 3 — НЕ изменился!</code></pre>

<h2>Когда передавать *[]int?</h2>
<pre><code>// Если функция должна ИЗМЕНИТЬ сам слайс (добавить/удалить элементы):
func addElem(s *[]int, val int) {
    *s = append(*s, val)
}

// Или лучше — возвращать новый слайс:
func addElem(s []int, val int) []int {
    return append(s, val)
}
nums = addElem(nums, 42)  // идиоматичный Go</code></pre>

<h2>Map — всегда ссылка</h2>
<pre><code>// Map внутри — это указатель на hash-таблицу
// Передача в функцию НЕ копирует данные
func addUser(m map[string]int, name string, age int) {
    m[name] = age  // ИЗМЕНЯЕТ оригинал!
}

users := map[string]int{"Alice": 25}
addUser(users, "Bob", 30)
fmt.Println(users)  // map[Alice:25 Bob:30]

// Но var m map[string]int = nil!
// Запись в nil map → PANIC!
var m map[string]int
m["key"] = 1  // panic: assignment to nil map
// Всегда: m := make(map[string]int)</code></pre>

<h2>Struct: значение vs указатель</h2>
<pre><code>type User struct {
    Name string
    Age  int
}

// Передача по значению — копия
func birthday(u User) {
    u.Age++  // копия! Оригинал не изменится
}

// Передача по указателю — оригинал
func birthdayPtr(u *User) {
    u.Age++  // Go автоматически разыменовывает: (*u).Age++
}

alice := User{Name: "Alice", Age: 25}
birthday(alice)
fmt.Println(alice.Age)  // 25 — не изменился

birthdayPtr(&alice)
fmt.Println(alice.Age)  // 26 — изменился</code></pre>

<h2>Подслайс разделяет память</h2>
<p>Это важная ловушка — из Хабра про внутренности слайса:</p>
<pre><code>original := []int{1, 2, 3, 4, 5}
sub := original[1:4]   // [2, 3, 4]

// sub и original смотрят на ОДИН и тот же массив!
sub[0] = 99
fmt.Println(original)  // [1 99 3 4 5] — изменился!

// Как сделать независимую копию:
independent := make([]int, len(sub))
copy(independent, sub)
independent[0] = 0
fmt.Println(original)  // [1 99 3 4 5] — не изменился</code></pre>

<h2>Escape analysis — куда уходят переменные</h2>
<p>Go-компилятор анализирует, "убегает" ли переменная за пределы функции (heap escape). Посмотреть можно через:</p>
<pre><code>go build -gcflags="-m" ./...
// ./main.go:5:2: moved to heap: x
// Значит x попала в кучу (GC будет за ней следить)</code></pre>
<p>Это полезно для оптимизации: если хочешь избежать GC-давления, не возвращай указатели на локальные переменные без нужды.</p>

<h2>Правила: когда указатель, когда значение</h2>
<ul>
<li><strong>Маленькие struct (&lt; 4 полей)</strong> → передавай по значению (быстрее, безопаснее)</li>
<li><strong>Большие struct</strong> → указатель (избегаем копирования)</li>
<li><strong>Нужно изменять</strong> → обязательно указатель</li>
<li><strong>Методы</strong> → если хотя бы один метод с *T, делай все с *T</li>
<li><strong>Map, slice, channel</strong> → уже содержат указатели внутри, передавай по значению</li>
</ul>

<h2>Читать глубже</h2>
<ul>
<li><a href="https://habr.com/ru/articles/477348/" target="_blank">Хабр: Внутренности слайсов Go</a> — как устроен slice header изнутри</li>
<li><a href="https://metanit.com/go/golang/3.11.php" target="_blank">Metanit: Слайсы и память</a> — подробно про copy и subslice</li>
</ul>`,

				Quiz: []Q{
					{
						Question:    "Почему изменение элемента слайса в функции видно снаружи, а append — нет?",
						Options:     []string{"Баг Go", "Слайс-заголовок копируется, но ptr указывает на тот же массив. append может создать новый массив, о котором копия заголовка не знает", "Всегда видно", "Никогда не видно"},
						Correct:     1,
						Explanation: "s[0] = x изменяет элемент через общий указатель. append может реаллоцировать массив — новый ptr записывается только в копию заголовка, оригинал не обновляется.",
					},
					{
						Question:    "Что произойдёт при записи в nil map?",
						Options:     []string{"Создастся новый map", "panic: assignment to entry in nil map", "Ничего", "Ошибка компиляции"},
						Correct:     1,
						Explanation: "nil map можно читать (возвращает zero value), но нельзя писать. Всегда инициализируй: m := make(map[K]V) или m := map[K]V{}.",
					},
					{
						Question:    "Что произойдёт: original := []int{1,2,3,4,5}; sub := original[1:4]; sub[0] = 99",
						Options:     []string{"original не изменится — sub независимая копия", "original[1] станет 99 — sub разделяет тот же массив", "Ошибка компиляции", "panic out of range"},
						Correct:     1,
						Explanation: "Подслайс original[1:4] — новый slice-заголовок (len=3, cap=4), но ptr указывает на тот же backing array. Изменение элемента через sub меняет original. Независимую копию: make + copy.",
					},
					{
						Question:    "Когда нужно передавать *[]int вместо []int?",
						Options:     []string{"Всегда — так безопаснее", "Когда функция только читает элементы", "Когда функция делает append и нужно видеть результат снаружи", "Никогда — лучше возвращать []int"},
						Correct:     2,
						Explanation: "s[i] = x видно через обычный []int. Но append может реаллоцировать массив — копия заголовка снаружи этого не увидит. Нужен *[]int или возврат нового слайса. Идиоматично: return append(s, val).",
					},
					{
						Question:    "Передаётся ли map по ссылке в Go?",
						Options:     []string{"Да — map это указатель на runtime.hmap", "Нет — map копируется полностью", "Только с оператором &m", "Зависит от размера map"},
						Correct:     0,
						Explanation: "Под капотом map — это указатель на runtime.hmap. При передаче в функцию копируется только указатель. Функция может добавлять/удалять ключи и это видно снаружи. Но сам указатель — value, как и slice-заголовок.",
					},
					{
						Question:    "Как сделать настоящую независимую копию слайса s []int?",
						Options:     []string{"copy2 := s", "copy2 := s[:]", "copy2 := make([]int, len(s)); copy(copy2, s)", "copy2 := &s"},
						Correct:     2,
						Explanation: "copy2 := s копирует только заголовок (ptr, len, cap) — оба слайса смотрят на тот же массив. Настоящая копия: выделить новый массив через make, скопировать данные через copy(). Тогда изменения в copy2 не затронут s.",
					},
				},
				Tasks: []T{
					{
						Title:      "Функция append через указатель",
						Difficulty: "medium",
						Description: `<p>Напиши две функции:</p>
<ul>
<li><code>appendBroken(s []int, val int)</code> — попытка добавить через значение (не работает)</li>
<li><code>appendFixed(s *[]int, val int)</code> — добавление через указатель (работает)</li>
</ul>
<p>Ввод: <code>3 1 2 3 4</code> (3 элемента, потом добавить 4)</p>
<p>Вывод:</p>
<pre><code>After broken: [1 2 3]
After fixed: [1 2 3 4]</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "*[]int", Definition: "Указатель на слайс. Нужен если функция должна изменить сам слайс (append), а не только его элементы."},
							{Term: "*s = append(*s, val)", Definition: "Разыменовать указатель на слайс, добавить элемент, записать обратно."},
						},
						TestCases: []TestCase{
							{Input: "3 1 2 3 4", ExpectedOutput: "After broken: [1 2 3]\nAfter fixed: [1 2 3 4]"},
						},
						StarterCode: `package main

import "fmt"

func appendBroken(s []int, val int) {
    s = append(s, val) // s — копия заголовка
}

func appendFixed(s *[]int, val int) {
    // Добавь через указатель
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums { fmt.Scan(&nums[i]) }
    var extra int
    fmt.Scan(&extra)

    test1 := make([]int, len(nums))
    copy(test1, nums)
    appendBroken(test1, extra)
    fmt.Println("After broken:", test1)

    test2 := make([]int, len(nums))
    copy(test2, nums)
    appendFixed(&test2, extra)
    fmt.Println("After fixed:", test2)
}`,
						Hints: `<p><code>*s = append(*s, val)</code> — разыменовать, добавить, записать обратно.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func appendBroken(s []int, val int) {
    s = append(s, val)
}

func appendFixed(s *[]int, val int) {
    *s = append(*s, val)
}

func main() {
    var n int
    fmt.Scan(&n)
    nums := make([]int, n)
    for i := range nums { fmt.Scan(&nums[i]) }
    var extra int
    fmt.Scan(&extra)

    test1 := make([]int, len(nums))
    copy(test1, nums)
    appendBroken(test1, extra)
    fmt.Println("After broken:", test1)

    test2 := make([]int, len(nums))
    copy(test2, nums)
    appendFixed(&test2, extra)
    fmt.Println("After fixed:", test2)
}</code></pre>`,
					},
					{
						Title:      "Счётчик слов через map",
						Difficulty: "easy",
						Description: `<p>Напиши функцию <code>countWords(text string, counts map[string]int)</code>, которая подсчитывает слова в строке и записывает результат в переданный map. Убедись, что изменения видны снаружи.</p>
<p>Ввод: <code>go is go and go is great</code></p>
<p>Вывод (порядок строк фиксирован):</p>
<pre><code>and: 1
go: 3
great: 1
is: 2</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "map as reference", Definition: "Map — это указатель на runtime.hmap. Передача в функцию копирует только указатель. Изменения внутри функции видны снаружи."},
							{Term: "strings.Fields", Definition: "Разбивает строку по пробелам, возвращает []string. Удобно для разбора слов."},
						},
						TestCases: []TestCase{
							{Input: "go is go and go is great", ExpectedOutput: "and: 1\ngo: 3\ngreat: 1\nis: 2"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strings"
)

func countWords(text string, counts map[string]int) {
    // Разбей на слова и считай через counts[word]++
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    counts := make(map[string]int)
    countWords(line, counts)

    keys := make([]string, 0, len(counts))
    for k := range counts { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, counts[k])
    }
}`,
						Hints: `<p><code>strings.Fields(text)</code> разбивает на слова. В цикле: <code>counts[word]++</code></p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "sort"
    "strings"
)

func countWords(text string, counts map[string]int) {
    for _, word := range strings.Fields(text) {
        counts[word]++
    }
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    counts := make(map[string]int)
    countWords(line, counts)

    keys := make([]string, 0, len(counts))
    for k := range counts { keys = append(keys, k) }
    sort.Strings(keys)
    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, counts[k])
    }
}</code></pre>`,
					},
					{
						Title:      "Подслайс vs копия",
						Difficulty: "easy",
						Description: `<p>Покажи разницу между подслайсом и независимой копией:</p>
<ol>
<li>Создай <code>original := []int{1, 2, 3, 4, 5}</code></li>
<li>Создай <code>sub := original[1:4]</code></li>
<li>Создай <code>independent</code> — настоящую копию через make + copy</li>
<li>Измени <code>sub[0] = 99</code> и <code>independent[0] = 0</code></li>
<li>Выведи original</li>
</ol>
<p>Вывод:</p>
<pre><code>[1 99 3 4 5]</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "subslice", Definition: "original[low:high] — новый заголовок, но тот же backing array. Изменения в subslice видны в original."},
							{Term: "deep copy", Definition: "make([]int, len(s)) + copy(dst, src) — полностью независимый слайс с собственным массивом."},
						},
						TestCases: []TestCase{
							{Input: "", ExpectedOutput: "[1 99 3 4 5]"},
						},
						StarterCode: `package main

import "fmt"

func main() {
    original := []int{1, 2, 3, 4, 5}
    sub := original[1:4]

    // Создай independent как настоящую копию sub
    independent := // ...

    sub[0] = 99
    independent[0] = 0

    fmt.Println(original)
}`,
						Hints: `<p><code>independent := make([]int, len(sub)); copy(independent, sub)</code></p>`,
						Solution: `<pre><code>package main

import "fmt"

func main() {
    original := []int{1, 2, 3, 4, 5}
    sub := original[1:4]

    independent := make([]int, len(sub))
    copy(independent, sub)

    sub[0] = 99
    independent[0] = 0

    fmt.Println(original) // [1 99 3 4 5]
}</code></pre>`,
					},
					{
						Title:      "Struct через указатель",
						Difficulty: "medium",
						Description: `<p>Реализуй структуру <code>BankAccount</code> с полем <code>Balance float64</code> и методами с pointer receiver:</p>
<ul>
<li><code>Deposit(amount float64)</code> — пополнить</li>
<li><code>Withdraw(amount float64) bool</code> — снять, если достаточно средств (иначе false)</li>
</ul>
<p>Ввод:</p>
<pre><code>deposit 100.50
withdraw 30.00
withdraw 200.00</code></pre>
<p>Вывод:</p>
<pre><code>Balance: 100.50
Balance: 70.50
Insufficient funds</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "pointer receiver", Definition: "func (a *BankAccount) — метод получает указатель на struct. Изменения сохраняются. Если receiver — значение (не *), изменения теряются."},
						},
						TestCases: []TestCase{
							{Input: "deposit 100.50\nwithdraw 30.00\nwithdraw 200.00", ExpectedOutput: "Balance: 100.50\nBalance: 70.50\nInsufficient funds"},
						},
						StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type BankAccount struct {
    Balance float64
}

func (a *BankAccount) Deposit(amount float64) {
    // Прибавь к Balance
}

func (a *BankAccount) Withdraw(amount float64) bool {
    // Если хватает — снять и вернуть true. Иначе false.
    return false
}

func main() {
    acc := &BankAccount{}
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        parts := strings.Fields(scanner.Text())
        var amount float64
        fmt.Sscanf(parts[1], "%f", &amount)
        switch parts[0] {
        case "deposit":
            acc.Deposit(amount)
            fmt.Printf("Balance: %.2f\n", acc.Balance)
        case "withdraw":
            if acc.Withdraw(amount) {
                fmt.Printf("Balance: %.2f\n", acc.Balance)
            } else {
                fmt.Println("Insufficient funds")
            }
        }
    }
}`,
						Hints: `<p>Deposit: <code>a.Balance += amount</code>. Withdraw: проверь <code>a.Balance >= amount</code>, потом <code>a.Balance -= amount</code>.</p>`,
						Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

type BankAccount struct {
    Balance float64
}

func (a *BankAccount) Deposit(amount float64) {
    a.Balance += amount
}

func (a *BankAccount) Withdraw(amount float64) bool {
    if a.Balance < amount {
        return false
    }
    a.Balance -= amount
    return true
}

func main() {
    acc := &BankAccount{}
    scanner := bufio.NewScanner(os.Stdin)
    for scanner.Scan() {
        parts := strings.Fields(scanner.Text())
        var amount float64
        fmt.Sscanf(parts[1], "%f", &amount)
        switch parts[0] {
        case "deposit":
            acc.Deposit(amount)
            fmt.Printf("Balance: %.2f\n", acc.Balance)
        case "withdraw":
            if acc.Withdraw(amount) {
                fmt.Printf("Balance: %.2f\n", acc.Balance)
            } else {
                fmt.Println("Insufficient funds")
            }
        }
    }
}</code></pre>`,
					},
					{
						Title:      "Разворот слайса на месте",
						Difficulty: "hard",
						Description: `<p>Напиши функцию <code>reverse(s []int)</code> которая разворачивает слайс <strong>на месте</strong> (in-place) через указатели-индексы. Без создания нового слайса. Потом напиши <code>rotateLeft(s []int, k int)</code> — сдвиг влево на k позиций (тоже in-place, через три reversal).</p>
<p>Ввод: <code>5 1 2 3 4 5 2</code> (5 элементов, сдвиг на 2)</p>
<p>Вывод:</p>
<pre><code>Reversed: [5 4 3 2 1]
Rotated: [3 4 5 1 2]</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "in-place reverse", Definition: "Два указателя: left=0, right=len-1. Меняем s[left], s[right] местами, двигаем навстречу."},
							{Term: "three reversal rotation", Definition: "rotateLeft(s, k): reverse(s[:k]), reverse(s[k:]), reverse(s). Сдвиг за O(n) без доп. памяти."},
						},
						TestCases: []TestCase{
							{Input: "5 1 2 3 4 5 2", ExpectedOutput: "Reversed: [5 4 3 2 1]\nRotated: [3 4 5 1 2]"},
						},
						StarterCode: `package main

import "fmt"

func reverse(s []int) {
    // Два указателя: l и r
    l, r := 0, len(s)-1
    for l < r {
        // Поменяй местами s[l] и s[r]
        l++
        r--
    }
}

func rotateLeft(s []int, k int) {
    n := len(s)
    k = k % n
    // Три reversal: reverse первых k, остальных, потом весь слайс
}

func main() {
    var n int
    fmt.Scan(&n)
    s := make([]int, n)
    for i := range s { fmt.Scan(&s[i]) }
    var k int
    fmt.Scan(&k)

    original := make([]int, n)
    copy(original, s)

    reverse(s)
    fmt.Println("Reversed:", s)

    rotateLeft(original, k)
    fmt.Println("Rotated:", original)
}`,
						Hints: `<p>reverse: <code>s[l], s[r] = s[r], s[l]</code>. rotateLeft: <code>reverse(s[:k]); reverse(s[k:]); reverse(s)</code>.</p>`,
						Solution: `<pre><code>package main

import "fmt"

func reverse(s []int) {
    l, r := 0, len(s)-1
    for l < r {
        s[l], s[r] = s[r], s[l]
        l++
        r--
    }
}

func rotateLeft(s []int, k int) {
    n := len(s)
    k = k % n
    reverse(s[:k])
    reverse(s[k:])
    reverse(s)
}

func main() {
    var n int
    fmt.Scan(&n)
    s := make([]int, n)
    for i := range s { fmt.Scan(&s[i]) }
    var k int
    fmt.Scan(&k)

    original := make([]int, n)
    copy(original, s)

    reverse(s)
    fmt.Println("Reversed:", s)

    rotateLeft(original, k)
    fmt.Println("Rotated:", original)
}</code></pre>`,
					},
				},
			},
		},
	}
}
