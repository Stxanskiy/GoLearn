# Reference solutions for pg-start lab 3 (ch-pgs-lab3): the first tables.
# A temporary file of its own while the labs are written in parallel; lab.sh
# sources every solutions*.sh. Its place is the end of solutions-pg.sh.
# Keyed sol_<lesson slug with - as _>_<n>, n = position among the lesson's
# checked tasks. Each runs as root in the lab container, like the student.

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
