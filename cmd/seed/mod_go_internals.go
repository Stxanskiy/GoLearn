package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Go Internals — Как Go работает под капотом
// Глубокое погружение в runtime, memory model, scheduler, GC
// ════════════════════════════════════════════════════════════════

func mod_go_internals() M {
	return M{
		Slug:          "go-internals",
		Title:         "Go Internals — Как Go работает под капотом",
		Description:   "Планировщик GMP, сборщик мусора, escape analysis, устройство интерфейсов, каналов и map. Знания для senior-уровня.",
		Order:         24,
		Track:         "backend",
		Difficulty:    "advanced",
		Prerequisites: []string{"concurrency", "interfaces", "pointers"},
		Lessons: []L{
			lesson_memory_model(),
			lesson_scheduler_deep(),
			lesson_gc_internals(),
			lesson_interface_internals(),
			lesson_channel_internals(),
			lesson_map_internals(),
			lesson_common_gotchas(),
		},
	}
}

// ── Урок 1: Memory Model ────────────────────────────────────────

func lesson_memory_model() L {
	return L{
		Slug: "memory-model", Title: "Модель памяти Go", Order: 1,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Модель памяти Go</h1>

<h2>Стек vs Куча</h2>
<p>Каждая горутина имеет свой <strong>стек</strong> — область памяти для локальных переменных. Начальный размер — <strong>2КБ</strong> (раньше был 8КБ). Стек растёт динамически до 1ГБ.</p>

<pre><code>// Стек горутины:
// ┌──────────────────┐ Высокие адреса
// │ frame: main()    │
// │   x := 42       │ ← локальная переменная, живёт на стеке
// │   p := &amp;x       │ ← указатель на стеке (но куда указывает?)
// ├──────────────────┤
// │ frame: helper()  │
// ���   y := "hello"  │
// └──────────────────┘ Низкие адреса (стек растёт вниз)</code></pre>

<h2>Escape Analysis — кто решает стек или куча?</h2>
<p><strong>Escape analysis</strong> �� это анализ компилятора, который определяет, может ли переменная остаться на стеке или должна "сбежать" (escape) в кучу.</p>

<pre><code>// Остаётся на стеке — быстро, без GC
func sum(a, b int) int {
    result := a + b  // result на стеке, освобождается при return
    return result
}

// Сбегает в кучу — медленнее, треб��ет GC
func newUser(name string) *User {
    u := User{Name: name}  // u НА КУЧЕ! Возвращаем указатель наружу
    return &amp;u               // компилятор видит: &amp;u покидает функцию → heap
}</code></pre>

<p><strong>Когда переменная сбегает в кучу:</strong></p>
<ul>
<li>Возвращаем указатель на локальную переменную</li>
<li>Переменная слишком большая для стека (обычно >64КБ)</li>
<li>Размер неизвестен на этапе компиляции (например, <code>make([]int, n)</code> где n — переменная)</li>
<li>Переменная захвачена замыканием (closure) которое переживает функцию</li>
<li>Присваивание в интерфейс (boxing)</li>
</ul>

<pre><code>// Как посмотреть escape analysis:
// go build -gcflags="-m" main.go
//
// Выво��:
// ./main.go:5:2: moved to heap: u
// ./main.go:3:6: can inline newUser</code></pre>

<h2>Рост стека (Stack Growth)</h2>
<p>В Go 1.4+ используется <strong>копирующий стек</strong> (contiguous stack). Когда стек заполняется:</p>
<ol>
<li>Аллоцируется новый стек в 2x больше</li>
<li>Все данные копируются</li>
<li>Все указатели внутри стека обновляются</li>
<li>Старый стек освобождается</li>
</ol>

<p><strong>Это причина, почему нельзя хранить указатель на стек горутины в C-коде!</strong> Адрес может измениться.</p>

<h2>Аллокатор памяти (tcmalloc-inspired)</h2>
<pre><code>// Go использует иерархический аллокатор:
//
// mcache (per-P) → mcentral (shared) → mheap (global) → OS
//
// Маленькие объекты (≤32КБ):
//   Берутся из mcache без блокировок (каждый P имеет свой)
//   Разбиты на ~70 размерных классов (8B, 16B, 32B, ..., 32KB)
//
// Большие объекты (>32КБ):
//   Аллоцируются напрямую из mheap (с блокировкой)</code></pre>

<h2>Выравнивание и padding</h2>
<pre><code>// Порядок полей влияет на размер структуры!
type Bad struct {
    a bool   // 1 byte + 7 padding
    b int64  // 8 bytes
    c bool   // 1 byte + 7 padding
}  // = 24 bytes

type Good struct {
    b int64  // 8 bytes
    a bool   // 1 byte
    c bool   // 1 byte + 6 padding
}  // = 16 bytes — экономия 33%!</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Функция возвращает *User. Где аллоцируется User?",
				Options:     []string{"Всегда на стеке", "Всегда в куче", "На куче, потому что указатель покидает функцию (escape)", "Зависит от размера структуры"},
				Correct:     2,
				Explanation: "Escape analysis видит, что указатель на User возвращается наружу — переменная 'сбегает' в кучу. Это решение компилятора, не программиста. Посмотреть можно через go build -gcflags='-m'.",
			},
			{
				Question:    "Стек горутины начинается с 2КБ. Что происходит когда он заполняется?",
				Options:     []string{"Программа падает с stack overflow", "Go аллоцирует новый стек 2x размера и копирует данные", "Go переключается на использование кучи", "Go создаёт новую горутину"},
				Correct:     1,
				Explanation: "Go использует 'contiguous stacks' — при заполнении стека аллоцируется новый в 2x больше, все данные копируются, указатели обновляются. Это значит указатели на стек нестабильны (но Go это скрывает).",
			},
			{
				Question:    "Почему порядок полей в структуре влияет на её размер?",
				Options:     []string{"Компилятор сортирует поля", "Из-за alignment padding — процессор требует выравнивания данных по их размеру", "Из-за кэш-линий", "Не влияет в Go"},
				Correct:     1,
				Explanation: "CPU требует чтобы int64 лежал по адресу кратному 8, int32 — кратному 4 и т.д. Компилятор вставляет padding (пустые байты) между полями для выравнивания. Правильный порядок ��олей (от больших к маленьким) минимизирует padding.",
			},
			{
				Question:    "make([]int, n) где n — переменная. Куда аллоцируется слайс?",
				Options:     []string{"Стек — это локальная переменная", "Куча — размер неизвестен на этапе компиляции", "Зависит от знач��ния n", "Стек если n < 1024, иначе куча"},
				Correct:     1,
				Explanation: "Если компилятор не может доказать размер на этапе компиляции, переменная escapes to heap. make([]int, 10) может остаться на стеке (размер известен), но make([]int, n) всегда в куче.",
			},
		},
		Tasks: []T{
			{
				Title:       "Анализатор структурного padding",
				Description: `<p>Напиши программу, которая по описанию полей структуры (тип и размер) рассчитывает общий размер с учётом alignment padding.</p><p>Правила выравнивания: каждое поле должно начинаться с адреса, кратного его размеру. В конце структуры добавляется padding до кратности максимального поля.</p>`,
				Difficulty:  "hard",
				Glossary: []GlossaryItem{
					{Term: "alignment", Definition: "Требование CPU: данные размера N должны лежать по адресу кратному N"},
					{Term: "padding", Definition: "Пустые байты, добавленные для выравнивания"},
					{Term: "struct size", Definition: "Суммарный размер всех полей + все padding"},
				},
				StarterCode: `package main

import (
	"fmt"
	"bufio"
	"os"
	"strconv"
	"strings"
)

func calcStructSize(fields []int) int {
	// fields — список размеров полей в байтах (например [1, 8, 1] для bool, int64, bool)
	// Верни общий размер структуры с padding
	// Правила:
	// 1. Каждое поле начинается с адреса кратного его размеру
	// 2. Общий размер кратен максимальному полю
	// TODO: реализуй
	return 0
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	parts := strings.Fields(scanner.Text())
	fields := make([]int, len(parts))
	for i, p := range parts {
		fields[i], _ = strconv.Atoi(p)
	}
	fmt.Println(calcStructSize(fields))
}`,
				Hints: `<ul><li>Держи переменную offset — текущая позиция в структуре</li><li>Для к��ждого поля: offset = roundUp(offset, fieldSize), потом offset += fieldSize</li><li>В конце: totalSize = roundUp(offset, maxField)</li><li>roundUp(x, align) = ((x + align - 1) / align) * align</li></ul>`,
				Solution: `<pre><code>func calcStructSize(fields []int) int {
    offset := 0
    maxAlign := 1
    for _, size := range fields {
        if size > maxAlign {
            maxAlign = size
        }
        // Align offset to field size
        if offset%size != 0 {
            offset += size - (offset % size)
        }
        offset += size
    }
    // Align total to max field
    if offset%maxAlign != 0 {
        offset += maxAlign - (offset % maxAlign)
    }
    return offset
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "1 8 1", ExpectedOutput: "24"},
					{Input: "8 1 1", ExpectedOutput: "16"},
					{Input: "4 4 4", ExpectedOutput: "12"},
					{Input: "1 1 1", ExpectedOutput: "3"},
					{Input: "8 4 2 1", ExpectedOutput: "16"},
					{Input: "1 2 4 8", ExpectedOutput: "24"},
				},
			},
			{
				Title:       "Escape Analysis симулятор",
				Description: `<p>Определи, сбежит ли переменная в кучу или останется на стеке. На вход подаются описания ситуаций, на выход — "heap" или "stack".</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "escape", Definition: "Переменная 'сбегает' в кучу когда компилятор не может гарантировать что она не переживёт функцию"},
					{Term: "heap", Definition: "Куча — динамическая память, управляемая GC"},
					{Term: "stack", Definition: "Стек — быстрая память функции, освобождается автоматически при return"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func escapeAnalysis(situation string) string {
	// Определи: "heap" или "stack"
	// Ситуации:
	// "return_pointer" — функция возвращает указатель на локальную переменную
	// "local_only" — переменная используется только внутри функции
	// "interface_assign" — присваивание в interface{}
	// "closure_escape" — переменная захвачена горутиной
	// "known_size_slice" — make([]int, 10) (размер известен, не возвращается)
	// "unknown_size_slice" — make([]int, n) где n — параметр
	// "large_struct" — структура > 64KB
	// "small_value" — int, bool, маленькая структура внутри функции
	// TODO: реализуй
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fmt.Println(escapeAnalysis(line))
	}
}`,
				Hints: `<ul><li>return_pointer → heap (указатель покидает функцию)</li><li>interface_assign → heap (boxing requires heap allocation)</li><li>closure_escape → heap (замыкание может пережить функцию)</li><li>unknown_size_slice → heap (размер неизвестен компилятору)</li><li>large_struct → heap (слишком большая для стека)</li></ul>`,
				Solution: `<pre><code>func escapeAnalysis(situation string) string {
    switch situation {
    case "return_pointer":
        return "heap"
    case "local_only":
        return "stack"
    case "interface_assign":
        return "heap"
    case "closure_escape":
        return "heap"
    case "known_size_slice":
        return "stack"
    case "unknown_size_slice":
        return "heap"
    case "large_struct":
        return "heap"
    case "small_value":
        return "stack"
    default:
        return "unknown"
    }
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "return_pointer\nlocal_only\ninterface_assign", ExpectedOutput: "heap\nstack\nheap"},
					{Input: "closure_escape\nknown_size_slice\nunknown_size_slice", ExpectedOutput: "heap\nstack\nheap"},
					{Input: "large_struct\nsmall_value", ExpectedOutput: "heap\nstack"},
				},
			},
		},
	}
}

