-- 014: remove course illustrations and embedded cover art from lesson content.
--
-- The images are being dropped from lessons (to be replaced later); the
-- data-URI covers also bloat every lesson by hundreds of KB. The seeder's
-- cleanContent() strips the same <img>/<picture> tags on import, so a re-seed
-- stays consistent with this migration.
UPDATE lessons
SET content = regexp_replace(
                regexp_replace(content, '<picture[^>]*>.*?</picture>', '', 'gi'),
                '<img[^>]*>', '', 'gi')
WHERE content ~* '<img|<picture';
