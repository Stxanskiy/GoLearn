package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 1: Первые шаги в Go (для абсолютных новичков)
// Каждый термин объясняется с нуля. Ноль предварительных знаний.
// ════════════════════════════════════════════════════════════════

func mod01_basics_new() M {
	return M{
		Slug:        "basics",
		Title:       "Первые шаги в Go",
		Description: "Начинаем с самого нуля: что такое программа, как с ней разговаривать, и как Go превращает текст в работающее приложение. Никакого предыдущего опыта не требуется.",
		Order:       1,
		Track:       "shared",
		Difficulty:  "beginner",
		Lessons: []L{
			lesson01_what_is_program(),
			lesson02_first_program(),
			lesson03_variables(),
			lesson04_types_and_conversions(),
			lesson05_input_output(),
			lesson06_conditions(),
			lesson07_loops(),
		},
	}
}

// ── Урок 1: Что такое программа ──────────────────────────────

func lesson01_what_is_program() L {
	return L{
		Slug: "what-is-program", Title: "Твоя первая программа", Order: 1,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Твоя первая программа на Go</h1>

<h2>Вот она — твоя первая программа</h2>
<p>Не читай — сначала посмотри на код:</p>

<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Привет, мир!")
}</code></pre>

<p>Запусти её — и ты увидишь на экране:</p>
<pre><code>Привет, мир!</code></pre>

<p><strong>Поздравляю!</strong> Ты только что написал работающую программу. Теперь разберём что каждая строка делает.</p>

<h2>Разбор по строкам</h2>

<pre><code>package main          // 1. Говорит Go: "это запускаемая программа"
import "fmt"          // 2. Подключаем пакет fmt — для вывода текста на экран
func main() {         // 3. Главная функция — программа начинается отсюда
    fmt.Println("Привет, мир!")  // 4. Выводит текст на экран
}                     // 5. Конец функции</code></pre>

<table>
<tr><th>Строка</th><th>Что делает</th><th>Аналогия</th></tr>
<tr><td><code>package main</code></td><td>Объявляет что это программа</td><td>Обложка книги — "это роман"</td></tr>
<tr><td><code>import "fmt"</code></td><td>Подключает инструмент для печати</td><td>Берём ручку чтобы писать</td></tr>
<tr><td><code>func main()</code></td><td>Точка входа — отсюда всё начинается</td><td>Первая страница книги</td></tr>
<tr><td><code>fmt.Println(...)</code></td><td>Печатает текст на экран</td><td>Говоришь вслух</td></tr>
</table>

<h2>Важные правила (запомни!)</h2>

<p><strong>1. Текст всегда в двойных кавычках:</strong></p>
<pre><code>fmt.Println("Привет")   // ✅ правильно — двойные кавычки
fmt.Println('Привет')   // ❌ ошибка — одинарные кавычки для ДРУГОГО</code></pre>

<p><strong>2. Каждая команда на новой строке:</strong></p>
<pre><code>fmt.Println("Первая строка")
fmt.Println("Вторая строка")   // каждый Println — отдельная строка</code></pre>

<p><strong>3. Фигурные скобки { } обязательны:</strong></p>
<pre><code>func main() {         // открывающая скобка — на той же строке
    // код здесь
}                     // закрывающая скобка — на отдельной строке</code></pre>

<p><strong>4. fmt.Println — точка, не запятая:</strong></p>
<pre><code>fmt.Println("текст")   // ✅ точка между fmt и Println
fmt,Println("текст")   // ❌ запятая — ошибка!</code></pre>

<h2>Несколько строк вывода</h2>
<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Строка 1")
    fmt.Println("Строка 2")
    fmt.Println("Строка 3")
}
// Вывод:
// Строка 1
// Строка 2
// Строка 3</code></pre>

<p>Каждый <code>fmt.Println</code> выводит текст и <strong>переходит на новую строку</strong>. Println = Print + Line (line = строка).</p>

<h2>Комментарии — заметки для людей</h2>
<pre><code>// Это комментарий — Go его ИГНОРИРУЕТ
// Пиши комментарии чтобы объяснить что делает код

fmt.Println("Это выполнится")  // можно после кода
// fmt.Println("А это НЕТ — закомментировано")</code></pre>

<h2>Частые ошибки новичков</h2>
<pre><code>// ОШИБКА 1: забыл кавычки
fmt.Println(Привет)    // ❌ Go думает что Привет — это переменная

// ОШИБКА 2: забыл import
func main() {
    fmt.Println("Привет")  // ❌ undefined: fmt — забыл import "fmt"
}

// ОШИБКА 3: Println с маленькой буквы
fmt.println("Привет")  // ❌ в Go заглавная буква = доступно снаружи</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что выведет fmt.Println(\"Go\")?",
				Options:     []string{"fmt.Println(\"Go\")", "Go", "Ошибку", "Ничего"},
				Correct:     1,
				Explanation: "fmt.Println выводит текст который в кавычках. \"Go\" → на экране появится Go (без кавычек).",
			},
			{
				Question:    "Что означает import \"fmt\"?",
				Options:     []string{"Создаёт переменную", "Подключает пакет fmt — набор инструментов для вывода текста на экран", "Запускает программу", "Скачивает файл"},
				Correct:     1,
				Explanation: "import подключает пакет. fmt (format) — стандартный пакет Go для ввода/вывода текста. Без import \"fmt\" использовать fmt.Println нельзя.",
			},
			{
				Question:    "Почему текст пишут в двойных кавычках \"...\"?",
				Options:     []string{"Для красоты", "Так Go отличает текст (строку) от команд и переменных", "Можно и без них", "Одинарные тоже подходят"},
				Correct:     1,
				Explanation: "Кавычки говорят Go: это текст (строка), не код. Без кавычек Go пытается найти переменную с таким именем. Двойные — для строк, одинарные — для одного символа (rune).",
			},
			{
				Question:    "Что не так в коде: fmt.println(\"Привет\")?",
				Options:     []string{"Всё правильно", "println с маленькой буквы — нужно Println с большой", "Не хватает точки с запятой", "Лишние кавычки"},
				Correct:     1,
				Explanation: "В Go заглавная первая буква = экспортированная функция (доступна снаружи пакета). println с маленькой — не существует. Правильно: fmt.Println.",
			},
			{
				Question:    "Зачем нужна строка package main?",
				Options:     []string{"Это комментарий", "Говорит Go что это запускаемая программа, а не библиотека", "Подключает пакет", "Создаёт папку"},
				Correct:     1,
				Explanation: "package main + func main() = точка входа. Go знает: эту программу можно запустить. Другие пакеты (не main) — это библиотеки, которые используются в других программах.",
			},
		},
		Tasks: []T{
			{
				Title:      "Привет, мир!",
				Difficulty: "easy",
				Description: `<p>Выведи на экран текст <code>Hello, Go!</code></p>
<p>Подсказка: используй <code>fmt.Println</code> — функцию для вывода текста. Текст пиши в двойных кавычках.</p>`,
				Glossary: []GlossaryItem{
					{Term: "fmt.Println(\"текст\")", Definition: "Выводит текст на экран и переходит на новую строку. Текст обязательно в двойных кавычках."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Hello, Go!"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    // Напиши здесь: fmt.Println("Hello, Go!")
}`,
				Hints:    `<p>Замени комментарий на <code>fmt.Println("Hello, Go!")</code>. Не забудь кавычки вокруг текста!</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Hello, Go!")
}</code></pre>`,
			},
			{
				Title:      "Три строки",
				Difficulty: "easy",
				Description: `<p>Выведи три строки текста — каждую на отдельной строке:</p>
<pre><code>Я учу Go
Это мой первый урок
Уже получается!</code></pre>
<p>Для каждой строки нужен отдельный <code>fmt.Println</code>.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Несколько Println", Definition: "Каждый fmt.Println выводит текст и переходит на новую строку. Три Println = три строки на экране."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Я учу Go\nЭто мой первый урок\nУже получается!"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    fmt.Println("Я учу Go")
    // Добавь ещё две строки по примеру первой
}`,
				Hints:    `<p>Скопируй строку <code>fmt.Println("Я учу Go")</code> и замени текст в кавычках на нужный.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Я учу Go")
    fmt.Println("Это мой первый урок")
    fmt.Println("Уже получается!")
}</code></pre>`,
			},
			{
				Title:      "Визитка",
				Difficulty: "easy",
				Description: `<p>Выведи свою визитку (используй эти данные):</p>
<pre><code>Имя: Алексей
Роль: Go-разработчик
Город: Москва</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Текст с двоеточием", Definition: "В кавычках можно писать любые символы: \"Имя: Алексей\" — двоеточие и пробел это часть текста."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Имя: Алексей\nРоль: Go-разработчик\nГород: Москва"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    fmt.Println("Имя: Алексей")
    fmt.Println("Роль: Go-разработчик")
    // Добавь третью строку: Город: Москва
}`,
				Hints:    `<p>Добавь <code>fmt.Println("Город: Москва")</code> после второй строки.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Имя: Алексей")
    fmt.Println("Роль: Go-разработчик")
    fmt.Println("Город: Москва")
}</code></pre>`,
			},
			{
				Title:      "Баннер",
				Difficulty: "medium",
				Description: `<p>Выведи баннер с рамкой из символов <code>=</code>:</p>
<pre><code>====================
  GoLearn Course
