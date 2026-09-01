package main

// Fixtures + auto-checks for "Kubernetes: основы" (k8s-intro).
//
// These lessons run in golearn/sandbox-k8s: a privileged lab container with a
// single-node k3s cluster. `k8s-start` boots it, imports the images the lessons
// deploy and brings up an offline registry mirror, so the cluster works without
// any network.

// k8sBoot waits for the cluster at the beginning of every lesson setup. In the
// Firecracker k8s golden, k3s auto-starts via systemd (k3s.service) and KUBECONFIG
// is exported globally, so the setup only has to wait until the API is ready.
const k8sBoot = `set -e
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
for i in $(seq 1 60); do kubectl get --raw=/readyz >/dev/null 2>&1 && break; sleep 1; done`

// kcheck fails with a clear message while the cluster is still coming up.
func kcheck(cond, good, bad string) string {
	return `export KUBECONFIG=/etc/rancher/k3s/k3s.yaml; ` +
		`if ! kubectl get --raw=/readyz >/dev/null 2>&1; then ` +
		fail("Кластер ещё поднимается. Подожди 10–20 секунд и нажми «Проверить» снова.") +
		`; fi; ` + check(cond, good, bad)
}

// jp runs a kubectl JSONPath query and compares the result with want.
func jp(args, path, want string) string {
	return `[ "$(kubectl ` + args + ` -o jsonpath='` + path + `' 2>/dev/null)" = ` + want + ` ]`
}

