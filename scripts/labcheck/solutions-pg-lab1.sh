# Reference solutions for pg-start lab 1 (ch-pgs-lab1): installing PostgreSQL.
# Keyed sol_ch_pgs_lab1_<n>, n = position among the lesson's checked tasks.
# Each runs as root in the lab container, the way the student types it.
sol_ch_pgs_lab1_1='apt update && apt install -y postgresql'
sol_ch_pgs_lab1_2='psql --version > /root/pg_version.txt'
sol_ch_pgs_lab1_3='pg_ctlcluster 16 main start && pg_lsclusters > /root/cluster.txt'
sol_ch_pgs_lab1_4='sudo -u postgres psql -c "SELECT current_user, version();" > /root/whoami.txt'
sol_ch_pgs_lab1_5='sudo -u postgres createuser --createdb root && sudo -u postgres createdb -O root root'
sol_ch_pgs_lab1_6='createdb pereplet'
sol_ch_pgs_lab1_7='grep "ready to accept connections" /var/log/postgresql/postgresql-16-main.log > /root/ready.txt'
