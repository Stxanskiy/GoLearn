# Reference solutions for pg-start, Лабораторная 4: наполняем магазин (ch-pgs-lab4).
# Keyed sol_ch_pgs_lab4_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.

# 1. Одна строка — одной командой из терминала.
sol_ch_pgs_lab4_1=$(cat <<'SOL'
psql -d pereplet -c "INSERT INTO authors (full_name, birth_year, country) VALUES ('Александр Пушкин', 1799, 'Россия');"
SOL
)

# 2. Три строки одним INSERT — в psql, команда в несколько строк.
sol_ch_pgs_lab4_2=$(cat <<'SOL'
psql -d pereplet <<'EOF'
INSERT INTO authors (full_name, birth_year, country) VALUES
    ('Лев Толстой',  1828, 'Россия'),
    ('Агата Кристи', 1890, 'Великобритания'),
    ('Рэй Брэдбери', 1920, 'США');
SELECT * FROM authors ORDER BY id;
EOF
SOL
)

# 3. RETURNING печатает номер; -At — без рамок, -q — без строки «INSERT 0 1».
sol_ch_pgs_lab4_3=$(cat <<'SOL'
psql -d pereplet -Atq -c "INSERT INTO books (title, price, page_count, published_year, in_stock) VALUES ('Евгений Онегин', 450.00, 320, 1833, true) RETURNING id" > /root/book_id.txt
cat /root/book_id.txt
SOL
)

# 4. Сначала SELECT с тем же WHERE, потом UPDATE формулой.
sol_ch_pgs_lab4_4=$(cat <<'SOL'
psql -d pereplet <<'EOF'
SELECT id, title, price FROM books WHERE title = 'Евгений Онегин';
UPDATE books SET price = price * 1.1 WHERE title = 'Евгений Онегин';
SELECT id, title, price FROM books WHERE title = 'Евгений Онегин';
EOF
SOL
)

# 5. Два столбца в одном SET.
sol_ch_pgs_lab4_5=$(cat <<'SOL'
psql -d pereplet -c "UPDATE books SET page_count = 128, published_year = 1836 WHERE title = 'Ревизор';"
SOL
)

# 6. Условие page_count < 100 само не берёт строку с NULL.
sol_ch_pgs_lab4_6=$(cat <<'SOL'
psql -d pereplet <<'EOF'
SELECT id, title, page_count, in_stock FROM books WHERE page_count < 100;
UPDATE books SET in_stock = false WHERE page_count < 100;
EOF
SOL
)

# 7. Посмотреть номера, отрепетировать удаление в транзакции и подтвердить.
sol_ch_pgs_lab4_7=$(cat <<'SOL'
psql -d pereplet <<'EOF'
SELECT id, title, isbn FROM books WHERE title = 'Мастер и Маргарита' ORDER BY id;
BEGIN;
DELETE FROM books WHERE title = 'Мастер и Маргарита' AND id <> 2;
SELECT id, title FROM books WHERE title = 'Мастер и Маргарита';
COMMIT;
EOF
SOL
)

# 8. Скрипт в файле: перенос одним INSERT … SELECT и запись в журнал.
sol_ch_pgs_lab4_8=$(cat <<'SOL'
cat > /root/sql/new_books.sql <<'EOF'
-- Новые поступления: из staging.new_books в каталог, все — в наличии.
INSERT INTO books (title, price, page_count, published_year, isbn, in_stock)
SELECT title, price, page_count, published_year, isbn, true
FROM staging.new_books;

-- Отметка о загрузке в журнале.
INSERT INTO staging.import_log (loaded_at, row_count)
VALUES (now(), 3);
EOF
psql -d pereplet -v ON_ERROR_STOP=1 -f /root/sql/new_books.sql
SOL
)
