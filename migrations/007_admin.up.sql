-- 007: roles + admin-managed specializations + content ownership

-- User roles. First-created user becomes admin (idempotent: only if no admin yet).
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(16) NOT NULL DEFAULT 'student';
UPDATE users SET role = 'admin'
 WHERE id = (SELECT MIN(id) FROM users)
   AND NOT EXISTS (SELECT 1 FROM users WHERE role = 'admin');

-- Specializations (top-level sections) — admin-manageable.
CREATE TABLE IF NOT EXISTS specializations (
    slug        VARCHAR(40) PRIMARY KEY,
    name        VARCHAR(80) NOT NULL,
    icon        VARCHAR(16) NOT NULL DEFAULT '',
    description TEXT        NOT NULL DEFAULT '',
    order_num   INT         NOT NULL DEFAULT 0
);
INSERT INTO specializations (slug, name, icon, description, order_num) VALUES
 ('devops',   'DevOps',            '♾️', 'Linux, Docker, Kubernetes, Git, CI/CD и мониторинг', 1),
 ('golang',   'Golang',            '🐹', 'Go с нуля до backend-разработчика',                  2),
 ('security', 'Кибербезопасность', '🛡️', 'Наступательная и защитная безопасность',             3),
 ('database', 'Базы данных',       '🗄️', 'SQL и проектирование баз данных',                    4)
ON CONFLICT (slug) DO NOTHING;

-- Content ownership: 'seed' rows are managed by the seeder; 'admin' rows are
-- created in the UI and must survive re-seeds.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS source VARCHAR(16) NOT NULL DEFAULT 'seed';
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS source VARCHAR(16) NOT NULL DEFAULT 'seed';
