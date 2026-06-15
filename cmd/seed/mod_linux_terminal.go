package main

// mod_linux_terminal — интерактивный модуль с реальной shell-песочницей.
// Теория оригинальная; темы соответствуют треку "Linux: старт в терминале".
func mod_linux_terminal() M {
	return M{
		Slug:        "linux-terminal-start",
		Title:       "Linux: Старт в терминале",
		Track:       "devops",
		Difficulty:  "beginner",
		Description: "Интерактивный модуль с настоящей Linux-песочницей в браузере. Учишься, выполняя реальные команды.",
		Lessons: []L{
			{
				Slug:       "navigation-and-links",
				Title:      "Навигация, файлы и ссылки",
				Track:      "devops",
				Difficulty: "beginner",
				Content: `<h1>Файловая система как дерево</h1>
<p>В Linux всё начинается с одного корня — каталога <code>/</code>. От него ветвями расходятся остальные каталоги: <code>/etc</code> (конфигурация), <code>/home</code> (домашние папки пользователей), <code>/tmp</code> (временные файлы), <code>/var</code> (изменяемые данные — логи, кеши). Любой файл в системе имеет адрес от корня — это <strong>абсолютный путь</strong>, например <code>/etc/hostname</code>.</p>

<blockquote>Ключевая идея: ты всегда «находишься» в каком-то каталоге. Он называется <strong>текущим рабочим каталогом</strong> (cwd). Команды без полного пути выполняются относительно него.</blockquote>

<h2>Где я? — pwd</h2>
<p>Команда <code>pwd</code> (print working directory) печатает абсолютный путь текущего каталога. Это первое, что стоит сделать, когда не понимаешь, где находишься.</p>

<h2>Что вокруг? — ls</h2>
<p><code>ls</code> показывает содержимое каталога. Полезные флаги:</p>
<ul>
<li><code>ls -l</code> — длинный формат: права, владелец, размер, дата.</li>
<li><code>ls -a</code> — показать скрытые файлы (те, что начинаются с точки, например <code>.bashrc</code>).</li>
<li><code>ls -la</code> — всё сразу. Флаги можно объединять.</li>
</ul>

<h2>Перемещение — cd и относительные пути</h2>
<p><code>cd</code> (change directory) меняет текущий каталог. Здесь важны три специальных обозначения:</p>
<ul>
<li><code>.</code> — текущий каталог;</li>
<li><code>..</code> — родительский каталог (на уровень вверх);</li>
<li><code>~</code> — домашний каталог текущего пользователя.</li>
</ul>
<p><strong>Абсолютный путь</strong> начинается с <code>/</code> и не зависит от того, где ты сейчас: <code>cd /var/log</code>. <strong>Относительный путь</strong> отсчитывается от текущего каталога: если ты в <code>/home/student</code>, то <code>cd ../..</code> приведёт в <code>/</code>.</p>

<h2>Под капотом: что такое файл на самом деле</h2>
<p>Имя файла и сам файл — разные вещи. Данные и метаданные (права, владелец, размер, время) хранятся в структуре под названием <strong>inode</strong>. А имя в каталоге — это всего лишь запись, которая указывает на номер inode. Каталог — это, по сути, таблица «имя → номер inode».</p>

<h2>Жёсткие ссылки (hard links)</h2>
<p>Команда <code>ln файл новое_имя</code> создаёт ещё одно имя, указывающее на <em>тот же самый inode</em>. Это не копия: оба имени равноправны, делят одни данные. Удалишь одно имя — данные останутся, пока на inode ссылается хотя бы одно имя. Жёсткие ссылки не могут пересекать границы файловой системы и обычно не делаются на каталоги.</p>

<h2>Символические ссылки (symlinks)</h2>
<p><code>ln -s цель имя_ссылки</code> создаёт <strong>символическую ссылку</strong> — маленький файл, внутри которого записан <em>путь</em> к цели. Это как ярлык. Symlink может указывать на что угодно, в том числе на несуществующий путь (тогда ссылка «битая») и на каталоги, и на другую файловую систему.</p>
<ul>
<li><code>ls -l</code> показывает symlink как <code>link -&gt; target</code>.</li>
<li><code>readlink link</code> печатает, куда ведёт ссылка.</li>
<li><code>readlink -f link</code> разрешает всю цепочку до реального файла.</li>
</ul>

<blockquote>Разница на практике: жёсткая ссылка — это «второе настоящее имя» данных; символическая — «указатель на путь». Symlink ломается, если цель переименовали; hard link — нет.</blockquote>

<h2>Создание и просмотр</h2>
<ul>
<li><code>mkdir имя</code> — создать каталог (<code>mkdir -p a/b/c</code> создаст всю цепочку).</li>
<li><code>touch файл</code> — создать пустой файл (или обновить время).</li>
<li><code>echo "текст" &gt; файл</code> — записать текст в файл (перезаписав).</li>
<li><code>cat файл</code> — вывести содержимое файла.</li>
</ul>

<p>Ниже — настоящая Linux-песочница. Команды выполняются по-честному в изолированном контейнере: текущий каталог и созданные файлы сохраняются между командами. Решай задания прямо в терминале и жми «Проверить».</p>`,
				Quiz: []Q{
					{
						Question:    "Что выведет команда pwd?",
						Options:     []string{"Список файлов в каталоге", "Абсолютный путь текущего каталога", "Содержимое файла", "Домашний каталог любого пользователя"},
						Correct:     1,
						Explanation: "pwd (print working directory) печатает абсолютный путь каталога, в котором ты сейчас находишься.",
					},
					{
						Question:    "Куда приведёт cd .. из каталога /var/log?",
						Options:     []string{"В /", "В /var", "В /var/log/..", "В домашний каталог"},
						Correct:     1,
						Explanation: ".. — это родительский каталог, на уровень вверх. Из /var/log это /var.",
					},
					{
						Question:    "Чем символическая ссылка (ln -s) отличается от жёсткой (ln)?",
						Options: []string{
							"Ничем, это синонимы",
							"Symlink хранит путь к цели и может ломаться; hard link — ещё одно имя того же inode",
							"Hard link можно делать только на каталоги",
							"Symlink копирует данные файла",
						},
						Correct:     1,
						Explanation: "Symlink — это файл-указатель на путь (как ярлык), он становится битым, если цель убрать. Hard link — равноправное второе имя того же inode.",
					},
					{
						Question:    "Что делает echo \"hi\" > notes.txt?",
						Options:     []string{"Печатает hi и имя файла на экран", "Создаёт символическую ссылку", "Записывает строку hi в файл notes.txt, перезаписав его", "Дописывает hi в конец файла"},
						Correct:     2,
						Explanation: "> перенаправляет вывод echo в файл, перезаписывая содержимое. Для дописывания используется >>.",
					},
				},
				Tasks: []T{
					{
						Title:        "Освойся: pwd, ls и переход",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>Осмотрись в системе и перейди в каталог временных файлов:</p>
<ol>
<li>Выведи текущий каталог командой <code>pwd</code>.</li>
<li>Посмотри содержимое <code>/etc</code> в длинном формате: <code>ls -l /etc</code>.</li>
<li>Перейди в каталог <code>/tmp</code> командой <code>cd /tmp</code>.</li>
</ol>
<p>Задание считается выполненным, когда твой текущий каталог — <code>/tmp</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "pwd", Definition: "печатает абсолютный путь текущего каталога"},
							{Term: "ls -l", Definition: "список файлов в длинном формате (права, владелец, размер)"},
							{Term: "cd /tmp", Definition: "перейти в каталог /tmp"},
						},
						CheckScript: `[ "$PWD" = /tmp ]`,
						Solution:    `<pre><code>pwd
ls -l /etc
cd /tmp</code></pre>`,
					},
					{
						Title:        "Создай структуру каталогов и файл",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>В домашнем каталоге (<code>~</code>) создай:</p>
<ol>
<li>каталог <code>workspace</code>;</li>
<li>внутри него файл <code>readme.txt</code> со словом <code>hello</code>.</li>
</ol>
<p>Подсказка: <code>mkdir ~/workspace</code>, затем <code>echo "hello" &gt; ~/workspace/readme.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "mkdir", Definition: "создать каталог"},
							{Term: "echo текст > файл", Definition: "записать строку в файл"},
							{Term: "~", Definition: "домашний каталог пользователя"},
						},
						CheckScript: `[ -d "$HOME/workspace" ] && grep -q "hello" "$HOME/workspace/readme.txt" 2>/dev/null`,
						Solution:    `<pre><code>mkdir ~/workspace
echo "hello" &gt; ~/workspace/readme.txt</code></pre>`,
					},
					{
						Title:        "Символическая ссылка",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>В домашнем каталоге создай файл и ссылку на него:</p>
<ol>
<li>Создай файл <code>data.txt</code> с текстом <code>GoLearn</code>.</li>
<li>Создай <strong>символическую</strong> ссылку <code>latest.txt</code>, указывающую на <code>data.txt</code> (<code>ln -s</code>).</li>
</ol>
<p>Проверь себя: <code>cat latest.txt</code> должен вывести <code>GoLearn</code>, а <code>ls -l</code> — показать <code>latest.txt -&gt; data.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "ln -s цель имя", Definition: "создать символическую ссылку (ярлык на путь)"},
							{Term: "readlink", Definition: "показать, куда ведёт ссылка"},
							{Term: "cat", Definition: "вывести содержимое файла"},
						},
						SetupScript: `cd /root; rm -f data.txt latest.txt`,
						CheckScript: `cd "$HOME"; [ -L latest.txt ] && [ "$(cat latest.txt 2>/dev/null)" = "GoLearn" ] && grep -q "GoLearn" data.txt 2>/dev/null`,
						Solution:    `<pre><code>cd ~
echo "GoLearn" &gt; data.txt
ln -s data.txt latest.txt
cat latest.txt   # GoLearn
ls -l            # latest.txt -&gt; data.txt</code></pre>`,
					},
				},
			},
			{
				Slug:       "reading-and-search",
				Title:      "Чтение файлов и поиск",
				Track:      "devops",
				Difficulty: "beginner",
				Content: `<h1>Читаем файлы</h1>
<p>В терминале нет «двойного клика» — содержимое файлов смотрят командами.</p>
<ul>
<li><code>cat файл</code> — вывести файл целиком (хорошо для коротких).</li>
<li><code>less файл</code> — постраничный просмотр (выход — <code>q</code>, поиск — <code>/слово</code>). Не грузит файл целиком в память — годится для гигабайтных логов.</li>
<li><code>head -n 5 файл</code> — первые 5 строк; <code>tail -n 5 файл</code> — последние 5.</li>
<li><code>tail -f файл</code> — «следить» за файлом в реальном времени (логи).</li>
<li><code>wc -l файл</code> — посчитать строки (<code>-w</code> — слова, <code>-c</code> — байты).</li>
</ul>

<h2>grep — поиск по содержимому</h2>
<p><code>grep шаблон файл</code> печатает строки, где встретился шаблон. Это рабочая лошадка администратора.</p>
<ul>
<li><code>-i</code> — игнорировать регистр;</li>
<li><code>-n</code> — показывать номера строк;</li>
<li><code>-c</code> — только количество совпадений;</li>
<li><code>-v</code> — инвертировать (строки <em>без</em> совпадения);</li>
<li><code>-r</code> — рекурсивно по каталогу.</li>
</ul>
<blockquote>Пример: <code>grep -in error /var/log/app.log</code> — найти все упоминания error (в любом регистре) с номерами строк.</blockquote>

<h2>find — поиск файлов</h2>
<p><code>grep</code> ищет <em>внутри</em> файлов, а <code>find</code> ищет <em>сами файлы</em> по имени, типу, размеру, времени.</p>
<ul>
<li><code>find /path -name "*.log"</code> — по маске имени;</li>
<li><code>find /path -type f</code> — только файлы (<code>-type d</code> — каталоги);</li>
<li><code>find /path -name "*.conf" -type f</code> — комбинация условий.</li>
</ul>
<p>Результаты find можно перенаправить в файл или передать дальше по конвейеру (об этом — в следующем уроке).</p>`,
				Quiz: []Q{
					{
						Question:    "Чем grep отличается от find?",
						Options:     []string{"Ничем", "grep ищет строки внутри файлов; find ищет сами файлы по имени/типу", "find быстрее grep", "grep работает только с логами"},
						Correct:     1,
						Explanation: "grep смотрит содержимое файлов и выводит совпавшие строки; find находит файлы в дереве каталогов по критериям (имя, тип, размер...).",
					},
					{
						Question:    "Что выведет grep -c ERROR app.log?",
						Options:     []string{"Сами строки с ERROR", "Количество строк, содержащих ERROR", "Строки без ERROR", "Номера строк"},
						Correct:     1,
						Explanation: "-c (count) выводит только число совпавших строк, а не сами строки.",
					},
					{
						Question:    "Почему для большого лога лучше less, а не cat?",
						Options:     []string{"less красивее", "less листает постранично и не грузит весь файл сразу", "cat не умеет читать логи", "разницы нет"},
						Correct:     1,
						Explanation: "less читает файл по мере прокрутки (постранично), поэтому открывает даже огромные файлы мгновенно; cat выливает всё содержимое в терминал разом.",
					},
				},
				Tasks: []T{
					{
						Title:        "Отфильтруй строки с grep",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>В файле <code>/root/log.txt</code> уже лежат логи. Выбери из них только строки со словом <code>ERROR</code> и сохрани их в <code>/root/errors.txt</code>.</p>
<p>Подсказка: <code>grep ERROR /root/log.txt &gt; /root/errors.txt</code>. Посмотреть исходник — <code>cat /root/log.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "grep шаблон файл", Definition: "вывести строки, содержащие шаблон"},
							{Term: "> файл", Definition: "перенаправить вывод в файл (перезаписав)"},
						},
						SetupScript: "printf 'INFO start\\nERROR disk full\\nINFO ok\\nERROR timeout\\nWARN low mem\\n' > /root/log.txt; rm -f /root/errors.txt",
						CheckScript: `[ -f /root/errors.txt ] && [ "$(grep -c ERROR /root/errors.txt)" = 2 ] && [ "$(wc -l < /root/errors.txt)" = 2 ]`,
						Solution:    "<pre><code>grep ERROR /root/log.txt &gt; /root/errors.txt</code></pre>",
					},
					{
						Title:        "Найди файлы через find",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>Внутри <code>/root/proj</code> лежат разные файлы. Найди все файлы с расширением <code>.md</code> и сохрани их пути в <code>/root/md.txt</code>.</p>
<p>Подсказка: <code>find /root/proj -name "*.md" &gt; /root/md.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: `find путь -name "*.md"`, Definition: "найти файлы по маске имени"},
						},
						SetupScript: "mkdir -p /root/proj/a /root/proj/b; touch /root/proj/a/notes.md /root/proj/b/readme.md /root/proj/a/data.txt; rm -f /root/md.txt",
						CheckScript: `[ -f /root/md.txt ] && [ "$(grep -c '\.md' /root/md.txt)" = 2 ]`,
						Solution:    "<pre><code>find /root/proj -name \"*.md\" &gt; /root/md.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "permissions",
				Title:      "Права доступа: rwx, chmod, chown",
				Track:      "devops",
				Difficulty: "beginner",
				Content: `<h1>Кто что может</h1>
<p>У каждого файла в Linux есть владелец, группа и набор прав для трёх категорий: <strong>владелец (user)</strong>, <strong>группа (group)</strong>, <strong>остальные (other)</strong>. Для каждой категории — три права:</p>
<ul>
<li><code>r</code> (read) — читать файл / смотреть список каталога;</li>
<li><code>w</code> (write) — изменять файл / создавать-удалять в каталоге;</li>
<li><code>x</code> (execute) — запускать файл как программу / <em>входить</em> в каталог (cd).</li>
</ul>

<h2>Читаем ls -l</h2>
<p>Строка вида <code>-rwxr-xr--</code> расшифровывается так:</p>
<ul>
<li>первый символ — тип (<code>-</code> файл, <code>d</code> каталог, <code>l</code> ссылка);</li>
<li>далее три триплета: <code>rwx</code> (владелец), <code>r-x</code> (группа), <code>r--</code> (остальные).</li>
</ul>

<h2>Числовая запись</h2>
<p>Каждое право — это бит: <code>r=4</code>, <code>w=2</code>, <code>x=1</code>. Сложив их в каждом триплете, получаем цифру:</p>
<ul>
<li><code>7</code> = 4+2+1 = rwx;</li>
<li><code>6</code> = 4+2 = rw-;</li>
<li><code>5</code> = 4+1 = r-x.</li>
</ul>
<p>Поэтому <code>755</code> = rwxr-xr-x (типично для программ и каталогов), <code>644</code> = rw-r--r-- (обычный файл), <code>600</code> = rw------- (приватный — например, секрет или SSH-ключ).</p>

<h2>Меняем права — chmod</h2>
<ul>
<li>Числом: <code>chmod 600 secret.txt</code>.</li>
<li>Символьно: <code>chmod +x script.sh</code> (добавить право запуска всем), <code>chmod u+w,go-w файл</code>.</li>
</ul>

<h2>Меняем владельца — chown</h2>
<p><code>chown user:group файл</code> меняет владельца и группу (обычно нужно root). <code>ls -l</code> покажет, кому файл принадлежит.</p>

<blockquote>Правило безопасности: приватные файлы (ключи, секреты) держат в <code>600</code> — иначе ssh откажется использовать ключ «UNPROTECTED PRIVATE KEY».</blockquote>`,
				Quiz: []Q{
					{
						Question:    "Что означает число 600 в правах?",
						Options:     []string{"rwxr-x---", "rw------- (владелец читает/пишет, остальным ничего)", "r--r--r--", "rwxrwxrwx"},
						Correct:     1,
						Explanation: "6=rw для владельца, 0 для группы, 0 для остальных. Это типичные права для приватного файла.",
					},
					{
						Question:    "Что делает chmod +x script.sh?",
						Options:     []string{"Удаляет файл", "Делает файл исполняемым (добавляет право x)", "Меняет владельца", "Шифрует файл"},
						Correct:     1,
						Explanation: "+x добавляет право на выполнение, после чего файл можно запустить как программу (./script.sh).",
					},
					{
						Question:    "Право x на КАТАЛОГ означает...",
						Options:     []string{"Можно удалить каталог", "Можно войти в каталог (cd) и обращаться к его содержимому", "Можно переименовать", "Ничего"},
						Correct:     1,
						Explanation: "Для каталога x — это право «входа»: без него нельзя сделать cd внутрь и обратиться к файлам по пути.",
					},
				},
				Tasks: []T{
					{
						Title:        "Сделай скрипт исполняемым",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>Создай файл <code>/root/run.sh</code> (любое содержимое) и сделай его исполняемым с помощью <code>chmod +x</code>.</p>
<p>Проверь результат: <code>ls -l /root/run.sh</code> — в правах должна появиться <code>x</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "chmod +x файл", Definition: "добавить право на выполнение"},
							{Term: "ls -l", Definition: "показать права файла"},
						},
						SetupScript: "rm -f /root/run.sh",
						CheckScript: `[ -f /root/run.sh ] && [ -x /root/run.sh ]`,
						Solution:    "<pre><code>echo '#!/bin/bash' &gt; /root/run.sh\nchmod +x /root/run.sh</code></pre>",
					},
					{
						Title:        "Закрой секрет правами 600",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>Файл <code>/root/secret.txt</code> сейчас доступен на чтение всем (644). Установи на него права <code>600</code>, чтобы читать/писать мог только владелец.</p>
<p>Подсказка: <code>chmod 600 /root/secret.txt</code>. Проверь: <code>ls -l /root/secret.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "chmod 600 файл", Definition: "rw для владельца, никаких прав остальным"},
							{Term: "stat -c '%a' файл", Definition: "показать права в числовом виде"},
						},
						SetupScript: "echo secret > /root/secret.txt; chmod 644 /root/secret.txt",
						CheckScript: `[ "$(stat -c '%a' /root/secret.txt)" = 600 ]`,
						Solution:    "<pre><code>chmod 600 /root/secret.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "redirection-and-pipes",
				Title:      "Перенаправление и конвейеры",
				Track:      "devops",
				Difficulty: "intermediate",
				Content: `<h1>Три потока</h1>
<p>У каждой программы есть три стандартных потока, у каждого свой номер (файловый дескриптор):</p>
<ul>
<li><strong>stdin</strong> (0) — ввод;</li>
<li><strong>stdout</strong> (1) — обычный вывод;</li>
<li><strong>stderr</strong> (2) — поток ошибок.</li>
</ul>
<p>Разделение stdout и stderr — важная идея: «полезный» результат и сообщения об ошибках можно направить в разные места.</p>

<h2>Перенаправление в файл</h2>
<ul>
<li><code>команда &gt; файл</code> — записать stdout в файл (перезаписав);</li>
<li><code>команда &gt;&gt; файл</code> — дописать в конец;</li>
<li><code>команда 2&gt; ошибки.txt</code> — записать только stderr;</li>
<li><code>команда &gt; out.txt 2&gt;&amp;1</code> — и stdout, и stderr в один файл.</li>
</ul>

<h2>Конвейеры (pipes)</h2>
<p>Символ <code>|</code> соединяет программы: stdout одной становится stdin следующей. Так из простых утилит собирают «обработку данных»:</p>
<pre><code>cat access.log | grep 404 | wc -l</code></pre>
<p>Читается так: взять лог → оставить строки с 404 → посчитать их количество. Каждый шаг делает одну вещь, а вместе они решают задачу — это и есть философия Unix.</p>

<h2>Полезное</h2>
<ul>
<li><code>tee файл</code> — раздвоить поток: и на экран, и в файл (<code>... | tee out.txt</code>);</li>
<li><code>cmd1 &amp;&amp; cmd2</code> — выполнить cmd2 только если cmd1 успешна;</li>
<li><code>cmd1 || cmd2</code> — выполнить cmd2 только если cmd1 провалилась.</li>
</ul>`,
				Quiz: []Q{
					{
						Question:    "Чем > отличается от >>?",
						Options:     []string{"Ничем", "> перезаписывает файл, >> дописывает в конец", ">> быстрее", "> только для текста"},
						Correct:     1,
						Explanation: "> создаёт/перезаписывает файл выводом команды; >> добавляет вывод в конец существующего файла.",
					},
					{
						Question:    "Что делает конвейер ls | grep .conf | wc -l?",
						Options:     []string{"Удаляет .conf файлы", "Считает, сколько имён в выводе ls содержат .conf", "Создаёт файл", "Открывает редактор"},
						Correct:     1,
						Explanation: "ls выдаёт список → grep оставляет строки с .conf → wc -l считает их количество. stdout каждого шага идёт в stdin следующего.",
					},
					{
						Question:    "Как направить ТОЛЬКО ошибки команды в файл?",
						Options:     []string{"cmd > err.txt", "cmd 2> err.txt", "cmd >> err.txt", "cmd | err.txt"},
						Correct:     1,
						Explanation: "stderr — это дескриптор 2, поэтому 2> перенаправляет именно поток ошибок.",
					},
				},
				Tasks: []T{
					{
						Title:        "Перенаправление: > и >>",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>Сделай два шага:</p>
<ol>
<li>Запиши список файлов каталога <code>/etc</code> в <code>/root/etc.txt</code> (<code>ls /etc &gt; /root/etc.txt</code>).</li>
<li>Допиши в конец этого файла строку <code>END</code> (<code>echo END &gt;&gt; /root/etc.txt</code>).</li>
</ol>
<p>Последняя строка файла должна быть ровно <code>END</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "> файл", Definition: "перенаправить вывод, перезаписав файл"},
							{Term: ">> файл", Definition: "дописать вывод в конец файла"},
						},
						SetupScript: "rm -f /root/etc.txt",
						CheckScript: `[ -f /root/etc.txt ] && [ "$(tail -n1 /root/etc.txt)" = "END" ] && [ "$(wc -l < /root/etc.txt)" -gt 1 ]`,
						Solution:    "<pre><code>ls /etc &gt; /root/etc.txt\necho END &gt;&gt; /root/etc.txt</code></pre>",
					},
					{
						Title:        "Конвейер: посчитай .conf",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>В каталоге <code>/root/cfg</code> лежат файлы. Собери конвейер, который посчитает, сколько из них имеют расширение <code>.conf</code>, и сохрани это число в <code>/root/n.txt</code>.</p>
<p>Подсказка: <code>ls /root/cfg | grep '\.conf$' | wc -l &gt; /root/n.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "|", Definition: "конвейер: stdout одной команды → stdin следующей"},
							{Term: "wc -l", Definition: "посчитать число строк"},
						},
						SetupScript: "mkdir -p /root/cfg; touch /root/cfg/a.conf /root/cfg/b.conf /root/cfg/c.txt; rm -f /root/n.txt",
						CheckScript: `[ "$(tr -d ' \n' < /root/n.txt 2>/dev/null)" = 2 ]`,
						Solution:    "<pre><code>ls /root/cfg | grep '\\.conf$' | wc -l &gt; /root/n.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "text-processing",
				Title:      "Текстовая обработка: sort, uniq, cut, sed",
				Track:      "devops",
				Difficulty: "intermediate",
				Content: `<h1>Текст — это данные</h1>
<p>В Unix-философии почти всё — текст, и для его обработки есть набор маленьких острых инструментов. Соединяя их конвейером (<code>|</code>), решают сложные задачи без программирования.</p>

<h2>sort — сортировка</h2>
<ul>
<li><code>sort файл</code> — по алфавиту;</li>
<li><code>sort -n</code> — как числа (иначе «10» окажется раньше «2»);</li>
<li><code>sort -r</code> — в обратном порядке.</li>
</ul>

<h2>uniq — убрать дубли</h2>
<p><code>uniq</code> удаляет <em>соседние</em> повторы, поэтому почти всегда идёт после <code>sort</code>: <code>sort файл | uniq</code>. Флаг <code>-c</code> добавит счётчик повторений.</p>

<h2>cut — вырезать колонки</h2>
<p>Когда строки разбиты разделителем, <code>cut</code> достаёт нужные поля: <code>cut -d: -f1 /etc/passwd</code> — взять первое поле (имя пользователя), где разделитель — двоеточие. <code>-d</code> задаёт разделитель, <code>-f</code> — номера полей.</p>

<h2>sed — потоковый редактор</h2>
<p><code>sed 's/старое/новое/g' файл</code> — заменить текст «на лету», не открывая редактор. Полезно в скриптах и конвейерах. Буква <code>g</code> в конце — заменять все вхождения в строке, а не только первое.</p>

<h2>awk — мини-язык для колонок</h2>
<p><code>awk '{print $2}'</code> печатает второе поле каждой строки (поля по умолчанию разделены пробелами). awk умеет куда больше, но даже <code>print $N</code> уже часто выручает.</p>

<blockquote>Сила в комбинации: <code>cut -d: -f1 /etc/passwd | sort | uniq</code> — отсортированный список уникальных пользователей.</blockquote>`,
				Quiz: []Q{
					{
						Question:    "Почему uniq обычно ставят после sort?",
						Options:     []string{"Так быстрее", "uniq удаляет только соседние дубли, а sort ставит одинаковые строки рядом", "uniq не работает без sort вообще", "Это требование синтаксиса"},
						Correct:     1,
						Explanation: "uniq схлопывает только подряд идущие одинаковые строки. Чтобы убрать все дубли, их сначала надо сгруппировать сортировкой.",
					},
					{
						Question:    "Что сделает cut -d: -f1 /etc/passwd?",
						Options:     []string{"Удалит первую строку", "Выведет первое поле (до первого :) каждой строки", "Отсортирует файл", "Заменит : на пробел"},
						Correct:     1,
						Explanation: "-d: задаёт разделитель «двоеточие», -f1 выбирает первое поле — это имена пользователей в /etc/passwd.",
					},
					{
						Question:    "Зачем нужен -n у sort?",
						Options:     []string{"Не выводить ничего", "Сортировать как числа, а не как строки", "Нумеровать строки", "Сортировать в обратном порядке"},
						Correct:     1,
						Explanation: "Без -n сортировка идёт лексикографически, и «10» оказывается раньше «2». -n сравнивает значения как числа.",
					},
				},
				Tasks: []T{
					{
						Title:        "Уникальные значения: sort | uniq",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>В файле <code>/root/names.txt</code> есть повторяющиеся имена. Получи отсортированный список <strong>без дублей</strong> и сохрани его в <code>/root/unique.txt</code>.</p>
<p>Подсказка: <code>sort /root/names.txt | uniq &gt; /root/unique.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "sort", Definition: "отсортировать строки"},
							{Term: "uniq", Definition: "убрать соседние повторы (после sort)"},
						},
						SetupScript: "printf 'bob\\nalice\\nbob\\ncarol\\nalice\\n' > /root/names.txt; rm -f /root/unique.txt",
						CheckScript: `[ -f /root/unique.txt ] && [ "$(wc -l < /root/unique.txt)" = 3 ] && [ "$(head -1 /root/unique.txt)" = alice ]`,
						Solution:    "<pre><code>sort /root/names.txt | uniq &gt; /root/unique.txt</code></pre>",
					},
					{
						Title:        "Вырежи колонку: cut",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>В <code>/root/users.csv</code> строки вида <code>имя:uid</code>. Извлеки только имена (первое поле, разделитель <code>:</code>) и сохрани в <code>/root/just_names.txt</code>.</p>
<p>Подсказка: <code>cut -d: -f1 /root/users.csv &gt; /root/just_names.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "cut -d: -f1", Definition: "взять первое поле, разделитель — двоеточие"},
						},
						SetupScript: "printf 'alice:1001\\nbob:1002\\ncarol:1003\\n' > /root/users.csv; rm -f /root/just_names.txt",
						CheckScript: `[ -f /root/just_names.txt ] && [ "$(wc -l < /root/just_names.txt)" = 3 ] && grep -qx alice /root/just_names.txt && ! grep -q ':' /root/just_names.txt`,
						Solution:    "<pre><code>cut -d: -f1 /root/users.csv &gt; /root/just_names.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "env-variables",
				Title:      "Переменные окружения и PATH",
				Track:      "devops",
				Difficulty: "intermediate",
				Content: `<h1>Переменные окружения</h1>
<p>Окружение — это набор пар «имя=значение», которые видны процессам. Через них настраивают поведение программ, не меняя их код.</p>

<h2>Создание и чтение</h2>
<ul>
<li><code>VAR=значение</code> — задать переменную (в текущей оболочке);</li>
<li><code>echo $VAR</code> — прочитать значение;</li>
<li><code>export VAR=значение</code> — сделать переменную видимой <em>дочерним</em> процессам;</li>
<li><code>env</code> — показать все переменные окружения.</li>
</ul>
<p>Без <code>export</code> переменная остаётся «локальной» для оболочки и не передаётся запускаемым из неё программам.</p>

<h2>Важные системные переменные</h2>
<ul>
<li><code>$HOME</code> — домашний каталог;</li>
<li><code>$USER</code> — имя текущего пользователя;</li>
<li><code>$PWD</code> — текущий каталог;</li>
<li><code>$PATH</code> — список каталогов, где оболочка ищет исполняемые команды.</li>
</ul>

<h2>PATH — как находятся команды</h2>
<p>Когда ты пишешь <code>ls</code>, оболочка по очереди ищет файл <code>ls</code> в каталогах из <code>$PATH</code> (например <code>/usr/bin:/bin</code>) и запускает первый найденный. Поэтому свои программы кладут в каталог из PATH или дополняют PATH: <code>export PATH="$HOME/bin:$PATH"</code>.</p>

<blockquote>Постоянство: переменные, заданные в терминале, живут только до закрытия сессии. Чтобы они появлялись всегда, их прописывают в <code>~/.bashrc</code> или <code>~/.profile</code>.</blockquote>`,
				Quiz: []Q{
					{
						Question:    "Чем отличается VAR=x от export VAR=x?",
						Options:     []string{"Ничем", "export делает переменную видимой дочерним процессам; без него она только в текущей оболочке", "export сохраняет навсегда", "VAR=x работает только для чисел"},
						Correct:     1,
						Explanation: "export помечает переменную для передачи в окружение запускаемых программ. Без export её увидит только сама оболочка.",
					},
					{
						Question:    "Что хранит переменная PATH?",
						Options:     []string{"Текущий каталог", "Список каталогов, где оболочка ищет исполняемые команды", "Историю команд", "Домашний каталог"},
						Correct:     1,
						Explanation: "PATH — это разделённый двоеточиями список каталогов; в них оболочка ищет файл команды, которую ты ввёл.",
					},
					{
						Question:    "Где прописать переменную, чтобы она была всегда?",
						Options:     []string{"Нигде, это невозможно", "В ~/.bashrc или ~/.profile", "В /tmp", "В PATH"},
						Correct:     1,
						Explanation: "Файлы вроде ~/.bashrc выполняются при старте оболочки, поэтому заданные там переменные появляются в каждой новой сессии.",
					},
				},
				Tasks: []T{
					{
						Title:        "Прочитай переменную окружения",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "easy",
						Description: `<p>Запиши значение переменной <code>$HOME</code> в файл <code>/root/home.txt</code>.</p>
<p>Подсказка: <code>echo $HOME &gt; /root/home.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "echo $HOME", Definition: "вывести значение переменной HOME"},
							{Term: "$HOME", Definition: "домашний каталог пользователя"},
						},
						SetupScript: "rm -f /root/home.txt",
						CheckScript: `[ "$(cat /root/home.txt 2>/dev/null)" = "/root" ]`,
						Solution:    "<pre><code>echo $HOME &gt; /root/home.txt</code></pre>",
					},
					{
						Title:        "Своя переменная",
						Kind:         "shell",
						SandboxImage: "ubuntu:24.04",
						Difficulty:   "medium",
						Description: `<p>В одной команде: создай переменную <code>GREETING</code> со значением <code>hello</code> и выведи её значение в файл <code>/root/greet.txt</code>.</p>
<p>Подсказка: <code>GREETING=hello; echo $GREETING &gt; /root/greet.txt</code> (в одной строке, потому что переменная живёт только в пределах команды).</p>`,
						Glossary: []GlossaryItem{
							{Term: "VAR=значение", Definition: "задать переменную"},
							{Term: "; ", Definition: "разделить несколько команд в одной строке"},
						},
						SetupScript: "rm -f /root/greet.txt",
						CheckScript: `[ "$(cat /root/greet.txt 2>/dev/null)" = "hello" ]`,
						Solution:    "<pre><code>GREETING=hello; echo $GREETING &gt; /root/greet.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "processes-and-signals",
				Title:      "Процессы и сигналы",
				Track:      "devops",
				Difficulty: "intermediate",
				Content: `<h1>Что такое процесс</h1>
<p>Процесс — это запущенная программа со своим адресным пространством, состоянием и уникальным номером <strong>PID</strong> (process id). У каждого процесса есть родитель (<strong>PPID</strong>); всё дерево растёт из процесса с PID 1 (init).</p>

<h2>Смотрим процессы — ps</h2>
<ul>
<li><code>ps aux</code> — все процессы в системе (пользователь, PID, %CPU, %MEM, команда);</li>
<li><code>ps -ef</code> — то же в другом формате, с PPID;</li>
<li><code>ps -p 1 -o comm=</code> — имя конкретного процесса по PID.</li>
</ul>
<p>Чтобы найти нужный процесс, выводы <code>ps</code> часто фильтруют через grep: <code>ps aux | grep nginx</code>.</p>

<h2>Фоновые процессы</h2>
<p>Символ <code>&amp;</code> в конце команды запускает её в фоне: <code>sleep 100 &amp;</code> — оболочка не ждёт завершения и сразу даёт ввести следующую команду. <code>jobs</code> покажет фоновые задачи текущей оболочки.</p>

<h2>Сигналы — как «разговаривают» с процессами</h2>
<p>Процессу нельзя «нажать крестик» — ему посылают <strong>сигнал</strong>. Главные:</p>
<ul>
<li><code>SIGTERM</code> (15) — вежливая просьба завершиться (по умолчанию у <code>kill</code>); процесс может корректно закрыть файлы и выйти;</li>
<li><code>SIGKILL</code> (9) — немедленное убийство без шанса на уборку (<code>kill -9</code>); крайняя мера;</li>
<li><code>SIGHUP</code> (1) — «терминал закрылся».</li>
</ul>
<p>Команды: <code>kill PID</code> (SIGTERM), <code>kill -9 PID</code> (SIGKILL). По имени удобнее <code>pkill имя</code> или найти PID через <code>pgrep имя</code>.</p>

<blockquote>Правило: сначала <code>kill</code> (SIGTERM) — дай процессу закрыться чисто; <code>kill -9</code> — только если завис и не реагирует.</blockquote>`,
				Quiz: []Q{
					{
						Question:    "Чем SIGKILL (9) отличается от SIGTERM (15)?",
						Options:     []string{"Ничем", "SIGTERM просит завершиться корректно; SIGKILL убивает мгновенно без уборки", "SIGKILL вежливее", "SIGTERM работает только от root"},
						Correct:     1,
						Explanation: "SIGTERM можно перехватить и красиво завершиться; SIGKILL не перехватывается и обрывает процесс немедленно — поэтому это крайняя мера.",
					},
					{
						Question:    "Что делает sleep 100 &?",
						Options:     []string{"Спит 100 минут", "Запускает sleep в фоне, не блокируя оболочку", "Удаляет процесс", "Повышает приоритет"},
						Correct:     1,
						Explanation: "& запускает команду в фоновом режиме: оболочка сразу освобождается для следующих команд.",
					},
					{
						Question:    "Как найти PID процесса по имени?",
						Options:     []string{"kill имя", "pgrep имя (или ps aux | grep имя)", "ls имя", "cd имя"},
						Correct:     1,
						Explanation: "pgrep ищет процессы по имени и печатает их PID; альтернатива — отфильтровать вывод ps через grep.",
					},
				},
				Tasks: []T{
					{
						Title:        "Найди и заверши процесс",
						Kind:         "shell",
						SandboxImage: "golearn/linux:latest",
						Difficulty:   "medium",
						Description: `<p>В песочнице уже работает фоновый процесс <code>sleep 100000</code>. Найди его PID и заверши его.</p>
<ol>
<li>Посмотри процессы: <code>ps aux | grep sleep</code> (или <code>pgrep -af sleep</code>).</li>
<li>Заверши нужный: <code>kill &lt;PID&gt;</code>.</li>
</ol>
<p>Задание зачтётся, когда процесс <code>sleep 100000</code> перестанет существовать.</p>`,
						Glossary: []GlossaryItem{
							{Term: "ps aux | grep sleep", Definition: "найти процесс по имени"},
							{Term: "pgrep -af sleep", Definition: "вывести PID и команду процессов sleep"},
							{Term: "kill PID", Definition: "послать процессу сигнал завершения (SIGTERM)"},
						},
						SetupScript: "pkill -f 'sleep 100000' 2>/dev/null; nohup sleep 100000 >/dev/null 2>&1 & disown; sleep 0.3",
						CheckScript: `! pgrep -f 'sleep 100000' >/dev/null 2>&1`,
						Solution:    "<pre><code>pgrep -af sleep        # узнать PID\nkill &lt;PID&gt;             # заменить на реальный PID\n# или одной командой:\npkill -f 'sleep 100000'</code></pre>",
					},
					{
						Title:        "Кто такой PID 1",
						Kind:         "shell",
						SandboxImage: "golearn/linux:latest",
						Difficulty:   "easy",
						Description: `<p>У главного процесса контейнера всегда PID 1. Выясни <strong>имя</strong> этого процесса и запиши его в <code>/root/pid1.txt</code>.</p>
<p>Подсказка: <code>ps -p 1 -o comm= &gt; /root/pid1.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "ps -p 1", Definition: "показать процесс с PID 1"},
							{Term: "-o comm=", Definition: "вывести только имя команды, без заголовка"},
						},
						SetupScript: "rm -f /root/pid1.txt",
						CheckScript: `[ -s /root/pid1.txt ] && grep -qi sleep /root/pid1.txt`,
						Solution:    "<pre><code>ps -p 1 -o comm= &gt; /root/pid1.txt</code></pre>",
					},
				},
			},
			{
				Slug:       "users-groups-sudo",
				Title:      "Пользователи, группы и sudo",
				Track:      "devops",
				Difficulty: "intermediate",
				Content: `<h1>Кто есть кто в системе</h1>
<p>Linux — многопользовательская система. Каждый пользователь — это запись с числовым идентификатором <strong>UID</strong>. Особый пользователь — <strong>root</strong> (UID 0): ему можно всё. Остальным — только то, что разрешено правами.</p>

<h2>Где живут пользователи</h2>
<p>Список пользователей — в файле <code>/etc/passwd</code>, по строке на каждого, поля через <code>:</code>:</p>
<pre><code>alice:x:1001:1001:Alice:/home/alice:/bin/bash</code></pre>
<p>Это: имя, заглушка пароля (<code>x</code> — настоящий хеш в <code>/etc/shadow</code>), UID, GID основной группы, описание, домашний каталог, оболочка входа.</p>

<h2>Узнать про себя</h2>
<ul>
<li><code>whoami</code> — имя текущего пользователя;</li>
<li><code>id</code> — UID, GID и группы (<code>id -u</code> — только UID);</li>
<li><code>groups</code> — список групп текущего пользователя.</li>
</ul>

<h2>Группы</h2>
<p>Группа объединяет пользователей, чтобы давать права сразу нескольким. У файла есть владелец и группа-владелец; права <code>rwx</code> задаются отдельно для владельца, группы и остальных (вспомни урок про права). Список групп — в <code>/etc/group</code>.</p>

<h2>Создание и изменение</h2>
<ul>
<li><code>useradd alice</code> — создать пользователя (обычно нужно root);</li>
<li><code>userdel alice</code> — удалить;</li>
<li><code>chown пользователь файл</code> — сменить владельца;</li>
<li><code>chgrp группа файл</code> (или <code>chown :группа файл</code>) — сменить группу-владельца.</li>
</ul>

<h2>su и sudo</h2>
<p>Работать под root постоянно — опасно (одна ошибка и снёс систему). Поэтому действует <strong>принцип наименьших привилегий</strong>: живёшь под обычным пользователем, а отдельные команды выполняешь с повышением прав через <code>sudo команда</code>. Кому что можно — описано в <code>/etc/sudoers</code>. <code>su</code> же полностью переключает на другого пользователя.</p>

<blockquote>Хорошая привычка: не <code>su root</code> на весь сеанс, а <code>sudo</code> точечно — так меньше шансов случайно навредить.</blockquote>`,
				Quiz: []Q{
					{
						Question:    "Какой UID у пользователя root?",
						Options:     []string{"1", "1000", "0", "Любой"},
						Correct:     2,
						Explanation: "root всегда имеет UID 0 — по этому признаку ядро и разрешает ему всё.",
					},
					{
						Question:    "Что показывает команда id?",
						Options:     []string{"IP-адрес", "UID, GID и группы текущего пользователя", "Список файлов", "Версию системы"},
						Correct:     1,
						Explanation: "id печатает числовые идентификаторы пользователя (uid), его основной группы (gid) и все группы, в которых он состоит.",
					},
					{
						Question:    "Почему sudo предпочтительнее, чем постоянно работать под root?",
						Options:     []string{"sudo быстрее", "Принцип наименьших привилегий: повышаешь права точечно, меньше риск случайно навредить", "root не умеет ставить пакеты", "Разницы нет"},
						Correct:     1,
						Explanation: "Под root любая ошибка фатальна для всей системы. sudo даёт права только на конкретную команду, ограничивая зону поражения.",
					},
				},
				Tasks: []T{
					{
						Title:        "Кто я в системе",
						Kind:         "shell",
						SandboxImage: "golearn/linux:latest",
						Difficulty:   "easy",
						Description: `<p>Запиши в файл <code>/root/me.txt</code> строку вида <code>имя:uid</code> для текущего пользователя.</p>
<p>Подсказка: <code>echo "$(whoami):$(id -u)" &gt; /root/me.txt</code>. В песочнице ты root, так что ждём <code>root:0</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "whoami", Definition: "имя текущего пользователя"},
							{Term: "id -u", Definition: "числовой UID текущего пользователя"},
						},
						SetupScript: "rm -f /root/me.txt",
						CheckScript: `[ "$(cat /root/me.txt 2>/dev/null)" = "root:0" ]`,
						Solution:    "<pre><code>echo \"$(whoami):$(id -u)\" &gt; /root/me.txt</code></pre>",
					},
					{
						Title:        "Создай пользователя",
						Kind:         "shell",
						SandboxImage: "golearn/linux:latest",
						Difficulty:   "medium",
						Description: `<p>Создай пользователя с именем <code>alice</code> командой <code>useradd</code>.</p>
<p>Проверь результат: <code>id alice</code> или <code>grep alice /etc/passwd</code> — должна появиться запись.</p>`,
						Glossary: []GlossaryItem{
							{Term: "useradd alice", Definition: "создать пользователя alice"},
							{Term: "id alice", Definition: "показать UID/группы пользователя"},
							{Term: "/etc/passwd", Definition: "файл со списком пользователей"},
						},
						SetupScript: "userdel -r alice 2>/dev/null; true",
						CheckScript: `id alice >/dev/null 2>&1 && grep -q '^alice:' /etc/passwd`,
						Solution:    "<pre><code>useradd alice\nid alice</code></pre>",
					},
					{
						Title:        "Смени группу-владельца файла",
						Kind:         "shell",
						SandboxImage: "golearn/linux:latest",
						Difficulty:   "medium",
						Description: `<p>У файла <code>/root/report.txt</code> сейчас группа-владелец <code>root</code>. Поменяй её на группу <code>daemon</code>.</p>
<p>Подсказка: <code>chgrp daemon /root/report.txt</code> (или <code>chown :daemon /root/report.txt</code>). Проверь: <code>ls -l /root/report.txt</code>.</p>`,
						Glossary: []GlossaryItem{
							{Term: "chgrp группа файл", Definition: "сменить группу-владельца"},
							{Term: "stat -c '%G' файл", Definition: "показать имя группы-владельца"},
						},
						SetupScript: "echo report > /root/report.txt; chgrp root /root/report.txt",
						CheckScript: `[ "$(stat -c '%G' /root/report.txt 2>/dev/null)" = daemon ]`,
						Solution:    "<pre><code>chgrp daemon /root/report.txt</code></pre>",
					},
				},
			},
		},
	}
}
