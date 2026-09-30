ALTER TABLE modules         ADD COLUMN IF NOT EXISTS icon_url TEXT NOT NULL DEFAULT '';
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS icon     VARCHAR(16) NOT NULL DEFAULT '';
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS icon_url TEXT NOT NULL DEFAULT '';
-- The uploaded icon URLs themselves are not restored: the objects they pointed
-- at are deleted with the columns.