====================</code></pre>
<p>Обрати внимание: текст "GoLearn Course" начинается с двух пробелов.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Пробелы в строке", Definition: "Пробелы внутри кавычек — часть текста. \"  GoLearn\" начинается с двух пробелов."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "====================\n  GoLearn Course\n===================="},
				},
				StarterCode: `package main

import "fmt"

func main() {
    fmt.Println("====================")
    fmt.Println("  GoLearn Course")
    // Добавь закрывающую линию из = (как первая)
}`,
				Hints:    `<p>Скопируй первую строку <code>fmt.Println("====================")</code> и вставь в конец.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    fmt.Println("====================")
    fmt.Println("  GoLearn Course")
    fmt.Println("====================")
}</code></pre>`,
			},
			{
				Title:      "Найди ошибку",
				Difficulty: "hard",
				Description: `<p>В этом коде <strong>3 ошибки</strong>. Найди и исправь их, чтобы программа вывела:</p>
<pre><code>Go — это просто!
Учись каждый день
Ты справишься!</code></pre>
<p>Подсказка: ошибки в синтаксисе — кавычки, регистр букв, import.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Ошибки компиляции", Definition: "Go проверяет код ПЕРЕД запуском. Если есть ошибка — программа не запустится. Прочитай сообщение об ошибке — оно подсказывает где проблема."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Go — это просто!\nУчись каждый день\nТы справишься!"},
				},
				StarterCode: `package main

func main() {
    fmt.Println("Go — это просто!")
    fmt.println("Учись каждый день")
    fmt.Println(Ты справишься!)
}`,
				Hints:    `<p>1) Нет <code>import "fmt"</code> 2) <code>println</code> → <code>Println</code> (заглавная P) 3) <code>Ты справишься!</code> без кавычек → добавь <code>"Ты справишься!"</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    fmt.Println("Go — это просто!")
    fmt.Println("Учись каждый день")
    fmt.Println("Ты справишься!")
}</code></pre>`,
			},
		},
	}
}

// ── Урок 2: Первая программа ─────────────────────────────────

func lesson02_first_program() L {
	return L{
		Slug: "variables", Title: "Переменные — коробки для данных", Order: 2,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Переменные — коробки для данных</h1>

<h2>Что такое переменная?</h2>
<p>Представь коробку с наклейкой. На наклейке написано имя, а внутри — значение:</p>

<pre><code>package main

import "fmt"

func main() {
    name := "Алексей"        // создаём коробку "name", кладём туда "Алексей"
    age := 25                // создаём коробку "age", кладём туда 25
    fmt.Println(name)        // достаём из коробки и выводим: Алексей
    fmt.Println(age)         // выводим: 25
}</code></pre>

<p>Результат:</p>
<pre><code>Алексей
25</code></pre>

<h2>Как создать переменную</h2>
<p>В Go самый частый способ — оператор <code>:=</code> (короткое объявление):</p>

<pre><code>имя := значение</code></pre>

<p>Go сам определяет тип по значению:</p>
<pre><code>name := "Go"       // текст (строка) — тип string
age := 25          // целое число — тип int
price := 9.99      // дробное число — тип float64
active := true     // да/нет — тип bool</code></pre>

<h2>Вывод переменных</h2>
<pre><code>name := "Go"
version := 1.22

// Способ 1: просто вывести
fmt.Println(name)              // Go

// Способ 2: вывести несколько через запятую
fmt.Println("Язык:", name)     // Язык: Go

// Способ 3: вывести несколько переменных
fmt.Println(name, version)     // Go 1.22</code></pre>

<p><strong>Важно:</strong> <code>fmt.Println</code> сам ставит пробелы между аргументами.</p>

<h2>Изменение переменной</h2>
<pre><code>score := 0         // создали переменную
fmt.Println(score)  // 0

score = 10          // ИЗМЕНИЛИ значение (без двоеточия!)
fmt.Println(score)  // 10

score = score + 5   // прибавили 5
fmt.Println(score)  // 15</code></pre>

<p><strong>Правило:</strong> <code>:=</code> — создать НОВУЮ переменную. <code>=</code> — изменить СУЩЕСТВУЮЩУЮ.</p>

<h2>Правила именования</h2>
<pre><code>// ✅ Хорошие имена — понятно что внутри
userName := "Alice"
videoCount := 42
isActive := true

// ❌ Плохие имена — непонятно
x := "Alice"
n := 42
flag := true</code></pre>

<p>В Go принято писать имена в <strong>camelCase</strong>: первое слово с маленькой буквы, каждое следующее — с большой: <code>userName</code>, <code>videoCount</code>, <code>isActive</code>.</p>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА 1: использовать := для уже существующей переменной
name := "Alice"
name := "Bob"      // ❌ "name already declared" — используй = без двоеточия

// ОШИБКА 2: создать переменную и не использовать
unused := 42       // ❌ "unused declared but not used" — Go не разрешает мусор

// ОШИБКА 3: вывести имя переменной вместо значения
name := "Alice"
fmt.Println("name")  // выведет: name (текст в кавычках!)
fmt.Println(name)    // выведет: Alice (значение переменной!)</code></pre>

<p><strong>Запомни:</strong> <code>"name"</code> в кавычках — это текст. <code>name</code> без кавычек — это переменная.</p>`,

		Quiz: []Q{
			{
				Question:    "Что выведет код: x := 5; fmt.Println(x)?",
				Options:     []string{"x", "5", "x := 5", "Ошибку"},
				Correct:     1,
				Explanation: "x без кавычек — это переменная. fmt.Println(x) выводит ЗНАЧЕНИЕ переменной, то есть 5. Если бы было fmt.Println(\"x\") — вывело бы букву x.",
			},
			{
				Question:    "Чем := отличается от =?",
				Options:     []string{"Ничем", ":= создаёт НОВУЮ переменную, = изменяет СУЩЕСТВУЮЩУЮ", ":= для строк, = для чисел", ":= медленнее"},
				Correct:     1,
				Explanation: "name := \"Alice\" — создать новую переменную name. name = \"Bob\" — изменить уже существующую. Если использовать := повторно для той же переменной — ошибка.",
			},
			{
				Question:    "Что не так: age := 25; age := 30?",
				Options:     []string{"Всё правильно", "Ошибка: age уже объявлена. Нужно age = 30 (без двоеточия)", "Нельзя менять числа", "Нужны кавычки"},
				Correct:     1,
				Explanation: ":= создаёт НОВУЮ переменную. Второй раз создать age нельзя — она уже есть. Для изменения используй = (без двоеточия): age = 30.",
			},
			{
				Question:    "Что произойдёт если создать переменную и не использовать?",
				Options:     []string{"Ничего", "Ошибка компиляции: declared but not used", "Предупреждение", "Переменная удалится"},
				Correct:     1,
				Explanation: "Go строгий: неиспользованная переменная = ошибка. Это заставляет держать код чистым. Если переменная не нужна — удали её.",
			},
			{
				Question:    "Чем отличается fmt.Println(\"name\") от fmt.Println(name)?",
				Options:     []string{"Ничем", "\"name\" выведет текст name, а name выведет ЗНАЧЕНИЕ переменной name", "Первое быстрее", "Второе вызовет ошибку"},
				Correct:     1,
				Explanation: "В кавычках — это текст (строковый литерал). Без кавычек — это имя переменной. fmt.Println(\"name\") → name. fmt.Println(name) → Alice (если name = \"Alice\").",
			},
		},
		Tasks: []T{
			{
				Title: "Создай переменную",
				Difficulty: "easy",
				Description: `<p>Создай переменную <code>language</code> со значением <code>"Go"</code> и выведи её:</p>
<pre><code>Go</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: ":= (короткое объявление)", Definition: "Создаёт новую переменную: name := \"значение\". Go сам определит тип."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Go"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    // Создай переменную language со значением "Go"
    // и выведи её через fmt.Println
}`,
				Hints:    `<p><code>language := "Go"</code> — создаёт переменную. <code>fmt.Println(language)</code> — выводит её значение.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    language := "Go"
    fmt.Println(language)
}</code></pre>`,
			},
			{
				Title: "Визитка с переменными",
				Difficulty: "easy",
				Description: `<p>Создай три переменные и выведи визитку:</p>
<pre><code>Имя: Алексей
Возраст: 25
Город: Москва</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "fmt.Println(\"текст\", переменная)", Definition: "Можно выводить текст и переменную вместе: fmt.Println(\"Имя:\", name) → Имя: Алексей"},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Имя: Алексей\nВозраст: 25\nГород: Москва"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    name := "Алексей"
    age := 25
    city := "Москва"

    fmt.Println("Имя:", name)
    // Добавь вывод возраста и города по тому же примеру
}`,
				Hints:    `<p><code>fmt.Println("Возраст:", age)</code> и <code>fmt.Println("Город:", city)</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    name := "Алексей"
    age := 25
    city := "Москва"

    fmt.Println("Имя:", name)
    fmt.Println("Возраст:", age)
    fmt.Println("Город:", city)
}</code></pre>`,
			},
			{
				Title: "Счётчик",
				Difficulty: "easy",
				Description: `<p>Создай переменную <code>count</code> = 0, увеличь её три раза на 1 и выведи результат:</p>
<pre><code>3</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "count = count + 1", Definition: "Изменяет переменную: берёт текущее значение, прибавляет 1, записывает обратно. Короткая форма: count++"},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "3"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    count := 0
    count = count + 1  // теперь count = 1
    count = count + 1  // теперь count = 2
    // Увеличь count ещё на 1
    fmt.Println(count)
}`,
				Hints:    `<p>Добавь ещё одну строку <code>count = count + 1</code> перед fmt.Println.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    count := 0
    count = count + 1
    count = count + 1
    count = count + 1
    fmt.Println(count)
}</code></pre>`,
			},
			{
				Title: "Обмен значений",
				Difficulty: "medium",
				Description: `<p>Поменяй значения двух переменных местами и выведи результат:</p>
<pre><code>a = 20
b = 10</code></pre>
<p>Подсказка: тебе понадобится третья переменная (временная).</p>`,
				Glossary: []GlossaryItem{
					{Term: "Временная переменная", Definition: "Для обмена: tmp := a; a = b; b = tmp. Без tmp одно значение потеряется."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "a = 20\nb = 10"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    a := 10
    b := 20

    // Поменяй a и b местами
    // Подсказка: сохрани a во временную переменную tmp

    fmt.Println("a =", a)
    fmt.Println("b =", b)
}`,
				Hints:    `<p><code>tmp := a</code> — сохранить. <code>a = b</code> — перезаписать. <code>b = tmp</code> — восстановить.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    a := 10
    b := 20

    tmp := a
    a = b
    b = tmp

    fmt.Println("a =", a)
    fmt.Println("b =", b)
}</code></pre>`,
			},
			{
				Title: "Найди ошибки в переменных",
				Difficulty: "hard",
				Description: `<p>В коде <strong>3 ошибки</strong> с переменными. Исправь их чтобы вывести:</p>
<pre><code>Привет, Go!
Версия: 1.22
Год: 2024</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Типичные ошибки", Definition: "1) := вместо = для существующей. 2) Неиспользованная переменная. 3) Кавычки вместо переменной."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Привет, Go!\nВерсия: 1.22\nГод: 2024"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    greeting := "Привет, Go!"
    version := 1.22
    year := 2024
    unused := "ненужное"   // Ошибка 1: переменная не используется

    fmt.Println(greeting)
    fmt.Println("Версия:", "version")  // Ошибка 2: version в кавычках
    year := 2025                       // Ошибка 3: := вместо =
    fmt.Println("Год:", year)

    _ = unused  // Убери эту строку и удали переменную unused
}`,
				Hints:    `<p>1) Удали <code>unused</code> и <code>_ = unused</code>. 2) Убери кавычки вокруг version. 3) Замени <code>:=</code> на <code>=</code> для year.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    greeting := "Привет, Go!"
    version := 1.22
    year := 2024

    fmt.Println(greeting)
    fmt.Println("Версия:", version)
    year = 2024
    fmt.Println("Год:", year)
}</code></pre>`,
			},
		},
	}
}


// ── Урок 3: Типы данных и форматирование ─────────────────────

func lesson03_variables() L {
	return L{
		Slug: "types-and-format", Title: "Типы данных и форматирование", Order: 3,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Типы данных и форматирование</h1>

<h2>Четыре основных типа</h2>
<p>Каждая переменная в Go имеет <strong>тип</strong> — он определяет что можно хранить и что делать:</p>

<pre><code>name := "Алексей"    // string — текст (строка)
age := 25            // int — целое число
price := 9.99        // float64 — дробное число
active := true       // bool — да (true) или нет (false)</code></pre>

<table>
<tr><th>Тип</th><th>Что хранит</th><th>Примеры</th></tr>
<tr><td><code>string</code></td><td>Текст в кавычках</td><td>"Привет", "Go", ""</td></tr>
<tr><td><code>int</code></td><td>Целые числа</td><td>0, 42, -100</td></tr>
<tr><td><code>float64</code></td><td>Дробные числа</td><td>3.14, 9.99, 0.0</td></tr>
<tr><td><code>bool</code></td><td>Истина или ложь</td><td>true, false</td></tr>
</table>

<h2>Зачем нужны типы?</h2>
<pre><code>// Go не даст сложить строку и число — это ошибка
name := "Go"
age := 25
// result := name + age  // ❌ ошибка: нельзя сложить string и int

// Числа можно складывать
total := 10 + 20    // 30
price := 9.99 + 0.01  // 10.0</code></pre>

<h2>fmt.Printf — форматированный вывод</h2>
<p>Когда нужно вставить переменную в середину текста — используй <code>fmt.Printf</code>:</p>

<pre><code>name := "Алексей"
age := 25

fmt.Printf("Привет, %s! Тебе %d лет.\n", name, age)
// Вывод: Привет, Алексей! Тебе 25 лет.</code></pre>

<p><strong>Заполнители</strong> (подставляют значение):</p>
<table>
<tr><th>Заполнитель</th><th>Для какого типа</th><th>Пример</th></tr>
<tr><td><code>%s</code></td><td>string (текст)</td><td>"Алексей" → Алексей</td></tr>
<tr><td><code>%d</code></td><td>int (целое число)</td><td>25 → 25</td></tr>
<tr><td><code>%f</code></td><td>float64 (дробное)</td><td>9.99 → 9.990000</td></tr>
<tr><td><code>%.2f</code></td><td>float64 (2 знака)</td><td>9.99 → 9.99</td></tr>
<tr><td><code>%v</code></td><td>любой тип</td><td>что угодно</td></tr>
</table>

<p><strong>Важно:</strong> <code>\n</code> в конце — перенос строки. Printf НЕ добавляет его автоматически (в отличие от Println).</p>

<h2>Арифметика</h2>
<pre><code>a := 10
b := 3

fmt.Println(a + b)    // 13 — сложение
fmt.Println(a - b)    // 7  — вычитание
fmt.Println(a * b)    // 30 — умножение
fmt.Println(a / b)    // 3  — деление (целочисленное! не 3.33)
fmt.Println(a % b)    // 1  — остаток от деления</code></pre>

<p><strong>Ловушка:</strong> <code>10 / 3 = 3</code> (не 3.33!) — целое делим на целое, результат целый. Для дробного: <code>10.0 / 3.0 = 3.333...</code></p>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: перепутал %s и %d
fmt.Printf("Возраст: %s\n", 25)   // ❌ %s для строк, 25 — число

// ОШИБКА: забыл \n
fmt.Printf("Строка 1")
fmt.Printf("Строка 2")
// Вывод: Строка 1Строка 2  ← всё на одной строке!

// ПРАВИЛЬНО:
fmt.Printf("Строка 1\n")
fmt.Printf("Строка 2\n")</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Какой тип у переменной price := 9.99?",
				Options:     []string{"int", "string", "float64", "bool"},
				Correct:     2,
				Explanation: "9.99 — число с точкой (дробное). Go автоматически определяет тип как float64. Целое число (без точки) было бы int.",
			},
			{
				Question:    "Что выведет: fmt.Printf(\"%s — %d лет\\n\", \"Go\", 15)?",
				Options:     []string{"%s — %d лет", "Go — 15 лет", "Ошибку", "Go — Go лет"},
				Correct:     1,
				Explanation: "%s заменяется на \"Go\", %d на 15. \\n переводит на новую строку. Результат: Go — 15 лет",
			},
			{
				Question:    "Что выведет: fmt.Println(10 / 3)?",
				Options:     []string{"3.33", "3", "3.0", "Ошибку"},
				Correct:     1,
				Explanation: "10 и 3 — целые числа (int). Деление целых = целый результат: 10/3 = 3 (остаток отбрасывается). Для 3.33 нужно: 10.0 / 3.0",
			},
			{
				Question:    "Зачем нужен \\n в fmt.Printf?",
				Options:     []string{"Для красоты", "Перенос строки — Printf не добавляет его автоматически", "Это пробел", "Завершает программу"},
				Correct:     1,
				Explanation: "Println добавляет \\n автоматически. Printf — нет. Без \\n следующий вывод будет на той же строке.",
			},
			{
				Question:    "Что не так: fmt.Printf(\"Возраст: %s\\n\", 25)?",
				Options:     []string{"Всё правильно", "%s для строк, а 25 — число. Нужен %d", "Лишний \\n", "Не хватает кавычек"},
				Correct:     1,
				Explanation: "%s ожидает строку (string), а 25 — число (int). Правильно: %d для чисел. Go может вывести что-то, но это неправильное использование.",
			},
		},
		Tasks: []T{
			{
				Title: "Калькулятор", Difficulty: "easy",
				Description: `<p>Создай две переменные <code>a = 15</code> и <code>b = 4</code>, выведи результат всех операций:</p>