var k8sIntroLabs = map[string]labSpec{
	// ── Lab 1: первый Pod ──
	"ch-k8si-lab1": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod nginx-pod --ignore-not-found >/dev/null 2>&1
rm -f /root/nginx_pod_ip.txt`,
		Checks: map[int]string{
			1: kcheck(jp("get pod nginx-pod", "{.spec.containers[0].image}", "nginx:alpine")+` && `+
				jp("get pod nginx-pod", "{.metadata.labels.run}", "nginx-pod"),
				"Pod nginx-pod создан с нужным образом и label",
				"kubectl run nginx-pod --image=nginx:alpine (kubectl run сам ставит label run=<имя>)"),
			2: kcheck(jp("get pod nginx-pod", "{.status.phase}", "Running"),
				"Pod в статусе Running",
				"Подожди и посмотри: kubectl get pod nginx-pod -w (если Pending — kubectl describe pod nginx-pod покажет причину)"),
			3: kcheck(`[ -s /root/nginx_pod_ip.txt ] && [ "$(tr -d ' \n' < /root/nginx_pod_ip.txt)" = "$(kubectl get pod nginx-pod -o jsonpath='{.status.podIP}' 2>/dev/null)" ]`,
				"IP пода сохранён верно",
				"kubectl get pod nginx-pod -o jsonpath='{.status.podIP}' > /root/nginx_pod_ip.txt"),
			4: kcheck(`! kubectl get pod nginx-pod >/dev/null 2>&1`,
				"Pod удалён",
				"kubectl delete pod nginx-pod"),
		},
	},

	// ── Lab 2: Deployment ──
	"ch-k8si-lab2": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment nginx-deploy --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy nginx-deploy", "{.spec.replicas}", "2")+` && `+
				jp("get deploy nginx-deploy", "{.spec.template.spec.containers[0].image}", "nginx:alpine"),
				"Deployment создан с 2 репликами",
				"kubectl create deployment nginx-deploy --image=nginx:alpine --replicas=2"),
			2: kcheck(jp("get deploy nginx-deploy", "{.status.readyReplicas}", "2"),
				"обе реплики готовы",
				"kubectl rollout status deployment/nginx-deploy"),
			3: kcheck(jp("get deploy nginx-deploy", "{.spec.replicas}", "1"),
				"Deployment масштабирован до 1 реплики",
				"kubectl scale deployment nginx-deploy --replicas=1"),
			4: kcheck(jp("get deploy nginx-deploy", "{.spec.strategy.type}", "RollingUpdate")+` && `+
				jp("get deploy nginx-deploy", "{.spec.strategy.rollingUpdate.maxSurge}", "1")+` && `+
				jp("get deploy nginx-deploy", "{.spec.strategy.rollingUpdate.maxUnavailable}", "0"),
				"стратегия RollingUpdate настроена",
				`kubectl patch deployment nginx-deploy -p '{"spec":{"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxSurge":1,"maxUnavailable":0}}}}'`),
			5: kcheck(jp("get deploy nginx-deploy", "{.spec.template.spec.containers[0].image}", "nginx:1.25-alpine")+` && `+
				`[ "$(kubectl get deploy nginx-deploy -o jsonpath='{.status.readyReplicas}' 2>/dev/null)" -ge 1 ]`,
				"образ обновлён, rollout завершён",
				"kubectl set image deployment/nginx-deploy nginx=nginx:1.25-alpine && kubectl rollout status deployment/nginx-deploy"),
			6: kcheck(jp("get deploy nginx-deploy", "{.spec.template.spec.containers[0].image}", "nginx:alpine")+` && `+
				`[ "$(kubectl get deploy nginx-deploy -o jsonpath='{.status.readyReplicas}' 2>/dev/null)" -ge 1 ]`,
				"откат выполнен",
				"kubectl rollout undo deployment/nginx-deploy && kubectl rollout status deployment/nginx-deploy"),
			7: kcheck(`! kubectl get deploy nginx-deploy >/dev/null 2>&1 && `+
				`for i in $(seq 1 15); do [ -z "$(kubectl get pods -l app=nginx-deploy --no-headers 2>/dev/null)" ] && break; sleep 1; done; `+
				`[ -z "$(kubectl get pods -l app=nginx-deploy --no-headers 2>/dev/null)" ]`,
				"Deployment и его Pod'ы удалены",
				"kubectl delete deployment nginx-deploy"),
		},
	},

	// ── Lab 2m: декларативные манифесты ──
	"ch-k8si-lab2-manifests": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment manifest-web broken-web --ignore-not-found >/dev/null 2>&1
rm -rf /root/manifest-deploy && mkdir -p /root/manifest-deploy
cat > /root/manifest-deploy/deployment.yaml <<'YEOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: manifest-web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: manifest-web
  template:
    metadata:
      labels:
        app: manifest-web
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: 80
YEOF
cat > /root/manifest-deploy/broken-deployment.yaml <<'YEOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: broken-web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: broken-web
  template:
    metadata:
      labels:
        app: wrong-label
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: "80"
YEOF`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy manifest-web", "{.status.readyReplicas}", "2"),
				"манифест применён, 2 реплики готовы",
				"kubectl apply -f /root/manifest-deploy/deployment.yaml && kubectl rollout status deployment/manifest-web"),
			2: kcheck(jp("get deploy manifest-web", "{.spec.replicas}", "3")+` && `+
				jp("get deploy manifest-web", "{.status.readyReplicas}", "3")+` && `+
				`grep -qE '^\s*replicas:\s*3' /root/manifest-deploy/deployment.yaml`,
				"replicas изменено в файле и применено",
				"Поправь spec.replicas на 3 в /root/manifest-deploy/deployment.yaml и снова kubectl apply -f …"),
			3: kcheck(jp("get deploy manifest-web", "{.spec.template.spec.containers[0].image}", "nginx:1.25-alpine")+` && `+
				`grep -q 'nginx:1.25-alpine' /root/manifest-deploy/deployment.yaml`,
				"образ обновлён через манифест",
				"Замени image на nginx:1.25-alpine в файле и примени: kubectl apply -f /root/manifest-deploy/deployment.yaml"),
			4: kcheck(`[ "$(kubectl get deploy broken-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)" -ge 1 ]`,
				"сломанный манифест исправлен и задеплоен",
				"В broken-deployment.yaml две ошибки: label в template не совпадает с selector (app=wrong-label вместо app=broken-web) и containerPort указан строкой \"80\" вместо числа 80"),
			5: kcheck(`! kubectl get deploy manifest-web >/dev/null 2>&1 && ! kubectl get deploy broken-web >/dev/null 2>&1`,
				"оба Deployment удалены",
				"kubectl delete deployment manifest-web broken-web"),
		},
	},

	// ── Lab 2r: история и откаты ──
	"ch-k8si-lab2-rollout": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment rollout-app --ignore-not-found >/dev/null 2>&1
