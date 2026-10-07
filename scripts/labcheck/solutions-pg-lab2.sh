# Reference solutions for pg-start, Лабораторная 2: осваиваем psql (ch-pgs-lab2).
# Keyed sol_ch_pgs_lab2_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way a student would type it.

# 1. Имена баз без рамок: tuples only (-t), unaligned (-A).
sol_ch_pgs_lab2_1='psql -At -c "SELECT datname FROM pg_database" > /root/dbs.txt'

# 2. Готовый скрипт — в базе pereplet.
sol_ch_pgs_lab2_2='psql -d pereplet -f /root/sql/hello.sql'

# 3. Одно число.
sol_ch_pgs_lab2_3='psql -d pereplet -At -c "SELECT count(*) FROM authors" > /root/authors_count.txt'

# 4. Изнутри psql: \t \a, вывод в файл через \o.
sol_ch_pgs_lab2_4=$(cat <<'SOL'
psql -d pereplet <<'EOF'
\t
\a
\o /root/titles.txt
SELECT title FROM books ORDER BY title;
\o
EOF
SOL
)

# 5. Синтаксис из \h CREATE DATABASE.
sol_ch_pgs_lab2_5='psql -c "CREATE DATABASE sandbox CONNECTION LIMIT 5;"'

# 6. ~/.psqlrc: замер времени и NULL как ∅.
sol_ch_pgs_lab2_6=$(cat <<'SOL'
cat > ~/.psqlrc <<'EOF'
\timing on
\pset null '∅'
EOF
SOL
)
