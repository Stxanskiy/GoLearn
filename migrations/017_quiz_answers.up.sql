-- 017: per-user quiz answers; the first answer to a question is kept until the attempt is reset.
CREATE TABLE IF NOT EXISTS quiz_answers (
    user_id        INT         NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    question_id    INT         NOT NULL REFERENCES quiz_questions(id) ON DELETE CASCADE,
    selected_index INT         NOT NULL,
    answered_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, question_id)
);
