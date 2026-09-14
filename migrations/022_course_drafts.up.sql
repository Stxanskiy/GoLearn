-- 022: drafts of published courses; review requests target whole courses.
ALTER TABLE modules        ADD COLUMN IF NOT EXISTS draft_of  INT REFERENCES modules(id) ON DELETE CASCADE;
ALTER TABLE lessons        ADD COLUMN IF NOT EXISTS origin_id INT;
ALTER TABLE quiz_questions ADD COLUMN IF NOT EXISTS origin_id INT;
ALTER TABLE tasks          ADD COLUMN IF NOT EXISTS origin_id INT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_modules_draft_of ON modules(draft_of) WHERE draft_of IS NOT NULL;

DROP INDEX IF EXISTS idx_review_requests_pending;
DELETE FROM review_requests WHERE lesson_id IS NOT NULL;
ALTER TABLE review_requests DROP COLUMN IF EXISTS lesson_id;
ALTER TABLE review_requests ADD COLUMN IF NOT EXISTS kind VARCHAR(16) NOT NULL DEFAULT 'publish';
CREATE UNIQUE INDEX IF NOT EXISTS idx_review_requests_pending ON review_requests(module_id) WHERE status = 'pending';
