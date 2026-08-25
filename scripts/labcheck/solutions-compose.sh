# Reference solutions for the Docker Compose course (sourced by run.sh).

sol_ch_dkc_lab1_1='cd /root/app && docker compose up -d >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab1_2='cd /root/app && docker compose exec -T web cat /etc/nginx/nginx.conf > /root/nginx_conf.txt'
sol_ch_dkc_lab1_3='cd /root/app && printf "  cache:\n    image: redis:alpine\n    volumes:\n      - cache_data:/data\n\nvolumes:\n  cache_data:\n" >> docker-compose.yml && docker compose up -d >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab1_4='cd /root/app && docker compose exec -T cache redis-cli ping > /root/redis_ping.txt'
sol_ch_dkc_lab1_5='cd /root/app && docker compose stop cache >/dev/null 2>&1'
sol_ch_dkc_lab1_6='cd /root/app && docker compose down >/dev/null 2>&1'

sol_ch_dkc_lab2_1='cd /root/webapp && docker compose up -d >/dev/null 2>&1; sleep 4'
sol_ch_dkc_lab2_2='cd /root/webapp && docker compose exec -T web getent hosts app | awk "{print \$1}" > /root/app_dns_ip.txt'
sol_ch_dkc_lab2_3='curl -s http://localhost/ > /root/app_response.txt'
sol_ch_dkc_lab2_4='docker network ls --format "{{.Name}}" | grep webapp > /root/compose_network.txt'
sol_ch_dkc_lab2_5='cd /root/webapp && docker compose up -d --scale app=2 >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab2_6='true'
sol_ch_dkc_lab2_7='cd /root/webapp && docker compose up -d --scale app=1 >/dev/null 2>&1 && docker compose restart web >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab2_8='cd /root/webapp && docker compose down >/dev/null 2>&1'

sol_ch_dkc_lab3_1='cd /root/dbapp && docker compose up -d >/dev/null 2>&1; for i in $(seq 1 30); do docker compose exec -T db pg_isready -U postgres >/dev/null 2>&1 && break; sleep 1; done'
sol_ch_dkc_lab3_2='cd /root/dbapp && docker compose exec -T db psql -U postgres -c "CREATE TABLE users (id SERIAL PRIMARY KEY, name VARCHAR(100)); INSERT INTO users (name) VALUES (\$\$Alice\$\$), (\$\$Bob\$\$);" >/dev/null 2>&1'
sol_ch_dkc_lab3_3='cd /root/dbapp && docker compose exec -T db psql -U postgres -tAc "select count(*) from users" > /root/users_count_before.txt'
sol_ch_dkc_lab3_4='cd /root/dbapp && docker compose down >/dev/null 2>&1 && docker compose up -d >/dev/null 2>&1; for i in $(seq 1 30); do docker compose exec -T db pg_isready -U postgres >/dev/null 2>&1 && break; sleep 1; done; docker compose exec -T db psql -U postgres -tAc "select count(*) from users" > /root/users_count_after.txt'
sol_ch_dkc_lab3_5='docker volume ls --format "{{.Name}}" | grep dbapp > /root/db_volume_name.txt'
sol_ch_dkc_lab3_6='mkdir -p /root/backups && cd /root/dbapp && docker compose exec -T db pg_dump -U postgres postgres > /root/backups/dump.sql'
sol_ch_dkc_lab3_7='docker volume inspect -f "{{.Mountpoint}}" dbapp_db_data > /root/volume_mountpoint.txt'
sol_ch_dkc_lab3_8='cd /root/dbapp && docker compose down -v >/dev/null 2>&1'

