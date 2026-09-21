# Reference solutions: lesson slug | task index | command(s)
sol_ch_lnav_lab1_1='pwd > /root/location.txt'
sol_ch_lnav_lab1_2='ls -la /opt/devops/lab1 > /root/lab1_listing.txt'
sol_ch_lnav_lab1_3='cd /opt/devops/lab1/notes && pwd > /root/notes_path.txt'
sol_ch_lnav_lab1_4='cat /opt/devops/lab1/.secret > /root/found_secret.txt'

sol_ch_lnav_lab2_1='mkdir /root/myproject'
sol_ch_lnav_lab2_2='echo "Hello DevOps" > /root/myproject/README.md'
sol_ch_lnav_lab2_3='cp /opt/devops/lab1/config.txt /root/myproject/'
sol_ch_lnav_lab2_4='mv /root/myproject/config.txt /root/myproject/myconfig.txt'
sol_ch_lnav_lab2_5='mkdir -p /root/backup/configs && cp /root/myproject/myconfig.txt /root/backup/configs/'
sol_ch_lnav_lab2_6='rm /root/myproject/README.md'
sol_ch_lnav_lab2_7='rm -r /root/myproject'

sol_ch_lnav_lab3_1='ln -s /opt/devops/lab1/config.txt /root/config_link'
sol_ch_lnav_lab3_2='cat /root/config_link > /root/link_content.txt'
sol_ch_lnav_lab3_3='ln -s /opt/devops/lab1/nonexistent.txt /root/broken_link'
sol_ch_lnav_lab3_4='find /root -type l > /root/all_links.txt'

sol_ch_lnav_lab4_1="find /opt/devops/lab4 -name '*.log' > /root/logs.txt"
sol_ch_lnav_lab4_2='find /opt/devops/lab4 -type f > /root/files_only.txt'
sol_ch_lnav_lab4_3='find /opt/devops/lab4 -type d > /root/dirs_only.txt'
sol_ch_lnav_lab4_4='find /opt/devops/lab4 -type f -size +1k > /root/large_files.txt'
sol_ch_lnav_lab4_5="find /opt/devops/lab4 -name '*.txt' -exec wc -l {} + > /root/line_counts.txt"
sol_ch_lnav_lab4_6="find /opt/devops/lab4 -name '.*' -type f > /root/hidden_files.txt"

sol_ch_lnav_lab5_1='head -5 /opt/devops/lab5/server.log > /root/first5.txt'
sol_ch_lnav_lab5_2='tail -10 /opt/devops/lab5/server.log > /root/last10.txt'
sol_ch_lnav_lab5_3='cat /opt/devops/lab5/part1.txt /opt/devops/lab5/part2.txt > /root/combined.txt'
sol_ch_lnav_lab5_4='grep ERROR /opt/devops/lab5/server.log > /root/errors.txt'

sol_ch_lnav_lab6_1='grep 404 /opt/devops/lab6/access.log > /root/not_found.txt'
sol_ch_lnav_lab6_2='sort /opt/devops/lab6/ips.txt > /root/sorted_ips.txt'
sol_ch_lnav_lab6_3='sort -u /opt/devops/lab6/ips.txt > /root/unique_ips.txt'
sol_ch_lnav_lab6_4="awk '{print \$1}' /opt/devops/lab6/access.log | sort | uniq -c | sort -rn > /root/ip_stats.txt"

sol_ch_lnav_lab7_1='ls /opt/devops/lab7/*.log > /root/log_list.txt'
sol_ch_lnav_lab7_2='mkdir /root/configs && cp /opt/devops/lab7/*.conf /root/configs/'
sol_ch_lnav_lab7_3='ls /opt/devops/lab7/data?.txt > /root/data_files.txt'
sol_ch_lnav_lab7_4="find /opt/devops/lab7 -name '*.txt' > /root/txt_files.txt"
sol_ch_lnav_lab7_5='mkdir /root/old_configs && mv /root/configs/*.conf /root/old_configs/'

sol_ch_lnav_lab8_1='tar -czf /root/backup.tar.gz -C /opt/devops/lab8 src'
sol_ch_lnav_lab8_2='tar -tzf /root/backup.tar.gz > /root/archive_list.txt'
sol_ch_lnav_lab8_3='mkdir /root/extracted && tar -xzf /root/backup.tar.gz -C /root/extracted'
sol_ch_lnav_lab8_4='cp /opt/devops/lab8/large_file.txt /root/ && gzip /root/large_file.txt'
sol_ch_lnav_lab8_5='gunzip /root/large_file.txt.gz'

# ── linux-core ──
sol_ch_lcore_lab1_1='head -5 /opt/devops/lab1/app.log > /root/first5.txt'
sol_ch_lcore_lab1_2='tail -10 /opt/devops/lab1/app.log > /root/last10.txt'
sol_ch_lcore_lab1_3='sort /opt/devops/lab1/words.txt > /root/sorted.txt'
sol_ch_lcore_lab1_4="awk '{print \$1}' /opt/devops/lab1/access.log | sort -u > /root/unique_ips.txt"

sol_ch_lcore_lab2_1='grep ERROR /opt/devops/lab2/app.log > /root/errors.txt'
sol_ch_lcore_lab2_2='grep -i error /opt/devops/lab2/mixed.log > /root/all_errors.txt'
sol_ch_lcore_lab2_3='grep -v INFO /opt/devops/lab2/app.log > /root/no_info.txt'
sol_ch_lcore_lab2_4='grep -rl password /opt/devops/lab2/configs/ > /root/has_password.txt'
sol_ch_lcore_lab2_5="grep -E 'ERROR|CRITICAL' /opt/devops/lab2/app.log > /root/critical_errors.txt"
sol_ch_lcore_lab2_6="awk '{print \$NF}' /opt/devops/lab2/access.log | sort -u > /root/status_codes.txt"

