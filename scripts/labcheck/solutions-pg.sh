# Reference solutions for the PostgreSQL course labs (pg-start).
# Keyed sol_<lesson slug with - as _>_<n>, n = position among the lesson's
# checked tasks. Each runs as root in the lab container, like the student.

# Lab 1 (ch-pgs-lab1): installing PostgreSQL.
sol_ch_pgs_lab1_1='apt update && apt install -y postgresql'
sol_ch_pgs_lab1_2='sudo -u postgres createuser --superuser root'
sol_ch_pgs_lab1_3='printf "CREATE DATABASE pereplet;\n" | psql postgres'
