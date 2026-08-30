package main

// Fixtures + auto-checks for "Ansible" (module ansible).
//
// Ansible fits the offline sandbox perfectly: every lab runs against localhost
// with `connection: local`, so no SSH and no network are needed — the student
// writes real playbooks that really execute in the container. The base sandbox
// image must ship `ansible` (added to deploy/sandbox/Dockerfile); until that
// image is rebuilt the labs open a terminal but the checks that invoke
// ansible-playbook will report it missing.
//
// Every lesson shares /root/ansible-lab, pre-seeded with ansible.cfg and a
// localhost inventory so `ansible-playbook <file>` just works.

const ansibleInit = `set -e
rm -rf /root/ansible-lab
mkdir -p /root/ansible-lab/templates
cd /root/ansible-lab
cat > ansible.cfg <<'CFG'
[defaults]
inventory = inventory.ini
host_key_checking = False
retry_files_enabled = False
deprecation_warnings = False
CFG
cat > inventory.ini <<'INI'
[local]
localhost ansible_connection=local
INI`

// ansibleIdempotent runs the given playbook and passes only if the recap shows
// changed=0 (nothing left to do on a second run).
func ansibleIdempotent(playbook string) string {
	return `cd /root/ansible-lab && ansible-playbook ` + playbook +
		` 2>/dev/null | grep -qE 'changed=0.*unreachable=0'`
}

