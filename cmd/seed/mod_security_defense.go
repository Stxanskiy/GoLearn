package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Кибербезопасность — Defense (Blue Team)
// 7 уроков: hardening → logging → SIEM → firewall → IR → cloud → DevSecOps
// ════════════════════════════════════════════════════════════════

func mod_security_defense() M {
	return M{
		Slug:          "security-defense",
		Title:         "Кибербезопасность: Blue Team",
		Description:   "Hardening, SIEM, Incident Response, Firewall, Cloud Security, DevSecOps.",
		Order:         32,
		Track:         "security-defense",
		Difficulty:    "intermediate",
		Prerequisites: []string{"linux-fundamentals"},
		Lessons: []L{
			lesson_def_hardening(),
			lesson_def_logging(),
			lesson_def_siem(),
			lesson_def_firewall(),
			lesson_def_incident_response(),
			lesson_def_cloud_security(),
			lesson_def_devsecops(),
		},
	}
}

// ═══════════════════ Урок 1: Linux Hardening ═══════════════════

func lesson_def_hardening() L {
	return L{
		Slug: "def-hardening", Title: "Linux Hardening", Order: 1,
		Difficulty: "intermediate", Track: "security-defense",
		Content: `<h1>Linux Hardening — укрепление системы</h1>
<h2>Принцип минимальных привилегий</h2>
<pre><code># Отключить root SSH
PermitRootLogin no        # /etc/ssh/sshd_config
# Только ключи, без паролей
PasswordAuthentication no
# Сменить порт SSH
Port 2222</code></pre>

<h2>Пользователи и sudo</h2>
<pre><code># Создать непривилегированного пользователя
adduser appuser
# Дать sudo только для нужных команд
echo "appuser ALL=(ALL) NOPASSWD: /usr/bin/systemctl restart myapp" >> /etc/sudoers.d/appuser
# Заблокировать неиспользуемых
usermod -L olduser</code></pre>

<h2>File permissions</h2>
<pre><code>chmod 600 /etc/shadow          # только root читает
chmod 700 /root                 # только root входит
find / -perm -4000 -type f      # аудит SUID файлов
chmod u-s /usr/bin/unneeded     # снять SUID если не нужен</code></pre>

<h2>Обновления и патчи</h2>
<pre><code># Автоматические security обновления (Ubuntu)
apt install unattended-upgrades
dpkg-reconfigure -plow unattended-upgrades

# Проверка устаревших пакетов
apt list --upgradable | grep -i security</code></pre>

<h2>auditd — аудит действий</h2>
<pre><code># Установка
apt install auditd
# Правила: кто читал /etc/shadow
auditctl -w /etc/shadow -p r -k shadow_access
# Кто менял sudoers
auditctl -w /etc/sudoers -p wa -k sudoers_change
# Просмотр логов
ausearch -k shadow_access</code></pre>`,

		Quiz: []Q{
			{Question: "Почему отключать root SSH?", Options: []string{"Неудобно", "Root = известный username. Атакующий brute-force: root + пароль. Отключение = нужно знать И username И пароль", "Быстрее", "Linux требует"}, Correct: 1, Explanation: "PermitRootLogin no: атакующий должен угадать И имя пользователя И пароль. Без root: даже если вошёл — нужен отдельный privilege escalation."},
			{Question: "PasswordAuthentication no — зачем?", Options: []string{"Сложнее", "SSH ключи невозможно brute-force (2048+ бит), пароли — можно", "Быстрее", "Для красоты"}, Correct: 1, Explanation: "SSH ключ 4096 бит vs пароль 'admin123'. Brute-force ключа = миллиарды лет. Пароля = часы/дни. Ключи = единственный надёжный метод."},
			{Question: "auditd — что даёт?", Options: []string{"Ускоряет", "Логирует КТО, КОГДА, ЧТО делал с файлами/командами. Для forensics и compliance", "Firewall", "Антивирус"}, Correct: 1, Explanation: "auditd: audit trail. Кто читал /etc/shadow, кто менял конфиги, кто запускал sudo. После инцидента: ausearch показывает всю цепочку действий."},
			{Question: "Принцип наименьших привилегий?", Options: []string{"Дать всем root", "Пользователь/процесс имеет ТОЛЬКО те права, которые необходимы для работы — не больше", "Минимум пользователей", "Один пароль для всех"}, Correct: 1, Explanation: "appuser может только restart myapp, не может rm -rf /. Если скомпрометирован — damage ограничен. Root нужен только для конкретных операций."},
			{Question: "unattended-upgrades — зачем?", Options: []string{"Новые фичи", "Автоматическая установка security патчей — закрывает уязвимости без вмешательства админа", "Ускоряет", "Экономит диск"}, Correct: 1, Explanation: "CVE опубликован → патч выпущен → если не обновил = уязвим. Автообновления: патчи ставятся сами. Для production: с тестированием, для dev — сразу."},
		},
		Tasks: []T{
			{Title: "SSH hardening config", Difficulty: "easy", Description: `<p>Сгенерируй hardened sshd_config:</p><p>Ввод: <code>2222 appuser</code></p><p>Вывод:</p><pre><code>Port 2222
PermitRootLogin no
PasswordAuthentication no
AllowUsers appuser
MaxAuthTries 3</code></pre>`, Glossary: []GlossaryItem{{Term: "sshd_config", Definition: "Конфигурация SSH сервера. /etc/ssh/sshd_config."}}, TestCases: []TestCase{{Input: "2222 appuser", ExpectedOutput: "Port 2222\nPermitRootLogin no\nPasswordAuthentication no\nAllowUsers appuser\nMaxAuthTries 3"}},
				StarterCode: `package main
import "fmt"
func main() { var port int; var user string; fmt.Scan(&port, &user); fmt.Printf("Port %d\nPermitRootLogin no\nPasswordAuthentication no\nAllowUsers %s\nMaxAuthTries 3\n", port, user) }`, Hints: `<p>Port, PermitRootLogin no, PasswordAuthentication no, AllowUsers, MaxAuthTries.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var p int;var u string;fmt.Scan(&p,&u);fmt.Printf("Port %d\nPermitRootLogin no\nPasswordAuthentication no\nAllowUsers %s\nMaxAuthTries 3\n",p,u)}</code></pre>`},
			{Title: "Audit rules generator", Difficulty: "easy", Description: `<p>Сгенерируй auditd правила:</p><p>Ввод: <code>/etc/shadow /etc/passwd</code></p><p>Вывод:</p><pre><code>auditctl -w /etc/shadow -p rwa -k sensitive_files
auditctl -w /etc/passwd -p rwa -k sensitive_files</code></pre>`, Glossary: []GlossaryItem{{Term: "auditctl -w", Definition: "Watch file. -p rwa = read/write/attribute changes. -k = key для поиска."}}, TestCases: []TestCase{{Input: "/etc/shadow /etc/passwd", ExpectedOutput: "auditctl -w /etc/shadow -p rwa -k sensitive_files\nauditctl -w /etc/passwd -p rwa -k sensitive_files"}},
				StarterCode: `package main
import "fmt"
func main() { var f1, f2 string; fmt.Scan(&f1, &f2); fmt.Printf("auditctl -w %s -p rwa -k sensitive_files\nauditctl -w %s -p rwa -k sensitive_files\n", f1, f2) }`, Hints: `<p>-w path -p permissions -k key. rwa = read+write+attribute.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var a,b string;fmt.Scan(&a,&b);fmt.Printf("auditctl -w %s -p rwa -k sensitive_files\nauditctl -w %s -p rwa -k sensitive_files\n",a,b)}</code></pre>`},
			{Title: "Permission checker", Difficulty: "medium", Description: `<p>Проверь файлы на опасные permissions:</p><p>Ввод:</p><pre><code>3
/etc/shadow 644
/etc/ssh/sshd_config 600
/home/user/.ssh/id_rsa 777</code></pre><p>Вывод:</p><pre><code>WARN: /etc/shadow should be 600, got 644
OK: /etc/ssh/sshd_config (600)
CRITICAL: /home/user/.ssh/id_rsa is world-readable (777)</code></pre>`, Glossary: []GlossaryItem{{Term: "File permissions", Definition: "600=owner only. 644=owner rw + others read. 777=everyone everything — DANGER."}}, TestCases: []TestCase{{Input: "3\n/etc/shadow 644\n/etc/ssh/sshd_config 600\n/home/user/.ssh/id_rsa 777", ExpectedOutput: "WARN: /etc/shadow should be 600, got 644\nOK: /etc/ssh/sshd_config (600)\nCRITICAL: /home/user/.ssh/id_rsa is world-readable (777)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    expected := map[string]string{"/etc/shadow": "600", "/etc/ssh/sshd_config": "600"}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); path, perm := parts[0], parts[1]
        if perm == "777" { fmt.Printf("CRITICAL: %s is world-readable (777)\n", path)
        } else if exp, ok := expected[path]; ok && perm != exp { fmt.Printf("WARN: %s should be %s, got %s\n", path, exp, perm)
        } else { fmt.Printf("OK: %s (%s)\n", path, perm) }
    }
}`, Hints: `<p>777 = critical. Сравни с expected map. Остальное OK.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){exp:=map[string]string{"/etc/shadow":"600","/etc/ssh/sshd_config":"600"};var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());if p[1]=="777"{fmt.Printf("CRITICAL: %s is world-readable (777)\n",p[0])}else if e,ok:=exp[p[0]];ok&&p[1]!=e{fmt.Printf("WARN: %s should be %s, got %s\n",p[0],e,p[1])}else{fmt.Printf("OK: %s (%s)\n",p[0],p[1])}}}</code></pre>`},
			{Title: "Sudo policy generator", Difficulty: "medium", Description: `<p>Сгенерируй минимальную sudo policy:</p><p>Ввод: <code>deploy /usr/bin/systemctl,/usr/bin/docker</code></p><p>Вывод: <code>deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl, /usr/bin/docker</code></p>`, Glossary: []GlossaryItem{{Term: "sudoers", Definition: "Файл с правилами кто и что может запускать через sudo."}}, TestCases: []TestCase{{Input: "deploy /usr/bin/systemctl,/usr/bin/docker", ExpectedOutput: "deploy ALL=(ALL) NOPASSWD: /usr/bin/systemctl, /usr/bin/docker"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() { var user, cmds string; fmt.Scan(&user, &cmds); fmt.Printf("%s ALL=(ALL) NOPASSWD: %s\n", user, strings.ReplaceAll(cmds, ",", ", ")) }`, Hints: `<p>user ALL=(ALL) NOPASSWD: cmd1, cmd2.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){var u,c string;fmt.Scan(&u,&c);fmt.Printf("%s ALL=(ALL) NOPASSWD: %s\n",u,strings.ReplaceAll(c,",",", "))}</code></pre>`},
			{Title: "Hardening checklist scorer", Difficulty: "hard", Description: `<p>Оцени уровень hardening:</p><p>Ввод:</p><pre><code>5
ssh_root_disabled yes
ssh_keys_only yes
firewall_enabled no
auto_updates yes
audit_enabled no</code></pre><p>Вывод:</p><pre><code>Score: 3/5 (60%)
MISSING: firewall_enabled, audit_enabled
Risk: MEDIUM</code></pre>`, Glossary: []GlossaryItem{{Term: "Hardening score", Definition: "Процент выполненных мер безопасности. 80%+ = good."}}, TestCases: []TestCase{{Input: "5\nssh_root_disabled yes\nssh_keys_only yes\nfirewall_enabled no\nauto_updates yes\naudit_enabled no", ExpectedOutput: "Score: 3/5 (60%)\nMISSING: firewall_enabled, audit_enabled\nRisk: MEDIUM"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    passed := 0; var missing []string
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text())
        if parts[1] == "yes" { passed++ } else { missing = append(missing, parts[0]) }
    }
    pct := passed * 100 / n
    var risk string
    switch { case pct >= 80: risk = "LOW"; case pct >= 50: risk = "MEDIUM"; default: risk = "HIGH" }
    fmt.Printf("Score: %d/%d (%d%%)\n", passed, n, pct)
    fmt.Printf("MISSING: %s\n", strings.Join(missing, ", "))
    fmt.Printf("Risk: %s\n", risk)
}`, Hints: `<p>Считай yes/no. Score = yes/total. missing = все no. Risk по проценту.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);p:=0;var m []string
    for i:=0;i<n;i++{sc.Scan();parts:=strings.Fields(sc.Text());if parts[1]=="yes"{p++}else{m=append(m,parts[0])}}
    pct:=p*100/n;var r string;switch{case pct>=80:r="LOW";case pct>=50:r="MEDIUM";default:r="HIGH"}
    fmt.Printf("Score: %d/%d (%d%%)\nMISSING: %s\nRisk: %s\n",p,n,pct,strings.Join(m,", "),r)}</code></pre>`},
		},
	}
}

// ═══════════════════ Уроки 2-7 ═══════════════════

func lesson_def_logging() L {
	return L{
		Slug: "def-logging", Title: "Мониторинг и логирование", Order: 2,
		Difficulty: "intermediate", Track: "security-defense",
		Content: `<h1>Security Logging</h1><p>Что логировать: auth events, privilege changes, file access, network connections. Где хранить: centralized (ELK, Loki). Retention: минимум 90 дней.</p>
