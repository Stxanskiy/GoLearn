-- 005: Course showcase metadata on modules (catalog cards: cover, tag, label, duration)

-- Explicit category tag shown on the card (e.g. Linux, Docker). Empty -> derived in handler.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS category VARCHAR(40) NOT NULL DEFAULT '';

-- Course type label: Старт | Практика | Вызов. Empty -> derived from difficulty.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS label VARCHAR(40) NOT NULL DEFAULT '';

-- Free-form topic tags shown as chips: JSON array of strings.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]';

-- Real cover image URL/path. Empty -> generated SVG at /api/courses/{slug}/cover.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS cover_image TEXT NOT NULL DEFAULT '';

-- Gradient accent key for the generated cover. Empty -> derived from category.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS accent VARCHAR(40) NOT NULL DEFAULT '';

-- Estimated time to complete, minutes. 0 -> derived from lesson count.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS est_minutes INT NOT NULL DEFAULT 0;
