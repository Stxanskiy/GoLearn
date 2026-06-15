package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Nginx — от архитектуры до продакшн-конфигурации
// ════════════════════════════════════════════════════════════════

func mod_nginx() M {
	return M{
		Slug:          "nginx",
		Title:         "Nginx — веб-сервер и reverse proxy",
		Description:   "Архитектура event-driven, reverse proxy, балансировка нагрузки, SSL/TLS, кэширование, rate limiting. Реальные продакшн-конфиги и подводные камни.",
		Order:         20,
		Track:         "devops",
		Difficulty:    "intermediate",
		Prerequisites: []string{"linux-fundamentals", "http-server"},
		Lessons: []L{
			lesson_nginx_architecture(),
			lesson_nginx_config_basics(),
			lesson_nginx_reverse_proxy(),
			lesson_nginx_load_balancing(),
			lesson_nginx_ssl_tls(),
			lesson_nginx_performance(),
			lesson_nginx_rate_limiting(),
		},
	}
}

// ── Урок 1: Архитектура Nginx ──────────────────────────────────

func lesson_nginx_architecture() L {
	return L{
		Slug: "nginx-architecture", Title: "Архитектура Nginx: event-driven модель", Order: 1,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Архитектура Nginx: event-driven модель</h1>

<h2>Проблема C10K</h2>
<p>В конце 1990-х появилась <strong>проблема C10K</strong> (concurrent 10,000 connections): как обслужить 10 000 одновременных соединений на одном сервере? Apache с моделью "один процесс/поток на соединение" потреблял огромное количество памяти и не масштабировался.</p>

<p>Каждый поток в Apache занимает ~2-8 МБ памяти. При 10 000 соединениях это 20-80 ГБ RAM только на потоки. Большинство потоков при этом <strong>простаивают</strong> — ждут данных от клиента или от бэкенда.</p>

<h2>Решение: event-driven архитектура</h2>
<p>Nginx использует принципиально другой подход — <strong>событийно-ориентированную</strong> (event-driven) модель:</p>

<pre><code>                    ┌───────────────────────────┐
                    │      Master Process        │
                    │  (читает конфиг, управляет │
                    │   worker-ами, bind портов) │
                    └─────────────┬─────────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
     ┌────────▼────────┐ ┌───────▼────────┐ ┌───────▼────────┐
     │  Worker Process  │ │ Worker Process │ │ Worker Process │
     │  (event loop)    │ │ (event loop)   │ │ (event loop)   │
     │  epoll/kqueue    │ │ epoll/kqueue   │ │ epoll/kqueue   │
     │  1000+ conn      │ │ 1000+ conn     │ │ 1000+ conn     │
     └─────────────────┘ └────────────────┘ └────────────────┘</code></pre>

<h2>Master и Worker процессы</h2>
<p><strong>Master process</strong> (PID 1 в контейнере или основной процесс):</p>
<ul>
<li>Читает и валидирует конфигурацию</li>
<li>Биндит порты (80, 443)</li>
<li>Запускает и управляет worker-процессами</li>
<li>Обрабатывает сигналы (reload, stop, reopen logs)</li>
<li>НЕ обрабатывает клиентские запросы</li>
</ul>

<p><strong>Worker process</strong> (обычно = количеству CPU ядер):</p>
<ul>
<li>Обрабатывает все клиентские соединения</li>
<li>Использует event loop (epoll на Linux, kqueue на BSD/macOS)</li>
<li>Один worker обрабатывает тысячи соединений одновременно</li>
<li>Никогда не блокируется на I/O — всё через асинхронные callbacks</li>
</ul>

<h2>Event Loop — как это работает</h2>
<pre><code>while (true) {
    events = epoll_wait(epoll_fd, timeout);  // Ждём события
    for each event in events {
        if (event.type == NEW_CONNECTION) {
            accept(event.fd);                // Принимаем соединение
            register_read_event(event.fd);   // Регистрируем на чтение
        }
        if (event.type == READABLE) {
            data = read(event.fd);           // Читаем данные (не блокируемся)
            process_request(data);           // Обрабатываем запрос
            register_write_event(event.fd);  // Регистрируем на запись
        }
        if (event.type == WRITABLE) {
            write(event.fd, response);       // Пишем ответ
            close_or_keepalive(event.fd);
        }
    }
}</code></pre>

<p>Ключевое отличие от Apache: worker <strong>никогда не ждёт</strong>. Если данные не готовы — он переходит к следующему событию. Когда данные придут — ОС уведомит через epoll.</p>

<h2>Сравнение с Apache</h2>
<table>
<tr><th>Характеристика</th><th>Apache (prefork/worker)</th><th>Nginx</th></tr>
<tr><td>Модель</td><td>Процесс/поток на соединение</td><td>Event loop, неблокирующий I/O</td></tr>
<tr><td>Память на 10K соединений</td><td>~20-80 ГБ</td><td>~200-300 МБ</td></tr>
<tr><td>Переключение контекста</td><td>Высокое (OS scheduler)</td><td>Минимальное (userspace)</td></tr>
<tr><td>Статика</td><td>Через модули, медленнее</td><td>Нативно, sendfile(), zero-copy</td></tr>
<tr><td>Динамический контент</td><td>mod_php встроен в процесс</td><td>Только через FastCGI/proxy_pass</td></tr>
<tr><td>.htaccess</td><td>Да (удобно, но медленно)</td><td>Нет (всё в одном конфиге)</td></tr>
</table>

<h2>Сигналы управления</h2>
<pre><code># Graceful reload — плавная перезагрузка конфига
nginx -s reload    # или: kill -HUP <master_pid>

# Graceful stop — дождаться завершения текущих запросов
nginx -s quit      # или: kill -QUIT <master_pid>

# Немедленная остановка
nginx -s stop      # или: kill -TERM <master_pid>

# Переоткрытие лог-файлов (для logrotate)
nginx -s reopen    # или: kill -USR1 <master_pid></code></pre>

<p><strong>Что происходит при reload:</strong></p>
<ol>
<li>Master получает SIGHUP</li>
<li>Проверяет новый конфиг синтаксически</li>
<li>Если ок — запускает новые worker-ы с новым конфигом</li>
<li>Старые worker-ы получают сигнал "graceful shutdown"</li>
<li>Старые worker-ы дорабатывают текущие запросы и завершаются</li>
<li><strong>Ноль downtime</strong></li>
</ol>

<h2>Проверка количества worker-ов</h2>
<pre><code># Посмотреть процессы nginx
ps aux | grep nginx
# root     12345  master process /usr/sbin/nginx
# www-data 12346  worker process
# www-data 12347  worker process
# www-data 12348  worker process
# www-data 12349  worker process

# Или через конфиг
grep worker_processes /etc/nginx/nginx.conf
# worker_processes auto;   — по количеству CPU ядер
# worker_processes 4;      — фиксированное значение</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему Nginx потребляет значительно меньше памяти чем Apache при большом количестве соединений?",
				Options:     []string{"Nginx написан на более эффективном языке", "Nginx использует event loop и один worker обрабатывает тысячи соединений, а Apache создаёт отдельный процесс/поток на каждое", "Nginx сжимает данные в памяти", "Nginx использует swap вместо RAM"},
				Correct:     1,
				Explanation: "Apache в модели prefork/worker создаёт отдельный процесс (2-8 МБ) или поток на каждое соединение. При 10K соединений это 20-80 ГБ. Nginx использует event loop с epoll/kqueue — один worker-процесс обрабатывает тысячи соединений без создания новых потоков.",
			},
			{
				Question:    "Что произойдёт, если при nginx -s reload новый конфиг содержит синтаксическую ошибку?",
				Options:     []string{"Nginx упадёт", "Старая конфигурация продолжит работать, ошибка будет записана в error.log", "Nginx перезапустится с частично невалидным конфигом", "Worker-процессы начнут падать один за другим"},
				Correct:     1,
				Explanation: "Master-процесс ВСЕГДА проверяет синтаксис нового конфига ПЕРЕД запуском новых worker-ов. Если конфиг невалиден — reload не произойдёт, старые worker-ы продолжат работать с предыдущей конфигурацией. Это фундаментальная гарантия zero-downtime.",
			},
			{
				Question:    "Какую системную функцию использует Nginx для эффективного мультиплексирования соединений на Linux?",
				Options:     []string{"select()", "poll()", "epoll_wait()", "fork()"},
				Correct:     2,
				Explanation: "На Linux Nginx использует epoll — интерфейс ядра для эффективного мониторинга тысяч файловых дескрипторов. В отличие от select/poll (O(n) при каждом вызове), epoll работает O(1) — возвращает только готовые дескрипторы. На BSD/macOS используется kqueue.",
			},
			{
				Question:    "Почему worker_processes обычно устанавливают равным количеству CPU ядер?",
				Options:     []string{"Это ограничение ОС", "Больше worker-ов чем ядер создаёт конкуренцию за CPU и контекстные переключения, меньше — недоиспользует ресурсы", "Nginx не может запустить больше worker-ов", "Для совместимости с Docker"},
				Correct:     1,
				Explanation: "Каждый worker — это один процесс, привязанный к event loop. Он потребляет 100% одного ядра CPU при нагрузке. Больше worker-ов чем ядер = OS scheduler тратит время на переключение контекста. Меньше = ядра простаивают. worker_processes auto определяет это автоматически.",
			},
			{
				Question:    "Чем sendfile() отличается от обычного чтения файла?",
				Options:     []string{"Ничем", "sendfile() передаёт файл из ядра в сеть БЕЗ копирования в userspace — zero-copy, в 2-3 раза быстрее для статики", "sendfile() сжимает данные", "sendfile() шифрует"},
				Correct:     1,
				Explanation: "Обычно: ядро читает файл → копирует в user buffer → приложение копирует обратно в ядро → отправка. sendfile(): ядро сразу из файлового кеша в сетевой буфер. Нет лишних копирований. Nginx включает это по умолчанию для статики.",
			},
		},
		Tasks: []T{
			{
				Title:      "Симуляция event loop",
				Difficulty: "medium",
				Description: `<p>Реализуйте упрощённую модель event loop Nginx.</p>
<p>На вход подаётся список событий (построчно) в формате: <code>тип_события fd данные</code></p>
<p>Типы событий:</p>
<ul>
<li><code>CONNECT fd</code> — новое соединение, вывести: <code>[fd] connected</code></li>
<li><code>READ fd данные</code> — данные получены, вывести: <code>[fd] read: данные</code></li>
<li><code>WRITE fd данные</code> — данные отправлены, вывести: <code>[fd] write: данные</code></li>
<li><code>CLOSE fd</code> — соединение закрыто, вывести: <code>[fd] closed</code></li>
</ul>
<p>Также подсчитайте и выведите в конце: <code>total_events: N, active_connections: M</code></p>
<p>Где M — количество соединений которые были CONNECT но не CLOSE.</p>`,
				Hints: `<p>Используйте map[string]bool для отслеживания активных соединений. При CONNECT ставьте true, при CLOSE удаляйте. В конце len(map) = active.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Event Loop", Definition: "Цикл обработки событий — центральная абстракция неблокирующего I/O. Один поток обрабатывает множество соединений, переключаясь между готовыми."},
					{Term: "File Descriptor (fd)", Definition: "Целочисленный идентификатор открытого ресурса (сокета, файла) в Unix. Каждое соединение = отдельный fd."},
					{Term: "epoll", Definition: "Механизм ядра Linux для эффективного мониторинга множества fd. Возвращает только те, которые готовы к I/O."},
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
	active := make(map[string]bool)
	totalEvents := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		eventType := parts[0]
		fd := parts[1]
		totalEvents++

		// TODO: обработать каждый тип события
		// CONNECT — добавить fd в active, вывести "[fd] connected"
		// READ — вывести "[fd] read: данные"
		// WRITE — вывести "[fd] write: данные"
		// CLOSE — удалить fd из active, вывести "[fd] closed"
		_ = eventType
		_ = fd
	}

	// TODO: вывести total_events и active_connections
	_ = active
	_ = totalEvents
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	active := make(map[string]bool)
	totalEvents := 0

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 3)
		eventType := parts[0]
		fd := parts[1]
		totalEvents++

		switch eventType {
		case "CONNECT":
			active[fd] = true
			fmt.Printf("[%s] connected\n", fd)
		case "READ":
			data := parts[2]
			fmt.Printf("[%s] read: %s\n", fd, data)
		case "WRITE":
			data := parts[2]
			fmt.Printf("[%s] write: %s\n", fd, data)
		case "CLOSE":
			delete(active, fd)
			fmt.Printf("[%s] closed\n", fd)
		}
	}

	fmt.Printf("total_events: %d, active_connections: %d\n", totalEvents, len(active))
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "CONNECT 5\nREAD 5 GET /index.html\nWRITE 5 200 OK\nCONNECT 6\nREAD 6 GET /api/users\nCLOSE 5\nWRITE 6 200 OK\nCLOSE 6",
						ExpectedOutput: "[5] connected\n[5] read: GET /index.html\n[5] write: 200 OK\n[6] connected\n[6] read: GET /api/users\n[5] closed\n[6] write: 200 OK\n[6] closed\ntotal_events: 8, active_connections: 0",
					},
					{
						Input:          "CONNECT 1\nCONNECT 2\nCONNECT 3\nREAD 1 hello\nCLOSE 1",
						ExpectedOutput: "[1] connected\n[2] connected\n[3] connected\n[1] read: hello\n[1] closed\ntotal_events: 5, active_connections: 2",
					},
				},
			},
			{
				Title:      "Анализ worker_processes",
				Difficulty: "easy",
				Description: `<p>Напишите программу, которая принимает количество CPU ядер и текущую нагрузку (запросов/сек), и выдаёт рекомендацию по настройке worker_processes и worker_connections.</p>
<p>Формат входа: <code>cpu_cores requests_per_sec</code></p>
<p>Правила:</p>
<ul>
<li><code>worker_processes</code> = cpu_cores (всегда)</li>
<li><code>worker_connections</code>: если rps &lt;= 1000 → 1024, если rps &lt;= 10000 → 4096, иначе → 8192</li>
<li><code>max_clients</code> = worker_processes * worker_connections</li>
</ul>
<p>Формат вывода (3 строки):</p>
<pre><code>worker_processes N;
worker_connections N;
# max_clients: N</code></pre>`,
				Hints: `<p>Просто прочитайте два числа через fmt.Scan, примените условия для worker_connections и выведите результат.</p>`,
				Glossary: []GlossaryItem{
					{Term: "worker_processes", Definition: "Директива Nginx — количество worker-процессов. Обычно = количество CPU ядер."},
					{Term: "worker_connections", Definition: "Максимальное количество одновременных соединений на один worker. max_clients = worker_processes * worker_connections."},
				},
				StarterCode: `package main

import "fmt"

func main() {
	var cpuCores, rps int
	fmt.Scan(&cpuCores, &rps)

	workerProcesses := cpuCores
	var workerConnections int

	// TODO: определить workerConnections по rps
	// rps <= 1000 → 1024
	// rps <= 10000 → 4096
	// иначе → 8192

	_ = workerProcesses
	_ = workerConnections

	// TODO: вывести результат в формате:
	// worker_processes N;
	// worker_connections N;
	// # max_clients: N
}`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
	var cpuCores, rps int
	fmt.Scan(&cpuCores, &rps)

	workerProcesses := cpuCores
	var workerConnections int

	switch {
	case rps <= 1000:
		workerConnections = 1024
	case rps <= 10000:
		workerConnections = 4096
	default:
		workerConnections = 8192
	}

	maxClients := workerProcesses * workerConnections
	fmt.Printf("worker_processes %d;\n", workerProcesses)
	fmt.Printf("worker_connections %d;\n", workerConnections)
	fmt.Printf("# max_clients: %d\n", maxClients)
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "4 500",
						ExpectedOutput: "worker_processes 4;\nworker_connections 1024;\n# max_clients: 4096",
					},
					{
						Input:          "8 5000",
						ExpectedOutput: "worker_processes 8;\nworker_connections 4096;\n# max_clients: 32768",
					},
					{
						Input:          "16 50000",
						ExpectedOutput: "worker_processes 16;\nworker_connections 8192;\n# max_clients: 131072",
					},
				},
			},
			{
				Title:      "Apache vs Nginx сравнение",
					Difficulty: "easy",
					Description: `<p>По характеристикам определи Apache или Nginx:</p>
<p>Ввод:</p><pre><code>3
event-driven
process-per-connection
zero-copy sendfile</code></pre>
<p>Вывод:</p><pre><code>event-driven: Nginx
process-per-connection: Apache
zero-copy sendfile: Nginx</code></pre>`,
					Glossary: []GlossaryItem{
						{Term: "event-driven", Definition: "Один процесс обрабатывает множество соединений через event loop — модель Nginx."},
					},
					TestCases: []TestCase{
						{Input: "3\nevent-driven\nprocess-per-connection\nzero-copy sendfile", ExpectedOutput: "event-driven: Nginx\nprocess-per-connection: Apache\nzero-copy sendfile: Nginx"},
					},
					StarterCode: `package main
import ("bufio"; "fmt"; "os")
func main() {
    nginx := map[string]bool{"event-driven": true, "zero-copy sendfile": true, "non-blocking I/O": true, "epoll": true, "no .htaccess": true}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); f := sc.Text()
        if nginx[f] { fmt.Printf("%s: Nginx\n", f) } else { fmt.Printf("%s: Apache\n", f) }
    }
}`,
					Hints: `<p>Map с Nginx-характеристиками. Остальное — Apache.</p>`,
					Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os")
func main() { ng := map[string]bool{"event-driven":true,"zero-copy sendfile":true,"non-blocking I/O":true,"epoll":true}; var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); f := sc.Text(); if ng[f] { fmt.Printf("%s: Nginx\n", f) } else { fmt.Printf("%s: Apache\n", f) } } }</code></pre>`,
				},
				{
					Title:      "Nginx signal handler",
					Difficulty: "medium",
					Description: `<p>По сигналу определи что произойдёт с Nginx:</p>
