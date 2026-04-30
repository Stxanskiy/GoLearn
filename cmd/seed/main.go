package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://golearn:golearn@localhost:5433/golearn?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal("connect:", err)
	}
	defer pool.Close()

	// Clear existing data
	pool.Exec(ctx, "DELETE FROM progress")
	pool.Exec(ctx, "DELETE FROM quiz_questions")
	pool.Exec(ctx, "DELETE FROM quizzes")
	pool.Exec(ctx, "DELETE FROM tasks")
	pool.Exec(ctx, "DELETE FROM lessons")
	pool.Exec(ctx, "DELETE FROM modules")

	// Reset sequences
	pool.Exec(ctx, "ALTER SEQUENCE modules_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE lessons_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE quizzes_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE quiz_questions_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE tasks_id_seq RESTART WITH 1")

	modules := getModules()

	for _, mod := range modules {
		var moduleID int
		err := pool.QueryRow(ctx,
			`INSERT INTO modules (slug, title, description, order_num) VALUES ($1, $2, $3, $4) RETURNING id`,
			mod.Slug, mod.Title, mod.Description, mod.Order).Scan(&moduleID)
		if err != nil {
			log.Fatalf("insert module %s: %v", mod.Slug, err)
		}
		fmt.Printf("Module: %s (id=%d)\n", mod.Title, moduleID)

		for _, lesson := range mod.Lessons {
			var lessonID int
			err := pool.QueryRow(ctx,
				`INSERT INTO lessons (module_id, slug, title, content, order_num) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
				moduleID, lesson.Slug, lesson.Title, lesson.Content, lesson.Order).Scan(&lessonID)
			if err != nil {
				log.Fatalf("insert lesson %s: %v", lesson.Slug, err)
			}
			fmt.Printf("  Lesson %d: %s\n", lesson.Order, lesson.Title)

			if len(lesson.QuizQuestions) > 0 {
				var quizID int
				err := pool.QueryRow(ctx,
					`INSERT INTO quizzes (lesson_id, title) VALUES ($1, $2) RETURNING id`,
					lessonID, "Quiz: "+lesson.Title).Scan(&quizID)
				if err != nil {
					log.Fatalf("insert quiz: %v", err)
				}

				for qi, q := range lesson.QuizQuestions {
					optJSON, _ := json.Marshal(q.Options)
					_, err := pool.Exec(ctx,
						`INSERT INTO quiz_questions (quiz_id, question, options, correct_index, explanation, order_num) VALUES ($1, $2, $3, $4, $5, $6)`,
						quizID, q.Question, optJSON, q.Correct, q.Explanation, qi+1)
					if err != nil {
						log.Fatalf("insert question: %v", err)
					}
				}
				fmt.Printf("    Quiz: %d questions\n", len(lesson.QuizQuestions))
			}

			for ti, t := range lesson.Tasks {
				_, err := pool.Exec(ctx,
					`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num) VALUES ($1, $2, $3, $4, $5, $6)`,
					lessonID, t.Title, t.Description, t.Hints, t.Solution, ti+1)
				if err != nil {
					log.Fatalf("insert task: %v", err)
				}
			}
			if len(lesson.Tasks) > 0 {
				fmt.Printf("    Tasks: %d\n", len(lesson.Tasks))
			}
		}
	}

	fmt.Println("\nSeed completed successfully!")
}

type SeedModule struct {
	Slug        string
	Title       string
	Description string
	Order       int
	Lessons     []SeedLesson
}

type SeedLesson struct {
	Slug          string
	Title         string
	Content       string
	Order         int
	QuizQuestions []SeedQuestion
	Tasks         []SeedTask
}

type SeedQuestion struct {
	Question    string
	Options     []string
	Correct     int
	Explanation string
}

type SeedTask struct {
	Title       string
	Description string
	Hints       string
	Solution    string
}

func getModules() []SeedModule {
	return []SeedModule{
		module1_GoFundamentals(),
		module2_HTTPAndRouting(),
		module3_DatabaseAndSQL(),
		module4_Architecture(),
		module5_Testing(),
		module6_AuthAndSecurity(),
		module7_Concurrency(),
		module8_DevOpsFoundations(),
		module9_CICD(),
		module10_Monitoring(),
		module11_Advanced(),
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 1: Go Fundamentals
// ═══════════════════════════════════════════════════════════

func module1_GoFundamentals() SeedModule {
	return SeedModule{
		Slug:        "go-fundamentals",
		Title:       "Module 1: Go Fundamentals",
		Description: "Core Go language: types, structs, interfaces, error handling, packages. The foundation for everything else.",
		Order:       1,
		Lessons: []SeedLesson{
			{
				Slug:  "go-toolchain",
				Title: "Go Toolchain & Project Setup",
				Order: 1,
				Content: `<h1>Go Toolchain & Project Setup</h1>

<h2>What You'll Learn</h2>
<ul>
<li>How Go organizes code with modules and packages</li>
<li>Essential CLI commands: <code>go mod</code>, <code>go build</code>, <code>go run</code>, <code>go vet</code></li>
<li>How to read Go documentation and understand function signatures</li>
</ul>

<h2>Go Modules — The Foundation</h2>
<p>Every Go project starts with <code>go mod init</code>. This creates a <code>go.mod</code> file — the heart of dependency management.</p>

<pre><code>go mod init github.com/yourname/projectname</code></pre>

<p>What <code>go.mod</code> actually contains:</p>
<pre><code>module github.com/yourname/projectname

go 1.22

require (
    github.com/go-chi/chi/v5 v5.0.12
    github.com/jackc/pgx/v5 v5.5.5
)</code></pre>

<p><strong>Key insight:</strong> the module path is NOT a URL that must exist. It's an identifier. Convention is <code>github.com/user/repo</code> but it could be <code>mycompany.com/internal/tool</code>.</p>

<h2>Package System</h2>
<p>In Go, every <code>.go</code> file belongs to a package. The package name is declared at the top of the file:</p>
<pre><code>package main  // executable
package handler  // library package</code></pre>

<p><strong>Rules you must know:</strong></p>
<ul>
<li>All files in the same directory must have the same package name</li>
<li><code>package main</code> + <code>func main()</code> = executable binary</li>
<li>Uppercase names are exported (public), lowercase are unexported (private)</li>
<li>There's no "protected" — it's either exported or not</li>
</ul>

<h2>How to Read Go Documentation</h2>
<p>This is a critical skill. When you encounter a new package, here's the process:</p>

<pre><code># View docs in terminal
go doc fmt.Println
go doc net/http.ListenAndServe

# Open in browser
go doc -all net/http</code></pre>

<p>When you see a function signature like:</p>
<pre><code>func ListenAndServe(addr string, handler Handler) error</code></pre>

<p>Read it as:</p>
<ul>
<li><code>addr string</code> — takes a string (like ":8080")</li>
<li><code>handler Handler</code> — takes something implementing the Handler interface</li>
<li><code>error</code> — returns an error (nil if success)</li>
</ul>

<p><strong>Pro tip:</strong> Always check what an interface requires. <code>Handler</code> interface needs just one method: <code>ServeHTTP(ResponseWriter, *Request)</code>.</p>

<h2>Essential Commands</h2>
<pre><code># Run without building
go run cmd/server/main.go

# Build binary
go build -o myapp cmd/server/main.go

# Download dependencies
go mod tidy

# Check for bugs
go vet ./...

# Format code
gofmt -w .

# Run tests
go test ./...

# See what go vet checks
go vet -help</code></pre>

<h2>Standard Project Layout</h2>
<pre><code>project/
├── cmd/
│   └── server/
│       └── main.go        # Entry point
├── internal/              # Private packages (can't be imported by others)
│   ├── handler/           # HTTP handlers
│   ├── service/           # Business logic
│   ├── repository/        # Database layer
│   ├── model/             # Data structures
│   └── config/            # Configuration
├── pkg/                   # Public packages (can be imported)
├── migrations/            # SQL migrations
├── go.mod
├── go.sum
└── Makefile</code></pre>

<p><strong>Why internal/?</strong> Go compiler enforces that packages under <code>internal/</code> cannot be imported by code outside the module. This is a hard guarantee, not a convention.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What does `go mod init` create?",
						Options:     []string{"main.go file", "go.mod file", "go.sum file", "Makefile"},
						Correct:     1,
						Explanation: "go mod init creates go.mod — the file that defines the module path and tracks dependencies.",
					},
					{
						Question:    "How does Go determine if a name is exported (public)?",
						Options:     []string{"Using the `public` keyword", "First letter is uppercase", "Using the `export` keyword", "Placing it in a pkg/ directory"},
						Correct:     1,
						Explanation: "In Go, any identifier starting with an uppercase letter is exported. This is a language rule, not a convention.",
					},
					{
						Question:    "What is special about the `internal/` directory?",
						Options:     []string{"It's just a naming convention", "Go compiler prevents importing from outside the module", "Files are compiled differently", "It's hidden from go doc"},
						Correct:     1,
						Explanation: "The Go compiler enforces that packages under internal/ cannot be imported by code outside the parent of internal/.",
					},
					{
						Question:    "What does `go mod tidy` do?",
						Options:     []string{"Formats go.mod nicely", "Adds missing and removes unused dependencies", "Updates all dependencies to latest", "Cleans the build cache"},
						Correct:     1,
						Explanation: "go mod tidy scans your code, adds any missing dependencies to go.mod, and removes any that are no longer needed.",
					},
					{
						Question:    "Which command checks for common bugs without running the code?",
						Options:     []string{"go test", "go build", "go vet", "go lint"},
						Correct:     2,
						Explanation: "go vet analyzes code for common mistakes like unreachable code, wrong printf format strings, etc.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Initialize WatchTogether project",
						Description: `<p>Your first task: set up the WatchTogether project structure.</p>
<ol>
<li>Navigate to <code>GolandProjects/WatchTogether/</code></li>
<li>Verify <code>go.mod</code> exists and check its contents</li>
<li>Create the directory structure: <code>cmd/server/</code>, <code>internal/{handler,service,repository,model,config}/</code></li>
<li>Create <code>cmd/server/main.go</code> with a minimal <code>package main</code> that prints "WatchTogether starting..."</li>
<li>Run it with <code>go run cmd/server/main.go</code></li>
</ol>`,
						Hints: `<p>Use <code>mkdir -p</code> to create nested directories. The main.go file needs <code>package main</code> and <code>func main()</code>.</p>`,
						Solution: `<pre><code>// cmd/server/main.go
package main

import "fmt"

func main() {
    fmt.Println("WatchTogether starting...")
}</code></pre>`,
					},
				},
			},
			{
				Slug:  "types-and-structs",
				Title: "Types, Structs & Methods",
				Order: 2,
				Content: `<h1>Types, Structs & Methods</h1>

<h2>Why Types Matter</h2>
<p>Go is statically typed. Every variable has a type known at compile time. This catches bugs before your code runs.</p>

<h2>Basic Types</h2>
<pre><code>// Numeric
var i int        // platform-dependent (32 or 64 bit)
var i64 int64    // always 64 bit
var f float64    // double precision

// String (immutable, UTF-8)
var s string

// Boolean
var b bool

// Byte and Rune
var by byte   // alias for uint8
var r rune    // alias for int32 (Unicode code point)</code></pre>

<p><strong>Critical nuance:</strong> <code>string</code> in Go is immutable. When you "modify" a string, you create a new one. For building strings in loops, use <code>strings.Builder</code>:</p>
<pre><code>var b strings.Builder
for i := 0; i < 1000; i++ {
    b.WriteString("hello ")
}
result := b.String() // one allocation</code></pre>

<h2>Structs — Custom Types</h2>
<pre><code>type User struct {
    ID        int64     ` + "`" + `json:"id" db:"id"` + "`" + `
    Username  string    ` + "`" + `json:"username" db:"username"` + "`" + `
    Email     string    ` + "`" + `json:"email" db:"email"` + "`" + `
    CreatedAt time.Time ` + "`" + `json:"created_at" db:"created_at"` + "`" + `
}</code></pre>

<p><strong>Struct tags</strong> (the backtick parts) are metadata. Libraries use them via reflection:</p>
<ul>
<li><code>json:"id"</code> — used by <code>encoding/json</code> for JSON field names</li>
<li><code>db:"id"</code> — used by database libraries like sqlx</li>
<li><code>json:"-"</code> — skip this field in JSON</li>
<li><code>json:"name,omitempty"</code> — omit if zero value</li>
</ul>

<h2>Methods</h2>
<p>Methods are functions with a receiver:</p>
<pre><code>// Value receiver — works on a copy
func (u User) FullName() string {
    return u.FirstName + " " + u.LastName
}

// Pointer receiver — can modify the struct
func (u *User) SetEmail(email string) {
    u.Email = email
}</code></pre>

<p><strong>When to use pointer receiver:</strong></p>
<ul>
<li>When the method needs to modify the struct</li>
<li>When the struct is large (avoids copying)</li>
<li><strong>Convention:</strong> if any method uses pointer receiver, ALL methods should</li>
</ul>

<h2>Zero Values</h2>
<p>Every type has a zero value. No nulls, no undefined:</p>
<pre><code>var i int       // 0
var s string    // ""
var b bool      // false
var p *int      // nil
var sl []int    // nil (but len(sl) == 0, works fine)
var m map[string]int  // nil (DANGER: writing to nil map panics!)</code></pre>

<p><strong>Map trap:</strong> Always initialize maps before writing:</p>
<pre><code>// WRONG — panics at runtime
var m map[string]int
m["key"] = 1  // panic: assignment to entry in nil map

// CORRECT
m := make(map[string]int)
m["key"] = 1  // works</code></pre>

<h2>Type Aliases vs Type Definitions</h2>
<pre><code>type UserID int64  // New type — cannot mix with int64 without conversion
type Byte = uint8  // Alias — same type, different name</code></pre>

<p>Type definitions create type safety. <code>UserID(42)</code> and <code>int64(42)</code> are different types.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What happens when you write to a nil map?",
						Options:     []string{"Nothing, the entry is ignored", "The map is auto-initialized", "Runtime panic", "Compile error"},
						Correct:     2,
						Explanation: "Writing to a nil map causes a runtime panic. Always use make(map[K]V) or a map literal to initialize.",
					},
					{
						Question:    "When should you use a pointer receiver?",
						Options:     []string{"Always", "When the method reads data", "When the method needs to modify the struct or the struct is large", "Never, Go handles it automatically"},
						Correct:     2,
						Explanation: "Pointer receivers are needed when the method modifies the receiver or when the struct is large enough that copying would be expensive.",
					},
					{
						Question:    "What does `json:\"-\"` in a struct tag do?",
						Options:     []string{"Names the JSON field '-'", "Skips the field during JSON encoding/decoding", "Makes the field required", "Sets default value to '-'"},
						Correct:     1,
						Explanation: "The special tag json:\"-\" tells encoding/json to completely skip this field during both marshaling and unmarshaling.",
					},
					{
						Question:    "What is the zero value of a string in Go?",
						Options:     []string{"nil", "null", "\"\" (empty string)", "undefined"},
						Correct:     2,
						Explanation: "The zero value of string is an empty string \"\". Go has no null or undefined for value types.",
					},
					{
						Question:    "What is the difference between `type UserID int64` and `type UserID = int64`?",
						Options:     []string{"No difference", "First creates a new type, second creates an alias", "First is faster", "Second can't have methods"},
						Correct:     1,
						Explanation: "type UserID int64 creates a new distinct type that cannot be used interchangeably with int64. type UserID = int64 creates an alias — they are the same type.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Create WatchTogether models",
						Description: `<p>Create the core data models for WatchTogether in <code>internal/model/</code>:</p>
<ol>
<li><code>Video</code> — ID, Title, FilePath, Duration, Size, MimeType, CreatedAt</li>
<li><code>Room</code> — ID, Name, VideoID, CreatedAt, IsActive</li>
<li><code>RoomMember</code> — ID, RoomID, Username, JoinedAt</li>
</ol>
<p>Use proper struct tags for JSON and db. Add a method <code>SizeHuman() string</code> on Video that returns human-readable size (e.g., "1.5 GB").</p>`,
						Hints: `<p>For SizeHuman, divide bytes by 1024 repeatedly. Use fmt.Sprintf("%.1f GB", ...). Think about which receiver type to use.</p>`,
						Solution: `<pre><code>package model

import (
    "fmt"
    "time"
)

type Video struct {
    ID        int64     ` + "`" + `json:"id" db:"id"` + "`" + `
    Title     string    ` + "`" + `json:"title" db:"title"` + "`" + `
    FilePath  string    ` + "`" + `json:"-" db:"file_path"` + "`" + `
    Duration  int       ` + "`" + `json:"duration" db:"duration"` + "`" + `
    Size      int64     ` + "`" + `json:"size" db:"size"` + "`" + `
    MimeType  string    ` + "`" + `json:"mime_type" db:"mime_type"` + "`" + `
    CreatedAt time.Time ` + "`" + `json:"created_at" db:"created_at"` + "`" + `
}

func (v Video) SizeHuman() string {
    const (
        KB = 1024
        MB = KB * 1024
        GB = MB * 1024
    )
    switch {
    case v.Size >= GB:
        return fmt.Sprintf("%.1f GB", float64(v.Size)/float64(GB))
    case v.Size >= MB:
        return fmt.Sprintf("%.1f MB", float64(v.Size)/float64(MB))
    default:
        return fmt.Sprintf("%.1f KB", float64(v.Size)/float64(KB))
    }
}</code></pre>`,
					},
				},
			},
			{
				Slug:  "interfaces",
				Title: "Interfaces & Dependency Injection",
				Order: 3,
				Content: `<h1>Interfaces & Dependency Injection</h1>

<h2>The Most Important Concept in Go</h2>
<p>Interfaces in Go are <strong>implicit</strong>. You don't write "implements". If your struct has the right methods, it satisfies the interface. Period.</p>

<pre><code>type Reader interface {
    Read(p []byte) (n int, err error)
}

// os.File satisfies Reader because it has a Read method
// bytes.Buffer satisfies Reader because it has a Read method
// http.Response.Body satisfies Reader because it has a Read method
// No "implements" keyword needed!</code></pre>

<h2>Why Interfaces Matter for Real Projects</h2>
<p>Without interfaces, your handler depends on a concrete repository:</p>
<pre><code>// BAD — tightly coupled
type UserHandler struct {
    repo *UserRepository  // concrete type
}
// Can't test without a real database!</code></pre>

<p>With interfaces:</p>
<pre><code>// GOOD — loosely coupled
type UserStore interface {
    GetByID(ctx context.Context, id int64) (*User, error)
    Create(ctx context.Context, u *User) error
}

type UserHandler struct {
    store UserStore  // interface — any implementation works
}
// In production: real database
// In tests: mock/fake</code></pre>

<h2>Interface Design Rules</h2>
<ul>
<li><strong>Keep interfaces small.</strong> 1-3 methods ideal. Go's io.Reader has ONE method.</li>
<li><strong>Define interfaces where they're used</strong>, not where they're implemented. The consumer decides what it needs.</li>
<li><strong>Accept interfaces, return structs.</strong> Functions should take interfaces as parameters but return concrete types.</li>
</ul>

<h2>The Empty Interface</h2>
<pre><code>interface{}  // Go < 1.18
any          // Go >= 1.18 (alias for interface{})</code></pre>
<p>Any type satisfies the empty interface. Use sparingly — you lose type safety.</p>

<h2>Type Assertions</h2>
<pre><code>var i interface{} = "hello"

s := i.(string)        // panics if not string
s, ok := i.(string)    // safe — ok is false if not string

// Type switch
switch v := i.(type) {
case string:
    fmt.Println("string:", v)
case int:
    fmt.Println("int:", v)
default:
    fmt.Println("unknown")
}</code></pre>

<h2>Dependency Injection Pattern</h2>
<p>DI is NOT a framework in Go. It's just passing dependencies through constructors:</p>
<pre><code>func NewUserService(store UserStore, logger *slog.Logger) *UserService {
    return &UserService{
        store:  store,
        logger: logger,
    }
}

// In main.go — wire everything together
func main() {
    db := connectDB()
    userRepo := repository.NewUserRepo(db)
    userService := service.NewUserService(userRepo, logger)
    userHandler := handler.NewUserHandler(userService)
}</code></pre>

<p><strong>The key insight:</strong> main.go is the "composition root" — the only place that knows about all concrete types. Everything else works with interfaces.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "How does a type implement an interface in Go?",
						Options:     []string{"Using the `implements` keyword", "By registering with the runtime", "By having all the methods the interface requires", "By embedding the interface"},
						Correct:     2,
						Explanation: "Go interfaces are implicit. A type implements an interface by simply having all the required methods with the correct signatures.",
					},
					{
						Question:    "Where should you define interfaces?",
						Options:     []string{"In the same package as the implementation", "In a shared interfaces package", "Where they are used (consumer side)", "In main.go"},
						Correct:     2,
						Explanation: "Go best practice: define interfaces where they are consumed. The consumer knows what methods it needs, not the implementor.",
					},
					{
						Question:    "What does `s, ok := i.(string)` do if i is not a string?",
						Options:     []string{"Panics", "Returns empty string and false", "Returns nil and false", "Compile error"},
						Correct:     1,
						Explanation: "The two-value type assertion returns the zero value of the type and false if the assertion fails. It never panics.",
					},
					{
						Question:    "What is the Go way of Dependency Injection?",
						Options:     []string{"A DI framework like Spring", "Annotation-based injection", "Passing dependencies through constructor functions", "Global service locator"},
						Correct:     2,
						Explanation: "In Go, DI is simply passing interfaces through constructor functions. No framework needed. main.go wires everything together.",
					},
					{
						Question:    "What does 'Accept interfaces, return structs' mean?",
						Options:     []string{"All parameters must be interfaces", "Function parameters should be interfaces for flexibility, but return concrete types for usability", "Never return interfaces", "Structs are faster than interfaces"},
						Correct:     1,
						Explanation: "This principle means your functions should accept interfaces (flexible, testable) but return concrete types (the caller can always assign to an interface variable if needed).",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Add interfaces to WatchTogether",
						Description: `<p>Refactor WatchTogether to use interfaces:</p>
<ol>
<li>In <code>internal/service/</code>, define a <code>VideoStore</code> interface with methods: <code>GetAll</code>, <code>GetByID</code>, <code>Create</code></li>
<li>Create <code>VideoService</code> struct that accepts <code>VideoStore</code> interface</li>
<li>Create <code>NewVideoService(store VideoStore) *VideoService</code> constructor</li>
<li>Implement a fake in-memory store for testing</li>
</ol>`,
						Hints: `<p>The interface should be defined in the service package, not the repository package. The in-memory implementation can use a simple slice with a mutex for thread safety.</p>`,
						Solution: `<pre><code>// internal/service/video.go
package service

import "context"
import "github.com/backendraz/watchtogether/internal/model"

type VideoStore interface {
    GetAll(ctx context.Context) ([]model.Video, error)
    GetByID(ctx context.Context, id int64) (*model.Video, error)
    Create(ctx context.Context, v *model.Video) error
}

type VideoService struct {
    store VideoStore
}

func NewVideoService(store VideoStore) *VideoService {
    return &VideoService{store: store}
}

func (s *VideoService) ListVideos(ctx context.Context) ([]model.Video, error) {
    return s.store.GetAll(ctx)
}</code></pre>`,
					},
				},
			},
			{
				Slug:  "error-handling",
				Title: "Error Handling Deep Dive",
				Order: 4,
				Content: `<h1>Error Handling Deep Dive</h1>

<h2>Errors Are Values</h2>
<p>Go doesn't have exceptions. Errors are just values that implement the <code>error</code> interface:</p>
<pre><code>type error interface {
    Error() string
}</code></pre>

<p>Every function that can fail returns an error as the last return value:</p>
<pre><code>file, err := os.Open("config.json")
if err != nil {
    return fmt.Errorf("open config: %w", err)
}
defer file.Close()</code></pre>

<h2>Error Wrapping with %w</h2>
<p>When you pass an error up the call stack, wrap it with context:</p>
<pre><code>func (r *UserRepo) GetByID(ctx context.Context, id int64) (*User, error) {
    var u User
    err := r.pool.QueryRow(ctx, "SELECT ...", id).Scan(&u.ID, &u.Name)
    if err != nil {
        return nil, fmt.Errorf("get user %d: %w", id, err)
    }
    return &u, nil
}</code></pre>

<p><strong>%w vs %v:</strong></p>
<ul>
<li><code>%w</code> — wraps the error (preserves the chain, can be unwrapped later with <code>errors.Is</code>/<code>errors.As</code>)</li>
<li><code>%v</code> — just includes the text (breaks the chain)</li>
</ul>

<h2>Sentinel Errors</h2>
<pre><code>var ErrNotFound = errors.New("not found")
var ErrConflict = errors.New("already exists")

// Usage
if errors.Is(err, ErrNotFound) {
    // handle 404
}</code></pre>

<h2>Custom Error Types</h2>
<pre><code>type AppError struct {
    Code    int
    Message string
    Err     error
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

// Check with errors.As
var appErr *AppError
if errors.As(err, &appErr) {
    w.WriteHeader(appErr.Code)
}</code></pre>

<h2>Common Anti-Patterns</h2>
<pre><code>// WRONG: Ignoring errors
data, _ := json.Marshal(v)  // What if it fails?

// WRONG: Checking error string
if err.Error() == "not found" {  // Fragile!

// WRONG: Wrapping without context
return fmt.Errorf("%w", err)  // Adds nothing useful

// CORRECT: Meaningful context
return fmt.Errorf("fetch user %d from DB: %w", id, err)</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What is the difference between %w and %v in fmt.Errorf?",
						Options:     []string{"No difference", "%w wraps the error preserving the chain, %v just formats as text", "%v is newer", "%w is for warnings"},
						Correct:     1,
						Explanation: "%w wraps the error, allowing errors.Is() and errors.As() to find it later. %v converts to string text, breaking the error chain.",
					},
					{
						Question:    "How do you check if an error is a specific type?",
						Options:     []string{"err == ErrNotFound", "err.Error() == \"not found\"", "errors.Is(err, ErrNotFound) or errors.As(err, &target)", "switch err.(type)"},
						Correct:     2,
						Explanation: "errors.Is checks if any error in the chain matches a value. errors.As checks if any error in the chain matches a type. These handle wrapped errors correctly.",
					},
					{
						Question:    "Why should you NOT write `data, _ := json.Marshal(v)` in production code?",
						Options:     []string{"It's slower", "It ignores potential errors that should be handled", "It doesn't compile", "Marshal never fails"},
						Correct:     1,
						Explanation: "While json.Marshal rarely fails for simple types, ignoring errors is dangerous. If it does fail (circular reference, unsupported type), you get silent data corruption.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Create AppError system for WatchTogether",
						Description: `<p>Create a unified error handling system:</p>
<ol>
<li>Define <code>AppError</code> struct with Code, Message, Err fields</li>
<li>Add helper constructors: <code>NotFound(msg)</code>, <code>BadRequest(msg)</code>, <code>Internal(err)</code></li>
<li>Create an error handler middleware that converts AppError to HTTP responses</li>
</ol>`,
						Hints: `<p>Use errors.As in the middleware to check if the error is *AppError. If not, treat it as 500 Internal Server Error.</p>`,
						Solution: `<pre><code>package apperror

import (
    "errors"
    "fmt"
    "net/http"
)

type AppError struct {
    Code    int    ` + "`" + `json:"-"` + "`" + `
    Message string ` + "`" + `json:"error"` + "`" + `
    Err     error  ` + "`" + `json:"-"` + "`" + `
}

func (e *AppError) Error() string { return e.Message }
func (e *AppError) Unwrap() error { return e.Err }

func NotFound(msg string) *AppError {
    return &AppError{Code: http.StatusNotFound, Message: msg}
}

func BadRequest(msg string) *AppError {
    return &AppError{Code: http.StatusBadRequest, Message: msg}
}

func Internal(err error) *AppError {
    return &AppError{
        Code:    http.StatusInternalServerError,
        Message: "internal server error",
        Err:     err,
    }
}

func HandleError(w http.ResponseWriter, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        http.Error(w, appErr.Message, appErr.Code)
        return
    }
    http.Error(w, "internal error", 500)
}</code></pre>`,
					},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 2: HTTP & Routing
// ═══════════════════════════════════════════════════════════

func module2_HTTPAndRouting() SeedModule {
	return SeedModule{
		Slug:        "http-routing",
		Title:       "Module 2: HTTP Server & Routing",
		Description: "net/http, chi router, middleware, JSON encoding/decoding, request lifecycle.",
		Order:       2,
		Lessons: []SeedLesson{
			{
				Slug:  "net-http-basics",
				Title: "net/http From Scratch",
				Order: 1,
				Content: `<h1>net/http From Scratch</h1>

<h2>The Standard Library HTTP Server</h2>
<p>Go's net/http is production-grade. Companies like Cloudflare and Uber use it directly.</p>

<pre><code>package main

import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
    })
    http.ListenAndServe(":8080", nil)
}</code></pre>

<h2>Understanding http.Handler</h2>
<pre><code>type Handler interface {
    ServeHTTP(ResponseWriter, *Request)
}</code></pre>
<p>This is the core of Go's HTTP. Everything is a Handler. Middleware? Handler wrapping Handler. Router? Handler that dispatches to other Handlers.</p>

<h2>http.ResponseWriter</h2>
<pre><code>type ResponseWriter interface {
    Header() http.Header       // get response headers
    Write([]byte) (int, error) // write body
    WriteHeader(statusCode int) // set status code
}

// IMPORTANT: WriteHeader must be called BEFORE Write
// IMPORTANT: Header().Set() must be called BEFORE WriteHeader</code></pre>

<h2>http.Request</h2>
<pre><code>r.Method           // "GET", "POST", etc.
r.URL.Path         // "/users/42"
r.URL.Query()      // query parameters
r.Header.Get("X")  // request headers
r.Body             // io.ReadCloser — the request body
r.Context()        // context for cancellation/timeout</code></pre>

<h2>JSON Response Pattern</h2>
<pre><code>func respondJSON(w http.ResponseWriter, status int, data any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    json.NewEncoder(w).Encode(data)
}

func handleGetUser(w http.ResponseWriter, r *http.Request) {
    user := User{ID: 1, Name: "John"}
    respondJSON(w, http.StatusOK, user)
}</code></pre>

<h2>JSON Request Decoding (Safe)</h2>
<pre><code>func decodeJSON(r *http.Request, dst any) error {
    // Limit body size to 1MB
    r.Body = http.MaxBytesReader(nil, r.Body, 1_048_576)

    dec := json.NewDecoder(r.Body)
    dec.DisallowUnknownFields() // reject extra fields

    if err := dec.Decode(dst); err != nil {
        return fmt.Errorf("decode json: %w", err)
    }
    return nil
}</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What order must HTTP response operations happen?",
						Options:     []string{"Write → Header → WriteHeader", "WriteHeader → Write → Header", "Header().Set() → WriteHeader() → Write()", "Any order"},
						Correct:     2,
						Explanation: "Headers must be set before WriteHeader, and WriteHeader must be called before Write. Once you call Write, the status is implicitly 200.",
					},
					{
						Question:    "Why use http.MaxBytesReader when decoding JSON?",
						Options:     []string{"For better performance", "To limit request body size and prevent memory abuse", "It's required by the standard", "For compression"},
						Correct:     1,
						Explanation: "MaxBytesReader limits how much data can be read from the body. Without it, an attacker could send a huge body and exhaust your server's memory.",
					},
					{
						Question:    "What does DisallowUnknownFields() do on json.Decoder?",
						Options:     []string{"Rejects JSON with fields not in the target struct", "Ignores unknown fields", "Panics on unknown fields", "Logs unknown fields"},
						Correct:     0,
						Explanation: "DisallowUnknownFields causes Decode to return an error if the JSON contains fields that don't match any field in the destination struct.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Build WatchTogether HTTP server",
						Description: `<p>Create a basic HTTP server for WatchTogether:</p>
<ol>
<li>Set up chi router with Logger and Recoverer middleware</li>
<li>Create a <code>GET /api/health</code> endpoint returning <code>{"status": "ok"}</code></li>
<li>Create <code>respondJSON</code> and <code>decodeJSON</code> helper functions</li>
<li>Add graceful shutdown with signal handling</li>
<li>Run and test with <code>curl localhost:8080/api/health</code></li>
</ol>`,
						Hints: `<p>Use os/signal with syscall.SIGINT and syscall.SIGTERM. Create a context with timeout for shutdown.</p>`,
						Solution: `<pre><code>// Refer to the GoLearn cmd/server/main.go for the pattern.
// The key parts: chi.NewRouter(), middleware, signal.Notify,
// srv.Shutdown(ctx) in a goroutine.</code></pre>`,
					},
				},
			},
			{
				Slug:  "chi-router",
				Title: "Chi Router & Middleware",
				Order: 2,
				Content: `<h1>Chi Router & Middleware</h1>

<h2>Why Chi?</h2>
<p>Chi is lightweight, idiomatic, and compatible with net/http. No magic, no reflection.</p>

<pre><code>r := chi.NewRouter()

// Global middleware
r.Use(middleware.Logger)
r.Use(middleware.Recoverer)
r.Use(middleware.Timeout(30 * time.Second))

// Routes
r.Get("/api/videos", h.ListVideos)
r.Post("/api/videos", h.CreateVideo)
r.Route("/api/videos/{id}", func(r chi.Router) {
    r.Get("/", h.GetVideo)
    r.Put("/", h.UpdateVideo)
    r.Delete("/", h.DeleteVideo)
})

// URL params
id := chi.URLParam(r, "id")</code></pre>

<h2>Middleware — How It Actually Works</h2>
<pre><code>// Middleware signature
type Middleware func(http.Handler) http.Handler

// Example: request timing
func TimingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)  // call next handler
        duration := time.Since(start)
        slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration", duration)
    })
}</code></pre>

<p>Middleware wraps handlers like layers of an onion. Request goes in through all layers, response comes back out.</p>

<h2>Route Groups</h2>
<pre><code>r.Route("/api", func(r chi.Router) {
    // Public routes
    r.Get("/health", h.Health)

    // Protected routes
    r.Group(func(r chi.Router) {
        r.Use(AuthMiddleware)
        r.Get("/me", h.GetCurrentUser)
        r.Get("/rooms", h.ListRooms)
    })

    // Admin routes
    r.Group(func(r chi.Router) {
        r.Use(AuthMiddleware)
        r.Use(AdminMiddleware)
        r.Delete("/users/{id}", h.DeleteUser)
    })
})</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What is the signature of a middleware function in Go?",
						Options:     []string{"func(w, r)", "func(http.Handler) http.Handler", "func(next) next", "func(*chi.Mux)"},
						Correct:     1,
						Explanation: "A middleware takes an http.Handler (the next handler in the chain) and returns an http.Handler (itself wrapping the next one).",
					},
					{
						Question:    "How do you get URL parameters with chi?",
						Options:     []string{"r.Param(\"id\")", "r.URL.Query().Get(\"id\")", "chi.URLParam(r, \"id\")", "r.PathValue(\"id\")"},
						Correct:     2,
						Explanation: "chi.URLParam(r, \"name\") extracts path parameters defined in the route pattern like /users/{id}.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Create WatchTogether video CRUD endpoints",
						Description: `<p>Build the video management API:</p>
<ol>
<li><code>GET /api/videos</code> — list all videos (return JSON array)</li>
<li><code>GET /api/videos/{id}</code> — get single video</li>
<li><code>POST /api/videos</code> — create video entry</li>
<li><code>DELETE /api/videos/{id}</code> — delete video</li>
<li>Use chi route groups and URL params</li>
<li>For now, use an in-memory slice as storage</li>
</ol>`,
						Hints: `<p>Use sync.Mutex to protect the in-memory slice from race conditions. Parse ID with strconv.Atoi(chi.URLParam(r, "id")).</p>`,
						Solution: `<pre><code>// Pattern: handler struct with methods
type VideoHandler struct {
    mu     sync.Mutex
    videos []model.Video
    nextID int64
}

func (h *VideoHandler) List(w http.ResponseWriter, r *http.Request) {
    h.mu.Lock()
    defer h.mu.Unlock()
    respondJSON(w, 200, h.videos)
}</code></pre>`,
					},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 3: Database & SQL
// ═══════════════════════════════════════════════════════════

func module3_DatabaseAndSQL() SeedModule {
	return SeedModule{
		Slug:        "database-sql",
		Title:       "Module 3: PostgreSQL & pgx",
		Description: "Database design, SQL queries, pgx driver, connection pools, migrations, transactions.",
		Order:       3,
		Lessons: []SeedLesson{
			{
				Slug:  "pgx-basics",
				Title: "pgx Connection & Queries",
				Order: 1,
				Content: `<h1>pgx — PostgreSQL Driver for Go</h1>

<h2>Why pgx?</h2>
<p>pgx is the best PostgreSQL driver for Go. It's faster than database/sql, supports PostgreSQL-specific features, and has a built-in connection pool.</p>

<h2>Connection Pool</h2>
<pre><code>import "github.com/jackc/pgx/v5/pgxpool"

pool, err := pgxpool.New(ctx, "postgres://user:pass@localhost:5432/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer pool.Close()

// Pool manages connections automatically
// Default: max connections = number of CPUs * 4</code></pre>

<h2>Query Patterns</h2>
<pre><code>// Single row
var user User
err := pool.QueryRow(ctx,
    "SELECT id, name, email FROM users WHERE id = $1", userID,
).Scan(&user.ID, &user.Name, &user.Email)

// Multiple rows
rows, err := pool.Query(ctx, "SELECT id, name FROM users WHERE active = $1", true)
if err != nil {
    return nil, err
}
defer rows.Close()

var users []User
for rows.Next() {
    var u User
    if err := rows.Scan(&u.ID, &u.Name); err != nil {
        return nil, err
    }
    users = append(users, u)
}
if err := rows.Err(); err != nil {
    return nil, err  // Don't forget this!
}

// Execute (INSERT/UPDATE/DELETE)
tag, err := pool.Exec(ctx,
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "John", "john@example.com")
// tag.RowsAffected() — how many rows changed</code></pre>

<p><strong>Critical:</strong> Always check <code>rows.Err()</code> after the loop! The loop can end early if there's a network error.</p>

<h2>SQL Injection Prevention</h2>
<pre><code>// NEVER do this — SQL injection!
query := "SELECT * FROM users WHERE name = '" + name + "'"

// ALWAYS use parameterized queries
pool.QueryRow(ctx, "SELECT * FROM users WHERE name = $1", name)</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Why must you call rows.Err() after iterating query results?",
						Options:     []string{"It's not required", "To close the connection", "To catch errors that may have occurred during iteration", "To commit the transaction"},
						Correct:     2,
						Explanation: "rows.Next() can return false either because there are no more rows OR because an error occurred. rows.Err() catches the error case.",
					},
					{
						Question:    "What does $1 in pgx queries represent?",
						Options:     []string{"A string literal", "A parameterized placeholder preventing SQL injection", "A regex pattern", "A column reference"},
						Correct:     1,
						Explanation: "Parameterized placeholders ($1, $2, ...) are safely escaped by the driver, preventing SQL injection. Never concatenate user input into queries.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Connect WatchTogether to PostgreSQL",
						Description: `<p>Set up database for WatchTogether:</p>
<ol>
<li>Add docker-compose.yml with PostgreSQL service</li>
<li>Create migration SQL for videos, rooms, room_members tables</li>
<li>Create <code>repository/video_repo.go</code> using pgxpool</li>
<li>Implement GetAll, GetByID, Create, Delete methods</li>
<li>Connect it in main.go</li>
</ol>`,
						Hints: `<p>Use pgxpool.New(ctx, dbURL). Don't forget defer pool.Close(). Check rows.Err() after iteration loops.</p>`,
						Solution: `<p>Follow the pattern from GoLearn's repository/ package. Key: always use $1 params, always check rows.Err(), always defer rows.Close().</p>`,
					},
				},
			},
			{
				Slug:  "transactions-migrations",
				Title: "Transactions & Migrations",
				Order: 2,
				Content: `<h1>Transactions & Migrations</h1>

<h2>Transactions</h2>
<p>A transaction groups multiple operations into one atomic unit. Either all succeed or all fail.</p>
<pre><code>tx, err := pool.Begin(ctx)
if err != nil {
    return err
}
defer tx.Rollback(ctx) // Rollback if not committed

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance - $1 WHERE id = $2", amount, fromID)
if err != nil {
    return err
}

_, err = tx.Exec(ctx, "UPDATE accounts SET balance = balance + $1 WHERE id = $2", amount, toID)
if err != nil {
    return err
}

return tx.Commit(ctx) // Commit only if everything succeeded</code></pre>

<p><strong>Key pattern:</strong> <code>defer tx.Rollback(ctx)</code> is safe even after Commit — it becomes a no-op.</p>

<h2>Migrations with golang-migrate</h2>
<pre><code># Install
brew install golang-migrate

# Create migration
migrate create -ext sql -dir migrations -seq add_users_table

# Run migrations
migrate -path migrations -database "postgres://..." up

# Rollback last migration
migrate -path migrations -database "postgres://..." down 1</code></pre>

<p>Each migration has two files:</p>
<ul>
<li><code>000001_add_users_table.up.sql</code> — apply changes</li>
<li><code>000001_add_users_table.down.sql</code> — undo changes</li>
</ul>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What does `defer tx.Rollback(ctx)` do after a successful Commit?",
						Options:     []string{"Undoes the commit", "Panics", "Nothing — it's a no-op after commit", "Returns an error"},
						Correct:     2,
						Explanation: "After Commit(), Rollback() is a no-op. This pattern is safe and ensures cleanup if any error prevents reaching Commit().",
					},
					{
						Question:    "What are the two files in a golang-migrate migration?",
						Options:     []string{"create.sql and destroy.sql", "up.sql and down.sql", "apply.sql and revert.sql", "init.sql and cleanup.sql"},
						Correct:     1,
						Explanation: "up.sql applies the migration (create tables, add columns). down.sql reverses it (drop tables, remove columns). This allows rollbacks.",
					},
				},
				Tasks: []SeedTask{},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 4: Architecture
// ═══════════════════════════════════════════════════════════

func module4_Architecture() SeedModule {
	return SeedModule{
		Slug:        "architecture",
		Title:       "Module 4: Clean Architecture",
		Description: "Layered architecture, repository pattern, service layer, separation of concerns.",
		Order:       4,
		Lessons: []SeedLesson{
			{
				Slug:  "layers",
				Title: "Handler → Service → Repository",
				Order: 1,
				Content: `<h1>Layered Architecture</h1>

<h2>The Three Layers</h2>
<pre><code>Handler (HTTP)  →  Service (Business Logic)  →  Repository (Database)
   ↑                      ↑                           ↑
Knows HTTP          Knows nothing about      Knows nothing about
request/response    HTTP or database          HTTP or business rules</code></pre>

<h2>Why Layers?</h2>
<ul>
<li><strong>Testability:</strong> Mock the layer below, test the layer above</li>
<li><strong>Flexibility:</strong> Swap PostgreSQL for MongoDB? Only repository changes</li>
<li><strong>Clarity:</strong> Each layer has one responsibility</li>
</ul>

<h2>Layer Rules</h2>
<ol>
<li><strong>Handler:</strong> Parse HTTP request → call service → format HTTP response. No business logic here.</li>
<li><strong>Service:</strong> Business rules, validation, orchestration. No HTTP, no SQL.</li>
<li><strong>Repository:</strong> Database queries only. Returns domain models, not database rows.</li>
</ol>

<h2>Example Flow: Create a Room</h2>
<pre><code>// Handler — HTTP concerns only
func (h *RoomHandler) Create(w http.ResponseWriter, r *http.Request) {
    var input CreateRoomInput
    if err := decodeJSON(r, &input); err != nil {
        respondError(w, http.StatusBadRequest, "invalid json")
        return
    }

    room, err := h.service.CreateRoom(r.Context(), input)
    if err != nil {
        handleError(w, err)
        return
    }

    respondJSON(w, http.StatusCreated, room)
}

// Service — business logic
func (s *RoomService) CreateRoom(ctx context.Context, input CreateRoomInput) (*Room, error) {
    // Validate
    if input.Name == "" {
        return nil, apperror.BadRequest("room name required")
    }

    // Check video exists
    video, err := s.videoStore.GetByID(ctx, input.VideoID)
    if err != nil {
        return nil, apperror.NotFound("video not found")
    }

    // Create room
    room := &Room{
        Name:    input.Name,
        VideoID: video.ID,
    }
    if err := s.roomStore.Create(ctx, room); err != nil {
        return nil, fmt.Errorf("create room: %w", err)
    }

    return room, nil
}

// Repository — database only
func (r *RoomRepo) Create(ctx context.Context, room *Room) error {
    return r.pool.QueryRow(ctx,
        "INSERT INTO rooms (name, video_id) VALUES ($1, $2) RETURNING id, created_at",
        room.Name, room.VideoID,
    ).Scan(&room.ID, &room.CreatedAt)
}</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Which layer should contain input validation logic?",
						Options:     []string{"Handler", "Service", "Repository", "Middleware"},
						Correct:     1,
						Explanation: "Business validation belongs in the Service layer. The Handler only parses HTTP requests. Basic format validation (JSON parsing) is in the Handler.",
					},
					{
						Question:    "Can the Repository layer import the Handler package?",
						Options:     []string{"Yes, if needed", "No — dependencies flow one direction: Handler → Service → Repository", "Yes, through interfaces", "Only for error types"},
						Correct:     1,
						Explanation: "Dependencies must flow inward. Handler knows about Service, Service knows about Repository. Never the reverse. This prevents circular dependencies.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Implement full CRUD for WatchTogether Rooms",
						Description: `<p>Build the Room management with proper layered architecture:</p>
<ol>
<li>Repository: RoomRepo with Create, GetByID, ListActive, Delete</li>
<li>Service: RoomService with business logic (validate name, check video exists)</li>
<li>Handler: RoomHandler with HTTP endpoints</li>
<li>Wire everything in main.go</li>
</ol>`,
						Hints: `<p>Follow the exact pattern from the lesson. Service defines the interface, repository implements it.</p>`,
						Solution: `<p>The key is the flow: Handler decodes HTTP → Service validates and orchestrates → Repository does SQL. Each layer only knows about the layer directly below it through interfaces.</p>`,
					},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 5: Testing
// ═══════════════════════════════════════════════════════════

func module5_Testing() SeedModule {
	return SeedModule{
		Slug:        "testing",
		Title:       "Module 5: Testing",
		Description: "Unit tests, integration tests, table-driven tests, mocks, testcontainers, benchmarks.",
		Order:       5,
		Lessons: []SeedLesson{
			{
				Slug:  "unit-tests",
				Title: "Unit Tests & Table-Driven Tests",
				Order: 1,
				Content: `<h1>Unit Tests & Table-Driven Tests</h1>

<h2>Test File Convention</h2>
<pre><code>// File: service/room.go      → Test: service/room_test.go
// File: handler/video.go     → Test: handler/video_test.go
// Package: same as the file being tested</code></pre>

<h2>Basic Test</h2>
<pre><code>func TestSizeHuman(t *testing.T) {
    v := Video{Size: 1536 * 1024 * 1024} // 1.5 GB
    got := v.SizeHuman()
    want := "1.5 GB"
    if got != want {
        t.Errorf("SizeHuman() = %q, want %q", got, want)
    }
}</code></pre>

<h2>Table-Driven Tests — The Go Way</h2>
<pre><code>func TestSizeHuman(t *testing.T) {
    tests := []struct {
        name string
        size int64
        want string
    }{
        {"zero", 0, "0.0 KB"},
        {"kilobytes", 512 * 1024, "512.0 KB"},
        {"megabytes", 50 * 1024 * 1024, "50.0 MB"},
        {"gigabytes", 1536 * 1024 * 1024, "1.5 GB"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            v := Video{Size: tt.size}
            if got := v.SizeHuman(); got != tt.want {
                t.Errorf("SizeHuman() = %q, want %q", got, tt.want)
            }
        })
    }
}</code></pre>

<h2>Testing with Mocks</h2>
<pre><code>// Define a mock
type mockVideoStore struct {
    videos []Video
    err    error
}

func (m *mockVideoStore) GetAll(ctx context.Context) ([]Video, error) {
    return m.videos, m.err
}

// Use in test
func TestVideoService_List(t *testing.T) {
    mock := &mockVideoStore{
        videos: []Video{{ID: 1, Title: "Test"}},
    }
    svc := NewVideoService(mock)

    videos, err := svc.ListVideos(context.Background())
    if err != nil {
        t.Fatal(err)
    }
    if len(videos) != 1 {
        t.Errorf("got %d videos, want 1", len(videos))
    }
}</code></pre>

<h2>Running Tests</h2>
<pre><code>go test ./...                    # all tests
go test ./internal/service/...   # specific package
go test -v ./...                 # verbose
go test -run TestSizeHuman ./... # specific test
go test -count=1 ./...           # no caching
go test -cover ./...             # coverage
go test -race ./...              # detect data races</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What is the naming convention for test files in Go?",
						Options:     []string{"test_filename.go", "filename_test.go", "filename.test.go", "tests/filename.go"},
						Correct:     1,
						Explanation: "Test files must end with _test.go. They are automatically excluded from regular builds but included when running go test.",
					},
					{
						Question:    "What does `go test -race ./...` detect?",
						Options:     []string{"Performance issues", "Memory leaks", "Data race conditions (concurrent access bugs)", "Syntax errors"},
						Correct:     2,
						Explanation: "The -race flag enables Go's race detector, which finds concurrent accesses to shared memory that aren't properly synchronized.",
					},
					{
						Question:    "Why use table-driven tests instead of separate test functions?",
						Options:     []string{"They're faster", "Less code duplication, easy to add new cases, clearer test names", "They're required by Go", "Better error messages"},
						Correct:     1,
						Explanation: "Table-driven tests reduce duplication. Adding a new test case is just one more entry in the table. t.Run gives each case a clear sub-test name.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Write tests for WatchTogether VideoService",
						Description: `<p>Write comprehensive tests:</p>
<ol>
<li>Create a mock implementation of VideoStore interface</li>
<li>Write table-driven tests for ListVideos (empty store, multiple videos, error case)</li>
<li>Write tests for GetVideo (found, not found)</li>
<li>Run with <code>go test -v -race ./internal/service/...</code></li>
</ol>`,
						Hints: `<p>Your mock struct should have fields for the data to return and the error to return. This lets each test case configure different behavior.</p>`,
						Solution: `<p>Mock pattern: struct with data/err fields → constructor → implement interface methods returning those fields. Each test configures different scenarios.</p>`,
					},
				},
			},
			{
				Slug:  "integration-tests",
				Title: "Integration Tests with Testcontainers",
				Order: 2,
				Content: `<h1>Integration Tests with Testcontainers</h1>

<h2>Why Integration Tests?</h2>
<p>Unit tests with mocks verify logic. Integration tests verify your code actually works with real PostgreSQL.</p>

<h2>Testcontainers-Go</h2>
<pre><code>import "github.com/testcontainers/testcontainers-go"
import "github.com/testcontainers/testcontainers-go/modules/postgres"

func setupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    ctx := context.Background()

    pgContainer, err := postgres.Run(ctx,
        "postgres:16-alpine",
        postgres.WithDatabase("testdb"),
        postgres.WithUsername("test"),
        postgres.WithPassword("test"),
        testcontainers.WithWaitStrategy(
            wait.ForLog("database system is ready").
                WithOccurrence(2).
                WithStartupTimeout(5*time.Second)),
    )
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { pgContainer.Terminate(ctx) })

    connStr, _ := pgContainer.ConnectionString(ctx, "sslmode=disable")
    pool, err := pgxpool.New(ctx, connStr)
    if err != nil {
        t.Fatal(err)
    }
    t.Cleanup(func() { pool.Close() })

    // Run migrations
    // ...

    return pool
}</code></pre>

<p><strong>How it works:</strong> Testcontainers spins up a real PostgreSQL in Docker for each test. After the test — it's destroyed. Clean slate every time.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What does testcontainers do?",
						Options:     []string{"Mocks the database", "Runs a real database in a Docker container for testing", "Creates in-memory database", "Generates test data"},
						Correct:     1,
						Explanation: "Testcontainers starts real Docker containers (PostgreSQL, Redis, etc.) for integration tests, then cleans them up after.",
					},
				},
				Tasks: []SeedTask{},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 6: Auth & Security
// ═══════════════════════════════════════════════════════════

func module6_AuthAndSecurity() SeedModule {
	return SeedModule{
		Slug:        "auth-security",
		Title:       "Module 6: Authentication & Security",
		Description: "JWT tokens, auth middleware, password hashing, CORS, security headers, OWASP basics.",
		Order:       6,
		Lessons: []SeedLesson{
			{
				Slug:  "jwt-auth",
				Title: "JWT Authentication",
				Order: 1,
				Content: `<h1>JWT Authentication</h1>

<h2>How JWT Works</h2>
<p>JWT (JSON Web Token) is a self-contained token. The server doesn't need to store sessions — the token itself contains the user info.</p>

<pre><code>Header.Payload.Signature

// Header: {"alg": "HS256", "typ": "JWT"}
// Payload: {"user_id": 42, "exp": 1735689600}
// Signature: HMAC-SHA256(header + payload, secret)</code></pre>

<h2>JWT in Go</h2>
<pre><code>import "github.com/golang-jwt/jwt/v5"

type Claims struct {
    UserID int64  ` + "`" + `json:"user_id"` + "`" + `
    jwt.RegisteredClaims
}

// Generate token
func GenerateToken(userID int64, secret string) (string, error) {
    claims := Claims{
        UserID: userID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(secret))
}

// Parse token
func ParseToken(tokenStr, secret string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{},
        func(t *jwt.Token) (interface{}, error) {
            return []byte(secret), nil
        })
    if err != nil {
        return nil, err
    }
    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, errors.New("invalid token")
    }
    return claims, nil
}

// Auth middleware
func AuthMiddleware(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            header := r.Header.Get("Authorization")
            if !strings.HasPrefix(header, "Bearer ") {
                http.Error(w, "unauthorized", 401)
                return
            }

            claims, err := ParseToken(strings.TrimPrefix(header, "Bearer "), secret)
            if err != nil {
                http.Error(w, "invalid token", 401)
                return
            }

            ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}</code></pre>

<h2>Password Hashing</h2>
<pre><code>import "golang.org/x/crypto/bcrypt"

// Hash password
hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

// Verify password
err := bcrypt.CompareHashAndPassword(hash, []byte(password))
if err != nil {
    // Wrong password
}</code></pre>

<p><strong>NEVER store plain text passwords. NEVER use MD5/SHA for passwords. Always bcrypt or argon2.</strong></p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Why is JWT called 'self-contained'?",
						Options:     []string{"It includes its own database", "The token itself contains user info — no server-side session storage needed", "It works without internet", "It encrypts everything"},
						Correct:     1,
						Explanation: "JWT tokens contain the payload (user ID, expiration, etc.) and a signature. The server can verify the token without looking up any stored session.",
					},
					{
						Question:    "Which algorithm should you use for password hashing?",
						Options:     []string{"MD5", "SHA-256", "bcrypt or argon2", "Base64"},
						Correct:     2,
						Explanation: "bcrypt and argon2 are designed for password hashing — they're intentionally slow and include salt. MD5/SHA are fast hash functions, not suitable for passwords.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Add auth to WatchTogether",
						Description: `<p>Implement authentication:</p>
<ol>
<li>Create users table with username, email, password_hash</li>
<li>Implement Register and Login endpoints</li>
<li>Generate JWT on login</li>
<li>Create auth middleware that protects room creation</li>
<li>Store user_id in context for downstream handlers</li>
</ol>`,
						Hints: `<p>Use bcrypt for password hashing. The JWT secret should come from environment variable. Never log tokens or passwords.</p>`,
						Solution: `<p>Follow the JWT pattern from the lesson. Key: middleware extracts token from Authorization header, validates it, puts user_id in context.</p>`,
					},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 7: Concurrency
// ═══════════════════════════════════════════════════════════

func module7_Concurrency() SeedModule {
	return SeedModule{
		Slug:        "concurrency",
		Title:       "Module 7: Goroutines & Concurrency",
		Description: "Goroutines, channels, sync package, context, worker pools, race conditions.",
		Order:       7,
		Lessons: []SeedLesson{
			{
				Slug:  "goroutines-channels",
				Title: "Goroutines & Channels",
				Order: 1,
				Content: `<h1>Goroutines & Channels</h1>

<h2>Goroutines</h2>
<p>A goroutine is a lightweight thread managed by Go runtime. Creating one is trivial:</p>
<pre><code>go func() {
    // runs concurrently
    fmt.Println("hello from goroutine")
}()

// Or with a named function
go processVideo(videoID)</code></pre>

<p><strong>Cost:</strong> ~2KB stack (grows as needed). You can easily run millions of goroutines.</p>

<h2>Channels — Communication Between Goroutines</h2>
<pre><code>// Create a channel
ch := make(chan string)    // unbuffered
ch := make(chan string, 10) // buffered (capacity 10)

// Send
ch <- "hello"

// Receive
msg := <-ch

// Close (only sender should close)
close(ch)</code></pre>

<h2>Common Patterns</h2>

<h3>Fan-out: Multiple workers processing from one channel</h3>
<pre><code>func processVideos(videos []Video) {
    jobs := make(chan Video, len(videos))
    results := make(chan error, len(videos))

    // Start 5 workers
    for i := 0; i < 5; i++ {
        go func() {
            for video := range jobs {
                results <- transcode(video)
            }
        }()
    }

    // Send jobs
    for _, v := range videos {
        jobs <- v
    }
    close(jobs)

    // Collect results
    for range videos {
        if err := <-results; err != nil {
            log.Println("error:", err)
        }
    }
}</code></pre>

<h3>Select — Multiplexing Channels</h3>
<pre><code>select {
case msg := <-ch1:
    fmt.Println("from ch1:", msg)
case msg := <-ch2:
    fmt.Println("from ch2:", msg)
case <-time.After(5 * time.Second):
    fmt.Println("timeout!")
case <-ctx.Done():
    fmt.Println("cancelled!")
}</code></pre>

<h2>sync Package</h2>
<pre><code>// Mutex — protect shared data
var mu sync.Mutex
mu.Lock()
sharedData++
mu.Unlock()

// RWMutex — multiple readers, one writer
var rw sync.RWMutex
rw.RLock()   // multiple goroutines can read
rw.RUnlock()
rw.Lock()    // exclusive write
rw.Unlock()

// WaitGroup — wait for goroutines to finish
var wg sync.WaitGroup
for i := 0; i < 10; i++ {
    wg.Add(1)
    go func() {
        defer wg.Done()
        // do work
    }()
}
wg.Wait() // blocks until all done

// Once — run initialization exactly once
var once sync.Once
once.Do(func() {
    // expensive initialization
})</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What happens when you send to an unbuffered channel with no receiver?",
						Options:     []string{"The value is dropped", "Compile error", "The goroutine blocks until someone receives", "Runtime panic"},
						Correct:     2,
						Explanation: "Unbuffered channels are synchronous. A send blocks until another goroutine receives. This is how goroutines synchronize.",
					},
					{
						Question:    "Who should close a channel?",
						Options:     []string{"The receiver", "The sender", "Both", "The runtime automatically"},
						Correct:     1,
						Explanation: "Only the sender should close a channel. Closing a channel signals to receivers that no more values will be sent. Sending to a closed channel panics.",
					},
					{
						Question:    "When should you use sync.RWMutex instead of sync.Mutex?",
						Options:     []string{"Always — it's faster", "When you have many concurrent readers and few writers", "When you don't need locks", "For channels"},
						Correct:     1,
						Explanation: "RWMutex allows multiple concurrent readers with RLock(). Use it when reads are much more frequent than writes.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Implement video file scanner with goroutines",
						Description: `<p>Build the video scanner for WatchTogether:</p>
<ol>
<li>Create <code>pkg/videoscanner/scanner.go</code></li>
<li>Scan a directory tree for video files (.mp4, .mkv, .avi) using goroutines</li>
<li>Use a worker pool (3 workers) to extract file metadata (size, name) concurrently</li>
<li>Return results through a channel</li>
<li>Use context for cancellation</li>
</ol>`,
						Hints: `<p>Use filepath.WalkDir to traverse. Send file paths to a jobs channel. Workers read from jobs, stat the file, send results to results channel. Main goroutine collects results.</p>`,
						Solution: `<p>Pattern: WalkDir → jobs chan → N workers → results chan → collect. Always use context for cancellation, WaitGroup to know when workers are done, close results channel after wg.Wait().</p>`,
					},
				},
			},
			{
				Slug:  "context-patterns",
				Title: "Context & Cancellation",
				Order: 2,
				Content: `<h1>Context & Cancellation</h1>

<h2>Why Context?</h2>
<p>Context carries deadlines, cancellation signals, and request-scoped values across API boundaries and goroutines.</p>

<pre><code>// In HTTP handler — request cancelled if client disconnects
func (h *Handler) GetVideo(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context() // this context is cancelled when client disconnects

    video, err := h.service.GetVideo(ctx, id)
    // If client disconnected, ctx.Done() fires, and the database query
    // (which also takes ctx) will be cancelled
}</code></pre>

<h2>Creating Contexts</h2>
<pre><code>// With timeout
ctx, cancel := context.WithTimeout(parent, 5*time.Second)
defer cancel() // ALWAYS defer cancel

// With cancellation
ctx, cancel := context.WithCancel(parent)
defer cancel()

// With deadline
ctx, cancel := context.WithDeadline(parent, time.Now().Add(time.Hour))
defer cancel()</code></pre>

<p><strong>Rule:</strong> Always <code>defer cancel()</code>. Forgetting this leaks resources.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What happens if you forget `defer cancel()` after context.WithTimeout?",
						Options:     []string{"Nothing", "The program panics", "Resources are leaked until the parent context is done", "The timeout doesn't work"},
						Correct:     2,
						Explanation: "Forgetting cancel() leaks the internal timer goroutine. It will only be cleaned up when the parent context is cancelled. Always defer cancel().",
					},
				},
				Tasks: []SeedTask{},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 8: DevOps Foundations
// ═══════════════════════════════════════════════════════════

func module8_DevOpsFoundations() SeedModule {
	return SeedModule{
		Slug:        "devops-foundations",
		Title:       "Module 8: DevOps Foundations",
		Description: "Docker, docker-compose, multi-stage builds, networking, volumes, Linux basics.",
		Order:       8,
		Lessons: []SeedLesson{
			{
				Slug:  "docker-deep-dive",
				Title: "Docker Deep Dive",
				Order: 1,
				Content: `<h1>Docker Deep Dive</h1>

<h2>What Docker Actually Does</h2>
<p>Docker packages your app + all dependencies into an isolated container. Same binary runs on your Mac, on CI, on production Linux server.</p>

<h2>Multi-Stage Build (Production Pattern)</h2>
<pre><code>## Stage 1: Build
FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download          # cached if go.mod unchanged

COPY . .
RUN CGO_ENABLED=0 go build -o /app/server ./cmd/server

## Stage 2: Run
FROM alpine:3.19

RUN apk --no-cache add ca-certificates
RUN adduser -D -g '' appuser

COPY --from=builder /app/server /server

USER appuser
EXPOSE 8080

CMD ["/server"]</code></pre>

<h2>Why Multi-Stage?</h2>
<ul>
<li><strong>Stage 1 image:</strong> ~800MB (Go compiler, source code, build tools)</li>
<li><strong>Final image:</strong> ~15MB (just the binary + Alpine)</li>
<li>Smaller image = faster deploys, smaller attack surface</li>
</ul>

<h2>Docker Compose for Development</h2>
<pre><code>services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://user:pass@db:5432/mydb?sslmode=disable
    depends_on:
      db:
        condition: service_healthy

  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_DB: mydb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: pass
    volumes:
      - pgdata:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U user"]
      interval: 5s

volumes:
  pgdata:</code></pre>

<h2>Essential Docker Commands</h2>
<pre><code>docker compose up -d          # start in background
docker compose logs -f app    # follow logs
docker compose down           # stop and remove
docker compose build --no-cache  # rebuild from scratch

docker exec -it container_name sh  # shell into container
docker stats                       # resource usage</code></pre>

<h2>Layer Caching</h2>
<p>Docker caches each layer. Order matters:</p>
<pre><code># GOOD — dependencies cached separately from source code
COPY go.mod go.sum ./    # changes rarely
RUN go mod download      # cached
COPY . .                 # changes often
RUN go build             # rebuilds

# BAD — any change invalidates everything
COPY . .
RUN go mod download && go build</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Why use multi-stage Docker builds?",
						Options:     []string{"Required by Docker", "Final image is much smaller (no build tools)", "It's faster to build", "Better logging"},
						Correct:     1,
						Explanation: "Multi-stage builds let you use a large image for building (with compilers, tools) but copy only the final binary to a tiny runtime image.",
					},
					{
						Question:    "Why copy go.mod before COPY . . in Dockerfile?",
						Options:     []string{"It's required", "Docker layer caching — dependencies change less often than source code", "go.mod must be first", "For security"},
						Correct:     1,
						Explanation: "Docker caches layers. If go.mod didn't change, `go mod download` is cached. Copying all source code first would invalidate this cache on every code change.",
					},
					{
						Question:    "What does `USER appuser` do in a Dockerfile?",
						Options:     []string{"Creates a login", "Runs subsequent commands and the final CMD as a non-root user", "Sets an environment variable", "Nothing important"},
						Correct:     1,
						Explanation: "Running as non-root is a security best practice. If the container is compromised, the attacker has limited permissions.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Dockerize WatchTogether",
						Description: `<p>Create production-ready Docker setup:</p>
<ol>
<li>Write multi-stage Dockerfile for WatchTogether</li>
<li>Create docker-compose.yml with app + PostgreSQL + Redis</li>
<li>Add healthchecks for all services</li>
<li>Mount a local video directory as a volume</li>
<li>Build and run: <code>docker compose up</code></li>
</ol>`,
						Hints: `<p>Use CGO_ENABLED=0 for static binary. Mount videos with - ./videos:/data/videos in docker-compose. Add depends_on with condition: service_healthy.</p>`,
						Solution: `<p>Key: multi-stage (golang:1.22-alpine → alpine:3.19), non-root user, volume for videos, healthcheck for postgres, depends_on for startup order.</p>`,
					},
				},
			},
			{
				Slug:  "linux-networking",
				Title: "Linux & Networking Basics",
				Order: 2,
				Content: `<h1>Linux & Networking for DevOps</h1>

<h2>Essential Linux Commands</h2>
<pre><code># Process management
ps aux | grep myapp        # find process
kill -SIGTERM pid          # graceful stop
kill -SIGKILL pid          # force kill

# Disk & Memory
df -h                      # disk usage
free -h                    # memory usage
du -sh /var/log/           # directory size

# Networking
ss -tlnp                   # listening ports
curl -v http://localhost   # verbose HTTP request
dig example.com            # DNS lookup
ip addr                    # network interfaces

# Logs
journalctl -u myservice -f  # systemd service logs
tail -f /var/log/syslog      # follow log file

# File permissions
chmod 755 script.sh        # rwxr-xr-x
chown user:group file      # change owner</code></pre>

<h2>Systemd — Running Services</h2>
<pre><code># /etc/systemd/system/watchtogether.service
[Unit]
Description=WatchTogether Service
After=network.target postgresql.service

[Service]
Type=simple
User=appuser
WorkingDirectory=/opt/watchtogether
ExecStart=/opt/watchtogether/server
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=DATABASE_URL=postgres://...

[Install]
WantedBy=multi-user.target</code></pre>

<pre><code>systemctl enable watchtogether   # start on boot
systemctl start watchtogether    # start now
systemctl status watchtogether   # check status
systemctl restart watchtogether  # restart</code></pre>

<h2>Nginx as Reverse Proxy</h2>
<pre><code>server {
    listen 80;
    server_name watchtogether.local;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /ws {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What is the difference between SIGTERM and SIGKILL?",
						Options:     []string{"No difference", "SIGTERM asks process to stop gracefully, SIGKILL forces immediate termination", "SIGKILL is slower", "SIGTERM is for terminals only"},
						Correct:     1,
						Explanation: "SIGTERM (15) allows the process to clean up (close connections, flush data). SIGKILL (9) immediately kills it with no chance to clean up. Always try SIGTERM first.",
					},
				},
				Tasks: []SeedTask{},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 9: CI/CD
// ═══════════════════════════════════════════════════════════

func module9_CICD() SeedModule {
	return SeedModule{
		Slug:        "cicd",
		Title:       "Module 9: CI/CD Pipelines",
		Description: "GitHub Actions, linting, testing in CI, automated deploys, secrets management.",
		Order:       9,
		Lessons: []SeedLesson{
			{
				Slug:  "github-actions",
				Title: "GitHub Actions Pipeline",
				Order: 1,
				Content: `<h1>GitHub Actions CI/CD</h1>

<h2>What is CI/CD?</h2>
<ul>
<li><strong>CI (Continuous Integration):</strong> Automatically test & lint every push</li>
<li><strong>CD (Continuous Deployment):</strong> Automatically deploy after CI passes</li>
</ul>

<h2>Basic Pipeline</h2>
<pre><code># .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4

  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:16-alpine
        env:
          POSTGRES_DB: testdb
          POSTGRES_USER: test
          POSTGRES_PASSWORD: test
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go test -race -cover ./...
        env:
          DATABASE_URL: postgres://test:test@localhost:5432/testdb?sslmode=disable

  build:
    needs: [lint, test]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: CGO_ENABLED=0 go build -o server ./cmd/server
      - uses: actions/upload-artifact@v4
        with:
          name: server
          path: server</code></pre>

<h2>Secrets Management</h2>
<pre><code># In workflow
env:
  JWT_SECRET: ` + "${{ secrets.JWT_SECRET }}" + `
  DATABASE_URL: ` + "${{ secrets.DATABASE_URL }}" + `

# Set secrets in: GitHub → Settings → Secrets → Actions</code></pre>

<p><strong>NEVER commit secrets to git. NEVER hardcode passwords in CI files.</strong></p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What does `needs: [lint, test]` mean in a GitHub Actions job?",
						Options:     []string{"The job needs lint and test packages", "The job waits for lint and test jobs to pass before starting", "It runs lint and test first in the same job", "It's optional"},
						Correct:     1,
						Explanation: "needs defines job dependencies. The build job only starts after lint and test jobs succeed. If either fails, build is skipped.",
					},
					{
						Question:    "Where should you store secrets like DATABASE_URL for CI?",
						Options:     []string{"In the workflow YAML file", "In .env file committed to git", "In GitHub Settings → Secrets → Actions", "In Dockerfile"},
						Correct:     2,
						Explanation: "GitHub Actions secrets are encrypted and only exposed to workflow runs. Never commit secrets to the repository.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Create CI pipeline for WatchTogether",
						Description: `<p>Set up GitHub Actions:</p>
<ol>
<li>Create <code>.github/workflows/ci.yml</code></li>
<li>Add lint job with golangci-lint</li>
<li>Add test job with PostgreSQL service container</li>
<li>Add build job that produces a binary</li>
<li>Make build depend on lint + test</li>
</ol>`,
						Hints: `<p>Use the services key for PostgreSQL. Set DATABASE_URL as env var in the test step. Use actions/setup-go@v5.</p>`,
						Solution: `<p>Follow the exact YAML from the lesson. Key: services for postgres, needs for job dependencies, secrets for sensitive values.</p>`,
					},
				},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 10: Monitoring & Observability
