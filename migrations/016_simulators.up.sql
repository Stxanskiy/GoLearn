-- 016: simulators become admin-manageable. The full turn-based scenario (metrics,
-- turns, choices, effects) lives in `data` JSONB; the columns mirror the fields used
-- for listing so the catalogue is cheap. Published/owner match courses/sections.
CREATE TABLE IF NOT EXISTS simulators (
    slug       VARCHAR(60)  PRIMARY KEY,
    title      VARCHAR(200) NOT NULL DEFAULT '',
    icon       VARCHAR(16)  NOT NULL DEFAULT '',
    role       VARCHAR(160) NOT NULL DEFAULT '',
    order_num  INT          NOT NULL DEFAULT 0,
    published  BOOLEAN      NOT NULL DEFAULT TRUE,
    owner_id   INT REFERENCES users(id) ON DELETE SET NULL,
    data       JSONB        NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
