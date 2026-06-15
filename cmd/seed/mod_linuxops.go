package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Linux Fundamentals
// 10 уроков от обзора ОС до troubleshooting
// Track: devops, Difficulty: beginner
// ════════════════════════════════════════════════════════════════

func mod_linux() M {
	return M{
		Slug:          "linux-fundamentals",
		Title:         "Linux: Основные инструменты",
		Description:   "Всё что нужно знать о Linux для backend/devops разработчика: файловая система, процессы, systemd, сети, bash-скрипты, диагностика.",
		Order:         19,
		Track:         "devops",
		Difficulty:    "beginner",
		Prerequisites: []string{"basics"},
		Lessons: []L{
			linuxLesson01Overview(),
			linuxLesson02Filesystem(),
			linuxLesson03Permissions(),
			linuxLesson04Users(),
			linuxLesson05Processes(),
			linuxLesson06Systemd(),
			linuxLesson07Packages(),
			linuxLesson08Networking(),
			linuxLesson09Bash(),
			linuxLesson10Logs(),
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 1: Linux Overview
// ══════════════════════════════════════════════════════════════════

func linuxLesson01Overview() L {
	return L{
		Slug: "linux-overview", Title: "Что такое Linux и зачем он нужен", Order: 1,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Что такое Linux и зачем он нужен</h1>

<h2>Linux — это ядро, не ОС</h2>
<p>Когда говорят "Linux", обычно имеют в виду целую операционную систему. Но технически <strong>Linux</strong> — это только <strong>ядро</strong> (kernel). Ядро — посредник между программами и железом (CPU, RAM, диск, сеть).</p>

<p>Полная ОС = ядро Linux + системные утилиты (GNU) + пакетный менеджер + shell. Правильное название: <strong>GNU/Linux</strong>, но все говорят просто "Linux".</p>

<h2>Kernel vs Userspace</h2>
<pre><code>┌─────────────────────────────────────────┐
│         Userspace (пользователь)        │
│  bash, nginx, postgres, go-программы    │
├─────────────────────────────────────────┤
│         System Calls (syscalls)         │
│  open(), read(), write(), fork(), ...   │
├─────────────────────────────────────────┤
│         Kernel (ядро Linux)             │
│  управление памятью, процессами,        │
│  файловыми системами, сетью, драйверы  │
├─────────────────────────────────────────┤
│         Hardware (железо)               │
│  CPU, RAM, SSD, NIC                     │
└─────────────────────────────────────────┘</code></pre>

<p><strong>Userspace</strong> — всё что работает вне ядра. Программы не могут напрямую общаться с железом — они просят ядро через <strong>системные вызовы</strong> (syscalls). Например, чтобы прочитать файл, программа вызывает <code>read()</code>, а ядро обращается к диску.</p>

<h2>Почему серверы используют Linux?</h2>
<ul>
<li><strong>Бесплатный и открытый</strong> — нет лицензий, можно модифицировать</li>
<li><strong>Стабильный</strong> — серверы работают годами без перезагрузки</li>
<li><strong>Безопасный</strong> — быстрые патчи, разделение привилегий</li>
<li><strong>Легковесный</strong> — сервер может работать с 512MB RAM без GUI</li>
<li><strong>Автоматизируемый</strong> — всё управляется через текстовые конфиги и CLI</li>
<li><strong>Docker/K8s</strong> — контейнеры работают только на Linux (даже Docker Desktop на Mac крутит Linux VM)</li>
</ul>

<p><strong>Факт:</strong> 96%+ серверов в мире работают на Linux. Все крупные облака (AWS, GCP, Azure) используют Linux как основу.</p>

<h2>Дистрибутивы (distros)</h2>
<p>Дистрибутив = ядро Linux + набор программ + пакетный менеджер + философия.</p>

<table>
<tr><th>Семейство</th><th>Дистрибутивы</th><th>Пакетный менеджер</th><th>Где используется</th></tr>
<tr><td>Debian-based</td><td>Ubuntu, Debian</td><td>apt</td><td>Серверы, рабочие станции</td></tr>
<tr><td>RHEL-based</td><td>Rocky, AlmaLinux, CentOS</td><td>dnf/yum</td><td>Enterprise</td></tr>
<tr><td>Alpine</td><td>Alpine Linux</td><td>apk</td><td>Docker-контейнеры (5MB base)</td></tr>
<tr><td>Arch-based</td><td>Arch, Manjaro</td><td>pacman</td><td>Рабочие станции</td></tr>
</table>

<p><strong>Для DevOps важны:</strong> Ubuntu Server (самый популярный на серверах), Alpine (Docker), Rocky/Alma (enterprise замена CentOS).</p>

<h2>Всё — файл</h2>
<p>Философия Linux: <strong>"Everything is a file"</strong>. Устройства, процессы, сокеты — всё представлено как файл:</p>
<ul>
<li><code>/dev/sda</code> — жёсткий диск</li>
<li><code>/dev/null</code> — "чёрная дыра" (всё что записано — исчезает)</li>
<li><code>/proc/cpuinfo</code> — информация о процессоре (виртуальный файл)</li>
<li><code>/sys/class/net/eth0</code> — сетевой интерфейс</li>
</ul>

<p>Это даёт единый интерфейс: для работы с любым ресурсом используются одни и те же вызовы <code>open/read/write/close</code>.</p>

<h2>Оболочка (Shell)</h2>
<p><strong>Shell</strong> — программа, которая принимает текстовые команды и исполняет их. Самые популярные:</p>
<ul>
<li><code>bash</code> — стандарт на большинстве серверов (Bourne Again Shell)</li>
<li><code>zsh</code> — расширенный bash (стандарт на macOS)</li>
<li><code>sh</code> — минималистичный POSIX shell (для скриптов)</li>
</ul>

<pre><code># Узнать свой текущий shell
echo $SHELL

# Посмотреть все доступные shells
cat /etc/shells</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что такое Linux в строгом смысле?",
				Options:     []string{"Полноценная операционная система", "Только ядро (kernel)", "Дистрибутив Ubuntu", "Графический интерфейс"},
				Correct:     1,
				Explanation: "Linux — это ядро, управляющее железом. Полная ОС (Ubuntu, Debian и т.д.) = ядро Linux + утилиты GNU + пакетный менеджер. Поэтому точное название: GNU/Linux.",
			},
			{
				Question:    "Что делают системные вызовы (syscalls)?",
				Options:     []string{"Вызывают техподдержку", "Позволяют программам просить ядро выполнить операции с железом", "Запускают другие программы", "Перезагружают систему"},
				Correct:     1,
				Explanation: "Программы в userspace не имеют прямого доступа к железу. Syscalls (open, read, write, fork, mmap) — единственный способ попросить ядро сделать что-то привилегированное. Go runtime активно использует syscalls: goroutine создаётся через clone(), файлы читаются через read().",
			},
			{
				Question:    "Почему Alpine Linux популярен в Docker-контейнерах?",
				Options:     []string{"Он самый безопасный", "Его базовый образ весит ~5MB (musl вместо glibc, нет лишних пакетов)", "Он поддерживает больше пакетов", "Он работает быстрее всех"},
				Correct:     1,
				Explanation: "Alpine использует musl libc вместо glibc и busybox вместо GNU coreutils. Это даёт образ ~5MB vs ~80MB у Debian. Меньше размер = быстрее pull, меньше поверхность атаки. Но: Go-программы с CGO могут не работать на Alpine без дополнительных настроек (из-за musl).",
			},
			{
				Question:    "Что означает философия 'Everything is a file'?",
				Options:     []string{"Все данные хранятся в файлах", "Устройства, процессы и сокеты представлены как файлы с единым интерфейсом open/read/write", "Linux может открыть любой формат файлов", "Конфигурация хранится только в файлах"},
				Correct:     1,
				Explanation: "Единый интерфейс: /dev/sda (диск), /proc/1/status (процесс), /dev/urandom (генератор случайных чисел) — все работают через read()/write(). Это упрощает программирование и автоматизацию: один инструмент (cat, echo) работает с чем угодно.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер информации о дистрибутиве",
				Description: "Напиши Go-программу, которая парсит содержимое файла /etc/os-release (стандартный формат для идентификации дистрибутива Linux) и выводит ключевые поля.",
				Hints:       "Формат файла — KEY=VALUE или KEY=\"VALUE\" на каждой строке. Используй strings.Cut() для разделения по '='. Убирай кавычки через strings.Trim().",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// parseOSRelease парсит содержимое /etc/os-release из stdin
// Формат: KEY=VALUE или KEY="VALUE WITH SPACES"
func parseOSRelease(lines []string) map[string]string {
	result := make(map[string]string)
	// TODO: распарси каждую строку
	// 1. Пропусти пустые строки и комментарии (#)
	// 2. Раздели по первому '='
	// 3. Убери кавычки из значения
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	info := parseOSRelease(lines)
	fmt.Printf("Name: %s\n", info["NAME"])
	fmt.Printf("Version: %s\n", info["VERSION_ID"])
	fmt.Printf("ID: %s\n", info["ID"])
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func parseOSRelease(lines []string) map[string]string {
	result := make(map[string]string)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		value = strings.Trim(value, "\"")
		result[key] = value
	}
	return result
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	info := parseOSRelease(lines)
	fmt.Printf("Name: %s\n", info["NAME"])
	fmt.Printf("Version: %s\n", info["VERSION_ID"])
	fmt.Printf("ID: %s\n", info["ID"])
}`,
				TestCases: []TestCase{
					{
						Input:          "NAME=\"Ubuntu\"\nVERSION_ID=\"22.04\"\nID=ubuntu\nID_LIKE=debian\n",
						ExpectedOutput: "Name: Ubuntu\nVersion: 22.04\nID: ubuntu\n",
					},
					{
						Input:          "NAME=\"Alpine Linux\"\nVERSION_ID=3.19\nID=alpine\n",
						ExpectedOutput: "Name: Alpine Linux\nVersion: 3.19\nID: alpine\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Kernel", Definition: "Ядро ОС — управляет железом, памятью, процессами. Программы общаются с ним через syscalls."},
					{Term: "Distro", Definition: "Дистрибутив — готовая сборка: ядро + утилиты + пакетный менеджер (Ubuntu, Alpine, Rocky)."},
					{Term: "Userspace", Definition: "Всё что работает вне ядра: пользовательские программы, демоны, shell."},
				},
			},
			{
				Title:       "Определитель семейства дистрибутива",
				Description: "Напиши программу, которая по имени дистрибутива определяет его семейство (debian, rhel, alpine, arch, unknown) и пакетный менеджер.",
				Hints:       "Используй map или switch. Debian-семейство: ubuntu, debian, mint, pop. RHEL: centos, rocky, alma, fedora, rhel. Arch: arch, manjaro.",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strings"
)

type DistroInfo struct {
	Family         string // debian, rhel, alpine, arch, unknown
	PackageManager string // apt, dnf, apk, pacman, unknown
}

// identifyDistro определяет семейство и пакетный менеджер по имени дистрибутива
func identifyDistro(name string) DistroInfo {
	name = strings.ToLower(strings.TrimSpace(name))
	// TODO: реализуй определение семейства
	// debian-based: ubuntu, debian, mint, pop → apt
	// rhel-based: centos, rocky, alma, fedora, rhel → dnf
	// alpine → apk
	// arch-based: arch, manjaro → pacman
	return DistroInfo{Family: "unknown", PackageManager: "unknown"}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: distro <name>")
		return
	}
	info := identifyDistro(os.Args[1])
	fmt.Printf("Family: %s\nPackage Manager: %s\n", info.Family, info.PackageManager)
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strings"
)

type DistroInfo struct {
	Family         string
	PackageManager string
}

func identifyDistro(name string) DistroInfo {
	name = strings.ToLower(strings.TrimSpace(name))

	debianFamily := []string{"ubuntu", "debian", "mint", "pop", "kali", "elementary"}
	rhelFamily := []string{"centos", "rocky", "alma", "fedora", "rhel", "oracle"}
	archFamily := []string{"arch", "manjaro", "endeavour"}

	for _, d := range debianFamily {
		if strings.Contains(name, d) {
			return DistroInfo{Family: "debian", PackageManager: "apt"}
		}
	}
	for _, d := range rhelFamily {
		if strings.Contains(name, d) {
			return DistroInfo{Family: "rhel", PackageManager: "dnf"}
		}
	}
	for _, d := range archFamily {
		if strings.Contains(name, d) {
			return DistroInfo{Family: "arch", PackageManager: "pacman"}
		}
	}
	if strings.Contains(name, "alpine") {
		return DistroInfo{Family: "alpine", PackageManager: "apk"}
	}
	return DistroInfo{Family: "unknown", PackageManager: "unknown"}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: distro <name>")
		return
	}
	info := identifyDistro(os.Args[1])
	fmt.Printf("Family: %s\nPackage Manager: %s\n", info.Family, info.PackageManager)
}`,
				TestCases: []TestCase{
					{
						Input:          "ubuntu",
						ExpectedOutput: "Family: debian\nPackage Manager: apt\n",
					},
					{
						Input:          "Rocky",
						ExpectedOutput: "Family: rhel\nPackage Manager: dnf\n",
					},
					{
						Input:          "alpine",
						ExpectedOutput: "Family: alpine\nPackage Manager: apk\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Package Manager", Definition: "Утилита для установки/обновления/удаления программ. apt (Debian), dnf (RHEL), apk (Alpine)."},
					{Term: "Repository", Definition: "Удалённый сервер с пакетами. Пакетный менеджер скачивает оттуда программы."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 2: Filesystem & Navigation
// ══════════════════════════════════════════════════════════════════

func linuxLesson02Filesystem() L {
	return L{
		Slug: "filesystem-navigation", Title: "Файловая система и навигация", Order: 2,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Файловая система Linux и навигация</h1>

<h2>FHS — Filesystem Hierarchy Standard</h2>
<p>В Linux всё начинается с <strong>/</strong> (root, корень). Нет дисков C: и D: — всё монтируется в единое дерево. <strong>FHS</strong> — стандарт, определяющий что где лежит:</p>

<pre><code>/
├── bin/          → основные утилиты (ls, cp, cat) — нужны для загрузки
├── sbin/         → системные утилиты (fdisk, iptables) — для root
├── etc/          → конфигурация ВСЕХ программ (текстовые файлы)
├── var/          → изменяемые данные (логи, БД, кеш, почта)
│   ├── log/      → логи системы и сервисов
│   ├── lib/      → данные сервисов (PostgreSQL, Docker)
│   └── cache/    → кеш пакетных менеджеров
├── home/         → домашние каталоги пользователей
│   └── user/     → /home/user (~)
├── root/         → домашняя папка суперпользователя
├── tmp/          → временные файлы (очищается при reboot)
├── usr/          → пользовательские программы (read-only)
│   ├── bin/      → основная масса программ
│   ├── lib/      → библиотеки
│   └── local/    → вручную установленное ПО
├── proc/         → виртуальная ФС — информация о процессах (RAM)
├── sys/          → виртуальная ФС — информация о железе (RAM)
├── dev/          → файлы устройств (диски, терминалы, null)
├── mnt/          → точки монтирования (временные)
└── opt/          → опциональное ПО (крупные пакеты)</code></pre>

<h2>Важные каталоги для DevOps</h2>
<ul>
<li><code>/etc/</code> — ВСЕ конфиги: nginx.conf, postgresql.conf, ssh/sshd_config, hosts, resolv.conf</li>
<li><code>/var/log/</code> — логи: syslog, auth.log, nginx/access.log</li>
<li><code>/var/lib/docker/</code> — все данные Docker (образы, контейнеры, volumes)</li>
<li><code>/proc/</code> — живая информация: /proc/meminfo, /proc/cpuinfo, /proc/[PID]/</li>
</ul>

<h2>Навигация</h2>
<pre><code># Где я?
pwd                    # /home/user/projects

# Что тут лежит?
ls                     # файлы и папки (кратко)
ls -la                 # подробно: права, владелец, размер, дата
ls -lah                # + человекочитаемые размеры (1K, 5M, 2G)
ls -lt                 # сортировка по времени (новые сверху)

# Перейти
cd /var/log            # абсолютный путь
cd ../                 # на уровень выше
cd ~                   # в домашнюю папку
cd -                   # в предыдущий каталог (toggle)

# Структура каталогов
tree -L 2              # дерево, глубина 2
tree -d                # только директории</code></pre>

<h2>Поиск файлов</h2>
<pre><code># find — мощный поиск
find /etc -name "*.conf"              # все .conf файлы в /etc
find /var/log -mtime -1               # изменённые за последние сутки
find / -size +100M                    # файлы больше 100MB
find /home -type d -name "node_modules"  # директории с именем node_modules

# locate — быстрый поиск (по индексу, обновляется updatedb)
locate nginx.conf

# which / whereis — найти исполняемый файл
which go             # /usr/local/go/bin/go
whereis nginx        # nginx: /usr/sbin/nginx /etc/nginx /usr/share/nginx</code></pre>

<h2>Абсолютные vs относительные пути</h2>
<ul>
<li><strong>Абсолютный</strong> — от корня: <code>/var/log/nginx/access.log</code></li>
<li><strong>Относительный</strong> — от текущей папки: <code>../config/app.yaml</code></li>
<li><code>.</code> — текущий каталог</li>
<li><code>..</code> — родительский каталог</li>
<li><code>~</code> — домашний каталог текущего пользователя</li>
</ul>

<h2>Полезные команды для работы с файлами</h2>
<pre><code># Создание
mkdir -p project/cmd/server    # создать всё дерево (-p)
touch file.txt                 # создать пустой файл

# Копирование и перемещение
cp file.txt backup.txt         # копировать
cp -r dir/ dir_backup/         # копировать директорию (-r рекурсивно)
mv old.txt new.txt             # переименовать / переместить

# Удаление
rm file.txt                    # удалить файл
rm -rf dir/                    # удалить директорию со всем содержимым
                               # ⚠️ ОСТОРОЖНО! Нет корзины!

# Просмотр содержимого
cat file.txt                   # весь файл
head -20 file.txt              # первые 20 строк
tail -50 file.txt              # последние 50 строк
tail -f /var/log/syslog        # следить за файлом в реальном времени
less file.txt                  # постраничный просмотр (q — выход)</code></pre>

<p><strong>Продовый совет:</strong> Никогда не делай <code>rm -rf /</code> или <code>rm -rf *</code> от root. Одна ошибка в пути — и сервер мёртв. Всегда дважды проверяй путь перед rm -rf.</p>`,

		Quiz: []Q{
			{
				Question:    "Где в Linux хранятся конфигурационные файлы сервисов?",
				Options:     []string{"/home/config/", "/etc/", "/usr/config/", "/var/config/"},
				Correct:     1,
				Explanation: "/etc/ (от 'et cetera' или 'editable text configuration') — стандартное место для всех конфигов. Nginx: /etc/nginx/, SSH: /etc/ssh/, PostgreSQL: /etc/postgresql/. Это текстовые файлы, которые можно редактировать и версионировать в Git.",
			},
			{
				Question:    "Что такое /proc/ и /sys/?",
				Options:     []string{"Обычные каталоги на диске с файлами процессов", "Виртуальные файловые системы в RAM — интерфейс к информации ядра", "Каталоги для хранения процедур и системных скриптов", "Резервные копии системы"},
				Correct:     1,
				Explanation: "/proc/ и /sys/ не занимают места на диске — это виртуальные ФС, через которые ядро предоставляет информацию. /proc/meminfo показывает RAM, /proc/[PID]/status — информацию о процессе. Они существуют только в оперативной памяти и генерируются ядром на лету.",
			},
			{
				Question:    "Чем отличается find от locate?",
				Options:     []string{"Ничем — это алиасы", "find ищет в реальном времени (медленнее), locate ищет по предварительно построенному индексу (быстрее, но может быть устаревшим)", "locate ищет только в /usr/", "find работает только с именами файлов"},
				Correct:     1,
				Explanation: "find обходит файловую систему прямо сейчас — всегда актуален, но медленный на больших ФС. locate использует базу данных (обновляется через updatedb/cron), находит мгновенно, но может показать удалённые файлы. На серверах чаще используют find.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер вывода ls -la",
				Description: "Напиши Go-программу, которая парсит вывод команды `ls -la` и выводит информацию в структурированном виде: имя файла, тип (файл/директория/симлинк), размер.",
				Hints:       "Каждая строка ls -la имеет формат: permissions links owner group size month day time name. Первый символ permissions: d=директория, l=симлинк, -=файл. Используй strings.Fields() для разбивки по пробелам.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type FileEntry struct {
	Name     string
	Type     string // "file", "directory", "symlink"
	Size     string
	Owner    string
	Perms    string
}

// parseLsLine парсит одну строку вывода ls -la
// Пример: "drwxr-xr-x 2 root root 4096 Jan 15 10:30 config"
func parseLsLine(line string) (FileEntry, bool) {
	// TODO: распарси строку
	// 1. Пропусти строку "total ..." (первая строка ls -la)
	// 2. Разбей по пробелам (strings.Fields)
	// 3. Определи тип по первому символу прав: d, l, -
	// 4. Верни FileEntry
	return FileEntry{}, false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		entry, ok := parseLsLine(scanner.Text())
		if !ok {
			continue
		}
		fmt.Printf("%-12s %-10s %8s %s\n", entry.Type, entry.Owner, entry.Size, entry.Name)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type FileEntry struct {
	Name  string
	Type  string
	Size  string
	Owner string
	Perms string
}

func parseLsLine(line string) (FileEntry, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "total") {
		return FileEntry{}, false
	}

	fields := strings.Fields(line)
	if len(fields) < 9 {
		return FileEntry{}, false
	}

	perms := fields[0]
	var fileType string
	switch perms[0] {
	case 'd':
		fileType = "directory"
	case 'l':
		fileType = "symlink"
	default:
		fileType = "file"
	}

	name := strings.Join(fields[8:], " ")

	return FileEntry{
		Name:  name,
		Type:  fileType,
		Size:  fields[4],
		Owner: fields[2],
		Perms: perms,
	}, true
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		entry, ok := parseLsLine(scanner.Text())
		if !ok {
			continue
		}
		fmt.Printf("%-12s %-10s %8s %s\n", entry.Type, entry.Owner, entry.Size, entry.Name)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "total 48\ndrwxr-xr-x 2 root root 4096 Jan 15 10:30 config\n-rw-r--r-- 1 user user 1234 Feb 01 09:00 main.go\nlrwxrwxrwx 1 root root 11 Mar 05 12:00 link -> /etc/hosts\n",
						ExpectedOutput: "directory    root           4096 config\nfile         user           1234 main.go\nsymlink      root             11 link -> /etc/hosts\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "FHS", Definition: "Filesystem Hierarchy Standard — стандарт, описывающий назначение каждого каталога в Linux."},
					{Term: "Inode", Definition: "Структура данных в файловой системе, хранящая метаданные файла (права, владелец, указатели на блоки данных)."},
					{Term: "Mount point", Definition: "Каталог, к которому примонтирована файловая система (диск, раздел, NFS)."},
				},
			},
			{
				Title:       "Навигатор по FHS",
				Description: "Напиши Go-программу, которая по названию типа данных возвращает правильный путь в Linux FHS. Программа должна знать где хранить логи, конфиги, временные файлы, данные БД и т.д.",
				Hints:       "Создай map с описанием стандартных путей. Логи → /var/log/, конфиги → /etc/, данные → /var/lib/, временные → /tmp/, бинарники → /usr/bin/.",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strings"
)

// FHSPath возвращает рекомендованный путь в Linux для данного типа ресурса
func FHSPath(resourceType string) string {
	resourceType = strings.ToLower(strings.TrimSpace(resourceType))
	// TODO: верни правильный путь FHS для типа ресурса:
	// "logs" → "/var/log/"
	// "config" → "/etc/"
	// "database" → "/var/lib/"
	// "temp" → "/tmp/"
	// "binary" → "/usr/bin/"
	// "user-data" → "/home/"
	// "system-binary" → "/sbin/"
	// "cache" → "/var/cache/"
	// иначе → "unknown"
	return "unknown"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fhs <resource-type>")
		return
	}
	path := FHSPath(os.Args[1])
	fmt.Println(path)
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strings"
)

func FHSPath(resourceType string) string {
	resourceType = strings.ToLower(strings.TrimSpace(resourceType))

	paths := map[string]string{
		"logs":          "/var/log/",
		"config":        "/etc/",
		"database":      "/var/lib/",
		"temp":          "/tmp/",
		"binary":        "/usr/bin/",
		"user-data":     "/home/",
		"system-binary": "/sbin/",
		"cache":         "/var/cache/",
	}

	if path, ok := paths[resourceType]; ok {
		return path
	}
	return "unknown"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fhs <resource-type>")
		return
	}
	path := FHSPath(os.Args[1])
	fmt.Println(path)
}`,
				TestCases: []TestCase{
					{Input: "logs", ExpectedOutput: "/var/log/\n"},
					{Input: "config", ExpectedOutput: "/etc/\n"},
					{Input: "database", ExpectedOutput: "/var/lib/\n"},
					{Input: "temp", ExpectedOutput: "/tmp/\n"},
					{Input: "blah", ExpectedOutput: "unknown\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "FHS", Definition: "Filesystem Hierarchy Standard — описывает назначение каждого каталога в корне Linux."},
					{Term: "Root (/)", Definition: "Корневой каталог файловой системы. Все пути начинаются с /."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 3: Files & Permissions
// ══════════════════════════════════════════════════════════════════

func linuxLesson03Permissions() L {
	return L{
		Slug: "file-permissions", Title: "Файлы и права доступа", Order: 3,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Файлы и права доступа в Linux</h1>

<h2>Три типа прав</h2>
<p>Каждый файл имеет три набора прав для трёх категорий:</p>

<pre><code>-rwxr-xr-- 1 user group 4096 Jan 15 10:30 script.sh
│└┬┘└┬┘└┬┘
│ │  │  └── Others (остальные): r-- (только чтение)
│ │  └───── Group (группа): r-x (чтение + выполнение)
│ └──────── User/Owner (владелец): rwx (всё)
└────────── Тип: - файл, d директория, l симлинк</code></pre>

<table>
<tr><th>Буква</th><th>Значение</th><th>Для файла</th><th>Для директории</th></tr>
<tr><td>r</td><td>read (4)</td><td>Читать содержимое</td><td>Листать содержимое (ls)</td></tr>
<tr><td>w</td><td>write (2)</td><td>Изменять содержимое</td><td>Создавать/удалять файлы внутри</td></tr>
<tr><td>x</td><td>execute (1)</td><td>Запускать как программу</td><td>Входить в каталог (cd)</td></tr>
</table>

<h2>Восьмеричная (octal) нотация</h2>
<p>Каждый набор rwx конвертируется в число 0-7:</p>
<pre><code>r=4, w=2, x=1

rwx = 4+2+1 = 7
r-x = 4+0+1 = 5
r-- = 4+0+0 = 4
--- = 0+0+0 = 0

Примеры:
chmod 755 file  → rwxr-xr-x (владелец=всё, остальные=чтение+выполнение)
chmod 644 file  → rw-r--r-- (стандарт для обычных файлов)
chmod 600 file  → rw------- (только владелец читает/пишет)
chmod 700 dir   → rwx------ (только владелец заходит)</code></pre>

<h2>chmod — изменение прав</h2>
<pre><code># Числовой формат
chmod 755 script.sh        # rwxr-xr-x
chmod 600 secrets.env      # rw------- (только владелец)

# Символьный формат
chmod u+x script.sh        # добавить execute владельцу
chmod g-w file.txt         # убрать write у группы
chmod o-rwx private/       # убрать все права у others
chmod a+r readme.txt       # добавить read всем (a=all)
chmod -R 755 project/      # рекурсивно для всех файлов</code></pre>

<h2>chown — смена владельца</h2>
<pre><code># Сменить владельца
chown nginx /var/www/html

# Сменить владельца и группу
chown nginx:www-data /var/www/html

# Рекурсивно
chown -R postgres:postgres /var/lib/postgresql/</code></pre>

<h2>umask — права по умолчанию</h2>
<p><strong>umask</strong> определяет какие права <strong>убирать</strong> при создании новых файлов:</p>
<pre><code># Максимальные права: файл=666, директория=777
# umask вычитается из максимума

umask 022  → файлы: 644 (666-022), директории: 755 (777-022)
umask 077  → файлы: 600, директории: 700 (только владелец)

# Посмотреть текущую umask
umask         # 0022
umask -S      # u=rwx,g=rx,o=rx</code></pre>

<h2>Специальные биты: SUID, SGID, Sticky</h2>
<pre><code># SUID (Set User ID) — выполняется от имени владельца файла
chmod u+s /usr/bin/passwd    # 4755 — passwd работает от root
ls -la /usr/bin/passwd       # -rwsr-xr-x (s вместо x)
# Зачем: passwd должен писать в /etc/shadow, а он принадлежит root

# SGID (Set Group ID) — наследование группы каталога
chmod g+s /shared/project/   # 2775 — новые файлы получают группу каталога
# Зачем: совместная работа над проектом

# Sticky bit — удалять может только владелец
chmod +t /tmp                # 1777 — /tmp доступен всем, но удалять свои файлы
ls -la / | grep tmp          # drwxrwxrwt (t в конце)</code></pre>

<h2>Под капотом: inode</h2>
<p>Права хранятся в <strong>inode</strong> — структуре данных файловой системы:</p>
<pre><code># Посмотреть inode файла
stat file.txt
# File: file.txt
# Size: 1234      Blocks: 8      IO Block: 4096
# Inode: 2621441  Links: 1
# Access: (0644/-rw-r--r--)  Uid: (1000/user)  Gid: (1000/user)
# Access: 2024-01-15 10:30:00
# Modify: 2024-01-14 09:00:00
# Change: 2024-01-14 09:00:00  ← метаданные (chmod) менялись тут</code></pre>

<p><strong>Продовый совет:</strong> <code>chmod 777</code> — почти всегда ошибка. Если сервис не работает и ты ставишь 777 — ты не решил проблему, а создал дыру в безопасности. Разберись кто реально владелец и какие минимальные права нужны.</p>`,

		Quiz: []Q{
			{
				Question:    "Что означает chmod 600 file.txt?",
				Options:     []string{"Все могут читать и писать", "Только владелец может читать и писать, остальные — ничего", "Все могут выполнять", "Файл удалён"},
				Correct:     1,
				Explanation: "6=rw- (read+write для владельца), 0=--- (ничего для группы), 0=--- (ничего для остальных). Это стандарт для секретов: SSH-ключи (~/.ssh/id_rsa), .env файлы, API-токены. SSH откажется работать с ключом, если права шире 600.",
			},
			{
				Question:    "Как chmod работает на уровне inode?",
				Options:     []string{"Меняет содержимое файла", "Обновляет 12-битное поле mode в inode файла (9 бит rwx + 3 специальных бита)", "Создаёт новый файл с другими правами", "Меняет имя файла"},
				Correct:     1,
				Explanation: "Inode хранит метаданные файла, включая 12-битное поле mode: 3 бита (SUID/SGID/sticky) + 9 бит (rwx для user/group/other). chmod обновляет это поле через системный вызов chmod(). Ctime (change time) обновляется при каждом изменении метаданных.",
			},
			{
				Question:    "Что случится, если сделать chmod 777 /etc/shadow?",
				Options:     []string{"Ничего страшного", "Критическая уязвимость — любой пользователь сможет прочитать хеши паролей всех аккаунтов", "Файл удалится", "Система перезагрузится"},
				Correct:     1,
				Explanation: "/etc/shadow содержит хеши паролей и должен иметь права 640 (root:shadow). chmod 777 позволит любому пользователю прочитать и даже перезаписать хеши. Утилиты безопасности (auditd, AIDE) поднимут alert. Некоторые системы (SELinux) могут заблокировать это через MAC.",
			},
			{
				Question:    "Для чего нужен Sticky bit на /tmp?",
				Options:     []string{"Ускоряет работу с файлами", "Запрещает удаление чужих файлов — каждый может удалять только свои", "Делает файлы read-only", "Шифрует содержимое"},
				Correct:     1,
				Explanation: "/tmp имеет права 1777 (rwxrwxrwt). Без sticky bit любой пользователь мог бы удалить файлы другого пользователя (т.к. w на директорию = право удалять файлы). Sticky bit (t) добавляет ограничение: удалять может только владелец файла или root.",
			},
		},

		Tasks: []T{
			{
				Title:       "Калькулятор chmod",
				Description: "Напиши Go-программу, которая конвертирует между символьным (rwxr-xr--) и восьмеричным (754) представлением прав доступа Linux. Программа принимает один формат и выводит оба.",
				Hints:       "Для конвертации символьного → числового: r=4, w=2, x=1, -=0. Суммируй для каждой тройки. Для числового → символьного: разложи каждую цифру на биты.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// symbolicToOctal конвертирует "rwxr-xr--" → "754"
func symbolicToOctal(symbolic string) string {
	// TODO: реализуй конвертацию
	// Обрабатывай по 3 символа: rwx → 7, r-x → 5, r-- → 4
	return ""
}

// octalToSymbolic конвертирует "754" → "rwxr-xr--"
func octalToSymbolic(octal string) string {
	// TODO: реализуй конвертацию
	// Для каждой цифры: 7→rwx, 6→rw-, 5→r-x, 4→r--, и т.д.
	return ""
}

// isOctal проверяет, является ли строка восьмеричным числом (3 цифры 0-7)
func isOctal(s string) bool {
	// TODO
	return false
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: chmod-calc <permissions>")
		fmt.Println("  chmod-calc 755")
		fmt.Println("  chmod-calc rwxr-xr-x")
		return
	}

	input := os.Args[1]
	if isOctal(input) {
		sym := octalToSymbolic(input)
		fmt.Printf("Octal:    %s\nSymbolic: %s\n", input, sym)
	} else {
		oct := symbolicToOctal(input)
		fmt.Printf("Symbolic: %s\nOctal:    %s\n", input, oct)
	}
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func symbolicToOctal(symbolic string) string {
	if len(symbolic) != 9 {
		return "000"
	}
	var result strings.Builder
	for i := 0; i < 9; i += 3 {
		val := 0
		if symbolic[i] == 'r' {
			val += 4
		}
		if symbolic[i+1] == 'w' {
			val += 2
		}
		if symbolic[i+2] == 'x' {
			val += 1
		}
		result.WriteString(strconv.Itoa(val))
	}
	return result.String()
}

func octalToSymbolic(octal string) string {
	var result strings.Builder
	for _, ch := range octal {
		digit := int(ch - '0')
		if digit&4 != 0 {
			result.WriteByte('r')
		} else {
			result.WriteByte('-')
		}
		if digit&2 != 0 {
			result.WriteByte('w')
		} else {
			result.WriteByte('-')
		}
		if digit&1 != 0 {
			result.WriteByte('x')
		} else {
			result.WriteByte('-')
		}
	}
	return result.String()
}

func isOctal(s string) bool {
	if len(s) != 3 {
		return false
	}
	for _, ch := range s {
		if ch < '0' || ch > '7' {
			return false
		}
	}
	return true
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: chmod-calc <permissions>")
		fmt.Println("  chmod-calc 755")
		fmt.Println("  chmod-calc rwxr-xr-x")
		return
	}

	input := os.Args[1]
	if isOctal(input) {
		sym := octalToSymbolic(input)
		fmt.Printf("Octal:    %s\nSymbolic: %s\n", input, sym)
	} else {
		oct := symbolicToOctal(input)
		fmt.Printf("Symbolic: %s\nOctal:    %s\n", input, oct)
	}
}`,
				TestCases: []TestCase{
					{Input: "755", ExpectedOutput: "Octal:    755\nSymbolic: rwxr-xr-x\n"},
					{Input: "644", ExpectedOutput: "Octal:    644\nSymbolic: rw-r--r--\n"},
					{Input: "rwxr-xr--", ExpectedOutput: "Symbolic: rwxr-xr--\nOctal:    754\n"},
					{Input: "600", ExpectedOutput: "Octal:    600\nSymbolic: rw-------\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "chmod", Definition: "Change mode — изменяет права доступа к файлу. Работает с inode."},
					{Term: "Octal notation", Definition: "Восьмеричная запись прав: каждая цифра 0-7 кодирует rwx для user/group/other."},
					{Term: "umask", Definition: "Маска прав, которые УБИРАЮТСЯ при создании новых файлов. umask 022 → новые файлы 644."},
				},
			},
			{
				Title:       "Анализатор безопасности прав",
				Description: "Напиши программу, которая проверяет восьмеричные права на типичные проблемы безопасности и выдаёт предупреждения: world-writable, too-open для секретов, отсутствие execute для скрипта и т.д.",
				Hints:       "Анализируй last digit (others): если он содержит w (2, 3, 6, 7) — world-writable. Для секретных файлов (передай тип) проверяй что group и others = 0.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strings"
)

type SecurityCheck struct {
	Level   string // "ok", "warning", "critical"
	Message string
}

// analyzePermissions проверяет права на проблемы безопасности
// fileType: "secret", "script", "config", "public"
func analyzePermissions(octal string, fileType string) []SecurityCheck {
	var checks []SecurityCheck
	// TODO:
	// 1. Парсим octal в три цифры (user, group, other)
	// 2. Проверяем world-writable (other имеет w)
	// 3. Для "secret": group и other должны быть 0
	// 4. Для "script": user должен иметь x
	// 5. Если 777 — всегда critical
	return checks
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: perms-check <octal> <type>")
		fmt.Println("Types: secret, script, config, public")
		return
	}
	checks := analyzePermissions(os.Args[1], os.Args[2])
	for _, c := range checks {
		fmt.Printf("[%s] %s\n", strings.ToUpper(c.Level), c.Message)
	}
	if len(checks) == 0 {
		fmt.Println("[OK] No issues found")
	}
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strings"
)

type SecurityCheck struct {
	Level   string
	Message string
}

func analyzePermissions(octal string, fileType string) []SecurityCheck {
	var checks []SecurityCheck

	if len(octal) != 3 {
		return []SecurityCheck{{Level: "critical", Message: "Invalid octal format"}}
	}

	user := int(octal[0] - '0')
	group := int(octal[1] - '0')
	other := int(octal[2] - '0')

	if octal == "777" {
		checks = append(checks, SecurityCheck{"critical", "777 grants full access to everyone"})
		return checks
	}

	if other&2 != 0 {
		checks = append(checks, SecurityCheck{"warning", "World-writable: anyone can modify"})
	}

	switch fileType {
	case "secret":
		if group != 0 || other != 0 {
			checks = append(checks, SecurityCheck{"critical", "Secret file accessible by group/others (should be 600)"})
		}
	case "script":
		if user&1 == 0 {
			checks = append(checks, SecurityCheck{"warning", "Script not executable by owner"})
		}
	case "config":
		if other&2 != 0 {
			checks = append(checks, SecurityCheck{"warning", "Config is world-writable"})
		}
	}

	return checks
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: perms-check <octal> <type>")
		fmt.Println("Types: secret, script, config, public")
		return
	}
	checks := analyzePermissions(os.Args[1], os.Args[2])
	for _, c := range checks {
		fmt.Printf("[%s] %s\n", strings.ToUpper(c.Level), c.Message)
	}
	if len(checks) == 0 {
		fmt.Println("[OK] No issues found")
	}
}`,
				TestCases: []TestCase{
					{Input: "777 secret", ExpectedOutput: "[CRITICAL] 777 grants full access to everyone\n"},
					{Input: "644 secret", ExpectedOutput: "[CRITICAL] Secret file accessible by group/others (should be 600)\n"},
					{Input: "600 secret", ExpectedOutput: "[OK] No issues found\n"},
					{Input: "644 script", ExpectedOutput: "[WARNING] Script not executable by owner\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "World-writable", Definition: "Файл, который может изменить любой пользователь системы. Почти всегда уязвимость."},
					{Term: "SUID", Definition: "Set User ID — при запуске файл исполняется от имени владельца, а не запустившего."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 4: Users & Groups
// ══════════════════════════════════════════════════════════════════

func linuxLesson04Users() L {
	return L{
		Slug: "users-groups", Title: "Пользователи и группы", Order: 4,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Пользователи и группы в Linux</h1>

<h2>Зачем нужны пользователи?</h2>
<p>Linux — многопользовательская система. Даже на сервере, где один человек, работает множество <strong>системных пользователей</strong>: nginx, postgres, docker. Каждый процесс запускается от имени какого-то пользователя, и это определяет его права.</p>

<h2>/etc/passwd — реестр пользователей</h2>
<pre><code># Формат: username:x:UID:GID:info:home:shell
root:x:0:0:root:/root:/bin/bash
user:x:1000:1000:Regular User:/home/user:/bin/bash
nginx:x:101:101:nginx web server:/var/cache/nginx:/usr/sbin/nologin
postgres:x:108:108:PostgreSQL:/var/lib/postgresql:/bin/bash
nobody:x:65534:65534:nobody:/nonexistent:/usr/sbin/nologin</code></pre>

<table>
<tr><th>Поле</th><th>Значение</th><th>Пример</th></tr>
<tr><td>username</td><td>Имя пользователя</td><td>nginx</td></tr>
<tr><td>x</td><td>Пароль (в /etc/shadow)</td><td>x</td></tr>
<tr><td>UID</td><td>ID пользователя</td><td>101</td></tr>
<tr><td>GID</td><td>ID основной группы</td><td>101</td></tr>
<tr><td>info</td><td>Описание (GECOS)</td><td>nginx web server</td></tr>
<tr><td>home</td><td>Домашняя директория</td><td>/var/cache/nginx</td></tr>
<tr><td>shell</td><td>Оболочка при входе</td><td>/usr/sbin/nologin</td></tr>
</table>

<p><strong>Важно:</strong> <code>/usr/sbin/nologin</code> или <code>/bin/false</code> — значит пользователь НЕ может залогиниться интерактивно. Это стандарт для сервисных аккаунтов (nginx, postgres) — они нужны только для разделения привилегий.</p>

<h2>/etc/shadow — хеши паролей</h2>
<pre><code># Доступен только root! Права: 640
# Формат: username:hash:lastchange:min:max:warn:inactive:expire
user:$6$salt$hash...:19500:0:99999:7:::
nginx:!:19000::::::</code></pre>
<p><code>!</code> или <code>*</code> в поле hash означает — аккаунт заблокирован (нельзя войти по паролю).</p>

<h2>UID: кто есть кто</h2>
<ul>
<li><code>UID 0</code> — root (суперпользователь, может всё)</li>
<li><code>UID 1-999</code> — системные пользователи (сервисы)</li>
<li><code>UID 1000+</code> — обычные пользователи</li>
</ul>

<h2>Управление пользователями</h2>
<pre><code># Создать пользователя
useradd -m -s /bin/bash username     # -m создаёт home, -s задаёт shell
useradd -r -s /usr/sbin/nologin app  # -r системный (без home, UID < 1000)

# Задать пароль
passwd username

# Изменить пользователя
usermod -aG docker username    # добавить в группу docker (-a = append!)
usermod -s /bin/zsh username   # сменить shell
usermod -L username            # заблокировать (Lock)

# Удалить
userdel -r username            # -r удаляет и домашнюю папку</code></pre>

<p><strong>Критическая ошибка:</strong> <code>usermod -G docker user</code> без <code>-a</code> ПЕРЕЗАПИШЕТ все группы пользователя! Всегда <code>-aG</code> для добавления.</p>

<h2>Группы</h2>
<pre><code># Посмотреть группы пользователя
groups username
id username        # uid=1000(user) gid=1000(user) groups=1000(user),998(docker)

# Создать группу
groupadd developers

# Добавить пользователя в группу
usermod -aG developers username

# /etc/group
docker:x:998:user,deploy
developers:x:1001:user,alice,bob</code></pre>

<h2>sudo — делегирование привилегий</h2>
<pre><code># Выполнить команду от root
sudo systemctl restart nginx
sudo cat /etc/shadow

# Стать root
sudo -i                    # интерактивный root shell
sudo su -                  # тоже самое

# Настройка: /etc/sudoers (редактировать ТОЛЬКО через visudo!)
# username ALL=(ALL:ALL) ALL           — может всё
# %docker ALL=(ALL) NOPASSWD: /usr/bin/docker  — группа docker может docker без пароля</code></pre>

<p><strong>Best practice:</strong> Не работай от root постоянно. Используй sudo для конкретных команд. На продовых серверах root-логин по SSH запрещён (<code>PermitRootLogin no</code> в sshd_config).</p>

<h2>su — переключение пользователей</h2>
<pre><code>su - username       # стать username (- загружает его окружение)
su -                # стать root (нужен пароль root)
sudo -u postgres psql  # выполнить команду от имени postgres</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что означает shell /usr/sbin/nologin у пользователя?",
				Options:     []string{"Пользователь удалён", "Пользователь не может войти интерактивно — это сервисный аккаунт для разделения привилегий", "Пользователь заблокирован навсегда", "У пользователя нет пароля"},
				Correct:     1,
				Explanation: "nologin не даёт пользователю получить shell при попытке логина (SSH, su). Но процессы могут работать от его имени (nginx, postgres). Это безопасность: даже если атакующий узнал пароль сервисного аккаунта, он не получит shell.",
			},
			{
				Question:    "Что произойдёт при usermod -G docker user (БЕЗ -a)?",
				Options:     []string{"Пользователь добавится в группу docker", "Все предыдущие группы пользователя будут УДАЛЕНЫ, останется только docker", "Ошибка синтаксиса", "Ничего не изменится"},
				Correct:     1,
				Explanation: "Без -a (append) команда ПЕРЕЗАПИСЫВАЕТ список групп. Если user был в groups: user, docker, developers, sudo — после usermod -G docker user останется только docker. Всегда используй -aG! Это одна из самых частых ошибок, которая может сломать доступ на сервере.",
			},
			{
				Question:    "Почему /etc/shadow хранится отдельно от /etc/passwd?",
				Options:     []string{"Для экономии места", "Потому что /etc/passwd читаем всеми (используется программами для маппинга UID→имя), а shadow хранит хеши паролей и доступен только root", "Историческая традиция без практического значения", "Shadow хранит данные в бинарном формате"},
				Correct:     1,
				Explanation: "Раньше хеши были в /etc/passwd (который читаем всеми для id→name маппинга). Любой пользователь мог брутфорсить пароли. Shadow с правами 640 (root:shadow) решил эту проблему. Разделение данных по уровню секретности — классический security pattern.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер /etc/passwd",
				Description: "Напиши Go-программу, которая парсит строки формата /etc/passwd и выводит информацию о пользователях: кто системный, кто обычный, у кого есть shell.",
				Hints:       "Разделитель — ':'. UID < 1000 — системный (кроме 0 = root). Shell содержит 'nologin' или 'false' — пользователь не может войти.",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Username string
	UID      int
	GID      int
	Home     string
	Shell    string
	Type     string // "root", "system", "regular"
	CanLogin bool
}

// parsePasswdLine парсит одну строку /etc/passwd
func parsePasswdLine(line string) (User, error) {
	// TODO: разбей строку по ':' и заполни структуру
	// Формат: username:x:UID:GID:info:home:shell
	return User{}, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		user, err := parsePasswdLine(line)
		if err != nil {
			continue
		}
		login := "no-login"
		if user.CanLogin {
			login = "can-login"
		}
		fmt.Printf("%-12s UID=%-5d %-8s %s\n", user.Username, user.UID, user.Type, login)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type User struct {
	Username string
	UID      int
	GID      int
	Home     string
	Shell    string
	Type     string
	CanLogin bool
}

func parsePasswdLine(line string) (User, error) {
	parts := strings.Split(line, ":")
	if len(parts) < 7 {
		return User{}, fmt.Errorf("invalid format")
	}

	uid, err := strconv.Atoi(parts[2])
	if err != nil {
		return User{}, err
	}
	gid, err := strconv.Atoi(parts[3])
	if err != nil {
		return User{}, err
	}

	userType := "regular"
	if uid == 0 {
		userType = "root"
	} else if uid < 1000 {
		userType = "system"
	}

	shell := parts[6]
	canLogin := !strings.Contains(shell, "nologin") && !strings.Contains(shell, "false")

	return User{
		Username: parts[0],
		UID:      uid,
		GID:      gid,
		Home:     parts[5],
		Shell:    shell,
		Type:     userType,
		CanLogin: canLogin,
	}, nil
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		user, err := parsePasswdLine(line)
		if err != nil {
			continue
		}
		login := "no-login"
		if user.CanLogin {
			login = "can-login"
		}
		fmt.Printf("%-12s UID=%-5d %-8s %s\n", user.Username, user.UID, user.Type, login)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "root:x:0:0:root:/root:/bin/bash\nnginx:x:101:101:nginx:/var/cache/nginx:/usr/sbin/nologin\nuser:x:1000:1000:User:/home/user:/bin/bash\n",
						ExpectedOutput: "root         UID=0     root     can-login\nnginx        UID=101   system   no-login\nuser         UID=1000  regular  can-login\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "UID", Definition: "User ID — числовой идентификатор пользователя. 0=root, 1-999=системные, 1000+=обычные."},
					{Term: "GID", Definition: "Group ID — числовой идентификатор группы."},
					{Term: "GECOS", Definition: "Поле описания пользователя в /etc/passwd (имя, телефон, кабинет). Историческое название."},
				},
			},
			{
				Title:       "Валидатор sudoers-правил",
				Description: "Напиши программу, которая парсит упрощённые строки sudoers и определяет: кто может выполнять какие команды, нужен ли пароль, от какого пользователя.",
				Hints:       "Формат: user/group ALL=(runas) [NOPASSWD:] command. Группа начинается с %. NOPASSWD — опционально. Парси через strings.Fields и проверяй наличие маркеров.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type SudoRule struct {
	Subject    string // user или %group
	IsGroup    bool
	RunAs      string // от какого пользователя
	NoPassword bool
	Command    string
}

// parseSudoersLine парсит строку вида:
// user ALL=(ALL) NOPASSWD: /usr/bin/docker
// %admin ALL=(ALL:ALL) ALL
func parseSudoersLine(line string) (SudoRule, bool) {
	// TODO: реализуй парсинг
	// 1. Пропусти комментарии (#) и пустые строки
	// 2. Первое слово — subject (% = группа)
	// 3. Найди (runas) в скобках
	// 4. Проверь NOPASSWD:
	// 5. Последнее — команда
	return SudoRule{}, false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rule, ok := parseSudoersLine(scanner.Text())
		if !ok {
			continue
		}
		subjectType := "user"
		if rule.IsGroup {
			subjectType = "group"
		}
		passwd := "password-required"
		if rule.NoPassword {
			passwd = "no-password"
		}
		fmt.Printf("%-8s %-10s run-as=%-6s %-18s cmd=%s\n",
			subjectType, rule.Subject, rule.RunAs, passwd, rule.Command)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type SudoRule struct {
	Subject    string
	IsGroup    bool
	RunAs      string
	NoPassword bool
	Command    string
}

func parseSudoersLine(line string) (SudoRule, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "#") {
		return SudoRule{}, false
	}

	parts := strings.Fields(line)
	if len(parts) < 3 {
		return SudoRule{}, false
	}

	rule := SudoRule{}
	rule.Subject = parts[0]
	rule.IsGroup = strings.HasPrefix(parts[0], "%")

	// Find (runas)
	rule.RunAs = "ALL"
	for _, p := range parts {
		if strings.HasPrefix(p, "(") && strings.Contains(p, ")") {
			rule.RunAs = strings.Trim(p, "()")
			if idx := strings.Index(rule.RunAs, ":"); idx != -1 {
				rule.RunAs = rule.RunAs[:idx]
			}
			break
		}
	}

	// Check NOPASSWD
	fullLine := strings.Join(parts[2:], " ")
	rule.NoPassword = strings.Contains(fullLine, "NOPASSWD:")

	// Command is the last part
	if rule.NoPassword {
		idx := strings.Index(fullLine, "NOPASSWD:")
		after := strings.TrimSpace(fullLine[idx+len("NOPASSWD:"):])
		rule.Command = after
	} else {
		// Command after (runas)
		rule.Command = parts[len(parts)-1]
	}

	return rule, true
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		rule, ok := parseSudoersLine(scanner.Text())
		if !ok {
			continue
		}
		subjectType := "user"
		if rule.IsGroup {
			subjectType = "group"
		}
		passwd := "password-required"
		if rule.NoPassword {
			passwd = "no-password"
		}
		fmt.Printf("%-8s %-10s run-as=%-6s %-18s cmd=%s\n",
			subjectType, rule.Subject, rule.RunAs, passwd, rule.Command)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "user ALL=(ALL) NOPASSWD: /usr/bin/docker\n%admin ALL=(ALL:ALL) ALL\n",
						ExpectedOutput: "user     user       run-as=ALL    no-password        cmd=/usr/bin/docker\ngroup    %admin     run-as=ALL    password-required  cmd=ALL\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "sudo", Definition: "Super User DO — выполнить команду с правами другого пользователя (обычно root)."},
					{Term: "sudoers", Definition: "Файл /etc/sudoers — определяет кто и что может выполнять через sudo. Редактируется ТОЛЬКО через visudo."},
					{Term: "NOPASSWD", Definition: "Директива sudoers — не спрашивать пароль. Используется для автоматизации (CI/CD, скрипты)."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 5: Process Management
// ══════════════════════════════════════════════════════════════════

func linuxLesson05Processes() L {
	return L{
		Slug: "process-management", Title: "Управление процессами", Order: 5,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Управление процессами в Linux</h1>

<h2>Что такое процесс?</h2>
<p><strong>Процесс</strong> — запущенная программа. Каждый процесс имеет:</p>
<ul>
<li><strong>PID</strong> — уникальный идентификатор (Process ID)</li>
<li><strong>PPID</strong> — PID родительского процесса</li>
<li><strong>UID</strong> — от какого пользователя запущен</li>
<li><strong>State</strong> — состояние (running, sleeping, zombie, stopped)</li>
<li><strong>Своё адресное пространство</strong> — изолированная память</li>
</ul>

<p>Первый процесс — <code>init</code> (PID 1). В современных системах это <strong>systemd</strong>. Все процессы — его потомки (дерево).</p>

<h2>Жизненный цикл процесса</h2>
<pre><code>fork() → exec() → [работает] → exit() → [zombie] → wait() → [удалён]

Родитель вызывает fork() → создаётся копия (дочерний процесс)
Дочерний вызывает exec() → заменяет себя новой программой
Когда завершается → становится zombie (ждёт пока родитель прочитает exit code)
Родитель вызывает wait() → zombie удаляется</code></pre>

<h2>Состояния процессов</h2>
<table>
<tr><th>Состояние</th><th>Код</th><th>Описание</th></tr>
<tr><td>Running</td><td>R</td><td>Выполняется или готов к выполнению</td></tr>
<tr><td>Sleeping</td><td>S</td><td>Ждёт события (I/O, сигнал)</td></tr>
<tr><td>Disk Sleep</td><td>D</td><td>Ждёт I/O (нельзя прервать — нельзя kill!)</td></tr>
<tr><td>Stopped</td><td>T</td><td>Остановлен (Ctrl+Z или SIGSTOP)</td></tr>
<tr><td>Zombie</td><td>Z</td><td>Завершился, но родитель не прочитал exit code</td></tr>
</table>

<h2>Просмотр процессов</h2>
<pre><code># ps — snapshot процессов
ps aux                         # все процессы (BSD-стиль)
ps -ef                         # все процессы (System V стиль)
ps aux | grep nginx            # найти nginx
ps -o pid,ppid,user,%mem,%cpu,comm -p 1234  # конкретные поля

# top / htop — живой мониторинг
top                            # базовый (есть везде)
htop                           # расширенный (нужно установить)
# В top: M=сортировка по памяти, P=по CPU, k=kill, q=выход

# pgrep / pidof — найти PID по имени
pgrep nginx                    # все PID процессов nginx
pidof postgres                 # то же, но для одного процесса

# pstree — дерево процессов
pstree -p                      # с PID
pstree -p 1                    # дерево от init</code></pre>

<h2>Сигналы</h2>
<p>Сигналы — способ общения между процессами и ядром:</p>
<pre><code># Основные сигналы
kill -15 PID    # SIGTERM — вежливая просьба завершиться (можно обработать)
kill -9 PID     # SIGKILL — принудительное убийство (нельзя перехватить!)
kill -1 PID     # SIGHUP — перечитать конфиг (nginx, systemd)
kill -19 PID    # SIGSTOP — заморозить процесс
kill -18 PID    # SIGCONT — разморозить

# Ctrl+C = SIGINT (2)  — прервать
# Ctrl+Z = SIGTSTP (20) — остановить (suspend)
# Ctrl+\ = SIGQUIT (3)  — quit + core dump

# killall / pkill — kill по имени
killall nginx              # SIGTERM всем процессам nginx
pkill -9 -u baduser        # SIGKILL все процессы пользователя</code></pre>

<p><strong>Правило:</strong> Всегда сначала SIGTERM (15), дай процессу 5 секунд закрыть соединения и файлы, только потом SIGKILL (9). SIGKILL — крайняя мера.</p>

<h2>Foreground / Background</h2>
<pre><code># Запуск в background
./server &                     # & — запустить в фоне
nohup ./server &               # + не умрёт при закрытии терминала

# Управление jobs
jobs                           # список фоновых задач
fg %1                          # вернуть job 1 в foreground
bg %1                          # продолжить job 1 в background
Ctrl+Z                         # остановить текущий процесс
bg                             # продолжить в фоне

# disown — отвязать от терминала
./long-task &
disown %1                      # теперь не умрёт при закрытии терминала</code></pre>

<p><strong>Для серверов:</strong> Не запускай сервисы через <code>nohup &</code>. Используй systemd — он перезапустит при крашах, ведёт логи, управляет зависимостями.</p>`,

		Quiz: []Q{
			{
				Question:    "Чем SIGTERM (15) отличается от SIGKILL (9)?",
				Options:     []string{"Ничем — оба убивают процесс", "SIGTERM можно перехватить и обработать gracefully, SIGKILL нельзя — ядро убивает процесс немедленно", "SIGKILL медленнее", "SIGTERM работает только для root"},
				Correct:     1,
				Explanation: "SIGTERM — вежливая просьба. Процесс может перехватить его, закрыть соединения, сохранить данные и завершиться чисто. SIGKILL ядро обрабатывает само — процесс не получает шанса ничего сделать. Файлы могут остаться незаписанными, соединения — открытыми. Поэтому: сначала TERM, подожди, потом KILL.",
			},
			{
				Question:    "Что такое zombie-процесс и опасен ли он?",
				Options:     []string{"Вирус, который убивает систему", "Процесс завершился, но его exit code не прочитан родителем — занимает только запись в таблице процессов", "Процесс потребляет 100% CPU", "Процесс без прав доступа"},
				Correct:     1,
				Explanation: "Zombie (Z) — процесс уже мёртв, не потребляет CPU/RAM. Он занимает только PID и запись в /proc. Опасность: если их тысячи — может закончиться пул PID. Лечение: kill родителя (SIGCHLD) или перезапустить родительский процесс. Если родитель — PID 1 (systemd), он автоматически собирает zombies.",
			},
			{
				Question:    "Что означает состояние D (Disk Sleep / Uninterruptible Sleep)?",
				Options:     []string{"Процесс спит и экономит энергию", "Процесс ждёт I/O и НЕ может быть прерван даже SIGKILL — ядро защищает целостность данных", "Процесс загружает файл", "Процесс заблокирован другим процессом"},
				Correct:     1,
				Explanation: "D-state означает процесс ждёт завершения дискового I/O. SIGKILL не работает! Ядро не может прервать операцию, иначе файловая система может повредиться. Если процесс завис в D-state навсегда — обычно проблема с диском, NFS или драйвером. Может потребоваться перезагрузка.",
			},
			{
				Question:    "Зачем нужен nohup при запуске процесса в фоне?",
				Options:     []string{"Ускоряет выполнение", "Предотвращает получение SIGHUP при закрытии терминала — процесс продолжит работу", "Делает процесс невидимым", "Уменьшает потребление памяти"},
				Correct:     1,
				Explanation: "Когда ты закрываешь SSH-сессию, shell отправляет SIGHUP всем дочерним процессам — они завершаются. nohup перехватывает SIGHUP и игнорирует его. Вывод перенаправляется в nohup.out. Но для серверов лучше systemd: он обеспечивает автоперезапуск, логирование и управление зависимостями.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер вывода ps aux",
				Description: "Напиши Go-программу, которая парсит вывод команды `ps aux` и находит: top-3 процесса по CPU, top-3 по памяти, и общее количество процессов каждого пользователя.",
				Hints:       "В ps aux: поле 0=USER, 1=PID, 2=%CPU, 3=%MEM, 10+=COMMAND. Используй strings.Fields(). Первая строка — заголовок (пропусти). Сортируй с sort.Slice().",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Process struct {
	User    string
	PID     int
	CPU     float64
	Memory  float64
	Command string
}

// parsePsLine парсит одну строку ps aux (не заголовок)
func parsePsLine(line string) (Process, bool) {
	fields := strings.Fields(line)
	if len(fields) < 11 {
		return Process{}, false
	}
	// TODO: распарси поля
	// fields[0]=USER, [1]=PID, [2]=%CPU, [3]=%MEM, [10:]=COMMAND
	return Process{}, false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var processes []Process
	first := true

	for scanner.Scan() {
		if first {
			first = false // пропускаем заголовок
			continue
		}
		p, ok := parsePsLine(scanner.Text())
		if ok {
			processes = append(processes, p)
		}
	}

	// TODO: выведи top-3 по CPU
	fmt.Println("=== Top CPU ===")

	// TODO: выведи top-3 по Memory
	fmt.Println("=== Top Memory ===")
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Process struct {
	User    string
	PID     int
	CPU     float64
	Memory  float64
	Command string
}

func parsePsLine(line string) (Process, bool) {
	fields := strings.Fields(line)
	if len(fields) < 11 {
		return Process{}, false
	}

	pid, err := strconv.Atoi(fields[1])
	if err != nil {
		return Process{}, false
	}
	cpu, err := strconv.ParseFloat(fields[2], 64)
	if err != nil {
		return Process{}, false
	}
	mem, err := strconv.ParseFloat(fields[3], 64)
	if err != nil {
		return Process{}, false
	}

	cmd := strings.Join(fields[10:], " ")

	return Process{
		User:    fields[0],
		PID:     pid,
		CPU:     cpu,
		Memory:  mem,
		Command: cmd,
	}, true
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var processes []Process
	first := true

	for scanner.Scan() {
		if first {
			first = false
			continue
		}
		p, ok := parsePsLine(scanner.Text())
		if ok {
			processes = append(processes, p)
		}
	}

	fmt.Println("=== Top CPU ===")
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].CPU > processes[j].CPU
	})
	for i := 0; i < 3 && i < len(processes); i++ {
		fmt.Printf("  %5.1f%% %s (%s)\n", processes[i].CPU, processes[i].Command, processes[i].User)
	}

	fmt.Println("=== Top Memory ===")
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].Memory > processes[j].Memory
	})
	for i := 0; i < 3 && i < len(processes); i++ {
		fmt.Printf("  %5.1f%% %s (%s)\n", processes[i].Memory, processes[i].Command, processes[i].User)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "USER       PID %CPU %MEM    VSZ   RSS TTY      STAT START   TIME COMMAND\nroot         1  0.0  0.1 169316 13092 ?        Ss   Jan01   5:30 /sbin/init\npostgres   500 25.0  5.2 500000 53000 ?        Ss   Jan01  10:00 postgres\nnginx      600  2.5  1.0 100000 10240 ?        S    Jan01   2:00 nginx: worker\nuser      1000 50.0 10.0 800000 102400 ?       R    09:00   1:00 go run main.go\n",
						ExpectedOutput: "=== Top CPU ===\n  50.0% go run main.go (user)\n  25.0% postgres (postgres)\n   2.5% nginx: worker (nginx)\n=== Top Memory ===\n  10.0% go run main.go (user)\n   5.2% postgres (postgres)\n   1.0% nginx: worker (nginx)\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "PID", Definition: "Process ID — уникальный номер процесса в системе."},
					{Term: "Signal", Definition: "Механизм IPC: ядро или процесс отправляет числовое уведомление другому процессу."},
					{Term: "Zombie", Definition: "Процесс завершился, но его exit status не прочитан родителем. Занимает только PID."},
				},
			},
			{
				Title:       "Симулятор простого планировщика процессов",
				Description: "Напиши Go-программу, которая симулирует Round-Robin планировщик: принимает список процессов с их burst time (время CPU) и квант времени, выводит порядок выполнения и время ожидания каждого процесса.",
				Hints:       "Round-Robin: каждый процесс получает квант времени. Если не закончил — идёт в конец очереди. Используй slice как очередь. Считай waiting time = время начала текущего кванта - время прибытия.",
				Difficulty:  "hard",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	Name      string
	BurstTime int // сколько CPU-тиков нужно
	Remaining int // сколько осталось
}

// roundRobin симулирует RR-планировщик
// Возвращает лог выполнения: "A A B A B C C" (кто работал каждый тик)
func roundRobin(tasks []Task, quantum int) string {
	// TODO:
	// 1. Пока есть незавершённые задачи:
	//    - Берём первую из очереди
	//    - Даём ей min(quantum, remaining) тиков
	//    - Записываем в лог
	//    - Если remaining > 0 — в конец очереди
	// 2. Возвращаем лог
	return ""
}

func main() {
	// Вход: "A:4 B:3 C:2" quantum=2
	if len(os.Args) < 3 {
		fmt.Println("Usage: scheduler <tasks> <quantum>")
		fmt.Println("  scheduler 'A:4 B:3 C:2' 2")
		return
	}

	parts := strings.Fields(os.Args[1])
	quantum, _ := strconv.Atoi(os.Args[2])

	var tasks []Task
	for _, p := range parts {
		kv := strings.Split(p, ":")
		burst, _ := strconv.Atoi(kv[1])
		tasks = append(tasks, Task{Name: kv[0], BurstTime: burst, Remaining: burst})
	}

	log := roundRobin(tasks, quantum)
	fmt.Println(log)
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Task struct {
	Name      string
	BurstTime int
	Remaining int
}

func roundRobin(tasks []Task, quantum int) string {
	queue := make([]int, len(tasks))
	for i := range tasks {
		queue[i] = i
	}

	var log []string
	for len(queue) > 0 {
		idx := queue[0]
		queue = queue[1:]

		ticks := quantum
		if tasks[idx].Remaining < ticks {
			ticks = tasks[idx].Remaining
		}

		for t := 0; t < ticks; t++ {
			log = append(log, tasks[idx].Name)
		}
		tasks[idx].Remaining -= ticks

		if tasks[idx].Remaining > 0 {
			queue = append(queue, idx)
		}
	}

	return strings.Join(log, " ")
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: scheduler <tasks> <quantum>")
		fmt.Println("  scheduler 'A:4 B:3 C:2' 2")
		return
	}

	parts := strings.Fields(os.Args[1])
	quantum, _ := strconv.Atoi(os.Args[2])

	var tasks []Task
	for _, p := range parts {
		kv := strings.Split(p, ":")
		burst, _ := strconv.Atoi(kv[1])
		tasks = append(tasks, Task{Name: kv[0], BurstTime: burst, Remaining: burst})
	}

	log := roundRobin(tasks, quantum)
	fmt.Println(log)
}`,
				TestCases: []TestCase{
					{
						Input:          "A:4 B:3 C:2 2",
						ExpectedOutput: "A A B B C C A A B\n",
					},
					{
						Input:          "X:3 Y:1 1",
						ExpectedOutput: "X Y X X\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Scheduler", Definition: "Планировщик — компонент ядра, решающий какой процесс получит CPU и на сколько."},
					{Term: "Round-Robin", Definition: "Алгоритм планирования: каждый процесс получает равный квант времени по кругу."},
					{Term: "Context Switch", Definition: "Переключение CPU между процессами. Ядро сохраняет/восстанавливает регистры. Дорогая операция."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 6: Systemd & Services
// ══════════════════════════════════════════════════════════════════

func linuxLesson06Systemd() L {
	return L{
		Slug: "systemd-services", Title: "Systemd и управление сервисами", Order: 6,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Systemd и управление сервисами</h1>

<h2>Что такое systemd?</h2>
<p><strong>systemd</strong> — init-система и менеджер сервисов в большинстве современных Linux-дистрибутивов. Он запускается первым (PID 1) и управляет всеми остальными сервисами.</p>

<p>systemd заменил старый SysVinit потому что:</p>
<ul>
<li><strong>Параллельный запуск</strong> — сервисы стартуют одновременно (быстрая загрузка)</li>
<li><strong>Зависимости</strong> — nginx стартует только после network.target</li>
<li><strong>Автоперезапуск</strong> — упавший сервис перезапускается автоматически</li>
<li><strong>Cgroups</strong> — ограничение ресурсов (CPU, RAM) для сервиса</li>
<li><strong>Журналирование</strong> — все логи в одном месте (journald)</li>
</ul>

<h2>systemctl — управление сервисами</h2>
<pre><code># Базовые операции
systemctl start nginx          # запустить
systemctl stop nginx           # остановить
systemctl restart nginx        # перезапустить (stop + start)
systemctl reload nginx         # перечитать конфиг (без downtime)
systemctl status nginx         # статус + последние логи

# Автозагрузка
systemctl enable nginx         # включить автозапуск при boot
systemctl disable nginx        # отключить автозапуск
systemctl enable --now nginx   # enable + start одной командой

# Просмотр
systemctl list-units --type=service              # все запущенные сервисы
systemctl list-units --type=service --all        # все (включая неактивные)
systemctl list-unit-files --type=service         # все файлы юнитов
systemctl is-active nginx                        # active/inactive
systemctl is-enabled nginx                       # enabled/disabled</code></pre>

<h2>Unit-файлы — конфигурация сервисов</h2>
<pre><code># Расположение:
# /etc/systemd/system/    — пользовательские (приоритет!)
# /lib/systemd/system/    — из пакетов (не редактировать!)

# Пример: /etc/systemd/system/myapp.service
[Unit]
Description=My Go Application
Documentation=https://github.com/me/myapp
After=network.target postgresql.service
Requires=postgresql.service

[Service]
Type=simple
User=appuser
Group=appuser
WorkingDirectory=/opt/myapp
ExecStart=/opt/myapp/server
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure
RestartSec=5
LimitNOFILE=65535

# Безопасность
ProtectSystem=strict
ProtectHome=true
NoNewPrivileges=true
ReadWritePaths=/var/lib/myapp /var/log/myapp

# Environment
Environment=PORT=8080
EnvironmentFile=/etc/myapp/env

[Install]
WantedBy=multi-user.target</code></pre>

<h2>Ключевые параметры [Service]</h2>
<table>
<tr><th>Параметр</th><th>Значение</th></tr>
<tr><td>Type=simple</td><td>Основной процесс = ExecStart (по умолчанию)</td></tr>
<tr><td>Type=forking</td><td>Процесс форкается (как старые демоны)</td></tr>
<tr><td>Type=oneshot</td><td>Запустить и завершиться (миграции, скрипты)</td></tr>
<tr><td>Restart=on-failure</td><td>Перезапуск при ненулевом exit code</td></tr>
<tr><td>Restart=always</td><td>Перезапуск при любом завершении</td></tr>
<tr><td>RestartSec=5</td><td>Пауза 5 сек перед перезапуском</td></tr>
<tr><td>User=appuser</td><td>Запуск от указанного пользователя</td></tr>
<tr><td>LimitNOFILE=65535</td><td>Лимит открытых файлов</td></tr>
</table>

<h2>journalctl — чтение логов</h2>
<pre><code># Логи конкретного сервиса
journalctl -u nginx                    # все логи nginx
journalctl -u nginx -f                 # follow (как tail -f)
journalctl -u nginx --since "1 hour ago"
journalctl -u nginx --since today
journalctl -u nginx -n 100            # последние 100 строк

# Системные логи
journalctl -b                          # с последней загрузки
journalctl -b -1                       # предыдущая загрузка
journalctl -p err                      # только ошибки
journalctl -k                          # kernel messages (dmesg)

# Поиск
journalctl -u nginx --grep "error"
journalctl --disk-usage                # сколько места занимают логи</code></pre>

<h2>Targets (ранлевелы)</h2>
<pre><code># Target = группа сервисов (аналог runlevel)
systemctl get-default                  # текущий target
systemctl set-default multi-user.target  # загрузка без GUI
systemctl set-default graphical.target   # загрузка с GUI

# Основные targets:
# poweroff.target  = shutdown
# rescue.target    = single-user (восстановление)
# multi-user.target = консольный сервер (стандарт)
# graphical.target  = с GUI</code></pre>

<h2>Рабочий процесс создания сервиса</h2>
<pre><code># 1. Создать unit-файл
sudo vim /etc/systemd/system/myapp.service

# 2. Перечитать конфигурацию (ОБЯЗАТЕЛЬНО после изменений!)
sudo systemctl daemon-reload

# 3. Запустить и включить автозагрузку
sudo systemctl enable --now myapp

# 4. Проверить
systemctl status myapp
journalctl -u myapp -f</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что делает systemctl daemon-reload?",
				Options:     []string{"Перезапускает все сервисы", "Перечитывает файлы юнитов с диска (после изменений в .service файлах)", "Перезагружает систему", "Обновляет systemd до новой версии"},
				Correct:     1,
				Explanation: "systemd кэширует конфигурацию юнитов в памяти. Если ты изменил .service файл — systemd не знает об этом до daemon-reload. Без reload изменения не применятся. Это НЕ перезапускает сервисы — только перечитывает конфиги. Потом нужен restart для применения.",
			},
			{
				Question:    "Чем Restart=on-failure отличается от Restart=always?",
				Options:     []string{"Ничем", "on-failure перезапускает только при ненулевом exit code (ошибка), always — при любом завершении (включая нормальное)", "always быстрее перезапускает", "on-failure работает только для root"},
				Correct:     1,
				Explanation: "on-failure: перезапуск при exit code != 0, сигналах (кроме SIGHUP/SIGINT/SIGTERM), таймауте. НЕ перезапускает при `systemctl stop` или нормальном выходе (exit 0). always: перезапуск ВСЕГДА кроме `systemctl stop`. Для серверов обычно on-failure (чтобы не перезапускать при graceful shutdown).",
			},
			{
				Question:    "Где создавать собственные unit-файлы?",
				Options:     []string{"/lib/systemd/system/ (там остальные)", "/etc/systemd/system/ (пользовательские, имеют приоритет)", "/usr/systemd/", "/var/systemd/"},
				Correct:     1,
				Explanation: "/etc/systemd/system/ — для пользовательских юнитов и оверрайдов. /lib/systemd/system/ — файлы из пакетов (dpkg/rpm). Файлы из /etc/ имеют приоритет. Не редактируй /lib/ напрямую — при обновлении пакета изменения потеряются. Используй systemctl edit для создания override.",
			},
		},

		Tasks: []T{
			{
				Title:       "Генератор systemd unit-файла",
				Description: "Напиши Go-программу, которая по параметрам (имя, бинарь, пользователь, порт) генерирует корректный systemd .service файл для Go-приложения.",
				Hints:       "Используй text/template. Шаблон должен включать [Unit], [Service], [Install]. Не забудь After=network.target, Restart=on-failure, Environment для порта.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type ServiceConfig struct {
	Name        string
	Description string
	ExecStart   string
	User        string
	Port        string
	WorkDir     string
}

const unitTemplate = ` + "`" + `# TODO: напиши шаблон systemd unit-файла
# Секции: [Unit], [Service], [Install]
# Используй: .Name, .Description, .ExecStart, .User, .Port, .WorkDir
` + "`" + `

func generateUnit(cfg ServiceConfig) string {
	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	var sb strings.Builder
	tmpl.Execute(&sb, cfg)
	return sb.String()
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: unitgen <name> <exec> <user> <port>")
		return
	}
	cfg := ServiceConfig{
		Name:        os.Args[1],
		Description: os.Args[1] + " service",
		ExecStart:   os.Args[2],
		User:        os.Args[3],
		Port:        os.Args[4],
		WorkDir:     "/opt/" + os.Args[1],
	}
	fmt.Print(generateUnit(cfg))
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strings"
	"text/template"
)

type ServiceConfig struct {
	Name        string
	Description string
	ExecStart   string
	User        string
	Port        string
	WorkDir     string
}

const unitTemplate = ` + "`" + `[Unit]
Description={{.Description}}
After=network.target

[Service]
Type=simple
User={{.User}}
Group={{.User}}
WorkingDirectory={{.WorkDir}}
ExecStart={{.ExecStart}}
Restart=on-failure
RestartSec=5
Environment=PORT={{.Port}}

[Install]
WantedBy=multi-user.target
` + "`" + `

func generateUnit(cfg ServiceConfig) string {
	tmpl, err := template.New("unit").Parse(unitTemplate)
	if err != nil {
		return "ERROR: " + err.Error()
	}
	var sb strings.Builder
	tmpl.Execute(&sb, cfg)
	return sb.String()
}

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: unitgen <name> <exec> <user> <port>")
		return
	}
	cfg := ServiceConfig{
		Name:        os.Args[1],
		Description: os.Args[1] + " service",
		ExecStart:   os.Args[2],
		User:        os.Args[3],
		Port:        os.Args[4],
		WorkDir:     "/opt/" + os.Args[1],
	}
	fmt.Print(generateUnit(cfg))
}`,
				TestCases: []TestCase{
					{
						Input:          "myapp /opt/myapp/server appuser 8080",
						ExpectedOutput: "[Unit]\nDescription=myapp service\nAfter=network.target\n\n[Service]\nType=simple\nUser=appuser\nGroup=appuser\nWorkingDirectory=/opt/myapp\nExecStart=/opt/myapp/server\nRestart=on-failure\nRestartSec=5\nEnvironment=PORT=8080\n\n[Install]\nWantedBy=multi-user.target\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "systemd", Definition: "Init-система Linux (PID 1). Управляет запуском, зависимостями и мониторингом всех сервисов."},
					{Term: "Unit file", Definition: ".service файл, описывающий как запускать, останавливать и перезапускать сервис."},
					{Term: "journald", Definition: "Подсистема логирования systemd. Собирает stdout/stderr всех сервисов. Доступ через journalctl."},
				},
			},
			{
				Title:       "Парсер вывода systemctl status",
				Description: "Напиши программу, которая парсит вывод `systemctl status` и извлекает ключевые поля: имя, статус (active/inactive/failed), PID основного процесса, uptime.",
				Hints:       "Вывод status содержит строки вида 'Active: active (running) since ...', 'Main PID: 1234 (nginx)'. Используй strings.Contains и strings.Fields для извлечения данных.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ServiceStatus struct {
	Name   string
	State  string // "active", "inactive", "failed"
	Sub    string // "running", "dead", "exited"
	PID    string
	Memory string
}

// parseSystemctlStatus парсит вывод systemctl status
func parseSystemctlStatus(lines []string) ServiceStatus {
	var status ServiceStatus
	// TODO:
	// Первая строка: "● nginx.service - A high performance web server"
	// "Active: active (running) since ..."
	// "Main PID: 1234 (nginx)"
	// "Memory: 10.5M"
	return status
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	s := parseSystemctlStatus(lines)
	fmt.Printf("Service: %s\nState: %s (%s)\nPID: %s\nMemory: %s\n",
		s.Name, s.State, s.Sub, s.PID, s.Memory)
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ServiceStatus struct {
	Name   string
	State  string
	Sub    string
	PID    string
	Memory string
}

func parseSystemctlStatus(lines []string) ServiceStatus {
	var status ServiceStatus
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasSuffix(line, ".service") || strings.Contains(line, ".service -") {
			line = strings.TrimPrefix(line, "● ")
			line = strings.TrimPrefix(line, "○ ")
			parts := strings.SplitN(line, " - ", 2)
			status.Name = strings.TrimSpace(parts[0])
		}

		if strings.HasPrefix(line, "Active:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				status.State = parts[1]
				status.Sub = strings.Trim(parts[2], "()")
			}
		}

		if strings.HasPrefix(line, "Main PID:") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				status.PID = parts[2]
			}
		}

		if strings.HasPrefix(line, "Memory:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				status.Memory = parts[1]
			}
		}
	}
	return status
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	s := parseSystemctlStatus(lines)
	fmt.Printf("Service: %s\nState: %s (%s)\nPID: %s\nMemory: %s\n",
		s.Name, s.State, s.Sub, s.PID, s.Memory)
}`,
				TestCases: []TestCase{
					{
						Input:          "● nginx.service - A high performance web server\n   Active: active (running) since Mon 2024-01-15 10:30:00 UTC; 5 days ago\n   Main PID: 1234 (nginx)\n   Memory: 10.5M\n",
						ExpectedOutput: "Service: nginx.service\nState: active (running)\nPID: 1234\nMemory: 10.5M\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Target", Definition: "Группа юнитов systemd (аналог runlevel). multi-user.target = консольный сервер."},
					{Term: "daemon-reload", Definition: "Команда systemd перечитать файлы юнитов. Обязательна после изменений .service файлов."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 7: Package Management
// ══════════════════════════════════════════════════════════════════

func linuxLesson07Packages() L {
	return L{
		Slug: "package-management", Title: "Управление пакетами", Order: 7,
		Difficulty: "beginner", Track: "devops",
		Content: `<h1>Управление пакетами в Linux</h1>

<h2>Зачем нужен пакетный менеджер?</h2>
<p>Без пакетного менеджера установка программы = скачать исходники, установить зависимости вручную, скомпилировать, разложить по каталогам. Пакетный менеджер делает это за тебя:</p>
<ul>
<li><strong>Разрешение зависимостей</strong> — автоматически ставит всё необходимое</li>
<li><strong>Обновления</strong> — одна команда обновляет всё</li>
<li><strong>Удаление</strong> — чистое удаление без мусора</li>
<li><strong>Верификация</strong> — проверяет GPG-подписи пакетов</li>
</ul>

<h2>APT (Debian/Ubuntu)</h2>
<pre><code># Обновить список пакетов (из репозиториев)
sudo apt update                        # обязательно перед install!

# Установить
sudo apt install nginx postgresql-16 htop

# Обновить все пакеты
sudo apt upgrade                       # безопасное обновление
sudo apt full-upgrade                  # + удаляет старые пакеты если нужно

# Удалить
sudo apt remove nginx                  # удалить пакет (конфиги остаются)
sudo apt purge nginx                   # удалить + конфиги
sudo apt autoremove                    # удалить ненужные зависимости

# Поиск и информация
apt search redis                       # найти пакет
apt show nginx                         # информация о пакете
apt list --installed                   # все установленные
apt list --upgradable                  # доступные обновления

# Конкретная версия
apt install nginx=1.24.0-1ubuntu1
apt-mark hold nginx                    # заморозить версию (не обновлять!)
apt-mark unhold nginx                  # разморозить</code></pre>

<h2>DNF (RHEL/Rocky/Alma/Fedora)</h2>
<pre><code># Аналоги APT-команд
sudo dnf install nginx
sudo dnf remove nginx
sudo dnf update                        # = apt upgrade
sudo dnf search redis
sudo dnf info nginx
sudo dnf list installed

# Группы пакетов
sudo dnf group install "Development Tools"

# Модули (SCL)
sudo dnf module enable postgresql:16
sudo dnf module install postgresql:16</code></pre>

<h2>APK (Alpine Linux)</h2>
<pre><code># Минималистичный менеджер для Alpine (Docker!)
apk update                             # обновить индекс
apk add nginx curl go                  # установить
apk del nginx                          # удалить
apk search redis                       # найти
apk info nginx                         # информация

# В Dockerfile (Alpine):
RUN apk add --no-cache curl ca-certificates
# --no-cache = не сохранять кэш индекса (меньше размер образа)</code></pre>

<h2>Репозитории</h2>
<pre><code># APT: /etc/apt/sources.list или /etc/apt/sources.list.d/
# Формат: deb http://archive.ubuntu.com/ubuntu jammy main universe

# Добавить PPA (Ubuntu)
sudo add-apt-repository ppa:deadsnakes/ppa

# Добавить внешний репозиторий (с GPG-ключом)
curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker.gpg
echo "deb [signed-by=/usr/share/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu jammy stable" | sudo tee /etc/apt/sources.list.d/docker.list
sudo apt update</code></pre>

<h2>Version Pinning — зачем и как</h2>
<p>На проде нельзя бесконтрольно обновлять пакеты — может сломаться совместимость:</p>
<pre><code># APT: заморозить версию
sudo apt-mark hold postgresql-16
# Теперь apt upgrade НЕ обновит postgres

# APT: файл предпочтений /etc/apt/preferences.d/
Package: postgresql-16
Pin: version 16.2*
Pin-Priority: 1001

# DNF: versionlock
sudo dnf install dnf-plugin-versionlock
sudo dnf versionlock add postgresql-16</code></pre>

<p><strong>Продовый совет:</strong> На серверах: отключи unattended-upgrades для критичных пакетов (PostgreSQL, Redis). Обновляй вручную после тестирования. Используй <code>apt-mark hold</code> для закрепления версий.</p>

<h2>Практика: типичный setup сервера</h2>
<pre><code># 1. Обновить всё
sudo apt update && sudo apt upgrade -y

# 2. Установить базовые утилиты
sudo apt install -y curl wget git htop tree jq unzip net-tools

# 3. Установить основной софт
sudo apt install -y nginx postgresql-16 redis-server

# 4. Заморозить версии критичных пакетов
sudo apt-mark hold postgresql-16 redis-server</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Почему нужно делать apt update перед apt install?",
				Options:     []string{"Для красоты", "Чтобы обновить локальный индекс пакетов — без этого apt не знает о новых версиях и может не найти пакет", "Чтобы обновить сам apt", "Это необязательно"},
				Correct:     1,
				Explanation: "apt хранит локальный кэш списка пакетов (/var/lib/apt/lists/). Без update он может быть устаревшим: новые пакеты не найдутся, версии будут старые. На CI/Docker всегда делаем: RUN apt-get update && apt-get install -y ... в одном RUN для консистентности.",
			},
			{
				Question:    "Чем apt remove отличается от apt purge?",
				Options:     []string{"Ничем", "remove удаляет программу но оставляет конфигурационные файлы, purge удаляет всё включая конфиги", "purge удаляет и данные пользователя", "remove не удаляет зависимости"},
				Correct:     1,
				Explanation: "remove: удаляет бинарники и данные пакета. Конфиги в /etc/ остаются (удобно при переустановке). purge: удаляет всё, включая конфиги. Для полной очистки: apt purge nginx && apt autoremove (autoremove удалит зависимости, которые больше никому не нужны).",
			},
			{
				Question:    "Зачем в Docker Alpine используют apk add --no-cache?",
				Options:     []string{"Для ускорения установки", "Чтобы не сохранять индекс пакетов в образе — уменьшает размер Docker-образа", "Чтобы пакеты не кешировались в RAM", "Это баг Alpine"},
				Correct:     1,
				Explanation: "Без --no-cache Alpine сохраняет индекс (/var/cache/apk/) в каждом слое Docker-образа. Это лишние 5-10MB. С --no-cache индекс скачивается, используется и удаляется. В Debian-based: RUN apt-get update && apt-get install -y ... && rm -rf /var/lib/apt/lists/* (аналогичная оптимизация).",
			},
			{
				Question:    "Для чего нужен apt-mark hold?",
				Options:     []string{"Заморозить версию пакета — apt upgrade не будет его обновлять", "Приостановить загрузку пакета", "Отметить пакет как любимый", "Заблокировать удаление пакета"},
				Correct:     0,
				Explanation: "hold 'замораживает' пакет: apt upgrade пропустит его. Это критично на проде: если PostgreSQL обновится с 16.2 до 16.3 без тестирования — может сломаться совместимость. Стратегия: hold критичные пакеты, обновлять их вручную после проверки на staging.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер apt list --installed",
				Description: "Напиши Go-программу, которая парсит вывод `apt list --installed` и выводит: имя пакета, версию, архитектуру. Также считает общее количество установленных пакетов.",
				Hints:       "Формат строки: name/repository version arch [installed,automatic]. Разделяй по '/' и ' '. Пропускай строки 'Listing...'.",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Package struct {
	Name    string
	Version string
	Arch    string
}

// parseAptLine парсит строку формата "nginx/jammy,now 1.24.0-1 amd64 [installed]"
func parseAptLine(line string) (Package, bool) {
	// TODO: парси строку
	// 1. Пропусти "Listing..." и пустые строки
	// 2. Имя — до первого '/'
	// 3. Версия — второе "слово" после пробела
	// 4. Архитектура — третье "слово"
	return Package{}, false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	count := 0
	for scanner.Scan() {
		pkg, ok := parseAptLine(scanner.Text())
		if !ok {
			continue
		}
		count++
		fmt.Printf("%-30s %-20s %s\n", pkg.Name, pkg.Version, pkg.Arch)
	}
	fmt.Printf("\nTotal: %d packages\n", count)
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Package struct {
	Name    string
	Version string
	Arch    string
}

func parseAptLine(line string) (Package, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "Listing") {
		return Package{}, false
	}

	slashIdx := strings.Index(line, "/")
	if slashIdx == -1 {
		return Package{}, false
	}
	name := line[:slashIdx]

	rest := line[slashIdx+1:]
	parts := strings.Fields(rest)
	if len(parts) < 3 {
		return Package{}, false
	}

	return Package{
		Name:    name,
		Version: parts[1],
		Arch:    parts[2],
	}, true
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	count := 0
	for scanner.Scan() {
		pkg, ok := parseAptLine(scanner.Text())
		if !ok {
			continue
		}
		count++
		fmt.Printf("%-30s %-20s %s\n", pkg.Name, pkg.Version, pkg.Arch)
	}
	fmt.Printf("\nTotal: %d packages\n", count)
}`,
				TestCases: []TestCase{
					{
						Input:          "Listing...\nnginx/jammy,now 1.24.0-1 amd64 [installed]\ncurl/jammy,now 7.81.0-1 amd64 [installed,automatic]\n",
						ExpectedOutput: "nginx                          1.24.0-1             amd64\ncurl                           7.81.0-1             amd64\n\nTotal: 2 packages\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Repository", Definition: "Удалённый сервер с .deb/.rpm пакетами. apt update скачивает индекс пакетов оттуда."},
					{Term: "Dependency", Definition: "Пакет, необходимый для работы другого пакета. Пакетный менеджер устанавливает их автоматически."},
				},
			},
			{
				Title:       "Сравнение версий пакетов (semver)",
				Description: "Напиши Go-программу, которая сравнивает две версии пакетов в формате semver (major.minor.patch) и определяет какая новее, а также тип обновления (major/minor/patch).",
				Hints:       "Разбей версию по '.', конвертируй в числа. Сравнивай слева направо: сначала major, потом minor, потом patch. Тип обновления определяется по тому, какая часть изменилась.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch int
}

// parseVersion парсит "1.24.3" в структуру
func parseVersion(s string) (Version, error) {
	// TODO: разбей по '.', конвертируй в int
	return Version{}, nil
}

// compareVersions возвращает:
//  1 если a > b, -1 если a < b, 0 если равны
func compareVersions(a, b Version) int {
	// TODO
	return 0
}

// upgradeType возвращает "major", "minor", "patch" или "none"
func upgradeType(from, to Version) string {
	// TODO: определи тип обновления
	return "none"
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vercmp <version1> <version2>")
		return
	}
	v1, err1 := parseVersion(os.Args[1])
	v2, err2 := parseVersion(os.Args[2])
	if err1 != nil || err2 != nil {
		fmt.Println("Invalid version format")
		return
	}

	cmp := compareVersions(v1, v2)
	switch {
	case cmp > 0:
		fmt.Printf("%s > %s\n", os.Args[1], os.Args[2])
	case cmp < 0:
		fmt.Printf("%s < %s\n", os.Args[1], os.Args[2])
		fmt.Printf("Upgrade type: %s\n", upgradeType(v1, v2))
	default:
		fmt.Printf("%s == %s\n", os.Args[1], os.Args[2])
	}
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Version struct {
	Major, Minor, Patch int
}

func parseVersion(s string) (Version, error) {
	parts := strings.Split(s, ".")
	if len(parts) != 3 {
		return Version{}, fmt.Errorf("invalid version: %s", s)
	}
	major, err := strconv.Atoi(parts[0])
	if err != nil {
		return Version{}, err
	}
	minor, err := strconv.Atoi(parts[1])
	if err != nil {
		return Version{}, err
	}
	patch, err := strconv.Atoi(parts[2])
	if err != nil {
		return Version{}, err
	}
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

func compareVersions(a, b Version) int {
	if a.Major != b.Major {
		if a.Major > b.Major {
			return 1
		}
		return -1
	}
	if a.Minor != b.Minor {
		if a.Minor > b.Minor {
			return 1
		}
		return -1
	}
	if a.Patch != b.Patch {
		if a.Patch > b.Patch {
			return 1
		}
		return -1
	}
	return 0
}

func upgradeType(from, to Version) string {
	if to.Major > from.Major {
		return "major"
	}
	if to.Minor > from.Minor {
		return "minor"
	}
	if to.Patch > from.Patch {
		return "patch"
	}
	return "none"
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: vercmp <version1> <version2>")
		return
	}
	v1, err1 := parseVersion(os.Args[1])
	v2, err2 := parseVersion(os.Args[2])
	if err1 != nil || err2 != nil {
		fmt.Println("Invalid version format")
		return
	}

	cmp := compareVersions(v1, v2)
	switch {
	case cmp > 0:
		fmt.Printf("%s > %s\n", os.Args[1], os.Args[2])
	case cmp < 0:
		fmt.Printf("%s < %s\n", os.Args[1], os.Args[2])
		fmt.Printf("Upgrade type: %s\n", upgradeType(v1, v2))
	default:
		fmt.Printf("%s == %s\n", os.Args[1], os.Args[2])
	}
}`,
				TestCases: []TestCase{
					{Input: "1.24.0 1.25.1", ExpectedOutput: "1.24.0 < 1.25.1\nUpgrade type: minor\n"},
					{Input: "2.0.0 1.9.9", ExpectedOutput: "2.0.0 > 1.9.9\n"},
					{Input: "1.0.0 1.0.0", ExpectedOutput: "1.0.0 == 1.0.0\n"},
					{Input: "1.0.0 2.0.0", ExpectedOutput: "1.0.0 < 2.0.0\nUpgrade type: major\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "Semver", Definition: "Semantic Versioning: MAJOR.MINOR.PATCH. Major=breaking changes, Minor=новые фичи, Patch=багфиксы."},
					{Term: "Version pinning", Definition: "Фиксация версии пакета для предотвращения неожиданных обновлений на проде."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 8: Networking Basics
// ══════════════════════════════════════════════════════════════════

func linuxLesson08Networking() L {
	return L{
		Slug: "networking-basics", Title: "Основы сетей в Linux", Order: 8,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Основы сетей в Linux</h1>

<h2>Сетевые интерфейсы</h2>
<pre><code># Посмотреть интерфейсы и IP
ip addr show            # или ip a
ip addr show eth0       # конкретный интерфейс

# Что покажет:
# eth0: <BROADCAST,MULTICAST,UP,LOWER_UP>
#     inet 192.168.1.10/24 brd 192.168.1.255 scope global eth0
#     inet6 fe80::1/64 scope link

# Старая команда (deprecated, но встречается):
ifconfig</code></pre>

<h2>Порты и соединения</h2>
<pre><code># ss (Socket Statistics) — замена netstat
ss -tlnp                     # TCP, Listening, Numeric, Process
# State    Recv-Q  Send-Q  Local Address:Port  Peer Address:Port  Process
# LISTEN   0       128     0.0.0.0:80          0.0.0.0:*          users:(("nginx",pid=1234))
# LISTEN   0       128     0.0.0.0:5432        0.0.0.0:*          users:(("postgres",pid=500))

ss -tlnp | grep :8080       # кто слушает на порту 8080
ss -tnp                      # все TCP-соединения (не только listening)
ss -s                        # статистика (сколько соединений)

# netstat (устаревшая, но знать надо)
netstat -tlnp                # аналог ss -tlnp</code></pre>

<h2>Ключевые порты</h2>
<table>
<tr><th>Порт</th><th>Сервис</th><th>Протокол</th></tr>
<tr><td>22</td><td>SSH</td><td>TCP</td></tr>
<tr><td>80</td><td>HTTP</td><td>TCP</td></tr>
<tr><td>443</td><td>HTTPS</td><td>TCP</td></tr>
<tr><td>5432</td><td>PostgreSQL</td><td>TCP</td></tr>
<tr><td>6379</td><td>Redis</td><td>TCP</td></tr>
<tr><td>3306</td><td>MySQL</td><td>TCP</td></tr>
<tr><td>8080</td><td>HTTP alt</td><td>TCP</td></tr>
<tr><td>53</td><td>DNS</td><td>TCP/UDP</td></tr>
</table>

<h2>DNS и /etc/hosts</h2>
<pre><code># Резолв DNS
dig example.com              # подробный DNS-запрос
dig +short example.com       # только IP
nslookup example.com         # простой DNS-запрос
host example.com             # ещё проще

# /etc/hosts — локальный DNS (приоритет над DNS-сервером!)
# 127.0.0.1   localhost
# 192.168.1.50  db.local
# 10.0.0.10    api.internal

# /etc/resolv.conf — какие DNS-серверы использовать
# nameserver 8.8.8.8
# nameserver 1.1.1.1

# Порядок резолва: /etc/nsswitch.conf
# hosts: files dns   ← сначала /etc/hosts, потом DNS</code></pre>

<h2>Маршрутизация</h2>
<pre><code># Таблица маршрутов
ip route show              # или ip r
# default via 192.168.1.1 dev eth0    ← шлюз по умолчанию
# 192.168.1.0/24 dev eth0 scope link  ← локальная сеть

# Трассировка пути до хоста
traceroute example.com     # показывает каждый hop
mtr example.com            # интерактивный traceroute + ping</code></pre>

<h2>Firewall: iptables / nftables</h2>
<pre><code># iptables — классический файрвол
# Три цепочки: INPUT (входящие), OUTPUT (исходящие), FORWARD (проходящие)

# Посмотреть правила
sudo iptables -L -n -v

# Разрешить SSH
sudo iptables -A INPUT -p tcp --dport 22 -j ACCEPT

# Разрешить HTTP/HTTPS
sudo iptables -A INPUT -p tcp --dport 80 -j ACCEPT
sudo iptables -A INPUT -p tcp --dport 443 -j ACCEPT

# Заблокировать всё остальное
sudo iptables -A INPUT -j DROP

# UFW — упрощённый фронтенд для iptables (Ubuntu)
sudo ufw enable
sudo ufw allow 22/tcp
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw deny from 10.0.0.5
sudo ufw status</code></pre>

<h2>Диагностика сети</h2>
<pre><code># Ping — проверка доступности
ping -c 4 example.com       # 4 пакета и выход

# curl — HTTP-запросы
curl -I https://example.com       # только заголовки
curl -v https://example.com       # verbose (подробно)
curl -o /dev/null -s -w "%{http_code} %{time_total}s\n" https://example.com

# wget — скачивание
wget https://example.com/file.tar.gz

# Проверить открыт ли порт
nc -zv example.com 443      # netcat — проверка порта
timeout 3 bash -c "echo > /dev/tcp/example.com/443" && echo OK</code></pre>

<p><strong>Продовый совет:</strong> Не открывай порты наружу без необходимости. PostgreSQL (5432) и Redis (6379) НИКОГДА не должны быть доступны из интернета напрямую. Используй SSH-туннели или VPN для доступа к внутренним сервисам.</p>`,

		Quiz: []Q{
			{
				Question:    "Что показывает команда ss -tlnp?",
				Options:     []string{"Все файлы в системе", "TCP-сокеты в состоянии LISTEN с номерами портов и PID процессов", "Сетевой трафик в реальном времени", "DNS-записи"},
				Correct:     1,
				Explanation: "ss = Socket Statistics. Флаги: -t=TCP, -l=Listening (только слушающие), -n=Numeric (порты числом, не именем), -p=Process (показать PID). Это первая команда при 'почему не подключиться к сервису': проверяем что он слушает на нужном порту и нужном интерфейсе (0.0.0.0 vs 127.0.0.1).",
			},
			{
				Question:    "Почему сервис на 127.0.0.1:5432 недоступен с другого сервера?",
				Options:     []string{"Порт закрыт файрволом", "127.0.0.1 (loopback) принимает соединения ТОЛЬКО с localhost — нужно слушать на 0.0.0.0 или конкретном IP", "DNS не работает", "Нужен HTTPS"},
				Correct:     1,
				Explanation: "127.0.0.1 = loopback = только локальные соединения. Если PostgreSQL слушает на 127.0.0.1:5432, то другие серверы не подключатся. Нужно: listen_addresses = '0.0.0.0' (все интерфейсы) или конкретный IP. НО: перед этим настрой pg_hba.conf и firewall! Не открывай базу наружу.",
			},
			{
				Question:    "В каком порядке Linux резолвит hostname?",
				Options:     []string{"Только DNS-сервер", "Сначала /etc/hosts, потом DNS-серверы из /etc/resolv.conf (порядок в /etc/nsswitch.conf)", "Сначала DNS, потом /etc/hosts", "Случайным образом"},
				Correct:     1,
				Explanation: "/etc/nsswitch.conf определяет порядок: 'hosts: files dns'. files = /etc/hosts (приоритет!), потом dns = серверы из /etc/resolv.conf. Поэтому запись в /etc/hosts переопределяет любой DNS. Это используется для: блокировки сайтов, локальных dev-окружений, service discovery без DNS.",
			},
		},

		Tasks: []T{
			{
				Title:       "Парсер вывода ss -tlnp",
				Description: "Напиши Go-программу, которая парсит вывод `ss -tlnp` и выводит список слушающих сервисов: порт, процесс, адрес. Также определяет потенциально опасные открытые порты.",
				Hints:       "В ss -tlnp: Local Address:Port — 4-е поле. Process — последнее поле, содержит 'users:((\"name\",pid=N,...))'. Опасные порты на 0.0.0.0: 5432, 6379, 3306 (БД не должны торчать наружу).",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ListeningPort struct {
	Address string
	Port    string
	Process string
	Risk    string // "safe", "warning", "danger"
}

// parseSsLine парсит строку вывода ss -tlnp
func parseSsLine(line string) (ListeningPort, bool) {
	// TODO:
	// Пример: "LISTEN 0 128 0.0.0.0:5432 0.0.0.0:* users:((\"postgres\",pid=500,fd=5))"
	// 1. Пропусти заголовок (начинается с "State")
	// 2. Извлеки Local Address:Port (4-е поле по пробелам)
	// 3. Извлеки имя процесса из users:((...))
	// 4. Определи риск: 0.0.0.0 + порт БД = danger
	return ListeningPort{}, false
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lp, ok := parseSsLine(scanner.Text())
		if !ok {
			continue
		}
		fmt.Printf("[%-7s] %s:%-5s %s\n", lp.Risk, lp.Address, lp.Port, lp.Process)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ListeningPort struct {
	Address string
	Port    string
	Process string
	Risk    string
}

var dangerPorts = map[string]bool{
	"5432": true, "6379": true, "3306": true,
	"27017": true, "9200": true,
}

func parseSsLine(line string) (ListeningPort, bool) {
	line = strings.TrimSpace(line)
	if line == "" || strings.HasPrefix(line, "State") {
		return ListeningPort{}, false
	}

	fields := strings.Fields(line)
	if len(fields) < 5 {
		return ListeningPort{}, false
	}

	// Local address is field 3 (0-indexed)
	localAddr := fields[3]
	lastColon := strings.LastIndex(localAddr, ":")
	if lastColon == -1 {
		return ListeningPort{}, false
	}
	addr := localAddr[:lastColon]
	port := localAddr[lastColon+1:]

	// Process name
	process := "unknown"
	for _, f := range fields {
		if strings.Contains(f, "users:") {
			start := strings.Index(f, "\"")
			end := strings.Index(f[start+1:], "\"")
			if start != -1 && end != -1 {
				process = f[start+1 : start+1+end]
			}
			break
		}
	}

	risk := "safe"
	if addr == "0.0.0.0" || addr == "*" || addr == "::" {
		if dangerPorts[port] {
			risk = "danger"
		} else {
			risk = "warning"
		}
	}

	return ListeningPort{
		Address: addr,
		Port:    port,
		Process: process,
		Risk:    risk,
	}, true
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		lp, ok := parseSsLine(scanner.Text())
		if !ok {
			continue
		}
		fmt.Printf("[%-7s] %s:%-5s %s\n", lp.Risk, lp.Address, lp.Port, lp.Process)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "State  Recv-Q Send-Q Local Address:Port  Peer Address:Port Process\nLISTEN 0      128    0.0.0.0:22          0.0.0.0:*         users:((\"sshd\",pid=800,fd=3))\nLISTEN 0      128    127.0.0.1:5432      0.0.0.0:*         users:((\"postgres\",pid=500,fd=5))\nLISTEN 0      128    0.0.0.0:6379        0.0.0.0:*         users:((\"redis\",pid=600,fd=6))\n",
						ExpectedOutput: "[warning] 0.0.0.0:22    sshd\n[safe   ] 127.0.0.1:5432  postgres\n[danger ] 0.0.0.0:6379  redis\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Socket", Definition: "Комбинация IP:Port, идентифицирующая конечную точку сетевого соединения."},
					{Term: "Loopback (127.0.0.1)", Definition: "Виртуальный интерфейс, доступный только локально. Соединения не выходят в сеть."},
					{Term: "Firewall", Definition: "Фильтр сетевых пакетов. В Linux: iptables/nftables. Определяет какой трафик пропускать."},
				},
			},
			{
				Title:       "Парсер /etc/hosts",
				Description: "Напиши Go-программу, которая парсит файл /etc/hosts и реализует простой DNS-резолвинг: по имени хоста возвращает IP, по IP — все связанные имена.",
				Hints:       "Формат: IP hostname [alias...]. Пропускай пустые строки и комментарии (#). Одному IP может соответствовать несколько имён.",
				Difficulty:  "easy",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type HostsDB struct {
	// ip -> []hostnames
	byIP map[string][]string
	// hostname -> ip
	byName map[string]string
}

func newHostsDB() *HostsDB {
	return &HostsDB{
		byIP:   make(map[string][]string),
		byName: make(map[string]string),
	}
}

// parseLine парсит одну строку /etc/hosts
func (db *HostsDB) parseLine(line string) {
	// TODO:
	// 1. Обрежи комментарии (всё после #)
	// 2. Пропусти пустые строки
	// 3. Первое поле — IP, остальные — hostnames
	// 4. Заполни оба map
}

// resolve ищет IP по hostname
func (db *HostsDB) resolve(name string) string {
	// TODO
	return ""
}

// reverseResolve ищет hostnames по IP
func (db *HostsDB) reverseResolve(ip string) []string {
	// TODO
	return nil
}

func main() {
	db := newHostsDB()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		db.parseLine(scanner.Text())
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: hosts-resolve <name-or-ip>")
		return
	}

	query := os.Args[1]
	if ip := db.resolve(query); ip != "" {
		fmt.Printf("%s -> %s\n", query, ip)
	} else if names := db.reverseResolve(query); len(names) > 0 {
		fmt.Printf("%s -> %s\n", query, strings.Join(names, ", "))
	} else {
		fmt.Printf("%s -> not found\n", query)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type HostsDB struct {
	byIP   map[string][]string
	byName map[string]string
}

func newHostsDB() *HostsDB {
	return &HostsDB{
		byIP:   make(map[string][]string),
		byName: make(map[string]string),
	}
}

func (db *HostsDB) parseLine(line string) {
	if idx := strings.Index(line, "#"); idx != -1 {
		line = line[:idx]
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	fields := strings.Fields(line)
	if len(fields) < 2 {
		return
	}

	ip := fields[0]
	for _, name := range fields[1:] {
		db.byIP[ip] = append(db.byIP[ip], name)
		db.byName[name] = ip
	}
}

func (db *HostsDB) resolve(name string) string {
	return db.byName[name]
}

func (db *HostsDB) reverseResolve(ip string) []string {
	return db.byIP[ip]
}

func main() {
	db := newHostsDB()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		db.parseLine(scanner.Text())
	}

	if len(os.Args) < 2 {
		fmt.Println("Usage: hosts-resolve <name-or-ip>")
		return
	}

	query := os.Args[1]
	if ip := db.resolve(query); ip != "" {
		fmt.Printf("%s -> %s\n", query, ip)
	} else if names := db.reverseResolve(query); len(names) > 0 {
		fmt.Printf("%s -> %s\n", query, strings.Join(names, ", "))
	} else {
		fmt.Printf("%s -> not found\n", query)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "127.0.0.1 localhost\n192.168.1.10 db.local postgres.local\n# comment\n10.0.0.1 api.internal\n",
						ExpectedOutput: "db.local -> 192.168.1.10\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "DNS", Definition: "Domain Name System — переводит имена (example.com) в IP-адреса (93.184.216.34)."},
					{Term: "/etc/hosts", Definition: "Локальный файл DNS-маппингов. Имеет приоритет над DNS-серверами."},
					{Term: "Порт", Definition: "Число 0-65535, идентифицирующее конкретный сервис на IP-адресе. 80=HTTP, 443=HTTPS, 22=SSH."},
				},
			},
			{
				Title:       "Калькулятор подсетей (CIDR)",
				Description: "Напиши Go-программу, которая по CIDR-нотации (192.168.1.0/24) вычисляет: сетевой адрес, broadcast, маску подсети, количество доступных хостов.",
				Hints:       "CIDR /24 = маска 255.255.255.0 = 24 единичных бита. Количество хостов = 2^(32-prefix) - 2. Network = IP AND mask. Broadcast = IP OR (NOT mask). Работай с uint32 для IP.",
				Difficulty:  "hard",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SubnetInfo struct {
	Network   string
	Broadcast string
	Mask      string
	Hosts     int
	Prefix    int
}

// parseCIDR парсит "192.168.1.0/24"
func parseCIDR(cidr string) (uint32, int, error) {
	// TODO: разбей на IP и prefix
	// Конвертируй IP в uint32 (4 октета по 8 бит)
	return 0, 0, nil
}

// ipToUint32 конвертирует "192.168.1.0" в uint32
func ipToUint32(ip string) (uint32, error) {
	// TODO
	return 0, nil
}

// uint32ToIP конвертирует uint32 обратно в "192.168.1.0"
func uint32ToIP(n uint32) string {
	// TODO
	return ""
}

// calculateSubnet вычисляет информацию о подсети
func calculateSubnet(cidr string) (SubnetInfo, error) {
	// TODO
	return SubnetInfo{}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: subnet <CIDR>")
		fmt.Println("  subnet 192.168.1.0/24")
		return
	}
	info, err := calculateSubnet(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Network:   %s\n", info.Network)
	fmt.Printf("Broadcast: %s\n", info.Broadcast)
	fmt.Printf("Mask:      %s\n", info.Mask)
	fmt.Printf("Hosts:     %d\n", info.Hosts)
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SubnetInfo struct {
	Network   string
	Broadcast string
	Mask      string
	Hosts     int
	Prefix    int
}

func ipToUint32(ip string) (uint32, error) {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return 0, fmt.Errorf("invalid IP")
	}
	var result uint32
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 || n > 255 {
			return 0, fmt.Errorf("invalid octet")
		}
		result |= uint32(n) << (24 - 8*i)
	}
	return result, nil
}

func uint32ToIP(n uint32) string {
	return fmt.Sprintf("%d.%d.%d.%d",
		(n>>24)&0xFF, (n>>16)&0xFF, (n>>8)&0xFF, n&0xFF)
}

func parseCIDR(cidr string) (uint32, int, error) {
	parts := strings.Split(cidr, "/")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid CIDR")
	}
	ip, err := ipToUint32(parts[0])
	if err != nil {
		return 0, 0, err
	}
	prefix, err := strconv.Atoi(parts[1])
	if err != nil || prefix < 0 || prefix > 32 {
		return 0, 0, fmt.Errorf("invalid prefix")
	}
	return ip, prefix, nil
}

func calculateSubnet(cidr string) (SubnetInfo, error) {
	ip, prefix, err := parseCIDR(cidr)
	if err != nil {
		return SubnetInfo{}, err
	}

	var mask uint32
	if prefix > 0 {
		mask = ^uint32(0) << (32 - prefix)
	}

	network := ip & mask
	broadcast := network | ^mask
	hosts := (1 << (32 - prefix)) - 2
	if prefix >= 31 {
		hosts = 0
	}

	return SubnetInfo{
		Network:   uint32ToIP(network),
		Broadcast: uint32ToIP(broadcast),
		Mask:      uint32ToIP(mask),
		Hosts:     hosts,
		Prefix:    prefix,
	}, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: subnet <CIDR>")
		fmt.Println("  subnet 192.168.1.0/24")
		return
	}
	info, err := calculateSubnet(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Network:   %s\n", info.Network)
	fmt.Printf("Broadcast: %s\n", info.Broadcast)
	fmt.Printf("Mask:      %s\n", info.Mask)
	fmt.Printf("Hosts:     %d\n", info.Hosts)
}`,
				TestCases: []TestCase{
					{Input: "192.168.1.0/24", ExpectedOutput: "Network:   192.168.1.0\nBroadcast: 192.168.1.255\nMask:      255.255.255.0\nHosts:     254\n"},
					{Input: "10.0.0.0/8", ExpectedOutput: "Network:   10.0.0.0\nBroadcast: 10.255.255.255\nMask:      255.0.0.0\nHosts:     16777214\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "CIDR", Definition: "Classless Inter-Domain Routing — /24 означает 24 бита сетевой части, 8 бит хостовой."},
					{Term: "Subnet mask", Definition: "Маска подсети определяет какая часть IP — сеть, а какая — хост. 255.255.255.0 = /24."},
					{Term: "Broadcast", Definition: "Адрес для отправки пакета ВСЕМ хостам в подсети. Последний адрес в диапазоне."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 9: Bash Scripting
// ══════════════════════════════════════════════════════════════════

func linuxLesson09Bash() L {
	return L{
		Slug: "bash-scripting", Title: "Bash-скрипты", Order: 9,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Bash-скрипты</h1>

<h2>Зачем нужны скрипты?</h2>
<p>Bash-скрипт — файл с последовательностью команд. Вместо того чтобы каждый раз набирать 10 команд вручную, ты пишешь скрипт и запускаешь его одной командой. Применения:</p>
<ul>
<li>Автоматизация деплоя</li>
<li>Бэкапы</li>
<li>Мониторинг и алерты</li>
<li>Настройка серверов (provisioning)</li>
<li>CI/CD пайплайны</li>
</ul>

<h2>Основы</h2>
<pre><code>#!/bin/bash
# ^ shebang — указывает какой интерпретатор использовать

# Переменные (без пробелов вокруг =!)
NAME="world"
PORT=8080
echo "Hello, $NAME on port $PORT"

# Чтение ввода
read -p "Enter name: " USERNAME
echo "Hi, $USERNAME"

# Подстановка команд
TODAY=$(date +%Y-%m-%d)
FILES=$(ls | wc -l)
echo "Today: $TODAY, Files: $FILES"</code></pre>

<h2>Условия</h2>
<pre><code># if / elif / else
if [ -f "/etc/nginx/nginx.conf" ]; then
    echo "Nginx config exists"
elif [ -d "/etc/nginx" ]; then
    echo "Nginx dir exists but no config"
else
    echo "Nginx not installed"
fi

# Операторы сравнения
[ "$A" -eq "$B" ]     # числа: equal
[ "$A" -ne "$B" ]     # числа: not equal
[ "$A" -gt "$B" ]     # числа: greater than
[ "$A" -lt "$B" ]     # числа: less than
[ "$A" = "$B" ]       # строки: equal
[ "$A" != "$B" ]      # строки: not equal
[ -z "$A" ]           # строка пустая
[ -n "$A" ]           # строка не пустая

# Файловые тесты
[ -f file ]           # файл существует
[ -d dir ]            # директория существует
[ -r file ]           # есть право на чтение
[ -w file ]           # есть право на запись
[ -x file ]           # есть право на выполнение
[ -s file ]           # файл не пустой

# Двойные скобки (bash-расширения — предпочтительнее!)
[[ "$NAME" == "admin" ]]    # == работает
[[ "$FILE" =~ \.log$ ]]     # regex!</code></pre>

<h2>Циклы</h2>
<pre><code># for
for file in /var/log/*.log; do
    echo "Processing $file"
    wc -l "$file"
done

# for с range
for i in {1..10}; do
    echo "Iteration $i"
done

# for в стиле C
for ((i=0; i<5; i++)); do
    echo $i
done

# while
while true; do
    if ! pgrep -x nginx > /dev/null; then
        echo "Nginx down! Restarting..."
        systemctl restart nginx
    fi
    sleep 60
done

# Чтение файла построчно
while IFS= read -r line; do
    echo "Line: $line"
done < /etc/hosts</code></pre>

<h2>Функции</h2>
<pre><code># Определение функции
log() {
    local level=$1    # local — локальная переменная
    local message=$2
    echo "[$(date '+%H:%M:%S')] [$level] $message"
}

# Использование
log "INFO" "Starting deploy"
log "ERROR" "Connection failed"

# Возврат значения (через echo + подстановку)
get_free_memory() {
    free -m | awk '/Mem:/ {print $4}'
}
FREE_MB=$(get_free_memory)
echo "Free RAM: ${FREE_MB}MB"</code></pre>

<h2>Exit codes и обработка ошибок</h2>
<pre><code># Exit code: 0 = успех, 1-255 = ошибка
# $? — exit code последней команды

command_that_might_fail
if [ $? -ne 0 ]; then
    echo "Command failed!"
    exit 1
fi

# set -e — прервать скрипт при любой ошибке
set -e          # exit on error
set -u          # error on undefined variable
set -o pipefail # error if any command in pipe fails
set -x          # debug: печатать каждую команду

# Комбинация (best practice для продовых скриптов):
#!/bin/bash
set -euo pipefail

# trap — выполнить при выходе (cleanup)
cleanup() {
    rm -f /tmp/lockfile
    echo "Cleaned up"
}
trap cleanup EXIT    # вызвать cleanup при любом exit</code></pre>

<h2>Pipes и перенаправление</h2>
<pre><code># Pipe: stdout одной команды → stdin другой
cat /var/log/syslog | grep ERROR | wc -l

# Перенаправление
echo "hello" > file.txt      # перезаписать (создать)
echo "world" >> file.txt     # дописать в конец
command 2>/dev/null           # скрыть stderr
command > out.txt 2>&1        # stdout и stderr в файл
command &>/dev/null           # скрыть всё (stdout + stderr)

# Here document
cat << 'EOF' > config.yaml
server:
  port: 8080
  host: 0.0.0.0
EOF</code></pre>

<h2>Полезные паттерны</h2>
<pre><code># Проверка что запущены от root
if [[ $EUID -ne 0 ]]; then
    echo "This script must be run as root"
    exit 1
fi

# Lock-файл (предотвращение двойного запуска)
LOCKFILE="/tmp/deploy.lock"
if [ -f "$LOCKFILE" ]; then
    echo "Already running!"
    exit 1
fi
trap "rm -f $LOCKFILE" EXIT
touch "$LOCKFILE"</code></pre>`,

		Quiz: []Q{
			{
				Question:    "Что делает set -euo pipefail в начале bash-скрипта?",
				Options:     []string{"Ускоряет выполнение", "Включает строгий режим: -e=exit при ошибке, -u=ошибка на неопределённые переменные, -o pipefail=ошибка если любая команда в pipe провалилась", "Включает отладку", "Отключает вывод"},
				Correct:     1,
				Explanation: "Без set -e скрипт продолжит работу после ошибки — опасно для деплоя! Без -u опечатка в имени переменной ($PROT вместо $PORT) молча подставит пустую строку. Без pipefail: 'grep x | wc' вернёт 0 даже если grep упал. Это must-have для продовых скриптов.",
			},
			{
				Question:    "Почему NAME = \"hello\" (с пробелами вокруг =) — ошибка в bash?",
				Options:     []string{"Пробелы запрещены в bash", "Bash интерпретирует это как команду 'NAME' с аргументами '=' и 'hello', а не как присваивание", "Это работает, просто медленнее", "Нужно использовать let"},
				Correct:     1,
				Explanation: "В bash пробелы разделяют команду и аргументы. 'NAME = hello' = запустить программу NAME с аргументами '=' и 'hello'. Правильно: NAME=\"hello\" (без пробелов). Это одна из самых частых ошибок новичков. В отличие от Python/Go/JS, присваивание в bash не терпит пробелов.",
			},
			{
				Question:    "Что делает trap cleanup EXIT?",
				Options:     []string{"Вызывает функцию cleanup каждую секунду", "Регистрирует функцию cleanup для вызова при ЛЮБОМ завершении скрипта (нормальном, ошибке, Ctrl+C)", "Перехватывает ошибки", "Блокирует завершение скрипта"},
				Correct:     1,
				Explanation: "trap регистрирует обработчик сигнала. EXIT — псевдо-сигнал, срабатывает при любом exit (включая set -e). Используется для cleanup: удалить lock-файл, temp-файлы, отправить уведомление. Аналог defer в Go или finally в Java. Без trap временные файлы останутся при ошибке.",
			},
		},

		Tasks: []T{
			{
				Title:       "Интерпретатор условий bash",
				Description: "Напиши Go-программу, которая вычисляет bash-подобные условия из [ ... ] выражений: сравнение чисел (-eq, -ne, -gt, -lt), строк (=, !=), проверки файлов (-f, -d).",
				Hints:       "Парси выражение по пробелам. Определяй тип оператора: числовой (-eq,-ne,-gt,-lt,-ge,-le), строковый (=, !=), файловый (-f, -d, -z, -n). Для файловых проверок используй os.Stat().",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// evalCondition вычисляет bash-условие
// Поддерживает: num -eq num, num -gt num, str = str, -z str, -n str
func evalCondition(parts []string) (bool, error) {
	// TODO:
	// Унарные: -z (пустая строка), -n (непустая строка)
	// Бинарные числовые: -eq, -ne, -gt, -lt, -ge, -le
	// Бинарные строковые: =, !=
	return false, fmt.Errorf("unknown condition")
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basheval <condition>")
		fmt.Println("  basheval '5 -gt 3'")
		fmt.Println("  basheval 'hello = hello'")
		fmt.Println("  basheval '-z '")
		return
	}

	expr := strings.Join(os.Args[1:], " ")
	parts := strings.Fields(expr)
	result, err := evalCondition(parts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
	if result {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func evalCondition(parts []string) (bool, error) {
	if len(parts) == 0 {
		return false, fmt.Errorf("empty condition")
	}

	// Unary operators
	if len(parts) == 2 {
		switch parts[0] {
		case "-z":
			return parts[1] == "", nil
		case "-n":
			return parts[1] != "", nil
		}
	}
	if len(parts) == 1 {
		if parts[0] == "-z" {
			return true, nil // -z with no arg = empty string = true
		}
	}

	// Binary operators
	if len(parts) == 3 {
		left, op, right := parts[0], parts[1], parts[2]

		switch op {
		case "-eq", "-ne", "-gt", "-lt", "-ge", "-le":
			l, err1 := strconv.Atoi(left)
			r, err2 := strconv.Atoi(right)
			if err1 != nil || err2 != nil {
				return false, fmt.Errorf("not numbers")
			}
			switch op {
			case "-eq":
				return l == r, nil
			case "-ne":
				return l != r, nil
			case "-gt":
				return l > r, nil
			case "-lt":
				return l < r, nil
			case "-ge":
				return l >= r, nil
			case "-le":
				return l <= r, nil
			}
		case "=", "==":
			return left == right, nil
		case "!=":
			return left != right, nil
		}
	}

	return false, fmt.Errorf("unknown condition: %s", strings.Join(parts, " "))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: basheval <condition>")
		fmt.Println("  basheval '5 -gt 3'")
		fmt.Println("  basheval 'hello = hello'")
		fmt.Println("  basheval '-z '")
		return
	}

	expr := strings.Join(os.Args[1:], " ")
	parts := strings.Fields(expr)
	result, err := evalCondition(parts)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(2)
	}
	if result {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}`,
				TestCases: []TestCase{
					{Input: "5 -gt 3", ExpectedOutput: "true\n"},
					{Input: "5 -eq 5", ExpectedOutput: "true\n"},
					{Input: "hello = world", ExpectedOutput: "false\n"},
					{Input: "hello = hello", ExpectedOutput: "true\n"},
					{Input: "10 -lt 5", ExpectedOutput: "false\n"},
				},
				Glossary: []GlossaryItem{
					{Term: "Shebang (#!/bin/bash)", Definition: "Первая строка скрипта, указывающая какой интерпретатор использовать для выполнения."},
					{Term: "Exit code", Definition: "Числовой код завершения программы. 0=успех, 1-255=ошибка. Проверяется через $?."},
					{Term: "Pipe (|)", Definition: "Соединяет stdout одной команды с stdin другой. Позволяет строить цепочки обработки данных."},
				},
			},
			{
				Title:       "Парсер bash-переменных из скрипта",
				Description: "Напиши Go-программу, которая читает bash-скрипт и извлекает все определения переменных (NAME=value), подстановки команд $(cmd), и environment-переменные ($VAR).",
				Hints:       "Ищи паттерны: VARNAME=VALUE (присваивание), $(command) (подстановка), $VAR или ${VAR} (использование). Регулярки или ручной парсинг по символам.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ScriptVar struct {
	Name  string
	Value string
	Type  string // "assignment", "substitution", "env-reference"
	Line  int
}

// extractVars извлекает переменные из строк bash-скрипта
func extractVars(lines []string) []ScriptVar {
	var vars []ScriptVar
	// TODO:
	// Для каждой строки:
	// 1. Пропусти комментарии (начинаются с #, но не #!)
	// 2. Найди присваивания: NAME=value или NAME="value"
	// 3. Найди использования: $VAR или ${VAR}
	return vars
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	vars := extractVars(lines)
	for _, v := range vars {
		fmt.Printf("L%d [%-12s] %s=%s\n", v.Line, v.Type, v.Name, v.Value)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type ScriptVar struct {
	Name  string
	Value string
	Type  string
	Line  int
}

func extractVars(lines []string) []ScriptVar {
	var vars []ScriptVar

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || (strings.HasPrefix(trimmed, "#") && !strings.HasPrefix(trimmed, "#!")) {
			continue
		}

		// Find assignments: VAR=value
		if idx := strings.Index(trimmed, "="); idx > 0 {
			name := trimmed[:idx]
			if isValidVarName(name) && !strings.ContainsAny(name, " \t$") {
				value := trimmed[idx+1:]
				value = strings.Trim(value, "\"'")
				vars = append(vars, ScriptVar{
					Name: name, Value: value,
					Type: "assignment", Line: i + 1,
				})
			}
		}

		// Find references: $VAR or ${VAR}
		for j := 0; j < len(trimmed); j++ {
			if trimmed[j] == '$' && j+1 < len(trimmed) {
				var name string
				if trimmed[j+1] == '{' {
					end := strings.Index(trimmed[j+2:], "}")
					if end != -1 {
						name = trimmed[j+2 : j+2+end]
					}
				} else if trimmed[j+1] == '(' {
					continue // skip $(command)
				} else {
					end := j + 1
					for end < len(trimmed) && (isAlphaNum(trimmed[end]) || trimmed[end] == '_') {
						end++
					}
					name = trimmed[j+1 : end]
				}
				if name != "" && name != "?" {
					vars = append(vars, ScriptVar{
						Name: name, Value: "",
						Type: "env-reference", Line: i + 1,
					})
				}
			}
		}
	}
	return vars
}

func isValidVarName(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, ch := range s {
		if i == 0 && ch >= '0' && ch <= '9' {
			return false
		}
		if !isAlphaNum(byte(ch)) && ch != '_' {
			return false
		}
	}
	return true
}

func isAlphaNum(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_'
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	vars := extractVars(lines)
	for _, v := range vars {
		fmt.Printf("L%d [%-12s] %s=%s\n", v.Line, v.Type, v.Name, v.Value)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "#!/bin/bash\nNAME=\"world\"\nPORT=8080\necho \"Hello $NAME on port $PORT\"\n",
						ExpectedOutput: "L2 [assignment  ] NAME=world\nL3 [assignment  ] PORT=8080\nL4 [env-reference] NAME=\nL4 [env-reference] PORT=\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "Variable expansion", Definition: "$VAR или ${VAR} — подстановка значения переменной в строку."},
					{Term: "Command substitution", Definition: "$(command) — выполнить команду и подставить её stdout."},
					{Term: "trap", Definition: "Регистрация обработчика сигнала. trap cleanup EXIT — вызвать при выходе."},
				},
			},
		},
	}
}

// ══════════════════════════════════════════════════════════════════
// Урок 10: Logs & Troubleshooting
// ══════════════════════════════════════════════════════════════════

func linuxLesson10Logs() L {
	return L{
		Slug: "logs-troubleshooting", Title: "Логи и диагностика", Order: 10,
		Difficulty: "intermediate", Track: "devops",
		Content: `<h1>Логи и диагностика проблем</h1>

<h2>Где искать логи</h2>
<p>В Linux все важные логи собраны в <code>/var/log/</code>:</p>

<pre><code>/var/log/
├── syslog         # основной системный лог (Debian/Ubuntu)
├── messages       # основной системный лог (RHEL/CentOS)
├── auth.log       # аутентификация (SSH-логины, sudo)
├── kern.log       # сообщения ядра
├── dmesg          # загрузка ядра, железо
├── nginx/
│   ├── access.log # HTTP-запросы
│   └── error.log  # ошибки nginx
├── postgresql/
│   └── postgresql-16-main.log
├── journal/       # бинарные логи systemd (journalctl)
└── apt/
    ├── history.log # что устанавливалось
    └── term.log    # вывод apt</code></pre>

<h2>journalctl — логи systemd</h2>
<pre><code># Логи конкретного сервиса
journalctl -u nginx -f                       # follow (live)
journalctl -u nginx --since "2024-01-15"
journalctl -u nginx --since "1 hour ago"
journalctl -u nginx -p err                   # только ошибки

# Системные логи
journalctl -b                                # текущая загрузка
journalctl -b -1                             # предыдущая загрузка (после ребута)
journalctl --list-boots                      # все загрузки
journalctl -k                                # kernel messages
journalctl -p crit --since today             # критические за сегодня

# JSON-формат (для парсинга)
journalctl -u nginx -o json-pretty -n 5

# Размер логов
journalctl --disk-usage
sudo journalctl --vacuum-time=7d             # удалить логи старше 7 дней
sudo journalctl --vacuum-size=500M           # ограничить до 500MB</code></pre>

<h2>dmesg — сообщения ядра</h2>
<pre><code># Загрузка, железо, ошибки драйверов
dmesg | tail -20                     # последние сообщения
dmesg -T                             # с человекочитаемым временем
dmesg -l err,warn                    # только ошибки и предупреждения
dmesg | grep -i "oom"                # Out Of Memory events
dmesg | grep -i "error\|fail"        # ошибки железа</code></pre>

<h2>Диагностика: диск</h2>
<pre><code># Место на диске
df -h                                # все файловые системы
df -h /var                           # конкретный путь
du -sh /var/log/*                    # размер каждой папки в /var/log
du -sh /var/lib/docker               # сколько жрёт Docker

# Самые большие файлы
du -ah / | sort -rh | head -20       # top-20 больших файлов
find / -size +100M -type f 2>/dev/null  # файлы > 100MB

# Inodes (могут кончиться даже при свободном месте!)
df -i                                # использование inodes
find /tmp -type f | wc -l            # миллионы мелких файлов = нет inodes

# Типичные пожиратели диска:
# /var/log/ — логи (ротация!)
# /var/lib/docker/ — образы и контейнеры
# /tmp/ — не очищается автоматически на всех дистрибутивах
# core dumps в рабочих каталогах</code></pre>

<h2>Диагностика: память (OOM)</h2>
<pre><code># Текущее использование RAM
free -h
# total    used    free   shared  buff/cache   available
# 16G      8G      1G     200M    7G           7.5G

# ⚠️ "free" не значит "доступно"!
# available = free + buff/cache (которые ядро освободит при необходимости)

# OOM Killer — ядро убивает процесс при нехватке памяти
dmesg | grep -i "oom"
journalctl -k | grep -i "oom"
# Out of memory: Kill process 1234 (java) score 900 or sacrifice child

# Кто потребляет память
ps aux --sort=-%mem | head -10
cat /proc/meminfo                    # подробная статистика</code></pre>

<h2>strace — отладка на уровне syscalls</h2>
<pre><code># Увидеть ВСЕ syscalls программы
strace ls                            # все вызовы ls
strace -e open,read ls               # только open и read
strace -p 1234                       # подключиться к запущенному процессу
strace -c ls                         # статистика syscalls

# Примеры когда strace спасает:
# - Программа зависает → strace показывает на каком read/connect она ждёт
# - "Permission denied" непонятно где → strace показывает какой файл open() вернул EACCES
# - "File not found" → strace показывает какие пути программа пробует</code></pre>

<h2>Диагностика: CPU</h2>
<pre><code># Загрузка CPU
top                                  # живой мониторинг
uptime                               # load average (1/5/15 минут)
# load average: 2.50, 1.80, 0.90
# Если load > количество ядер — система перегружена

mpstat -P ALL 1                      # загрузка по ядрам (каждую секунду)
sar -u 1 5                           # CPU статистика (5 замеров по 1 сек)</code></pre>

<h2>Комплексная диагностика (чеклист)</h2>
<pre><code># 1. Диск
df -h && df -i

# 2. Память
free -h

# 3. CPU / Load
uptime

# 4. Процессы
ps aux --sort=-%cpu | head -5
ps aux --sort=-%mem | head -5

# 5. Логи
journalctl -p err --since "10 minutes ago"
dmesg | tail -20

# 6. Сеть
ss -tlnp                             # кто слушает
ping -c 1 8.8.8.8                    # есть ли интернет

# 7. Проблемный сервис
systemctl status problematic-service
journalctl -u problematic-service -n 50</code></pre>

<p><strong>Продовый совет:</strong> Когда сервер "тормозит" — проверяй в порядке: диск (df), память (free), CPU (top), сеть (ss), логи (journalctl). 80% проблем — полный диск или OOM.</p>`,

		Quiz: []Q{
			{
				Question:    "В 'free -h' показывает 1GB free и 7GB buff/cache. Система скоро упадёт?",
				Options:     []string{"Да, осталось только 1GB", "Нет — buff/cache автоматически освобождается при необходимости, смотри на колонку 'available'", "Нужно перезагрузить", "Зависит от swap"},
				Correct:     1,
				Explanation: "Linux использует свободную RAM как кэш для дисковых операций (buff/cache). Это ХОРОШО — ускоряет чтение. При нехватке ядро мгновенно освобождает кэш. Реальный показатель — 'available'. Если available близок к 0 — тогда проблема. Типичная ошибка мониторинга: alert на 'free' вместо 'available'.",
			},
			{
				Question:    "Что такое OOM Killer и когда он срабатывает?",
				Options:     []string{"Антивирус Linux", "Компонент ядра, который убивает процесс с наибольшим 'oom_score' когда RAM и swap полностью исчерпаны", "Системный процесс для перезагрузки", "Утилита для мониторинга памяти"},
				Correct:     1,
				Explanation: "Когда RAM + swap = 0, ядро не может выделить память для ни одного процесса (даже для kill). OOM Killer выбирает 'жертву' по oom_score (учитывает размер памяти, возраст, привилегии) и убивает через SIGKILL. Защита: oom_score_adj=-1000 для критичных процессов. Или настроить swap и лимиты памяти в systemd.",
			},
			{
				Question:    "Зачем нужен strace при отладке проблем?",
				Options:     []string{"Для ускорения программы", "Показывает все системные вызовы процесса — видно какие файлы открывает, на каких вызовах зависает, какие ошибки получает от ядра", "Для отладки bash-скриптов", "Для просмотра сетевого трафика"},
				Correct:     1,
				Explanation: "strace перехватывает syscalls между программой и ядром. Видно: open('/etc/config.yaml') = -1 ENOENT (файл не найден), connect('10.0.0.5:5432') = -1 ETIMEDOUT (сервер не отвечает). Это 'рентген' для программы: не нужен исходный код, работает с любым бинарником. Для Go-программ также полезен ltrace (library calls).",
			},
			{
				Question:    "Что означает load average 8.0 на 4-ядерном сервере?",
				Options:     []string{"Всё нормально", "Система перегружена вдвое — в среднем 4 процесса ждут CPU в очереди помимо 4 работающих", "Используется 80% CPU", "Нужно добавить RAM"},
				Correct:     1,
				Explanation: "Load average = среднее количество процессов в состоянии Running или Waiting. На 4 ядрах: load 4.0 = полная загрузка (каждое ядро занято), load 8.0 = перегрузка вдвое (4 работают + 4 ждут). НО: load включает процессы в D-state (I/O wait). Высокий load + низкий CPU% = проблема с диском, не с CPU.",
			},
		},

		Tasks: []T{
			{
				Title:       "Анализатор лог-файла",
				Description: "Напиши Go-программу, которая парсит строки лога формата syslog (timestamp hostname process[pid]: message) и выводит статистику: количество сообщений по уровню severity, top-5 процессов по количеству сообщений.",
				Hints:       "Формат syslog: 'Jan 15 10:30:45 server nginx[1234]: message'. Разделяй по пробелам. Процесс — поле с [PID]. Уровень определяй по ключевым словам в message: error, warning, info, debug.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type LogStats struct {
	TotalLines int
	ByLevel    map[string]int // error, warning, info, other
	ByProcess  map[string]int // process name → count
}

// classifyLevel определяет уровень по содержимому сообщения
func classifyLevel(message string) string {
	msg := strings.ToLower(message)
	// TODO: определи уровень по ключевым словам
	// error, fail, fatal → "error"
	// warn → "warning"
	// info, notice → "info"
	// остальное → "other"
	_ = msg
	return "other"
}

// parseLogLine извлекает процесс и сообщение из строки syslog
func parseLogLine(line string) (process, message string, ok bool) {
	// TODO: парси формат "Jan 15 10:30:45 hostname process[pid]: message"
	return "", "", false
}

func main() {
	stats := LogStats{
		ByLevel:   make(map[string]int),
		ByProcess: make(map[string]int),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		proc, msg, ok := parseLogLine(line)
		if !ok {
			continue
		}
		stats.TotalLines++
		stats.ByProcess[proc]++
		level := classifyLevel(msg)
		stats.ByLevel[level]++
		_ = msg
	}

	fmt.Printf("Total: %d lines\n", stats.TotalLines)
	fmt.Println("\n--- By Level ---")
	for _, level := range []string{"error", "warning", "info", "other"} {
		if count, ok := stats.ByLevel[level]; ok {
			fmt.Printf("  %-8s %d\n", level, count)
		}
	}

	fmt.Println("\n--- Top Processes ---")
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range stats.ByProcess {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })
	for i := 0; i < 5 && i < len(sorted); i++ {
		fmt.Printf("  %-15s %d\n", sorted[i].Key, sorted[i].Value)
	}
}`,
				Solution: `package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

type LogStats struct {
	TotalLines int
	ByLevel    map[string]int
	ByProcess  map[string]int
}

func classifyLevel(message string) string {
	msg := strings.ToLower(message)
	if strings.Contains(msg, "error") || strings.Contains(msg, "fail") || strings.Contains(msg, "fatal") {
		return "error"
	}
	if strings.Contains(msg, "warn") {
		return "warning"
	}
	if strings.Contains(msg, "info") || strings.Contains(msg, "notice") {
		return "info"
	}
	return "other"
}

func parseLogLine(line string) (process, message string, ok bool) {
	// Format: "Jan 15 10:30:45 hostname process[pid]: message"
	// Fields: 0=month, 1=day, 2=time, 3=hostname, 4=process[pid]:, 5+=message
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return "", "", false
	}

	procField := fields[4]
	// Remove [pid]: suffix
	if idx := strings.Index(procField, "["); idx != -1 {
		process = procField[:idx]
	} else {
		process = strings.TrimSuffix(procField, ":")
	}

	message = strings.Join(fields[5:], " ")
	return process, message, true
}

func main() {
	stats := LogStats{
		ByLevel:   make(map[string]int),
		ByProcess: make(map[string]int),
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		proc, msg, ok := parseLogLine(line)
		if !ok {
			continue
		}
		stats.TotalLines++
		stats.ByProcess[proc]++
		level := classifyLevel(msg)
		stats.ByLevel[level]++
		_ = msg
	}

	fmt.Printf("Total: %d lines\n", stats.TotalLines)
	fmt.Println("\n--- By Level ---")
	for _, level := range []string{"error", "warning", "info", "other"} {
		if count, ok := stats.ByLevel[level]; ok {
			fmt.Printf("  %-8s %d\n", level, count)
		}
	}

	fmt.Println("\n--- Top Processes ---")
	type kv struct {
		Key   string
		Value int
	}
	var sorted []kv
	for k, v := range stats.ByProcess {
		sorted = append(sorted, kv{k, v})
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Value > sorted[j].Value })
	for i := 0; i < 5 && i < len(sorted); i++ {
		fmt.Printf("  %-15s %d\n", sorted[i].Key, sorted[i].Value)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "Jan 15 10:30:45 server nginx[1234]: Connection error from 10.0.0.1\nJan 15 10:30:46 server nginx[1234]: GET /api/users 200 info\nJan 15 10:30:47 server postgres[500]: WARNING: table bloat detected\nJan 15 10:31:00 server sshd[800]: Accepted publickey for user\n",
						ExpectedOutput: "Total: 4 lines\n\n--- By Level ---\n  error    1\n  warning  1\n  info     1\n  other    1\n\n--- Top Processes ---\n  nginx           2\n  postgres        1\n  sshd            1\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "syslog", Definition: "Стандарт формата логов Linux. Включает timestamp, hostname, facility, severity, message."},
					{Term: "OOM Killer", Definition: "Out-Of-Memory Killer — компонент ядра, убивающий процессы при полном исчерпании RAM."},
					{Term: "Load average", Definition: "Среднее количество процессов в состоянии Running/Waiting за 1/5/15 минут. Сравнивай с числом ядер CPU."},
				},
			},
			{
				Title:       "Симулятор диагностики сервера",
				Description: "Напиши Go-программу, которая принимает 'метрики сервера' (CPU%, RAM%, disk%, load) и выдаёт диагностику с приоритетом: какая проблема самая критическая, что проверить, какие команды выполнить.",
				Hints:       "Определи пороги: disk>90%=critical, RAM(available)<10%=critical, load>2*CPUs=high. Выдавай конкретные рекомендации и команды для каждой проблемы.",
				Difficulty:  "medium",
				StarterCode: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ServerMetrics struct {
	CPUPercent  float64
	RAMPercent  float64
	DiskPercent float64
	LoadAvg    float64
	NumCPUs    int
}

type Diagnosis struct {
	Level       string // "critical", "warning", "ok"
	Problem     string
	Command     string // что выполнить для диагностики
	Action      string // что делать
}

// diagnose анализирует метрики и возвращает диагнозы
func diagnose(m ServerMetrics) []Diagnosis {
	var results []Diagnosis
	// TODO:
	// 1. Disk > 90% → critical (df -h, du -sh /var/log/*, очистка)
	// 2. Disk > 80% → warning
	// 3. RAM > 90% → critical (free -h, ps aux --sort=-%mem)
	// 4. Load > 2*CPUs → critical (top, проверь D-state)
	// 5. Load > CPUs → warning
	// 6. CPU > 90% → warning
	return results
}

func main() {
	// Вход: "cpu=85 ram=92 disk=45 load=12.5 cpus=4"
	if len(os.Args) < 2 {
		fmt.Println("Usage: diagnose 'cpu=85 ram=92 disk=45 load=12.5 cpus=4'")
		return
	}

	m := ServerMetrics{NumCPUs: 4}
	for _, arg := range strings.Fields(os.Args[1]) {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue
		}
		val, _ := strconv.ParseFloat(kv[1], 64)
		switch kv[0] {
		case "cpu":
			m.CPUPercent = val
		case "ram":
			m.RAMPercent = val
		case "disk":
			m.DiskPercent = val
		case "load":
			m.LoadAvg = val
		case "cpus":
			m.NumCPUs = int(val)
		}
	}

	results := diagnose(m)
	if len(results) == 0 {
		fmt.Println("[OK] All metrics within normal range")
		return
	}
	for _, d := range results {
		fmt.Printf("[%-8s] %s\n", strings.ToUpper(d.Level), d.Problem)
		fmt.Printf("           Check: %s\n", d.Command)
		fmt.Printf("           Fix:   %s\n\n", d.Action)
	}
}`,
				Solution: `package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type ServerMetrics struct {
	CPUPercent  float64
	RAMPercent  float64
	DiskPercent float64
	LoadAvg    float64
	NumCPUs    int
}

type Diagnosis struct {
	Level   string
	Problem string
	Command string
	Action  string
}

func diagnose(m ServerMetrics) []Diagnosis {
	var results []Diagnosis

	if m.DiskPercent > 90 {
		results = append(results, Diagnosis{
			Level:   "critical",
			Problem: fmt.Sprintf("Disk usage at %.0f%% — server may stop writing", m.DiskPercent),
			Command: "df -h && du -sh /var/log/* /var/lib/docker",
			Action:  "Clear logs: journalctl --vacuum-time=3d, docker system prune",
		})
	} else if m.DiskPercent > 80 {
		results = append(results, Diagnosis{
			Level:   "warning",
			Problem: fmt.Sprintf("Disk usage at %.0f%%", m.DiskPercent),
			Command: "df -h && du -sh /var/log/*",
			Action:  "Monitor growth, set up log rotation",
		})
	}

	if m.RAMPercent > 90 {
		results = append(results, Diagnosis{
			Level:   "critical",
			Problem: fmt.Sprintf("RAM usage at %.0f%% — OOM Killer may trigger", m.RAMPercent),
			Command: "free -h && ps aux --sort=-%mem | head -5",
			Action:  "Identify memory leak, restart offending process, add swap",
		})
	}

	if m.LoadAvg > float64(m.NumCPUs)*2 {
		results = append(results, Diagnosis{
			Level:   "critical",
			Problem: fmt.Sprintf("Load %.1f on %d CPUs — system overloaded", m.LoadAvg, m.NumCPUs),
			Command: "top -bn1 | head -20 && ps aux --sort=-%cpu | head -5",
			Action:  "Check for D-state processes (IO wait), kill runaway processes",
		})
	} else if m.LoadAvg > float64(m.NumCPUs) {
		results = append(results, Diagnosis{
			Level:   "warning",
			Problem: fmt.Sprintf("Load %.1f on %d CPUs — above capacity", m.LoadAvg, m.NumCPUs),
			Command: "uptime && top -bn1 | head -10",
			Action:  "Monitor trend, consider scaling",
		})
	}

	if m.CPUPercent > 90 {
		results = append(results, Diagnosis{
			Level:   "warning",
			Problem: fmt.Sprintf("CPU at %.0f%%", m.CPUPercent),
			Command: "top -bn1 -o %%CPU | head -15",
			Action:  "Identify CPU-intensive process, check for infinite loops",
		})
	}

	return results
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: diagnose 'cpu=85 ram=92 disk=45 load=12.5 cpus=4'")
		return
	}

	m := ServerMetrics{NumCPUs: 4}
	for _, arg := range strings.Fields(os.Args[1]) {
		kv := strings.Split(arg, "=")
		if len(kv) != 2 {
			continue
		}
		val, _ := strconv.ParseFloat(kv[1], 64)
		switch kv[0] {
		case "cpu":
			m.CPUPercent = val
		case "ram":
			m.RAMPercent = val
		case "disk":
			m.DiskPercent = val
		case "load":
			m.LoadAvg = val
		case "cpus":
			m.NumCPUs = int(val)
		}
	}

	results := diagnose(m)
	if len(results) == 0 {
		fmt.Println("[OK] All metrics within normal range")
		return
	}
	for _, d := range results {
		fmt.Printf("[%-8s] %s\n", strings.ToUpper(d.Level), d.Problem)
		fmt.Printf("           Check: %s\n", d.Command)
		fmt.Printf("           Fix:   %s\n\n", d.Action)
	}
}`,
				TestCases: []TestCase{
					{
						Input:          "cpu=50 ram=60 disk=45 load=2.0 cpus=4",
						ExpectedOutput: "[OK] All metrics within normal range\n",
					},
					{
						Input:          "cpu=50 ram=95 disk=92 load=12.0 cpus=4",
						ExpectedOutput: "[CRITICAL] Disk usage at 92% — server may stop writing\n           Check: df -h && du -sh /var/log/* /var/lib/docker\n           Fix:   Clear logs: journalctl --vacuum-time=3d, docker system prune\n\n[CRITICAL] RAM usage at 95% — OOM Killer may trigger\n           Check: free -h && ps aux --sort=-%mem | head -5\n           Fix:   Identify memory leak, restart offending process, add swap\n\n[CRITICAL] Load 12.0 on 4 CPUs — system overloaded\n           Check: top -bn1 | head -20 && ps aux --sort=-%cpu | head -5\n           Fix:   Check for D-state processes (IO wait), kill runaway processes\n\n",
					},
				},
				Glossary: []GlossaryItem{
					{Term: "strace", Definition: "Инструмент отладки, перехватывающий системные вызовы программы. Показывает взаимодействие с ядром."},
					{Term: "dmesg", Definition: "Сообщения кольцевого буфера ядра. Показывает информацию о загрузке, железе, ошибках драйверов."},
					{Term: "Log rotation", Definition: "Автоматическое архивирование и удаление старых логов (logrotate). Предотвращает заполнение диска."},
				},
			},
		},
	}
}
