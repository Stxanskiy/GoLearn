package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/backendraz/golearn/internal/migrate"
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

	// The seeder is often the first thing run against a fresh database, so it
	// applies pending migrations too (same bookkeeping table as the server).
	if ran, err := migrate.Up(ctx, pool, os.Getenv("MIGRATIONS_DIR")); err != nil {
		log.Fatalf("apply migrations: %v", err)
	} else if len(ran) > 0 {
		fmt.Printf("Applied %d migration(s): %v\n", len(ran), ran)
	}

	// Idempotent seed: rebuild quizzes/questions/tasks (no user progress lives
	// there), but UPSERT modules/lessons by slug (stable IDs) below and KEEP the
	// progress table — so user progress survives re-seeds on every deploy.
	pool.Exec(ctx, "DELETE FROM quiz_questions")
	pool.Exec(ctx, "DELETE FROM quizzes")
	pool.Exec(ctx, "DELETE FROM tasks") // cascades submissions (code-attempt history)
	keepModuleSlugs := []string{}

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
		tagsJSON, _ := json.Marshal(mod.Tags)

		keepModuleSlugs = append(keepModuleSlugs, mod.Slug)
		var moduleID int
		err := pool.QueryRow(ctx,
			`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites,
			   category, label, tags, cover_image, accent, est_minutes, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, 'seed')
			 ON CONFLICT (slug) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description,
			   track=EXCLUDED.track, difficulty=EXCLUDED.difficulty,
			   prerequisites=EXCLUDED.prerequisites, category=EXCLUDED.category, label=EXCLUDED.label,
			   tags=EXCLUDED.tags, accent=EXCLUDED.accent, est_minutes=EXCLUDED.est_minutes,
			   cover_image=COALESCE(NULLIF(EXCLUDED.cover_image,''), modules.cover_image)
			 RETURNING id`,
			mod.Slug, mod.Title, mod.Description, mod.Order, track, difficulty, prereqJSON,
			mod.Category, mod.Label, tagsJSON, mod.CoverImage, mod.Accent, mod.EstMinutes).Scan(&moduleID)
		if err != nil {
			log.Fatalf("upsert module %s: %v", mod.Slug, err)
		}
		fmt.Printf("Module %d: %s [%s/%s]\n", mod.Order, mod.Title, track, difficulty)
		keepLessonSlugs := []string{}

		for _, lesson := range mod.Lessons {
			lTrack := lesson.Track
			if lTrack == "" {
				lTrack = track
			}
			lDiff := lesson.Difficulty
			if lDiff == "" {
				lDiff = difficulty
			}
			lKind := lesson.Kind
			if lKind == "" {
				lKind = "theory"
			}

			keepLessonSlugs = append(keepLessonSlugs, lesson.Slug)
			var lessonID int
			err := pool.QueryRow(ctx,
				`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, vm_image, vm_init, source)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'seed')
				 ON CONFLICT (module_id, slug) DO UPDATE SET title=EXCLUDED.title, content=EXCLUDED.content,
				   difficulty=EXCLUDED.difficulty, track=EXCLUDED.track,
				   kind=EXCLUDED.kind, vm_image=EXCLUDED.vm_image, vm_init=EXCLUDED.vm_init
				 RETURNING id`,
				moduleID, lesson.Slug, lesson.Title, lesson.Content, lesson.Order, lDiff, lTrack,
				lKind, lesson.VMImage, lesson.VMInit).Scan(&lessonID)
			if err != nil {
				log.Fatalf("upsert lesson %s: %v", lesson.Slug, err)
			}
			fmt.Printf("  Lesson %d: %s [%s]\n", lesson.Order, lesson.Title, lDiff)

			if len(lesson.Quiz) > 0 {
				var quizID int
				pool.QueryRow(ctx,
					`INSERT INTO quizzes (lesson_id, title) VALUES ($1, $2) RETURNING id`,
					lessonID, "Квиз: "+lesson.Title).Scan(&quizID)
				for qi, q := range lesson.Quiz {
					optJSON, _ := json.Marshal(q.Options)
					oexplJSON, _ := json.Marshal(q.OptionExpl)
					pool.Exec(ctx,
						`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num) VALUES ($1,$2,$3,$4,$5,$6,$7)`,
						quizID, q.Question, optJSON, oexplJSON, q.Correct, q.Explanation, qi+1)
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
		if len(keepLessonSlugs) > 0 {
			pool.Exec(ctx, `DELETE FROM lessons WHERE module_id=$1 AND source='seed' AND slug <> ALL($2)`, moduleID, keepLessonSlugs)
		}
	}
	if len(keepModuleSlugs) > 0 {
		// Only prune seed-managed modules; admin-created courses survive re-seeds.
		pool.Exec(ctx, `DELETE FROM modules WHERE source='seed' AND slug <> ALL($1)`, keepModuleSlugs)
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
	Category                 string   // explicit catalog tag; empty -> derived in handler
	Label                    string   // Старт | Практика | Вызов; empty -> derived
	Tags                     []string // topic chips
	CoverImage               string   // real photo URL/path; empty -> generated SVG
	Accent                   string   // gradient key; empty -> by category
	EstMinutes               int      // 0 -> derived from lesson count
	Lessons                  []L
}
type L struct {
	Slug, Title, Content string
	Order                int
	Difficulty           string // beginner | intermediate | advanced | expert
	Track                string // backend | devops | shared
	Kind                 string // theory | quiz | lab | sim (empty -> theory)
	VMImage              string // lab terminal image
	VMInit               string // lab setup reference/script
	Quiz                 []Q
	Tasks                []T
}
type Q struct {
	Question, Explanation string
	Options               []string
	OptionExpl            []string // per-option explanation (parallel to Options)
	Correct               int
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
	Difficulty                          string // easy | medium | hard
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
	// The Go courses were removed from the platform: GoLearn is a DevOps course
	// now. Their content still lives in cmd/seed/mod*.go — registering that list
	// here again is all it takes to bring them back.

	// ── Section: Кибербезопасность ──
	security := []M{
		mod_security_offense(),
		mod_security_defense(),
	}
	for i := range security {
		security[i].Track = "security"
	}
	// Replace the legacy Go-coding tasks with real shell labs (auto-checked in
	// the sandbox terminal), matching the DevOps courses.
	applySecurityLabs(security)

	// DevOps + Database sections come fully from the devops404 export.
	var mods []M
	mods = append(mods, mod_linux_terminal()) // interactive Linux module with auto-checked shell tasks
	mods = append(mods, importedModules()...)
	mods = append(mods, sqlAcademyModules()...)
	mods = append(mods, security...)

	if err := assignOrder(mods); err != nil {
		log.Fatalf("curriculum order: %v", err)
	}
	return mods
}
