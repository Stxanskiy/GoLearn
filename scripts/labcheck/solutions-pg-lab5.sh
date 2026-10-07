# Reference solutions for pg-start, Лабораторная 5: связываем таблицы (ch-pgs-lab5).
# Keyed sol_ch_pgs_lab5_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.

# 1. Новый столбец сразу со ссылкой — одной командой из терминала.
sol_ch_pgs_lab5_1=$(cat <<'SOL'
psql -d pereplet -c "ALTER TABLE books ADD COLUMN author_id bigint REFERENCES authors (id);"
SOL
)

# 2. Посмотреть номера авторов и расставить их книгам: по команде на автора.
sol_ch_pgs_lab5_2=$(cat <<'SOL'
psql -d pereplet <<'EOF'
SELECT id, full_name FROM authors ORDER BY id;
UPDATE books SET author_id = 1 WHERE title IN ('Преступление и наказание', 'Идиот');
UPDATE books SET author_id = 2 WHERE title = 'Мастер и Маргарита';
UPDATE books SET author_id = 3 WHERE title IN ('Вишнёвый сад', 'Каштанка');
UPDATE books SET author_id = 4 WHERE title IN ('Евгений Онегин', 'Капитанская дочка');
UPDATE books SET author_id = 5 WHERE title IN ('Анна Каренина', 'Война и мир');
UPDATE books SET author_id = 6 WHERE title = 'Убийство в Восточном экспрессе';
SELECT title FROM books WHERE author_id IS NULL;
EOF
SOL
)

# 3. Сначала исправить данные, потом добавить правило — в обратном порядке
#    ALTER TABLE откажет: «Война и мир» с ценой 0 нарушает его.
sol_ch_pgs_lab5_3=$(cat <<'SOL'
psql -d pereplet <<'EOF'
UPDATE books SET price = 890.00 WHERE title = 'Война и мир';
ALTER TABLE books ADD CONSTRAINT books_price_positive CHECK (price > 0);
EOF
SOL
)

# 4. Схема таблицы — в файле, чтобы её было легко поправить и запустить снова.
sol_ch_pgs_lab5_4=$(cat <<'SOL'
cat > /root/sql/customers.sql <<'EOF'
CREATE TABLE customers (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email         text NOT NULL UNIQUE,
    full_name     text NOT NULL,
    city          text,
    registered_at timestamptz NOT NULL DEFAULT now()
);
EOF
psql -d pereplet -v ON_ERROR_STOP=1 -f /root/sql/customers.sql
SOL
)

# 5. Заказы: ссылка на покупателя, статус из четырёх значений.
sol_ch_pgs_lab5_5=$(cat <<'SOL'
cat > /root/sql/orders.sql <<'EOF'
CREATE TABLE orders (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id bigint NOT NULL REFERENCES customers (id),
    created_at  timestamptz NOT NULL DEFAULT now(),
    status      text NOT NULL DEFAULT 'new'
                CHECK (status IN ('new', 'paid', 'shipped', 'cancelled'))
);
EOF
psql -d pereplet -v ON_ERROR_STOP=1 -f /root/sql/orders.sql
SOL
)

# 6. Позиции заказа: составной первичный ключ и каскад только от заказа.
sol_ch_pgs_lab5_6=$(cat <<'SOL'
psql -d pereplet -v ON_ERROR_STOP=1 <<'EOF'
CREATE TABLE order_items (
    order_id bigint REFERENCES orders (id) ON DELETE CASCADE,
    book_id  bigint REFERENCES books (id),
    quantity integer NOT NULL CHECK (quantity > 0),
    price    numeric(8,2) NOT NULL,
    PRIMARY KEY (order_id, book_id)
);
\d order_items
EOF
SOL
)

# 7. Первый заказ: покупатель, заказ, позиции — номера берём запросами,
#    цену копируем из books.
sol_ch_pgs_lab5_7=$(cat <<'SOL'
psql -d pereplet -v ON_ERROR_STOP=1 <<'EOF'
INSERT INTO customers (email, full_name, city)
VALUES ('irina.sokolova@example.com', 'Ирина Соколова', 'Казань')
RETURNING id;

INSERT INTO orders (customer_id)
SELECT id FROM customers WHERE email = 'irina.sokolova@example.com'
RETURNING id, status;

INSERT INTO order_items (order_id, book_id, quantity, price)
SELECT o.id, b.id, 1, b.price
FROM orders o, books b
WHERE o.customer_id = (SELECT id FROM customers WHERE email = 'irina.sokolova@example.com')
  AND b.title = 'Мастер и Маргарита';

INSERT INTO order_items (order_id, book_id, quantity, price)
SELECT o.id, b.id, 2, b.price
FROM orders o, books b
WHERE o.customer_id = (SELECT id FROM customers WHERE email = 'irina.sokolova@example.com')
  AND b.title = 'Капитанская дочка';

SELECT * FROM order_items;
EOF
SOL
)

# 8. Жанры и связующая таблица, потом три связи.
sol_ch_pgs_lab5_8=$(cat <<'SOL'
psql -d pereplet -v ON_ERROR_STOP=1 <<'EOF'
CREATE TABLE genres (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE book_genres (
    book_id  bigint REFERENCES books (id) ON DELETE CASCADE,
    genre_id bigint REFERENCES genres (id),
    PRIMARY KEY (book_id, genre_id)
);

INSERT INTO genres (name) VALUES ('Роман'), ('Фантастика');

INSERT INTO book_genres (book_id, genre_id)
SELECT b.id, g.id FROM books b, genres g
WHERE b.title = 'Мастер и Маргарита' AND g.name IN ('Роман', 'Фантастика');

INSERT INTO book_genres (book_id, genre_id)
SELECT b.id, g.id FROM books b, genres g
WHERE b.title = 'Анна Каренина' AND g.name = 'Роман';
EOF
SOL
)