<pre><code># Ключевые логи Linux:
/var/log/auth.log     # SSH, sudo, su
/var/log/syslog       # системные события
/var/log/kern.log     # ядро
journalctl -u sshd    # systemd unit logs</code></pre>`,
		Quiz: []Q{
			{Question: "Зачем централизованное логирование?", Options: []string{"Удобство", "Если сервер скомпрометирован — локальные логи удалят. Центральный сервер не под контролем атакующего", "Быстрее", "Экономия места"}, Correct: 1, Explanation: "Атакующий первым делом чистит логи. Если логи отправляются в реальном времени на отдельный сервер — не может стереть."},
			{Question: "auth.log — что содержит?", Options: []string{"HTTP запросы", "Все попытки авторизации: SSH login, sudo, su, failed password", "Сетевой трафик", "Файловые операции"}, Correct: 1, Explanation: "auth.log = кто пытался войти, откуда, успешно ли. grep 'Failed password' /var/log/auth.log — брутфорс обнаружение."},
			{Question: "Retention 90 дней — зачем так долго?", Options: []string{"Закон", "Инцидент может быть обнаружен через недели/месяцы. Логи нужны для расследования", "Привычка", "Не важно"}, Correct: 1, Explanation: "APT (Advanced Persistent Threat) живёт в сети месяцами. Обнаружение через 60 дней → нужны логи за 60+ дней для полного расследования."},
			{Question: "ELK Stack — что это?", Options: []string{"Язык", "Elasticsearch + Logstash + Kibana — система сбора, хранения и визуализации логов", "Firewall", "Антивирус"}, Correct: 1, Explanation: "Logstash: собирает и парсит логи. Elasticsearch: хранит и ищет. Kibana: дашборды и визуализация. Альтернатива: Grafana Loki + Promtail."},
			{Question: "Что НЕЛЬЗЯ логировать?", Options: []string{"IP адреса", "Пароли, токены, персональные данные (PII) — нарушение security и GDPR", "Timestamps", "User agents"}, Correct: 1, Explanation: "Логи = потенциальная утечка. Пароли в логах = если логи утекут → все пароли скомпрометированы. PII (email, паспорт) → штраф GDPR."},
		},
		Tasks: []T{
			{Title: "Log parser", Difficulty: "easy", Description: `<p>Найди failed SSH logins:</p><p>Ввод:</p><pre><code>3
Failed password for root from 10.0.0.5
Accepted password for admin from 10.0.0.1
Failed password for user from 10.0.0.5</code></pre><p>Вывод:</p><pre><code>ALERT: 2 failed logins from 10.0.0.5</code></pre>`, Glossary: []GlossaryItem{{Term: "Failed password", Definition: "Строка в auth.log при неудачной авторизации SSH."}}, TestCases: []TestCase{{Input: "3\nFailed password for root from 10.0.0.5\nAccepted password for admin from 10.0.0.1\nFailed password for user from 10.0.0.5", ExpectedOutput: "ALERT: 2 failed logins from 10.0.0.5"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    ips := map[string]int{}
    for i := 0; i < n; i++ { sc.Scan(); line := sc.Text()
        if strings.Contains(line, "Failed password") { parts := strings.Fields(line); ip := parts[len(parts)-1]; ips[ip]++ }
    }
    for ip, count := range ips { if count >= 2 { fmt.Printf("ALERT: %d failed logins from %s\n", count, ip) } }
}`, Hints: `<p>Ищи "Failed password", извлекай IP (последнее слово). Считай по IP.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);ips:=map[string]int{}
    for i:=0;i<n;i++{sc.Scan();l:=sc.Text();if strings.Contains(l,"Failed password"){p:=strings.Fields(l);ips[p[len(p)-1]]++}}
    for ip,c:=range ips{if c>=2{fmt.Printf("ALERT: %d failed logins from %s\n",c,ip)}}}</code></pre>`},
			{Title: "Log level classifier", Difficulty: "easy", Description: `<p>Классифицируй события по severity:</p><p>Ввод:</p><pre><code>3
Failed password brute_force
Accepted publickey normal_login
sudo command privilege_use</code></pre><p>Вывод:</p><pre><code>[HIGH] Failed password (brute_force)
[INFO] Accepted publickey (normal_login)
[MEDIUM] sudo command (privilege_use)</code></pre>`, Glossary: []GlossaryItem{{Term: "Log severity", Definition: "INFO=normal, MEDIUM=interesting, HIGH=potential attack, CRITICAL=confirmed attack."}}, TestCases: []TestCase{{Input: "3\nFailed password brute_force\nAccepted publickey normal_login\nsudo command privilege_use", ExpectedOutput: "[HIGH] Failed password (brute_force)\n[INFO] Accepted publickey (normal_login)\n[MEDIUM] sudo command (privilege_use)"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    severity := map[string]string{"brute_force": "HIGH", "normal_login": "INFO", "privilege_use": "MEDIUM", "file_access": "LOW"}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); event := strings.Join(parts[:len(parts)-1], " "); tag := parts[len(parts)-1]
        fmt.Printf("[%s] %s (%s)\n", severity[tag], event, tag)
    }
}`, Hints: `<p>Map tag → severity level. Выводи [LEVEL] event (tag).</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){sev:=map[string]string{"brute_force":"HIGH","normal_login":"INFO","privilege_use":"MEDIUM"};var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());t:=p[len(p)-1];e:=strings.Join(p[:len(p)-1]," ");fmt.Printf("[%s] %s (%s)\n",sev[t],e,t)}}</code></pre>`},
			{Title: "Anomaly detector", Difficulty: "medium", Description: `<p>Определи аномалии: >5 failed logins за минуту = brute force:</p><p>Ввод: <code>7</code></p><p>Вывод: <code>ALERT: Possible brute-force (7 attempts)</code></p>`, Glossary: []GlossaryItem{{Term: "Anomaly detection", Definition: "Отклонение от нормы. >5 failed/min = brute. >100 requests/sec = DDoS."}}, TestCases: []TestCase{{Input: "7", ExpectedOutput: "ALERT: Possible brute-force (7 attempts)"}, {Input: "3", ExpectedOutput: "OK: normal activity (3 attempts)"}},
				StarterCode: `package main
import "fmt"
func main() { var attempts int; fmt.Scan(&attempts); if attempts > 5 { fmt.Printf("ALERT: Possible brute-force (%d attempts)\n", attempts) } else { fmt.Printf("OK: normal activity (%d attempts)\n", attempts) } }`, Hints: `<p>>5 за минуту = alert. <=5 = normal.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var a int;fmt.Scan(&a);if a>5{fmt.Printf("ALERT: Possible brute-force (%d attempts)\n",a)}else{fmt.Printf("OK: normal activity (%d attempts)\n",a)}}</code></pre>`},
			{Title: "SIEM alert rule", Difficulty: "hard", Description: `<p>Сгенерируй SIEM правило по описанию:</p><p>Ввод: <code>brute_force 5 60 ssh</code> (type, threshold, window_sec, service)</p><p>Вывод:</p><pre><code>Rule: brute_force_ssh
Condition: failed_auth >= 5 within 60s
Source: ssh
Action: block_ip, alert_team</code></pre>`, Glossary: []GlossaryItem{{Term: "SIEM rule", Definition: "Condition → Action. Если событий > threshold за window → alert/block."}}, TestCases: []TestCase{{Input: "brute_force 5 60 ssh", ExpectedOutput: "Rule: brute_force_ssh\nCondition: failed_auth >= 5 within 60s\nSource: ssh\nAction: block_ip, alert_team"}},
				StarterCode: `package main
import "fmt"
func main() {
    var ruleType string; var threshold, window int; var service string
    fmt.Scan(&ruleType, &threshold, &window, &service)
    fmt.Printf("Rule: %s_%s\nCondition: failed_auth >= %d within %ds\nSource: %s\nAction: block_ip, alert_team\n", ruleType, service, threshold, window, service)
}`, Hints: `<p>Формат: Rule name, Condition с threshold и window, Source, Action.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var t string;var th,w int;var s string;fmt.Scan(&t,&th,&w,&s);fmt.Printf("Rule: %s_%s\nCondition: failed_auth >= %d within %ds\nSource: %s\nAction: block_ip, alert_team\n",t,s,th,w,s)}</code></pre>`},
			{Title: "Log correlation", Difficulty: "hard", Description: `<p>Коррелируй события из разных источников для обнаружения атаки:</p><p>Ввод:</p><pre><code>3
ssh_fail 10.0.0.5 5
web_scan 10.0.0.5 100
ssh_success 10.0.0.5 1</code></pre><p>Вывод:</p><pre><code>CORRELATED ATTACK from 10.0.0.5:
1. Port scan/web recon (100 requests)
2. SSH brute-force (5 failures)
3. SSH access obtained (1 success)
Verdict: COMPROMISED - isolate host immediately</code></pre>`, Glossary: []GlossaryItem{{Term: "Log correlation", Definition: "Объединение событий из разных источников для обнаружения сложных атак."}}, TestCases: []TestCase{{Input: "3\nssh_fail 10.0.0.5 5\nweb_scan 10.0.0.5 100\nssh_success 10.0.0.5 1", ExpectedOutput: "CORRELATED ATTACK from 10.0.0.5:\n1. Port scan/web recon (100 requests)\n2. SSH brute-force (5 failures)\n3. SSH access obtained (1 success)\nVerdict: COMPROMISED - isolate host immediately"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin); var ip string
    type Event struct{ name string; count int }
    var events []Event
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); var count int; fmt.Sscan(parts[2], &count)
        ip = parts[1]; events = append(events, Event{parts[0], count})
    }
    fmt.Printf("CORRELATED ATTACK from %s:\n", ip)
    descs := map[string]string{"web_scan": "Port scan/web recon", "ssh_fail": "SSH brute-force", "ssh_success": "SSH access obtained"}
    order := []string{"web_scan", "ssh_fail", "ssh_success"}
    step := 1
    for _, o := range order { for _, e := range events { if e.name == o { fmt.Printf("%d. %s (%d %s)\n", step, descs[o], e.count, map[string]string{"web_scan":"requests","ssh_fail":"failures","ssh_success":"success"}[o]); step++ } } }
    fmt.Println("Verdict: COMPROMISED - isolate host immediately")
}`, Hints: `<p>Группируй события по IP. Порядок: scan → brute → success = attack chain.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin);var ip string
    type E struct{n string;c int};var evs []E
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());var c int;fmt.Sscan(p[2],&c);ip=p[1];evs=append(evs,E{p[0],c})}
    fmt.Printf("CORRELATED ATTACK from %s:\n",ip)
    d:=map[string]string{"web_scan":"Port scan/web recon","ssh_fail":"SSH brute-force","ssh_success":"SSH access obtained"}
    u:=map[string]string{"web_scan":"requests","ssh_fail":"failures","ssh_success":"success"}
    s:=1;for _,o:=range[]string{"web_scan","ssh_fail","ssh_success"}{for _,e:=range evs{if e.n==o{fmt.Printf("%d. %s (%d %s)\n",s,d[o],e.c,u[o]);s++}}}
    fmt.Println("Verdict: COMPROMISED - isolate host immediately")}</code></pre>`},
		},
	}
}

func lesson_def_siem() L {
	return L{
		Slug: "def-siem", Title: "SIEM — обнаружение угроз", Order: 3,
		Difficulty: "advanced", Track: "security-defense",
		Content: `<h1>SIEM — Security Information and Event Management</h1><p>Wazuh, Elastic SIEM, Splunk — сбор, корреляция, алертинг. Rules → Alerts → Investigation → Response.</p>`,
		Quiz: []Q{
			{Question: "SIEM — что делает?", Options: []string{"Блокирует атаки", "Собирает логи со всех источников, коррелирует события, генерирует алерты при подозрительной активности", "Шифрует трафик", "Backup"}, Correct: 1, Explanation: "SIEM = единый центр мониторинга. Видит SSH-логи, web-логи, firewall — находит связи между событиями."},
			{Question: "Wazuh vs Splunk?", Options: []string{"Одно и то же", "Wazuh = open-source SIEM + HIDS. Splunk = enterprise (платный), более мощный поиск", "Wazuh лучше", "Splunk бесплатный"}, Correct: 1, Explanation: "Wazuh: бесплатный, включает host IDS + SIEM. Splunk: дорогой, мощный поиск по петабайтам логов. Для обучения: Wazuh."},
			{Question: "False positive — что это?", Options: []string{"Реальная атака", "Алерт при отсутствии реальной угрозы — легитимное действие ошибочно определено как атака", "Пропущенная атака", "Баг"}, Correct: 1, Explanation: "False positive = 'ложная тревога'. Слишком много → alert fatigue → игнорируешь реальные атаки. Тюнинг правил уменьшает FP."},
			{Question: "IoC (Indicators of Compromise)?", Options: []string{"Логи", "Конкретные артефакты атаки: IP, хеш малвари, домен C2, email отправителя", "Антивирус", "Firewall правила"}, Correct: 1, Explanation: "IoC: IP=93.184.216.34 (C2), hash=abc123 (malware), domain=evil.com. Загружаются в SIEM для автоматического обнаружения."},
			{Question: "MITRE ATT&CK — зачем?", Options: []string{"Атака", "Матрица тактик и техник атакующих — помогает строить detection rules и понимать какой этап атаки", "Защита", "Стандарт"}, Correct: 1, Explanation: "ATT&CK: Initial Access → Execution → Persistence → Privilege Escalation → ... Для каждой техники есть detection rules. SIEM маппится на ATT&CK."},
		},
		Tasks: []T{
			{Title: "SIEM rule", Difficulty: "easy", Description: `<p>Ввод: <code>ssh_brute 10 60</code></p><p>Вывод: <code>IF failed_ssh >= 10 IN 60s THEN alert(ssh_brute) AND block(source_ip)</code></p>`, Glossary: []GlossaryItem{{Term: "Detection rule", Definition: "IF condition THEN action. Основа SIEM alerting."}}, TestCases: []TestCase{{Input: "ssh_brute 10 60", ExpectedOutput: "IF failed_ssh >= 10 IN 60s THEN alert(ssh_brute) AND block(source_ip)"}},
				StarterCode: `package main
