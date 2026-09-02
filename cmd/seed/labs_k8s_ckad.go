package main

// Fixtures + auto-checks for "Kubernetes CKAD" (k8s-ckad). Runs in the k3s golden
// (sandboxImageK8s); k3s auto-starts, KUBECONFIG is exported. Checks validate the
// resource the task creates via kubectl jsonpath (exit 0 = solved).
//
// Covered here: the self-contained labs (Pods/PVC/SecurityContext/RBAC/Probes/
// Jobs/HPA), the deploy-strategies lab (lab16) and the two debug labs (lab13/lab14)
// — the last three drop pre-authored broken/ready manifests into /root via the Setup.
// The debug labs only auto-check the "Исправить"/cleanup tasks; the investigate and
// kubectl-debug (ephemeral container) steps stay manual on purpose.
//
// Still without checks: Gateway API (lab9) — needs its CRDs + a controller, which the
// offline k3s golden doesn't ship; it keeps the manual "Готово" until we vendor them.

var k8sCkadLabs = map[string]labSpec{
	// ── Lab 1: Init и Sidecar containers ──
	"ch-ckad-lab1": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod init-pod sidecar-pod multi-init --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get pod init-pod", "{.spec.initContainers[0].name}", "init-data")+` && `+
				jp("get pod init-pod", "{.spec.initContainers[0].image}", "nginx:alpine")+` && `+
				jp("get pod init-pod", "{.status.phase}", "Running"),
				"init-pod с init container init-data запущен",
				"опиши Pod init-pod: initContainers: [{name: init-data, image: nginx:alpine, ...}], общий emptyDir volume, kubectl apply -f"),
			2: kcheck(jp("get pod sidecar-pod", "{.spec.containers[*].name}", "'app log-reader'")+` && `+
				jp("get pod sidecar-pod", "{.status.phase}", "Running"),
				"sidecar-pod с контейнерами app и log-reader запущен",
				"Pod sidecar-pod: containers: [{name: app}, {name: log-reader}] на nginx:alpine, общий emptyDir /logs"),
			3: kcheck(jp("get pod multi-init", "{.spec.initContainers[*].name}", "'step1 step2'")+` && `+
				jp("get pod multi-init", "{.status.phase}", "Running"),
				"multi-init с двумя init containers запущен",
				"Pod multi-init: initContainers: [{name: step1}, {name: step2}], основной контейнер app"),
			4: kcheck(`! kubectl get pod init-pod >/dev/null 2>&1 && ! kubectl get pod sidecar-pod >/dev/null 2>&1 && ! kubectl get pod multi-init >/dev/null 2>&1`,
				"тестовые Pod'ы удалены",
				"kubectl delete pod init-pod sidecar-pod multi-init --ignore-not-found"),
		},
	},

	// ── Lab 2: PVC и данные между Pod'ами ──
	"ch-ckad-lab2": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod pvc-writer-pod pvc-reader-pod emptydir-demo --ignore-not-found >/dev/null 2>&1