<pre><code>15 + 4 = 19
15 - 4 = 11
15 * 4 = 60
15 / 4 = 3
15 % 4 = 3</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "% (остаток)", Definition: "15 % 4 = 3, потому что 15 = 4*3 + 3. Остаток от деления."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "15 + 4 = 19\n15 - 4 = 11\n15 * 4 = 60\n15 / 4 = 3\n15 % 4 = 3"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    a := 15
    b := 4

    fmt.Println(a, "+", b, "=", a+b)
    fmt.Println(a, "-", b, "=", a-b)
    // Допиши умножение, деление и остаток
}`,
				Hints:    `<p><code>fmt.Println(a, "*", b, "=", a*b)</code> — по аналогии для *, / и %.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    a := 15
    b := 4
    fmt.Println(a, "+", b, "=", a+b)
    fmt.Println(a, "-", b, "=", a-b)
    fmt.Println(a, "*", b, "=", a*b)
    fmt.Println(a, "/", b, "=", a/b)
    fmt.Println(a, "%", b, "=", a%b)
}</code></pre>`,
			},
			{
				Title: "Профиль пользователя", Difficulty: "easy",
				Description: `<p>Используй <code>fmt.Printf</code> для вывода профиля:</p>
<pre><code>Имя: Alice
Возраст: 25
Баланс: $99.50</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "%.2f", Definition: "Выводит дробное число с 2 знаками после точки: 99.5 → 99.50"},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Имя: Alice\nВозраст: 25\nБаланс: $99.50"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    name := "Alice"
    age := 25
    balance := 99.50

    fmt.Printf("Имя: %s\n", name)
    fmt.Printf("Возраст: %d\n", age)
    // Допиши вывод баланса с %.2f
}`,
				Hints:    `<p><code>fmt.Printf("Баланс: $%.2f\n", balance)</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    name := "Alice"
    age := 25
    balance := 99.50
    fmt.Printf("Имя: %s\n", name)
    fmt.Printf("Возраст: %d\n", age)
    fmt.Printf("Баланс: $%.2f\n", balance)
}</code></pre>`,
			},
			{
				Title: "Чек магазина", Difficulty: "medium",
				Description: `<p>Рассчитай стоимость покупки:</p>
