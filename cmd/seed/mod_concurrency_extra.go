package main

// ════════════════════════════════════════════════════════════════
// Дополнительные уроки для модуля конкурентности (4-6)
// ════════════════════════════════════════════════════════════════

func lesson_race_conditions() L {
	return L{
		Slug: "race-conditions", Title: "Гонки данных и безопасность", Order: 4,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Гонки данных (Race Conditions)</h1>

<h2>Что такое гонка данных?</h2>
<p>Гонка — когда две горутины одновременно обращаются к переменной и хотя бы одна пишет. Результат непредсказуем:</p>
<pre><code>var counter int

func increment(wg *sync.WaitGroup) {
    defer wg.Done()
    for i := 0; i < 1000; i++ {
        counter++ // READ → MODIFY → WRITE — не атомарно!
    }
}
// 10 горутин × 1000 = ожидаем 10000, получаем ~7000-9000</code></pre>

<h2>Детектор: go run -race</h2>
<pre><code>go run -race main.go
// WARNING: DATA RACE
go test -race ./...  // ОБЯЗАТЕЛЬНО в CI</code></pre>

<h2>Решение 1: sync.Mutex</h2>
<pre><code>var mu sync.Mutex
mu.Lock()
counter++
mu.Unlock()</code></pre>

<h2>Решение 2: sync/atomic</h2>
<pre><code>var counter int64
atomic.AddInt64(&counter, 1) // одна CPU-инструкция</code></pre>

<h2>Решение 3: каналы (no shared state)</h2>
<pre><code>results := make(chan int, 10)
// каждая горутина отправляет свой результат — нет shared state</code></pre>

<h2>sync.RWMutex</h2>
<pre><code>var mu sync.RWMutex
mu.RLock()   // параллельное чтение OK
mu.RUnlock()
mu.Lock()    // эксклюзивная запись
mu.Unlock()</code></pre>

<h2>Читать глубже</h2>
<ul>
<li><a href="https://habr.com/ru/articles/412715/" target="_blank">Хабр: Гонки данных в Go</a></li>
<li><a href="https://metanit.com/go/golang/8.4.php" target="_blank">Metanit: Мьютексы</a></li>
</ul>`,

		Quiz: []Q{
			{Question: "Что такое гонка данных?", Options: []string{"Ошибка компиляции", "Две горутины обращаются к переменной без синхронизации, хотя бы одна пишет", "Медленная программа", "Deadlock"}, Correct: 1, Explanation: "Data race = concurrent unsynchronized access. Результат непредсказуем."},
			{Question: "Как обнаружить гонку?", Options: []string{"go build", "go run -race / go test -race", "fmt.Println", "Дебаггер"}, Correct: 1, Explanation: "-race использует ThreadSanitizer. Обязателен в CI."},
			{Question: "Когда atomic лучше Mutex?", Options: []string{"Всегда", "Для простых операций с одной переменной (счётчик)", "Никогда", "Только для строк"}, Correct: 1, Explanation: "Atomic = hardware-level. Быстрее для простых случаев. Mutex — для группы операций."},
			{Question: "Чем RWMutex лучше Mutex?", Options: []string{"Быстрее всегда", "Параллельное чтение, блокирует только запись", "Нет разницы", "Проще"}, Correct: 1, Explanation: "RWMutex: RLock — много читателей. Lock — один писатель. Идеален для cache."},
			{Question: "Идиоматичный подход к конкурентности в Go?", Options: []string{"Shared memory + Mutex", "Каналы — share by communicating", "Глобальные переменные", "Lock-free"}, Correct: 1, Explanation: "'Don't communicate by sharing memory; share memory by communicating.'"},
		},
		Tasks: []T{
			{
				Title: "Mutex счётчик", Difficulty: "easy",
				Description: `<p>10 горутин увеличивают counter по 1000 раз. Используй Mutex:</p><p>Вывод: <code>Counter: 10000</code></p>`,
				Glossary:    []GlossaryItem{{Term: "sync.Mutex", Definition: "Lock/Unlock. Только одна горутина между ними."}},
				TestCases:   []TestCase{{Input: "", ExpectedOutput: "Counter: 10000"}},
				StarterCode: `package main
import ("fmt"; "sync")
var (counter int; mu sync.Mutex)
func increment(wg *sync.WaitGroup) {
    defer wg.Done()
    for i := 0; i < 1000; i++ { mu.Lock(); counter++; mu.Unlock() }
}
func main() {
    wg := &sync.WaitGroup{}
    for i := 0; i < 10; i++ { wg.Add(1); go increment(wg) }
    wg.Wait()
    fmt.Printf("Counter: %d\n", counter)
}`,
				Hints: `<p>mu.Lock() перед counter++, mu.Unlock() после.</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync")
var (counter int; mu sync.Mutex)
func increment(wg *sync.WaitGroup) { defer wg.Done(); for i := 0; i < 1000; i++ { mu.Lock(); counter++; mu.Unlock() } }
func main() { wg := &sync.WaitGroup{}; for i := 0; i < 10; i++ { wg.Add(1); go increment(wg) }; wg.Wait(); fmt.Printf("Counter: %d\n", counter) }</code></pre>`,
			},
			{
				Title: "Atomic счётчик", Difficulty: "easy",
				Description: `<p>Перепиши через sync/atomic:</p><p>Вывод: <code>Counter: 10000</code></p>`,
				Glossary:    []GlossaryItem{{Term: "atomic.AddInt64", Definition: "Атомарный инкремент. Одна CPU-инструкция."}},
				TestCases:   []TestCase{{Input: "", ExpectedOutput: "Counter: 10000"}},
				StarterCode: `package main
import ("fmt"; "sync"; "sync/atomic")
var counter int64
func increment(wg *sync.WaitGroup) { defer wg.Done(); for i := 0; i < 1000; i++ { atomic.AddInt64(&counter, 1) } }
func main() { wg := &sync.WaitGroup{}; for i := 0; i < 10; i++ { wg.Add(1); go increment(wg) }; wg.Wait(); fmt.Printf("Counter: %d\n", atomic.LoadInt64(&counter)) }`,
				Hints: `<p>atomic.AddInt64(&counter, 1)</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync"; "sync/atomic")
var counter int64
func increment(wg *sync.WaitGroup) { defer wg.Done(); for i := 0; i < 1000; i++ { atomic.AddInt64(&counter, 1) } }
func main() { wg := &sync.WaitGroup{}; for i := 0; i < 10; i++ { wg.Add(1); go increment(wg) }; wg.Wait(); fmt.Printf("Counter: %d\n", atomic.LoadInt64(&counter)) }</code></pre>`,
			},
			{
				Title: "Thread-safe cache", Difficulty: "medium",
				Description: `<p>RWMutex cache: set/get команды:</p>
<p>Ввод:</p><pre><code>5
set name Alice
set age 25
get name
get age
get city</code></pre>
<p>Вывод:</p><pre><code>Alice
25
NOT FOUND</code></pre>`,
				Glossary:  []GlossaryItem{{Term: "RWMutex", Definition: "RLock — параллельное чтение. Lock — запись."}},
				TestCases: []TestCase{{Input: "5\nset name Alice\nset age 25\nget name\nget age\nget city", ExpectedOutput: "Alice\n25\nNOT FOUND"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings"; "sync")
type Cache struct { mu sync.RWMutex; data map[string]string }
func (c *Cache) Get(k string) (string, bool) { c.mu.RLock(); defer c.mu.RUnlock(); v, ok := c.data[k]; return v, ok }
func (c *Cache) Set(k, v string) { c.mu.Lock(); defer c.mu.Unlock(); c.data[k] = v }
func main() {
    cache := &Cache{data: make(map[string]string)}
    var n int; fmt.Scan(&n)
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); p := strings.Fields(scanner.Text())
        switch p[0] { case "set": cache.Set(p[1], p[2]); case "get": if v, ok := cache.Get(p[1]); ok { fmt.Println(v) } else { fmt.Println("NOT FOUND") } }
    }
}`,
				Hints: `<p>Get: RLock/RUnlock. Set: Lock/Unlock.</p>`,
				Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings"; "sync")
type Cache struct { mu sync.RWMutex; data map[string]string }
func (c *Cache) Get(k string) (string, bool) { c.mu.RLock(); defer c.mu.RUnlock(); v, ok := c.data[k]; return v, ok }
func (c *Cache) Set(k, v string) { c.mu.Lock(); defer c.mu.Unlock(); c.data[k] = v }
func main() { cache := &Cache{data: make(map[string]string)}; var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); p := strings.Fields(scanner.Text()); switch p[0] { case "set": cache.Set(p[1], p[2]); case "get": if v, ok := cache.Get(p[1]); ok { fmt.Println(v) } else { fmt.Println("NOT FOUND") } } } }</code></pre>`,
			},
			{
				Title: "Параллельная сумма", Difficulty: "medium",
				Description: `<p>Раздели массив на 2 части, посчитай сумму параллельно через канал:</p>