kubectl delete pvc lab-pvc --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get pvc lab-pvc", "{.spec.accessModes[0]}", "ReadWriteOnce")+` && `+
				jp("get pvc lab-pvc", "{.spec.resources.requests.storage}", "100Mi"),
				"PVC lab-pvc создан (RWO, 100Mi)",
				"kubectl apply -f: PersistentVolumeClaim lab-pvc, accessModes: [ReadWriteOnce], resources.requests.storage: 100Mi"),
			2: kcheck(jp("get pod pvc-writer-pod", "{.spec.volumes[0].persistentVolumeClaim.claimName}", "lab-pvc")+` && `+
				jp("get pod pvc-writer-pod", "{.status.phase}", "Running"),
				"pvc-writer-pod подключил PVC и запущен",
				"Pod pvc-writer-pod: volumes: [{persistentVolumeClaim: {claimName: lab-pvc}}], volumeMounts на /data"),
			3: kcheck(`! kubectl get pod pvc-writer-pod >/dev/null 2>&1 && ! kubectl get pvc lab-pvc >/dev/null 2>&1`,
				"Pod'ы и PVC удалены",
				"kubectl delete pod pvc-writer-pod pvc-reader-pod emptydir-demo --ignore-not-found; kubectl delete pvc lab-pvc --ignore-not-found"),
		},
	},

	// ── Lab 3: SecurityContext ──
	"ch-ckad-lab3": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod secure-pod readonly-pod noprivilege-pod --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(`[ "$(kubectl get pod secure-pod -o jsonpath='{.spec.securityContext.runAsUser}{.spec.containers[0].securityContext.runAsUser}' 2>/dev/null)" = 1000 ]`,
				"secure-pod запускается под runAsUser 1000",
				"Pod secure-pod на nginx:alpine, securityContext.runAsUser: 1000"),
			2: kcheck(jp("get pod readonly-pod", "{.spec.containers[0].securityContext.readOnlyRootFilesystem}", "true"),
				"readonly-pod с readOnlyRootFilesystem",
				"Pod readonly-pod: containers[0].securityContext.readOnlyRootFilesystem: true, emptyDir на /tmp"),
			3: kcheck(jp("get pod noprivilege-pod", "{.spec.containers[0].securityContext.allowPrivilegeEscalation}", "false")+` && `+
				jp("get pod noprivilege-pod", "{.spec.containers[0].securityContext.capabilities.drop[0]}", "ALL"),
				"noprivilege-pod без эскалации привилегий",
				"Pod noprivilege-pod: securityContext.allowPrivilegeEscalation: false, capabilities.drop: [ALL]"),
			4: kcheck(`! kubectl get pod secure-pod >/dev/null 2>&1 && ! kubectl get pod readonly-pod >/dev/null 2>&1 && ! kubectl get pod noprivilege-pod >/dev/null 2>&1`,
				"тестовые Pod'ы удалены",
				"kubectl delete pod secure-pod readonly-pod noprivilege-pod --ignore-not-found"),
		},
	},

	// ── Lab 4: RBAC ── (our kubeconfig is cluster-admin, so the binding succeeds)
	"ch-ckad-lab8": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod sa-pod --ignore-not-found >/dev/null 2>&1
