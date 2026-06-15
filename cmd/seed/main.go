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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	var pool *pgxpool.Pool
	for attempts := 0; attempts < 10; attempts++ {
		var err error
		pool, err = pgxpool.New(ctx, dbURL)
		if err == nil {
			if pingErr := pool.Ping(ctx); pingErr == nil {
				break
			}
			pool.Close()
		}
		if attempts < 9 {
			log.Printf("waiting for database... (attempt %d/10)", attempts+1)
			time.Sleep(time.Second)
		} else {
			log.Fatal("could not connect to database after 10 attempts")
		}
	}
	defer pool.Close()

	pool.Exec(ctx, "DELETE FROM submissions")
	pool.Exec(ctx, "DELETE FROM progress")
	pool.Exec(ctx, "DELETE FROM quiz_questions")
	pool.Exec(ctx, "DELETE FROM quizzes")
	pool.Exec(ctx, "DELETE FROM tasks")
	pool.Exec(ctx, "DELETE FROM lessons")
	pool.Exec(ctx, "DELETE FROM modules")
	pool.Exec(ctx, "ALTER SEQUENCE modules_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE lessons_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE quizzes_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE quiz_questions_id_seq RESTART WITH 1")
	pool.Exec(ctx, "ALTER SEQUENCE tasks_id_seq RESTART WITH 1")

	modules := getAllModules()
	for _, mod := range modules {
		track := mod.Track
		if track == "" {
			track = "backend"
		}
		difficulty := mod.Difficulty
		if difficulty == "" {
			difficulty = "beginner"
		}
		prereqJSON, _ := json.Marshal(mod.Prerequisites)

		var moduleID int
		err := pool.QueryRow(ctx,
			`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
			mod.Slug, mod.Title, mod.Description, mod.Order, track, difficulty, prereqJSON).Scan(&moduleID)
		if err != nil {
			log.Fatalf("insert module %s: %v", mod.Slug, err)
		}
		fmt.Printf("Module %d: %s [%s/%s]\n", mod.Order, mod.Title, track, difficulty)

		for _, lesson := range mod.Lessons {
			lTrack := lesson.Track
			if lTrack == "" {
				lTrack = track
			}
			lDiff := lesson.Difficulty
			if lDiff == "" {
				lDiff = difficulty
			}

			var lessonID int
			err := pool.QueryRow(ctx,
				`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`,
				moduleID, lesson.Slug, lesson.Title, lesson.Content, lesson.Order, lDiff, lTrack).Scan(&lessonID)
			if err != nil {
				log.Fatalf("insert lesson %s: %v", lesson.Slug, err)
			}
			fmt.Printf("  Lesson %d: %s [%s]\n", lesson.Order, lesson.Title, lDiff)

			if len(lesson.Quiz) > 0 {
				var quizID int
				pool.QueryRow(ctx,
					`INSERT INTO quizzes (lesson_id, title) VALUES ($1, $2) RETURNING id`,
					lessonID, "Квиз: "+lesson.Title).Scan(&quizID)
				for qi, q := range lesson.Quiz {
					optJSON, _ := json.Marshal(q.Options)
					pool.Exec(ctx,
						`INSERT INTO quiz_questions (quiz_id, question, options, correct_index, explanation, order_num) VALUES ($1,$2,$3,$4,$5,$6)`,
						quizID, q.Question, optJSON, q.Correct, q.Explanation, qi+1)
				}
				fmt.Printf("    Quiz: %d questions\n", len(lesson.Quiz))
			}
			for ti, t := range lesson.Tasks {
				tDiff := t.Difficulty
				if tDiff == "" {
					tDiff = "easy"
				}
				glossaryJSON, _ := json.Marshal(t.Glossary)
				testCasesJSON, _ := json.Marshal(t.TestCases)
				kind := t.Kind
				if kind == "" {
					kind = "go"
				}
				pool.Exec(ctx,
					`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, kind, sandbox_image, setup_script, check_script) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)`,
					lessonID, t.Title, t.Description, t.Hints, t.Solution, ti+1, tDiff, glossaryJSON, testCasesJSON, t.StarterCode, kind, t.SandboxImage, t.SetupScript, t.CheckScript)
			}
			if len(lesson.Tasks) > 0 {
				fmt.Printf("    Tasks: %d\n", len(lesson.Tasks))
			}
		}
	}
	fmt.Println("\nSeed completed!")
}

// ── Types ──

type M struct {
	Slug, Title, Description string
	Order                    int
	Track                    string   // backend | devops | shared
	Difficulty               string   // beginner | intermediate | advanced | expert
	Prerequisites            []string // module slugs
	Lessons                  []L
}
type L struct {
	Slug, Title, Content string
	Order                int
	Difficulty           string // beginner | intermediate | advanced | expert
	Track                string // backend | devops | shared
	Quiz                 []Q
	Tasks                []T
}
type Q struct {
	Question, Explanation string
	Options              []string
	Correct              int
}
type GlossaryItem struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}
type TestCase struct {
	Input          string `json:"input"`
	ExpectedOutput string `json:"expected_output"`
}
type T struct {
	Title, Description, Hints, Solution string
	Difficulty                          string         // easy | medium | hard
	Glossary                            []GlossaryItem
	TestCases                           []TestCase
	StarterCode                         string
	Kind                                string // "" -> go | shell
	SandboxImage                        string
	SetupScript                         string
	CheckScript                         string
}

// ── Registry ──

func getAllModules() []M {
	mods := []M{
		// ── Track: shared — фундамент ──
		mod01_basics_new(),       // 1: Первые шаги в Go
		mod02_collections_new(),  // 2: Коллекции данных
		mod03_functions_new(),    // 3: Функции
		mod_pointers(),           // 4: Указатели и память
		mod04_structs_new(),      // 5: Структуры и методы
		mod05_interfaces_new(),   // 6: Интерфейсы
		mod06_errors_new(),       // 7: Обработка ошибок
		mod_milestone_beginner(), // 8: ★ Проект CLI (milestone — после всех основ)
		mod_generics(),           // 9: Generics
		mod08_packages(),         // 10: Пакеты и модули
		mod_context(),            // 11: context.Context
		mod_git(),                // 12: Git

		// ── Track: backend ──
		mod07_files_json(),       // 12: Файлы и JSON
		mod09_http(),             // 13: HTTP Сервер
		mod10_database(),         // 14: PostgreSQL
		mod_architecture_full(),  // 15: Архитектура и SOLID (5 уроков)
		mod_testing_full(),       // 16: Тестирование (5 уроков)
		mod_milestone_intermediate(), // 17: ★ Проект REST API (milestone — после архитектуры+тестов)
		mod13_auth(),             // 18: Аутентификация
		mod14_concurrency_full(), // 19: Конкурентность (3 урока)

		// ── Track: backend — advanced ──
		mod_go_internals(),       // Go Internals (memory, scheduler, GC, interfaces)

		// ── Track: devops ──
		mod_linux_terminal(),     // Linux: Старт в терминале (интерактив, песочница)
		mod_linux(),              // Linux: Основные инструменты
		mod_docker_full(),        // Docker
		mod_helm(),               // Kubernetes & Helm
		mod_nginx(),              // Nginx
		mod_ansible(),            // Ansible
		mod_grafana(),            // Grafana + Prometheus (мониторинг)
		mod18_advanced(),         // Advanced: WebSocket, Redis, K8s

		// ── Track: security ──
		mod_security_offense(),   // Кибербез: Пентест (8 уроков)
		mod_security_defense(),   // Кибербез: Blue Team (7 уроков)
	}

	// Fix order numbers
	for i := range mods {
		mods[i].Order = i + 1
	}
	return mods
}
