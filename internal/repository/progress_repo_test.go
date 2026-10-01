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

// Every release ran the seeder, the seeder deleted a lesson's tasks before
// inserting them again, and submissions reference tasks with ON DELETE CASCADE.
// So each deploy erased the lab progress of everyone on the platform: a student
// returned to a lesson they had finished and found it blank. Tasks carry a
// source_key now and are updated in place; this is the property that must hold.
func TestSubmissionsSurviveATaskRefresh(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	upsert := func(title string) int {
		t.Helper()
		var id int
		err := pool.QueryRow(ctx,
			`INSERT INTO tasks (lesson_id, title, order_num, kind, check_script, source_key)
			 VALUES ($1, $2, 1, 'shell', 'true', 'lab_t1')
			 ON CONFLICT (lesson_id, source_key) WHERE source_key <> '' DO UPDATE
			   SET title = EXCLUDED.title
			 RETURNING id`, lessonID, title).Scan(&id)
		if err != nil {
			t.Fatalf("upsert task: %v", err)
		}
		return id
	}

	first := upsert("Шаг 1")
	if err := NewSubmissionRepo(pool).Save(ctx, userID, first, "", "", "", true); err != nil {
		t.Fatalf("save submission: %v", err)
	}

	// What a reseed does: the same task, with edited text.
	again := upsert("Шаг 1, формулировка поправлена")
	if again != first {
		t.Fatalf("task was recreated: id %d became %d", first, again)
	}

	passed, err := NewSubmissionRepo(pool).PassedTaskIDs(ctx, userID, lessonID)
	if err != nil {
		t.Fatalf("passed tasks: %v", err)
	}
	if !passed[first] {
		t.Error("the student's answer did not survive a reseed")
	}
	var title string
	if err := pool.QueryRow(ctx, `SELECT title FROM tasks WHERE id=$1`, first).Scan(&title); err != nil {
		t.Fatalf("read back: %v", err)
	}
	if title != "Шаг 1, формулировка поправлена" {
		t.Errorf("the edit did not land: title is %q", title)
	}

	// And a task an author added through the studio has no key in the content, so
	// the prune has to leave it alone. Without that the fix above only covered
	// the seeder's own tasks: a studio task matched "not named by the content",
	// was deleted on the next deploy, and took its submissions with it.
	var studio int
	if err := pool.QueryRow(ctx,
		`INSERT INTO tasks (lesson_id, title, order_num, kind, check_script)
		 VALUES ($1, 'Добавлено в студии', 2, 'shell', 'true') RETURNING id`,
		lessonID).Scan(&studio); err != nil {
		t.Fatalf("create studio task: %v", err)
	}
	if err := NewSubmissionRepo(pool).Save(ctx, userID, studio, "", "", "", true); err != nil {
		t.Fatalf("save submission on studio task: %v", err)
	}

	// Exactly what the seeder issues after upserting a lesson's tasks.
	if _, err := pool.Exec(ctx,
		`DELETE FROM tasks
		 WHERE lesson_id=$1 AND source_key <> '' AND source_key <> ALL($2)`,
		lessonID, []string{"lab_t1"}); err != nil {
		t.Fatalf("prune: %v", err)
	}

	passed, err = NewSubmissionRepo(pool).PassedTaskIDs(ctx, userID, lessonID)
	if err != nil {
		t.Fatalf("passed tasks after prune: %v", err)
	}
	if !passed[studio] {
		t.Error("the prune deleted a task created in the studio, and the student's answer with it")
	}
	if !passed[first] {
		t.Error("the prune deleted a task the content still names")
	}

	// It must still remove what genuinely left the content.
	var gone int
	if err := pool.QueryRow(ctx,
		`INSERT INTO tasks (lesson_id, title, order_num, kind, check_script, source_key)
		 VALUES ($1, 'Удалённое из контента', 3, 'shell', 'true', 'lab_t_removed') RETURNING id`,
		lessonID).Scan(&gone); err != nil {
		t.Fatalf("create removed task: %v", err)
	}
	if _, err := pool.Exec(ctx,
		`DELETE FROM tasks
		 WHERE lesson_id=$1 AND source_key <> '' AND source_key <> ALL($2)`,
		lessonID, []string{"lab_t1"}); err != nil {
		t.Fatalf("second prune: %v", err)
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tasks WHERE id=$1`, gone).Scan(&n); err != nil {
		t.Fatalf("count removed task: %v", err)
	}
	if n != 0 {
		t.Error("a task that left the content was not pruned")
	}
}

// The dashboard counts a lab as done when every task in it passed. It used to
// count the rows of a join against submissions instead of the tasks, so a
// student who got a task right on the second attempt had three "tasks" in a
// two-task lab and could never reach the total. The visible effect was the one
// students reported: a lab they had finished never showed as finished.
func TestOverviewCountsFinishedLabDespiteRetries(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	userID, lessonID := testUser(t, pool), testLesson(t, pool)

	// Overview only looks at published content, which the helper does not publish.
	// One statement per Exec: pgx prepares them, and a prepared statement cannot
	// hold two commands.
	for _, q := range []string{
		`UPDATE lessons SET published = true WHERE id = $1`,
		`UPDATE modules SET published = true
		 WHERE id = (SELECT module_id FROM lessons WHERE id = $1)`,
	} {
		if _, err := pool.Exec(ctx, q, lessonID); err != nil {
			t.Fatalf("publish test lesson: %v", err)
		}
	}

	var taskIDs []int
	for i := 1; i <= 2; i++ {
		var id int
		err := pool.QueryRow(ctx,
			`INSERT INTO tasks (lesson_id, title, order_num, kind, check_script)
			 VALUES ($1, 'task', $2, 'shell', 'true') RETURNING id`, lessonID, i).Scan(&id)
		if err != nil {
			t.Fatalf("create task %d: %v", i, err)
		}
		taskIDs = append(taskIDs, id)
	}

	subs := NewSubmissionRepo(pool)
	// Four attempts over two tasks: the shape that used to break the count.
	for _, passed := range []bool{false, false, true} {
		if err := subs.Save(ctx, userID, taskIDs[0], "", "", "", passed); err != nil {
			t.Fatalf("save attempt: %v", err)
		}
	}
	if err := subs.Save(ctx, userID, taskIDs[1], "", "", "", true); err != nil {
		t.Fatalf("save second task: %v", err)
	}

	// A user created for this test has done nothing else, so these are exact
	// numbers rather than deltas, whatever else the database holds.
	o, err := NewProgressRepo(pool).Overview(ctx, userID)
	if err != nil {
		t.Fatalf("overview: %v", err)
	}
	if o.LabsDone != 1 {
		t.Errorf("LabsDone = %d, want 1: every task in the lab passed", o.LabsDone)
	}
	if o.TasksSolved != 2 {
		t.Errorf("TasksSolved = %d, want 2: retries are attempts, not tasks", o.TasksSolved)
	}
}
