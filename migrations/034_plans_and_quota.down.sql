DROP TABLE IF EXISTS sandbox_launches;
ALTER TABLE payments      DROP COLUMN IF EXISTS plan;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS plan;
-- A lifetime row has no date to put back, so it has to go before the column can
-- be NOT NULL again.
DELETE FROM subscriptions WHERE expires_at IS NULL;
ALTER TABLE subscriptions ALTER COLUMN expires_at SET NOT NULL;
