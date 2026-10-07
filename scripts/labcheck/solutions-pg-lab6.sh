# Reference solutions for pg-start, Лабораторная 6: импорт каталога (ch-pgs-lab6).
# Keyed sol_ch_pgs_lab6_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.

# 1. Посмотреть на файл и загрузить его одной командой \copy из терминала.
sol_ch_pgs_lab6_1=$(cat <<'SOL'
head -n 3 /root/import/authors.csv
psql -d pereplet -c "\copy authors (full_name, birth_year, country) FROM '/root/import/authors.csv' WITH (FORMAT csv, HEADER)"
psql -d pereplet -c "SELECT count(*) FROM authors;"
SOL
)

# 2. Другой разделитель и свой порядок столбцов — в psql, как в интерактивной сессии.
sol_ch_pgs_lab6_2=$(cat <<'SOL'
head -n 3 /root/import/books.csv
psql -d pereplet <<'EOF'
\copy books (isbn, title, price, page_count, published_year, in_stock) FROM '/root/import/books.csv' WITH (FORMAT csv, HEADER, DELIMITER ';')
SELECT count(*) FROM books;
SELECT title, page_count FROM books WHERE page_count IS NULL;
EOF
SOL
)

# 3. Первая попытка падает и называет строку; исправить её в файле и загрузить заново.
sol_ch_pgs_lab6_3=$(cat <<'SOL'
psql -d pereplet -c "\copy customers (email, full_name, city, registered_at) FROM '/root/import/customers.csv' WITH (FORMAT csv, HEADER)"
sed -n '9p' /root/import/customers.csv
sed -i '9s/31\.12\.2023/2023-12-31/' /root/import/customers.csv
sed -n '9p' /root/import/customers.csv
psql -d pereplet -c "\copy customers (email, full_name, city, registered_at) FROM '/root/import/customers.csv' WITH (FORMAT csv, HEADER)"
SOL
)

# 4. Выгрузка запроса в файл.
sol_ch_pgs_lab6_4=$(cat <<'SOL'
psql -d pereplet -c "\copy (SELECT title, price FROM books WHERE NOT in_stock ORDER BY title) TO '/root/export/out_of_stock.csv' WITH (FORMAT csv, HEADER)"
cat /root/export/out_of_stock.csv
SOL
)

# 5. Номера из файла, неудачный INSERT, сдвиг счётчика и INSERT ещё раз.
sol_ch_pgs_lab6_5=$(cat <<'SOL'
psql -d pereplet <<'EOF'
\copy genres (id, name) FROM '/root/import/genres.csv' WITH (FORMAT csv, HEADER)
INSERT INTO genres (name) VALUES ('Приключения');
SELECT setval(pg_get_serial_sequence('genres', 'id'), (SELECT max(id) FROM genres));
INSERT INTO genres (name) VALUES ('Приключения') RETURNING id;
EOF
SOL
)

# 6. Сначала посмотреть на три строки, потом вставить тысячу одной командой.
sol_ch_pgs_lab6_6=$(cat <<'SOL'
psql -d pereplet <<'EOF'
SELECT 'test' || g || '@example.com', 'Тестовый покупатель ' || g FROM generate_series(1, 3) AS g;
INSERT INTO customers (email, full_name)
SELECT 'test' || g || '@example.com', 'Тестовый покупатель ' || g
FROM generate_series(1, 1000) AS g;
EOF
SOL
)
