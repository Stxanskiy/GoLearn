package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Конкурентность — расширенный (3 урока)
// Заменяет mod14_concurrency() из mod10_18.go
// ════════════════════════════════════════════════════════════════

func mod14_concurrency_full() M {
	return M{
		Slug:          "concurrency",
		Title:         "Конкурентность и параллелизм",
		Description:   "Горутины, каналы, sync, select, worker pool, race conditions. Ключевая сила Go.",
		Order:         14,
		Track:         "backend",
		Difficulty:    "advanced",
		Prerequisites: []string{"auth"},
		Lessons: []L{
			lesson_goroutines_channels(),
			lesson_sync_patterns(),
			lesson_advanced_concurrency(),
			lesson_race_conditions(),
			lesson_worker_pool(),
			lesson_context_concurrency(),
		},
	}
}

func lesson_goroutines_channels() L {
	return L{
		Slug: "goroutines-channels", Title: "Горутины и каналы", Order: 1,
		Difficulty: "intermediate", Track: "backend",
		Content: `<h1>Конкурентность vs Параллелизм</h1>

<h2>Разница, которую все путают</h2>
<p><strong>Конкурентность</strong> (concurrency) — <em>структура</em> программы. Несколько задач могут быть в процессе одновременно, но не обязательно выполняются в один момент. Как один повар жонглирует тремя блюдами.</p>

<p><strong>Параллелизм</strong> (parallelism) — <em>выполнение</em>. Задачи буквально выполняются одновременно на разных CPU. Как три повара — каждый готовит своё блюдо.</p>

<pre><code>// Go даёт конкурентность через горутины
// Параллелизм — бонус, если есть несколько CPU
// GOMAXPROCS = сколько потоков ОС использовать (по умолчанию = кол-во CPU)</code></pre>

<p><strong>Rob Pike:</strong> "Concurrency is about dealing with lots of things at once. Parallelism is about doing lots of things at once."</p>

<h2>Под капотом: планировщик Go (GMP)</h2>
<p>Go runtime содержит собственный планировщик, не зависящий от ОС:</p>

<pre><code>// G (Goroutine) — горутина. Стек начинается с 2КБ, растёт до 1ГБ.
// M (Machine) — поток ОС. Реально выполняет код.
// P (Processor) — логический процессор. GOMAXPROCS штук.
//
// Каждый P имеет локальную очередь горутин (local run queue).
// Есть глобальная очередь (global run queue) для переполнения.
//
// Цикл планировщика:
// 1. M берёт G из очереди P
// 2. Выполняет G
// 3. G завершилась или заблокирована → M берёт следующую G
// 4. Если очередь P пуста → work stealing из другого P</code></pre>

<p><strong>Почему горутины дешёвые:</strong></p>
<ul>
<li>Стек 2КБ vs 1МБ у потока → можно создать 1 000 000 горутин</li>
<li>Переключение контекста в user-space (не ядро) → ~10нс vs ~1мкс</li>
<li>Создание горутины ~0.3мкс vs ~10мкс поток ОС</li>
</ul>

<pre><code>// Горутина — это просто go перед вызовом функции
go processVideo(path)

// Анонимная горутина
go func() {
    fmt.Println("работаю в фоне")
}()

// ВНИМАНИЕ: main() не ждёт горутины!
func main() {
    go fmt.Println("hello")  // может не успеть напечатать!
}
// main() завершается → все горутины убиваются</code></pre>

<h2>Каналы — безопасная коммуникация</h2>
<p><strong>Девиз Go:</strong> "Don't communicate by sharing memory; share memory by communicating."</p>

<pre><code>// Канал — типизированная труба между горутинами
ch := make(chan string)

// Отправка
go func() {
    ch <- "результат"  // блокируется пока кто-то не прочитает
}()

// Чтение
msg := <-ch  // блокируется пока кто-то не отправит
fmt.Println(msg)  // "результат"</code></pre>

<h2>Небуферизированный vs Буферизированный</h2>
<pre><code>// НЕБУФЕРИЗИРОВАННЫЙ — синхронное рандеву
ch := make(chan int)
// Отправитель ждёт получателя. Получатель ждёт отправителя.
// Гарантия: данные доставлены в момент отправки.

// БУФЕРИЗИРОВАННЫЙ — асинхронная очередь
ch := make(chan int, 100)
// Отправитель НЕ ждёт, пока буфер не полон.
// Получатель НЕ ждёт, пока буфер не пуст.
// Используй когда: producer быстрее consumer.</code></pre>

<p><strong>Аналогия:</strong></p>
<ul>
<li>Небуферизированный = передать из рук в руки. Оба стоят и ждут.</li>
<li>Буферизированный = положить в почтовый ящик (размер = буфер). Если ящик полный — ждёшь.</li>
</ul>

<h2>Направленные каналы — type safety</h2>
<pre><code>// Канал только для отправки
func producer(out chan<- int) {
    for i := 0; i < 5; i++ {
        out <- i
    }
    close(out)  // сигнал: больше данных не будет
}

// Канал только для чтения
func consumer(in <-chan int) {
    for val := range in {  // range читает пока канал не закрыт
        fmt.Println(val)
    }
}

func main() {
    ch := make(chan int)     // двунаправленный
    go producer(ch)          // Go автоматически сужает тип
    consumer(ch)
}</code></pre>

<h2>close() и range — идиома завершения</h2>
<pre><code>ch := make(chan int, 5)
go func() {
    for i := 0; i < 5; i++ {
        ch <- i * i
    }
    close(ch)  // ВАЖНО: закрывает ОТПРАВИТЕЛЬ, не получатель
}()

// range автоматически завершается при close
for val := range ch {
    fmt.Println(val)  // 0, 1, 4, 9, 16
}

// Ручная проверка закрытия:
val, ok := <-ch  // ok == false если канал закрыт и пуст</code></pre>

<h2>Deadlock — самая частая ошибка</h2>
<pre><code>// DEADLOCK: отправка без получателя
ch := make(chan int)
ch <- 42  // заблокируется навечно → fatal error: all goroutines are asleep

// DEADLOCK: чтение без отправителя
ch := make(chan int)
val := <-ch  // заблокируется навечно

// ПРАВИЛЬНО: отправка в горутине
ch := make(chan int)
go func() { ch <- 42 }()
fmt.Println(<-ch)  // 42</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: горутина-утечка (goroutine leak)
ch := make(chan int)
go func() {
    result := heavyComputation()
    ch <- result  // если main уже ушёл — горутина висит навечно
}()
// Забыли прочитать из ch → горутина никогда не завершится!

// ОШИБКА: закрытие канала получателем
go func() {
    for val := range ch {
        process(val)
    }
    close(ch)  // ПАНИКА! Закрывать должен ОТПРАВИТЕЛЬ
}()

// ОШИБКА: двойное закрытие
close(ch)
close(ch)  // ПАНИКА: close of closed channel</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Чем конкурентность отличается от параллелизма?",
				Options:     []string{"Ничем", "Конкурентность — структура (несколько задач в процессе), параллелизм — выполнение (одновременно на разных CPU)", "Параллелизм быстрее", "Конкурентность — для одного CPU"},
				Correct:     1,
				Explanation: "Конкурентность — design pattern, параллелизм — execution model. Go даёт конкурентность через горутины. Параллелизм — бонус при нескольких CPU.",
			},
			{
				Question:    "Что произойдёт при отправке в небуферизированный канал без получателя?",
				Options:     []string{"Ничего", "Deadlock — горутина заблокируется навечно", "Данные потеряются", "Ошибка компиляции"},
				Correct:     1,
				Explanation: "Небуферизированный канал — рандеву. Отправитель ждёт получателя. Если получателя нет и других горутин нет — deadlock.",
			},
			{
				Question:    "Кто должен закрывать канал?",
				Options:     []string{"Получатель", "Отправитель — только он знает когда данных больше нет", "Любой", "Runtime"},
				Correct:     1,
				Explanation: "Отправитель закрывает канал, потому что только он знает когда данные закончились. Получатель проверяет через range или val, ok := <-ch.",
			},
			{
				Question:    "Сколько горутин можно создать на машине с 8ГБ RAM?",
				Options:     []string{"8", "1000", "Миллионы — стек горутины начинается с 2КБ", "Зависит от CPU"},
				Correct:     2,
				Explanation: "2КБ × 1 000 000 = 2ГБ. На 8ГБ машине — миллионы горутин. Но каждая потребляет CPU при активной работе.",
			},
		},
		Tasks: []T{
			{
				Title: "Ping-Pong на каналах", Difficulty: "easy",
				Glossary: []GlossaryItem{
					{Term: "ch <- val", Definition: "Отправить val в канал. Блокируется пока кто-то не прочитает (для unbuffered)."},
					{Term: "<-ch", Definition: "Прочитать из канала. Блокируется пока кто-то не отправит."},
				},
				Description: `<p>Две горутины играют в пинг-понг через канал. Первая отправляет "ping", вторая отвечает "pong". N раундов.</p>
<p>Ввод: <code>3</code></p>
<p>Вывод:</p>
<pre><code>ping
pong
ping
pong
ping
pong</code></pre>`,
				StarterCode: `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	ping := make(chan struct{})
	pong := make(chan struct{})
	done := make(chan struct{})

	// Горутина "ping"
	go func() {
		for i := 0; i < n; i++ {
			// TODO: напечатай "ping", отправь в pong, подожди ping
			fmt.Println("ping")
			pong <- struct{}{}
			if i < n-1 { <-ping }
		}
		done <- struct{}{}
	}()

	// Горутина "pong"
	go func() {
		for i := 0; i < n; i++ {
			<-pong
			fmt.Println("pong")
			if i < n-1 { ping <- struct{}{} }
		}
	}()

	<-done
}`,
				TestCases: []TestCase{
					{Input: "3", ExpectedOutput: "ping\npong\nping\npong\nping\npong"},
					{Input: "1", ExpectedOutput: "ping\npong"},
				},
				Hints: `<p>ping горутина: println → pong <- → <-ping. pong горутина: <-pong → println → ping <-.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func main() {
	var n int; fmt.Scan(&n)
	ping := make(chan struct{}); pong := make(chan struct{}); done := make(chan struct{})
	go func() {
		for i := 0; i < n; i++ { fmt.Println("ping"); pong <- struct{}{}; if i < n-1 { <-ping } }
		done <- struct{}{}
	}()
	go func() {
		for i := 0; i < n; i++ { <-pong; fmt.Println("pong"); if i < n-1 { ping <- struct{}{} } }
	}()
	<-done
}</code></pre>`,
			},
			{
				Title: "Конвейер (pipeline) из каналов", Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "pipeline", Definition: "Цепочка горутин, соединённых каналами: stage1 → ch → stage2 → ch → stage3."},
					{Term: "close(ch)", Definition: "Закрывает канал. range по каналу завершится. Закрывает только отправитель."},
				},
				Description: `<p>Построй конвейер из 3 стадий через каналы:</p>
<ol>
<li><code>generate</code> — отправляет числа от 1 до N в канал</li>
<li><code>square</code> — читает, возводит в квадрат, отправляет дальше</li>
<li><code>print</code> — читает и печатает</li>
</ol>
<p>Ввод: <code>4</code> → Вывод: <code>1\n4\n9\n16</code></p>`,
				StarterCode: `package main

import "fmt"

func generate(n int) <-chan int {
	out := make(chan int)
	go func() {
		for i := 1; i <= n; i++ { out <- i }
		close(out)
	}()
	return out
}

func square(in <-chan int) <-chan int {
	out := make(chan int)
	// TODO: горутина: range in, отправляй v*v в out, потом close(out)
	return out
}

func main() {
	var n int
	fmt.Scan(&n)
	for val := range square(generate(n)) {
		fmt.Println(val)
	}
}`,
				TestCases: []TestCase{
					{Input: "4", ExpectedOutput: "1\n4\n9\n16"},
					{Input: "3", ExpectedOutput: "1\n4\n9"},
				},
				Hints: `<p>square: go func() { for v := range in { out <- v * v }; close(out) }()</p>`,
				Solution: `<pre><code>package main
import "fmt"
func generate(n int) <-chan int {
	out := make(chan int)
	go func() { for i := 1; i <= n; i++ { out <- i }; close(out) }()
	return out
}
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() { for v := range in { out <- v * v }; close(out) }()
	return out
}
func main() {
	var n int; fmt.Scan(&n)
	for val := range square(generate(n)) { fmt.Println(val) }
}</code></pre>`,
			},
		},
	}
}

func lesson_sync_patterns() L {
	return L{
		Slug: "sync-primitives", Title: "sync: Mutex, WaitGroup и защита данных", Order: 2,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>sync пакет — защита общих данных</h1>

<h2>Проблема: Race Condition</h2>
<p>Когда две горутины читают/пишут одни данные без синхронизации — <strong>гонка данных</strong>:</p>

<pre><code>// БАГИ: гонка данных — результат непредсказуем
counter := 0
for i := 0; i < 1000; i++ {
    go func() {
        counter++  // чтение + инкремент + запись — НЕ атомарно!
    }()
}
// counter может быть 980, 995, 1000 — каждый раз разный!

// Как найти: go test -race ./...
// WARNING: DATA RACE → строка и стек вызовов</code></pre>

<h2>sync.Mutex — взаимное исключение</h2>
<pre><code>type SafeCounter struct {
    mu    sync.Mutex
    value int
}

func (c *SafeCounter) Inc() {
    c.mu.Lock()         // захватить замок
    defer c.mu.Unlock() // ВСЕГДА defer Unlock!
    c.value++           // теперь безопасно
}

func (c *SafeCounter) Get() int {
    c.mu.Lock()
    defer c.mu.Unlock()
    return c.value
}</code></pre>

<p><strong>Правила Mutex:</strong></p>
<ul>
<li>Всегда <code>defer mu.Unlock()</code> — даже при panic</li>
<li>Не копируй struct с Mutex (передавай указатель)</li>
<li>Держи замок минимальное время</li>
<li>Не вызывай Lock() из горутины, уже держащей Lock() → deadlock</li>
</ul>

<h2>sync.RWMutex — много читателей, один писатель</h2>
<pre><code>type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()          // множество горутин могут RLock одновременно
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}

func (c *Cache) Set(key, val string) {
    c.mu.Lock()           // эксклюзивный доступ — ждёт пока все RLock отпустят
    defer c.mu.Unlock()
    c.data[key] = val
}

// Когда использовать RWMutex vs Mutex:
// 90% чтений, 10% записей → RWMutex (читатели не блокируют друг друга)
// 50/50 чтений/записей → обычный Mutex (RWMutex создаёт overhead)</code></pre>

<h2>sync.WaitGroup — ждать группу горутин</h2>
<pre><code>var wg sync.WaitGroup

for i := 0; i < 10; i++ {
    wg.Add(1)           // ДО запуска горутины!
    go func(id int) {
        defer wg.Done() // уменьшает счётчик
        process(id)
    }(i)                // передаём i как аргумент (копия!)
}

wg.Wait()  // блокируется пока счётчик не станет 0
fmt.Println("все горутины завершились")</code></pre>

<p><strong>Частая ошибка:</strong> wg.Add(1) ВНУТРИ горутины → горутина может не успеть вызвать Add до Wait.</p>

<h2>sync.Once — ровно один раз</h2>
<pre><code>var (
    once     sync.Once
    instance *Database
)

func GetDB() *Database {
    once.Do(func() {
        // Этот код выполнится РОВНО ОДИН раз,
        // даже если 100 горутин вызовут GetDB() одновременно.
        instance = connectToDatabase()
    })
    return instance
}

// Все остальные вызовы GetDB() мгновенно вернут instance</code></pre>

<h2>sync.Map — конкурентная map</h2>
<pre><code>// Обычная map + горутины = PANIC: concurrent map writes
// Решение 1: map + Mutex (чаще лучше)
// Решение 2: sync.Map (для специфичных сценариев)

var cache sync.Map

cache.Store("key", "value")              // запись
val, ok := cache.Load("key")             // чтение
cache.Delete("key")                       // удаление
cache.LoadOrStore("key", "default")       // получить или записать

// sync.Map хорош когда:
// 1. Ключи стабильны (много чтений, мало записей)
// 2. Горутины работают с разными ключами (мало конфликтов)
// В остальных случаях: map + RWMutex</code></pre>

<h2>Атомарные операции — без замков</h2>
<pre><code>import "sync/atomic"

var counter int64

// Атомарный инкремент — без Mutex, без блокировки
atomic.AddInt64(&counter, 1)

// Атомарное чтение
val := atomic.LoadInt64(&counter)

// Когда использовать:
// Простые счётчики, флаги → atomic (быстрее Mutex)
// Сложная логика (map, struct) → Mutex</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что произойдёт при concurrent map writes без защиты?",
				Options:     []string{"Данные перепутаются", "Panic: concurrent map writes — Go runtime обнаружит и упадёт", "Ничего", "Замедление"},
				Correct:     1,
				Explanation: "Go runtime детектирует конкурентную запись в map и вызывает panic. Это сделано намеренно — лучше упасть явно, чем молча повредить данные.",
			},
			{
				Question:    "Почему wg.Add(1) должен быть ДО go func()?",
				Options:     []string{"Для красоты", "Если Add внутри горутины — горутина может не успеть вызвать Add до того как Wait вернётся", "Не обязательно", "Для скорости"},
				Correct:     1,
				Explanation: "wg.Wait() может вернуться до того как горутина начнёт выполняться. Add() до go гарантирует что счётчик увеличен до запуска горутины.",
			},
			{
				Question:    "Когда sync.RWMutex лучше обычного Mutex?",
				Options:     []string{"Всегда", "Когда большинство операций — чтение (90%+ читателей)", "Когда много записей", "Для одной горутины"},
				Correct:     1,
				Explanation: "RWMutex позволяет множественное чтение без блокировки. Если 90% операций — чтение, читатели не блокируют друг друга. При частых записях overhead RWMutex больше чем выигрыш.",
			},
		},
		Tasks: []T{
			{
				Title: "Потокобезопасный счётчик", Difficulty: "easy",
				Glossary: []GlossaryItem{
					{Term: "sync.Mutex", Definition: "Взаимное исключение. Lock() → критическая секция → Unlock(). Только одна горутина внутри."},
				},
				Description: `<p>Создай SafeCounter с Mutex. N горутин инкрементируют счётчик. Выведи итог.</p>
