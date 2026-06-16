-- 008: per-option explanations for quiz questions (feedback on any answer)
ALTER TABLE quiz_questions ADD COLUMN IF NOT EXISTS option_explanations JSONB NOT NULL DEFAULT '[]';
