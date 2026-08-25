package main

// Fixtures + auto-checks for "Linux: продвинутая эксплуатация" (linux-advanced).
// Script-writing tasks are checked by RUNNING the student's script and looking
// at its output — that verifies the script actually works, not merely that a
// file with the right name exists.

var linuxAdvancedLabs = map[string]labSpec{
	// ── Lab 1: curl ──
	"ch-ladv-lab1": {
		Setup: `set -e
mkdir -p /opt/devops/lab1/www
printf '<html><body><h1>Hello from lab1</h1></body></html>\n' > /opt/devops/lab1/www/index.html
pkill -x python3 >/dev/null 2>&1 || true
(cd /opt/devops/lab1/www && setsid python3 -m http.server 80 >/var/log/lab1-http.log 2>&1 &)
sleep 0.5
rm -f /root/http_code.txt /root/page.html`,
		Checks: map[int]string{
			1: check(`[ "$(tr -d ' \n' < /root/http_code.txt 2>/dev/null)" = 200 ]`,
				"код ответа 200 сохранён",
				"Сохрани только код: curl -s -o /dev/null -w '%{http_code}' http://localhost/ > /root/http_code.txt"),
			2: check(`grep -q 'Hello from lab1' /root/page.html 2>/dev/null`,
				"HTML-страница скачана",
				"Скачай тело ответа: curl -s http://localhost/ > /root/page.html"),
		},
	},

	// ── Lab 2: SSH-ключи ──
	"ch-ladv-lab2": {
		Setup: `set -e
rm -rf /root/.ssh
mkdir -p /root/.ssh && chmod 700 /root/.ssh
rm -f /root/key_info.txt`,
		Checks: map[int]string{
			1: check(`[ -f /root/.ssh/mykey ] && [ -f /root/.ssh/mykey.pub ] && grep -q 'ssh-ed25519' /root/.ssh/mykey.pub`,
				"пара ключей ed25519 создана",
				"ssh-keygen -t ed25519 -f /root/.ssh/mykey -N '' (флаг -N '' — без пароля)"),
			2: check(`[ "$(stat -c '%a' /root/.ssh/mykey 2>/dev/null)" = 600 ]`,
				"права на приватный ключ — 600",
				"chmod 600 /root/.ssh/mykey — сейчас $(stat -c '%a' /root/.ssh/mykey 2>/dev/null)"),
			3: check(`[ -f /root/.ssh/authorized_keys ] && grep -q "$(cut -d' ' -f2 /root/.ssh/mykey.pub)" /root/.ssh/authorized_keys`,
				"публичный ключ добавлен в authorized_keys",
				"cat /root/.ssh/mykey.pub >> /root/.ssh/authorized_keys"),
			4: check(`grep -qi 'SHA256:' /root/key_info.txt 2>/dev/null && grep -qi 'ed25519' /root/key_info.txt`,
				"fingerprint ключа сохранён",
				"ssh-keygen -lf /root/.ssh/mykey.pub > /root/key_info.txt"),
			5: check(`grep -qi 'Host myserver' /root/.ssh/config 2>/dev/null && grep -q '127\.0\.0\.1' /root/.ssh/config && grep -qi 'IdentityFile' /root/.ssh/config && grep -qi 'User root' /root/.ssh/config`,
				"SSH config для myserver настроен",
				"Создай /root/.ssh/config с блоком: Host myserver / HostName 127.0.0.1 / User root / IdentityFile ~/.ssh/mykey"),
		},
	},

	// ── Lab 3: переменные окружения ──
	"ch-ladv-lab3": {
		Setup: `set -e
rm -rf /opt/devops/lab3
mkdir -p /opt/devops/lab3/bin
printf 'DB_HOST=db.internal\nDB_PORT=5432\nAPP_ENV=production\n' > /opt/devops/lab3/app.env
printf '#!/bin/bash\necho "myapp running"\n' > /opt/devops/lab3/bin/myapp
chmod +x /opt/devops/lab3/bin/myapp
rm -f /root/db_host.txt /root/myapp_path.txt /root/.zshrc`,
		Checks: map[int]string{
			1: check(`[ "$(tr -d ' \n' < /root/db_host.txt 2>/dev/null)" = db.internal ]`,
				"значение DB_HOST сохранено",
				"Загрузи переменные и сохрани значение: set -a; . /opt/devops/lab3/app.env; set +a; echo \\$DB_HOST > /root/db_host.txt"),
			2: check(`grep -qE '^\s*export\s+EDITOR=nano' /root/.zshrc 2>/dev/null`,
				"export EDITOR=nano добавлен в ~/.zshrc",
				"echo 'export EDITOR=nano' >> ~/.zshrc — затем source ~/.zshrc и echo \\$EDITOR"),
			3: check(`grep -q '/opt/devops/lab3/bin/myapp' /root/myapp_path.txt 2>/dev/null`,
				"myapp найден через PATH",
				"export PATH=\"\\$PATH:/opt/devops/lab3/bin\" затем which myapp > /root/myapp_path.txt"),
		},
	},

	// ── Lab 4: операторы && и || ──
	"ch-ladv-lab4": {
		Setup: `set -e
rm -rf /root/project /root/deploy
rm -f /root/safe_deploy.sh`,
		Checks: map[int]string{
			1: check(`[ -d /root/project ] && [ -s /root/project/README.md ]`,
				"каталог и README.md созданы",
				"mkdir /root/project && echo 'project' > /root/project/README.md"),
			2: check(`grep -q 'default config' /root/project/config.txt 2>/dev/null`,
				"config.txt создан веткой ||",
				"cat /root/project/config.txt || echo 'default config' > /root/project/config.txt"),
			3: check(`[ -x /root/safe_deploy.sh ] && grep -q 'set -e' /root/safe_deploy.sh && grep -q deployed /root/deploy/status.txt 2>/dev/null`,
				"скрипт с set -e создан и отработал",
				"Скрипт должен начинаться с set -e, создавать /root/deploy и писать deployed в status.txt; не забудь chmod +x и запуск"),
		},
	},

	// ── Lab 5: первые скрипты ──
	"ch-ladv-lab5": {
		Setup: `set -e
rm -rf /opt/devops/lab5
mkdir -p /opt/devops/lab5
printf 'app=1\n' > /opt/devops/lab5/config.txt
rm -f /root/hello.sh /root/version.sh /root/dated.sh /root/check_file.sh /root/greet.sh /root/healthcheck.sh`,
		Checks: map[int]string{
			1: check(`[ -x /root/hello.sh ] && head -1 /root/hello.sh | grep -q '^#!.*bash' && /root/hello.sh 2>/dev/null | grep -q 'Hello, DevOps!'`,
				"скрипт исполняемый и печатает нужную строку",
				"Первая строка #!/bin/bash, тело echo 'Hello, DevOps!', затем chmod +x /root/hello.sh"),
			2: check(`grep -q 'VERSION=' /root/version.sh 2>/dev/null && bash /root/version.sh 2>/dev/null | grep -q 'App version: 1\.0\.0'`,
				"версия выводится через переменную",
				"VERSION=1.0.0 и echo \"App version: \\$VERSION\" — запусти скрипт и проверь вывод"),
			3: check(`[ -x /root/dated.sh ] && grep -q '\$(' /root/dated.sh && /root/dated.sh 2>/dev/null | grep -qE 'Today: [0-9]{4}-[0-9]{2}-[0-9]{2}'`,
				"дата подставляется через $( )",
				"D=\\$(date +%Y-%m-%d); echo \"Today: \\$D\" — и chmod +x /root/dated.sh"),
			4: check(`[ -x /root/check_file.sh ] && /root/check_file.sh 2>/dev/null | grep -q '^found$'`,
				"проверка существования файла работает",
				"if [ -f /opt/devops/lab5/config.txt ]; then echo found; else echo 'not found'; fi — и chmod +x"),
			5: check(`[ -x /root/greet.sh ] && /root/greet.sh World 2>/dev/null | grep -q 'Hello, World!'`,
				"скрипт принимает имя аргументом",
				"echo \"Hello, \\$1!\" — проверь запуском: /root/greet.sh World"),
			6: check(`[ -x /root/healthcheck.sh ] && /root/healthcheck.sh 2>/dev/null | grep -qE '^(OK|FAIL)$'`,
				"healthcheck печатает OK или FAIL",
				"if pgrep sshd >/dev/null; then echo OK; else echo FAIL; fi — и chmod +x"),
		},
	},

	// ── Lab 6: циклы и функции ──
	"ch-ladv-lab6": {
		Setup: `set -e
rm -rf /opt/devops/lab6 /root/dev /root/staging /root/prod
mkdir -p /opt/devops/lab6
printf 'line1\nline2\nline3\n' > /opt/devops/lab6/app.log
printf 'a\nb\n'                > /opt/devops/lab6/sys.log
printf 'web1\nweb2\ndb1\n'     > /opt/devops/lab6/servers.txt
rm -f /root/list_envs.sh /root/count_logs.sh /root/count_to_five.sh /root/logger.sh /root/check_servers.sh /root/ensure_envs.sh`,
		Checks: map[int]string{
			1: check(`[ -x /root/list_envs.sh ] && /root/list_envs.sh 2>/dev/null | tr -d ' ' | diff -q - <(printf 'dev\nstaging\nprod\n') >/dev/null 2>&1`,
				"цикл for печатает три окружения по одному в строке",
				"for env in dev staging prod; do echo \\$env; done — и chmod +x /root/list_envs.sh"),
			2: check(`[ -x /root/count_logs.sh ] && out=$(/root/count_logs.sh 2>/dev/null) && echo "$out" | grep -q 'app\.log' && echo "$out" | grep -q 'sys\.log' && echo "$out" | grep -qE '[0-9]'`,
				"по каждому .log выведено количество строк",
				"for f in /opt/devops/lab6/*.log; do wc -l \"\\$f\"; done — и chmod +x"),
			3: check(`[ -x /root/count_to_five.sh ] && /root/count_to_five.sh 2>/dev/null | tr -d ' ' | diff -q - <(printf '1\n2\n3\n4\n5\n') >/dev/null 2>&1`,
				"while печатает числа от 1 до 5",
				"i=1; while [ \\$i -le 5 ]; do echo \\$i; i=\\$((i+1)); done — и chmod +x"),
			4: check(`[ -x /root/logger.sh ] && out=$(/root/logger.sh 2>/dev/null) && echo "$out" | grep -qE '\[[0-9]{2}:[0-9]{2}:[0-9]{2}\].*Script started' && echo "$out" | grep -q 'Script done'`,
				"функция log() печатает метку времени и оба сообщения",
				"log() { echo \"[\\$(date +%H:%M:%S)] \\$1\"; }; log 'Script started'; log 'Script done'"),
			5: check(`[ -x /root/check_servers.sh ] && out=$(/root/check_servers.sh 2>/dev/null) && echo "$out" | grep -q 'Checking: web1' && echo "$out" | grep -q 'Checking: db1'`,
				"файл прочитан построчно",
				"while read -r line; do echo \"Checking: \\$line\"; done < /opt/devops/lab6/servers.txt"),
			6: check(`[ -x /root/ensure_envs.sh ] && [ -d /root/dev ] && [ -d /root/staging ] && [ -d /root/prod ]`,
				"недостающие каталоги созданы скриптом",
				"for d in dev staging prod; do [ -d /root/\\$d ] || mkdir /root/\\$d; done — не забудь запустить скрипт"),
		},
	},

	// ── Lab 7: бэкапы ──
	"ch-ladv-lab7": {
		Setup: `set -e
rm -rf /opt/data /root/backups
rm -f /root/backup.sh /root/data_backup.tar.gz /root/old_backups.txt
mkdir -p /opt/data
printf 'important data 1\n' > /opt/data/file1.txt
printf 'important data 2\n' > /opt/data/file2.txt`,
		Checks: map[int]string{
			1: check(`[ -f /root/data_backup.tar.gz ] && tar -tzf /root/data_backup.tar.gz 2>/dev/null | grep -q 'file1\.txt'`,
				"архив /opt/data создан",
				"tar -czf /root/data_backup.tar.gz -C /opt data"),
			2: check(`[ -x /root/backup.sh ] && ls /root/backups/backup_$(date +%Y-%m-%d).tar.gz >/dev/null 2>&1`,
				"скрипт создал архив с сегодняшней датой в имени",
				"В скрипте: mkdir -p /root/backups && tar -czf /root/backups/backup_\\$(date +%F).tar.gz -C /opt data — затем chmod +x и запуск"),
			3: check(`grep -q '\.tar\.gz' /root/old_backups.txt 2>/dev/null`,
				"список архивов сохранён",
				"find /root/backups -name '*.tar.gz' > /root/old_backups.txt"),
		},
	},

	// ── Lab 8: пакеты ──
	// htop не предустановлен в образе: локальный офлайн-репозиторий даёт
	// возможность реально выполнить apt-get install без сети.
	"ch-ladv-lab8": {
		Setup: `apt-get remove -y htop >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: check(`command -v htop >/dev/null && dpkg -s htop >/dev/null 2>&1`,
				"пакет htop установлен",
				"Обнови индекс и поставь пакет без подтверждения: apt-get update && apt-get install -y htop"),
		},
	},

	// ── Lab 9: место на диске ──
	"ch-ladv-lab9": {
		Setup: `set -e
mkdir -p /var/lib/gldemo /var/log/gldemo /var/cache/gldemo
head -c 300000 /dev/zero | tr '\0' 'a' > /var/lib/gldemo/big1
head -c 200000 /dev/zero | tr '\0' 'b' > /var/log/gldemo/big2
head -c 100000 /dev/zero | tr '\0' 'c' > /var/cache/gldemo/big3
rm -f /root/var_sizes.txt /root/largest_files.txt`,
		Checks: map[int]string{
			1: check(`[ -s /root/var_sizes.txt ] && grep -q '/var/' /root/var_sizes.txt && grep -qE '[0-9]' /root/var_sizes.txt`,
				"размеры поддиректорий /var сохранены",
				"du -h --max-depth=1 /var 2>/dev/null | sort -hr > /root/var_sizes.txt"),
			2: check(`[ "$(grep -c . /root/largest_files.txt 2>/dev/null)" = 5 ] && grep -q '/var/' /root/largest_files.txt`,
				"найдены 5 самых больших файлов",
				"find /var -type f -exec du -h {} + 2>/dev/null | sort -hr | head -5 > /root/largest_files.txt (ровно 5 строк)"),
		},
	},

	// Lab 10 (монтирование tmpfs) намеренно без автопроверки: mount требует
	// CAP_SYS_ADMIN, которого у учебного контейнера нет.

	// ── Lab 11: bash-приёмы ──
	"ch-ladv-lab11": {
		Setup: `set -e
rm -rf /opt/devops/lab11
mkdir -p /opt/devops/lab11
printf 'server=HOSTNAME port=8080 backup=HOSTNAME\n' > /opt/devops/lab11/template.txt
printf 'nginx\npostgres\nredis\n'                    > /opt/devops/lab11/services.txt
rm -f /root/array_demo.sh /root/servers.txt /root/result.txt /root/log_demo.sh /root/process_lines.sh`,
		Checks: map[int]string{
			1: check(`[ -x /root/array_demo.sh ] && diff -q /root/servers.txt <(printf 'web1\nweb2\nweb3\n') >/dev/null 2>&1`,
				"массив перебран, имена записаны в файл",
				"servers=(web1 web2 web3); for s in \"\\${servers[@]}\"; do echo \\$s >> /root/servers.txt; done — и запусти скрипт"),
			2: check(`grep -q 'prod-server-01' /root/result.txt 2>/dev/null && ! grep -q 'HOSTNAME' /root/result.txt`,
				"подстановка выполнена через parameter expansion",
				"line=\\$(cat /opt/devops/lab11/template.txt); echo \"\\${line//HOSTNAME/prod-server-01}\" > /root/result.txt"),
			3: check(`[ -x /root/log_demo.sh ] && /root/log_demo.sh 2>/dev/null | grep -qE '\[[0-9]{2}:[0-9]{2}:[0-9]{2}\] INFO: Script started'`,
				"функция log_info печатает метку времени и сообщение",
				"log_info() { echo \"[\\$(date +%H:%M:%S)] INFO: \\$1\"; }; log_info 'Script started'"),
			4: check(`[ -x /root/process_lines.sh ] && out=$(/root/process_lines.sh 2>/dev/null) && echo "$out" | grep -q 'Checking: nginx' && echo "$out" | grep -q 'Checking: redis'`,
				"файл сервисов обработан построчно",
				"while read -r svc; do echo \"Checking: \\$svc\"; done < /opt/devops/lab11/services.txt"),
		},
	},
}
