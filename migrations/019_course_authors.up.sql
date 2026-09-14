-- 019: course co-authors; the owner stays in modules.owner_id.
CREATE TABLE IF NOT EXISTS course_authors (
    module_id  INT         NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    user_id    INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    added_by   INT         REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (module_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_course_authors_user ON course_authors(user_id);