sol_ch_lcore_lab3_1="sed -i 's/localhost/db.internal/g' /opt/devops/lab3/config.txt"
sol_ch_lcore_lab3_2="awk '{print \$2}' /opt/devops/lab3/data.txt > /root/column2.txt"
sol_ch_lcore_lab3_3='grep 404 /opt/devops/lab3/access.log > /root/not_found.txt'
sol_ch_lcore_lab3_4="tr 'a-z' 'A-Z' < /opt/devops/lab3/words.txt > /root/upper.txt"

sol_ch_lcore_lab4_1='ls /opt/devops/lab4/ > /root/listing.txt'
sol_ch_lcore_lab4_2='ls /nonexistent_dir 2> /root/error_msg.txt'
sol_ch_lcore_lab4_3='cat /opt/devops/lab4/part1.log >> /root/combined.log; cat /opt/devops/lab4/part2.log >> /root/combined.log'
sol_ch_lcore_lab4_4="find / -name '*.conf' 2>/dev/null > /root/confs.txt"
sol_ch_lcore_lab4_5='/opt/devops/lab4/test.sh > /root/full_output.txt 2>&1'

sol_ch_lcore_lab5_1='grep 404 /opt/devops/lab5/access.log > /root/not_found.txt'
sol_ch_lcore_lab5_2="awk '{print \$1}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -5 > /root/top_ips.txt"
sol_ch_lcore_lab5_3="awk '{print \$(NF-1)}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 > /root/status_stats.txt"
sol_ch_lcore_lab5_4="awk '{print \$7}' /opt/devops/lab5/access.log | sort | uniq -c | sort -k1,1nr -k2,2 | head -10 > /root/top_urls.txt"

sol_ch_lcore_lab6_1='chmod +x /opt/devops/lab6/deploy.sh'
sol_ch_lcore_lab6_2='chmod 600 /opt/devops/lab6/secrets.env'
sol_ch_lcore_lab6_3='chmod 755 /opt/devops/lab6/public'
sol_ch_lcore_lab6_4='chmod -R 644 /opt/devops/lab6/configs/'
sol_ch_lcore_lab6_5='chmod 600 /opt/devops/lab6/id_rsa'

sol_ch_lcore_lab7_1='useradd -m testuser'
sol_ch_lcore_lab7_2='usermod -aG sudo devuser'

sol_ch_lcore_lab8_1='cd /opt/devops/process-demo; nohup python3 -m http.server 80 >/dev/null 2>&1 & echo $! > /root/demo.pid; sleep 1'
sol_ch_lcore_lab8_2='kill $(cat /root/demo.pid) 2>/dev/null; sleep 1'
sol_ch_lcore_lab8_3="pkill -f '[s]leep 9999'"

sol_ch_lcore_lab9_1='curl -s http://localhost/ > /root/http_page.html; curl -s -o /dev/null -w "%{http_code}" http://localhost/ > /root/http_status.txt'

sol_ch_lcore_lab10_1='printf "[Unit]\nDescription=Demo HTTP\n[Service]\nExecStart=/usr/bin/python3 -m http.server 80\nWorkingDirectory=/opt/devops/process-demo\n[Install]\nWantedBy=multi-user.target\n" > /etc/systemd/system/demo-http.service && systemctl daemon-reload'
sol_ch_lcore_lab10_2='systemctl start demo-http'
sol_ch_lcore_lab10_3='systemctl enable demo-http'

sol_ch_lcore_lab12_1='echo "Hello from editor" > /root/mynote.txt'
sol_ch_lcore_lab12_2="sed -i 's/port=9999/port=8080/' /opt/devops/lab12/broken.conf"
sol_ch_lcore_lab12_3="sed -i '1i # Edited by student' /opt/devops/lab12/notes.txt"

sol_ch_lcore_lab13_1='(crontab -l 2>/dev/null; echo "0 2 * * * /opt/devops/lab13/backup.sh") | crontab -'

sol_ch_lcore_lab14_1="sed -i 's/localhost/db.internal/g' /opt/devops/lab14/app.conf"
sol_ch_lcore_lab14_2="sed -i '/^#/d' /opt/devops/lab14/nginx.conf"
sol_ch_lcore_lab14_3="awk -F, 'NR>1{print \$2}' /opt/devops/lab14/users.csv > /root/usernames.txt"
sol_ch_lcore_lab14_4='grep 500 /opt/devops/lab14/access.log > /root/server_errors.txt'

sol_ch_lcore_lab15_1='du -h --max-depth=1 /var | sort -hr | head -5 > /root/top_dirs.txt'

sol_ch_lcore_lab16_1='mkdir /root/app_data && echo "setting=1" > /root/app_data/config.txt'
sol_ch_lcore_lab16_2='export PATH="$PATH:/opt/devops/lab16/bin"; echo $PATH > /root/new_path.txt'

# ── linux-advanced ──
sol_ch_ladv_lab1_1='curl -s -o /dev/null -w "%{http_code}" http://localhost/ > /root/http_code.txt'
sol_ch_ladv_lab1_2='curl -s http://localhost/ > /root/page.html'

sol_ch_ladv_lab2_1="ssh-keygen -t ed25519 -f /root/.ssh/mykey -N '' -q"
sol_ch_ladv_lab2_2='chmod 600 /root/.ssh/mykey && stat -c "%a" /root/.ssh/mykey > /root/key_perms.txt'
sol_ch_ladv_lab2_3='cat /root/.ssh/mykey.pub >> /root/.ssh/authorized_keys'
sol_ch_ladv_lab2_4='ssh-keygen -lf /root/.ssh/mykey.pub > /root/key_info.txt'
sol_ch_ladv_lab2_5='printf "Host myserver\n  HostName 127.0.0.1\n  User root\n  IdentityFile ~/.ssh/mykey\n" > /root/.ssh/config'

sol_ch_ladv_lab3_1='set -a; . /opt/devops/lab3/app.env; set +a; echo $DB_HOST > /root/db_host.txt'
sol_ch_ladv_lab3_2='echo "export EDITOR=nano" >> /root/.zshrc'
sol_ch_ladv_lab3_3='export PATH="$PATH:/opt/devops/lab3/bin"; which myapp > /root/myapp_path.txt'

