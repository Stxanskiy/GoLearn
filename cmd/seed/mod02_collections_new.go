package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 2: Коллекции — массивы, слайсы, map
// Переработанный: для начинающих, с глоссарием и test cases
// ════════════════════════════════════════════════════════════════

func mod02_collections_new() M {
	return M{
		Slug:          "collections",
		Title:         "Коллекции данных",
		Description:   "Как хранить наборы данных: массивы, слайсы (динамические списки) и map (словари). Фундамент для любой программы.",
		Order:         2,
		Track:         "shared",
		Difficulty:    "beginner",
		Prerequisites: []string{"basics"},
		Lessons: []L{
			lesson_arrays_slices(),
			lesson_slice_operations(),
			lesson_strings_deep(),
			lesson_maps(),
		},
	}
}

func lesson_arrays_slices() L {
	return L{
		Slug: "arrays-slices", Title: "Массивы и слайсы — списки данных", Order: 1,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Массивы и слайсы — списки данных</h1>

<h2>Зачем нужны коллекции?</h2>
<p>До сих пор мы хранили данные в отдельных переменных: <code>name1</code>, <code>name2</code>, <code>name3</code>... Но что если пользователей 1000? Нужна структура, которая хранит <strong>набор значений</strong> одного типа.</p>

<h2>Массив — фиксированный размер</h2>
<p><strong>Массив</strong> — это набор элементов одного типа с <strong>фиксированным</strong> размером. Размер задаётся при создании и не меняется:</p>

<pre><code>// Создание массива из 3 строк
var names [3]string
names[0] = "Alice"
names[1] = "Bob"
names[2] = "Charlie"

// Сразу с значениями
scores := [4]int{90, 85, 78, 92}

// Go сам посчитает размер
colors := [...]string{"red", "green", "blue"}  // [3]string</code></pre>

<p><strong>Проблема массивов:</strong> размер является частью типа. <code>[3]int</code> и <code>[4]int</code> — это <strong>разные типы</strong>! Нельзя передать массив из 3 элементов в функцию, которая ожидает 4.</p>

<h2>Слайс — динамический список</h2>
<p><strong>Слайс</strong> (slice) — это динамический массив. Его размер может расти и уменьшаться. В 99% случаев в Go используют слайсы, а не массивы:</p>

<pre><code>// Создание слайса
names := []string{"Alice", "Bob", "Charlie"}  // без числа в скобках!
numbers := []int{10, 20, 30}

// Пустой слайс
var empty []int             // nil слайс (значение по умолчанию)
also_empty := []int{}       // пустой слайс
preallocated := make([]int, 0, 10)  // пустой, но с зарезервированным местом на 10

// Основные операции
fmt.Println(len(names))     // 3 — длина (сколько элементов)
fmt.Println(names[0])       // "Alice" — доступ по индексу (с 0!)
names[1] = "Robert"         // изменение элемента</code></pre>

<h2>append — добавление элементов</h2>
<pre><code>fruits := []string{"apple", "banana"}

// Добавить один элемент
fruits = append(fruits, "cherry")
// ["apple", "banana", "cherry"]

// Добавить несколько
fruits = append(fruits, "date", "elderberry")

// ВАЖНО: всегда присваивай результат обратно!
// append может создать новый слайс внутри</code></pre>

<h2>Перебор элементов — for range</h2>
<pre><code>names := []string{"Alice", "Bob", "Charlie"}

// i — индекс (номер), name — значение
for i, name := range names {
    fmt.Printf("%d: %s\n", i, name)
}
// 0: Alice
// 1: Bob
// 2: Charlie

// Если индекс не нужен — используй _
for _, name := range names {
    fmt.Println(name)
}</code></pre>

<h2>Что такое индекс?</h2>
<p><strong>Индекс</strong> — это номер позиции элемента в списке. В Go (и почти всех языках) нумерация начинается с <strong>0</strong>:</p>
<pre><code>// Индексы:  0        1       2
names := []string{"Alice", "Bob", "Charlie"}
names[0] // "Alice"  — первый элемент
names[2] // "Charlie" — третий элемент
names[3] // ОШИБКА! panic: index out of range</code></pre>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/4.1.php" target="_blank">Metanit: Массивы в Go</a> — подробный разбор на русском с примерами</li>
<li><a href="https://metanit.com/go/golang/4.3.php" target="_blank">Metanit: Срезы (слайсы)</a> — len, cap, append, copy</li>
<li><a href="https://golangify.com/slices" target="_blank">Golangify: Слайсы</a> — популярные паттерны и подводные камни</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Чем слайс отличается от массива?",
				Options:     []string{"Ничем", "Слайс может менять размер (динамический), массив — нет (фиксированный)", "Массив быстрее", "Слайс хранит только строки"},
				Correct:     1,
				Explanation: "Массив [N]T — фиксированный размер, часть типа. Слайс []T — динамический, может расти через append. В Go почти всегда используют слайсы.",
			},
			{
				Question:    "С какого числа начинается нумерация элементов в слайсе?",
				Options:     []string{"С 1", "С 0", "С -1", "Зависит от типа"},
				Correct:     1,
				Explanation: "Индексация с 0 — стандарт в большинстве языков. Первый элемент — [0], второй — [1], последний — [len-1].",
			},
			{
				Question:    "Почему нужно писать s = append(s, x)?",
				Options:     []string{"Для красоты", "append может создать новый слайс — без присваивания результат потеряется", "Go требует", "Можно и без ="},
				Correct:     1,
				Explanation: "Когда слайсу не хватает места, append создаёт новый массив внутри и возвращает новый слайс. Без s = append(...) ты потеряешь добавленный элемент.",
			},
			{
				Question:    "Что такое nil-слайс и чем он отличается от пустого []int{}?",
				Options:     []string{"Ничем, это одно и то же", "nil-слайс: var s []int (== nil, len=0, cap=0). Пустой: []int{} (не nil, len=0). Оба работают с append", "nil-слайс вызывает panic", "Пустой слайс — ошибка"},
				Correct:     1,
				Explanation: "var s []int — nil слайс. s == nil → true. []int{} — не nil. На практике оба работают с append и range. Отличие важно только при сравнении с nil.",
			},
			{
				Question:    "Что выведет: s := []int{1,2,3}; fmt.Println(s[1:3])?",
				Options:     []string{"[2 3 0]", "[2 3]", "[1 2]", "Ошибка"},
				Correct:     1,
				Explanation: "s[1:3] — элементы с индекса 1 по 2 включительно (3 не включается). s[1]=2, s[2]=3 → [2 3]. Формула: [начало:конец), конец не включается.",
			},
		},
		Tasks: []T{
			{
				Title:      "Список пользователей",
				Difficulty: "easy",
				Description: `<p>Создай слайс из трёх имён, добавь четвёртое через <code>append</code>, и выведи всех через цикл:</p>
<pre><code>1. Alice
2. Bob
3. Charlie
4. Diana</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "[]string{...}", Definition: "Создание слайса строк с начальными значениями. Квадратные скобки без числа = слайс (динамический)."},
					{Term: "append(slice, elem)", Definition: "Добавляет элемент в конец слайса и возвращает новый слайс. Обязательно: slice = append(slice, elem)."},
					{Term: "for i, v := range slice", Definition: "Перебор слайса. i — индекс (номер с 0), v — значение элемента."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "1. Alice\n2. Bob\n3. Charlie\n4. Diana"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    names := []string{"Alice", "Bob", "Charlie"}

    // Добавь "Diana" через append

    // Выведи всех через for range в формате "N. Name"

}`,
				Hints: `<p><code>names = append(names, "Diana")</code>. Затем <code>for i, name := range names { fmt.Printf("%d. %s\n", i+1, name) }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    names := []string{"Alice", "Bob", "Charlie"}
    names = append(names, "Diana")

    for i, name := range names {
        fmt.Printf("%d. %s\n", i+1, name)
    }
}</code></pre>`,
			},
			{
				Title:      "Сумма и максимум",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N чисел, находит их сумму и максимум:</p>
<p>Ввод:</p>
<pre><code>5
3 7 1 9 4</code></pre>
<p>Вывод:</p>
<pre><code>Sum: 24
Max: 9</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "make([]int, n)", Definition: "Создаёт слайс из n элементов, заполненных нулями. n — длина слайса."},
					{Term: "fmt.Scan(&var)", Definition: "Читает значение из ввода. & передаёт адрес переменной для записи."},
				},
				TestCases: []TestCase{
					{Input: "5\n3 7 1 9 4", ExpectedOutput: "Sum: 24\nMax: 9"},
					{Input: "3\n10 20 30", ExpectedOutput: "Sum: 60\nMax: 30"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Найди сумму и максимум

}`,
				Hints: `<p>Для суммы: <code>sum += nums[i]</code>. Для максимума: начни с <code>max := nums[0]</code>, затем <code>if nums[i] > max { max = nums[i] }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    sum := 0
    max := nums[0]
    for _, v := range nums {
        sum += v
        if v > max {
            max = v
        }
    }
    fmt.Printf("Sum: %d\n", sum)
    fmt.Printf("Max: %d\n", max)
}</code></pre>`,
			},
			{
				Title:      "Реверс слайса",
				Difficulty: "hard",
				Description: `<p>Напиши программу, которая читает N чисел и выводит их в обратном порядке:</p>
