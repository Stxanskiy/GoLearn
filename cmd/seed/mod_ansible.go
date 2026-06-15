package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Ansible — автоматизация инфраструктуры
// ════════════════════════════════════════════════════════════════

func mod_ansible() M {
	return M{
		Slug:          "ansible",
		Title:         "Ansible — автоматизация инфраструктуры",
		Description:   "Agentless-автоматизация: inventory, playbooks, роли, шаблоны, деплой Go-приложений. От первого ad-hoc до полного стека.",
		Order:         21,
		Track:         "devops",
		Difficulty:    "intermediate",
		Prerequisites: []string{"linux-fundamentals"},
		Lessons: []L{
			lesson_ansible_intro(),
			lesson_ansible_playbooks(),
			lesson_ansible_variables(),
			lesson_ansible_roles(),
			lesson_ansible_advanced(),
			lesson_ansible_realworld(),
		},
	}
}

// ── Урок 1: Введение в Ansible ──────────────────────────────

func lesson_ansible_intro() L {
	return L{
		Slug: "ansible-intro", Title: "Введение в Ansible", Order: 1,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Введение в Ansible</h1>

<h2>Зачем нужна автоматизация?</h2>
<p>Представь: у тебя 50 серверов и нужно на каждом обновить Nginx, изменить конфиг и перезапустить сервис. Вручную — это часы работы и гарантированные ошибки. <strong>Ansible</strong> позволяет описать желаемое состояние инфраструктуры и применить его на любое количество серверов одной командой.</p>

<h2>Что такое Ansible?</h2>
<p><strong>Ansible</strong> — инструмент автоматизации от Red Hat. Ключевые особенности:</p>
<ul>
<li><strong>Agentless</strong> — не нужно устанавливать агент на целевые серверы. Достаточно SSH-доступа</li>
<li><strong>Push-модель</strong> — управляющий узел (control node) инициирует подключение и "проталкивает" конфигурацию</li>
<li><strong>Декларативный</strong> — описываешь <em>что</em> должно быть, а не <em>как</em> это сделать</li>
<li><strong>Идемпотентный</strong> — повторный запуск не сломает систему, если она уже в нужном состоянии</li>
<li><strong>YAML</strong> — конфигурация пишется на простом YAML, не нужен язык программирования</li>
</ul>

<h2>Ansible vs Chef vs Puppet vs Salt</h2>
<table>
<tr><th>Критерий</th><th>Ansible</th><th>Chef</th><th>Puppet</th><th>Salt</th></tr>
<tr><td>Агент</td><td>Нет (SSH)</td><td>Да (chef-client)</td><td>Да (puppet-agent)</td><td>Да (salt-minion)</td></tr>
<tr><td>Модель</td><td>Push</td><td>Pull</td><td>Pull</td><td>Push/Pull</td></tr>
<tr><td>Язык</td><td>YAML</td><td>Ruby DSL</td><td>Puppet DSL</td><td>YAML/Python</td></tr>
<tr><td>Порог входа</td><td>Низкий</td><td>Высокий</td><td>Средний</td><td>Средний</td></tr>
<tr><td>Масштаб</td><td>Средний*</td><td>Большой</td><td>Большой</td><td>Большой</td></tr>
</table>
<p><em>* Ansible масштабируется хуже на тысячи серверов из-за SSH, но для 90% проектов этого достаточно. AWX/Tower решает проблему.</em></p>

<h2>Архитектура Ansible</h2>
<pre><code>┌─────────────────┐         SSH          ┌──────────────┐
│  Control Node   │ ──────────────────── │  Managed Host │
│  (твой ноут)    │                      │  (сервер)     │
│                 │         SSH          ├──────────────┤
│  - ansible      │ ──────────────────── │  Managed Host │
│  - ansible.cfg  │                      │  (сервер)     │
│  - inventory    │         SSH          ├──────────────┤
│  - playbooks/   │ ──────────────────── │  Managed Host │
└─────────────────┘                      └──────────────┘</code></pre>

<h2>Inventory — список серверов</h2>
<p><strong>Inventory</strong> — файл, описывающий на каких серверах выполнять задачи. Форматы: INI или YAML.</p>

<h3>INI-формат</h3>
<pre><code># inventory.ini
[webservers]
web1.example.com ansible_port=22
web2.example.com ansible_port=2222

[dbservers]
db1.example.com ansible_user=postgres
db2.example.com

[production:children]
webservers
dbservers

[webservers:vars]
http_port=80
max_connections=1000</code></pre>

<h3>YAML-формат</h3>
<pre><code># inventory.yml
all:
  children:
    webservers:
      hosts:
        web1.example.com:
          ansible_port: 22
        web2.example.com:
          ansible_port: 2222
      vars:
        http_port: 80
    dbservers:
      hosts:
        db1.example.com:
          ansible_user: postgres</code></pre>

<h2>Ad-hoc команды</h2>
<p>Простейший способ использовать Ansible — одноразовые (ad-hoc) команды:</p>
<pre><code># Пинг всех серверов
ansible all -i inventory.ini -m ping

# Выполнить команду на webservers
ansible webservers -i inventory.ini -m shell -a "uptime"

# Установить пакет
ansible dbservers -i inventory.ini -m apt -a "name=postgresql state=present" --become

# Скопировать файл
ansible all -i inventory.ini -m copy -a "src=./app.conf dest=/etc/app.conf"</code></pre>

<h2>Как Ansible выполняет задачу</h2>
<ol>
<li>Читает inventory — определяет целевые хосты</li>
<li>Подключается по SSH (параллельно, по умолчанию 5 forks)</li>
<li>Копирует Python-модуль на целевой хост во временную директорию</li>
<li>Выполняет модуль на хосте</li>
<li>Собирает результат (JSON)</li>
<li>Удаляет временные файлы</li>
<li>Выводит итог: changed / ok / failed</li>
</ol>

<p><strong>Важно:</strong> на целевых серверах нужен только Python (2.7+ или 3.5+) и SSH-сервер. Никаких агентов.</p>`,

		Quiz: []Q{
			{
				Question:    "Почему Ansible называют agentless?",
				Options:     []string{"Потому что он не требует лицензии", "Потому что на целевых серверах не нужно устанавливать агент — достаточно SSH", "Потому что он работает без сети", "Потому что он использует pull-модель"},
				Correct:     1,
				Explanation: "Agentless означает отсутствие постоянного процесса (агента) на управляемых серверах. Ansible подключается по SSH, выполняет задачу и отключается. Chef/Puppet требуют установки клиента на каждый сервер.",
			},
			{
				Question:    "Что такое идемпотентность в контексте Ansible?",
				Options:     []string{"Каждый запуск создаёт новый сервер", "Повторный запуск даёт тот же результат — если система уже в нужном состоянии, ничего не меняется", "Ansible работает только один раз", "Задачи выполняются параллельно"},
				Correct:     1,
				Explanation: "Идемпотентность — ключевое свойство. Модуль apt state=present не будет переустанавливать пакет если он уже установлен. Это делает безопасным повторный запуск плейбуков.",
			},
			{
				Question:    "Какой формат используется для inventory в Ansible?",
				Options:     []string{"Только JSON", "Только INI", "INI или YAML (а также динамический inventory через скрипты)", "XML"},
				Correct:     2,
				Explanation: "Ansible поддерживает INI и YAML для статического inventory. Также поддерживается динамический inventory — скрипт или плагин, который генерирует JSON (например, из AWS EC2, GCP).",
			},
			{
				Question:    "Чем push-модель Ansible отличается от pull-модели Puppet?",
				Options:     []string{"Push быстрее", "В push-модели управляющий узел инициирует подключение и доставляет конфигурацию; в pull — агент сам периодически запрашивает конфигурацию с сервера", "Push работает только локально", "Разницы нет"},
				Correct:     1,
				Explanation: "Push: control node → SSH → managed hosts (Ansible инициирует). Pull: managed host → запрос к puppet master → получение каталога (агент инициирует по расписанию). Push даёт мгновенный контроль, pull — автономность.",
			},
		},

		Tasks: []T{
			{
				Title:      "Парсер INI Inventory",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "strings.HasPrefix(s, prefix)", Definition: "Проверяет начинается ли строка с указанного префикса."},
					{Term: "strings.Trim(s, cutset)", Definition: "Убирает указанные символы с обоих концов строки."},
					{Term: "strings.Fields(s)", Definition: "Разбивает строку по пробелам, игнорируя множественные пробелы."},
				},
				Description: `<p>Реализуй парсер Ansible inventory в INI-формате.</p>
<p><strong>Формат входа:</strong></p>
<ul>
<li>Строки вида <code>[groupname]</code> — начало группы</li>
<li>Строки с именем хоста (возможно с параметрами через пробел) — хост в текущей группе</li>
<li>Пустые строки и строки начинающиеся с # — игнорировать</li>
</ul>
<p><strong>Формат выхода:</strong> для каждой группы выведи <code>group: host1, host2, ...</code> (в порядке появления).</p>
<p><em>Пример входа:</em></p>
<pre><code>[webservers]
web1.example.com ansible_port=22
web2.example.com
[dbservers]
db1.example.com</code></pre>
<p><em>Выход:</em></p>
<pre><code>webservers: web1.example.com, web2.example.com
dbservers: db1.example.com</code></pre>`,
				Hints: `<p>Отслеживай текущую группу. Если строка в квадратных скобках — новая группа. Иначе — хост (бери первое поле через strings.Fields). Выводи группы в порядке их появления.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var groups []string
	hosts := make(map[string][]string)
	currentGroup := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			currentGroup = line[1 : len(line)-1]
			groups = append(groups, currentGroup)
			continue
		}
		if currentGroup != "" {
			host := strings.Fields(line)[0]
			hosts[currentGroup] = append(hosts[currentGroup], host)
		}
	}

	for _, g := range groups {
		fmt.Printf("%s: %s\n", g, strings.Join(hosts[g], ", "))
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
	// TODO: отслеживай текущую группу и её хосты
	// Если строка [name] — новая группа
	// Иначе — хост (первое слово в строке)
	var groups []string
	hosts := make(map[string][]string)
	currentGroup := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		_ = line
		// TODO: реализуй парсинг
	}

	// TODO: выведи результат
	_ = groups
	_ = hosts
	_ = currentGroup
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "[webservers]\nweb1.example.com ansible_port=22\nweb2.example.com\n[dbservers]\ndb1.example.com\n", ExpectedOutput: "webservers: web1.example.com, web2.example.com\ndbservers: db1.example.com"},
					{Input: "[app]\nserver1.local\nserver2.local\nserver3.local\n", ExpectedOutput: "app: server1.local, server2.local, server3.local"},
				},
			},
			{
				Title:      "Симулятор идемпотентности",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "map[string]bool", Definition: "Множество (set) в Go — map со значением bool для проверки наличия элемента."},
					{Term: "strings.SplitN(s, sep, n)", Definition: "Разбивает строку на максимум n частей по разделителю."},
				},
				Description: `<p>Симулируй идемпотентное выполнение задач Ansible.</p>
<p><strong>Формат входа:</strong></p>
<ul>
<li>Первая строка — текущее состояние системы: пакеты через запятую (или "empty")</li>
<li>Остальные строки — команды в формате <code>action package</code></li>
</ul>
<p><strong>Действия:</strong></p>
<ul>
<li><code>install pkg</code> — установить пакет (state=present)</li>
<li><code>remove pkg</code> — удалить пакет (state=absent)</li>
</ul>
<p><strong>Для каждой команды выведи:</strong></p>
<ul>
<li><code>changed: install pkg</code> — если пакета не было и он установлен</li>
<li><code>changed: remove pkg</code> — если пакет был и он удалён</li>
<li><code>ok: install pkg</code> — если пакет уже установлен (идемпотентность)</li>
<li><code>ok: remove pkg</code> — если пакета уже нет</li>
</ul>`,
				Hints: `<p>Храни множество установленных пакетов в map[string]bool. Для install: если пакет уже есть — "ok", иначе добавь и выведи "changed". Для remove: если пакета нет — "ok", иначе удали и выведи "changed".</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	installed := make(map[string]bool)

	scanner.Scan()
	state := strings.TrimSpace(scanner.Text())
	if state != "empty" {
		for _, pkg := range strings.Split(state, ",") {
			installed[strings.TrimSpace(pkg)] = true
		}
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		action, pkg := parts[0], parts[1]

		switch action {
		case "install":
			if installed[pkg] {
				fmt.Printf("ok: install %s\n", pkg)
			} else {
				installed[pkg] = true
				fmt.Printf("changed: install %s\n", pkg)
			}
		case "remove":
			if !installed[pkg] {
				fmt.Printf("ok: remove %s\n", pkg)
			} else {
				delete(installed, pkg)
				fmt.Printf("changed: remove %s\n", pkg)
			}
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
	installed := make(map[string]bool)

	// Первая строка — текущее состояние
	scanner.Scan()
	state := strings.TrimSpace(scanner.Text())
	// TODO: разбери state в map (если не "empty")
	_ = state

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		action, pkg := parts[0], parts[1]
		// TODO: реализуй логику идемпотентности
		_ = action
		_ = pkg
	}
	_ = installed
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "nginx,curl\ninstall nginx\ninstall git\nremove curl\nremove curl\n", ExpectedOutput: "ok: install nginx\nchanged: install git\nchanged: remove curl\nok: remove curl"},
					{Input: "empty\ninstall docker\ninstall docker\nremove docker\ninstall docker\n", ExpectedOutput: "changed: install docker\nok: install docker\nchanged: remove docker\nchanged: install docker"},
				},
			},
		},
	}
}

// ── Урок 2: Playbooks ──────────────────────────────────────────

func lesson_ansible_playbooks() L {
	return L{
		Slug: "ansible-playbooks", Title: "Playbooks — сценарии автоматизации", Order: 2,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Playbooks — сценарии автоматизации</h1>

<h2>Что такое Playbook?</h2>
<p><strong>Playbook</strong> — YAML-файл, описывающий набор задач для выполнения на группе серверов. Это главный инструмент Ansible — в отличие от ad-hoc команд, playbooks можно версионировать, переиспользовать и документировать.</p>

<h2>Структура Playbook</h2>
<pre><code># deploy.yml
---
- name: Configure web servers       # Play (сценарий)
  hosts: webservers                  # На каких хостах
  become: yes                        # Sudo (повышение привилегий)
  vars:                              # Переменные
    http_port: 80
    app_version: "1.2.3"

  tasks:                             # Задачи (выполняются последовательно)
    - name: Install Nginx
      apt:
        name: nginx
        state: present
        update_cache: yes

    - name: Copy config
      template:
        src: nginx.conf.j2
        dest: /etc/nginx/nginx.conf
      notify: Restart Nginx          # Уведомить handler

    - name: Ensure Nginx is running
      service:
        name: nginx
        state: started
        enabled: yes

  handlers:                          # Выполняются только если notify сработал
    - name: Restart Nginx
      service:
        name: nginx
        state: restarted</code></pre>

<h2>Ключевые концепции</h2>
<ul>
<li><strong>Play</strong> — блок: "на этих хостах выполни эти задачи". Playbook может содержать несколько plays</li>
<li><strong>Task</strong> — одна атомарная операция (установить пакет, скопировать файл, запустить сервис)</li>
<li><strong>Module</strong> — "глагол" задачи (apt, copy, service, template, file, user и т.д.)</li>
<li><strong>Handler</strong> — задача, которая выполняется только при изменении (notify). Вызывается один раз в конце play</li>
</ul>

<h2>Основные модули</h2>

<h3>apt — управление пакетами (Debian/Ubuntu)</h3>
<pre><code>- name: Install packages
  apt:
    name:
      - nginx
      - curl
      - htop
    state: present        # present/absent/latest
    update_cache: yes
    cache_valid_time: 3600</code></pre>

<h3>copy — копирование файлов</h3>
<pre><code>- name: Copy app binary
  copy:
    src: ./build/myapp           # Локальный файл
    dest: /opt/myapp/bin/myapp   # Путь на сервере
    owner: deploy
    group: deploy
    mode: '0755'</code></pre>

<h3>template — Jinja2 шаблоны</h3>
<pre><code>- name: Render Nginx config
  template:
    src: nginx.conf.j2     # Шаблон с переменными {{ var }}
    dest: /etc/nginx/nginx.conf
    validate: "nginx -t -c %s"    # Проверить перед применением</code></pre>

<h3>service / systemd — управление сервисами</h3>
<pre><code>- name: Start and enable app
  systemd:
    name: myapp
    state: started       # started/stopped/restarted/reloaded
    enabled: yes         # Автозапуск при boot
    daemon_reload: yes   # systemctl daemon-reload</code></pre>

<h3>file — файлы и директории</h3>
<pre><code>- name: Create app directory
  file:
    path: /opt/myapp
    state: directory      # directory/file/link/absent
    owner: deploy
    mode: '0755'</code></pre>

<h2>Идемпотентность — глубокое погружение</h2>
<p>Каждый модуль Ansible проверяет текущее состояние перед действием:</p>
<pre><code># Первый запуск:
# apt: nginx не установлен → устанавливает → CHANGED

# Второй запуск:
# apt: nginx уже установлен, версия актуальна → OK (ничего не делает)

# Третий запуск после удаления nginx:
# apt: nginx не установлен → устанавливает → CHANGED</code></pre>

<p><strong>Антипаттерн — потеря идемпотентности:</strong></p>
<pre><code># ПЛОХО — shell/command НЕ идемпотентны!
- name: Create user
  shell: useradd deploy    # Упадёт при повторном запуске!

# ХОРОШО — модуль user идемпотентен
- name: Create user
  user:
    name: deploy
    state: present         # Не создаст повторно</code></pre>

<h2>Порядок выполнения</h2>
<ol>
<li><strong>pre_tasks</strong> — задачи до ролей</li>
<li><strong>roles</strong> — подключённые роли</li>
<li><strong>tasks</strong> — основные задачи</li>
<li><strong>post_tasks</strong> — задачи после</li>
<li><strong>handlers</strong> — только если были notify (в конце play)</li>
</ol>

<h2>Запуск Playbook</h2>
<pre><code># Базовый запуск
ansible-playbook -i inventory.ini deploy.yml

# Dry-run (проверка без изменений)
ansible-playbook -i inventory.ini deploy.yml --check

# Verbose (отладка)
ansible-playbook -i inventory.ini deploy.yml -vvv

# Ограничить хосты
ansible-playbook -i inventory.ini deploy.yml --limit web1.example.com

# Передать переменную
ansible-playbook -i inventory.ini deploy.yml -e "app_version=2.0.0"</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Когда выполняются handlers в Ansible playbook?",
				Options:     []string{"Сразу после задачи с notify", "В конце каждой задачи", "Один раз в конце play, и только если были вызваны через notify", "Перед всеми задачами"},
				Correct:     2,
				Explanation: "Handlers выполняются в конце play (не сразу после notify). Даже если notify вызван 5 раз одним handler — он выполнится один раз. Это оптимизация: перезапуск сервиса происходит один раз после всех изменений конфигов.",
			},
			{
				Question:    "Почему модуль shell/command считается неидемпотентным?",
				Options:     []string{"Он медленный", "Ansible не знает что именно делает команда и не может проверить нужно ли её выполнять — он всегда показывает changed", "Он не работает на Linux", "Он требует root"},
				Correct:     1,
				Explanation: "Shell/command — чёрный ящик для Ansible. Модуль apt знает: 'пакет уже есть? → ok'. Shell не знает что делает команда внутри, поэтому всегда changed. Используй creates/removes параметры или специализированные модули.",
			},
			{
				Question:    "Что делает флаг --check при запуске playbook?",
				Options:     []string{"Проверяет синтаксис YAML", "Выполняет dry-run: показывает что изменится, но не применяет изменения", "Проверяет SSH-подключение", "Запускает тесты"},
				Correct:     1,
				Explanation: "--check (dry-run) симулирует выполнение. Ansible подключается к серверам и проверяет текущее состояние, но не вносит изменений. Полезно для аудита: 'что изменит этот playbook?'. Не все модули поддерживают check mode.",
			},
			{
				Question:    "В каком порядке выполняются секции playbook?",
				Options:     []string{"tasks → roles → handlers", "roles → tasks → pre_tasks → handlers", "pre_tasks → roles → tasks → post_tasks → handlers", "handlers → tasks → roles"},
				Correct:     2,
				Explanation: "Порядок: pre_tasks → roles → tasks → post_tasks → handlers. Это позволяет подготовить окружение (pre_tasks), применить роли, выполнить специфичные задачи, и обработчики сработают в самом конце.",
			},
		},

		Tasks: []T{
			{
				Title:      "YAML Playbook Validator",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "strings.Count(s, substr)", Definition: "Считает количество вхождений подстроки в строку."},
					{Term: "strings.TrimSpace(s)", Definition: "Убирает пробелы и переводы строк с обоих концов строки."},
					{Term: "strings.Contains(s, substr)", Definition: "Проверяет содержит ли строка подстроку."},
				},
				Description: `<p>Проверь структуру упрощённого Ansible playbook (текстовый формат).</p>
