-- Only the courses this migration moved go back: trainers that still sit in devops.
UPDATE modules SET track = 'gym' WHERE is_trainer AND track = 'devops';
UPDATE lessons l SET track = 'gym' FROM modules m WHERE m.id = l.module_id AND m.track = 'gym';

ALTER TABLE modules DROP COLUMN IF EXISTS is_trainer;
