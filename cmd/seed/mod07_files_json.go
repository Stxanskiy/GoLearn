package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ 7: Работа с файлами и JSON (НОВЫЙ)
// ════════════════════════════════════════════════════════════════

func mod07_files_json() M {
	return M{
		Slug:          "files-json",
		Title:         "Файлы, I/O и JSON",
		Description:   "os, io, bufio, filepath, encoding/json — работа с файловой системой и сериализация данных.",
		Order:         7,
		Track:         "backend",
		Difficulty:    "intermediate",
		Prerequisites: []string{"errors"},
		Lessons: []L{
			{
				Slug: "files-io", Title: "Работа с файлами и потоками", Order: 1,
				Content: `<h1>Работа с файлами и потоками I/O</h1>

<h2>Под капотом: io.Reader и io.Writer</h2>
<p>Вся система ввода-вывода в Go построена на двух интерфейсах:</p>
<pre><code>type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}</code></pre>

<p><strong>Почему это гениально?</strong> Один интерфейс — и файл, и HTTP body, и сетевое соединение, и буфер в памяти, и gzip-поток — всё это Reader. Функция, написанная для Reader, работает с любым источником данных.</p>

<pre><code>// Все эти типы реализуют io.Reader:
os.File          // файл на диске
http.Response.Body // тело HTTP ответа
bytes.Buffer     // буфер в памяти
strings.NewReader("text") // строка как поток
gzip.NewReader(r) // распаковка gzip</code></pre>

<h2>Открытие и чтение файлов</h2>
<pre><code>// Прочитать весь файл (для маленьких файлов)
data, err := os.ReadFile("config.json")
if err != nil {
    return fmt.Errorf("чтение файла: %w", err)
}
fmt.Println(string(data))

// Открыть файл (для больших файлов и потоковой обработки)
f, err := os.Open("video.mp4") // только чтение
if err != nil {
    return err
}
defer f.Close() // ВСЕГДА defer Close!

// Информация о файле
info, err := f.Stat()
fmt.Println(info.Name())    // "video.mp4"
fmt.Println(info.Size())    // 2500000000
fmt.Println(info.IsDir())   // false
fmt.Println(info.ModTime()) // время модификации</code></pre>

<h2>Запись в файлы</h2>
<pre><code>// Простая запись (создаёт или перезаписывает)
err := os.WriteFile("output.txt", []byte("данные"), 0644)

// Создание файла для записи
f, err := os.Create("log.txt") // создаёт или обрезает
defer f.Close()
fmt.Fprintf(f, "Строка %d\n", 1)

// Открытие с флагами (append, создание)
f, err := os.OpenFile("log.txt",
    os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
defer f.Close()</code></pre>

<h2>Права доступа (0644, 0755)</h2>
<p>Число в восьмеричной системе: <code>owner-group-others</code></p>
<pre><code>0644 = rw-r--r--  → владелец: чтение+запись, остальные: чтение
0755 = rwxr-xr-x  → владелец: всё, остальные: чтение+выполнение
0600 = rw-------  → только владелец (для секретов!)</code></pre>

<h2>bufio — буферизованный ввод-вывод</h2>
<pre><code>// Чтение большого файла построчно (экономит память)
f, _ := os.Open("large.log")
defer f.Close()

scanner := bufio.NewScanner(f)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}
if err := scanner.Err(); err != nil {
    log.Fatal(err)
}</code></pre>

<p><strong>Зачем bufio?</strong> Без буфера каждый Read — системный вызов (дорого). bufio читает блоками по 4КБ и выдаёт данные из буфера. На 10000 строках — разница в 100 раз.</p>

<h2>filepath — работа с путями</h2>
<pre><code>import "path/filepath"

filepath.Join("videos", "comedy", "movie.mp4")  // "videos/comedy/movie.mp4"
filepath.Ext("movie.mp4")                        // ".mp4"
filepath.Base("/home/user/movie.mp4")            // "movie.mp4"
filepath.Dir("/home/user/movie.mp4")             // "/home/user"

// Обход дерева каталогов
filepath.WalkDir("/videos", func(path string, d fs.DirEntry, err error) error {
    if err != nil { return err }
    if !d.IsDir() {
        fmt.Println(path) // каждый файл
    }
    return nil
})</code></pre>

<h2>io.Copy — потоковое копирование</h2>
<pre><code>// Копирование файла (любого размера, без загрузки в память)
src, _ := os.Open("source.mp4")
defer src.Close()
dst, _ := os.Create("copy.mp4")
defer dst.Close()

written, err := io.Copy(dst, src) // потоковое копирование
fmt.Printf("Скопировано %d байт\n", written)</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА 1: забыть defer f.Close()
f, _ := os.Open("file.txt")
// ... если return до Close — файл утекает! (file descriptor leak)

// ОШИБКА 2: ReadFile для огромных файлов
data, _ := os.ReadFile("10gb-video.mp4") // 10ГБ в RAM!
// Используй потоковую обработку (io.Copy, bufio.Scanner)

// ОШИБКА 3: не проверять ошибку Close
defer f.Close() // ошибка записи может быть здесь!
// Для файлов на запись:
defer func() {
    if err := f.Close(); err != nil {
        log.Println("close error:", err)
    }
}()</code></pre>`,

				Quiz: []Q{
					{Question: "Почему io.Reader — один из важнейших интерфейсов Go?", Options: []string{"Он встроен в компилятор", "Один интерфейс для любого источника данных: файл, сеть, память, сжатие", "Он самый быстрый", "Он единственный интерфейс в Go"}, Correct: 1, Explanation: "io.Reader абстрагирует источник данных. Функция, работающая с Reader, автоматически работает с файлами, HTTP, буферами, архивами — без изменений кода."},
					{Question: "Зачем нужен bufio при чтении файлов?", Options: []string{"Для красивого вывода", "Буферизация уменьшает количество системных вызовов в десятки раз", "Для сжатия", "Не нужен — os.File достаточно"}, Correct: 1, Explanation: "Без буфера каждый Read — системный вызов к ОС (дорого). bufio читает блоками и выдаёт данные из памяти."},
					{Question: "Что значит права 0600?", Options: []string{"Всем можно всё", "Только владелец: чтение+запись, остальные — ничего", "Только чтение", "Запрет доступа"}, Correct: 1, Explanation: "0600 = rw------- → владелец может читать и писать, группа и остальные — ничего. Используй для секретных файлов."},
					{Question: "Как скопировать файл в 10ГБ без загрузки в память?", Options: []string{"os.ReadFile + os.WriteFile", "io.Copy(dst, src) — потоковое копирование", "Невозможно в Go", "bytes.Buffer"}, Correct: 1, Explanation: "io.Copy читает и пишет блоками, не загружая весь файл в память. Работает с любым размером файла."},
					{Question: "Почему ошибку Close() важно проверять для файлов на запись?", Options: []string{"Не важно", "Буферизованные данные сбрасываются на диск при Close — ошибка означает потерю данных", "Close не может вернуть ошибку", "Только для Linux"}, Correct: 1, Explanation: "При записи данные могут буферизоваться. Close() сбрасывает буфер на диск. Если диск полный или сеть упала — ошибка будет именно в Close."},
				},
				Tasks: []T{
					{
						Title:      "Потоковый анализатор лога",
						Difficulty: "easy",
						Glossary: []GlossaryItem{
							{Term: "bufio.NewScanner(r)", Definition: "Создаёт построчный сканер из io.Reader. Работает с os.Stdin, файлами, буферами — одинаково."},
							{Term: "strings.Fields(s)", Definition: "Разбивает строку по пробелам. Аналог split по whitespace."},
						},
						Description: `<p>Прочитай лог-строки из stdin (формат: <code>LEVEL message</code>) и выведи статистику:</p>
<p>Ввод:</p>
<pre><code>INFO Server started
ERROR Connection failed
INFO Request handled
WARN Slow query
ERROR Disk full</code></pre>
<p>Вывод:</p>
<pre><code>INFO: 2
WARN: 1
ERROR: 2</code></pre>`,
						Hints: `<p>Используй bufio.Scanner + map[string]int для счётчиков. strings.Fields(line)[0] — уровень.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	counts := map[string]int{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 0 {
			counts[fields[0]]++
		}
	}
	for _, level := range []string{"INFO", "WARN", "ERROR"} {
		if c, ok := counts[level]; ok {
			fmt.Printf("%s: %d\n", level, c)
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	counts := map[string]int{}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		// TODO: fields[0] — уровень (INFO/WARN/ERROR)
		// Увеличь счётчик: counts[level]++
		_ = fields
	}

	// TODO: выведи статистику в порядке: INFO, WARN, ERROR
	// Формат: "LEVEL: count"
	_ = counts
}`,
						TestCases: []TestCase{
							{Input: "INFO Server started\nERROR Connection failed\nINFO Request handled\nWARN Slow query\nERROR Disk full", ExpectedOutput: "INFO: 2\nWARN: 1\nERROR: 2"},
							{Input: "ERROR fail\nERROR retry\nERROR crash", ExpectedOutput: "ERROR: 3"},
						},
					},
					{
						Title:      "Фильтр расширений файлов",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "strings.ToLower(s)", Definition: "Приводит строку к нижнему регистру. Для сравнения расширений: .MP4 → .mp4."},
							{Term: "strings.LastIndex(s, sep)", Definition: "Находит последнюю позицию sep в строке. Для извлечения расширения."},
						},
						Description: `<p>Симулируй <code>filepath.WalkDir</code>: из stdin читай пути файлов. Отфильтруй видеофайлы (.mp4, .mkv, .avi, .mov).</p>
<p>Для каждого видео выведи имя файла (без пути) и расширение.</p>
<p>Ввод:</p>
<pre><code>videos/comedy/movie.mp4
docs/readme.txt
videos/drama/film.MKV
music/song.mp3
clips/short.avi</code></pre>
<p>Вывод:</p>
<pre><code>movie.mp4 [.mp4]
film.MKV [.mkv]
short.avi [.avi]</code></pre>`,
						Hints: `<p>Для имени: всё после последнего /. Для расширения: всё после последней точки. Сравнивай в нижнем регистре.</p>`,
						Solution: `<pre><code>package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	videoExts := map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".mov": true}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := scanner.Text()
		name := path
		if i := strings.LastIndex(path, "/"); i >= 0 {
			name = path[i+1:]
		}
		dot := strings.LastIndex(name, ".")
		if dot < 0 {
			continue
		}
		ext := name[dot:]
		if videoExts[strings.ToLower(ext)] {
			fmt.Printf("%s [%s]\n", name, strings.ToLower(ext))
		}
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	videoExts := map[string]bool{".mp4": true, ".mkv": true, ".avi": true, ".mov": true}
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		path := scanner.Text()

		// TODO: извлеки имя файла (после последнего /)
		// TODO: извлеки расширение (после последней точки)
		// TODO: если расширение (в нижнем регистре) есть в videoExts → выведи
		// Формат: "name [.ext]"
		_ = path
		_ = videoExts
	}
}`,
						TestCases: []TestCase{
							{Input: "videos/comedy/movie.mp4\ndocs/readme.txt\nvideos/drama/film.MKV\nmusic/song.mp3\nclips/short.avi", ExpectedOutput: "movie.mp4 [.mp4]\nfilm.MKV [.mkv]\nshort.avi [.avi]"},
						},
					},
					{
						Title:      "Счётчик строк и слов",
						Difficulty: "easy",
						Description: `<p>Прочитай текст из stdin и выведи статистику: количество строк, слов и символов (как <code>wc</code>):</p>
<p>Ввод:</p><pre><code>hello world
foo bar baz
end</code></pre>
<p>Вывод: <code>3 6 27</code></p>`,
						Glossary: []GlossaryItem{
							{Term: "bufio.Scanner", Definition: "Читает ввод построчно. Для подсчёта слов — strings.Fields(line)."},
						},
						TestCases: []TestCase{
							{Input: "hello world\nfoo bar baz\nend", ExpectedOutput: "3 6 27"},
						},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    sc := bufio.NewScanner(os.Stdin)
    lines, words, chars := 0, 0, 0
    for sc.Scan() {
        line := sc.Text()
        lines++
        words += len(strings.Fields(line))
        chars += len(line) + 1 // +1 for newline
    }
    fmt.Printf("%d %d %d\n", lines, words, chars)
}`,
						Hints: `<p><code>strings.Fields(line)</code> разбивает на слова. <code>len(line)+1</code> считает символы + newline.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin); l,w,c:=0,0,0
    for sc.Scan(){t:=sc.Text();l++;w+=len(strings.Fields(t));c+=len(t)+1}; fmt.Printf("%d %d %d\n",l,w,c) }</code></pre>`,
					},
					{
						Title:      "Grep на Go",
						Difficulty: "medium",
						Description: `<p>Реализуй упрощённый grep: найди строки содержащие подстроку (case-insensitive):</p>
<p>Ввод:</p><pre><code>error
[INFO] Server started
[ERROR] Connection failed
[WARN] Timeout
[ERROR] Disk full</code></pre>
<p>Вывод:</p><pre><code>2: [ERROR] Connection failed
4: [ERROR] Disk full</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "strings.ToLower", Definition: "Для case-insensitive поиска: сравнивай strings.ToLower(line) с strings.ToLower(pattern)."},
						},
						TestCases: []TestCase{
							{Input: "error\n[INFO] Server started\n[ERROR] Connection failed\n[WARN] Timeout\n[ERROR] Disk full", ExpectedOutput: "2: [ERROR] Connection failed\n4: [ERROR] Disk full"},
						},
						StarterCode: `package main
import ("bufio";"fmt";"os";"strings")
func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); pattern := strings.ToLower(sc.Text())
    lineNum := 0
    for sc.Scan() {
        lineNum++
        if strings.Contains(strings.ToLower(sc.Text()), pattern) {
            fmt.Printf("%d: %s\n", lineNum, sc.Text())
        }
    }
}`,
						Hints: `<p><code>strings.Contains(strings.ToLower(line), strings.ToLower(pattern))</code></p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin); sc.Scan(); p:=strings.ToLower(sc.Text()); n:=0
    for sc.Scan(){n++;if strings.Contains(strings.ToLower(sc.Text()),p){fmt.Printf("%d: %s\n",n,sc.Text())}} }</code></pre>`,
					},
					{
						Title:      "Tail -n",
						Difficulty: "hard",
						Description: `<p>Реализуй <code>tail -n N</code>: выведи последние N строк ввода. Используй кольцевой буфер:</p>
<p>Ввод:</p><pre><code>2
line 1
line 2
line 3
line 4
line 5</code></pre>
<p>Вывод:</p><pre><code>line 4
line 5</code></pre>`,
						Glossary: []GlossaryItem{
							{Term: "Ring buffer", Definition: "Массив фиксированного размера + индекс записи. Новые элементы перезаписывают старые по кругу."},
						},
						TestCases: []TestCase{
							{Input: "2\nline 1\nline 2\nline 3\nline 4\nline 5", ExpectedOutput: "line 4\nline 5"},
							{Input: "3\na\nb", ExpectedOutput: "a\nb"},
						},
						StarterCode: `package main
import ("bufio";"fmt";"os")
func main() {
    sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); var n int; fmt.Sscan(sc.Text(), &n)
    buf := make([]string, n); idx := 0; total := 0
    for sc.Scan() { buf[idx%n] = sc.Text(); idx++; total++ }
    start := 0; count := total
    if count > n { start = idx % n; count = n }
    for i := 0; i < count; i++ { fmt.Println(buf[(start+i)%n]) }
}`,
						Hints: `<p>Кольцевой буфер: <code>buf[idx%n] = line</code>. Потом выводи с позиции <code>(idx%n)</code> по кругу.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"fmt";"os")
func main() { sc:=bufio.NewScanner(os.Stdin); sc.Scan(); var n int; fmt.Sscan(sc.Text(),&n)
    b:=make([]string,n); x,t:=0,0; for sc.Scan(){b[x%n]=sc.Text();x++;t++}
    s,c:=0,t; if c>n{s=x%n;c=n}; for i:=0;i<c;i++{fmt.Println(b[(s+i)%n])} }</code></pre>`,
					},
				},
			},
			{
				Slug: "json-deep", Title: "JSON: сериализация данных", Order: 2,
				Content: `<h1>JSON в Go — полный разбор</h1>

<h2>Под капотом: encoding/json</h2>
<p>Go использует <strong>рефлексию</strong> для чтения struct tags и автоматического маппинга полей:</p>

<pre><code>type Video struct {
    ID       int64     ` + "`" + `json:"id"` + "`" + `
    Title    string    ` + "`" + `json:"title"` + "`" + `
    Duration int       ` + "`" + `json:"duration"` + "`" + `
    FilePath string    ` + "`" + `json:"-"` + "`" + `              // исключён из JSON
    Size     int64     ` + "`" + `json:"size,omitempty"` + "`" + ` // пропустить если 0
    Tags     []string  ` + "`" + `json:"tags"` + "`" + `
}

// Маршалинг (struct → JSON)
v := Video{ID: 1, Title: "Матрица", Duration: 8160}
data, err := json.Marshal(v)
// {"id":1,"title":"Матрица","duration":8160,"tags":null}

// Красивый JSON (для отладки)
data, err := json.MarshalIndent(v, "", "  ")

// Демаршалинг (JSON → struct)
var v2 Video
err := json.Unmarshal(data, &v2)</code></pre>

<h2>Ловушка: nil slice vs пустой slice в JSON</h2>
<pre><code>type Response struct {
    Items []string ` + "`" + `json:"items"` + "`" + `
}

r1 := Response{Items: nil}
// JSON: {"items":null}  ← фронтенд может сломаться!

r2 := Response{Items: []string{}}
// JSON: {"items":[]}    ← корректный пустой массив

// Лучшая практика: всегда инициализируй слайсы для JSON
if r.Items == nil {
    r.Items = []string{}
}</code></pre>

<h2>Потоковый JSON (json.Encoder / json.Decoder)</h2>
<pre><code>// Encoder — пишет JSON напрямую в Writer (без промежуточного []byte)
json.NewEncoder(w).Encode(data) // в http.ResponseWriter

// Decoder — читает из Reader
var input CreateVideoRequest
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields() // отклонить неизвестные поля
err := dec.Decode(&input)</code></pre>

<p><strong>Encoder/Decoder vs Marshal/Unmarshal:</strong></p>
<ul>
<li><code>Marshal/Unmarshal</code> — работают с <code>[]byte</code>, для данных в памяти</li>
<li><code>Encoder/Decoder</code> — работают с потоками (файлы, HTTP), эффективнее по памяти</li>
</ul>

<h2>Работа с динамическим JSON</h2>
<pre><code>// Когда структура JSON заранее неизвестна
var data map[string]interface{}
json.Unmarshal(rawJSON, &data)

// Или более типизированно
var data map[string]json.RawMessage
json.Unmarshal(rawJSON, &data)
// Потом парсим каждое поле отдельно</code></pre>

<h2>Custom Marshaler</h2>
<pre><code>// Когда стандартная сериализация не подходит
type Duration struct {
    time.Duration
}

func (d Duration) MarshalJSON() ([]byte, error) {
    return json.Marshal(d.String()) // "2h30m" вместо наносекунд
}

func (d *Duration) UnmarshalJSON(b []byte) error {
    var s string
    if err := json.Unmarshal(b, &s); err != nil {
        return err
    }
    dur, err := time.ParseDuration(s)
    d.Duration = dur
    return err
}</code></pre>

<h2>Частые ошибки</h2>
<pre><code>// ОШИБКА: неэкспортированные поля не сериализуются
type video struct {
    title string  // маленькая буква → НЕ попадёт в JSON!
}

// ОШИБКА: time.Time по умолчанию → ISO 8601
// "2024-01-15T10:30:00Z" — может не подойти фронтенду

// ОШИБКА: числа из JSON → float64 (не int!)
var data map[string]interface{}
json.Unmarshal([]byte(` + "`" + `{"count": 42}` + "`" + `), &data)
count := data["count"].(float64) // float64, НЕ int!</code></pre>`,

				Quiz: []Q{
					{Question: "Чем Encoder/Decoder отличается от Marshal/Unmarshal?", Options: []string{"Ничем", "Encoder/Decoder работают с потоками (io.Writer/Reader), Marshal/Unmarshal — с []byte в памяти", "Marshal быстрее", "Decoder устарел"}, Correct: 1, Explanation: "Encoder пишет напрямую в Writer (файл, HTTP). Marshal сначала создаёт []byte в памяти. Для HTTP всегда используй Encoder."},
					{Question: "Что будет если поле struct начинается с маленькой буквы?", Options: []string{"Нормальная сериализация", "Поле полностью игнорируется JSON — неэкспортированные поля невидимы", "Ошибка компиляции", "Поле будет null"}, Correct: 1, Explanation: "encoding/json использует рефлексию, которая не может получить доступ к неэкспортированным полям. Они просто невидимы."},
					{Question: "В какой тип Go превращается число из JSON при Unmarshal в interface{}?", Options: []string{"int", "int64", "float64", "string"}, Correct: 2, Explanation: "JSON не различает int и float. При декодировании в interface{} все числа становятся float64. Для точности используй конкретные struct."},
					{Question: "Зачем DisallowUnknownFields() при декодировании?", Options: []string{"Для скорости", "Отклоняет JSON с полями, которых нет в struct — защита от опечаток в API", "Для сжатия", "Не нужен"}, Correct: 1, Explanation: "Без этого JSON с полем 'tilte' (опечатка) молча пройдёт, а поле Title останется пустым. С DisallowUnknownFields — ошибка."},
					{Question: "Как сделать чтобы nil slice выводился как [] а не null в JSON?", Options: []string{"Использовать omitempty", "Инициализировать: Items = []string{} вместо nil", "Невозможно", "Использовать json:\"-\""}, Correct: 1, Explanation: "nil slice → null в JSON. Пустой slice []string{} → []. Для API всегда инициализируй слайсы перед маршалингом."},
				},
				Tasks: []T{
					{
						Title:      "JSON парсер видео",
						Difficulty: "medium",
						Glossary: []GlossaryItem{
							{Term: "json.Unmarshal(data, &v)", Definition: "Парсит JSON bytes в Go-структуру. Передавай указатель (&v)."},
							{Term: "json.Marshal(v)", Definition: "Превращает Go-значение в JSON bytes."},
							{Term: "json:\"name,omitempty\"", Definition: "Struct tag: name — имя поля в JSON, omitempty — пропустить если zero value."},
						},
						Description: `<p>На вход — JSON массив видео. Парси его и выведи отформатированную таблицу.</p>
<p>Ввод:</p>
<pre><code>[{"title":"Matrix","year":1999,"duration":136},{"title":"Inception","year":2010,"duration":148}]</code></pre>
<p>Вывод:</p>
<pre><code>1. Matrix (1999) 2h16m
2. Inception (2010) 2h28m
Total: 2 videos, 4h44m</code></pre>`,
						Hints: `<p>json.Unmarshal в []Video. Длительность: hours = d/60, mins = d%60. Суммируй total.</p>`,
						Solution: `<pre><code>package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Video struct {
	Title    string ` + "`" + `json:"title"` + "`" + `
	Year     int    ` + "`" + `json:"year"` + "`" + `
	Duration int    ` + "`" + `json:"duration"` + "`" + `
}

func main() {
	data, _ := io.ReadAll(os.Stdin)
	var videos []Video
	json.Unmarshal(data, &videos)

	total := 0
	for i, v := range videos {
		fmt.Printf("%d. %s (%d) %dh%02dm\n", i+1, v.Title, v.Year, v.Duration/60, v.Duration%60)
		total += v.Duration
	}
	fmt.Printf("Total: %d videos, %dh%02dm\n", len(videos), total/60, total%60)
}</code></pre>`,
						StarterCode: `package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Video struct {
	Title    string ` + "`" + `json:"title"` + "`" + `
	Year     int    ` + "`" + `json:"year"` + "`" + `
	Duration int    ` + "`" + `json:"duration"` + "`" + `
}

func main() {
	data, _ := io.ReadAll(os.Stdin)

	// TODO: json.Unmarshal(data, &videos)
	var videos []Video

	// TODO: выведи каждое видео: "N. Title (Year) XhYYm"
	// И итог: "Total: N videos, XhYYm"
	_ = videos
	_ = data
}`,
						TestCases: []TestCase{
							{Input: `[{"title":"Matrix","year":1999,"duration":136},{"title":"Inception","year":2010,"duration":148}]`, ExpectedOutput: "1. Matrix (1999) 2h16m\n2. Inception (2010) 2h28m\nTotal: 2 videos, 4h44m"},
							{Input: `[{"title":"Go Tour","year":2024,"duration":30}]`, ExpectedOutput: "1. Go Tour (2024) 0h30m\nTotal: 1 videos, 0h30m"},
						},
					},
					{
						Title:      "JSON → CSV конвертер",
						Difficulty: "hard",
						Glossary: []GlossaryItem{
							{Term: "json.NewDecoder(r).Decode(&v)", Definition: "Потоковый парсинг JSON из io.Reader. Эффективнее чем ReadAll + Unmarshal."},
							{Term: "fmt.Fprintf(w, format, args...)", Definition: "Форматированная запись в io.Writer. Для вывода CSV."},
						},
						Description: `<p>Конвертируй JSON массив объектов в CSV. Первая строка — заголовки (ключи), остальные — значения.</p>
<p>Ввод:</p>
<pre><code>[{"name":"Alice","age":30,"city":"Moscow"},{"name":"Bob","age":25,"city":"Berlin"}]</code></pre>
<p>Вывод:</p>
<pre><code>age,city,name
30,Moscow,Alice
25,Berlin,Bob</code></pre>
<p>Ключи — в алфавитном порядке.</p>`,
						Hints: `<p>Unmarshal в []map[string]any. Собери ключи из первого элемента, отсортируй. Для значений: fmt.Sprintf("%v", val).</p>`,
						Solution: `<pre><code>package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	data, _ := io.ReadAll(os.Stdin)
	var rows []map[string]any
	json.Unmarshal(data, &rows)

	if len(rows) == 0 {
		return
	}

	var keys []string
	for k := range rows[0] {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	fmt.Println(strings.Join(keys, ","))

	for _, row := range rows {
		vals := make([]string, len(keys))
		for i, k := range keys {
			vals[i] = fmt.Sprintf("%v", row[k])
		}
		fmt.Println(strings.Join(vals, ","))
	}
}</code></pre>`,
						StarterCode: `package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func main() {
	data, _ := io.ReadAll(os.Stdin)
	var rows []map[string]any
	json.Unmarshal(data, &rows)

	if len(rows) == 0 {
		return
	}

	// TODO: собери ключи из rows[0], отсортируй
	// TODO: выведи заголовки через запятую
	// TODO: для каждой строки выведи значения через запятую
	// Используй fmt.Sprintf("%v", val) для преобразования any → string

	_ = rows
}`,
						TestCases: []TestCase{
							{Input: `[{"name":"Alice","age":30,"city":"Moscow"},{"name":"Bob","age":25,"city":"Berlin"}]`, ExpectedOutput: "age,city,name\n30,Moscow,Alice\n25,Berlin,Bob"},
							{Input: `[{"x":1,"y":2},{"x":3,"y":4}]`, ExpectedOutput: "x,y\n1,2\n3,4"},
						},
					},
					{
						Title: "JSON валидатор", Difficulty: "easy",
						Description: `<p>Проверь каждую строку — валидный ли JSON. Если да — тип (object/array):</p>
<p>Ввод:</p><pre><code>3
{"name":"Alice"}
{bad
[1,2]</code></pre>
<p>Вывод:</p><pre><code>VALID: object
INVALID
VALID: array</code></pre>`,
						Glossary:  []GlossaryItem{{Term: "json.Valid", Definition: "Проверяет JSON без десериализации."}},
						TestCases: []TestCase{{Input: "3\n{\"name\":\"Alice\"}\n{bad\n[1,2]", ExpectedOutput: "VALID: object\nINVALID\nVALID: array"}},
						StarterCode: `package main
import ("bufio";"encoding/json";"fmt";"os";"strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { sc.Scan(); l := strings.TrimSpace(sc.Text())
        if !json.Valid([]byte(l)) { fmt.Println("INVALID") } else if strings.HasPrefix(l,"{") { fmt.Println("VALID: object") } else { fmt.Println("VALID: array") } } }`,
						Hints: `<p>json.Valid([]byte(s)). Первый символ определяет тип.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"encoding/json";"fmt";"os";"strings")
func main() { var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();l:=strings.TrimSpace(sc.Text());if !json.Valid([]byte(l)){fmt.Println("INVALID")}else if strings.HasPrefix(l,"{"){fmt.Println("VALID: object")}else{fmt.Println("VALID: array")}} }</code></pre>`,
					},
					{
						Title: "JSON merge", Difficulty: "medium",
						Description: `<p>Объедини два JSON-объекта (второй перезаписывает):</p>
<p>Ввод:</p><pre><code>{"name":"Alice","age":25}
{"age":26,"city":"Moscow"}</code></pre>
<p>Вывод: <code>{"age":26,"city":"Moscow","name":"Alice"}</code></p>`,
						Glossary:  []GlossaryItem{{Term: "map[string]any", Definition: "Для произвольного JSON. Marshal сортирует ключи."}},
						TestCases: []TestCase{{Input: "{\"name\":\"Alice\",\"age\":25}\n{\"age\":26,\"city\":\"Moscow\"}", ExpectedOutput: "{\"age\":26,\"city\":\"Moscow\",\"name\":\"Alice\"}"}},
						StarterCode: `package main
import ("bufio";"encoding/json";"fmt";"os")
func main() { sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); var a map[string]any; json.Unmarshal([]byte(sc.Text()), &a)
    sc.Scan(); var b map[string]any; json.Unmarshal([]byte(sc.Text()), &b)
    for k, v := range b { a[k] = v }; d, _ := json.Marshal(a); fmt.Println(string(d)) }`,
						Hints: `<p>Unmarshal оба → merge ключи → Marshal.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"encoding/json";"fmt";"os")
func main() { sc:=bufio.NewScanner(os.Stdin);sc.Scan();var a map[string]any;json.Unmarshal([]byte(sc.Text()),&a)
    sc.Scan();var b map[string]any;json.Unmarshal([]byte(sc.Text()),&b);for k,v:=range b{a[k]=v};d,_:=json.Marshal(a);fmt.Println(string(d)) }</code></pre>`,
					},
					{
						Title: "JSON path extractor", Difficulty: "hard",
						Description: `<p>Извлеки значение по точечному пути из вложенного JSON:</p>
<p>Ввод:</p><pre><code>{"user":{"address":{"city":"Moscow"}}}
user.address.city</code></pre>
<p>Вывод: <code>Moscow</code></p>`,
						Glossary: []GlossaryItem{{Term: "Nested traversal", Definition: "Разбей путь по точке, на каждом шаге cast к map[string]any."}},
						TestCases: []TestCase{
							{Input: "{\"user\":{\"address\":{\"city\":\"Moscow\"}}}\nuser.address.city", ExpectedOutput: "Moscow"},
							{Input: "{\"a\":{\"b\":{\"c\":42}}}\na.b.c", ExpectedOutput: "42"},
						},
						StarterCode: `package main
import ("bufio";"encoding/json";"fmt";"os";"strings")
func main() { sc := bufio.NewScanner(os.Stdin)
    sc.Scan(); var d map[string]any; json.Unmarshal([]byte(sc.Text()), &d)
    sc.Scan(); var c any = d
    for _, k := range strings.Split(sc.Text(), ".") { c = c.(map[string]any)[k] }
    if v, ok := c.(float64); ok { fmt.Printf("%g\n", v) } else { fmt.Println(c) } }`,
						Hints: `<p>На каждом шаге <code>c.(map[string]any)[key]</code>.</p>`,
						Solution: `<pre><code>package main
import ("bufio";"encoding/json";"fmt";"os";"strings")
func main() { sc:=bufio.NewScanner(os.Stdin);sc.Scan();var d map[string]any;json.Unmarshal([]byte(sc.Text()),&d)
    sc.Scan();var c any=d;for _,k:=range strings.Split(sc.Text(),"."){c=c.(map[string]any)[k]}
    if v,ok:=c.(float64);ok{fmt.Printf("%g\n",v)}else{fmt.Println(c)} }</code></pre>`,
					},
				},
			},
		},
	}
}