kubectl delete rolebinding pod-reader-binding --ignore-not-found >/dev/null 2>&1
kubectl delete role pod-reader --ignore-not-found >/dev/null 2>&1
kubectl delete sa app-sa --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(`kubectl get sa app-sa >/dev/null 2>&1`,
				"ServiceAccount app-sa создан",
				"kubectl create serviceaccount app-sa"),
			2: kcheck(jp("get role pod-reader", "{.rules[0].resources[0]}", "pods")+` && `+
				jp("get role pod-reader", "{.rules[0].verbs[*]}", "'get list watch'"),
				"Role pod-reader с get/list/watch на pods",
				"kubectl create role pod-reader --verb=get,list,watch --resource=pods"),
			3: kcheck(`kubectl get rolebinding pod-reader-binding >/dev/null 2>&1`,
				"RoleBinding создан",
				"kubectl create rolebinding pod-reader-binding --role=pod-reader --serviceaccount=default:app-sa"),
			4: kcheck(jp("get pod sa-pod", "{.spec.serviceAccountName}", "app-sa"),
				"sa-pod использует ServiceAccount app-sa",
				"Pod sa-pod на nginx:alpine, spec.serviceAccountName: app-sa"),
			5: kcheck(`! kubectl get pod sa-pod >/dev/null 2>&1 && ! kubectl get sa app-sa >/dev/null 2>&1`,
				"Pod и ServiceAccount удалены",
				"kubectl delete pod sa-pod; kubectl delete sa app-sa"),
		},
	},

	// ── Lab 5: Probes ──
	"ch-ckad-lab4": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod liveness-pod readiness-pod exec-probe tcp-probe --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get pod liveness-pod", "{.spec.containers[0].livenessProbe.httpGet.path}", "/")+` && `+
				jp("get pod liveness-pod", "{.spec.containers[0].livenessProbe.httpGet.port}", "80"),
				"liveness-pod с HTTP livenessProbe",
				"Pod liveness-pod на nginx:alpine, livenessProbe.httpGet: {path: /, port: 80}"),
			2: kcheck(jp("get pod readiness-pod", "{.spec.containers[0].readinessProbe.httpGet.path}", "/ready")+` && `+
				jp("get pod readiness-pod", "{.spec.containers[0].readinessProbe.httpGet.port}", "80"),
				"readiness-pod с HTTP readinessProbe",
				"Pod readiness-pod: readinessProbe.httpGet: {path: /ready, port: 80}"),
			3: kcheck(jp("get pod exec-probe", "{.spec.containers[0].livenessProbe.exec.command[*]}", "'cat /tmp/healthy'"),
				"exec-probe с exec livenessProbe",
				"Pod exec-probe: livenessProbe.exec.command: [cat, /tmp/healthy]"),
			4: kcheck(jp("get pod tcp-probe", "{.spec.containers[0].image}", "redis:alpine")+` && `+
				jp("get pod tcp-probe", "{.spec.containers[0].livenessProbe.tcpSocket.port}", "6379"),
				"tcp-probe с TCP livenessProbe",
				"Pod tcp-probe на redis:alpine: livenessProbe.tcpSocket.port: 6379"),
			5: kcheck(`! kubectl get pod liveness-pod >/dev/null 2>&1 && ! kubectl get pod readiness-pod >/dev/null 2>&1 && ! kubectl get pod exec-probe >/dev/null 2>&1 && ! kubectl get pod tcp-probe >/dev/null 2>&1`,
				"тестовые Pod'ы удалены",
				"kubectl delete pod liveness-pod readiness-pod exec-probe tcp-probe --ignore-not-found"),
		},
	},

	// ── Lab 6: Jobs и CronJobs ──
	"ch-ckad-lab5": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete job pi-job multi-job manual-run --ignore-not-found >/dev/null 2>&1
