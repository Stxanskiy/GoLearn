package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Grafana + Prometheus — мониторинг и observability
// ════════════════════════════════════════════════════════════════

func mod_grafana() M {
	return M{
		Slug:          "grafana-prometheus",
		Title:         "Grafana + Prometheus: мониторинг",
		Description:   "Полный стек мониторинга: Prometheus для сбора метрик, PromQL для запросов, Grafana для визуализации и алертинга. От инструментирования Go-сервиса до SLO-based alerting в проде.",
		Order:         23,
		Track:         "devops",
		Difficulty:    "advanced",
		Prerequisites: []string{"linux-fundamentals", "http-server"},
		Lessons: []L{
			lesson_observability_fundamentals(),
			lesson_prometheus_architecture(),
			lesson_promql_deep_dive(),
			lesson_exporters_instrumentation(),
			lesson_grafana_dashboards(),
			lesson_alerting_pipeline(),
			lesson_production_monitoring(),
		},
	}
}

// ── Урок 1: Observability Fundamentals ──────────────────────────

func lesson_observability_fundamentals() L {
	return L{
		Slug: "observability-fundamentals", Title: "Основы Observability", Order: 1,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Основы Observability</h1>

<h2>Observability ≠ Monitoring</h2>
<p><strong>Мониторинг</strong> отвечает на вопрос "что сломалось?" — заранее заданные дашборды и алерты. <strong>Observability</strong> отвечает на вопрос "почему сломалось?" — возможность задавать произвольные вопросы системе без деплоя нового кода.</p>

<p>Три столпа observability:</p>

<h2>1. Метрики (Metrics)</h2>
<p>Числовые значения, агрегированные во времени. Дёшевы в хранении, быстры в запросах.</p>
<pre><code># Примеры метрик
http_requests_total{method="GET", status="200"}  → 142857
http_request_duration_seconds{quantile="0.99"}   → 0.250
process_resident_memory_bytes                     → 52428800</code></pre>
<p><strong>Когда использовать:</strong> тренды, алерты, capacity planning. "Сколько запросов в секунду? Какой 99-перцентиль латенси?"</p>

<h2>2. Логи (Logs)</h2>
<p>Текстовые записи событий с временной меткой. Дорогие в хранении, медленные в поиске без индексации.</p>
<pre><code>2024-01-15T10:23:45Z level=ERROR msg="connection refused"
    host="db-primary.internal" port=5432 retry=3 trace_id=abc123</code></pre>
<p><strong>Когда использовать:</strong> детали конкретного события, debug. "Что именно произошло в 10:23?"</p>

<h2>3. Трейсы (Traces)</h2>
<p>Путь запроса через распределённую систему. Каждый span — одна операция.</p>
<pre><code>TraceID: abc123
├── [API Gateway] 250ms
│   ├── [Auth Service] 15ms
│   └── [Order Service] 230ms
│       ├── [PostgreSQL] 45ms
│       └── [Redis Cache] 2ms (MISS → hit DB)</code></pre>
<p><strong>Когда использовать:</strong> латенси в распределённых системах. "Почему этот запрос шёл 5 секунд?"</p>

<h2>SLI / SLO / SLA</h2>
<table>
<tr><th>Термин</th><th>Что это</th><th>Пример</th></tr>
<tr><td><strong>SLI</strong> (Service Level Indicator)</td><td>Конкретная метрика качества</td><td>% запросов быстрее 200ms</td></tr>
<tr><td><strong>SLO</strong> (Service Level Objective)</td><td>Целевое значение SLI</td><td>99.9% запросов быстрее 200ms</td></tr>
<tr><td><strong>SLA</strong> (Service Level Agreement)</td><td>Контракт с клиентом (штрафы)</td><td>"Uptime 99.9%, иначе возврат"</td></tr>
</table>

<p><strong>Error Budget:</strong> если SLO = 99.9%, то за 30 дней можно "потратить" 43.2 минуты даунтайма. Это и есть error budget.</p>

<h2>Golden Signals (Google SRE Book)</h2>
<p>Четыре сигнала, которые описывают здоровье любого сервиса:</p>
<ol>
<li><strong>Latency</strong> — время обработки запроса (отдельно для success и error!)</li>
<li><strong>Traffic</strong> — сколько запросов приходит (RPS)</li>
<li><strong>Errors</strong> — доля неуспешных запросов (5xx, таймауты)</li>
<li><strong>Saturation</strong> — насколько ресурс загружен (CPU, memory, disk IO, goroutines)</li>
</ol>

<h2>RED Method (для сервисов)</h2>
<ul>
<li><strong>R</strong>ate — запросов в секунду</li>
<li><strong>E</strong>rrors — количество ошибок в секунду</li>
<li><strong>D</strong>uration — распределение латенси (гистограмма)</li>
</ul>

<h2>USE Method (для ресурсов: CPU, RAM, disk, network)</h2>
<ul>
<li><strong>U</strong>tilization — % использования (CPU 85%)</li>
<li><strong>S</strong>aturation — очередь ожидания (load average, disk queue)</li>
<li><strong>E</strong>rrors — аппаратные/системные ошибки</li>
</ul>

<h2>Когда что применять</h2>
<pre><code>┌─────────────────────┐     ┌─────────────────────┐
│  RED — для сервисов │     │  USE — для ресурсов  │
│  (API, микросервис) │     │  (CPU, RAM, диск)    │
│  Rate/Errors/Duratn │     │  Util/Satur/Errors   │
└─────────────────────┘     └─────────────────────┘
         ↓                            ↓
    Golden Signals = RED + Saturation (объединяет оба)</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Чем observability отличается от мониторинга?",
				Options:     []string{"Ничем, это синонимы", "Observability позволяет задавать произвольные вопросы системе без деплоя нового кода", "Observability — это только логи", "Мониторинг дороже observability"},
				Correct:     1,
				Explanation: "Мониторинг = заранее заданные дашборды и алерты (known-unknowns). Observability = возможность исследовать unknown-unknowns, задавая ad-hoc запросы к метрикам, логам, трейсам.",
			},
			{
				Question:    "Что такое Error Budget при SLO 99.9% за 30 дней?",
				Options:     []string{"0 минут даунтайма", "43.2 минуты допустимого даунтайма", "4.32 часа допустимого даунтайма", "Бесконечное время"},
				Correct:     1,
				Explanation: "30 дней = 43200 минут. 0.1% от 43200 = 43.2 минуты. Это бюджет ошибок — сколько можно «потратить» и не нарушить SLO.",
			},
			{
				Question:    "Какой метод применяется для мониторинга CPU и памяти?",
				Options:     []string{"RED Method", "USE Method", "Golden Signals", "SLI/SLO"},
				Correct:     1,
				Explanation: "USE (Utilization, Saturation, Errors) — метод для ресурсов (CPU, RAM, диск, сеть). RED — для сервисов (API endpoints).",
			},
			{
				Question:    "Почему важно измерять латенси ошибочных запросов отдельно от успешных?",
				Options:     []string{"Не важно, можно объединять", "Ошибки часто быстрее (мгновенный 500) или медленнее (таймаут), что искажает общую картину", "Ошибки не имеют латенси", "Для красоты графиков"},
				Correct:     1,
				Explanation: "Запрос вернувший 500 за 1ms (мгновенный reject) или за 30s (таймаут) — два разных паттерна. Если смешать с success latency, можно пропустить деградацию.",
			},
		},
		Tasks: []T{
			{
				Title:      "Калькулятор Error Budget",
				Difficulty: "easy",
				Description: `<p>Напиши программу, которая вычисляет error budget.</p>
<p><strong>Вход:</strong> две строки:</p>
<ol>
<li>SLO в процентах (например: <code>99.9</code>)</li>
<li>Период в днях (например: <code>30</code>)</li>
</ol>
<p><strong>Выход:</strong> допустимое количество минут даунтайма, округлённое до одного знака после запятой.</p>
<p>Формула: <code>budget_minutes = (100 - SLO) / 100 * days * 24 * 60</code></p>`,
				Hints: `<p>Используй <code>fmt.Scanf</code> для чтения SLO и дней. Для форматирования используй <code>fmt.Printf("%.1f\n", result)</code>.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
	var slo float64
	var days int
	fmt.Scan(&slo)
	fmt.Scan(&days)

	budget := (100.0 - slo) / 100.0 * float64(days) * 24.0 * 60.0
	fmt.Printf("%.1f\n", budget)
}</code></pre>`,
				StarterCode: `package main

import "fmt"

func main() {
	var slo float64
	var days int
	fmt.Scan(&slo)
	fmt.Scan(&days)

	// TODO: вычисли error budget в минутах
	// budget = (100 - SLO) / 100 * days * 24 * 60
	// Выведи с одним знаком после запятой
	_ = slo
	_ = days
}`,
				Glossary: []GlossaryItem{
					{Term: "fmt.Scan(&x)", Definition: "Читает значение из stdin и сохраняет в переменную x."},
					{Term: "fmt.Printf(\"%.1f\", x)", Definition: "Выводит float с 1 знаком после запятой."},
					{Term: "float64(n)", Definition: "Преобразует int в float64 для деления без потери дробной части."},
				},
				TestCases: []TestCase{
					{Input: "99.9\n30", ExpectedOutput: "43.2"},
					{Input: "99.0\n30", ExpectedOutput: "432.0"},
					{Input: "99.99\n7", ExpectedOutput: "1.0"},
					{Input: "95.0\n30", ExpectedOutput: "21600.0"},
				},
			},
			{
				Title:      "Классификатор Golden Signals",
				Difficulty: "medium",
				Description: `<p>На вход подаются строки с именами метрик. Для каждой определи, к какому Golden Signal она относится.</p>
<p><strong>Вход:</strong> каждая строка — имя метрики.</p>
<p><strong>Выход:</strong> для каждой строки выведи <code>имя_метрики: SIGNAL</code>, где SIGNAL — один из: <code>LATENCY</code>, <code>TRAFFIC</code>, <code>ERRORS</code>, <code>SATURATION</code>.</p>
<p><strong>Правила классификации:</strong></p>
<ul>
<li>Содержит "duration" или "latency" → <code>LATENCY</code></li>
<li>Содержит "requests_total" или "rps" → <code>TRAFFIC</code></li>
<li>Содержит "errors" или "5xx" → <code>ERRORS</code></li>
<li>Содержит "cpu" или "memory" или "saturation" → <code>SATURATION</code></li>
</ul>`,
				Hints: `<p>Используй <code>strings.Contains(metric, substr)</code> для проверки вхождения подстроки. Проверяй в порядке приоритета.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func classify(metric string) string {
	m := strings.ToLower(metric)
	switch {
	case strings.Contains(m, "duration") || strings.Contains(m, "latency"):
		return "LATENCY"
	case strings.Contains(m, "requests_total") || strings.Contains(m, "rps"):
		return "TRAFFIC"
	case strings.Contains(m, "errors") || strings.Contains(m, "5xx"):
		return "ERRORS"
	case strings.Contains(m, "cpu") || strings.Contains(m, "memory") || strings.Contains(m, "saturation"):
		return "SATURATION"
	default:
		return "UNKNOWN"
	}
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		metric := scanner.Text()
		fmt.Printf("%s: %s\n", metric, classify(metric))
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func classify(metric string) string {
	// TODO: определи к какому Golden Signal относится метрика
	// Используй strings.Contains для проверки вхождения подстроки
	_ = strings.Contains("", "")
	return "UNKNOWN"
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		metric := scanner.Text()
		fmt.Printf("%s: %s\n", metric, classify(metric))
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "strings.Contains(s, substr)", Definition: "Возвращает true если строка s содержит подстроку substr."},
					{Term: "strings.ToLower(s)", Definition: "Приводит строку к нижнему регистру для case-insensitive сравнения."},
					{Term: "switch без выражения", Definition: "switch { case cond1: ... } — каждый case проверяет своё условие, как цепочка if/else if."},
				},
				TestCases: []TestCase{
					{Input: "http_request_duration_seconds\nhttp_requests_total\nhttp_errors_total\nprocess_cpu_seconds", ExpectedOutput: "http_request_duration_seconds: LATENCY\nhttp_requests_total: TRAFFIC\nhttp_errors_total: ERRORS\nprocess_cpu_seconds: SATURATION"},
					{Input: "api_latency_ms\ngateway_rps\ndb_5xx_count\nnode_memory_bytes", ExpectedOutput: "api_latency_ms: LATENCY\ngateway_rps: TRAFFIC\ndb_5xx_count: ERRORS\nnode_memory_bytes: SATURATION"},
				},
			},
		},
	}
}

// ── Урок 2: Prometheus Architecture ─────────────────────────────

func lesson_prometheus_architecture() L {
	return L{
		Slug: "prometheus-architecture", Title: "Архитектура Prometheus", Order: 2,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Архитектура Prometheus</h1>

<h2>Pull Model — ключевое отличие</h2>
<p>В отличие от push-систем (StatsD, InfluxDB Telegraf), Prometheus <strong>сам ходит за метриками</strong> к целям (targets). Это называется <strong>scraping</strong>.</p>

<pre><code>┌──────────────┐     GET /metrics      ┌─────────────────┐
│  Prometheus  │ ────────────────────→  │  Your Go Service │
│   (scraper)  │ ←────────────────────  │  :8080/metrics   │
│              │    text/plain           └─────────────────┘
│   ┌──────┐   │
│   │ TSDB │   │     GET /metrics      ┌─────────────────┐
│   └──────┘   │ ────────────────────→  │  node_exporter   │
└──────────────┘                        │  :9100/metrics   │
                                        └─────────────────┘</code></pre>

<h3>Преимущества pull-модели:</h3>
<ul>
<li>Prometheus знает, жив ли target (если scrape failed → up=0)</li>
<li>Targets не знают о Prometheus — нет зависимости</li>
<li>Легко дебажить — можно открыть /metrics в браузере</li>
<li>Нет проблемы с backpressure (Prometheus контролирует частоту)</li>
</ul>

<h2>Компоненты Prometheus</h2>
<pre><code>┌─────────────────────────────────────────────────────────┐
│                     PROMETHEUS SERVER                     │
├─────────────┬───────────────┬───────────────────────────┤
│  Retrieval  │    TSDB       │      HTTP Server          │
│  (scraper)  │  (хранение)   │   (PromQL API + UI)       │
│             │               │                           │
│  - targets  │  - chunks     │  - /api/v1/query          │
│  - interval │  - WAL        │  - /api/v1/query_range    │
│  - timeout  │  - compaction │  - /graph (встроенный UI) │
└─────────────┴───────────────┴───────────────────────────┘</code></pre>

<h2>TSDB (Time Series Database)</h2>
<p>Prometheus хранит данные в собственной TSDB, оптимизированной для временных рядов:</p>
<ul>
<li><strong>WAL (Write-Ahead Log)</strong> — все записи сначала идут в WAL для durability</li>
<li><strong>Head block</strong> — последние 2 часа данных в памяти (быстрый доступ)</li>
<li><strong>Persistent blocks</strong> — старые данные на диске, сжаты</li>
<li><strong>Compaction</strong> — Prometheus периодически объединяет мелкие блоки в крупные</li>
</ul>

<h2>Service Discovery</h2>
<p>Prometheus умеет автоматически находить targets:</p>
<pre><code># prometheus.yml
scrape_configs:
  # Статический список
  - job_name: 'my-app'
    static_configs:
      - targets: ['app1:8080', 'app2:8080']

  # Kubernetes SD — автоматически находит pods
  - job_name: 'k8s-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true

  # Docker SD
  - job_name: 'docker'
    docker_sd_configs:
      - host: unix:///var/run/docker.sock

  # Consul SD
  - job_name: 'consul-services'
    consul_sd_configs:
      - server: 'consul:8500'</code></pre>

<h2>Federation — масштабирование</h2>
<p>Для крупных инфраструктур используют иерархию Prometheus:</p>
<pre><code>        ┌─────────────────────┐
        │  Global Prometheus   │  ← агрегированные метрики
        │  (long retention)    │
        └──────┬──────┬────────┘
               │      │
     ┌─────────┴┐   ┌─┴─────────┐
     │  DC-East  │   │  DC-West   │  ← per-datacenter
     │ Prometheus│   │ Prometheus │
     └─────┬─────┘   └─────┬─────┘
           │                │
    ┌──────┴──────┐  ┌──────┴──────┐
    │  Targets    │  │   Targets   │
    └─────────────┘  └─────────────┘</code></pre>

<h2>Retention и ресурсы</h2>
<ul>
<li>По умолчанию: 15 дней retention</li>
<li>Формула памяти: ~1-2 байта на sample, ~3KB на time series в head block</li>
<li>Для long-term storage → Thanos, Cortex, Mimir (remote write)</li>
</ul>

<h2>Конфигурация scrape</h2>
<pre><code>global:
  scrape_interval: 15s      # как часто scrape (по умолчанию)
  evaluation_interval: 15s  # как часто оценивать правила алертинга
  scrape_timeout: 10s       # таймаут одного scrape

scrape_configs:
  - job_name: 'go-app'
    scrape_interval: 5s     # override для конкретного job
    metrics_path: '/metrics'
    scheme: 'http'
    static_configs:
      - targets: ['localhost:8080']
        labels:
          env: 'production'
          team: 'backend'</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему Prometheus использует pull-модель, а не push?",
				Options:     []string{"Push медленнее", "Pull позволяет Prometheus знать, жив ли target, и контролировать нагрузку", "Pull проще реализовать", "Push не поддерживает labels"},
				Correct:     1,
				Explanation: "Pull-модель: Prometheus контролирует частоту scrape, может определить, что target недоступен (up=0), не страдает от backpressure. Targets не зависят от Prometheus.",
			},
			{
				Question:    "Что произойдёт если Prometheus не может scrape target?",
				Options:     []string{"Target перезапустится", "Метрика up=0 для этого target, данные за interval потеряны", "Prometheus упадёт", "Данные буферизируются на target"},
				Correct:     1,
				Explanation: "При неудачном scrape Prometheus записывает up=0. Данные за этот интервал теряются — Prometheus не хранит буфер на target. Это one of trade-offs pull-модели.",
			},
			{
				Question:    "Для чего нужна Federation в Prometheus?",
				Options:     []string{"Для бэкапов", "Для масштабирования — иерархия серверов Prometheus для больших инфраструктур", "Для шифрования", "Для аутентификации"},
				Correct:     1,
				Explanation: "Federation позволяет одному Prometheus scrape-ить агрегированные метрики с других Prometheus. Типичный паттерн: per-DC Prometheus + global Prometheus для общей картины.",
			},
		},
		Tasks: []T{
			{
				Title:      "Парсер prometheus.yml targets",
				Difficulty: "medium",
				Description: `<p>Напиши программу, которая парсит упрощённый формат конфигурации Prometheus и выводит список targets.</p>
<p><strong>Формат входа:</strong> строки вида <code>job_name target1 target2 ...</code> (имя job и через пробел его targets).</p>
<p><strong>Выход:</strong> для каждого target выведи <code>job_name → target</code>, по одному на строку.</p>`,
				Hints: `<p>Используй <code>strings.Fields(line)</code> для разделения строки по пробелам. Первый элемент — job name, остальные — targets.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		job := fields[0]
		for _, target := range fields[1:] {
			fmt.Printf("%s → %s\n", job, target)
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// TODO: fields[0] — имя job, fields[1:] — targets
		// Для каждого target выведи: jobName → target
		_ = fields
		_ = fmt.Sprintf("")
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "strings.Fields(s)", Definition: "Разделяет строку по любым пробельным символам, убирая пустые элементы. \"a  b  c\" → [\"a\", \"b\", \"c\"]"},
					{Term: "slice[1:]", Definition: "Slice от индекса 1 до конца — все элементы кроме первого."},
					{Term: "range slice", Definition: "Итерация по slice: for i, v := range slice { ... }"},
				},
				TestCases: []TestCase{
					{Input: "webapp app1:8080 app2:8080 app3:8080\nredis redis1:9121 redis2:9121", ExpectedOutput: "webapp → app1:8080\nwebapp → app2:8080\nwebapp → app3:8080\nredis → redis1:9121\nredis → redis2:9121"},
					{Input: "node node1:9100\npostgres pg1:9187 pg2:9187", ExpectedOutput: "node → node1:9100\npostgres → pg1:9187\npostgres → pg2:9187"},
				},
			},
			{
				Title:      "Симулятор scrape status",
				Difficulty: "medium",
				Description: `<p>Prometheus записывает метрику <code>up</code> для каждого target: 1 = scrape успешен, 0 = ошибка.</p>
<p><strong>Вход:</strong> строки формата <code>target status</code>, где status — "ok" или "fail".</p>
<p><strong>Выход:</strong> для каждого target выведи <code>up{instance="target"} VALUE</code>, где VALUE = 1 для ok, 0 для fail. После всех строк выведи итог: <code>healthy: N, unhealthy: M</code>.</p>`,
				Hints: `<p>Считай количество ok и fail в отдельных переменных. Используй <code>strings.Fields</code> для разделения.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	healthy, unhealthy := 0, 0

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		target, status := fields[0], fields[1]
		if status == "ok" {
			fmt.Printf("up{instance=\"%s\"} 1\n", target)
			healthy++
		} else {
			fmt.Printf("up{instance=\"%s\"} 0\n", target)
			unhealthy++
		}
	}
	fmt.Printf("healthy: %d, unhealthy: %d\n", healthy, unhealthy)
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	healthy, unhealthy := 0, 0

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		target, status := fields[0], fields[1]
		// TODO: если status == "ok" — вывести up{instance="target"} 1, иначе 0
		// Увеличить healthy или unhealthy
		_ = target
		_ = status
	}
	// TODO: вывести итог: healthy: N, unhealthy: M
	_ = healthy
	_ = unhealthy
}`,
				Glossary: []GlossaryItem{
					{Term: "up{instance=\"x\"}", Definition: "Стандартная метрика Prometheus: 1 = target доступен, 0 = scrape failed."},
					{Term: "fmt.Printf с кавычками", Definition: "Для вывода кавычек внутри строки используй \\\" : fmt.Printf(\"up{instance=\\\"%s\\\"}\"...)"},
				},
				TestCases: []TestCase{
					{Input: "app1:8080 ok\napp2:8080 fail\napp3:8080 ok", ExpectedOutput: "up{instance=\"app1:8080\"} 1\nup{instance=\"app2:8080\"} 0\nup{instance=\"app3:8080\"} 1\nhealthy: 2, unhealthy: 1"},
					{Input: "db:5432 fail\ncache:6379 fail", ExpectedOutput: "up{instance=\"db:5432\"} 0\nup{instance=\"cache:6379\"} 0\nhealthy: 0, unhealthy: 2"},
				},
			},
		},
	}
}