<p>Ввод: <code>10 1 2 3 4 5 6 7 8 9 10</code></p><p>Вывод: <code>Sum: 55</code></p>`,
				Glossary:  []GlossaryItem{{Term: "Fan-out", Definition: "Разбить работу → горутины → собрать результаты."}},
				TestCases: []TestCase{{Input: "10 1 2 3 4 5 6 7 8 9 10", ExpectedOutput: "Sum: 55"}},
				StarterCode: `package main
import "fmt"
func partialSum(nums []int, ch chan<- int) { s := 0; for _, n := range nums { s += n }; ch <- s }
func main() {
    var n int; fmt.Scan(&n); nums := make([]int, n); for i := range nums { fmt.Scan(&nums[i]) }
    ch := make(chan int, 2); mid := n / 2
    go partialSum(nums[:mid], ch); go partialSum(nums[mid:], ch)
    fmt.Printf("Sum: %d\n", <-ch + <-ch)
}`,
				Hints: `<p>Раздели пополам, две горутины, собери 2 результата из канала.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func partialSum(nums []int, ch chan<- int) { s := 0; for _, n := range nums { s += n }; ch <- s }
func main() { var n int; fmt.Scan(&n); nums := make([]int, n); for i := range nums { fmt.Scan(&nums[i]) }; ch := make(chan int, 2); mid := n/2; go partialSum(nums[:mid], ch); go partialSum(nums[mid:], ch); fmt.Printf("Sum: %d\n", <-ch + <-ch) }</code></pre>`,
			},
			{
				Title: "sync.Map для счётчика URL", Difficulty: "hard",
				Description: `<p>sync.Map + atomic для подсчёта посещений URL:</p>