<p>Ввод:</p><pre><code>4
reload
stop
quit
reopen</code></pre>
<p>Вывод:</p><pre><code>reload (HUP): graceful config reload, zero downtime
stop (TERM): immediate shutdown
quit (QUIT): graceful shutdown, finish current requests
reopen (USR1): reopen log files</code></pre>`,
					Glossary: []GlossaryItem{
						{Term: "nginx -s reload", Definition: "SIGHUP → проверить конфиг → новые workers → старые дорабатывают → zero downtime."},
					},
					TestCases: []TestCase{
						{Input: "4\nreload\nstop\nquit\nreopen", ExpectedOutput: "reload (HUP): graceful config reload, zero downtime\nstop (TERM): immediate shutdown\nquit (QUIT): graceful shutdown, finish current requests\nreopen (USR1): reopen log files"},
					},
					StarterCode: `package main
import ("bufio"; "fmt"; "os")
func main() {
    signals := map[string]string{
        "reload": "reload (HUP): graceful config reload, zero downtime",
        "stop": "stop (TERM): immediate shutdown",
        "quit": "quit (QUIT): graceful shutdown, finish current requests",
        "reopen": "reopen (USR1): reopen log files",
    }
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); fmt.Println(signals[sc.Text()]) }
}`,
					Hints: `<p>Map signal → описание с Unix-сигналом.</p>`,
					Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os")
func main() { s := map[string]string{"reload":"reload (HUP): graceful config reload, zero downtime","stop":"stop (TERM): immediate shutdown","quit":"quit (QUIT): graceful shutdown, finish current requests","reopen":"reopen (USR1): reopen log files"}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin); for i := 0; i < n; i++ { sc.Scan(); fmt.Println(s[sc.Text()]) } }</code></pre>`,
				},
				{
					Title:      "Capacity planner",
					Difficulty: "hard",
					Description: `<p>Рассчитай capacity Nginx сервера:</p>
<p>Ввод: <code>8 4096 500</code> (cores, worker_connections, avg_request_ms)</p>
<p>Вывод:</p><pre><code>Workers: 8
Max concurrent: 32768
Theoretical RPS: 65536
Bottleneck: CPU-bound at 8 cores</code></pre>`,
					Glossary: []GlossaryItem{
						{Term: "Capacity planning", Definition: "max_concurrent = workers * connections. RPS = max_concurrent * (1000/avg_ms)."},
					},
					TestCases: []TestCase{
						{Input: "8 4096 500", ExpectedOutput: "Workers: 8\nMax concurrent: 32768\nTheoretical RPS: 65536\nBottleneck: CPU-bound at 8 cores"},
					},
					StarterCode: `package main
import "fmt"
func main() {
    var cores, conns, avgMs int; fmt.Scan(&cores, &conns, &avgMs)
    maxConc := cores * conns
    rps := maxConc * (1000 / avgMs)
    fmt.Printf("Workers: %d\nMax concurrent: %d\nTheoretical RPS: %d\nBottleneck: CPU-bound at %d cores\n", cores, maxConc, rps, cores)
}`,
					Hints: `<p>RPS = max_concurrent * (1000ms / avg_request_ms).</p>`,
					Solution: `<pre><code>package main
import "fmt"
func main() { var c, conn, ms int; fmt.Scan(&c, &conn, &ms); mc := c * conn; fmt.Printf("Workers: %d\nMax concurrent: %d\nTheoretical RPS: %d\nBottleneck: CPU-bound at %d cores\n", c, mc, mc*(1000/ms), c) }</code></pre>`,
				},
			},
	}
}

// ── Урок 2: Основы конфигурации ────────────────────────────────

func lesson_nginx_config_basics() L {
	return L{
		Slug: "nginx-config-basics", Title: "Конфигурация Nginx: структура и контексты", Order: 2,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Конфигурация Nginx: структура и контексты</h1>

<h2>Файловая структура</h2>
<p>Конфигурация Nginx обычно живёт в:</p>
<pre><code>/etc/nginx/
├── nginx.conf              # Главный файл конфигурации
├── conf.d/                 # Дополнительные конфиги (*.conf подключаются автоматически)
│   ├── default.conf
│   └── mysite.conf
├── sites-available/        # Все доступные виртуальные хосты (Debian/Ubuntu)
│   ├── default
│   └── mysite.com
├── sites-enabled/          # Симлинки на активные хосты
│   └── mysite.com -> ../sites-available/mysite.com
├── snippets/               # Переиспользуемые фрагменты конфига
│   ├── ssl-params.conf
│   └── proxy-params.conf
└── mime.types              # MIME-типы для расширений файлов</code></pre>

<h2>Иерархия контекстов</h2>
<p>Nginx конфиг — это дерево <strong>контекстов</strong> (блоков), вложенных друг в друга:</p>
<pre><code># Глобальный контекст (main)
worker_processes auto;
error_log /var/log/nginx/error.log warn;

events {                          # Контекст events
    worker_connections 1024;
    multi_accept on;
}

http {                            # Контекст http
    include /etc/nginx/mime.types;
    sendfile on;
    keepalive_timeout 65;
    gzip on;

    server {                      # Контекст server (виртуальный хост)
        listen 80;
        server_name example.com;
        root /var/www/example;

        location / {              # Контекст location (URL-путь)
            try_files $uri $uri/ =404;
        }

        location /api/ {          # Другой location
            proxy_pass http://backend:8080;
        }

        location ~* \.(jpg|png|css|js)$ {  # Регулярное выражение
            expires 30d;
            add_header Cache-Control "public, immutable";
        }
    }
}</code></pre>

<h2>Типы директив</h2>
<p><strong>Простые директивы</strong> — оканчиваются точкой с запятой:</p>
<pre><code>worker_processes 4;
listen 80;
server_name example.com;</code></pre>

<p><strong>Блочные директивы</strong> — содержат другие директивы в фигурных скобках:</p>
<pre><code>server {
    location / {
        ...
    }
}</code></pre>

<h2>Наследование директив</h2>
<p>Директивы наследуются от родительского контекста к дочернему. Дочерний может переопределить:</p>
<pre><code>http {
    gzip on;                      # Включено для ВСЕХ серверов

    server {
        server_name api.example.com;
        gzip off;                 # Переопределено — выключено для этого сервера
    }

    server {
        server_name web.example.com;
        # gzip on — унаследовано от http
    }
}</code></pre>

<h2>Приоритет location</h2>
<p>Nginx выбирает location по приоритету (от высшего к низшему):</p>
<ol>
<li><code>location = /exact</code> — точное совпадение (наивысший приоритет)</li>
<li><code>location ^~ /prefix</code> — префикс без проверки regex</li>
<li><code>location ~* \.php$</code> — регулярное выражение (case-insensitive)</li>
<li><code>location ~ \.php$</code> — регулярное выражение (case-sensitive)</li>
<li><code>location /prefix</code> — обычный префикс (выбирается самый длинный)</li>
</ol>

<pre><code># Примеры — какой location сработает для /images/photo.jpg?
location / { }              # Совпадает (/)
location /images/ { }       # Совпадает (/images/) — длиннее
location ~* \.(jpg|png)$ {} # Совпадает (regex)
# Ответ: regex победит обычный префикс, но ^~ /images/ победил бы regex</code></pre>

<h2>Проверка конфигурации</h2>
<pre><code># Проверить синтаксис (ОБЯЗАТЕЛЬНО перед reload!)
nginx -t
# nginx: the configuration file /etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /etc/nginx/nginx.conf test is successful

# Показать итоговую конфигурацию (с includes)
nginx -T

# Проверить конкретный файл
nginx -t -c /path/to/custom.conf</code></pre>

<h2>Переменные Nginx</h2>
<pre><code># Встроенные переменные:
$host           # Имя хоста из запроса (Host header)
$uri            # Текущий URI (может меняться rewrite-ами)
$request_uri    # Оригинальный URI запроса (с аргументами)
$remote_addr    # IP клиента
$scheme         # http или https
$args           # Query string (?key=value)
$request_method # GET, POST, PUT...

# Пользовательские переменные:
set $backend "http://127.0.0.1:8080";
proxy_pass $backend;</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Какой location из перечисленных будет выбран для запроса GET /api/users/123?",
				Options:     []string{"location /", "location /api/", "location = /api/users/123", "location ~ /api/.*"},
				Correct:     2,
				Explanation: "Точное совпадение (=) имеет наивысший приоритет. location = /api/users/123 матчит запрос полностью, поэтому побеждает и префикс /api/, и regex. Если бы точного совпадения не было — regex победил бы обычный префикс.",
			},
			{
				Question:    "Что произойдёт если выполнить nginx -s reload без предварительной проверки nginx -t и конфиг невалиден?",
				Options:     []string{"Nginx упадёт и перестанет обслуживать запросы", "Master-процесс сам проверит конфиг и откажется от reload, старая конфигурация продолжит работать", "Nginx применит только валидные части конфига", "Worker-процессы перезапустятся с ошибками"},
				Correct:     1,
				Explanation: "nginx -s reload = kill -HUP master. Master ВСЕГДА проверяет конфиг перед reload. Даже без явного nginx -t, master сам сделает проверку. Если невалидно — reload не произойдёт. Но лучше всегда делать nginx -t перед reload для раннего обнаружения ошибок.",
			},
			{
				Question:    "Зачем используются sites-available и sites-enabled как два отдельных каталога?",
				Options:     []string{"Для разделения прав доступа между пользователями", "Чтобы можно было хранить все конфиги сайтов, но активировать только нужные через симлинки — включение/выключение сайта без удаления конфига", "Это требование безопасности", "Для ускорения парсинга конфига"},
				Correct:     1,
				Explanation: "sites-available хранит ВСЕ конфиги. sites-enabled содержит только симлинки на активные. Отключить сайт = удалить симлинку (конфиг остаётся). Включить = создать симлинку. Это паттерн из Debian/Ubuntu, в RHEL/CentOS используют conf.d/.",
			},
		},
		Tasks: []T{
			{
				Title:      "Парсер nginx.conf",
				Difficulty: "medium",
				Description: `<p>Напишите парсер упрощённого формата nginx.conf.</p>
<p>На вход подаётся конфигурация. Нужно извлечь все директивы <code>server_name</code> и <code>listen</code> из каждого блока <code>server { }</code>.</p>
<p>Формат вывода для каждого server-блока (по одной строке):</p>
<pre><code>server: listen=PORT server_name=NAME</code></pre>
<p>Если директивы нет — используйте значение по умолчанию: listen=80, server_name=_</p>
<p>Входной формат упрощён: каждая директива на отдельной строке, блоки открываются и закрываются на отдельных строках.</p>`,
				Hints: `<p>Используйте стек или счётчик глубины для отслеживания вхождения в блок server. Внутри server ищите строки начинающиеся с listen или server_name.</p>`,
				Glossary: []GlossaryItem{
					{Term: "server_name", Definition: "Директива Nginx определяющая доменное имя виртуального хоста. Nginx выбирает server-блок по Host header запроса."},
					{Term: "listen", Definition: "Директива определяющая порт (и опционально IP) на котором server-блок принимает соединения."},
					{Term: "Виртуальный хост", Definition: "Возможность обслуживать несколько доменов на одном IP-адресе. Каждый server-блок = один виртуальный хост."},
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
	inServer := false
	listen := "80"
	serverName := "_"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// TODO: Определите вход в блок server (строка == "server {")
		// TODO: Внутри server ищите директивы listen и server_name
		// TODO: При выходе из блока (строка == "}") выведите результат
		// и сбросьте значения по умолчанию
		_ = line
		_ = inServer
		_ = listen
		_ = serverName
	}
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	inServer := false
	listen := "80"
	serverName := "_"

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "server {" {
			inServer = true
			listen = "80"
			serverName = "_"
			continue
		}

		if inServer && line == "}" {
			fmt.Printf("server: listen=%s server_name=%s\n", listen, serverName)
			inServer = false
			continue
		}

		if inServer {
			if strings.HasPrefix(line, "listen ") {
				listen = strings.TrimSuffix(strings.TrimPrefix(line, "listen "), ";")
			}
			if strings.HasPrefix(line, "server_name ") {
				serverName = strings.TrimSuffix(strings.TrimPrefix(line, "server_name "), ";")
			}
		}
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "server {\nlisten 80;\nserver_name example.com;\n}\nserver {\nlisten 443;\nserver_name api.example.com;\n}",
						ExpectedOutput: "server: listen=80 server_name=example.com\nserver: listen=443 server_name=api.example.com",
					},
					{
						Input:          "server {\nserver_name test.local;\n}\nserver {\nlisten 8080;\n}",
						ExpectedOutput: "server: listen=80 server_name=test.local\nserver: listen=8080 server_name=_",
					},
				},
			},
			{
				Title:      "Матчер location приоритетов",
				Difficulty: "hard",
				Description: `<p>Реализуйте алгоритм выбора location блока Nginx.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество location правил N</li>
<li>Следующие N строк — правила в формате: <code>тип паттерн</code> (тип: exact, prefix, regex)</li>
<li>Последняя строка — URI запроса для проверки</li>
</ol>
<p><strong>Приоритеты (от высшего к низшему):</strong></p>
<ol>
<li><code>exact</code> — точное совпадение (=)</li>
<li><code>regex</code> — регулярное выражение (~)</li>
<li><code>prefix</code> — самый длинный совпавший префикс</li>
</ol>
<p>Выведите номер сработавшего правила (1-based) или <code>no match</code>.</p>`,
				Hints: `<p>Проверяйте в порядке приоритета: сначала все exact, потом regex (первый совпавший), потом prefix (самый длинный). Для regex используйте regexp.MatchString.</p>`,
				Glossary: []GlossaryItem{
					{Term: "location = /path", Definition: "Точное совпадение URI. Наивысший приоритет. Если URI == /path — сработает."},
					{Term: "location ~ regex", Definition: "Регулярное выражение (case-sensitive). Проверяются в порядке конфига, первое совпадение побеждает."},
					{Term: "location /prefix", Definition: "Префиксный матч. Побеждает самый длинный совпавший префикс."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	type rule struct {
		ruleType string // "exact", "prefix", "regex"
		pattern  string
		index    int
	}

	rules := make([]rule, 0, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(scanner.Text(), " ", 2)
		rules = append(rules, rule{ruleType: parts[0], pattern: parts[1], index: i + 1})
	}

	scanner.Scan()
	uri := scanner.Text()

	// TODO: Реализуйте алгоритм выбора location
	// 1. Проверить все exact — если есть точное совпадение, вернуть его
	// 2. Проверить все regex по порядку — первое совпадение побеждает
	// 3. Проверить все prefix — самый длинный совпавший побеждает
	// 4. Если ничего — "no match"
	_ = uri
	_ = rules
	_ = regexp.MatchString
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	type rule struct {
		ruleType string
		pattern  string
		index    int
	}

	rules := make([]rule, 0, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(scanner.Text(), " ", 2)
		rules = append(rules, rule{ruleType: parts[0], pattern: parts[1], index: i + 1})
	}

	scanner.Scan()
	uri := scanner.Text()

	// 1. Exact match
	for _, r := range rules {
		if r.ruleType == "exact" && uri == r.pattern {
			fmt.Println(r.index)
			return
		}
	}

	// 2. Regex (first match wins)
	for _, r := range rules {
		if r.ruleType == "regex" {
			matched, _ := regexp.MatchString(r.pattern, uri)
			if matched {
				fmt.Println(r.index)
				return
			}
		}
	}

	// 3. Prefix (longest match wins)
	bestIdx := -1
	bestLen := 0
	for _, r := range rules {
		if r.ruleType == "prefix" && strings.HasPrefix(uri, r.pattern) {
			if len(r.pattern) > bestLen {
				bestLen = len(r.pattern)
				bestIdx = r.index
			}
		}
	}

	if bestIdx != -1 {
		fmt.Println(bestIdx)
	} else {
		fmt.Println("no match")
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "4\nprefix /\nprefix /api/\nregex \\.(jpg|png)$\nexact /api/users/123\n/api/users/123",
						ExpectedOutput: "4",
					},
					{
						Input:          "3\nprefix /\nprefix /images/\nregex \\.(jpg|png)$\n/images/photo.jpg",
						ExpectedOutput: "3",
					},
					{
						Input:          "2\nprefix /api/\nprefix /api/v2/\n/api/v2/users",
						ExpectedOutput: "2",
					},
				},
			},
		},
	}
}

