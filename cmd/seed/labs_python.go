package main

// Fixtures for the Python path. Only Setup for now: the first pass of this
// course ships without auto-checks, so a task's result is the student's own
// "Готово". Setup is still required — without it the tasks talk about files
// that do not exist.
//
// The base sandbox image already carries python3 and python3-venv (see
// deploy/sandbox/Dockerfile), so the first courses of the path need no new
// image. That changes at postgres-psycopg and pyqt6-start.

var pythonFirstStepsLabs = map[string]labSpec{
	// Lesson 4. The tasks read hello.py, then break it on purpose, then repair
	// a file broken for them.
	"pfs-04": {
		Setup: `
cat > /root/hello.py <<'PYEOF'
print("Привет из файла")
print("Меня запустили")
PYEOF

# Задание 13: одна ошибка, и номер строки в сообщении будет НЕ тот, что сломан.
# Скобка не закрыта в строке 3, а Python споткнётся на строке 4 — именно это
# и надо увидеть, прежде чем править.
cat > /root/task_broken.py <<'PYEOF'
print("Отчёт за сентябрь")
print()
print("Заказов:", 128
print("Выручка:", 45900)
print("Средний чек:", 358)
PYEOF

# Чистим файлы, которые студент создаёт сам: лабораторную можно начать заново.
rm -f /root/me.py /root/one.py /root/blocks.py /root/math.py
`,
	},
}
