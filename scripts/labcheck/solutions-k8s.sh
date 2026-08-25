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
sol_ch_k8si_lab3_4='kubectl get svc >/dev/null'
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
sol_ch_k8si_lab4_3='kubectl logs cm-pod >/dev/null 2>&1'
sol_ch_k8si_lab4_4='kubectl create secret generic db-secret --from-literal=DB_USER=admin --from-literal=DB_PASSWORD=supersecret123 >/dev/null'
sol_ch_k8si_lab4_5='kubectl get secret db-secret -o jsonpath="{.data.DB_PASSWORD}" | base64 -d >/dev/null'
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
