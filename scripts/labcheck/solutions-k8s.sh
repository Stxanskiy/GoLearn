# Reference solutions for the Kubernetes courses (sourced by run.sh).
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# helpers used inside the container by several solutions
K_WAIT_DEPLOY='f(){ kubectl rollout status deployment/$1 --timeout=180s >/dev/null 2>&1; }; f'

sol_ch_k8si_lab1_1='kubectl run nginx-pod --image=nginx:alpine >/dev/null'
sol_ch_k8si_lab1_2='kubectl wait --for=condition=Ready pod/nginx-pod --timeout=180s >/dev/null 2>&1'
sol_ch_k8si_lab1_3="kubectl get pod nginx-pod -o jsonpath='{.status.podIP}' > /root/nginx_pod_ip.txt"
sol_ch_k8si_lab1_4='kubectl delete pod nginx-pod --wait >/dev/null'

sol_ch_k8si_lab2_1='kubectl create deployment nginx-deploy --image=nginx:alpine --replicas=2 >/dev/null'
sol_ch_k8si_lab2_2='kubectl rollout status deployment/nginx-deploy --timeout=180s >/dev/null'
sol_ch_k8si_lab2_3='kubectl scale deployment nginx-deploy --replicas=1 >/dev/null'
sol_ch_k8si_lab2_4='kubectl patch deployment nginx-deploy -p "{\"spec\":{\"strategy\":{\"type\":\"RollingUpdate\",\"rollingUpdate\":{\"maxSurge\":1,\"maxUnavailable\":0}}}}" >/dev/null'
sol_ch_k8si_lab2_5='kubectl set image deployment/nginx-deploy nginx=nginx:1.25-alpine >/dev/null && kubectl rollout status deployment/nginx-deploy --timeout=180s >/dev/null'
sol_ch_k8si_lab2_6='kubectl rollout undo deployment/nginx-deploy >/dev/null && kubectl rollout status deployment/nginx-deploy --timeout=180s >/dev/null'
sol_ch_k8si_lab2_7='kubectl delete deployment nginx-deploy --wait >/dev/null'

sol_ch_k8si_lab2_manifests_1='kubectl apply -f /root/manifest-deploy/deployment.yaml >/dev/null && kubectl rollout status deployment/manifest-web --timeout=180s >/dev/null'
sol_ch_k8si_lab2_manifests_2='sed -i "s/replicas: 2/replicas: 3/" /root/manifest-deploy/deployment.yaml && kubectl apply -f /root/manifest-deploy/deployment.yaml >/dev/null && kubectl rollout status deployment/manifest-web --timeout=180s >/dev/null'
sol_ch_k8si_lab2_manifests_3='sed -i "s|image: nginx:alpine|image: nginx:1.25-alpine|" /root/manifest-deploy/deployment.yaml && kubectl apply -f /root/manifest-deploy/deployment.yaml >/dev/null && kubectl rollout status deployment/manifest-web --timeout=180s >/dev/null'
sol_ch_k8si_lab2_manifests_4='sed -i "s/app: wrong-label/app: broken-web/; s/containerPort: \"80\"/containerPort: 80/" /root/manifest-deploy/broken-deployment.yaml && kubectl apply -f /root/manifest-deploy/broken-deployment.yaml >/dev/null && kubectl rollout status deployment/broken-web --timeout=180s >/dev/null'
sol_ch_k8si_lab2_manifests_5='kubectl delete deployment manifest-web broken-web --wait >/dev/null'

