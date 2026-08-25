package main

// Fixtures + auto-checks for "Linux: продвинутый" (linux-core).
// Same contract as linux-start: one setup per lesson creates everything the
// lesson's tasks reference, and each task gets a check that inspects the result.

var linuxCoreLabs = map[string]labSpec{
	// ── Lab 1: просмотр и сортировка ──
	"ch-lcore-lab1": {
		Setup: `set -e
rm -rf /opt/devops/lab1; mkdir -p /opt/devops/lab1
for i in $(seq 1 20); do echo "2024-05-01 10:$(printf '%02d' $i):00 INFO  event number $i"; done > /opt/devops/lab1/app.log
printf 'banana\napple\ncherry\napple\ndate\nbanana\n' > /opt/devops/lab1/words.txt
{
  echo '10.0.0.5 - - "GET / HTTP/1.1" 200'
  echo '10.0.0.7 - - "GET /a HTTP/1.1" 200'
  echo '10.0.0.5 - - "GET /b HTTP/1.1" 404'
  echo '10.0.0.9 - - "GET /c HTTP/1.1" 500'
  echo '10.0.0.7 - - "GET /d HTTP/1.1" 200'
} > /opt/devops/lab1/access.log
rm -f /root/first5.txt /root/last10.txt /root/sorted.txt /root/unique_ips.txt`,
		Checks: map[int]string{
			1: check(`head -5 /opt/devops/lab1/app.log | diff -q - /root/first5.txt >/dev/null 2>&1`,
				"первые 5 строк совпадают",
				"Нужны ровно первые 5 строк: head -5 /opt/devops/lab1/app.log > /root/first5.txt"),
			2: check(`tail -10 /opt/devops/lab1/app.log | diff -q - /root/last10.txt >/dev/null 2>&1`,
				"последние 10 строк совпадают",
				"Нужны последние 10 строк: tail -10 /opt/devops/lab1/app.log > /root/last10.txt"),
			3: check(`sort /opt/devops/lab1/words.txt | diff -q - /root/sorted.txt >/dev/null 2>&1`,
				"файл отсортирован",
				"Отсортируй строки: sort /opt/devops/lab1/words.txt > /root/sorted.txt"),
			4: check(`awk '{print $1}' /opt/devops/lab1/access.log | sort -u | diff -q - <(sort -u /root/unique_ips.txt 2>/dev/null) >/dev/null 2>&1`,
				"получены 3 уникальных IP",
				"Возьми первое поле и убери повторы: awk '{print \\$1}' /opt/devops/lab1/access.log | sort -u > /root/unique_ips.txt"),
		},
	},

	// ── Lab 2: grep ──
	"ch-lcore-lab2": {
		Setup: `set -e
rm -rf /opt/devops/lab2; mkdir -p /opt/devops/lab2/configs
{
  echo "2024-05-01 INFO  service started"
  echo "2024-05-01 ERROR cannot open socket"
  echo "2024-05-01 INFO  retry in 5s"
  echo "2024-05-01 CRITICAL data corruption detected"
  echo "2024-05-01 WARN  memory high"
  echo "2024-05-01 ERROR db timeout"
  echo "2024-05-01 INFO  shutdown"
} > /opt/devops/lab2/app.log
{
  echo "ERROR uppercase form"
  echo "Error mixed form"
  echo "error lowercase form"
  echo "INFO not an error line at all"
} > /opt/devops/lab2/mixed.log
printf 'user=admin\npassword=secret123\n' > /opt/devops/lab2/configs/db.conf
printf 'listen 80;\n'                     > /opt/devops/lab2/configs/nginx.conf
printf 'token=abc\npassword=hunter2\n'    > /opt/devops/lab2/configs/app.conf
{
  echo '10.0.0.1 - - "GET / HTTP/1.1" 200'
  echo '10.0.0.2 - - "GET /x HTTP/1.1" 404'
  echo '10.0.0.3 - - "POST /y HTTP/1.1" 500'
  echo '10.0.0.4 - - "GET /z HTTP/1.1" 200'
  echo '10.0.0.5 - - "GET /w HTTP/1.1" 301'
} > /opt/devops/lab2/access.log
rm -f /root/errors.txt /root/all_errors.txt /root/no_info.txt /root/has_password.txt /root/critical_errors.txt /root/status_codes.txt`,
		Checks: map[int]string{
			1: check(`grep ERROR /opt/devops/lab2/app.log | diff -q - /root/errors.txt >/dev/null 2>&1`,
				"строки ERROR выбраны",
				"grep ERROR /opt/devops/lab2/app.log > /root/errors.txt"),
			2: check(`grep -i error /opt/devops/lab2/mixed.log | diff -q - /root/all_errors.txt >/dev/null 2>&1`,
				"найдены все 3 формы написания error",
				"Регистр игнорируется флагом -i: grep -i error /opt/devops/lab2/mixed.log > /root/all_errors.txt"),
			3: check(`grep -v INFO /opt/devops/lab2/app.log | diff -q - /root/no_info.txt >/dev/null 2>&1`,
				"строки с INFO исключены",
				"Инвертируй условие: grep -v INFO /opt/devops/lab2/app.log > /root/no_info.txt"),
			4: check(`grep -q 'db\.conf' /root/has_password.txt 2>/dev/null && grep -q 'app\.conf' /root/has_password.txt && ! grep -q 'nginx\.conf' /root/has_password.txt`,
				"найдены оба конфига с паролем",
				"Нужен список ИМЁН файлов: grep -rl password /opt/devops/lab2/configs/ > /root/has_password.txt"),
			5: check(`grep -E 'ERROR|CRITICAL' /opt/devops/lab2/app.log | diff -q - /root/critical_errors.txt >/dev/null 2>&1`,
				"выбраны ERROR и CRITICAL",
				"Расширенные регулярки — флаг -E: grep -E 'ERROR|CRITICAL' /opt/devops/lab2/app.log > /root/critical_errors.txt"),
			6: check(`grep -oE '(200|301|404|500)' /root/status_codes.txt 2>/dev/null | sort -u | diff -q - <(printf '200\n301\n404\n500\n') >/dev/null 2>&1`,
				"собраны все уникальные коды ответа",
				"Извлеки последнее поле и убери повторы: awk '{print \\$NF}' /opt/devops/lab2/access.log | sort -u > /root/status_codes.txt"),
		},
	},

	// ── Lab 3: sed и awk ──
	"ch-lcore-lab3": {
		Setup: `set -e
rm -rf /opt/devops/lab3; mkdir -p /opt/devops/lab3
printf 'db_host=localhost\ncache_host=localhost\napp_port=8080\n' > /opt/devops/lab3/config.txt
printf 'alpha 10 x\nbeta 20 y\ngamma 30 z\n'                     > /opt/devops/lab3/data.txt
{
  echo '10.0.0.1 - - "GET /index.html HTTP/1.1" 200'
  echo '10.0.0.2 - - "GET /missing HTTP/1.1" 404'
  echo '10.0.0.3 - - "GET /old HTTP/1.1" 404'
  echo '10.0.0.4 - - "POST /api HTTP/1.1" 500'
} > /opt/devops/lab3/access.log
printf 'linux\ndevops\nkernel\n' > /opt/devops/lab3/words.txt
rm -f /root/column2.txt /root/not_found.txt /root/upper.txt`,
		Checks: map[int]string{
			1: check(`! grep -q localhost /opt/devops/lab3/config.txt && [ "$(grep -c db\.internal /opt/devops/lab3/config.txt)" = 2 ]`,
				"localhost заменён на db.internal прямо в файле",
				"Замена на месте — флаг -i: sed -i 's/localhost/db.internal/g' /opt/devops/lab3/config.txt"),
			2: check(`awk '{print $2}' /opt/devops/lab3/data.txt | diff -q - /root/column2.txt >/dev/null 2>&1`,
				"второй столбец извлечён",
				"awk '{print \\$2}' /opt/devops/lab3/data.txt > /root/column2.txt"),
			3: check(`grep 404 /opt/devops/lab3/access.log | diff -q - /root/not_found.txt >/dev/null 2>&1`,
				"строки с кодом 404 выбраны",
				"grep 404 /opt/devops/lab3/access.log > /root/not_found.txt (должно быть 2 строки)"),
			4: check(`tr 'a-z' 'A-Z' < /opt/devops/lab3/words.txt | diff -q - /root/upper.txt >/dev/null 2>&1`,
				"текст переведён в верхний регистр",
				"tr 'a-z' 'A-Z' < /opt/devops/lab3/words.txt > /root/upper.txt"),
		},
	},

	// ── Lab 4: потоки ввода-вывода ──
	"ch-lcore-lab4": {
		Setup: `set -e
rm -rf /opt/devops/lab4; mkdir -p /opt/devops/lab4
printf 'alpha\nbeta\n'   > /opt/devops/lab4/part1.log
printf 'gamma\ndelta\n'  > /opt/devops/lab4/part2.log
printf 'file one\n'      > /opt/devops/lab4/one.txt
printf 'file two\n'      > /opt/devops/lab4/two.txt
cat > /opt/devops/lab4/test.sh <<'SH'
#!/bin/bash
echo "STDOUT_LINE: всё хорошо"
echo "STDERR_LINE: что-то пошло не так" >&2
SH
chmod +x /opt/devops/lab4/test.sh
mkdir -p /etc/demo && printf 'demo=1\n' > /etc/demo/demo.conf
rm -f /root/listing.txt /root/error_msg.txt /root/combined.log /root/confs.txt /root/full_output.txt`,
		Checks: map[int]string{
			1: check(`grep -q 'one\.txt' /root/listing.txt 2>/dev/null && grep -q 'two\.txt' /root/listing.txt`,
				"вывод ls перенаправлен в файл",
				"ls /opt/devops/lab4/ > /root/listing.txt"),
			2: check(`[ -s /root/error_msg.txt ] && grep -qiE 'no such file|нет такого' /root/error_msg.txt`,
				"текст ошибки попал в файл из stderr",
				"Поток ошибок — это 2: ls /nonexistent_dir 2> /root/error_msg.txt"),
			3: check(`cat /opt/devops/lab4/part1.log /opt/devops/lab4/part2.log | diff -q - /root/combined.log >/dev/null 2>&1`,
				"обе части дописаны по порядку",
				"Дописывай через >>: cat /opt/devops/lab4/part1.log >> /root/combined.log; cat /opt/devops/lab4/part2.log >> /root/combined.log"),
			4: check(`[ -s /root/confs.txt ] && grep -q '\.conf' /root/confs.txt && ! grep -qiE 'permission denied|отказано в доступе' /root/confs.txt`,
				"найдены .conf, ошибки доступа отброшены",
				"Отправь stderr в /dev/null: find / -name '*.conf' 2>/dev/null > /root/confs.txt"),
			5: check(`grep -q STDOUT_LINE /root/full_output.txt 2>/dev/null && grep -q STDERR_LINE /root/full_output.txt`,
				"stdout и stderr собраны в один файл",
				"Объедини потоки: /opt/devops/lab4/test.sh > /root/full_output.txt 2>&1"),
		},
	},

	// ── Lab 5: анализ access.log ──
	"ch-lcore-lab5": {
		Setup: `set -e
rm -rf /opt/devops/lab5; mkdir -p /opt/devops/lab5
gen() { # ip url code count
  for i in $(seq 1 $4); do echo "$1 - - [01/May/2024:10:00:$i +0000] \"GET $2 HTTP/1.1\" $3 512"; done
}
{
  gen 192.168.1.1 /index.html 200 12
  gen 192.168.1.2 /about.html 200 9
  gen 192.168.1.3 /missing    404 7
  gen 192.168.1.4 /api/users  200 5
  gen 192.168.1.5 /api/orders 500 3
  gen 192.168.1.6 /contact    200 2
  gen 192.168.1.7 /old-page   404 1
} > /opt/devops/lab5/access.log
rm -f /root/not_found.txt /root/top_ips.txt /root/status_stats.txt /root/top_urls.txt`,
		Checks: map[int]string{
			1: check(`grep ' 404 ' /opt/devops/lab5/access.log | diff -q - /root/not_found.txt >/dev/null 2>&1 || grep 404 /opt/devops/lab5/access.log | diff -q - /root/not_found.txt >/dev/null 2>&1`,
				"строки с 404 выбраны (их 8)",
				"grep 404 /opt/devops/lab5/access.log > /root/not_found.txt"),
			2: check(`awk '{print $1}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -5 | awk '{print $1" "$2}' | diff -q - <(sed 's/^[[:space:]]*//; s/[[:space:]]\+/ /g' /root/top_ips.txt 2>/dev/null) >/dev/null 2>&1`,
				"топ-5 IP в правильном порядке и формате",
				"awk '{print \\$1}' access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -5 — формат строки: <количество> <IP>"),
			3: check(`awk '{print $(NF-1)}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 | awk '{print $1" "$2}' | diff -q - <(sed 's/^[[:space:]]*//; s/[[:space:]]\+/ /g' /root/status_stats.txt 2>/dev/null) >/dev/null 2>&1`,
				"статистика по кодам ответа верна",
				"Код ответа — предпоследнее поле: awk '{print \\$(NF-1)}' access.log | sort | uniq -c | sort -k1,1nr -k2,2 — формат: <количество> <код>"),
			4: check(`awk '{print $7}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -10 | awk '{print $1" "$2}' | diff -q - <(sed 's/^[[:space:]]*//; s/[[:space:]]\+/ /g' /root/top_urls.txt 2>/dev/null) >/dev/null 2>&1`,
				"топ URL в правильном порядке и формате",
				"URL — седьмое поле: awk '{print \\$7}' access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -10 — формат: <количество> <URL>"),
		},
	},

	// ── Lab 6: права доступа ──
	"ch-lcore-lab6": {
		Setup: `set -e
rm -rf /opt/devops/lab6; mkdir -p /opt/devops/lab6/public /opt/devops/lab6/configs
printf '#!/bin/bash\necho deploying\n' > /opt/devops/lab6/deploy.sh
printf 'API_KEY=supersecret\n'         > /opt/devops/lab6/secrets.env
printf 'app=1\n'                       > /opt/devops/lab6/configs/app.conf
printf 'PRIVATE KEY DATA\n'            > /opt/devops/lab6/id_rsa
printf 'index\n'                       > /opt/devops/lab6/public/index.html
chmod 644 /opt/devops/lab6/deploy.sh /opt/devops/lab6/secrets.env /opt/devops/lab6/id_rsa
chmod 700 /opt/devops/lab6/public
chmod 600 /opt/devops/lab6/configs/app.conf`,
		Checks: map[int]string{
			1: check(`[ -x /opt/devops/lab6/deploy.sh ]`,
				"deploy.sh стал исполняемым",
				"Добавь право выполнения: chmod +x /opt/devops/lab6/deploy.sh"),
			2: check(`[ "$(stat -c '%a' /opt/devops/lab6/secrets.env)" = 600 ]`,
				"права на secrets.env — 600",
				"chmod 600 /opt/devops/lab6/secrets.env — сейчас $(stat -c '%a' /opt/devops/lab6/secrets.env)"),
			3: check(`[ "$(stat -c '%a' /opt/devops/lab6/public)" = 755 ]`,
				"права на каталог public — 755",
				"chmod 755 /opt/devops/lab6/public — сейчас $(stat -c '%a' /opt/devops/lab6/public)"),
			4: check(`[ "$(stat -c '%a' /opt/devops/lab6/configs/app.conf)" = 644 ]`,
				"права выставлены рекурсивно (644 на app.conf)",
				"chmod -R 644 /opt/devops/lab6/configs/ — сейчас у app.conf $(stat -c '%a' /opt/devops/lab6/configs/app.conf)"),
			5: check(`[ "$(stat -c '%a' /opt/devops/lab6/id_rsa)" = 600 ]`,
				"права на приватный ключ — 600",
				"chmod 600 /opt/devops/lab6/id_rsa — сейчас $(stat -c '%a' /opt/devops/lab6/id_rsa)"),
		},
	},

	// ── Lab 7: пользователи и группы ──
	"ch-lcore-lab7": {
		Setup: `set -e
id devuser >/dev/null 2>&1 || useradd -m devuser
gpasswd -d devuser sudo >/dev/null 2>&1 || true
userdel -r testuser >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`id testuser >/dev/null 2>&1 && [ -d "$(getent passwd testuser | cut -d: -f6)" ]`,
				"пользователь testuser создан вместе с домашней директорией",
				"Флаг -m создаёт домашний каталог: useradd -m testuser"),
			2: check(`id -nG devuser 2>/dev/null | tr ' ' '\n' | grep -qx sudo`,
				"devuser добавлен в группу sudo",
				"Добавь в группу, не затирая остальные: usermod -aG sudo devuser"),
		},
	},

	// ── Lab 8: процессы ──
	"ch-lcore-lab8": {
		Setup: `set -e
mkdir -p /opt/devops/process-demo
printf '<html><body><h1>Process demo</h1></body></html>\n' > /opt/devops/process-demo/index.html
pkill -x python3 >/dev/null 2>&1 || true
rm -f /root/demo.pid
(setsid sleep 9999 >/dev/null 2>&1 &)
sleep 0.2`,
		Checks: map[int]string{
			1: check(`[ -s /root/demo.pid ] && kill -0 "$(cat /root/demo.pid)" 2>/dev/null && (ss -ltn 2>/dev/null | grep -q ':80 ' || curl -s -o /dev/null http://localhost/)`,
				"сервер запущен в фоне, PID сохранён, порт 80 слушается",
				"Запусти в фоне и сохрани PID: cd /opt/devops/process-demo && python3 -m http.server 80 & echo \\$! > /root/demo.pid"),
			2: check(`! ss -ltn 2>/dev/null | grep -q ':80 ' && ! pgrep -x python3 >/dev/null`,
				"demo-сервер остановлен, порт 80 свободен",
				"Останови процесс: kill $(cat /root/demo.pid) — затем проверь ss -ltnp | grep :80"),
			3: check(`[ -z "$(pgrep -f 'sleep 99[9]9')" ]`,
				"фоновый процесс sleep завершён",
				"Найди фоновый процесс: pgrep -af sleep — затем заверши его: kill <PID>"),
		},
	},

	// ── Lab 9: сеть и curl ──
	"ch-lcore-lab9": {
		Setup: `set -e
mkdir -p /opt/devops/netlab
printf '<html><body><h1>Linux network tools lab</h1></body></html>\n' > /opt/devops/netlab/index.html
pkill -x python3 >/dev/null 2>&1 || true
(cd /opt/devops/netlab && setsid python3 -m http.server 80 >/var/log/netlab.log 2>&1 &)
sleep 0.5
rm -f /root/http_page.html /root/http_status.txt`,
		Checks: map[int]string{
			1: check(`grep -q 'Linux network tools lab' /root/http_page.html 2>/dev/null && grep -q '200' /root/http_status.txt 2>/dev/null`,
				"HTML-ответ и код 200 сохранены",
				"curl http://localhost/ > /root/http_page.html и curl -s -o /dev/null -w '%{http_code}' http://localhost/ > /root/http_status.txt"),
		},
	},

	// ── Lab 10: systemd ──
	"ch-lcore-lab10": {
		Setup: `set -e
mkdir -p /opt/devops/process-demo
printf '<html><body><h1>Process demo</h1></body></html>\n' > /opt/devops/process-demo/index.html
systemctl stop demo-http >/dev/null 2>&1 || true
systemctl disable demo-http >/dev/null 2>&1 || true
rm -f /etc/systemd/system/demo-http.service
pkill -x python3 >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`[ -f /etc/systemd/system/demo-http.service ] && grep -qE '^\s*ExecStart=.*python3 -m http\.server 80' /etc/systemd/system/demo-http.service`,
				"unit-файл создан и содержит правильный ExecStart",
				"Создай /etc/systemd/system/demo-http.service с ExecStart=/usr/bin/python3 -m http.server 80 и выполни systemctl daemon-reload"),
			2: check(`systemctl is-active demo-http >/dev/null 2>&1`,
				"сервис demo-http запущен",
				"Запусти сервис: systemctl start demo-http — состояние смотри через systemctl status demo-http"),
			3: check(`systemctl is-enabled demo-http >/dev/null 2>&1`,
				"автозапуск demo-http включён",
				"Включи автозапуск: systemctl enable demo-http"),
		},
	},

	// ── Lab 12: редактирование файлов ──
	"ch-lcore-lab12": {
		Setup: `set -e
rm -rf /opt/devops/lab12; mkdir -p /opt/devops/lab12
printf 'host=127.0.0.1\nport=9999\ndebug=false\n' > /opt/devops/lab12/broken.conf
printf 'первая строка\nвторая строка\n'           > /opt/devops/lab12/notes.txt
rm -f /root/mynote.txt`,
		Checks: map[int]string{
			1: check(`grep -q 'Hello from editor' /root/mynote.txt 2>/dev/null`,
				"файл создан и содержит нужную строку",
				"Открой nano /root/mynote.txt, впиши Hello from editor и сохрани (Ctrl+O, Enter, Ctrl+X)"),
			2: check(`grep -q '^port=8080$' /opt/devops/lab12/broken.conf && ! grep -q '9999' /opt/devops/lab12/broken.conf`,
				"порт исправлен на месте",
				"Без редактора это делает sed: sed -i 's/port=9999/port=8080/' /opt/devops/lab12/broken.conf"),
			3: check(`[ "$(head -1 /opt/devops/lab12/notes.txt)" = '# Edited by student' ]`,
				"строка добавлена в начало файла",
				"Добавь строку первой: sed -i '1i # Edited by student' /opt/devops/lab12/notes.txt"),
		},
	},

	// ── Lab 13: cron ──
	"ch-lcore-lab13": {
		Setup: `set -e
rm -rf /opt/devops/lab13; mkdir -p /opt/devops/lab13
printf '#!/bin/bash\necho backup done\n' > /opt/devops/lab13/backup.sh
chmod +x /opt/devops/lab13/backup.sh
crontab -r >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`crontab -l 2>/dev/null | grep -qE '^\s*0\s+2\s+\*\s+\*\s+\*\s+.*backup\.sh'`,
				"задача на 02:00 добавлена в crontab",
				"Добавь строку в crontab (crontab -e): 0 2 * * * /opt/devops/lab13/backup.sh — проверь через crontab -l"),
		},
	},

	// ── Lab 14: sed/awk на конфигах ──
	"ch-lcore-lab14": {
		Setup: `set -e
rm -rf /opt/devops/lab14; mkdir -p /opt/devops/lab14
printf 'db_host=localhost\ncache=localhost\nport=8080\n' > /opt/devops/lab14/app.conf
{
  echo "# главный конфиг"
  echo "server {"
  echo "    listen 80;"
  echo "# закомментированная строка"
  echo "    root /var/www;"
  echo "}"
} > /opt/devops/lab14/nginx.conf
printf 'id,username,email\n1,alice,alice@example.com\n2,bob,bob@example.com\n3,carol,carol@example.com\n' > /opt/devops/lab14/users.csv
{
  echo '10.0.0.1 - - "GET / HTTP/1.1" 200'
  echo '10.0.0.2 - - "POST /api HTTP/1.1" 500'
  echo '10.0.0.3 - - "GET /x HTTP/1.1" 404'
  echo '10.0.0.4 - - "POST /pay HTTP/1.1" 500'
} > /opt/devops/lab14/access.log
rm -f /root/usernames.txt /root/server_errors.txt`,
		Checks: map[int]string{
			1: check(`! grep -q localhost /opt/devops/lab14/app.conf && [ "$(grep -c db\.internal /opt/devops/lab14/app.conf)" = 2 ]`,
				"хост заменён на месте",
				"sed -i 's/localhost/db.internal/g' /opt/devops/lab14/app.conf"),
			2: check(`! grep -q '^#' /opt/devops/lab14/nginx.conf && grep -q 'listen 80;' /opt/devops/lab14/nginx.conf`,
				"комментарии удалены, конфиг цел",
				"Удали строки, начинающиеся с #: sed -i '/^#/d' /opt/devops/lab14/nginx.conf"),
			3: check(`grep -q '^alice$' /root/usernames.txt 2>/dev/null && grep -q '^carol$' /root/usernames.txt && ! grep -q '@' /root/usernames.txt`,
				"имена пользователей извлечены",
				"Второе поле с разделителем-запятой: awk -F, 'NR>1{print \\$2}' /opt/devops/lab14/users.csv > /root/usernames.txt"),
			4: check(`grep 500 /opt/devops/lab14/access.log | diff -q - /root/server_errors.txt >/dev/null 2>&1`,
				"строки с кодом 500 выбраны",
				"grep 500 /opt/devops/lab14/access.log > /root/server_errors.txt (должно быть 2 строки)"),
		},
	},

	// ── Lab 15: дисковое пространство ──
	"ch-lcore-lab15": {
		Setup: `set -e
mkdir -p /var/lib/gldemo /var/log/gldemo /var/cache/gldemo
head -c 300000 /dev/zero | tr '\0' 'a' > /var/lib/gldemo/big1
head -c 200000 /dev/zero | tr '\0' 'b' > /var/log/gldemo/big2
head -c 100000 /dev/zero | tr '\0' 'c' > /var/cache/gldemo/big3
rm -f /root/top_dirs.txt`,
		Checks: map[int]string{
			1: check(`[ "$(grep -c . /root/top_dirs.txt 2>/dev/null)" = 5 ] && grep -q '/var/' /root/top_dirs.txt && grep -qE '[0-9]+([.,][0-9]+)?[KMG]' /root/top_dirs.txt`,
				"5 крупнейших каталогов /var с human-readable размерами",
				"du -h --max-depth=1 /var | sort -hr | head -5 > /root/top_dirs.txt (нужно ровно 5 строк с размерами вида 12M)"),
		},
	},

	// ── Lab 16:操作 операторы и PATH ──
	"ch-lcore-lab16": {
		Setup: `set -e
rm -rf /root/app_data /opt/devops/lab16; mkdir -p /opt/devops/lab16/bin
printf '#!/bin/bash\necho hello from lab16\n' > /opt/devops/lab16/bin/lab16-hello
chmod +x /opt/devops/lab16/bin/lab16-hello
rm -f /root/new_path.txt`,
		Checks: map[int]string{
			1: check(`[ -d /root/app_data ] && [ -s /root/app_data/config.txt ]`,
				"каталог и непустой файл созданы",
				"Одной строкой через &&: mkdir /root/app_data && echo 'setting=1' > /root/app_data/config.txt"),
			2: check(`grep -q '/opt/devops/lab16/bin' /root/new_path.txt 2>/dev/null && grep -q ':' /root/new_path.txt`,
				"каталог добавлен в PATH и значение сохранено",
				"export PATH=\"\\$PATH:/opt/devops/lab16/bin\" затем echo \\$PATH > /root/new_path.txt"),
		},
	},
}