kubectl delete cronjob hello-cron --ignore-not-found >/dev/null 2>&1`,
		Checks: map[int]string{
			1: kcheck(jp("get job pi-job", "{.spec.template.spec.restartPolicy}", "Never")+` && `+
				`[ "$(kubectl get job pi-job -o jsonpath='{.status.succeeded}' 2>/dev/null)" -ge 1 ] 2>/dev/null`,
				"Job pi-job выполнен",
				`kubectl apply -f: Job pi-job, template.spec.restartPolicy: Never, command: ["sh","-c","echo 3.14159"] на nginx:alpine`),
			2: kcheck(jp("get job multi-job", "{.spec.completions}", "3")+` && `+
				jp("get job multi-job", "{.spec.parallelism}", "2"),
				"Job multi-job с completions:3, parallelism:2",
				"Job multi-job: spec.completions: 3, spec.parallelism: 2"),
			3: kcheck(`[ "$(kubectl get job multi-job -o jsonpath='{.status.succeeded}' 2>/dev/null)" = 3 ]`,
				"multi-job завершился (3/3)",
				"kubectl wait --for=condition=complete job/multi-job --timeout=90s"),
			4: kcheck(jp("get cronjob hello-cron", "{.spec.schedule}", "'*/1 * * * *'")+` && `+
				jp("get cronjob hello-cron", "{.spec.jobTemplate.spec.template.spec.restartPolicy}", "OnFailure"),
				"CronJob hello-cron с расписанием */1 * * * *",
				"CronJob hello-cron: schedule: '*/1 * * * *', jobTemplate restartPolicy: OnFailure"),
			5: kcheck(jp("get cronjob hello-cron", "{.spec.schedule}", "'*/1 * * * *'"),
				"расписание CronJob корректное",
				"kubectl get cronjob hello-cron -o jsonpath='{.spec.schedule}'"),
			6: kcheck(`[ "$(kubectl get job manual-run -o jsonpath='{.status.succeeded}' 2>/dev/null)" -ge 1 ] 2>/dev/null`,
				"Job manual-run из CronJob выполнен",
				"kubectl create job manual-run --from=cronjob/hello-cron"),
			7: kcheck(`! kubectl get job pi-job >/dev/null 2>&1 && ! kubectl get job multi-job >/dev/null 2>&1 && ! kubectl get cronjob hello-cron >/dev/null 2>&1`,
				"Jobs и CronJob удалены",
				"kubectl delete job pi-job multi-job manual-run; kubectl delete cronjob hello-cron"),
		},
	},

	// ── Lab 7: HPA ── (no metrics-server offline, so HPA never actually scales; we
	// only validate the HPA object's spec, which is what the tasks ask for).
	"ch-ckad-lab6": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete hpa hpa-app --ignore-not-found >/dev/null 2>&1
kubectl delete deployment hpa-app --ignore-not-found >/dev/null 2>&1`,
		Descs: map[int]string{
			1: "Создай Deployment <code>hpa-app</code> на образе <code>nginx:alpine</code> (1 реплика). Он станет целью для HPA. Пример: <code>kubectl create deployment hpa-app --image=nginx:alpine</code>",
		},
		Checks: map[int]string{
			1: kcheck(jp("get deploy hpa-app", "{.spec.template.spec.containers[0].image}", "nginx:alpine"),
				"Deployment hpa-app создан",
				"kubectl create deployment hpa-app --image=nginx:alpine"),
			2: kcheck(jp("get hpa hpa-app", "{.spec.minReplicas}", "1")+` && `+
				jp("get hpa hpa-app", "{.spec.maxReplicas}", "3"),
				"HPA hpa-app создан (min 1, max 3)",
				"kubectl apply -f: HorizontalPodAutoscaler hpa-app, scaleTargetRef deployment hpa-app, minReplicas: 1, maxReplicas: 3, metric memory 50%"),
			3: kcheck(`[ "$(kubectl get hpa hpa-app -o jsonpath='{.spec.minReplicas}/{.spec.maxReplicas}' 2>/dev/null)" = "1/3" ]`,
				"лимиты реплик HPA: 1/3",
				"kubectl get hpa hpa-app -o jsonpath='{.spec.minReplicas}/{.spec.maxReplicas}'"),
			4: kcheck(`! kubectl get hpa hpa-app >/dev/null 2>&1 && ! kubectl get deployment hpa-app >/dev/null 2>&1`,
				"HPA и Deployment удалены",
				"kubectl delete hpa hpa-app; kubectl delete deployment hpa-app"),
		},
	},

	// ── Lab 11: Стратегии деплоя (canary / blue-green) ── the Setup drops the
	// ready manifests the tasks apply; checks validate the applied resources and
	// the Service selector (offline nginx:alpine pods can't return distinct bodies,
	// so we verify the selector switch rather than the HTTP response).
	"ch-ckad-lab16-deployment-strategies": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deploy shop-v1 shop-v2 checkout-blue checkout-green --ignore-not-found >/dev/null 2>&1
kubectl delete svc shop-svc checkout-svc --ignore-not-found >/dev/null 2>&1
kubectl delete pod strategy-debug --ignore-not-found >/dev/null 2>&1
mkdir -p /root/deploy-strategies
cat > /root/deploy-strategies/canary-v1.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: shop-v1}
spec:
  replicas: 3
  selector: {matchLabels: {app: shop, version: v1}}
  template:
    metadata: {labels: {app: shop, track: active, version: v1}}
    spec: {containers: [{name: nginx, image: nginx:alpine}]}