<p>Ввод:</p>
<pre><code>5
1 2 3 4 5</code></pre>
<p>Вывод:</p>
<pre><code>5 4 3 2 1</code></pre>
<p>Числа в выводе разделены пробелами, без пробела в конце.</p>`,
				Glossary: []GlossaryItem{
					{Term: "len(slice) - 1", Definition: "Индекс последнего элемента слайса. Помни: индексация с 0, поэтому последний = длина минус 1."},
				},
				TestCases: []TestCase{
					{Input: "5\n1 2 3 4 5", ExpectedOutput: "5 4 3 2 1"},
					{Input: "3\n10 20 30", ExpectedOutput: "30 20 10"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Выведи числа в обратном порядке через пробел

}`,
				Hints: `<p>Цикл от последнего к первому: <code>for i := len(nums) - 1; i >= 0; i--</code>. Для вывода без пробела в конце: проверяй <code>if i > 0 { fmt.Print(" ") }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    for i := len(nums) - 1; i >= 0; i-- {
        if i < len(nums)-1 {
            fmt.Print(" ")
        }
        fmt.Print(nums[i])
    }
    fmt.Println()
}</code></pre>`,
			},
			{
				Title:      "Среднее значение",
				Difficulty: "easy",
				Description: `<p>Напиши программу, которая читает N чисел и вычисляет их среднее значение:</p>
<p>Ввод:</p>
<pre><code>4
10 20 30 40</code></pre>
<p>Вывод:</p>
<pre><code>Average: 25.00</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "float64(sum) / float64(n)", Definition: "Деление int на int — целочисленное. Чтобы получить дробный результат, приводим к float64."},
				},
				TestCases: []TestCase{
					{Input: "4\n10 20 30 40", ExpectedOutput: "Average: 25.00"},
					{Input: "3\n1 2 3", ExpectedOutput: "Average: 2.00"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Вычисли сумму, затем среднее как float64

}`,
				Hints: `<p>Сумму накапливай как int: <code>sum += v</code>. Среднее: <code>avg := float64(sum) / float64(n)</code>. Вывод: <code>fmt.Printf("Average: %.2f\n", avg)</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    sum := 0
    for _, v := range nums {
        sum += v
    }
    avg := float64(sum) / float64(n)
    fmt.Printf("Average: %.2f\n", avg)
}</code></pre>`,
			},
			{
				Title:      "Уникальные элементы",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N чисел и выводит только уникальные (без повторений), в порядке первого появления:</p>
<p>Ввод:</p>
<pre><code>6
1 2 3 2 1 4</code></pre>
<p>Вывод:</p>
<pre><code>1 2 3 4</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "map для проверки уникальности", Definition: "seen := map[int]bool{}. Если seen[v] — уже видели, пропустить. Иначе seen[v] = true и добавить в результат."},
				},
				TestCases: []TestCase{
					{Input: "6\n1 2 3 2 1 4", ExpectedOutput: "1 2 3 4"},
					{Input: "4\n5 5 5 5", ExpectedOutput: "5"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    seen := map[int]bool{}
    result := []int{}

    // Для каждого числа: если не видели — добавь в result и отметь в seen

    for i, v := range result {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}`,
				Hints: `<p>В цикле: <code>if !seen[v] { seen[v] = true; result = append(result, v) }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    seen := map[int]bool{}
    result := []int{}
    for _, v := range nums {
        if !seen[v] {
            seen[v] = true
            result = append(result, v)
        }
    }

    for i, v := range result {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}</code></pre>`,
			},
		},
	}
}

func lesson_slice_operations() L {
	return L{
		Slug: "slice-operations", Title: "Операции со слайсами", Order: 2,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Операции со слайсами</h1>

<h2>Подслайс (slicing)</h2>
<p>Можно получить "кусок" слайса, указав диапазон индексов:</p>

<pre><code>nums := []int{10, 20, 30, 40, 50}

part := nums[1:3]   // [20, 30] — от индекса 1 до 3 (не включая 3)
first := nums[:3]   // [10, 20, 30] — от начала до 3
last := nums[2:]    // [30, 40, 50] — от 2 до конца</code></pre>

<p><strong>Формула:</strong> <code>slice[начало:конец]</code> — начало включается, конец НЕ включается.</p>

<h2>Удаление элемента</h2>
<pre><code>s := []int{1, 2, 3, 4, 5}

// Удалить элемент с индексом 2 (число 3)
i := 2
s = append(s[:i], s[i+1:]...)  // [1, 2, 4, 5]

// Что здесь происходит:
// s[:2]    = [1, 2]
// s[3:]    = [4, 5]
// append   = [1, 2, 4, 5]</code></pre>

<h2>copy — копирование слайса</h2>
<pre><code>original := []int{1, 2, 3}
clone := make([]int, len(original))
copy(clone, original)

clone[0] = 999
fmt.Println(original[0])  // 1 — оригинал не изменился!</code></pre>

<h2>len и cap</h2>
<p>У слайса два размера:</p>
<ul>
<li><code>len</code> — сколько элементов сейчас</li>
<li><code>cap</code> — сколько элементов может поместиться без выделения новой памяти</li>
</ul>
<pre><code>s := make([]int, 3, 10)
fmt.Println(len(s))  // 3 — три элемента
fmt.Println(cap(s))  // 10 — место для 10</code></pre>

<h2>sort — сортировка</h2>
<pre><code>import "sort"

nums := []int{5, 3, 1, 4, 2}
sort.Ints(nums)          // [1, 2, 3, 4, 5]

names := []string{"Charlie", "Alice", "Bob"}
sort.Strings(names)      // ["Alice", "Bob", "Charlie"]</code></pre>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/4.4.php" target="_blank">Metanit: Операции со слайсами</a> — копирование, нарезка, удаление</li>
<li><a href="https://habr.com/ru/articles/325468/" target="_blank">Хабр: Срезы в Go: всё что нужно знать</a> — внутренняя структура, capacity, gotchas</li>
<li><a href="https://pkg.go.dev/sort" target="_blank">Документация пакета sort</a> — все функции сортировки</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Что вернёт nums[1:3] для []int{10, 20, 30, 40, 50}?",
				Options:     []string{"[10, 20, 30]", "[20, 30]", "[20, 30, 40]", "[10, 20]"},
				Correct:     1,
				Explanation: "slice[1:3] берёт элементы с индексом 1 и 2 (не включая 3). nums[1]=20, nums[2]=30 → [20, 30].",
			},
			{
				Question:    "Как удалить элемент из слайса?",
				Options:     []string{"delete(s, i)", "s = append(s[:i], s[i+1:]...)", "s.Remove(i)", "s[i] = nil"},
				Correct:     1,
				Explanation: "В Go нет встроенной функции удаления из слайса. Используют append: склеивают часть до элемента и часть после.",
			},
			{
				Question:    "Что произойдёт если изменить элемент подслайса?",
				Options:     []string{"Изменится только подслайс", "Изменится и оригинальный слайс — они разделяют один массив", "Ошибка компиляции", "Изменится только если использовать copy"},
				Correct:     1,
				Explanation: "s[1:3] не копирует данные — это другой заголовок (len, cap, указатель) на тот же массив. Изменение sub[0] = 99 изменит s[1]. Для независимой копии используй copy().",
			},
			{
				Question:    "Как правильно скопировать слайс чтобы оригинал не изменился?",
				Options:     []string{"clone := s", "clone := s[:]", "clone := make([]int, len(s)); copy(clone, s)", "clone := append([]int{}, s...)"},
				Correct:     2,
				Explanation: "clone := s просто копирует заголовок — оба указывают на один массив. Для настоящей копии: make + copy, или append([]int{}, s...). Оба варианта [2] и [3] верны, но [2] более явный.",
			},
			{
				Question:    "Что такое capacity (cap) слайса?",
				Options:     []string{"Максимальный размер в байтах", "Сколько элементов можно добавить без нового выделения памяти", "То же что len", "Количество удалённых элементов"},
				Correct:     1,
				Explanation: "cap — размер underlying array. Пока len <= cap, append не выделяет новый массив. Когда len == cap, append удваивает cap (примерно) и копирует данные — это дорого.",
			},
		},
		Tasks: []T{
			{
				Title:      "Фильтрация чётных",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N чисел и выводит только чётные:</p>
<p>Ввод:</p>
<pre><code>6
1 2 3 4 5 6</code></pre>
<p>Вывод:</p>
<pre><code>2 4 6</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "n % 2 == 0", Definition: "Проверка на чётность. % — остаток от деления. Если остаток от деления на 2 равен 0, число чётное."},
					{Term: "append(result, v)", Definition: "Добавить элемент v в слайс result. Используй для создания нового слайса с отфильтрованными значениями."},
				},
				TestCases: []TestCase{
					{Input: "6\n1 2 3 4 5 6", ExpectedOutput: "2 4 6"},
					{Input: "4\n10 15 20 25", ExpectedOutput: "10 20"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Собери чётные числа в новый слайс и выведи через пробел

}`,
				Hints: `<p>Создай <code>even := []int{}</code>, в цикле проверяй <code>if v % 2 == 0</code> и добавляй через <code>append</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    even := []int{}
    for _, v := range nums {
        if v%2 == 0 {
            even = append(even, v)
        }
    }

    for i, v := range even {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}</code></pre>`,
			},
			{
				Title:      "Сортировка и поиск",
				Difficulty: "hard",
				Description: `<p>Напиши программу, которая читает N чисел, сортирует их и выводит отсортированный список, минимум и максимум:</p>
<p>Ввод:</p>
<pre><code>5
38 27 43 3 9</code></pre>
<p>Вывод:</p>
<pre><code>Sorted: 3 9 27 38 43
Min: 3
Max: 43</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "sort.Ints(slice)", Definition: "Сортирует слайс int по возрастанию. Изменяет слайс на месте (не создаёт новый). Нужен import \"sort\"."},
				},
				TestCases: []TestCase{
					{Input: "5\n38 27 43 3 9", ExpectedOutput: "Sorted: 3 9 27 38 43\nMin: 3\nMax: 43"},
					{Input: "3\n5 1 3", ExpectedOutput: "Sorted: 1 3 5\nMin: 1\nMax: 5"},
				},
				StarterCode: `package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Отсортируй, выведи список, минимум и максимум
    _ = sort.Ints // убери эту строку когда используешь sort

}`,
				Hints: `<p>После <code>sort.Ints(nums)</code> первый элемент — минимум, последний — максимум: <code>nums[0]</code> и <code>nums[len(nums)-1]</code>.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    sort.Ints(nums)

    fmt.Print("Sorted:")
    for _, v := range nums {
        fmt.Printf(" %d", v)
    }
    fmt.Println()
    fmt.Printf("Min: %d\n", nums[0])
    fmt.Printf("Max: %d\n", nums[len(nums)-1])
}</code></pre>`,
			},
			{
				Title:      "Срез подслайса",
				Difficulty: "easy",
				Description: `<p>Напиши программу с фиксированным слайсом <code>[]int{10, 20, 30, 40, 50}</code> и выведи три подслайса:</p>