// ── Урок 3: PromQL Deep Dive ─────────────────────────────────────

func lesson_promql_deep_dive() L {
	return L{
		Slug: "promql-deep-dive", Title: "PromQL — язык запросов Prometheus", Order: 3,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>PromQL Deep Dive</h1>

<h2>Типы данных в PromQL</h2>
<ul>
<li><strong>Instant vector</strong> — набор time series, каждая с одним значением в момент времени</li>
<li><strong>Range vector</strong> — набор time series, каждая с массивом значений за период</li>
<li><strong>Scalar</strong> — одно число (1.5, 42)</li>
<li><strong>String</strong> — строка (почти не используется)</li>
</ul>

<h2>Selectors</h2>
<pre><code># Instant vector selector
http_requests_total                           # все series с этим именем
http_requests_total{method="GET"}             # фильтр по label
http_requests_total{status=~"5.."}            # regex: 500, 501, 502...
http_requests_total{method!="DELETE"}          # отрицание
http_requests_total{handler=~"/api/.*"}       # regex match

# Range vector selector — добавляем [duration]
http_requests_total[5m]      # последние 5 минут значений
http_requests_total[1h]      # последний час

# Offset — сдвиг во времени
http_requests_total offset 1h    # значение час назад
http_requests_total[5m] offset 1d  # 5 минут данных, но сутки назад</code></pre>

<h2>rate() vs irate()</h2>
<p>Для counter-метрик (которые только растут) нужно вычислять скорость изменения:</p>
<pre><code># rate() — средняя скорость за период (сглаженная)
rate(http_requests_total[5m])
# = (last_value - first_value) / duration_seconds

# irate() — мгновенная скорость (по двум последним точкам)
irate(http_requests_total[5m])
# = (last_value - prev_value) / time_diff

# ПРАВИЛО: rate для алертов, irate для графиков (более детальный)
# rate() корректно обрабатывает resets (перезапуск counter)</code></pre>

<h3>Частая ошибка: rate без [range]</h3>
<pre><code># НЕПРАВИЛЬНО — rate нужен range vector!
rate(http_requests_total)  # ERROR

# ПРАВИЛЬНО
rate(http_requests_total[5m])

# НЕПРАВИЛЬНО — range должен быть >= 4x scrape_interval
# Если scrape_interval = 15s, минимум rate()[1m] (4 * 15s = 60s)
rate(http_requests_total[15s])  # может дать пустой результат</code></pre>

<h2>Агрегации</h2>
<pre><code># sum — сумма по всем instances
sum(rate(http_requests_total[5m]))

# sum by — группировка
sum by (method) (rate(http_requests_total[5m]))

# sum without — исключение label
sum without (instance) (rate(http_requests_total[5m]))

# Другие агрегации:
avg, min, max, count, stddev, stdvar
topk(5, ...), bottomk(3, ...), quantile(0.99, ...)</code></pre>

<h2>histogram_quantile() — перцентили из гистограмм</h2>
<pre><code># 99-перцентиль латенси
histogram_quantile(0.99,
  sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
)

# ВАЖНО: le (less than or equal) — обязательный label для гистограмм
# Нельзя агрегировать по le вместе с другими labels!

# 50-перцентиль (медиана) по каждому endpoint:
histogram_quantile(0.50,
  sum by (le, handler) (rate(http_request_duration_seconds_bucket[5m]))
)</code></pre>

<h2>Recording Rules</h2>
<p>Предвычисленные запросы, которые сохраняются как новые time series:</p>
<pre><code># prometheus_rules.yml
groups:
  - name: sli_rules
    interval: 30s
    rules:
      - record: job:http_requests:rate5m
        expr: sum by (job) (rate(http_requests_total[5m]))

      - record: job:http_request_duration:p99
        expr: |
          histogram_quantile(0.99,
            sum by (job, le) (rate(http_request_duration_seconds_bucket[5m]))
          )</code></pre>
<p><strong>Зачем:</strong> ускорить тяжёлые запросы в дашбордах, упростить выражения в алертах.</p>

<h2>Полезные паттерны</h2>
<pre><code># Error rate
sum(rate(http_requests_total{status=~"5.."}[5m]))
/
sum(rate(http_requests_total[5m]))

# Saturation — горутины
go_goroutines / go_threads

# Availability (%)
1 - (
  sum(rate(http_requests_total{status=~"5.."}[5m]))
  /
  sum(rate(http_requests_total[5m]))
)

# QPS per instance
sum by (instance) (rate(http_requests_total[5m]))</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Чем отличается rate() от irate()?",
				Options:     []string{"Ничем", "rate() — средняя скорость за весь range, irate() — мгновенная по двум последним точкам", "irate() медленнее", "rate() только для gauge"},
				Correct:     1,
				Explanation: "rate() вычисляет среднюю скорость за [range], сглаживая всплески. irate() берёт только два последних sample — показывает мгновенную скорость. rate для алертов (стабильнее), irate для графиков (детальнее).",
			},
			{
				Question:    "Почему rate(http_requests_total[15s]) при scrape_interval=15s может вернуть пустой результат?",
				Options:     []string{"15s слишком большой интервал", "В range [15s] может быть только одна точка, а rate() нужны минимум две", "Prometheus не поддерживает 15s", "Это баг"},
				Correct:     1,
				Explanation: "rate() нужны минимум 2 точки в range. При scrape_interval=15s в окне [15s] будет максимум 1 точка. Правило: range >= 4 * scrape_interval (для надёжности при пропущенных scrape).",
			},
			{
				Question:    "Что делает histogram_quantile(0.99, sum by (le) (rate(..._bucket[5m])))?",
				Options:     []string{"Считает сумму", "Вычисляет 99-перцентиль из гистограммы (значение, ниже которого 99% запросов)", "Фильтрует по label le", "Удаляет 99% данных"},
				Correct:     1,
				Explanation: "histogram_quantile оценивает значение перцентиля из bucket-ов гистограммы. 0.99 → 99% запросов имеют значение НИЖЕ результата. le (less-or-equal) — границы bucket-ов.",
			},
			{
				Question:    "Зачем нужны Recording Rules?",
				Options:     []string{"Для красоты", "Предвычисляют тяжёлые PromQL-запросы и сохраняют как новые time series для быстрого доступа", "Для удаления метрик", "Для шифрования"},
				Correct:     1,
				Explanation: "Recording rules выполняются периодически и записывают результат как новую time series. Это ускоряет дашборды (pre-computed) и упрощает алерты (ссылаешься на короткое имя вместо длинного выражения).",
			},
		},
		Tasks: []T{
			{
				Title:      "Парсер PromQL selectors",
				Difficulty: "medium",
				Description: `<p>Напиши парсер, который извлекает имя метрики и labels из PromQL selector.</p>
<p><strong>Вход:</strong> строки с PromQL selectors формата <code>metric_name{label1="value1",label2="value2"}</code></p>
<p><strong>Выход:</strong> для каждой строки выведи имя метрики, затем labels по одному на строку в формате <code>  label = value</code>, затем пустую строку.</p>
<p>Если labels нет (просто <code>metric_name</code>) — выводи только имя и пустую строку.</p>`,
				Hints: `<p>Разделяй по "{" чтобы получить имя метрики. Затем удали "}" и разделяй по "," для получения labels. Каждый label разделяй по "=" и убирай кавычки через strings.Trim.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "{", 2)
		metric := parts[0]
		fmt.Println(metric)

		if len(parts) == 2 {
			labelsStr := strings.TrimSuffix(parts[1], "}")
			labels := strings.Split(labelsStr, ",")
			for _, l := range labels {
				kv := strings.SplitN(l, "=", 2)
				if len(kv) == 2 {
					key := strings.TrimSpace(kv[0])
					val := strings.Trim(kv[1], "\"")
					fmt.Printf("  %s = %s\n", key, val)
				}
			}
		}
		fmt.Println()
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: разделяй строку по "{" — первая часть = имя метрики
		// Если есть вторая часть — это labels: убери "}", разделяй по ","
		// Каждый label: разделяй по "=", убирай кавычки через strings.Trim(val, "\"")
		// Выводи: имя метрики, затем "  key = value" для каждого label, пустая строка в конце
		_ = line
		_ = strings.SplitN("", "", 0)
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "strings.SplitN(s, sep, 2)", Definition: "Разделяет строку максимум на 2 части. Полезно когда sep может встречаться в значении."},
					{Term: "strings.TrimSuffix(s, suffix)", Definition: "Удаляет суффикс из строки если он есть. TrimSuffix(\"abc}\", \"}\") → \"abc\""},
					{Term: "strings.Trim(s, cutset)", Definition: "Удаляет символы из cutset с обоих концов строки. Trim(`\"hello\"`, `\"`) → \"hello\""},
				},
				TestCases: []TestCase{
					{Input: `http_requests_total{method="GET",status="200"}`, ExpectedOutput: "http_requests_total\n  method = GET\n  status = 200\n"},
					{Input: "go_goroutines\nprocess_cpu_seconds{instance=\"app:8080\"}", ExpectedOutput: "go_goroutines\n\nprocess_cpu_seconds\n  instance = app:8080\n"},
				},
			},
			{
				Title:      "Вычисление rate() по точкам",
				Difficulty: "hard",
				Description: `<p>Реализуй упрощённый rate() — вычисление средней скорости роста counter-а.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка — число N (количество точек)</li>
<li>Следующие N строк — пары <code>timestamp value</code> (timestamp в секундах, value — значение counter-а)</li>
</ol>
<p><strong>Формула:</strong> <code>rate = (last_value - first_value) / (last_timestamp - first_timestamp)</code></p>
<p><strong>Выход:</strong> значение rate с точностью до 4 знаков после запятой.</p>
<p><strong>Обработка reset:</strong> если value[i] < value[i-1], это counter reset — считай что value[i] начался с 0 (добавь value[i] к аккумулятору). Итоговый total increase = сумма всех increases с учётом resets.</p>`,
				Hints: `<p>Пройди по всем парам соседних точек. Если текущее значение >= предыдущему, increase += current - prev. Если current < prev (reset), increase += current (считаем что counter сбросился в 0 и дорос до current). rate = total_increase / (last_ts - first_ts).</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	timestamps := make([]float64, n)
	values := make([]float64, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&timestamps[i], &values[i])
	}

	var totalIncrease float64
	for i := 1; i < n; i++ {
		if values[i] >= values[i-1] {
			totalIncrease += values[i] - values[i-1]
		} else {
			totalIncrease += values[i]
		}
	}

	duration := timestamps[n-1] - timestamps[0]
	rate := totalIncrease / duration
	fmt.Printf("%.4f\n", rate)
}</code></pre>`,
				StarterCode: `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	timestamps := make([]float64, n)
	values := make([]float64, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&timestamps[i], &values[i])
	}

	// TODO: вычисли total increase с обработкой counter resets
	// Если values[i] >= values[i-1]: increase += values[i] - values[i-1]
	// Если values[i] < values[i-1] (reset): increase += values[i]
	// rate = totalIncrease / (last_timestamp - first_timestamp)
	var totalIncrease float64
	_ = totalIncrease
	_ = timestamps
}`,
				Glossary: []GlossaryItem{
					{Term: "Counter reset", Definition: "Когда counter-метрика сбрасывается в 0 (перезапуск сервиса). Prometheus rate() корректно обрабатывает это."},
					{Term: "make([]float64, n)", Definition: "Создаёт slice из n элементов типа float64, инициализированных нулями."},
					{Term: "rate формула", Definition: "rate = total_increase / total_duration. Результат — среднее количество событий в секунду."},
				},
				TestCases: []TestCase{
					{Input: "4\n0 100\n15 150\n30 200\n45 250", ExpectedOutput: "3.3333"},
					{Input: "5\n0 900\n15 950\n30 1000\n45 50\n60 100", ExpectedOutput: "4.1667"},
					{Input: "3\n0 0\n60 120\n120 240", ExpectedOutput: "2.0000"},
				},
			},
			{
				Title:      "SLO калькулятор из метрик",
				Difficulty: "medium",
				Description: `<p>Вычисли текущий SLI (availability) из метрик запросов.</p>
<p><strong>Вход:</strong> строки формата <code>status_code count</code> — сколько запросов с каким статусом.</p>
<p><strong>Выход:</strong> два значения:</p>
<ol>
<li><code>availability: XX.XXX%</code> — процент успешных запросов (status < 500), 3 знака после запятой</li>
<li><code>error_budget_remaining: XX.XXX%</code> — при SLO=99.9%, сколько бюджета осталось. Формула: <code>(availability - slo) / (100 - slo) * 100</code>. Если отрицательное — вывести 0.000.</li>
</ol>`,
				Hints: `<p>Суммируй total и error (status >= 500) запросы. availability = (total - errors) / total * 100. Error budget remaining = max(0, (availability - 99.9) / 0.1 * 100).</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var total, errors int

	for scanner.Scan() {
		var status, count int
		fmt.Sscanf(scanner.Text(), "%d %d", &status, &count)
		total += count
		if status >= 500 {
			errors += count
		}
	}

	availability := float64(total-errors) / float64(total) * 100.0
	fmt.Printf("availability: %.3f%%\n", availability)

	slo := 99.9
	budgetRemaining := (availability - slo) / (100.0 - slo) * 100.0
	if budgetRemaining < 0 {
		budgetRemaining = 0
	}
	fmt.Printf("error_budget_remaining: %.3f%%\n", budgetRemaining)
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var total, errors int

	for scanner.Scan() {
		var status, count int
		fmt.Sscanf(scanner.Text(), "%d %d", &status, &count)
		total += count
		if status >= 500 {
			errors += count
		}
	}

	// TODO: вычисли availability = (total - errors) / total * 100
	// Вывести: availability: XX.XXX%
	// SLO = 99.9%. budget_remaining = (availability - 99.9) / 0.1 * 100
	// Если < 0, вывести 0.000
	// Вывести: error_budget_remaining: XX.XXX%
	_ = total
	_ = errors
}`,
				Glossary: []GlossaryItem{
					{Term: "SLI (availability)", Definition: "Доля успешных запросов: (total - errors) / total * 100%. Успешные = status < 500."},
					{Term: "Error budget remaining", Definition: "Сколько % бюджета ошибок осталось: (current_availability - SLO) / (100 - SLO) * 100. Отрицательное = бюджет исчерпан."},
					{Term: "fmt.Sscanf(s, format, &vars...)", Definition: "Парсит строку s по формату, записывая значения в переменные."},
				},
				TestCases: []TestCase{
					{Input: "200 9990\n404 5\n500 5", ExpectedOutput: "availability: 99.950%\nerror_budget_remaining: 50.000%"},
					{Input: "200 999\n500 1", ExpectedOutput: "availability: 99.900%\nerror_budget_remaining: 0.000%"},
					{Input: "200 10000\n201 500\n500 50\n502 10", ExpectedOutput: "availability: 99.432%\nerror_budget_remaining: 0.000%"},
				},
			},
		},
	}
}