// ═══════════════════════════════════════════════════════════

func module10_Monitoring() SeedModule {
	return SeedModule{
		Slug:        "monitoring",
		Title:       "Module 10: Monitoring & Logging",
		Description: "Structured logging, Prometheus metrics, Grafana dashboards, health checks, alerting.",
		Order:       10,
		Lessons: []SeedLesson{
			{
				Slug:  "structured-logging",
				Title: "Structured Logging with slog",
				Order: 1,
				Content: `<h1>Structured Logging with slog</h1>

<h2>Why Structured Logging?</h2>
<p>Text logs are for humans. Structured logs are for machines (log aggregators, searching, alerting).</p>

<pre><code>// BAD — unstructured
log.Printf("user %d created order %d for $%.2f", userID, orderID, amount)
// Output: 2024/01/15 10:30:00 user 42 created order 123 for $99.99
// How do you search for all orders > $50? Good luck parsing that.

// GOOD — structured (slog)
slog.Info("order created",
    "user_id", userID,
    "order_id", orderID,
    "amount", amount,
)
// Output: {"time":"2024-01-15T10:30:00Z","level":"INFO","msg":"order created","user_id":42,"order_id":123,"amount":99.99}
// Now you can query: amount > 50 AND user_id = 42</code></pre>

<h2>slog Setup</h2>
<pre><code>import "log/slog"

// JSON handler for production
logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelInfo,
}))

// Text handler for development
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))

// Set as default
slog.SetDefault(logger)</code></pre>

<h2>Log Levels</h2>
<pre><code>slog.Debug("detailed info for debugging")    // development only
slog.Info("normal operation events")          // request handled, job completed
slog.Warn("something unexpected but handled") // retry succeeded, cache miss
slog.Error("something failed", "error", err)  // DB error, external API down</code></pre>

<h2>What NOT to Log</h2>
<ul>
<li>Passwords, tokens, API keys — NEVER</li>
<li>Full request/response bodies — PII risk</li>
<li>Every successful database query — too noisy</li>
</ul>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Why use structured logging over fmt.Println/log.Printf?",
						Options:     []string{"It's faster", "Structured logs can be searched, filtered, and aggregated by machines", "It's required by Go", "Better colors"},
						Correct:     1,
						Explanation: "Structured logs output key-value pairs (often JSON) that log aggregators like Grafana Loki, ELK, or Datadog can search and visualize.",
					},
					{
						Question:    "Which log level should you use for a failed database query?",
						Options:     []string{"Debug", "Info", "Warn", "Error"},
						Correct:     3,
						Explanation: "Failed database queries are errors — they indicate something went wrong that likely affects the user. Use slog.Error with the error value included.",
					},
				},
				Tasks: []SeedTask{},
			},
			{
				Slug:  "prometheus-metrics",
				Title: "Prometheus Metrics",
				Order: 2,
				Content: `<h1>Prometheus Metrics</h1>

<h2>What is Prometheus?</h2>
<p>Prometheus scrapes /metrics endpoint from your app, stores time-series data, and powers Grafana dashboards and alerts.</p>

<h2>Adding Metrics to Go</h2>
<pre><code>import "github.com/prometheus/client_golang/prometheus"
import "github.com/prometheus/client_golang/prometheus/promhttp"

var (
    httpRequests = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total HTTP requests",
        },
        []string{"method", "path", "status"},
    )

    httpDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Help:    "HTTP request duration",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
)

func init() {
    prometheus.MustRegister(httpRequests, httpDuration)
}

// Middleware to track metrics
func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

        next.ServeHTTP(ww, r)

        duration := time.Since(start).Seconds()
        status := strconv.Itoa(ww.Status())

        httpRequests.WithLabelValues(r.Method, r.URL.Path, status).Inc()
        httpDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
    })
}

