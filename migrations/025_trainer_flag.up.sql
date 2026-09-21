ALTER TABLE modules ADD COLUMN IF NOT EXISTS is_trainer BOOLEAN NOT NULL DEFAULT FALSE;

-- The gym pseudo-track becomes a flag; its courses are Linux and Git, so they move to devops.
UPDATE modules SET is_trainer = TRUE, track = 'devops' WHERE track = 'gym';
UPDATE lessons SET track = 'devops' WHERE track = 'gym';