// ── Урок 4: Exporters & Instrumentation ─────────────────────────

func lesson_exporters_instrumentation() L {
	return L{
		Slug: "exporters-instrumentation", Title: "Exporters и инструментирование Go", Order: 4,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Exporters и инструментирование Go-сервиса</h1>

<h2>Что такое Exporter?</h2>
<p>Exporter — программа, которая собирает метрики из системы/приложения и отдаёт их в формате Prometheus по HTTP.</p>

<h2>Популярные exporters</h2>
<table>
<tr><th>Exporter</th><th>Что мониторит</th><th>Порт</th></tr>
<tr><td>node_exporter</td><td>Linux: CPU, RAM, disk, network</td><td>9100</td></tr>
<tr><td>postgres_exporter</td><td>PostgreSQL: connections, queries, locks</td><td>9187</td></tr>
<tr><td>redis_exporter</td><td>Redis: memory, keys, commands</td><td>9121</td></tr>
<tr><td>blackbox_exporter</td><td>HTTP/TCP/DNS probes (внешний мониторинг)</td><td>9115</td></tr>
<tr><td>cadvisor</td><td>Docker контейнеры: CPU, RAM, network</td><td>8080</td></tr>
</table>

<h2>Типы метрик в Prometheus</h2>
<ol>
<li><strong>Counter</strong> — только растёт (total requests, errors). Нельзя уменьшить!</li>
<li><strong>Gauge</strong> — произвольное значение (температура, goroutines, queue size)</li>
<li><strong>Histogram</strong> — распределение значений по bucket-ам (latency)</li>
<li><strong>Summary</strong> — квантили на стороне клиента (как histogram, но вычисляется в приложении)</li>
</ol>

<h2>Histogram vs Summary</h2>
<table>
<tr><th>Свойство</th><th>Histogram</th><th>Summary</th></tr>
<tr><td>Квантили</td><td>Сервер (PromQL)</td><td>Клиент (приложение)</td></tr>
<tr><td>Агрегация</td><td>Можно суммировать bucket-ы</td><td>Нельзя агрегировать quantiles!</td></tr>
<tr><td>Точность</td><td>Зависит от bucket-ов</td><td>Фиксированная ошибка</td></tr>
<tr><td>Рекомендация</td><td><strong>Используй histogram</strong></td><td>Только если нужны точные квантили</td></tr>
</table>

<h2>Инструментирование Go-сервиса (prometheus/client_golang)</h2>
<pre><code>import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Counter — считаем запросы
var httpRequestsTotal = promauto.NewCounterVec(
    prometheus.CounterOpts{
        Namespace: "myapp",
        Subsystem: "http",
        Name:      "requests_total",
        Help:      "Total number of HTTP requests.",
    },
    []string{"method", "status", "handler"},
)

// Histogram — измеряем латенси
var httpRequestDuration = promauto.NewHistogramVec(
    prometheus.HistogramOpts{
        Namespace: "myapp",
        Subsystem: "http",
        Name:      "request_duration_seconds",
        Help:      "HTTP request duration in seconds.",
        Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "handler"},
)

// Gauge — текущее значение
var activeConnections = promauto.NewGauge(
    prometheus.GaugeOpts{
        Namespace: "myapp",
        Name:      "active_connections",
        Help:      "Number of active connections.",
    },
)

// Middleware для автоматического сбора
func metricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        wrapped := &responseWriter{ResponseWriter: w, status: 200}

        next.ServeHTTP(wrapped, r)

        duration := time.Since(start).Seconds()
        httpRequestsTotal.WithLabelValues(r.Method, strconv.Itoa(wrapped.status), r.URL.Path).Inc()
        httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

// Endpoint /metrics
mux.Handle("/metrics", promhttp.Handler())</code></pre>

<h2>Naming Conventions</h2>
<pre><code># Формат: namespace_subsystem_name_unit
# Пример: myapp_http_request_duration_seconds

# Правила:
# 1. snake_case
# 2. suffix _total для counters
# 3. suffix с единицей: _seconds, _bytes, _total
# 4. base units (seconds, not milliseconds; bytes, not megabytes)
# 5. ПОМОЩЬ: Help должен быть полезным!

# ХОРОШО:
http_request_duration_seconds
process_resident_memory_bytes
http_requests_total

# ПЛОХО:
requestLatency          # camelCase, нет единицы
http_request_ms         # milliseconds вместо seconds
requests                # непонятно: counter? gauge? чей?</code></pre>

<h2>Формат /metrics endpoint</h2>
<pre><code># HELP myapp_http_requests_total Total number of HTTP requests.
# TYPE myapp_http_requests_total counter
myapp_http_requests_total{method="GET",status="200",handler="/api/users"} 1542
myapp_http_requests_total{method="POST",status="201",handler="/api/users"} 89

# HELP myapp_http_request_duration_seconds HTTP request duration in seconds.
# TYPE myapp_http_request_duration_seconds histogram
myapp_http_request_duration_seconds_bucket{method="GET",handler="/api/users",le="0.005"} 100
myapp_http_request_duration_seconds_bucket{method="GET",handler="/api/users",le="0.01"} 250
myapp_http_request_duration_seconds_bucket{method="GET",handler="/api/users",le="+Inf"} 1542
myapp_http_request_duration_seconds_sum{method="GET",handler="/api/users"} 45.23
myapp_http_request_duration_seconds_count{method="GET",handler="/api/users"} 1542</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему histogram предпочтительнее summary?",
				Options:     []string{"Histogram точнее", "Bucket-ы histogram можно агрегировать (суммировать), а quantiles summary — нельзя", "Summary занимает больше памяти", "Нет разницы"},
				Correct:     1,
				Explanation: "Главное преимущество histogram: bucket-ы можно суммировать по instances (sum by (le)). Summary quantiles нельзя агрегировать математически — p99 от p99 это не p99.",
			},
			{
				Question:    "Какой тип метрики использовать для количества горутин?",
				Options:     []string{"Counter", "Gauge", "Histogram", "Summary"},
				Correct:     1,
				Explanation: "Gauge — для значений, которые могут расти и уменьшаться: горутины, температура, память, размер очереди. Counter только растёт.",
			},
			{
				Question:    "Почему метрики измеряются в секундах, а не миллисекундах?",
				Options:     []string{"Миллисекунды неточные", "Prometheus convention: base units (seconds, bytes) — единообразие и предсказуемость", "Секунды экономят место", "Так быстрее"},
				Correct:     1,
				Explanation: "Prometheus convention: всегда base units (seconds, bytes, не ms/KB). Это обеспечивает предсказуемость — все duration в секундах, не нужно гадать.",
			},
		},
		Tasks: []T{
			{
				Title:      "Генератор /metrics output",
				Difficulty: "medium",
				Description: `<p>Сгенерируй вывод в формате Prometheus exposition format.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: <code>metric_name type help_text</code> (type: counter, gauge, histogram)</li>
<li>Остальные строки: <code>label_key=label_value value</code> — значения метрики с labels.</li>
</ol>
<p><strong>Выход:</strong> корректный Prometheus exposition format:</p>
<pre><code># HELP metric_name help_text
# TYPE metric_name type
metric_name{label_key="label_value"} value</code></pre>`,
				Hints: `<p>Первую строку разделяй по пробелам (Fields): [0]=name, [1]=type, [2:]=help. Для labels: разделяй по "=" → key и value, оформляй как {key="value"}.</p>`,
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
	fields := strings.Fields(scanner.Text())
	name := fields[0]
	mtype := fields[1]
	help := strings.Join(fields[2:], " ")

	fmt.Printf("# HELP %s %s\n", name, help)
	fmt.Printf("# TYPE %s %s\n", name, mtype)

	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			continue
		}
		labelPart := parts[0]
		value := parts[1]

		kv := strings.SplitN(labelPart, "=", 2)
		fmt.Printf("%s{%s=\"%s\"} %s\n", name, kv[0], kv[1], value)
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	// Читаем заголовок: name type help_text
	scanner.Scan()
	fields := strings.Fields(scanner.Text())
	name := fields[0]
	mtype := fields[1]
	help := strings.Join(fields[2:], " ")

	// TODO: вывести # HELP name help
	// TODO: вывести # TYPE name type
	_ = name
	_ = mtype
	_ = help

	// Читаем строки с данными: label_key=label_value value
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) != 2 {
			continue
		}
		// TODO: разделить parts[0] по "=" → key, value label
		// Вывести: metric_name{key="value"} metric_value
		_ = parts
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "Exposition format", Definition: "Текстовый формат метрик Prometheus: # HELP, # TYPE, затем строки name{labels} value."},
					{Term: "strings.Join(slice, sep)", Definition: "Объединяет элементы slice в одну строку через sep. Join([\"a\",\"b\"], \" \") → \"a b\""},
					{Term: "strings.Fields(s)", Definition: "Разделяет строку по whitespace в slice строк."},
				},
				TestCases: []TestCase{
					{Input: "http_requests_total counter Total HTTP requests\nmethod=GET 1542\nmethod=POST 89", ExpectedOutput: "# HELP http_requests_total Total HTTP requests\n# TYPE http_requests_total counter\nhttp_requests_total{method=\"GET\"} 1542\nhttp_requests_total{method=\"POST\"} 89"},
					{Input: "go_goroutines gauge Current number of goroutines\ninstance=app1 142\ninstance=app2 89", ExpectedOutput: "# HELP go_goroutines Current number of goroutines\n# TYPE go_goroutines gauge\ngo_goroutines{instance=\"app1\"} 142\ngo_goroutines{instance=\"app2\"} 89"},
				},
			},
			{
				Title:      "Реализация Counter в Go",
				Difficulty: "hard",
				Description: `<p>Реализуй упрощённый потокобезопасный Counter с labels.</p>
<p><strong>Вход:</strong> команды, по одной на строку:</p>
<ul>
<li><code>inc label_value</code> — увеличить counter для данного label на 1</li>
<li><code>add label_value N</code> — увеличить counter для данного label на N</li>
<li><code>get label_value</code> — вывести текущее значение для label</li>
<li><code>dump</code> — вывести все labels и значения в формате <code>metric{label="value"} count</code> (отсортировано по label)</li>
</ul>
<p>Имя метрики для dump: <code>requests_total</code>.</p>`,
				Hints: `<p>Используй map[string]int для хранения значений по label. Для сортированного вывода: собери ключи в slice, отсортируй через sort.Strings, выводи по порядку.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	counters := make(map[string]int)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "inc":
			counters[fields[1]]++
		case "add":
			var n int
			fmt.Sscan(fields[2], &n)
			counters[fields[1]] += n
		case "get":
			fmt.Println(counters[fields[1]])
		case "dump":
			keys := make([]string, 0, len(counters))
			for k := range counters {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				fmt.Printf("requests_total{label=\"%s\"} %d\n", k, counters[k])
			}
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	counters := make(map[string]int)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "inc":
			// TODO: увеличить counters[fields[1]] на 1
		case "add":
			// TODO: увеличить counters[fields[1]] на число из fields[2]
		case "get":
			// TODO: вывести значение counters[fields[1]]
		case "dump":
			// TODO: отсортировать ключи, вывести: requests_total{label="key"} value
			_ = sort.Strings
		}
	}
	_ = counters
	_ = fmt.Println
	_ = strings.Fields
}`,
				Glossary: []GlossaryItem{
					{Term: "map[string]int", Definition: "Словарь с ключом string и значением int. counters[\"GET\"]++ увеличивает значение."},
					{Term: "sort.Strings(slice)", Definition: "Сортирует slice строк на месте (in-place) в алфавитном порядке."},
					{Term: "fmt.Sscan(s, &n)", Definition: "Парсит строку s и записывает значение в переменную n."},
				},
				TestCases: []TestCase{
					{Input: "inc GET\ninc GET\ninc POST\nadd GET 5\nget GET\nget POST\ndump", ExpectedOutput: "7\n1\nrequests_total{label=\"GET\"} 7\nrequests_total{label=\"POST\"} 1"},
					{Input: "add api 100\nadd web 50\nadd api 25\ndump", ExpectedOutput: "requests_total{label=\"api\"} 125\nrequests_total{label=\"web\"} 50"},
				},
			},
		},
	}
}