<p>Ввод:</p><pre><code>6
/home
/about
/home
/home
/about
/contact</code></pre>
<p>Вывод:</p><pre><code>/about: 2
/contact: 1
/home: 3</code></pre>`,
				Glossary:  []GlossaryItem{{Term: "sync.Map", Definition: "Потокобезопасный map. LoadOrStore, Store, Load, Range."}},
				TestCases: []TestCase{{Input: "6\n/home\n/about\n/home\n/home\n/about\n/contact", ExpectedOutput: "/about: 2\n/contact: 1\n/home: 3"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "sort"; "sync"; "sync/atomic")
func main() {
    var visits sync.Map; var n int; fmt.Scan(&n); var wg sync.WaitGroup
    scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); url := scanner.Text(); wg.Add(1)
        go func(u string) { defer wg.Done(); val, _ := visits.LoadOrStore(u, new(int64)); atomic.AddInt64(val.(*int64), 1) }(url) }
    wg.Wait()
    var keys []string; visits.Range(func(k, v any) bool { keys = append(keys, k.(string)); return true })
    sort.Strings(keys)
    for _, k := range keys { val, _ := visits.Load(k); fmt.Printf("%s: %d\n", k, atomic.LoadInt64(val.(*int64))) }
}`,
				Hints: `<p>LoadOrStore: вернёт existing или сохранит new(int64). Потом atomic.AddInt64.</p>`,
				Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "sort"; "sync"; "sync/atomic")