<p><strong>Формат входа:</strong> строки playbook. Проверить:</p>
<ol>
<li>Есть ли поле <code>hosts:</code></li>
<li>Есть ли секция <code>tasks:</code></li>
<li>Каждая задача имеет <code>name:</code></li>
</ol>
<p><strong>Вывод:</strong></p>
<ul>
<li><code>VALID</code> — если все проверки пройдены</li>
<li><code>ERROR: missing hosts</code> — нет hosts</li>
<li><code>ERROR: missing tasks</code> — нет tasks</li>
<li><code>ERROR: task without name</code> — задача без name (строка с модулем без предшествующего name)</li>
</ul>
<p><em>Примечание:</em> задача определяется как строка с отступом содержащая <code>- name:</code> или модуль (apt:, copy:, service:, template:, file:, shell:, command:). Если модуль встречается без предшествующего <code>- name:</code> — ошибка.</p>`,
				Hints: `<p>Пройди по строкам: отслеживай наличие hosts и tasks. После tasks: каждая строка с "- name:" начинает задачу. Строка с модулем (apt:/copy:/service: и т.д.) без предшествующего name — ошибка.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	hasHosts := false
	hasTasks := false
	inTasks := false
	hasName := false
	modules := []string{"apt:", "copy:", "service:", "template:", "file:", "shell:", "command:", "systemd:"}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "hosts:") {
			hasHosts = true
		}
		if strings.HasPrefix(trimmed, "tasks:") {
			hasTasks = true
			inTasks = true
			continue
		}
		if inTasks {
			if strings.HasPrefix(trimmed, "- name:") {
				hasName = true
				continue
			}
			for _, mod := range modules {
				if strings.HasPrefix(trimmed, mod) || strings.Contains(trimmed, " "+mod) {
					if !hasName {
						fmt.Println("ERROR: task without name")
						return
					}
					hasName = false
					break
				}
			}
		}
	}

	if !hasHosts {
		fmt.Println("ERROR: missing hosts")
	} else if !hasTasks {
		fmt.Println("ERROR: missing tasks")
	} else {
		fmt.Println("VALID")
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
	hasHosts := false
	hasTasks := false
	inTasks := false
	hasName := false
	modules := []string{"apt:", "copy:", "service:", "template:", "file:", "shell:", "command:", "systemd:"}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		// TODO: проверь наличие hosts:, tasks:
		// В секции tasks: отслеживай - name: и модули
		_ = trimmed
		_ = modules
	}

	// TODO: выведи результат проверки
	_ = hasHosts
	_ = hasTasks
	_ = inTasks
	_ = hasName
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "- name: Deploy app\n  hosts: webservers\n  tasks:\n    - name: Install nginx\n      apt:\n        name: nginx\n", ExpectedOutput: "VALID"},
					{Input: "- name: Deploy\n  tasks:\n    - name: Install\n      apt:\n        name: nginx\n", ExpectedOutput: "ERROR: missing hosts"},
					{Input: "- name: Deploy\n  hosts: all\n  become: yes\n", ExpectedOutput: "ERROR: missing tasks"},
				},
			},
			{
				Title:      "Генератор Playbook из спецификации",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "strings.Repeat(s, count)", Definition: "Повторяет строку count раз. Repeat(\"  \", 2) → \"    \" (4 пробела)."},
					{Term: "fmt.Fprintf(w, format, ...)", Definition: "Форматированная запись в io.Writer (os.Stdout)."},
				},
				Description: `<p>По спецификации сгенерируй YAML playbook.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка: <code>hosts_value</code></li>
<li>Вторая строка: число N (количество задач)</li>
<li>Следующие N строк в формате: <code>module_name param_key param_value task_description</code></li>
</ol>
<p><strong>Формат выхода:</strong> YAML playbook (2 пробела для отступов).</p>
<p><em>Пример входа:</em></p>
<pre><code>webservers
2
apt name=nginx Install Nginx
service name=nginx,state=started Start Nginx</code></pre>
<p><em>Выход:</em></p>
<pre><code>---
- hosts: webservers
  become: yes
  tasks:
    - name: Install Nginx
      apt:
        name: nginx
    - name: Start Nginx
      service:
        name: nginx
        state: started</code></pre>
<p><em>Параметры через запятую: key=val,key2=val2</em></p>`,
				Hints: `<p>Разбери каждую строку задачи: первое слово — модуль, второе — параметры (key=val через запятую), остальное — описание (name). Выведи YAML с правильными отступами.</p>`,
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
	hosts := strings.TrimSpace(scanner.Text())

	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	fmt.Println("---")
	fmt.Printf("- hosts: %s\n", hosts)
	fmt.Println("  become: yes")
	fmt.Println("  tasks:")

	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), " ", 3)
		module := parts[0]
		params := parts[1]
		taskName := parts[2]

		fmt.Printf("    - name: %s\n", taskName)
		fmt.Printf("      %s:\n", module)

		for _, p := range strings.Split(params, ",") {
			kv := strings.SplitN(p, "=", 2)
			fmt.Printf("        %s: %s\n", kv[0], kv[1])
		}
	}
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
	hosts := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	// TODO: выведи YAML playbook
	// ---
	// - hosts: <hosts>
	//   become: yes
	//   tasks:
	//     - name: <task_name>
	//       <module>:
	//         <key>: <value>
	_ = hosts
	_ = n
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "webservers\n2\napt name=nginx Install Nginx\nservice name=nginx,state=started Start Nginx\n", ExpectedOutput: "---\n- hosts: webservers\n  become: yes\n  tasks:\n    - name: Install Nginx\n      apt:\n        name: nginx\n    - name: Start Nginx\n      service:\n        name: nginx\n        state: started"},
					{Input: "all\n1\ncopy src=/app,dest=/opt/app Deploy binary\n", ExpectedOutput: "---\n- hosts: all\n  become: yes\n  tasks:\n    - name: Deploy binary\n      copy:\n        src: /app\n        dest: /opt/app"},
				},
			},
		},
	}
}