// ── Урок 5: Grafana Dashboards ───────────────────────────────────

func lesson_grafana_dashboards() L {
	return L{
		Slug: "grafana-dashboards", Title: "Grafana: дашборды и визуализация", Order: 5,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Grafana: дашборды и визуализация</h1>

<h2>Что такое Grafana?</h2>
<p>Grafana — платформа для визуализации метрик. Подключается к Prometheus (и другим sources) и показывает красивые дашборды с графиками, таблицами, алертами.</p>

<h2>Data Sources</h2>
<p>Grafana не хранит данные — она запрашивает их из backend:</p>
<ul>
<li><strong>Prometheus</strong> — основной для метрик (PromQL)</li>
<li><strong>Loki</strong> — для логов (LogQL)</li>
<li><strong>Tempo</strong> — для трейсов</li>
<li><strong>PostgreSQL</strong> — для бизнес-метрик из БД</li>
<li><strong>InfluxDB, Elasticsearch, CloudWatch...</strong></li>
</ul>

<h2>Панели (Panels)</h2>
<pre><code>┌─────────────────────────────────────────────────────────────┐
│  Dashboard: "Service Overview"                               │
├─────────────────┬───────────────────┬───────────────────────┤
│  Time series    │  Stat panel       │  Table               │
│  (RPS graph)    │  (Current p99)    │  (Top endpoints)     │
│  ~~~~~~~~~~~~   │     ┌───┐         │  /api/users  12ms    │
│  ~~~~/\~~~~     │     │250│ ms      │  /api/orders 45ms    │
│  ~~~/  \~~~~    │     └───┘         │  /healthz    1ms     │
├─────────────────┼───────────────────┤───────────────────────┤
│  Gauge          │  Heatmap          │  Alert list          │
│  (Error rate)   │  (Latency dist)   │  High error rate     │
│    [||||  ]     │                   │  Memory > 80%        │
│    0.5%         │                   │  CPU normal          │
└─────────────────┴───────────────────┴───────────────────────┘</code></pre>

<h2>Типы панелей</h2>
<ul>
<li><strong>Time series</strong> — основной график (линии, бары, точки)</li>
<li><strong>Stat</strong> — одно большое число (текущее значение)</li>
<li><strong>Gauge</strong> — шкала с пороговыми значениями</li>
<li><strong>Table</strong> — табличные данные</li>
<li><strong>Heatmap</strong> — тепловая карта (идеально для histogram)</li>
<li><strong>Logs</strong> — потоковый вывод логов (Loki)</li>
<li><strong>Bar chart, Pie chart, Node graph...</strong></li>
</ul>

<h2>Variables (Template Variables)</h2>
<p>Позволяют делать дашборды динамическими — один дашборд для всех сервисов:</p>
<pre><code># Variable: $instance
# Query: label_values(up{job="my-app"}, instance)
# Result: dropdown с app1:8080, app2:8080...

# Использование в панели:
rate(http_requests_total{instance="$instance"}[5m])

# Multi-value variable (regex)
rate(http_requests_total{instance=~"$instance"}[5m])</code></pre>

<h2>Annotations</h2>
<p>Вертикальные маркеры на графике — отмечают события (деплой, инцидент):</p>
<pre><code># Annotation query (Prometheus):
changes(deployment_info{app="myapp"}[1m]) > 0

# Или через Grafana API:
POST /api/annotations
{
  "time": 1704067200000,
  "text": "Deploy v2.3.1",
  "tags": ["deploy", "myapp"]
}</code></pre>

<h2>Dashboard as Code (Provisioning)</h2>
<p>Дашборды в Git, автоматический деплой:</p>
<pre><code># grafana/provisioning/dashboards/provider.yml
apiVersion: 1
providers:
  - name: 'default'
    folder: 'Provisioned'
    type: file
    options:
      path: /var/lib/grafana/dashboards
      foldersFromFilesStructure: true

# Дашборд как JSON:
# grafana/dashboards/service-overview.json
{
  "title": "Service Overview",
  "panels": [
    {
      "type": "timeseries",
      "title": "RPS",
      "targets": [{
        "expr": "sum(rate(http_requests_total{job=\"$job\"}[5m]))"
      }]
    }
  ]
}</code></pre>

<h2>Grafana + docker-compose</h2>
<pre><code># docker-compose.yml
services:
  grafana:
    image: grafana/grafana:10.2.0
    ports:
      - "3000:3000"
    volumes:
      - ./grafana/provisioning:/etc/grafana/provisioning
      - ./grafana/dashboards:/var/lib/grafana/dashboards
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=secret
      - GF_AUTH_ANONYMOUS_ENABLED=true

  prometheus:
    image: prom/prometheus:v2.48.0
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Зачем нужны Template Variables в Grafana?",
				Options:     []string{"Для красоты", "Один дашборд работает для всех сервисов/instances — выбираешь из dropdown", "Для безопасности", "Для скорости"},
				Correct:     1,
				Explanation: "Template variables делают дашборд универсальным: вместо хардкода instance=\"app1\" используешь $instance — пользователь выбирает из списка. Один дашборд вместо 50.",
			},
			{
				Question:    "Как хранить дашборды Grafana в Git (Dashboard as Code)?",
				Options:     []string{"Скриншоты дашбордов", "JSON-экспорт дашбордов + provisioning config — Grafana загружает с диска при старте", "Копировать базу Grafana", "Через API при каждом запуске"},
				Correct:     1,
				Explanation: "Dashboard as Code: экспортируешь JSON дашборда, кладёшь в Git, настраиваешь provisioning provider — Grafana автоматически загружает дашборды из папки. Версионирование, code review, rollback.",
			},
			{
				Question:    "Какой тип панели лучше всего подходит для визуализации распределения латенси (histogram)?",
				Options:     []string{"Stat panel", "Heatmap", "Table", "Pie chart"},
				Correct:     1,
				Explanation: "Heatmap идеально показывает распределение значений по времени — цветом кодируется частота попадания в bucket. Для histogram-метрик это самая информативная визуализация.",
			},
		},
		Tasks: []T{
			{
				Title:      "Генератор Grafana JSON panel",
				Difficulty: "medium",
				Description: `<p>Сгенерируй упрощённый JSON для Grafana panel.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: <code>panel_title</code></li>
<li>Вторая строка: <code>panel_type</code> (timeseries, stat, gauge)</li>
<li>Остальные строки: PromQL выражения (targets)</li>
</ol>
<p><strong>Выход:</strong> JSON (без лишних пробелов в значениях):</p>
<pre><code>{"title":"panel_title","type":"panel_type","targets":[{"expr":"query1"},{"expr":"query2"}]}</code></pre>`,
				Hints: `<p>Используй fmt.Sprintf для формирования JSON строки. Собери targets в slice строк, затем объедини через strings.Join.</p>`,
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
	title := scanner.Text()
	scanner.Scan()
	panelType := scanner.Text()

	var targets []string
	for scanner.Scan() {
		expr := scanner.Text()
		targets = append(targets, fmt.Sprintf("{\"expr\":\"%s\"}", expr))
	}

	targetsJSON := "[" + strings.Join(targets, ",") + "]"
	fmt.Printf("{\"title\":\"%s\",\"type\":\"%s\",\"targets\":%s}\n", title, panelType, targetsJSON)
}</code></pre>`,
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
	title := scanner.Text()
	scanner.Scan()
	panelType := scanner.Text()

	var targets []string
	for scanner.Scan() {
		expr := scanner.Text()
		// TODO: добавить в targets строку формата {"expr":"..."}
		_ = expr
	}

	// TODO: сформировать JSON:
	// {"title":"...","type":"...","targets":[...]}
	// targets объединить через strings.Join(targets, ",")
	_ = title
	_ = panelType
	_ = targets
	_ = strings.Join
	_ = fmt.Sprintf
}`,
				Glossary: []GlossaryItem{
					{Term: "strings.Join(slice, sep)", Definition: "Объединяет slice строк в одну строку, вставляя sep между элементами."},
					{Term: "fmt.Sprintf", Definition: "Форматирует строку (как Printf), но возвращает результат вместо печати."},
					{Term: "JSON в Grafana", Definition: "Дашборды Grafana хранятся как JSON — панели, запросы, переменные, всё описано в одном файле."},
				},
				TestCases: []TestCase{
					{Input: "Request Rate\ntimeseries\nsum(rate(http_requests_total[5m]))", ExpectedOutput: `{"title":"Request Rate","type":"timeseries","targets":[{"expr":"sum(rate(http_requests_total[5m]))"}]}`},
					{Input: "Error Rate\nstat\nsum(rate(http_requests_total{status=~\"5..\"}[5m]))\nsum(rate(http_requests_total[5m]))", ExpectedOutput: `{"title":"Error Rate","type":"stat","targets":[{"expr":"sum(rate(http_requests_total{status=~\"5..\"}[5m]))"},{"expr":"sum(rate(http_requests_total[5m]))"}]}`},
				},
			},
			{
				Title:      "Парсер Grafana variable query",
				Difficulty: "easy",
				Description: `<p>Grafana template variables используют функцию <code>label_values(metric, label)</code> для получения списка значений label.</p>
<p><strong>Вход:</strong> строки формата <code>label_values(metric_name, label_name)</code></p>
<p><strong>Выход:</strong> для каждой строки выведи: <code>metric: metric_name, label: label_name</code></p>`,
				Hints: `<p>Убери "label_values(" в начале и ")" в конце. Разделяй оставшуюся часть по ", ".</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		inner := strings.TrimPrefix(line, "label_values(")
		inner = strings.TrimSuffix(inner, ")")
		parts := strings.SplitN(inner, ", ", 2)
		fmt.Printf("metric: %s, label: %s\n", parts[0], parts[1])
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		// TODO: убери "label_values(" и ")" с краёв
		// Разделяй по ", " чтобы получить metric и label
		// Выведи: metric: X, label: Y
		_ = line
		_ = strings.TrimPrefix
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "strings.TrimPrefix(s, prefix)", Definition: "Удаляет prefix из начала строки если он есть."},
					{Term: "strings.TrimSuffix(s, suffix)", Definition: "Удаляет suffix из конца строки если он есть."},
					{Term: "label_values(metric, label)", Definition: "Функция Grafana для получения всех уникальных значений label из указанной метрики."},
				},
				TestCases: []TestCase{
					{Input: "label_values(up, instance)\nlabel_values(http_requests_total, method)", ExpectedOutput: "metric: up, label: instance\nmetric: http_requests_total, label: method"},
					{Input: "label_values(node_cpu_seconds_total, cpu)", ExpectedOutput: "metric: node_cpu_seconds_total, label: cpu"},
				},
			},
		},
	}
}

// ── Урок 6: Alerting Pipeline ────────────────────────────────────

func lesson_alerting_pipeline() L {
	return L{
		Slug: "alerting-pipeline", Title: "Alerting: от правил до уведомлений", Order: 6,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Alerting Pipeline</h1>

<h2>Архитектура алертинга</h2>
<pre><code>┌──────────────┐    fire     ┌──────────────────┐   route   ┌──────────────┐
│  Prometheus  │ ──────────→ │  Alertmanager    │ ────────→ │  Receivers   │
│  (rules)     │             │  (routing,       │           │  - Slack     │
│              │             │   grouping,      │           │  - PagerDuty │
│  alerting    │   resolve   │   silences,      │           │  - Email     │
│  rules.yml   │ ──────────→ │   inhibitions)   │           │  - Telegram  │
└──────────────┘             └──────────────────┘           └──────────────┘</code></pre>

<h2>Prometheus Alerting Rules</h2>
<pre><code># alert_rules.yml
groups:
  - name: service_alerts
    rules:
      # Высокий error rate
      - alert: HighErrorRate
        expr: |
          sum(rate(http_requests_total{status=~"5.."}[5m]))
          /
          sum(rate(http_requests_total[5m]))
          > 0.01
        for: 5m    # должно быть true 5 минут подряд!
        labels:
          severity: critical
          team: backend
        annotations:
          summary: "High error rate: {{ $value | humanizePercentage }}"
          description: "Error rate above 1% for 5 minutes"
          runbook_url: "https://wiki/runbooks/high-error-rate"

      # Сервис недоступен
      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "{{ $labels.instance }} is down"

      # Диск заполняется
      - alert: DiskSpaceLow
        expr: |
          (node_filesystem_avail_bytes / node_filesystem_size_bytes) < 0.1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "Disk space below 10% on {{ $labels.instance }}"</code></pre>

<h2>Состояния алерта</h2>
<pre><code>Inactive  →  Pending  →  Firing  →  Resolved
                ↑           │
                └───────────┘  (if condition becomes false before 'for' duration)

- Inactive: условие false
- Pending: условие true, но ещё не прошло 'for' (ожидание)
- Firing: условие true дольше 'for' → отправлен в Alertmanager
- Resolved: было Firing, стало false → уведомление о восстановлении</code></pre>

<h2>Alertmanager: Routing</h2>
<pre><code># alertmanager.yml
route:
  receiver: 'default-slack'
  group_by: ['alertname', 'job']
  group_wait: 30s       # ждём 30s, собирая алерты в группу
  group_interval: 5m    # между повторными уведомлениями группы
  repeat_interval: 4h   # повторное уведомление если не resolved

  routes:
    # Critical → PagerDuty (немедленно)
    - match:
        severity: critical
      receiver: 'pagerduty'
      group_wait: 10s
      continue: false    # stop routing

    # Warning от team=db → Slack канал #db-alerts
    - match:
        severity: warning
        team: db
      receiver: 'db-slack'

receivers:
  - name: 'default-slack'
    slack_configs:
      - channel: '#alerts'
        send_resolved: true

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'xxx'

  - name: 'db-slack'
    slack_configs:
      - channel: '#db-alerts'</code></pre>

<h2>Silences (Заглушки)</h2>
<p>Временное подавление алертов (плановые работы, известная проблема):</p>
<pre><code># CLI
amtool silence add alertname=HighErrorRate --duration=2h \
  --comment="Deploying new version, expected errors"

# Или через UI Alertmanager: http://alertmanager:9093/#/silences</code></pre>

<h2>Inhibitions (Подавление каскадов)</h2>
<p>Если сервис полностью down — не слать алерты о его метриках:</p>
<pre><code>inhibit_rules:
  # Если ServiceDown → подавить все другие алерты этого instance
  - source_match:
      alertname: ServiceDown
    target_match_re:
      alertname: '.+'
    equal: ['instance']  # совпадение по instance label</code></pre>

<h2>Лучшие практики</h2>
<ul>
<li><strong>for: 5m</strong> — всегда ставь 'for' чтобы избежать flapping (мигающих алертов)</li>
<li><strong>runbook_url</strong> — каждый алерт должен ссылаться на инструкцию "что делать"</li>
<li><strong>Labels</strong>: severity (critical/warning/info), team, service — для routing</li>
<li><strong>Не алертить на симптомы ресурсов</strong> — алертить на user-facing impact (SLO-based)</li>
<li><strong>group_by</strong> — группируй, чтобы не получать 100 отдельных уведомлений</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Зачем в alerting rule ставят for: 5m?",
				Options:     []string{"Для задержки уведомления", "Чтобы условие должно быть true 5 минут подряд, избегая ложных срабатываний при кратковременных всплесках", "Для повторения алерта", "Это таймаут"},
				Correct:     1,
				Explanation: "'for' — это grace period. Если условие true менее 5m (кратковременный spike) → алерт остаётся в Pending и не fires. Защита от flapping и noise.",
			},
			{
				Question:    "Что делают Inhibition Rules в Alertmanager?",
				Options:     []string{"Удаляют алерты навсегда", "Подавляют каскадные алерты: если source alert firing, target alerts не отправляются", "Создают новые алерты", "Повторяют алерты"},
				Correct:     1,
				Explanation: "Inhibitions предотвращают каскад: если сервер down (source), не нужно слать отдельные алерты о CPU, RAM, disk этого сервера (targets). equal: ['instance'] — подавление по совпадающему label.",
			},
			{
				Question:    "Разница между Silence и Inhibition?",
				Options:     []string{"Нет разницы", "Silence — ручное временное подавление (плановые работы), Inhibition — автоматическое подавление каскадов", "Silence permanent, Inhibition temporary", "Silence для warning, Inhibition для critical"},
				Correct:     1,
				Explanation: "Silence — человек вручную глушит конкретные алерты на время (деплой, maintenance). Inhibition — автоматическое правило: если X firing, то не слать Y (предотвращение alert storm).",
			},
			{
				Question:    "Почему лучше алертить на SLO (user-facing impact), а не на CPU > 90%?",
				Options:     []string{"CPU не важен", "CPU > 90% не значит проблему для пользователя; SLO-based алерт срабатывает когда есть реальное влияние на сервис", "SLO проще настроить", "CPU нельзя мониторить"},
				Correct:     1,
				Explanation: "CPU 90% может быть нормой для batch-воркера. SLO-based алерт (error_rate > budget) означает РЕАЛЬНУЮ деградацию для пользователей. Меньше шума, больше actionable alerts.",
			},
		},
		Tasks: []T{
			{
				Title:      "Генератор alerting rules из спецификации",
				Difficulty: "medium",
				Description: `<p>По спецификации сгенерируй Prometheus alerting rule в YAML-подобном формате.</p>
<p><strong>Вход:</strong> строки формата <code>alert_name metric operator threshold for_duration severity</code></p>
<p><strong>Выход:</strong> для каждой строки сгенерируй:</p>
<pre><code>- alert: AlertName
  expr: metric operator threshold
  for: for_duration
  labels:
    severity: severity</code></pre>`,
				Hints: `<p>Разделяй каждую строку по пробелам (Fields). Элементы по индексам: [0]=name, [1]=metric, [2]=operator, [3]=threshold, [4]=for, [5]=severity.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 6 {
			continue
		}
		name := fields[0]
		metric := fields[1]
		op := fields[2]
		threshold := fields[3]
		forDur := fields[4]
		severity := fields[5]

		fmt.Printf("- alert: %s\n", name)
		fmt.Printf("  expr: %s %s %s\n", metric, op, threshold)
		fmt.Printf("  for: %s\n", forDur)
		fmt.Printf("  labels:\n")
		fmt.Printf("    severity: %s\n", severity)
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 6 {
			continue
		}
		// fields: [name, metric, operator, threshold, for_duration, severity]
		// TODO: вывести YAML-формат alerting rule
		// - alert: Name
		//   expr: metric op threshold
		//   for: duration
		//   labels:
		//     severity: sev
		_ = fields
		_ = fmt.Printf
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "Alerting rule", Definition: "Правило в Prometheus: если PromQL-выражение true в течение 'for' — генерируется alert."},
					{Term: "'for' duration", Definition: "Сколько времени условие должно быть true перед переходом в Firing. Защита от кратковременных всплесков."},
					{Term: "severity label", Definition: "Уровень критичности: critical (PagerDuty, немедленно), warning (Slack), info (для дашбордов)."},
				},
				TestCases: []TestCase{
					{Input: "HighErrorRate error_rate > 0.01 5m critical\nHighLatency p99_latency > 1.0 10m warning", ExpectedOutput: "- alert: HighErrorRate\n  expr: error_rate > 0.01\n  for: 5m\n  labels:\n    severity: critical\n- alert: HighLatency\n  expr: p99_latency > 1.0\n  for: 10m\n  labels:\n    severity: warning"},
					{Input: "ServiceDown up == 0 1m critical", ExpectedOutput: "- alert: ServiceDown\n  expr: up == 0\n  for: 1m\n  labels:\n    severity: critical"},
				},
			},
			{
				Title:      "Симулятор alert routing",
				Difficulty: "hard",
				Description: `<p>Реализуй упрощённый routing алертов как в Alertmanager.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: число N (количество routing rules)</li>
<li>Следующие N строк: правила формата <code>label_name=label_value receiver</code></li>
<li>Далее строка <code>---</code> (разделитель)</li>
<li>Остальные строки: алерты формата <code>alertname label1=val1 label2=val2 ...</code></li>
</ol>
<p><strong>Логика:</strong> для каждого алерта проверь rules сверху вниз. Первое совпавшее правило определяет receiver. Если ничего не совпало — receiver = "default".</p>
<p><strong>Выход:</strong> <code>alertname → receiver</code></p>`,
				Hints: `<p>Сохрани rules как slice struct{label, value, receiver}. Для каждого алерта разбирай labels в map, проверяй rules по порядку.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type rule struct {
	label, value, receiver string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	rules := make([]rule, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		fields := strings.Fields(scanner.Text())
		kv := strings.SplitN(fields[0], "=", 2)
		rules[i] = rule{label: kv[0], value: kv[1], receiver: fields[1]}
	}

	scanner.Scan() // skip "---"

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		alertName := fields[0]
		labels := make(map[string]string)
		for _, f := range fields[1:] {
			kv := strings.SplitN(f, "=", 2)
			if len(kv) == 2 {
				labels[kv[0]] = kv[1]
			}
		}

		receiver := "default"
		for _, r := range rules {
			if labels[r.label] == r.value {
				receiver = r.receiver
				break
			}
		}
		fmt.Printf("%s → %s\n", alertName, receiver)
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type rule struct {
	label, value, receiver string
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	var n int
	fmt.Sscan(scanner.Text(), &n)

	rules := make([]rule, n)
	for i := 0; i < n; i++ {
		scanner.Scan()
		fields := strings.Fields(scanner.Text())
		kv := strings.SplitN(fields[0], "=", 2)
		rules[i] = rule{label: kv[0], value: kv[1], receiver: fields[1]}
	}

	scanner.Scan() // skip "---"

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		alertName := fields[0]
		// TODO: разбери fields[1:] в map[string]string labels
		// Проверь rules по порядку: если labels[r.label] == r.value → receiver = r.receiver
		// Если ничего не совпало → receiver = "default"
		// Выведи: alertName → receiver
		_ = alertName
		_ = rules
	}
}`,
				Glossary: []GlossaryItem{
					{Term: "Routing rules", Definition: "Alertmanager проверяет labels алерта сверху вниз по правилам. Первое совпавшее определяет receiver (канал уведомлений)."},
					{Term: "map[string]string", Definition: "Словарь строка→строка. Удобен для хранения labels: labels[\"severity\"] = \"critical\"."},
					{Term: "struct{}", Definition: "Тип-структура. type rule struct { label, value, receiver string } — группирует связанные поля."},
				},
				TestCases: []TestCase{
					{Input: "2\nseverity=critical pagerduty\nteam=db db-slack\n---\nHighErrorRate severity=critical team=backend\nDiskLow severity=warning team=db\nMemoryHigh severity=info team=frontend", ExpectedOutput: "HighErrorRate → pagerduty\nDiskLow → db-slack\nMemoryHigh → default"},
					{Input: "1\nseverity=critical oncall\n---\nServiceDown severity=critical\nHighLatency severity=warning", ExpectedOutput: "ServiceDown → oncall\nHighLatency → default"},
				},
			},
		},
	}
}