sol_ch_k8si_lab2_rollout_1='kubectl create deployment rollout-app --image=nginx:alpine --replicas=3 >/dev/null && kubectl annotate deployment rollout-app kubernetes.io/change-cause="Initial deployment" --overwrite >/dev/null && kubectl rollout status deployment/rollout-app --timeout=180s >/dev/null'
sol_ch_k8si_lab2_rollout_2='kubectl patch deployment rollout-app -p "{\"spec\":{\"strategy\":{\"type\":\"RollingUpdate\",\"rollingUpdate\":{\"maxSurge\":1,\"maxUnavailable\":0}}}}" >/dev/null'
sol_ch_k8si_lab2_rollout_3='kubectl set image deployment/rollout-app nginx=nginx:1.25-alpine >/dev/null && kubectl rollout status deployment/rollout-app --timeout=180s >/dev/null && kubectl annotate deployment rollout-app kubernetes.io/change-cause="Update nginx to 1.25" --overwrite >/dev/null'
sol_ch_k8si_lab2_rollout_4='kubectl rollout history deployment/rollout-app > /root/rollout_history.txt'
sol_ch_k8si_lab2_rollout_5='kubectl rollout undo deployment/rollout-app >/dev/null && kubectl rollout status deployment/rollout-app --timeout=180s >/dev/null && kubectl annotate deployment rollout-app kubernetes.io/change-cause="Rollback to nginx:alpine" --overwrite >/dev/null'
sol_ch_k8si_lab2_rollout_6='kubectl rollout history deployment/rollout-app > /root/revision1.txt'
sol_ch_k8si_lab2_rollout_7='kubectl delete deployment rollout-app --wait >/dev/null'

sol_ch_k8si_lab7_1='kubectl run web-prod --image=nginx:alpine --labels=app=web,env=prod >/dev/null; kubectl run web-staging --image=nginx:alpine --labels=app=web,env=staging >/dev/null; kubectl run db-prod --image=nginx:alpine --labels=app=db,env=prod >/dev/null'
sol_ch_k8si_lab7_2='kubectl label pod web-prod monitored=true >/dev/null'
sol_ch_k8si_lab7_3='kubectl label pod web-prod monitored- >/dev/null'
sol_ch_k8si_lab7_4='kubectl delete pod web-prod web-staging db-prod --wait >/dev/null'

sol_ch_k8si_lab3_1='kubectl create deployment web-app --image=nginx:alpine --replicas=2 >/dev/null && kubectl rollout status deployment/web-app --timeout=180s >/dev/null'
sol_ch_k8si_lab3_2='kubectl expose deployment web-app --name=web-svc --port=80 --target-port=80 >/dev/null'
sol_ch_k8si_lab3_3='kubectl expose deployment web-app --name=web-nodeport --type=NodePort --port=80 --target-port=80 --dry-run=client -o yaml > /root/web-nodeport.yaml && kubectl apply -f /root/web-nodeport.yaml >/dev/null'
sol_ch_k8si_lab3_4='kubectl get svc > /root/services.txt'
sol_ch_k8si_lab3_5='kubectl delete svc web-svc web-nodeport --wait >/dev/null'

sol_ch_k8si_lab8_service_manifests_1='kubectl apply -f /root/service-manifests/deployment.yaml -f /root/service-manifests/service.yaml >/dev/null && kubectl rollout status deployment/store-web --timeout=180s >/dev/null'
sol_ch_k8si_lab8_service_manifests_2='sed -i "s/app: shop/app: store/" /root/service-manifests/service.yaml && kubectl apply -f /root/service-manifests/service.yaml >/dev/null; sleep 3'
sol_ch_k8si_lab8_service_manifests_3='sed -i "s/targetPort: 8080/targetPort: 80/" /root/service-manifests/service.yaml && kubectl apply -f /root/service-manifests/service.yaml >/dev/null; sleep 2'
sol_ch_k8si_lab8_service_manifests_4='kubectl run store-client --image=busybox:1.28 --restart=Never --command -- sleep 3600 >/dev/null 2>&1; kubectl wait --for=condition=Ready pod/store-client --timeout=180s >/dev/null 2>&1; kubectl exec store-client -- wget -qO- http://store-svc > /root/store_http.txt 2>/dev/null'
sol_ch_k8si_lab8_service_manifests_5='sed -i "s/replicas: 2/replicas: 3/" /root/service-manifests/deployment.yaml && kubectl apply -f /root/service-manifests/deployment.yaml >/dev/null && kubectl rollout status deployment/store-web --timeout=180s >/dev/null; sleep 3'
sol_ch_k8si_lab8_service_manifests_6='kubectl delete svc store-svc --wait >/dev/null; kubectl delete deployment store-web --wait >/dev/null; kubectl delete pod store-client --ignore-not-found --wait >/dev/null'