func main() { var visits sync.Map; var n int; fmt.Scan(&n); var wg sync.WaitGroup; scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan(); url := scanner.Text(); wg.Add(1); go func(u string) { defer wg.Done(); val, _ := visits.LoadOrStore(u, new(int64)); atomic.AddInt64(val.(*int64), 1) }(url) }
    wg.Wait(); var keys []string; visits.Range(func(k, v any) bool { keys = append(keys, k.(string)); return true }); sort.Strings(keys)
    for _, k := range keys { val, _ := visits.Load(k); fmt.Printf("%s: %d\n", k, atomic.LoadInt64(val.(*int64))) } }</code></pre>`,
			},
		},
	}
}

func lesson_worker_pool() L {
	return L{
		Slug: "worker-pool", Title: "Worker Pool и Fan-out/Fan-in", Order: 5,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Worker Pool — производственный паттерн</h1>

<h2>Зачем?</h2>
<p>Нельзя запустить 1M горутин для HTTP — убьёшь сервер. Worker Pool ограничивает параллелизм:</p>
<pre><code>const workers = 10
jobs := make(chan string, 100)
for i := 0; i < workers; i++ { go worker(jobs, results) }
for _, url := range urls { jobs <- url }
close(jobs)</code></pre>

<h2>Semaphore через буферизированный канал</h2>
<pre><code>sem := make(chan struct{}, 10)
sem <- struct{}{} // захват
<-sem             // освобождение</code></pre>

<h2>Pipeline</h2>
<pre><code>func generate(n int) <-chan int { ... }
func square(in <-chan int) <-chan int { ... }
// Цепочка: for v := range square(generate(n)) { ... }</code></pre>

<h2>errgroup</h2>
<pre><code>g, ctx := errgroup.WithContext(ctx)
g.SetLimit(5)
g.Go(func() error { return fetch(ctx, url) })
err := g.Wait()</code></pre>`,

		Quiz: []Q{
			{Question: "Зачем Worker Pool если горутины дешёвые?", Options: []string{"Не нужен", "Ресурсы (TCP, файлы) ограничены — Pool контролирует параллелизм", "Для красоты", "Только CPU"}, Correct: 1, Explanation: "Горутины дешёвые, TCP/файлы — нет. Pool = не более N одновременных I/O операций."},
			{Question: "Как реализовать semaphore в Go?", Options: []string{"sync.Semaphore", "make(chan struct{}, N) — буферизированный канал", "runtime.SetMax", "Mutex+counter"}, Correct: 1, Explanation: "Идиоматичный семафор: chan struct{} с буфером N."},
			{Question: "Что такое Fan-out/Fan-in?", Options: []string{"Вентиляторы", "Fan-out: 1→N воркеров. Fan-in: N→1 канал результатов", "Типы каналов", "БД-паттерн"}, Correct: 1, Explanation: "Fan-out распределяет. Fan-in собирает. Вместе = Worker Pool."},
			{Question: "Зачем close(jobs)?", Options: []string{"Экономия памяти", "Сигнал воркерам: задач больше нет, range завершится", "Обязательно", "Скорость"}, Correct: 1, Explanation: "Без close — воркеры навечно ждут. close + range = чистое завершение."},
			{Question: "errgroup.SetLimit(5)?", Options: []string{"5 ошибок", "Макс 5 горутин одновременно", "5 попыток", "Таймаут"}, Correct: 1, Explanation: "errgroup из x/sync — Worker Pool + первая ошибка. SetLimit = семафор внутри."},
		},
		Tasks: []T{
			{
				Title: "Базовый Worker Pool", Difficulty: "easy",
				Description: `<p>3 воркера удваивают числа. Выведи сумму:</p><p>Ввод: <code>6 1 2 3 4 5 6</code></p><p>Вывод: <code>Sum: 42</code></p>`,
				Glossary:    []GlossaryItem{{Term: "worker(jobs <-chan int, results chan<- int)", Definition: "Воркер: range jobs → process → results."}},
				TestCases:   []TestCase{{Input: "6 1 2 3 4 5 6", ExpectedOutput: "Sum: 42"}},
				StarterCode: `package main
import "fmt"
func worker(jobs <-chan int, results chan<- int) { for j := range jobs { results <- j * 2 } }
func main() {
    var n int; fmt.Scan(&n); nums := make([]int, n); for i := range nums { fmt.Scan(&nums[i]) }
    jobs := make(chan int, n); results := make(chan int, n)
    for w := 0; w < 3; w++ { go worker(jobs, results) }
    for _, num := range nums { jobs <- num }; close(jobs)
    sum := 0; for i := 0; i < n; i++ { sum += <-results }; fmt.Printf("Sum: %d\n", sum)
}`,
				Hints: `<p>3 воркера, close(jobs), собрать N результатов.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func worker(jobs <-chan int, results chan<- int) { for j := range jobs { results <- j * 2 } }
func main() { var n int; fmt.Scan(&n); nums := make([]int, n); for i := range nums { fmt.Scan(&nums[i]) }; jobs := make(chan int, n); results := make(chan int, n); for w := 0; w < 3; w++ { go worker(jobs, results) }; for _, num := range nums { jobs <- num }; close(jobs); sum := 0; for i := 0; i < n; i++ { sum += <-results }; fmt.Printf("Sum: %d\n", sum) }</code></pre>`,
			},
			{
				Title: "Semaphore", Difficulty: "easy",
				Description: `<p>Ограничь параллелизм до 2. Обработай N задач:</p><p>Ввод: <code>5</code></p><p>Вывод: <code>done: 5</code></p>`,
				Glossary:    []GlossaryItem{{Term: "make(chan struct{}, N)", Definition: "Семафор на N слотов."}},
				TestCases:   []TestCase{{Input: "5", ExpectedOutput: "done: 5"}},
				StarterCode: `package main
import ("fmt"; "sync"; "sync/atomic")
func main() {
    var n int; fmt.Scan(&n); sem := make(chan struct{}, 2); var wg sync.WaitGroup; var done int64
    for i := 0; i < n; i++ { wg.Add(1); sem <- struct{}{}
        go func() { defer wg.Done(); defer func() { <-sem }(); atomic.AddInt64(&done, 1) }() }
    wg.Wait(); fmt.Printf("done: %d\n", atomic.LoadInt64(&done))
}`,
				Hints: `<p>sem <- struct{}{} захват. <-sem освобождение в defer.</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync"; "sync/atomic")
func main() { var n int; fmt.Scan(&n); sem := make(chan struct{}, 2); var wg sync.WaitGroup; var done int64; for i := 0; i < n; i++ { wg.Add(1); sem <- struct{}{}; go func() { defer wg.Done(); defer func() { <-sem }(); atomic.AddInt64(&done, 1) }() }; wg.Wait(); fmt.Printf("done: %d\n", atomic.LoadInt64(&done)) }</code></pre>`,
			},
			{
				Title: "Pipeline", Difficulty: "medium",
				Description: `<p>generate → square → print:</p><p>Ввод: <code>5</code></p><p>Вывод: <code>1 4 9 16 25</code></p>`,
				Glossary:    []GlossaryItem{{Term: "Pipeline", Definition: "Цепочка стадий через каналы. Каждая — горутина."}},
				TestCases:   []TestCase{{Input: "5", ExpectedOutput: "1 4 9 16 25"}},
				StarterCode: `package main
import "fmt"
func generate(n int) <-chan int { out := make(chan int); go func() { for i := 1; i <= n; i++ { out <- i }; close(out) }(); return out }
func square(in <-chan int) <-chan int { out := make(chan int); go func() { for v := range in { out <- v * v }; close(out) }(); return out }
func main() { var n int; fmt.Scan(&n); first := true; for v := range square(generate(n)) { if !first { fmt.Print(" ") }; fmt.Print(v); first = false }; fmt.Println() }`,
				Hints: `<p>Каждая стадия: make(chan), go func(){range in → out}, close(out), return out.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func generate(n int) <-chan int { out := make(chan int); go func() { for i := 1; i <= n; i++ { out <- i }; close(out) }(); return out }
func square(in <-chan int) <-chan int { out := make(chan int); go func() { for v := range in { out <- v*v }; close(out) }(); return out }
func main() { var n int; fmt.Scan(&n); first := true; for v := range square(generate(n)) { if !first { fmt.Print(" ") }; fmt.Print(v); first = false }; fmt.Println() }</code></pre>`,
			},
			{
				Title: "Fan-in: merge каналов", Difficulty: "hard",
				Description: `<p>3 producer-а по N чисел. fanIn объединяет:</p><p>Ввод: <code>3</code></p><p>Вывод: <code>Count: 9</code></p>`,
				Glossary:    []GlossaryItem{{Term: "Fan-in", Definition: "N каналов → 1. WaitGroup + close(merged)."}},
				TestCases:   []TestCase{{Input: "3", ExpectedOutput: "Count: 9"}},
				StarterCode: `package main
import ("fmt"; "sync")
func producer(id, count int) <-chan int { ch := make(chan int); go func() { for i := 0; i < count; i++ { ch <- id*100+i }; close(ch) }(); return ch }
func fanIn(chs ...<-chan int) <-chan int { var wg sync.WaitGroup; merged := make(chan int); for _, ch := range chs { wg.Add(1); go func(c <-chan int) { defer wg.Done(); for v := range c { merged <- v } }(ch) }; go func() { wg.Wait(); close(merged) }(); return merged }
func main() { var n int; fmt.Scan(&n); count := 0; for range fanIn(producer(1,n), producer(2,n), producer(3,n)) { count++ }; fmt.Printf("Count: %d\n", count) }`,
				Hints: `<p>fanIn: для каждого канала горутина с range. WaitGroup → close(merged).</p>`,
				Solution: `<pre><code>package main
import ("fmt"; "sync")
func producer(id, count int) <-chan int { ch := make(chan int); go func() { for i := 0; i < count; i++ { ch <- id*100+i }; close(ch) }(); return ch }
func fanIn(chs ...<-chan int) <-chan int { var wg sync.WaitGroup; merged := make(chan int); for _, ch := range chs { wg.Add(1); go func(c <-chan int) { defer wg.Done(); for v := range c { merged <- v } }(ch) }; go func() { wg.Wait(); close(merged) }(); return merged }
func main() { var n int; fmt.Scan(&n); count := 0; for range fanIn(producer(1,n), producer(2,n), producer(3,n)) { count++ }; fmt.Printf("Count: %d\n", count) }</code></pre>`,
			},
			{
				Title: "Rate limiter", Difficulty: "hard",
				Description: `<p>Обработай N задач с ограничением скорости:</p><p>Ввод: <code>6</code></p><p>Вывод: <code>Processed: 6</code></p>`,
				Glossary:    []GlossaryItem{{Term: "time.Ticker", Definition: "Канал с периодическим сигналом. Для rate limiting."}},
				TestCases:   []TestCase{{Input: "6", ExpectedOutput: "Processed: 6"}},
				StarterCode: `package main
import "fmt"
func main() { var n int; fmt.Scan(&n); processed := 0; for i := 0; i < n; i++ { processed++ }; fmt.Printf("Processed: %d\n", processed) }`,
				Hints: `<p>В реальности: ticker := time.NewTicker(interval); <-ticker.C перед операцией.</p>`,
				Solution: `<pre><code>package main
import "fmt"
func main() { var n int; fmt.Scan(&n); p := 0; for i := 0; i < n; i++ { p++ }; fmt.Printf("Processed: %d\n", p) }</code></pre>`,
			},
		},
	}
}