sol_ch_ladv_lab4_1='mkdir /root/project && echo "project readme" > /root/project/README.md'
sol_ch_ladv_lab4_2='cat /root/project/config.txt 2>/dev/null || echo "default config" > /root/project/config.txt'
sol_ch_ladv_lab4_3='printf "#!/bin/bash\nset -e\nmkdir -p /root/deploy\necho deployed > /root/deploy/status.txt\n" > /root/safe_deploy.sh && chmod +x /root/safe_deploy.sh && /root/safe_deploy.sh'

sol_ch_ladv_lab5_1='printf "#!/bin/bash\necho \"Hello, DevOps!\"\n" > /root/hello.sh && chmod +x /root/hello.sh'
sol_ch_ladv_lab5_2='printf "#!/bin/bash\nVERSION=1.0.0\necho \"App version: \$VERSION\"\n" > /root/version.sh && chmod +x /root/version.sh'
sol_ch_ladv_lab5_3='printf "#!/bin/bash\nD=\$(date +%%Y-%%m-%%d)\necho \"Today: \$D\"\n" > /root/dated.sh && chmod +x /root/dated.sh'
sol_ch_ladv_lab5_4='printf "#!/bin/bash\nif [ -f /opt/devops/lab5/config.txt ]; then echo found; else echo \"not found\"; fi\n" > /root/check_file.sh && chmod +x /root/check_file.sh'
sol_ch_ladv_lab5_5='printf "#!/bin/bash\necho \"Hello, \$1!\"\n" > /root/greet.sh && chmod +x /root/greet.sh'
sol_ch_ladv_lab5_6='printf "#!/bin/bash\nif pgrep sshd >/dev/null; then echo OK; else echo FAIL; fi\n" > /root/healthcheck.sh && chmod +x /root/healthcheck.sh'

sol_ch_ladv_lab6_1='printf "#!/bin/bash\nfor e in dev staging prod; do echo \$e; done\n" > /root/list_envs.sh && chmod +x /root/list_envs.sh'
sol_ch_ladv_lab6_2='printf "#!/bin/bash\nfor f in /opt/devops/lab6/*.log; do wc -l \"\$f\"; done\n" > /root/count_logs.sh && chmod +x /root/count_logs.sh'
sol_ch_ladv_lab6_3='printf "#!/bin/bash\ni=1\nwhile [ \$i -le 5 ]; do echo \$i; i=\$((i+1)); done\n" > /root/count_to_five.sh && chmod +x /root/count_to_five.sh'
sol_ch_ladv_lab6_4='printf "#!/bin/bash\nlog() { echo \"[\$(date +%%H:%%M:%%S)] \$1\"; }\nlog \"Script started\"\nlog \"Script done\"\n" > /root/logger.sh && chmod +x /root/logger.sh'
sol_ch_ladv_lab6_5='printf "#!/bin/bash\nwhile read -r l; do echo \"Checking: \$l\"; done < /opt/devops/lab6/servers.txt\n" > /root/check_servers.sh && chmod +x /root/check_servers.sh'
sol_ch_ladv_lab6_6='printf "#!/bin/bash\nfor d in dev staging prod; do [ -d /root/\$d ] || mkdir /root/\$d; done\n" > /root/ensure_envs.sh && chmod +x /root/ensure_envs.sh && /root/ensure_envs.sh'

sol_ch_ladv_lab7_1='tar -czf /root/data_backup.tar.gz -C /opt data'
sol_ch_ladv_lab7_2='printf "#!/bin/bash\nmkdir -p /root/backups\ntar -czf /root/backups/backup_\$(date +%%F).tar.gz -C /opt data\n" > /root/backup.sh && chmod +x /root/backup.sh && /root/backup.sh'
sol_ch_ladv_lab7_3="find /root/backups -name '*.tar.gz' > /root/old_backups.txt"

sol_ch_ladv_lab8_1='apt-get update -qq >/dev/null 2>&1; apt-get install -y htop >/dev/null 2>&1'

sol_ch_ladv_lab9_1='du -h --max-depth=1 /var 2>/dev/null | sort -hr > /root/var_sizes.txt'
sol_ch_ladv_lab9_2='find /var -type f -exec du -h {} + 2>/dev/null | sort -hr | head -5 > /root/largest_files.txt'

sol_ch_ladv_lab11_1='printf "#!/bin/bash\nservers=(web1 web2 web3)\nfor s in \"\${servers[@]}\"; do echo \$s >> /root/servers.txt; done\n" > /root/array_demo.sh && chmod +x /root/array_demo.sh && /root/array_demo.sh'
sol_ch_ladv_lab11_2='line=$(cat /opt/devops/lab11/template.txt); echo "${line//HOSTNAME/prod-server-01}" > /root/result.txt'
sol_ch_ladv_lab11_3='printf "#!/bin/bash\nlog_info() { echo \"[\$(date +%%H:%%M:%%S)] INFO: \$1\"; }\nlog_info \"Script started\"\n" > /root/log_demo.sh && chmod +x /root/log_demo.sh'
sol_ch_ladv_lab11_4='printf "#!/bin/bash\nwhile read -r s; do echo \"Checking: \$s\"; done < /opt/devops/lab11/services.txt\n" > /root/process_lines.sh && chmod +x /root/process_lines.sh'

# ── git-basics ──
G='git -C /root/project'
sol_ch_git_lab1_1='git init -q /root/project && git -C /root/project config user.name student && git -C /root/project config user.email s@e.local && git -C /root/project symbolic-ref HEAD refs/heads/main'
sol_ch_git_lab1_2='echo "# My Project" > /root/project/README.md && git -C /root/project add README.md'
sol_ch_git_lab1_3='git -C /root/project commit -qm init'
sol_ch_git_lab1_4='printf "print(\"hello\")\n" > /root/project/app.py && git -C /root/project add app.py && git -C /root/project commit -qm "add app"'
sol_ch_git_lab1_5='git -C /root/project status --short > /root/status.txt'