<pre><code>First 3: [10 20 30]
Last 3: [30 40 50]
Middle: [20 30 40]</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "s[:n]", Definition: "Первые n элементов. Эквивалент s[0:n]."},
					{Term: "s[n:]", Definition: "Элементы начиная с индекса n до конца."},
					{Term: "fmt.Println(slice)", Definition: "Выводит слайс в формате [a b c] автоматически."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "First 3: [10 20 30]\nLast 3: [30 40 50]\nMiddle: [20 30 40]"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    s := []int{10, 20, 30, 40, 50}

    // Выведи первые 3, последние 3, средние 3

}`,
				Hints: `<p><code>s[:3]</code> — первые 3. <code>s[2:]</code> — последние 3. <code>s[1:4]</code> — средние.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    s := []int{10, 20, 30, 40, 50}

    fmt.Println("First 3:", s[:3])
    fmt.Println("Last 3:", s[2:])
    fmt.Println("Middle:", s[1:4])
}</code></pre>`,
			},
			{
				Title:      "Удаление дубликатов",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N чисел (отсортированный слайс) и удаляет дубликаты на месте:</p>
<p>Ввод:</p>
<pre><code>7
1 1 2 3 3 3 4</code></pre>
<p>Вывод:</p>
<pre><code>1 2 3 4</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "два указателя (two pointers)", Definition: "Классический алгоритм: один указатель идёт вперёд, второй пишет уникальные значения. Эффективно для отсортированных данных."},
				},
				TestCases: []TestCase{
					{Input: "7\n1 1 2 3 3 3 4", ExpectedOutput: "1 2 3 4"},
					{Input: "5\n1 1 1 1 1", ExpectedOutput: "1"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    // Собери уникальные: если nums[i] != предыдущего, добавь
    result := []int{}
    for i, v := range nums {
        if i == 0 || v != nums[i-1] {
            result = append(result, v)
        }
    }

    for i, v := range result {
        if i > 0 {
            fmt.Print(" ")
        }
        fmt.Print(v)
    }
    fmt.Println()
}`,
				Hints: `<p>Если слайс отсортирован — дубликат стоит рядом с предыдущим. Проверь: <code>if i == 0 || v != nums[i-1]</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    result := []int{}
    for i, v := range nums {
        if i == 0 || v != nums[i-1] {
            result = append(result, v)
        }
    }

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
				Title:      "Второй максимум",
				Difficulty: "hard",
				Description: `<p>Напиши программу, которая находит второй по величине уникальный элемент в слайсе:</p>
