-- 015: draft/publish state + per-course ownership + user blocking.
-- Existing content defaults to published=TRUE so nothing disappears on migration.
-- owner_id is nullable: seed/system content has no owner (visible to every admin);
-- admin-created courses get the creator's id and their drafts stay private to them.

ALTER TABLE modules         ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE lessons         ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS published BOOLEAN NOT NULL DEFAULT TRUE;

ALTER TABLE modules         ADD COLUMN IF NOT EXISTS owner_id INT REFERENCES users(id) ON DELETE SET NULL;
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS owner_id INT REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE users           ADD COLUMN IF NOT EXISTS blocked BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX IF NOT EXISTS idx_modules_published ON modules(published);
CREATE INDEX IF NOT EXISTS idx_lessons_published ON lessons(published);