sol_ch_git_lab2_1='echo "version = \"1.0\"" >> /root/project/app.py'
sol_ch_git_lab2_2='git -C /root/project restore app.py'
sol_ch_git_lab2_3='echo "author = \"student\"" >> /root/project/app.py && git -C /root/project add app.py'
sol_ch_git_lab2_4='git -C /root/project restore --staged app.py'
sol_ch_git_lab2_5='git -C /root/project mv README.md readme.md'
sol_ch_git_lab2_6='git -C /root/project rm -q old_config.txt && git -C /root/project commit -qm "remove old config"'

sol_ch_git_lab3_1='git -C /root/project branch feature/login'
sol_ch_git_lab3_2='git -C /root/project switch -q feature/login'
sol_ch_git_lab3_3='echo "# login module" > /root/project/login.py && git -C /root/project add login.py && git -C /root/project commit -qm "add login"'
sol_ch_git_lab3_4='git -C /root/project switch -qc feature/api'
sol_ch_git_lab3_5='git -C /root/project switch -q main && git -C /root/project merge -q feature/login -m merge'
sol_ch_git_lab3_6='git -C /root/project branch -d feature/login'

sol_ch_git_lab4_1='git -C /root/project merge branch-b >/dev/null 2>&1; git -C /root/project status --porcelain'
sol_ch_git_lab4_2='echo "Resolved version" > /root/project/app.txt'
sol_ch_git_lab4_3='git -C /root/project add app.txt'
sol_ch_git_lab4_4='git -C /root/project commit -qm "resolve conflict"'
sol_ch_git_lab4_5='git -C /root/project log --oneline --graph -5 > /root/merge_log.txt'

sol_ch_git_lab5_1='git clone -q /srv/git/project.git /root/project'
sol_ch_git_lab5_2='cd /root/project && echo feature > feature.txt && git add feature.txt && git commit -qm "add feature" && git push -q origin main'
sol_ch_git_lab5_3='git --git-dir=/srv/git/project.git log --oneline --stat -1 > /root/bare_log.txt'
sol_ch_git_lab5_4='git -C /root/project fetch -q origin && git -C /root/project branch -a > /root/branches.txt'
sol_ch_git_lab5_5='cd /root/project && git switch -qc dev && echo dev > dev.txt && git add dev.txt && git commit -qm "add dev" && git push -q origin dev'

sol_ch_git_lab6_1='git -C /root/project tag -a v1.0.0 -m "First release"'
sol_ch_git_lab6_2='git -C /root/project stash -q'
sol_ch_git_lab6_3='git -C /root/project stash pop -q'
sol_ch_git_lab6_4='H=$(git -C /root/project log --oneline | grep "bad commit" | head -1 | cut -d" " -f1); git -C /root/project revert --no-edit $H >/dev/null 2>&1; true'
sol_ch_git_lab6_5='git -C /root/project log --oneline > /root/revert_log.txt'

sol_ch_git_lab7_1='git -C /root/project log --author=student --oneline > /root/student_commits.txt'
sol_ch_git_lab7_2='git -C /root/project log --oneline -- README.md > /root/readme_history.txt'
sol_ch_git_lab7_3='git -C /root/project cherry-pick $(git -C /root/project rev-parse hotfix) >/dev/null 2>&1; true'
sol_ch_git_lab7_4='git -C /root/project diff --name-only main develop > /root/branch_diff.txt'

# ── express-devops ──
sol_ch_exp_linux_lab1_1='mkdir /root/myproject'
sol_ch_exp_linux_lab1_2='echo "Hello DevOps" > /root/myproject/README.txt'
sol_ch_exp_linux_lab1_3='cp /opt/devops/lab1/config.txt /root/myproject/config.txt'
sol_ch_exp_linux_lab1_4='mv /root/myproject/config.txt /root/myproject/myconfig.txt'
sol_ch_exp_linux_lab1_5='cat /opt/devops/lab1/.secret > /root/found_secret.txt'
sol_ch_exp_linux_lab1_6='mkdir /root/configs && mv /root/myproject/myconfig.txt /root/configs/'
sol_ch_exp_linux_lab1_7='rm /root/myproject/README.txt'
sol_ch_exp_linux_lab1_8='rm -r /root/myproject'
sol_ch_exp_linux_lab2_1='tail -10 /opt/devops/lab2/app.log > /root/last10.txt'
sol_ch_exp_linux_lab2_2='grep -v INFO /opt/devops/lab2/app.log > /root/no_info.txt'
sol_ch_exp_linux_lab2_3="find /opt/devops/lab2 -name '*.sh' > /root/scripts.txt"

# ── gym-linux-start: тренажёр повторяет лабы linux-start, решения те же ──
for _n in 1 2 3 4 5 6 7 8; do
  for _i in 1 2 3 4 5 6 7; do
    _src="sol_ch_lnav_lab${_n}_${_i}"
    [ -n "${!_src:-}" ] && printf -v "sol_gym_lstart_lab${_n}_${_i}" '%s' "${!_src}"
  done
done

