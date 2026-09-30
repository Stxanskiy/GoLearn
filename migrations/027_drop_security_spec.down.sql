INSERT INTO specializations (slug, name, icon, description, order_num) VALUES
 ('security', 'Кибербезопасность', '🛡️', 'Наступательная и защитная безопасность', 3)
ON CONFLICT (slug) DO NOTHING;
-- The SQL Academy modules are not restored: their source was removed from the
-- seeder, so re-creating empty modules here would only bring the empty section back.