---
apiVersion: v1
kind: Service
metadata: {name: shop-svc}
spec:
  selector: {app: shop, track: active}
  ports: [{port: 80, targetPort: 80}]
YAML
cat > /root/deploy-strategies/canary-v2.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: shop-v2}
spec:
  replicas: 1
  selector: {matchLabels: {app: shop, version: v2}}
  template:
    metadata: {labels: {app: shop, track: active, version: v2}}
    spec: {containers: [{name: nginx, image: nginx:alpine}]}
YAML
cat > /root/deploy-strategies/blue-green.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: checkout-blue}
spec:
  replicas: 1
  selector: {matchLabels: {app: checkout, color: blue}}
  template:
    metadata: {labels: {app: checkout, color: blue}}
    spec: {containers: [{name: nginx, image: nginx:alpine}]}
---
apiVersion: apps/v1
kind: Deployment
metadata: {name: checkout-green}
spec:
  replicas: 1
  selector: {matchLabels: {app: checkout, color: green}}
  template:
    metadata: {labels: {app: checkout, color: green}}
    spec: {containers: [{name: nginx, image: nginx:alpine}]}
---
apiVersion: v1
kind: Service
metadata: {name: checkout-svc}
spec:
  selector: {app: checkout, color: blue}
  ports: [{port: 80, targetPort: 80}]