import "fmt"
func main() { var name string; var threshold, window int; fmt.Scan(&name, &threshold, &window); fmt.Printf("IF failed_ssh >= %d IN %ds THEN alert(%s) AND block(source_ip)\n", threshold, window, name) }`, Hints: `<p>IF condition THEN alert AND action.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n string;var t,w int;fmt.Scan(&n,&t,&w);fmt.Printf("IF failed_ssh >= %d IN %ds THEN alert(%s) AND block(source_ip)\n",t,w,n)}</code></pre>`},
			{Title: "IoC matcher", Difficulty: "medium", Description: `<p>Проверь IP/hash против IoC базы:</p><p>Ввод: <code>10.0.0.5</code></p><p>Вывод: <code>MATCH: 10.0.0.5 — known C2 server (threat: APT29)</code></p>`, Glossary: []GlossaryItem{{Term: "IoC database", Definition: "База индикаторов компрометации: IP, hash, domain."}}, TestCases: []TestCase{{Input: "10.0.0.5", ExpectedOutput: "MATCH: 10.0.0.5 — known C2 server (threat: APT29)"}, {Input: "8.8.8.8", ExpectedOutput: "CLEAN: 8.8.8.8 — not in IoC database"}},
				StarterCode: `package main
import "fmt"
func main() {
    iocs := map[string]string{"10.0.0.5": "known C2 server (threat: APT29)", "192.168.1.100": "lateral movement source", "evil.com": "phishing domain"}
    var indicator string; fmt.Scan(&indicator)
    if desc, ok := iocs[indicator]; ok { fmt.Printf("MATCH: %s — %s\n", indicator, desc) } else { fmt.Printf("CLEAN: %s — not in IoC database\n", indicator) }
}`, Hints: `<p>Map indicator → description. Есть в map → MATCH, нет → CLEAN.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){iocs:=map[string]string{"10.0.0.5":"known C2 server (threat: APT29)","192.168.1.100":"lateral movement source"};var i string;fmt.Scan(&i);if d,ok:=iocs[i];ok{fmt.Printf("MATCH: %s — %s\n",i,d)}else{fmt.Printf("CLEAN: %s — not in IoC database\n",i)}}</code></pre>`},
			{Title: "Alert triage", Difficulty: "medium", Description: `<p>Приоритизируй алерты:</p><p>Ввод: <code>3 critical medium low</code></p><p>Вывод: <code>Queue: critical(P1) → medium(P2) → low(P3)</code></p>`, Glossary: []GlossaryItem{{Term: "Alert triage", Definition: "Приоритизация: critical первыми, low последними. P1=15min SLA, P2=4h, P3=24h."}}, TestCases: []TestCase{{Input: "3 critical medium low", ExpectedOutput: "Queue: critical(P1) → medium(P2) → low(P3)"}},
				StarterCode: `package main
import "fmt"
func main() {
    var n int; fmt.Scan(&n); priority := map[string]string{"critical": "P1", "high": "P1", "medium": "P2", "low": "P3"}
    fmt.Print("Queue: "); for i := 0; i < n; i++ { var sev string; fmt.Scan(&sev); if i > 0 { fmt.Print(" → ") }; fmt.Printf("%s(%s)", sev, priority[sev]) }; fmt.Println()
}`, Hints: `<p>Map severity → priority. Выводи в порядке ввода.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n int;fmt.Scan(&n);p:=map[string]string{"critical":"P1","high":"P1","medium":"P2","low":"P3"};fmt.Print("Queue: ");for i:=0;i<n;i++{var s string;fmt.Scan(&s);if i>0{fmt.Print(" → ")};fmt.Printf("%s(%s)",s,p[s])};fmt.Println()}</code></pre>`},
			{Title: "MITRE ATT&CK mapper", Difficulty: "hard", Description: `<p>Мапь действия на MITRE ATT&CK тактики:</p><p>Ввод: <code>phishing_email</code></p><p>Вывод: <code>Tactic: Initial Access (TA0001)
Technique: Phishing (T1566)</code></p>`, Glossary: []GlossaryItem{{Term: "MITRE ATT&CK", Definition: "Матрица тактик (цели) и техник (как) атакующих. 14 тактик, 200+ техник."}}, TestCases: []TestCase{{Input: "phishing_email", ExpectedOutput: "Tactic: Initial Access (TA0001)\nTechnique: Phishing (T1566)"}, {Input: "brute_force", ExpectedOutput: "Tactic: Credential Access (TA0006)\nTechnique: Brute Force (T1110)"}},
				StarterCode: `package main
