-- Specialization for the Python / demo-exam path. Order 2 puts it right after
-- devops in the catalog; the module order_num band is 200 (see specBand in
-- cmd/seed/curriculum.go), which was left free when the SQL course was dropped.
INSERT INTO specializations (slug, name, description, order_num, published) VALUES
  ('python',
   'Python: от нуля до демоэкзамена',
   'Путь с первой строки кода до готового приложения с базой и интерфейсом: основы языка, качество кода, PostgreSQL с psycopg3, PyQt6 и сборка под демонстрационный экзамен.',
   2,
   -- hidden until the first course of the path is published: an empty section in
   -- the catalog reads as a bug, not as work in progress
   false)
ON CONFLICT (slug) DO NOTHING;
