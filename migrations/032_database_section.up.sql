-- 032: the "Базы данных" section returns for the PostgreSQL path.
--
-- 030 removed it together with the withdrawn SQL course. It comes back
-- published, because pg-start ships with it: ListPublished filters on
-- `WHERE published`, so a hidden section would hide the course inside it and
-- "published pg-start" would show an empty catalogue. pg-sql, pg-ops and
-- gym-sql stay drafts (importSpec.Draft) and remain invisible until an admin
-- releases them, so the heading is not empty - it carries one finished course.
--
-- order_num 3, not 2: 031 gave 2 to the python specialization, and the
-- column is not unique, so a duplicate would not error - the two sections
-- would just sort against each other by slug.
INSERT INTO specializations (slug, name, description, order_num, published) VALUES
 ('database', 'Базы данных', 'PostgreSQL с нуля: от установки и SQL до эксплуатации', 3, TRUE)
ON CONFLICT (slug) DO NOTHING;