// ── Урок 3: Reverse Proxy ──────────────────────────────────────

func lesson_nginx_reverse_proxy() L {
	return L{
		Slug: "nginx-reverse-proxy", Title: "Reverse Proxy: проксирование запросов", Order: 3,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Reverse Proxy: проксирование запросов</h1>

<h2>Что такое Reverse Proxy?</h2>
<p><strong>Reverse proxy</strong> — сервер, который принимает запросы от клиентов и перенаправляет их на backend-серверы. Клиент не знает о существовании backend-ов — он общается только с proxy.</p>

<pre><code>                                    ┌─────────────┐
                                ┌──▶│  Backend 1  │ (Go app :8080)
┌────────┐      ┌─────────┐    │   └─────────────┘
│ Client │─────▶│  Nginx  │────┤
└────────┘      │  :443   │    │   ┌─────────────┐
                └─────────┘    └──▶│  Backend 2  │ (Go app :8081)
                                   └─────────────┘</code></pre>

<p><strong>Зачем нужен:</strong></p>
<ul>
<li><strong>SSL termination</strong> — Nginx обрабатывает HTTPS, backend работает по HTTP</li>
<li><strong>Балансировка нагрузки</strong> — распределение запросов между несколькими backend-ами</li>
<li><strong>Кэширование</strong> — хранение ответов backend-а</li>
<li><strong>Защита</strong> — backend не виден из интернета напрямую</li>
<li><strong>Статика</strong> — Nginx отдаёт файлы сам, не нагружая backend</li>
</ul>

<h2>Базовая конфигурация proxy_pass</h2>
<pre><code>server {
    listen 80;
    server_name api.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;

        # Передаём оригинальные заголовки клиента
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}</code></pre>

<h2>Важные заголовки</h2>
<p>Без явной передачи заголовков backend не знает реальный IP клиента и протокол:</p>
<table>
<tr><th>Заголовок</th><th>Назначение</th><th>Значение</th></tr>
<tr><td>X-Real-IP</td><td>Реальный IP клиента</td><td>$remote_addr</td></tr>
<tr><td>X-Forwarded-For</td><td>Цепочка прокси (client, proxy1, proxy2)</td><td>$proxy_add_x_forwarded_for</td></tr>
<tr><td>X-Forwarded-Proto</td><td>Оригинальный протокол (http/https)</td><td>$scheme</td></tr>
<tr><td>Host</td><td>Оригинальный Host header</td><td>$host</td></tr>
</table>

<p><strong>Типичная ошибка:</strong> без <code>proxy_set_header Host $host</code> backend получит Host: 127.0.0.1:8080 вместо api.example.com. Это ломает виртуальные хосты на backend-е.</p>

<h2>Upstream блоки</h2>
<pre><code>upstream backend_pool {
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    server 10.0.0.3:8080 backup;   # Используется только если остальные недоступны
    server 10.0.0.4:8080 down;     # Временно исключён

    keepalive 32;    # Пул постоянных соединений к backend-ам
}

server {
    location /api/ {
        proxy_pass http://backend_pool;
        proxy_http_version 1.1;                    # Нужно для keepalive
        proxy_set_header Connection "";             # Сброс для keepalive
    }
}</code></pre>

<h2>Health Checks</h2>
<pre><code>upstream backend {
    server 10.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 10.0.0.2:8080 max_fails=3 fail_timeout=30s;
}

# max_fails=3 — после 3 неудачных попыток сервер считается "down"
# fail_timeout=30s — сервер "down" на 30 секунд, потом снова проверяется</code></pre>

<h2>WebSocket проксирование</h2>
<pre><code>location /ws/ {
    proxy_pass http://backend;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400s;    # WebSocket может жить долго
    proxy_send_timeout 86400s;
}</code></pre>

<p><strong>Ключевой момент:</strong> WebSocket начинается как HTTP-запрос с заголовками Upgrade: websocket и Connection: Upgrade. Nginx должен передать эти заголовки backend-у для переключения протокола.</p>

<h2>Таймауты</h2>
<pre><code>location / {
    proxy_pass http://backend;

    proxy_connect_timeout 5s;   # Тайм-аут на установку соединения с backend
    proxy_send_timeout 10s;     # Тайм-аут на отправку запроса backend-у
    proxy_read_timeout 30s;     # Тайм-аут на получение ответа от backend-а

    # Если backend не ответил — попробовать следующий
    proxy_next_upstream error timeout http_502 http_503;
    proxy_next_upstream_tries 2;
}</code></pre>

<h2>Трейлинг слеш в proxy_pass</h2>
<pre><code># БЕЗ трейлинг слеша — URI передаётся как есть
location /app/ {
    proxy_pass http://backend;
    # /app/users → backend получит /app/users
}

# С трейлинг слешом — location-префикс заменяется
location /app/ {
    proxy_pass http://backend/;
    # /app/users → backend получит /users  (!!!)
}</code></pre>
<p><strong>Это одна из самых частых ошибок в продакшне!</strong> Один символ (/) меняет поведение кардинально.</p>`,

		Quiz: []Q{
			{
				Question:    "Чем отличается proxy_pass http://backend и proxy_pass http://backend/ (с трейлинг слешом)?",
				Options:     []string{"Ничем не отличается", "С трейлинг слешом Nginx заменяет location-префикс в URI перед отправкой на backend, без — передаёт URI как есть", "Трейлинг слеш включает HTTPS", "Трейлинг слеш добавляет host header"},
				Correct:     1,
				Explanation: "Это критическая разница! location /app/ + proxy_pass http://backend/ → запрос /app/users попадёт на backend как /users (location prefix /app/ убирается). Без слеша backend получит полный URI /app/users. Частая причина 404 на проде.",
			},
			{
				Question:    "Зачем нужен proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for?",
				Options:     []string{"Для ускорения запросов", "Чтобы backend знал реальный IP клиента и цепочку прокси через которые прошёл запрос", "Для SSL", "Для кэширования"},
				Correct:     1,
				Explanation: "Без этого заголовка backend видит remote_addr = IP nginx (127.0.0.1). X-Forwarded-For содержит цепочку: client_ip, proxy1_ip, proxy2_ip. $proxy_add_x_forwarded_for автоматически дописывает текущий $remote_addr к существующему заголовку.",
			},
			{
				Question:    "Что нужно добавить в конфиг для проксирования WebSocket соединений?",
				Options:     []string{"Только proxy_pass", "proxy_set_header Upgrade $http_upgrade и Connection \"upgrade\" + proxy_http_version 1.1", "Специальный модуль websocket_proxy", "Включить SSL"},
				Correct:     1,
				Explanation: "WebSocket использует HTTP Upgrade mechanism. Nginx по умолчанию не передаёт заголовки Upgrade и Connection. Без них backend не получит запрос на переключение протокола и WebSocket не установится. Также нужен HTTP/1.1 (не 1.0) для поддержки Upgrade.",
			},
		},
		Tasks: []T{
			{
				Title:      "Генератор конфига reverse proxy",
				Difficulty: "medium",
				Description: `<p>Напишите программу, которая генерирует конфигурацию Nginx reverse proxy из параметров.</p>
<p><strong>Формат входа (построчно):</strong></p>
<ol>
<li>Доменное имя</li>
<li>Порт listen</li>
<li>Количество backend-ов N</li>
<li>Следующие N строк — адреса backend-ов (host:port)</li>
</ol>
<p><strong>Формат вывода:</strong></p>
<pre><code>upstream backend {
    server host1:port1;
    server host2:port2;
}

server {
    listen PORT;
    server_name DOMAIN;

    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}</code></pre>`,
				Hints: `<p>Последовательно считывайте данные через scanner.Scan(). Для числа backend-ов используйте fmt.Sscan. Выводите точный формат — отступы 4 пробела.</p>`,
				Glossary: []GlossaryItem{
					{Term: "upstream", Definition: "Блок Nginx определяющий группу backend-серверов для балансировки нагрузки."},
					{Term: "proxy_pass", Definition: "Директива перенаправляющая запрос на указанный backend или upstream-группу."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	domain := scanner.Text()
	scanner.Scan()
	var port int
	fmt.Sscan(scanner.Text(), &port)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	backends := make([]string, 0, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		backends = append(backends, scanner.Text())
	}

	// TODO: Сгенерируйте и выведите nginx-конфиг
	// Используйте fmt.Printf / fmt.Println
	// Отступы: 4 пробела
	_ = domain
	_ = port
	_ = backends
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	domain := scanner.Text()
	scanner.Scan()
	var port int
	fmt.Sscan(scanner.Text(), &port)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	backends := make([]string, 0, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		backends = append(backends, scanner.Text())
	}

	fmt.Println("upstream backend {")
	for _, b := range backends {
		fmt.Printf("    server %s;\n", b)
	}
	fmt.Println("}")
	fmt.Println()
	fmt.Println("server {")
	fmt.Printf("    listen %d;\n", port)
	fmt.Printf("    server_name %s;\n", domain)
	fmt.Println()
	fmt.Println("    location / {")
	fmt.Println("        proxy_pass http://backend;")
	fmt.Println("        proxy_set_header Host $host;")
	fmt.Println("        proxy_set_header X-Real-IP $remote_addr;")
	fmt.Println("        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;")
	fmt.Println("    }")
	fmt.Println("}")
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "api.example.com\n443\n2\n10.0.0.1:8080\n10.0.0.2:8080",
						ExpectedOutput: "upstream backend {\n    server 10.0.0.1:8080;\n    server 10.0.0.2:8080;\n}\n\nserver {\n    listen 443;\n    server_name api.example.com;\n\n    location / {\n        proxy_pass http://backend;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n    }\n}",
					},
					{
						Input:          "web.local\n80\n1\n127.0.0.1:3000",
						ExpectedOutput: "upstream backend {\n    server 127.0.0.1:3000;\n}\n\nserver {\n    listen 80;\n    server_name web.local;\n\n    location / {\n        proxy_pass http://backend;\n        proxy_set_header Host $host;\n        proxy_set_header X-Real-IP $remote_addr;\n        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;\n    }\n}",
					},
				},
			},
			{
				Title:      "Парсер X-Forwarded-For",
				Difficulty: "easy",
				Description: `<p>Напишите программу, которая парсит заголовок X-Forwarded-For и определяет реальный IP клиента.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество доверенных прокси N</li>
<li>Следующие N строк — IP-адреса доверенных прокси</li>
<li>Последняя строка — значение X-Forwarded-For (через запятую с пробелом)</li>
</ol>
<p><strong>Логика:</strong> Идём по цепочке X-Forwarded-For справа налево. Первый IP, которого нет в списке доверенных — это реальный IP клиента.</p>
<p>Выведите: <code>client: IP</code></p>`,
				Hints: `<p>Сплитите X-Forwarded-For по ", " и обходите массив с конца. Первый IP не в trusted-set — ответ.</p>`,
				Glossary: []GlossaryItem{
					{Term: "X-Forwarded-For", Definition: "HTTP-заголовок содержащий цепочку IP-адресов: клиент, прокси1, прокси2. Каждый прокси дописывает свой remote_addr."},
					{Term: "Trusted proxy", Definition: "Прокси которому мы доверяем. Его IP нужно пропустить при определении реального клиента."},
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	trusted := make(map[string]bool)
	for i := 0; i < n; i++ {
		scanner.Scan()
		trusted[scanner.Text()] = true
	}

	scanner.Scan()
	xff := scanner.Text()
	ips := strings.Split(xff, ", ")

	// TODO: Обойти ips с конца (справа налево)
	// Первый IP, которого нет в trusted — реальный клиент
	// Вывести: client: IP
	_ = ips
	_ = trusted
}`,
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	trusted := make(map[string]bool)
	for i := 0; i < n; i++ {
		scanner.Scan()
		trusted[scanner.Text()] = true
	}

	scanner.Scan()
	xff := scanner.Text()
	ips := strings.Split(xff, ", ")

	for i := len(ips) - 1; i >= 0; i-- {
		if !trusted[ips[i]] {
			fmt.Printf("client: %s\n", ips[i])
			return
		}
	}
	fmt.Println("client: unknown")
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "2\n10.0.0.1\n10.0.0.2\n203.0.113.50, 10.0.0.1, 10.0.0.2",
						ExpectedOutput: "client: 203.0.113.50",
					},
					{
						Input:          "1\n192.168.1.1\n8.8.8.8, 1.1.1.1, 192.168.1.1",
						ExpectedOutput: "client: 1.1.1.1",
					},
					{
						Input:          "0\n5.5.5.5, 6.6.6.6",
						ExpectedOutput: "client: 6.6.6.6",
					},
				},
			},
		},
	}
}