// ── Урок 2: Планировщик ─────────────────────────────────────────

func lesson_scheduler_deep() L {
	return L{
		Slug: "scheduler-deep", Title: "Планировщик Go (GMP) — глубокое погружение", Order: 2,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Планировщик Go (GMP)</h1>

<h2>Три сущности</h2>
<pre><code>// G (Goroutine) — единица работы
// Содержит: стек, instruction pointer, статус, привязку к каналу
// Статусы: _Gidle, _Grunnable, _Grunning, _Gsyscall, _Gwaiting, _Gdead

// M (Machine) — поток ОС
// Каждый M привязан к одному P для выполнения G
// M может быть заблокирован в syscall (тогда P отвязывается)

// P (Processor) — логический процессор (GOMAXPROCS штук)
// Содержит: local run queue (до 256 G), mcache, таймеры
// P — это "право н�� выполнение". Без P горутина не запустится.</code></pre>

<h2>Как работает scheduling цикл</h2>
<pre><code>// Упрощённый цикл runtime.schedule():
func schedule() {
    // 1. Каждый 61-й тик — проверить global queue (fairness)
    if tick%61 == 0 {
        g = globrunqget(pp)
    }

    // 2. Взять G из local queue
    if g == nil {
        g = runqget(pp)
    }

    // 3. Если пусто — findrunnable() (блокирующий)
    //    Пробует: global queue, netpoll, work stealing
    if g == nil {
        g = findrunnable()
    }

    // 4. Выполнить G
    execute(g)
}</code></pre>

<h2>Work Stealing</h2>
<p>Когда у P пустая очередь, он <strong>крадёт</strong> горутины у другого P:</p>
<pre><code>// runtime.stealWork():
// 1. Выбрать случайный P (жертву)
// 2. Забрать половину его local queue
// 3. Если все P пусты — проверить global queue и netpoll
//
// Это обеспеч��вает автоматическую балансировку нагрузки</code></pre>

<h2>Preemption (вытеснение)</h2>
<p>До Go 1.14 горутина могла монополизировать M если не вызывала функций (tight loop). С 1.14 — <strong>asynchronous preemption</strong>:</p>
<pre><code>// Sysmon (системный монитор) — специальная горутина без P:
// - Запускается каждые 20мкс-10мс
// - Обнаруживает горутины работающие >10ms без preemption point
// - Отправляет сигнал SIGURG → горутина прерывается
// - Также: retake P у горутин в syscall >20мс, запуск GC</code></pre>

<h2>Netpoll — интеграция с сетью</h2>
<pre><code>// Когда горутина делает net.Read() на сокете:
// 1. Горутина НЕ блокирует M (это не syscall!)
// 2. Сокет регистрируется в epoll/kqueue
// 3. Горутина паркуется (Gwaiting)
// 4. M берёт другую G из очереди
// 5. Когда данные приходят: netpoll возвращает G в runnable queue
//
// Результат: 1 млн конкурентных соединений на ~4 потоках</code></pre>

<h2>Syscall handling</h2>
<pre><code>// Когда горутина входит в blocking syscall (file I/O):
// 1. G переходит в Gsyscall
// 2. P ОТВЯЗЫВАЕТСЯ от M (handoff)
// 3. P находит другой M (или создаёт новый)
// 4. P продолжает выполнять другие G
// 5. Когда syscall завершается:
//    - G пытается вернуть свой P (если свободен)
//    - Иначе G идёт в global queue</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Горутина крутится в tight loop (for {}) без вызовов функций. Что произойдёт в Go 1.22?",
				Options:     []string{"Она навсегда заблокирует поток", "Sysmon обнаружит и отправит SIGURG для preemption", "Другие горутины на этом P голодают навсе��да", "Go паникует"},
				Correct:     1,
				Explanation: "С Go 1.14 реализовано asynchronous preemption через SIGURG. Sysmon каждые 10ms проверяет, не работает ли горутина слишком долго, и прерывает её через сигнал. До 1.14 tight loop действительно мог заблокировать P.",
			},
			{
				Question:    "Горутина вызвала net.Read(). Что происходит с потоком (M)?",
				Options:     []string{"M блокируется до получения данных", "M отвя��ывается от P и блокируется", "M НЕ блокируется — горутина паркуется, M берёт другую G", "Создаётся новый M"},
				Correct:     2,
				Explanation: "Сетевой I/O в Go — НЕ blocking syscall. Это netpoll (epoll/kqueue). Горутина паркуется (Gwaiting), M продолжает работать с другими G. Ко��да данные приходят, горутина возвращается в очередь. Именно поэтому Go эффективно обрабатывает миллионы соединений.",
			},
			{
				Question:    "У P пустая local queue. Что делает планировщик?",
				Options:     []string{"P засыпает и ждёт новых горутин", "P крадёт половину горутин у случайного другого P (work stealing)", "P берёт все горутины из global queue", "P уничтожается"},
				Correct:     1,
				Explanation: "Work stealing — ключевой механизм балансировки. P выбирает случайную жертву и забирает половину её local queue. Также проверяются global queue и netpoll. Это обеспечивает равномерную нагрузку без центрального координатора.",
			},
			{
				Question:    "Зачем каждый 61-й тик планировщик проверяет global queue?",
				Options:     []string{"Это магическое число без смысла", "Для fairness — чтобы горутины в global queue не голодали", "Для сборки мусора", "Для синхронизации таймеров"},
				Correct:     1,
				Explanation: "Без этого горутины в global queue могли бы голодать, если local queues всегда заняты. Число 61 выбрано как простое число (не создаёт паттернов с другими интервалами). Это гарантия fairness — каждая горутина рано или поздно получит CPU.",
			},
		},
		Tasks: []T{
			{
				Title:       "Симулятор work stealing",
				Description: `<p>Реализуй упрощённый work stealing планировщик. У тебя N процессоров, каждый с очередью задач. Когда очередь пуста — процессор крадёт половину задач у случайного соседа.</p><p>Вход: число процессоров и начальные задачи для каждого. Выход: порядок выполнения задач каждым процессором.</p>`,
				Difficulty:  "hard",
				Glossary: []GlossaryItem{
					{Term: "work stealing", Definition: "Стратегия балансировки: idle процессор крадёт задачи у загруженного"},
					{Term: "local queue", Definition: "Очередь задач привязанная к конкретному процессору"},
					{Term: "fairness", Definition: "Гарантия что каждая задача рано или поздно будет выполнена"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func simulate(queues [][]string) [][]string {
	// queues[i] — задачи для процессора i
	// Правила:
	// 1. Каждый процессор берёт задачу из головы своей очереди
	// 2. Если очередь пуста — крадёт половину у процессора с наибольшей очередью
	// 3. "Половина" = len/2, берутся с конца очереди жертвы
	// 4. Если все очереди пусты — стоп
	// Верни: results[i] — порядок задач выполненных процессором i
	results := make([][]string, len(queues))
	for i := range results {
		results[i] = []string{}
	}

	// TODO: реализуй
	return results
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	scanner.Scan()
	n := 0
	fmt.Sscanf(scanner.Text(), "%d", &n)

	queues := make([][]string, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			queues[i] = []string{}
		} else {
			queues[i] = strings.Fields(line)
		}
	}

	results := simulate(queues)
	for _, r := range results {
		if len(r) == 0 {
			fmt.Println("idle")
		} else {
			fmt.Println(strings.Join(r, " "))
		}
	}
}`,
				Hints: `<ul><li>Основной цикл: while хоть одна очередь непуста</li><li>Для каждого процессора: если очередь непуста — pop front, иначе найди процессор с max очередью и укради половину с конца</li><li>Одна итерация = один "тик" — все процессоры работают одновременно</li></ul>`,
				Solution: `<pre><code>func simulate(queues [][]string) [][]string {
    results := make([][]string, len(queues))
    for i := range results {
        results[i] = []string{}
    }
    for {
        allEmpty := true
        for i := range queues {
            if len(queues[i]) > 0 {
                allEmpty = false
                results[i] = append(results[i], queues[i][0])
                queues[i] = queues[i][1:]
            } else {
                // Work stealing: find largest queue
                maxIdx, maxLen := -1, 0
                for j := range queues {
                    if j != i && len(queues[j]) > maxLen {
                        maxIdx = j
                        maxLen = len(queues[j])
                    }
                }
                if maxIdx != -1 && maxLen > 1 {
                    half := maxLen / 2
                    stolen := queues[maxIdx][maxLen-half:]
                    queues[maxIdx] = queues[maxIdx][:maxLen-half]
                    queues[i] = append(queues[i], stolen...)
                    allEmpty = false
                }
            }
        }
        if allEmpty { break }
    }
    return results
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "2\nA B C D\n", ExpectedOutput: "A B C\nD"},
					{Input: "2\nA B C D E F\n", ExpectedOutput: "A B C D\nE F"},
				},
			},
		},
	}
}

