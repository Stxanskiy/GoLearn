-- The SQL course is withdrawn.
--
-- "Экспресс-курс по SQL" (sql-express) was the last course in the database
-- track, and it never worked: the first task of every one of its seven labs
-- told the student to run `psql -U student`, while the sandbox only ever
-- created a `root` role, so the very first command of each lab failed. One lab
-- had its tasks working against a database its setup never created. The course
-- was never published, so no student met any of this.
--
-- Removing the module cascades to its lessons, tasks and quizzes, and through
-- them to any progress and submissions against them. There are no submissions;
-- the handful of progress rows are from testing.
--
-- The seeder no longer builds this module (its content and fixtures are gone),
-- so this clears what earlier runs left behind.
DELETE FROM modules WHERE slug = 'sql-express';

-- The database specialization has nothing left in it. An empty section in the
-- catalogue is exactly what migration 027 removed for Кибербезопасность, and
-- for the same reason: students see a heading that leads nowhere.
--
-- Modules reference a specialization by their `track` column rather than a
-- foreign key, so this only removes the catalogue entry.
DELETE FROM specializations WHERE slug = 'database';
