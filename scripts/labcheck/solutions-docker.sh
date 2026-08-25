# Reference solutions for the Docker courses (sourced by solutions.sh).
# Kept in a separate file so the main list stays readable.

sol_ch_dock_lab1_1='docker run --name lab1demo nginx:alpine -v >/dev/null 2>&1 || true'
sol_ch_dock_lab1_2='docker run -d --name mynginx nginx:alpine >/dev/null'
sol_ch_dock_lab1_3="docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' mynginx > /root/mynginx_ip.txt"
sol_ch_dock_lab1_4='sleep 4; docker stop mynginx >/dev/null && docker start mynginx >/dev/null'
sol_ch_dock_lab1_5='docker stop mynginx >/dev/null && docker rm mynginx >/dev/null'

sol_ch_dock_lab2_1='docker pull -q redis:alpine >/dev/null'
sol_ch_dock_lab2_2='docker tag redis:alpine myredis:v1'
sol_ch_dock_lab2_3='docker rmi myredis:v1 >/dev/null'

sol_ch_dock_lab3_1='cd /root/myapp && docker build -q -t myapp:v1 . >/dev/null'
sol_ch_dock_lab3_2='docker run -d --name myapp -p 5000:5000 myapp:v1 >/dev/null; sleep 2'
sol_ch_dock_lab3_3='curl -s http://localhost:5000/ >/dev/null'
sol_ch_dock_lab3_4='docker rm -f myapp >/dev/null; printf "LABEL version=\"2.0\"\nENV APP_VERSION=2.0\n" >> /root/myapp/Dockerfile; cd /root/myapp && docker build -q -t myapp:v2 . >/dev/null'
sol_ch_dock_lab3_5='docker run -d --name myapp2 -p 5001:5000 myapp:v2 >/dev/null; sleep 2'
sol_ch_dock_lab3_6='docker rm -f myapp2 >/dev/null'

sol_ch_dock_lab4_1='cd /root/goapp && docker build -q -f Dockerfile.multistage -t goapp:multi . >/dev/null'
sol_ch_dock_lab4_2='docker run -d --name goapp -p 8080:8080 goapp:multi >/dev/null; sleep 2'
sol_ch_dock_lab4_3='curl -s http://localhost:8080/ >/dev/null'
sol_ch_dock_lab4_4='cd /root/goapp && docker build -q --no-cache -f Dockerfile.multistage -t goapp:nocache . >/dev/null'
sol_ch_dock_lab4_5='docker rm -f goapp >/dev/null'
sol_ch_dock_lab4_6='true'

sol_ch_dock_lab5_1='docker network create mynet >/dev/null'
sol_ch_dock_lab5_2='docker run -d --name web1 --network mynet -p 80:80 -v /root/docker-lab5/web1:/usr/share/nginx/html nginx:alpine >/dev/null; sleep 2'
sol_ch_dock_lab5_3='docker run -d --name web2 --network mynet -v /root/docker-lab5/web2:/usr/share/nginx/html nginx:alpine >/dev/null; sleep 2'
sol_ch_dock_lab5_4="docker inspect -f '{{(index .NetworkSettings.Networks \"mynet\").IPAddress}}' web1 > /root/web1_ip.txt"
sol_ch_dock_lab5_5='docker rm -f web1 web2 outsider >/dev/null; docker network rm mynet >/dev/null'

sol_ch_dock_lab6_1='docker volume create mydata >/dev/null'
sol_ch_dock_lab6_2='docker run -d --name vol_test -v mydata:/data alpine:latest sleep 3600 >/dev/null'
sol_ch_dock_lab6_3='docker exec vol_test sh -c "echo persistent data > /data/test.txt"'
sol_ch_dock_lab6_4='docker rm -f vol_test >/dev/null; docker run --name vol_reader -v mydata:/data alpine:latest cat /data/test.txt >/dev/null'
sol_ch_dock_lab6_5='docker run -d --name nginx_bind -p 80:80 -v /root/html:/usr/share/nginx/html nginx:alpine >/dev/null; sleep 2'
sol_ch_dock_lab6_6="docker volume inspect -f '{{.Mountpoint}}' mydata > /root/mydata_mountpoint.txt"
sol_ch_dock_lab6_7='docker rm -f nginx_bind vol_reader >/dev/null; docker volume rm mydata >/dev/null'

sol_ch_dock_lab7_1='docker tag nginx:alpine registry.company.local/devops/mynginx:v1'
sol_ch_dock_lab7_2='docker images registry.company.local/devops/mynginx >/dev/null'
sol_ch_dock_lab7_3='docker pull -q redis:7-alpine >/dev/null'
sol_ch_dock_lab7_4='docker tag redis:7-alpine registry.company.local/devops/myredis:prod'
sol_ch_dock_lab7_5='docker rmi registry.company.local/devops/mynginx:v1 registry.company.local/devops/myredis:prod >/dev/null'

# Lab 8: итоговый стек — студент пишет Dockerfile'ы сам, эталон делает то же.
sol_ch_dock_lab8_1='mkdir -p /root/final-stack/api && printf "FROM python:3.12-alpine\nWORKDIR /app\nCOPY app.py .\nENV APP_VERSION=1.0\nENV WORKER_URL=http://worker:8080\nEXPOSE 5000\nCMD [\"python\", \"app.py\"]\n" > /root/final-stack/api/Dockerfile && printf "tmp\n*.log\n" > /root/final-stack/api/.dockerignore && cd /root/final-stack/api && docker build -q -t final-api:1.0 . >/dev/null'
sol_ch_dock_lab8_2='printf "FROM golang:1.21-alpine AS builder\nWORKDIR /src\nCOPY go.mod ./\nCOPY main.go ./\nENV CGO_ENABLED=0 GOFLAGS=-mod=mod GOPROXY=off\nRUN go build -o /out/worker .\n\nFROM alpine:latest\nWORKDIR /app\nCOPY --from=builder /out/worker /app/worker\nENV WORKER_VERSION=1.0\nEXPOSE 8080\nCMD [\"/app/worker\"]\n" > /root/final-stack/worker/Dockerfile && printf "tmp\n*.log\n" > /root/final-stack/worker/.dockerignore && cd /root/final-stack/worker && docker build -q -t final-worker:1.0 . >/dev/null'
sol_ch_dock_lab8_3='printf "FROM nginx:alpine\nCOPY index.html /usr/share/nginx/html/index.html\nCOPY nginx.conf /etc/nginx/conf.d/default.conf\nEXPOSE 80\n" > /root/final-stack/frontend/Dockerfile && cd /root/final-stack/frontend && docker build -q -t final-frontend:1.0 . >/dev/null'
sol_ch_dock_lab8_4='docker network create final-net >/dev/null'
sol_ch_dock_lab8_5='docker run -d --name final-worker --network final-net --network-alias worker final-worker:1.0 >/dev/null; sleep 2'
sol_ch_dock_lab8_6='docker run -d --name final-api --network final-net --network-alias api -e APP_VERSION=1.0 -e WORKER_URL=http://worker:8080 final-api:1.0 >/dev/null; sleep 2'
sol_ch_dock_lab8_7='docker run -d --name final-frontend --network final-net --network-alias frontend -p 80:80 final-frontend:1.0 >/dev/null; sleep 2'
sol_ch_dock_lab8_8='docker rm -f final-frontend final-api final-worker >/dev/null; docker network rm final-net >/dev/null'