sol_ch_k8si_lab_headless_stateful_1='kubectl apply -f /root/headless-lab/web-deploy.yaml -f /root/headless-lab/web-sts.yaml >/dev/null && kubectl rollout status deployment/web-deploy --timeout=180s >/dev/null && kubectl rollout status statefulset/web-sts --timeout=300s >/dev/null'
sol_ch_k8si_lab_headless_stateful_2='kubectl apply -f /root/headless-lab/web-headless.yaml >/dev/null; sleep 3'
sol_ch_k8si_lab_headless_stateful_3='kubectl apply -f /root/headless-lab/web-headless-broken.yaml >/dev/null; sleep 3'
sol_ch_k8si_lab_headless_stateful_4='kubectl scale statefulset web-sts --replicas=5 >/dev/null && kubectl rollout status statefulset/web-sts --timeout=300s >/dev/null'
sol_ch_k8si_lab_headless_stateful_5='kubectl delete statefulset web-sts --wait >/dev/null; kubectl delete svc web-headless web-svc --wait >/dev/null; kubectl delete deployment web-deploy --wait >/dev/null'

sol_ch_k8si_lab4_1='kubectl create configmap app-config --from-literal=APP_ENV=production --from-literal=LOG_LEVEL=info --from-literal=PORT=8080 >/dev/null'
sol_ch_k8si_lab4_2='python3 - <<PYEOF
p="/root/cm-pod.yaml"
s=open(p).read()
s=s.replace("""      command: ["sh", "-c", "echo APP_ENV=\$APP_ENV; sleep 3600"]""","""      command: ["sh", "-c", "echo APP_ENV=\$APP_ENV; sleep 3600"]
      envFrom:
        - configMapRef:
            name: app-config""")
open(p,"w").write(s)
PYEOF
kubectl apply -f /root/cm-pod.yaml >/dev/null && kubectl wait --for=condition=Ready pod/cm-pod --timeout=180s >/dev/null 2>&1'
sol_ch_k8si_lab4_3='kubectl logs cm-pod > /root/cm_env.txt'
sol_ch_k8si_lab4_4='kubectl create secret generic db-secret --from-literal=DB_USER=admin --from-literal=DB_PASSWORD=supersecret123 >/dev/null'
sol_ch_k8si_lab4_5='kubectl get secret db-secret -o jsonpath="{.data.DB_PASSWORD}" | base64 -d > /root/db_password.txt'
sol_ch_k8si_lab4_6='printf "server.port=8080\nlog.level=debug\ndb.pool.size=10\n" > /root/app.conf && kubectl create configmap file-config --from-file=/root/app.conf >/dev/null'
sol_ch_k8si_lab4_7='kubectl delete configmap app-config file-config --wait >/dev/null; kubectl delete secret db-secret --wait >/dev/null; kubectl delete pod cm-pod --wait >/dev/null'

sol_ch_k8si_lab4_manifests_1='kubectl apply -f /root/config-manifests/app-config.yaml -f /root/config-manifests/app-secret.yaml >/dev/null'
sol_ch_k8si_lab4_manifests_2='kubectl apply -f /root/config-manifests/app-pod.yaml >/dev/null && kubectl wait --for=condition=Ready pod/manifest-config-app --timeout=180s >/dev/null 2>&1'
sol_ch_k8si_lab4_manifests_3='sed -i "s/APP_ENV: staging/APP_ENV: production/; s/FEATURE_FLAG: disabled/FEATURE_FLAG: enabled/" /root/config-manifests/app-config.yaml && kubectl apply -f /root/config-manifests/app-config.yaml >/dev/null'
sol_ch_k8si_lab4_manifests_4='kubectl delete pod manifest-config-app --wait >/dev/null && kubectl apply -f /root/config-manifests/app-pod.yaml >/dev/null && kubectl wait --for=condition=Ready pod/manifest-config-app --timeout=180s >/dev/null 2>&1'
sol_ch_k8si_lab4_manifests_5='sed -i "s/name: wrong-secret/name: manifest-secret/" /root/config-manifests/broken-secret-pod.yaml && kubectl delete pod secret-broken-app --ignore-not-found --wait >/dev/null && kubectl apply -f /root/config-manifests/broken-secret-pod.yaml >/dev/null && kubectl wait --for=condition=Ready pod/secret-broken-app --timeout=180s >/dev/null 2>&1; sleep 2'
sol_ch_k8si_lab4_manifests_6='kubectl delete pod manifest-config-app secret-broken-app --wait >/dev/null; kubectl delete configmap manifest-config --wait >/dev/null; kubectl delete secret manifest-secret --wait >/dev/null'