// ── Урок 3: GC Internals ────────────────────────────────────────

func lesson_gc_internals() L {
	return L{
		Slug: "gc-internals", Title: "Сборщик мусора Go", Order: 3,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Сборщик мусора Go</h1>

<h2>Tri-color Mark & Sweep</h2>
<p>Go использует <strong>concurrent, tri-color, mark-sweep</strong> GC с <strong>write barrier</strong>:</p>

<pre><code>// Три цвета:
// Белый — объект не посещён (потенциально мусор)
// Серый — объект посещён, но его ссылки не проверены
// Чёрный — объект и все его ссылки проверены (живой)
//
// Алгоритм:
// 1. Все объекты — белые
// 2. Root set (глобалы, стеки, регистры) → серые
// 3. Берём серый объект:
//    - Все его ссылки → серые
//    - Сам → чёрный
// 4. Повторяем пока есть серые
// 5. Белые = мусор → освобождаем</code></pre>

<h2>Почему concurrent?</h2>
<p>GC работает <strong>одновременно</strong> с мутатором (вашим кодом). Это минимизирует STW (stop-the-world) паузы:</p>
<pre><code>// Фазы GC:
// 1. Mark Setup (STW ~10-30μs) — включить write barrier
// 2. Marking (concurrent) — обход графа объектов параллельно с вашим кодом
// 3. Mark Termination (STW ~10-30μs) — выключить write barrier
// 4. Sweeping (concurrent) — освобождение белых объектов
//
// STW паузы: обычно <1ms даже для огромных хипов</code></pre>

<h2>Write Barrier</h2>
<p>Проблема: пока GC сканирует, мутатор может переместить указатель. Это ��арушит инвариант:</p>
<pre><code>// Инвариант: чёрный ��бъект НИКОГДА не указывает на белый напрямую
// (иначе белый будет собран хотя он жив)
//
// Write barrier: при записи указателя runtime перехватывает операцию:
// Если dst чёрный и src белый → пометить src серым
//
// Стоимость write barrier: ~несколько наносекунд на каждое присваивание указате��я
// Включен ТОЛЬКО во время marking phase</code></pre>

<h2>Когда запускается GC?</h2>
<pre><code>// GOGC=100 (по умолчанию):
// GC запускается когда heap вырос на 100% от размера после прошлой сборки
//
// Пример: после GC heap = 10MB → следующий GC при heap = 20MB
//
// GOGC=200 → при 30MB (реже, больше память)
// GOGC=50 → при 15MB (чаще, меньше память)
// GOGC=off → GC отключен (осторожно!)
//
// Go 1.19+: GOMEMLIMIT — жёсткий лимит памяти
// runtime/debug.SetMemoryLimit(512 << 20) // 512MB</code></pre>

<h2>GC Tuning на практике</h2>
<pre><code>// Проблема: GC слишком частый (высокий CPU)
// Решение: увеличить GOGC или установить GOMEMLIMIT

// Проблема: большие аллокации создают давление на GC
// Решение: sync.Pool для переиспользования объектов

// Проблема: latency spikes
// Решение: GODEBUG=gctrace=1 для диагностики

// runtime/debug.ReadGCStats() — статистика GC
// runtime.ReadMemStats() — статистика памяти</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что такое write barrier в контексте GC?",
				Options:     []string{"Запрет записи в память", "Механизм перехвата присваивания указателей для поддержания инварианта tri-color GC", "Блокировка потоков во время GC", "Защита от race conditions"},
				Correct:     1,
				Explanation: "Write barrier — это runtime-код, который вставляется при каждом присваивании указателя во время marking phase. Он гарантирует, что чёрный объект не может указывать на белый напрямую (иначе живой объект будет ошибочно собран).",
			},
			{
				Question:    "GOGC=100 и после GC heap = 50MB. Когда запустится следующий GC?",
				Options:     []string{"При 75MB", "При 100MB (heap вырос на 100%)", "При 150MB", "При 51MB"},
				Correct:     1,
				Explanation: "GOGC=100 означает: запустить GC когда heap вырос на 100% от размера после прошлого GC. 50MB * 2 = 100MB — следующий GC при достижении 100MB heap.",
			},
			{
				Question:    "Максимальная STW пауза в Go 1.22 обычно составляет:",
				Options:     []string{"Десятки миллисекунд", "Менее 1 миллисекунды", "Несколько секунд", "Зависит от размера heap"},
				Correct:     1,
				Explanation: "Go GC оптимизирован для минимальных STW пауз. С Go 1.5+ (concurrent GC) паузы обычно <500μs, независимо от размера heap. STW происходит только для setup/teardown write barrier, а сам marking — concurrent.",
			},
		},
		Tasks: []T{
			{
				Title:       "Симулятор tri-color marking",
				Description: `<p>Реализуй алгоритм tri-color mark & sweep. На вход: гра�� объектов (ссылки между ними) и root set. На выход: какие объекты будут собраны (garbage).</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "root set", Definition: "Начальные объекты, от которых начинается обход (стек, глобалы)"},
					{Term: "reachable", Definition: "Объект достижим из root set через цепочку ссылок"},
					{Term: "garbage", Definition: "Объекты недостижимые из root set — их память можно освободить"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func markAndSweep(objects []string, edges map[string][]string, roots []string) []string {
	// objects — все объекты
	// edges — ссылки: edges["A"] = ["B", "C"] значит A указывает на B и C
	// roots — корневые объекты
	// Верни отсортированный список МУСОРА (недостижимых объектов)
	// TODO: р��ализуй tri-color marking
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	// Line 1: objects (space-separated)
	scanner.Scan()
	objects := strings.Fields(scanner.Text())

	// Line 2: number of edges
	scanner.Scan()
	var n int
	fmt.Sscanf(scanner.Text(), "%d", &n)

	// Lines 3..n+2: edges "A B" means A -> B
	edges := make(map[string][]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		edges[parts[0]] = append(edges[parts[0]], parts[1])
	}

	// Last line: roots
	scanner.Scan()
	roots := strings.Fields(scanner.Text())

	garbage := markAndSweep(objects, edges, roots)
	if len(garbage) == 0 {
		fmt.Println("no garbage")
	} else {
		fmt.Println(strings.Join(garbage, " "))
	}
}`,
				Hints: `<ul><li>BFS/DFS от roots через edges</li><li>Все посещённые = живые</li><li>objects - живые = garbage</li><li>Отсортируй результат</li></ul>`,
				Solution: `<pre><code>func markAndSweep(objects []string, edges map[string][]string, roots []string) []string {
    alive := make(map[string]bool)
    queue := append([]string{}, roots...)
    for len(queue) > 0 {
        obj := queue[0]
        queue = queue[1:]
        if alive[obj] { continue }
        alive[obj] = true
        for _, ref := range edges[obj] {
            if !alive[ref] { queue = append(queue, ref) }
        }
    }
    var garbage []string
    for _, obj := range objects {
        if !alive[obj] { garbage = append(garbage, obj) }
    }
    sort.Strings(garbage)
    return garbage
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "A B C D E\n3\nA B\nB C\nD E\nA", ExpectedOutput: "D E"},
					{Input: "A B C\n2\nA B\nB C\nA", ExpectedOutput: "no garbage"},
					{Input: "A B C D\n1\nA B\nA", ExpectedOutput: "C D"},
				},
			},
		},
	}
}

// ── Урок 4: Interface Internals ─────────────────────────────────

func lesson_interface_internals() L {
	return L{
		Slug: "interface-internals", Title: "Интерфейсы изнутри", Order: 4,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Интерфейсы изнутри</h1>

<h2>Два типа интерфейсов в runtime</h2>
<pre><code>// runtime.iface — интерфейс с методами
type iface struct {
    tab  *itab    // type info + method table
    data unsafe.Pointer  // указатель на данные
}

// runtime.eface — пустой интерфейс (interface{} / any)
type eface struct {
    _type *_type          // только type info
    data  unsafe.Pointer  // указатель на данные
}

// Размер: 2 слова (16 байт на 64-bit)
// Это причина почему ��нтерфейс != указатель на значение</code></pre>

<h2>itab — таблица методов</h2>
<pre><code>type itab struct {
    inter *interfacetype  // тип интерфейса (какие методы нужны)
    _type *_type          // конкретный тип значения
    hash  uint32          // для быстрых type assertions
    _     [4]byte
    fun   [1]uintptr      // массив указателей на методы (variable-size)
}

// itab кэшируется глобально!
// Первый раз: runtime ищет методы конкретного типа, заполняет itab
// В��орой раз: берёт из кэша (O(1) lookup)</code></pre>

<h2>Nil Interface Gotcha</h2>
<pre><code>// САМАЯ ЧАСТАЯ ОШИБКА с интерфейсами:
var p *MyStruct = nil
var i interface{} = p

fmt.Println(i == nil)  // false!!!

// Почему? i содержит:
// tab: *itab для (*MyStruct, interface{})
// data: nil
// i != nil потому что tab != nil!

// nil interface: И tab И data равны nil
var j interface{} = nil
fmt.Println(j == nil)  // true (оба поля nil)</code></pre>

<h2>Boxing — упаковка значения в интерфейс</h2>
<pre><code>var x int = 42
var i interface{} = x  // boxing!

// Что происходит:
// 1. Аллоцируется память для копии x (на куче или стек — escape analysis)
// 2. x копируется в эту память
// 3. eface.data = указатель на копию
// 4. eface._type = *_type для int
//
// Оптимизация: маленькие значения (≤ pointer size) хранятся прямо в data
// без дополнительной аллокации</code></pre>

<h2>Type Assertion и Type Switch</h2>
<pre><code>// Type assertion:
val, ok := i.(ConcreteType)
// Runtime проверяет: itab._type == *_type(ConcreteType)?
// Если да — возвращает data, иначе — zero value + false

// Type switch:
switch v := i.(type) {
case string: // hash comparison для быстрого matching
case int:
case fmt.Stringer: // interface → interface: проверка itab.inter
}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "var p *MyStruct = nil; var i error = p; i == nil?",
				Options:     []string{"true", "false — interface содержит type info даже когда data=nil", "panic", "зависит от MyStruct"},
				Correct:     1,
				Explanation: "Классическая gotcha! Interface nil только когда ОБА поля (tab и data) равны nil. Здесь tab != nil (содержит itab для *MyStruct), а data = nil. Поэтому i != nil, хотя значение внутри nil. Это частая причина багов при возврате *ConcreteType as error.",
			},
			{
				Question:    "Сколько байт занимает interface{} на 64-bit системе?",
				Options:     []string{"8 б��й��", "16 байт (два указателя: type + data)", "Зависит от значения внутри", "24 байта"},
				Correct:     1,
				Explanation: "interface{} (eface) состоит из двух полей: *_type и data — оба pointer-size (8 байт на 64-bit). Итого 16 байт. Само значение хранится отдельно (на куче или inline для маленьких значений).",
			},
			{
				Question:    "Первый вызов type assertion для пары (тип, интерфейс) медленный. Почему второй быстрый?",
				Options:     []string{"JIT компиляция", "itab кэшируется глобально — повторный lookup O(1)", "Компилятор инлайнит", "Branch prediction"},
				Correct:     1,
				Explanation: "Go runtime поддерживает глобальный ��эш itab. При первом type assertion runtime и��ет методы типа и заполняет itab. При повторном — просто берёт готовый itab из кэша по хэшу (type, interface).",
			},
		},
		Tasks: []T{
			{
				Title:       "Nil interface детектор",
				Description: `<p>Определи, будет ли сравнение interface с nil давать true или false. На вход: описание того как создаётся интерфейс.</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "iface", Definition: "Runtime-представление интерфейса: {tab *itab, data unsafe.Pointer}"},
					{Term: "nil interface", Definition: "Интерфейс где и tab и data равны nil"},
					{Term: "typed nil", Definition: "Интерфейс с type info но nil data — НЕ равен nil!"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func isNilInterface(desc string) string {
	// Описания:
	// "nil_direct" — var i interface{} = nil
	// "typed_nil_ptr" — var p *T = nil; var i interface{} = p
	// "nil_interface_var" — var i error (zero value)
	// "non_nil_value" — var i interface{} = 42
	// "nil_ptr_to_interface" — передали nil конк��етный тип в interface параметр функции
	// "func_returns_nil_ptr_as_error" — func f() error { var p *MyErr = nil; return p }
	// "func_returns_nil_error" — func f() error { return nil }
	// Верни "true" если == nil, "false" если != nil
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		fmt.Println(isNilInterface(line))
	}
}`,
				Hints: `<ul><li>nil_direct и nil_interface_var — оба поля nil → true</li><li>typed_nil_ptr — type info заполнен → false!</li><li>func_returns_nil_ptr_as_error — то же что typed_nil_ptr → false</li></ul>`,
				Solution: `<pre><code>func isNilInterface(desc string) string {
    switch desc {
    case "nil_direct":
        return "true"
    case "typed_nil_ptr":
        return "false"
    case "nil_interface_var":
        return "true"
    case "non_nil_value":
        return "false"
    case "nil_ptr_to_interface":
        return "false"
    case "func_returns_nil_ptr_as_error":
        return "false"
    case "func_returns_nil_error":
        return "true"
    }
    return "unknown"
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "nil_direct\ntyped_nil_ptr\nnil_interface_var", ExpectedOutput: "true\nfalse\ntrue"},
					{Input: "non_nil_value\nfunc_returns_nil_ptr_as_error\nfunc_returns_nil_error", ExpectedOutput: "false\nfalse\ntrue"},
				},
			},
		},
	}
}