<p>Ввод: <code>1000</code> (горутин) → Вывод: <code>1000</code></p>`,
				StarterCode: `package main

import (
	"fmt"
	"sync"
)

type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	// TODO: Lock, value++, Unlock
}

func (c *SafeCounter) Get() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func main() {
	var n int
	fmt.Scan(&n)

	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}

	wg.Wait()
	fmt.Println(counter.Get())
}`,
				TestCases: []TestCase{
					{Input: "1000", ExpectedOutput: "1000"},
					{Input: "500", ExpectedOutput: "500"},
				},
				Hints: `<p>c.mu.Lock(); c.value++; c.mu.Unlock() — или defer c.mu.Unlock()</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync")
type SafeCounter struct { mu sync.Mutex; value int }
func (c *SafeCounter) Inc() { c.mu.Lock(); defer c.mu.Unlock(); c.value++ }
func (c *SafeCounter) Get() int { c.mu.Lock(); defer c.mu.Unlock(); return c.value }
func main() {
	var n int; fmt.Scan(&n)
	counter := &SafeCounter{}; var wg sync.WaitGroup
	for i := 0; i < n; i++ { wg.Add(1); go func() { defer wg.Done(); counter.Inc() }() }
	wg.Wait(); fmt.Println(counter.Get())
}</code></pre>`,
			},
			{
				Title: "Worker Pool", Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "jobs := make(chan int)", Definition: "Канал заданий. Воркеры читают из него. close(jobs) — сигнал что заданий нет."},
					{Term: "range jobs", Definition: "Читает из канала пока он не закрыт. Идиома для воркеров."},
				},
				Description: `<p>Реализуй worker pool: N воркеров обрабатывают задания (умножают на 2) из канала. Выведи сумму результатов.</p>