// ── Урок 7: Production Monitoring ────────────────────────────────

func lesson_production_monitoring() L {
	return L{
		Slug: "production-monitoring", Title: "Мониторинг Go-сервиса в проде", Order: 7,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Production Monitoring: Go-сервис end-to-end</h1>

<h2>Полный стек мониторинга</h2>
<pre><code>┌─────────────────────────────────────────────────────────────────┐
│                        YOUR GO SERVICE                           │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Instrumentation (prometheus/client_golang)               │   │
│  │  - http_requests_total (Counter)                          │   │
│  │  - http_request_duration_seconds (Histogram)              │   │
│  │  - active_connections (Gauge)                             │   │
│  │  - business_events_total (Counter)                        │   │
│  └──────────────────────────────────────────────────────────┘   │
│                            │ :8080/metrics                        │
└────────────────────────────┼─────────────────────────────────────┘
                             ↓ scrape every 15s
┌──────────────────┐   ┌──────────────┐   ┌───────────────────┐
│   Prometheus     │──→│  Alertmanager │──→│  Slack/PagerDuty  │
│   + Rules        │   └──────────────┘   └───────────────────┘
└────────┬─────────┘
         │ data source
         ↓
┌──────────────────┐   ┌───────────────────┐
│     Grafana      │   │   Grafana Loki    │ ← logs via promtail
│   dashboards     │   │   (LogQL)         │
└──────────────────┘   └───────────────────┘</code></pre>

<h2>Метрики Go runtime</h2>
<p>prometheus/client_golang автоматически экспортирует Go runtime метрики:</p>
<pre><code># Горутины (утечка горутин — частая проблема)
go_goroutines                          # текущее количество
go_threads                             # OS threads

# Память
go_memstats_alloc_bytes                # allocated heap (сейчас)
go_memstats_sys_bytes                  # total memory from OS
go_memstats_heap_inuse_bytes           # heap in use

# GC (Garbage Collector)
go_gc_duration_seconds                 # GC pause duration
go_memstats_gc_sys_bytes               # GC overhead

# Process
process_resident_memory_bytes          # RSS (реальное потребление)
process_cpu_seconds_total              # CPU usage
process_open_fds                       # open file descriptors</code></pre>

<h2>SLO-based Alerting</h2>
<p>Вместо "CPU > 80%" алертим на "error budget burning too fast":</p>
<pre><code># Многоокорный алерт (Multi Burn Rate)
# Идея: алертим если error budget сгорает слишком быстро

# 1-hour burn rate: if this continues, budget exhausted in 1 day
- alert: ErrorBudgetBurn_1h
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[1h]))
      /
      sum(rate(http_requests_total[1h]))
    ) > (14.4 * 0.001)   # 14.4x burn rate
  for: 2m
  labels:
    severity: critical
    window: 1h