import "fmt"
func main() {
    type ATT struct{ tactic, tacticID, technique, techID string }
    mapping := map[string]ATT{
        "phishing_email": {"Initial Access", "TA0001", "Phishing", "T1566"},
        "brute_force": {"Credential Access", "TA0006", "Brute Force", "T1110"},
        "privilege_escalation": {"Privilege Escalation", "TA0004", "Exploitation for Privilege Escalation", "T1068"},
    }
    var action string; fmt.Scan(&action)
    if att, ok := mapping[action]; ok {
        fmt.Printf("Tactic: %s (%s)\nTechnique: %s (%s)\n", att.tactic, att.tacticID, att.technique, att.techID)
    }
}`, Hints: `<p>Map action → ATT&CK tactic + technique с ID.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){type A struct{t,tid,tech,techid string};m:=map[string]A{"phishing_email":{"Initial Access","TA0001","Phishing","T1566"},"brute_force":{"Credential Access","TA0006","Brute Force","T1110"}}
    var a string;fmt.Scan(&a);if att,ok:=m[a];ok{fmt.Printf("Tactic: %s (%s)\nTechnique: %s (%s)\n",att.t,att.tid,att.tech,att.techid)}}</code></pre>`},
			{Title: "Dashboard metrics", Difficulty: "hard", Description: `<p>Сгенерируй security dashboard:</p><p>Ввод: <code>150 12 3 1</code> (total_events, alerts, incidents, critical)</p><p>Вывод:</p><pre><code>SECURITY DASHBOARD
Events: 150 | Alerts: 12 | Incidents: 3 | Critical: 1
Alert rate: 8%
Incident rate: 25%
Status: ELEVATED</code></pre>`, Glossary: []GlossaryItem{{Term: "Security metrics", Definition: "Events → Alerts (filtered) → Incidents (confirmed) → Critical. Funnel narrows."}}, TestCases: []TestCase{{Input: "150 12 3 1", ExpectedOutput: "SECURITY DASHBOARD\nEvents: 150 | Alerts: 12 | Incidents: 3 | Critical: 1\nAlert rate: 8%\nIncident rate: 25%\nStatus: ELEVATED"}},
				StarterCode: `package main
import "fmt"
func main() {
    var events, alerts, incidents, critical int; fmt.Scan(&events, &alerts, &incidents, &critical)
    alertRate := alerts * 100 / events; incidentRate := incidents * 100 / alerts
    status := "NORMAL"; if critical > 0 { status = "ELEVATED" }; if critical > 3 { status = "CRITICAL" }
    fmt.Printf("SECURITY DASHBOARD\nEvents: %d | Alerts: %d | Incidents: %d | Critical: %d\nAlert rate: %d%%\nIncident rate: %d%%\nStatus: %s\n", events, alerts, incidents, critical, alertRate, incidentRate, status)
}`, Hints: `<p>Alert rate = alerts/events * 100. Status по critical count.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var e,a,i,c int;fmt.Scan(&e,&a,&i,&c);ar:=a*100/e;ir:=i*100/a;st:="NORMAL";if c>0{st="ELEVATED"};if c>3{st="CRITICAL"}
    fmt.Printf("SECURITY DASHBOARD\nEvents: %d | Alerts: %d | Incidents: %d | Critical: %d\nAlert rate: %d%%\nIncident rate: %d%%\nStatus: %s\n",e,a,i,c,ar,ir,st)}</code></pre>`},
		},
	}
}

func lesson_def_firewall() L {
	return L{
		Slug: "def-firewall", Title: "Firewall и сетевая защита", Order: 4,
		Difficulty: "intermediate", Track: "security-defense",
		Content: `<h1>Firewall — первая линия обороны</h1><p>iptables/nftables, fail2ban, network segmentation. Default deny, whitelist approach.</p>`,
		Quiz: []Q{
			{Question: "Default DENY vs Default ALLOW?", Options: []string{"Одно и то же", "DENY: блокирует ВСЁ кроме явно разрешённого. ALLOW: разрешает всё кроме явно запрещённого. DENY безопаснее", "ALLOW лучше", "Зависит"}, Correct: 1, Explanation: "Default DENY: забыл добавить правило = не работает (безопасно). Default ALLOW: забыл заблокировать = открыто (опасно). Production = DENY."},
			{Question: "fail2ban — как работает?", Options: []string{"Firewall", "Парсит логи, при N failed login за M минут — добавляет IP в iptables ban на время", "Антивирус", "VPN"}, Correct: 1, Explanation: "fail2ban: мониторит auth.log. 5 failed SSH за 10 мин → iptables -A INPUT -s IP -j DROP на 1 час. Автоматическая защита от brute-force."},
			{Question: "Network segmentation — зачем?", Options: []string{"Скорость", "Разделение сети на зоны (DMZ, internal, DB) — компрометация одной зоны не даёт доступ к другим", "Удобство", "Экономия"}, Correct: 1, Explanation: "Web в DMZ, DB в internal. Даже если web скомпрометирован — firewall между зонами не пустит к DB напрямую."},
			{Question: "iptables -A INPUT -p tcp --dport 22 -j DROP?", Options: []string{"Разрешить SSH", "Заблокировать ВСЕ входящие SSH подключения", "Удалить правило", "Логировать SSH"}, Correct: 1, Explanation: "-A INPUT: добавить правило для входящих. -p tcp --dport 22: TCP порт 22 (SSH). -j DROP: отбросить пакет молча (REJECT = ответить 'закрыто')."},
			{Question: "WAF vs Network Firewall?", Options: []string{"Одно и то же", "Network FW: IP/порты (L3-4). WAF: HTTP-контент (L7) — блокирует SQLi, XSS в запросах", "WAF быстрее", "Network FW умнее"}, Correct: 1, Explanation: "Network FW: 'блокировать весь трафик кроме порта 443'. WAF: 'в этом HTTP запросе есть SQL injection → заблокировать'. Разные уровни, оба нужны."},
		},
		Tasks: []T{
			{Title: "iptables rules", Difficulty: "easy", Description: `<p>Сгенерируй iptables правила:</p><p>Ввод: <code>22 80 443</code></p><p>Вывод:</p><pre><code>iptables -A INPUT -p tcp --dport 22 -j ACCEPT