<p>Ввод:</p>
<pre><code>5
3 1 4 1 5</code></pre>
<p>Вывод: <code>Second max: 4</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "math.MinInt", Definition: "Минимально возможное значение int. Удобно как начальное значение для поиска максимума."},
				},
				TestCases: []TestCase{
					{Input: "5\n3 1 4 1 5", ExpectedOutput: "Second max: 4"},
					{Input: "4\n10 10 5 5", ExpectedOutput: "Second max: 5"},
				},
				StarterCode: `package main

import (
    "fmt"
    "math"
)

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    first := math.MinInt
    second := math.MinInt

    for _, v := range nums {
        if v > first {
            second = first
            first = v
        } else if v > second && v != first {
            second = v
        }
    }

    fmt.Printf("Second max: %d\n", second)
}`,
				Hints: `<p>Следи за двумя переменными: <code>first</code> (максимум) и <code>second</code> (второй). Если новое значение больше first — сдвигай. Если между second и first — обновляй second.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "math"
)

func main() {
    var n int
    fmt.Scan(&n)

    nums := make([]int, n)
    for i := 0; i < n; i++ {
        fmt.Scan(&nums[i])
    }

    first := math.MinInt
    second := math.MinInt

    for _, v := range nums {
        if v > first {
            second = first
            first = v
        } else if v > second && v != first {
            second = v
        }
    }

    fmt.Printf("Second max: %d\n", second)
}</code></pre>`,
			},
		},
	}
}

