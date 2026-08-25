package main

// Fixtures + auto-checks for the Linux troubleshooting trainer.
//
// These labs are the opposite of the course labs: the setup deliberately BREAKS
// something (a shell config, a unit file, a permission, an occupied port) and
// the check asserts the service actually works again — not that a particular
// command was typed. That keeps every road to a working system valid.

var gymTroubleshootLabs = map[string]labSpec{
	// ── Lab 1: сервисы и запуск ──
	"ch-ltrouble-lab1": {
		Setup: `set -e
mkdir -p /opt/app

# 1. Сломанный zsh-конфиг: новые интерактивные сессии теряют PATH.
printf 'export PATH=/nonexistent\n' >> /root/.zshrc

# 2. Unit без политики перезапуска.
cat > /etc/systemd/system/simpleapp.service <<'UNIT'
[Unit]
Description=Simple app

[Service]
ExecStart=/usr/bin/python3 -c "import time
while True: time.sleep(5)"

[Install]
WantedBy=multi-user.target
UNIT

# 3. Unit с неверным путём до бинарника (сам бинарник лежит в /opt/app).
printf '#!/bin/bash\nwhile true; do sleep 5; done\n' > /opt/app/myapp
chmod +x /opt/app/myapp
cat > /etc/systemd/system/myapp.service <<'UNIT'
[Unit]
Description=My app

[Service]
ExecStart=/usr/local/bin/myapp

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload >/dev/null 2>&1 || true

# 4. Скрипт, которому нужен явный режим запуска.
cat > /opt/app/start.sh <<'SH'
#!/bin/bash
if [ "$1" != "--production" ]; then
    echo "ERROR: не указан режим запуска. Требуется: $0 --production" >&2
    exit 1
fi
echo started > /opt/app/started.flag
echo "app started in production mode"
SH
chmod +x /opt/app/start.sh
rm -f /opt/app/started.flag`,
		Checks: map[int]string{
			1: check(`zsh -i -c 'command -v ls' >/dev/null 2>&1`,
				"новые zsh-сессии снова находят команды",
				"Посмотри ~/.zshrc: там строка, которая затирает PATH вместо того чтобы дополнять его"),
			2: check(`grep -qE '^\s*Restart=(on-failure|always)' /etc/systemd/system/simpleapp.service`,
				"в unit добавлена политика перезапуска",
				"В секцию [Service] файла /etc/systemd/system/simpleapp.service добавь Restart=on-failure и выполни systemctl daemon-reload"),
			3: check(`systemctl is-active myapp >/dev/null 2>&1`,
				"myapp.service в состоянии active",
				"Сравни ExecStart в /etc/systemd/system/myapp.service с реальным путём бинарника (ls /opt/app), поправь, затем daemon-reload и start"),
			4: check(`[ -f /opt/app/started.flag ]`,
				"приложение стартовало",
				"Запусти /opt/app/start.sh — он сам печатает, какого аргумента ему не хватает"),
		},
	},

	// ── Lab 2: права и владельцы ──
	"ch-ltrouble-lab2": {
		Setup: `set -e
mkdir -p /opt/app /opt/app/reports
id svcuser    >/dev/null 2>&1 || useradd -m svcuser
id appuser    >/dev/null 2>&1 || useradd -m appuser
id reportuser >/dev/null 2>&1 || useradd -m reportuser

# 1. Слишком открытые права на authorized_keys — sshd такое отвергает.
mkdir -p /root/.ssh
ssh-keygen -q -t ed25519 -f /root/.ssh/id_ed25519 -N '' <<<y >/dev/null 2>&1 || true
cat /root/.ssh/id_ed25519.pub > /root/.ssh/authorized_keys
chmod 777 /root/.ssh
chmod 666 /root/.ssh/authorized_keys

# 2. Конфиг читает сервисный пользователь, но владелец — root (mode менять нельзя).
printf 'db_host=localhost\n' > /opt/app/config.yml
chown root:root /opt/app/config.yml
chmod 600 /opt/app/config.yml
cat > /opt/app/run-config-app.sh <<'SH'
#!/bin/bash
# config-app всегда работает от сервисного пользователя svcuser
su -s /bin/bash -c 'cat /opt/app/config.yml > /dev/null' svcuser
SH
chmod +x /opt/app/run-config-app.sh

# 3. db.conf недоступен appuser.
printf 'dsn=postgres://localhost/app\n' > /opt/app/db.conf
chown root:root /opt/app/db.conf
chmod 640 /opt/app/db.conf

# 4. Каталог отчётов не пишется от имени reportuser.
cat > /opt/app/run-report-app.sh <<'SH'
#!/bin/bash
su -s /bin/bash -c 'echo "report body" > /opt/app/reports/report.txt' reportuser
SH
chmod +x /opt/app/run-report-app.sh
chown root:root /opt/app/reports
chmod 755 /opt/app/reports
rm -f /opt/app/reports/report.txt`,
		Checks: map[int]string{
			1: check(`[ "$(stat -c '%a' /root/.ssh/authorized_keys)" -le 600 ] && [ "$(stat -c '%a' /root/.ssh)" -le 700 ]`,
				"права на ~/.ssh и authorized_keys приведены к безопасным",
				"sshd игнорирует ключи при слишком открытых правах: chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys (сейчас $(stat -c '%a' /root/.ssh) и $(stat -c '%a' /root/.ssh/authorized_keys))"),
			2: check(`/opt/app/run-config-app.sh >/dev/null 2>&1 && [ "$(stat -c '%a' /opt/app/config.yml)" = 600 ]`,
				"config-app запускается, права файла остались 600",
				"Права менять нельзя — поменяй владельца: chown svcuser /opt/app/config.yml"),
			3: check(`su -s /bin/bash -c 'test -r /opt/app/db.conf' appuser 2>/dev/null && [ "$(stat -c '%a' /opt/app/db.conf)" = 640 ]`,
				"appuser читает db.conf, mode не изменён",
				"Файл 640 — читать его может владелец и группа. Отдай файл нужному пользователю или группе: chown appuser /opt/app/db.conf (или chgrp)"),
			4: check(`/opt/app/run-report-app.sh >/dev/null 2>&1 && [ -s /opt/app/reports/report.txt ] && [ "$(stat -c '%a' /opt/app/reports)" = 755 ]`,
				"отчёт записан, права каталога не изменены",
				"755 даёт запись только владельцу — сделай владельцем каталога reportuser: chown reportuser /opt/app/reports"),
		},
	},

	// ── Lab 3: cron и порты ──
	"ch-ltrouble-lab3": {
		Setup: `set -e
mkdir -p /opt/app

# 1. В crontab остался путь к старому скрипту.
printf '#!/bin/bash\necho backup ok\n' > /opt/app/backup.sh
chmod +x /opt/app/backup.sh
printf '0 3 * * * /opt/app/old-backup.sh\n' | crontab -

# 2. Сервис слушает не тот порт.
cat > /etc/systemd/system/webapp.service <<'UNIT'
[Unit]
Description=Web app

[Service]
ExecStart=/usr/bin/python3 -m http.server 9090
WorkingDirectory=/opt/app

[Install]
WantedBy=multi-user.target
UNIT
systemctl daemon-reload >/dev/null 2>&1 || true
systemctl stop webapp >/dev/null 2>&1 || true
systemctl start webapp >/dev/null 2>&1 || true

# 3. Порт 3000 занят посторонним процессом.
cd /opt/app
nohup python3 -m http.server 3000 >/var/log/stray.log 2>&1 &
echo $! > /run/gl-stray.pid
cd /
sleep 0.5

# 4. Файл cron-задания не исполняемый.
printf '#!/bin/bash\necho maintenance done\n' > /opt/app/maintenance.sh
chmod 644 /opt/app/maintenance.sh`,
		Checks: map[int]string{
			1: check(`crontab -l 2>/dev/null | grep -q '/opt/app/backup\.sh' && ! crontab -l 2>/dev/null | grep -q 'old-backup\.sh'`,
				"cron-задача указывает на существующий скрипт",
				"Посмотри crontab -l: путь ведёт на несуществующий old-backup.sh. Поправь через crontab -e на /opt/app/backup.sh"),
			2: check(`grep -q '8080' /etc/systemd/system/webapp.service && ss -ltn 2>/dev/null | grep -q ':8080 '`,
				"webapp слушает порт 8080",
				"Поправь порт в ExecStart файла /etc/systemd/system/webapp.service, затем systemctl daemon-reload && systemctl restart webapp"),
			3: check(`! kill -0 "$(cat /run/gl-stray.pid 2>/dev/null)" 2>/dev/null && ss -ltn 2>/dev/null | grep -q ':3000 '`,
				"посторонний процесс убран, на 3000 работает твоё приложение",
				"Найди, кто держит порт: ss -ltnp | grep 3000 — заверши этот процесс и запусти на 3000 своё приложение"),
			4: check(`[ -x /opt/app/maintenance.sh ]`,
				"файл задания стал исполняемым",
				"cron запускает файл как программу: chmod +x /opt/app/maintenance.sh"),
		},
	},

	// ── Lab 4: разбор инцидентов ──
	// Задание 1 (магазин) остаётся ручным: его сценарий завязан на Web Preview
	// и на несколько сервисов, которых нет в изолированной песочнице.
	"ch-ltrouble-lab4": {
		Setup: `set -e
mkdir -p /opt/payment
cat > /usr/local/bin/payment_logger <<'SH'
#!/bin/bash
LOG=/opt/payment/audit.log
exec 3<> "$LOG"
echo "INFO  payment service started" >&3
echo "AUDIT recovery-code: GL-RECOVER-7731" >&3
rm -f "$LOG"
while true; do echo "INFO  heartbeat" >&3; sleep 5; done
SH
chmod +x /usr/local/bin/payment_logger
(setsid /usr/local/bin/payment_logger >/dev/null 2>&1 &)
sleep 0.5
rm -f /root/recovered-payment-token.txt`,
		Checks: map[int]string{
			2: check(`grep -q 'GL-RECOVER-7731' /root/recovered-payment-token.txt 2>/dev/null`,
				"recovery-code восстановлен из удалённого, но открытого файла",
				"Файл удалён, но процесс держит его открытым: найди PID (ps aux | grep payment_logger), посмотри ls -l /proc/<PID>/fd — у удалённого файла будет пометка (deleted). Прочитай его: cat /proc/<PID>/fd/<N> > /root/recovered-payment-token.txt"),
		},
	},
}
