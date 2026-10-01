-- Brings the catalogue section back, empty.
--
-- The course itself is not restored: its content and lab fixtures were deleted
-- from the seeder, so re-creating the module here would only produce a heading
-- with nothing behind it. To bring the course back, restore
-- cmd/seed/content/module_postgres_sql and cmd/seed/labs_sql_express.go, put
-- the entry back in importedModules() and the curriculum, and re-run the seeder
-- — and fix the `psql -U student` problem first.
INSERT INTO specializations (slug, name, description, order_num) VALUES
 ('database', 'Базы данных', 'SQL и работа с данными', 2)
ON CONFLICT (slug) DO NOTHING;