YAML
kubectl run strategy-debug --image=busybox:1.36 --restart=Never --command -- sleep 3600 >/dev/null 2>&1 || true`,
		Checks: map[int]string{
			1: kcheck(jp("get deploy shop-v1", "{.spec.replicas}", "3")+` && kubectl get svc shop-svc >/dev/null 2>&1`,
				"stable-версия shop-v1 (3 реплики) и Service shop-svc запущены",
				"kubectl apply -f /root/deploy-strategies/canary-v1.yaml"),
			2: kcheck(jp("get deploy shop-v2", "{.spec.replicas}", "1")+` && `+
				jp("get deploy shop-v2", "{.spec.template.metadata.labels.track}", "active"),
				"canary shop-v2 добавлен (Service видит обе версии)",
				"kubectl apply -f /root/deploy-strategies/canary-v2.yaml"),
			3: kcheck(`kubectl get deploy checkout-blue >/dev/null 2>&1 && kubectl get deploy checkout-green >/dev/null 2>&1 && `+
				jp("get svc checkout-svc", "{.spec.selector.color}", "blue"),
				"blue-green развёрнут, Service указывает на blue",
				"kubectl apply -f /root/deploy-strategies/blue-green.yaml"),
			4: kcheck(jp("get svc checkout-svc", "{.spec.selector.color}", "green"),
				"traffic переключён на green",
				`kubectl patch svc checkout-svc -p '{"spec":{"selector":{"app":"checkout","color":"green"}}}'`),
			5: kcheck(jp("get svc checkout-svc", "{.spec.selector.color}", "blue"),
				"traffic откатан обратно на blue",
				`kubectl patch svc checkout-svc -p '{"spec":{"selector":{"app":"checkout","color":"blue"}}}'`),
			6: kcheck(`! kubectl get deploy shop-v1 >/dev/null 2>&1 && ! kubectl get deploy checkout-blue >/dev/null 2>&1 && ! kubectl get svc shop-svc >/dev/null 2>&1 && ! kubectl get svc checkout-svc >/dev/null 2>&1 && ! kubectl get pod strategy-debug >/dev/null 2>&1`,
				"все ресурсы лабораторной удалены",
				"kubectl delete deploy shop-v1 shop-v2 checkout-blue checkout-green; kubectl delete svc shop-svc checkout-svc; kubectl delete pod strategy-debug"),
		},
	},

	// ── Lab 9 (debug startup) ── Setup drops four intentionally-broken workloads in
	// /root/debug-workloads; the student diagnoses each and edits the YAML. Only the
	// "Исправить"/cleanup tasks get checks (the "Исследовать" tasks stay manual — they
	// are read-and-understand steps). Each bug is real and offline-reproducible:
	//   app-a  frontend-api  bad image tag  -> ErrImagePull   (fix: nginx:alpine)
	//   app-b  worker-api    missing env    -> CrashLoop      (fix: add DATABASE_URL)
	//   app-c  billing-api   wrong CM mount -> initFailed     (fix: mountPath /config)
	//   app-d  catalog-api   probe /readyz  -> never Ready    (fix: probe path /healthz)
	"ch-ckad-lab13-debug-startup": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete pod frontend-api worker-api billing-api --ignore-not-found >/dev/null 2>&1
kubectl delete deploy catalog-api --ignore-not-found >/dev/null 2>&1
kubectl delete configmap billing-config --ignore-not-found >/dev/null 2>&1
mkdir -p /root/debug-workloads
cat > /root/debug-workloads/app-a.yaml <<'YAML'
# BUG: неверный тег образа — Pod не может стянуть image (офлайн-кластер).
apiVersion: v1
kind: Pod
metadata: {name: frontend-api, labels: {app: frontend-api}}
spec:
  containers:
  - name: app
    image: nginx:alpne
    ports: [{containerPort: 80}]
YAML
cat > /root/debug-workloads/app-b.yaml <<'YAML'
# command имитирует стартовую команду собранного приложения — не переписывай её.
# Процесс падает, потому что ему не передана ожидаемая конфигурация.
apiVersion: v1
kind: Pod
metadata: {name: worker-api, labels: {app: worker-api}}
spec:
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh","-c","if [ -z \"$DATABASE_URL\" ]; then echo 'FATAL: DATABASE_URL is required' >&2; exit 1; fi; echo worker started; sleep 3600"]
YAML
cat > /root/debug-workloads/app-c.yaml <<'YAML'
# init container prepare-config копирует конфиг из ConfigMap в общий emptyDir.
# BUG: том с конфигом смонтирован не туда, где init его читает.
apiVersion: v1
kind: ConfigMap
metadata: {name: billing-config}
data: {billing.conf: "rate=0.2\n"}
---
apiVersion: v1
kind: Pod
metadata: {name: billing-api, labels: {app: billing-api}}
spec:
  initContainers:
  - name: prepare-config
    image: busybox:1.36
    command: ["sh","-c","cp /config/billing.conf /shared/billing.conf && echo copied"]
    volumeMounts:
    - {name: cfg, mountPath: /etc/wrong}
    - {name: shared, mountPath: /shared}
  containers:
  - name: app
    image: busybox:1.36
    command: ["sh","-c","test -f /shared/billing.conf && echo billing ready && sleep 3600"]
    volumeMounts:
    - {name: shared, mountPath: /shared}
  volumes:
  - {name: cfg, configMap: {name: billing-config}}
  - {name: shared, emptyDir: {}}
YAML
cat > /root/debug-workloads/app-d.yaml <<'YAML'
# Приложение отдаёт health endpoint по /healthz, но readinessProbe стучится в /readyz
# -> проверка готовности не проходит, rollout не завершается.
apiVersion: apps/v1
kind: Deployment
metadata: {name: catalog-api, labels: {app: catalog-api}}
spec:
  replicas: 1
  selector: {matchLabels: {app: catalog-api}}
  template:
    metadata: {labels: {app: catalog-api}}
    spec:
      containers:
      - name: app
        image: nginx:alpine
        command: ["sh","-c","echo ok > /usr/share/nginx/html/healthz && exec nginx -g 'daemon off;'"]
        ports: [{containerPort: 80}]
        readinessProbe:
          httpGet: {path: /readyz, port: 80}
          initialDelaySeconds: 2
          periodSeconds: 3
YAML`,
		Checks: map[int]string{
			2: kcheck(jp("get pod frontend-api", "{.status.containerStatuses[0].ready}", "true"),
				"frontend-api исправлен и в состоянии Ready",
				"причина в теге образа: в app-a.yaml image: nginx:alpne -> nginx:alpine; пересоздай Pod"),
			4: kcheck(jp("get pod worker-api", "{.status.containerStatuses[0].ready}", "true"),
				"worker-api больше не падает и в состоянии Ready",
				"процесс требует переменную DATABASE_URL: добавь в app-b.yaml env: [{name: DATABASE_URL, value: \"postgres://db:5432/app\"}]"),
			6: kcheck(jp("get pod billing-api", "{.status.containerStatuses[0].ready}", "true"),
				"billing-api: init container отработал, Pod Ready",
				"init читает /config/billing.conf, а том смонтирован в /etc/wrong: поменяй mountPath тома cfg на /config"),
			8: kcheck(jp("get deploy catalog-api", "{.status.readyReplicas}", "1"),
				"catalog-api: rollout завершён, реплика Ready",
				"readinessProbe стучится в /readyz, а приложение отдаёт /healthz: поменяй httpGet.path на /healthz"),
			10: kcheck(`! kubectl get pod frontend-api >/dev/null 2>&1 && ! kubectl get pod worker-api >/dev/null 2>&1 && ! kubectl get pod billing-api >/dev/null 2>&1 && ! kubectl get deploy catalog-api >/dev/null 2>&1 && ! kubectl get configmap billing-config >/dev/null 2>&1`,
				"все debug-ресурсы удалены",
				"kubectl delete pod frontend-api worker-api billing-api; kubectl delete deploy catalog-api; kubectl delete configmap billing-config"),
		},
	},

	// ── Lab 10 (debug Service/ConfigMap/routing) ── Setup drops three broken stacks in
	// /root/debug-routing. Only the "Исправить" tasks + cleanup are auto-checked; the
	// investigate / kubectl-debug (ephemeral container) steps stay manual. Bugs:
	//   app-a  orders   Service selector wrong (empty endpoints) + env configMapKeyRef
	//                   key wrong (app serves empty config)  -> fix selector, fix key
	//   app-b  reports  Service targetPort 8080 != container 80  -> fix targetPort 80
	//   app-c  profile  Deployment envFrom a ConfigMap that doesn't exist -> create it
	"ch-ckad-lab14-debug-service-config": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
