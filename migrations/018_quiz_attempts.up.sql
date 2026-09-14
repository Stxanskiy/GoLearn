-- 018: every quiz submission is kept as an attempt (score + per-question answers) for statistics.
CREATE TABLE IF NOT EXISTS quiz_attempts (
    id         BIGSERIAL   PRIMARY KEY,
    user_id    INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id  INT         NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    score      INT         NOT NULL,
    total      INT         NOT NULL,
    answers    JSONB       NOT NULL DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_quiz_attempts_user_lesson ON quiz_attempts (user_id, lesson_id);