sol_ch_k8si_lab6_1='cat > /root/resource-pod.yaml <<YEOF
apiVersion: v1
kind: Pod
metadata:
  name: resource-pod
  labels:
    run: resource-pod
spec:
  containers:
    - name: nginx
      image: nginx:alpine
      resources:
        requests:
          cpu: 50m
          memory: 64Mi
        limits:
          cpu: 200m
          memory: 128Mi
YEOF
kubectl apply -f /root/resource-pod.yaml >/dev/null && kubectl wait --for=condition=Ready pod/resource-pod --timeout=180s >/dev/null 2>&1'
sol_ch_k8si_lab6_2='cat > /root/limited-web.yaml <<YEOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: limited-web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: limited-web
  template:
    metadata:
      labels:
        app: limited-web
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 200m
              memory: 128Mi
YEOF
kubectl apply -f /root/limited-web.yaml >/dev/null && kubectl rollout status deployment/limited-web --timeout=180s >/dev/null'
sol_ch_k8si_lab6_3='sed -i "s/memory: 128Mi/memory: 256Mi/" /root/limited-web.yaml && kubectl apply -f /root/limited-web.yaml >/dev/null && kubectl rollout status deployment/limited-web --timeout=180s >/dev/null'
sol_ch_k8si_lab6_4='kubectl delete pod resource-pod --ignore-not-found --wait >/dev/null; kubectl delete deployment limited-web --wait >/dev/null; kubectl delete pod too-large-pod --ignore-not-found --wait >/dev/null'

sol_ch_k8si_lab9_dns_1='kubectl create deployment web-svc-app --image=nginx:alpine --replicas=2 >/dev/null && kubectl rollout status deployment/web-svc-app --timeout=180s >/dev/null && kubectl expose deployment web-svc-app --name=web-svc --port=80 >/dev/null'

sol_ch_k8si_lab10_ingress_1='kubectl apply -f /root/ingress-lab/app.yaml >/dev/null && for d in frontend catalog cart; do kubectl rollout status deployment/$d --timeout=180s >/dev/null; done'
sol_ch_k8si_lab10_ingress_2='kubectl create ingress store-ingress --rule="/*=frontend-svc:80" --rule="/catalog*=catalog-svc:80" --rule="/cart*=cart-svc:80" >/dev/null; kubectl -n kube-system rollout status deployment/traefik --timeout=180s >/dev/null 2>&1; for i in $(seq 1 30); do kubectl get ingress store-ingress -o jsonpath="{.status.loadBalancer.ingress[0].ip}" 2>/dev/null | grep -q . && break; sleep 2; done'
sol_ch_k8si_lab10_ingress_3='for i in $(seq 1 45); do curl -s --max-time 5 http://10.55.0.2/ | grep -q "store frontend" && break; sleep 2; done; curl -s --max-time 10 http://10.55.0.2/ > /root/ingress_root.txt'
sol_ch_k8si_lab10_ingress_4='kubectl delete ingress store-ingress --ignore-not-found --wait >/dev/null; kubectl delete service frontend-svc catalog-svc cart-svc --ignore-not-found --wait >/dev/null; kubectl delete deployment frontend catalog cart --ignore-not-found --wait >/dev/null; kubectl delete configmap frontend-content catalog-source cart-source --ignore-not-found --wait >/dev/null'

# ── Helm ──
# These run in the k3s golden, so installs really happen. --wait where the check
# reads readyReplicas: without it the check races the rollout and fails for a
# reason that has nothing to do with the student.
sol_ch_helm_lab1_1='helm install my-nginx /root/charts/nginx-chart --wait --timeout 120s >/dev/null'
sol_ch_helm_lab1_2='kubectl rollout status deploy/my-nginx-nginx --timeout=120s >/dev/null'
sol_ch_helm_lab1_3='helm status my-nginx >/dev/null'
sol_ch_helm_lab1_4='helm uninstall my-nginx >/dev/null'