kubectl delete deploy orders-api reports-api profile-api --ignore-not-found >/dev/null 2>&1
kubectl delete svc orders-svc reports-svc profile-svc --ignore-not-found >/dev/null 2>&1
kubectl delete configmap orders-config profile-config --ignore-not-found >/dev/null 2>&1
kubectl delete pod debug-client --ignore-not-found >/dev/null 2>&1
mkdir -p /root/debug-routing
cat > /root/debug-routing/app-a.yaml <<'YAML'
apiVersion: v1
kind: ConfigMap
metadata: {name: orders-config}
data: {app_mode: "production"}
---
apiVersion: apps/v1
kind: Deployment
metadata: {name: orders-api, labels: {app: orders-api}}
spec:
  replicas: 1
  selector: {matchLabels: {app: orders-api}}
  template:
    metadata: {labels: {app: orders-api}}
    spec:
      containers:
      - name: web
        image: nginx:alpine
        command: ["sh","-c","echo \"mode=$APP_MODE\" > /usr/share/nginx/html/index.html && exec nginx -g 'daemon off;'"]
        # BUG: ключ должен быть app_mode (см. ConfigMap) — сейчас mode, значение пустое.
        env:
        - name: APP_MODE
          valueFrom: {configMapKeyRef: {name: orders-config, key: mode, optional: true}}
        ports: [{containerPort: 80}]