// ── Урок 4: Балансировка нагрузки ──────────────────────────────

func lesson_nginx_load_balancing() L {
	return L{
		Slug: "nginx-load-balancing", Title: "Балансировка нагрузки: алгоритмы и failover", Order: 4,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Балансировка нагрузки: алгоритмы и failover</h1>

<h2>Зачем балансировать нагрузку?</h2>
<p>Один сервер имеет предел — по CPU, памяти, сетевому каналу. Когда нагрузка растёт:</p>
<ul>
<li><strong>Горизонтальное масштабирование</strong> — добавляем серверы</li>
<li><strong>Load balancer</strong> — распределяет запросы между серверами</li>
<li><strong>High Availability</strong> — если один сервер упал, остальные обслуживают</li>
</ul>

<h2>Алгоритмы балансировки</h2>

<h3>1. Round Robin (по умолчанию)</h3>
<pre><code>upstream backend {
    server 10.0.0.1:8080;   # Запрос 1, 4, 7...
    server 10.0.0.2:8080;   # Запрос 2, 5, 8...
    server 10.0.0.3:8080;   # Запрос 3, 6, 9...
}</code></pre>
<p>Распределяет запросы по кругу. Прост, но не учитывает реальную нагрузку серверов.</p>

<h3>2. Weighted Round Robin</h3>
<pre><code>upstream backend {
    server 10.0.0.1:8080 weight=5;   # Получит 5/8 запросов
    server 10.0.0.2:8080 weight=2;   # Получит 2/8 запросов
    server 10.0.0.3:8080 weight=1;   # Получит 1/8 запросов
}</code></pre>
<p>Для серверов разной мощности. Мощный сервер получает больше запросов.</p>

<h3>3. Least Connections</h3>
<pre><code>upstream backend {
    least_conn;
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    server 10.0.0.3:8080;
}</code></pre>
<p>Запрос идёт на сервер с наименьшим количеством активных соединений. Идеален для запросов с разным временем обработки (короткие API + длинные отчёты).</p>

<h3>4. IP Hash</h3>
<pre><code>upstream backend {
    ip_hash;
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
    server 10.0.0.3:8080;
}</code></pre>
<p>Один и тот же клиент (IP) всегда попадает на один и тот же сервер. Нужен для:</p>
<ul>
<li>Session affinity (если сессии хранятся в памяти сервера)</li>
<li>Локальный кэш (данные клиента уже в кэше этого сервера)</li>
</ul>
<p><strong>Проблема:</strong> если сервер упал — его клиенты перераспределяются. Лучше хранить сессии в Redis.</p>

<h2>Health Checks и Failover</h2>
<pre><code>upstream backend {
    server 10.0.0.1:8080 max_fails=3 fail_timeout=30s;
    server 10.0.0.2:8080 max_fails=3 fail_timeout=30s;
    server 10.0.0.3:8080 backup;    # Включается только если основные недоступны
}

# Что считается "fail":
# - connection refused
# - timeout (proxy_connect_timeout)
# - HTTP 502, 503, 504 (при proxy_next_upstream)

server {
    location / {
        proxy_pass http://backend;
        proxy_next_upstream error timeout http_502 http_503;
        proxy_next_upstream_tries 2;     # Макс. 2 попытки
        proxy_next_upstream_timeout 10s; # Общий тайм-аут на все попытки
    }
}</code></pre>

<h2>Sticky Sessions (коммерческий Nginx Plus)</h2>
<pre><code># Nginx Plus only:
upstream backend {
    sticky cookie srv_id expires=1h;
    server 10.0.0.1:8080;
    server 10.0.0.2:8080;
}

# Open-source альтернатива — ip_hash или сторонний модуль nginx-sticky-module</code></pre>

<h2>Мониторинг upstream</h2>
<pre><code># Логирование для отслеживания распределения
log_format upstream_log '$remote_addr - $request '
    'upstream: $upstream_addr '
    'status: $upstream_status '
    'time: $upstream_response_time';

access_log /var/log/nginx/upstream.log upstream_log;</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Когда least_conn лучше чем round-robin?",
				Options:     []string{"Всегда", "Когда запросы имеют сильно разное время обработки — быстрые API и медленные отчёты на одном upstream", "Когда серверов мало", "Когда нужен SSL"},
				Correct:     1,
				Explanation: "Round-robin распределяет равномерно по количеству запросов. Но если один запрос длится 10мс, а другой 10с — сервер с медленным запросом будет перегружен. least_conn учитывает АКТИВНЫЕ соединения — сервер с длинным запросом не получит новых, пока не освободится.",
			},
			{
				Question:    "Что произойдёт с upstream-сервером после max_fails=3 fail_timeout=30s при 3 подряд неудачных ответах?",
				Options:     []string{"Сервер навсегда исключается", "Сервер помечается как unavailable на 30 секунд, затем Nginx снова попробует отправить на него запрос", "Nginx перезапускает сервер", "Генерируется алерт"},
				Correct:     1,
				Explanation: "После 3 неудач сервер помечается 'down' на fail_timeout (30с). По истечении 30с Nginx снова включит его в ротацию и отправит тестовый запрос. Если ответ успешен — сервер полностью восстановлен. Это passive health check (в отличие от active в Nginx Plus).",
			},
			{
				Question:    "В чём основная проблема ip_hash для session persistence?",
				Options:     []string{"Он медленный", "Клиенты за одним NAT/прокси (корпоративная сеть) получат один IP и все пойдут на один сервер, создавая дисбаланс", "Он не работает с HTTPS", "Он несовместим с upstream"},
				Correct:     1,
				Explanation: "ip_hash хеширует IP клиента. Но за корпоративным NAT могут быть тысячи пользователей с одним внешним IP — все попадут на один backend. Также при падении сервера его клиенты перераспределяются, теряя сессии. Лучше хранить сессии в Redis/Memcached.",
			},
			{
				Question:    "Зачем нужен параметр backup в upstream?",
				Options:     []string{"Для логирования", "Backup-сервер получает запросы ТОЛЬКО когда все основные серверы недоступны — это резервный сервер для аварийных ситуаций", "Для хранения бэкапов", "Для SSL-offloading"},
				Correct:     1,
				Explanation: "backup-сервер не получает трафик в нормальном режиме. Он включается только когда ВСЕ основные серверы помечены как down. Используется как последний рубеж — например, сервер который отдаёт страницу 'Technical maintenance' или минимальный readonly-режим.",
			},
		},
		Tasks: []T{
			{
				Title:      "Симулятор Round Robin балансировки",
				Difficulty: "easy",
				Description: `<p>Реализуйте алгоритм Weighted Round Robin балансировки нагрузки.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество серверов N</li>
<li>Следующие N строк — имя_сервера вес (через пробел)</li>
<li>Последняя строка — количество запросов R</li>
</ol>
<p><strong>Алгоритм (Smooth Weighted Round Robin):</strong></p>
<ul>
<li>Для каждого запроса: каждому серверу прибавить его вес к текущему_весу</li>
<li>Выбрать сервер с максимальным текущим_весом</li>
<li>У выбранного сервера отнять сумму всех весов от текущего_веса</li>
</ul>
<p>Для каждого запроса выведите имя выбранного сервера.</p>`,
				Hints: `<p>Smooth Weighted Round Robin от Nginx: currentWeight += effectiveWeight для всех, выбрать max, у выбранного currentWeight -= totalWeight. Это даёт равномерное распределение без скоплений.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Round Robin", Definition: "Алгоритм балансировки по кругу: 1,2,3,1,2,3... Простейший, не учитывает нагрузку."},
					{Term: "Weighted Round Robin", Definition: "Round Robin с весами: сервер с weight=3 получает в 3 раза больше запросов чем с weight=1."},
					{Term: "Smooth Weighted RR", Definition: "Улучшенный WRR от Nginx — запросы распределяются равномерно (не пачками), например для weights 5,1,1 → a,a,b,a,c,a,a вместо a,a,a,a,a,b,c."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type server struct {
	name          string
	weight        int
	currentWeight int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]server, n)
	totalWeight := 0
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		var w int
		fmt.Sscan(parts[1], &w)
		servers[i] = server{name: parts[0], weight: w}
		totalWeight += w
	}

	scanner.Scan()
	var requests int
	fmt.Sscan(scanner.Text(), &requests)

	// TODO: Для каждого запроса реализуйте Smooth Weighted Round Robin:
	// 1. Каждому серверу: currentWeight += weight
	// 2. Выбрать сервер с максимальным currentWeight
	// 3. У выбранного: currentWeight -= totalWeight
	// 4. Вывести имя выбранного сервера
	_ = servers
	_ = totalWeight
	_ = requests
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type server struct {
	name          string
	weight        int
	currentWeight int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]server, n)
	totalWeight := 0
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		var w int
		fmt.Sscan(parts[1], &w)
		servers[i] = server{name: parts[0], weight: w}
		totalWeight += w
	}

	scanner.Scan()
	var requests int
	fmt.Sscan(scanner.Text(), &requests)

	for r := 0; r < requests; r++ {
		bestIdx := 0
		for i := range servers {
			servers[i].currentWeight += servers[i].weight
			if servers[i].currentWeight > servers[bestIdx].currentWeight {
				bestIdx = i
			}
		}
		fmt.Println(servers[bestIdx].name)
		servers[bestIdx].currentWeight -= totalWeight
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3\nbackend1 5\nbackend2 1\nbackend3 1\n7",
						ExpectedOutput: "backend1\nbackend1\nbackend2\nbackend1\nbackend3\nbackend1\nbackend1",
					},
					{
						Input:          "2\nserver-a 1\nserver-b 1\n4",
						ExpectedOutput: "server-a\nserver-b\nserver-a\nserver-b",
					},
				},
			},
			{
				Title:      "Least Connections балансировщик",
				Difficulty: "medium",
				Description: `<p>Реализуйте алгоритм Least Connections.</p>
<p><strong>Формат входа (построчно):</strong></p>
<ol>
<li>Первая строка — количество серверов N</li>
<li>Следующие N строк — имена серверов</li>
<li>Остальные строки — события: <code>REQUEST</code> (новый запрос) или <code>DONE имя_сервера</code> (запрос завершён)</li>
</ol>
<p><strong>Логика:</strong></p>
<ul>
<li>REQUEST — выбрать сервер с минимальным числом активных соединений, вывести его имя, увеличить счётчик</li>
<li>DONE server — уменьшить счётчик активных соединений на указанном сервере</li>
<li>При равном количестве соединений — выбрать первый по порядку</li>
</ul>`,
				Hints: `<p>Используйте map[string]int для подсчёта активных соединений на каждом сервере. При REQUEST ищите минимум.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Least Connections", Definition: "Алгоритм балансировки: новый запрос идёт на сервер с наименьшим числом текущих активных соединений."},
					{Term: "Active connections", Definition: "Количество соединений которые сейчас обрабатываются сервером (запрос отправлен, ответ ещё не получен)."},
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]string, n)
	conns := make(map[string]int)
	for i := 0; i < n; i++ {
		scanner.Scan()
		servers[i] = scanner.Text()
		conns[servers[i]] = 0
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "REQUEST" {
			// TODO: Найти сервер с минимальным conns[server]
			// При равных — первый по порядку в servers[]
			// Вывести имя, увеличить conns
		} else if strings.HasPrefix(line, "DONE ") {
			// TODO: Уменьшить conns для указанного сервера
		}
	}

	_ = conns
	_ = servers
}`,
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]string, n)
	conns := make(map[string]int)
	for i := 0; i < n; i++ {
		scanner.Scan()
		servers[i] = scanner.Text()
		conns[servers[i]] = 0
	}

	for scanner.Scan() {
		line := scanner.Text()

		if line == "REQUEST" {
			best := servers[0]
			for _, s := range servers {
				if conns[s] < conns[best] {
					best = s
				}
			}
			fmt.Println(best)
			conns[best]++
		} else if strings.HasPrefix(line, "DONE ") {
			srv := strings.TrimPrefix(line, "DONE ")
			conns[srv]--
		}
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3\nweb1\nweb2\nweb3\nREQUEST\nREQUEST\nREQUEST\nREQUEST\nDONE web1\nREQUEST",
						ExpectedOutput: "web1\nweb2\nweb3\nweb1\nweb1",
					},
					{
						Input:          "2\nalpha\nbeta\nREQUEST\nREQUEST\nDONE alpha\nREQUEST\nDONE beta\nREQUEST",
						ExpectedOutput: "alpha\nbeta\nalpha\nalpha",
					},
				},
			},
			{
				Title:      "IP Hash балансировка",
				Difficulty: "medium",
				Description: `<p>Реализуйте алгоритм IP Hash для sticky sessions.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество серверов N</li>
