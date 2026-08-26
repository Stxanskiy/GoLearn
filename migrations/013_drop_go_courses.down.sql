-- Restoring the Go courses means re-registering them in cmd/seed and re-seeding;
-- only the specialization row can be recreated here.
INSERT INTO specializations (slug, name, icon, description, order_num)
VALUES ('golang', 'Golang', '🐹', 'Go с нуля до backend-разработчика', 2)
ON CONFLICT (slug) DO NOTHING;