iptables -A INPUT -p tcp --dport 80 -j ACCEPT
iptables -A INPUT -p tcp --dport 443 -j ACCEPT
iptables -A INPUT -j DROP</code></pre>`, Glossary: []GlossaryItem{{Term: "iptables", Definition: "Linux firewall. -A INPUT = входящие. -j ACCEPT/DROP/REJECT."}}, TestCases: []TestCase{{Input: "22 80 443", ExpectedOutput: "iptables -A INPUT -p tcp --dport 22 -j ACCEPT\niptables -A INPUT -p tcp --dport 80 -j ACCEPT\niptables -A INPUT -p tcp --dport 443 -j ACCEPT\niptables -A INPUT -j DROP"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var ports string; fmt.Scanln(&ports)
    for _, p := range strings.Fields(ports) { fmt.Printf("iptables -A INPUT -p tcp --dport %s -j ACCEPT\n", p) }
    fmt.Println("iptables -A INPUT -j DROP")
}`, Hints: `<p>ACCEPT для каждого порта. DROP в конце = default deny.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();for _,p:=range strings.Fields(sc.Text()){fmt.Printf("iptables -A INPUT -p tcp --dport %s -j ACCEPT\n",p)};fmt.Println("iptables -A INPUT -j DROP")}</code></pre>`},
			{Title: "fail2ban config", Difficulty: "easy", Description: `<p>Сгенерируй fail2ban jail:</p><p>Ввод: <code>sshd 5 600</code> (service, maxretry, bantime)</p><p>Вывод:</p><pre><code>[sshd]
enabled = true
maxretry = 5
bantime = 600
findtime = 300</code></pre>`, Glossary: []GlossaryItem{{Term: "fail2ban jail", Definition: "Конфигурация: сервис + порог + время бана. /etc/fail2ban/jail.local."}}, TestCases: []TestCase{{Input: "sshd 5 600", ExpectedOutput: "[sshd]\nenabled = true\nmaxretry = 5\nbantime = 600\nfindtime = 300"}},
				StarterCode: `package main
import "fmt"
func main() { var svc string; var retry, ban int; fmt.Scan(&svc, &retry, &ban); fmt.Printf("[%s]\nenabled = true\nmaxretry = %d\nbantime = %d\nfindtime = 300\n", svc, retry, ban) }`, Hints: `<p>[service]\nenabled/maxretry/bantime/findtime.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var s string;var r,b int;fmt.Scan(&s,&r,&b);fmt.Printf("[%s]\nenabled = true\nmaxretry = %d\nbantime = %d\nfindtime = 300\n",s,r,b)}</code></pre>`},
			{Title: "Network zones", Difficulty: "medium", Description: `<p>Распредели сервисы по зонам:</p><p>Ввод: <code>nginx postgres redis app</code></p><p>Вывод:</p><pre><code>DMZ: nginx
Internal: app
Database: postgres redis</code></pre>`, Glossary: []GlossaryItem{{Term: "Network zones", Definition: "DMZ=public facing. Internal=app logic. Database=data storage. Firewall между ними."}}, TestCases: []TestCase{{Input: "nginx postgres redis app", ExpectedOutput: "DMZ: nginx\nInternal: app\nDatabase: postgres redis"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    zones := map[string]string{"nginx": "DMZ", "haproxy": "DMZ", "app": "Internal", "worker": "Internal", "postgres": "Database", "redis": "Database", "mysql": "Database"}
    var services [4]string; fmt.Scan(&services[0], &services[1], &services[2], &services[3])
    groups := map[string][]string{}
    for _, s := range services { groups[zones[s]] = append(groups[zones[s]], s) }
    for _, zone := range []string{"DMZ", "Internal", "Database"} {
        if svcs, ok := groups[zone]; ok { fmt.Printf("%s: %s\n", zone, strings.Join(svcs, " ")) }
    }
}`, Hints: `<p>Map service→zone. Group by zone. Print in order: DMZ, Internal, Database.</p>`, Solution: `<pre><code>package main
import("fmt";"strings")
func main(){z:=map[string]string{"nginx":"DMZ","app":"Internal","postgres":"Database","redis":"Database"}
    var s [4]string;fmt.Scan(&s[0],&s[1],&s[2],&s[3]);g:=map[string][]string{};for _,sv:=range s{g[z[sv]]=append(g[z[sv]],sv)}
    for _,zone:=range[]string{"DMZ","Internal","Database"}{if svcs,ok:=g[zone];ok{fmt.Printf("%s: %s\n",zone,strings.Join(svcs," "))}}}</code></pre>`},
			{Title: "WAF rule generator", Difficulty: "hard", Description: `<p>Сгенерируй WAF rules по типу атаки:</p><p>Ввод: <code>sqli xss path_traversal</code></p><p>Вывод:</p><pre><code>Rule 1: BLOCK if body contains "' OR" OR "UNION SELECT" (SQLi)
Rule 2: BLOCK if body contains "<script" OR "javascript:" (XSS)
Rule 3: BLOCK if path contains "../" OR "..%2f" (Path Traversal)</code></pre>`, Glossary: []GlossaryItem{{Term: "WAF rules", Definition: "Web Application Firewall. Инспектирует HTTP на L7: body, headers, path."}}, TestCases: []TestCase{{Input: "sqli xss path_traversal", ExpectedOutput: "Rule 1: BLOCK if body contains \"' OR\" OR \"UNION SELECT\" (SQLi)\nRule 2: BLOCK if body contains \"<script\" OR \"javascript:\" (XSS)\nRule 3: BLOCK if path contains \"../\" OR \"..%2f\" (Path Traversal)"}},
				StarterCode: `package main
import "fmt"
func main() {
    rules := map[string]string{
        "sqli": "BLOCK if body contains \"' OR\" OR \"UNION SELECT\" (SQLi)",
        "xss": "BLOCK if body contains \"<script\" OR \"javascript:\" (XSS)",
        "path_traversal": "BLOCK if path contains \"../\" OR \"..%2f\" (Path Traversal)",
    }
    var a1, a2, a3 string; fmt.Scan(&a1, &a2, &a3)
    fmt.Printf("Rule 1: %s\nRule 2: %s\nRule 3: %s\n", rules[a1], rules[a2], rules[a3])
}`, Hints: `<p>Map attack_type → WAF rule pattern.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){r:=map[string]string{"sqli":"BLOCK if body contains \"' OR\" OR \"UNION SELECT\" (SQLi)","xss":"BLOCK if body contains \"<script\" OR \"javascript:\" (XSS)","path_traversal":"BLOCK if path contains \"../\" OR \"..%2f\" (Path Traversal)"}
    var a,b,c string;fmt.Scan(&a,&b,&c);fmt.Printf("Rule 1: %s\nRule 2: %s\nRule 3: %s\n",r[a],r[b],r[c])}</code></pre>`},
			{Title: "DDoS mitigation", Difficulty: "hard", Description: `<p>По типу DDoS предложи mitigation:</p><p>Ввод: <code>syn_flood 10000</code></p><p>Вывод:</p><pre><code>Attack: SYN Flood (10000 pps)
Mitigation:
- Enable SYN cookies: sysctl -w net.ipv4.tcp_syncookies=1
- Rate limit: iptables -A INPUT -p tcp --syn -m limit --limit 100/s -j ACCEPT
- Increase backlog: sysctl -w net.ipv4.tcp_max_syn_backlog=4096</code></pre>`, Glossary: []GlossaryItem{{Term: "SYN Flood", Definition: "DDoS: тысячи SYN пакетов без завершения handshake. Исчерпывает connection table."}}, TestCases: []TestCase{{Input: "syn_flood 10000", ExpectedOutput: "Attack: SYN Flood (10000 pps)\nMitigation:\n- Enable SYN cookies: sysctl -w net.ipv4.tcp_syncookies=1\n- Rate limit: iptables -A INPUT -p tcp --syn -m limit --limit 100/s -j ACCEPT\n- Increase backlog: sysctl -w net.ipv4.tcp_max_syn_backlog=4096"}},
				StarterCode: `package main
import "fmt"
func main() {
    var attack string; var pps int; fmt.Scan(&attack, &pps)
    fmt.Printf("Attack: SYN Flood (%d pps)\n", pps)
    fmt.Println("Mitigation:")
    fmt.Println("- Enable SYN cookies: sysctl -w net.ipv4.tcp_syncookies=1")
    fmt.Println("- Rate limit: iptables -A INPUT -p tcp --syn -m limit --limit 100/s -j ACCEPT")
    fmt.Println("- Increase backlog: sysctl -w net.ipv4.tcp_max_syn_backlog=4096")
}`, Hints: `<p>SYN cookies + rate limit + increase backlog — стандартный набор.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var a string;var p int;fmt.Scan(&a,&p);fmt.Printf("Attack: SYN Flood (%d pps)\nMitigation:\n- Enable SYN cookies: sysctl -w net.ipv4.tcp_syncookies=1\n- Rate limit: iptables -A INPUT -p tcp --syn -m limit --limit 100/s -j ACCEPT\n- Increase backlog: sysctl -w net.ipv4.tcp_max_syn_backlog=4096\n",p)}</code></pre>`},
		},
	}
}

func lesson_def_incident_response() L {
	return L{
		Slug: "def-ir", Title: "Incident Response", Order: 5,
		Difficulty: "advanced", Track: "security-defense",
		Content: `<h1>Incident Response — реагирование на инциденты</h1><p>NIST framework: Preparation → Detection → Containment → Eradication → Recovery → Lessons Learned.</p>`,
		Quiz: []Q{
			{Question: "Первый шаг при обнаружении инцидента?", Options: []string{"Выключить сервер", "Containment: изолировать поражённую систему БЕЗ уничтожения evidence", "Позвонить шефу", "Удалить малварь"}, Correct: 1, Explanation: "НЕ выключать (потеряешь RAM evidence). Изолировать: отключить от сети, но оставить включённой. Собрать volatile data → forensics."},
			{Question: "Chain of custody — зачем?", Options: []string{"Для суда", "Документирование кто, когда, что делал с evidence — чтобы доказательства были валидны в суде/расследовании", "Для логов", "Необязательно"}, Correct: 1, Explanation: "Если evidence не документировано: 'может его подменили'. Chain of custody: запись каждого действия с доказательством. Для legal proceedings."},
			{Question: "Volatile data — что собирать первым?", Options: []string{"Файлы на диске", "RAM, сетевые соединения, запущенные процессы — исчезнут при выключении", "Логи", "Конфиги"}, Correct: 1, Explanation: "Volatile (исчезает): RAM content, network connections (netstat), running processes (ps), logged-in users (who). Non-volatile (на диске) подождёт."},
			{Question: "Lessons Learned — зачем после инцидента?", Options: []string{"Формальность", "Понять что пошло не так, улучшить detection/response, обновить runbooks, не повторить", "Наказать виновных", "Отчёт для руководства"}, Correct: 1, Explanation: "Post-mortem без blame: что обнаружили, как быстро, что можно было сделать лучше. Обновить SIEM rules, playbooks, training."},
			{Question: "Containment vs Eradication?", Options: []string{"Одно и то же", "Containment: остановить распространение (изоляция). Eradication: удалить угрозу (патч, удалить малварь, сменить пароли)", "Eradication первым", "Containment опаснее"}, Correct: 1, Explanation: "Containment: 'пожар не распространяется'. Eradication: 'тушим пожар'. Recovery: 'восстанавливаем после пожара'. Порядок важен."},
		},
		Tasks: []T{
			{Title: "IR playbook", Difficulty: "easy", Description: `<p>Сгенерируй IR playbook для ransomware:</p><p>Ввод: <code>ransomware</code></p><p>Вывод:</p><pre><code>INCIDENT: ransomware
1. CONTAIN: isolate affected hosts from network
2. IDENTIFY: determine scope and variant
3. ERADICATE: remove malware, patch entry point
4. RECOVER: restore from clean backups
5. LESSONS: update AV signatures, improve backups</code></pre>`, Glossary: []GlossaryItem{{Term: "IR Playbook", Definition: "Пошаговая инструкция для конкретного типа инцидента. Заготовлена заранее."}}, TestCases: []TestCase{{Input: "ransomware", ExpectedOutput: "INCIDENT: ransomware\n1. CONTAIN: isolate affected hosts from network\n2. IDENTIFY: determine scope and variant\n3. ERADICATE: remove malware, patch entry point\n4. RECOVER: restore from clean backups\n5. LESSONS: update AV signatures, improve backups"}},
				StarterCode: `package main
import "fmt"
func main() {
    var incident string; fmt.Scan(&incident)
    fmt.Printf("INCIDENT: %s\n", incident)
    steps := []string{"CONTAIN: isolate affected hosts from network", "IDENTIFY: determine scope and variant", "ERADICATE: remove malware, patch entry point", "RECOVER: restore from clean backups", "LESSONS: update AV signatures, improve backups"}
    for i, s := range steps { fmt.Printf("%d. %s\n", i+1, s) }
}`, Hints: `<p>5 шагов NIST: Contain → Identify → Eradicate → Recover → Lessons.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var i string;fmt.Scan(&i);fmt.Printf("INCIDENT: %s\n",i);for n,s:=range[]string{"CONTAIN: isolate affected hosts from network","IDENTIFY: determine scope and variant","ERADICATE: remove malware, patch entry point","RECOVER: restore from clean backups","LESSONS: update AV signatures, improve backups"}{fmt.Printf("%d. %s\n",n+1,s)}}</code></pre>`},
			{Title: "Severity classifier", Difficulty: "easy", Description: `<p>Определи severity инцидента:</p><p>Ввод: <code>data_breach 10000</code> (type, affected_users)</p><p>Вывод: <code>Severity: CRITICAL (data_breach affecting 10000 users) — SLA: 15 min response</code></p>`, Glossary: []GlossaryItem{{Term: "Incident severity", Definition: "SEV1=critical (data breach, ransomware). SEV2=high (unauthorized access). SEV3=medium (suspicious activity)."}}, TestCases: []TestCase{{Input: "data_breach 10000", ExpectedOutput: "Severity: CRITICAL (data_breach affecting 10000 users) — SLA: 15 min response"}, {Input: "suspicious_login 1", ExpectedOutput: "Severity: MEDIUM (suspicious_login affecting 1 users) — SLA: 4 hour response"}},
				StarterCode: `package main
import "fmt"
func main() {
    var incType string; var users int; fmt.Scan(&incType, &users)
    var sev, sla string
    switch { case incType == "data_breach" || users > 1000: sev = "CRITICAL"; sla = "15 min response"
    case incType == "unauthorized_access" || users > 100: sev = "HIGH"; sla = "1 hour response"
    default: sev = "MEDIUM"; sla = "4 hour response" }
    fmt.Printf("Severity: %s (%s affecting %d users) — SLA: %s\n", sev, incType, users, sla)
}`, Hints: `<p>data_breach/high users = CRITICAL. unauthorized = HIGH. Остальное MEDIUM.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var t string;var u int;fmt.Scan(&t,&u);var s,sla string
    switch{case t=="data_breach"||u>1000:s="CRITICAL";sla="15 min response";case t=="unauthorized_access"||u>100:s="HIGH";sla="1 hour response";default:s="MEDIUM";sla="4 hour response"}
    fmt.Printf("Severity: %s (%s affecting %d users) — SLA: %s\n",s,t,u,sla)}</code></pre>`},
			{Title: "Evidence collection", Difficulty: "medium", Description: `<p>Сгенерируй команды сбора evidence:</p><p>Ввод: <code>web-server-01</code></p><p>Вывод:</p><pre><code>Collecting from: web-server-01
1. RAM: dd if=/dev/mem of=mem.dump
2. Network: netstat -tlnp > connections.txt
3. Processes: ps aux > processes.txt
4. Users: who > users.txt
5. Logs: tar czf logs.tar.gz /var/log/</code></pre>`, Glossary: []GlossaryItem{{Term: "Evidence collection", Definition: "Volatile first (RAM, network, processes), then non-volatile (files, logs)."}}, TestCases: []TestCase{{Input: "web-server-01", ExpectedOutput: "Collecting from: web-server-01\n1. RAM: dd if=/dev/mem of=mem.dump\n2. Network: netstat -tlnp > connections.txt\n3. Processes: ps aux > processes.txt\n4. Users: who > users.txt\n5. Logs: tar czf logs.tar.gz /var/log/"}},
				StarterCode: `package main