// ── Урок 3: Переменные и шаблоны ───────────────────────────────

func lesson_ansible_variables() L {
	return L{
		Slug: "ansible-variables", Title: "Переменные и шаблоны Jinja2", Order: 3,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Переменные и шаблоны Jinja2</h1>

<h2>Переменные в Ansible</h2>
<p>Переменные делают playbooks гибкими и переиспользуемыми. Вместо хардкода значений — параметризация:</p>

<pre><code># ПЛОХО — хардкод
- name: Install app
  copy:
    src: ./build/myapp-1.2.3
    dest: /opt/myapp/bin/myapp

# ХОРОШО — переменная
- name: Install app
  copy:
    src: "./build/myapp-{{ app_version }}"
    dest: /opt/myapp/bin/myapp</code></pre>

<h2>Где определять переменные</h2>

<h3>1. В playbook (vars:)</h3>
<pre><code>- hosts: webservers
  vars:
    http_port: 80
    app_user: deploy
    packages:
      - nginx
      - curl
  tasks:
    - name: Install packages
      apt:
        name: "{{ item }}"
        state: present
      loop: "{{ packages }}"</code></pre>

<h3>2. В отдельном файле (vars_files:)</h3>
<pre><code># vars/production.yml
---
db_host: db.prod.internal
db_port: 5432
db_name: myapp_prod

# playbook.yml
- hosts: webservers
  vars_files:
    - vars/production.yml</code></pre>

<h3>3. В inventory</h3>
<pre><code># group_vars/webservers.yml — для группы
http_port: 80

# host_vars/web1.example.com.yml — для конкретного хоста
nginx_worker_processes: 4</code></pre>

<h3>4. Через командную строку (-e)</h3>
<pre><code>ansible-playbook deploy.yml -e "app_version=2.0.0 env=staging"</code></pre>

<h2>Приоритет переменных (от низшего к высшему)</h2>
<ol>
<li>Role defaults (defaults/main.yml)</li>
<li>Inventory group_vars</li>
<li>Inventory host_vars</li>
<li>Play vars</li>
<li>Play vars_files</li>
<li>Role vars (vars/main.yml)</li>
<li>Task vars (set_fact, register)</li>
<li>Extra vars (-e) — <strong>всегда побеждают!</strong></li>
</ol>

<h2>Facts — автоматические переменные</h2>
<p>Ansible автоматически собирает информацию о хостах (facts):</p>
<pre><code># Посмотреть все facts
ansible web1 -m setup

# Использование в playbook
- name: Show OS info
  debug:
    msg: "OS: {{ ansible_distribution }} {{ ansible_distribution_version }}"
    # → OS: Ubuntu 22.04

# Полезные facts:
# ansible_hostname          → web1
# ansible_default_ipv4.address → 10.0.0.5
# ansible_memtotal_mb       → 4096
# ansible_processor_cores   → 2
# ansible_os_family         → Debian</code></pre>

<h2>register — сохранение результата</h2>
<pre><code>- name: Check if app is running
  shell: pgrep -f myapp
  register: app_status
  ignore_errors: yes    # Не падать если процесс не найден

- name: Start app if not running
  systemd:
    name: myapp
    state: started
  when: app_status.rc != 0    # rc = return code</code></pre>

<h2>Jinja2 шаблоны</h2>
<p>Ansible использует Jinja2 — мощный шаблонизатор из Python мира:</p>

<h3>Базовый синтаксис</h3>
<pre><code># templates/nginx.conf.j2

server {
    listen {{ http_port }};                    # Переменная
    server_name {{ ansible_hostname }};        # Fact

    {% if ssl_enabled %}                       # Условие
    listen 443 ssl;
    ssl_certificate /etc/ssl/{{ domain }}.crt;
    {% endif %}

    {% for upstream in app_servers %}          # Цикл
    upstream backend {
        server {{ upstream }}:{{ app_port }};
    }
    {% endfor %}
}</code></pre>

<h3>Фильтры Jinja2</h3>
<pre><code>{{ name | upper }}              → MYAPP
{{ name | lower }}              → myapp
{{ list | join(', ') }}         → "a, b, c"
{{ value | default('none') }}   → значение или 'none'
{{ path | basename }}           → filename.txt
{{ dict | to_nice_yaml }}       → YAML строка
{{ secret | hash('sha256') }}   → хеш</code></pre>

<h2>Ansible Vault — шифрование секретов</h2>
<pre><code># Зашифровать файл с переменными
ansible-vault encrypt vars/secrets.yml

# Содержимое secrets.yml (до шифрования):
db_password: "super-secret-pass"
api_key: "sk-123456"

# После шифрования — нечитаемый текст
$ANSIBLE_VAULT;1.1;AES256
3766...encrypted...data

# Запуск playbook с vault
ansible-playbook deploy.yml --ask-vault-pass
# или
ansible-playbook deploy.yml --vault-password-file ~/.vault_pass</code></pre>

<p><strong>Важно:</strong> никогда не коммить незашифрованные секреты! Vault-файлы безопасно хранить в Git.</p>`,

		Quiz: []Q{
			{
				Question:    "Какой приоритет у extra vars (-e) в Ansible?",
				Options:     []string{"Самый низкий", "Средний (после play vars)", "Самый высокий — extra vars всегда побеждают все остальные источники", "Зависит от порядка определения"},
				Correct:     2,
				Explanation: "Extra vars (-e при запуске) имеют наивысший приоритет — 22 из 22 в полной иерархии Ansible. Это позволяет переопределить любую переменную при запуске, что полезно для CI/CD и hot-fix.",
			},
			{
				Question:    "Что такое Ansible Facts?",
				Options:     []string{"Переменные определённые пользователем", "Автоматически собранная информация о хосте (ОС, IP, память, CPU) через модуль setup", "Результаты тестов", "Логи выполнения"},
				Correct:     1,
				Explanation: "Facts — информация собранная через модуль setup: ОС, IP-адреса, RAM, CPU, диски, сетевые интерфейсы. Собираются автоматически в начале play (можно отключить: gather_facts: no). Доступны как ansible_* переменные.",
			},
			{
				Question:    "Для чего нужен Ansible Vault?",
				Options:     []string{"Для бэкапа серверов", "Для шифрования файлов с секретами (пароли, ключи), чтобы безопасно хранить их в Git", "Для хранения логов", "Для ускорения playbooks"},
				Correct:     1,
				Explanation: "Vault шифрует файлы AES-256. Позволяет хранить secrets.yml в Git без страха утечки. При запуске playbook расшифровывает на лету через --ask-vault-pass или --vault-password-file.",
			},
		},

		Tasks: []T{
			{
				Title:      "Механизм приоритета переменных",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "map[string]string", Definition: "Словарь строка→строка в Go. Поиск за O(1)."},
					{Term: "strings.SplitN(s, sep, n)", Definition: "Разбивает строку на максимум n частей."},
				},
				Description: `<p>Реализуй упрощённый механизм приоритета переменных Ansible.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Строка с числом N — количество источников (от низшего приоритета к высшему)</li>
<li>N блоков, каждый начинается с <code>SOURCE source_name</code>, затем строки <code>key=value</code>, блок заканчивается пустой строкой</li>
<li>Строка <code>RESOLVE</code></li>
<li>Строки с именами переменных для разрешения</li>
</ol>
<p><strong>Вывод:</strong> для каждой переменной: <code>key=value (from source_name)</code> или <code>key=UNDEFINED</code></p>`,
				Hints: `<p>Обрабатывай источники по порядку. Каждый следующий перезаписывает предыдущий (более высокий приоритет). Сохраняй откуда пришло значение.</p>`,
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
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	values := make(map[string]string)
	sources := make(map[string]string)

	for i := 0; i < n; i++ {
		scanner.Scan()
		srcLine := strings.TrimSpace(scanner.Text())
		srcName := strings.TrimPrefix(srcLine, "SOURCE ")

		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				break
			}
			parts := strings.SplitN(line, "=", 2)
			values[parts[0]] = parts[1]
			sources[parts[0]] = srcName
		}
	}

	scanner.Scan() // RESOLVE

	for scanner.Scan() {
		key := strings.TrimSpace(scanner.Text())
		if key == "" {
			continue
		}
		if val, ok := values[key]; ok {
			fmt.Printf("%s=%s (from %s)\n", key, val, sources[key])
		} else {
			fmt.Printf("%s=UNDEFINED\n", key)
		}
	}
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
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	values := make(map[string]string)
	sources := make(map[string]string)

	// TODO: обработай N источников
	// Каждый следующий перезаписывает предыдущий (выше приоритет)
	_ = n

	// TODO: после RESOLVE выведи значения переменных
	_ = values
	_ = sources
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "3\nSOURCE defaults\nport=80\nhost=localhost\n\nSOURCE group_vars\nport=8080\ndb_host=db.local\n\nSOURCE extra_vars\nport=9090\n\nRESOLVE\nport\nhost\ndb_host\nmissing\n", ExpectedOutput: "port=9090 (from extra_vars)\nhost=localhost (from defaults)\ndb_host=db.local (from group_vars)\nmissing=UNDEFINED"},
				},
			},
			{
				Title:      "Jinja2 Template Renderer",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "strings.ReplaceAll(s, old, new)", Definition: "Заменяет все вхождения подстроки."},
					{Term: "strings.Index(s, substr)", Definition: "Возвращает индекс первого вхождения или -1."},
				},
				Description: `<p>Реализуй упрощённый рендерер Jinja2-шаблонов.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка — число N (количество переменных)</li>
<li>N строк: <code>key=value</code></li>
<li>Строка <code>TEMPLATE</code></li>
<li>Остальные строки — шаблон</li>
</ol>
<p><strong>Правила:</strong> замени все <code>{{ key }}</code> на соответствующее значение. Пробелы внутри скобок игнорируй.</p>
<p><em>Вход:</em></p>
<pre><code>2
port=8080
host=myserver
TEMPLATE
server {{ host }} listens on {{ port }}</code></pre>
<p><em>Выход:</em> <code>server myserver listens on 8080</code></p>`,
				Hints: `<p>Собери переменные в map. Для каждой строки шаблона замени все вхождения {{ key }} (с пробелами и без) на значение. Используй strings.ReplaceAll для каждой переменной.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"strconv"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	vars := make(map[string]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), "=", 2)
		vars[parts[0]] = parts[1]
	}

	scanner.Scan() // TEMPLATE

	for scanner.Scan() {
		line := scanner.Text()
		for key, val := range vars {
			line = strings.ReplaceAll(line, "{{ "+key+" }}", val)
			line = strings.ReplaceAll(line, "{{"+key+"}}", val)
		}
		fmt.Println(line)
	}
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
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	vars := make(map[string]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), "=", 2)
		vars[parts[0]] = parts[1]
	}

	scanner.Scan() // TEMPLATE line

	// TODO: для каждой строки шаблона замени {{ key }} на значение
	for scanner.Scan() {
		line := scanner.Text()
		_ = line
		_ = vars
		fmt.Println("TODO")
	}
}`,
				TestCases: []TestCase{
					{Input: "2\nport=8080\nhost=myserver\nTEMPLATE\nserver {{ host }} listens on {{ port }}\n", ExpectedOutput: "server myserver listens on 8080"},
					{Input: "3\napp=golearn\nenv=production\nversion=1.0\nTEMPLATE\nDeploy {{ app }} v{{ version }} to {{ env }}\nDone.\n", ExpectedOutput: "Deploy golearn v1.0 to production\nDone."},
				},
			},
		},
	}
}

// ── Урок 4: Роли и Galaxy ───────────────────────────────────────

func lesson_ansible_roles() L {
	return L{
		Slug: "ansible-roles", Title: "Роли и Ansible Galaxy", Order: 4,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Роли и Ansible Galaxy</h1>

<h2>Зачем нужны роли?</h2>
<p>Когда playbook растёт — он становится неуправляемым. <strong>Роль</strong> — это способ организовать автоматизацию в переиспользуемые модули. Вместо одного огромного файла — структурированные компоненты.</p>

<pre><code># БЕЗ ролей — один большой playbook
# deploy.yml: 500 строк задач для nginx, postgres, app, monitoring...

# С РОЛЯМИ — чистая структура
- hosts: webservers
  roles:
    - nginx
    - app
    - monitoring</code></pre>

<h2>Структура роли</h2>
<pre><code>roles/
└── nginx/
    ├── defaults/
    │   └── main.yml        # Переменные по умолчанию (низший приоритет)
    ├── vars/
    │   └── main.yml        # Переменные роли (высокий приоритет)
    ├── tasks/
    │   └── main.yml        # Основные задачи
    ├── handlers/
    │   └── main.yml        # Обработчики (restart, reload)
    ├── templates/
    │   └── nginx.conf.j2   # Jinja2 шаблоны
    ├── files/
    │   └── index.html      # Статические файлы
    ├── meta/
    │   └── main.yml        # Зависимости роли, метаданные
    └── README.md           # Документация</code></pre>

<h3>defaults/main.yml — значения по умолчанию</h3>
<pre><code># roles/nginx/defaults/main.yml
---
nginx_port: 80
nginx_worker_processes: auto
nginx_worker_connections: 1024
nginx_client_max_body_size: "10m"
nginx_ssl_enabled: false</code></pre>
<p>Defaults — самый низкий приоритет. Любой другой источник переменных их перезапишет. Это "безопасные значения из коробки".</p>

<h3>tasks/main.yml — задачи</h3>
<pre><code># roles/nginx/tasks/main.yml
---
- name: Install Nginx
  apt:
    name: nginx
    state: present
    update_cache: yes

- name: Configure Nginx
  template:
    src: nginx.conf.j2
    dest: /etc/nginx/nginx.conf
  notify: Reload Nginx

- name: Enable and start Nginx
  systemd:
    name: nginx
    state: started
    enabled: yes</code></pre>

<h3>handlers/main.yml</h3>
<pre><code># roles/nginx/handlers/main.yml
---
- name: Reload Nginx
  systemd:
    name: nginx
    state: reloaded

- name: Restart Nginx
  systemd:
    name: nginx
    state: restarted</code></pre>

<h3>meta/main.yml — зависимости</h3>
<pre><code># roles/app/meta/main.yml
---
dependencies:
  - role: nginx
    vars:
      nginx_port: 8080
  - role: certbot
    when: ssl_enabled</code></pre>

<h2>Использование ролей в Playbook</h2>
<pre><code># Классический способ
- hosts: webservers
  roles:
    - nginx
    - { role: app, app_version: "2.0.0" }
    - { role: ssl, when: env == "production" }

# Через include_role (динамически)
- hosts: webservers
  tasks:
    - name: Deploy app
      include_role:
        name: app
      vars:
        app_version: "{{ version }}"

# Через import_role (статически — при парсинге)
- hosts: webservers
  tasks:
    - import_role:
        name: common</code></pre>

<h2>Ansible Galaxy</h2>
<p><strong>Ansible Galaxy</strong> — публичный репозиторий ролей (аналог npm/PyPI для Ansible).</p>

<pre><code># Поиск роли
ansible-galaxy search nginx

# Установка роли
ansible-galaxy install geerlingguy.nginx

# Файл зависимостей — requirements.yml
---
roles:
  - name: geerlingguy.nginx
    version: "3.1.0"
  - name: geerlingguy.postgresql
    version: "3.4.0"
  - src: https://github.com/user/role.git
    name: custom_role
    version: main

# Установить все зависимости
ansible-galaxy install -r requirements.yml</code></pre>

<h2>Создание своей роли</h2>
<pre><code># Scaffold роли
ansible-galaxy init roles/myapp

# Создаст структуру:
roles/myapp/
├── defaults/main.yml
├── files/
├── handlers/main.yml
├── meta/main.yml
├── tasks/main.yml
├── templates/
├── tests/
│   ├── inventory
│   └── test.yml
└── vars/main.yml</code></pre>

<h2>Best Practices для ролей</h2>
<ul>
<li><strong>Одна ответственность</strong> — роль nginx не должна настраивать PostgreSQL</li>
<li><strong>Defaults для всего</strong> — все переменные в defaults с разумными значениями</li>
<li><strong>Документация</strong> — README с описанием переменных и примерами</li>
<li><strong>Идемпотентность</strong> — повторный запуск безопасен</li>
<li><strong>Тесты</strong> — molecule для тестирования ролей в контейнерах</li>
<li><strong>Версионирование</strong> — semver тегами в Git</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "В чём разница между defaults/ и vars/ в роли?",
				Options:     []string{"Никакой разницы", "defaults/ имеет самый низкий приоритет (легко переопределить), vars/ — высокий приоритет (трудно переопределить)", "vars/ для строк, defaults/ для чисел", "defaults/ запускается первым"},
				Correct:     1,
				Explanation: "defaults/main.yml — приоритет 2 (самый низкий), любой group_vars/play vars их переопределит. vars/main.yml — приоритет 18, переопределить можно только через set_fact, include_vars, или extra_vars. Используй defaults для настраиваемых параметров.",
			},
			{
				Question:    "Что делает meta/main.yml в роли?",
				Options:     []string{"Определяет задачи", "Описывает зависимости роли — какие другие роли должны быть выполнены до неё", "Хранит секреты", "Определяет хосты"},
				Correct:     1,
				Explanation: "meta/main.yml — метаданные: зависимости (dependencies), поддерживаемые платформы, автор, лицензия. Dependencies выполняются ДО основных задач роли. Роль app с dependency: nginx гарантирует что nginx установлен.",
			},
			{
				Question:    "Чем import_role отличается от include_role?",
				Options:     []string{"Ничем", "import_role — статический (обрабатывается при парсинге playbook), include_role — динамический (обрабатывается при выполнении, поддерживает loops и when)", "import_role быстрее", "include_role устарел"},
				Correct:     1,
				Explanation: "import = статический (pre-processing): все задачи видны в --list-tasks, нельзя в loop. include = динамический (runtime): поддерживает loop, when на весь блок, переменные определённые ранее. Правило: import для предсказуемости, include для динамики.",
			},
		},

		Tasks: []T{
			{
				Title:      "Валидатор структуры роли",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "map[string]bool", Definition: "Используется как множество для проверки наличия элементов."},
					{Term: "sort.Strings(slice)", Definition: "Сортирует срез строк по алфавиту in-place."},
				},
				Description: `<p>Проверь что роль Ansible содержит все обязательные директории/файлы.</p>
