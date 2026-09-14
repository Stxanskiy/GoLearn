-- 023: tracks hold specialization slugs, which are up to 40 characters.
ALTER TABLE modules ALTER COLUMN track TYPE VARCHAR(40);
ALTER TABLE lessons ALTER COLUMN track TYPE VARCHAR(40);