<pre><code>Товар: Наушники
Цена: $49.99
Количество: 3
Итого: $149.97</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "float64 * int", Definition: "Нельзя напрямую умножить float64 на int. Нужно: price * float64(quantity)"},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Товар: Наушники\nЦена: $49.99\nКоличество: 3\nИтого: $149.97"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    item := "Наушники"
    price := 49.99
    quantity := 3

    total := price * float64(quantity)

    fmt.Printf("Товар: %s\n", item)
    fmt.Printf("Цена: $%.2f\n", price)
    fmt.Printf("Количество: %d\n", quantity)
    // Допиши вывод итого
}`,
				Hints:    `<p><code>fmt.Printf("Итого: $%.2f\n", total)</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    item := "Наушники"
    price := 49.99
    quantity := 3
    total := price * float64(quantity)
    fmt.Printf("Товар: %s\n", item)
    fmt.Printf("Цена: $%.2f\n", price)
    fmt.Printf("Количество: %d\n", quantity)
    fmt.Printf("Итого: $%.2f\n", total)
}</code></pre>`,
			},
			{
				Title: "Конвертер температуры", Difficulty: "medium",
				Description: `<p>Переведи температуру из Цельсия в Фаренгейт. Формула: F = C * 9/5 + 32</p>
<pre><code>100°C = 212.00°F</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Дробное деление", Definition: "9/5 = 1 (целые!). Нужно 9.0/5.0 = 1.8 чтобы получить дробный результат."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "100°C = 212.00°F"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    celsius := 100.0
    fahrenheit := celsius * 9.0/5.0 + 32.0
    fmt.Printf("%.0f°C = %.2f°F\n", celsius, fahrenheit)
}`,
				Hints:    `<p>Формула уже написана. Убедись что используешь 9.0/5.0 (не 9/5).</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    celsius := 100.0
    fahrenheit := celsius*9.0/5.0 + 32.0
    fmt.Printf("%.0f°C = %.2f°F\n", celsius, fahrenheit)
}</code></pre>`,
			},
			{
				Title: "Отладка типов", Difficulty: "hard",
				Description: `<p>Исправь 3 ошибки связанные с типами и форматированием:</p>
<pre><code>Пользователь: Bob (30 лет)
Баланс: $150.75
Активен: true</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "%v", Definition: "Универсальный заполнитель — подставит любой тип. Используй когда не уверен."},
				},
				TestCases: []TestCase{
					{Input: "", ExpectedOutput: "Пользователь: Bob (30 лет)\nБаланс: $150.75\nАктивен: true"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    name := "Bob"
    age := 30
    balance := 150.75
    active := true

    fmt.Printf("Пользователь: %d (%s лет)\n", name, age)  // Ошибка 1: перепутаны %d и %s
    fmt.Printf("Баланс: $%d\n", balance)                    // Ошибка 2: %d для float64
    fmt.Printf("Активен: %s\n", active)                     // Ошибка 3: %s для bool
}`,
				Hints:    `<p>1) name — строка (%s), age — число (%d). 2) balance — дробное (%.2f). 3) active — bool (%v или %t).</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    name := "Bob"
    age := 30
    balance := 150.75
    active := true

    fmt.Printf("Пользователь: %s (%d лет)\n", name, age)
    fmt.Printf("Баланс: $%.2f\n", balance)
    fmt.Printf("Активен: %v\n", active)
}</code></pre>`,
			},
		},
	}
}

// ── Урок 4: Ввод от пользователя ────────────────────────────────

func lesson04_types_and_conversions() L {
	return L{
		Slug: "user-input", Title: "Ввод от пользователя", Order: 4,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Ввод от пользователя</h1>

<h2>fmt.Scan — читаем данные</h2>
<p>До сих пор мы только выводили текст. Теперь научимся <strong>получать данные от пользователя</strong>:</p>

<pre><code>package main

import "fmt"

func main() {
    var name string           // создаём переменную через var
    fmt.Println("Как тебя зовут?")
    fmt.Scan(&name)           // ждём ввод от пользователя
    fmt.Println("Привет,", name)
}</code></pre>

<p>Запускаем, вводим "Алексей" → получаем:</p>
<pre><code>Как тебя зовут?
Привет, Алексей</code></pre>

<h2>var — другой способ создать переменную</h2>
<p>Раньше мы писали <code>name := "Go"</code>. Но для <code>fmt.Scan</code> нужна <strong>пустая переменная</strong> — создаём через <code>var</code>:</p>

<pre><code>var name string    // пустая строка ""
var age int        // ноль: 0
var price float64  // ноль: 0.0

// var создаёт переменную с "нулевым значением"
// := создаёт сразу с конкретным значением
name := "Go"      // сразу есть значение</code></pre>

<h2>Почему &name а не просто name?</h2>
<p>Символ <code>&</code> перед переменной означает "запиши значение СЮДА". Просто запомни: <strong>fmt.Scan всегда с &</strong>. Подробно объясним в уроке про указатели.</p>

<pre><code>var age int
fmt.Scan(&age)     // ✅ правильно — с &
fmt.Scan(age)      // ❌ ошибка — без & не работает</code></pre>

<h2>Читаем несколько значений</h2>
<pre><code>var name string
var age int
fmt.Scan(&name, &age)  // читает два значения через пробел

// Пользователь вводит: Alice 25
// name = "Alice", age = 25</code></pre>

<h2>Пример: простой калькулятор</h2>
<pre><code>package main

import "fmt"

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Println(a + b)
}
// Ввод: 10 20
// Вывод: 30</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА 1: забыл &
var n int
fmt.Scan(n)        // ❌ нужно fmt.Scan(&n)

// ОШИБКА 2: неправильный тип
var age int
// Пользователь ввёл "двадцать" → age = 0 (Go не может преобразовать)

// ОШИБКА 3: не создал переменную
fmt.Scan(&name)    // ❌ undefined: name — сначала var name string</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Зачем перед переменной стоит & в fmt.Scan(&name)?",
				Options:     []string{"Для красоты", "Означает 'запиши значение сюда' — передаёт адрес переменной", "Это умножение", "Это логическое И"},
				Correct:     1,
				Explanation: "& передаёт адрес переменной. fmt.Scan записывает введённое значение по этому адресу. Без & — Scan не знает КУДА записать.",
			},
			{
				Question:    "Чем var name string отличается от name := \"Go\"?",
				Options:     []string{"Ничем", "var создаёт пустую переменную (нулевое значение), := создаёт с конкретным значением", "var медленнее", "var только для строк"},
				Correct:     1,
				Explanation: "var name string → name = \"\" (пустая строка). name := \"Go\" → name = \"Go\" (сразу с значением). var нужен когда значение придёт позже (из Scan, из файла).",
			},
			{
				Question:    "Что произойдёт если ввести текст когда ожидается число? var n int; fmt.Scan(&n)",
				Options:     []string{"Ошибка компиляции", "n останется 0 — Go не может преобразовать текст в число", "Программа упадёт", "Текст станет числом"},
				Correct:     1,
				Explanation: "Scan молча вернёт ошибку и оставит n = 0. Программа продолжит работу, но с неправильным значением. В реальном коде нужно проверять ошибку.",
			},
			{
				Question:    "Сколько значений прочитает fmt.Scan(&a, &b)?",
				Options:     []string{"Одно", "Два — разделённые пробелом или переносом строки", "Зависит от ввода", "Ноль"},
				Correct:     1,
				Explanation: "Scan читает столько значений, сколько аргументов. &a, &b = два аргумента = два значения. Разделитель: пробел, табуляция или перенос строки.",
			},
			{
				Question:    "Какое нулевое значение у var count int?",
				Options:     []string{"nil", "0", "\"\"", "false"},
				Correct:     1,
				Explanation: "int → 0, string → \"\" (пустая строка), bool → false, float64 → 0.0. Go всегда инициализирует переменные нулевым значением.",
			},
		},
		Tasks: []T{
			{
				Title: "Приветствие", Difficulty: "easy",
				Description: `<p>Прочитай имя пользователя и поприветствуй его:</p>
