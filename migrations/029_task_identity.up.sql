-- Student progress was wiped by every deploy.
--
-- The seeder refreshed a lesson by deleting all of its tasks and inserting them
-- again, and submissions reference tasks with ON DELETE CASCADE. So each release
-- took every lab answer with it: a student came back to a lesson they had
-- finished and found it blank. Quiz answers went the same way through
-- quiz_questions.
--
-- source_key is the task's identity in the content ("lnav_lab1_t2_pwd"), which
-- lets the seeder update a task in place instead of recreating it. Tasks that
-- carry no such key in the source get a positional one.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS source_key TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_lesson_source_key
  ON tasks (lesson_id, source_key) WHERE source_key <> '';

-- Same for quiz questions: quiz_answers cascade from them.
ALTER TABLE quiz_questions ADD COLUMN IF NOT EXISTS source_key TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_quiz_questions_source_key
  ON quiz_questions (quiz_id, source_key) WHERE source_key <> '';

-- A lesson has at most one quiz, but nothing said so: the seeder could only
-- refresh it by deleting it, which cascaded the answers away. One per lesson is
-- what the product always meant.
DELETE FROM quizzes q USING quizzes k
  WHERE q.lesson_id = k.lesson_id AND q.id > k.id;
CREATE UNIQUE INDEX IF NOT EXISTS idx_quizzes_lesson ON quizzes (lesson_id);
