# Reference solutions for the PostgreSQL course labs (pg-start).
# Keyed sol_<lesson slug with - as _>_<n>, n = position among the lesson's
# checked tasks. Each runs as root in the lab container, like the student.

# Lab 1 (ch-pgs-lab1): installing PostgreSQL.
sol_ch_pgs_lab1_1='apt update && apt install -y postgresql'
sol_ch_pgs_lab1_2='sudo -u postgres createuser --superuser root'
sol_ch_pgs_lab1_3='printf "CREATE DATABASE pereplet;\n" | psql postgres'

# Lab 2 (ch-pgs-lab2): working in psql.
sol_ch_pgs_lab2_1='printf "CREATE DATABASE sandbox;\n" | psql pereplet'
sol_ch_pgs_lab2_2='printf "DROP DATABASE old_shop;\n" | psql pereplet'
sol_ch_pgs_lab2_3='printf "CREATE DATABASE archive_2025 OWNER postgres;\n" | psql pereplet'

# Lab 3 (ch-pgs-lab3): the first tables.
sol_ch_pgs_lab3_1='psql pereplet <<"SQL"
CREATE TABLE authors (
    id         serial PRIMARY KEY,
    full_name  text,
    birth_year integer,
    country    text
);
SQL'
sol_ch_pgs_lab3_2='psql pereplet <<"SQL"
CREATE TABLE books (
    id         serial PRIMARY KEY,
    title      text,
    price      numeric(8,2),
    page_count integer,
    in_stock   boolean
);
SQL'
sol_ch_pgs_lab3_3='printf "ALTER TABLE books ADD COLUMN published_year integer;\n" | psql pereplet'
sol_ch_pgs_lab3_4='printf "DROP TABLE bookz;\n" | psql pereplet'

# Lab 4 (ch-pgs-lab4): filling the catalog.
# Double-quoted here because the SQL is full of single-quoted text.
sol_ch_pgs_lab4_1="printf \"INSERT INTO authors (full_name, birth_year, country) VALUES ('Антон Чехов', 1860, 'Россия');\n\" | psql pereplet"
sol_ch_pgs_lab4_2="psql pereplet <<'SQL'
INSERT INTO authors (full_name, birth_year, country)
VALUES
    ('Лев Толстой',      1828, 'Россия'),
    ('Марк Твен',        1835, 'США'),
    ('Артур Конан Дойл', 1859, 'Великобритания');
SQL"
sol_ch_pgs_lab4_3="psql pereplet <<'SQL'
INSERT INTO books (title, price, page_count, in_stock, published_year)
VALUES
    ('Собака Баскервилей',      450.00, 256, true,  1902),
    ('Приключения Тома Сойера', 380.00, 288, false, 1876);
SQL"
sol_ch_pgs_lab4_4="printf \"INSERT INTO books (title, price) VALUES ('Анна Каренина', 870.00);\n\" | psql pereplet"
sol_ch_pgs_lab4_5="printf \"INSERT INTO books (title, price, page_count, in_stock, published_year) VALUES ('Затерянный мир', 430.00, 320, true, 1912);\n\" | psql pereplet"

# Lab 5 (ch-pgs-lab5): fixing the catalog lab 4 left, plus the intern's duplicate.
sol_ch_pgs_lab5_1='printf "UPDATE books SET price = 460.00 WHERE id = 2;\n" | psql pereplet'
sol_ch_pgs_lab5_2='printf "UPDATE books SET price = price * 1.1 WHERE id = 5;\n" | psql pereplet'
sol_ch_pgs_lab5_3='printf "UPDATE books SET in_stock = false WHERE page_count < 200;\n" | psql pereplet'
sol_ch_pgs_lab5_4='printf "UPDATE books SET page_count = 864, in_stock = true, published_year = 1878 WHERE id = 13;\n" | psql pereplet'
sol_ch_pgs_lab5_5="psql pereplet <<'SQL'
BEGIN;
DELETE FROM books WHERE id = 15;
SELECT id, title FROM books WHERE title = 'Затерянный мир';
COMMIT;
SQL"