sol_ch_helm_lab2_1='helm install scaled-nginx /root/charts/nginx-chart --set replicaCount=2 --wait --timeout 120s >/dev/null'
sol_ch_helm_lab2_2='kubectl rollout status deploy/scaled-nginx-nginx --timeout=120s >/dev/null'
sol_ch_helm_lab2_3='cat > /root/prod-values.yaml <<YML
replicaCount: 3
image:
  repository: nginx
  tag: alpine
YML'
sol_ch_helm_lab2_4='helm install prod-nginx /root/charts/nginx-chart -f /root/prod-values.yaml --wait --timeout 180s >/dev/null'
sol_ch_helm_lab2_5='helm upgrade prod-nginx /root/charts/nginx-chart -f /root/prod-values.yaml --set replicaCount=1 --wait --timeout 120s >/dev/null'
sol_ch_helm_lab2_6='helm lint /root/charts/nginx-chart >/dev/null'
sol_ch_helm_lab2_7='helm uninstall scaled-nginx prod-nginx >/dev/null'

sol_ch_helm_lab3_1='helm create /root/myapp-chart >/dev/null'
sol_ch_helm_lab3_2='cd /root/myapp-chart && sed -i "s/^name: .*/name: myapp-chart/; s/^description: .*/description: TOT training chart/" Chart.yaml'
sol_ch_helm_lab3_3='cd /root/myapp-chart && sed -i "s/^replicaCount: .*/replicaCount: 2/" values.yaml && sed -i "s/^  tag: .*/  tag: \"alpine\"/" values.yaml'
sol_ch_helm_lab3_4='cat > /root/myapp-chart/templates/deployment.yaml <<YML
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "myapp-chart.fullname" . }}
  labels:
    {{- include "myapp-chart.labels" . | nindent 4 }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      {{- include "myapp-chart.selectorLabels" . | nindent 6 }}
  template:
    metadata:
      labels:
        {{- include "myapp-chart.selectorLabels" . | nindent 8 }}
    spec:
      containers:
      - name: {{ .Chart.Name }}
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag | default .Chart.AppVersion }}"
        imagePullPolicy: {{ .Values.image.pullPolicy }}
        ports:
        - name: http
          containerPort: {{ .Values.service.port }}
          protocol: TCP