<p>Ввод:</p>
<pre><code>3 5
1 2 3 4 5</code></pre>
<p>(3 воркера, 5 заданий)</p>
<p>Вывод: <code>30</code> (2+4+6+8+10)</p>`,
				StarterCode: `package main

import (
	"fmt"
	"sync"
)

func main() {
	var workers, n int
	fmt.Scan(&workers, &n)
	nums := make([]int, n)
	for i := range nums { fmt.Scan(&nums[i]) }

	jobs := make(chan int, n)
	results := make(chan int, n)

	// TODO: запусти workers горутин-воркеров
	// Каждый: for job := range jobs { results <- job * 2 }
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			// TODO: range jobs, отправляй job*2 в results
		}()
	}

	// Отправь задания
	for _, num := range nums { jobs <- num }
	close(jobs)

	// Закрой results после завершения воркеров
	go func() { wg.Wait(); close(results) }()

	sum := 0
	for v := range results { sum += v }
	fmt.Println(sum)
}`,
				TestCases: []TestCase{
					{Input: "3 5\n1 2 3 4 5", ExpectedOutput: "30"},
					{Input: "2 4\n10 20 30 40", ExpectedOutput: "200"},
				},
				Hints: `<p>Воркер: for job := range jobs { results <- job * 2 }. close(jobs) после отправки. wg.Wait() + close(results) в горутине.</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync")