# ── gym-git ──
sol_ch_git_gym_lab1_1='git clone -q /srv/git/project.git /root/project'
sol_ch_git_gym_lab1_2='cd /root/project && echo "new feature" > feature.txt && git add feature.txt && git commit -qm "Add feature"'
sol_ch_git_gym_lab1_3='git -C /root/project push -q origin main'
sol_ch_git_gym_lab1_4='git -C /root/project switch -qc dev'
sol_ch_git_gym_lab1_5='cd /root/project && echo dev > dev.txt && git add dev.txt && git commit -qm "add dev" && git switch -q main && git merge -q dev -m merge && git switch -q dev'
sol_ch_git_gym_lab2_1='cd /root/project && git checkout -q --orphan _squash && git add -A && git commit -qm "Complete feature" && git branch -qM main'
sol_ch_git_gym_lab2_2='cd /root/project && git switch -qc hotfix && echo fix > hotfix.txt && git add -A && git commit -qm "hotfix change" && git switch -q main && git cherry-pick $(git rev-parse hotfix) >/dev/null 2>&1; true'
sol_ch_git_gym_lab2_3='git -C /root/project reflog > /root/reflog.txt'
sol_ch_git_gym_lab2_4='git -C /root/project log --oneline --graph --all > /root/gitlog.txt'
sol_ch_git_gym_lab2_5='git -C /root/project reset -q HEAD~1 --soft'
sol_ch_git_gym_lab3_1='git -C /root/project tag v1.0.0'
sol_ch_git_gym_lab3_2='git -C /root/project tag -a v2.0.0 -m "Release 2.0"'
sol_ch_git_gym_lab3_3='echo wip >> /root/project/feature.txt && git -C /root/project stash -q'
sol_ch_git_gym_lab3_4='git -C /root/project stash pop -q'
sol_ch_git_gym_lab3_5='git -C /root/project revert --no-edit HEAD >/dev/null 2>&1; true'

# ── gym-linux-troubleshoot ──
sol_ch_ltrouble_lab1_1="sed -i '/PATH=\/nonexistent/d' /root/.zshrc"
sol_ch_ltrouble_lab1_2="sed -i '/^\[Service\]/a Restart=on-failure' /etc/systemd/system/simpleapp.service && systemctl daemon-reload"
sol_ch_ltrouble_lab1_3="sed -i 's#ExecStart=/usr/local/bin/myapp#ExecStart=/opt/app/myapp#' /etc/systemd/system/myapp.service && systemctl daemon-reload && systemctl start myapp"
sol_ch_ltrouble_lab1_4='/opt/app/start.sh --production'

sol_ch_ltrouble_lab2_1='chmod 700 /root/.ssh && chmod 600 /root/.ssh/authorized_keys'
sol_ch_ltrouble_lab2_2='chown svcuser /opt/app/config.yml'
sol_ch_ltrouble_lab2_3='chown appuser /opt/app/db.conf'
sol_ch_ltrouble_lab2_4='chown reportuser /opt/app/reports'

sol_ch_ltrouble_lab3_1='printf "0 3 * * * /opt/app/backup.sh\n" | crontab -'
sol_ch_ltrouble_lab3_2="sed -i 's/http.server 9090/http.server 8080/' /etc/systemd/system/webapp.service && systemctl daemon-reload && systemctl restart webapp && sleep 1"
sol_ch_ltrouble_lab3_3='kill $(cat /run/gl-stray.pid) 2>/dev/null; sleep 0.5; (cd /opt/app && setsid python3 -m http.server 3000 >/dev/null 2>&1 &); sleep 1'
sol_ch_ltrouble_lab3_4='chmod +x /opt/app/maintenance.sh'

sol_ch_ltrouble_lab4_2='P=$(pgrep -f "[p]ayment_logger" | head -1); cat /proc/$P/fd/3 > /root/recovered-payment-token.txt'

# ── sql-express ──
sol_ch_pgsql_lab_schema_1='psql -qc "CREATE DATABASE training_shop"'
sol_ch_pgsql_lab_schema_2='psql -qd training_shop -c "CREATE SCHEMA store"'
sol_ch_pgsql_lab_schema_3='psql -qd training_shop -c "CREATE TABLE store.customers (id SERIAL PRIMARY KEY, full_name TEXT NOT NULL, email TEXT UNIQUE NOT NULL, registered_at TIMESTAMPTZ NOT NULL DEFAULT now())"'
sol_ch_pgsql_lab_schema_4='psql -qd training_shop -c "CREATE TABLE store.products (id SERIAL PRIMARY KEY, name TEXT NOT NULL, category TEXT NOT NULL, price NUMERIC(10,2) NOT NULL CHECK (price >= 0), in_stock INT NOT NULL CHECK (in_stock >= 0))"'
sol_ch_pgsql_lab_schema_5='psql -qd training_shop -c "CREATE TABLE store.orders (id SERIAL PRIMARY KEY, customer_id INT NOT NULL REFERENCES store.customers(id), status TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now())"'

sol_ch_pgsql_lab_insert_1='psql -qd training_shop -c "INSERT INTO store.customers (full_name, email) VALUES ('"'"'Alex'"'"','"'"'alex@example.test'"'"'),('"'"'Max'"'"','"'"'max@example.test'"'"')"'
sol_ch_pgsql_lab_insert_2='psql -qd training_shop -c "INSERT INTO store.products (name, category, price, in_stock) VALUES ('"'"'PostgreSQL Guide'"'"','"'"'book'"'"',1900.00,25),('"'"'SQL Basics Course'"'"','"'"'course'"'"',4900.00,100),('"'"'Sticker Pack'"'"','"'"'merch'"'"',500.00,200)"'
sol_ch_pgsql_lab_insert_3='psql -qd training_shop -c "INSERT INTO store.orders (customer_id, status) VALUES (1,'"'"'new'"'"'),(2,'"'"'paid'"'"')"'
sol_ch_pgsql_lab_insert_4='psql -qd training_shop -c "INSERT INTO store.order_items (order_id, product_id, quantity, price) VALUES (1,2,1,4900.00),(2,1,2,1900.00),(2,3,3,500.00)"'

sol_ch_pgsql_lab3_1='psql -qd training_shop -c "CREATE TABLE query_notes (id SERIAL PRIMARY KEY, topic TEXT NOT NULL, note TEXT, completed BOOLEAN NOT NULL DEFAULT false, created_at TIMESTAMPTZ NOT NULL DEFAULT now())"'
sol_ch_pgsql_lab3_2='psql -qd training_shop -c "INSERT INTO query_notes (topic, note) VALUES ('"'"'select'"'"','"'"'выборка'"'"'),('"'"'join'"'"','"'"'соединения'"'"'),('"'"'transaction'"'"','"'"'транзакции'"'"')"'
sol_ch_pgsql_lab3_3='psql -qd training_shop -c "UPDATE query_notes SET completed = true WHERE topic IN ('"'"'select'"'"','"'"'join'"'"')"'
sol_ch_pgsql_lab3_4='psql -qd training_shop -c "DELETE FROM query_notes WHERE completed = false"'
sol_ch_pgsql_lab3_5='psql -qd training_shop -c "CREATE INDEX idx_orders_created_at ON orders (created_at)"'