func lesson_strings_deep() L {
	return L{
		Slug: "strings", Title: "Строки и работа с текстом", Order: 3,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Строки и работа с текстом</h1>

<h2>Строки в Go — это набор байтов</h2>
<p>Строка в Go хранится в кодировке <strong>UTF-8</strong>. Это значит, что латинские буквы занимают 1 байт, а кириллица — 2 байта:</p>

<pre><code>s := "Hello"
fmt.Println(len(s))  // 5 — каждая буква = 1 байт

s2 := "Привет"
fmt.Println(len(s2)) // 12 — каждая буква = 2 байта!</code></pre>

<h2>Пакет strings — работа с текстом</h2>
<pre><code>import "strings"

// Поиск
strings.Contains("WatchTogether", "Watch")   // true — содержит?
strings.HasPrefix("video.mp4", "video")      // true — начинается с?
strings.HasSuffix("video.mp4", ".mp4")       // true — заканчивается на?

// Преобразование
strings.ToUpper("hello")                      // "HELLO"
strings.ToLower("HELLO")                      // "hello"
strings.TrimSpace("  hello  ")               // "hello" — убрать пробелы

// Разделение и объединение
strings.Split("a,b,c", ",")                  // ["a", "b", "c"]
strings.Join([]string{"a", "b", "c"}, "-")   // "a-b-c"

// Замена
strings.ReplaceAll("foo bar foo", "foo", "baz")  // "baz bar baz"

// Подсчёт
strings.Count("banana", "a")                 // 3</code></pre>

<h2>Преобразование строка ↔ число</h2>
<pre><code>import "strconv"

// Число → строка
s := strconv.Itoa(42)          // "42"

// Строка → число
n, err := strconv.Atoi("42")  // n=42, err=nil
n, err := strconv.Atoi("abc") // n=0, err=ошибка!</code></pre>

<h2>rune — символ Unicode</h2>
<pre><code>// Для подсчёта символов (не байтов) — конвертируй в руны
s := "Привет"
runes := []rune(s)
fmt.Println(len(runes))  // 6 — символов

// for range по строке итерирует по символам автоматически
for _, ch := range "Привет" {
    fmt.Printf("%c ", ch)  // П р и в е т
}</code></pre>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/4.7.php" target="_blank">Metanit: Строки в Go</a> — bytes, runes, UTF-8 на русском</li>
<li><a href="https://habr.com/ru/articles/704772/" target="_blank">Хабр: Строки в Go — разбираем детали</a> — как устроена строка, UTF-8, rune</li>
<li><a href="https://pkg.go.dev/strings" target="_blank">Документация пакета strings</a> — все функции с примерами</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Что вернёт len(\"Привет\")?",
				Options:     []string{"6", "12", "7", "Ошибку"},
				Correct:     1,
				Explanation: "len() считает байты, не символы. Кириллица в UTF-8 = 2 байта на символ. 6 символов × 2 = 12 байт.",
			},
			{
				Question:    "Как разделить строку \"a,b,c\" по запятым?",
				Options:     []string{"\"a,b,c\".Split(\",\")", "strings.Split(\"a,b,c\", \",\")", "split(\"a,b,c\", \",\")", "strings.Divide(\"a,b,c\", \",\")"},
				Correct:     1,
				Explanation: "strings.Split(str, sep) разделяет строку по разделителю и возвращает слайс строк.",
			},
			{
				Question:    "Как правильно посчитать количество символов (не байт) в кириллической строке?",
				Options:     []string{"len(s)", "len([]byte(s))", "len([]rune(s))", "strings.Count(s, \"\")"},
				Correct:     2,
				Explanation: "len(s) считает байты. Кириллица = 2 байта/символ. Для подсчёта символов: len([]rune(s)) или strings.Count(s, \"\") (последний работает с Unicode правильно).",
			},
			{
				Question:    "Что делает strings.TrimSpace?",
				Options:     []string{"Удаляет все пробелы внутри строки", "Удаляет пробелы и \\n в начале и конце строки", "Удаляет только пробелы в начале", "Заменяет пробелы на _"},
				Correct:     1,
				Explanation: "TrimSpace удаляет все пробельные символы (пробел, \\t, \\n, \\r) с обоих концов строки. Полезно для обработки пользовательского ввода.",
			},
			{
				Question:    "Чем strings.Contains отличается от strings.Index?",
				Options:     []string{"Ничем", "Contains возвращает bool (есть/нет), Index возвращает позицию (-1 если нет)", "Index быстрее", "Contains только для символов"},
				Correct:     1,
				Explanation: "Contains → true/false. Index → позиция первого вхождения или -1. Используй Contains когда нужно только знать «есть ли», Index — когда нужна позиция.",
			},
		},
		Tasks: []T{
			{
				Title:      "Подсчёт слов",
				Difficulty: "easy",
				Description: `<p>Напиши программу, которая читает строку и считает количество слов:</p>
<p>Ввод: <code>Hello World Go</code></p>
<p>Вывод: <code>Words: 3</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "strings.Fields(s)", Definition: "Разделяет строку по пробелам (любому количеству) и возвращает слайс слов. Лучше Split для слов."},
					{Term: "bufio.NewScanner(os.Stdin)", Definition: "Для чтения целой строки (с пробелами). fmt.Scan читает только до первого пробела."},
				},
				TestCases: []TestCase{
					{Input: "Hello World Go", ExpectedOutput: "Words: 3"},
					{Input: "one", ExpectedOutput: "Words: 1"},
				},
				StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    // Посчитай слова и выведи "Words: N"
    _ = strings.Fields // убери когда используешь
    _ = line

}`,
				Hints: `<p><code>words := strings.Fields(line)</code> разделит строку по пробелам. Затем <code>len(words)</code> даст количество.</p>`,
				Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    words := strings.Fields(line)
    fmt.Printf("Words: %d\n", len(words))
}</code></pre>`,
			},
			{
				Title:      "Форматирование имени",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает имя и фамилию (через пробел) и форматирует:</p>
<p>Ввод: <code>alice smith</code></p>
<p>Вывод:</p>
<pre><code>Full: Alice Smith
Username: asmith
Email: alice.smith@watchtogether.com</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "strings.Title(s) / strings.ToUpper", Definition: "Title делает первую букву заглавной. Для новых версий Go используй cases.Title."},
					{Term: "strings.ToLower(s)", Definition: "Переводит всю строку в нижний регистр."},
					{Term: "s[0:1]", Definition: "Взять первый байт строки как подстроку. Для ASCII-символов это первая буква."},
				},
				TestCases: []TestCase{
					{Input: "alice smith", ExpectedOutput: "Full: Alice Smith\nUsername: asmith\nEmail: alice.smith@watchtogether.com"},
					{Input: "bob jones", ExpectedOutput: "Full: Bob Jones\nUsername: bjones\nEmail: bob.jones@watchtogether.com"},
				},
				StarterCode: `package main

import (
    "fmt"
    "strings"
)

func main() {
    var first, last string
    fmt.Scan(&first, &last)

    first = strings.ToLower(first)
    last = strings.ToLower(last)

    // Сформируй и выведи Full, Username, Email

}`,
				Hints: `<p>Capitalize: <code>strings.ToUpper(s[:1]) + s[1:]</code>. Username: <code>first[:1] + last</code>. Email: <code>first + "." + last + "@watchtogether.com"</code></p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "strings"
)

func main() {
    var first, last string
    fmt.Scan(&first, &last)

    first = strings.ToLower(first)
    last = strings.ToLower(last)

    capFirst := strings.ToUpper(first[:1]) + first[1:]
    capLast := strings.ToUpper(last[:1]) + last[1:]

    fmt.Printf("Full: %s %s\n", capFirst, capLast)
    fmt.Printf("Username: %s%s\n", first[:1], last)
    fmt.Printf("Email: %s.%s@watchtogether.com\n", first, last)
}</code></pre>`,
			},
			{
				Title:      "Проверка палиндрома",
				Difficulty: "easy",
				Description: `<p>Напиши программу, которая проверяет — является ли строка палиндромом (читается одинаково слева направо и справа налево):</p>
<p>Ввод: <code>racecar</code> → Вывод: <code>true</code></p>
<p>Ввод: <code>hello</code> → Вывод: <code>false</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "[]rune(s)", Definition: "Конвертирует строку в слайс символов. Позволяет обращаться к символам по индексу и разворачивать строку."},
				},
				TestCases: []TestCase{
					{Input: "racecar", ExpectedOutput: "true"},
					{Input: "hello", ExpectedOutput: "false"},
					{Input: "madam", ExpectedOutput: "true"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var s string
    fmt.Scan(&s)

    runes := []rune(s)
    isPalindrome := true

    // Проверь: первый символ == последнему, второй == предпоследнему и т.д.

    fmt.Println(isPalindrome)
}`,
				Hints: `<p>Цикл до len(runes)/2. Сравни <code>runes[i]</code> с <code>runes[len(runes)-1-i]</code>. Если хоть одна пара не совпала — false.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var s string
    fmt.Scan(&s)

    runes := []rune(s)
    isPalindrome := true
    for i := 0; i < len(runes)/2; i++ {
        if runes[i] != runes[len(runes)-1-i] {
            isPalindrome = false
            break
        }
    }
    fmt.Println(isPalindrome)
}</code></pre>`,
			},
			{
				Title:      "CSV парсер",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает CSV-строку и выводит каждое поле с номером:</p>
<p>Ввод: <code>Alice,25,premium,active</code></p>
<p>Вывод:</p>
<pre><code>1: Alice
2: 25
3: premium
4: active</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "strings.Split(s, \",\")", Definition: "Разделяет строку по запятой. Возвращает []string. Базовый CSV-парсер для простых случаев без кавычек."},
				},
				TestCases: []TestCase{
					{Input: "Alice,25,premium,active", ExpectedOutput: "1: Alice\n2: 25\n3: premium\n4: active"},
					{Input: "Bob,30,free", ExpectedOutput: "1: Bob\n2: 30\n3: free"},
				},
				StarterCode: `package main