func main() {
	var workers, n int; fmt.Scan(&workers, &n)
	nums := make([]int, n)
	for i := range nums { fmt.Scan(&nums[i]) }
	jobs := make(chan int, n); results := make(chan int, n)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ { wg.Add(1); go func() { defer wg.Done(); for j := range jobs { results <- j * 2 } }() }
	for _, num := range nums { jobs <- num }; close(jobs)
	go func() { wg.Wait(); close(results) }()
	sum := 0; for v := range results { sum += v }; fmt.Println(sum)
}</code></pre>`,
			},
		},
	}
}

func lesson_advanced_concurrency() L {
	return L{
		Slug: "select-patterns", Title: "Select, таймауты и продвинутые паттерны", Order: 3,
		Difficulty: "expert", Track: "backend",
		Content: `<h1>Select и продвинутые паттерны</h1>

<h2>Select — мультиплексирование каналов</h2>
<p><code>select</code> ждёт первый готовый канал:</p>

<pre><code>select {
case msg := <-msgCh:
    handleMessage(msg)
case err := <-errCh:
    handleError(err)
case <-ctx.Done():
    return ctx.Err()
case <-time.After(5 * time.Second):
    return errors.New("таймаут")
}

// Если несколько каналов готовы одновременно — Go выбирает случайный
// default — если ни один канал не готов (non-blocking)</code></pre>

