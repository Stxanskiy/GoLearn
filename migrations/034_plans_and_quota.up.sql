-- 034: what a student can buy, and what the free tier gets without buying.

-- A lifetime subscription has no expiry, and NULL says exactly that. Every
-- query already reads correctly against it: `expires_at <= now()` is never true
-- for NULL, so the sweep leaves a lifetime row alone, and `ORDER BY expires_at
-- DESC` puts NULLs first in PostgreSQL, so a lifetime row wins over a dated one
-- when both exist.
--
-- The alternative — a date far in the future — looks simpler and then quietly
-- expires on a day nobody is watching.
ALTER TABLE subscriptions ALTER COLUMN expires_at DROP NOT NULL;

-- Which plan was bought. Every row that exists today was a month, which is why
-- that is the default rather than something like 'unknown'.
ALTER TABLE subscriptions ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT 'month';
ALTER TABLE payments      ADD COLUMN IF NOT EXISTS plan TEXT NOT NULL DEFAULT 'month';

-- One row per sandbox start, so the free tier's weekly allowance can be counted.
--
-- A row per launch rather than a counter on the user: a counter cannot answer
-- "when do I get more", which is the first thing a student asks when they run
-- out, and it cannot be audited when someone claims they were charged wrongly.
--
-- sandbox_key is text, not a lesson id: trainers and simulators have sandboxes
-- too and no lesson row behind them.
CREATE TABLE IF NOT EXISTS sandbox_launches (
    id          BIGSERIAL   PRIMARY KEY,
    user_id     INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    sandbox_key TEXT        NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The only query is "how many did this user start since <date>", newest first.
CREATE INDEX IF NOT EXISTS sandbox_launches_user_idx
    ON sandbox_launches (user_id, created_at DESC);
