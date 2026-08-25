package main

// Fixtures + auto-checks for the "Linux: основные инструменты" course.
// Every lesson creates the /opt/devops/labN tree its tasks refer to, then each
// task is validated by inspecting the result the task asked for.

const lab1Config = "server_port=8080\\nlog_level=info\\nmax_connections=100\\n"

var linuxStartLabs = map[string]labSpec{
	// ── Lab 1: навигация по файловой системе ──
	"ch-lnav-lab1": {
		Setup: `set -e
mkdir -p /opt/devops/lab1/notes
printf '` + lab1Config + `' > /opt/devops/lab1/config.txt
printf 'TODO: проверить бэкапы\n' > /opt/devops/lab1/notes/todo.md
printf 'README проекта lab1\n'     > /opt/devops/lab1/README.md
printf 'DEVOPS_SECRET_2024\n'      > /opt/devops/lab1/.secret
rm -f /root/location.txt /root/lab1_listing.txt /root/notes_path.txt /root/found_secret.txt`,
		Checks: map[int]string{
			1: check(`[ -s /root/location.txt ] && grep -qE '^/' /root/location.txt`,
				"location.txt содержит текущий путь",
				"Нужен файл /root/location.txt с текущим путём: pwd > /root/location.txt"),
			2: check(`grep -q 'config\.txt' /root/lab1_listing.txt 2>/dev/null && grep -q '\.secret' /root/lab1_listing.txt && grep -q 'notes' /root/lab1_listing.txt`,
				"листинг содержит и обычные, и скрытые файлы",
				"В /root/lab1_listing.txt должен быть полный листинг со скрытыми файлами: ls -la /opt/devops/lab1 > /root/lab1_listing.txt"),
			3: check(`[ "$(tr -d ' \n' < /root/notes_path.txt 2>/dev/null)" = /opt/devops/lab1/notes ]`,
				"путь каталога notes сохранён верно",
				"Перейди в /opt/devops/lab1/notes и запиши pwd в /root/notes_path.txt. Сейчас там: $(cat /root/notes_path.txt 2>/dev/null)"),
			4: check(`grep -q DEVOPS_SECRET_2024 /root/found_secret.txt 2>/dev/null`,
				"содержимое скрытого файла найдено",
				"Скрытый файл начинается с точки — найди его через ls -a и скопируй содержимое в /root/found_secret.txt"),
		},
	},

	// ── Lab 2: создание, копирование, удаление ──
	"ch-lnav-lab2": {
		Setup: `set -e
mkdir -p /opt/devops/lab1
printf '` + lab1Config + `' > /opt/devops/lab1/config.txt
rm -rf /root/myproject /root/backup`,
		Checks: map[int]string{
			1: check(`[ -d /root/myproject ]`,
				"каталог /root/myproject создан",
				"Создай каталог: mkdir /root/myproject"),
			2: check(`grep -q 'Hello DevOps' /root/myproject/README.md 2>/dev/null`,
				"README.md содержит нужную строку",
				"Нужен /root/myproject/README.md со строкой Hello DevOps: echo 'Hello DevOps' > /root/myproject/README.md"),
			3: check(`[ -f /root/myproject/config.txt ] && grep -q server_port /root/myproject/config.txt`,
				"config.txt скопирован в проект",
				"Скопируй файл: cp /opt/devops/lab1/config.txt /root/myproject/"),
			4: check(`[ -f /root/myproject/myconfig.txt ] && [ ! -e /root/myproject/config.txt ]`,
				"файл переименован в myconfig.txt",
				"Переименуй именно этот файл (mv), а не копируй: mv /root/myproject/config.txt /root/myproject/myconfig.txt"),
			5: check(`[ -f /root/backup/configs/myconfig.txt ]`,
				"вложенная структура создана и файл скопирован",
				"Создай всю структуру сразу и скопируй файл: mkdir -p /root/backup/configs && cp /root/myproject/myconfig.txt /root/backup/configs/"),
			6: check(`[ ! -e /root/myproject/README.md ]`,
				"README.md удалён",
				"Удали файл: rm /root/myproject/README.md"),
			7: check(`[ ! -e /root/myproject ]`,
				"каталог myproject удалён вместе с содержимым",
				"Удали каталог рекурсивно: rm -r /root/myproject"),
		},
	},

	// ── Lab 3: символические ссылки ──
	"ch-lnav-lab3": {
		Setup: `set -e
mkdir -p /opt/devops/lab1
printf '` + lab1Config + `' > /opt/devops/lab1/config.txt
rm -f /root/config_link /root/broken_link /root/link_content.txt /root/all_links.txt`,
		Checks: map[int]string{
			1: check(`[ -L /root/config_link ] && [ "$(readlink /root/config_link)" = /opt/devops/lab1/config.txt ]`,
				"симлинк config_link указывает на config.txt",
				"Нужна символическая ссылка: ln -s /opt/devops/lab1/config.txt /root/config_link"),
			2: check(`[ -s /root/link_content.txt ] && diff -q /root/link_content.txt /opt/devops/lab1/config.txt >/dev/null`,
				"содержимое по ссылке прочитано верно",
				"Прочитай файл через ссылку и сохрани: cat /root/config_link > /root/link_content.txt"),
			3: check(`[ -L /root/broken_link ] && [ ! -e /root/broken_link ]`,
				"битый симлинк создан",
				"Создай ссылку на несуществующий путь: ln -s /opt/devops/lab1/nonexistent.txt /root/broken_link"),
			4: check(`grep -q config_link /root/all_links.txt 2>/dev/null && grep -q broken_link /root/all_links.txt`,
				"обе ссылки попали в список",
				"Найди симлинки и сохрани список: find /root -type l > /root/all_links.txt"),
		},
	},

	// ── Lab 4: поиск через find ──
	"ch-lnav-lab4": {
		Setup: `set -e
rm -rf /opt/devops/lab4
mkdir -p /opt/devops/lab4/logs /opt/devops/lab4/archive
printf 'INFO boot\nERROR disk full\nINFO ready\n' > /opt/devops/lab4/app.log
printf 'INFO cron started\n'                     > /opt/devops/lab4/logs/cron.log
printf 'alpha\nbeta\ngamma\n'                    > /opt/devops/lab4/notes.txt
printf 'one\ntwo\n'                              > /opt/devops/lab4/data.txt
head -c 4096 /dev/zero | tr '\0' 'x'             > /opt/devops/lab4/big.bin
printf 'скрытая заметка\n'                       > /opt/devops/lab4/.hidden_note
ln -sf /opt/devops/lab4/notes.txt /opt/devops/lab4/link_notes
rm -f /root/logs.txt /root/files_only.txt /root/dirs_only.txt /root/large_files.txt /root/line_counts.txt /root/hidden_files.txt`,
		Checks: map[int]string{
			1: check(`grep -q 'app\.log' /root/logs.txt 2>/dev/null && grep -q 'cron\.log' /root/logs.txt`,
				"найдены оба .log файла (включая вложенный)",
				"Нужны все .log, в том числе в подкаталогах: find /opt/devops/lab4 -name '*.log' > /root/logs.txt"),
			2: check(`grep -q 'big\.bin' /root/files_only.txt 2>/dev/null && grep -q 'notes\.txt' /root/files_only.txt && ! grep -q 'link_notes' /root/files_only.txt`,
				"в списке только обычные файлы, без симлинков",
				"Отбери именно обычные файлы: find /opt/devops/lab4 -type f > /root/files_only.txt (симлинк link_notes попасть не должен)"),
			3: check(`grep -q 'logs' /root/dirs_only.txt 2>/dev/null && grep -q 'archive' /root/dirs_only.txt`,
				"каталоги logs и archive найдены",
				"Ищи каталоги: find /opt/devops/lab4 -type d > /root/dirs_only.txt"),
			4: check(`grep -q 'big\.bin' /root/large_files.txt 2>/dev/null && ! grep -q 'data\.txt' /root/large_files.txt`,
				"найден только файл больше 1K",
				"Фильтруй по размеру: find /opt/devops/lab4 -type f -size +1k > /root/large_files.txt"),
			5: check(`grep -q 'notes\.txt' /root/line_counts.txt 2>/dev/null && grep -q 'data\.txt' /root/line_counts.txt && grep -qE '[0-9]' /root/line_counts.txt`,
				"посчитаны строки в .txt файлах",
				"Посчитай строки найденных файлов: find /opt/devops/lab4 -name '*.txt' -exec wc -l {} + > /root/line_counts.txt"),
			6: check(`grep -q 'hidden_note' /root/hidden_files.txt 2>/dev/null`,
				"скрытый файл найден",
				"Скрытые файлы начинаются с точки: find /opt/devops/lab4 -name '.*' -type f > /root/hidden_files.txt"),
		},
	},

	// ── Lab 5: чтение файлов ──
	"ch-lnav-lab5": {
		Setup: `set -e
rm -rf /opt/devops/lab5; mkdir -p /opt/devops/lab5
{
  echo "2024-05-01 10:00:01 INFO  server started on :8080"
  echo "2024-05-01 10:00:02 INFO  config loaded"
  echo "2024-05-01 10:00:03 INFO  db connection ok"
  echo "2024-05-01 10:01:10 WARN  slow query 1.2s"
  echo "2024-05-01 10:02:00 INFO  request GET /health"
  echo "2024-05-01 10:02:31 ERROR db connection lost"
  echo "2024-05-01 10:02:32 INFO  reconnecting"
  echo "2024-05-01 10:02:35 INFO  db connection ok"
  echo "2024-05-01 10:05:00 INFO  request GET /api/users"
  echo "2024-05-01 10:06:12 WARN  cache miss rate 40%"
  echo "2024-05-01 10:07:45 ERROR timeout talking to payments"
  echo "2024-05-01 10:07:50 INFO  retry scheduled"
  echo "2024-05-01 10:09:00 INFO  request POST /api/orders"
  echo "2024-05-01 10:10:20 ERROR disk usage 95%"
  echo "2024-05-01 10:11:00 INFO  cleanup started"
  echo "2024-05-01 10:11:30 INFO  cleanup finished"
  echo "2024-05-01 10:12:00 INFO  request GET /metrics"
  echo "2024-05-01 10:13:00 INFO  heartbeat ok"
} > /opt/devops/lab5/server.log
printf 'первая часть, строка 1\nпервая часть, строка 2\n' > /opt/devops/lab5/part1.txt
printf 'вторая часть, строка 1\nвторая часть, строка 2\n' > /opt/devops/lab5/part2.txt
rm -f /root/first5.txt /root/last10.txt /root/combined.txt /root/errors.txt`,
		Checks: map[int]string{
			1: check(`head -5 /opt/devops/lab5/server.log | diff -q - /root/first5.txt >/dev/null 2>&1`,
				"первые 5 строк совпадают",
				"Нужны ровно первые 5 строк: head -5 /opt/devops/lab5/server.log > /root/first5.txt"),
			2: check(`tail -10 /opt/devops/lab5/server.log | diff -q - /root/last10.txt >/dev/null 2>&1`,
				"последние 10 строк совпадают",
				"Нужны последние 10 строк: tail -10 /opt/devops/lab5/server.log > /root/last10.txt"),
			3: check(`cat /opt/devops/lab5/part1.txt /opt/devops/lab5/part2.txt | diff -q - /root/combined.txt >/dev/null 2>&1`,
				"файлы объединены в правильном порядке",
				"Склей файлы по порядку: cat /opt/devops/lab5/part1.txt /opt/devops/lab5/part2.txt > /root/combined.txt"),
			4: check(`grep ERROR /opt/devops/lab5/server.log | diff -q - /root/errors.txt >/dev/null 2>&1`,
				"все строки ERROR выбраны",
				"Отфильтруй строки: grep ERROR /opt/devops/lab5/server.log > /root/errors.txt (их должно быть 3)"),
		},
	},

	// ── Lab 6: grep, sort, uniq ──
	"ch-lnav-lab6": {
		Setup: `set -e
rm -rf /opt/devops/lab6; mkdir -p /opt/devops/lab6
{
  echo '10.0.0.1 - - [01/May/2024:10:00:01] "GET /index.html HTTP/1.1" 200 512'
  echo '10.0.0.2 - - [01/May/2024:10:00:05] "GET /missing HTTP/1.1" 404 120'
  echo '10.0.0.1 - - [01/May/2024:10:00:09] "GET /style.css HTTP/1.1" 200 880'
  echo '10.0.0.3 - - [01/May/2024:10:00:12] "POST /api/login HTTP/1.1" 500 64'
  echo '10.0.0.1 - - [01/May/2024:10:00:15] "GET /app.js HTTP/1.1" 200 2048'
  echo '10.0.0.2 - - [01/May/2024:10:00:18] "GET /old-page HTTP/1.1" 404 120'
  echo '10.0.0.1 - - [01/May/2024:10:00:21] "GET /api/users HTTP/1.1" 200 300'
  echo '10.0.0.4 - - [01/May/2024:10:00:25] "GET /favicon.ico HTTP/1.1" 404 0'
} > /opt/devops/lab6/access.log
printf '10.0.0.1\n10.0.0.3\n10.0.0.1\n10.0.0.2\n10.0.0.3\n10.0.0.1\n10.0.0.4\n' > /opt/devops/lab6/ips.txt
rm -f /root/not_found.txt /root/sorted_ips.txt /root/unique_ips.txt /root/ip_stats.txt`,
		Checks: map[int]string{
			1: check(`grep ' 404 ' /opt/devops/lab6/access.log | diff -q - /root/not_found.txt >/dev/null 2>&1 || grep 404 /opt/devops/lab6/access.log | diff -q - /root/not_found.txt >/dev/null 2>&1`,
				"выбраны все строки с кодом 404",
				"Отфильтруй по коду: grep 404 /opt/devops/lab6/access.log > /root/not_found.txt (должно быть 3 строки)"),
			2: check(`sort /opt/devops/lab6/ips.txt | diff -q - /root/sorted_ips.txt >/dev/null 2>&1`,
				"строки отсортированы",
				"Отсортируй все строки, повторы сохраняются: sort /opt/devops/lab6/ips.txt > /root/sorted_ips.txt"),
			3: check(`sort -u /opt/devops/lab6/ips.txt | diff -q - <(sort -u /root/unique_ips.txt 2>/dev/null) >/dev/null 2>&1`,
				"остались только уникальные адреса",
				"Убери повторы: sort -u /opt/devops/lab6/ips.txt > /root/unique_ips.txt (должно остаться 4 адреса)"),
			4: check(`head -1 /root/ip_stats.txt 2>/dev/null | grep -q '10\.0\.0\.1' && grep -qE '(^|[^0-9])4([^0-9]|$)' <(head -1 /root/ip_stats.txt)`,
				"статистика отсортирована по убыванию, сверху самый частый IP",
				"Посчитай и отсортируй по убыванию: awk '{print \\$1}' /opt/devops/lab6/access.log | sort | uniq -c | sort -rn > /root/ip_stats.txt (сверху 10.0.0.1 с 4 запросами)"),
		},
	},

	// ── Lab 7: wildcards ──
	"ch-lnav-lab7": {
		Setup: `set -e
rm -rf /opt/devops/lab7 /root/configs /root/old_configs
mkdir -p /opt/devops/lab7/sub
printf 'INFO app\n'    > /opt/devops/lab7/app.log
printf 'INFO system\n' > /opt/devops/lab7/sys.log
printf 'listen 80;\n'  > /opt/devops/lab7/nginx.conf
printf 'port 6379\n'   > /opt/devops/lab7/redis.conf
printf 'data one\n'    > /opt/devops/lab7/data1.txt
printf 'data two\n'    > /opt/devops/lab7/data2.txt
printf 'data three\n'  > /opt/devops/lab7/data3.txt
printf 'other stuff\n' > /opt/devops/lab7/other.txt
printf 'nested\n'      > /opt/devops/lab7/sub/nested.txt
rm -f /root/log_list.txt /root/data_files.txt /root/txt_files.txt`,
		Checks: map[int]string{
			1: check(`grep -q 'app\.log' /root/log_list.txt 2>/dev/null && grep -q 'sys\.log' /root/log_list.txt && ! grep -q '\.conf' /root/log_list.txt`,
				"в списке только .log файлы",
				"Используй шаблон: ls /opt/devops/lab7/*.log > /root/log_list.txt"),
			2: check(`[ -f /root/configs/nginx.conf ] && [ -f /root/configs/redis.conf ]`,
				"оба .conf скопированы в /root/configs",
				"Сначала создай каталог, потом копируй по шаблону: mkdir /root/configs && cp /opt/devops/lab7/*.conf /root/configs/"),
			3: check(`grep -q 'data1\.txt' /root/data_files.txt 2>/dev/null && grep -q 'data3\.txt' /root/data_files.txt && ! grep -q 'other\.txt' /root/data_files.txt`,
				"шаблон с ? выбрал только data1-3",
				"Один символ заменяется на ?: ls /opt/devops/lab7/data?.txt > /root/data_files.txt (other.txt попасть не должен)"),
			4: check(`grep -q 'nested\.txt' /root/txt_files.txt 2>/dev/null && grep -q 'data1\.txt' /root/txt_files.txt`,
				"найдены .txt и в подкаталогах",
				"Шаблон * не заходит в подкаталоги — нужен find: find /opt/devops/lab7 -name '*.txt' > /root/txt_files.txt"),
			5: check(`[ -f /root/old_configs/nginx.conf ] && [ -f /root/old_configs/redis.conf ] && [ -z "$(ls /root/configs/*.conf 2>/dev/null)" ]`,
				"конфиги перемещены, в /root/configs их больше нет",
				"Перемести (mv), а не копируй: mkdir /root/old_configs && mv /root/configs/*.conf /root/old_configs/"),
		},
	},

	// ── Lab 8: архивы ──
	"ch-lnav-lab8": {
		Setup: `set -e
rm -rf /opt/devops/lab8 /root/extracted
mkdir -p /opt/devops/lab8/src
printf 'print("hello")\n'  > /opt/devops/lab8/src/main.py
printf 'helper code\n'     > /opt/devops/lab8/src/utils.py
printf '# Project\n'       > /opt/devops/lab8/src/README.md
head -c 8192 /dev/zero | tr '\0' 'y' > /opt/devops/lab8/large_file.txt
rm -f /root/backup.tar.gz /root/archive_list.txt /root/large_file.txt /root/large_file.txt.gz`,
		Checks: map[int]string{
			1: check(`[ -f /root/backup.tar.gz ] && tar -tzf /root/backup.tar.gz >/dev/null 2>&1 && tar -tzf /root/backup.tar.gz | grep -q 'main\.py'`,
				"архив создан и содержит файлы из src",
				"Создай сжатый архив: tar -czf /root/backup.tar.gz -C /opt/devops/lab8 src"),
			2: check(`grep -q 'main\.py' /root/archive_list.txt 2>/dev/null && grep -q 'utils\.py' /root/archive_list.txt`,
				"список содержимого архива сохранён",
				"Посмотри архив без распаковки: tar -tzf /root/backup.tar.gz > /root/archive_list.txt"),
			3: check(`[ -d /root/extracted ] && [ -n "$(find /root/extracted -name 'main.py' -print -quit)" ]`,
				"архив распакован в /root/extracted",
				"Создай каталог и распакуй в него: mkdir /root/extracted && tar -xzf /root/backup.tar.gz -C /root/extracted"),
			4: check(`[ -f /root/large_file.txt.gz ] && [ ! -e /root/large_file.txt ] && gzip -t /root/large_file.txt.gz 2>/dev/null`,
				"копия сжата, оригинал заменён архивом",
				"Скопируй и сожми: cp /opt/devops/lab8/large_file.txt /root/ && gzip /root/large_file.txt"),
			5: check(`[ -f /root/large_file.txt ] && [ ! -e /root/large_file.txt.gz ]`,
				"файл распакован обратно",
				"Распакуй: gunzip /root/large_file.txt.gz"),
		},
	},
}
