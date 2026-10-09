# Reference solutions for pg-start lab 2, kept apart from solutions-pg.sh while
# the labs are written in parallel; lab.sh sources every solutions*.sh. The
# section below moves to the end of solutions-pg.sh as it is.

# Lab 2 (ch-pgs-lab2): working in psql.
sol_ch_pgs_lab2_1='printf "CREATE DATABASE sandbox;\n" | psql pereplet'
sol_ch_pgs_lab2_2='printf "DROP DATABASE old_shop;\n" | psql pereplet'
sol_ch_pgs_lab2_3='printf "CREATE DATABASE archive_2025 OWNER postgres;\n" | psql pereplet'