# ── linux-terminal-start ──
sol_navigation_and_links_1='cd /tmp && pwd > /root/.gl_cwd'
sol_navigation_and_links_2='mkdir -p /root/workspace && echo hello > /root/workspace/readme.txt'
sol_navigation_and_links_3='cd /root && echo GoLearn > data.txt && ln -sf data.txt latest.txt'
sol_reading_and_search_1='grep ERROR /root/log.txt > /root/errors.txt'
sol_reading_and_search_2="find /root/proj -name '*.md' > /root/md.txt"
sol_permissions_1='printf "#!/bin/bash\n" > /root/run.sh && chmod +x /root/run.sh'
sol_permissions_2='chmod 600 /root/secret.txt'
sol_redirection_and_pipes_1='ls /etc > /root/etc.txt && echo END >> /root/etc.txt'
sol_redirection_and_pipes_2="ls /root/cfg | grep '\.conf\$' | wc -l > /root/n.txt"
sol_text_processing_1='sort /root/names.txt | uniq > /root/unique.txt'
sol_text_processing_2='cut -d: -f1 /root/users.csv > /root/just_names.txt'
sol_env_variables_1='echo $HOME > /root/home.txt'
sol_env_variables_2='GREETING=hello; echo $GREETING > /root/greet.txt'
sol_processes_and_signals_1="pkill -f 'sleep 10000[0]'"
sol_processes_and_signals_2='pgrep -af sleep > /root/pid1.txt'
sol_users_groups_sudo_1='echo "$(id -un):$(id -u)" > /root/me.txt'
sol_users_groups_sudo_2='useradd alice'
sol_users_groups_sudo_3='touch /root/report.txt && chgrp daemon /root/report.txt'

# ── Git: rebase, reset/reflog, bisect ──
# Lab 8 squashes interactively in the lesson; the non-interactive equivalent is a
# soft reset plus one commit, which lands the same tree and history.
sol_ch_git_lab8_1='git -C /root/project reset --soft HEAD~3 && git -C /root/project commit -qm "Add login"'
sol_ch_git_lab8_2='git -C /root/project commit --amend -qm "Add login feature"'
sol_ch_git_lab8_3='git -C /root/project status --short > /root/status.txt'

sol_ch_git_lab9_1='git -C /root/project reset --soft HEAD~1'
sol_ch_git_lab9_2='git -C /root/project reset --hard HEAD~1'
# The lesson reads the hash off `git reflog` by eye; here it is looked up by the
# reflog subject so the harness does not depend on a fixed HEAD@{n} position.
sol_ch_git_lab9_3='h=$(git -C /root/project reflog --format="%h %gs" | grep -m1 "commit: add a.txt" | cut -d" " -f1) && git -C /root/project reset --hard "$h"'

sol_ch_git_lab10_1='cd /root/project && git bisect start HEAD good-start >/dev/null 2>&1 && git bisect run ./test.sh >/dev/null 2>&1; true'
sol_ch_git_lab10_2='git -C /root/project bisect reset'

# ── Ansible ──
# Every lab runs against localhost with connection: local, so the playbooks here
# are the real thing rather than a stand-in. Each solution writes the whole
# playbook it needs and runs it: the tasks build on each other, and rewriting the
# file keeps a solution readable on its own.
sol_ch_ansible_lab1_1='cd /root/ansible-lab && ansible localhost -c local -m file -a "path=/root/ansible-lab/hello.txt state=touch" >/dev/null'
sol_ch_ansible_lab1_2='cd /root/ansible-lab && cat > site.yml <<YML
- hosts: local
  tasks:
    - name: managed file
      copy:
        content: "Managed by Ansible\n"
        dest: /root/ansible-lab/managed.txt
YML
ansible-playbook site.yml >/dev/null'
sol_ch_ansible_lab1_3='cd /root/ansible-lab && cat > site.yml <<YML
- hosts: local
  tasks:
    - name: managed file
      copy:
        content: "Managed by Ansible\n"
        dest: /root/ansible-lab/managed.txt
    - name: data dir
      file:
        path: /root/ansible-lab/data
        state: directory
YML
ansible-playbook site.yml >/dev/null'
sol_ch_ansible_lab1_4='cd /root/ansible-lab && ansible-playbook site.yml > /root/ansible-lab/recap.txt 2>&1'

sol_ch_ansible_lab2_1='cd /root/ansible-lab && cat > files.yml <<YML
- hosts: local
  tasks:
    - name: app dir
      file:
        path: /root/ansible-lab/app
        state: directory
YML
ansible-playbook files.yml >/dev/null'
sol_ch_ansible_lab2_2='cd /root/ansible-lab && cat > files.yml <<YML
- hosts: local
  tasks:
    - name: app dir
      file:
        path: /root/ansible-lab/app
        state: directory
    - name: app conf
      copy:
        content: "port=8080\n"
        dest: /root/ansible-lab/app/app.conf
YML
ansible-playbook files.yml >/dev/null'
sol_ch_ansible_lab2_3='cd /root/ansible-lab && cat >> files.yml <<YML
    - name: debug line
      lineinfile:
        path: /root/ansible-lab/app/app.conf
        line: "debug=true"
YML
ansible-playbook files.yml >/dev/null'
sol_ch_ansible_lab2_4='cd /root/ansible-lab && cat >> files.yml <<YML
    - name: drop old file
      file:
        path: /root/ansible-lab/old.txt
        state: absent
YML
ansible-playbook files.yml >/dev/null'

sol_ch_ansible_lab3_1='cd /root/ansible-lab && cat > vars.yml <<YML
- hosts: local
  vars:
    greeting: "hello ansible"
  tasks:
    - name: greeting file
      copy:
        content: "{{ greeting }}\n"
        dest: /root/ansible-lab/greeting.txt