<h2>Паттерн: Таймаут операции</h2>
<pre><code>func fetchWithTimeout(url string, timeout time.Duration) (string, error) {
    resultCh := make(chan string, 1)
    errCh := make(chan error, 1)

    go func() {
        result, err := fetch(url)
        if err != nil {
            errCh <- err
            return
        }
        resultCh <- result
    }()

    select {
    case result := <-resultCh:
        return result, nil
    case err := <-errCh:
        return "", err
    case <-time.After(timeout):
        return "", fmt.Errorf("timeout after %v", timeout)
    }
}</code></pre>

<h2>Паттерн: Fan-out / Fan-in</h2>
<pre><code>// Fan-out: одни данные → несколько обработчиков
func fanOut(in <-chan int, workers int) []<-chan int {
    channels := make([]<-chan int, workers)
    for i := 0; i < workers; i++ {
        ch := make(chan int)
        go func() {
            for val := range in {
                ch <- val * val
            }
            close(ch)
        }()
        channels[i] = ch
    }
    return channels
}

// Fan-in: несколько каналов → один
func fanIn(channels ...<-chan int) <-chan int {
    var wg sync.WaitGroup
    merged := make(chan int)
    for _, ch := range channels {
        wg.Add(1)
        go func(c <-chan int) {
            defer wg.Done()
            for val := range c {
                merged <- val
            }
        }(ch)
    }
    go func() {
        wg.Wait()
        close(merged)
    }()
    return merged
}</code></pre>

