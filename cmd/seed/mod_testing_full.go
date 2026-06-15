package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Тестирование — расширенный (5 уроков)
// Заменяет mod12_testing()
// ════════════════════════════════════════════════════════════════

func mod_testing_full() M {
	return M{
		Slug: "testing", Title: "Тестирование в Go", Order: 12,
		Description: "Unit тесты, table-driven, моки, интеграционные тесты, benchmarks, coverage.",
		Track: "backend", Difficulty: "advanced", Prerequisites: []string{"architecture"},
		Lessons: []L{
			{
				Slug: "unit-testing", Title: "Unit тесты — основы", Order: 1,
				Difficulty: "intermediate", Track: "backend",
				Content: `<h1>Unit тестирование в Go</h1>

<h2>Философия тестирования в Go</h2>
<p>Go имеет встроенную поддержку тестов — не нужны внешние фреймворки:</p>
<pre><code>// Файл: math.go
package math

func Add(a, b int) int { return a + b }

// Файл: math_test.go (суффикс _test.go!)
package math

import "testing"

func TestAdd(t *testing.T) {
    result := Add(2, 3)
    if result != 5 {
        t.Errorf("Add(2, 3) = %d, want 5", result)
    }
}</code></pre>

<h2>Запуск тестов</h2>
<pre><code>go test ./...              # все тесты
go test -v ./...           # verbose — видно каждый тест
go test -run TestAdd       # конкретный тест
go test -race ./...        # с детектором гонок
go test -count=1 ./...     # без кеша</code></pre>

<h2>Методы *testing.T</h2>
<pre><code>t.Error("msg")   // ошибка, тест продолжается
t.Errorf(...)    // то же с форматом
t.Fatal("msg")   // ошибка, тест ОСТАНАВЛИВАЕТСЯ
t.Fatalf(...)    // то же с форматом
t.Log("info")    // вывод (виден с -v)
t.Skip("reason") // пропустить тест
t.Helper()       // пометить как helper (для stack trace)</code></pre>

<h2>Организация тестов</h2>
<pre><code>// Тест в том же пакете — доступ к приватным полям
package user

// Тест в пакете _test — тестирует публичный API
package user_test

import "myapp/internal/user"

func TestCreate(t *testing.T) {
    u := user.New("Alice")  // только публичный API
}</code></pre>

<h2>testify — популярная библиотека</h2>
<pre><code>import "github.com/stretchr/testify/assert"

func TestAdd(t *testing.T) {
    assert.Equal(t, 5, Add(2, 3))
    assert.NotNil(t, result)
    assert.Error(t, err)
    assert.Contains(t, str, "hello")
}</code></pre>

<h2>Читать глубже</h2>
<ul>
<li><a href="https://habr.com/ru/articles/568036/" target="_blank">Хабр: Тестирование в Go</a></li>
<li><a href="https://metanit.com/go/golang/9.1.php" target="_blank">Metanit: Тестирование</a></li>
</ul>`,

				Quiz: []Q{
					{Question: "Как Go находит файлы с тестами?", Options: []string{"Через конфиг", "По суффиксу _test.go — автоматически, без регистрации", "По папке tests/", "Через main"}, Correct: 1, Explanation: "Соглашение Go: файл *_test.go + функция Test*(t *testing.T). go test находит автоматически. Не нужен test runner."},
					{Question: "t.Error vs t.Fatal?", Options: []string{"Одно и то же", "Error продолжает тест (report failure). Fatal останавливает тест немедленно", "Fatal мягче", "Error для warnings"}, Correct: 1, Explanation: "Error: 'этот check провалился, но проверь остальные'. Fatal: 'дальше проверять бессмысленно' (например, nil pointer)."},
					{Question: "go test -race — зачем?", Options: []string{"Ускоряет", "Включает детектор гонок данных — находит concurrent bugs", "Для бенчмарков", "Пропускает медленные"}, Correct: 1, Explanation: "-race использует ThreadSanitizer. Замедляет в ~10x но находит data races. Обязателен в CI для concurrent кода."},
					{Question: "package user_test vs package user — в чём разница?", Options: []string{"Нет разницы", "user_test = black-box (только public API). user = white-box (доступ к приватным полям)", "user_test быстрее", "user устарел"}, Correct: 1, Explanation: "user_test тестирует как внешний пользователь пакета — лучше для API-стабильности. user — для тестирования внутренней логики."},
					{Question: "Зачем t.Helper()?", Options: []string{"Ускоряет", "Помечает функцию как helper — ошибка покажет строку вызывающего, не helper-а", "Пропускает тест", "Создаёт mock"}, Correct: 1, Explanation: "Без Helper(): ошибка показывает строку внутри helper-функции. С Helper(): показывает строку где helper был вызван — проще найти проблему."},
				},
				Tasks: []T{
					{Title: "Первый тест", Difficulty: "easy", Description: `<p>Напиши тест для функции <code>Max(a, b int) int</code>:</p><p>Вывод (при запуске): <code>PASS</code></p>`, Glossary: []GlossaryItem{{Term: "func TestX(t *testing.T)", Definition: "Сигнатура теста. Test + CamelCase имя. t — объект для ассертов."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS"}},
						StarterCode: `package main
import "fmt"
func Max(a, b int) int { if a > b { return a }; return b }
func testMax() string {
    if Max(3, 7) != 7 { return "FAIL: Max(3,7)" }
    if Max(10, 5) != 10 { return "FAIL: Max(10,5)" }
    if Max(-1, -5) != -1 { return "FAIL: Max(-1,-5)" }
    return "PASS"
}
func main() { fmt.Println(testMax()) }`, Hints: `<p>Проверяй несколько случаев: обычный, равные, отрицательные.</p>`, Solution: `<pre><code>package main
import "fmt"
func Max(a, b int) int { if a > b { return a }; return b }
func testMax() string {
    cases := [][3]int{{3,7,7},{10,5,10},{-1,-5,-1},{0,0,0}}
    for _, c := range cases { if Max(c[0], c[1]) != c[2] { return fmt.Sprintf("FAIL: Max(%d,%d)=%d, want %d", c[0], c[1], Max(c[0],c[1]), c[2]) } }
    return "PASS"
}
func main() { fmt.Println(testMax()) }</code></pre>`},
					{Title: "Тест с несколькими проверками", Difficulty: "easy", Description: `<p>Протестируй <code>IsEven(n int) bool</code>:</p><p>Вывод: <code>PASS: 4/4 checks</code></p>`, Glossary: []GlossaryItem{{Term: "multiple assertions", Definition: "Проверяй разные случаи: positive, negative, zero, edge cases."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: 4/4 checks"}},
						StarterCode: `package main
import "fmt"
func IsEven(n int) bool { return n%2 == 0 }
func main() {
    checks := 0; total := 4
    if IsEven(4) == true { checks++ }
    if IsEven(7) == false { checks++ }
    if IsEven(0) == true { checks++ }
    if IsEven(-2) == true { checks++ }
    fmt.Printf("PASS: %d/%d checks\n", checks, total)
}`, Hints: `<p>Проверь: чётное true, нечётное false, ноль, отрицательное.</p>`, Solution: `<pre><code>package main
import "fmt"
func IsEven(n int) bool { return n%2 == 0 }
func main() { c := 0; if IsEven(4) { c++ }; if !IsEven(7) { c++ }; if IsEven(0) { c++ }; if IsEven(-2) { c++ }; fmt.Printf("PASS: %d/%d checks\n", c, 4) }</code></pre>`},
					{Title: "Error testing", Difficulty: "medium", Description: `<p>Протестируй функцию которая возвращает error:</p><p>Вывод: <code>PASS: divide tests</code></p>`, Glossary: []GlossaryItem{{Term: "error testing", Definition: "Проверяй и happy path (err == nil) и error path (err != nil, правильное сообщение)."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: divide tests"}},
						StarterCode: `package main
import "fmt"
func divide(a, b float64) (float64, error) {
    if b == 0 { return 0, fmt.Errorf("division by zero") }
    return a / b, nil
}
func main() {
    r, err := divide(10, 2)
    if err != nil || r != 5 { fmt.Println("FAIL: happy path"); return }
    _, err = divide(10, 0)
    if err == nil { fmt.Println("FAIL: should error on zero"); return }
    fmt.Println("PASS: divide tests")
}`, Hints: `<p>Тестируй и success (err==nil, result correct) и failure (err!=nil).</p>`, Solution: `<pre><code>package main
import "fmt"
func divide(a, b float64) (float64, error) { if b == 0 { return 0, fmt.Errorf("division by zero") }; return a / b, nil }
func main() { r, err := divide(10, 2); if err != nil || r != 5 { fmt.Println("FAIL"); return }; _, err = divide(10, 0); if err == nil { fmt.Println("FAIL"); return }; fmt.Println("PASS: divide tests") }</code></pre>`},
					{Title: "Helper function", Difficulty: "medium", Description: `<p>Напиши helper <code>assertEqual(got, want int) string</code> для удобных проверок:</p><p>Вывод: <code>PASS: all equal</code></p>`, Glossary: []GlossaryItem{{Term: "t.Helper()", Definition: "В реальных тестах helper показывает строку вызова, не helper-а."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: all equal"}},
						StarterCode: `package main
import "fmt"
func assertEqual(got, want int) string {
    if got != want { return fmt.Sprintf("got %d, want %d", got, want) }
    return ""
}
func main() {
    errs := []string{}
    if e := assertEqual(2+2, 4); e != "" { errs = append(errs, e) }
    if e := assertEqual(3*3, 9); e != "" { errs = append(errs, e) }
    if e := assertEqual(10/2, 5); e != "" { errs = append(errs, e) }
    if len(errs) > 0 { for _, e := range errs { fmt.Println("FAIL:", e) } } else { fmt.Println("PASS: all equal") }
}`, Hints: `<p>assertEqual возвращает "" при успехе, описание ошибки при fail.</p>`, Solution: `<pre><code>package main
import "fmt"
func assertEqual(got, want int) string { if got != want { return fmt.Sprintf("got %d, want %d", got, want) }; return "" }
func main() { var errs []string; for _, c := range [][2]int{{4,4},{9,9},{5,5}} { if e := assertEqual(c[0], c[1]); e != "" { errs = append(errs, e) } }; if len(errs) > 0 { fmt.Println("FAIL") } else { fmt.Println("PASS: all equal") } }</code></pre>`},
					{Title: "Test coverage report", Difficulty: "hard", Description: `<p>Имитируй coverage: по списку функций определи какие протестированы:</p><p>Ввод:</p><pre><code>5 3
Add Sub Mul Div Mod
Add Mul Div</code></pre><p>Вывод:</p><pre><code>Coverage: 60% (3/5)
Untested: Sub Mod</code></pre>`, Glossary: []GlossaryItem{{Term: "go test -cover", Definition: "Показывает процент покрытия кода тестами."}}, TestCases: []TestCase{{Input: "5 3\nAdd Sub Mul Div Mod\nAdd Mul Div", ExpectedOutput: "Coverage: 60% (3/5)\nUntested: Sub Mod"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var total, tested int; fmt.Scan(&total, &tested)
    scanner := bufio.NewScanner(os.Stdin)
    scanner.Scan(); funcs := strings.Fields(scanner.Text())
    scanner.Scan(); testedFuncs := strings.Fields(scanner.Text())
    testedSet := map[string]bool{}
    for _, f := range testedFuncs { testedSet[f] = true }
    var untested []string
    for _, f := range funcs { if !testedSet[f] { untested = append(untested, f) } }
    pct := tested * 100 / total
    fmt.Printf("Coverage: %d%% (%d/%d)\n", pct, tested, total)
    fmt.Printf("Untested: %s\n", strings.Join(untested, " "))
}`, Hints: `<p>Set из tested. Пройди по всем funcs — кто не в set = untested.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { var t, td int; fmt.Scan(&t, &td); sc := bufio.NewScanner(os.Stdin); sc.Scan(); funcs := strings.Fields(sc.Text()); sc.Scan(); tf := strings.Fields(sc.Text())
    s := map[string]bool{}; for _, f := range tf { s[f] = true }; var u []string; for _, f := range funcs { if !s[f] { u = append(u, f) } }
    fmt.Printf("Coverage: %d%% (%d/%d)\nUntested: %s\n", td*100/t, td, t, strings.Join(u, " ")) }</code></pre>`},
				},
			},
			{
				Slug: "table-driven", Title: "Table-Driven Tests", Order: 2,
				Difficulty: "intermediate", Track: "backend",
				Content: `<h1>Table-Driven Tests — паттерн Go</h1>

<h2>Проблема: дублирование</h2>
<pre><code>// ПЛОХО: один тест = одна проверка
func TestAdd2and3(t *testing.T) { ... }
func TestAdd0and0(t *testing.T) { ... }
func TestAddNegative(t *testing.T) { ... }
// 50 тестов = 50 функций...</code></pre>

<h2>Решение: таблица тестов</h2>
<pre><code>func TestAdd(t *testing.T) {
    tests := []struct {
        name     string
        a, b     int
        expected int
    }{
        {"positive", 2, 3, 5},
        {"zeros", 0, 0, 0},
        {"negative", -1, -2, -3},
        {"mixed", -1, 5, 4},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := Add(tt.a, tt.b)
            if got != tt.expected {
                t.Errorf("Add(%d, %d) = %d, want %d", tt.a, tt.b, got, tt.expected)
            }
        })
    }
}</code></pre>

<h2>t.Run — подтесты</h2>
<pre><code>// t.Run создаёт подтест с именем:
// --- FAIL: TestAdd/negative
//     math_test.go:15: Add(-1, -2) = -4, want -3
// Можно запустить конкретный: go test -run TestAdd/negative</code></pre>

<h2>Шаблон для ошибок</h2>
<pre><code>func TestDivide(t *testing.T) {
    tests := []struct {
        name    string
        a, b    float64
        want    float64
        wantErr bool
    }{
        {"normal", 10, 2, 5, false},
        {"zero div", 10, 0, 0, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := Divide(tt.a, tt.b)
            if (err != nil) != tt.wantErr {
                t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if got != tt.want {
                t.Errorf("got %v, want %v", got, tt.want)
            }
        })
    }
}</code></pre>`,

				Quiz: []Q{
					{Question: "Зачем table-driven tests?", Options: []string{"Мода", "Одна структура теста для всех случаев — DRY, легко добавить новый case", "Быстрее", "Go требует"}, Correct: 1, Explanation: "Добавить новый тест = добавить строку в таблицу. Не нужно писать новую функцию. Видно все edge cases в одном месте."},
					{Question: "Что делает t.Run(name, func)?", Options: []string{"Параллельный запуск", "Создаёт именованный подтест — видно в отчёте, можно запустить отдельно", "Пропускает тест", "Создаёт горутину"}, Correct: 1, Explanation: "t.Run(\"case_name\", ...) создаёт подтест. go test -run TestAdd/zeros — запустить конкретный. В отчёте: TestAdd/zeros PASS."},
					{Question: "Что тестировать: wantErr bool или конкретную ошибку?", Options: []string{"Только wantErr", "wantErr для наличия, errors.Is/As для конкретного типа ошибки", "Только строку", "Не тестировать ошибки"}, Correct: 1, Explanation: "wantErr: bool достаточно если неважно какая ошибка. Для точности: wantErr error + errors.Is(got, want). Для типа: errors.As."},
					{Question: "Как запустить один конкретный подтест?", Options: []string{"Нельзя", "go test -run TestFunc/subtest_name", "Только через IDE", "go test -name"}, Correct: 1, Explanation: "-run принимает regex. TestAdd/negative запустит только подтест 'negative' из TestAdd. Удобно для отладки одного failing case."},
					{Question: "Зачем поле name в таблице тестов?", Options: []string{"Обязательно", "Описание case — при fail видно КАКОЙ именно случай сломался", "Для документации", "Для скорости"}, Correct: 1, Explanation: "Без name: 'TestAdd failed at line 15'. С name: 'TestAdd/negative_numbers failed'. Сразу понятно что сломалось без чтения кода."},
				},
				Tasks: []T{
					{Title: "Table-driven для FizzBuzz", Difficulty: "easy", Description: `<p>Напиши table-driven тесты для FizzBuzz:</p><p>Вывод: <code>PASS: 5/5 cases</code></p>`, Glossary: []GlossaryItem{{Term: "[]struct{input; want}", Definition: "Таблица тестов: слайс анонимных структур с входом и ожидаемым выходом."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: 5/5 cases"}},
						StarterCode: `package main
import "fmt"
func fizzbuzz(n int) string {
    if n%15 == 0 { return "FizzBuzz" }
    if n%3 == 0 { return "Fizz" }
    if n%5 == 0 { return "Buzz" }
    return fmt.Sprintf("%d", n)
}
func main() {
    tests := []struct{ input int; want string }{
        {1, "1"}, {3, "Fizz"}, {5, "Buzz"}, {15, "FizzBuzz"}, {7, "7"},
    }
    passed := 0
    for _, tt := range tests { if fizzbuzz(tt.input) == tt.want { passed++ } }
    fmt.Printf("PASS: %d/%d cases\n", passed, len(tests))
}`, Hints: `<p>Таблица: []struct{input int; want string}. Цикл проверяет каждый case.</p>`, Solution: `<pre><code>package main
import "fmt"
func fizzbuzz(n int) string { if n%15==0{return "FizzBuzz"}; if n%3==0{return "Fizz"}; if n%5==0{return "Buzz"}; return fmt.Sprintf("%d",n) }
func main() { tests := []struct{i int; w string}{{1,"1"},{3,"Fizz"},{5,"Buzz"},{15,"FizzBuzz"},{7,"7"}}; p := 0; for _, tt := range tests { if fizzbuzz(tt.i)==tt.w { p++ } }; fmt.Printf("PASS: %d/%d cases\n", p, len(tests)) }</code></pre>`},
					{Title: "Table-driven с ошибками", Difficulty: "medium", Description: `<p>Table-driven для функции с error:</p><p>Вывод: <code>PASS: 4/4 cases</code></p>`, Glossary: []GlossaryItem{{Term: "wantErr bool", Definition: "Поле в таблице: ожидаем ли ошибку от функции."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: 4/4 cases"}},
						StarterCode: `package main
import "fmt"
func parseInt(s string) (int, error) {
    n := 0; neg := false
    if len(s) == 0 { return 0, fmt.Errorf("empty") }
    start := 0
    if s[0] == '-' { neg = true; start = 1 }
    for _, c := range s[start:] { if c < '0' || c > '9' { return 0, fmt.Errorf("invalid: %s", s) }; n = n*10 + int(c-'0') }
    if neg { n = -n }
    return n, nil
}
func main() {
    tests := []struct{ input string; want int; wantErr bool }{
        {"42", 42, false}, {"-7", -7, false}, {"abc", 0, true}, {"", 0, true},
    }
    passed := 0
    for _, tt := range tests {
        got, err := parseInt(tt.input)
        errOk := (err != nil) == tt.wantErr
        valOk := !tt.wantErr && got == tt.want || tt.wantErr
        if errOk && valOk { passed++ }
    }
    fmt.Printf("PASS: %d/%d cases\n", passed, len(tests))
}`, Hints: `<p>Проверяй (err != nil) == tt.wantErr для error case. got == tt.want для value.</p>`, Solution: `<pre><code>package main
import "fmt"
func parseInt(s string) (int, error) { n:=0; neg:=false; if len(s)==0{return 0,fmt.Errorf("empty")}; st:=0; if s[0]=='-'{neg=true;st=1}; for _,c:=range s[st:]{if c<'0'||c>'9'{return 0,fmt.Errorf("invalid")}; n=n*10+int(c-'0')}; if neg{n=-n}; return n,nil }
func main() { ts:=[]struct{i string;w int;e bool}{{"42",42,false},{"-7",-7,false},{"abc",0,true},{"",0,true}}; p:=0
    for _,tt:=range ts{g,err:=parseInt(tt.i); if (err!=nil)==tt.e && (tt.e || g==tt.w){p++}}; fmt.Printf("PASS: %d/%d cases\n",p,len(ts)) }</code></pre>`},
					{Title: "Test runner с подтестами", Difficulty: "medium", Description: `<p>Имитируй t.Run — запускай подтесты и выводи результат каждого:</p><p>Вывод:</p><pre><code>--- PASS: TestAdd/positive
--- PASS: TestAdd/zero
--- PASS: TestAdd/negative
PASS (3/3)</code></pre>`, Glossary: []GlossaryItem{{Term: "t.Run(name, fn)", Definition: "Подтест. Именованный, можно запустить отдельно."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "--- PASS: TestAdd/positive\n--- PASS: TestAdd/zero\n--- PASS: TestAdd/negative\nPASS (3/3)"}},
						StarterCode: `package main
import "fmt"
func add(a, b int) int { return a + b }
func main() {
    tests := []struct{ name string; a, b, want int }{
        {"positive", 2, 3, 5}, {"zero", 0, 0, 0}, {"negative", -1, -2, -3},
    }
    passed := 0
    for _, tt := range tests {
        if add(tt.a, tt.b) == tt.want { fmt.Printf("--- PASS: TestAdd/%s\n", tt.name); passed++ } else { fmt.Printf("--- FAIL: TestAdd/%s\n", tt.name) }
    }
    fmt.Printf("PASS (%d/%d)\n", passed, len(tests))
}`, Hints: `<p>Для каждого case: если pass → "--- PASS: TestName/case". Итого внизу.</p>`, Solution: `<pre><code>package main
import "fmt"
func add(a,b int) int { return a+b }
func main() { ts := []struct{n string; a,b,w int}{{"positive",2,3,5},{"zero",0,0,0},{"negative",-1,-2,-3}}; p:=0
    for _,tt:=range ts { if add(tt.a,tt.b)==tt.w { fmt.Printf("--- PASS: TestAdd/%s\n",tt.n); p++ } else { fmt.Printf("--- FAIL: TestAdd/%s\n",tt.n) } }
    fmt.Printf("PASS (%d/%d)\n",p,len(ts)) }</code></pre>`},
					{Title: "Edge case generator", Difficulty: "hard", Description: `<p>Для заданной функции сгенерируй edge cases:</p><p>Ввод: <code>divide int int</code></p><p>Вывод:</p><pre><code>divide(0, 1) = 0
divide(1, 0) = error
divide(-1, -1) = 1
divide(MAX_INT, 1) = MAX_INT</code></pre>`, Glossary: []GlossaryItem{{Term: "Edge cases", Definition: "Граничные случаи: zero, negative, MAX, empty, nil. Самые частые источники багов."}}, TestCases: []TestCase{{Input: "divide int int", ExpectedOutput: "divide(0, 1) = 0\ndivide(1, 0) = error\ndivide(-1, -1) = 1\ndivide(MAX_INT, 1) = MAX_INT"}},
						StarterCode: `package main
import ("fmt"; "strings")
func main() {
    var name string; fmt.Scan(&name)
    var args string; fmt.Scanln(&args)
    _ = strings.TrimSpace(args)
    fmt.Printf("%s(0, 1) = 0\n", name)
    fmt.Printf("%s(1, 0) = error\n", name)
    fmt.Printf("%s(-1, -1) = 1\n", name)
    fmt.Printf("%s(MAX_INT, 1) = MAX_INT\n", name)
}`, Hints: `<p>Стандартные edge cases для числовых функций: 0, negative, max, division by zero.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; fmt.Scan(&n); fmt.Printf("%s(0, 1) = 0\n%s(1, 0) = error\n%s(-1, -1) = 1\n%s(MAX_INT, 1) = MAX_INT\n", n, n, n, n) }</code></pre>`},
					{Title: "Parametrized test generator", Difficulty: "hard", Description: `<p>По описанию функции сгенерируй Go test code:</p><p>Ввод: <code>Add 3</code> (функция, кол-во cases)</p><p>Вывод:</p><pre><code>func TestAdd(t *testing.T) {
    tests := []struct{ a, b, want int }{
        {1, 2, 3},
        {0, 0, 0},
        {-1, 1, 0},
    }
    for _, tt := range tests {
        if got := Add(tt.a, tt.b); got != tt.want {
            t.Errorf("Add(%d,%d) = %d, want %d", tt.a, tt.b, got, tt.want)
        }
    }
}</code></pre>`, Glossary: []GlossaryItem{{Term: "Code generation", Definition: "Генерация boilerplate тестов. В реальности: gotests CLI tool."}}, TestCases: []TestCase{{Input: "Add 3", ExpectedOutput: "func TestAdd(t *testing.T) {\n    tests := []struct{ a, b, want int }{\n        {1, 2, 3},\n        {0, 0, 0},\n        {-1, 1, 0},\n    }\n    for _, tt := range tests {\n        if got := Add(tt.a, tt.b); got != tt.want {\n            t.Errorf(\"Add(%d,%d) = %d, want %d\", tt.a, tt.b, got, tt.want)\n        }\n    }\n}"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; var n int; fmt.Scan(&name, &n)
    fmt.Printf("func Test%s(t *testing.T) {\n", name)
    fmt.Printf("    tests := []struct{ a, b, want int }{\n")
    cases := [][3]int{{1,2,3},{0,0,0},{-1,1,0}}
    for i := 0; i < n && i < len(cases); i++ { fmt.Printf("        {%d, %d, %d},\n", cases[i][0], cases[i][1], cases[i][2]) }
    fmt.Printf("    }\n")
    fmt.Printf("    for _, tt := range tests {\n")
    fmt.Printf("        if got := %s(tt.a, tt.b); got != tt.want {\n", name)
    fmt.Printf("            t.Errorf(\"%s(%%d,%%d) = %%d, want %%d\", tt.a, tt.b, got, tt.want)\n", name)
    fmt.Printf("        }\n    }\n}\n")
}`, Hints: `<p>Шаблон table-driven test с подстановкой имени функции.</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { var n string; var c int; fmt.Scan(&n, &c); cs := [][3]int{{1,2,3},{0,0,0},{-1,1,0}}
    fmt.Printf("func Test%s(t *testing.T) {\n    tests := []struct{ a, b, want int }{\n", n)
    for i:=0;i<c&&i<len(cs);i++{fmt.Printf("        {%d, %d, %d},\n",cs[i][0],cs[i][1],cs[i][2])}
    fmt.Printf("    }\n    for _, tt := range tests {\n        if got := %s(tt.a, tt.b); got != tt.want {\n            t.Errorf(\"%s(%%d,%%d) = %%d, want %%d\", tt.a, tt.b, got, tt.want)\n        }\n    }\n}\n", n, n) }</code></pre>`},
				},
			},
			{
				Slug: "mocks", Title: "Моки и dependency injection", Order: 3,
				Difficulty: "advanced", Track: "backend",
				Content: `<h1>Моки — тестирование без зависимостей</h1>

<h2>Проблема: как тестировать код с БД?</h2>
<p>UserService зависит от PostgreSQL. В тесте не хочется поднимать БД. Решение: интерфейс + мок.</p>

<h2>Паттерн: Interface → Mock</h2>
<pre><code>// 1. Определи интерфейс
type UserStore interface {
    GetByID(id int) (*User, error)
    Create(u *User) error
}

// 2. Реальная реализация
type PostgresStore struct { pool *pgxpool.Pool }
func (s *PostgresStore) GetByID(id int) (*User, error) { /* SQL query */ }

// 3. Mock для тестов
type MockStore struct {
    users map[int]*User
    err   error  // можно подставить ошибку
}
func (m *MockStore) GetByID(id int) (*User, error) {
    if m.err != nil { return nil, m.err }
    return m.users[id], nil
}

// 4. Тест использует mock
func TestGetUser(t *testing.T) {
    store := &MockStore{users: map[int]*User{1: {Name: "Alice"}}}
    svc := NewUserService(store)
    user, err := svc.GetUser(1)
    assert.NoError(t, err)
    assert.Equal(t, "Alice", user.Name)
}</code></pre>

<h2>Тестирование ошибок через mock</h2>
<pre><code>func TestGetUser_NotFound(t *testing.T) {
    store := &MockStore{err: ErrNotFound}
    svc := NewUserService(store)
    _, err := svc.GetUser(999)
    assert.ErrorIs(t, err, ErrNotFound)
}</code></pre>

<h2>httptest — мок HTTP</h2>
<pre><code>func TestHandler(t *testing.T) {
    req := httptest.NewRequest("GET", "/users/1", nil)
    w := httptest.NewRecorder()
    handler.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)
}</code></pre>`,

				Quiz: []Q{
					{Question: "Зачем нужны моки?", Options: []string{"Для красоты", "Тестировать бизнес-логику без реальных зависимостей (БД, HTTP, файлы)", "Ускоряют тесты", "Go требует"}, Correct: 1, Explanation: "Mock подменяет зависимость. Тест быстрый (нет БД), изолированный (не зависит от сети), детерминированный (всегда одинаковый результат)."},
					{Question: "Интерфейс нужен для мока потому что...", Options: []string{"Go требует", "Без интерфейса нельзя подменить реализацию — функция привязана к конкретному типу", "Для скорости", "Для документации"}, Correct: 1, Explanation: "Service зависит от interface, не от struct. В проде — PostgresStore. В тесте — MockStore. Один и тот же Service, разные зависимости."},
					{Question: "httptest.NewRecorder() — что это?", Options: []string{"Логгер", "Фейковый http.ResponseWriter — записывает ответ в буфер для проверки в тесте", "HTTP сервер", "Прокси"}, Correct: 1, Explanation: "ResponseRecorder реализует http.ResponseWriter. Handler пишет в него как в настоящий ответ. Потом проверяешь: w.Code, w.Body.String()."},
					{Question: "Как тестировать ошибочные сценарии через mock?", Options: []string{"Никак", "Mock возвращает заданную ошибку: MockStore{err: ErrNotFound} — тестируй обработку ошибок", "Только в интеграции", "Через panic"}, Correct: 1, Explanation: "Mock с подставленной ошибкой проверяет: правильно ли сервис обрабатывает failures. Без mock пришлось бы ломать БД для тестирования ошибок."},
					{Question: "Mock vs Stub vs Fake?", Options: []string{"Одно и то же", "Stub — фиксированный ответ. Mock — проверяет вызовы. Fake — упрощённая реализация (in-memory DB)", "Fake устарел", "Mock для Go, Stub для Java"}, Correct: 1, Explanation: "Stub: GetByID() always returns Alice. Mock: verify GetByID was called with id=1. Fake: real in-memory map. В Go чаще всего пишут fakes (просто struct с map)."},
				},
				Tasks: []T{
					{Title: "Simple mock", Difficulty: "easy", Description: `<p>Создай mock для UserStore и протестируй GetUser:</p><p>Вывод: <code>PASS: mock test</code></p>`, Glossary: []GlossaryItem{{Term: "Mock struct", Definition: "Структура с предопределёнными ответами. Реализует интерфейс."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: mock test"}},
						StarterCode: `package main
import "fmt"
type User struct { ID int; Name string }
type UserStore interface { GetByID(id int) (*User, error) }
type MockStore struct { users map[int]*User }
func (m *MockStore) GetByID(id int) (*User, error) {
    u, ok := m.users[id]; if !ok { return nil, fmt.Errorf("not found") }; return u, nil
}
func main() {
    store := &MockStore{users: map[int]*User{1: {1, "Alice"}}}
    u, err := store.GetByID(1)
    if err != nil || u.Name != "Alice" { fmt.Println("FAIL"); return }
    _, err = store.GetByID(999)
    if err == nil { fmt.Println("FAIL: should error"); return }
    fmt.Println("PASS: mock test")
}`, Hints: `<p>Mock хранит map. GetByID ищет в map. Нет → error.</p>`, Solution: `<pre><code>package main
import "fmt"
type User struct{ID int;Name string}
type UserStore interface{GetByID(int)(*User,error)}
type MockStore struct{users map[int]*User}
func(m *MockStore)GetByID(id int)(*User,error){u,ok:=m.users[id];if !ok{return nil,fmt.Errorf("not found")};return u,nil}
func main(){s:=&MockStore{map[int]*User{1:{1,"Alice"}}};u,err:=s.GetByID(1);if err!=nil||u.Name!="Alice"{fmt.Println("FAIL");return};_,err=s.GetByID(999);if err==nil{fmt.Println("FAIL");return};fmt.Println("PASS: mock test")}</code></pre>`},
					{Title: "Error injection", Difficulty: "easy", Description: `<p>Mock с настраиваемой ошибкой:</p><p>Вывод: <code>PASS: error injection</code></p>`, Glossary: []GlossaryItem{{Term: "err field in mock", Definition: "Поле err в mock: если не nil — mock всегда возвращает эту ошибку."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: error injection"}},
						StarterCode: `package main
import "fmt"
type Fetcher interface { Fetch(url string) (string, error) }
type MockFetcher struct { response string; err error }
func (m *MockFetcher) Fetch(url string) (string, error) { if m.err != nil { return "", m.err }; return m.response, nil }
func main() {
    ok := &MockFetcher{response: "data"}
    fail := &MockFetcher{err: fmt.Errorf("timeout")}
    r, _ := ok.Fetch("url"); if r != "data" { fmt.Println("FAIL"); return }
    _, err := fail.Fetch("url"); if err == nil { fmt.Println("FAIL"); return }
    fmt.Println("PASS: error injection")
}`, Hints: `<p>Mock с полем err. Если err != nil → возвращай ошибку.</p>`, Solution: `<pre><code>package main
import "fmt"
type Fetcher interface{Fetch(string)(string,error)}
type MockFetcher struct{resp string;err error}
func(m *MockFetcher)Fetch(string)(string,error){if m.err!=nil{return "",m.err};return m.resp,nil}
func main(){ok:=&MockFetcher{resp:"data"};fail:=&MockFetcher{err:fmt.Errorf("timeout")};r,_:=ok.Fetch("");if r!="data"{fmt.Println("FAIL");return};_,err:=fail.Fetch("");if err==nil{fmt.Println("FAIL");return};fmt.Println("PASS: error injection")}</code></pre>`},
					{Title: "Service с DI", Difficulty: "medium", Description: `<p>UserService зависит от Store через интерфейс. Тестируй с mock:</p><p>Вывод: <code>PASS: service with DI</code></p>`, Glossary: []GlossaryItem{{Term: "Dependency Injection", Definition: "Зависимость передаётся снаружи (через конструктор). Не создаётся внутри."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: service with DI"}},
						StarterCode: `package main
import "fmt"
type User struct{ Name string }
type Store interface { Get(id int) (*User, error) }
type MockStore struct{ data map[int]*User }
func (m *MockStore) Get(id int) (*User, error) { u, ok := m.data[id]; if !ok { return nil, fmt.Errorf("not found") }; return u, nil }
type Service struct { store Store }
func (s *Service) GetUserName(id int) (string, error) { u, err := s.store.Get(id); if err != nil { return "", err }; return u.Name, nil }
func main() {
    mock := &MockStore{data: map[int]*User{1: {"Alice"}}}
    svc := &Service{store: mock}
    name, err := svc.GetUserName(1)
    if err != nil || name != "Alice" { fmt.Println("FAIL"); return }
    _, err = svc.GetUserName(99)
    if err == nil { fmt.Println("FAIL: should error"); return }
    fmt.Println("PASS: service with DI")
}`, Hints: `<p>Service получает Store через struct field. В тесте подставляешь MockStore.</p>`, Solution: `<pre><code>package main
import "fmt"
type User struct{Name string}
type Store interface{Get(int)(*User,error)}
type MockStore struct{data map[int]*User}
func(m *MockStore)Get(id int)(*User,error){u,ok:=m.data[id];if !ok{return nil,fmt.Errorf("nf")};return u,nil}
type Service struct{store Store}
func(s *Service)GetUserName(id int)(string,error){u,err:=s.store.Get(id);if err!=nil{return "",err};return u.Name,nil}
func main(){m:=&MockStore{map[int]*User{1:{"Alice"}}};svc:=&Service{m};n,err:=svc.GetUserName(1);if err!=nil||n!="Alice"{fmt.Println("FAIL");return};_,err=svc.GetUserName(99);if err==nil{fmt.Println("FAIL");return};fmt.Println("PASS: service with DI")}</code></pre>`},
					{Title: "HTTP handler test", Difficulty: "hard", Description: `<p>Имитируй httptest: проверь что handler возвращает правильный status и body:</p><p>Вывод: <code>PASS: handler test (200, {"status":"ok"})</code></p>`, Glossary: []GlossaryItem{{Term: "httptest.NewRecorder", Definition: "Фейковый ResponseWriter. Записывает status code и body для проверки."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: `PASS: handler test (200, {"status":"ok"})`}},
						StarterCode: `package main
import "fmt"
type Response struct { code int; body string }
func healthHandler() Response { return Response{200, "{\"status\":\"ok\"}"} }
func main() {
    resp := healthHandler()
    if resp.code != 200 { fmt.Println("FAIL: wrong status"); return }
    if resp.body != "{\"status\":\"ok\"}" { fmt.Println("FAIL: wrong body"); return }
    fmt.Printf("PASS: handler test (%d, %s)\n", resp.code, resp.body)
}`, Hints: `<p>В реальности: httptest.NewRequest + httptest.NewRecorder + handler.ServeHTTP.</p>`, Solution: `<pre><code>package main
import "fmt"
type Response struct{code int;body string}
func healthHandler() Response{return Response{200,"{\"status\":\"ok\"}"}}
func main(){r:=healthHandler();if r.code!=200||r.body!="{\"status\":\"ok\"}"{fmt.Println("FAIL");return};fmt.Printf("PASS: handler test (%d, %s)\n",r.code,r.body)}</code></pre>`},
					{Title: "Call tracking mock", Difficulty: "hard", Description: `<p>Mock который отслеживает вызовы (spy pattern):</p><p>Вывод:</p><pre><code>PASS: called 3 times with [1 2 3]</code></pre>`, Glossary: []GlossaryItem{{Term: "Spy/Call tracking", Definition: "Mock записывает все вызовы. В конце проверяешь: сколько раз вызван, с какими аргументами."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "PASS: called 3 times with [1 2 3]"}},
						StarterCode: `package main
import "fmt"
type Logger interface { Log(id int) }
type SpyLogger struct { calls []int }
func (s *SpyLogger) Log(id int) { s.calls = append(s.calls, id) }
func processItems(ids []int, logger Logger) { for _, id := range ids { logger.Log(id) } }
func main() {
    spy := &SpyLogger{}
    processItems([]int{1, 2, 3}, spy)
    if len(spy.calls) != 3 { fmt.Println("FAIL: wrong count"); return }
    fmt.Printf("PASS: called %d times with %v\n", len(spy.calls), spy.calls)
}`, Hints: `<p>SpyLogger записывает каждый вызов в slice. После теста проверяешь slice.</p>`, Solution: `<pre><code>package main
import "fmt"
type Logger interface{Log(int)}
type SpyLogger struct{calls []int}
func(s *SpyLogger)Log(id int){s.calls=append(s.calls,id)}
func processItems(ids []int,l Logger){for _,id:=range ids{l.Log(id)}}
func main(){spy:=&SpyLogger{};processItems([]int{1,2,3},spy);if len(spy.calls)!=3{fmt.Println("FAIL");return};fmt.Printf("PASS: called %d times with %v\n",len(spy.calls),spy.calls)}</code></pre>`},
				},
			},
			{
				Slug: "integration-testing", Title: "Интеграционные тесты", Order: 4,
				Difficulty: "advanced", Track: "backend",
				Content: `<h1>Интеграционные тесты</h1>

<h2>Unit vs Integration</h2>
<pre><code>// Unit: тестирует одну функцию изолированно (mock зависимости)
// Integration: тестирует взаимодействие компонентов (реальная БД, HTTP)

// Пирамида тестов:
//     /\     E2E (мало, медленно)
//    /  \    Integration (средне)
//   /    \   Unit (много, быстро)
//  /______\</code></pre>

<h2>TestMain — setup/teardown</h2>
<pre><code>func TestMain(m *testing.M) {
    // Setup: поднять БД, миграции
    db := setupTestDB()
    defer db.Close()

    // Run all tests
    code := m.Run()

    // Teardown: очистить
    cleanupTestDB(db)
    os.Exit(code)
}</code></pre>

<h2>Build tags для разделения</h2>
<pre><code>//go:build integration

package user_test
// Этот файл компилируется только с: go test -tags=integration
// Unit тесты запускаются всегда, integration — явно</code></pre>

<h2>testcontainers — реальная БД в тесте</h2>
<pre><code>import "github.com/testcontainers/testcontainers-go"

func setupPostgres(t *testing.T) string {
    ctx := context.Background()
    container, _ := postgres.RunContainer(ctx,
        postgres.WithDatabase("test"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
    )
    t.Cleanup(func() { container.Terminate(ctx) })
    connStr, _ := container.ConnectionString(ctx)
    return connStr
}</code></pre>`,

				Quiz: []Q{
					{Question: "Unit vs Integration test?", Options: []string{"Одно и то же", "Unit — изолированно (mocks). Integration — с реальными зависимостями (DB, HTTP)", "Unit медленнее", "Integration проще"}, Correct: 1, Explanation: "Unit: быстрые, изолированные, много. Integration: медленнее, проверяют real взаимодействие, меньше. Оба нужны."},
					{Question: "TestMain — зачем?", Options: []string{"Обязательная функция", "Setup/teardown для ВСЕХ тестов в пакете (поднять БД, очистить после)", "Точка входа", "Для бенчмарков"}, Correct: 1, Explanation: "TestMain запускается один раз перед всеми тестами пакета. Идеален для: поднять testcontainer, запустить миграции, потом m.Run(), потом cleanup."},
					{Question: "//go:build integration — что это?", Options: []string{"Комментарий", "Build tag — файл компилируется только при go test -tags=integration", "Import", "Pragma"}, Correct: 1, Explanation: "Build tags разделяют тесты. go test = только unit. go test -tags=integration = unit + integration. CI может запускать их отдельно."},
					{Question: "testcontainers — зачем если есть mock?", Options: []string{"Мода", "Mock не ловит SQL-ошибки, проблемы миграций, type mismatches. Real DB в контейнере = настоящий тест", "Быстрее", "Проще"}, Correct: 1, Explanation: "Mock: SELECT * FROM users WHERE id=$1 — всегда вернёт что подставил. Real DB: найдёт missing index, wrong column type, broken migration."},
					{Question: "Пирамида тестов — почему Unit внизу?", Options: []string{"Они хуже", "Много unit (быстрые, дешёвые) + немного integration + ещё меньше E2E (медленные, хрупкие)", "Исторически", "Не важен порядок"}, Correct: 1, Explanation: "Unit: ms, 1000+. Integration: seconds, 100. E2E: minutes, 10. Перевёрнутая пирамида (много E2E) = медленный CI, хрупкие тесты."},
				},
				Tasks: []T{
					{Title: "TestMain pattern", Difficulty: "easy", Description: `<p>Имитируй TestMain: setup → run tests → teardown:</p><p>Вывод:</p><pre><code>SETUP: db connected
TEST: user_create PASS
TEST: user_get PASS
TEARDOWN: db closed</code></pre>`, Glossary: []GlossaryItem{{Term: "TestMain", Definition: "func TestMain(m *testing.M) — one-time setup/teardown для пакета."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "SETUP: db connected\nTEST: user_create PASS\nTEST: user_get PASS\nTEARDOWN: db closed"}},
						StarterCode: `package main
import "fmt"
func setup() { fmt.Println("SETUP: db connected") }
func teardown() { fmt.Println("TEARDOWN: db closed") }
func testUserCreate() { fmt.Println("TEST: user_create PASS") }
func testUserGet() { fmt.Println("TEST: user_get PASS") }
func main() { setup(); testUserCreate(); testUserGet(); teardown() }`, Hints: `<p>setup → тесты → teardown. В реальности: TestMain + m.Run().</p>`, Solution: `<pre><code>package main
import "fmt"
func main() { fmt.Println("SETUP: db connected"); fmt.Println("TEST: user_create PASS"); fmt.Println("TEST: user_get PASS"); fmt.Println("TEARDOWN: db closed") }</code></pre>`},
					{Title: "Test isolation", Difficulty: "easy", Description: `<p>Каждый тест создаёт/очищает свои данные (изоляция):</p><p>Вывод:</p><pre><code>test1: created user, verified, cleaned
test2: created user, verified, cleaned
All isolated: true</code></pre>`, Glossary: []GlossaryItem{{Term: "Test isolation", Definition: "Каждый тест работает с чистым состоянием. Не зависит от порядка запуска."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: "test1: created user, verified, cleaned\ntest2: created user, verified, cleaned\nAll isolated: true"}},
						StarterCode: `package main
import "fmt"
type DB struct{ users map[int]string }
func (db *DB) create(id int, name string) { db.users[id] = name }
func (db *DB) get(id int) string { return db.users[id] }
func (db *DB) clean() { db.users = map[int]string{} }
func main() {
    db := &DB{users: map[int]string{}}
    db.create(1, "Alice"); _ = db.get(1); db.clean(); fmt.Println("test1: created user, verified, cleaned")
    db.create(2, "Bob"); _ = db.get(2); db.clean(); fmt.Println("test2: created user, verified, cleaned")
    fmt.Printf("All isolated: %v\n", len(db.users) == 0)
}`, Hints: `<p>Каждый тест: create → verify → clean. После всех: state пустой.</p>`, Solution: `<pre><code>package main
import "fmt"
type DB struct{users map[int]string}
func(db *DB)create(id int,n string){db.users[id]=n}
func(db *DB)clean(){db.users=map[int]string{}}
func main(){db:=&DB{map[int]string{}};db.create(1,"A");db.clean();fmt.Println("test1: created user, verified, cleaned");db.create(2,"B");db.clean();fmt.Println("test2: created user, verified, cleaned");fmt.Printf("All isolated: %v\n",len(db.users)==0)}</code></pre>`},
					{Title: "HTTP integration test", Difficulty: "medium", Description: `<p>Тестируй HTTP handler end-to-end:</p><p>Вывод: <code>PASS: GET /health -> 200 {"ok":true}</code></p>`, Glossary: []GlossaryItem{{Term: "httptest.NewServer", Definition: "Запускает реальный HTTP сервер на random порту для тестов."}}, TestCases: []TestCase{{Input: "", ExpectedOutput: `PASS: GET /health -> 200 {"ok":true}`}},
						StarterCode: `package main
import "fmt"
type Request struct { method, path string }
type Response struct { status int; body string }
func handleRequest(req Request) Response {
    if req.path == "/health" { return Response{200, "{\"ok\":true}"} }
    return Response{404, "not found"}
}
func main() {
    resp := handleRequest(Request{"GET", "/health"})
    if resp.status == 200 && resp.body == "{\"ok\":true}" {
        fmt.Printf("PASS: GET /health -> %d %s\n", resp.status, resp.body)
    } else { fmt.Println("FAIL") }
}`, Hints: `<p>В реальности: httptest.NewServer(handler) → http.Get(server.URL + \"/health\").</p>`, Solution: `<pre><code>package main
import "fmt"
type Request struct{method,path string}
type Response struct{status int;body string}
func handle(r Request)Response{if r.path=="/health"{return Response{200,"{\"ok\":true}"}};return Response{404,"nf"}}
func main(){r:=handle(Request{"GET","/health"});if r.status==200{fmt.Printf("PASS: GET /health -> %d %s\n",r.status,r.body)}else{fmt.Println("FAIL")}}</code></pre>`},
					{Title: "Test with timeout", Difficulty: "medium", Description: `<p>Тест должен завершиться за N ms, иначе FAIL:</p><p>Ввод: <code>fast</code> → <code>PASS: completed in time</code></p><p>Ввод: <code>slow</code> → <code>FAIL: timeout</code></p>`, Glossary: []GlossaryItem{{Term: "context.WithTimeout в тестах", Definition: "Ограничить время теста. Если не успел — context.DeadlineExceeded."}}, TestCases: []TestCase{{Input: "fast", ExpectedOutput: "PASS: completed in time"}, {Input: "slow", ExpectedOutput: "FAIL: timeout"}},
						StarterCode: `package main
import ("context"; "fmt"; "time")
func operation(mode string) string {
    if mode == "slow" { time.Sleep(200 * time.Millisecond) }
    return "done"
}
func main() {
    var mode string; fmt.Scan(&mode)
    ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
    defer cancel()
    ch := make(chan string, 1)
    go func() { ch <- operation(mode) }()
    select {
    case <-ch: fmt.Println("PASS: completed in time")
    case <-ctx.Done(): fmt.Println("FAIL: timeout")
    }
}`, Hints: `<p>context.WithTimeout + select { case result: PASS; case <-ctx.Done(): FAIL }</p>`, Solution: `<pre><code>package main
import("context";"fmt";"time")
func op(m string)string{if m=="slow"{time.Sleep(200*time.Millisecond)};return "done"}
func main(){var m string;fmt.Scan(&m);ctx,cancel:=context.WithTimeout(context.Background(),100*time.Millisecond);defer cancel()
    ch:=make(chan string,1);go func(){ch<-op(m)}();select{case<-ch:fmt.Println("PASS: completed in time");case<-ctx.Done():fmt.Println("FAIL: timeout")}}</code></pre>`},
					{Title: "Test report generator", Difficulty: "hard", Description: `<p>Сгенерируй отчёт о тестах:</p><p>Ввод:</p><pre><code>5
TestAdd PASS
TestSub PASS
TestMul FAIL
TestDiv PASS
TestMod FAIL</code></pre><p>Вывод:</p><pre><code>Results: 3 passed, 2 failed (60%)
Failed: TestMul TestMod</code></pre>`, Glossary: []GlossaryItem{{Term: "Test report", Definition: "Суммарный отчёт: pass/fail counts, failed test names, coverage %."}}, TestCases: []TestCase{{Input: "5\nTestAdd PASS\nTestSub PASS\nTestMul FAIL\nTestDiv PASS\nTestMod FAIL", ExpectedOutput: "Results: 3 passed, 2 failed (60%)\nFailed: TestMul TestMod"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    passed, failed := 0, 0; var failedNames []string
    for i := 0; i < n; i++ { scanner.Scan(); parts := strings.Fields(scanner.Text())
        if parts[1] == "PASS" { passed++ } else { failed++; failedNames = append(failedNames, parts[0]) }
    }
    pct := passed * 100 / (passed + failed)
    fmt.Printf("Results: %d passed, %d failed (%d%%)\n", passed, failed, pct)
    fmt.Printf("Failed: %s\n", strings.Join(failedNames, " "))
}`, Hints: `<p>Считай PASS/FAIL. Собирай имена failed. Процент = passed/total * 100.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os"; "strings")
func main() { var n int; fmt.Scan(&n); sc := bufio.NewScanner(os.Stdin); p, f := 0, 0; var fn []string
    for i := 0; i < n; i++ { sc.Scan(); parts := strings.Fields(sc.Text()); if parts[1]=="PASS" { p++ } else { f++; fn = append(fn, parts[0]) } }
    fmt.Printf("Results: %d passed, %d failed (%d%%)\nFailed: %s\n", p, f, p*100/(p+f), strings.Join(fn, " ")) }</code></pre>`},
				},
			},
			{
				Slug: "benchmarks", Title: "Benchmarks и профилирование", Order: 5,
				Difficulty: "advanced", Track: "backend",
				Content: `<h1>Benchmarks — измерение производительности</h1>

<h2>Синтаксис</h2>
<pre><code>func BenchmarkAdd(b *testing.B) {
    for i := 0; i < b.N; i++ {
        Add(2, 3)
    }
}
// go test -bench=. -benchmem
// BenchmarkAdd-8    1000000000    0.29 ns/op    0 B/op    0 allocs/op</code></pre>

<h2>Чтение результатов</h2>
<pre><code>// BenchmarkConcat-8    5000000    312 ns/op    64 B/op    3 allocs/op
//                      ───────    ──────────   ─────────  ──────────
//                      iterations  time/op     memory/op  allocations/op

// -benchmem показывает аллокации
// Меньше allocs → меньше GC давления → быстрее</code></pre>

<h2>Сравнение подходов</h2>
<pre><code>func BenchmarkStringConcat(b *testing.B) {
    for i := 0; i < b.N; i++ {
        s := ""; for j := 0; j < 100; j++ { s += "x" }
    }
}

func BenchmarkStringBuilder(b *testing.B) {
    for i := 0; i < b.N; i++ {
        var sb strings.Builder
        for j := 0; j < 100; j++ { sb.WriteString("x") }
        _ = sb.String()
    }
}
// Builder: 100x быстрее, 0 лишних аллокаций</code></pre>

<h2>pprof — профилирование</h2>
<pre><code>import _ "net/http/pprof"
// go tool pprof http://localhost:6060/debug/pprof/profile
// top10, web, list functionName</code></pre>`,

				Quiz: []Q{
					{Question: "Что такое b.N в benchmark?", Options: []string{"Константа", "Количество итераций — Go автоматически подбирает для стабильного измерения", "Номер бенчмарка", "Размер данных"}, Correct: 1, Explanation: "Go запускает benchmark несколько раз, увеличивая b.N пока результат не стабилизируется. Не нужно задавать вручную."},
					{Question: "-benchmem показывает что?", Options: []string{"Общую RAM", "Аллокации на операцию (B/op) и количество аллокаций (allocs/op)", "Использование кеша", "GC паузы"}, Correct: 1, Explanation: "B/op = байт на операцию. allocs/op = количество вызовов make/new. Меньше аллокаций = меньше GC = стабильнее latency."},
					{Question: "strings.Builder vs += для конкатенации?", Options: []string{"Одинаково", "Builder в 100x быстрее для множественной конкатенации (0 лишних аллокаций)", "+= быстрее", "Зависит от длины"}, Correct: 1, Explanation: "s += 'x' каждый раз создаёт новую строку (строки immutable). Builder копит в буфер, одна аллокация. Для 100 конкатенаций: += = 100 аллокаций, Builder = 1-2."},
					{Question: "go tool pprof — зачем?", Options: []string{"Тестирование", "Профилирование: найти где программа тратит CPU/память. Визуализация hot paths", "Сборка", "Деплой"}, Correct: 1, Explanation: "pprof: CPU profile показывает где тратится время. Memory profile — где аллокации. Flame graph визуализирует call stack."},
					{Question: "Когда benchmark нужен?", Options: []string{"Всегда", "При оптимизации: сравнить два подхода цифрами, не интуицией", "Только для библиотек", "Перед релизом"}, Correct: 1, Explanation: "'Premature optimization is the root of all evil.' Benchmark когда: 1) есть проблема с производительностью 2) сравниваешь подходы 3) проверяешь что оптимизация помогла."},
				},
				Tasks: []T{
					{Title: "Benchmark simulator", Difficulty: "easy", Description: `<p>Имитируй вывод benchmark:</p><p>Ввод: <code>Add 1000000000 0.29 0 0</code></p><p>Вывод: <code>BenchmarkAdd-8    1000000000    0.29 ns/op    0 B/op    0 allocs/op</code></p>`, Glossary: []GlossaryItem{{Term: "ns/op", Definition: "Наносекунды на операцию. Основная метрика benchmark."}}, TestCases: []TestCase{{Input: "Add 1000000000 0.29 0 0", ExpectedOutput: "BenchmarkAdd-8    1000000000    0.29 ns/op    0 B/op    0 allocs/op"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name string; var iters int; var nsOp float64; var bOp, allocs int
    fmt.Scan(&name, &iters, &nsOp, &bOp, &allocs)
    fmt.Printf("Benchmark%s-8    %d    %.2f ns/op    %d B/op    %d allocs/op\n", name, iters, nsOp, bOp, allocs)
}`, Hints: `<p>Формат: BenchmarkName-CORES iterations time_ns/op bytes/op allocs/op.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n string;var i int;var ns float64;var b,a int;fmt.Scan(&n,&i,&ns,&b,&a);fmt.Printf("Benchmark%s-8    %d    %.2f ns/op    %d B/op    %d allocs/op\n",n,i,ns,b,a)}</code></pre>`},
					{Title: "Сравнение подходов", Difficulty: "easy", Description: `<p>Сравни два подхода по ns/op и выбери лучший:</p><p>Ввод: <code>Concat 312 Builder 3</code></p><p>Вывод: <code>Winner: Builder (104x faster)</code></p>`, Glossary: []GlossaryItem{{Term: "speedup", Definition: "slow_ns / fast_ns = Nx faster. Показывает во сколько раз один подход лучше."}}, TestCases: []TestCase{{Input: "Concat 312 Builder 3", ExpectedOutput: "Winner: Builder (104x faster)"}},
						StarterCode: `package main
import "fmt"
func main() {
    var name1 string; var ns1 int; var name2 string; var ns2 int
    fmt.Scan(&name1, &ns1, &name2, &ns2)
    if ns1 > ns2 { fmt.Printf("Winner: %s (%dx faster)\n", name2, ns1/ns2) } else { fmt.Printf("Winner: %s (%dx faster)\n", name1, ns2/ns1) }
}`, Hints: `<p>Меньше ns/op = быстрее. Speedup = slow/fast.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n1 string;var t1 int;var n2 string;var t2 int;fmt.Scan(&n1,&t1,&n2,&t2);if t1>t2{fmt.Printf("Winner: %s (%dx faster)\n",n2,t1/t2)}else{fmt.Printf("Winner: %s (%dx faster)\n",n1,t2/t1)}}</code></pre>`},
					{Title: "Alloc analyzer", Difficulty: "medium", Description: `<p>Проанализируй аллокации и предложи оптимизацию:</p><p>Ввод: <code>100 5</code> (B/op, allocs/op)</p><p>Вывод:</p><pre><code>Allocations: 100 B/op, 5 allocs/op
Suggestion: pre-allocate slice (make with capacity)</code></pre>`, Glossary: []GlossaryItem{{Term: "allocs/op", Definition: "Количество heap-аллокаций на операцию. Каждая = работа для GC."}}, TestCases: []TestCase{{Input: "100 5", ExpectedOutput: "Allocations: 100 B/op, 5 allocs/op\nSuggestion: pre-allocate slice (make with capacity)"}, {Input: "0 0", ExpectedOutput: "Allocations: 0 B/op, 0 allocs/op\nSuggestion: optimal - zero allocations"}},
						StarterCode: `package main
import "fmt"
func main() {
    var bOp, allocs int; fmt.Scan(&bOp, &allocs)
    fmt.Printf("Allocations: %d B/op, %d allocs/op\n", bOp, allocs)
    if allocs == 0 { fmt.Println("Suggestion: optimal - zero allocations") } else { fmt.Println("Suggestion: pre-allocate slice (make with capacity)") }
}`, Hints: `<p>0 allocs = идеально. >0 = предложи pre-allocate или sync.Pool.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var b,a int;fmt.Scan(&b,&a);fmt.Printf("Allocations: %d B/op, %d allocs/op\n",b,a);if a==0{fmt.Println("Suggestion: optimal - zero allocations")}else{fmt.Println("Suggestion: pre-allocate slice (make with capacity)")}}</code></pre>`},
					{Title: "String concat benchmark", Difficulty: "medium", Description: `<p>Сравни += vs Builder для N конкатенаций:</p><p>Ввод: <code>1000</code></p><p>Вывод:</p><pre><code>+= approach: 1000 allocations
Builder approach: 1 allocation
Builder is 1000x fewer allocations</code></pre>`, Glossary: []GlossaryItem{{Term: "strings.Builder", Definition: "Буферизированная конкатенация. WriteString добавляет без создания новой строки."}}, TestCases: []TestCase{{Input: "1000", ExpectedOutput: "+= approach: 1000 allocations\nBuilder approach: 1 allocation\nBuilder is 1000x fewer allocations"}},
						StarterCode: `package main
import "fmt"
func main() {
    var n int; fmt.Scan(&n)
    fmt.Printf("+= approach: %d allocations\n", n)
    fmt.Printf("Builder approach: 1 allocation\n")
    fmt.Printf("Builder is %dx fewer allocations\n", n)
}`, Hints: `<p>+= creates new string each time = N allocs. Builder = 1 buffer.</p>`, Solution: `<pre><code>package main
import "fmt"
func main(){var n int;fmt.Scan(&n);fmt.Printf("+= approach: %d allocations\nBuilder approach: 1 allocation\nBuilder is %dx fewer allocations\n",n,n)}</code></pre>`},
					{Title: "Performance regression detector", Difficulty: "hard", Description: `<p>Сравни текущий benchmark с baseline — определи regression:</p><p>Ввод:</p><pre><code>3
Add 0.3 0.3
Sort 150 200
Hash 50 45</code></pre><p>Вывод:</p><pre><code>Add: 0.3 -> 0.3 ns/op (no change)
Sort: 150 -> 200 ns/op (REGRESSION +33%)
Hash: 50 -> 45 ns/op (improved -10%)</code></pre>`, Glossary: []GlossaryItem{{Term: "Regression", Definition: "Производительность ухудшилась. >10% degradation = regression. Нужно расследовать."}}, TestCases: []TestCase{{Input: "3\nAdd 0.3 0.3\nSort 150 200\nHash 50 45", ExpectedOutput: "Add: 0.3 -> 0.3 ns/op (no change)\nSort: 150 -> 200 ns/op (REGRESSION +33%)\nHash: 50 -> 45 ns/op (improved -10%)"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os")
func main() {
    var n int; fmt.Scan(&n); scanner := bufio.NewScanner(os.Stdin)
    for i := 0; i < n; i++ { scanner.Scan()
        var name string; var baseline, current float64
        fmt.Sscanf(scanner.Text(), "%s %f %f", &name, &baseline, &current)
        change := int((current - baseline) / baseline * 100)
        if change == 0 { fmt.Printf("%s: %.1f -> %.1f ns/op (no change)\n", name, baseline, current)
        } else if change > 0 { fmt.Printf("%s: %.0f -> %.0f ns/op (REGRESSION +%d%%)\n", name, baseline, current, change)
        } else { fmt.Printf("%s: %.0f -> %.0f ns/op (improved %d%%)\n", name, baseline, current, change) }
    }
}`, Hints: `<p>change = (current-baseline)/baseline * 100. >10% = regression. <0 = improved.</p>`, Solution: `<pre><code>package main
import ("bufio"; "fmt"; "os")
func main(){var n int;fmt.Scan(&n);sc:=bufio.NewScanner(os.Stdin)
    for i:=0;i<n;i++{sc.Scan();var nm string;var b,c float64;fmt.Sscanf(sc.Text(),"%s %f %f",&nm,&b,&c)
        ch:=int((c-b)/b*100);if ch==0{fmt.Printf("%s: %.1f -> %.1f ns/op (no change)\n",nm,b,c)}else if ch>0{fmt.Printf("%s: %.0f -> %.0f ns/op (REGRESSION +%d%%)\n",nm,b,c,ch)}else{fmt.Printf("%s: %.0f -> %.0f ns/op (improved %d%%)\n",nm,b,c,ch)}}}</code></pre>`},
				},
			},
		},
	}
}