// ── Урок 5: Channel Internals ───────────────────────────────────

func lesson_channel_internals() L {
	return L{
		Slug: "channel-internals", Title: "Каналы изнутри", Order: 5,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Каналы изнутри</h1>

<h2>Структура hchan</h2>
<pre><code>// runtime.hchan — канал под капотом
type hchan struct {
    qcount   uint           // текущее количество элементов в буфере
    dataqsiz uint           // размер буфера (cap)
    buf      unsafe.Pointer // кольцевой буфер
    elemsize uint16         // размер одного элемента
    closed   uint32         // канал закрыт?
    sendx    uint           // индекс для следующей записи
    recvx    uint           // индекс для следующего чтения
    recvq    waitq          // очередь ожидающих получателей (sudog linked list)
    sendq    waitq          // очередь ожидающих отправителей
    lock     mutex          // мьютекс для всех операций
}</code></pre>

<h2>Три режима работы</h2>
<pre><code>// 1. Буферизованный канал (make(chan int, 5)):
//    - За��исывает в кольцевой буфер пока не заполнен
//    - Блокирует отправителя только когда буфер полон
//    - Блокирует получателя только когда буфер пуст

// 2. Небуферизованный канал (make(chan int)):
//    - dataqsiz = 0, буфера нет
//    - Отправитель блокируется ДО получателя
//    - Происходит direct copy: данные копируются из стека sender в стек receiver
//    - Это САМЫЙ быстрый путь — нет промежуточного буфера

// 3. nil канал:
//    - var ch chan int (не инициализирован)
//    - Отправка и получение БЛОКИРУЮТСЯ НАВСЕГДА
//    - Полезно в select для отключения ветки</code></pre>

<h2>Что происходит при ch &lt;- value</h2>
<pre><code>// 1. Lock(ch.lock)
// 2. Есть ли ждущий receiver в recvq?
//    ДА → direct send: копируем данные прямо в стек receiver, будим его
//    НЕТ → буфер не полон? копируем в buf[sendx], sendx++
//    НЕТ И ПОЛОН → создаём sudog, паркуем горутину в sendq
// 3. Unlock(ch.lock)
//
// Direct send — опт��мизация: обходим буфер пол��остью
// Данные идут sender.stack → receiver.stack без копий</code></pre>

<h2>Select statement</h2>
<pre><code>// select компилируется в runtime.selectgo()
// Алгоритм:
// 1. Перемешать cases (randomize order) — для fairness
// 2. Отсортировать по адресу канала — для lock ordering (предотвращает deadlock)
// 3. Пройти по cases: если кто-то ready → выполнить, return
// 4. Никто не ready → зарегистрировать горутину во всех каналах, park
// 5. Когда любой канал разбудит — dequeue из остальных, выполнить case</code></pre>

<h2>Close semantics</h2>
<pre><code>// close(ch):
// 1. ch.closed = 1
// 2. Все ждущие receivers получают zero value
// 3. Все ждущие senders → PANIC
//
// Правила:
// - Только один горутина должна закрывать канал (обычно sender)
// - Отправка в закрытый канал → panic
// - Получение из закрытого канала → ok (буферные данные), потом zero values
// - Повторное закрытие → panic</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Небуферизованный канал. Горутина A отправляет, B ждёт. Куда копируются данные?",
				Options:     []string{"В промежуточный буфер", "Напрямую из стека A в стек B (direct send)", "В кучу", "В регистры"},
				Correct:     1,
				Explanation: "Direct send — оптимизация для unbuffered channels. Когда receiver уже ждёт, данные копируются напрямую из стека sender в стек receiver без промежуточного буфера. Это самый быстрый путь передачи данных между горутинами.",
			},
			{
				Question:    "Что происходит при чтении из nil канала?",
				Options:     []string{"Panic", "Возвращает zero value", "Блокирует навсегда", "Compile error"},
				Correct:     2,
				Explanation: "Чтение/запись в nil канал блокирует навсегда (горутина паркуется без возможности проснуться). Это свойство полезно в select: присвоив каналу nil, вы отключаете эту ветку select без изменения кода.",
			},
			{
				Question:    "Почему select перемешивает порядок cases?",
				Options:     []string{"Для безопасности", "Для fairness — чтобы один case не имел приоритета", "Для скорости", "Для совместимости"},
				Correct:     1,
				Explanation: "Если несколько cases ready одновременно, Go выбир��ет случайный (не первый по порядку). Без рандомизации первый case имел бы несправедливый приоритет, и нижние cases могли бы голодать при высокой нагрузке.",
			},
		},
		Tasks: []T{
			{
				Title:       "Симулятор кольцевого буфера канала",
				Description: `<p>Реализуй кольцевой буфер фиксированного размера (как в buffered channel). Операции: send (добавить) и recv (забрать). При полном буфере send возвращает "blocked", при ��устом recv возвращ��ет "blocked".</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "ring buffer", Definition: "Кольцевой буфер — массив с двумя индексами: sendx (запись) и recvx (чтение), wrap-around при достижении конца"},
					{Term: "sendx", Definition: "Индекс для следующей записи в буфер"},
					{Term: "recvx", Definition: "Индекс для следующего чтения из буфера"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type RingBuffer struct {
	buf   []string
	size  int
	count int
	sendx int
	recvx int
}

func NewRingBuffer(capacity int) *RingBuffer {
	return &RingBuffer{buf: make([]string, capacity), size: capacity}
}

func (rb *RingBuffer) Send(val string) string {
	// Если буфер полон — вернуть "blocked"
	// Иначе — добавить в buf[sendx], инкрементировать sendx (с wrap-around)
	// Вернуть "ok"
	// TODO
	return ""
}

func (rb *RingBuffer) Recv() string {
	// Если буфер пуст — вернуть "blocked"
	// Иначе — взять из buf[recvx], инкрементировать recvx (с wrap-around)
	// Вернуть значение
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var cap int
	fmt.Sscanf(scanner.Text(), "%d", &cap)
	rb := NewRingBuffer(cap)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		parts := strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "send":
			fmt.Println(rb.Send(parts[1]))
		case "recv":
			fmt.Println(rb.Recv())
		}
	}
}`,
				Hints: `<ul><li>Send: if count == size return "blocked"; buf[sendx] = val; sendx = (sendx+1) % size; count++</li><li>Recv: if count == 0 return "blocked"; val = buf[recvx]; recvx = (recvx+1) % size; count--</li></ul>`,
				Solution: `<pre><code>func (rb *RingBuffer) Send(val string) string {
    if rb.count == rb.size { return "blocked" }
    rb.buf[rb.sendx] = val
    rb.sendx = (rb.sendx + 1) % rb.size
    rb.count++
    return "ok"
}
func (rb *RingBuffer) Recv() string {
    if rb.count == 0 { return "blocked" }
    val := rb.buf[rb.recvx]
    rb.recvx = (rb.recvx + 1) % rb.size
    rb.count--
    return val
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "3\nsend A\nsend B\nsend C\nsend D\nrecv\nrecv\nsend E\nrecv", ExpectedOutput: "ok\nok\nok\nblocked\nA\nB\nok\nC"},
					{Input: "1\nrecv\nsend X\nrecv\nrecv", ExpectedOutput: "blocked\nok\nX\nblocked"},
				},
			},
		},
	}
}

// ── Урок 6: Map Internals ───────────────────────────────────────

func lesson_map_internals() L {
	return L{
		Slug: "map-internals", Title: "Map изнутри", Order: 6,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Map изнутри</h1>

<h2>Структура runtime.hmap</h2>
<pre><code>// runtime.hmap
type hmap struct {
    count     int    // количество элементов
    flags     uint8  // состояние (writing, iterating, etc.)
    B         uint8  // log2(количество бакетов). buckets = 2^B
    noverflow uint16 // приблизительное количество overflow бакетов
    hash0     uint32 // seed для хэш-функции (рандом при создании map)
    buckets   unsafe.Pointer  // массив из 2^B бакетов
    oldbuckets unsafe.Pointer // при росте: старые бакеты
    nevacuate  uintptr        // счётчик эвакуированных бакетов (growing)
}

// Каждый ба��ет хранит 8 key-value пар:
type bmap struct {
    tophash [8]uint8  // верхние 8 бит хэша каждого ключа (для быстрого поиска)
    // За ним идут: 8 ключей подряд, потом 8 значений подряд
    // (keys и values разделены для лучшего alignment!)
    // В конце: указатель на overflow бакет (если >8 элементов)
}</code></pre>

<h2>Как работает lookup</h2>
<pre><code>// map[key]:
// 1. hash = hashFunc(key, hmap.hash0)
// 2. bucket_index = hash & (2^B - 1)  // нижние B бит
// 3. tophash = hash >> (64 - 8)  // верхние 8 бит
// 4. В бакете: сравниваем tophash (быстро — 1 байт)
//    Если tophash совпал → сравниваем полный ключ
//    Если нет → проверяем overflow бакеты
//
// Tophash — это bloom-filter для одного бакета:
// Отсеивает ~255/256 ложных кандидатов одним сравнением байта</code></pre>

<h2>Рост map (Growing)</h2>
<pre><code>// Когда растёт:
// load factor > 6.5 (слишком много элементов на бакет)
// ИЛИ слишком много overflow бакетов
//
// Процесс (incremental):
// 1. Создать новый массив бакетов 2x (или same-size для overflow cleanup)
// 2. oldbuckets = текущие бакеты
// 3. При каждом доступе: эвакуировать 1-2 старых бакета в новые
// 4. Когда все эвакуированы → освободить oldbuckets
//
// Инкрементальный рост = нет one-time spike latency!</code></pre>

<h2>Почему map не concurrent-safe</h2>
<pre><code>// flags содержит бит hashWriting
// При записи: flags |= hashWriting
// При любом доступе: if flags&hashWriting != 0 → FATAL (не panic, а fatal!)
//
// Это FATAL — не recover() — программа умирает
// Сделано специально чтобы ловить race conditions
//
// Решения для concurrent access:
// 1. sync.Mutex / sync.RWMutex вокруг map
// 2. sync.Map (оптимизирован для read-heavy + keys mostly stable)
// 3. Шардирование: []map с mutex на каждый шард</code></pre>

<h2>Порядок итерации</h2>
<pre><code>// for k, v := range m { ... }
// Порядок РАНДОМИЗИРОВАН в каждом запуске!
//
// Go специально добавил рандомизацию чтобы программисты
// не полагались на порядок (он может измениться между версиями Go)
//
// Стартовая позиция итерации: random bucket + random offset в бакете</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что хранится в tophash[8] каждого бакета?",
				Options:     []string{"Полные хэши ключей", "Верхние 8 бит хэша — для быстрого отсеивания при lookup", "Индексы ��лючей", "Флаги состояния"},
				Correct:     1,
				Explanation: "tophash — это быстрый фильтр. При поиске ключа сначала сравниваем 1-байтовый tophash (отсеивает 255/256 кандидатов), и только при совпадении — сравниваем полный ключ. Это экономит дорогие сравнения ключей.",
			},
			{
				Question:    "Одновременная запись и чтение в map из разных горутин вызывает:",
				Options:     []string{"Panic (можно recover)", "Fatal (программа умирает, нельзя recover)", "Data race без последствий", "Deadlock"},
				Correct:     1,
				Explanation: "Go runtime использует FATAL (throw), а не panic. Это НЕПЕРЕХВАТЫВАЕМО — программа гарантированно завершается. Сделано специально чтобы race conditions на map не проходили незамеченными. Используй sync.Mutex или sync.Map.",
			},
			{
				Question:    "Map растёт инкрементально. Что это значит?",
				Options:     []string{"Все элементы копируются сразу в новые бакеты", "При каждом доступе эвакуируются 1-2 старых бакета — нет одного большого spike", "Map не растёт — паникует при переполнении", "Растёт только при GC"},
				Correct:     1,
				Explanation: "Инкрементальный рост: новые бакеты создаются сразу, но элементы переносятся понемногу (при каждом read/write). Это распределяет стоимость роста по времени, предотвращая latency spike. oldbuckets хранит старые данные пока эвакуация не завершена.",
			},
		},
		Tasks: []T{
			{
				Title:       "Hash map с tophash оптимизацией",
				Description: `<p>Реализуй простую hash map с tophash-оптимизацией: каждый бакет хранит до 4 элементов, tophash используется для быстрого фильтра.</p>`,
				Difficulty:  "hard",
				Glossary: []GlossaryItem{
					{Term: "tophash", Definition: "Верхние биты хэша для быстрого отсеивания кандидатов в бакете"},
					{Term: "bucket", Definition: "Группа слотов в hash table, выбирается по нижним битам хэша"},
					{Term: "load factor", Definition: "Отношение количества элементов к количеству бакетов"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

const bucketSize = 4
const numBuckets = 8

type Entry struct {
	key   string
	value string
}

type Bucket struct {
	tophash [bucketSize]uint8
	entries [bucketSize]Entry
	count   int
}

type HashMap struct {
	buckets [numBuckets]Bucket
}

func hash(key string) uint64 {
	// Simple FNV-1a
	h := uint64(14695981039346656037)
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= 1099511628211
	}
	return h
}

func (hm *HashMap) Put(key, value string) {
	h := hash(key)
	idx := h % numBuckets
	top := uint8(h >> 56) // верхние 8 бит
	if top == 0 { top = 1 } // 0 зарезервирован для "пусто"

	b := &hm.buckets[idx]
	// TODO: 1. Проверить tophash, если совпал — проверить ключ, обновить
	// TODO: 2. Если не найден — добавить в свободный слот
	_ = b
	_ = top
}

func (hm *HashMap) Get(key string) (string, bool) {
	h := hash(key)
	idx := h % numBuckets
	top := uint8(h >> 56)
	if top == 0 { top = 1 }

	b := &hm.buckets[idx]
	// TODO: Найти по tophash, потом по ключу
	_ = b
	_ = top
	return "", false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	hm := &HashMap{}
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		parts := strings.SplitN(line, " ", 3)
		switch parts[0] {
		case "put":
			hm.Put(parts[1], parts[2])
			fmt.Println("ok")
		case "get":
			if val, ok := hm.Get(parts[1]); ok {
				fmt.Println(val)
			} else {
				fmt.Println("not found")
			}
		}
	}
}`,
				Hints: `<ul><li>Put: iterate bucket slots, check tophash match then key match for update. If not found, find empty slot (tophash==0)</li><li>Get: iterate slots, if tophash[i]==top && entries[i].key==key return value</li><li>tophash==0 means empty slot</li></ul>`,
				Solution: `<pre><code>func (hm *HashMap) Put(key, value string) {
    h := hash(key)
    idx := h % numBuckets
    top := uint8(h >> 56)
    if top == 0 { top = 1 }
    b := &hm.buckets[idx]
    for i := 0; i < b.count; i++ {
        if b.tophash[i] == top && b.entries[i].key == key {
            b.entries[i].value = value
            return
        }
    }
    if b.count < bucketSize {
        b.tophash[b.count] = top
        b.entries[b.count] = Entry{key, value}
        b.count++
    }
}
func (hm *HashMap) Get(key string) (string, bool) {
    h := hash(key)
    idx := h % numBuckets
    top := uint8(h >> 56)
    if top == 0 { top = 1 }
    b := &hm.buckets[idx]
    for i := 0; i < b.count; i++ {
        if b.tophash[i] == top && b.entries[i].key == key {
            return b.entries[i].value, true
        }
    }
    return "", false
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "put name Alice\nput age 30\nget name\nget age\nget city", ExpectedOutput: "ok\nok\nAlice\n30\nnot found"},
					{Input: "put x 1\nput x 2\nget x", ExpectedOutput: "ok\nok\n2"},
				},
			},
		},
	}
}

// ── Урок 7: Common Gotchas ──────────────────────────────────────

func lesson_common_gotchas() L {
	return L{
		Slug: "common-gotchas", Title: "Частые ошибки и подводные камни Go", Order: 7,
		Difficulty: "intermediate", Track: "backend",
		Content: `<h1>Частые ошибки и подводные камни Go</h1>

<h2>1. Loop variable capture (до Go 1.22)</h2>
<pre><code>// БАГИ В Go &lt;1.22:
funcs := []func(){}
for _, v := range []int{1, 2, 3} {
    funcs = append(funcs, func() {
        fmt.Println(v) // v — ОДНА переменная, меняется в loop!
    })
}
for _, f := range funcs {
    f() // Печатает: 3, 3, 3 (не 1, 2, 3!)
}

// Фикс (до Go 1.22): копировать переменную
for _, v := range items {
    v := v // shadow — создать новую переменную
    go func() { process(v) }()
}

// Go 1.22+: каждая итерация = новая переменная (фикс по умолчанию)</code></pre>

<h2>2. Slice append gotcha</h2>
<pre><code>// Slice = {ptr, len, cap}. Append может НЕ создать новый underlying array!
a := make([]int, 3, 5) // len=3, cap=5
b := append(a, 4)      // cap ещё есть → b разделяет array с a!
b[0] = 999             // МЕНЯЕТ И a[0]!

// Безопасно: полная копия
b := append([]int{}, a...)
b := make([]int, len(a))
copy(b, a)

// Или: срез с ограничением cap
b := append(a[:len(a):len(a)], 4) // cap = len → всегда новый array</code></pre>

<h2>3. Defer в цикле</h2>
<pre><code>// ПРОБЛЕМА: defer не выполняется в конце итерации — только при return!
for _, f := range files {
    file, _ := os.Open(f)
    defer file.Close() // УТЕЧКА! Все файлы закроются только при выходе из функции
}

// ФИКС 1: замыкание
for _, f := range files {
    func() {
        file, _ := os.Open(f)
        defer file.Close()
        // работаем с file
    }()
}

// ФИКС 2: явное закрытие
for _, f := range files {
    file, _ := os.Open(f)
    processFile(file)
    file.Close()
}</code></pre>

<h2>4. Goroutine leak</h2>
<pre><code>// УТЕЧКА: горутина заблокирована навсегда
func search(query string) string {
    ch := make(chan string)
    go func() { ch &lt;- searchDB(query) }()
    go func() { ch &lt;- searchCache(query) }()
    return &lt;-ch  // берём первый результат
    // ВТОРАЯ горутина навсегда заблокирована на ch &lt;- !
}

// ФИКС: буферизованный канал
func search(query string) string {
    ch := make(chan string, 2) // буфер = количество горутин
    go func() { ch &lt;- searchDB(query) }()
    go func() { ch &lt;- searchCache(query) }()
    return &lt;-ch // вторая запишет в буфер и завершится
}</code></pre>

<h2>5. Error handling anti-patterns</h2>
<pre><code>// ПЛОХО: сравнение строк
if err.Error() == "not found" { ... }

// ХОРОШО: sentinel errors
if errors.Is(err, ErrNotFound) { ... }

// ПЛОХО: игнорирование ошибки
json.Unmarshal(data, &amp;result)

// ХОРОШО: всегда проверять
if err := json.Unmarshal(data, &amp;result); err != nil {
    return fmt.Errorf("parse config: %w", err) // wrap с контекстом
}

// ПЛОХО: fmt.Errorf без %w
return fmt.Errorf("failed: %v", err) // теряем error chain!

// ХОРОШО: %w для wrapping
return fmt.Errorf("open config: %w", err) // errors.Is/As работает</code></pre>

<h2>6. Race condition с map</h2>
<pre><code>// FATAL: concurrent map read and write
m := make(map[string]int)
go func() { m["a"] = 1 }()
go func() { _ = m["a"] }()
// → fatal error: concurrent map read and map write

// ФИКС: sync.RWMutex
type SafeMap struct {
    mu sync.RWMutex
    m  map[string]int
}
func (sm *SafeMap) Get(key string) int {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    return sm.m[key]
}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "a := make([]int, 3, 5); b := append(a, 4); b[0] = 999. Что будет с a[0]?",
				Options:     []string{"a[0] = 0 (не изменится)", "a[0] = 999 (b разделяет underlying array с a!)", "Panic", "Compile error"},
				Correct:     1,
				Explanation: "append не со��даёт новый array если есть свободный cap! a имеет cap=5, len=3 → append(a, 4) добавляет в позицию [3] того же array. b и a указывают на один array, поэтому b[0]=999 меняет a[0]. Fix: append(a[:len(a):len(a)], 4) — ограничить cap.",
			},
			{
				Question:    "defer в цикле for. Когда выполнятся deferred вызовы?",
				Options:     []string{"В конце каждой итерации", "При return из функции (ВСЕ СРАЗУ) — потенциальная утечка ресурсов", "При break из цикла", "Никогда"},
				Correct:     1,
				Explanation: "defer привязан к ФУНКЦИИ, не к блоку или итерации. В цикле defer накапливает вызовы до return. 1000 итераций с defer file.Close() = 1000 открытых файлов до выхода из функции. Fix: вынести в отдельную функцию или закрывать вручную.",
			},
			{
				Question:    "Горутина пишет в unbuffered канал, но receiver уже ушёл. Что происходит?",
				Options:     []string{"Данные теряются", "Горутина заблокирована навсегда (goroutine leak)", "Panic", "Канал автоматически освобождает горутину через GC"},
				Correct:     1,
				Explanation: "Горутина навсегда заблокирована на записи — это goroutine leak. GC НЕ собирает заблокированные горутины (они считаются живыми). Fix: использовать буферизованный канал, select с default, или context.Context для cancellation.",
			},
			{
				Question:    "fmt.Errorf(\"failed: %v\", err) vs fmt.Errorf(\"failed: %w\", err). В чём разница?",
				Options:     []string{"Нет разницы", "%w сохраняет error chain — errors.Is/As работает; %v теряет chain", "%w быстрее", "%v для production, %w для debug"},
				Correct:     1,
				Explanation: "%w wraps ошибку, сохраняя цепочку. errors.Is(wrappedErr, originalErr) вернёт true. %v просто форматирует строку — цепочка теряется, Is/As не сработают. Всегда используйте %w если хотите сохранить возможность проверки оригинальной ошибки.",
			},
		},
		Tasks: []T{
			{
				Title:       "Детектор goroutine leak паттернов",
				Description: `<p>Определи, есть ли goroutine leak в описанном паттерне. На вход: описание паттерна. На выход: "leak" или "safe".</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "goroutine leak", Definition: "Горутина навсегда заблокирована и никогда не завершится — потребляет память"},
					{Term: "buffered channel", Definition: "Канал с буфером — запись не блокируется пока буфер не полон"},
					{Term: "context cancellation", Definition: "Механизм отмены — позволяет горутинам узнать что результат больше не нужен"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func detectLeak(pattern string) string {
	// Паттерны:
	// "unbuf_no_receiver" — go func(){ ch <- v }() но никто не читает ch
	// "unbuf_first_wins" — 2 горутины пишут в unbuffered ch, берём 1 результат
	// "buffered_all_write" — 2 горутины пишут в ch с cap=2, берём 1 результат
	// "context_cancel" — горутина проверяет ctx.Done() в select
	// "range_closed" — горутина делает range по ch, ch закрывается
	// "ticker_no_stop" — time.NewTicker без defer ticker.Stop()
	// "http_body_no_close" — http.Get без defer resp.Body.Close()
	// TODO: верни "leak" или "safe"
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		fmt.Println(detectLeak(line))
	}
}`,
				Hints: `<ul><li>unbuf_no_receiver: никто не читает → sender заблокирован навсегда → leak</li><li>unbuf_first_wins: 2 пишут, 1 читает → второй навсегда заблокирован → leak</li><li>buffered_all_write: cap=2, обе записи поместятся → safe</li><li>context_cancel: select с ctx.Done() → горутина узнает об отмене → safe</li></ul>`,
				Solution: `<pre><code>func detectLeak(pattern string) string {
    switch pattern {
    case "unbuf_no_receiver":
        return "leak"
    case "unbuf_first_wins":
        return "leak"
    case "buffered_all_write":
        return "safe"
    case "context_cancel":
        return "safe"
    case "range_closed":
        return "safe"
    case "ticker_no_stop":
        return "leak"
    case "http_body_no_close":
        return "leak"
    }
    return "unknown"
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "unbuf_no_receiver\nunbuf_first_wins\nbuffered_all_write", ExpectedOutput: "leak\nleak\nsafe"},
					{Input: "context_cancel\nrange_closed\nticker_no_stop\nhttp_body_no_close", ExpectedOutput: "safe\nsafe\nleak\nleak"},
				},
			},
			{
				Title:       "Slice capacity trap детектор",
				Description: `<p>Определи, разделяют ли два слайса underlying array после операции append. На вход: начальные len и cap, и операция.</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "underlying array", Definition: "Реальный массив в памяти, на который указывает слайс"},
					{Term: "cap", Definition: "Capacity — максимальное количество элементов без реаллокации"},
					{Term: "shared array", Definition: "Два слайса указывают на один array — изменения в одном видны в другом"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func sharesArray(lenA, capA, appendCount int) string {
	// a := make([]int, lenA, capA)
	// b := append(a, <appendCount elements>...)
	// Вопрос: b[0] = X изменит a[0]?
	// "shared" если да (один array), "independent" если ��ет (новый array)
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		var l, c, n int
		fmt.Sscanf(line, "%d %d %d", &l, &c, &n)
		fmt.Println(sharesArray(l, c, n))
	}
}`,
				Hints: `<ul><li>Если lenA + appendCount <= capA → append не создаёт новый array → shared</li><li>Если lenA + appendCount > capA → нужен новый array → independent</li></ul>`,
				Solution: `<pre><code>func sharesArray(lenA, capA, appendCount int) string {
    if lenA+appendCount <= capA {
        return "shared"
    }
    return "independent"
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "3 5 1\n3 5 3\n3 3 1\n0 0 1", ExpectedOutput: "shared\nindependent\nindependent\nindependent"},
					{Input: "5 10 5\n5 10 6\n0 5 5", ExpectedOutput: "shared\nindependent\nshared"},
				},
			},
		},
	}
}
