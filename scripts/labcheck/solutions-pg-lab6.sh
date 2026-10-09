# Reference solutions for lab 6 of pg-start (ch-pgs-lab6): linking the tables.
# A file of its own while the labs are written in parallel — lab.sh sources
# every solutions*.sh; it belongs at the end of solutions-pg.sh. Keyed
# sol_ch_pgs_lab6_<n>, n = position among the lesson's checked tasks. Each runs
# as root in the lab container, like the student: SQL typed into psql pereplet.
#
# Task 6 uses ids 1 for the customer and the order: the checks never take a
# number from a sequence, so the student's first rows get 1 too. Double-quoted
# where the SQL is full of single-quoted text.

# Lab 6 (ch-pgs-lab6): linking the tables.
sol_ch_pgs_lab6_1="psql pereplet <<'SQL'
ALTER TABLE books ADD COLUMN author_id integer REFERENCES authors (id);
SELECT id, full_name FROM authors;
UPDATE books SET author_id = 1 WHERE title IN ('Евгений Онегин', 'Капитанская дочка');
UPDATE books SET author_id = 2 WHERE title IN ('Мёртвые души', 'Шинель', 'Ревизор');
UPDATE books SET author_id = 3 WHERE title = 'Идиот';
UPDATE books SET author_id = 4 WHERE title = 'Анна Каренина';
UPDATE books SET author_id = 5 WHERE title IN ('Дама с собачкой', 'Человек в футляре');
SELECT id, title, author_id FROM books ORDER BY id;
SQL"
sol_ch_pgs_lab6_2='psql pereplet <<"SQL"
ALTER TABLE books ALTER COLUMN title SET NOT NULL;
ALTER TABLE books ALTER COLUMN price SET NOT NULL;
ALTER TABLE books ADD CHECK (price > 0);
SELECT id, title, price FROM books WHERE price <= 0;
UPDATE books SET price = 671.00 WHERE id = 7;
ALTER TABLE books ADD CHECK (price > 0);
ALTER TABLE books ALTER COLUMN in_stock SET NOT NULL;
ALTER TABLE books ALTER COLUMN in_stock SET DEFAULT true;
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
SELECT id, title FROM books WHERE title IN ('Евгений Онегин', 'Ревизор');
INSERT INTO order_items (order_id, book_id, quantity) VALUES (1, 1, 2), (1, 8, 1);
SQL"
