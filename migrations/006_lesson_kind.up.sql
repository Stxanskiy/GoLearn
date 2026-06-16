-- 006: Lesson type (theory/quiz/lab/sim) so each chapter is one typed item,
-- matching the reference course structure (Урок / Тест / Лаб. работа / Симулятор).

ALTER TABLE lessons ADD COLUMN IF NOT EXISTS kind VARCHAR(16) NOT NULL DEFAULT 'theory';
-- theory | quiz | lab | sim

-- Terminal lab config (for kind='lab').
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS vm_image VARCHAR(64) NOT NULL DEFAULT '';
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS vm_init  TEXT NOT NULL DEFAULT '';