<li>Следующие N строк — имена серверов</li>
<li>Остальные строки — IP-адреса запросов</li>
</ol>
<p><strong>Алгоритм хеширования:</strong> Сложите все байты IP-адреса (числа между точками) и возьмите остаток от деления на N.</p>
<p>Формула: <code>hash = (octet1 + octet2 + octet3 + octet4) % N</code></p>
<p>Для каждого запроса выведите: <code>IP -> имя_сервера</code></p>`,
				Hints: `<p>Распарсите IP через strings.Split(ip, ".") и сложите октеты через fmt.Sscan или strconv.Atoi.</p>`,
				Glossary: []GlossaryItem{
					{Term: "ip_hash", Definition: "Алгоритм балансировки по хешу IP клиента. Один клиент всегда попадает на один сервер (session affinity)."},
					{Term: "Session affinity", Definition: "Гарантия что запросы одного клиента всегда попадают на один backend. Нужно для серверных сессий."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]string, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		servers[i] = scanner.Text()
	}

	for scanner.Scan() {
		ip := scanner.Text()
		// TODO: Разбить IP по ".", сложить октеты, взять % n
		// Вывести: IP -> servers[hash]
		_ = ip
		_ = strconv.Atoi
		_ = strings.Split
	}
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	servers := make([]string, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		servers[i] = scanner.Text()
	}

	for scanner.Scan() {
		ip := scanner.Text()
		parts := strings.Split(ip, ".")
		sum := 0
		for _, p := range parts {
			v, _ := strconv.Atoi(p)
			sum += v
		}
		idx := sum % n
		fmt.Printf("%s -> %s\n", ip, servers[idx])
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3\nweb1\nweb2\nweb3\n192.168.1.1\n192.168.1.1\n10.0.0.1\n172.16.0.5",
						ExpectedOutput: "192.168.1.1 -> web2\n192.168.1.1 -> web2\n10.0.0.1 -> web2\n172.16.0.5 -> web3",
					},
					{
						Input:          "2\nalpha\nbeta\n1.2.3.4\n5.6.7.8",
						ExpectedOutput: "1.2.3.4 -> alpha\n5.6.7.8 -> beta",
					},
				},
			},
		},
	}
}

// ── Урок 5: SSL/TLS ───────────────────────────────────────────

func lesson_nginx_ssl_tls() L {
	return L{
		Slug: "nginx-ssl-tls", Title: "SSL/TLS: сертификаты и безопасность", Order: 5,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>SSL/TLS: сертификаты и безопасность</h1>

<h2>Зачем нужен HTTPS?</h2>
<p>HTTP передаёт данные <strong>открытым текстом</strong>. Любой на пути (Wi-Fi, ISP, CDN) может:</p>
<ul>
<li><strong>Прочитать</strong> — пароли, cookies, данные форм</li>
<li><strong>Подменить</strong> — внедрить рекламу, вредоносный JS</li>
<li><strong>Перехватить</strong> — session hijacking</li>
</ul>
<p>HTTPS = HTTP + TLS (Transport Layer Security). TLS обеспечивает:</p>
<ul>
<li><strong>Шифрование</strong> — данные зашифрованы, перехватчик видит мусор</li>
<li><strong>Аутентификация</strong> — сертификат подтверждает что сервер = тот за кого себя выдаёт</li>
<li><strong>Целостность</strong> — данные не могут быть модифицированы в пути</li>
</ul>

<h2>Цепочка сертификатов</h2>
<pre><code>Root CA (в браузере/ОС)
    └── Intermediate CA (подписан Root)
            └── Server Certificate (подписан Intermediate)
                    example.com</code></pre>
<p>Браузер доверяет сайту если может построить цепочку от сертификата сервера до Root CA, которому он доверяет по умолчанию.</p>

<h2>Базовая конфигурация SSL в Nginx</h2>
<pre><code>server {
    listen 443 ssl http2;
    server_name example.com;

    # Сертификат и ключ
    ssl_certificate     /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    # Протоколы — только TLS 1.2 и 1.3
    ssl_protocols TLSv1.2 TLSv1.3;

    # Шифронаборы — только сильные
    ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
    ssl_prefer_server_ciphers off;  # В TLS 1.3 клиент выбирает

    # Session cache для переиспользования TLS-сессий
    ssl_session_cache shared:SSL:10m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;   # Отключаем для Forward Secrecy

    # OCSP Stapling — ускоряет проверку сертификата клиентом
    ssl_stapling on;
    ssl_stapling_verify on;
    ssl_trusted_certificate /etc/letsencrypt/live/example.com/chain.pem;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
}

# Редирект HTTP → HTTPS
server {
    listen 80;
    server_name example.com;
    return 301 https://$host$request_uri;
}</code></pre>

<h2>Let's Encrypt — бесплатные сертификаты</h2>
<pre><code># Установка certbot
apt install certbot python3-certbot-nginx

# Получение сертификата (автоматически настраивает nginx)
certbot --nginx -d example.com -d www.example.com

# Только получить (без изменения конфига nginx)
certbot certonly --webroot -w /var/www/html -d example.com

# Автообновление (cron / systemd timer)
certbot renew --post-hook "nginx -s reload"

# Файлы сертификата:
# /etc/letsencrypt/live/example.com/fullchain.pem  — сертификат + intermediate
# /etc/letsencrypt/live/example.com/privkey.pem    — приватный ключ
# /etc/letsencrypt/live/example.com/chain.pem      — только intermediate</code></pre>

<h2>Security Headers</h2>
<pre><code>server {
    # HSTS — браузер запоминает что сайт только HTTPS
    add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;

    # Защита от clickjacking
    add_header X-Frame-Options "SAMEORIGIN" always;

    # Защита от MIME-sniffing
    add_header X-Content-Type-Options "nosniff" always;

    # XSS protection (устарел, но не вредит)
    add_header X-XSS-Protection "1; mode=block" always;

    # Referrer policy
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;

    # Content Security Policy (пример)
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'" always;
}</code></pre>

<h2>HSTS — важные нюансы</h2>
<p><strong>Strict-Transport-Security</strong> говорит браузеру: "следующие N секунд обращайся к этому домену ТОЛЬКО по HTTPS". Это защита от downgrade-атак (sslstrip).</p>
<p><strong>Опасность:</strong> если выставили max-age=2 года, а потом решили убрать HTTPS — пользователи не смогут зайти на HTTP-версию 2 года! Начинайте с малого max-age при тестировании.</p>

<h2>HTTP/2</h2>
<pre><code># Включается одним словом:
listen 443 ssl http2;

# Преимущества HTTP/2:
# - Мультиплексирование (много запросов в одном TCP-соединении)
# - Header compression (HPACK)
# - Server push
# - Binary protocol (быстрее парсить чем текстовый HTTP/1.1)</code></pre>

<h2>Проверка SSL-конфигурации</h2>
<pre><code># Проверить сертификат
openssl s_client -connect example.com:443 -servername example.com

# Проверить цепочку
openssl s_client -connect example.com:443 -showcerts

# Проверить срок действия
echo | openssl s_client -connect example.com:443 2>/dev/null | openssl x509 -noout -dates

# Онлайн-тесты:
# https://www.ssllabs.com/ssltest/
# https://observatory.mozilla.org/</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему ssl_session_tickets рекомендуется отключать?",
				Options:     []string{"Они замедляют соединение", "Session tickets шифруются одним ключом на сервере — если ключ скомпрометирован, можно расшифровать все прошлые сессии, нарушая Forward Secrecy", "Они не поддерживаются современными браузерами", "Они увеличивают размер сертификата"},
				Correct:     1,
				Explanation: "Forward Secrecy означает: компрометация ключа сервера не позволяет расшифровать ПРОШЛЫЕ сессии. Session tickets используют симметричный ключ для шифрования TLS-состояния. Если этот ключ утёк — все записанные сессии расшифруемы. Без tickets каждая сессия использует уникальный эфемерный ключ.",
			},
			{
				Question:    "Что такое OCSP Stapling и зачем оно нужно?",
				Options:     []string{"Это способ шифрования трафика", "Сервер сам получает подтверждение валидности сертификата от CA и отдаёт его клиенту — клиенту не нужно делать отдельный запрос к CA, это ускоряет TLS handshake", "Это альтернатива Let's Encrypt", "Это механизм кэширования SSL-сессий"},
				Correct:     1,
				Explanation: "Без OCSP Stapling клиент при TLS handshake делает запрос к OCSP-серверу CA: 'этот сертификат ещё валиден?'. Это добавляет задержку и зависимость от доступности CA. С stapling — Nginx сам периодически получает OCSP-ответ и 'прикрепляет' (staples) его к TLS handshake. Быстрее и надёжнее.",
			},
			{
				Question:    "Чем опасно выставление HSTS с max-age=63072000 без предварительного тестирования?",
				Options:     []string{"Замедляет загрузку сайта", "Если потребуется отключить HTTPS — пользователи с закешированным HSTS не смогут зайти на HTTP-версию сайта 2 года, даже удаление заголовка не поможет", "Блокирует Google-индексацию", "Увеличивает нагрузку на CPU"},
				Correct:     1,
				Explanation: "HSTS заголовок кешируется в браузере на max-age секунд. 63072000 = 2 года. Если выставили, а потом нужен HTTP (проблемы с сертификатом, миграция) — браузеры будут принудительно перенаправлять на HTTPS. Единственный способ — ждать или пользователь руками очищает HSTS-кеш. Начинайте с max-age=300 при тестировании.",
			},
		},
		Tasks: []T{
			{
				Title:      "Валидатор цепочки сертификатов",
				Difficulty: "medium",
				Description: `<p>Напишите программу, которая проверяет валидность цепочки сертификатов.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество сертификатов N</li>