<p>Ввод: <code>Алексей</code></p>
<p>Вывод: <code>Привет, Алексей!</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "fmt.Scan(&name)", Definition: "Читает текст от пользователя и записывает в переменную name. & обязателен."},
				},
				TestCases: []TestCase{
					{Input: "Алексей", ExpectedOutput: "Привет, Алексей!"},
					{Input: "Go", ExpectedOutput: "Привет, Go!"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var name string
    fmt.Scan(&name)
    // Выведи: Привет, <имя>!
    // Подсказка: fmt.Printf("Привет, %s!\n", name)
}`,
				Hints:    `<p><code>fmt.Printf("Привет, %s!\n", name)</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var name string
    fmt.Scan(&name)
    fmt.Printf("Привет, %s!\n", name)
}</code></pre>`,
			},
			{
				Title: "Сумма двух чисел", Difficulty: "easy",
				Description: `<p>Прочитай два числа и выведи их сумму:</p>
<p>Ввод: <code>10 20</code></p>
<p>Вывод: <code>30</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "var a, b int", Definition: "Можно объявить несколько переменных одного типа через запятую."},
				},
				TestCases: []TestCase{
					{Input: "10 20", ExpectedOutput: "30"},
					{Input: "100 200", ExpectedOutput: "300"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    // Выведи сумму a + b
}`,
				Hints:    `<p><code>fmt.Println(a + b)</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var a, b int
    fmt.Scan(&a, &b)
    fmt.Println(a + b)
}</code></pre>`,
			},
			{
				Title: "Площадь прямоугольника", Difficulty: "easy",
				Description: `<p>Прочитай ширину и высоту, выведи площадь:</p>
<p>Ввод: <code>5 3</code></p>
<p>Вывод: <code>Площадь: 15</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "a * b", Definition: "Умножение. Площадь прямоугольника = ширина * высота."},
				},
				TestCases: []TestCase{
					{Input: "5 3", ExpectedOutput: "Площадь: 15"},
					{Input: "10 10", ExpectedOutput: "Площадь: 100"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var width, height int
    fmt.Scan(&width, &height)
    area := width * height
    fmt.Println("Площадь:", area)
}`,
				Hints:    `<p>Код уже написан! Просто запусти его.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var width, height int
    fmt.Scan(&width, &height)
    fmt.Println("Площадь:", width*height)
}</code></pre>`,
			},
			{
				Title: "Конвертер валют", Difficulty: "medium",
				Description: `<p>Прочитай сумму в рублях и курс доллара. Выведи сумму в долларах:</p>
<p>Ввод: <code>10000 90.5</code></p>
<p>Вывод: <code>$110.50</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "float64 деление", Definition: "10000.0 / 90.5 = 110.497... С %.2f выведет 110.50"},
				},
				TestCases: []TestCase{
					{Input: "10000 90.5", ExpectedOutput: "$110.50"},
					{Input: "5000 100.0", ExpectedOutput: "$50.00"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var rubles, rate float64
    fmt.Scan(&rubles, &rate)
    dollars := rubles / rate
    fmt.Printf("$%.2f\n", dollars)
}`,
				Hints:    `<p>Код уже готов. Изучи как работает <code>%.2f</code> — два знака после точки.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var rubles, rate float64
    fmt.Scan(&rubles, &rate)
    fmt.Printf("$%.2f\n", rubles/rate)
}</code></pre>`,
			},
			{
				Title: "Визитка из ввода", Difficulty: "hard",
				Description: `<p>Прочитай имя, возраст и город. Выведи красивую визитку:</p>
<p>Ввод: <code>Alice 25 Moscow</code></p>
<p>Вывод:</p>
<pre><code>====================
  Name: Alice
  Age: 25
  City: Moscow
====================</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Комбинация Scan + Printf", Definition: "Scan читает данные, Printf красиво выводит. Стандартный паттерн: ввод → обработка → вывод."},
				},
				TestCases: []TestCase{
					{Input: "Alice 25 Moscow", ExpectedOutput: "====================\n  Name: Alice\n  Age: 25\n  City: Moscow\n===================="},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var name string
    var age int
    var city string
    fmt.Scan(&name, &age, &city)

    fmt.Println("====================")
    fmt.Printf("  Name: %s\n", name)
    // Допиши вывод Age и City
    // Потом закрывающую линию ====================
}`,
				Hints:    `<p><code>fmt.Printf("  Age: %d\n", age)</code> и <code>fmt.Printf("  City: %s\n", city)</code>. В конце <code>fmt.Println("====================")</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var name string
    var age int
    var city string
    fmt.Scan(&name, &age, &city)
    fmt.Println("====================")
    fmt.Printf("  Name: %s\n", name)
    fmt.Printf("  Age: %d\n", age)
    fmt.Printf("  City: %s\n", city)
    fmt.Println("====================")
}</code></pre>`,
			},
		},
	}
}

// ── Урок 5: Условия ────────────────────────────────────────────

func lesson05_input_output() L {
	return L{
		Slug: "conditions", Title: "Условия — принимаем решения", Order: 5,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Условия — принимаем решения</h1>

<h2>if — если условие верно</h2>
<p>Программа может делать разные вещи в зависимости от данных:</p>

<pre><code>age := 18

if age >= 18 {
    fmt.Println("Доступ разрешён")
}
// Вывод: Доступ разрешён</code></pre>

<p><strong>Как читать:</strong> "ЕСЛИ age больше или равно 18 — выполни код в фигурных скобках".</p>

<h2>if — else: иначе</h2>
<pre><code>age := 15

if age >= 18 {
    fmt.Println("Доступ разрешён")
} else {
    fmt.Println("Доступ запрещён")
}
// Вывод: Доступ запрещён</code></pre>

<h2>if — else if — else: несколько условий</h2>
<pre><code>score := 75

if score >= 90 {
    fmt.Println("Отлично")
} else if score >= 70 {
    fmt.Println("Хорошо")
} else if score >= 50 {
    fmt.Println("Удовлетворительно")
} else {
    fmt.Println("Неудовлетворительно")
}
// Вывод: Хорошо</code></pre>

<p>Go проверяет условия <strong>сверху вниз</strong>. Первое истинное — выполняется, остальные пропускаются.</p>

<h2>Операторы сравнения</h2>
<table>
<tr><th>Оператор</th><th>Значение</th><th>Пример</th></tr>
<tr><td><code>==</code></td><td>Равно</td><td>x == 5</td></tr>
<tr><td><code>!=</code></td><td>Не равно</td><td>x != 0</td></tr>
<tr><td><code>&lt;</code></td><td>Меньше</td><td>x &lt; 10</td></tr>
<tr><td><code>&gt;</code></td><td>Больше</td><td>x &gt; 0</td></tr>
<tr><td><code>&lt;=</code></td><td>Меньше или равно</td><td>x &lt;= 100</td></tr>
<tr><td><code>&gt;=</code></td><td>Больше или равно</td><td>x &gt;= 18</td></tr>
</table>

<p><strong>Важно:</strong> <code>==</code> (два знака равно) — сравнение. <code>=</code> (один) — присваивание. Не путай!</p>

<h2>Логические операторы</h2>
<pre><code>age := 25
hasLicense := true

// && — И (оба условия должны быть true)
if age >= 18 && hasLicense {
    fmt.Println("Можно водить")
}

// || — ИЛИ (хотя бы одно true)
if age < 12 || age > 65 {
    fmt.Println("Льготный билет")
}

// ! — НЕ (инвертирует)
if !hasLicense {
    fmt.Println("Нужны права")
}</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА 1: = вместо ==
if age = 18 {      // ❌ присваивание, не сравнение!
if age == 18 {     // ✅ сравнение

// ОШИБКА 2: скобка на следующей строке
if age >= 18
{                  // ❌ Go требует { на той же строке что и if
if age >= 18 {     // ✅ правильно</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что выведет: if 10 > 5 { fmt.Println(\"Да\") } else { fmt.Println(\"Нет\") }?",
				Options:     []string{"Нет", "Да", "Ошибку", "Ничего"},
				Correct:     1,
				Explanation: "10 > 5 — истина (true). Выполняется код в первых фигурных скобках: \"Да\".",
			},
			{
				Question:    "Чем == отличается от =?",
				Options:     []string{"Ничем", "== сравнивает значения, = присваивает значение переменной", "== быстрее", "= для строк, == для чисел"},
				Correct:     1,
				Explanation: "x == 5 — вопрос 'x равен 5?'. x = 5 — команда 'запиши 5 в x'. В if используй только ==.",
			},
			{
				Question:    "Что означает && в условии?",
				Options:     []string{"Или", "И — оба условия должны быть true", "Не", "Сложение"},
				Correct:     1,
				Explanation: "&& — логическое И. age >= 18 && hasID — true только если И возраст >= 18, И есть документ. Если хоть одно false — всё false.",
			},
			{
				Question:    "Почему if age = 18 { } — ошибка?",
				Options:     []string{"18 — не переменная", "= это присваивание, нужно == для сравнения", "Не хватает else", "age не объявлена"},
				Correct:     1,
				Explanation: "Go не даст присвоить значение внутри if. Нужно == (сравнение): if age == 18 { }. Это защита от частой ошибки.",
			},
			{
				Question:    "В каком порядке проверяются условия в if/else if/else?",
				Options:     []string{"Случайно", "Сверху вниз — первое истинное выполняется, остальные пропускаются", "Все одновременно", "Снизу вверх"},
				Correct:     1,
				Explanation: "Go проверяет по порядку. Как только нашёл true — выполняет этот блок и пропускает все остальные else if/else.",
			},
		},
		Tasks: []T{
			{
				Title: "Совершеннолетие", Difficulty: "easy",
				Description: `<p>Прочитай возраст. Если >= 18 — выведи "Доступ разрешён", иначе "Доступ запрещён":</p>
