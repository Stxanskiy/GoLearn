-- GoLearn LMS Schema

CREATE TABLE IF NOT EXISTS modules (
    id          SERIAL PRIMARY KEY,
    slug        VARCHAR(100) UNIQUE NOT NULL,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    order_num   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS lessons (
    id          SERIAL PRIMARY KEY,
    module_id   INT NOT NULL REFERENCES modules(id) ON DELETE CASCADE,
    slug        VARCHAR(100) NOT NULL,
    title       VARCHAR(255) NOT NULL,
    content     TEXT NOT NULL DEFAULT '',
    order_num   INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(module_id, slug)
);

CREATE TABLE IF NOT EXISTS quizzes (
    id        SERIAL PRIMARY KEY,
    lesson_id INT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title     VARCHAR(255) NOT NULL DEFAULT 'Quiz'
);

CREATE TABLE IF NOT EXISTS quiz_questions (
    id            SERIAL PRIMARY KEY,
    quiz_id       INT NOT NULL REFERENCES quizzes(id) ON DELETE CASCADE,
    question      TEXT NOT NULL,
    options       JSONB NOT NULL DEFAULT '[]',
    correct_index INT NOT NULL DEFAULT 0,
    explanation   TEXT NOT NULL DEFAULT '',
    order_num     INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tasks (
    id          SERIAL PRIMARY KEY,
    lesson_id   INT NOT NULL REFERENCES lessons(id) ON DELETE CASCADE,
    title       VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    hints       TEXT NOT NULL DEFAULT '',
    solution    TEXT NOT NULL DEFAULT '',
    order_num   INT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS progress (
    id           SERIAL PRIMARY KEY,
    lesson_id    INT NOT NULL UNIQUE REFERENCES lessons(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'not_started',
    quiz_score   INT,
    quiz_total   INT,
    notes        TEXT NOT NULL DEFAULT '',
    completed_at TIMESTAMPTZ,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_lessons_module ON lessons(module_id);
CREATE INDEX idx_progress_status ON progress(status);
