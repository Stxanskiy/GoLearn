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

# Lab 6 (ch-pgs-lab6): linking the tables.
# The catalog is lab 5's: authors 1-8 in lab 4's order (5 is Чехов, who has no
# books), books 1-14. Task 6 uses ids 1 for the customer and the order: the
# checks never take a number from a sequence, so the student's first rows get
# 1 too. Double-quoted where the SQL is full of single-quoted text.
sol_ch_pgs_lab6_1="psql pereplet <<'SQL'
ALTER TABLE books ADD COLUMN author_id integer REFERENCES authors (id);
SELECT id, full_name FROM authors;
SELECT id, title FROM books ORDER BY id;
UPDATE books SET author_id = 1 WHERE title = 'Басни';
UPDATE books SET author_id = 2 WHERE title IN ('Евгений Онегин', 'Капитанская дочка', 'Повести Белкина');
UPDATE books SET author_id = 3 WHERE title IN ('Идиот', 'Братья Карамазовы', 'Белые ночи');
UPDATE books SET author_id = 4 WHERE id IN (8, 9, 10);
UPDATE books SET author_id = 6 WHERE title = 'Анна Каренина';
UPDATE books SET author_id = 7 WHERE title = 'Приключения Тома Сойера';
UPDATE books SET author_id = 8 WHERE title IN ('Собака Баскервилей', 'Затерянный мир');
SELECT id, title, author_id FROM books ORDER BY id;
SQL"
sol_ch_pgs_lab6_2='psql pereplet <<"SQL"
ALTER TABLE books ALTER COLUMN title SET NOT NULL;
ALTER TABLE books ALTER COLUMN price SET NOT NULL;
ALTER TABLE books ADD CHECK (price > 0);
SELECT id, title, price FROM books WHERE price <= 0;
UPDATE books SET price = 715.00 WHERE id = 5;
ALTER TABLE books ADD CHECK (price > 0);
ALTER TABLE books ALTER COLUMN in_stock SET NOT NULL;
ALTER TABLE books ALTER COLUMN in_stock SET DEFAULT true;
ALTER TABLE authors ALTER COLUMN full_name SET NOT NULL;
\d books
\d authors
SQL'
sol_ch_pgs_lab6_3='psql pereplet <<"SQL"
CREATE TABLE customers (
    id            serial PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    full_name     text NOT NULL,
    city          text,
    registered_at date NOT NULL DEFAULT current_date
);
SQL'
sol_ch_pgs_lab6_4="psql pereplet <<'SQL'
CREATE TABLE orders (
    id          serial PRIMARY KEY,
    customer_id integer NOT NULL REFERENCES customers (id),
    created_at  date NOT NULL DEFAULT current_date,
    status      text NOT NULL DEFAULT 'new'
                CHECK (status IN ('new', 'paid', 'shipped', 'cancelled'))
);
SQL"
sol_ch_pgs_lab6_5='psql pereplet <<"SQL"
CREATE TABLE order_items (
    order_id integer NOT NULL REFERENCES orders (id),
    book_id  integer NOT NULL REFERENCES books (id),
    quantity integer NOT NULL CHECK (quantity > 0),
    PRIMARY KEY (order_id, book_id)
);
SQL'
sol_ch_pgs_lab6_6="psql pereplet <<'SQL'
INSERT INTO customers (email, full_name, city) VALUES ('anna.smirnova@example.com', 'Анна Смирнова', 'Москва');
SELECT id FROM customers WHERE email = 'anna.smirnova@example.com';
INSERT INTO orders (customer_id) VALUES (1);
SELECT * FROM orders;
SELECT id, title FROM books WHERE title IN ('Евгений Онегин', 'Анна Каренина');
INSERT INTO order_items (order_id, book_id, quantity) VALUES (1, 2, 2), (1, 13, 1);
SQL"

