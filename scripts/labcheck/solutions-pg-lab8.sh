# Lab 8 (ch-pgs-lab8): the project, a database for a helpdesk.
# A section of solutions-pg.sh written on its own while the labs are written
# in parallel; lab.sh sources every solutions*.sh. The schema goes into
# /root/helpdesk.sql and runs with \i, the way lesson 18 teaches: each step
# rewrites the file with one table more, as a student adds to it in nano, and
# running it again rebuilds the tables from scratch. The rows of the last task
# are typed in psql after a look at the ids (fresh tables: 1, 2 and 1).
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