YML
helm lint /root/myapp-chart >/dev/null'
sol_ch_helm_lab3_5='helm install my-custom-app /root/myapp-chart --wait --timeout 180s >/dev/null'
sol_ch_helm_lab3_6='printf "configmap:\n  enabled: true\n" >> /root/myapp-chart/values.yaml && cat > /root/myapp-chart/templates/configmap.yaml <<YML
{{- if .Values.configmap.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-configmap
data:
  greeting: hello
{{- end }}
YML'
sol_ch_helm_lab3_7='helm package /root/myapp-chart -d /root >/dev/null'
sol_ch_helm_lab3_8='helm install from-package $(ls /root/myapp-chart-*.tgz | head -1) --wait --timeout 180s >/dev/null'
sol_ch_helm_lab3_9='helm uninstall my-custom-app from-package >/dev/null'

sol_ch_helm_lab4_1='helm install webapp /root/charts/nginx-chart --set image.tag=1.25-alpine --wait --timeout 120s >/dev/null'
sol_ch_helm_lab4_2='helm history webapp >/dev/null'
sol_ch_helm_lab4_3='helm upgrade webapp /root/charts/nginx-chart --set image.tag=alpine --wait --timeout 120s >/dev/null'
sol_ch_helm_lab4_4='helm history webapp >/dev/null'
sol_ch_helm_lab4_5='helm rollback webapp 1 --wait --timeout 120s >/dev/null'
sol_ch_helm_lab4_6='helm get values webapp >/dev/null'
sol_ch_helm_lab4_7='helm upgrade --install new-webapp /root/charts/nginx-chart --wait --timeout 120s >/dev/null'
sol_ch_helm_lab4_8='helm upgrade webapp /root/charts/nginx-chart --set image.tag=alpine --set replicaCount=2 --wait --timeout 180s >/dev/null'
sol_ch_helm_lab4_9='helm history webapp >/dev/null'
sol_ch_helm_lab4_10='helm uninstall webapp new-webapp >/dev/null'

# Template lab: the chart ships a static Deployment, and each step replaces one
# hard-coded value with the template construct the task is about.
sol_ch_helm_lab_templates_1='cd /root/template-lab/webchart && sed -i "s/^  name: webchart$/  name: {{ include \"webchart.fullname\" . }}/" templates/deployment.yaml'
sol_ch_helm_lab_templates_2='cd /root/template-lab/webchart && printf "containerPorts:\n  - name: http\n    containerPort: 9090\n" >> values.yaml && python3 - <<PY
import io
p="templates/deployment.yaml"; s=io.open(p).read()
s=s.replace("        ports:\n        - name: http\n          containerPort: 80\n",
            "        ports:\n        {{- range .Values.containerPorts }}\n        - name: {{ .name }}\n          containerPort: {{ .containerPort }}\n        {{- end }}\n")
io.open(p,"w").write(s)
PY'
sol_ch_helm_lab_templates_3='cd /root/template-lab/webchart && printf "resources:\n  limits:\n    memory: 128Mi\n  requests:\n    memory: 128Mi\n" >> values.yaml && python3 - <<PY
import io
p="templates/deployment.yaml"; s=io.open(p).read()
s=s.replace("        image: \"{{ .Values.image.repository }}:{{ .Values.image.tag }}\"\n",
            "        image: \"{{ .Values.image.repository }}:{{ .Values.image.tag }}\"\n        resources:\n{{ toYaml .Values.resources | nindent 10 }}\n")
io.open(p,"w").write(s)
PY'
sol_ch_helm_lab_templates_4='cd /root/template-lab/webchart && sed -i "s/^  replicas: 1$/  replicas: {{ .Values.replicaCount | default 1 }}/" templates/deployment.yaml'
sol_ch_helm_lab_templates_5='cd /root/template-lab/webchart && printf "serviceAccount:\n  enabled: true\n" >> values.yaml && cat > templates/serviceaccount.yaml <<YML
{{- if .Values.serviceAccount.enabled }}
apiVersion: v1
kind: ServiceAccount
metadata:
  name: {{ include "webchart.fullname" . }}
{{- end }}
YML'
sol_ch_helm_lab_templates_6='cd /root/template-lab/webchart && python3 - <<PY
import io
p="templates/deployment.yaml"; s=io.open(p).read()
s=s.replace("{{ .Values.image.repository }}", "{{ required \"image.repository is required\" .Values.image.repository }}")
io.open(p,"w").write(s)
PY'

# Hooks lab: creating a hook manifest and recording it in the release are separate
# tasks, so the manifest steps and the install/upgrade steps are split.
sol_ch_helm_lab5_1='cat > /root/hooks-chart/templates/pre-install-job.yaml <<YML
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-pre-install
  annotations:
    "helm.sh/hook": pre-install
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: pre-install
        image: busybox:1.28
        command: ["sh","-c","echo pre-install"]
YML'
sol_ch_helm_lab5_2='helm install hooked-app /root/hooks-chart --wait --timeout 180s >/dev/null'
sol_ch_helm_lab5_3='cat > /root/hooks-chart/templates/post-upgrade-job.yaml <<YML
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-post-upgrade
  annotations:
    "helm.sh/hook": post-upgrade
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: post-upgrade
        image: busybox:1.28
        command: ["sh","-c","echo post-upgrade"]
YML'
sol_ch_helm_lab5_4='helm upgrade hooked-app /root/hooks-chart --wait --timeout 180s >/dev/null'
sol_ch_helm_lab5_5='cd /root/hooks-chart/templates && for n in early:-10 late:10; do w=${n#*:}; k=${n%%:*}; cat > pre-upgrade-$k-job.yaml <<YML
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-pre-upgrade-$k
  annotations:
    "helm.sh/hook": pre-upgrade
    "helm.sh/hook-weight": "$w"
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: pre-upgrade-$k
        image: busybox:1.28
        command: ["sh","-c","echo pre-upgrade-$k"]
YML
done
helm upgrade hooked-app /root/hooks-chart --wait --timeout 180s >/dev/null'
sol_ch_helm_lab5_6='cat > /root/hooks-chart/templates/pre-delete-job.yaml <<YML
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Release.Name }}-pre-delete
  annotations:
    "helm.sh/hook": pre-delete
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: pre-delete
        image: busybox:1.28
        command: ["sh","-c","echo pre-delete"]