<p><strong>Обязательные элементы:</strong> tasks/main.yml, defaults/main.yml, handlers/main.yml, meta/main.yml</p>
<p><strong>Формат входа:</strong> строки с путями файлов роли (относительно корня роли).</p>
<p><strong>Вывод:</strong></p>
<ul>
<li><code>VALID</code> — если все обязательные элементы есть</li>
<li><code>MISSING: file1, file2</code> — отсортированный список недостающих (через запятую с пробелом)</li>
</ul>`,
				Hints: `<p>Создай множество обязательных файлов. Для каждого входного пути убирай из множества. Что осталось — MISSING.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

func main() {
	required := map[string]bool{
		"tasks/main.yml":    true,
		"defaults/main.yml": true,
		"handlers/main.yml": true,
		"meta/main.yml":     true,
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		delete(required, path)
	}

	if len(required) == 0 {
		fmt.Println("VALID")
	} else {
		var missing []string
		for k := range required {
			missing = append(missing, k)
		}
		sort.Strings(missing)
		fmt.Printf("MISSING: %s\n", strings.Join(missing, ", "))
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
	required := map[string]bool{
		"tasks/main.yml":    true,
		"defaults/main.yml": true,
		"handlers/main.yml": true,
		"meta/main.yml":     true,
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := strings.TrimSpace(scanner.Text())
		// TODO: если path есть в required — удали
		_ = path
	}

	// TODO: если required пуст → VALID, иначе → MISSING: ...
	_ = required
	_ = sort.Strings
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "tasks/main.yml\ndefaults/main.yml\nhandlers/main.yml\nmeta/main.yml\ntemplates/app.conf.j2\nREADME.md\n", ExpectedOutput: "VALID"},
					{Input: "tasks/main.yml\nREADME.md\ntemplates/nginx.conf.j2\n", ExpectedOutput: "MISSING: defaults/main.yml, handlers/main.yml, meta/main.yml"},
				},
			},
			{
				Title:      "Разрешение зависимостей ролей",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "Топологическая сортировка", Definition: "Алгоритм упорядочивания узлов графа так, чтобы зависимости шли перед зависимыми."},
				},
				Description: `<p>Определи порядок выполнения ролей с учётом зависимостей (meta/dependencies).</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Число N — количество ролей</li>
<li>N строк: <code>role_name:dep1,dep2</code> или <code>role_name:</code> (без зависимостей)</li>
</ol>
<p><strong>Вывод:</strong> роли в порядке выполнения (зависимости сначала). Если несколько ролей без зависимостей — по алфавиту.</p>
<p><em>Пример:</em></p>
<pre><code>3
app:nginx,common
nginx:common
common:</code></pre>
<p><em>Выход:</em> <code>common, nginx, app</code></p>`,
				Hints: `<p>Топологическая сортировка: начни с ролей без зависимостей, затем добавляй те, все зависимости которых уже добавлены. При равенстве — по алфавиту.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	deps := make(map[string][]string)
	var roles []string

	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), ":", 2)
		role := parts[0]
		roles = append(roles, role)
		if parts[1] != "" {
			deps[role] = strings.Split(parts[1], ",")
		}
	}

	resolved := make(map[string]bool)
	var order []string

	for len(order) < len(roles) {
		var ready []string
		for _, r := range roles {
			if resolved[r] {
				continue
			}
			allMet := true
			for _, d := range deps[r] {
				if !resolved[d] {
					allMet = false
					break
				}
			}
			if allMet {
				ready = append(ready, r)
			}
		}
		sort.Strings(ready)
		for _, r := range ready {
			resolved[r] = true
			order = append(order, r)
		}
	}

	fmt.Println(strings.Join(order, ", "))
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	deps := make(map[string][]string)
	var roles []string

	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), ":", 2)
		role := parts[0]
		roles = append(roles, role)
		if parts[1] != "" {
			deps[role] = strings.Split(parts[1], ",")
		}
	}

	// TODO: топологическая сортировка
	// На каждой итерации бери роли, все зависимости которых уже resolved
	// При равенстве — по алфавиту
	_ = sort.Strings
	_ = deps
	_ = roles
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "3\napp:nginx,common\nnginx:common\ncommon:\n", ExpectedOutput: "common, nginx, app"},
					{Input: "4\nmonitoring:app\napp:db,nginx\nnginx:\ndb:\n", ExpectedOutput: "db, nginx, app, monitoring"},
				},
			},
		},
	}
}

// ── Урок 5: Продвинутые паттерны ────────────────────────────────

func lesson_ansible_advanced() L {
	return L{
		Slug: "ansible-advanced", Title: "Продвинутые паттерны", Order: 5,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Продвинутые паттерны Ansible</h1>

<h2>Условия (when)</h2>
<p>Выполнение задачи по условию:</p>
<pre><code># Только на Ubuntu
- name: Install apt packages
  apt:
    name: nginx
  when: ansible_os_family == "Debian"

# Только если переменная определена
- name: Configure SSL
  template:
    src: ssl.conf.j2
    dest: /etc/nginx/ssl.conf
  when: ssl_cert_path is defined

# Комбинация условий
- name: Deploy to production
  copy:
    src: app
    dest: /opt/app
  when:
    - env == "production"
    - app_version is defined
    - deploy_enabled | bool

# На основе результата предыдущей задачи
- name: Check disk space
  shell: df -h / | tail -1 | awk '{print $5}' | tr -d '%'
  register: disk_usage

- name: Alert if disk full
  debug:
    msg: "WARNING: Disk usage {{ disk_usage.stdout }}%"
  when: disk_usage.stdout | int > 80</code></pre>

<h2>Циклы (loop)</h2>
<pre><code># Простой цикл
- name: Create users
  user:
    name: "{{ item }}"
    state: present
  loop:
    - deploy
    - monitoring
    - backup

# Цикл по словарям
- name: Create users with groups
  user:
    name: "{{ item.name }}"
    groups: "{{ item.groups }}"
  loop:
    - { name: deploy, groups: "sudo,docker" }
    - { name: monitoring, groups: "docker" }

# Цикл с index
- name: Create numbered configs
  template:
    src: worker.conf.j2
    dest: "/etc/workers/worker-{{ idx }}.conf"
  loop: "{{ workers }}"
  loop_control:
    index_var: idx
    label: "{{ item.name }}"    # Что показывать в логе</code></pre>

<h2>Делегация (delegate_to)</h2>
<pre><code># Выполнить на другом хосте
- name: Remove from load balancer
  uri:
    url: "http://lb.internal/api/remove"
    body: '{"host": "{{ inventory_hostname }}"}'
  delegate_to: localhost      # Выполнить на control node

# Типичный паттерн: rolling deploy
- name: Disable in LB
  uri:
    url: "http://{{ lb_host }}/disable/{{ inventory_hostname }}"
  delegate_to: "{{ lb_host }}"

- name: Deploy new version
  copy:
    src: app
    dest: /opt/app

- name: Enable in LB
  uri:
    url: "http://{{ lb_host }}/enable/{{ inventory_hostname }}"
  delegate_to: "{{ lb_host }}"</code></pre>

<h2>Обработка ошибок (block/rescue/always)</h2>
<pre><code>- name: Deploy with rollback
  block:
    - name: Deploy new version
      copy:
        src: "app-{{ new_version }}"
        dest: /opt/app/bin/app

    - name: Restart service
      systemd:
        name: myapp
        state: restarted

    - name: Health check
      uri:
        url: "http://localhost:8080/health"
        status_code: 200
      retries: 5
      delay: 3

  rescue:    # Выполняется только если block упал
    - name: Rollback to previous version
      copy:
        src: "app-{{ old_version }}"
        dest: /opt/app/bin/app

    - name: Restart with old version
      systemd:
        name: myapp
        state: restarted

    - name: Notify about failure
      slack:
        msg: "Deploy {{ new_version }} FAILED, rolled back to {{ old_version }}"

  always:    # Выполняется всегда
    - name: Cleanup temp files
      file:
        path: /tmp/deploy-artifacts
        state: absent</code></pre>

<h2>Асинхронные задачи</h2>
<pre><code># Долгая задача — не ждать
- name: Run database migration (long)
  shell: /opt/app/migrate.sh
  async: 3600      # Максимум 1 час
  poll: 0          # Не ждать (fire-and-forget)
  register: migration_job

# Потом проверить статус
- name: Wait for migration
  async_status:
    jid: "{{ migration_job.ansible_job_id }}"
  register: job_result
  until: job_result.finished
  retries: 60
  delay: 30</code></pre>

<h2>Serial — Rolling Deploys</h2>
<pre><code># Обновлять по 2 сервера за раз (из 10)
- hosts: webservers
  serial: 2          # или "25%" — четверть за раз
  max_fail_percentage: 25    # Стоп если >25% хостов упали

  tasks:
    - name: Disable in LB
      ...
    - name: Deploy
      ...
    - name: Health check
      ...
    - name: Enable in LB
      ...

# Сложная стратегия: сначала 1 (canary), потом остальные
- hosts: webservers
  serial:
    - 1        # Первый — canary
    - "100%"   # Остальные все</code></pre>

<h2>Стратегии выполнения</h2>
<pre><code># По умолчанию: linear — все хосты выполняют задачу, потом следующую
- hosts: all
  strategy: linear   # default

# Free — каждый хост идёт с максимальной скоростью
- hosts: all
  strategy: free     # Не ждут друг друга

# Mitogen — ускорение через persistent connection (плагин)
[defaults]
strategy = mitogen_linear</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что происходит в секции rescue блока block/rescue/always?",
				Options:     []string{"Выполняется всегда", "Выполняется только если задачи в block завершились с ошибкой — аналог try/catch", "Выполняется перед block", "Выполняется параллельно с block"},
				Correct:     1,
				Explanation: "block/rescue/always — аналог try/catch/finally. block = try (основной путь), rescue = catch (выполняется только при ошибке), always = finally (выполняется всегда). Идеально для deploy с rollback.",
			},
			{
				Question:    "Что делает serial: 2 в playbook?",
				Options:     []string{"Запускает 2 задачи параллельно", "Обрабатывает хосты пакетами по 2 — rolling deploy, не все сразу", "Повторяет задачу 2 раза", "Ждёт 2 секунды"},
				Correct:     1,
				Explanation: "serial ограничивает количество хостов обрабатываемых одновременно. При serial: 2 из 10 серверов — сначала обновляются 2, потом следующие 2, и т.д. Это zero-downtime rolling deploy: часть серверов всегда доступна.",
			},
			{
				Question:    "Зачем нужен delegate_to: localhost?",
				Options:     []string{"Для ускорения", "Чтобы выполнить задачу на control node (например, вызвать API load balancer) в контексте текущего хоста", "Для отладки", "Для работы без SSH"},
				Correct:     1,
				Explanation: "delegate_to перенаправляет выполнение на другой хост. delegate_to: localhost выполняет задачу на управляющем узле (часто для API-вызовов, уведомлений). При этом переменные (inventory_hostname) остаются от исходного хоста.",
			},
			{
				Question:    "Что означает async: 3600, poll: 0?",
				Options:     []string{"Задача выполнится за 3600 секунд", "Fire-and-forget: запустить задачу асинхронно (макс. 1 час), не ждать завершения, продолжить playbook", "Повторять каждые 3600 секунд", "Таймаут подключения"},
				Correct:     1,
				Explanation: "async задаёт максимальное время выполнения. poll: 0 означает 'не ждать' — Ansible запускает задачу и идёт дальше. Результат можно проверить позже через async_status. Используется для долгих операций (миграции, бэкапы).",
			},
		},

		Tasks: []T{
			{
				Title:      "Симулятор Rolling Deploy",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "strconv.Atoi(s)", Definition: "Конвертирует строку в int."},
				},
				Description: `<p>Симулируй rolling deploy с параметром serial.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Первая строка: число serial (сколько серверов за раз)</li>
