-- 002: Add tracks, difficulty levels, glossary, and code runner support

-- Track: which learning path a module belongs to
ALTER TABLE modules ADD COLUMN IF NOT EXISTS track VARCHAR(20) NOT NULL DEFAULT 'backend';
-- backend | devops | shared (both tracks)

-- Difficulty level for modules
ALTER TABLE modules ADD COLUMN IF NOT EXISTS difficulty VARCHAR(20) NOT NULL DEFAULT 'beginner';
-- beginner | intermediate | advanced | expert

-- Prerequisites: JSON array of module slugs that should be completed first
ALTER TABLE modules ADD COLUMN IF NOT EXISTS prerequisites JSONB NOT NULL DEFAULT '[]';

-- Difficulty level for individual tasks
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS difficulty VARCHAR(20) NOT NULL DEFAULT 'easy';
-- easy | medium | hard

-- Glossary: JSON array of {term, definition} explaining functions/concepts used in the task
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS glossary JSONB NOT NULL DEFAULT '[]';

-- Test cases for automatic code checking: JSON array of {input, expected_output}
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS test_cases JSONB NOT NULL DEFAULT '[]';

-- Starter code template for the code editor
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS starter_code TEXT NOT NULL DEFAULT '';

-- Difficulty level for lessons
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS difficulty VARCHAR(20) NOT NULL DEFAULT 'beginner';

-- Track for lessons (inherits from module but can override)
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS track VARCHAR(20) NOT NULL DEFAULT 'backend';

-- Code submissions: track user's code attempts
CREATE TABLE IF NOT EXISTS submissions (
    id          SERIAL PRIMARY KEY,
    task_id     INT NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
    code        TEXT NOT NULL,
    output      TEXT NOT NULL DEFAULT '',
    errors      TEXT NOT NULL DEFAULT '',
    passed      BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_submissions_task ON submissions(task_id);
CREATE INDEX IF NOT EXISTS idx_modules_track ON modules(track);
CREATE INDEX IF NOT EXISTS idx_modules_difficulty ON modules(difficulty);
CREATE INDEX IF NOT EXISTS idx_lessons_track ON lessons(track);