<p>Ввод: <code>20</code> → Вывод: <code>Доступ разрешён</code></p>
<p>Ввод: <code>15</code> → Вывод: <code>Доступ запрещён</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "if condition { } else { }", Definition: "Если условие истинно — первый блок, иначе — второй."},
				},
				TestCases: []TestCase{
					{Input: "20", ExpectedOutput: "Доступ разрешён"},
					{Input: "15", ExpectedOutput: "Доступ запрещён"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var age int
    fmt.Scan(&age)

    if age >= 18 {
        fmt.Println("Доступ разрешён")
    }
    // Добавь else с "Доступ запрещён"
}`,
				Hints:    `<p>После закрывающей скобки <code>}</code> добавь <code>else { fmt.Println("Доступ запрещён") }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var age int
    fmt.Scan(&age)
    if age >= 18 {
        fmt.Println("Доступ разрешён")
    } else {
        fmt.Println("Доступ запрещён")
    }
}</code></pre>`,
			},
			{
				Title: "Чётное или нечётное", Difficulty: "easy",
				Description: `<p>Определи чётное число или нет:</p>
<p>Ввод: <code>4</code> → Вывод: <code>Чётное</code></p>
<p>Ввод: <code>7</code> → Вывод: <code>Нечётное</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "% (остаток)", Definition: "n % 2 == 0 — число чётное. n % 2 != 0 — нечётное."},
				},
				TestCases: []TestCase{
					{Input: "4", ExpectedOutput: "Чётное"},
					{Input: "7", ExpectedOutput: "Нечётное"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    // Если n % 2 == 0 → "Чётное", иначе "Нечётное"
}`,
				Hints:    `<p><code>if n % 2 == 0 { fmt.Println("Чётное") } else { fmt.Println("Нечётное") }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    if n%2 == 0 {
        fmt.Println("Чётное")
    } else {
        fmt.Println("Нечётное")
    }
}</code></pre>`,
			},
			{
				Title: "Оценка по баллу", Difficulty: "medium",
				Description: `<p>По баллу (0-100) выведи оценку:</p>
<ul>
<li>90+ → "Отлично"</li>
<li>70-89 → "Хорошо"</li>
<li>50-69 → "Удовлетворительно"</li>
<li>&lt;50 → "Неудовлетворительно"</li>
</ul>
<p>Ввод: <code>85</code> → Вывод: <code>Хорошо</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "else if", Definition: "Дополнительное условие: if ... else if ... else if ... else. Проверяются по порядку."},
				},
				TestCases: []TestCase{
					{Input: "95", ExpectedOutput: "Отлично"},
					{Input: "85", ExpectedOutput: "Хорошо"},
					{Input: "55", ExpectedOutput: "Удовлетворительно"},
					{Input: "30", ExpectedOutput: "Неудовлетворительно"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var score int
    fmt.Scan(&score)

    if score >= 90 {
        fmt.Println("Отлично")
    } else if score >= 70 {
        fmt.Println("Хорошо")
    }
    // Допиши else if для >= 50 и else
}`,
				Hints:    `<p>Добавь <code>else if score >= 50 { ... } else { ... }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var score int
    fmt.Scan(&score)
    if score >= 90 {
        fmt.Println("Отлично")
    } else if score >= 70 {
        fmt.Println("Хорошо")
    } else if score >= 50 {
        fmt.Println("Удовлетворительно")
    } else {
        fmt.Println("Неудовлетворительно")
    }
}</code></pre>`,
			},
			{
				Title: "Максимум из трёх", Difficulty: "medium",
				Description: `<p>Прочитай три числа и выведи наибольшее:</p>
<p>Ввод: <code>5 12 8</code> → Вывод: <code>12</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "Вложенные if", Definition: "if внутри if. Или: сохрани max и сравнивай каждое число с ним."},
				},
				TestCases: []TestCase{
					{Input: "5 12 8", ExpectedOutput: "12"},
					{Input: "100 50 75", ExpectedOutput: "100"},
					{Input: "1 1 1", ExpectedOutput: "1"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var a, b, c int
    fmt.Scan(&a, &b, &c)

    max := a
    if b > max {
        max = b
    }
    // Добавь проверку для c
    fmt.Println(max)
}`,
				Hints:    `<p>Добавь <code>if c > max { max = c }</code> перед fmt.Println.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var a, b, c int
    fmt.Scan(&a, &b, &c)
    max := a
    if b > max { max = b }
    if c > max { max = c }
    fmt.Println(max)
}</code></pre>`,
			},
			{
				Title: "Калькулятор с операцией", Difficulty: "hard",
				Description: `<p>Прочитай два числа и операцию (+, -, *, /). Выведи результат:</p>
<p>Ввод: <code>10 3 +</code> → Вывод: <code>13</code></p>
<p>Ввод: <code>10 3 /</code> → Вывод: <code>3</code></p>
<p>При делении на 0: <code>Ошибка: деление на ноль</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "string сравнение", Definition: "op == \"+\" — сравниваем строку с строкой. Кавычки обязательны."},
				},
				TestCases: []TestCase{
					{Input: "10 3 +", ExpectedOutput: "13"},
					{Input: "10 3 -", ExpectedOutput: "7"},
					{Input: "10 3 *", ExpectedOutput: "30"},
					{Input: "10 3 /", ExpectedOutput: "3"},
					{Input: "10 0 /", ExpectedOutput: "Ошибка: деление на ноль"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &b, &op)

    if op == "+" {
        fmt.Println(a + b)
    } else if op == "-" {
        fmt.Println(a - b)
    }
    // Допиши * и / (с проверкой деления на 0)
}`,
				Hints:    `<p>Для /: <code>if b == 0 { fmt.Println("Ошибка: деление на ноль") } else { fmt.Println(a/b) }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &b, &op)

    if op == "+" {
        fmt.Println(a + b)
    } else if op == "-" {
        fmt.Println(a - b)
    } else if op == "*" {
        fmt.Println(a * b)
    } else if op == "/" {
        if b == 0 {
            fmt.Println("Ошибка: деление на ноль")
        } else {
            fmt.Println(a / b)
        }
    }
}</code></pre>`,
			},
		},
	}
}

// ── Урок 6: Switch ─────────────────────────────────────────────

func lesson06_conditions() L {
	return L{
		Slug: "switch", Title: "Switch — выбор из вариантов", Order: 6,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Switch — выбор из вариантов</h1>

<h2>Когда if/else слишком много</h2>
<p>Если нужно проверить одну переменную на множество значений — <code>switch</code> читабельнее:</p>

<pre><code>day := "Monday"

switch day {
case "Monday":
    fmt.Println("Понедельник")
case "Tuesday":
    fmt.Println("Вторник")
case "Wednesday":
    fmt.Println("Среда")
default:
    fmt.Println("Другой день")
}</code></pre>

<p><strong>Как работает:</strong> Go сравнивает <code>day</code> с каждым <code>case</code>. Первое совпадение — выполняется. <code>default</code> — если ничего не совпало.</p>

<h2>Несколько значений в одном case</h2>
<pre><code>day := "Saturday"

switch day {
case "Saturday", "Sunday":
    fmt.Println("Выходной!")
default:
    fmt.Println("Рабочий день")
}</code></pre>

<h2>Switch без значения — замена if/else</h2>
<pre><code>score := 85

switch {
case score >= 90:
    fmt.Println("A")
case score >= 80:
    fmt.Println("B")
case score >= 70:
    fmt.Println("C")
default:
    fmt.Println("F")
}
// Вывод: B</code></pre>

<p>Switch без значения проверяет условия — как if/else if, но чище.</p>

<h2>Отличия от других языков</h2>
<pre><code>// В Go НЕ нужен break!
// В C/Java каждый case "проваливается" (fall through) в следующий.
// В Go — автоматическая остановка после первого совпадения.

switch n {
case 1:
    fmt.Println("Один")    // выполнится только это
case 2:
    fmt.Println("Два")     // НЕ выполнится
}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Когда лучше использовать switch вместо if/else?",
				Options:     []string{"Никогда", "Когда проверяем одну переменную на множество конкретных значений", "switch быстрее", "Только для строк"},
				Correct:     1,
				Explanation: "switch day { case \"Mon\": ... case \"Tue\": ... } читабельнее чем 7 if/else if. Для 2-3 условий if проще, для 4+ — switch.",
			},
			{
				Question:    "Нужен ли break в Go switch?",
				Options:     []string{"Да, как в Java/C", "Нет — Go автоматически останавливается после первого совпавшего case", "Только для чисел", "Зависит от версии"},
				Correct:     1,
				Explanation: "В Go break в switch не нужен. Каждый case автоматически завершается. Это осознанное решение — меньше ошибок.",
			},
			{
				Question:    "Что делает default в switch?",
				Options:     []string{"Ничего", "Выполняется если ни один case не совпал", "Первый вариант", "Обязателен"},
				Correct:     1,
				Explanation: "default — как else в if. Ни один case не подошёл → выполняется default. Не обязателен, но рекомендуется для обработки неожиданных значений.",
			},
			{
				Question:    "Что выведет: switch { case 5 > 3: fmt.Println(\"A\") case 10 > 5: fmt.Println(\"B\") }?",
				Options:     []string{"A и B", "Только A — первое истинное условие", "Только B", "Ошибку"},
				Correct:     1,
				Explanation: "Switch без значения проверяет условия по порядку. 5 > 3 — true, выполняется \"A\", дальше не проверяется.",
			},
			{
				Question:    "Можно ли написать case \"a\", \"b\", \"c\"?",
				Options:     []string{"Нет", "Да — несколько значений через запятую в одном case", "Только для чисел", "Нужен отдельный case для каждого"},
				Correct:     1,
				Explanation: "case \"a\", \"b\", \"c\": — совпадение с любым из значений. Удобно для группировки: case \"Mon\", \"Tue\", \"Wed\": fmt.Println(\"Будний\")",
			},
		},
		Tasks: []T{
			{
				Title: "День недели", Difficulty: "easy",
				Description: `<p>Прочитай номер дня (1-7) и выведи название:</p>
