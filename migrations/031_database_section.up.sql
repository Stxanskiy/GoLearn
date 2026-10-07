-- 031: the "Базы данных" section returns for the PostgreSQL path.
--
-- 030 removed it together with the withdrawn SQL course. The new courses are
-- seeded as drafts (importSpec.Draft), so the section comes back hidden too:
-- an empty heading in the catalogue is what 030 was cleaning up. An admin
-- publishes the section and its courses together when they are ready.
INSERT INTO specializations (slug, name, description, order_num, published) VALUES
 ('database', 'Базы данных', 'PostgreSQL с нуля: от установки и SQL до эксплуатации', 2, FALSE)
ON CONFLICT (slug) DO NOTHING;
