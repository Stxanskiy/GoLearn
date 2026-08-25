-- 003: Multi-user support (auth + per-user progress)

CREATE TABLE IF NOT EXISTS users (
    id            SERIAL PRIMARY KEY,
    email         VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name          VARCHAR(100) NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS sessions (
    token      VARCHAR(64) PRIMARY KEY,
    user_id    INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

-- Add user_id to progress (nullable first for migration safety)
ALTER TABLE progress ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id) ON DELETE CASCADE;

-- Drop old unique constraint (lesson_id only) and add composite.
-- ADD CONSTRAINT has no IF NOT EXISTS, so drop the target name first: the file
-- must be safe to apply twice (a database may already carry it from an older
-- deploy that applied migrations without recording them).
ALTER TABLE progress DROP CONSTRAINT IF EXISTS progress_lesson_id_key;
ALTER TABLE progress DROP CONSTRAINT IF EXISTS progress_user_lesson_unique;
ALTER TABLE progress ADD CONSTRAINT progress_user_lesson_unique UNIQUE(user_id, lesson_id);

-- Add user_id to submissions
ALTER TABLE submissions ADD COLUMN IF NOT EXISTS user_id INT REFERENCES users(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS idx_progress_user ON progress(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_user ON submissions(user_id);