# 6-hour burn rate: slower burn, still concerning
- alert: ErrorBudgetBurn_6h
  expr: |
    (
      sum(rate(http_requests_total{status=~"5.."}[6h]))
      /
      sum(rate(http_requests_total[6h]))
    ) > (6 * 0.001)      # 6x burn rate
  for: 5m
  labels:
    severity: warning
    window: 6h</code></pre>

<h2>Capacity Planning</h2>
<p>Используй метрики для прогнозирования роста:</p>
<pre><code># Линейная регрессия: когда диск заполнится?
predict_linear(node_filesystem_avail_bytes[24h], 7*24*3600) < 0
# "Если тренд последних 24h сохранится, диск кончится менее чем за 7 дней"

# Аналогично для памяти:
predict_linear(process_resident_memory_bytes[1h], 4*3600) > 1e9
# "При текущем росте через 4 часа RSS превысит 1GB"

# Traffic growth (year over year):
sum(rate(http_requests_total[7d]))
/
sum(rate(http_requests_total[7d] offset 365d))
# Коэффициент роста трафика год к году</code></pre>

<h2>Grafana Loki — логи рядом с метриками</h2>
<pre><code># Loki — как Prometheus, но для логов
# Индексирует labels, а тело лога — не индексирует (дёшево!)