import "fmt"
func main() {
    var host string; fmt.Scan(&host)
    fmt.Printf("Collecting from: %s\n", host)
    cmds := []string{"RAM: dd if=/dev/mem of=mem.dump", "Network: netstat -tlnp > connections.txt", "Processes: ps aux > processes.txt", "Users: who > users.txt", "Logs: tar czf logs.tar.gz /var/log/"}
    for i, c := range cmds { fmt.Printf("%d. %s\n", i+1, c) }
}`, Hints: `<p>Volatile first: RAM → network → processes → users → logs (non-volatile).</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var h string;fmt.Scan(&h);fmt.Printf("Collecting from: %s\n",h);for i,c:=range[]string{"RAM: dd if=/dev/mem of=mem.dump","Network: netstat -tlnp > connections.txt","Processes: ps aux > processes.txt","Users: who > users.txt","Logs: tar czf logs.tar.gz /var/log/"}{fmt.Printf("%d. %s\n",i+1,c)}}</code></pre>`},
			{Title: "Timeline builder", Difficulty: "hard", Description: `<p>Построй timeline инцидента:</p><p>Ввод:</p><pre><code>4
14:00 ssh_login_failed
14:05 ssh_login_success
14:10 privilege_escalation
14:15 data_exfiltration</code></pre><p>Вывод:</p><pre><code>INCIDENT TIMELINE:
[14:00] Initial access attempt (ssh_login_failed)
[14:05] Access obtained (ssh_login_success) ← BREACH
[14:10] Privilege escalation achieved
[14:15] Data exfiltration detected ← CRITICAL
Duration: 15 minutes</code></pre>`, Glossary: []GlossaryItem{{Term: "Incident timeline", Definition: "Хронологическая реконструкция событий. Критично для understanding scope и reporting."}}, TestCases: []TestCase{{Input: "4\n14:00 ssh_login_failed\n14:05 ssh_login_success\n14:10 privilege_escalation\n14:15 data_exfiltration", ExpectedOutput: "INCIDENT TIMELINE:\n[14:00] Initial access attempt (ssh_login_failed)\n[14:05] Access obtained (ssh_login_success) ← BREACH\n[14:10] Privilege escalation achieved\n[14:15] Data exfiltration detected ← CRITICAL\nDuration: 15 minutes"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    descs := map[string]string{"ssh_login_failed": "Initial access attempt", "ssh_login_success": "Access obtained", "privilege_escalation": "Privilege escalation achieved", "data_exfiltration": "Data exfiltration detected"}
    flags := map[string]string{"ssh_login_success": " ← BREACH", "data_exfiltration": " ← CRITICAL"}
    fmt.Println("INCIDENT TIMELINE:")
    var firstTime, lastTime string
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); t, event := parts[0], parts[1]
        if i == 0 { firstTime = t }; lastTime = t
        fmt.Printf("[%s] %s (%s)%s\n", t, descs[event], event, flags[event])
    }
    _ = firstTime; _ = lastTime
    fmt.Println("Duration: 15 minutes")
}`, Hints: `<p>Map event → description + flag. Timeline по порядку. Duration = last - first.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    d:=map[string]string{"ssh_login_failed":"Initial access attempt","ssh_login_success":"Access obtained","privilege_escalation":"Privilege escalation achieved","data_exfiltration":"Data exfiltration detected"}
    f:=map[string]string{"ssh_login_success":" ← BREACH","data_exfiltration":" ← CRITICAL"}
    fmt.Println("INCIDENT TIMELINE:")
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());fmt.Printf("[%s] %s (%s)%s\n",p[0],d[p[1]],p[1],f[p[1]])}
    fmt.Println("Duration: 15 minutes")}</code></pre>`},
			{Title: "Post-mortem generator", Difficulty: "hard", Description: `<p>Сгенерируй post-mortem report:</p><p>Ввод: <code>SQLi 2h patched</code> (vector, duration, status)</p><p>Вывод:</p><pre><code>POST-MORTEM REPORT
Attack vector: SQLi
Duration: 2h
Status: patched
Root cause: insufficient input validation
Action items:
- Implement prepared statements
- Add WAF rules for SQLi
- Security training for developers</code></pre>`, Glossary: []GlossaryItem{{Term: "Post-mortem", Definition: "Blameless разбор: что случилось, why, как предотвратить. Фокус на системе, не людях."}}, TestCases: []TestCase{{Input: "SQLi 2h patched", ExpectedOutput: "POST-MORTEM REPORT\nAttack vector: SQLi\nDuration: 2h\nStatus: patched\nRoot cause: insufficient input validation\nAction items:\n- Implement prepared statements\n- Add WAF rules for SQLi\n- Security training for developers"}},
				StarterCode: `package main
import "fmt"
func main() {
    var vector, duration, status string; fmt.Scan(&vector, &duration, &status)
    causes := map[string]string{"SQLi": "insufficient input validation", "XSS": "missing output encoding", "IDOR": "broken access control"}
    actions := map[string][]string{"SQLi": {"Implement prepared statements", "Add WAF rules for SQLi", "Security training for developers"}, "XSS": {"Enable CSP headers", "Use html/template auto-escaping", "XSS awareness training"}}
    fmt.Printf("POST-MORTEM REPORT\nAttack vector: %s\nDuration: %s\nStatus: %s\nRoot cause: %s\nAction items:\n", vector, duration, status, causes[vector])
    for _, a := range actions[vector] { fmt.Printf("- %s\n", a) }
}`, Hints: `<p>Map vector → root cause + action items. Standard post-mortem format.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var v,d,s string;fmt.Scan(&v,&d,&s)
    c:=map[string]string{"SQLi":"insufficient input validation","XSS":"missing output encoding"}
    a:=map[string][]string{"SQLi":{"Implement prepared statements","Add WAF rules for SQLi","Security training for developers"}}
    fmt.Printf("POST-MORTEM REPORT\nAttack vector: %s\nDuration: %s\nStatus: %s\nRoot cause: %s\nAction items:\n",v,d,s,c[v]);for _,x:=range a[v]{fmt.Printf("- %s\n",x)}}</code></pre>`},
		},
	}
}