var ansibleLabs = map[string]labSpec{
	// ── Lab 1: ad-hoc + первый playbook ──
	"ch-ansible-lab1": {
		Setup: ansibleInit,
		Checks: map[int]string{
			1: check(`[ -f /root/ansible-lab/hello.txt ]`,
				"файл hello.txt создан ad-hoc модулем file",
				`ansible localhost -c local -m file -a "path=/root/ansible-lab/hello.txt state=touch"`),
			2: check(`grep -q 'Managed by Ansible' /root/ansible-lab/managed.txt 2>/dev/null`,
				"playbook создал managed.txt через copy",
				"в site.yml задачей copy запиши 'Managed by Ansible' в managed.txt, затем ansible-playbook site.yml"),
			3: check(`[ -d /root/ansible-lab/data ]`,
				"каталог data создан задачей file",
				"добавь в site.yml задачу file со state: directory для /root/ansible-lab/data"),
			4: check(ansibleIdempotent("site.yml"),
				"playbook идемпотентен (повторный прогон: changed=0)",
				"второй ansible-playbook site.yml должен дать changed=0 — используй модули file/copy, не command"),
		},
	},

	// ── Lab 2: модули для файлов ──
	"ch-ansible-lab2": {
		Setup: ansibleInit + `
touch /root/ansible-lab/old.txt`,
		Checks: map[int]string{
			1: check(`[ -d /root/ansible-lab/app ]`,
				"каталог app создан модулем file",
				"задача file: path: /root/ansible-lab/app, state: directory"),
			2: check(`grep -q 'port=8080' /root/ansible-lab/app/app.conf 2>/dev/null`,
				"app.conf создан модулем copy (port=8080)",
				"задача copy: content: \"port=8080\\n\", dest: /root/ansible-lab/app/app.conf"),
			3: check(`grep -q 'debug=true' /root/ansible-lab/app/app.conf 2>/dev/null`,
				"строка debug=true добавлена модулем lineinfile",
				"задача lineinfile: path: .../app.conf, line: 'debug=true'"),
			4: check(`[ ! -e /root/ansible-lab/old.txt ]`,
				"old.txt удалён (state: absent)",
				"задача file: path: /root/ansible-lab/old.txt, state: absent — затем ansible-playbook files.yml"),
		},
	},

	// ── Lab 3: переменные и register ──
	"ch-ansible-lab3": {
		Setup: ansibleInit,
		Checks: map[int]string{
			1: check(`grep -q 'hello ansible' /root/ansible-lab/greeting.txt 2>/dev/null`,
				"переменная greeting подставлена в файл",
				"vars: greeting: \"hello ansible\"; copy content: \"{{ greeting }}\\n\" в greeting.txt"),
			2: check(`[ -s /root/ansible-lab/os.txt ] && grep -qE '^[A-Za-z]' /root/ansible-lab/os.txt`,
				"дистрибутив из ansible_facts записан в os.txt",
				"copy content: \"{{ ansible_facts['distribution'] }}\\n\" в os.txt (facts включены по умолчанию)"),
			3: check(`[ -s /root/ansible-lab/host.txt ] && [ "$(cat /root/ansible-lab/host.txt)" = "$(cat /etc/hostname)" ]`,
				"результат команды сохранён через register и записан в host.txt",
				"command: cat /etc/hostname (register: host_out, changed_when: false); copy content: \"{{ host_out.stdout }}\\n\""),
		},
	},

	// ── Lab 4: template и handler ──
	"ch-ansible-lab4": {
		Setup: ansibleInit,
		Checks: map[int]string{
			1: check(`grep -qE 'port = 909[01]' /root/ansible-lab/service.conf 2>/dev/null`,
				"шаблон отрендерен (port = 9090)",
				"templates/service.conf.j2 со строкой 'port = {{ svc_port }}', vars svc_port: 9090, задача template -> service.conf"),
			2: check(`grep -q 'debug = on' /root/ansible-lab/service.conf 2>/dev/null`,
				"условие Jinja2 добавило debug = on",
				"в шаблон: {% if enable_debug %}debug = on{% endif %}; vars enable_debug: true; перезапусти playbook"),
			3: check(`[ -f /root/ansible-lab/reloaded.marker ] && grep -q 'port = 9091' /root/ansible-lab/service.conf 2>/dev/null`,
				"изменение шаблона через notify запустило handler",
				"смени svc_port на 9091, добавь notify: reload service и handler, создающий reloaded.marker; перезапусти playbook"),
		},
	},

	// ── Lab 5: when и loop ──
	"ch-ansible-lab5": {
		Setup: ansibleInit,
		Checks: map[int]string{
			1: check(`[ -d /root/ansible-lab/logs ] && [ -d /root/ansible-lab/data ] && [ -d /root/ansible-lab/cache ]`,
				"три каталога созданы одним loop",
				"file (state: directory, path: /root/ansible-lab/{{ item }}) с loop: [logs, data, cache]"),
			2: check(`[ -f /root/ansible-lab/alpha.flag ] && [ -f /root/ansible-lab/gamma.flag ] && [ ! -e /root/ansible-lab/beta.flag ]`,
				"loop + when пропустил beta",
				"loop: [alpha, beta, gamma] создаёт <item>.flag, но when: item != 'beta'"),
			3: check(`[ -f /root/ansible-lab/prod.marker ]`,
				"задача с when: make_prod выполнилась",
				"vars make_prod: true; задача создаёт prod.marker с when: make_prod | bool"),
		},
	},

	// ── Lab 6: роль, идемпотентность, vault ──
	"ch-ansible-lab6": {
		Setup: ansibleInit,
		Checks: map[int]string{
			1: check(`[ -f /root/ansible-lab/roles/webapp/tasks/main.yml ]`,
				"заготовка роли webapp создана",
				"cd /root/ansible-lab && ansible-galaxy init roles/webapp"),
			2: check(`[ -f /root/ansible-lab/webapp/index.html ]`,
				"роль создала /root/ansible-lab/webapp/index.html",
				"в roles/webapp/tasks/main.yml — file(directory) + copy(index.html); site.yml с roles: [webapp]; ansible-playbook site.yml"),
			3: check(ansibleIdempotent("site.yml"),
				"роль идемпотентна (повторный прогон: changed=0)",
				"второй ansible-playbook site.yml должен дать changed=0 (не используй command/shell без creates)"),
			4: check(`head -1 /root/ansible-lab/secrets.yml 2>/dev/null | grep -q 'ANSIBLE_VAULT'`,
				"secrets.yml зашифрован через ansible-vault",
				"создай secrets.yml и: ansible-vault encrypt secrets.yml"),
		},
	},
}