# docker-compose.yml
services:
  loki:
    image: grafana/loki:2.9.0
    ports: ["3100:3100"]

  promtail:
    image: grafana/promtail:2.9.0
    volumes:
      - /var/log:/var/log
      - ./promtail.yml:/etc/promtail/config.yml

# LogQL запросы в Grafana:
{app="my-service"} |= "error"                    # простой фильтр
{app="my-service"} | json | level="error"         # parse JSON logs
{app="my-service"} | json | latency_ms > 1000     # фильтр по полю

# Метрики из логов!
rate({app="my-service"} |= "error" [5m])          # errors per second
count_over_time({app="my-service"} | json | status >= 500 [1h])  # 5xx за час</code></pre>

<h2>Чеклист мониторинга Go-сервиса в проде</h2>
<ol>
<li>prometheus/client_golang подключён, /metrics endpoint работает</li>
<li>HTTP middleware считает requests_total и duration_seconds</li>
<li>Business metrics: signups, orders, payments</li>
<li>Prometheus scrape config добавлен</li>
<li>Grafana дашборд: Golden Signals + Go runtime</li>
<li>Alerting rules: SLO-based burn rate + критические (ServiceDown)</li>
<li>Loki + promtail для structured logs</li>
<li>Runbook URL в каждом алерте</li>
<li>predict_linear для capacity planning</li>
<li>Dashboards provisioned (в Git)</li>
</ol>`,

		Quiz: []Q{
			{
				Question:    "Что показывает метрика go_goroutines и на что она может указывать?",
				Options:     []string{"Количество CPU", "Текущее число горутин — рост без плато указывает на goroutine leak", "Скорость GC", "Размер стека"},
				Correct:     1,
				Explanation: "go_goroutines показывает текущее количество горутин. Если оно постоянно растёт (без стабилизации) — это goroutine leak: создаются горутины, которые не завершаются (забытый cancel, blocked channel).",
			},
			{
				Question:    "Что делает predict_linear(metric[24h], 7*24*3600)?",
				Options:     []string{"Показывает значение за прошлую неделю", "Прогнозирует значение метрики через 7 дней на основе тренда последних 24 часов", "Усредняет за неделю", "Удаляет данные старше 7 дней"},
				Correct:     1,
				Explanation: "predict_linear делает линейную экстраполяцию: берёт тренд за указанный range [24h] и предсказывает значение через заданное количество секунд (7 дней). Идеально для capacity planning.",
			},
			{
				Question:    "Почему Multi Burn Rate alerting лучше, чем простой error_rate > threshold?",
				Options:     []string{"Проще настроить", "Учитывает скорость сгорания error budget — быстрый burn = critical, медленный = warning, не шумит при кратковременных всплесках", "Работает быстрее", "Использует меньше ресурсов"},
				Correct:     1,
				Explanation: "Multi burn rate алертит пропорционально скорости сгорания бюджета: 14.4x за 1 час = critical (бюджет сгорит за 1 день), 6x за 6 часов = warning. Это SRE best practice из Google SRE Book.",
			},
		},
		Tasks: []T{
			{
				Title:      "Детектор goroutine leak",
				Difficulty: "medium",
				Description: `<p>По временному ряду значений go_goroutines определи, есть ли goroutine leak.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: число N (количество измерений)</li>