rm -f /root/rollout_history.txt /root/revision1.txt`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy rollout-app", "{.spec.replicas}", "3")+` && `+
				`kubectl get deploy rollout-app -o jsonpath='{.metadata.annotations.kubernetes\.io/change-cause}' 2>/dev/null | grep -q 'Initial deployment'`,
				"Deployment создан с аннотацией change-cause",
				"kubectl create deployment rollout-app --image=nginx:alpine --replicas=3 затем kubectl annotate deployment rollout-app kubernetes.io/change-cause='Initial deployment'"),
			2: kcheck(jp("get deploy rollout-app", "{.spec.strategy.rollingUpdate.maxSurge}", "1")+` && `+
				jp("get deploy rollout-app", "{.spec.strategy.rollingUpdate.maxUnavailable}", "0"),
				"стратегия обновления настроена",
				`kubectl patch deployment rollout-app -p '{"spec":{"strategy":{"type":"RollingUpdate","rollingUpdate":{"maxSurge":1,"maxUnavailable":0}}}}'`),
			3: kcheck(jp("get deploy rollout-app", "{.spec.template.spec.containers[0].image}", "nginx:1.25-alpine")+` && `+
				`kubectl get deploy rollout-app -o jsonpath='{.metadata.annotations.kubernetes\.io/change-cause}' 2>/dev/null | grep -qi 'nginx'`,
				"образ обновлён и ревизия подписана",
				"kubectl set image deployment/rollout-app nginx=nginx:1.25-alpine, дождись rollout status и подпиши: kubectl annotate deployment rollout-app kubernetes.io/change-cause='Update nginx to 1.25' --overwrite"),
			4: kcheck(`grep -qi 'REVISION' /root/rollout_history.txt 2>/dev/null && [ "$(grep -cE '^[0-9]+' /root/rollout_history.txt)" -ge 2 ]`,
				"история rollout сохранена",
				"kubectl rollout history deployment/rollout-app > /root/rollout_history.txt (в файле должно быть минимум две ревизии)"),
			5: kcheck(jp("get deploy rollout-app", "{.spec.template.spec.containers[0].image}", "nginx:alpine")+` && `+
				`kubectl get deploy rollout-app -o jsonpath='{.metadata.annotations.kubernetes\.io/change-cause}' 2>/dev/null | grep -qi 'roll'`,
				"откат выполнен и подписан",
				"kubectl rollout undo deployment/rollout-app, дождись rollout status, затем kubectl annotate deployment rollout-app kubernetes.io/change-cause='Rollback to nginx:alpine' --overwrite"),
			6: kcheck(`grep -qi 'REVISION' /root/revision1.txt 2>/dev/null && `+
				`[ "$(kubectl get deploy rollout-app -o jsonpath='{.status.readyReplicas}' 2>/dev/null)" -ge 1 ]`,
				"история после отката сохранена, Deployment готов",
				"kubectl rollout history deployment/rollout-app > /root/revision1.txt"),
			7: kcheck(`! kubectl get deploy rollout-app >/dev/null 2>&1`,
				"Deployment удалён",
				"kubectl delete deployment rollout-app"),
		},
	},

	// ── Lab 7: labels ──
	"ch-k8si-lab7": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod web-prod web-staging db-prod --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get pod web-prod", "{.metadata.labels.app}", "web")+` && `+
				jp("get pod web-prod", "{.metadata.labels.env}", "prod")+` && `+
				jp("get pod web-staging", "{.metadata.labels.env}", "staging")+` && `+
				jp("get pod db-prod", "{.metadata.labels.app}", "db"),
				"три Pod'а созданы с нужными labels",
				"kubectl run web-prod --image=nginx:alpine --labels=app=web,env=prod (аналогично web-staging и db-prod)"),
			2: kcheck(jp("get pod web-prod", "{.metadata.labels.monitored}", "true"),
				"label monitored=true добавлен",
				"kubectl label pod web-prod monitored=true"),
			3: kcheck(`[ -z "$(kubectl get pod web-prod -o jsonpath='{.metadata.labels.monitored}' 2>/dev/null)" ]`,
				"label monitored удалён",
				"kubectl label pod web-prod monitored-"),
			4: kcheck(`! kubectl get pod web-prod >/dev/null 2>&1 && ! kubectl get pod web-staging >/dev/null 2>&1 && ! kubectl get pod db-prod >/dev/null 2>&1`,
				"тестовые Pod'ы удалены",
				"kubectl delete pod web-prod web-staging db-prod"),
		},
	},

	// ── Lab 3: Services ──
	"ch-k8si-lab3": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment web-app --ignore-not-found >/dev/null 2>&1
kubectl delete svc web-svc web-nodeport --ignore-not-found >/dev/null 2>&1
rm -f /root/web-nodeport.yaml`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy web-app", "{.status.readyReplicas}", "2"),
				"Deployment web-app готов",
				"kubectl create deployment web-app --image=nginx:alpine --replicas=2"),
			2: kcheck(jp("get svc web-svc", "{.spec.type}", "ClusterIP")+` && `+
				jp("get svc web-svc", "{.spec.selector.app}", "web-app")+` && `+
				jp("get svc web-svc", "{.spec.ports[0].port}", "80"),
				"Service web-svc типа ClusterIP создан",
				"kubectl expose deployment web-app --name=web-svc --port=80 --target-port=80"),
			3: kcheck(jp("get svc web-nodeport", "{.spec.type}", "NodePort")+` && `+
				`grep -q 'NodePort' /root/web-nodeport.yaml 2>/dev/null`,
				"Service типа NodePort создан и сохранён в файл",
				"kubectl expose deployment web-app --name=web-nodeport --type=NodePort --port=80 --target-port=80 -o yaml --dry-run=client > /root/web-nodeport.yaml, затем примени файл"),
			4: kcheck(`kubectl get svc web-svc >/dev/null 2>&1 && kubectl get svc web-nodeport >/dev/null 2>&1`,
				"оба сервиса на месте",
				"kubectl get svc — в списке должны быть web-svc и web-nodeport"),
			5: kcheck(`! kubectl get svc web-svc >/dev/null 2>&1 && ! kubectl get svc web-nodeport >/dev/null 2>&1`,
				"оба сервиса удалены",
				"kubectl delete svc web-svc web-nodeport"),
		},
	},

	// ── Lab 8: Service через манифесты ──
	"ch-k8si-lab8-service-manifests": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment store-web --ignore-not-found >/dev/null 2>&1
