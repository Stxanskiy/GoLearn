package main

// ════════════════════════════════════════════════════════════════
// МОДУЛЬ: Архитектура и SOLID — расширенный (5 уроков)
// Заменяет mod11_architecture()
// ════════════════════════════════════════════════════════════════

func mod_architecture_full() M {
	return M{
		Slug: "architecture", Title: "Архитектура и SOLID", Order: 11,
		Description: "SOLID принципы, чистая архитектура, dependency injection, паттерны проектирования в Go.",
		Track: "backend", Difficulty: "advanced", Prerequisites: []string{"database"},
		Lessons: []L{
			{
				Slug: "solid-principles", Title: "SOLID принципы в Go", Order: 1,
				Difficulty: "intermediate", Track: "backend",
				Content: `<h1>SOLID в Go</h1>

<h2>S — Single Responsibility</h2>
<p>Один struct/package — одна причина для изменения:</p>
<pre><code>// ПЛОХО: handler делает всё
func CreateUser(w http.ResponseWriter, r *http.Request) {
    // парсит JSON + валидирует + хеширует пароль + пишет в БД + отвечает
}

// ХОРОШО: каждый слой — своя ответственность
type UserHandler struct { svc *UserService }     // HTTP
type UserService struct { repo UserRepository }  // бизнес-логика
type UserRepository interface { Create(*User) error } // данные</code></pre>

<h2>O — Open/Closed</h2>
<p>Открыт для расширения, закрыт для модификации:</p>
<pre><code>// ПЛОХО: добавление нового типа уведомлений = менять switch
func Notify(method string, msg string) {
    switch method {
    case "email": sendEmail(msg)
    case "sms": sendSMS(msg)
    // каждый новый тип = менять эту функцию
    }
}

// ХОРОШО: новый тип = новая структура, не трогая существующий код
type Notifier interface { Send(msg string) error }
type EmailNotifier struct{}
type SMSNotifier struct{}
type TelegramNotifier struct{} // новый — без изменения старого кода</code></pre>

<h2>L — Liskov Substitution</h2>
<p>Подтипы должны быть заменяемы базовым типом:</p>
<pre><code>// В Go = интерфейс + реализации ведут себя одинаково
type Storage interface {
    Save(key string, data []byte) error
    Load(key string) ([]byte, error)
}
// FileStorage, S3Storage, MemoryStorage — все заменяемы</code></pre>

<h2>I — Interface Segregation</h2>
<p>Маленькие интерфейсы лучше больших:</p>
<pre><code>// ПЛОХО: один огромный интерфейс
type UserRepository interface {
    Create(*User) error
    GetByID(int) (*User, error)
    Update(*User) error
    Delete(int) error
    ListAll() ([]*User, error)
    Search(query string) ([]*User, error)
}

// ХОРОШО: разделить по потребителям
type UserReader interface { GetByID(int) (*User, error) }
type UserWriter interface { Create(*User) error }
type UserSearcher interface { Search(string) ([]*User, error) }</code></pre>

<h2>D — Dependency Inversion</h2>
<p>Зависи от абстракций, не от реализаций:</p>
<pre><code>// ПЛОХО: Service зависит от конкретной PostgresRepo
type Service struct { repo *PostgresRepo }

// ХОРОШО: Service зависит от интерфейса
type Service struct { repo UserRepository }
// В проде: PostgresRepo. В тестах: MockRepo.</code></pre>`,

				Quiz: []Q{
					{Question: "S в SOLID — что значит?", Options: []string{"Super", "Single Responsibility — один struct/package = одна причина для изменения", "Static", "Secure"}, Correct: 1, Explanation: "UserHandler отвечает за HTTP. UserService — за бизнес-логику. Если меняется формат API — трогаем только handler."},
					{Question: "Зачем маленькие интерфейсы (Interface Segregation)?", Options: []string{"Для красоты", "Потребитель зависит только от нужных методов. Легче мокать, понятнее контракт", "Быстрее", "Go требует"}, Correct: 1, Explanation: "io.Reader = 1 метод. Файл, сеть, буфер — все реализуют. Если бы Reader имел 20 методов — мало кто бы его реализовал."},
					{Question: "Dependency Inversion на практике в Go?", Options: []string{"Наследование", "Struct поле = interface, не конкретный тип. В конструкторе передаётся реализация", "Global variable", "init()"}, Correct: 1, Explanation: "type Service struct { repo Repository } — interface. NewService(repo) — в проде PostgresRepo, в тестах MockRepo."},
					{Question: "Open/Closed в Go — как реализуется?", Options: []string{"abstract class", "Через интерфейсы: новая реализация = новый struct, старый код не меняется", "Через наследование", "Через generics"}, Correct: 1, Explanation: "Добавить TelegramNotifier = создать struct + реализовать Notifier interface. Существующий код (функция notify(Notifier)) не меняется."},
					{Question: "Liskov Substitution в Go?", Options: []string{"Наследование", "Любая реализация интерфейса должна вести себя предсказуемо — не нарушать контракт", "Type assertion", "Embedding"}, Correct: 1, Explanation: "Если функция принимает Storage interface — FileStorage и S3Storage должны работать одинаково. Нельзя чтобы S3Storage паниковал там где FileStorage возвращает error."},
				},
				Tasks: []T{
					{Title: "SRP — разделение", Difficulty: "easy", Description: `<p>Раздели монолитную функцию на 3 ответственности: parse, validate, process:</p><p>Ввод: <code>Alice 25</code></p><p>Вывод: <code>Parsed: Alice,25 | Valid: true | Result: User(Alice,25)</code></p>`, Glossary: []GlossaryItem{{Term: "SRP", Definition: "Single Responsibility. Одна функция — одно дело."}}, TestCases: []TestCase{{Input: "Alice 25", ExpectedOutput: "Parsed: Alice,25 | Valid: true | Result: User(Alice,25)"}},
						StarterCode: `package main
import "fmt"
type User struct{ Name string; Age int }
func parse(name string, age int) User { return User{name, age} }
func validate(u User) bool { return u.Name != "" && u.Age > 0 && u.Age < 150 }
func process(u User) string { return fmt.Sprintf("User(%s,%d)", u.Name, u.Age) }
func main() {
    var name string; var age int; fmt.Scan(&name, &age)
    u := parse(name, age)
    fmt.Printf("Parsed: %s,%d | Valid: %v | Result: %s\n", u.Name, u.Age, validate(u), process(u))
}`, Hints: `<p>Три функции: parse (string→struct), validate (struct→bool), process (struct→string).</p>`, Solution: `<pre><code>package main
import "fmt"
type User struct{Name string;Age int}
func parse(n string,a int)User{return User{n,a}}
func validate(u User)bool{return u.Name!=""&&u.Age>0}
func process(u User)string{return fmt.Sprintf("User(%s,%d)",u.Name,u.Age)}
func main(){var n string;var a int;fmt.Scan(&n,&a);u:=parse(n,a);fmt.Printf("Parsed: %s,%d | Valid: %v | Result: %s\n",u.Name,u.Age,validate(u),process(u))}</code></pre>`},
					{Title: "Open/Closed — новый Notifier", Difficulty: "easy", Description: `<p>Добавь TelegramNotifier без изменения существующего кода:</p><p>Ввод: <code>telegram Hello!</code></p><p>Вывод: <code>[Telegram] Hello!</code></p>`, Glossary: []GlossaryItem{{Term: "OCP", Definition: "Open for extension (новый struct), Closed for modification (старый код не трогаем)."}}, TestCases: []TestCase{{Input: "telegram Hello!", ExpectedOutput: "[Telegram] Hello!"}, {Input: "email Hello!", ExpectedOutput: "[Email] Hello!"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
type Notifier interface { Send(msg string) string }
type EmailNotifier struct{}
func (e EmailNotifier) Send(msg string) string { return "[Email] " + msg }
type TelegramNotifier struct{}
func (t TelegramNotifier) Send(msg string) string { return "[Telegram] " + msg }
func notify(n Notifier, msg string) { fmt.Println(n.Send(msg)) }
func main() {
    scanner := bufio.NewScanner(os.Stdin); scanner.Scan()
    parts := strings.SplitN(scanner.Text(), " ", 2)
    var n Notifier
    switch parts[0] { case "email": n = EmailNotifier{}; case "telegram": n = TelegramNotifier{} }
    notify(n, parts[1])
}`, Hints: `<p>Новый notifier = новый struct + реализация Send(). notify() не меняется.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
type Notifier interface{Send(string)string}
type EmailNotifier struct{}; func(EmailNotifier)Send(m string)string{return "[Email] "+m}
type TelegramNotifier struct{}; func(TelegramNotifier)Send(m string)string{return "[Telegram] "+m}
func notify(n Notifier,m string){fmt.Println(n.Send(m))}
func main(){sc:=bufio.NewScanner(os.Stdin);sc.Scan();p:=strings.SplitN(sc.Text()," ",2);var n Notifier;switch p[0]{case "email":n=EmailNotifier{};case "telegram":n=TelegramNotifier{}};notify(n,p[1])}</code></pre>`},
					{Title: "DI — constructor injection", Difficulty: "medium", Description: `<p>Реализуй Service с DI: принимает Repository через конструктор:</p><p>Ввод: <code>1 Alice</code></p><p>Вывод: <code>Created: Alice (id=1)</code></p>`, Glossary: []GlossaryItem{{Term: "Constructor Injection", Definition: "Зависимость передаётся через NewService(repo). Не создаётся внутри."}}, TestCases: []TestCase{{Input: "1 Alice", ExpectedOutput: "Created: Alice (id=1)"}},
						StarterCode: `package main
import "fmt"
type User struct{ ID int; Name string }
type Repository interface { Save(u User) error }
type MemoryRepo struct{ users []User }
func (r *MemoryRepo) Save(u User) error { r.users = append(r.users, u); return nil }
type Service struct { repo Repository }
func NewService(repo Repository) *Service { return &Service{repo: repo} }
func (s *Service) CreateUser(id int, name string) { s.repo.Save(User{id, name}); fmt.Printf("Created: %s (id=%d)\n", name, id) }
func main() { var id int; var name string; fmt.Scan(&id, &name); svc := NewService(&MemoryRepo{}); svc.CreateUser(id, name) }`, Hints: `<p>NewService принимает interface. В main передаём &MemoryRepo{}. В тестах — &MockRepo{}.</p>`, Solution: `<pre><code>package main
import "fmt"
type User struct{ID int;Name string}
type Repository interface{Save(User)error}
type MemoryRepo struct{users []User}
func(r *MemoryRepo)Save(u User)error{r.users=append(r.users,u);return nil}
type Service struct{repo Repository}
func NewService(r Repository)*Service{return &Service{r}}
func(s *Service)CreateUser(id int,n string){s.repo.Save(User{id,n});fmt.Printf("Created: %s (id=%d)\n",n,id)}
func main(){var id int;var n string;fmt.Scan(&id,&n);NewService(&MemoryRepo{}).CreateUser(id,n)}</code></pre>`},
					{Title: "Interface Segregation", Difficulty: "medium", Description: `<p>Раздели большой интерфейс на мелкие по потребителям:</p><p>Ввод: <code>read 1</code></p><p>Вывод: <code>Read user 1: Alice</code></p><p>Ввод: <code>write Alice</code></p><p>Вывод: <code>Created: Alice</code></p>`, Glossary: []GlossaryItem{{Term: "ISP", Definition: "Interface Segregation. Потребитель зависит только от нужных методов."}}, TestCases: []TestCase{{Input: "read 1", ExpectedOutput: "Read user 1: Alice"}, {Input: "write Bob", ExpectedOutput: "Created: Bob"}},
						StarterCode: `package main
import "fmt"
type UserReader interface { GetByID(id int) string }
type UserWriter interface { Create(name string) }
type UserStore struct { users map[int]string }
func (s *UserStore) GetByID(id int) string { return s.users[id] }
func (s *UserStore) Create(name string) { fmt.Printf("Created: %s\n", name) }
func readUser(r UserReader, id int) { fmt.Printf("Read user %d: %s\n", id, r.GetByID(id)) }
func writeUser(w UserWriter, name string) { w.Create(name) }
func main() {
    store := &UserStore{users: map[int]string{1: "Alice"}}
    var cmd, arg string; fmt.Scan(&cmd, &arg)
    switch cmd {
    case "read": var id int; fmt.Sscan(arg, &id); readUser(store, id)
    case "write": writeUser(store, arg)
    }
}`, Hints: `<p>readUser принимает только UserReader. writeUser — только UserWriter. Один store реализует оба.</p>`, Solution: `<pre><code>package main
import "fmt"
type UserReader interface{GetByID(int)string}
type UserWriter interface{Create(string)}
type UserStore struct{users map[int]string}
func(s *UserStore)GetByID(id int)string{return s.users[id]}
func(s *UserStore)Create(n string){fmt.Printf("Created: %s\n",n)}
func readUser(r UserReader,id int){fmt.Printf("Read user %d: %s\n",id,r.GetByID(id))}
func writeUser(w UserWriter,n string){w.Create(n)}
func main(){store:=&UserStore{map[int]string{1:"Alice"}};var c,a string;fmt.Scan(&c,&a);switch c{case "read":var id int;fmt.Sscan(a,&id);readUser(store,id);case "write":writeUser(store,a)}}</code></pre>`},
					{Title: "Full SOLID refactoring", Difficulty: "hard", Description: `<p>Рефакторинг: монолитный handler → 3 слоя (Handler→Service→Repo):</p><p>Ввод: <code>create Alice</code></p><p>Вывод: <code>[Handler] received: create Alice
[Service] validating: Alice
[Repo] saved: Alice
[Handler] response: 201 Created</code></p>`, Glossary: []GlossaryItem{{Term: "Layered Architecture", Definition: "Handler (HTTP) → Service (бизнес-логика) → Repository (данные). Каждый слой знает только о следующем."}}, TestCases: []TestCase{{Input: "create Alice", ExpectedOutput: "[Handler] received: create Alice\n[Service] validating: Alice\n[Repo] saved: Alice\n[Handler] response: 201 Created"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
type Repo struct{}
func (r *Repo) Save(name string) { fmt.Printf("[Repo] saved: %s\n", name) }
type Service struct { repo *Repo }
func (s *Service) Create(name string) { fmt.Printf("[Service] validating: %s\n", name); s.repo.Save(name) }
type Handler struct { svc *Service }
func (h *Handler) Handle(cmd, arg string) { fmt.Printf("[Handler] received: %s %s\n", cmd, arg); h.svc.Create(arg); fmt.Println("[Handler] response: 201 Created") }
func main() {
    repo := &Repo{}; svc := &Service{repo}; handler := &Handler{svc}
    scanner := bufio.NewScanner(os.Stdin); scanner.Scan()
    parts := strings.SplitN(scanner.Text(), " ", 2)
    handler.Handle(parts[0], parts[1])
}`, Hints: `<p>Handler → Service → Repo. Каждый выводит свою строку. Поток сверху вниз.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
type Repo struct{}; func(*Repo)Save(n string){fmt.Printf("[Repo] saved: %s\n",n)}
type Service struct{repo *Repo}; func(s *Service)Create(n string){fmt.Printf("[Service] validating: %s\n",n);s.repo.Save(n)}
type Handler struct{svc *Service}; func(h *Handler)Handle(c,a string){fmt.Printf("[Handler] received: %s %s\n",c,a);h.svc.Create(a);fmt.Println("[Handler] response: 201 Created")}
func main(){h:=&Handler{&Service{&Repo{}}};sc:=bufio.NewScanner(os.Stdin);sc.Scan();p:=strings.SplitN(sc.Text()," ",2);h.Handle(p[0],p[1])}</code></pre>`},
				},
			},
			{
				Slug: "clean-architecture", Title: "Чистая архитектура на практике", Order: 2,
				Difficulty: "advanced", Track: "backend",
				Content: `<h1>Чистая архитектура в Go</h1>

<h2>Handler → Service → Repository</h2>
<pre><code>// Handler — HTTP слой (парсит запрос, вызывает service, формирует ответ)
// Service — бизнес-логика (валидация, правила, координация)
// Repository — данные (SQL, Redis, файлы)

// Зависимости направлены ВНУТРЬ:
// Handler → знает о Service (через interface)
// Service → знает о Repository (через interface)
// Repository → знает только о DB driver</code></pre>

<h2>Структура проекта</h2>
<pre><code>internal/
  handler/       ← HTTP handlers (зависит от service)
  service/       ← бизнес-логика (зависит от repository interface)
  repository/    ← SQL queries (реализует repository interface)
  model/         ← общие типы данных (не зависит ни от кого)
cmd/
  server/        ← main.go (собирает всё вместе — wiring)</code></pre>

<h2>Wiring в main.go</h2>
<pre><code>func main() {
    // 1. Инфраструктура
    pool := connectDB()

    // 2. Repository (конкретная реализация)
    userRepo := repository.NewPostgresUserRepo(pool)

    // 3. Service (зависит от interface)
    userSvc := service.NewUserService(userRepo)

    // 4. Handler (зависит от service interface)
    userHandler := handler.NewUserHandler(userSvc)

    // 5. Router
    r := chi.NewRouter()
    r.Post("/users", userHandler.Create)
}</code></pre>`,

				Quiz: []Q{
					{Question: "Почему зависимости направлены внутрь?", Options: []string{"Так проще", "Handler может измениться (другой фреймворк) без изменения бизнес-логики. DB может измениться без изменения service", "Go требует", "Для скорости"}, Correct: 1, Explanation: "Бизнес-логика — ядро. Она не должна знать о HTTP или SQL. Если меняем chi на gin — трогаем только handler. Если меняем PostgreSQL на MongoDB — только repository."},
					{Question: "Где происходит wiring (сборка)?", Options: []string{"В каждом пакете", "В main.go — единственное место где знаем о всех конкретных реализациях", "В config", "Автоматически"}, Correct: 1, Explanation: "main.go = composition root. Только здесь создаём PostgresRepo и передаём в Service. Service не знает что это Postgres — видит только interface."},
					{Question: "model/ зависит от чего?", Options: []string{"От repository", "Ни от чего — это чистые типы данных без зависимостей", "От handler", "От всего"}, Correct: 1, Explanation: "model.User — plain struct. Нет импортов из handler/service/repository. Все слои импортируют model, но model не импортирует никого."},
					{Question: "Когда чистая архитектура = оверинжиниринг?", Options: []string{"Никогда", "Маленький проект, один разработчик, прототип, CRUD без бизнес-логики", "Всегда", "Для микросервисов"}, Correct: 1, Explanation: "Для todo-app из 3 endpoints — layers добавят файлов без пользы. Правило: если бизнес-логика = 'сохрани в БД' → не нужны слои. Если есть rules/validation → нужны."},
					{Question: "Service возвращает HTTP-коды?", Options: []string{"Да", "Нет — service возвращает domain errors. Handler маппит их в HTTP-коды", "Зависит от проекта", "Только 500"}, Correct: 1, Explanation: "Service: return ErrNotFound, ErrForbidden. Handler: if errors.Is(err, ErrNotFound) → 404. Service не знает о HTTP — может использоваться из CLI, gRPC, etc."},
				},
				Tasks: []T{
					{Title: "Layered structure", Difficulty: "easy", Description: `<p>Реализуй три слоя: handler → service → repo. Каждый делает своё:</p><p>Ввод: <code>get 1</code></p><p>Вывод:</p><pre><code>[Repo] SELECT * FROM users WHERE id=1
[Service] found user: Alice
[Handler] 200: {"name":"Alice"}</code></pre>`, Glossary: []GlossaryItem{{Term: "Three layers", Definition: "Handler=HTTP, Service=logic, Repo=data. Каждый вызывает следующий."}}, TestCases: []TestCase{{Input: "get 1", ExpectedOutput: "[Repo] SELECT * FROM users WHERE id=1\n[Service] found user: Alice\n[Handler] 200: {\"name\":\"Alice\"}"}},
						StarterCode: `package main
import "fmt"
type Repo struct{}
func (r *Repo) GetByID(id int) string { fmt.Printf("[Repo] SELECT * FROM users WHERE id=%d\n", id); return "Alice" }
type Service struct { repo *Repo }
func (s *Service) GetUser(id int) string { name := s.repo.GetByID(id); fmt.Printf("[Service] found user: %s\n", name); return name }
type Handler struct { svc *Service }
func (h *Handler) Get(id int) { name := h.svc.GetUser(id); fmt.Printf("[Handler] 200: {\"name\":\"%s\"}\n", name) }
func main() { var cmd string; var id int; fmt.Scan(&cmd, &id); h := &Handler{&Service{&Repo{}}}; h.Get(id) }`, Hints: `<p>Handler вызывает Service, Service вызывает Repo. Каждый выводит своё.</p>`, Solution: `<pre><code>package main
import "fmt"
type Repo struct{}; func(*Repo)GetByID(id int)string{fmt.Printf("[Repo] SELECT * FROM users WHERE id=%d\n",id);return "Alice"}
type Service struct{repo *Repo}; func(s *Service)GetUser(id int)string{n:=s.repo.GetByID(id);fmt.Printf("[Service] found user: %s\n",n);return n}
type Handler struct{svc *Service}; func(h *Handler)Get(id int){n:=h.svc.GetUser(id);fmt.Printf("[Handler] 200: {\"name\":\"%s\"}\n",n)}
func main(){var c string;var id int;fmt.Scan(&c,&id);(&Handler{&Service{&Repo{}}}).Get(id)}</code></pre>`},
					{Title: "Error mapping", Difficulty: "medium", Description: `<p>Service возвращает domain error, Handler маппит в HTTP code:</p><p>Ввод: <code>999</code></p><p>Вывод: <code>404: user not found</code></p><p>Ввод: <code>1</code></p><p>Вывод: <code>200: Alice</code></p>`, Glossary: []GlossaryItem{{Term: "Error mapping", Definition: "Service → ErrNotFound. Handler → 404. Service не знает о HTTP."}}, TestCases: []TestCase{{Input: "999", ExpectedOutput: "404: user not found"}, {Input: "1", ExpectedOutput: "200: Alice"}},
						StarterCode: `package main
import ("errors"; "fmt")
var ErrNotFound = errors.New("not found")
type Service struct{}
func (s *Service) GetUser(id int) (string, error) {
    if id == 1 { return "Alice", nil }
    return "", ErrNotFound
}
func main() {
    var id int; fmt.Scan(&id); svc := &Service{}
    name, err := svc.GetUser(id)
    if errors.Is(err, ErrNotFound) { fmt.Println("404: user not found") } else { fmt.Printf("200: %s\n", name) }
}`, Hints: `<p>Service: ErrNotFound. Handler: errors.Is → HTTP code.</p>`, Solution: `<pre><code>package main
import("errors";"fmt")
var ErrNotFound=errors.New("not found")
type Service struct{}
func(*Service)GetUser(id int)(string,error){if id==1{return "Alice",nil};return "",ErrNotFound}
func main(){var id int;fmt.Scan(&id);n,err:=(&Service{}).GetUser(id);if errors.Is(err,ErrNotFound){fmt.Println("404: user not found")}else{fmt.Printf("200: %s\n",n)}}</code></pre>`},
					{Title: "Wiring в main", Difficulty: "medium", Description: `<p>Собери все слои в main (composition root):</p><p>Ввод: <code>create Bob</code></p><p>Вывод: <code>Wired: Repo→Service→Handler
Created: Bob</code></p>`, Glossary: []GlossaryItem{{Term: "Composition root", Definition: "main.go где все зависимости собираются вместе. Единственное место знающее о конкретных типах."}}, TestCases: []TestCase{{Input: "create Bob", ExpectedOutput: "Wired: Repo→Service→Handler\nCreated: Bob"}},
						StarterCode: `package main
import ("bufio"; "fmt"; "os"; "strings")
type Repo struct{}; func(*Repo) Save(n string) { fmt.Printf("Created: %s\n", n) }
type Service struct{repo *Repo}; func(s *Service)Create(n string){s.repo.Save(n)}
type Handler struct{svc *Service}; func(h *Handler)Handle(cmd,arg string){h.svc.Create(arg)}
func main() {
    repo := &Repo{}; svc := &Service{repo}; h := &Handler{svc}
    fmt.Println("Wired: Repo→Service→Handler")
    sc := bufio.NewScanner(os.Stdin); sc.Scan(); p := strings.SplitN(sc.Text(), " ", 2)
    h.Handle(p[0], p[1])
}`, Hints: `<p>main: создать Repo → передать в Service → передать в Handler. Вызвать Handler.</p>`, Solution: `<pre><code>package main
import("bufio";"fmt";"os";"strings")
type Repo struct{}; func(*Repo)Save(n string){fmt.Printf("Created: %s\n",n)}
type Service struct{repo *Repo}; func(s *Service)Create(n string){s.repo.Save(n)}
type Handler struct{svc *Service}; func(h *Handler)Handle(_,a string){h.svc.Create(a)}
func main(){fmt.Println("Wired: Repo→Service→Handler");sc:=bufio.NewScanner(os.Stdin);sc.Scan();p:=strings.SplitN(sc.Text()," ",2);(&Handler{&Service{&Repo{}}}).Handle(p[0],p[1])}</code></pre>`},
					{Title: "Domain events", Difficulty: "hard", Description: `<p>После создания user — отправить event (email, analytics). Через interface:</p><p>Ввод: <code>Alice</code></p><p>Вывод:</p><pre><code>User created: Alice
[Email] Welcome Alice!
[Analytics] new_user: Alice</code></pre>`, Glossary: []GlossaryItem{{Term: "Domain events", Definition: "Service публикует событие. Подписчики (email, analytics) реагируют. Loose coupling."}}, TestCases: []TestCase{{Input: "Alice", ExpectedOutput: "User created: Alice\n[Email] Welcome Alice!\n[Analytics] new_user: Alice"}},
						StarterCode: `package main
import "fmt"
type EventHandler interface { Handle(event string, data string) }
type EmailHandler struct{}; func (EmailHandler) Handle(_, data string) { fmt.Printf("[Email] Welcome %s!\n", data) }
type AnalyticsHandler struct{}; func (AnalyticsHandler) Handle(event, data string) { fmt.Printf("[Analytics] %s: %s\n", event, data) }
type Service struct { handlers []EventHandler }
func (s *Service) CreateUser(name string) {
    fmt.Printf("User created: %s\n", name)
    for _, h := range s.handlers { h.Handle("new_user", name) }
}
func main() {
    var name string; fmt.Scan(&name)
    svc := &Service{handlers: []EventHandler{EmailHandler{}, AnalyticsHandler{}}}
    svc.CreateUser(name)
}`, Hints: `<p>Service хранит []EventHandler. После действия — вызывает все handlers. Loose coupling.</p>`, Solution: `<pre><code>package main
import "fmt"
type EventHandler interface{Handle(string,string)}
type EmailHandler struct{}; func(EmailHandler)Handle(_,d string){fmt.Printf("[Email] Welcome %s!\n",d)}
type AnalyticsHandler struct{}; func(AnalyticsHandler)Handle(e,d string){fmt.Printf("[Analytics] %s: %s\n",e,d)}
type Service struct{handlers []EventHandler}
func(s *Service)CreateUser(n string){fmt.Printf("User created: %s\n",n);for _,h:=range s.handlers{h.Handle("new_user",n)}}
func main(){var n string;fmt.Scan(&n);(&Service{[]EventHandler{EmailHandler{},AnalyticsHandler{}}}).CreateUser(n)}</code></pre>`},
					{Title: "Full clean arch", Difficulty: "hard", Description: `<p>Полная реализация: Handler→Service→Repo с interfaces, DI, error mapping:</p><p>Ввод: <code>get 1</code> → <code>200: {"id":1,"name":"Alice"}</code></p><p>Ввод: <code>get 99</code> → <code>404: not found</code></p>`, Glossary: []GlossaryItem{{Term: "Clean Architecture", Definition: "Полный pipeline: HTTP→бизнес-логика→данные через interfaces."}}, TestCases: []TestCase{{Input: "get 1", ExpectedOutput: "200: {\"id\":1,\"name\":\"Alice\"}"}, {Input: "get 99", ExpectedOutput: "404: not found"}},
						StarterCode: `package main
import ("errors"; "fmt")
var ErrNotFound = errors.New("not found")
type User struct{ ID int; Name string }
type UserRepo interface { FindByID(int) (*User, error) }
type MemRepo struct{ users map[int]*User }
func (r *MemRepo) FindByID(id int) (*User, error) { u, ok := r.users[id]; if !ok { return nil, ErrNotFound }; return u, nil }
type UserService struct { repo UserRepo }
func (s *UserService) Get(id int) (*User, error) { return s.repo.FindByID(id) }
type UserHandler struct { svc *UserService }
func (h *UserHandler) Get(id int) {
    u, err := h.svc.Get(id)
    if errors.Is(err, ErrNotFound) { fmt.Println("404: not found"); return }
    fmt.Printf("200: {\"id\":%d,\"name\":\"%s\"}\n", u.ID, u.Name)
}
func main() {
    repo := &MemRepo{users: map[int]*User{1: {1, "Alice"}}}
    svc := &UserService{repo}; h := &UserHandler{svc}
    var cmd string; var id int; fmt.Scan(&cmd, &id); h.Get(id)
}`, Hints: `<p>Repo: interface. Service: зависит от interface. Handler: маппит errors→HTTP. Wiring в main.</p>`, Solution: `<pre><code>package main
import("errors";"fmt")
var ErrNotFound=errors.New("not found")
type User struct{ID int;Name string}
type UserRepo interface{FindByID(int)(*User,error)}
type MemRepo struct{u map[int]*User}; func(r *MemRepo)FindByID(id int)(*User,error){u,ok:=r.u[id];if !ok{return nil,ErrNotFound};return u,nil}
type Svc struct{repo UserRepo}; func(s *Svc)Get(id int)(*User,error){return s.repo.FindByID(id)}
type Handler struct{svc *Svc}
func(h *Handler)Get(id int){u,err:=h.svc.Get(id);if errors.Is(err,ErrNotFound){fmt.Println("404: not found");return};fmt.Printf("200: {\"id\":%d,\"name\":\"%s\"}\n",u.ID,u.Name)}
func main(){var c string;var id int;fmt.Scan(&c,&id);(&Handler{&Svc{&MemRepo{map[int]*User{1:{1,"Alice"}}}}}).Get(id)}</code></pre>`},
				},
			},
		},
	}
}
