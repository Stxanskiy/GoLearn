# Lab 7 (ch-pgs-lab7): importing the catalog.
# A section of solutions-pg.sh written on its own while the labs are written
# in parallel; lab.sh sources every solutions*.sh. Double-quoted where the
# SQL has single quotes: \\ stands for the one backslash of \copy.
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
# The student opens the file in nano, goes to line 11 from CONTEXT and cuts it.
sol_ch_pgs_lab7_4="sed -i 11d /root/book_genres.csv && psql pereplet <<'SQL'
\\copy book_genres (book_id, genre_id) FROM '/root/book_genres.csv' WITH (FORMAT csv, HEADER)
SQL"
sol_ch_pgs_lab7_5="psql pereplet <<'SQL'
\\copy (SELECT title, price FROM books WHERE in_stock = true ORDER BY price) TO '/root/in_stock.csv' WITH (FORMAT csv, HEADER, DELIMITER ';')
SQL"