<h2>Паттерн: Semaphore (ограничение параллелизма)</h2>
<pre><code>// Не более 10 параллельных HTTP-запросов
sem := make(chan struct{}, 10)

for _, url := range urls {
    sem <- struct{}{}  // захватить слот (блокируется если 10 заняты)
    go func(u string) {
        defer func() { <-sem }()  // освободить слот
        fetch(u)
    }(url)
}</code></pre>

<h2>Паттерн: Graceful Shutdown</h2>
<pre><code>func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // Запускаем воркеров
    go worker(ctx, "worker-1")
    go worker(ctx, "worker-2")

    // Ждём SIGINT/SIGTERM
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    fmt.Println("shutting down...")
    cancel()  // отменяем context → все воркеры завершаются
    time.Sleep(2 * time.Second)  // даём время на cleanup
}

func worker(ctx context.Context, name string) {
    for {
        select {
        case <-ctx.Done():
            fmt.Printf("%s: stopped\n", name)
            return
        default:
            doWork()
        }
    }
}</code></pre>

<h2>Паттерн: Rate Limiter через Ticker</h2>
<pre><code>// Не более 10 запросов в секунду
limiter := time.NewTicker(100 * time.Millisecond)
defer limiter.Stop()

for _, req := range requests {
    <-limiter.C  // ждёт следующий тик (100мс)
    go process(req)
}</code></pre>