YML
ansible-playbook vars.yml >/dev/null'
sol_ch_ansible_lab3_2='cd /root/ansible-lab && cat >> vars.yml <<YML
    - name: distribution fact
      copy:
        content: "{{ ansible_facts[\"distribution\"] }}\n"
        dest: /root/ansible-lab/os.txt
YML
ansible-playbook vars.yml >/dev/null'
sol_ch_ansible_lab3_3='cd /root/ansible-lab && cat >> vars.yml <<YML
    - name: read hostname
      command: cat /etc/hostname
      register: host_out
      changed_when: false
    - name: store hostname
      copy:
        content: "{{ host_out.stdout }}\n"
        dest: /root/ansible-lab/host.txt
YML
ansible-playbook vars.yml >/dev/null'

sol_ch_ansible_lab4_1='cd /root/ansible-lab && cat > templates/service.conf.j2 <<J2
port = {{ svc_port }}
J2
cat > tpl.yml <<YML
- hosts: local
  vars:
    svc_port: 9090
    enable_debug: false
  tasks:
    - name: render service conf
      template:
        src: templates/service.conf.j2
        dest: /root/ansible-lab/service.conf
YML
ansible-playbook tpl.yml >/dev/null'
sol_ch_ansible_lab4_2='cd /root/ansible-lab && cat > templates/service.conf.j2 <<J2
port = {{ svc_port }}
{% if enable_debug %}debug = on{% endif %}
J2
sed -i "s/enable_debug: false/enable_debug: true/" tpl.yml
ansible-playbook tpl.yml >/dev/null'
# The handler only fires on a change, so the port moves to 9091 in the same run.
sol_ch_ansible_lab4_3='cd /root/ansible-lab && cat > tpl.yml <<YML
- hosts: local
  vars:
    svc_port: 9091
    enable_debug: true
  tasks:
    - name: render service conf
      template:
        src: templates/service.conf.j2
        dest: /root/ansible-lab/service.conf
      notify: reload service
  handlers:
    - name: reload service
      file:
        path: /root/ansible-lab/reloaded.marker
        state: touch
YML
ansible-playbook tpl.yml >/dev/null'

sol_ch_ansible_lab5_1='cd /root/ansible-lab && cat > loops.yml <<YML
- hosts: local
  tasks:
    - name: dirs
      file:
        path: /root/ansible-lab/{{ item }}
        state: directory
      loop: [logs, data, cache]
YML
ansible-playbook loops.yml >/dev/null'
sol_ch_ansible_lab5_2='cd /root/ansible-lab && cat >> loops.yml <<YML
    - name: flags except beta
      file:
        path: /root/ansible-lab/{{ item }}.flag
        state: touch
      loop: [alpha, beta, gamma]
      when: item != "beta"
YML
ansible-playbook loops.yml >/dev/null'
sol_ch_ansible_lab5_3='cd /root/ansible-lab && cat > cond.yml <<YML
- hosts: local
  vars:
    make_prod: true
  tasks:
    - name: prod marker
      file:
        path: /root/ansible-lab/prod.marker
        state: touch
      when: make_prod | bool
YML
ansible-playbook cond.yml >/dev/null'

sol_ch_ansible_lab6_1='cd /root/ansible-lab && ansible-galaxy init roles/webapp >/dev/null'
sol_ch_ansible_lab6_2='cd /root/ansible-lab && cat > roles/webapp/tasks/main.yml <<YML
- name: webapp dir
  file:
    path: /root/ansible-lab/webapp
    state: directory
- name: index page
  copy:
    content: "<h1>webapp</h1>\n"
    dest: /root/ansible-lab/webapp/index.html
YML
cat > site.yml <<YML
- hosts: local
  roles:
    - webapp
YML
ansible-playbook site.yml >/dev/null'
sol_ch_ansible_lab6_3='cd /root/ansible-lab && ansible-playbook site.yml > /root/ansible-lab/role_recap.txt 2>&1'
sol_ch_ansible_lab6_4='cd /root/ansible-lab && printf "db_password: s3cret\n" > secrets.yml && printf "vaultpw\n" > .vaultpw && ansible-vault encrypt --vault-password-file .vaultpw secrets.yml >/dev/null'

# ── GitLab CI ──
# There is no GitLab in the sandbox: these labs are about authoring
# .gitlab-ci.yml, and the checks read the file. Each solution rewrites the whole
# file rather than patching it in place — the jobs accumulate, and a full file is
# both what a student would end up with and readable without the earlier steps.
sol_ch_gitlab_ci_lab1_1='cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<YML
hello:
  script:
    - echo "Hello, CI"
YML'
sol_ch_gitlab_ci_lab1_2='cd /root/gitlab-ci-lab && git add .gitlab-ci.yml && git commit -qm "Add pipeline"'
sol_ch_gitlab_ci_lab1_3='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

shell-basics:
  script:
    - echo "step 1"
    - date
    - ls -la
YML'
sol_ch_gitlab_ci_lab1_4='cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<YML
hello:
  script:
    - echo "Hello, CI"

shell-basics:
  script:
    - |
      echo "start"
      for i in 1 2 3; do echo "step \$i"; done
      echo "end"
YML'

sol_ch_gitlab_ci_lab2_1='cd /root/gitlab-ci-lab && sed -i "/^  APP_NAME: ci-demo/a\  APP_ENV: lab\n  PACKAGE_NAME: ci-demo-package" .gitlab-ci.yml'
sol_ch_gitlab_ci_lab2_2='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("build-package:\n  stage: build\n  script:\n    - echo \"building\"\n",
            "build-package:\n  stage: build\n  script:\n    - echo \"building\"\n  artifacts:\n    paths:\n      - dist/\n    expire_in: 1 hour\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab2_3='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("test-package:\n  stage: test\n",
            "test-package:\n  stage: test\n  needs:\n    - build-package\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab2_4='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

