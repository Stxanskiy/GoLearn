package main

// Fixtures + auto-checks for the Linux labs of "Экспресс-погружение в DevOps".
//
// Only the Linux chapters are auto-checked here: the Docker and Kubernetes
// chapters of this course need a container runtime inside the sandbox, which
// the current isolated (network-less, unprivileged) lab container does not
// provide — those steps stay manual.

var expressDevopsLabs = map[string]labSpec{
	// ── Lab 1: файлы и каталоги ──
	"ch-exp-linux-lab1": {
		Setup: `set -e
rm -rf /root/myproject /root/configs
mkdir -p /opt/devops/lab1
printf 'server_port=8080\nlog_level=info\n' > /opt/devops/lab1/config.txt
printf 'EXPRESS_SECRET_42\n'                > /opt/devops/lab1/.secret
rm -f /root/found_secret.txt`,
		Checks: map[int]string{
			1: check(`[ -d /root/myproject ]`,
				"каталог /root/myproject создан",
				"mkdir /root/myproject"),
			2: check(`grep -q 'Hello DevOps' /root/myproject/README.txt 2>/dev/null`,
				"README.txt содержит нужную строку",
				"echo 'Hello DevOps' > /root/myproject/README.txt"),
			3: check(`[ -f /root/myproject/config.txt ] && grep -q server_port /root/myproject/config.txt`,
				"config.txt скопирован",
				"cp /opt/devops/lab1/config.txt /root/myproject/config.txt"),
			4: check(`[ -f /root/myproject/myconfig.txt ] && [ ! -e /root/myproject/config.txt ]`,
				"файл переименован",
				"mv /root/myproject/config.txt /root/myproject/myconfig.txt"),
			5: check(`grep -q EXPRESS_SECRET_42 /root/found_secret.txt 2>/dev/null`,
				"скрытый файл найден",
				"Скрытые файлы видны через ls -a: cat /opt/devops/lab1/.secret > /root/found_secret.txt"),
			6: check(`[ -f /root/configs/myconfig.txt ] && [ ! -e /root/myproject/myconfig.txt ]`,
				"файл перемещён в /root/configs",
				"mkdir /root/configs && mv /root/myproject/myconfig.txt /root/configs/"),
			7: check(`[ ! -e /root/myproject/README.txt ]`,
				"README.txt удалён",
				"rm /root/myproject/README.txt"),
			8: check(`[ ! -e /root/myproject ]`,
				"каталог удалён рекурсивно",
				"rm -r /root/myproject"),
		},
	},

	// ── Lab 2: чтение и поиск ──
	"ch-exp-linux-lab2": {
		Setup: `set -e
rm -rf /opt/devops/lab2
mkdir -p /opt/devops/lab2/scripts
for i in $(seq 1 15); do echo "2024-05-01 10:00:$i INFO  request handled"; done > /opt/devops/lab2/app.log
{
  echo "2024-05-01 10:01:00 ERROR db timeout"
  echo "2024-05-01 10:01:05 WARN  retry"
} >> /opt/devops/lab2/app.log
printf '#!/bin/bash\necho deploy\n'  > /opt/devops/lab2/deploy.sh
printf '#!/bin/bash\necho rollback\n' > /opt/devops/lab2/scripts/rollback.sh
printf 'notes\n'                      > /opt/devops/lab2/notes.txt
rm -f /root/last10.txt /root/no_info.txt /root/scripts.txt`,
		Checks: map[int]string{
			1: check(`tail -10 /opt/devops/lab2/app.log | diff -q - /root/last10.txt >/dev/null 2>&1`,
				"последние 10 строк сохранены",
				"tail -10 /opt/devops/lab2/app.log > /root/last10.txt"),
			2: check(`grep -v INFO /opt/devops/lab2/app.log | diff -q - /root/no_info.txt >/dev/null 2>&1`,
				"строки без INFO отобраны",
				"grep -v INFO /opt/devops/lab2/app.log > /root/no_info.txt"),
			3: check(`grep -q 'deploy\.sh' /root/scripts.txt 2>/dev/null && grep -q 'rollback\.sh' /root/scripts.txt`,
				"найдены оба .sh файла, включая вложенный",
				"find /opt/devops/lab2 -name '*.sh' > /root/scripts.txt"),
		},
	},
}