YML
helm upgrade hooked-app /root/hooks-chart --wait --timeout 180s >/dev/null'
sol_ch_helm_lab5_7='helm uninstall hooked-app >/dev/null 2>&1; for i in $(seq 1 60); do kubectl get jobs -o name 2>/dev/null | grep -qi pre-delete && break; sleep 1; done'

# Bitnami repo lab: update the index, pull the chart two ways, then install it
# from the repo with the image overridden to the familiar nginx:alpine.
sol_ch_helm_lab_repos_1='helm repo update bitnami >/dev/null'
sol_ch_helm_lab_repos_2='mkdir -p /root/bitnami-charts && helm pull bitnami/nginx --destination /root/bitnami-charts >/dev/null'
sol_ch_helm_lab_repos_3='mkdir -p /root/bitnami-unpacked && helm pull bitnami/nginx --untar --untardir /root/bitnami-unpacked >/dev/null'
sol_ch_helm_lab_repos_4='helm install bitnami-nginx bitnami/nginx --set image.registry=docker.io --set image.repository=nginx --set image.tag=alpine --set service.type=ClusterIP --set containerPorts.http=80 --set podSecurityContext.enabled=false --set containerSecurityContext.enabled=false --set containerSecurityContext.readOnlyRootFilesystem=false --set pdb.create=false --set networkPolicy.enabled=false --wait --timeout 180s >/dev/null'

# Helmfile lab: write the two values files and the helmfile, then sync/change/destroy.
# NOTE: needs the `helmfile` binary — present in the production k8s golden, absent
# from a locally built golearn/sandbox-k8s, so these fail locally until it is rebuilt.
sol_ch_helm_lab_helmfile_1='mkdir -p /root/helmfile-lab/values && cat > /root/helmfile-lab/values/dev.yaml <<YML
replicaCount: 1
environment: dev
image:
  repository: nginx
  tag: alpine
service:
  type: ClusterIP
  port: 8080
  targetPort: 80
YML
cat > /root/helmfile-lab/values/staging.yaml <<YML
replicaCount: 2
environment: staging
image:
  repository: nginx
  tag: alpine
service:
  type: ClusterIP
  port: 8080
  targetPort: 80
YML'
sol_ch_helm_lab_helmfile_2='cat > /root/helmfile-lab/helmfile.yaml.gotmpl <<YML
environments:
  dev:
    values:
      - values/dev.yaml
  staging:
    values:
      - values/staging.yaml
---
releases:
  - name: web-{{ .Environment.Name }}
    chart: ./charts/webapp
    values:
      - values/{{ .Environment.Name }}.yaml
YML'
sol_ch_helm_lab_helmfile_3='cd /root/helmfile-lab && helmfile -f helmfile.yaml.gotmpl -e dev sync >/dev/null'
sol_ch_helm_lab_helmfile_4='cd /root/helmfile-lab && helmfile -f helmfile.yaml.gotmpl -e staging sync >/dev/null'
sol_ch_helm_lab_helmfile_5='cd /root/helmfile-lab && sed -i "s/^replicaCount: 1$/replicaCount: 2/" values/dev.yaml && helmfile -f helmfile.yaml.gotmpl -e dev sync >/dev/null'
sol_ch_helm_lab_helmfile_6='cd /root/helmfile-lab && helmfile -f helmfile.yaml.gotmpl -e dev destroy >/dev/null 2>&1; helmfile -f helmfile.yaml.gotmpl -e staging destroy >/dev/null 2>&1; true'

# ── CKAD debug labs ──────────────────────────────────────────────────────────
# Each lab ships deliberately broken manifests in /root/debug-*; the reference
# solution fixes the one root cause named in the task and re-applies.

