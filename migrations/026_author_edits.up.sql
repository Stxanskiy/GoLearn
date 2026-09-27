-- Marks a row an author changed through the studio.
--
-- Until now the seeder ran on every deploy with ON CONFLICT DO UPDATE and a
-- prune, so it rewrote every course it knows about. Every module and lesson in
-- the database is source='seed', which meant an edit made in the studio lived
-- exactly until the next deploy and then vanished without a word.
--
-- With this column the seeder can leave alone anything a human has touched.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ;
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS edited_at TIMESTAMPTZ;