<p>Ввод: <code>1</code> → Вывод: <code>Понедельник</code></p>
<p>Ввод: <code>7</code> → Вывод: <code>Воскресенье</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "switch n { case 1: ... }", Definition: "Сравнивает n с каждым case. Первое совпадение — выполняется."},
				},
				TestCases: []TestCase{
					{Input: "1", ExpectedOutput: "Понедельник"},
					{Input: "5", ExpectedOutput: "Пятница"},
					{Input: "7", ExpectedOutput: "Воскресенье"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var day int
    fmt.Scan(&day)

    switch day {
    case 1:
        fmt.Println("Понедельник")
    case 2:
        fmt.Println("Вторник")
    // Допиши остальные дни: 3-Среда, 4-Четверг, 5-Пятница, 6-Суббота, 7-Воскресенье
    }
}`,
				Hints:    `<p>Добавь case 3: ... case 4: ... и так до case 7.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var day int
    fmt.Scan(&day)
    switch day {
    case 1: fmt.Println("Понедельник")
    case 2: fmt.Println("Вторник")
    case 3: fmt.Println("Среда")
    case 4: fmt.Println("Четверг")
    case 5: fmt.Println("Пятница")
    case 6: fmt.Println("Суббота")
    case 7: fmt.Println("Воскресенье")
    }
}</code></pre>`,
			},
			{
				Title: "Выходной или рабочий", Difficulty: "easy",
				Description: `<p>По номеру дня определи — рабочий (1-5) или выходной (6-7):</p>
<p>Ввод: <code>3</code> → Вывод: <code>Рабочий день</code></p>
<p>Ввод: <code>6</code> → Вывод: <code>Выходной!</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "case 6, 7:", Definition: "Несколько значений в одном case — совпадение с любым."},
				},
				TestCases: []TestCase{
					{Input: "3", ExpectedOutput: "Рабочий день"},
					{Input: "6", ExpectedOutput: "Выходной!"},
					{Input: "7", ExpectedOutput: "Выходной!"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var day int
    fmt.Scan(&day)

    switch day {
    case 6, 7:
        fmt.Println("Выходной!")
    default:
        // Допиши
    }
}`,
				Hints:    `<p><code>fmt.Println("Рабочий день")</code> в блоке default.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var day int
    fmt.Scan(&day)
    switch day {
    case 6, 7: fmt.Println("Выходной!")
    default: fmt.Println("Рабочий день")
    }
}</code></pre>`,
			},
			{
				Title: "Времена года", Difficulty: "medium",
				Description: `<p>По номеру месяца (1-12) выведи время года:</p>
<p>Ввод: <code>7</code> → Вывод: <code>Лето</code></p>
<p>12, 1, 2 → Зима. 3-5 → Весна. 6-8 → Лето. 9-11 → Осень.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Группировка case", Definition: "case 12, 1, 2: — зима. Логически объединяет несколько значений."},
				},
				TestCases: []TestCase{
					{Input: "1", ExpectedOutput: "Зима"},
					{Input: "4", ExpectedOutput: "Весна"},
					{Input: "7", ExpectedOutput: "Лето"},
					{Input: "10", ExpectedOutput: "Осень"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var month int
    fmt.Scan(&month)

    switch month {
    case 12, 1, 2:
        fmt.Println("Зима")
    // Допиши: Весна (3,4,5), Лето (6,7,8), Осень (9,10,11)
    }
}`,
				Hints:    `<p>case 3, 4, 5: "Весна". case 6, 7, 8: "Лето". case 9, 10, 11: "Осень".</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var month int
    fmt.Scan(&month)
    switch month {
    case 12, 1, 2: fmt.Println("Зима")
    case 3, 4, 5: fmt.Println("Весна")
    case 6, 7, 8: fmt.Println("Лето")
    case 9, 10, 11: fmt.Println("Осень")
    }
}</code></pre>`,
			},
			{
				Title: "Калькулятор на switch", Difficulty: "medium",
				Description: `<p>Перепиши калькулятор из прошлого урока используя switch:</p>
<p>Ввод: <code>10 5 *</code> → Вывод: <code>50</code></p>`,
				Glossary: []GlossaryItem{
					{Term: "switch op { case \"+\": ... }", Definition: "Switch по строке — чище чем if op == \"+\" { } else if ..."},
				},
				TestCases: []TestCase{
					{Input: "10 5 +", ExpectedOutput: "15"},
					{Input: "10 5 *", ExpectedOutput: "50"},
					{Input: "10 0 /", ExpectedOutput: "Ошибка: деление на ноль"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &b, &op)

    switch op {
    case "+":
        fmt.Println(a + b)
    case "-":
        fmt.Println(a - b)
    // Допиши * и /
    }
}`,
				Hints:    `<p>Для /: внутри case "/" проверь <code>if b == 0</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var a, b int
    var op string
    fmt.Scan(&a, &b, &op)
    switch op {
    case "+": fmt.Println(a + b)
    case "-": fmt.Println(a - b)
    case "*": fmt.Println(a * b)
    case "/":
        if b == 0 { fmt.Println("Ошибка: деление на ноль") } else { fmt.Println(a / b) }
    }
}</code></pre>`,
			},
			{
				Title: "Статус HTTP кода", Difficulty: "hard",
				Description: `<p>По HTTP-коду выведи статус:</p>
<p>200 → "OK", 301 → "Moved", 404 → "Not Found", 500 → "Server Error", остальное → "Unknown"</p>
<p>Ввод: <code>3</code> (количество), потом 3 кода</p>
<pre><code>3
200 404 500</code></pre>
<p>Вывод:</p>
<pre><code>200: OK
404: Not Found
500: Server Error</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "switch + fmt.Printf", Definition: "Комбинация: switch определяет текст, Printf форматирует вывод."},
				},
				TestCases: []TestCase{
					{Input: "3\n200 404 500", ExpectedOutput: "200: OK\n404: Not Found\n500: Server Error"},
					{Input: "2\n301 999", ExpectedOutput: "301: Moved\n999: Unknown"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    for i := 0; i < n; i++ {
        var code int
        fmt.Scan(&code)

        var status string
        switch code {
        case 200: status = "OK"
        case 301: status = "Moved"
        // Допиши 404 и 500
        default: status = "Unknown"
        }
        fmt.Printf("%d: %s\n", code, status)
    }
}`,
				Hints:    `<p>case 404: status = "Not Found". case 500: status = "Server Error".</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    for i := 0; i < n; i++ {
        var code int
        fmt.Scan(&code)
        var status string
        switch code {
        case 200: status = "OK"
        case 301: status = "Moved"
        case 404: status = "Not Found"
        case 500: status = "Server Error"
        default: status = "Unknown"
        }
        fmt.Printf("%d: %s\n", code, status)
    }
}</code></pre>`,
			},
		},
	}
}

// ── Урок 7: Циклы ──────────────────────────────────────────────

