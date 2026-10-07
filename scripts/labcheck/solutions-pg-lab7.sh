# Reference solutions for pg-start, Проект: база для службы поддержки (ch-pgs-lab7).
# Keyed sol_ch_pgs_lab7_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.
#
# The student builds the schema in /root/project/schema.sql step by step: the
# columns first, then keys, then rules — each time the file is rewritten (in
# nano, here with a heredoc), the empty tables are dropped and the file is run
# again. The plain CREATE TABLE of the first version fails on a second run;
# making the script repeatable is the ★ task.

# 1. База и таблицы — только столбцы и типы.
sol_ch_pgs_lab7_1=$(cat <<'SOL'
createdb helpdesk
cat > /root/project/schema.sql <<'EOF'
-- Схема базы службы поддержки «Переплёта».
-- Запуск: psql -d helpdesk -f /root/project/schema.sql
CREATE TABLE users (
    id         bigint,
    email      text,
    full_name  text,
    created_at timestamptz
);

CREATE TABLE tickets (
    id         bigint,
    author_id  bigint,
    title      text,
    priority   text,
    status     text,
    created_at timestamptz
);

CREATE TABLE comments (
    id         bigint,
    ticket_id  bigint,
    author_id  bigint,
    body       text,
    created_at timestamptz
);
EOF
psql -d helpdesk -f /root/project/schema.sql
SOL
)

# 2. Ключи: identity, первичные и внешние. Таблицы пока пустые — проще удалить
# их и запустить исправленный файл заново. Удаляют в обратном порядке: сначала
# те, что ссылаются.
sol_ch_pgs_lab7_2=$(cat <<'SOL'
cat > /root/project/schema.sql <<'EOF'
-- Схема базы службы поддержки «Переплёта».
-- Запуск: psql -d helpdesk -f /root/project/schema.sql
-- Порядок таблиц: сначала те, на кого ссылаются.
CREATE TABLE users (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      text,
    full_name  text,
    created_at timestamptz
);

CREATE TABLE tickets (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id  bigint REFERENCES users (id),
    title      text,
    priority   text,
    status     text,
    created_at timestamptz
);

CREATE TABLE comments (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ticket_id  bigint REFERENCES tickets (id),
    author_id  bigint REFERENCES users (id),
    body       text,
    created_at timestamptz
);
EOF
psql -d helpdesk -c "DROP TABLE comments, tickets, users;"
psql -d helpdesk -f /root/project/schema.sql
SOL
)

# 3. Правила: NOT NULL, UNIQUE, CHECK, DEFAULT и каскад для переписки.
sol_ch_pgs_lab7_3=$(cat <<'SOL'
cat > /root/project/schema.sql <<'EOF'
-- Схема базы службы поддержки «Переплёта».
-- Запуск: psql -d helpdesk -f /root/project/schema.sql
-- Порядок таблиц: сначала те, на кого ссылаются.
CREATE TABLE users (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      text NOT NULL UNIQUE,
    full_name  text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tickets (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    author_id  bigint NOT NULL REFERENCES users (id),
    title      text NOT NULL,
    priority   text NOT NULL DEFAULT 'normal'
               CHECK (priority IN ('low', 'normal', 'high', 'critical')),
    status     text NOT NULL DEFAULT 'open'
               CHECK (status IN ('open', 'in_progress', 'closed')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE comments (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    ticket_id  bigint NOT NULL REFERENCES tickets (id) ON DELETE CASCADE,
    author_id  bigint NOT NULL REFERENCES users (id),
    body       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
EOF
psql -d helpdesk -c "DROP TABLE comments, tickets, users;"
psql -d helpdesk -f /root/project/schema.sql
SOL
)

# 4. Данные — в порядке ссылок: пользователи, обращения, переписка. id — из файлов.
sol_ch_pgs_lab7_4=$(cat <<'SOL'
cd /root/project
psql -d helpdesk <<'EOF'
\copy users (id, email, full_name) FROM 'users.csv' WITH (FORMAT csv, HEADER)
\copy tickets (id, author_id, title, priority, status) FROM 'tickets.csv' WITH (FORMAT csv, HEADER)
\copy comments (id, ticket_id, author_id, body) FROM 'comments.csv' WITH (FORMAT csv, HEADER)
SELECT count(*) FROM users;
SELECT count(*) FROM tickets;
SELECT count(*) FROM comments;
EOF
SOL
)

# 5. Счётчики — за наибольший загруженный id, потом новые строки без id,
# как их добавит сайт. Номер нового пользователя берём из RETURNING.
sol_ch_pgs_lab7_5=$(cat <<'SOL'
psql -d helpdesk <<'EOF'
SELECT setval(pg_get_serial_sequence('users', 'id'), (SELECT max(id) FROM users));
SELECT setval(pg_get_serial_sequence('tickets', 'id'), (SELECT max(id) FROM tickets));
SELECT setval(pg_get_serial_sequence('comments', 'id'), (SELECT max(id) FROM comments));
INSERT INTO users (email, full_name) VALUES ('t.zakharov@example.com', 'Тимур Захаров') RETURNING id;
EOF
uid=$(psql -d helpdesk -At -c "SELECT id FROM users WHERE email = 't.zakharov@example.com'")
psql -d helpdesk -c "INSERT INTO tickets (author_id, title, priority) VALUES ($uid, 'Не приходит письмо с подтверждением регистрации', 'high') RETURNING id, status, created_at;"
SOL
)

# 6. ★ Повторяемый скрипт: IF NOT EXISTS у каждой таблицы. Проверка себя — на
# отдельной базе, два запуска подряд, базу helpdesk не трогаем.
sol_ch_pgs_lab7_6=$(cat <<'SOL'
sed -i 's/^CREATE TABLE /CREATE TABLE IF NOT EXISTS /' /root/project/schema.sql
createdb schema_test
psql -v ON_ERROR_STOP=1 -d schema_test -f /root/project/schema.sql
psql -v ON_ERROR_STOP=1 -d schema_test -f /root/project/schema.sql
dropdb schema_test
psql -d helpdesk -f /root/project/schema.sql
SOL
)
