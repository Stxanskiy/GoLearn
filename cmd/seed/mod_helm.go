package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Helm & Kubernetes
// 6 уроков: K8s concepts → Helm → Charts → Templates → Advanced → Production
// ════════════════════════════════════════════════════════════════

func mod_helm() M {
	return M{
		Slug:          "helm-k8s",
		Title:         "Helm",
		Description:   "Kubernetes-концепции, Helm charts, templating, releases, production patterns с ArgoCD и Helmfile.",
		Order:         22,
		Track:         "devops",
		Difficulty:    "advanced",
		Prerequisites: []string{"linux-fundamentals", "docker-linux"},
		Lessons: []L{
			helmLesson01K8sConcepts(),
			helmLesson02Overview(),
			helmLesson03ChartStructure(),
			helmLesson04Templating(),
			helmLesson05Advanced(),
			helmLesson06Production(),
		},
	}
}

func helmLesson01K8sConcepts() L {
	return L{
		Slug: "k8s-concepts", Title: "Kubernetes — ключевые концепции", Order: 1,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Kubernetes — ключевые концепции</h1>

<h2>Зачем Kubernetes?</h2>
<p>Docker запускает контейнеры на ОДНОЙ машине. Kubernetes (K8s) — это <strong>оркестратор</strong>: он управляет контейнерами на КЛАСТЕРЕ машин. Решает проблемы:</p>
<ul>
<li><strong>Масштабирование</strong> — запустить 50 копий сервиса на 10 нодах</li>
<li><strong>Self-healing</strong> — если контейнер упал, перезапустить автоматически</li>
<li><strong>Service discovery</strong> — сервисы находят друг друга по имени, не по IP</li>
<li><strong>Rolling updates</strong> — обновить без downtime</li>
</ul>

<h2>Архитектура кластера</h2>
<pre><code>┌─── Control Plane (master) ────────────────┐
│ API Server → etcd (хранилище состояния)  │
│ Scheduler → выбирает ноду для Pod        │
│ Controller Manager → следит за desired state │
└───────────────────────────────────────────┘
        │
┌───────┼──── Worker Nodes ──────────────────┐
│ Node1: kubelet + container runtime + kube-proxy │
│ Node2: kubelet + container runtime + kube-proxy │
│ Node3: kubelet + container runtime + kube-proxy │
└────────────────────────────────────────────┘</code></pre>

<h2>Pod — минимальная единица</h2>
<pre><code>apiVersion: v1
kind: Pod
metadata:
  name: my-app
  labels:
    app: web
spec:
  containers:
  - name: app
    image: myapp:1.0
    ports:
    - containerPort: 8080
    resources:
      requests:
        memory: "64Mi"
        cpu: "250m"
      limits:
        memory: "128Mi"
        cpu: "500m"</code></pre>
<p>Pod = 1+ контейнеров с общим network namespace и volumes. Обычно 1 Pod = 1 контейнер.</p>

<h2>Deployment — управление репликами</h2>
<pre><code>apiVersion: apps/v1
kind: Deployment
metadata:
  name: web
spec:
  replicas: 3
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
      - name: web
        image: myapp:2.0
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1        # сколько лишних Pod при обновлении
      maxUnavailable: 0  # 0 = zero-downtime</code></pre>

<h2>Service — сетевой доступ</h2>
<pre><code>apiVersion: v1
kind: Service
metadata:
  name: web-service
spec:
  selector:
    app: web        # найти все Pod с label app=web
  ports:
  - port: 80       # порт Service
    targetPort: 8080  # порт контейнера
  type: ClusterIP   # внутренний (LoadBalancer для внешнего)</code></pre>
<p><strong>Service types:</strong> ClusterIP (внутренний), NodePort (порт на каждой ноде), LoadBalancer (облачный LB), ExternalName (DNS alias).</p>

<h2>Declarative vs Imperative</h2>
<pre><code># Imperative (как): "создай 3 реплики"
kubectl create deployment web --image=myapp:1.0 --replicas=3

# Declarative (что): "я хочу чтобы было 3 реплики"
kubectl apply -f deployment.yaml
# K8s сам разбирается как достичь desired state</code></pre>
<p>K8s — <strong>декларативная система</strong>. Ты описываешь желаемое состояние, K8s сам приводит реальность к нему через reconciliation loop.</p>`,

		Quiz: []Q{
			{
				Question:    "Pod упал на Node2. Что делает Kubernetes?",
				Options:     []string{"Ничего — ждёт ручного вмешательства", "Controller Manager замечает расхождение desired/actual state и создаёт новый Pod", "Перезагружает Node2", "Останавливает весь Deployment"},
				Correct:     1,
				Explanation: "K8s работает через reconciliation loop: Controller Manager постоянно сравнивает desired state (3 реплики) с actual state (2 реплики). При расхождении — создаёт новый Pod. Scheduler выбирает ноду. Это self-healing.",
			},
			{
				Question:    "Чем Service отличается от Pod по сетевой доступности?",
				Options:     []string{"Ничем", "Service имеет стабильный IP/DNS, Pod — нет (IP меняется при рестарте)", "Pod доступен снаружи, Service — нет", "Service быстрее"},
				Correct:     1,
				Explanation: "Pod получает случайный IP при создании, который меняется при рестарте. Service имеет стабильный ClusterIP и DNS-имя (web-service.default.svc.cluster.local). Service балансирует трафик между Pod по labels.",
			},
			{
				Question:    "resources.requests vs resources.limits. В чём разница?",
				Options:     []string{"Нет разницы", "requests — гарантированный минимум ресурсов; limits — максимум (при превышении CPU throttling, при превышении memory — OOMKill)", "requests для production, limits для dev", "limits опционально"},
				Correct:     1,
				Explanation: "requests = сколько ресурсов зарезервировано (scheduler учитывает). limits = максимум. CPU > limit → throttling (замедление). Memory > limit → OOMKilled (убит). Без limits Pod может съесть все ресурсы ноды.",
			},
		},
		Tasks: []T{
			{
				Title:       "Парсер Kubernetes Deployment YAML",
				Description: `<p>Напиши парсер упрощённого K8s Deployment YAML. Из входных параметров (name, image, replicas, port) сгенерируй корректный YAML.</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "Deployment", Definition: "K8s ресурс управляющий репликами Pod с rolling update стратегией"},
					{Term: "replicas", Definition: "Количество одновременно работающих копий Pod"},
					{Term: "selector", Definition: "Критерий по которому Deployment находит свои Pod (обычно по labels)"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func generateDeploymentYAML(name, image string, replicas, port int) string {
	// Сгенерируй минимальный Deployment YAML:
	// apiVersion: apps/v1
	// kind: Deployment
	// metadata.name: <name>
	// spec.replicas: <replicas>
	// spec.selector.matchLabels.app: <name>
	// spec.template.metadata.labels.app: <name>
	// spec.template.spec.containers[0].name: <name>
	// spec.template.spec.containers[0].image: <image>
	// spec.template.spec.containers[0].ports[0].containerPort: <port>
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	parts := strings.Fields(scanner.Text())
	// input: name image replicas port
	var replicas, port int
	fmt.Sscanf(parts[2], "%d", &replicas)
	fmt.Sscanf(parts[3], "%d", &port)
	fmt.Print(generateDeploymentYAML(parts[0], parts[1], replicas, port))
}`,
				Hints: `<ul><li>Используй fmt.Sprintf с форматированием YAML</li><li>Отступы в YAML: 2 пробела на уровень</li><li>labels в selector и template должны совпадать</li></ul>`,
				Solution: `<pre><code>func generateDeploymentYAML(name, image string, replicas, port int) string {
    return fmt.Sprintf("apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: %s\nspec:\n  replicas: %d\n  selector:\n    matchLabels:\n      app: %s\n  template:\n    metadata:\n      labels:\n        app: %s\n    spec:\n      containers:\n      - name: %s\n        image: %s\n        ports:\n        - containerPort: %d\n", name, replicas, name, name, name, image, port)
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "web nginx:1.21 3 80", ExpectedOutput: "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: web\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: web\n  template:\n    metadata:\n      labels:\n        app: web\n    spec:\n      containers:\n      - name: web\n        image: nginx:1.21\n        ports:\n        - containerPort: 80"},
				},
			},
		},
	}
}