func lesson_context_concurrency() L {
	return L{
		Slug: "context-concurrency", Title: "Context: отмена, таймауты, graceful shutdown", Order: 6,
		Difficulty: "advanced", Track: "backend",
		Content: `<h1>Context в конкурентности</h1>

<h2>Зачем?</h2>
<p>Без context горутину нельзя отменить, ограничить по времени, передать deadline.</p>
<pre><code>ctx, cancel := context.WithCancel(context.Background())
go func(ctx context.Context) {
    for { select { case <-ctx.Done(): return; default: doWork() } }
}(ctx)
cancel() // все горутины с этим ctx завершатся</code></pre>

<h2>WithTimeout / WithDeadline</h2>
<pre><code>ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel() // ВСЕГДА</code></pre>

<h2>Graceful Shutdown</h2>
<pre><code>ctx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
defer stop()
<-ctx.Done() // ждём сигнал
srv.Shutdown(timeoutCtx)</code></pre>

<h2>Правила</h2>
<ul>
<li>ctx — первый параметр: func Do(ctx context.Context, ...)</li>
<li>Не хранить в struct</li>
<li>Всегда defer cancel()</li>
<li>Проверять ctx.Done() в длинных операциях</li>
</ul>`,

		Quiz: []Q{
			{Question: "Зачем defer cancel() даже если таймаут сработает?", Options: []string{"Красота", "Освободить ресурсы runtime (внутренние горутины)", "Ускоряет", "Необязательно"}, Correct: 1, Explanation: "WithTimeout запускает горутину. cancel() её останавливает. Без cancel = утечка."},
			{Question: "WithTimeout vs WithDeadline?", Options: []string{"Нет разницы", "Timeout = через N от сейчас. Deadline = в конкретный момент", "Timeout быстрее", "Deadline для HTTP"}, Correct: 1, Explanation: "WithTimeout(5s) = WithDeadline(now+5s). Внутри одно и то же."},
			{Question: "Как горутина узнаёт об отмене?", Options: []string{"panic", "select { case <-ctx.Done(): }", "Глобальная переменная", "return"}, Correct: 1, Explanation: "ctx.Done() — канал, закрывается при отмене. select проверяет."},
			{Question: "Хранить context в struct?", Options: []string{"Да", "Нет — передавать первым аргументом", "Только HTTP", "Только тесты"}, Correct: 1, Explanation: "Context = lifecycle запроса. В struct = потенциальная утечка/протухший ctx."},
			{Question: "ctx.Err() после cancel()?", Options: []string{"nil", "context.Canceled", "panic", "строка"}, Correct: 1, Explanation: "cancel() → Canceled. Timeout → DeadlineExceeded. До отмены → nil."},
		},
		Tasks: []T{
			{
				Title: "Cancel горутины", Difficulty: "easy",
				Description: `<p>Запусти горутину, отмени через cancel():</p><p>Вывод: <code>Stopped</code></p>`,
				Glossary:    []GlossaryItem{{Term: "context.WithCancel", Definition: "ctx + cancel(). cancel() закрывает Done()."}},
				TestCases:   []TestCase{{Input: "", ExpectedOutput: "Stopped"}},
				StarterCode: `package main
import ("context"; "fmt")
func counter(ctx context.Context, done chan<- struct{}) { for { select { case <-ctx.Done(): done <- struct{}{}; return; default: } } }
func main() { ctx, cancel := context.WithCancel(context.Background()); done := make(chan struct{}); go counter(ctx, done); cancel(); <-done; fmt.Println("Stopped") }`,
				Hints: `<p>select { case <-ctx.Done(): return }</p>`,
				Solution: `<pre><code>package main
import ("context"; "fmt")
func counter(ctx context.Context, done chan<- struct{}) { for { select { case <-ctx.Done(): done <- struct{}{}; return; default: } } }
func main() { ctx, cancel := context.WithCancel(context.Background()); done := make(chan struct{}); go counter(ctx, done); cancel(); <-done; fmt.Println("Stopped") }</code></pre>`,
			},
			{
				Title: "Timeout операции", Difficulty: "easy",
				Description: `<p>WithTimeout: fast → результат, slow → timeout:</p>
<p>Ввод: <code>fast</code> → <code>Result: done</code></p><p>Ввод: <code>slow</code> → <code>timeout</code></p>`,
				Glossary:  []GlossaryItem{{Term: "WithTimeout", Definition: "Автоотмена через duration."}},
				TestCases: []TestCase{{Input: "fast", ExpectedOutput: "Result: done"}, {Input: "slow", ExpectedOutput: "timeout"}},
				StarterCode: `package main
import ("context"; "fmt"; "time")
func operation(ctx context.Context, mode string) (string, error) {
    ch := make(chan string, 1); go func() { if mode == "slow" { time.Sleep(200*time.Millisecond) }; ch <- "done" }()
    select { case r := <-ch: return r, nil; case <-ctx.Done(): return "", ctx.Err() }
}
func main() { var mode string; fmt.Scan(&mode); ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond); defer cancel()
    r, err := operation(ctx, mode); if err != nil { fmt.Println("timeout") } else { fmt.Printf("Result: %s\n", r) } }`,
				Hints: `<p>select { case r := <-ch: ok; case <-ctx.Done(): timeout }</p>`,
				Solution: `<pre><code>package main
import ("context"; "fmt"; "time")
func operation(ctx context.Context, mode string) (string, error) { ch := make(chan string, 1); go func() { if mode == "slow" { time.Sleep(200*time.Millisecond) }; ch <- "done" }(); select { case r := <-ch: return r, nil; case <-ctx.Done(): return "", ctx.Err() } }
func main() { var m string; fmt.Scan(&m); ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond); defer cancel(); r, err := operation(ctx, m); if err != nil { fmt.Println("timeout") } else { fmt.Printf("Result: %s\n", r) } }</code></pre>`,
			},
			{
				Title: "Context propagation", Difficulty: "medium",
				Description: `<p>Передай ctx по цепочке, проверяй Done():</p><p>Ввод: <code>3</code></p><p>Вывод: <code>Processed: 3</code></p>`,
				Glossary:    []GlossaryItem{{Term: "propagation", Definition: "ctx передаётся вниз. Отмена наверху → все ниже завершаются."}},
				TestCases:   []TestCase{{Input: "3", ExpectedOutput: "Processed: 3"}},
				StarterCode: `package main
import ("context"; "fmt")
func process(ctx context.Context, items []int) int {
    count := 0; for range items { select { case <-ctx.Done(): return count; default: count++ } }; return count
}
func main() { var n int; fmt.Scan(&n); items := make([]int, n); fmt.Printf("Processed: %d\n", process(context.Background(), items)) }`,
				Hints: `<p>select { case <-ctx.Done(): early exit; default: work }</p>`,
				Solution: `<pre><code>package main
import ("context"; "fmt")
func process(ctx context.Context, items []int) int { c := 0; for range items { select { case <-ctx.Done(): return c; default: c++ } }; return c }
func main() { var n int; fmt.Scan(&n); fmt.Printf("Processed: %d\n", process(context.Background(), make([]int, n))) }</code></pre>`,
			},
			{
				Title: "Worker Pool + context", Difficulty: "hard",
				Description: `<p>Воркеры проверяют ctx.Done(). При отмене — чистый выход:</p><p>Ввод: <code>5</code></p><p>Вывод: <code>Done: 5</code></p>`,
				Glossary:    []GlossaryItem{{Term: "Worker+ctx", Definition: "select { case <-ctx.Done(): return; case job := <-jobs: process }"}},
				TestCases:   []TestCase{{Input: "5", ExpectedOutput: "Done: 5"}},
				StarterCode: `package main
import ("context"; "fmt"; "sync")
func worker(ctx context.Context, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
    defer wg.Done(); for { select { case <-ctx.Done(): return; case j, ok := <-jobs: if !ok { return }; results <- j*2 } }
}
func main() { var n int; fmt.Scan(&n); ctx := context.Background(); jobs := make(chan int, n); results := make(chan int, n); var wg sync.WaitGroup
    for w := 0; w < 3; w++ { wg.Add(1); go worker(ctx, jobs, results, &wg) }
    for i := 1; i <= n; i++ { jobs <- i }; close(jobs); go func() { wg.Wait(); close(results) }()
    count := 0; for range results { count++ }; fmt.Printf("Done: %d\n", count) }`,
				Hints: `<p>select с двумя case: ctx.Done() и jobs. close(jobs) → ok=false → return.</p>`,
				Solution: `<pre><code>package main
import ("context"; "fmt"; "sync")
func worker(ctx context.Context, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) { defer wg.Done(); for { select { case <-ctx.Done(): return; case j, ok := <-jobs: if !ok { return }; results <- j*2 } } }
func main() { var n int; fmt.Scan(&n); jobs := make(chan int, n); results := make(chan int, n); var wg sync.WaitGroup
    for w := 0; w < 3; w++ { wg.Add(1); go worker(context.Background(), jobs, results, &wg) }
    for i := 1; i <= n; i++ { jobs <- i }; close(jobs); go func() { wg.Wait(); close(results) }()
    c := 0; for range results { c++ }; fmt.Printf("Done: %d\n", c) }</code></pre>`,
			},
			{
				Title: "Graceful shutdown", Difficulty: "hard",
				Description: `<p>Фоновая работа → shutdown → вывод:</p><p>Ввод: <code>3</code></p><p>Вывод: <code>Shutdown complete: 3 tasks done</code></p>`,
				Glossary:    []GlossaryItem{{Term: "Graceful shutdown", Definition: "Сигнал → прекращаем новые → дожидаемся текущие → выход."}},
				TestCases:   []TestCase{{Input: "3", ExpectedOutput: "Shutdown complete: 3 tasks done"}},
				StarterCode: `package main
import ("context"; "fmt"; "sync/atomic")
func main() { var n int; fmt.Scan(&n); _, cancel := context.WithCancel(context.Background()); var done int64; finished := make(chan struct{})
    go func() { for i := 0; i < n; i++ { atomic.AddInt64(&done, 1) }; finished <- struct{}{} }()
    <-finished; cancel(); fmt.Printf("Shutdown complete: %d tasks done\n", atomic.LoadInt64(&done)) }`,
				Hints: `<p>Горутина работает → finished. Main ждёт → cancel → print.</p>`,
				Solution: `<pre><code>package main
import ("context"; "fmt"; "sync/atomic")
func main() { var n int; fmt.Scan(&n); _, cancel := context.WithCancel(context.Background()); var done int64; finished := make(chan struct{})
    go func() { for i := 0; i < n; i++ { atomic.AddInt64(&done, 1) }; finished <- struct{}{} }()
    <-finished; cancel(); fmt.Printf("Shutdown complete: %d tasks done\n", atomic.LoadInt64(&done)) }</code></pre>`,
			},
		},
	}
}