<li>Следующие N строк — сертификаты в формате: <code>subject issuer days_valid</code></li>
<li>Последняя строка — имя Root CA (которому мы доверяем)</li>
</ol>
<p><strong>Правила валидации:</strong></p>
<ul>
<li>Цепочка строится от server cert вверх: subject[i].issuer == subject[i+1].subject</li>
<li>Верхний сертификат в цепочке должен быть подписан Root CA (issuer == root_ca)</li>
<li>Все сертификаты должны иметь days_valid > 0</li>
</ul>
<p><strong>Вывод:</strong></p>
<ul>
<li><code>VALID</code> — если цепочка полная и все сертификаты не истекли</li>
<li><code>EXPIRED: subject</code> — если сертификат истёк (первый найденный)</li>
<li><code>BROKEN CHAIN: subject</code> — если не удаётся построить цепочку</li>
</ul>
<p>Сертификаты подаются в порядке: server, intermediate(s), ... (от leaf к root).</p>`,
				Hints: `<p>Пройдите по цепочке с начала: проверьте days_valid > 0 (если нет — EXPIRED). Проверьте что cert[i].issuer == cert[i+1].subject (если нет — BROKEN CHAIN). Последний cert.issuer должен == root_ca.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Certificate Chain", Definition: "Цепочка доверия: server cert → intermediate CA → root CA. Браузер проверяет каждое звено."},
					{Term: "Subject", Definition: "Кому выдан сертификат (доменное имя или название CA)."},
					{Term: "Issuer", Definition: "Кто выдал (подписал) сертификат. Issuer одного = Subject вышестоящего."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type cert struct {
	subject   string
	issuer    string
	daysValid int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	certs := make([]cert, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		days, _ := strconv.Atoi(parts[2])
		certs[i] = cert{subject: parts[0], issuer: parts[1], daysValid: days}
	}

	scanner.Scan()
	rootCA := scanner.Text()

	// TODO: Validate the certificate chain
	// 1. Check each cert's daysValid > 0 → otherwise "EXPIRED: subject"
	// 2. Check chain: certs[i].issuer == certs[i+1].subject → otherwise "BROKEN CHAIN: subject"
	// 3. Check last cert's issuer == rootCA → otherwise "BROKEN CHAIN: subject"
	// 4. If all good → "VALID"
	_ = certs
	_ = rootCA
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type cert struct {
	subject   string
	issuer    string
	daysValid int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	certs := make([]cert, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		days, _ := strconv.Atoi(parts[2])
		certs[i] = cert{subject: parts[0], issuer: parts[1], daysValid: days}
	}

	scanner.Scan()
	rootCA := scanner.Text()

	for _, c := range certs {
		if c.daysValid <= 0 {
			fmt.Printf("EXPIRED: %s\n", c.subject)
			return
		}
	}

	for i := 0; i < len(certs)-1; i++ {
		if certs[i].issuer != certs[i+1].subject {
			fmt.Printf("BROKEN CHAIN: %s\n", certs[i].subject)
			return
		}
	}

	if certs[len(certs)-1].issuer != rootCA {
		fmt.Printf("BROKEN CHAIN: %s\n", certs[len(certs)-1].subject)
		return
	}

	fmt.Println("VALID")
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3\nexample.com LetsEncrypt-R3 90\nLetsEncrypt-R3 ISRG-Root-X1 365\nISRG-Root-X1 DST-Root-CA-X3 730\nDST-Root-CA-X3",
						ExpectedOutput: "VALID",
					},
					{
						Input:          "2\nmysite.org IntermediateCA 0\nIntermediateCA RootCA 365\nRootCA",
						ExpectedOutput: "EXPIRED: mysite.org",
					},
					{
						Input:          "2\napp.io SomeCA 90\nOtherCA RootCA 365\nRootCA",
						ExpectedOutput: "BROKEN CHAIN: app.io",
					},
				},
			},
			{
				Title:      "Генератор Security Headers",
				Difficulty: "easy",
				Description: `<p>Напишите программу, которая генерирует блок security headers для Nginx на основе параметров.</p>
<p><strong>Формат входа (построчно):</strong></p>
<ol>
<li><code>hsts_max_age</code> — значение max-age для HSTS (число секунд, 0 = не добавлять HSTS)</li>
<li><code>frame_options</code> — значение X-Frame-Options (DENY, SAMEORIGIN, или пустая строка = не добавлять)</li>
<li><code>content_type_nosniff</code> — yes/no</li>
</ol>
<p><strong>Формат вывода:</strong> директивы add_header (только для включённых опций), по одной на строку.</p>
<pre><code>add_header Strict-Transport-Security "max-age=N; includeSubDomains" always;
add_header X-Frame-Options "VALUE" always;
add_header X-Content-Type-Options "nosniff" always;</code></pre>`,
				Hints: `<p>Просто проверяйте условия: если hsts > 0 — вывести HSTS header, если frame_options не пустой — вывести X-Frame-Options, и т.д.</p>`,
				Glossary: []GlossaryItem{
					{Term: "HSTS", Definition: "HTTP Strict Transport Security — заголовок принуждающий браузер использовать только HTTPS для данного домена в течение max-age секунд."},
					{Term: "X-Frame-Options", Definition: "Защита от clickjacking. DENY = запрет встраивания во фреймы. SAMEORIGIN = только с того же домена."},
					{Term: "X-Content-Type-Options", Definition: "Значение 'nosniff' запрещает браузеру угадывать MIME-тип — использовать только тот, что в Content-Type."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	var hstsMaxAge int
	fmt.Sscan(scanner.Text(), &hstsMaxAge)

	scanner.Scan()
	frameOptions := scanner.Text()

	scanner.Scan()
	contentTypeNosniff := scanner.Text()

	// TODO: Генерируйте add_header директивы
	// Если hstsMaxAge > 0 → add_header Strict-Transport-Security "max-age=N; includeSubDomains" always;
	// Если frameOptions не пустой → add_header X-Frame-Options "VALUE" always;
	// Если contentTypeNosniff == "yes" → add_header X-Content-Type-Options "nosniff" always;
	_ = hstsMaxAge
	_ = frameOptions
	_ = contentTypeNosniff
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	var hstsMaxAge int
	fmt.Sscan(scanner.Text(), &hstsMaxAge)

	scanner.Scan()
	frameOptions := scanner.Text()

	scanner.Scan()
	contentTypeNosniff := scanner.Text()

	if hstsMaxAge > 0 {
		fmt.Printf("add_header Strict-Transport-Security \"max-age=%d; includeSubDomains\" always;\n", hstsMaxAge)
	}
	if frameOptions != "" {
		fmt.Printf("add_header X-Frame-Options \"%s\" always;\n", frameOptions)
	}
	if contentTypeNosniff == "yes" {
		fmt.Println("add_header X-Content-Type-Options \"nosniff\" always;")
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "63072000\nSAMEORIGIN\nyes",
						ExpectedOutput: "add_header Strict-Transport-Security \"max-age=63072000; includeSubDomains\" always;\nadd_header X-Frame-Options \"SAMEORIGIN\" always;\nadd_header X-Content-Type-Options \"nosniff\" always;",
					},
					{
						Input:          "0\nDENY\nno",
						ExpectedOutput: "add_header X-Frame-Options \"DENY\" always;",
					},
					{
						Input:          "31536000\n\nyes",
						ExpectedOutput: "add_header Strict-Transport-Security \"max-age=31536000; includeSubDomains\" always;\nadd_header X-Content-Type-Options \"nosniff\" always;",
					},
				},
			},
		},
	}
}

// ── Урок 6: Производительность и кэширование ──────────────────

