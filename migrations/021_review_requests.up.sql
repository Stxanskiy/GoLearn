-- 021: moderation requests for publishing courses and lessons by non-admin authors.
CREATE TABLE IF NOT EXISTS review_requests (
    id            SERIAL      PRIMARY KEY,
    module_id     INT         NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    lesson_id     INT         REFERENCES lessons(id) ON DELETE CASCADE,
    status        VARCHAR(16) NOT NULL DEFAULT 'pending',
    note          TEXT        NOT NULL DEFAULT '',
    requested_by  INT         REFERENCES users(id) ON DELETE SET NULL,
    decision_note TEXT        NOT NULL DEFAULT '',
    decided_by    INT         REFERENCES users(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_at    TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_review_requests_pending
    ON review_requests (module_id, COALESCE(lesson_id, 0)) WHERE status = 'pending';
CREATE INDEX IF NOT EXISTS idx_review_requests_status ON review_requests (status, created_at);