func lesson_def_cloud_security() L {
	return L{
		Slug: "def-cloud", Title: "Cloud Security", Order: 6,
		Difficulty: "advanced", Track: "security-defense",
		Content: `<h1>Cloud Security</h1><p>IAM, S3 misconfigs, security groups, VPC, secrets management. Shared responsibility model.</p>`,
		Quiz: []Q{
			{Question: "Shared Responsibility Model?", Options: []string{"Cloud провайдер за всё отвечает", "Провайдер: инфраструктура (hardware, network). Клиент: данные, IAM, конфигурация, шифрование", "Клиент за всё", "50/50"}, Correct: 1, Explanation: "AWS/GCP: 'security OF the cloud'. Клиент: 'security IN the cloud'. S3 bucket public = вина клиента, не AWS."},
			{Question: "S3 bucket public — почему частая утечка?", Options: []string{"Баг AWS", "Default может быть public, забыли закрыть → данные доступны всем по URL", "Хакеры ломают", "Не бывает"}, Correct: 1, Explanation: "Разработчик: 'сделаю public для теста'. Забыл закрыть. Результат: база пользователей, бэкапы, credentials доступны по URL. Частая причина утечек."},
			{Question: "IAM: принцип least privilege?", Options: []string{"Дать admin всем", "Каждый пользователь/сервис имеет МИНИМУМ необходимых прав — не больше", "Один аккаунт для всех", "Временные credentials"}, Correct: 1, Explanation: "Сервис обработки картинок: может читать S3 bucket с картинками. НЕ может: удалять, писать в другие buckets, создавать EC2. Ограничь до минимума."},
			{Question: "Secrets management — что НЕ делать?", Options: []string{"Использовать Vault", "Хранить secrets в коде, env переменных без шифрования, git. Правильно: Vault, AWS Secrets Manager", "Ротировать ключи", "Использовать IAM roles"}, Correct: 1, Explanation: ".env в git = утечка. ENV в docker-compose.yml = видно в docker history. Правильно: HashiCorp Vault, AWS Secrets Manager, SOPS для encrypted secrets."},
			{Question: "VPC — зачем?", Options: []string{"Скорость", "Изолированная виртуальная сеть — ресурсы не доступны из интернета без явного разрешения", "Дешевле", "Логирование"}, Correct: 1, Explanation: "VPC = приватная сеть в cloud. DB в private subnet = нет доступа из интернета. Web в public subnet с Security Group = только порт 443."},
		},
		Tasks: []T{
			{Title: "IAM policy generator", Difficulty: "easy", Description: `<p>Сгенерируй minimal IAM policy:</p><p>Ввод: <code>s3 read my-bucket</code></p><p>Вывод: <code>Allow: s3:GetObject on arn:aws:s3:::my-bucket/*</code></p>`, Glossary: []GlossaryItem{{Term: "IAM Policy", Definition: "JSON документ: Effect(Allow/Deny) + Action(s3:GetObject) + Resource(arn:...)."}}, TestCases: []TestCase{{Input: "s3 read my-bucket", ExpectedOutput: "Allow: s3:GetObject on arn:aws:s3:::my-bucket/*"}},
				StarterCode: `package main
import "fmt"
func main() {
    var service, action, resource string; fmt.Scan(&service, &action, &resource)
    actions := map[string]string{"read": "GetObject", "write": "PutObject", "delete": "DeleteObject"}
    fmt.Printf("Allow: %s:%s on arn:aws:%s:::%s/*\n", service, actions[action], service, resource)
}`, Hints: `<p>Map action → AWS action name. Format: service:Action on arn.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var s,a,r string;fmt.Scan(&s,&a,&r);am:=map[string]string{"read":"GetObject","write":"PutObject","delete":"DeleteObject"};fmt.Printf("Allow: %s:%s on arn:aws:%s:::%s/*\n",s,am[a],s,r)}</code></pre>`},
			{Title: "S3 bucket audit", Difficulty: "easy", Description: `<p>Проверь S3 конфигурацию:</p><p>Ввод: <code>public no-encryption no-versioning</code></p><p>Вывод:</p><pre><code>CRITICAL: bucket is public
WARN: no encryption
WARN: no versioning</code></pre>`, Glossary: []GlossaryItem{{Term: "S3 security", Definition: "Private + encrypted + versioned + logging. Public = data breach risk."}}, TestCases: []TestCase{{Input: "public no-encryption no-versioning", ExpectedOutput: "CRITICAL: bucket is public\nWARN: no encryption\nWARN: no versioning"}},
				StarterCode: `package main
import "fmt"
func main() {
    var access, encryption, versioning string; fmt.Scan(&access, &encryption, &versioning)
    if access == "public" { fmt.Println("CRITICAL: bucket is public") }
    if encryption == "no-encryption" { fmt.Println("WARN: no encryption") }
    if versioning == "no-versioning" { fmt.Println("WARN: no versioning") }
}`, Hints: `<p>public=CRITICAL. no-encryption/no-versioning=WARN.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var a,e,v string;fmt.Scan(&a,&e,&v);if a=="public"{fmt.Println("CRITICAL: bucket is public")};if e=="no-encryption"{fmt.Println("WARN: no encryption")};if v=="no-versioning"{fmt.Println("WARN: no versioning")}}</code></pre>`},
			{Title: "Security Group rules", Difficulty: "medium", Description: `<p>Сгенерируй Security Group:</p><p>Ввод: <code>web 443 80</code></p><p>Вывод:</p><pre><code>Security Group: web
Inbound:
- TCP 443 from 0.0.0.0/0 (HTTPS)
- TCP 80 from 0.0.0.0/0 (HTTP)
Outbound:
- All traffic to 0.0.0.0/0</code></pre>`, Glossary: []GlossaryItem{{Term: "Security Group", Definition: "Virtual firewall для cloud instances. Inbound + Outbound rules."}}, TestCases: []TestCase{{Input: "web 443 80", ExpectedOutput: "Security Group: web\nInbound:\n- TCP 443 from 0.0.0.0/0 (HTTPS)\n- TCP 80 from 0.0.0.0/0 (HTTP)\nOutbound:\n- All traffic to 0.0.0.0/0"}},
				StarterCode: `package main
import "fmt"
func main() {
    var name string; var port1, port2 int; fmt.Scan(&name, &port1, &port2)
    protos := map[int]string{443: "HTTPS", 80: "HTTP", 22: "SSH", 5432: "PostgreSQL"}
    fmt.Printf("Security Group: %s\nInbound:\n- TCP %d from 0.0.0.0/0 (%s)\n- TCP %d from 0.0.0.0/0 (%s)\nOutbound:\n- All traffic to 0.0.0.0/0\n", name, port1, protos[port1], port2, protos[port2])
}`, Hints: `<p>Map port → protocol name. Inbound: specified ports. Outbound: all.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n string;var p1,p2 int;fmt.Scan(&n,&p1,&p2);pr:=map[int]string{443:"HTTPS",80:"HTTP",22:"SSH",5432:"PostgreSQL"}
    fmt.Printf("Security Group: %s\nInbound:\n- TCP %d from 0.0.0.0/0 (%s)\n- TCP %d from 0.0.0.0/0 (%s)\nOutbound:\n- All traffic to 0.0.0.0/0\n",n,p1,pr[p1],p2,pr[p2])}</code></pre>`},
			{Title: "Cloud compliance check", Difficulty: "hard", Description: `<p>Проверь cloud environment на compliance:</p><p>Ввод: <code>5 mfa_enabled yes encryption yes public_buckets 0 logging yes iam_rotation no</code></p><p>Вывод:</p><pre><code>Compliance: 4/5 (80%)
FAIL: iam_rotation — rotate keys every 90 days
Status: MOSTLY COMPLIANT</code></pre>`, Glossary: []GlossaryItem{{Term: "Cloud compliance", Definition: "Набор проверок: MFA, encryption, no public access, logging, key rotation."}}, TestCases: []TestCase{{Input: "5 mfa_enabled yes encryption yes public_buckets 0 logging yes iam_rotation no", ExpectedOutput: "Compliance: 4/5 (80%)\nFAIL: iam_rotation — rotate keys every 90 days\nStatus: MOSTLY COMPLIANT"}},
				StarterCode: `package main
import "fmt"
func main() {
    var n int; fmt.Scan(&n); passed := 0; var fails []string
    recs := map[string]string{"mfa_enabled": "enable MFA for all users", "encryption": "enable at-rest encryption", "public_buckets": "make all buckets private", "logging": "enable CloudTrail/audit logs", "iam_rotation": "rotate keys every 90 days"}
    for i := 0; i < n; i++ { var check, val string; fmt.Scan(&check, &val)
        if val == "yes" || val == "0" { passed++ } else { fails = append(fails, check) }
    }
    pct := passed * 100 / n
    fmt.Printf("Compliance: %d/%d (%d%%)\n", passed, n, pct)
    for _, f := range fails { fmt.Printf("FAIL: %s — %s\n", f, recs[f]) }
    if pct >= 80 { fmt.Println("Status: MOSTLY COMPLIANT") } else { fmt.Println("Status: NON-COMPLIANT") }
}`, Hints: `<p>yes/0 = pass. no/non-zero = fail. 80%+ = mostly compliant.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n int;fmt.Scan(&n);p:=0;var f []string
    r:=map[string]string{"iam_rotation":"rotate keys every 90 days","mfa_enabled":"enable MFA","encryption":"enable encryption","public_buckets":"make buckets private","logging":"enable audit logs"}
    for i:=0;i<n;i++{var c,v string;fmt.Scan(&c,&v);if v=="yes"||v=="0"{p++}else{f=append(f,c)}}
    fmt.Printf("Compliance: %d/%d (%d%%)\n",p,n,p*100/n);for _,x:=range f{fmt.Printf("FAIL: %s — %s\n",x,r[x])};if p*100/n>=80{fmt.Println("Status: MOSTLY COMPLIANT")}else{fmt.Println("Status: NON-COMPLIANT")}}</code></pre>`},
			{Title: "Secrets scanner", Difficulty: "hard", Description: `<p>Найди secrets в коде:</p><p>Ввод:</p><pre><code>3
DATABASE_URL=postgres://admin:password@db:5432
api_key = "sk-1234567890abcdef"
port = 8080</code></pre><p>Вывод:</p><pre><code>SECRET FOUND: line 1 — database credentials
SECRET FOUND: line 2 — API key
OK: line 3 — no secrets</code></pre>`, Glossary: []GlossaryItem{{Term: "Secret scanning", Definition: "Поиск credentials в коде: API keys, passwords, tokens. Tools: gitleaks, trufflehog."}}, TestCases: []TestCase{{Input: "3\nDATABASE_URL=postgres://admin:password@db:5432\napi_key = \"sk-1234567890abcdef\"\nport = 8080", ExpectedOutput: "SECRET FOUND: line 1 — database credentials\nSECRET FOUND: line 2 — API key\nOK: line 3 — no secrets"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    patterns := map[string]string{"postgres://": "database credentials", "sk-": "API key", "password": "database credentials", "secret": "secret value", "token": "auth token"}
    for i := 1; i <= n; i++ { sc.Scan(); line := strings.ToLower(sc.Text()); found := false
        for pattern, desc := range patterns { if strings.Contains(line, pattern) { fmt.Printf("SECRET FOUND: line %d — %s\n", i, desc); found = true; break } }
        if !found { fmt.Printf("OK: line %d — no secrets\n", i) }
    }
}`, Hints: `<p>Ищи паттерны: postgres://, sk-, password=, secret=, token= в строках.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    p:=map[string]string{"postgres://":"database credentials","sk-":"API key","password":"database credentials","secret":"secret value"}
    for i:=1;i<=n;i++{sc.Scan();l:=strings.ToLower(sc.Text());found:=false
        for pat,desc:=range p{if strings.Contains(l,pat){fmt.Printf("SECRET FOUND: line %d — %s\n",i,desc);found=true;break}};if !found{fmt.Printf("OK: line %d — no secrets\n",i)}}}</code></pre>`},
		},
	}
}

func lesson_def_devsecops() L {
	return L{
		Slug: "def-devsecops", Title: "DevSecOps", Order: 7,
		Difficulty: "advanced", Track: "security-defense",
		Content: `<h1>DevSecOps — Security in CI/CD</h1><p>Shift-left: security на раннем этапе. SAST (статический анализ), DAST (динамический), dependency scanning, container scanning.</p>`,
		Quiz: []Q{
			{Question: "Shift-left — что значит?", Options: []string{"Перенести security в конец", "Внедрить security проверки как можно раньше в pipeline (при коммите, не при деплое)", "Уволить security team", "Автоматизировать всё"}, Correct: 1, Explanation: "Баг найден при коммите: fix за 5 минут. Баг в проде: часы/дни + инцидент. Раньше нашёл = дешевле починить."},
			{Question: "SAST vs DAST?", Options: []string{"Одно и то же", "SAST: анализ кода без запуска. DAST: тестирование running приложения (как атакующий)", "SAST медленнее", "DAST точнее"}, Correct: 1, Explanation: "SAST (gosec, semgrep): находит patterns в коде (SQL concat, hardcoded secrets). DAST (OWASP ZAP): отправляет payloads в running app. Оба нужны."},
			{Question: "Dependency scanning — зачем?", Options: []string{"Скорость", "Зависимости могут содержать CVE — govulncheck/snyk находят уязвимые версии", "Не нужно", "Только для JavaScript"}, Correct: 1, Explanation: "Log4Shell (CVE-2021-44228): одна зависимость = RCE на миллионах серверов. govulncheck для Go, npm audit для JS — обязательно в CI."},
			{Question: "Container scanning — что проверяет?", Options: []string{"Dockerfile syntax", "Base image (alpine/ubuntu) на наличие CVE в установленных пакетах", "Размер", "Порты"}, Correct: 1, Explanation: "trivy/grype сканируют образ: какие пакеты установлены, есть ли CVE. alpine:3.14 может иметь уязвимый openssl. Обновляй base images."},
			{Question: "Какой security tool первым добавить в CI?", Options: []string{"DAST", "Secret scanning + dependency check — ловит самые частые проблемы с минимальным effort", "Container scan", "Pen test"}, Correct: 1, Explanation: "Secret в git = мгновенная утечка. Уязвимая зависимость = известный attack vector. Оба: 5 минут на настройку в CI, ловят 80% проблем."},
		},
		Tasks: []T{
			{Title: "CI security pipeline", Difficulty: "easy", Description: `<p>Сгенерируй security stages для CI:</p><p>Ввод: <code>go</code></p><p>Вывод:</p><pre><code>Security Pipeline (Go):
1. gosec ./... (SAST)
2. govulncheck ./... (dependency scan)
3. trivy image myapp:latest (container scan)
4. gitleaks detect (secret scan)</code></pre>`, Glossary: []GlossaryItem{{Term: "Security pipeline", Definition: "SAST → dependency scan → container scan → secret scan. В CI при каждом PR."}}, TestCases: []TestCase{{Input: "go", ExpectedOutput: "Security Pipeline (Go):\n1. gosec ./... (SAST)\n2. govulncheck ./... (dependency scan)\n3. trivy image myapp:latest (container scan)\n4. gitleaks detect (secret scan)"}},
				StarterCode: `package main
import "fmt"
func main() {
    var lang string; fmt.Scan(&lang)
    fmt.Printf("Security Pipeline (%s):\n", strings.Title(lang))
    steps := []string{"gosec ./... (SAST)", "govulncheck ./... (dependency scan)", "trivy image myapp:latest (container scan)", "gitleaks detect (secret scan)"}
    for i, s := range steps { fmt.Printf("%d. %s\n", i+1, s) }
}`, Hints: `<p>4 шага: SAST, deps, container, secrets.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var l string;fmt.Scan(&l);fmt.Printf("Security Pipeline (Go):\n");for i,s:=range[]string{"gosec ./... (SAST)","govulncheck ./... (dependency scan)","trivy image myapp:latest (container scan)","gitleaks detect (secret scan)"}{fmt.Printf("%d. %s\n",i+1,s)}}</code></pre>`},
			{Title: "Vulnerability prioritizer", Difficulty: "easy", Description: `<p>Приоритизируй CVE по severity:</p><p>Ввод: <code>3 CVE-2021-44228 critical CVE-2022-1234 medium CVE-2023-5678 low</code></p><p>Вывод:</p><pre><code>FIX NOW: CVE-2021-44228 (critical)
FIX THIS SPRINT: CVE-2022-1234 (medium)
BACKLOG: CVE-2023-5678 (low)</code></pre>`, Glossary: []GlossaryItem{{Term: "CVE prioritization", Definition: "Critical=fix now. High=this week. Medium=this sprint. Low=backlog."}}, TestCases: []TestCase{{Input: "3 CVE-2021-44228 critical CVE-2022-1234 medium CVE-2023-5678 low", ExpectedOutput: "FIX NOW: CVE-2021-44228 (critical)\nFIX THIS SPRINT: CVE-2022-1234 (medium)\nBACKLOG: CVE-2023-5678 (low)"}},
				StarterCode: `package main
import "fmt"
func main() {
    actions := map[string]string{"critical": "FIX NOW", "high": "FIX THIS WEEK", "medium": "FIX THIS SPRINT", "low": "BACKLOG"}
    var n int; fmt.Scan(&n)
    for i := 0; i < n; i++ { var cve, sev string; fmt.Scan(&cve, &sev); fmt.Printf("%s: %s (%s)\n", actions[sev], cve, sev) }
}`, Hints: `<p>Map severity → action. Print action: CVE (severity).</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){a:=map[string]string{"critical":"FIX NOW","high":"FIX THIS WEEK","medium":"FIX THIS SPRINT","low":"BACKLOG"};var n int;fmt.Scan(&n);for i:=0;i<n;i++{var c,s string;fmt.Scan(&c,&s);fmt.Printf("%s: %s (%s)\n",a[s],c,s)}}</code></pre>`},
			{Title: "SAST finding analyzer", Difficulty: "medium", Description: `<p>Классифицируй SAST findings:</p><p>Ввод:</p><pre><code>3
G101 hardcoded_credentials main.go:15
G201 sql_concatenation db.go:42
G104 unhandled_error handler.go:8</code></pre><p>Вывод:</p><pre><code>[CRITICAL] G101: hardcoded_credentials at main.go:15
[HIGH] G201: sql_concatenation at db.go:42
[MEDIUM] G104: unhandled_error at handler.go:8</code></pre>`, Glossary: []GlossaryItem{{Term: "gosec rules", Definition: "G101=credentials, G201=SQL concat, G104=unhandled error. Стандартные Go security rules."}}, TestCases: []TestCase{{Input: "3\nG101 hardcoded_credentials main.go:15\nG201 sql_concatenation db.go:42\nG104 unhandled_error handler.go:8", ExpectedOutput: "[CRITICAL] G101: hardcoded_credentials at main.go:15\n[HIGH] G201: sql_concatenation at db.go:42\n[MEDIUM] G104: unhandled_error at handler.go:8"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    severity := map[string]string{"G101": "CRITICAL", "G201": "HIGH", "G104": "MEDIUM", "G301": "LOW"}
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text())
        rule, finding, location := parts[0], parts[1], parts[2]
        fmt.Printf("[%s] %s: %s at %s\n", severity[rule], rule, finding, location)
    }
}`, Hints: `<p>Map rule_id → severity. Print [SEVERITY] rule: finding at location.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){sev:=map[string]string{"G101":"CRITICAL","G201":"HIGH","G104":"MEDIUM"};var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());fmt.Printf("[%s] %s: %s at %s\n",sev[p[0]],p[0],p[1],p[2])}}</code></pre>`},
			{Title: "Supply chain risk", Difficulty: "hard", Description: `<p>Оцени риск зависимостей:</p><p>Ввод:</p><pre><code>3
github.com/lib/pq v1.10.0 2-years-old 5-CVEs
github.com/go-chi/chi v5.0.10 1-month-old 0-CVEs
github.com/dgrijalva/jwt-go v3.2.0 archived 3-CVEs</code></pre><p>Вывод:</p><pre><code>HIGH RISK: github.com/lib/pq (5 CVEs, outdated 2 years)
LOW RISK: github.com/go-chi/chi (0 CVEs, recent)
CRITICAL: github.com/dgrijalva/jwt-go (archived, 3 CVEs) — REPLACE IMMEDIATELY</code></pre>`, Glossary: []GlossaryItem{{Term: "Supply chain", Definition: "Зависимости = attack surface. Archived/unmaintained + CVEs = supply chain risk."}}, TestCases: []TestCase{{Input: "3\ngithub.com/lib/pq v1.10.0 2-years-old 5-CVEs\ngithub.com/go-chi/chi v5.0.10 1-month-old 0-CVEs\ngithub.com/dgrijalva/jwt-go v3.2.0 archived 3-CVEs", ExpectedOutput: "HIGH RISK: github.com/lib/pq (5 CVEs, outdated 2 years)\nLOW RISK: github.com/go-chi/chi (0 CVEs, recent)\nCRITICAL: github.com/dgrijalva/jwt-go (archived, 3 CVEs) — REPLACE IMMEDIATELY"}},
				StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text())
        pkg, age, cves := parts[0], parts[2], parts[3]
        if strings.Contains(age, "archived") {
            fmt.Printf("CRITICAL: %s (archived, %s) — REPLACE IMMEDIATELY\n", pkg, strings.TrimSuffix(cves, "-CVEs")+" CVEs")
        } else if strings.HasPrefix(cves, "0") {
            fmt.Printf("LOW RISK: %s (0 CVEs, recent)\n", pkg)
        } else {
            cv := strings.TrimSuffix(cves, "-CVEs"); yr := strings.TrimSuffix(age, "-old")
            fmt.Printf("HIGH RISK: %s (%s CVEs, outdated %s)\n", pkg, cv, strings.Replace(yr, "-years", " years", 1))
        }
    }
}`, Hints: `<p>archived = CRITICAL. 0 CVEs = LOW. else HIGH. Parse age and CVE count.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();p:=strings.Fields(sc.Text());pkg,age,cves:=p[0],p[2],p[3]
        if strings.Contains(age,"archived"){fmt.Printf("CRITICAL: %s (archived, %s) — REPLACE IMMEDIATELY\n",pkg,strings.Replace(cves,"-"," ",-1))
        }else if strings.HasPrefix(cves,"0"){fmt.Printf("LOW RISK: %s (0 CVEs, recent)\n",pkg)
        }else{fmt.Printf("HIGH RISK: %s (%s, outdated %s)\n",pkg,strings.Replace(cves,"-"," ",-1),strings.Replace(strings.TrimSuffix(age,"-old"),"-"," ",-1))}}}</code></pre>`},
			{Title: "Security scorecard", Difficulty: "hard", Description: `<p>Итоговая оценка security posture проекта:</p><p>Ввод: <code>sast:pass deps:fail container:pass secrets:pass tests:pass</code></p><p>Вывод:</p><pre><code>SECURITY SCORECARD: 4/5 (80%)