<li>Вторая строка: список серверов через запятую</li>
</ol>
<p><strong>Вывод:</strong> пакеты серверов (batch), по serial штук. Каждый пакет на отдельной строке.</p>
<p><em>Пример:</em></p>
<pre><code>2
web1,web2,web3,web4,web5</code></pre>
<p><em>Выход:</em></p>
<pre><code>batch 1: web1, web2
batch 2: web3, web4
batch 3: web5</code></pre>`,
				Hints: `<p>Разбей серверы на chunks размером serial. Последний chunk может быть меньше.</p>`,
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
	serial, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	scanner.Scan()
	servers := strings.Split(strings.TrimSpace(scanner.Text()), ",")

	batch := 1
	for i := 0; i < len(servers); i += serial {
		end := i + serial
		if end > len(servers) {
			end = len(servers)
		}
		fmt.Printf("batch %d: %s\n", batch, strings.Join(servers[i:end], ", "))
		batch++
	}
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
	serial, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	scanner.Scan()
	servers := strings.Split(strings.TrimSpace(scanner.Text()), ",")

	// TODO: разбей servers на пакеты по serial штук
	// Выведи: batch N: server1, server2
	_ = serial
	_ = servers
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "2\nweb1,web2,web3,web4,web5\n", ExpectedOutput: "batch 1: web1, web2\nbatch 2: web3, web4\nbatch 3: web5"},
					{Input: "3\napp1,app2,app3,app4,app5,app6\n", ExpectedOutput: "batch 1: app1, app2, app3\nbatch 2: app4, app5, app6"},
					{Input: "1\nserver1,server2,server3\n", ExpectedOutput: "batch 1: server1\nbatch 2: server2\nbatch 3: server3"},
				},
			},
			{
				Title:      "Block/Rescue симулятор",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "strings.HasPrefix(s, prefix)", Definition: "Проверяет начинается ли строка с указанного prefix."},
				},
				Description: `<p>Симулируй выполнение block/rescue/always.</p>
