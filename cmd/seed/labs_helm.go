package main

// Fixtures + auto-checks for the Helm course. Runs in the k3s golden
// (sandboxImageK8s): helm + kubectl + a running single-node k3s, offline.
//
// Covered: the self-contained labs that only need a LOCAL chart whose image is a
// baked one (nginx:alpine / nginx:1.25-alpine). The chart the tasks reference,
// /root/charts/nginx-chart, is created by the Setup (helmNginxChart).
//
// NOT covered yet (kept as the manual "Готово"):
//   - ch-helm-lab-repos    — Bitnami charts from a remote repo (no network offline);
//   - ch-helm-lab-helmfile — needs the `helmfile` binary + remote charts;
//   - ch-helm-lab-templates & ch-helm-lab5 (hooks) — the student edits chart templates;
//     verifying rendered output reliably needs richer fixtures. TODO.

// helmNginxChart writes a minimal chart at /root/charts/nginx-chart. Its Deployment
// is named <release>-nginx and honours replicaCount + image.repository/tag, which is
// exactly what the install/values/upgrade labs drive.
const helmNginxChart = `
mkdir -p /root/charts/nginx-chart/templates
cat > /root/charts/nginx-chart/Chart.yaml <<'EOF'
apiVersion: v2
name: nginx
description: TOT nginx chart
type: application
version: 0.1.0
appVersion: "1.0"
EOF
cat > /root/charts/nginx-chart/values.yaml <<'EOF'
replicaCount: 1
image:
  repository: nginx
  tag: alpine
EOF
cat > /root/charts/nginx-chart/templates/deployment.yaml <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-nginx
  labels:
    app: {{ .Release.Name }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ .Release.Name }}
  template:
    metadata:
      labels:
        app: {{ .Release.Name }}
    spec:
      containers:
      - name: nginx
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
EOF`

// hdeployed passes when a helm release is present and in status "deployed".
func hdeployed(rel string) string {
	return `helm status ` + rel + ` 2>/dev/null | grep -qi 'STATUS: deployed'`
}

var helmLabs = map[string]labSpec{
	// ── Lab 1: Первый helm install ──
	"ch-helm-lab1": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
helm uninstall my-nginx >/dev/null 2>&1` + helmNginxChart,
		Checks: map[int]string{
			1: kcheck(hdeployed("my-nginx"),
				"release my-nginx установлен (deployed)",
				"helm install my-nginx /root/charts/nginx-chart"),
			2: kcheck(`[ "$(kubectl get pods -l app=my-nginx --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)" -ge 1 ]`,
				"Pod'ы релиза my-nginx запущены",
				"kubectl get pods -l app=my-nginx (после helm install)"),
			3: kcheck(hdeployed("my-nginx"),
				"статус релиза — deployed",
				"helm status my-nginx"),
			4: kcheck(`! helm status my-nginx >/dev/null 2>&1`,
				"release my-nginx удалён",
				"helm uninstall my-nginx"),
		},
	},

	// ── Lab 3: Values и override ──
	"ch-helm-lab2": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
helm uninstall scaled-nginx prod-nginx >/dev/null 2>&1
rm -f /root/prod-values.yaml` + helmNginxChart,
		Checks: map[int]string{
			1: kcheck(jp("get deploy scaled-nginx-nginx", "{.spec.replicas}", "2"),
				"scaled-nginx установлен с 2 репликами",
				"helm install scaled-nginx /root/charts/nginx-chart --set replicaCount=2"),
			2: kcheck(jp("get deploy scaled-nginx-nginx", "{.status.readyReplicas}", "2"),
				"обе реплики готовы",
				"kubectl get deploy scaled-nginx-nginx (2/2)"),
			3: kcheck(`[ -f /root/prod-values.yaml ] && grep -qE 'replicaCount:\s*3' /root/prod-values.yaml && grep -qE 'tag:\s*"?alpine"?' /root/prod-values.yaml`,
				"prod-values.yaml создан (replicaCount 3, tag alpine)",
				"создай /root/prod-values.yaml с replicaCount: 3 и image.tag: alpine"),
			4: kcheck(hdeployed("prod-nginx"),
				"prod-nginx установлен с values-файлом",
				"helm install prod-nginx /root/charts/nginx-chart -f /root/prod-values.yaml"),
			5: kcheck(jp("get deploy prod-nginx-nginx", "{.spec.replicas}", "1"),
				"replicaCount переопределён до 1 через --set",
				"helm upgrade prod-nginx /root/charts/nginx-chart -f /root/prod-values.yaml --set replicaCount=1"),
			6: kcheck(`helm lint /root/charts/nginx-chart >/dev/null 2>&1`,
				"helm lint без ошибок",
				"helm lint /root/charts/nginx-chart"),
			7: kcheck(`! helm status scaled-nginx >/dev/null 2>&1 && ! helm status prod-nginx >/dev/null 2>&1`,
				"оба релиза удалены",
				"helm uninstall scaled-nginx prod-nginx"),
		},
	},

	// ── Lab 5: Создание своего chart ──
	"ch-helm-lab3": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