<h2>Race Detector — ваш лучший друг</h2>
<pre><code>// Запусти тесты с race detector:
go test -race ./...

// Запусти программу с race detector:
go run -race main.go

// Что он находит:
// - Конкурентная запись в map
// - Чтение/запись одной переменной из разных горутин без sync
// - Забытый Mutex.Lock

// WARNING: DATA RACE
// Write at 0x00c00001a0a0 by goroutine 7:
//   main.(*Counter).Inc()
//   main.go:15
// Previous read at 0x00c00001a0a0 by goroutine 6:
//   main.(*Counter).Get()
//   main.go:20</code></pre>

<h2>Чего НЕ делать</h2>
<pre><code>// ❌ Не используй time.Sleep для синхронизации
go doWork()
time.Sleep(time.Second)  // "должно хватить" — нет, не хватит!

// ❌ Не передавай Mutex по значению (копирует!)
type Bad struct { mu sync.Mutex }
func process(b Bad) { b.mu.Lock() }  // КОПИЯ мьютекса — не защищает!

// ❌ Не используй горутины для CPU-bound задач без ограничения
for _, item := range millionItems {
    go process(item)  // миллион горутин! Используй worker pool.
}

// ✅ Правило: горутина для I/O (HTTP, DB, файлы)
// ✅ Worker pool для CPU-bound задач</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что делает select с несколькими готовыми каналами?",
				Options:     []string{"Читает первый", "Выбирает случайный готовый канал — для fairness", "Читает все", "Ошибка"},
				Correct:     1,
				Explanation: "Go специально рандомизирует выбор среди готовых каналов. Это предотвращает starvation — ситуацию когда один канал всегда приоритетнее.",
			},
			{
				Question:    "Для чего используется буферизированный канал как семафор?",
				Options:     []string{"Для скорости", "Ограничить количество параллельных операций — sem <- struct{}{} блокируется когда лимит достигнут", "Для логирования", "Для кеширования"},
				Correct:     1,
				Explanation: "Канал с буфером N пропускает N горутин. N+1 горутина блокируется на отправке. Идеально для лимита параллельных HTTP-запросов, DB-соединений.",
			},
			{
				Question:    "Зачем go test -race?",
				Options:     []string{"Ускоряет тесты", "Находит гонки данных — конкурентные чтения/записи без синхронизации", "Проверяет покрытие", "Форматирует код"},
				Correct:     1,
				Explanation: "Race detector инструментирует код и отслеживает каждый доступ к памяти. Если две горутины обращаются к одной переменной без sync — WARNING: DATA RACE.",
			},
		},
		Tasks: []T{
			{
				Title: "Семафор: ограничение параллелизма", Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "sem := make(chan struct{}, N)", Definition: "Семафор — буферизированный канал. Ограничивает до N параллельных операций."},
				},
				Description: `<p>Обработай N чисел с ограничением параллелизма (макс K горутин одновременно). Каждое число × 3. Выведи сумму.</p>
<p>Ввод: <code>2 5\n1 2 3 4 5</code> (макс 2 параллельных, 5 чисел)</p>
<p>Вывод: <code>45</code> (3+6+9+12+15)</p>`,
				StarterCode: `package main

import (
	"fmt"
	"sync"
)

func main() {
	var maxParallel, n int
	fmt.Scan(&maxParallel, &n)
	nums := make([]int, n)
	for i := range nums { fmt.Scan(&nums[i]) }

	sem := make(chan struct{}, maxParallel)
	results := make(chan int, n)
	var wg sync.WaitGroup

	for _, num := range nums {
		wg.Add(1)
		sem <- struct{}{} // TODO: захвати слот семафора
		go func(v int) {
			defer wg.Done()
			defer func() { <-sem }() // TODO: освободи слот
			results <- v * 3
		}(num)
	}

	go func() { wg.Wait(); close(results) }()

	sum := 0
	for v := range results { sum += v }
	fmt.Println(sum)
}`,
				TestCases: []TestCase{
					{Input: "2 5\n1 2 3 4 5", ExpectedOutput: "45"},
					{Input: "1 3\n10 20 30", ExpectedOutput: "180"},
				},
				Hints: `<p>sem <- struct{}{} перед go. defer func() { <-sem }() внутри горутины.</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync")
func main() {
	var maxP, n int; fmt.Scan(&maxP, &n)
	nums := make([]int, n); for i := range nums { fmt.Scan(&nums[i]) }
	sem := make(chan struct{}, maxP); results := make(chan int, n); var wg sync.WaitGroup
	for _, num := range nums { wg.Add(1); sem <- struct{}{}
		go func(v int) { defer wg.Done(); defer func() { <-sem }(); results <- v * 3 }(num) }
	go func() { wg.Wait(); close(results) }()
	sum := 0; for v := range results { sum += v }; fmt.Println(sum)
}</code></pre>`,
			},
		},
	}
}