<p><strong>Формат входа:</strong></p>
<ul>
<li>Строки с задачами. Каждая: <code>section:task_name:result</code></li>
<li>section: block, rescue, always</li>
<li>result: ok или fail</li>
</ul>
<p><strong>Логика:</strong></p>
<ul>
<li>Выполняй задачи block по порядку. Если result=fail — переключись на rescue</li>
<li>Если block прошёл без ошибок — rescue пропускается</li>
<li>always выполняется всегда</li>
</ul>
<p><strong>Вывод:</strong> для каждой выполненной задачи: <code>TASK [task_name] => result</code></p>
<p>Невыполненные задачи (пропущенный rescue при успешном block, или оставшиеся block-задачи после fail) — не выводить.</p>`,
				Hints: `<p>Разбери задачи на три группы. Выполняй block пока нет fail. Если был fail — выполняй rescue. Always — всегда в конце.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	var blockTasks, rescueTasks, alwaysTasks []struct{ name, result string }

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		section, name, result := parts[0], parts[1], parts[2]
		task := struct{ name, result string }{name, result}
		switch section {
		case "block":
			blockTasks = append(blockTasks, task)
		case "rescue":
			rescueTasks = append(rescueTasks, task)
		case "always":
			alwaysTasks = append(alwaysTasks, task)
		}
	}

	blockFailed := false
	for _, t := range blockTasks {
		fmt.Printf("TASK [%s] => %s\n", t.name, t.result)
		if t.result == "fail" {
			blockFailed = true
			break
		}
	}

	if blockFailed {
		for _, t := range rescueTasks {
			fmt.Printf("TASK [%s] => %s\n", t.name, t.result)
		}
	}

	for _, t := range alwaysTasks {
		fmt.Printf("TASK [%s] => %s\n", t.name, t.result)
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

	type task struct{ name, result string }
	var blockTasks, rescueTasks, alwaysTasks []task

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		// TODO: распредели задачи по секциям
		_ = parts
	}

	// TODO: выполни block, при fail переключись на rescue, always — всегда
	_ = blockTasks
	_ = rescueTasks
	_ = alwaysTasks
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "block:Deploy app:ok\nblock:Restart:ok\nrescue:Rollback:ok\nalways:Cleanup:ok\n", ExpectedOutput: "TASK [Deploy app] => ok\nTASK [Restart] => ok\nTASK [Cleanup] => ok"},
					{Input: "block:Deploy app:ok\nblock:Health check:fail\nblock:Enable LB:ok\nrescue:Rollback:ok\nrescue:Notify:ok\nalways:Cleanup:ok\n", ExpectedOutput: "TASK [Deploy app] => ok\nTASK [Health check] => fail\nTASK [Rollback] => ok\nTASK [Notify] => ok\nTASK [Cleanup] => ok"},
				},
			},
			{
				Title:      "Обработка условий when",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "strconv.Atoi(s)", Definition: "Конвертирует строку в число. Возвращает (int, error)."},
					{Term: "strings.Contains(s, substr)", Definition: "Проверяет содержит ли строка подстроку."},
				},
				Description: `<p>Симулируй выполнение задач с условиями when.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Строка с переменными: <code>key=value,key2=value2</code></li>
<li>Остальные строки — задачи: <code>task_name|condition</code></li>
</ol>
<p><strong>Условия (упрощённые):</strong></p>
<ul>
<li><code>var == "value"</code> — равенство строк</li>
<li><code>var != "value"</code> — неравенство</li>
<li><code>var > N</code> — больше числа</li>
<li><code>var is defined</code> — переменная существует</li>
</ul>
<p><strong>Вывод:</strong> <code>RUN: task_name</code> или <code>SKIP: task_name</code></p>`,
				Hints: `<p>Разбери условие на оператор. Для == и != сравни строки. Для > преобразуй в числа. Для is defined проверь наличие ключа в map.</p>`,
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
	vars := make(map[string]string)
	for _, pair := range strings.Split(strings.TrimSpace(scanner.Text()), ",") {
		kv := strings.SplitN(pair, "=", 2)
		vars[kv[0]] = kv[1]
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		taskName := parts[0]
		cond := strings.TrimSpace(parts[1])

		result := evaluate(cond, vars)
		if result {
			fmt.Printf("RUN: %s\n", taskName)
		} else {
			fmt.Printf("SKIP: %s\n", taskName)
		}
	}
}

