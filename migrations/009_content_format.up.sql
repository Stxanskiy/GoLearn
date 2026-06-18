-- 009: content format (md|html) for admin-authored content
ALTER TABLE lessons ADD COLUMN IF NOT EXISTS format VARCHAR(8) NOT NULL DEFAULT 'html';
ALTER TABLE tasks   ADD COLUMN IF NOT EXISTS format VARCHAR(8) NOT NULL DEFAULT 'html';
