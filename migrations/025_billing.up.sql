-- 025: platform subscription.
--
-- Access is decided per course: a course is either free or part of the
-- subscription, and an admin moves it between the two. The default is free so
-- applying this migration locks nothing — today every course is open, and
-- silently closing them on deploy would be a surprise.
ALTER TABLE modules ADD COLUMN IF NOT EXISTS access_tier TEXT NOT NULL DEFAULT 'free'
    CHECK (access_tier IN ('free', 'subscription'));

-- One row per purchase period. Expiry is stored rather than derived so a manual
-- grant (support, a gift, a fixed trial) has the same shape as a paid one.
CREATE TABLE IF NOT EXISTS subscriptions (
    id           SERIAL PRIMARY KEY,
    user_id      INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       TEXT        NOT NULL CHECK (status IN ('active', 'expired', 'canceled')),
    started_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at   TIMESTAMPTZ NOT NULL,
    -- 'stub' until a real provider is wired in; provider_ref holds its id.
    provider     TEXT        NOT NULL DEFAULT 'stub',
    provider_ref TEXT        NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A user holds at most one active subscription; a renewal extends that row or
-- opens a new one once the old is no longer active.
CREATE UNIQUE INDEX IF NOT EXISTS subscriptions_one_active
    ON subscriptions (user_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS subscriptions_user_idx ON subscriptions (user_id);

-- Every attempt is recorded, not only the successful ones: without the failed
-- and abandoned rows there is no telling a provider outage from no demand.
CREATE TABLE IF NOT EXISTS payments (
    id              SERIAL PRIMARY KEY,
    user_id         INTEGER     NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id INTEGER     REFERENCES subscriptions(id) ON DELETE SET NULL,
    status          TEXT        NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'canceled')),
    -- Minor units (kopeks) so no amount is ever held in a float.
    amount_minor    BIGINT      NOT NULL DEFAULT 0,
    currency        TEXT        NOT NULL DEFAULT 'RUB',
    months          INTEGER     NOT NULL DEFAULT 1,
    provider        TEXT        NOT NULL DEFAULT 'stub',
    provider_ref    TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at         TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS payments_user_idx ON payments (user_id, created_at DESC);
-- A provider replays webhooks; this reference is what makes handling one twice safe.
CREATE UNIQUE INDEX IF NOT EXISTS payments_provider_ref
    ON payments (provider, provider_ref) WHERE provider_ref <> '';