---
# BUG: селектор не совпадает с labels Pod'ов (app: orders-api) -> пустые endpoints.
apiVersion: v1
kind: Service
metadata: {name: orders-svc}
spec:
  selector: {app: orders-frontend}
  ports: [{port: 80, targetPort: 80}]
YAML
cat > /root/debug-routing/app-b.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: reports-api, labels: {app: reports-api}}
spec:
  replicas: 1
  selector: {matchLabels: {app: reports-api}}
  template:
    metadata: {labels: {app: reports-api}}
    spec:
      containers:
      - {name: web, image: nginx:alpine, ports: [{containerPort: 80}]}
---
# BUG: targetPort 8080, а контейнер слушает 80 -> запрос через Service не доходит.
apiVersion: v1
kind: Service
metadata: {name: reports-svc}
spec:
  selector: {app: reports-api}
  ports: [{port: 80, targetPort: 8080}]
YAML
cat > /root/debug-routing/app-c.yaml <<'YAML'
apiVersion: apps/v1
kind: Deployment
metadata: {name: profile-api, labels: {app: profile-api}}
spec:
  replicas: 1
  selector: {matchLabels: {app: profile-api}}
  template:
    metadata: {labels: {app: profile-api}}
    spec:
      containers:
      # BUG: envFrom ссылается на ConfigMap profile-config, которого нет -> контейнер не стартует.
      - name: web
        image: nginx:alpine
        envFrom: [{configMapRef: {name: profile-config}}]
        ports: [{containerPort: 80}]
---
apiVersion: v1
kind: Service
metadata: {name: profile-svc}
spec:
  selector: {app: profile-api}
  ports: [{port: 80, targetPort: 80}]
YAML`,
		Checks: map[int]string{
			2: kcheck(`[ -n "$(kubectl get endpoints orders-svc -o jsonpath='{.subsets[0].addresses[0].ip}' 2>/dev/null)" ]`,
				"orders-svc: селектор исправлен, endpoints не пустые",
				"селектор Service не совпадает с labels Pod'ов: поменяй selector orders-svc на app: orders-api"),
			4: kcheck(jp("get deploy orders-api", "{.status.readyReplicas}", "1")+` && `+
				jp("get deploy orders-api", "{.spec.template.spec.containers[0].env[0].valueFrom.configMapKeyRef.key}", "app_mode"),
				"orders-api: ключ конфигурации исправлен (app_mode), rollout завершён",
				"configMapKeyRef.key указывает на mode, а в ConfigMap ключ app_mode: поменяй key на app_mode"),
			7: kcheck(jp("get svc reports-svc", "{.spec.ports[0].targetPort}", "80"),
				"reports-svc: targetPort совпадает с портом контейнера (80)",
				"targetPort 8080 не совпадает с портом контейнера 80: поменяй targetPort у reports-svc на 80"),
			9: kcheck(`kubectl get configmap profile-config >/dev/null 2>&1 && `+
				jp("get deploy profile-api", "{.status.readyReplicas}", "1"),
				"profile-config создан, profile-api стал Ready",
				"Deployment ждёт ConfigMap profile-config: kubectl create configmap profile-config --from-literal=app_env=prod"),
			12: kcheck(`! kubectl get deploy orders-api >/dev/null 2>&1 && ! kubectl get deploy reports-api >/dev/null 2>&1 && ! kubectl get deploy profile-api >/dev/null 2>&1 && ! kubectl get svc orders-svc >/dev/null 2>&1 && ! kubectl get svc reports-svc >/dev/null 2>&1 && ! kubectl get svc profile-svc >/dev/null 2>&1 && ! kubectl get pod debug-client >/dev/null 2>&1`,
				"все ресурсы лабораторной удалены",
				"kubectl delete deploy orders-api reports-api profile-api; kubectl delete svc orders-svc reports-svc profile-svc; kubectl delete configmap orders-config profile-config; kubectl delete pod debug-client"),
		},
	},
}