import (
    "fmt"
    "strings"
)

func main() {
    var line string
    fmt.Scan(&line)

    // Разбей по запятой и выведи с номерами

}`,
				Hints: `<p><code>fields := strings.Split(line, ",")</code>, затем <code>for i, f := range fields { fmt.Printf("%d: %s\n", i+1, f) }</code></p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "strings"
)

func main() {
    var line string
    fmt.Scan(&line)

    fields := strings.Split(line, ",")
    for i, f := range fields {
        fmt.Printf("%d: %s\n", i+1, f)
    }
}</code></pre>`,
			},
			{
				Title:      "Анализатор текста",
				Difficulty: "hard",
				Description: `<p>Напиши программу, которая читает строку и выводит статистику:</p>
<p>Ввод: <code>Hello World Go</code></p>
<p>Вывод:</p>
<pre><code>Words: 3
Chars: 15
Upper: 3
Lower: 9</code></pre>
<p>Chars — все символы включая пробелы. Upper — заглавные буквы. Lower — строчные.</p>`,
				Glossary: []GlossaryItem{
					{Term: "unicode.IsUpper(r)", Definition: "Проверяет, является ли символ заглавной буквой. Нужен import \"unicode\"."},
					{Term: "unicode.IsLower(r)", Definition: "Проверяет, является ли символ строчной буквой."},
				},
				TestCases: []TestCase{
					{Input: "Hello World Go", ExpectedOutput: "Words: 3\nChars: 14\nUpper: 3\nLower: 9"},
				},
				StarterCode: `package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "unicode"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    words := strings.Fields(line)
    upper, lower := 0, 0

    for _, ch := range line {
        if unicode.IsUpper(ch) {
            upper++
        } else if unicode.IsLower(ch) {
            lower++
        }
    }

    fmt.Printf("Words: %d\n", len(words))
    fmt.Printf("Chars: %d\n", len([]rune(line)))
    fmt.Printf("Upper: %d\n", upper)
    fmt.Printf("Lower: %d\n", lower)
}`,
				Hints: `<p>Используй <code>for _, ch := range line</code> для перебора символов. <code>unicode.IsUpper(ch)</code> и <code>unicode.IsLower(ch)</code> для определения регистра.</p>`,
				Solution: `<pre><code>package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "unicode"
)

