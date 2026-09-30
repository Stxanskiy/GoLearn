-- The Кибербезопасность specialization has been published since migration 007
-- with no courses in it: students saw an empty section in the catalogue. The two
-- security labs it was meant to hold were folded into other courses long ago.
--
-- Modules reference a specialization by the `track` column, not a foreign key,
-- so this only removes the catalogue entry. Nothing currently points at it.
DELETE FROM specializations WHERE slug = 'security';

-- SQL Academy shipped 165 exercises as theory-only lessons: the question was
-- there, but nothing to answer it with — no tasks, no checks, and no SQL editor
-- on the frontend since the rewrite. Published, unusable, and in a paid section.
-- The seeder no longer builds these modules; this clears what earlier runs left.
DELETE FROM modules WHERE slug IN ('sql-easy', 'sql-medium', 'sql-hard');