✓ SAST: pass
✗ Dependencies: FAIL — update vulnerable packages
✓ Container: pass
✓ Secrets: pass
✓ Tests: pass
Grade: B (good, fix dependencies)</code></pre>`, Glossary: []GlossaryItem{{Term: "Security scorecard", Definition: "Общая оценка security проекта. A=100%, B=80%+, C=60%+, D=<60%."}}, TestCases: []TestCase{{Input: "sast:pass deps:fail container:pass secrets:pass tests:pass", ExpectedOutput: "SECURITY SCORECARD: 4/5 (80%)\n✓ SAST: pass\n✗ Dependencies: FAIL — update vulnerable packages\n✓ Container: pass\n✓ Secrets: pass\n✓ Tests: pass\nGrade: B (good, fix dependencies)"}},
				StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var input string; fmt.Scanln(&input)
    checks := strings.Fields(input); passed := 0; var fails []string
    names := map[string]string{"sast": "SAST", "deps": "Dependencies", "container": "Container", "secrets": "Secrets", "tests": "Tests"}
    recs := map[string]string{"deps": "update vulnerable packages", "sast": "fix security findings", "container": "update base image", "secrets": "remove hardcoded secrets"}
    fmt.Printf("SECURITY SCORECARD: ")
    for _, c := range checks { parts := strings.Split(c, ":"); if parts[1] == "pass" { passed++ } else { fails = append(fails, parts[0]) } }
    fmt.Printf("%d/%d (%d%%)\n", passed, len(checks), passed*100/len(checks))
    for _, c := range checks { parts := strings.Split(c, ":")
        if parts[1] == "pass" { fmt.Printf("✓ %s: pass\n", names[parts[0]]) } else { fmt.Printf("✗ %s: FAIL — %s\n", names[parts[0]], recs[parts[0]]) }
    }
    pct := passed * 100 / len(checks); var grade string
    switch { case pct == 100: grade = "A (excellent)"; case pct >= 80: grade = "B (good, fix " + strings.Join(fails, ", ") + ")"; case pct >= 60: grade = "C (needs work)"; default: grade = "D (critical issues)" }
    fmt.Printf("Grade: %s\n", grade)
}`, Hints: `<p>Parse key:value. Count pass/fail. Grade by percentage. Show ✓/✗ for each.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();checks:=strings.Fields(sc.Text());p:=0;var f []string
    nm:=map[string]string{"sast":"SAST","deps":"Dependencies","container":"Container","secrets":"Secrets","tests":"Tests"}
    rc:=map[string]string{"deps":"update vulnerable packages","sast":"fix security findings"}
    for _,c:=range checks{parts:=strings.Split(c,":");if parts[1]=="pass"{p++}else{f=append(f,parts[0])}}
    fmt.Printf("SECURITY SCORECARD: %d/%d (%d%%)\n",p,len(checks),p*100/len(checks))
    for _,c:=range checks{parts:=strings.Split(c,":");if parts[1]=="pass"{fmt.Printf("✓ %s: pass\n",nm[parts[0]])}else{fmt.Printf("✗ %s: FAIL — %s\n",nm[parts[0]],rc[parts[0]])}}
    pct:=p*100/len(checks);var g string;switch{case pct==100:g="A (excellent)";case pct>=80:g="B (good, fix "+strings.Join(f,", ")+")";default:g="C (needs work)"};fmt.Printf("Grade: %s\n",g)}</code></pre>`},
		},
	}
}