kubectl delete svc store-svc --ignore-not-found >/dev/null 2>&1
kubectl delete pod store-client --ignore-not-found >/dev/null 2>&1
rm -rf /root/service-manifests && mkdir -p /root/service-manifests
cat > /root/service-manifests/deployment.yaml <<'YEOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: store-web
spec:
  replicas: 2
  selector:
    matchLabels:
      app: store
  template:
    metadata:
      labels:
        app: store
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          ports:
            - containerPort: 80
YEOF
cat > /root/service-manifests/service.yaml <<'YEOF'
apiVersion: v1
kind: Service
metadata:
  name: store-svc
spec:
  selector:
    app: shop
  ports:
    - port: 80
      targetPort: 8080
YEOF
rm -f /root/store_http.txt`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy store-web", "{.status.readyReplicas}", "2")+` && kubectl get svc store-svc >/dev/null 2>&1`,
				"манифесты применены (Endpoints пока пустые — так и задумано)",
				"kubectl apply -f /root/service-manifests/deployment.yaml -f /root/service-manifests/service.yaml"),
			2: kcheck(jp("get svc store-svc", "{.spec.selector.app}", "store")+` && `+
				`[ -n "$(kubectl get endpoints store-svc -o jsonpath='{.subsets[0].addresses[0].ip}' 2>/dev/null)" ]`,
				"selector исправлен, Endpoints появились",
				"В /root/service-manifests/service.yaml замени selector app: shop на app: store и примени файл заново"),
			3: kcheck(jp("get svc store-svc", "{.spec.ports[0].targetPort}", "80")+` && `+
				`grep -qE 'targetPort:\s*80' /root/service-manifests/service.yaml`,
				"targetPort исправлен на 80",
				"В service.yaml поставь targetPort: 80 и примени файл"),
			4: kcheck(`grep -q 'Welcome to nginx' /root/store_http.txt 2>/dev/null`,
				"ответ сервиса получен по DNS-имени",
				"kubectl run store-client --image=busybox:1.28 --restart=Never --command -- sleep 3600, затем kubectl exec store-client -- wget -qO- http://store-svc > /root/store_http.txt"),
			5: kcheck(`[ "$(kubectl get endpoints store-svc -o jsonpath='{.subsets[0].addresses[*].ip}' 2>/dev/null | wc -w)" = 3 ]`,
				"после масштабирования 3 Endpoints",
				"Поставь replicas: 3 в deployment.yaml, примени и проверь: kubectl get endpoints store-svc"),
			6: kcheck(`! kubectl get svc store-svc >/dev/null 2>&1 && ! kubectl get deploy store-web >/dev/null 2>&1 && ! kubectl get pod store-client >/dev/null 2>&1`,
				"ресурсы лаборатории удалены",
				"kubectl delete svc store-svc; kubectl delete deployment store-web; kubectl delete pod store-client --ignore-not-found"),
		},
	},

	// ── Lab: Headless Service и StatefulSet ──
	"ch-k8si-lab-headless-stateful": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete statefulset web-sts --ignore-not-found >/dev/null 2>&1
kubectl delete deployment web-deploy --ignore-not-found >/dev/null 2>&1
kubectl delete svc web-svc web-headless --ignore-not-found >/dev/null 2>&1
rm -rf /root/headless-lab && mkdir -p /root/headless-lab
cat > /root/headless-lab/web-deploy.yaml <<'YEOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-deploy
spec:
  replicas: 2
  selector:
    matchLabels:
      app: web
  template:
    metadata:
      labels:
        app: web
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
---
apiVersion: v1
kind: Service
metadata:
  name: web-svc
spec:
  selector:
    app: web
  ports:
    - port: 80
YEOF
cat > /root/headless-lab/web-sts.yaml <<'YEOF'
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: web-sts
spec:
  serviceName: web-headless
  replicas: 3
  selector:
    matchLabels:
      app: web-sts
  template:
    metadata:
      labels:
        app: web-sts
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
YEOF
cat > /root/headless-lab/web-headless.yaml <<'YEOF'
apiVersion: v1
kind: Service
metadata:
  name: web-headless
spec:
  clusterIP: None
  selector:
    app: web-sts
  ports:
    - port: 80
YEOF
cat > /root/headless-lab/web-headless-broken.yaml <<'YEOF'
apiVersion: v1
kind: Service
metadata:
  name: web-headless
spec:
  clusterIP: None
  selector:
    app: wrong-sts
  ports:
    - port: 80
YEOF`,
		Descs: map[int]string{
			// Экспорт курса потерял текст этого задания (в базе осталось "$26").
			1: `<p>Разверни обычную пару «Deployment + Service» и StatefulSet, с которыми будешь дальше сравнивать поведение Headless Service.</p>
<p>Готовые манифесты уже лежат в <code>/root/headless-lab/</code>:</p>
<pre><code>kubectl apply -f /root/headless-lab/web-deploy.yaml
kubectl apply -f /root/headless-lab/web-sts.yaml</code></pre>
<p>Дождись, пока <code>web-deploy</code> поднимет 2 Ready-реплики, а StatefulSet <code>web-sts</code> — 3 Pod'а с именами <code>web-sts-0</code>, <code>web-sts-1</code>, <code>web-sts-2</code>.</p>`,
		},
		Checks: map[int]string{
			1: kcheck(jp("get deploy web-deploy", "{.status.readyReplicas}", "2")+` && `+
				jp("get statefulset web-sts", "{.status.readyReplicas}", "3")+` && `+
				`kubectl get pod web-sts-0 >/dev/null 2>&1`,
				"Deployment и StatefulSet подняты",
				"kubectl apply -f /root/headless-lab/web-deploy.yaml -f /root/headless-lab/web-sts.yaml — затем дождись Ready"),
			2: kcheck(jp("get svc web-headless", "{.spec.clusterIP}", "None")+` && `+
				`[ "$(kubectl get endpoints web-headless -o jsonpath='{.subsets[0].addresses[*].ip}' 2>/dev/null | wc -w)" -ge 3 ]`,
				"Headless Service создан, Endpoints указывают на Pod'ы",
				"kubectl apply -f /root/headless-lab/web-headless.yaml"),
			3: kcheck(jp("get svc web-headless", "{.spec.selector.app}", "wrong-sts")+` && `+
				`[ -z "$(kubectl get endpoints web-headless -o jsonpath='{.subsets[0].addresses[0].ip}' 2>/dev/null)" ]`,
				"со сломанным selector Endpoints исчезли — эффект воспроизведён",
				"kubectl apply -f /root/headless-lab/web-headless-broken.yaml, затем kubectl get endpoints web-headless"),
			4: kcheck(jp("get statefulset web-sts", "{.status.readyReplicas}", "5"),
				"StatefulSet масштабирован до 5 Pod'ов",
				"kubectl scale statefulset web-sts --replicas=5 — Pod'ы создаются по очереди, дождись Ready"),
			5: kcheck(`! kubectl get statefulset web-sts >/dev/null 2>&1 && ! kubectl get svc web-headless >/dev/null 2>&1 && `+
				`! kubectl get deploy web-deploy >/dev/null 2>&1 && ! kubectl get svc web-svc >/dev/null 2>&1`,
				"все ресурсы лаборатории удалены",
				"kubectl delete statefulset web-sts; kubectl delete svc web-headless web-svc; kubectl delete deployment web-deploy"),
		},
	},

	// ── Lab 4: ConfigMap и Secret ──
	"ch-k8si-lab4": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete configmap app-config file-config --ignore-not-found >/dev/null 2>&1