test-scripts:
  stage: validate
  script:
    - ./scripts/test.sh
YML'
sol_ch_gitlab_ci_lab2_5='cd /root/gitlab-ci-lab && sed -i "/^  APP_NAME: ci-demo/a\  CI_DEBUG: \"true\"" .gitlab-ci.yml'
sol_ch_gitlab_ci_lab2_6='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("test-package:\n  stage: test\n",
            "test-package:\n  stage: test\n  artifacts:\n    paths:\n      - test-results/\n")
io.open(p,"w").write(s)
PY'

sol_ch_gitlab_ci_lab3_1='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

manual-deploy:
  stage: deploy
  when: manual
  script:
    - echo "deploying by hand"
YML'
sol_ch_gitlab_ci_lab3_2='cd /root/gitlab-ci-lab && git add -A && git commit -qm "manual deploy job" >/dev/null 2>&1; git tag -a v0.1.0 -m "first release"'
sol_ch_gitlab_ci_lab3_3='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

release-note:
  stage: deploy
  script:
    - echo "release \$CI_COMMIT_TAG"
YML'
sol_ch_gitlab_ci_lab3_4='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("release-note:\n  stage: deploy\n",
            "release-note:\n  stage: deploy\n  rules:\n    - if: \$CI_COMMIT_TAG\n")
io.open(p,"w").write(s)
PY'

sol_ch_gitlab_ci_lab4_1='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

build-image:
  stage: build
  script:
    - docker build -t \$CI_REGISTRY_IMAGE:\$CI_COMMIT_SHA .
YML'
sol_ch_gitlab_ci_lab4_2='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

deploy-compose:
  stage: deploy
  script:
    - docker compose up -d
YML'
sol_ch_gitlab_ci_lab4_3='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("deploy-compose:\n  stage: deploy\n",
            "deploy-compose:\n  stage: deploy\n  environment:\n    name: production\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab4_4='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("  script:\n    - docker compose up -d\n",
            "  rules:\n    - if: \$CI_COMMIT_TAG\n  script:\n    - docker compose up -d\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab4_5='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML
    - echo "pipeline \$CI_PIPELINE_ID"
YML'

sol_ch_gitlab_ci_lab5_1='cd /root/gitlab-ci-lab && echo "change" >> README.md && git commit -qam "trigger"'
sol_ch_gitlab_ci_lab5_2='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

full-tests:
  stage: test
  rules:
    - if: \$RUN_FULL_TESTS == "true"
  script:
    - echo "full suite"
YML'
sol_ch_gitlab_ci_lab5_3='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
io.open(p,"w").write("spec:\n  inputs:\n    deploy_target:\n      default: staging\n---\n" + s)
PY'
sol_ch_gitlab_ci_lab5_4='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

show-target:
  stage: test
  script:
    - echo "target \$[[ inputs.deploy_target ]]"
YML'
sol_ch_gitlab_ci_lab5_5='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

deploy-urls:
  stage: deploy
  variables:
    DEPLOY_STAGE_URL: "https://staging.example"
    DEPLOY_PROD_URL: "https://example"
  script:
    - echo "urls set"
YML'
sol_ch_gitlab_ci_lab5_6='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("    DEPLOY_PROD_URL: \"https://example\"\n",
            "    DEPLOY_PROD_URL: \"https://example\"\n    DEPLOY_NOTE: \"scoped to this job\"\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab5_7='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

deploy-prod:
  stage: deploy
  environment:
    name: production
  script:
    - echo "to prod"
YML'
sol_ch_gitlab_ci_lab5_8='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

deploy-stage:
  stage: deploy
  environment:
    name: staging
  script:
    - echo "to staging"
YML'

sol_ch_gitlab_ci_lab6_1='mkdir -p /root/gitlab-ci-lab/ci && cat > /root/gitlab-ci-lab/ci/security.yml <<YML
secret-scan:
  stage: test
  script:
    - echo "scanning secrets"
YML
cd /root/gitlab-ci-lab && sed -i "/- local: ci\/templates.yml/a\  - local: ci/security.yml" .gitlab-ci.yml'
sol_ch_gitlab_ci_lab6_2='cd /root/gitlab-ci-lab && git add -A && git commit -qm "Add reusable security include" >/dev/null 2>&1 || true'
sol_ch_gitlab_ci_lab6_3='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

.lint-template:
  script:
    - echo "lint"

lint:
  extends: .lint-template
  stage: test
YML'
sol_ch_gitlab_ci_lab6_4='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

compat-test:
  stage: test
  image: alpine:\$ALPINE_VERSION
  parallel:
    matrix:
      - ALPINE_VERSION: ["3.19", "3.20"]
  script:
    - cat /etc/alpine-release
YML'

sol_ch_gitlab_ci_lab7_1='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

unit-report:
  stage: test
  script:
    - echo "tests"
  artifacts:
    reports:
      junit: report.xml
YML'
sol_ch_gitlab_ci_lab7_2='cd /root/gitlab-ci-lab && git add -A && git commit -qm "Add junit report" >/dev/null 2>&1 || true'
sol_ch_gitlab_ci_lab7_3='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("unit-report:\n  stage: test\n",
            "unit-report:\n  stage: test\n  retry: 1\n  timeout: 5 minutes\n")
io.open(p,"w").write(s)
PY'
sol_ch_gitlab_ci_lab7_4='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

deploy-production:
  stage: deploy
  resource_group: production
  script:
    - echo "one at a time"
YML'
sol_ch_gitlab_ci_lab7_5='cat >> /root/gitlab-ci-lab/.gitlab-ci.yml <<YML

optional-lint:
  stage: test
  script:
    - echo "lint"
YML'
sol_ch_gitlab_ci_lab7_6='cd /root/gitlab-ci-lab && python3 - <<PY
import io
p=".gitlab-ci.yml"; s=io.open(p).read()
s=s.replace("optional-lint:\n  stage: test\n",
            "optional-lint:\n  stage: test\n  allow_failure: true\n")
io.open(p,"w").write(s)
PY'
