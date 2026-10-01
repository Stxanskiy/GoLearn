DROP INDEX IF EXISTS idx_quizzes_lesson;
DROP INDEX IF EXISTS idx_quiz_questions_source_key;
ALTER TABLE quiz_questions DROP COLUMN IF EXISTS source_key;
DROP INDEX IF EXISTS idx_tasks_lesson_source_key;
ALTER TABLE tasks DROP COLUMN IF EXISTS source_key;
