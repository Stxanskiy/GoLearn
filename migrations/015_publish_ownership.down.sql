DROP INDEX IF EXISTS idx_lessons_published;
DROP INDEX IF EXISTS idx_modules_published;
ALTER TABLE users           DROP COLUMN IF EXISTS blocked;
ALTER TABLE specializations DROP COLUMN IF EXISTS owner_id;
ALTER TABLE modules         DROP COLUMN IF EXISTS owner_id;
ALTER TABLE specializations DROP COLUMN IF EXISTS published;
ALTER TABLE lessons         DROP COLUMN IF EXISTS published;
ALTER TABLE modules         DROP COLUMN IF EXISTS published;
