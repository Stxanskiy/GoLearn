-- 013: GoLearn is a DevOps course — the Go track is removed from the platform.
--
-- The seeder prunes seed-managed modules it no longer registers, so this only
-- has to remove what the seeder does not own: the specialization row itself
-- (added by 007), which would otherwise show as an empty section in the
-- catalog. Deleting the modules here too keeps a database that is not
-- re-seeded consistent; lessons, quizzes, tasks, progress and submissions go
-- with them through ON DELETE CASCADE.
DELETE FROM modules WHERE track = 'golang';
DELETE FROM specializations WHERE slug = 'golang';
