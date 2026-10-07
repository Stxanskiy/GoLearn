-- Removes the section again. Modules link to it by track, not by a foreign key,
-- so the courses themselves are untouched; they just lose their heading.
DELETE FROM specializations WHERE slug = 'database';
