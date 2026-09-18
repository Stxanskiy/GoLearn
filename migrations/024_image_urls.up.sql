-- 024: course and specialization icons are uploaded images stored in object storage.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS icon_url TEXT NOT NULL DEFAULT '';
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS icon_url TEXT NOT NULL DEFAULT '';
