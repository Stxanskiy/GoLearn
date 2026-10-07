# Reference solutions for pg-start, Лабораторная 3: создаём базу магазина (ch-pgs-lab3).
# Keyed sol_ch_pgs_lab3_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.

# 1. База — утилитой из терминала, как в первой лабораторной.
sol_ch_pgs_lab3_1='createdb pereplet'

# 2. authors — в psql, команда в несколько строк.
sol_ch_pgs_lab3_2=$(cat <<'SOL'
psql -d pereplet <<'EOF'
CREATE TABLE authors (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name  text,
    birth_year integer,
    country    text
);
\d authors
EOF
SOL
)

# 3. books — через файл и psql -f, как советует памятка лабы.
sol_ch_pgs_lab3_3=$(cat <<'SOL'
cat > /root/sql/books.sql <<'EOF'
CREATE TABLE books (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title          text,
    price          numeric(8,2),
    pages          integer,
    published_year integer,
    in_stock       boolean
);
EOF
psql -d pereplet -f /root/sql/books.sql
SOL
)

# 4. Новый столбец в готовой таблице.
sol_ch_pgs_lab3_4='psql -d pereplet -c "ALTER TABLE books ADD COLUMN isbn text;"'

# 5. Переименование — данные остаются на месте.
sol_ch_pgs_lab3_5='psql -d pereplet -c "ALTER TABLE books RENAME COLUMN pages TO page_count;"'

# 6. Сначала схема, потом таблица в ней.
sol_ch_pgs_lab3_6=$(cat <<'SOL'
psql -d pereplet <<'EOF'
CREATE SCHEMA staging;
CREATE TABLE staging.import_log (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    loaded_at timestamptz,
    row_count integer
);
EOF
SOL
)

# 7. Имя создано в кавычках с заглавной буквы — и удалять его нужно так же.
sol_ch_pgs_lab3_7=$(cat <<'SOL'
psql <<'EOF'
DROP TABLE "Books";
EOF
SOL
)

# 8. Итоговый вид таблиц, IF NOT EXISTS везде; два запуска подряд — проверка себя.
sol_ch_pgs_lab3_8=$(cat <<'SOL'
cat > /root/sql/schema.sql <<'EOF'
-- Схема базы «Переплёта», первая версия.
-- Базу выбирают при запуске: psql -d pereplet -f /root/sql/schema.sql
CREATE TABLE IF NOT EXISTS authors (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    full_name  text,
    birth_year integer,
    country    text
);

CREATE TABLE IF NOT EXISTS books (
    id             bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title          text,
    price          numeric(8,2),
    page_count     integer,
    published_year integer,
    in_stock       boolean,
    isbn           text
);

CREATE SCHEMA IF NOT EXISTS staging;

CREATE TABLE IF NOT EXISTS staging.import_log (
    id        bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    loaded_at timestamptz,
    row_count integer
);
EOF
createdb schema_test
psql -v ON_ERROR_STOP=1 -d schema_test -f /root/sql/schema.sql
psql -v ON_ERROR_STOP=1 -d schema_test -f /root/sql/schema.sql
dropdb schema_test
SOL
)
