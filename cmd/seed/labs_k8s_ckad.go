package main

// Fixtures + auto-checks for "Kubernetes CKAD" (k8s-ckad). Runs in the k3s golden
// (sandboxImageK8s); k3s auto-starts, KUBECONFIG is exported. Checks validate the
// resource the task creates via kubectl jsonpath (exit 0 = solved).
//
// Covered here: the self-contained labs (Pods/PVC/SecurityContext/RBAC/Probes/
// Jobs/HPA). The debug (lab13/lab14), Gateway API (lab9) and deploy-strategies
// (lab16) labs need pre-authored broken/ready manifests dropped into /root by the
// Setup (and Gateway API needs its CRDs+controller, which the offline k3s lacks) —
// those are a separate fixtures task and are intentionally left without checks for
// now, so they keep the manual "Готово" button rather than a broken "Проверить".

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
}