# Lab 7 (ch-pgs-lab7): importing the catalog.
# Double-quoted where the SQL has single quotes: \\ stands for the one
# backslash of \copy.
sol_ch_pgs_lab7_1='psql pereplet <<"SQL"
CREATE TABLE genres (
    id   integer PRIMARY KEY,
    name text NOT NULL UNIQUE
);
SQL'
sol_ch_pgs_lab7_2="psql pereplet <<'SQL'
\\copy genres (id, name) FROM '/root/genres.csv' WITH (FORMAT csv, HEADER, DELIMITER ';')
SQL"
sol_ch_pgs_lab7_3='psql pereplet <<"SQL"
CREATE TABLE book_genres (
    book_id  integer NOT NULL REFERENCES books (id),
    genre_id integer NOT NULL REFERENCES genres (id),
    PRIMARY KEY (book_id, genre_id)
);
SQL'
# The student opens the file in nano, goes to line 12 from CONTEXT and cuts it.
sol_ch_pgs_lab7_4="sed -i 12d /root/book_genres.csv && psql pereplet <<'SQL'
\\copy book_genres (book_id, genre_id) FROM '/root/book_genres.csv' WITH (FORMAT csv, HEADER)
SQL"
sol_ch_pgs_lab7_5="psql pereplet <<'SQL'
\\copy (SELECT title, price FROM books WHERE in_stock = true ORDER BY price) TO '/root/in_stock.csv' WITH (FORMAT csv, HEADER, DELIMITER ';')
SQL"

# Lab 8 (ch-pgs-lab8): the project, a database for a helpdesk.
# The schema goes into /root/helpdesk.sql and runs with \i, the way lesson 18
# teaches: each step rewrites the file with one table more, as a student adds
# to it in nano, and running it again rebuilds the tables from scratch. The rows
# of the last task are typed in psql after a look at the ids (fresh tables: 1,
# 2 and 1).
sol_ch_pgs_lab8_1='printf "CREATE DATABASE helpdesk;\n" | psql postgres'
sol_ch_pgs_lab8_2='cat > /root/helpdesk.sql <<"SQL"
-- Служба поддержки: клиенты.
DROP TABLE IF EXISTS clients;

CREATE TABLE clients (
    id        serial PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL UNIQUE,
    phone     text
);
SQL
psql helpdesk <<"SQL"
\i /root/helpdesk.sql
SQL'
sol_ch_pgs_lab8_3='cat > /root/helpdesk.sql <<"SQL"
-- Служба поддержки: клиенты и сотрудники.
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS clients;

CREATE TABLE clients (
    id        serial PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL UNIQUE,
    phone     text
);

CREATE TABLE agents (
    id        serial PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL UNIQUE,
    is_active boolean NOT NULL DEFAULT true
);
SQL
psql helpdesk <<"SQL"
\i /root/helpdesk.sql
SQL'
sol_ch_pgs_lab8_4='cat > /root/helpdesk.sql <<"SQL"
-- Служба поддержки: клиенты, сотрудники, обращения.
-- Файл можно выполнять много раз: сначала удаляем старые таблицы.
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS agents;
DROP TABLE IF EXISTS clients;

CREATE TABLE clients (
    id        serial PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL UNIQUE,
    phone     text
);

CREATE TABLE agents (
    id        serial PRIMARY KEY,
    full_name text NOT NULL,
    email     text NOT NULL UNIQUE,
    is_active boolean NOT NULL DEFAULT true
);

CREATE TABLE tickets (
    id         serial PRIMARY KEY,
    client_id  integer NOT NULL REFERENCES clients (id),
    agent_id   integer REFERENCES agents (id),
    subject    text NOT NULL,
    priority   text NOT NULL DEFAULT '\''normal'\''
               CHECK (priority IN ('\''low'\'', '\''normal'\'', '\''high'\'')),
    status     text NOT NULL DEFAULT '\''open'\''
               CHECK (status IN ('\''open'\'', '\''in_progress'\'', '\''closed'\'')),
    created_at date NOT NULL DEFAULT current_date,
    closed_at  date,
    CHECK (closed_at >= created_at)
);
SQL
psql helpdesk <<"SQL"
\i /root/helpdesk.sql
SQL'
sol_ch_pgs_lab8_5="psql helpdesk <<'SQL'
INSERT INTO clients (full_name, email, phone) VALUES
    ('Мария Волкова', 'maria.volkova@example.com', '+7 912 345-67-89');
INSERT INTO clients (full_name, email) VALUES
    ('Денис Орлов', 'denis.orlov@example.com');
INSERT INTO agents (full_name, email) VALUES
    ('Елена Морозова', 'morozova@support.example.com');
SELECT id, full_name FROM clients;
SELECT id, full_name FROM agents;
INSERT INTO tickets (client_id, agent_id, subject, priority, status, created_at, closed_at) VALUES
    (1, 1, 'Не доставили заказ', 'high', 'closed', '2026-10-01', '2026-10-03');
INSERT INTO tickets (client_id, agent_id, subject, status, created_at) VALUES
    (2, 1, 'В коробке не хватает фишек', 'in_progress', '2026-10-05');
INSERT INTO tickets (client_id, subject, priority) VALUES
    (1, 'Как вернуть игру', 'low');
SQL"