sol_ch_dkc_lab4_1='cp /root/envapp/.env.example /root/envapp/.env'
sol_ch_dkc_lab4_2='cd /root/envapp && docker compose up -d >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab4_3='cd /root/envapp && APP_DEBUG=false docker compose up -d --force-recreate app >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab4_4='cd /root/envapp && printf "  monitor:\n    image: nginx:alpine\n    ports:\n      - \"\${MONITOR_PORT:-9090}:80\"\n" >> docker-compose.yml && docker compose up -d >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab4_5='printf "APP_DEBUG=false\nAPP_PORT=80\n" > /root/envapp/.env.prod'
sol_ch_dkc_lab4_6='cd /root/envapp && docker compose down >/dev/null 2>&1'

sol_ch_dkc_lab5_1='cd /root/override && docker compose -f docker-compose.yml up -d >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab5_2='cd /root/override && docker compose -f docker-compose.yml down >/dev/null 2>&1'
sol_ch_dkc_lab5_3='cd /root/override && docker compose -f docker-compose.yml -f docker-compose.dev.yml up -d >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab5_4='printf "services:\n  app:\n    environment:\n      AUTO_OVERRIDE: \"true\"\n" > /root/override/docker-compose.override.yml && cd /root/override && docker compose up -d --force-recreate >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab5_5='cd /root/override && printf "  debugger:\n    image: alpine:latest\n    command: sh -c \"while true; do sleep 5; done\"\n    profiles: [debug]\n" >> docker-compose.yml && docker compose --profile debug up -d >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab5_6='cd /root/override && docker compose --profile debug down >/dev/null 2>&1'

sol_ch_dkc_lab6_1='python3 - <<PYEOF
import re
p="/root/healthapp/docker-compose.yml"
s=open(p).read()
s=s.replace("""  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: postgres""","""  db:
    image: postgres:16-alpine
    environment:
      POSTGRES_PASSWORD: postgres
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 3s
      retries: 10""")
s=s.replace("""    depends_on:
      - db""","""    depends_on:
      db:
        condition: service_healthy""")
open(p,"w").write(s)
PYEOF'
sol_ch_dkc_lab6_2='cd /root/healthapp && docker compose up -d >/dev/null 2>&1; sleep 5'
sol_ch_dkc_lab6_3='for i in $(seq 1 30); do [ "$(docker inspect -f "{{.State.Health.Status}}" healthapp-db-1 2>/dev/null)" = healthy ] && break; sleep 1; done'
sol_ch_dkc_lab6_4='python3 - <<PYEOF
p="/root/healthapp/docker-compose.yml"
s=open(p).read()
s=s.replace("""  web:
    image: nginx:alpine""","""  web:
    image: nginx:alpine
    healthcheck:
      test: ["CMD-SHELL", "wget -qO- http://localhost/ || exit 1"]
      interval: 5s
      timeout: 3s
      retries: 10""")
open(p,"w").write(s)
PYEOF'
sol_ch_dkc_lab6_5='cd /root/healthapp && docker compose up -d >/dev/null 2>&1; for i in $(seq 1 40); do [ "$(docker inspect -f "{{.State.Health.Status}}" healthapp-web-1 2>/dev/null)" = healthy ] && [ "$(docker inspect -f "{{.State.Health.Status}}" healthapp-db-1 2>/dev/null)" = healthy ] && break; sleep 1; done'
sol_ch_dkc_lab6_6='cd /root/healthapp && docker compose down >/dev/null 2>&1'

sol_ch_dkc_lab7_1='cd /root/scaleapp && docker compose up -d >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab7_2='cd /root/scaleapp && docker compose up -d --scale worker=3 >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab7_3='true'
sol_ch_dkc_lab7_4='cd /root/scaleapp && printf "    deploy:\n      replicas: 2\n" >> docker-compose.yml && docker compose up -d >/dev/null 2>&1; sleep 3'
sol_ch_dkc_lab7_5='cd /root/scaleapp && docker compose up -d --scale worker=1 >/dev/null 2>&1; sleep 2'
sol_ch_dkc_lab7_6='cd /root/scaleapp && docker compose down >/dev/null 2>&1'