# Lab 13: broken startup — bad image tag, missing env, wrong init mountPath, wrong probe path.
sol_ch_ckad_lab13_debug_startup_1='sed -i "s/nginx:alpne/nginx:alpine/" /root/debug-workloads/app-a.yaml && kubectl delete pod frontend-api --ignore-not-found >/dev/null 2>&1 && kubectl apply -f /root/debug-workloads/app-a.yaml >/dev/null && kubectl wait --for=condition=Ready pod/frontend-api --timeout=180s >/dev/null'
sol_ch_ckad_lab13_debug_startup_2='printf "    env:\n    - {name: DATABASE_URL, value: \"postgres://db:5432/app\"}\n" >> /root/debug-workloads/app-b.yaml && kubectl delete pod worker-api --ignore-not-found >/dev/null 2>&1 && kubectl apply -f /root/debug-workloads/app-b.yaml >/dev/null && kubectl wait --for=condition=Ready pod/worker-api --timeout=180s >/dev/null'
sol_ch_ckad_lab13_debug_startup_3='sed -i "s#mountPath: /etc/wrong#mountPath: /config#" /root/debug-workloads/app-c.yaml && kubectl delete pod billing-api --ignore-not-found >/dev/null 2>&1 && kubectl apply -f /root/debug-workloads/app-c.yaml >/dev/null && kubectl wait --for=condition=Ready pod/billing-api --timeout=180s >/dev/null'
sol_ch_ckad_lab13_debug_startup_4='sed -i "s#path: /readyz#path: /healthz#" /root/debug-workloads/app-d.yaml && kubectl apply -f /root/debug-workloads/app-d.yaml >/dev/null && kubectl rollout status deploy/catalog-api --timeout=180s >/dev/null'
sol_ch_ckad_lab13_debug_startup_5='kubectl delete pod frontend-api worker-api billing-api --ignore-not-found >/dev/null 2>&1; kubectl delete deploy catalog-api --ignore-not-found >/dev/null 2>&1; kubectl delete configmap billing-config --ignore-not-found >/dev/null 2>&1; true'

# Lab 14: broken routing/config — Service selector, ConfigMap key, targetPort, missing ConfigMap.
# Task 3 is observational (no check), so no solution is needed for it.
sol_ch_ckad_lab14_debug_service_config_1='sed -i "s/app: orders-frontend/app: orders-api/" /root/debug-routing/app-a.yaml && kubectl apply -f /root/debug-routing/app-a.yaml >/dev/null && kubectl rollout status deploy/orders-api --timeout=180s >/dev/null && for i in $(seq 1 60); do [ -n "$(kubectl get endpoints orders-svc -o jsonpath="{.subsets[0].addresses[0].ip}" 2>/dev/null)" ] && break; sleep 1; done'
sol_ch_ckad_lab14_debug_service_config_2='sed -i "s/key: mode,/key: app_mode,/" /root/debug-routing/app-a.yaml && kubectl apply -f /root/debug-routing/app-a.yaml >/dev/null && kubectl rollout status deploy/orders-api --timeout=180s >/dev/null'
# Наблюдательное задание: студент просто обращается к сервису. Решение-«no-op»
# нужно, чтобы харнесс выполнил проверку и выдал OBS, а не SKIP.
sol_ch_ckad_lab14_debug_service_config_3='kubectl exec deploy/orders-api -- wget -qO- http://orders-svc >/dev/null 2>&1 || true'
sol_ch_ckad_lab14_debug_service_config_4='sed -i "s/targetPort: 8080/targetPort: 80/" /root/debug-routing/app-b.yaml && kubectl apply -f /root/debug-routing/app-b.yaml >/dev/null && kubectl rollout status deploy/reports-api --timeout=180s >/dev/null'
sol_ch_ckad_lab14_debug_service_config_5='kubectl create configmap profile-config --from-literal=app_env=prod >/dev/null 2>&1; kubectl apply -f /root/debug-routing/app-c.yaml >/dev/null && kubectl rollout status deploy/profile-api --timeout=180s >/dev/null'
sol_ch_ckad_lab14_debug_service_config_6='kubectl delete deploy orders-api reports-api profile-api --ignore-not-found >/dev/null 2>&1; kubectl delete svc orders-svc reports-svc profile-svc --ignore-not-found >/dev/null 2>&1; kubectl delete configmap orders-config profile-config --ignore-not-found >/dev/null 2>&1; kubectl delete pod debug-client --ignore-not-found >/dev/null 2>&1; true'