func evaluate(cond string, vars map[string]string) bool {
	if strings.Contains(cond, " is defined") {
		varName := strings.TrimSpace(strings.Split(cond, " is defined")[0])
		_, ok := vars[varName]
		return ok
	}
	if strings.Contains(cond, " == ") {
		parts := strings.SplitN(cond, " == ", 2)
		varName := strings.TrimSpace(parts[0])
		expected := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		return vars[varName] == expected
	}
	if strings.Contains(cond, " != ") {
		parts := strings.SplitN(cond, " != ", 2)
		varName := strings.TrimSpace(parts[0])
		expected := strings.Trim(strings.TrimSpace(parts[1]), "\"")
		return vars[varName] != expected
	}
	if strings.Contains(cond, " > ") {
		parts := strings.SplitN(cond, " > ", 2)
		varName := strings.TrimSpace(parts[0])
		threshold, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		val, _ := strconv.Atoi(vars[varName])
		return val > threshold
	}
	return false
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
	vars := make(map[string]string)
	for _, pair := range strings.Split(strings.TrimSpace(scanner.Text()), ",") {
		kv := strings.SplitN(pair, "=", 2)
		vars[kv[0]] = kv[1]
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "|", 2)
		taskName := parts[0]
		cond := strings.TrimSpace(parts[1])

		// TODO: вычисли условие cond и выведи RUN или SKIP
		_ = taskName
		_ = cond
		_ = vars
		_ = strconv.Atoi
	}
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "env=production,version=2,port=8080\nDeploy|env == \"production\"\nDebug|env == \"staging\"\nScale|port > 80\nSetup|missing is defined\n", ExpectedOutput: "RUN: Deploy\nSKIP: Debug\nRUN: Scale\nSKIP: Setup"},
					{Input: "os=ubuntu,memory=4096\nInstall apt|os == \"ubuntu\"\nInstall yum|os != \"ubuntu\"\nAdd swap|memory > 2048\nCheck var|os is defined\n", ExpectedOutput: "RUN: Install apt\nSKIP: Install yum\nRUN: Add swap\nRUN: Check var"},
				},
			},
		},
	}
}

// ── Урок 6: Real-World Playbooks ────────────────────────────────

