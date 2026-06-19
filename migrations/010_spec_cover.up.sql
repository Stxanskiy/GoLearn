-- 010: cover image for specializations (section cards)
ALTER TABLE specializations ADD COLUMN IF NOT EXISTS cover_image TEXT NOT NULL DEFAULT '';
