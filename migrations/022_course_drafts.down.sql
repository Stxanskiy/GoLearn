DROP INDEX IF EXISTS idx_review_requests_pending;
ALTER TABLE review_requests DROP COLUMN IF EXISTS kind;
ALTER TABLE review_requests ADD COLUMN IF NOT EXISTS lesson_id INT REFERENCES lessons(id) ON DELETE CASCADE;
CREATE UNIQUE INDEX IF NOT EXISTS idx_review_requests_pending
    ON review_requests (module_id, COALESCE(lesson_id, 0)) WHERE status = 'pending';
DELETE FROM modules WHERE draft_of IS NOT NULL;
DROP INDEX IF EXISTS idx_modules_draft_of;
ALTER TABLE tasks          DROP COLUMN IF EXISTS origin_id;
ALTER TABLE quiz_questions DROP COLUMN IF EXISTS origin_id;
ALTER TABLE lessons        DROP COLUMN IF EXISTS origin_id;
ALTER TABLE modules        DROP COLUMN IF EXISTS draft_of;