func lesson07_loops() L {
	return L{
		Slug: "loops", Title: "Циклы — повторение действий", Order: 7,
		Difficulty: "beginner", Track: "shared",
		Content: `<h1>Циклы — повторение действий</h1>

<h2>for — единственный цикл в Go</h2>
<p>В отличие от других языков, в Go есть только один цикл — <code>for</code>. Но он умеет всё:</p>

<pre><code>// Считаем от 1 до 5
for i := 1; i <= 5; i++ {
    fmt.Println(i)
}
// 1
// 2
// 3
// 4
// 5</code></pre>

<p>Разберём: <code>for НАЧАЛО; УСЛОВИЕ; ШАГ { тело }</code></p>
<ul>
<li><code>i := 1</code> — начинаем с 1</li>
<li><code>i &lt;= 5</code> — пока i меньше или равно 5</li>
<li><code>i++</code> — после каждого повторения увеличиваем на 1</li>
</ul>

<h2>Цикл как while (только условие)</h2>
<pre><code>n := 1
for n <= 100 {    // пока n <= 100
    n = n * 2     // удваиваем
}
fmt.Println(n)    // 128 (первое число > 100)</code></pre>

<h2>Бесконечный цикл</h2>
<pre><code>for {
    // бесконечный цикл — используй break для выхода
    break    // выходим
}</code></pre>

<h2>break и continue</h2>
<pre><code>// break — полностью выйти из цикла
for i := 1; i <= 10; i++ {
    if i == 5 {
        break    // стоп! выходим из цикла
    }
    fmt.Println(i)  // 1 2 3 4
}

// continue — пропустить текущую итерацию
for i := 1; i <= 5; i++ {
    if i == 3 {
        continue    // пропускаем 3
    }
    fmt.Println(i)  // 1 2 4 5
}</code></pre>

<h2>Суммирование и накопление</h2>
<pre><code>// Сумма чисел от 1 до N
n := 10
sum := 0
for i := 1; i <= n; i++ {
    sum = sum + i    // или sum += i
}
fmt.Println(sum)     // 55</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА 1: бесконечный цикл (забыл изменять условие)
i := 0
for i < 10 {
    fmt.Println(i)  // i всегда 0 → никогда не выйдет!
    // Нужно: i++
}

// ОШИБКА 2: off-by-one (ошибка на единицу)
for i := 0; i < 5; i++ {     // 0,1,2,3,4 — ПЯТЬ итераций
for i := 0; i <= 5; i++ {    // 0,1,2,3,4,5 — ШЕСТЬ итераций!
for i := 1; i <= 5; i++ {    // 1,2,3,4,5 — ПЯТЬ итераций</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Сколько раз выполнится: for i := 0; i < 3; i++ { }?",
				Options:     []string{"2", "3", "4", "Бесконечно"},
				Correct:     1,
				Explanation: "i=0 (< 3 ✓), i=1 (< 3 ✓), i=2 (< 3 ✓), i=3 (< 3 ✗ — стоп). Итого 3 раза: при i=0, 1, 2.",
			},
			{
				Question:    "Что делает break?",
				Options:     []string{"Ломает программу", "Немедленно выходит из текущего цикла", "Пропускает итерацию", "Перезапускает цикл"},
				Correct:     1,
				Explanation: "break — полный выход из цикла. Код после цикла продолжит выполняться.",
			},
			{
				Question:    "Что делает continue?",
				Options:     []string{"Выходит из цикла", "Пропускает оставшийся код в текущей итерации и переходит к следующей", "Перезапускает цикл", "Ничего"},
				Correct:     1,
				Explanation: "continue пропускает всё после себя в текущей итерации. Цикл продолжает со следующего значения i.",
			},
			{
				Question:    "Как написать while-цикл в Go?",
				Options:     []string{"while condition { }", "for condition { }", "do { } while(condition)", "loop { }"},
				Correct:     1,
				Explanation: "В Go нет while. Вместо этого: for condition { }. for — единственный цикл в Go, но с разными формами записи.",
			},
			{
				Question:    "Что выведет: for i := 1; i <= 3; i++ { if i == 2 { continue }; fmt.Print(i) }?",
				Options:     []string{"123", "13", "12", "1"},
				Correct:     1,
				Explanation: "i=1: выводим 1. i=2: continue — пропускаем вывод. i=3: выводим 3. Результат: 13.",
			},
		},
		Tasks: []T{
			{
				Title: "Считай до N", Difficulty: "easy",
				Description: `<p>Прочитай число N и выведи числа от 1 до N, каждое на новой строке:</p>
<p>Ввод: <code>5</code></p>
<p>Вывод:</p><pre><code>1
2
3
4
5</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "for i := 1; i <= n; i++", Definition: "Цикл: начиная с 1, пока i <= n, каждый раз увеличивая i на 1."},
				},
				TestCases: []TestCase{
					{Input: "5", ExpectedOutput: "1\n2\n3\n4\n5"},
					{Input: "3", ExpectedOutput: "1\n2\n3"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    for i := 1; i <= n; i++ {
        fmt.Println(i)
    }
}`,
				Hints:    `<p>Код уже написан! Запусти и разберись как работает цикл.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    for i := 1; i <= n; i++ {
        fmt.Println(i)
    }
}</code></pre>`,
			},
			{
				Title: "Сумма от 1 до N", Difficulty: "easy",
				Description: `<p>Посчитай сумму чисел от 1 до N:</p>
<p>Ввод: <code>10</code> → Вывод: <code>55</code></p>
<p>Ввод: <code>5</code> → Вывод: <code>15</code> (1+2+3+4+5)</p>`,
				Glossary: []GlossaryItem{
					{Term: "sum += i", Definition: "Короткая запись для sum = sum + i. Накапливаем сумму."},
				},
				TestCases: []TestCase{
					{Input: "10", ExpectedOutput: "55"},
					{Input: "5", ExpectedOutput: "15"},
					{Input: "1", ExpectedOutput: "1"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    sum := 0
    for i := 1; i <= n; i++ {
        // Добавь i к sum
    }
    fmt.Println(sum)
}`,
				Hints:    `<p><code>sum += i</code> или <code>sum = sum + i</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    sum := 0
    for i := 1; i <= n; i++ {
        sum += i
    }
    fmt.Println(sum)
}</code></pre>`,
			},
			{
				Title: "Таблица умножения", Difficulty: "medium",
				Description: `<p>Выведи таблицу умножения для числа N (от 1 до 10):</p>
<p>Ввод: <code>3</code></p>
<p>Вывод:</p><pre><code>3 x 1 = 3
3 x 2 = 6
3 x 3 = 9
3 x 4 = 12
3 x 5 = 15
3 x 6 = 18
3 x 7 = 21
3 x 8 = 24
3 x 9 = 27
3 x 10 = 30</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Printf в цикле", Definition: "fmt.Printf(\"%d x %d = %d\\n\", n, i, n*i) — форматированный вывод на каждой итерации."},
				},
				TestCases: []TestCase{
					{Input: "3", ExpectedOutput: "3 x 1 = 3\n3 x 2 = 6\n3 x 3 = 9\n3 x 4 = 12\n3 x 5 = 15\n3 x 6 = 18\n3 x 7 = 21\n3 x 8 = 24\n3 x 9 = 27\n3 x 10 = 30"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    for i := 1; i <= 10; i++ {
        fmt.Printf("%d x %d = %d\n", n, i, n*i)
    }
}`,
				Hints:    `<p>Код уже готов! Разберись как <code>n*i</code> вычисляет результат на каждой итерации.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    for i := 1; i <= 10; i++ {
        fmt.Printf("%d x %d = %d\n", n, i, n*i)
    }
}</code></pre>`,
			},
			{
				Title: "FizzBuzz", Difficulty: "medium",
				Description: `<p>Классическая задача: для чисел от 1 до N:</p>
<ul>
<li>Делится на 3 и 5 → "FizzBuzz"</li>
<li>Делится на 3 → "Fizz"</li>
<li>Делится на 5 → "Buzz"</li>
<li>Иначе → само число</li>
</ul>
<p>Ввод: <code>15</code></p>
<p>Вывод:</p><pre><code>1
2
Fizz
4
Buzz
Fizz
7
8
Fizz
Buzz
11
Fizz
13
14
FizzBuzz</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "i % 15 == 0", Definition: "Делится и на 3, и на 5. Проверяй ЭТО первым — иначе попадёт в case для 3 или 5."},
				},
				TestCases: []TestCase{
					{Input: "15", ExpectedOutput: "1\n2\nFizz\n4\nBuzz\nFizz\n7\n8\nFizz\nBuzz\n11\nFizz\n13\n14\nFizzBuzz"},
					{Input: "5", ExpectedOutput: "1\n2\nFizz\n4\nBuzz"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    for i := 1; i <= n; i++ {
        if i%15 == 0 {
            fmt.Println("FizzBuzz")
        } else if i%3 == 0 {
            fmt.Println("Fizz")
        }
        // Допиши: проверку на 5 и вывод числа (else)
    }
}`,
				Hints:    `<p>Добавь <code>else if i%5 == 0 { fmt.Println("Buzz") } else { fmt.Println(i) }</code></p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    for i := 1; i <= n; i++ {
        if i%15 == 0 {
            fmt.Println("FizzBuzz")
        } else if i%3 == 0 {
            fmt.Println("Fizz")
        } else if i%5 == 0 {
            fmt.Println("Buzz")
        } else {
            fmt.Println(i)
        }
    }
}</code></pre>`,
			},
			{
				Title: "Обратный отсчёт", Difficulty: "hard",
				Description: `<p>Прочитай N чисел. Выведи их сумму, минимум и максимум:</p>
<p>Ввод:</p><pre><code>5
3 7 1 9 4</code></pre>
<p>Вывод:</p><pre><code>Сумма: 24
Минимум: 1
Максимум: 9</code></pre>`,
				Glossary: []GlossaryItem{
					{Term: "Паттерн min/max", Definition: "Начни с первого элемента. В цикле: if x < min { min = x }. Аналогично для max."},
				},
				TestCases: []TestCase{
					{Input: "5\n3 7 1 9 4", ExpectedOutput: "Сумма: 24\nМинимум: 1\nМаксимум: 9"},
					{Input: "3\n10 10 10", ExpectedOutput: "Сумма: 30\nМинимум: 10\nМаксимум: 10"},
				},
				StarterCode: `package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)

    var first int
    fmt.Scan(&first)
    sum := first
    min := first
    max := first

    for i := 1; i < n; i++ {
        var x int
        fmt.Scan(&x)
        sum += x
        if x < min {
            min = x
        }
        // Допиши проверку для max
    }

    fmt.Println("Сумма:", sum)
    fmt.Println("Минимум:", min)
    fmt.Println("Максимум:", max)
}`,
				Hints:    `<p>Добавь <code>if x > max { max = x }</code> после проверки min.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
    var n int
    fmt.Scan(&n)
    var first int
    fmt.Scan(&first)
    sum, min, max := first, first, first
    for i := 1; i < n; i++ {
        var x int
        fmt.Scan(&x)
        sum += x
        if x < min { min = x }
        if x > max { max = x }
    }
    fmt.Println("Сумма:", sum)
    fmt.Println("Минимум:", min)
    fmt.Println("Максимум:", max)
}</code></pre>`,
			},
		},
	}
}