kubectl delete secret db-secret --ignore-not-found >/dev/null 2>&1
kubectl delete pod cm-pod --ignore-not-found >/dev/null 2>&1
rm -f /root/app.conf
cat > /root/cm-pod.yaml <<'YEOF'
apiVersion: v1
kind: Pod
metadata:
  name: cm-pod
  labels:
    run: cm-pod
spec:
  restartPolicy: Never
  containers:
    - name: app
      image: alpine:latest
      command: ["sh", "-c", "echo APP_ENV=$APP_ENV; sleep 3600"]
YEOF`,
		Checks: map[int]string{
			1: kcheck(jp("get configmap app-config", "{.data.APP_ENV}", "production")+` && `+
				jp("get configmap app-config", "{.data.LOG_LEVEL}", "info")+` && `+
				jp("get configmap app-config", "{.data.PORT}", "8080"),
				"ConfigMap app-config создан",
				"kubectl create configmap app-config --from-literal=APP_ENV=production --from-literal=LOG_LEVEL=info --from-literal=PORT=8080"),
			2: kcheck(`grep -q 'envFrom' /root/cm-pod.yaml && grep -q 'app-config' /root/cm-pod.yaml && kubectl get pod cm-pod >/dev/null 2>&1`,
				"Pod подключает ConfigMap через envFrom",
				"Добавь контейнеру в /root/cm-pod.yaml: envFrom: - configMapRef: name: app-config — затем kubectl apply -f /root/cm-pod.yaml"),
			3: kcheck(`kubectl logs cm-pod 2>/dev/null | grep -q 'APP_ENV=production'`,
				"переменная из ConfigMap попала в Pod",
				"kubectl logs cm-pod — в выводе должно быть APP_ENV=production"),
			4: kcheck(`[ "$(kubectl get secret db-secret -o jsonpath='{.data.DB_USER}' 2>/dev/null | base64 -d)" = admin ] && `+
				`[ "$(kubectl get secret db-secret -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d)" = supersecret123 ]`,
				"Secret db-secret создан",
				"kubectl create secret generic db-secret --from-literal=DB_USER=admin --from-literal=DB_PASSWORD=supersecret123"),
			5: kcheck(`[ "$(kubectl get secret db-secret -o jsonpath='{.data.DB_PASSWORD}' 2>/dev/null | base64 -d)" = supersecret123 ]`,
				"значение Secret декодируется",
				"kubectl get secret db-secret -o jsonpath='{.data.DB_PASSWORD}' | base64 -d"),
			6: kcheck(`grep -q 'server.port=8080' /root/app.conf 2>/dev/null && grep -q 'log.level=debug' /root/app.conf && grep -q 'db.pool.size=10' /root/app.conf && `+
				`kubectl get configmap file-config -o jsonpath='{.data.app\.conf}' 2>/dev/null | grep -q 'server.port=8080'`,
				"ConfigMap создан из файла",
				"Создай /root/app.conf с тремя строками, затем kubectl create configmap file-config --from-file=/root/app.conf"),
			7: kcheck(`! kubectl get configmap app-config >/dev/null 2>&1 && ! kubectl get configmap file-config >/dev/null 2>&1 && `+
				`! kubectl get secret db-secret >/dev/null 2>&1 && ! kubectl get pod cm-pod >/dev/null 2>&1`,
				"ресурсы удалены",
				"kubectl delete configmap app-config file-config; kubectl delete secret db-secret; kubectl delete pod cm-pod"),
		},
	},

	// ── Lab 4m: конфигурация через манифесты ──
	"ch-k8si-lab4-manifests": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod manifest-config-app secret-broken-app --ignore-not-found >/dev/null 2>&1
kubectl delete configmap manifest-config --ignore-not-found >/dev/null 2>&1
kubectl delete secret manifest-secret --ignore-not-found >/dev/null 2>&1
rm -rf /root/config-manifests && mkdir -p /root/config-manifests
cat > /root/config-manifests/app-config.yaml <<'YEOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: manifest-config
data:
  APP_ENV: staging
  FEATURE_FLAG: disabled
YEOF
cat > /root/config-manifests/app-secret.yaml <<'YEOF'
apiVersion: v1
kind: Secret
metadata:
  name: manifest-secret
type: Opaque
stringData:
  API_TOKEN: token-12345
YEOF
cat > /root/config-manifests/app-pod.yaml <<'YEOF'
apiVersion: v1
kind: Pod
metadata:
  name: manifest-config-app
spec:
  restartPolicy: Never
  containers:
    - name: app
      image: alpine:latest
      command: ["sh", "-c", "env | sort; sleep 3600"]
      envFrom:
        - configMapRef:
            name: manifest-config
YEOF
cat > /root/config-manifests/broken-secret-pod.yaml <<'YEOF'
apiVersion: v1
kind: Pod
metadata:
  name: secret-broken-app
spec:
  restartPolicy: Never
  containers:
    - name: app
      image: alpine:latest
      command: ["sh", "-c", "echo TOKEN=$API_TOKEN; sleep 3600"]
      env:
        - name: API_TOKEN
          valueFrom:
            secretKeyRef:
              name: wrong-secret
              key: API_TOKEN
YEOF`,
		Checks: map[int]string{
			1: kcheck(`kubectl get configmap manifest-config >/dev/null 2>&1 && kubectl get secret manifest-secret >/dev/null 2>&1`,
				"ConfigMap и Secret применены",
				"kubectl apply -f /root/config-manifests/app-config.yaml -f /root/config-manifests/app-secret.yaml"),
			2: kcheck(jp("get pod manifest-config-app", "{.status.phase}", "Running"),
				"Pod запущен",
				"kubectl apply -f /root/config-manifests/app-pod.yaml"),
			3: kcheck(jp("get configmap manifest-config", "{.data.APP_ENV}", "production")+` && `+
				jp("get configmap manifest-config", "{.data.FEATURE_FLAG}", "enabled"),
				"ConfigMap обновлён",
				"Поправь значения в /root/config-manifests/app-config.yaml (APP_ENV: production, FEATURE_FLAG: enabled) и примени файл"),
			4: kcheck(`kubectl exec manifest-config-app -- env 2>/dev/null | grep -q '^APP_ENV=production$'`,
				"пересозданный Pod видит новые значения",
				"Pod читает ConfigMap только при старте: kubectl delete pod manifest-config-app && kubectl apply -f /root/config-manifests/app-pod.yaml, затем kubectl exec manifest-config-app -- env | grep APP_ENV"),
			5: kcheck(jp("get pod secret-broken-app", "{.status.phase}", "Running")+` && `+
				`kubectl logs secret-broken-app 2>/dev/null | grep -q 'TOKEN=token-12345'`,
				"манифест исправлен, Pod получил значение из Secret",
				"Pod ссылается на несуществующий Secret wrong-secret — замени имя на manifest-secret, удали Pod и примени манифест снова"),
			6: kcheck(`! kubectl get pod manifest-config-app >/dev/null 2>&1 && ! kubectl get pod secret-broken-app >/dev/null 2>&1 && `+
				`! kubectl get configmap manifest-config >/dev/null 2>&1 && ! kubectl get secret manifest-secret >/dev/null 2>&1`,
				"ресурсы удалены",
				"kubectl delete pod manifest-config-app secret-broken-app; kubectl delete configmap manifest-config; kubectl delete secret manifest-secret"),
		},
	},

	// ── Lab 6: requests и limits ──
	"ch-k8si-lab6": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod resource-pod too-large-pod --ignore-not-found >/dev/null 2>&1