func main() {
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan()
    line := scanner.Text()

    words := strings.Fields(line)
    upper, lower := 0, 0

    for _, ch := range line {
        if unicode.IsUpper(ch) {
            upper++
        } else if unicode.IsLower(ch) {
            lower++
        }
    }

    fmt.Printf("Words: %d\n", len(words))
    fmt.Printf("Chars: %d\n", len([]rune(line)))
    fmt.Printf("Upper: %d\n", upper)
    fmt.Printf("Lower: %d\n", lower)
}</code></pre>`,
			},
		},
	}
}

func lesson_maps() L {
	return L{
		Slug: "maps", Title: "Map — словарь ключ-значение", Order: 4,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Map — словарь ключ-значение</h1>

<h2>Что такое map?</h2>
<p><strong>Map</strong> (карта, словарь) — структура данных, которая хранит пары <strong>ключ → значение</strong>. Как телефонная книга: имя → номер.</p>

<pre><code>// Создание map
ages := map[string]int{
    "Alice":   25,
    "Bob":     30,
    "Charlie": 35,
}

// Или пустой map
scores := make(map[string]int)

// Добавление/изменение
ages["Diana"] = 28     // добавить новый
ages["Alice"] = 26     // изменить существующий

// Чтение
fmt.Println(ages["Bob"])  // 30

// Удаление
delete(ages, "Charlie")

// Длина
fmt.Println(len(ages))  // 3</code></pre>

<h2>Проверка наличия ключа</h2>
<pre><code>// Если ключа нет — map вернёт нулевое значение (0 для int, "" для string)
fmt.Println(ages["Unknown"])  // 0 — но мы не знаем, это 0 или ключа нет!

// Правильный способ:
age, exists := ages["Unknown"]
if exists {
    fmt.Println("Возраст:", age)
} else {
    fmt.Println("Не найден")
}

// Короткая форма:
if age, ok := ages["Alice"]; ok {
    fmt.Println("Alice:", age)
}</code></pre>

<h2>Перебор map</h2>
<pre><code>for name, age := range ages {
    fmt.Printf("%s: %d\n", name, age)
}
// Порядок СЛУЧАЙНЫЙ! При каждом запуске может быть разным</code></pre>

<h2>Важные особенности</h2>
<ul>
<li>Ключи должны быть сравнимого типа (string, int, bool — да; slice, map — нет)</li>
<li>Порядок итерации случайный</li>
<li>Нулевой map (<code>var m map[string]int</code>) = nil — нельзя писать! Используй <code>make</code></li>
<li>Map передаётся по ссылке — изменения в функции видны снаружи</li>
</ul>

<h2>Рекомендуемые ресурсы</h2>
<ul>
<li><a href="https://metanit.com/go/golang/4.5.php" target="_blank">Metanit: Словари (map) в Go</a> — создание, итерация, работа с ключами</li>
<li><a href="https://habr.com/ru/articles/457728/" target="_blank">Хабр: Устройство map в Go изнутри</a> — как map работает под капотом (хэш-таблицы)</li>
<li><a href="https://golangify.com/maps" target="_blank">Golangify: Map в Go</a> — типичные ошибки и паттерны</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Что вернёт map если ключа нет?",
				Options:     []string{"Ошибку", "nil", "Нулевое значение типа (0 для int, \"\" для string)", "panic"},
				Correct:     2,
				Explanation: "Map возвращает нулевое значение если ключа нет. Для проверки наличия используй второе возвращаемое значение: val, ok := m[key].",
			},
			{
				Question:    "Как удалить элемент из map?",
				Options:     []string{"m[key] = nil", "delete(m, key)", "m.Remove(key)", "del m[key]"},
				Correct:     1,
				Explanation: "Встроенная функция delete(map, key) удаляет элемент по ключу. Если ключа нет — ничего не происходит (без ошибки).",
			},
			{
				Question:    "Почему нельзя использовать nil map для записи?",
				Options:     []string{"Можно, всё в порядке", "nil map вызовет panic: assignment to entry in nil map", "Это просто предупреждение", "nil map автоматически инициализируется при первой записи"},
				Correct:     1,
				Explanation: "var m map[string]int — nil map. Запись m[\"key\"] = 1 вызовет panic. Чтение из nil map безопасно (вернёт нулевое значение). Всегда инициализируй через make(map[K]V).",
			},
			{
				Question:    "Гарантирован ли порядок итерации по map?",
				Options:     []string{"Да, по порядку вставки", "Да, по алфавиту ключей", "Нет — порядок случайный и меняется от запуска к запуску", "Да, по хэш-значению"},
				Correct:     2,
				Explanation: "Go намеренно рандомизирует порядок итерации map с каждым запуском (security feature). Для стабильного порядка — соберай ключи в слайс, отсортируй, потом перебирай.",
			},
			{
				Question:    "Можно ли использовать слайс как ключ map?",
				Options:     []string{"Да, любой тип", "Нет — только comparable типы: string, int, bool, struct без слайсов/map", "Да, если слайс не nil", "Только []byte"},
				Correct:     1,
				Explanation: "Ключ map должен поддерживать == и !=. Слайс и map не поддерживают сравнение — нельзя. Строки, числа, bool, указатели, массивы (не слайсы!), struct из comparable полей — можно.",
			},
		},
		Tasks: []T{
			{
				Title:      "Подсчёт символов",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает строку и считает количество каждого символа, затем выводит в формате:</p>
<p>Ввод: <code>hello</code></p>
<p>Вывод (по алфавиту):</p>
<pre><code>e: 1
h: 1
l: 2
o: 1</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "map[rune]int", Definition: "Map где ключ — символ (rune), значение — количество (int). Идеально для подсчёта частоты символов."},
					{Term: "sort.Slice", Definition: "Сортирует слайс по произвольному условию. sort.Slice(s, func(i, j int) bool { return s[i] < s[j] })"},
				},
				TestCases: []TestCase{
					{Input: "hello", ExpectedOutput: "e: 1\nh: 1\nl: 2\no: 1"},
					{Input: "aab", ExpectedOutput: "a: 2\nb: 1"},
				},
				StarterCode: `package main

import (
    "fmt"
    "sort"
)

func main() {
    var s string
    fmt.Scan(&s)

    // Посчитай символы в map, отсортируй ключи, выведи
    _ = sort.Slice // убери когда используешь

}`,
				Hints: `<p>Создай <code>counts := map[rune]int{}</code>. Перебери строку: <code>for _, ch := range s { counts[ch]++ }</code>. Собери ключи в слайс, отсортируй, выведи.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func main() {
    var s string
    fmt.Scan(&s)

    counts := map[rune]int{}
    for _, ch := range s {
        counts[ch]++
    }

    keys := []rune{}
    for k := range counts {
        keys = append(keys, k)
    }
    sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

    for _, k := range keys {
        fmt.Printf("%c: %d\n", k, counts[k])
    }
}</code></pre>`,
			},
			{
				Title:      "Телефонная книга",
				Difficulty: "hard",
				Description: `<p>Реализуй простую телефонную книгу. Программа читает N операций:</p>