func lesson_ansible_realworld() L {
	return L{
		Slug: "ansible-realworld", Title: "Real-World Playbooks", Order: 6,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Real-World Playbooks</h1>

<h2>Деплой Go-приложения</h2>
<p>Полный playbook для деплоя Go бинарника на production:</p>

<pre><code># deploy-go-app.yml
---
- name: Deploy Go Application
  hosts: app_servers
  become: yes
  serial: 1    # Rolling deploy по одному

  vars:
    app_name: myapp
    app_user: deploy
    app_port: 8080
    app_dir: /opt/{{ app_name }}
    binary_src: "./build/{{ app_name }}"

  pre_tasks:
    - name: Remove from load balancer
      uri:
        url: "http://{{ lb_host }}/api/disable/{{ inventory_hostname }}"
        method: POST
      delegate_to: localhost

  tasks:
    - name: Create app directory
      file:
        path: "{{ app_dir }}/bin"
        state: directory
        owner: "{{ app_user }}"
        mode: '0755'

    - name: Stop application
      systemd:
        name: "{{ app_name }}"
        state: stopped
      ignore_errors: yes    # Может не существовать при первом деплое

    - name: Deploy binary
      copy:
        src: "{{ binary_src }}"
        dest: "{{ app_dir }}/bin/{{ app_name }}"
        owner: "{{ app_user }}"
        mode: '0755'

    - name: Deploy environment file
      template:
        src: templates/app.env.j2
        dest: "{{ app_dir }}/.env"
        owner: "{{ app_user }}"
        mode: '0600'    # Только владелец читает (секреты!)

    - name: Deploy systemd unit
      template:
        src: templates/app.service.j2
        dest: "/etc/systemd/system/{{ app_name }}.service"
      notify: Reload systemd

    - name: Start application
      systemd:
        name: "{{ app_name }}"
        state: started
        enabled: yes

    - name: Health check
      uri:
        url: "http://localhost:{{ app_port }}/health"
        status_code: 200
      retries: 10
      delay: 3
      register: health

    - name: Verify health
      assert:
        that: health.status == 200
        fail_msg: "Health check failed!"

  post_tasks:
    - name: Enable in load balancer
      uri:
        url: "http://{{ lb_host }}/api/enable/{{ inventory_hostname }}"
        method: POST
      delegate_to: localhost

  handlers:
    - name: Reload systemd
      systemd:
        daemon_reload: yes</code></pre>

<h3>Шаблон systemd unit</h3>
<pre><code># templates/app.service.j2
[Unit]
Description={{ app_name }} service
After=network.target

[Service]
Type=simple
User={{ app_user }}
WorkingDirectory={{ app_dir }}
ExecStart={{ app_dir }}/bin/{{ app_name }}
EnvironmentFile={{ app_dir }}/.env
Restart=always
RestartSec=5

# Hardening
NoNewPrivileges=yes
ProtectSystem=strict
ReadWritePaths={{ app_dir }}/data

[Install]
WantedBy=multi-user.target</code></pre>

<h2>Настройка Nginx + SSL</h2>
<pre><code># nginx-ssl.yml
---
- name: Setup Nginx with SSL
  hosts: webservers
  become: yes

  vars:
    domain: app.example.com
    upstream_port: 8080
    ssl_email: admin@example.com

  roles:
    - nginx
    - certbot

  tasks:
    - name: Deploy Nginx site config
      template:
        src: templates/nginx-site.conf.j2
        dest: "/etc/nginx/sites-available/{{ domain }}"
      notify: Reload Nginx

    - name: Enable site
      file:
        src: "/etc/nginx/sites-available/{{ domain }}"
        dest: "/etc/nginx/sites-enabled/{{ domain }}"
        state: link
      notify: Reload Nginx

    - name: Remove default site
      file:
        path: /etc/nginx/sites-enabled/default
        state: absent
      notify: Reload Nginx

    - name: Obtain SSL certificate
      command: >
        certbot certonly --nginx
        -d {{ domain }}
        --email {{ ssl_email }}
        --agree-tos
        --non-interactive
      args:
        creates: "/etc/letsencrypt/live/{{ domain }}/fullchain.pem"

    - name: Setup auto-renewal
      cron:
        name: "certbot renewal"
        minute: "30"
        hour: "2"
        job: "certbot renew --quiet --post-hook 'systemctl reload nginx'"

  handlers:
    - name: Reload Nginx
      systemd:
        name: nginx
        state: reloaded</code></pre>

<h2>Провижининг PostgreSQL</h2>
<pre><code># postgres.yml
---
- name: Setup PostgreSQL
  hosts: dbservers
  become: yes

  vars:
    pg_version: "16"
    pg_databases:
      - { name: myapp_prod, owner: app_user }
      - { name: myapp_staging, owner: app_user }
    pg_users:
      - { name: app_user, password: "{{ vault_db_password }}" }
    pg_hba_entries:
      - { type: host, database: all, user: app_user, address: "10.0.0.0/24", method: md5 }

  tasks:
    - name: Install PostgreSQL
      apt:
        name:
          - "postgresql-{{ pg_version }}"
          - "postgresql-client-{{ pg_version }}"
          - python3-psycopg2
        state: present

    - name: Configure pg_hba.conf
      template:
        src: templates/pg_hba.conf.j2
        dest: "/etc/postgresql/{{ pg_version }}/main/pg_hba.conf"
      notify: Reload PostgreSQL

    - name: Configure postgresql.conf
      lineinfile:
        path: "/etc/postgresql/{{ pg_version }}/main/postgresql.conf"
        regexp: "^{{ item.key }}"
        line: "{{ item.key }} = {{ item.value }}"
      loop:
        - { key: listen_addresses, value: "'*'" }
        - { key: max_connections, value: "200" }
        - { key: shared_buffers, value: "'256MB'" }
      notify: Restart PostgreSQL

    - name: Create users
      postgresql_user:
        name: "{{ item.name }}"
        password: "{{ item.password }}"
        state: present
      loop: "{{ pg_users }}"
      become_user: postgres

    - name: Create databases
      postgresql_db:
        name: "{{ item.name }}"
        owner: "{{ item.owner }}"
        state: present
      loop: "{{ pg_databases }}"
      become_user: postgres

  handlers:
    - name: Reload PostgreSQL
      systemd:
        name: postgresql
        state: reloaded
    - name: Restart PostgreSQL
      systemd:
        name: postgresql
        state: restarted</code></pre>

<h2>Полный стек: роли + orchestration</h2>
<pre><code># site.yml — мастер-playbook
---
- name: Common setup (all servers)
  hosts: all
  roles:
    - common        # SSH keys, NTP, firewall, monitoring agent
    - security      # fail2ban, sysctl hardening

- name: Database tier
  hosts: dbservers
  roles:
    - postgresql
    - backup

- name: Application tier
  hosts: app_servers
  serial: 1
  roles:
    - golang_app
  post_tasks:
    - name: Verify deployment
      uri:
        url: "http://localhost:8080/health"

- name: Web tier
  hosts: webservers
  roles:
    - nginx
    - certbot

# Запуск всего стека:
# ansible-playbook -i production site.yml --vault-password-file .vault

# Только приложение:
# ansible-playbook -i production site.yml --tags app

# Только конкретный сервер:
# ansible-playbook -i production site.yml --limit app1.example.com</code></pre>

<h2>Best Practices для Production</h2>
<ul>
<li><strong>Всегда --check сначала</strong> — dry-run перед боевым запуском</li>
<li><strong>Serial для критических сервисов</strong> — rolling deploy, не все сразу</li>
<li><strong>Health check после deploy</strong> — убедиться что приложение живо</li>
<li><strong>Vault для секретов</strong> — никогда не хардкодь пароли</li>
<li><strong>Тегирование</strong> — tags позволяют запускать часть playbook</li>
<li><strong>Идемпотентность</strong> — избегай shell/command, используй модули</li>
<li><strong>Inventory per env</strong> — production, staging, development — разные файлы</li>
<li><strong>CI/CD интеграция</strong> — ansible-playbook в pipeline (lint → check → apply)</li>
</ul>`,

		Quiz: []Q{
			{
				Question:    "Почему в шаблоне systemd unit используется RestartSec=5?",
				Options:     []string{"Для красоты", "Чтобы при крэше приложения systemd не пытался перезапускать слишком часто (защита от restart loop)", "Для ускорения старта", "Это обязательный параметр"},
				Correct:     1,
				Explanation: "Без RestartSec systemd будет перезапускать мгновенно. Если приложение крэшит сразу при старте — получим бесконечный цикл перезапусков. RestartSec=5 даёт паузу и ограничивает rate. По умолчанию systemd остановит сервис после 5 неудачных попыток за 10 секунд.",
			},
			{
				Question:    "Зачем в деплое Go-приложения сначала убирают сервер из LB?",
				Options:     []string{"Для экономии трафика", "Чтобы новые запросы не шли на сервер пока идёт деплой — zero-downtime rolling deploy", "LB мешает SSH", "Для ускорения деплоя"},
				Correct:     1,
				Explanation: "Паттерн zero-downtime: 1) Убрать из LB (новые запросы не приходят). 2) Дождаться drain (текущие запросы завершатся). 3) Остановить → обновить → запустить. 4) Health check. 5) Вернуть в LB. Пользователи не видят downtime.",
			},
			{
				Question:    "Зачем для .env файла используется mode: '0600'?",
				Options:     []string{"Так быстрее читается", "Только владелец может читать файл — защита секретов (пароли, API-ключи) от других пользователей системы", "Это стандарт", "Для совместимости"},
				Correct:     1,
				Explanation: "0600 = rw------- (только владелец читает/пишет). .env содержит секреты (DB_PASSWORD, API_KEY). Без ограничений любой пользователь системы может прочитать файл. 0600 + правильный owner — минимальные привилегии.",
			},
			{
				Question:    "Что делает args: creates в задаче certbot?",
				Options:     []string{"Создаёт директорию", "Пропускает задачу если указанный файл уже существует — обеспечивает идемпотентность для shell/command", "Проверяет сертификат", "Генерирует конфиг"},
				Correct:     1,
				Explanation: "creates — параметр идемпотентности для shell/command модулей. 'creates: /etc/letsencrypt/live/domain/fullchain.pem' означает: если файл существует — задача пропускается (ok). Это превращает не-идемпотентный command в идемпотентный.",
			},
		},

		Tasks: []T{
			{
				Title:      "Генератор systemd unit",
				Difficulty: "medium",
				Glossary: []GlossaryItem{
					{Term: "fmt.Printf(format, args...)", Definition: "Форматированный вывод. %s для строк, %d для чисел."},
					{Term: "strings.SplitN(s, sep, n)", Definition: "Разбивает строку на максимум n частей."},
				},
				Description: `<p>Сгенерируй systemd unit файл из параметров.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li><code>app_name</code></li>
<li><code>app_user</code></li>
<li><code>exec_start_command</code></li>
<li><code>working_dir</code></li>
</ol>
<p><strong>Выход:</strong> systemd unit в формате:</p>
<pre><code>[Unit]
Description=app_name service
After=network.target

[Service]
Type=simple
User=app_user
WorkingDirectory=working_dir
ExecStart=exec_start_command
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target</code></pre>`,
				Hints: `<p>Просто прочитай 4 строки и выведи шаблон unit-файла, подставив значения.</p>`,
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
	appName := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	appUser := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	execStart := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	workDir := strings.TrimSpace(scanner.Text())

	fmt.Println("[Unit]")
	fmt.Printf("Description=%s service\n", appName)
	fmt.Println("After=network.target")
	fmt.Println()
	fmt.Println("[Service]")
	fmt.Println("Type=simple")
	fmt.Printf("User=%s\n", appUser)
	fmt.Printf("WorkingDirectory=%s\n", workDir)
	fmt.Printf("ExecStart=%s\n", execStart)
	fmt.Println("Restart=always")
	fmt.Println("RestartSec=5")
	fmt.Println()
	fmt.Println("[Install]")
	fmt.Println("WantedBy=multi-user.target")
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
	appName := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	appUser := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	execStart := strings.TrimSpace(scanner.Text())
	scanner.Scan()
	workDir := strings.TrimSpace(scanner.Text())

	// TODO: выведи systemd unit файл используя шаблон
	_ = appName
	_ = appUser
	_ = execStart
	_ = workDir
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "golearn\ndeploy\n/opt/golearn/bin/golearn\n/opt/golearn\n", ExpectedOutput: "[Unit]\nDescription=golearn service\nAfter=network.target\n\n[Service]\nType=simple\nUser=deploy\nWorkingDirectory=/opt/golearn\nExecStart=/opt/golearn/bin/golearn\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=multi-user.target"},
					{Input: "api\nwww\n/usr/local/bin/api serve\n/var/www/api\n", ExpectedOutput: "[Unit]\nDescription=api service\nAfter=network.target\n\n[Service]\nType=simple\nUser=www\nWorkingDirectory=/var/www/api\nExecStart=/usr/local/bin/api serve\nRestart=always\nRestartSec=5\n\n[Install]\nWantedBy=multi-user.target"},
				},
			},
			{
				Title:      "Full-Stack Deploy Planner",
				Difficulty: "hard",
				Glossary: []GlossaryItem{
					{Term: "strings.Split(s, sep)", Definition: "Разбивает строку по разделителю на срез строк."},
				},
				Description: `<p>Определи порядок деплоя для full-stack приложения по зависимостям тиров.</p>
<p><strong>Формат входа:</strong></p>
<ol>
<li>Число N — количество тиров (компонентов)</li>
<li>N строк: <code>tier_name:depends_on1,depends_on2</code> или <code>tier_name:</code></li>
<li>Строка <code>DEPLOY</code></li>
<li>Строки с именами тиров для деплоя</li>
</ol>
<p><strong>Вывод:</strong> порядок деплоя с учётом зависимостей. Каждый тир тянет свои зависимости (если они есть в списке на деплой). Формат: <code>Step N: tier_name</code></p>
<p><em>Примечание:</em> если зависимость тира не входит в список для деплоя — она пропускается. Тиры на одном уровне — по алфавиту.</p>`,
				Hints: `<p>Топологическая сортировка: на каждом шаге бери тиры, все зависимости которых уже задеплоены (или не в списке деплоя). При равенстве — по алфавиту.</p>`,
				Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	deps := make(map[string][]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), ":", 2)
		name := parts[0]
		if parts[1] != "" {
			deps[name] = strings.Split(parts[1], ",")
		} else {
			deps[name] = nil
		}
	}

	scanner.Scan() // DEPLOY

	toDeploy := make(map[string]bool)
	var deployList []string
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t != "" {
			toDeploy[t] = true
			deployList = append(deployList, t)
		}
	}

	deployed := make(map[string]bool)
	step := 1

	for len(deployed) < len(deployList) {
		var ready []string
		for _, t := range deployList {
			if deployed[t] {
				continue
			}
			allMet := true
			for _, d := range deps[t] {
				if toDeploy[d] && !deployed[d] {
					allMet = false
					break
				}
			}
			if allMet {
				ready = append(ready, t)
			}
		}
		sort.Strings(ready)
		for _, t := range ready {
			fmt.Printf("Step %d: %s\n", step, t)
			deployed[t] = true
			step++
		}
	}
}</code></pre>`,
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	scanner.Scan()
	n, _ := strconv.Atoi(strings.TrimSpace(scanner.Text()))

	deps := make(map[string][]string)
	for i := 0; i < n; i++ {
		scanner.Scan()
		parts := strings.SplitN(strings.TrimSpace(scanner.Text()), ":", 2)
		name := parts[0]
		if parts[1] != "" {
			deps[name] = strings.Split(parts[1], ",")
		}
	}

	scanner.Scan() // DEPLOY line

	toDeploy := make(map[string]bool)
	var deployList []string
	for scanner.Scan() {
		t := strings.TrimSpace(scanner.Text())
		if t != "" {
			toDeploy[t] = true
			deployList = append(deployList, t)
		}
	}

	// TODO: топологическая сортировка для деплоя
	_ = deps
	_ = toDeploy
	_ = deployList
	_ = sort.Strings
	fmt.Println("TODO")
}`,
				TestCases: []TestCase{
					{Input: "4\ndb:\nnginx:app\napp:db\nmonitoring:app\nDEPLOY\nnginx\napp\ndb\nmonitoring\n", ExpectedOutput: "Step 1: db\nStep 2: app\nStep 3: monitoring\nStep 4: nginx"},
					{Input: "3\nfrontend:backend\nbackend:database\ndatabase:\nDEPLOY\nfrontend\nbackend\n", ExpectedOutput: "Step 1: backend\nStep 2: frontend"},
				},
			},
		},
	}
}