func lesson_nginx_performance() L {
	return L{
		Slug: "nginx-performance", Title: "Производительность: gzip, кэширование, буферы", Order: 6,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Производительность: gzip, кэширование, буферы</h1>

<h2>Gzip-сжатие</h2>
<p>Gzip уменьшает размер ответов на 60-90% для текстовых форматов (HTML, CSS, JS, JSON). Это экономит трафик и ускоряет загрузку.</p>

<pre><code>http {
    gzip on;
    gzip_vary on;                    # Добавляет Vary: Accept-Encoding
    gzip_proxied any;                # Сжимать даже проксированные ответы
    gzip_comp_level 6;               # Уровень сжатия (1-9, 6 = оптимум CPU/сжатие)
    gzip_min_length 256;             # Не сжимать файлы меньше 256 байт
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/javascript
        application/json
        application/xml
        application/rss+xml
        image/svg+xml;
}</code></pre>

<p><strong>Важно:</strong> НЕ сжимайте уже сжатые форматы (JPEG, PNG, MP4, ZIP) — они не уменьшатся, но CPU будет потрачен.</p>

<h2>Proxy Cache</h2>
<p>Nginx может кэшировать ответы backend-а — повторные запросы отдаются из кэша без обращения к backend:</p>
<pre><code>http {
    # Определяем зону кэша
    proxy_cache_path /var/cache/nginx
        levels=1:2                    # Структура директорий
        keys_zone=my_cache:10m        # 10 МБ для ключей в памяти
        max_size=1g                   # Макс. размер на диске
        inactive=60m                  # Удалять неиспользуемые через 60 мин
        use_temp_path=off;            # Писать сразу в кэш (без temp)

    server {
        location /api/ {
            proxy_pass http://backend;
            proxy_cache my_cache;
            proxy_cache_valid 200 10m;      # Кэшировать 200 на 10 минут
            proxy_cache_valid 404 1m;       # Кэшировать 404 на 1 минуту
            proxy_cache_key "$scheme$request_method$host$request_uri";
            proxy_cache_use_stale error timeout updating;  # Отдавать stale при ошибке

            # Заголовок для дебага — показывает HIT/MISS/EXPIRED
            add_header X-Cache-Status $upstream_cache_status;
        }
    }
}</code></pre>

<h2>Expires и Cache-Control для статики</h2>
<pre><code>location ~* \.(jpg|jpeg|png|gif|ico|css|js|woff2)$ {
    root /var/www/static;
    expires 30d;                           # Expires: +30 дней
    add_header Cache-Control "public, immutable";
    access_log off;                        # Не логировать запросы к статике
}</code></pre>

<h2>Буферы</h2>
<pre><code>http {
    # Буферы для проксирования
    proxy_buffering on;                   # Включить буферизацию (по умолчанию)
    proxy_buffer_size 4k;                 # Буфер для первой части ответа (заголовки)
    proxy_buffers 8 8k;                   # 8 буферов по 8 КБ для тела ответа
    proxy_busy_buffers_size 16k;          # Сколько можно отдавать клиенту пока буферизуем

    # Клиентские буферы
    client_body_buffer_size 16k;          # Буфер для тела запроса
    client_max_body_size 10m;             # Максимальный размер тела запроса
    client_header_buffer_size 1k;         # Буфер для заголовков запроса
    large_client_header_buffers 4 8k;     # Для больших заголовков/cookies
}</code></pre>

<h2>Open File Cache</h2>
<pre><code>http {
    open_file_cache max=10000 inactive=20s;   # Кэш метаданных файлов
    open_file_cache_valid 30s;                 # Перепроверять каждые 30с
    open_file_cache_min_uses 2;                # Кэшировать если запрошен >= 2 раз
    open_file_cache_errors on;                 # Кэшировать и ошибки (файл не найден)
}</code></pre>
<p>Open file cache хранит <strong>метаданные</strong> файлов (дескрипторы, размеры, время модификации) — не содержимое. Это экономит системные вызовы stat() и open().</p>

<h2>Worker Connections и события</h2>
<pre><code>worker_processes auto;

events {
    worker_connections 4096;         # Макс. соединений на 1 worker
    multi_accept on;                 # Принимать все ожидающие соединения за раз
    use epoll;                       # Явно указать epoll (Linux)
}

# Расчёт max clients:
# max_clients = worker_processes * worker_connections
# С reverse proxy каждый клиент = 2 соединения (client→nginx + nginx→backend)
# Реальный max = worker_processes * worker_connections / 2</code></pre>

<h2>Sendfile и TCP оптимизации</h2>
<pre><code>http {
    sendfile on;           # Передача файлов напрямую из FS в сокет (zero-copy)
    tcp_nopush on;         # Отправлять заголовки + начало файла одним пакетом
    tcp_nodelay on;        # Отключить алгоритм Nagle (для keepalive)
    keepalive_timeout 65;  # Тайм-аут keepalive-соединений
    keepalive_requests 100; # Макс. запросов на одно keepalive-соединение
}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему gzip_comp_level обычно ставят 6, а не 9?",
				Options:     []string{"9 не поддерживается браузерами", "После уровня 5-6 выигрыш в сжатии минимален (1-2%), но CPU-затраты растут экспоненциально — закон убывающей отдачи", "Уровень 9 создаёт несовместимый формат", "Nginx ограничивает максимум до 6"},
				Correct:     1,
				Explanation: "gzip_comp_level 6 даёт ~85% от максимального сжатия при ~50% CPU-затрат уровня 9. Уровень 9 сожмёт текст на 1-2% лучше, но потратит в 2-3 раза больше CPU. При высокой нагрузке это критично. В продакшне CPU — бутылочное горлышко, а не трафик.",
			},
			{
				Question:    "Что означает proxy_cache_use_stale error timeout updating?",
				Options:     []string{"Удалять старые записи кэша", "При ошибке backend-а, таймауте или обновлении кэша — отдавать клиенту устаревшую (stale) версию ответа из кэша вместо ошибки", "Кэшировать ответы с ошибками", "Обновлять кэш каждую секунду"},
				Correct:     1,
				Explanation: "Это критическая директива для надёжности. Если backend упал (error), не ответил вовремя (timeout), или кэш в процессе обновления (updating) — клиент получит предыдущую закэшированную версию. Пользователь увидит немного устаревшие данные, а не 502 Bad Gateway.",
			},
			{
				Question:    "Зачем sendfile on и почему это быстрее обычной отправки?",
				Options:     []string{"sendfile включает шифрование", "sendfile передаёт файл напрямую из файловой системы в сокет внутри ядра (zero-copy) минуя userspace — два системных вызова (read+write) заменяются одним", "sendfile включает HTTP/2", "sendfile увеличивает буферы"},
				Correct:     1,
				Explanation: "Без sendfile: ядро читает файл в userspace-буфер → Nginx копирует в сокетный буфер → ядро отправляет. С sendfile: ядро копирует данные из файлового кэша напрямую в сокетный буфер. Нет переключения user/kernel space, нет копирования данных. Особенно эффективно для больших статических файлов.",
			},
		},
		Tasks: []T{
			{
				Title:      "Калькулятор экономии gzip",
				Difficulty: "easy",
				Description: `<p>Напишите программу, которая рассчитывает экономию от gzip-сжатия.</p>
<p><strong>Формат входа (построчно):</strong></p>
<ol>
<li>Первая строка — количество файлов N</li>
<li>Следующие N строк: <code>тип_файла размер_байт</code></li>
</ol>
<p><strong>Коэффициенты сжатия по типу:</strong></p>
<ul>
<li>html, css, js, json, xml, svg — 0.30 (70% экономия)</li>
<li>jpg, png, gif, mp4, zip — 1.00 (не сжимается)</li>
<li>все остальные — 0.50 (50% экономия)</li>
</ul>
<p><strong>Формат вывода:</strong></p>
<pre><code>original: X bytes
compressed: Y bytes
saved: Z bytes (P%)</code></pre>
<p>Процент округлять до целого. Compressed размер = floor(original * ratio).</p>`,
				Hints: `<p>Определите тип по расширению, умножьте на коэффициент. Суммируйте original и compressed для всех файлов.</p>`,
				Glossary: []GlossaryItem{
					{Term: "gzip", Definition: "Алгоритм сжатия (LZ77 + Huffman). В Nginx сжимает HTTP-ответы для уменьшения трафика. Клиент декодирует прозрачно."},
					{Term: "Content-Encoding: gzip", Definition: "HTTP-заголовок указывающий что тело ответа сжато gzip. Браузер автоматически разжимает."},
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	textTypes := map[string]bool{
		"html": true, "css": true, "js": true,
		"json": true, "xml": true, "svg": true,
	}
	binaryTypes := map[string]bool{
		"jpg": true, "png": true, "gif": true,
		"mp4": true, "zip": true,
	}

	var totalOriginal, totalCompressed int

	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		var size int
		fmt.Sscan(parts[1], &size)
		fileType := parts[0]

		// TODO: определить ratio по типу файла
		// textTypes → 0.30, binaryTypes → 1.00, остальные → 0.50
		// compressed = int(float64(size) * ratio)
		// Суммировать totalOriginal и totalCompressed
		_ = fileType
		_ = textTypes
		_ = binaryTypes
	}

	// TODO: Вывести результат
	_ = totalOriginal
	_ = totalCompressed
}`,
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
	var n int
	fmt.Sscan(scanner.Text(), &n)

	textTypes := map[string]bool{
		"html": true, "css": true, "js": true,
		"json": true, "xml": true, "svg": true,
	}
	binaryTypes := map[string]bool{
		"jpg": true, "png": true, "gif": true,
		"mp4": true, "zip": true,
	}

	var totalOriginal, totalCompressed int

	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		var size int
		fmt.Sscan(parts[1], &size)
		fileType := parts[0]

		var ratio float64
		if textTypes[fileType] {
			ratio = 0.30
		} else if binaryTypes[fileType] {
			ratio = 1.00
		} else {
			ratio = 0.50
		}

		totalOriginal += size
		totalCompressed += int(float64(size) * ratio)
	}

	saved := totalOriginal - totalCompressed
	percent := 0
	if totalOriginal > 0 {
		percent = saved * 100 / totalOriginal
	}
	fmt.Printf("original: %d bytes\n", totalOriginal)
	fmt.Printf("compressed: %d bytes\n", totalCompressed)
	fmt.Printf("saved: %d bytes (%d%%)\n", saved, percent)
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "4\nhtml 10000\ncss 5000\njs 8000\njpg 50000",
						ExpectedOutput: "original: 73000 bytes\ncompressed: 56900 bytes\nsaved: 16100 bytes (22%)",
					},
					{
						Input:          "2\njson 1000\nxml 2000",
						ExpectedOutput: "original: 3000 bytes\ncompressed: 900 bytes\nsaved: 2100 bytes (70%)",
					},
				},
			},
			{
				Title:      "Симулятор proxy_cache",
				Difficulty: "medium",
				Description: `<p>Реализуйте упрощённый proxy_cache с TTL.</p>
<p><strong>Формат входа (построчно):</strong></p>
<ul>
<li><code>GET url timestamp</code> — запрос к URL в момент времени timestamp (целое число)</li>
<li><code>SET url timestamp ttl response</code> — backend вернул ответ, закэшировать на ttl секунд</li>
</ul>
<p><strong>Логика GET:</strong></p>
<ul>
<li>Если URL есть в кэше и (timestamp - cache_timestamp) < ttl → вывести: <code>HIT: response</code></li>
<li>Иначе → вывести: <code>MISS</code></li>
</ul>
<p><strong>Логика SET:</strong> сохранить url, timestamp, ttl, response в кэш (ничего не выводить).</p>`,
				Hints: `<p>Используйте map[string]struct{timestamp, ttl int; response string}. При GET проверяйте условие: текущий timestamp - сохранённый timestamp < ttl.</p>`,
				Glossary: []GlossaryItem{
					{Term: "proxy_cache", Definition: "Кэш ответов upstream-сервера. Повторные запросы отдаются из кэша без обращения к backend."},
					{Term: "TTL (Time To Live)", Definition: "Время жизни записи в кэше. По истечении TTL запись считается устаревшей (stale)."},
					{Term: "HIT/MISS", Definition: "HIT — ответ найден в кэше и ещё валиден. MISS — кэша нет или он устарел, нужен запрос к backend."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cacheEntry struct {
	timestamp int
	ttl       int
	response  string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cache := make(map[string]cacheEntry)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 5)
		action := parts[0]

		if action == "GET" {
			url := parts[1]
			var ts int
			fmt.Sscan(parts[2], &ts)

			// TODO: Проверить кэш
			// Если есть запись и (ts - entry.timestamp) < entry.ttl → HIT: response
			// Иначе → MISS
			_ = url
			_ = ts
		} else if action == "SET" {
			url := parts[1]
			var ts, ttl int
			fmt.Sscan(parts[2], &ts)
			fmt.Sscan(parts[3], &ttl)
			response := parts[4]

			// TODO: Сохранить в кэш
			_ = url
			_ = ts
			_ = ttl
			_ = response
		}
	}
	_ = cache
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cacheEntry struct {
	timestamp int
	ttl       int
	response  string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	cache := make(map[string]cacheEntry)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 5)
		action := parts[0]

		if action == "GET" {
			url := parts[1]
			var ts int
			fmt.Sscan(parts[2], &ts)

			if entry, ok := cache[url]; ok && (ts-entry.timestamp) < entry.ttl {
				fmt.Printf("HIT: %s\n", entry.response)
			} else {
				fmt.Println("MISS")
			}
		} else if action == "SET" {
			url := parts[1]
			var ts, ttl int
			fmt.Sscan(parts[2], &ts)
			fmt.Sscan(parts[3], &ttl)
			response := parts[4]
			cache[url] = cacheEntry{timestamp: ts, ttl: ttl, response: response}
		}
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "GET /api/users 100\nSET /api/users 100 60 [{\"id\":1}]\nGET /api/users 130\nGET /api/users 170\nGET /api/posts 130",
						ExpectedOutput: "MISS\nHIT: [{\"id\":1}]\nMISS\nMISS",
					},
					{
						Input:          "SET /page 0 10 hello\nGET /page 5\nGET /page 9\nGET /page 10",
						ExpectedOutput: "HIT: hello\nHIT: hello\nMISS",
					},
				},
			},
			{
				Title:      "Анализатор cache-status логов",
				Difficulty: "easy",
				Description: `<p>Напишите программу, которая анализирует логи Nginx с X-Cache-Status и считает статистику.</p>
<p><strong>Формат входа (построчно):</strong></p>
<pre><code>url status</code></pre>
<p>Где status: HIT, MISS, EXPIRED, STALE, BYPASS</p>
<p><strong>Формат вывода:</strong></p>
<pre><code>total: N
HIT: N (P%)
MISS: N (P%)
other: N (P%)</code></pre>
<p>other = EXPIRED + STALE + BYPASS. Процент = count * 100 / total (целочисленное деление).</p>`,
				Hints: `<p>Просто считайте HIT и MISS отдельно, всё остальное — other. Потом выведите статистику.</p>`,
				Glossary: []GlossaryItem{
					{Term: "$upstream_cache_status", Definition: "Переменная Nginx показывающая результат обращения к кэшу: HIT, MISS, EXPIRED, STALE, UPDATING, BYPASS."},
					{Term: "Cache Hit Ratio", Definition: "Процент запросов обслуженных из кэша. CHR = HIT / total * 100%. Хороший показатель > 80%."},
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
	total := 0
	hits := 0
	misses := 0
	other := 0

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		status := parts[1]
		total++

		// TODO: подсчитайте hits, misses, other
		_ = status
	}

	// TODO: Выведите статистику
	_ = total
	_ = hits
	_ = misses
	_ = other
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	total := 0
	hits := 0
	misses := 0
	other := 0

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		status := parts[1]
		total++

		switch status {
		case "HIT":
			hits++
		case "MISS":
			misses++
		default:
			other++
		}
	}

	fmt.Printf("total: %d\n", total)
	fmt.Printf("HIT: %d (%d%%)\n", hits, hits*100/total)
	fmt.Printf("MISS: %d (%d%%)\n", misses, misses*100/total)
	fmt.Printf("other: %d (%d%%)\n", other, other*100/total)
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "/api/users HIT\n/api/posts HIT\n/api/comments MISS\n/api/users HIT\n/static/app.js BYPASS",
						ExpectedOutput: "total: 5\nHIT: 3 (60%)\nMISS: 1 (20%)\nother: 1 (20%)",
					},
					{
						Input:          "/page HIT\n/page HIT\n/page HIT\n/page MISS",
						ExpectedOutput: "total: 4\nHIT: 3 (75%)\nMISS: 1 (25%)\nother: 0 (0%)",
					},
				},
			},
		},
	}
}

// ── Урок 7: Rate Limiting и безопасность ───────────────────────

func lesson_nginx_rate_limiting() L {
	return L{
		Slug: "nginx-rate-limiting", Title: "Rate Limiting и безопасность", Order: 7,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Rate Limiting и безопасность</h1>

<h2>Зачем нужен Rate Limiting?</h2>
<p>Rate limiting защищает от:</p>
<ul>
<li><strong>DDoS-атак</strong> — миллионы запросов с целью положить сервер</li>
<li><strong>Brute-force</strong> — подбор паролей, API-ключей</li>
<li><strong>Scraping</strong> — массовый парсинг контента</li>
<li><strong>Abuse</strong> — один клиент потребляет все ресурсы</li>
</ul>

<h2>limit_req — ограничение запросов</h2>
<pre><code>http {
    # Определяем зону rate limiting
    # $binary_remote_addr — по IP клиента (16 байт vs ~64 для $remote_addr)
    # zone=api:10m — 10 МБ shared memory (~160,000 IP-адресов)
    # rate=10r/s — 10 запросов в секунду на один IP
    limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;
    limit_req_zone $binary_remote_addr zone=login:1m rate=1r/s;

    server {
        location /api/ {
            # burst=20 — допускаем "всплеск" 20 запросов сверх лимита
            # nodelay — обрабатывать burst сразу (не задерживать в очереди)
            limit_req zone=api burst=20 nodelay;
            limit_req_status 429;    # HTTP 429 Too Many Requests
        }

        location /login {
            # Строгий лимит для логина — 1 запрос/сек, burst=5
            limit_req zone=login burst=5;
            # Без nodelay — запросы сверх rate ставятся в очередь
        }
    }
}</code></pre>

<h2>Как работает Leaky Bucket (алгоритм limit_req)</h2>
<pre><code>         Запросы
           │ │ │ │ │ │ │
           ▼ ▼ ▼ ▼ ▼ ▼ ▼
    ┌─────────────────────┐
    │    Bucket (burst)    │ ← Overflow = 429 Too Many Requests
    │  ┌─────────────────┐│
    │  │ req req req req ││ ← Запросы ждут в "ведре"
    │  └────────┬────────┘│
    └───────────┼─────────┘
                │ (rate = утечка)
                ▼
         Обработка (1 req каждые 100мс при rate=10r/s)</code></pre>

<p><strong>rate=10r/s</strong> означает: из "ведра" вытекает 1 запрос каждые 100мс.</p>
<p><strong>burst=20</strong> — размер "ведра". Если пришло 25 запросов разом: 1 обрабатывается, 20 в ведре, 4 отвергаются (429).</p>
<p><strong>nodelay</strong> — обрабатывать burst сразу, не по одному. Важно для API.</p>

<h2>limit_conn — ограничение соединений</h2>
<pre><code>http {
    limit_conn_zone $binary_remote_addr zone=addr:10m;

    server {
        location /download/ {
            limit_conn addr 5;          # Макс. 5 одновременных соединений с одного IP
            limit_conn_status 429;
            limit_rate 100k;            # Скорость скачивания — 100 КБ/с на соединение
            limit_rate_after 1m;        # Первый 1 МБ без ограничений
        }
    }
}</code></pre>

<h2>IP Allow/Deny</h2>
<pre><code># Закрыть админку для всех кроме офиса
location /admin/ {
    allow 10.0.0.0/8;        # Офисная сеть
    allow 192.168.1.0/24;    # VPN
    deny all;                # Все остальные — 403 Forbidden
    proxy_pass http://admin_backend;
}

# Заблокировать известных абьюзеров
location / {
    deny 1.2.3.4;
    deny 5.6.7.0/24;
    allow all;
    proxy_pass http://backend;
}

# Порядок важен! Nginx проверяет сверху вниз, первое совпадение побеждает.</code></pre>

<h2>WAF-подход: блокировка подозрительных запросов</h2>
<pre><code># Блокировка SQL injection попыток
location / {
    if ($query_string ~* "union.*select") { return 403; }
    if ($query_string ~* "insert.*into") { return 403; }
    if ($request_uri ~* "\.\.") { return 403; }    # Path traversal

    # Блокировка user-agents ботов
    if ($http_user_agent ~* "bot|crawl|spider|scrape") {
        return 403;
    }

    proxy_pass http://backend;
}</code></pre>
<p><strong>Важно:</strong> это НЕ полноценный WAF. Серьёзная защита — ModSecurity или CloudFlare.</p>

<h2>Интеграция с fail2ban</h2>
<pre><code># /etc/fail2ban/filter.d/nginx-limit-req.conf
[Definition]
failregex = limiting requests.*client: &lt;HOST&gt;

# /etc/fail2ban/jail.local
[nginx-limit-req]
enabled = true
filter = nginx-limit-req
logpath = /var/log/nginx/error.log
maxretry = 5          # 5 rate-limit нарушений
findtime = 60         # за 60 секунд
bantime = 3600        # бан на 1 час

# fail2ban добавит правило iptables:
# iptables -I INPUT -s 1.2.3.4 -j DROP</code></pre>

<h2>DDoS Mitigation</h2>
<pre><code># Базовые меры в Nginx
http {
    # Ограничить размер запросов
    client_body_timeout 5s;
    client_header_timeout 5s;
    client_max_body_size 1m;

    # Закрыть slowloris
    keepalive_timeout 5s;
    send_timeout 5s;

    # Лимит на запросы
    limit_req_zone $binary_remote_addr zone=global:20m rate=50r/s;

    server {
        limit_req zone=global burst=100 nodelay;

        # Отвергать запросы без Host header (сканеры)
        if ($host = "") { return 444; }  # 444 = drop connection

        # Лимит на размер заголовков (против Header floods)
        large_client_header_buffers 2 1k;
    }
}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "В чём разница между limit_req с nodelay и без него?",
				Options:     []string{"nodelay отключает rate limiting", "Без nodelay запросы из burst обрабатываются по одному с задержкой (rate), с nodelay — burst обрабатывается сразу, но дальнейшие запросы сверх burst+rate отвергаются", "nodelay отключает burst", "nodelay удваивает rate"},
				Correct:     1,
				Explanation: "При rate=10r/s burst=20 без nodelay: если пришло 20 запросов разом — они встанут в очередь и будут выпускаться по 1 каждые 100мс. Клиент ждёт до 2с. С nodelay — все 20 обрабатываются СРАЗУ, но 'ведро' заполнено и следующие запросы будут отвергаться пока оно не 'вытечет'. Для API — всегда nodelay.",
			},
			{
				Question:    "Почему используется $binary_remote_addr вместо $remote_addr в limit_req_zone?",
				Options:     []string{"binary_remote_addr быстрее парсится", "$binary_remote_addr занимает 16 байт (IPv6) вместо ~45 байт текстового IP — в shared memory зоне 10м помещается в ~3 раза больше записей", "binary_remote_addr поддерживает IPv6", "Нет разницы"},
				Correct:     1,
				Explanation: "В 10 МБ shared memory: с $remote_addr (текст, ~45-64 байта + state) → ~100K записей. С $binary_remote_addr (16 байт для IPv6, 4 для IPv4) → ~160K записей. При DDoS с тысяч IP это критично — с текстовым IP зона быстрее переполнится.",
			},
			{
				Question:    "Что такое HTTP-статус 444 в Nginx и зачем он используется?",
				Options:     []string{"Это стандартный HTTP-ответ 'Not Found'", "Это специальный код Nginx означающий 'закрыть TCP-соединение без отправки ответа' — экономит ресурсы при отсечении нежелательных запросов", "Это ошибка сервера", "Это редирект"},
				Correct:     1,
				Explanation: "444 — нестандартный код, специфичный для Nginx. При return 444 Nginx закрывает TCP-соединение НЕ отправляя никакого HTTP-ответа. Это экономит bandwidth и CPU: ботам/сканерам не тратим ресурсы даже на формирование ответа. Используется для drop вредоносного трафика.",
			},
			{
				Question:    "Почему порядок директив allow/deny важен?",
				Options:     []string{"Для красоты кода", "Nginx проверяет правила сверху вниз и применяет ПЕРВОЕ совпавшее — если deny all стоит перед allow, никто не получит доступ", "Порядок не важен, Nginx сортирует автоматически", "Для совместимости с Apache"},
				Correct:     1,
				Explanation: "allow/deny работают как firewall rules — первое совпадение останавливает проверку. allow 10.0.0.0/8; deny all; → офис пройдёт, остальные нет. Если поменять местами: deny all; allow 10.0.0.0/8; → НИКТО не пройдёт, потому что deny all матчит все IP первым.",
			},
		},
		Tasks: []T{
			{
				Title:      "Реализация Token Bucket rate limiter",
				Difficulty: "medium",
				Description: `<p>Реализуйте алгоритм Token Bucket для rate limiting.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка: <code>rate burst</code> (rate = токенов/сек, burst = макс. токенов в ведре)</li>
<li>Остальные строки: <code>timestamp client_id</code> — запрос от клиента в момент timestamp</li>
</ol>
<p><strong>Алгоритм Token Bucket:</strong></p>
<ul>
<li>Каждый клиент имеет своё "ведро" начинающее с burst токенов</li>
<li>Каждую секунду добавляется rate токенов (но не больше burst)</li>
<li>Каждый запрос тратит 1 токен</li>
<li>Если токенов 0 — запрос отклонён</li>
</ul>
<p>Для каждого запроса выведите: <code>ALLOW client_id</code> или <code>DENY client_id</code></p>`,
				Hints: `<p>При каждом запросе вычисляйте сколько токенов добавилось с последнего запроса этого клиента: elapsed * rate. Потом проверяйте >= 1 токен.</p>`,
				Glossary: []GlossaryItem{
					{Term: "Token Bucket", Definition: "Алгоритм rate limiting: 'ведро' заполняется токенами с фиксированной скоростью. Каждый запрос тратит токен. Пустое ведро = отказ."},
					{Term: "Burst", Definition: "Максимальная вместимость ведра. Позволяет кратковременные всплески трафика сверх rate."},
					{Term: "rate vs burst", Definition: "rate = долгосрочная скорость (сколько добавляется). burst = допустимый мгновенный всплеск (размер ведра)."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type bucket struct {
	tokens    float64
	lastTime  int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var rate, burst int
	fmt.Sscan(scanner.Text(), &rate, &burst)

	clients := make(map[string]*bucket)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		var ts int
		fmt.Sscan(parts[0], &ts)
		clientID := parts[1]

		// TODO: Получить или создать bucket для клиента
		// Новый клиент начинает с burst токенов
		// Пополнить токены: elapsed = ts - lastTime; добавить elapsed * rate (не больше burst)
		// Если tokens >= 1: tokens--; ALLOW
		// Иначе: DENY
		_ = clientID
		_ = ts
		_ = rate
		_ = burst
		_ = clients
	}
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type bucket struct {
	tokens   float64
	lastTime int
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var rate, burst int
	fmt.Sscan(scanner.Text(), &rate, &burst)

	clients := make(map[string]*bucket)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		var ts int
		fmt.Sscan(parts[0], &ts)
		clientID := parts[1]

		b, exists := clients[clientID]
		if !exists {
			b = &bucket{tokens: float64(burst), lastTime: ts}
			clients[clientID] = b
		}

		elapsed := ts - b.lastTime
		b.tokens += float64(elapsed * rate)
		if b.tokens > float64(burst) {
			b.tokens = float64(burst)
		}
		b.lastTime = ts

		if b.tokens >= 1 {
			b.tokens--
			fmt.Printf("ALLOW %s\n", clientID)
		} else {
			fmt.Printf("DENY %s\n", clientID)
		}
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "2 5\n0 userA\n0 userA\n0 userA\n0 userA\n0 userA\n0 userA\n1 userA",
						ExpectedOutput: "ALLOW userA\nALLOW userA\nALLOW userA\nALLOW userA\nALLOW userA\nDENY userA\nALLOW userA",
					},
					{
						Input:          "1 2\n0 client1\n0 client1\n0 client1\n0 client2\n1 client1",
						ExpectedOutput: "ALLOW client1\nALLOW client1\nDENY client1\nALLOW client2\nALLOW client1",
					},
				},
			},
			{
				Title:      "Фильтр allow/deny правил",
				Difficulty: "easy",
				Description: `<p>Реализуйте логику allow/deny Nginx.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — количество правил N</li>
<li>Следующие N строк — правила: <code>allow IP</code> или <code>deny IP</code> (или "all" вместо IP)</li>
<li>Остальные строки — IP-адреса для проверки</li>
</ol>
<p><strong>Логика:</strong> правила проверяются по порядку (сверху вниз). Первое совпавшее правило определяет результат. "all" матчит любой IP.</p>
<p>Для каждого IP выведите: <code>IP: ALLOWED</code> или <code>IP: DENIED</code></p>`,
				Hints: `<p>Для каждого IP пройдитесь по правилам в порядке. Если rule.ip == checkIP или rule.ip == "all" — применяйте это правило и выходите.</p>`,
				Glossary: []GlossaryItem{
					{Term: "allow/deny", Definition: "Директивы контроля доступа Nginx по IP-адресу. Проверяются последовательно, первое совпадение определяет результат."},
					{Term: "deny all", Definition: "Запретить доступ всем IP-адресам. Обычно ставится последним правилом после allow для нужных IP."},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type rule struct {
	action string // "allow" or "deny"
	ip     string // IP or "all"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	rules := make([]rule, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		rules[i] = rule{action: parts[0], ip: parts[1]}
	}

	for scanner.Scan() {
		ip := scanner.Text()
		// TODO: Проверить IP по правилам (первое совпадение побеждает)
		// Если rule.ip == ip или rule.ip == "all" → применить action
		// action "allow" → "IP: ALLOWED"
		// action "deny" → "IP: DENIED"
		_ = ip
		_ = rules
	}
}`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type rule struct {
	action string
	ip     string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	rules := make([]rule, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.Fields(scanner.Text())
		rules[i] = rule{action: parts[0], ip: parts[1]}
	}

	for scanner.Scan() {
		ip := scanner.Text()
		result := "DENIED"
		for _, r := range rules {
			if r.ip == ip || r.ip == "all" {
				if r.action == "allow" {
					result = "ALLOWED"
				} else {
					result = "DENIED"
				}
				break
			}
		}
		fmt.Printf("%s: %s\n", ip, result)
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3\nallow 10.0.0.1\nallow 10.0.0.2\ndeny all\n10.0.0.1\n10.0.0.2\n192.168.1.1\n8.8.8.8",
						ExpectedOutput: "10.0.0.1: ALLOWED\n10.0.0.2: ALLOWED\n192.168.1.1: DENIED\n8.8.8.8: DENIED",
					},
					{
						Input:          "2\ndeny 1.2.3.4\nallow all\n1.2.3.4\n5.6.7.8\n9.9.9.9",
						ExpectedOutput: "1.2.3.4: DENIED\n5.6.7.8: ALLOWED\n9.9.9.9: ALLOWED",
					},
				},
			},
			{
				Title:      "Детектор DDoS-паттернов",
				Difficulty: "hard",
				Description: `<p>Напишите программу, которая анализирует access-лог и обнаруживает потенциальные DDoS-атаки.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка: <code>threshold window</code> (порог запросов и окно в секундах)</li>
<li>Остальные строки: <code>timestamp ip</code> — лог-записи</li>
</ol>
<p><strong>Логика:</strong></p>
<ul>
<li>Для каждого IP подсчитайте количество запросов в скользящем окне (последние window секунд)</li>
<li>Если count > threshold — IP считается атакующим</li>
</ul>
<p><strong>Формат вывода:</strong> в конце выведите все заблокированные IP (в порядке первого нарушения), по одному на строку:</p>
<pre><code>BLOCK: IP (N requests in Ws window)</code></pre>
<p>Где N — количество запросов, W — window.</p>`,
				Hints: `<p>Для каждого IP храните список timestamps. При каждом новом запросе удалите из списка все timestamps старше (current - window). Если len(list) > threshold — IP в бан.</p>`,
				Glossary: []GlossaryItem{
					{Term: "DDoS", Definition: "Distributed Denial of Service — распределённая атака множеством запросов с целью исчерпать ресурсы сервера."},
					{Term: "Sliding window", Definition: "Скользящее окно — метод подсчёта событий за последние N секунд. Устаревшие события удаляются."},
					{Term: "fail2ban", Definition: "Утилита Linux, анализирует логи и блокирует IP на уровне iptables при превышении порога нарушений."},
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
	var threshold, window int
	fmt.Sscan(scanner.Text(), &threshold, &window)

	// Для каждого IP — список timestamps запросов
	requests := make(map[string][]int)
	// Заблокированные IP (в порядке обнаружения)
	blocked := make(map[string]bool)
	var blockedOrder []string

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		var ts int
		fmt.Sscan(parts[0], &ts)
		ip := parts[1]

		// TODO:
		// 1. Добавить ts к requests[ip]
		// 2. Отфильтровать старые (ts - entry < window для каждого entry)
		// 3. Если len > threshold и IP ещё не в blocked → добавить в blocked
		_ = ts
		_ = ip
		_ = requests
		_ = blocked
		_ = blockedOrder
	}

	// TODO: Вывести заблокированные IP
	_ = threshold
	_ = window
}`,
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
	var threshold, window int
	fmt.Sscan(scanner.Text(), &threshold, &window)

	requests := make(map[string][]int)
	blocked := make(map[string]bool)
	var blockedOrder []string
	blockedCount := make(map[string]int)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		var ts int
		fmt.Sscan(parts[0], &ts)
		ip := parts[1]

		requests[ip] = append(requests[ip], ts)

		// Filter old entries
		filtered := make([]int, 0)
		for _, t := range requests[ip] {
			if ts-t < window {
				filtered = append(filtered, t)
			}
		}
		requests[ip] = filtered

		if len(filtered) > threshold && !blocked[ip] {
			blocked[ip] = true
			blockedOrder = append(blockedOrder, ip)
			blockedCount[ip] = len(filtered)
		}
	}

	for _, ip := range blockedOrder {
		fmt.Printf("BLOCK: %s (%d requests in %ds window)\n", ip, blockedCount[ip], window)
	}
}</code></pre>`,
				TestCases: []TestCase{
					{
						Input:          "3 10\n1 192.168.1.1\n2 192.168.1.1\n3 192.168.1.1\n4 192.168.1.1\n5 10.0.0.1\n6 10.0.0.1",
						ExpectedOutput: "BLOCK: 192.168.1.1 (4 requests in 10s window)",
					},
					{
						Input:          "2 5\n1 1.1.1.1\n2 1.1.1.1\n3 1.1.1.1\n3 2.2.2.2\n4 2.2.2.2\n5 2.2.2.2",
						ExpectedOutput: "BLOCK: 1.1.1.1 (3 requests in 5s window)\nBLOCK: 2.2.2.2 (3 requests in 5s window)",
					},
				},
			},
		},
	}
}