<ul>
<li><code>add Name Number</code> — добавить/обновить контакт</li>
<li><code>find Name</code> — найти номер (вывести "Name: Number" или "Not found")</li>
<li><code>delete Name</code> — удалить контакт</li>
</ul>
<p>Ввод:</p>
<pre><code>5
add Alice 12345
add Bob 67890
find Alice
delete Bob
find Bob</code></pre>
<p>Вывод:</p>
<pre><code>Alice: 12345
Not found</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "make(map[string]string)", Definition: "Создаёт пустой map. Обязательно перед использованием — nil map вызовет panic при записи."},
					{Term: "delete(m, key)", Definition: "Удаляет пару ключ-значение из map. Если ключа нет — ничего не произойдёт."},
					{Term: "val, ok := m[key]", Definition: "Проверка наличия ключа. ok=true если ключ существует, false если нет."},
				},
				TestCases: []TestCase{
					{Input: "5\nadd Alice 12345\nadd Bob 67890\nfind Alice\ndelete Bob\nfind Bob", ExpectedOutput: "Alice: 12345\nNot found"},
					{Input: "3\nadd Test 111\nfind Test\nfind Missing", ExpectedOutput: "Test: 111\nNot found"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    book := make(map[string]string)

    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)

        // Обработай команды: add, find, delete
        _ = book
    }
}`,
				Hints: `<p>Для add: <code>fmt.Scan(&name, &number); book[name] = number</code>. Для find: <code>if val, ok := book[name]; ok { ... }</code>. Для delete: <code>delete(book, name)</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    book := make(map[string]string)

    for i := 0; i < n; i++ {
        var cmd string
        fmt.Scan(&cmd)

        switch cmd {
        case "add":
            var name, number string
            fmt.Scan(&name, &number)
            book[name] = number
        case "find":
            var name string
            fmt.Scan(&name)
            if val, ok := book[name]; ok {
                fmt.Printf("%s: %s\n", name, val)
            } else {
                fmt.Println("Not found")
            }
        case "delete":
            var name string
            fmt.Scan(&name)
            delete(book, name)
        }
    }
}</code></pre>`,
			},
			{
				Title:      "Счётчик слов",
				Difficulty: "easy",
				Description: `<p>Напиши программу, которая читает N слов и выводит сколько раз встретилось каждое:</p>
<p>Ввод:</p>
<pre><code>5
go go java python go</code></pre>
<p>Вывод (по алфавиту):</p>
<pre><code>go: 3
java: 1
python: 1</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "counts[word]++", Definition: "Если ключа нет — Go вернёт 0 (нулевое значение int), затем прибавит 1. Идиоматичный способ подсчёта."},
				},
				TestCases: []TestCase{
					{Input: "5\ngo go java python go", ExpectedOutput: "go: 3\njava: 1\npython: 1"},
					{Input: "3\na b a", ExpectedOutput: "a: 2\nb: 1"},
				},
				StarterCode: `package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    counts := make(map[string]int)
    for i := 0; i < n; i++ {
        var word string
        fmt.Scan(&word)
        counts[word]++
    }

    // Собери ключи, отсортируй, выведи
    keys := []string{}
    for k := range counts {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, counts[k])
    }
}`,
				Hints: `<p><code>counts[word]++</code> работает даже для новых ключей — Go автоматически инициализирует нулём. Для вывода по алфавиту: собери ключи в слайс и отсортируй.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    counts := make(map[string]int)
    for i := 0; i < n; i++ {
        var word string
        fmt.Scan(&word)
        counts[word]++
    }

    keys := []string{}
    for k := range counts {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %d\n", k, counts[k])
    }
}</code></pre>`,
			},
			{
				Title:      "Инвертирование map",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N пар ключ-значение и выводит инвертированный map (значение становится ключом):</p>
<p>Ввод:</p>
<pre><code>3
alice A
bob B
charlie C</code></pre>
<p>Вывод (по алфавиту):</p>
<pre><code>A: alice
B: bob
C: charlie</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "инвертирование map", Definition: "Создание нового map где ключи и значения меняются местами. Применяется для двунаправленного поиска."},
				},
				TestCases: []TestCase{
					{Input: "3\nalice A\nbob B\ncharlie C", ExpectedOutput: "A: alice\nB: bob\nC: charlie"},
				},
				StarterCode: `package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    original := make(map[string]string)
    for i := 0; i < n; i++ {
        var k, v string
        fmt.Scan(&k, &v)
        original[k] = v
    }

    // Создай инвертированный map и выведи по алфавиту ключей

    _ = sort.Strings
}`,
				Hints: `<p>Для инверсии: <code>for k, v := range original { inverted[v] = k }</code>. Затем отсортируй ключи inverted и выведи.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
)

func main() {
    var n int
    fmt.Scan(&n)

    original := make(map[string]string)
    for i := 0; i < n; i++ {
        var k, v string
        fmt.Scan(&k, &v)
        original[k] = v
    }

    inverted := make(map[string]string)
    for k, v := range original {
        inverted[v] = k
    }

    keys := []string{}
    for k := range inverted {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %s\n", k, inverted[k])
    }
}</code></pre>`,
			},
			{
				Title:      "Группировка по первой букве",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая читает N имён и группирует их по первой букве:</p>
<p>Ввод:</p>
<pre><code>5
Alice Bob Anna Charlie Bob</code></pre>
<p>Вывод (по алфавиту ключей):</p>
<pre><code>A: Alice Anna
B: Bob Bob
C: Charlie</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "map[string][]string", Definition: "Map где значение — слайс строк. Позволяет группировать элементы по ключу."},
					{Term: "string(name[0])", Definition: "Первый байт строки как строка. Для ASCII-имён (латиница) это первая буква."},
				},
				TestCases: []TestCase{
					{Input: "5\nAlice Bob Anna Charlie Bob", ExpectedOutput: "A: Alice Anna\nB: Bob Bob\nC: Charlie"},
				},
				StarterCode: `package main

import (
    "fmt"
    "sort"
    "strings"
)

func main() {
    var n int
    fmt.Scan(&n)

    groups := make(map[string][]string)
    for i := 0; i < n; i++ {
        var name string
        fmt.Scan(&name)
        key := string(name[0])
        groups[key] = append(groups[key], name)
    }

    keys := []string{}
    for k := range groups {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %s\n", k, strings.Join(groups[k], " "))
    }
}`,
				Hints: `<p>Ключ — первая буква: <code>key := string(name[0])</code>. Добавление в группу: <code>groups[key] = append(groups[key], name)</code>. Вывод группы: <code>strings.Join(groups[k], " ")</code>.</p>`,
				Solution: `<pre><code>package main

import (
    "fmt"
    "sort"
    "strings"
)

func main() {
    var n int
    fmt.Scan(&n)

    groups := make(map[string][]string)
    for i := 0; i < n; i++ {
        var name string
        fmt.Scan(&name)
        key := string(name[0])
        groups[key] = append(groups[key], name)
    }

    keys := []string{}
    for k := range groups {
        keys = append(keys, k)
    }
    sort.Strings(keys)

    for _, k := range keys {
        fmt.Printf("%s: %s\n", k, strings.Join(groups[k], " "))
    }
}</code></pre>`,
			},
		},
	}
}
