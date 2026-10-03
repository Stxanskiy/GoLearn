package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
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

	// Idempotent seed: modules and lessons are UPSERTed by slug so ids stay stable
	// and user progress survives. Quizzes and tasks are rebuilt, but per lesson and
	// only for lessons this seeder still owns — see the edited_at guard below.
	//
	// It used to wipe every task and quiz in the database before rebuilding, which
	// meant a task an author added in the studio lived until the next deploy.
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
			// published is set on insert only and never in DO UPDATE: a draft course
			// is created hidden, and once the author publishes it by hand the next
			// deploy must not hide it again.
			`INSERT INTO modules (slug, title, description, order_num, track, difficulty, prerequisites,
			   category, label, tags, cover_image, accent, est_minutes, is_trainer, published, source)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 'seed')
			 ON CONFLICT (slug) DO UPDATE SET title=EXCLUDED.title, description=EXCLUDED.description,
			   track=EXCLUDED.track, difficulty=EXCLUDED.difficulty,
			   prerequisites=EXCLUDED.prerequisites, category=EXCLUDED.category, label=EXCLUDED.label,
			   tags=EXCLUDED.tags, accent=EXCLUDED.accent, est_minutes=EXCLUDED.est_minutes,
			   is_trainer=EXCLUDED.is_trainer,
			   cover_image=COALESCE(NULLIF(EXCLUDED.cover_image,''), modules.cover_image)
			 WHERE modules.edited_at IS NULL
			 RETURNING id`,
			mod.Slug, mod.Title, mod.Description, mod.Order, track, difficulty, prereqJSON,
			mod.Category, mod.Label, tagsJSON, mod.CoverImage, mod.Accent, mod.EstMinutes, mod.Trainer,
			!mod.Draft).Scan(&moduleID)
		if errors.Is(err, pgx.ErrNoRows) {
			// The guard above refused the update: an author edited this course in
			// the studio. Their version stands — the point of edited_at is that a
			// deploy does not undo an author's work — so take the id and move on to
			// the lessons.
			if err = pool.QueryRow(ctx, `SELECT id FROM modules WHERE slug=$1`, mod.Slug).Scan(&moduleID); err != nil {
				log.Fatalf("locate edited module %s: %v", mod.Slug, err)
			}
			fmt.Printf("Module %s: edited in the studio, left as it is\n", mod.Slug)
		} else if err != nil {
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
			lFormat := lesson.Format
			if lFormat == "" {
				lFormat = "html"
			}
			if lKind == "" {
				lKind = "theory"
			}

			keepLessonSlugs = append(keepLessonSlugs, lesson.Slug)
			var lessonID int
			err := pool.QueryRow(ctx,
				`INSERT INTO lessons (module_id, slug, title, content, order_num, difficulty, track, kind, vm_image, vm_init, format, source)
				 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 'seed')
				 ON CONFLICT (module_id, slug) DO UPDATE SET title=EXCLUDED.title, content=EXCLUDED.content,
				   difficulty=EXCLUDED.difficulty, track=EXCLUDED.track,
				   kind=EXCLUDED.kind, vm_image=EXCLUDED.vm_image, vm_init=EXCLUDED.vm_init,
				   format=EXCLUDED.format,
				   -- order_num used to be insert-only, so moving a chapter in the content
				   -- file changed nothing: the lesson kept whatever position it got the
				   -- first time it was seeded. Author-edited lessons are excluded by the
				   -- guard below, so this only re-sequences lessons the seeder owns.
				   order_num=EXCLUDED.order_num
				 WHERE lessons.edited_at IS NULL
				 RETURNING id`,
				moduleID, lesson.Slug, lesson.Title, lesson.Content, lesson.Order, lDiff, lTrack,
				lKind, lesson.VMImage, lesson.VMInit, lFormat).Scan(&lessonID)
			if errors.Is(err, pgx.ErrNoRows) {
				// The guard above refused the update: an author owns this lesson now.
				// Leave it and everything under it exactly as they left it. It still
				// counts as kept, or the prune below would delete it instead.
				fmt.Printf("  Lesson %d: %s — правка автора, пропускаю\n", lesson.Order, lesson.Slug)
				continue
			}
			if err != nil {
				log.Fatalf("upsert lesson %s: %v", lesson.Slug, err)
			}
			fmt.Printf("  Lesson %d: %s [%s]\n", lesson.Order, lesson.Title, lDiff)

			// Refresh this lesson's quiz and tasks IN PLACE.
			//
			// This used to delete them and insert them again. submissions and
			// quiz_answers reference them with ON DELETE CASCADE, so every release
			// erased the lab and quiz progress of every student on the platform —
			// they came back to a lesson they had finished and found it blank.
			// Rows are matched by their key in the content instead, and only keys
			// that have genuinely disappeared are removed.
			if len(lesson.Quiz) > 0 {
				var quizID int
				if err := pool.QueryRow(ctx,
					`INSERT INTO quizzes (lesson_id, title) VALUES ($1, $2)
					 ON CONFLICT (lesson_id) DO UPDATE SET title=EXCLUDED.title
					 RETURNING id`,
					lessonID, "Квиз: "+lesson.Title).Scan(&quizID); err != nil {
					log.Fatalf("quiz of lesson %q: %v", lesson.Slug, err)
				}
				keepQuestionKeys := make([]string, 0, len(lesson.Quiz))
				for qi, q := range lesson.Quiz {
					optJSON, _ := json.Marshal(q.Options)
					oexplJSON, _ := json.Marshal(q.OptionExpl)
					// Questions carry no id in the source, so position is the key. A
					// reordered quiz keeps answers attached to the slot, which is the
					// best available and far better than losing them all.
					key := fmt.Sprintf("q%d", qi+1)
					keepQuestionKeys = append(keepQuestionKeys, key)
					if _, err := pool.Exec(ctx,
						`INSERT INTO quiz_questions (quiz_id, question, options, option_explanations, correct_index, explanation, order_num, source_key)
						 VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
						 ON CONFLICT (quiz_id, source_key) WHERE source_key <> '' DO UPDATE SET
						   question=EXCLUDED.question, options=EXCLUDED.options,
						   option_explanations=EXCLUDED.option_explanations,
						   correct_index=EXCLUDED.correct_index, explanation=EXCLUDED.explanation,
						   order_num=EXCLUDED.order_num`,
						quizID, q.Question, optJSON, oexplJSON, q.Correct, q.Explanation, qi+1, key); err != nil {
						log.Fatalf("quiz question of lesson %q: %v", lesson.Slug, err)
					}
				}
				if _, err := pool.Exec(ctx,
					`DELETE FROM quiz_questions
					 WHERE quiz_id=$1 AND source_key <> '' AND source_key <> ALL($2)`,
					quizID, keepQuestionKeys); err != nil {
					log.Fatalf("prune quiz questions of lesson %q: %v", lesson.Slug, err)
				}
				fmt.Printf("    Quiz: %d questions\n", len(lesson.Quiz))
			} else {
				// The content has no quiz for this lesson. Drop the questions the
				// seeder put there, but keep any an author added, and only remove the
				// quiz itself once nothing is left in it: deleting the row cascaded
				// away every question and every student answer under it.
				if _, err := pool.Exec(ctx,
					`DELETE FROM quiz_questions
					 WHERE source_key <> '' AND quiz_id IN (SELECT id FROM quizzes WHERE lesson_id=$1)`,
					lessonID); err != nil {
					log.Fatalf("prune quiz questions of lesson %q: %v", lesson.Slug, err)
				}
				if _, err := pool.Exec(ctx,
					`DELETE FROM quizzes WHERE lesson_id=$1
					 AND NOT EXISTS (SELECT 1 FROM quiz_questions q WHERE q.quiz_id = quizzes.id)`,
					lessonID); err != nil {
					log.Fatalf("drop quiz of lesson %q: %v", lesson.Slug, err)
				}
			}
			keepTaskKeys := make([]string, 0, len(lesson.Tasks))
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
				// A dropped task is invisible otherwise: the lesson just ships with fewer
				// steps than it was authored with. Fail loudly instead — one bad byte
				// in a check message silently removed nine tasks before this.
				// 44 of 846 tasks carry no key in the source; a positional one keeps
				// them stable as long as the lesson's task order does not change.
				key := t.SourceKey
				if key == "" {
					key = fmt.Sprintf("pos%d", ti+1)
				}
				keepTaskKeys = append(keepTaskKeys, key)
				if _, err := pool.Exec(ctx,
					`INSERT INTO tasks (lesson_id, title, description, hints, solution, order_num, difficulty, glossary, test_cases, starter_code, kind, sandbox_image, setup_script, check_script, source_key)
					 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
					 ON CONFLICT (lesson_id, source_key) WHERE source_key <> '' DO UPDATE SET
					   title=EXCLUDED.title, description=EXCLUDED.description, hints=EXCLUDED.hints,
					   solution=EXCLUDED.solution, order_num=EXCLUDED.order_num, difficulty=EXCLUDED.difficulty,
					   glossary=EXCLUDED.glossary, test_cases=EXCLUDED.test_cases,
					   starter_code=EXCLUDED.starter_code, kind=EXCLUDED.kind,
					   sandbox_image=EXCLUDED.sandbox_image, setup_script=EXCLUDED.setup_script,
					   check_script=EXCLUDED.check_script`,
					lessonID, t.Title, t.Description, t.Hints, t.Solution, ti+1, tDiff, glossaryJSON, testCasesJSON, t.StarterCode, kind, t.SandboxImage, t.SetupScript, t.CheckScript, key); err != nil {
					log.Fatalf("task %q of lesson %q: %v", t.Title, lesson.Slug, err)
				}
			}
			// Only keys that really left the content are removed; everything else keeps
			// its id, and with it every submission pointing at it.
			//
			// source_key <> '' is what keeps the studio out of this. A task an author
			// adds through the studio has no key in the content, so without that
			// condition it matched "not named by the content", was deleted on the
			// next deploy, and took every submission against it along the cascade.
			if _, err := pool.Exec(ctx,
				`DELETE FROM tasks
				 WHERE lesson_id=$1 AND source_key <> '' AND source_key <> ALL($2)`,
				lessonID, keepTaskKeys); err != nil {
				log.Fatalf("prune tasks of lesson %q: %v", lesson.Slug, err)
			}
			if len(lesson.Tasks) > 0 {
				fmt.Printf("    Tasks: %d\n", len(lesson.Tasks))
			}
		}
		if len(keepLessonSlugs) > 0 {
			if _, err := pool.Exec(ctx, `DELETE FROM lessons WHERE module_id=$1 AND source='seed' AND edited_at IS NULL AND slug <> ALL($2)`, moduleID, keepLessonSlugs); err != nil {
				log.Fatalf("prune lessons: %v", err)
			}
		}
	}
	if len(keepModuleSlugs) > 0 {
		// Only prune seed-managed modules; admin-created courses survive re-seeds.
		if _, err := pool.Exec(ctx, `DELETE FROM modules WHERE source='seed' AND edited_at IS NULL AND slug <> ALL($1)`, keepModuleSlugs); err != nil {
			log.Fatalf("prune modules: %v", err)
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
	Category                 string   // explicit catalog tag; empty -> derived in handler
	Label                    string   // Старт | Практика | Вызов; empty -> derived
	Tags                     []string // topic chips
	CoverImage               string   // real photo URL/path; empty -> generated SVG
	Accent                   string   // gradient key; empty -> by category
	EstMinutes               int      // 0 -> derived from lesson count
	Trainer                  bool     // practice-only course, listed under trainers
	Draft                    bool     // create unpublished; the author publishes it himself
	Lessons                  []L
}
type L struct {
	Slug, Title, Content string
	Order                int
	Difficulty           string // beginner | intermediate | advanced | expert
	Track                string // backend | devops | shared
	Kind                 string // theory | quiz | lab | sim (empty -> theory)
	Format               string // html | md (empty -> html)
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
	// SourceKey is the task's identity in the content ("lnav_lab1_t2_pwd"). The
	// seeder updates a task in place by it instead of deleting and re-inserting,
	// which used to take every student submission with it through ON DELETE
	// CASCADE — a release wiped the lab progress of everyone on the platform.
	SourceKey    string
	Difficulty   string // easy | medium | hard
	Glossary     []GlossaryItem
	TestCases    []TestCase
	StarterCode  string
	Kind         string // "" -> go | shell
	SandboxImage string
	SetupScript  string
	CheckScript  string
	// SelfCheck marks a task the source export typed "self": the student verifies
	// it themselves. It never gets an auto-check, and — crucially — it is skipped
	// when Checks/Descs are numbered, so adding these tasks does not shift the
	// index every authored check is keyed by.
	SelfCheck bool
}

// ── Registry ──

func getAllModules() []M {
	// The Go courses were removed from the platform: GoLearn is a DevOps course
	// now. Their content still lives in cmd/seed/mod*.go — registering that list
	// here again is all it takes to bring them back.

	// Every course now comes from data, not from Go: the export under
	// cmd/seed/content and the SQL academy files. The three courses that used to
	// be written out in Go — the terminal practicum and the two security ones —
	// were removed rather than migrated.
	var mods []M
	mods = append(mods, importedModules()...)

	if err := assignOrder(mods); err != nil {
		log.Fatalf("curriculum order: %v", err)
	}
	return mods
}
