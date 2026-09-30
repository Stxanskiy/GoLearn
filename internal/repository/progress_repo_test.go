package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Progress is the record of what a student actually did. A wrong UPSERT here
// does not fail — it quietly rewrites someone's history, which is why these run
// against a real database like the billing ones do.

// testLesson creates a throwaway module and lesson and returns the lesson id.
func testLesson(t *testing.T, pool *pgxpool.Pool) int {
	t.Helper()
	ctx := context.Background()
	suffix := time.Now().Format("150405.000000000")
	var moduleID int
	err := pool.QueryRow(ctx,
		`INSERT INTO modules (slug, title, description, order_num, source)
		 VALUES ($1, 'progress test', '', 9000, 'test') RETURNING id`,
		"progress-test-"+suffix).Scan(&moduleID)
	if err != nil {
		t.Fatalf("create module: %v", err)
	}
	var lessonID int
	err = pool.QueryRow(ctx,
		`INSERT INTO lessons (module_id, slug, title, content, order_num, source)
		 VALUES ($1, $2, 'progress test', '', 1, 'test') RETURNING id`,
		moduleID, "progress-test-"+suffix).Scan(&lessonID)
	if err != nil {
		t.Fatalf("create lesson: %v", err)
	}
	t.Cleanup(func() {
		pool.Exec(ctx, `DELETE FROM modules WHERE id=$1`, moduleID)
	})
	return lessonID
}

func progressRow(t *testing.T, pool *pgxpool.Pool, userID, lessonID int) (status string, score, total *int, completed *time.Time) {
	t.Helper()
	err := pool.QueryRow(context.Background(),
		`SELECT status, quiz_score, quiz_total, completed_at FROM progress WHERE user_id=$1 AND lesson_id=$2`,
		userID, lessonID).Scan(&status, &score, &total, &completed)
	if err != nil {
		t.Fatalf("read progress: %v", err)
	}
	return
}

// A finished lesson must stay finished. Saving a quiz result, notes, or calling
// Start again all write to the same row, and any of them setting status would
// demote a lesson the student already completed.
func TestProgressIsNeverDemoted(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewProgressRepo(pool)
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	if err := repo.Upsert(ctx, userID, lessonID, "completed"); err != nil {
		t.Fatalf("complete: %v", err)
	}

	for _, step := range []struct {
		name string
		run  func() error
	}{
		{"quiz result", func() error { return repo.SaveQuizResult(ctx, userID, lessonID, 4, 5) }},
		{"notes", func() error { return repo.SaveNotes(ctx, userID, lessonID, "мои заметки") }},
		{"start again", func() error { return repo.Start(ctx, userID, lessonID) }},
	} {
		if err := step.run(); err != nil {
			t.Fatalf("%s: %v", step.name, err)
		}
		if status, _, _, _ := progressRow(t, pool, userID, lessonID); status != "completed" {
			t.Fatalf("%s demoted a completed lesson to %q", step.name, status)
		}
	}

	// The quiz result still has to land, or the guard above would be satisfied by
	// a no-op write.
	_, score, total, _ := progressRow(t, pool, userID, lessonID)
	if score == nil || total == nil || *score != 4 || *total != 5 {
		t.Errorf("quiz result not stored: score=%v total=%v", score, total)
	}
}

func TestProgressCompletionTimeIsSetOnCompletion(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewProgressRepo(pool)
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	if err := repo.Start(ctx, userID, lessonID); err != nil {
		t.Fatalf("start: %v", err)
	}
	if _, _, _, completed := progressRow(t, pool, userID, lessonID); completed != nil {
		t.Error("a lesson that was only started already has a completion time")
	}
	if err := repo.Upsert(ctx, userID, lessonID, "completed"); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if _, _, _, completed := progressRow(t, pool, userID, lessonID); completed == nil {
		t.Error("completing the lesson did not record when")
	}
}

func TestProgressResetClearsTheRow(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewProgressRepo(pool)
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	if err := repo.SaveQuizResult(ctx, userID, lessonID, 3, 5); err != nil {
		t.Fatalf("save quiz: %v", err)
	}
	if err := repo.ResetLesson(ctx, userID, lessonID); err != nil {
		t.Fatalf("reset: %v", err)
	}
	var n int
	if err := pool.QueryRow(ctx,
		`SELECT count(*) FROM progress WHERE user_id=$1 AND lesson_id=$2 AND quiz_score IS NOT NULL`,
		userID, lessonID).Scan(&n); err != nil {
		t.Fatalf("count: %v", err)
	}
	if n != 0 {
		t.Error("reset left the old quiz score behind")
	}
}

func TestQuizAttemptsAccumulate(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewQuizAttemptRepo(pool)
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	if n, err := repo.Count(ctx, userID, lessonID); err != nil || n != 0 {
		t.Fatalf("fresh lesson reports %d attempts (err %v)", n, err)
	}
	for i := 1; i <= 3; i++ {
		if _, err := repo.Save(ctx, userID, lessonID, i, 5, nil); err != nil {
			t.Fatalf("attempt %d: %v", i, err)
		}
		n, err := repo.Count(ctx, userID, lessonID)
		if err != nil {
			t.Fatalf("count after %d: %v", i, err)
		}
		// Retries are history, not a single row that gets overwritten: the attempt
		// number is shown to the student and used to spot a lesson being ground
		// through by trial and error.
		if n != i {
			t.Fatalf("after %d attempts the repo reports %d", i, n)
		}
	}
}

// A lab is marked finished when every task in it has passed. The count of tasks
// came from COUNT(*) over tasks LEFT JOIN submissions, which counts rows, not
// tasks: each re-press of "Проверить" added one. Checking a single task twice —
// the normal way anyone uses a lab — pushed the total past the number of passed
// tasks and the lesson never showed as done, no matter how much the student did.
func TestLabCountsAsPassedDespiteRetries(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := NewSubmissionRepo(pool)
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	var taskIDs []int
	for i := 1; i <= 2; i++ {
		var id int
		err := pool.QueryRow(ctx,
			`INSERT INTO tasks (lesson_id, title, order_num, kind, check_script)
			 VALUES ($1, $2, $3, 'shell', 'true') RETURNING id`,
			lessonID, "task", i).Scan(&id)
		if err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
		taskIDs = append(taskIDs, id)
	}

	// The first task is checked three times — twice failing, then passing, which
	// is what actually happens in a lab.
	for _, passed := range []bool{false, false, true} {
		if err := repo.Save(ctx, userID, taskIDs[0], "", "", "", passed); err != nil {
			t.Fatalf("save attempt: %v", err)
		}
	}
	if err := repo.Save(ctx, userID, taskIDs[1], "", "", "", true); err != nil {
		t.Fatalf("save second task: %v", err)
	}

	status, err := repo.LessonLabStatus(ctx, userID)
	if err != nil {
		t.Fatalf("lab status: %v", err)
	}
	if !status[lessonID] {
		t.Error("every task passed, but the lab does not count as finished")
	}

	// And it must still be honest: one task left undone means not finished.
	other := testUser(t, pool)
	if err := repo.Save(ctx, other, taskIDs[0], "", "", "", true); err != nil {
		t.Fatalf("save for second user: %v", err)
	}
	status, err = repo.LessonLabStatus(ctx, other)
	if err != nil {
		t.Fatalf("lab status: %v", err)
	}
	if status[lessonID] {
		t.Error("a lab with one of two tasks passed counts as finished")
	}
}