func helmLesson02Overview() L {
	return L{
		Slug: "helm-overview", Title: "Что такое Helm и зачем он нужен", Order: 2,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Helm — пакетный менеджер для Kubernetes</h1>

<h2>Проблема без Helm</h2>
<p>Типичное приложение в K8s = 5-15 YAML файлов (Deployment, Service, ConfigMap, Secret, Ingress, PVC, RBAC...). Проблемы:</p>
<ul>
<li>Дублирование: имя приложения повторяется 20+ раз в разных файлах</li>
<li>Окружения: dev/staging/prod отличаются только значениями (replicas, image tag)</li>
<li>Версионирование: как откатить "набор YAML" к прошлой версии?</li>
<li>Шаринг: как опубликовать свой набор манифестов для других?</li>
</ul>

<h2>Helm решает всё это</h2>
<pre><code># Helm Chart = шаблонизированные K8s манифесты + значения
#
# Chart ≈ пакет (как .deb или .rpm, но для K8s)
# Release = установленный экземпляр Chart в кластере
# Repository = хранилище Charts (как apt repo)
#
# Команды:
helm install my-app ./chart      # установить
helm upgrade my-app ./chart      # обновить
helm rollback my-app 1           # откатить к ревизии 1
helm uninstall my-app            # удалить
helm list                         # показать releases</code></pre>

<h2>Helm 3 vs Helm 2</h2>
<pre><code># Helm 2: клиент + Tiller (серверный компонент в кластере)
# Проблема: Tiller имел cluster-admin права → security risk

# Helm 3 (текущий): только клиент, работает через kubeconfig
# Нет Tiller → безопаснее, проще
# Releases хранятся в Secrets/ConfigMaps в namespace</code></pre>

<h2>Values — параметризация</h2>
<pre><code># values.yaml (дефолты чарта):
replicaCount: 2
image:
  repository: myapp
  tag: "1.0.0"
service:
  type: ClusterIP
  port: 80

# Переопределение при установке:
helm install app ./chart -f production-values.yaml
helm install app ./chart --set image.tag=2.0.0

# Приоритет values (от низкого к высокому):
# 1. values.yaml (дефолты чарта)
# 2. -f custom-values.yaml (файл)
# 3. --set key=value (командная строка)</code></pre>`,

		Quiz: []Q{
			{
				Question:    "В Helm 3, где хранится информация о releases?",
				Options:     []string{"В отдельной базе данных", "В Secrets внутри namespace кластера", "На диске Helm клиента", "В Tiller"},
				Correct:     1,
				Explanation: "Helm 3 хранит release metadata в Kubernetes Secrets (тип helm.sh/release.v1) внутри namespace. Это значит: разные namespaces могут иметь releases с одинаковым именем, и информация переживает перезапуск клиента.",
			},
			{
				Question:    "helm install --set image.tag=2.0 и -f values.yaml оба задают image.tag. Что победит?",
				Options:     []string{"-f файл", "--set (командная строка имеет высший приоритет)", "Последний указанный", "Ошибка"},
				Correct:     1,
				Explanation: "--set всегда имеет высший приоритет. Порядок: defaults < -f file < --set. Это позволяет иметь base values в файле и переопределять конкретные значения в CI/CD pipeline через --set.",
			},
			{
				Question:    "Chart vs Release. В чём разница?",
				Options:     []string{"Нет разницы", "Chart — шаблон/пакет, Release — конкретная установка Chart в кластер с конкретными values", "Release — новая версия Chart", "Chart для dev, Release для prod"},
				Correct:     1,
				Explanation: "Chart = шаблон (как Docker image). Release = работающий экземпляр (как Docker container). Один Chart можно установить несколько раз с разными именами и values: helm install app1 ./chart, helm install app2 ./chart --set port=9090.",
			},
		},
		Tasks: []T{
			{
				Title:       "Values merge — глубокое слияние конфигов",
				Description: `<p>Реализуй deep merge для Helm values. Два словаря (base и override) сливаются рекурсивно: override перезаписывает значения base, вложенные словари мержатся.</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "deep merge", Definition: "Рекурсивное слияние: вложенные объекты мержатся, а не заменяются целиком"},
					{Term: "values precedence", Definition: "Порядок приоритета: defaults < -f file < --set"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Простой формат: key=value, вложенность через точку (image.tag=2.0)
func mergeValues(base, override map[string]string) map[string]string {
	// override перезаписывает base
	// Верни объединённый словарь
	// TODO
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	var n int
	fmt.Sscanf(scanner.Text(), "%d", &n)

	base := make(map[string]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(scanner.Text(), "=", 2)
		base[parts[0]] = parts[1]
	}

	scanner.Scan()
	var m int
	fmt.Sscanf(scanner.Text(), "%d", &m)

	override := make(map[string]string)
	for i := 0; i < m; i++ {
		scanner.Scan()
		parts := strings.SplitN(scanner.Text(), "=", 2)
		override[parts[0]] = parts[1]
	}

	result := mergeValues(base, override)
	keys := make([]string, 0, len(result))
	for k := range result {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, result[k])
	}
}`,
				Hints: `<ul><li>Скопируй все base в результат</li><li>Потом пройди по override и перезапиши/добавь</li></ul>`,
				Solution: `<pre><code>func mergeValues(base, override map[string]string) map[string]string {
    result := make(map[string]string)
    for k, v := range base {
        result[k] = v
    }
    for k, v := range override {
        result[k] = v
    }
    return result
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "3\nimage.repo=nginx\nimage.tag=1.0\nreplicas=2\n2\nimage.tag=2.0\nport=80", ExpectedOutput: "image.repo=nginx\nimage.tag=2.0\nport=80\nreplicas=2"},
				},
			},
		},
	}
}

func helmLesson03ChartStructure() L {
	return L{
		Slug: "chart-structure", Title: "Структура Helm Chart", Order: 3,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Структура Helm Chart</h1>

<h2>Файловое дерево</h2>
<pre><code>mychart/
├── Chart.yaml          # Метаданные (имя, версия, описание)
├── values.yaml         # Дефолтные значения
├── charts/             # Зависимости (subcharts)
├── templates/          # Шаблоны K8s манифестов
│   ├── deployment.yaml
│   ├── service.yaml
│   ├── ingress.yaml
│   ├── _helpers.tpl    # Переиспользуемые шаблоны (начинаются с _)
│   ├── NOTES.txt       # Текст после install (инструкции пользователю)
│   └── tests/          # Helm test манифесты
├── .helmignore         # Файлы для исключения из пакета
└── README.md</code></pre>

<h2>Chart.yaml</h2>
<pre><code>apiVersion: v2            # v2 для Helm 3
name: myapp
version: 1.2.3            # Версия CHART (не приложения!)
appVersion: "2.0.0"       # Версия приложения (informational)
description: My awesome app
type: application          # application | library
dependencies:
  - name: postgresql
    version: "12.x"
    repository: "https://charts.bitnami.com"
    condition: postgresql.enabled  # включить только если true в values</code></pre>

<h2>_helpers.tpl — DRY в шаблонах</h2>
<pre><code>{{/* Полное имя с release */}}
{{- define "myapp.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/* Общие labels */}}
{{- define "myapp.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
app.kubernetes.io/version: {{ .Chart.AppVersion }}
helm.sh/chart: {{ printf "%s-%s" .Chart.Name .Chart.Version }}
{{- end }}</code></pre>

<h2>NOTES.txt</h2>
<pre><code>Congratulations! {{ .Chart.Name }} has been deployed.

Get the application URL:
{{- if eq .Values.service.type "LoadBalancer" }}
  kubectl get svc {{ include "myapp.fullname" . }}
{{- else }}
  kubectl port-forward svc/{{ include "myapp.fullname" . }} 8080:{{ .Values.service.port }}
{{- end }}</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Chart.yaml: version vs appVersion. В чём разница?",
				Options:     []string{"Нет разницы", "version — версия самого Chart (пакета), appVersion — версия приложения внутри (informational)", "version для Helm 2, appVersion для Helm 3", "appVersion обязательна, version — нет"},
				Correct:     1,
				Explanation: "version — SemVer версия Chart-пакета. Меняется когда меняются templates/values. appVersion — версия приложения (image tag). Можно обновить Chart (поменять лимиты ресурсов) без изменения appVersion, и наоборот.",
			},
			{
				Question:    "Файл templates/_helpers.tpl. Почему начинается с подчёркивания?",
				Options:     []string{"Стиль кода", "Файлы с _ не рендерятся как K8s манифесты — только содержат определения шаблонов", "Для приватности", "Загружается первым"},
				Correct:     1,
				Explanation: "Helm рендерит все файлы в templates/ как K8s манифесты КРОМЕ файлов начинающихся с _. _helpers.tpl содержит {{define}} блоки которые переиспользуются через {{include}} в других шаблонах, но сам не генерирует YAML.",
			},
			{
				Question:    "type: library в Chart.yaml. Что это значит?",
				Options:     []string{"Чарт для библиотек кода", "Library chart не генерирует K8s ресурсы — только предоставляет шаблоны для использования другими charts", "Чарт для shared libraries", "Deprecated тип"},
				Correct:     1,
				Explanation: "Library charts содержат только _helpers.tpl определения без templates для K8s ресурсов. Их нельзя установить напрямую. Используются как dependencies другими charts для переиспользования шаблонов (DRY principle).",
			},
		},
		Tasks: []T{
			{
				Title:       "Валидатор Chart.yaml",
				Description: `<p>Проверь что Chart.yaml содержит все обязательные поля и версии в формате SemVer.</p>`,
				Difficulty:  "easy",
				Glossary: []GlossaryItem{
					{Term: "SemVer", Definition: "Semantic Versioning: MAJOR.MINOR.PATCH (1.2.3)"},
					{Term: "apiVersion", Definition: "Версия API Chart — v2 для Helm 3"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

func validateChart(fields map[string]string) string {
	// Обязательные поля: apiVersion, name, version
	// version должна быть SemVer (X.Y.Z где X,Y,Z — числа)
	// apiVersion должна быть "v2"
	// Верни "valid" или описание первой ошибки
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	fields := make(map[string]string)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			fields[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
		}
	}
	fmt.Println(validateChart(fields))
}`,
				Hints: `<ul><li>Проверь наличие apiVersion, name, version</li><li>SemVer regex: ^\d+\.\d+\.\d+$</li></ul>`,
				Solution: `<pre><code>func validateChart(fields map[string]string) string {
    if _, ok := fields["apiVersion"]; !ok { return "missing apiVersion" }
    if fields["apiVersion"] != "v2" { return "apiVersion must be v2" }
    if _, ok := fields["name"]; !ok { return "missing name" }
    v, ok := fields["version"]
    if !ok { return "missing version" }
    semver := regexp.MustCompile("^\\d+\\.\\d+\\.\\d+$")
    if !semver.MatchString(v) { return "version must be SemVer" }
    return "valid"
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "apiVersion: v2\nname: myapp\nversion: 1.0.0", ExpectedOutput: "valid"},
					{Input: "apiVersion: v2\nname: myapp\nversion: abc", ExpectedOutput: "version must be SemVer"},
					{Input: "apiVersion: v1\nname: myapp\nversion: 1.0.0", ExpectedOutput: "apiVersion must be v2"},
					{Input: "name: myapp\nversion: 1.0.0", ExpectedOutput: "missing apiVersion"},
				},
			},
		},
	}
}

func helmLesson04Templating() L {
	return L{
		Slug: "helm-templating", Title: "Helm Templating — Go templates в действии", Order: 4,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Helm Templating</h1>

<h2>Основы Go template в Helm</h2>
<pre><code># {{ }} — действие шаблона (template action)
# .Values — доступ к values.yaml
# .Release — информация о release (Name, Namespace, Revision)
# .Chart — из Chart.yaml

apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
  labels:
    {{- include "myapp.labels" . | nindent 4 }}
data:
  app_name: {{ .Values.app.name | quote }}
  replicas: {{ .Values.replicaCount | toString | quote }}</code></pre>

<h2>Условия и циклы</h2>
<pre><code># if/else:
{{- if .Values.ingress.enabled }}
apiVersion: networking.k8s.io/v1
kind: Ingress
...
{{- end }}

# range (цикл):
env:
{{- range $key, $val := .Values.env }}
  - name: {{ $key }}
    value: {{ $val | quote }}
{{- end }}

# with (scope):
{{- with .Values.resources }}
resources:
  {{- toYaml . | nindent 2 }}
{{- end }}</code></pre>

<h2>Дефис в {{ }} — trim whitespace</h2>
<pre><code># {{- ... }} — trim пробелы/переводы строк СЛЕВА
# {{ ... -}} — trim СПРАВА
# {{- ... -}} — trim с обеих сторон
#
# БЕЗ дефиса:
{{ if .Values.debug }}
debug: true
{{ end }}
# Результат: пустые строки вокруг!

# С дефисом:
{{- if .Values.debug }}
debug: true
{{- end }}
# Результат: чистый YAML</code></pre>

<h2>nindent — спаситель отступов</h2>
<pre><code># nindent N = newline + indent N пробелов
# Используй с | (pipe):
metadata:
  annotations:
    {{- include "myapp.annotations" . | nindent 4 }}

# indent (без n) = только отступ, без newline
# toYaml = конвертировать Go struct в YAML строку</code></pre>`,

		Quiz: []Q{
			{
				Question:    "{{- и {{ в Helm templates. В чём разница?",
				Options:     []string{"Нет разницы", "{{- удаляет whitespace (пробелы/переводы строк) слева от действия", "{{- это комментарий", "{{- для переменных"},
				Correct:     1,
				Explanation: "Дефис после {{ или перед }} — trim whitespace. {{- trim слева, -}} trim справа. Без этого Helm генерирует пустые строки в YAML (из-за строк с только {{ if }}/{{ end }}). Пустые строки в YAML могут ломать парсинг.",
			},
			{
				Question:    "| nindent 4 в {{ include \"x\" . | nindent 4 }}. Что делает?",
				Options:     []string{"Ничего", "Добавляет newline + 4 пробела отступа к каждой строке результата", "Обрезает до 4 символов", "Indent только первой строки"},
				Correct:     1,
				Explanation: "nindent = new line + indent. Добавляет перевод строки В НАЧАЛЕ, затем к каждой строке добавляет N пробелов. Критически важно для правильных отступов в YAML. Без nindent вложенные шаблоны ломают структуру YAML.",
			},
			{
				Question:    "{{ .Values.name | quote }} vs {{ .Values.name }}. Когда нужен quote?",
				Options:     []string{"Всегда нужен", "Когда значение может содержать спецсимволы YAML или начинаться с цифры — quote оборачивает в кавычки", "Никогда не нужен", "Только для строк"},
				Correct:     1,
				Explanation: "YAML без кавычек: 'true' парсится как bool, '123' как int, 'null' как null. | quote оборачивает значение в двойные кавычки, гарантируя что оно будет строкой. Best practice: всегда quote для значений из values.",
			},
		},
		Tasks: []T{
			{
				Title:       "Мини-шаблонизатор {{ .Values }}",
				Description: `<p>Реализуй простой template engine: замени {{ .Values.KEY }} на значение из словаря.</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "template action", Definition: "{{ }} — место в шаблоне для подстановки значений"},
					{Term: ".Values", Definition: "Корневой объект с пользовательскими значениями"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func renderTemplate(tmpl string, values map[string]string) string {
	// Замени все {{ .Values.KEY }} на values[KEY]
	// Если ключа нет — оставь <no value>
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	scanner.Scan()
	tmpl := scanner.Text()

	scanner.Scan()
	var n int
	fmt.Sscanf(scanner.Text(), "%d", &n)

	values := make(map[string]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(scanner.Text(), "=", 2)
		values[parts[0]] = parts[1]
	}

	fmt.Println(renderTemplate(tmpl, values))
}`,
				Hints: `<ul><li>Ищи "{{ .Values." в строке</li><li>Найди закрывающий " }}"</li><li>Извлеки ключ между .Values. и }}</li><li>Замени всю конструкцию на значение</li></ul>`,
				Solution: `<pre><code>func renderTemplate(tmpl string, values map[string]string) string {
    result := tmpl
    for {
        start := strings.Index(result, "{{ .Values.")
        if start == -1 { break }
        end := strings.Index(result[start:], " }}")
        if end == -1 { break }
        end += start + 3
        key := result[start+12 : end-3] // extract key
        val, ok := values[key]
        if !ok { val = "<no value>" }
        result = result[:start] + val + result[end:]
    }
    return result
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "name: {{ .Values.name }}\n2\nname=myapp\nport=8080", ExpectedOutput: "name: myapp"},
					{Input: "image: {{ .Values.repo }}:{{ .Values.tag }}\n2\nrepo=nginx\ntag=latest", ExpectedOutput: "image: nginx:latest"},
					{Input: "x: {{ .Values.missing }}\n1\nfoo=bar", ExpectedOutput: "x: <no value>"},
				},
			},
		},
	}
}

func helmLesson05Advanced() L {
	return L{
		Slug: "helm-advanced", Title: "Advanced Helm — hooks, dependencies, rollback", Order: 5,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Advanced Helm</h1>

<h2>Hooks — действия на этапах lifecycle</h2>
<pre><code># Hooks выполняются на определённых этапах release lifecycle:
# pre-install  → перед первой установкой
# post-install → после установки
# pre-upgrade  → перед обновлением
# post-upgrade → после обновления
# pre-delete   → перед удалением
# post-delete  → после удаления
# pre-rollback → перед откатом

# Пример: миграция БД перед upgrade
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-migrate
  annotations:
    "helm.sh/hook": pre-upgrade
    "helm.sh/hook-weight": "0"        # порядок среди hooks
    "helm.sh/hook-delete-policy": hook-succeeded  # удалить Job после успеха
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: {{ .Values.image.repository }}:{{ .Values.image.tag }}
        command: ["./migrate", "up"]
      restartPolicy: Never</code></pre>

<h2>Dependencies — подчарты</h2>
<pre><code># Chart.yaml:
dependencies:
  - name: postgresql
    version: "~12.0"    # >= 12.0.0, < 13.0.0
    repository: "https://charts.bitnami.com/bitnami"
    condition: postgresql.enabled
    alias: db           # использовать .Values.db вместо .Values.postgresql

# Обновить зависимости:
helm dependency update ./mychart
# Скачает charts/postgresql-12.x.tgz

# Values для subchart:
postgresql:       # имя subchart = ключ в values
  auth:
    postgresPassword: secret123
    database: myapp</code></pre>

<h2>Rollback стратегии</h2>
<pre><code># Откатить к предыдущей ревизии:
helm rollback myapp 1

# Atomic install/upgrade — автоматический rollback при ошибке:
helm upgrade myapp ./chart --atomic --timeout 5m
# Если Pod не стал Ready за 5 минут → автоматический rollback

# --wait: ждать пока все ресурсы станут Ready
helm upgrade myapp ./chart --wait --timeout 3m

# История ревизий:
helm history myapp
# REVISION  STATUS      DESCRIPTION
# 1         superseded  Install complete
# 2         deployed    Upgrade complete</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Hook pre-upgrade с Job. Job упал. Что произойдёт с upgrade?",
				Options:     []string{"Upgrade продолжится", "Upgrade остановится — pre-hooks должны успешно завершиться", "Job перезапустится бесконечно", "Helm проигнорирует ошибку"},
				Correct:     1,
				Explanation: "Pre-hooks БЛОКИРУЮТ основную операцию. Если pre-upgrade Job завершился с ошибкой — upgrade не произойдёт. Это механизм безопасности: миграция БД должна пройти перед деплоем нового кода.",
			},
			{
				Question:    "helm upgrade --atomic. Что происходит при ошибке?",
				Options:     []string{"Ничего", "Автоматический rollback к предыдущей рабочей ревизии", "Helm ретраит", "Кластер перезагружается"},
				Correct:     1,
				Explanation: "--atomic = --wait + автоматический rollback при ошибке. Если после upgrade Pod не стали Ready за timeout — Helm автоматически откатывает к прошлой рабочей ревизии. Must-have для CI/CD pipelines.",
			},
		},
		Tasks: []T{
			{
				Title:       "Release версионирование с rollback",
				Description: `<p>Реализуй систему управления ревизиями release. Поддержи операции: deploy (новая ревизия), rollback N (откат к ревизии N), history (показать все).</p>`,
				Difficulty:  "medium",
				Glossary: []GlossaryItem{
					{Term: "revision", Definition: "Версия release — инкрементируется при каждом upgrade/rollback"},
					{Term: "rollback", Definition: "Откат к предыдущему состоянию release"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Release struct {
	revisions []string // revision[i] = description
	current   int      // текущая ревизия (1-based)
}

func NewRelease() *Release {
	return &Release{}
}

func (r *Release) Deploy(desc string) string {
	// Добавить новую ревизию
	// Вернуть "deployed revision N"
	// TODO
	return ""
}

func (r *Release) Rollback(rev int) string {
	// Откатиться к ревизии rev
	// Создать НОВУЮ ревизию с описанием "rollback to N"
	// Вернуть "rolled back to revision N" или "revision not found"
	// TODO
	return ""
}

func (r *Release) History() string {
	// Вернуть все ревизии: "N: desc [current]" для текущей
	// TODO
	return ""
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	rel := NewRelease()
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" { continue }
		parts := strings.SplitN(line, " ", 2)
		switch parts[0] {
		case "deploy":
			fmt.Println(rel.Deploy(parts[1]))
		case "rollback":
			var n int
			fmt.Sscanf(parts[1], "%d", &n)
			fmt.Println(rel.Rollback(n))
		case "history":
			fmt.Print(rel.History())
		}
	}
}`,
				Hints: `<ul><li>Deploy: append описание, current = len(revisions)</li><li>Rollback: проверить что rev существует, добавить новую ревизию "rollback to N"</li><li>History: пройтись по всем ревизиям, пометить current</li></ul>`,
				Solution: `<pre><code>func (r *Release) Deploy(desc string) string {
    r.revisions = append(r.revisions, desc)
    r.current = len(r.revisions)
    return fmt.Sprintf("deployed revision %d", r.current)
}
func (r *Release) Rollback(rev int) string {
    if rev < 1 || rev > len(r.revisions) { return "revision not found" }
    desc := fmt.Sprintf("rollback to %d", rev)
    r.revisions = append(r.revisions, desc)
    r.current = len(r.revisions)
    return fmt.Sprintf("rolled back to revision %d", rev)
}
func (r *Release) History() string {
    var sb strings.Builder
    for i, d := range r.revisions {
        marker := ""
        if i+1 == r.current { marker = " [current]" }
        sb.WriteString(fmt.Sprintf("%d: %s%s\n", i+1, d, marker))
    }
    return sb.String()
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "deploy v1.0\ndeploy v2.0\nrollback 1\nhistory", ExpectedOutput: "deployed revision 1\ndeployed revision 2\nrolled back to revision 1\n1: v1.0\n2: v2.0\n3: rollback to 1 [current]"},
				},
			},
		},
	}
}

func helmLesson06Production() L {
	return L{
		Slug: "helm-production", Title: "Production Helm — Helmfile, GitOps, ArgoCD", Order: 6,
		Difficulty: "advanced", Track: "devops",
		Content: `<h1>Production Helm Patterns</h1>

<h2>Helmfile — декларативное управление releases</h2>
<pre><code># helmfile.yaml — описывает ВСЕ releases в кластере
repositories:
  - name: bitnami
    url: https://charts.bitnami.com/bitnami

releases:
  - name: app
    namespace: production
    chart: ./charts/myapp
    version: 1.2.3
    values:
      - values/common.yaml
      - values/{{ .Environment.Name }}.yaml
    set:
      - name: image.tag
        value: {{ requiredEnv "IMAGE_TAG" }}

environments:
  staging:
    values: [env/staging.yaml]
  production:
    values: [env/production.yaml]

# helmfile -e production apply</code></pre>

<h2>GitOps с ArgoCD</h2>
<pre><code># GitOps принцип: Git = single source of truth
# ArgoCD следит за Git repo и автоматически синхронизирует кластер

# Application CRD:
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: myapp
  namespace: argocd
spec:
  source:
    repoURL: https://github.com/org/deploy.git
    path: charts/myapp
    targetRevision: HEAD
    helm:
      valueFiles:
        - values-prod.yaml
  destination:
    server: https://kubernetes.default.svc
    namespace: production
  syncPolicy:
    automated:
      prune: true      # удалять ресурсы которых нет в Git
      selfHeal: true   # восстанавливать drift</code></pre>

<h2>Multi-environment стратегия</h2>
<pre><code>deploy-repo/
├── base/                # общие шаблоны
│   ├── Chart.yaml
│   └── templates/
├── environments/
│   ├── dev/
│   │   └── values.yaml   # replicas: 1, debug: true
│   ├── staging/
│   │   └── values.yaml   # replicas: 2, debug: false
│   └── production/
│       └── values.yaml   # replicas: 5, resources.limits.memory: 1Gi
└── helmfile.yaml</code></pre>`,

		Quiz: []Q{
			{
				Question:    "GitOps с ArgoCD. Кто-то вручную изменил replicas через kubectl. Что произойдёт?",
				Options:     []string{"Изменение останется", "ArgoCD обнаружит drift и вернёт к состоянию из Git (selfHeal)", "ArgoCD упадёт", "Нужен ручной sync"},
				Correct:     1,
				Explanation: "С syncPolicy.automated.selfHeal: true ArgoCD постоянно сравнивает actual state с desired state из Git. При drift (ручные изменения) — автоматически откатывает к Git-состоянию. Git = единственный источник правды.",
			},
			{
				Question:    "Helmfile vs helm install. Главное преимущество Helmfile?",
				Options:     []string{"Быстрее", "Декларативное описание ВСЕХ releases + environments в одном файле", "Лучше для одного сервиса", "Не нужен Helm"},
				Correct:     1,
				Explanation: "Helmfile — meta-tool поверх Helm. Позволяет описать десятки releases, их зависимости, multi-environment конфигурацию в одном декларативном файле. helmfile apply = установить/обновить ВСЁ одной командой.",
			},
		},
		Tasks: []T{
			{
				Title:       "Multi-env values resolver",
				Description: `<p>Реализуй resolver для multi-environment values. Base values + environment overrides = final values.</p>`,
				Difficulty:  "easy",
				Glossary: []GlossaryItem{
					{Term: "environment", Definition: "Окружение: dev, staging, production — каждое со своими настройками"},
					{Term: "override", Definition: "Перезапись значений base конфигурации для конкретного окружения"},
				},
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func resolveValues(base map[string]string, env map[string]string) map[string]string {
	// env перезаписывает base
	// TODO
	return nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	base := make(map[string]string)
	env := make(map[string]string)

	scanner.Scan()
	var nb int
	fmt.Sscanf(scanner.Text(), "%d", &nb)
	for i := 0; i < nb; i++ {
		scanner.Scan()
		p := strings.SplitN(scanner.Text(), "=", 2)
		base[p[0]] = p[1]
	}

	scanner.Scan()
	var ne int
	fmt.Sscanf(scanner.Text(), "%d", &ne)
	for i := 0; i < ne; i++ {
		scanner.Scan()
		p := strings.SplitN(scanner.Text(), "=", 2)
		env[p[0]] = p[1]
	}

	result := resolveValues(base, env)
	keys := make([]string, 0, len(result))
	for k := range result { keys = append(keys, k) }
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("%s=%s\n", k, result[k])
	}
}`,
				Hints: `<ul><li>Скопируй base, потом override с env</li></ul>`,
				Solution: `<pre><code>func resolveValues(base map[string]string, env map[string]string) map[string]string {
    r := make(map[string]string)
    for k, v := range base { r[k] = v }
    for k, v := range env { r[k] = v }
    return r
}</code></pre>`,
				TestCases: []TestCase{
					{Input: "3\nreplicas=2\ndebug=true\nport=8080\n2\nreplicas=5\ndebug=false", ExpectedOutput: "debug=false\nport=8080\nreplicas=5"},
				},
			},
		},
	}
}