helm uninstall my-custom-app from-package >/dev/null 2>&1
rm -rf /root/myapp-chart /root/myapp-chart-*.tgz`,
		Checks: map[int]string{
			1: kcheck(`[ -f /root/myapp-chart/Chart.yaml ] && [ -d /root/myapp-chart/templates ]`,
				"заготовка chart создана",
				"helm create /root/myapp-chart"),
			2: kcheck(`grep -qE 'name:\s*myapp-chart' /root/myapp-chart/Chart.yaml && grep -q 'TOT training chart' /root/myapp-chart/Chart.yaml`,
				"Chart.yaml обновлён (name, description)",
				"в /root/myapp-chart/Chart.yaml: name: myapp-chart, description: TOT training chart"),
			3: kcheck(`grep -qE 'replicaCount:\s*2' /root/myapp-chart/values.yaml && grep -qE 'tag:\s*"?alpine"?' /root/myapp-chart/values.yaml`,
				"values.yaml настроен (replicaCount 2, tag alpine)",
				"в /root/myapp-chart/values.yaml: replicaCount: 2, image.tag: alpine"),
			4: kcheck(`helm lint /root/myapp-chart >/dev/null 2>&1`,
				"helm lint chart без ошибок",
				"helm lint /root/myapp-chart"),
			5: kcheck(hdeployed("my-custom-app"),
				"свой chart установлен (deployed)",
				"helm install my-custom-app /root/myapp-chart"),
			6: kcheck(`[ -f /root/myapp-chart/templates/configmap.yaml ] && grep -q 'configmap' /root/myapp-chart/values.yaml && helm template /root/myapp-chart 2>/dev/null | grep -q 'kind: ConfigMap'`,
				"условный ConfigMap в chart рендерится",
				"добавь configmap.enabled: true в values и templates/configmap.yaml с {{- if .Values.configmap.enabled }}"),
			7: kcheck(`ls /root/myapp-chart-*.tgz >/dev/null 2>&1`,
				"chart упакован в .tgz",
				"helm package /root/myapp-chart -d /root"),
			8: kcheck(hdeployed("from-package"),
				"release from-package установлен из архива",
				"helm install from-package /root/myapp-chart-*.tgz"),
			9: kcheck(`! helm status my-custom-app >/dev/null 2>&1 && ! helm status from-package >/dev/null 2>&1`,
				"оба релиза удалены",
				"helm uninstall my-custom-app from-package"),
		},
	},

	// ── Lab 6: Upgrade и Rollback ──
	"ch-helm-lab4": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
helm uninstall webapp new-webapp >/dev/null 2>&1` + helmNginxChart,
		Checks: map[int]string{
			1: kcheck(hdeployed("webapp")+` && `+jp("get deploy webapp-nginx", "{.spec.template.spec.containers[0].image}", "nginx:1.25-alpine"),
				"webapp установлен с образом nginx:1.25-alpine",
				"helm install webapp /root/charts/nginx-chart --set image.tag=1.25-alpine"),
			2: kcheck(`[ "$(helm history webapp 2>/dev/null | grep -c deployed)" -ge 1 ]`,
				"у webapp есть история ревизий",
				"helm history webapp"),
			3: kcheck(hdeployed("webapp")+` && `+jp("get deploy webapp-nginx", "{.spec.template.spec.containers[0].image}", "nginx:alpine"),
				"webapp обновлён на nginx:alpine",
				"helm upgrade webapp /root/charts/nginx-chart --set image.tag=alpine"),
			4: kcheck(`[ "$(helm history webapp 2>/dev/null | grep -cE '^[0-9]')" -ge 2 ]`,
				"есть вторая ревизия",
				"helm history webapp (revision 2)"),
			5: kcheck(jp("get deploy webapp-nginx", "{.spec.template.spec.containers[0].image}", "nginx:1.25-alpine"),
				"rollback вернул nginx:1.25-alpine",
				"helm rollback webapp 1"),
			6: kcheck(`helm get values webapp 2>/dev/null | grep -q '1.25-alpine'`,
				"values соответствуют revision 1",
				"helm get values webapp (tag 1.25-alpine)"),
			7: kcheck(hdeployed("new-webapp"),
				"new-webapp создан через upgrade --install",
				"helm upgrade --install new-webapp /root/charts/nginx-chart"),
			8: kcheck(jp("get deploy webapp-nginx", "{.spec.template.spec.containers[0].image}", "nginx:alpine")+` && `+jp("get deploy webapp-nginx", "{.spec.replicas}", "2"),
				"webapp обновлён (alpine, 2 реплики)",
				"helm upgrade webapp /root/charts/nginx-chart --set image.tag=alpine --set replicaCount=2 --wait"),
			9: kcheck(`[ "$(helm history webapp 2>/dev/null | grep -cE '^[0-9]')" -ge 4 ]`,
				"история содержит все ревизии, включая rollback",
				"helm history webapp"),
			10: kcheck(`! helm status webapp >/dev/null 2>&1 && ! helm status new-webapp >/dev/null 2>&1`,
				"оба релиза удалены",
				"helm uninstall webapp new-webapp"),
		},
	},
}