<li>Следующие N строк: значение go_goroutines (целое число)</li>
</ol>
<p><strong>Логика:</strong></p>
<ul>
<li>Если значение выросло более чем на 50% от начального И ни разу не уменьшалось — это leak</li>
<li>Уменьшение = текущее значение < предыдущее</li>
</ul>
<p><strong>Выход:</strong> <code>LEAK</code> или <code>NORMAL</code></p>`,
				Hints: `<p>Сохрани первое значение. Пройди по всем значениям: если текущее < предыдущему — значит есть уменьшение (not leak). В конце проверь: last > first*1.5 И не было уменьшений → LEAK.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	values := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&values[i])
	}

	decreased := false
	for i := 1; i < n; i++ {
		if values[i] < values[i-1] {
			decreased = true
			break
		}
	}

	first := values[0]
	last := values[n-1]

	if !decreased && float64(last) > float64(first)*1.5 {
		fmt.Println("LEAK")
	} else {
		fmt.Println("NORMAL")
	}
}</code></pre>`,
				StarterCode: `package main

import "fmt"

func main() {
	var n int
	fmt.Scan(&n)

	values := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&values[i])
	}

	// TODO: проверь, было ли уменьшение (values[i] < values[i-1])
	// Если не было уменьшений И last > first * 1.5 → "LEAK"
	// Иначе → "NORMAL"
	_ = values
}`,
				Glossary: []GlossaryItem{
					{Term: "Goroutine leak", Definition: "Горутины создаются, но не завершаются. go_goroutines растёт бесконечно → OOM."},
					{Term: "float64(n)", Definition: "Преобразование int в float64 для операции умножения на 1.5."},
					{Term: "Monotonic growth", Definition: "Постоянный рост без уменьшений — признак утечки ресурсов."},
				},
				TestCases: []TestCase{
					{Input: "5\n100\n120\n150\n180\n200", ExpectedOutput: "LEAK"},
					{Input: "5\n100\n150\n120\n160\n130", ExpectedOutput: "NORMAL"},
					{Input: "4\n100\n105\n110\n115", ExpectedOutput: "NORMAL"},
					{Input: "6\n50\n60\n70\n80\n90\n100", ExpectedOutput: "LEAK"},
				},
			},
			{
				Title:      "Predict Linear (упрощённый)",
				Difficulty: "hard",
				Description: `<p>Реализуй упрощённую версию PromQL <code>predict_linear</code> — линейная экстраполяция.</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: число N (количество точек) и T (через сколько секунд предсказать)</li>
<li>Следующие N строк: <code>timestamp value</code> (timestamp в секундах)</li>
</ol>
<p><strong>Метод:</strong> Линейная регрессия (метод наименьших квадратов):</p>
<pre><code>slope = (N * sum(x*y) - sum(x) * sum(y)) / (N * sum(x^2) - sum(x)^2)
intercept = (sum(y) - slope * sum(x)) / N
prediction = slope * (last_timestamp + T) + intercept</code></pre>
<p><strong>Выход:</strong> предсказанное значение, округлённое до целого.</p>`,
				Hints: `<p>Вычисли суммы: sumX (timestamps), sumY (values), sumXY (timestamp*value), sumX2 (timestamp^2). Подставь в формулу slope/intercept. Prediction = slope * (lastTS + T) + intercept.</p>`,
				Solution: `<pre><code>package main

import "fmt"

func main() {
	var n, t int
	fmt.Scan(&n, &t)

	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&xs[i], &ys[i])
	}

	var sumX, sumY, sumXY, sumX2 float64
	for i := 0; i < n; i++ {
		sumX += xs[i]
		sumY += ys[i]
		sumXY += xs[i] * ys[i]
		sumX2 += xs[i] * xs[i]
	}

	nf := float64(n)
	slope := (nf*sumXY - sumX*sumY) / (nf*sumX2 - sumX*sumX)
	intercept := (sumY - slope*sumX) / nf

	predictAt := xs[n-1] + float64(t)
	prediction := slope*predictAt + intercept

	fmt.Printf("%d\n", int(prediction+0.5))
}</code></pre>`,
				StarterCode: `package main

import "fmt"

func main() {
	var n, t int
	fmt.Scan(&n, &t)

	xs := make([]float64, n)
	ys := make([]float64, n)
	for i := 0; i < n; i++ {
		fmt.Scan(&xs[i], &ys[i])
	}

	// TODO: вычисли sumX, sumY, sumXY, sumX2
	// slope = (N*sumXY - sumX*sumY) / (N*sumX2 - sumX^2)
	// intercept = (sumY - slope*sumX) / N
	// prediction = slope * (lastTimestamp + T) + intercept
	// Выведи округлённое до целого: int(prediction + 0.5)
	_ = xs
	_ = ys
	_ = t
}`,
				Glossary: []GlossaryItem{
					{Term: "predict_linear", Definition: "PromQL функция: линейная экстраполяция тренда. Прогнозирует будущее значение по текущему тренду."},
					{Term: "Linear regression", Definition: "Метод наименьших квадратов: находит прямую y = slope*x + intercept, лучше всего описывающую точки."},
					{Term: "int(x + 0.5)", Definition: "Округление float к ближайшему целому в Go (math.Round тоже работает)."},
				},
				TestCases: []TestCase{
					{Input: "4 3600\n0 100\n3600 200\n7200 300\n10800 400", ExpectedOutput: "500"},
					{Input: "3 7200\n0 1000000000\n3600 1100000000\n7200 1200000000", ExpectedOutput: "1400000000"},
					{Input: "5 60\n0 50\n15 55\n30 60\n45 65\n60 70", ExpectedOutput: "80"},
				},
			},
			{
				Title:      "Histogram bucket analyzer",
				Difficulty: "hard",
				Description: `<p>По данным histogram bucket-ов вычисли приближённый перцентиль (как histogram_quantile в PromQL).</p>
<p><strong>Вход:</strong></p>
<ol>
<li>Первая строка: quantile (например: 0.99)</li>
<li>Остальные строки: <code>le count</code> — граница bucket и кумулятивный count. Последний bucket всегда <code>+Inf</code>.</li>
</ol>
<p><strong>Алгоритм:</strong></p>
<ol>
<li>total = count последнего bucket (+Inf)</li>
<li>target = quantile * total</li>
<li>Найди первый bucket, где count >= target</li>
<li>Линейная интерполяция: result = lower_bound + (upper_bound - lower_bound) * (target - prev_count) / (curr_count - prev_count)</li>
<li>Для первого bucket: lower_bound = 0</li>
</ol>
<p><strong>Выход:</strong> значение перцентиля с 3 знаками после запятой.</p>`,
				Hints: `<p>Парси le и count в slices. target = quantile * total. Пройди по buckets: когда count[i] >= target, интерполируй между le[i-1] и le[i]. Не забудь skip +Inf при интерполяции.</p>`,
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
	var quantile float64
	fmt.Sscan(scanner.Text(), &quantile)

	var boundaries []float64
	var counts []float64

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		countVal, _ := strconv.ParseFloat(fields[1], 64)
		counts = append(counts, countVal)

		if fields[0] == "+Inf" {
			boundaries = append(boundaries, 0)
		} else {
			le, _ := strconv.ParseFloat(fields[0], 64)
			boundaries = append(boundaries, le)
		}
	}

	total := counts[len(counts)-1]
	target := quantile * total

	for i := 0; i < len(counts)-1; i++ {
		if counts[i] >= target {
			var lowerBound float64
			var prevCount float64
			if i > 0 {
				lowerBound = boundaries[i-1]
				prevCount = counts[i-1]
			}
			upperBound := boundaries[i]
			result := lowerBound + (upperBound-lowerBound)*(target-prevCount)/(counts[i]-prevCount)
			fmt.Printf("%.3f\n", result)
			return
		}
	}
	fmt.Printf("%.3f\n", boundaries[len(boundaries)-2])
}</code></pre>`,
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
	var quantile float64
	fmt.Sscan(scanner.Text(), &quantile)

	var boundaries []float64
	var counts []float64

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) != 2 {
			continue
		}
		countVal, _ := strconv.ParseFloat(fields[1], 64)
		counts = append(counts, countVal)

		if fields[0] == "+Inf" {
			boundaries = append(boundaries, 0)
		} else {
			le, _ := strconv.ParseFloat(fields[0], 64)
			boundaries = append(boundaries, le)
		}
	}

	// TODO: total = counts[last], target = quantile * total
	// Найди bucket где counts[i] >= target
	// Интерполируй: lowerBound + (upperBound - lowerBound) * (target - prevCount) / (count - prevCount)
	// Выведи с 3 знаками: fmt.Printf("%.3f\n", result)
	_ = quantile
	_ = boundaries
	_ = counts
	_ = strconv.ParseFloat
	_ = strings.Fields
}`,
				Glossary: []GlossaryItem{
					{Term: "histogram_quantile", Definition: "PromQL функция: оценивает значение перцентиля из bucket-ов гистограммы методом линейной интерполяции."},
					{Term: "le (less or equal)", Definition: "Label bucket-а гистограммы: верхняя граница. le=\"0.5\" значит: count запросов <= 0.5 секунд."},
					{Term: "strconv.ParseFloat(s, 64)", Definition: "Парсит строку в float64. Возвращает (value, error)."},
				},
				TestCases: []TestCase{
					{Input: "0.99\n0.005 10\n0.01 20\n0.025 40\n0.05 80\n0.1 150\n0.25 400\n0.5 800\n1.0 950\n2.5 990\n5.0 998\n10.0 1000\n+Inf 1000", ExpectedOutput: "2.250"},
					{Input: "0.50\n0.1 100\n0.5 400\n1.0 900\n+Inf 1000", ExpectedOutput: "0.225"},
				},
			},
		},
	}
}
