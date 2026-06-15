-- 003 rollback
ALTER TABLE submissions DROP COLUMN IF EXISTS user_id;
ALTER TABLE progress DROP CONSTRAINT IF EXISTS progress_user_lesson_unique;
ALTER TABLE progress DROP COLUMN IF EXISTS user_id;
ALTER TABLE progress ADD CONSTRAINT progress_lesson_id_key UNIQUE(lesson_id);
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