// Expose metrics endpoint
r.Handle("/metrics", promhttp.Handler())</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What are the three main Prometheus metric types?",
						Options:     []string{"int, float, string", "Counter, Gauge, Histogram", "Sum, Average, Max", "Request, Response, Error"},
						Correct:     1,
						Explanation: "Counter (only goes up), Gauge (goes up and down), Histogram (distribution of values). There's also Summary, but Histogram is preferred.",
					},
				},
				Tasks: []SeedTask{},
			},
		},
	}
}

// ═══════════════════════════════════════════════════════════
// MODULE 11: Advanced Topics
// ═══════════════════════════════════════════════════════════

func module11_Advanced() SeedModule {
	return SeedModule{
		Slug:        "advanced",
		Title:       "Module 11: Advanced Topics",
		Description: "WebSocket, Redis caching, file streaming, Kubernetes basics, Terraform intro.",
		Order:       11,
		Lessons: []SeedLesson{
			{
				Slug:  "websocket",
				Title: "WebSocket for Real-Time Sync",
				Order: 1,
				Content: `<h1>WebSocket for Real-Time Sync</h1>

<h2>Why WebSocket for WatchTogether?</h2>
<p>HTTP is request-response. WebSocket is bidirectional — both client and server can send messages anytime. Perfect for syncing video playback.</p>

<h2>gorilla/websocket</h2>
<pre><code>import "github.com/gorilla/websocket"

var upgrader = websocket.Upgrader{
    ReadBufferSize:  1024,
    WriteBufferSize: 1024,
    CheckOrigin: func(r *http.Request) bool {
        return true // In production, validate origin
    },
}

func (h *Handler) HandleWS(w http.ResponseWriter, r *http.Request) {
    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        slog.Error("ws upgrade", "error", err)
        return
    }
    defer conn.Close()

    // Read messages
    for {
        _, msg, err := conn.ReadMessage()
        if err != nil {
            break
        }
        // Handle message
        handleMessage(conn, msg)
    }
}</code></pre>

<h2>Room-Based Broadcasting</h2>
<pre><code>type Room struct {
    ID      string
    clients map[*websocket.Conn]bool
    mu      sync.RWMutex
}

func (r *Room) Broadcast(msg []byte) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    for conn := range r.clients {
        conn.WriteMessage(websocket.TextMessage, msg)
    }
}

// Sync events
type SyncEvent struct {
    Type     string  ` + "`" + `json:"type"` + "`" + `     // "play", "pause", "seek"
    Position float64 ` + "`" + `json:"position"` + "`" + ` // seconds
    UserID   int64   ` + "`" + `json:"user_id"` + "`" + `
}</code></pre>

<p>When one user plays/pauses/seeks, the event is broadcast to all room members. Their players sync to the same position.</p>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "Why use WebSocket instead of HTTP polling for real-time sync?",
						Options:     []string{"WebSocket is newer", "WebSocket maintains a persistent connection — instant bidirectional messaging without repeated requests", "HTTP doesn't support JSON", "WebSocket is more secure"},
						Correct:     1,
						Explanation: "HTTP polling creates many short-lived requests. WebSocket opens one connection that stays open, allowing the server to push updates instantly.",
					},
					{
						Question:    "Why use sync.RWMutex for the Room's clients map?",
						Options:     []string{"It's faster than Mutex", "Multiple goroutines may read (broadcast) while only one writes (join/leave)", "It prevents WebSocket errors", "It's required by gorilla/websocket"},
						Correct:     1,
						Explanation: "Broadcasting reads the clients map (many goroutines can do this concurrently with RLock). Adding/removing clients writes to it (needs exclusive Lock).",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Implement WatchTogether sync room",
						Description: `<p>Build the core feature — synchronized playback:</p>
<ol>
<li>Create <code>pkg/syncroom/</code> with Room and Hub types</li>
<li>Implement WebSocket handler at <code>/ws/room/{roomID}</code></li>
<li>Handle sync events: play, pause, seek</li>
<li>Broadcast events to all room members</li>
<li>Handle client disconnect gracefully</li>
</ol>`,
						Hints: `<p>Use a Hub pattern: one goroutine manages all rooms, clients send messages to it via channels. This avoids complex locking. The Hub has register/unregister/broadcast channels.</p>`,
						Solution: `<p>Hub pattern: Hub struct with rooms map, register/unregister channels. Hub.Run() in a goroutine processes events. Each client connection runs readPump and writePump goroutines.</p>`,
					},
				},
			},
			{
				Slug:  "redis-caching",
				Title: "Redis Caching",
				Order: 2,
				Content: `<h1>Redis Caching</h1>

<h2>Why Cache?</h2>
<p>Database queries take 1-50ms. Redis gets take <1ms. Cache frequently accessed data to reduce DB load.</p>

<h2>go-redis</h2>
<pre><code>import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// Set with expiration
err := rdb.Set(ctx, "video:42", videoJSON, 5*time.Minute).Err()

// Get
val, err := rdb.Get(ctx, "video:42").Result()
if err == redis.Nil {
    // Key doesn't exist — fetch from DB
} else if err != nil {
    // Real error
}

// Delete (cache invalidation)
rdb.Del(ctx, "video:42")</code></pre>

<h2>Cache-Aside Pattern</h2>
<pre><code>func (s *VideoService) GetByID(ctx context.Context, id int64) (*Video, error) {
    // 1. Try cache
    key := fmt.Sprintf("video:%d", id)
    cached, err := s.cache.Get(ctx, key).Result()
    if err == nil {
        var v Video
        json.Unmarshal([]byte(cached), &v)
        return &v, nil
    }

    // 2. Cache miss — fetch from DB
    video, err := s.repo.GetByID(ctx, id)
    if err != nil {
        return nil, err
    }

    // 3. Store in cache
    data, _ := json.Marshal(video)
    s.cache.Set(ctx, key, data, 5*time.Minute)

    return video, nil
}</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What does `redis.Nil` error mean?",
						Options:     []string{"Redis is down", "The key doesn't exist (cache miss)", "Invalid key format", "Connection timeout"},
						Correct:     1,
						Explanation: "redis.Nil is returned when GET finds no value for the key. This is a cache miss — you should fetch from the database and populate the cache.",
					},
				},
				Tasks: []SeedTask{},
			},
			{
				Slug:  "kubernetes-intro",
				Title: "Kubernetes Introduction",
				Order: 3,
				Content: `<h1>Kubernetes Introduction</h1>

<h2>What is Kubernetes?</h2>
<p>Kubernetes (K8s) orchestrates containers across multiple servers. It handles: scaling, load balancing, self-healing, rolling updates.</p>

<h2>Core Concepts</h2>
<ul>
<li><strong>Pod:</strong> Smallest deployable unit. Usually 1 container.</li>
<li><strong>Deployment:</strong> Manages pod replicas. Handles rolling updates.</li>
<li><strong>Service:</strong> Stable network endpoint for pods.</li>
<li><strong>Ingress:</strong> HTTP routing from outside the cluster.</li>
<li><strong>ConfigMap/Secret:</strong> Configuration and secrets.</li>
</ul>

<h2>Basic Deployment YAML</h2>
<pre><code>apiVersion: apps/v1
kind: Deployment
metadata:
  name: watchtogether
spec:
  replicas: 3
  selector:
    matchLabels:
      app: watchtogether
  template:
    metadata:
      labels:
        app: watchtogether
    spec:
      containers:
        - name: app
          image: watchtogether:latest
          ports:
            - containerPort: 8080
          env:
            - name: DATABASE_URL
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: database-url
          livenessProbe:
            httpGet:
              path: /api/health
              port: 8080
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            limits:
              memory: "256Mi"
              cpu: "500m"</code></pre>

<h2>Essential kubectl Commands</h2>
<pre><code>kubectl apply -f deployment.yaml   # deploy
kubectl get pods                   # list pods
kubectl logs pod-name              # view logs
kubectl describe pod pod-name      # detailed info
kubectl port-forward svc/app 8080  # local access
kubectl scale deployment app --replicas=5  # scale</code></pre>`,

				QuizQuestions: []SeedQuestion{
					{
						Question:    "What is the role of a Kubernetes Deployment?",
						Options:     []string{"Run a single container", "Manage pod replicas with rolling updates and self-healing", "Store configuration", "Route HTTP traffic"},
						Correct:     1,
						Explanation: "A Deployment manages a set of identical pods. It handles scaling (replicas), rolling updates (zero-downtime), and self-healing (restarts crashed pods).",
					},
					{
						Question:    "What does a liveness probe do?",
						Options:     []string{"Checks if the container can accept traffic", "Checks if the container is still running — restarts if not", "Monitors CPU usage", "Logs health status"},
						Correct:     1,
						Explanation: "Liveness probe periodically checks if the application is alive (e.g., hitting /health). If it fails, Kubernetes restarts the container.",
					},
				},
				Tasks: []SeedTask{
					{
						Title: "Create K8s manifests for WatchTogether",
						Description: `<p>Write Kubernetes deployment files:</p>
<ol>
<li>Deployment with 2 replicas, resource limits, liveness/readiness probes</li>
<li>Service (ClusterIP) exposing port 8080</li>
<li>Secret for DATABASE_URL and JWT_SECRET</li>
<li>ConfigMap for non-sensitive config</li>
</ol>`,
						Hints: `<p>Use separate YAML files or --- separator. For secrets, values must be base64 encoded. Use kubectl create secret for convenience.</p>`,
						Solution: `<p>Follow the YAML from the lesson. Key additions: readinessProbe (traffic routing), resource limits (prevent runaway pods), secrets via secretKeyRef.</p>`,
					},
				},
			},
		},
	}
}