kubectl delete deployment limited-web --ignore-not-found >/dev/null 2>&1
rm -f /root/resource-pod.yaml /root/limited-web.yaml`,
		Checks: map[int]string{
			1: kcheck(jp("get pod resource-pod", "{.spec.containers[0].resources.requests.cpu}", "50m")+` && `+
				jp("get pod resource-pod", "{.spec.containers[0].resources.requests.memory}", "64Mi")+` && `+
				jp("get pod resource-pod", "{.spec.containers[0].resources.limits.cpu}", "200m")+` && `+
				`[ -f /root/resource-pod.yaml ]`,
				"Pod с requests и limits создан",
				"Опиши в /root/resource-pod.yaml Pod resource-pod (label run=resource-pod, image nginx:alpine) с resources.requests/limits и примени файл"),
			2: kcheck(jp("get deploy limited-web", "{.spec.template.spec.containers[0].resources.requests.cpu}", "50m")+` && `+
				jp("get deploy limited-web", "{.status.readyReplicas}", "2")+` && `+
				`[ -f /root/limited-web.yaml ]`,
				"Deployment limited-web с лимитами готов",
				"Опиши Deployment в /root/limited-web.yaml (2 реплики, app=limited-web, те же resources) и примени"),
			3: kcheck(jp("get deploy limited-web", "{.spec.template.spec.containers[0].resources.limits.memory}", "256Mi")+` && `+
				`[ "$(kubectl get deploy limited-web -o jsonpath='{.status.readyReplicas}' 2>/dev/null)" -ge 1 ]`,
				"лимит памяти обновлён и rollout завершён",
				"Поставь limits.memory: 256Mi в /root/limited-web.yaml, примени и дождись kubectl rollout status deployment/limited-web"),
			4: kcheck(`! kubectl get pod resource-pod >/dev/null 2>&1 && ! kubectl get deploy limited-web >/dev/null 2>&1 && ! kubectl get pod too-large-pod >/dev/null 2>&1`,
				"ресурсы удалены",
				"kubectl delete pod resource-pod --ignore-not-found; kubectl delete deployment limited-web; kubectl delete pod too-large-pod --ignore-not-found"),
		},
	},

	// ── Lab 9: DNS ──
	"ch-k8si-lab9-dns": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deployment web-svc-app --ignore-not-found >/dev/null 2>&1
kubectl delete svc web-svc --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy web-svc-app", "{.status.readyReplicas}", "2")+` && `+
				jp("get svc web-svc", "{.spec.selector.app}", "web-svc-app")+` && `+
				jp("get svc web-svc", "{.spec.ports[0].port}", "80"),
				"Deployment и Service созданы, имена резолвятся через DNS",
				"kubectl create deployment web-svc-app --image=nginx:alpine --replicas=2 && kubectl expose deployment web-svc-app --name=web-svc --port=80"),
		},
	},

	// ── Lab 10: Ingress ──
	"ch-k8si-lab10-ingress": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete ingress store-ingress --ignore-not-found >/dev/null 2>&1
kubectl delete svc frontend-svc catalog-svc cart-svc --ignore-not-found >/dev/null 2>&1
kubectl delete deployment frontend catalog cart --ignore-not-found >/dev/null 2>&1
kubectl delete configmap frontend-content catalog-source cart-source --ignore-not-found >/dev/null 2>&1
rm -rf /root/ingress-lab && mkdir -p /root/ingress-lab
cat > /root/ingress-lab/app.yaml <<'YEOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: frontend-content
data:
  index.html: "<h1>store frontend</h1>"
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: catalog-source
data:
  default.conf: |
    server {
      listen 80;
      location / { default_type text/plain; return 200 "catalog service\n"; }
    }
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cart-source
data:
  default.conf: |
    server {
      listen 80;
      location / { default_type text/plain; return 200 "cart service\n"; }
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          volumeMounts:
            - name: content
              mountPath: /usr/share/nginx/html
      volumes:
        - name: content
          configMap:
            name: frontend-content
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: catalog
spec:
  replicas: 1
  selector:
    matchLabels:
      app: catalog
  template:
    metadata:
      labels:
        app: catalog
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          volumeMounts:
            - name: content
              mountPath: /etc/nginx/conf.d
      volumes:
        - name: content
          configMap:
            name: catalog-source
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cart
spec:
  replicas: 1
  selector:
    matchLabels:
      app: cart
  template:
    metadata:
      labels:
        app: cart
    spec:
      containers:
        - name: nginx
          image: nginx:alpine
          volumeMounts:
            - name: content
              mountPath: /etc/nginx/conf.d
      volumes:
        - name: content
          configMap:
            name: cart-source
---
apiVersion: v1
kind: Service
metadata:
  name: frontend-svc
spec:
  selector:
    app: frontend
  ports:
    - port: 80
---
apiVersion: v1
kind: Service
metadata:
  name: catalog-svc
spec:
  selector:
    app: catalog
  ports:
    - port: 80
---
apiVersion: v1
kind: Service
metadata:
  name: cart-svc
spec:
  selector:
    app: cart
  ports:
    - port: 80
YEOF`,
		Descs: map[int]string{
			// Экспорт курса потерял тексты этих двух заданий ("$2a" и "$2b").
			2: `<p>Создай Ingress с именем <code>store-ingress</code>, который маршрутизирует запросы по путям:</p>
<ul>
<li><code>/</code> → сервис <code>frontend-svc</code> (порт 80)</li>
<li><code>/catalog</code> → сервис <code>catalog-svc</code> (порт 80)</li>
<li><code>/cart</code> → сервис <code>cart-svc</code> (порт 80)</li>
</ul>
<p>Подсказка: <code>kubectl create ingress store-ingress --rule="/*=frontend-svc:80" --rule="/catalog*=catalog-svc:80" --rule="/cart*=cart-svc:80"</code></p>`,
			3: `<p>Проверь, что маршрутизация работает. В песочнице Ingress доступен по адресу узла <code>10.55.0.2</code>:</p>
<pre><code>curl http://10.55.0.2/
curl http://10.55.0.2/catalog
curl http://10.55.0.2/cart</code></pre>
<p>Сохрани ответ корневого пути в <code>/root/ingress_root.txt</code> — в нём должен быть текст <code>store frontend</code>.</p>`,
		},
		Checks: map[int]string{
			1: kcheck(jp("get deploy frontend", "{.status.readyReplicas}", "1")+` && `+
				jp("get deploy catalog", "{.status.readyReplicas}", "1")+` && `+
				jp("get deploy cart", "{.status.readyReplicas}", "1")+` && `+
				`kubectl get svc frontend-svc catalog-svc cart-svc >/dev/null 2>&1`,
				"приложение развёрнуто: три Deployment и три Service",
				"kubectl apply -f /root/ingress-lab/app.yaml — затем kubectl rollout status для каждого Deployment"),
			2: kcheck(`kubectl get ingress store-ingress >/dev/null 2>&1 && `+
				`kubectl get ingress store-ingress -o jsonpath='{.spec.rules[0].http.paths[*].backend.service.name}' 2>/dev/null | grep -q 'frontend-svc' && `+
				`kubectl get ingress store-ingress -o jsonpath='{.spec.rules[0].http.paths[*].backend.service.name}' 2>/dev/null | grep -q 'catalog-svc'`,
				"Ingress создан с маршрутами",
				`kubectl create ingress store-ingress --rule="/*=frontend-svc:80" --rule="/catalog*=catalog-svc:80" --rule="/cart*=cart-svc:80"`),
			3: kcheck(`grep -q 'store frontend' /root/ingress_root.txt 2>/dev/null && `+
				`for i in $(seq 1 20); do curl -s --max-time 5 http://10.55.0.2/catalog | grep -q 'catalog service' && break; sleep 2; done; `+
				`curl -s --max-time 5 http://10.55.0.2/catalog | grep -q 'catalog service'`,
				"маршрутизация через Ingress работает",
				"curl http://10.55.0.2/ > /root/ingress_root.txt и проверь curl http://10.55.0.2/catalog"),
			4: kcheck(`! kubectl get ingress store-ingress >/dev/null 2>&1 && ! kubectl get deploy frontend >/dev/null 2>&1 && `+
				`! kubectl get svc cart-svc >/dev/null 2>&1 && ! kubectl get configmap catalog-source >/dev/null 2>&1`,
				"ресурсы лаборатории удалены",
				"kubectl delete ingress store-ingress; kubectl delete service frontend-svc catalog-svc cart-svc; kubectl delete deployment frontend catalog cart; kubectl delete configmap frontend-content catalog-source cart-source"),
		},
	},
}
