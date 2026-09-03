package main

// Fixtures + auto-checks for the Helm course. Runs in the k3s golden
// (sandboxImageK8s): helm + kubectl + a running single-node k3s, offline.
//
// Covered: the self-contained labs that only need a LOCAL chart whose image is a
// baked one (nginx:alpine / nginx:1.25-alpine). The chart the tasks reference,
// /root/charts/nginx-chart, is created by the Setup (helmNginxChart).
//
// The Go-template lab (ch-helm-lab-templates) and the hooks lab (ch-helm-lab5) ship
// their own fixture charts (webChart / hooksChart) that the Setup drops into /root;
// their checks render the chart / read `helm get hooks` after install/upgrade/uninstall.
//
// The helmfile lab (ch-helm-lab-helmfile) uses the `helmfile` binary baked into the
// k8s golden + a local ./charts/webapp fixture. The repos lab (ch-helm-lab-repos) uses
// an offline Bitnami mirror baked into the golden (systemd python http.server on
// 127.0.0.1:8879 serving a vendored nginx chart) registered as the `bitnami` repo.
// All Helm labs are now covered.

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

// webChart writes the "before" chart at /root/template-lab/webchart that the Go-template
// lab edits: _helpers.tpl is pre-authored (fullname + labels), deployment.yaml starts
// with static name/labels/ports so the student converts them to include/range/toYaml/
// default/required and adds a conditional serviceaccount.yaml. It renders as-is.
const webChart = `
mkdir -p /root/template-lab/webchart/templates
cat > /root/template-lab/webchart/Chart.yaml <<'EOF'
apiVersion: v2
name: webchart
description: TOT Go-template lab chart
type: application
version: 0.1.0
appVersion: "1.0"
EOF
cat > /root/template-lab/webchart/values.yaml <<'EOF'
replicaCount: 1
image:
  repository: nginx
  tag: alpine
EOF
cat > /root/template-lab/webchart/templates/_helpers.tpl <<'EOF'
{{- define "webchart.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name -}}
{{- end -}}
{{- define "webchart.labels" -}}
app.kubernetes.io/name: {{ .Chart.Name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
EOF
cat > /root/template-lab/webchart/templates/deployment.yaml <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webchart
  labels:
    app: webchart
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webchart
  template:
    metadata:
      labels:
        app: webchart
    spec:
      containers:
      - name: web
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        ports:
        - name: http
          containerPort: 80
EOF
cat > /root/template-lab/webchart/templates/service.yaml <<'EOF'
apiVersion: v1
kind: Service
metadata:
  name: webchart
spec:
  selector:
    app: webchart
  ports:
  - port: 80
    targetPort: 80
EOF`

// hooksChart writes a minimal chart at /root/hooks-chart with one normal resource
// (a ConfigMap) so the release installs something; the student adds the hook Jobs
// (pre-install / post-upgrade / pre-upgrade weights / pre-delete) the lab is about.
const hooksChart = `
mkdir -p /root/hooks-chart/templates
cat > /root/hooks-chart/Chart.yaml <<'EOF'
apiVersion: v2
name: hooks-chart
description: TOT Helm hooks lab chart
type: application
version: 0.1.0
appVersion: "1.0"
EOF
cat > /root/hooks-chart/values.yaml <<'EOF'
{}
EOF
cat > /root/hooks-chart/templates/configmap.yaml <<'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Release.Name }}-config
data:
  greeting: hello
EOF`

// htpl renders the Go-template lab chart; helper keeps the checks readable.
const htpl = `helm template demo /root/template-lab/webchart`

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

	// ── Lab 4: Go template в Helm ── Setup drops the "before" webchart; the student
	// converts static YAML into helpers/range/toYaml/default/required + a conditional
	// ServiceAccount. Checks render the chart and assert the expected output (the
	// investigate/lint/quiz tasks stay manual). Each check fails on the "before" chart
	// and passes once the corresponding edit is made.
	"ch-helm-lab-templates": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + webChart,
		Checks: map[int]string{
			3: kcheck(htpl+` 2>/dev/null | grep -qE '^[[:space:]]*name: demo-webchart$'`,
				"metadata.name Deployment рендерится через include \"webchart.fullname\"",
				`в deployment.yaml: metadata.name: {{ include "webchart.fullname" . }}, labels: {{ include "webchart.labels" . | nindent 4 }}`),
			4: kcheck(htpl+` 2>/dev/null | grep -q 'containerPort: 9090'`,
				"порты контейнера рендерятся через range (http:80 + metrics:9090)",
				"добавь containerPorts в values.yaml и замени статичный ports на {{- range .Values.containerPorts }}"),
			5: kcheck(htpl+` 2>/dev/null | grep -q '128Mi'`,
				"resources вставлены через toYaml (limits/requests 200m/128Mi)",
				"добавь resources в values.yaml и вставь через {{ toYaml .Values.resources | nindent 12 }}"),
			6: kcheck(htpl+` --set replicaCount=3 2>/dev/null | grep -qE '^[[:space:]]*replicas: 3$'`,
				"replicas берутся из .Values.replicaCount | default 1 (override работает)",
				"замени replicas на {{ .Values.replicaCount | default 1 }}"),
			7: kcheck(htpl+` 2>/dev/null | grep -q 'kind: ServiceAccount' && ! `+htpl+` --set serviceAccount.enabled=false 2>/dev/null | grep -q 'kind: ServiceAccount'`,
				"ServiceAccount условный: есть при enabled=true, исчезает при false",
				"добавь serviceAccount.enabled: true и templates/serviceaccount.yaml в {{- if .Values.serviceAccount.enabled }}"),
			8: kcheck(htpl+` --set image.repository= 2>&1 | grep -qi 'image.repository is required'`,
				"image.repository обязателен через required",
				`image: "{{ required \"image.repository is required\" .Values.image.repository }}:..."`),
		},
	},

	// ── Lab 7: Helm Hooks ── Setup drops a minimal chart (one ConfigMap); the student
	// adds hook Jobs. Checks read `helm get hooks` / the surviving hook Jobs after
	// install/upgrade/uninstall. The create-manifest and observe tasks stay manual.
	"ch-helm-lab5": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
helm uninstall hooked-app >/dev/null 2>&1
kubectl delete job -l app.kubernetes.io/managed-by=Helm >/dev/null 2>&1` + hooksChart,
		Checks: map[int]string{
			3: kcheck(hdeployed("hooked-app")+` && helm get hooks hooked-app 2>/dev/null | grep -qi pre-install`,
				"release установлен, pre-install hook отработал",
				"helm install hooked-app /root/hooks-chart (после создания pre-install-job.yaml)"),
			6: kcheck(`helm get hooks hooked-app 2>/dev/null | grep -qi post-upgrade`,
				"post-upgrade hook описан в релизе",
				"добавь post-upgrade-job.yaml и helm upgrade hooked-app /root/hooks-chart"),
			7: kcheck(`helm get hooks hooked-app 2>/dev/null | grep -qi pre-upgrade-early && helm get hooks hooked-app 2>/dev/null | grep -qi pre-upgrade-late`,
				"оба pre-upgrade hook (early/late) с weight добавлены",
				"создай pre-upgrade-early-job.yaml (weight -10) и pre-upgrade-late-job.yaml (weight 10), затем helm upgrade"),
			9: kcheck(`helm get hooks hooked-app 2>/dev/null | grep -qi pre-delete`,
				"pre-delete hook записан в ревизию релиза",
				"добавь pre-delete-job.yaml и helm upgrade hooked-app /root/hooks-chart"),
			10: kcheck(`kubectl get jobs -o name 2>/dev/null | grep -qi pre-delete && ! helm status hooked-app >/dev/null 2>&1`,
				"release удалён, pre-delete hook выполнился",
				"helm uninstall hooked-app (pre-delete Job остаётся, т.к. без hook-succeeded)"),
		},
	},

	// ── Lab 8: Helmfile ── the helmfile binary is baked into the k8s golden; the Setup
	// drops the local chart ./charts/webapp the helmfile references. The student writes
	// values/{dev,staging}.yaml + helmfile.yaml.gotmpl, then renders/syncs/destroys.
	// Checks cover the sync (release deployed + replicas) and the destroy.
	"ch-helm-lab-helmfile": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
export PATH=$PATH:/usr/local/bin
helmfile -f /root/helmfile-lab/helmfile.yaml.gotmpl -e dev destroy >/dev/null 2>&1
helm uninstall web-dev web-staging >/dev/null 2>&1
mkdir -p /root/helmfile-lab/charts/webapp/templates /root/helmfile-lab/values
cat > /root/helmfile-lab/charts/webapp/Chart.yaml <<'EOF'
apiVersion: v2
name: webapp
description: TOT helmfile lab chart
type: application
version: 0.1.0
appVersion: "1.0"
EOF
cat > /root/helmfile-lab/charts/webapp/values.yaml <<'EOF'
replicaCount: 1
environment: dev
image:
  repository: nginx
  tag: alpine
service:
  type: ClusterIP
  port: 8080
  targetPort: 80
EOF
cat > /root/helmfile-lab/charts/webapp/templates/deployment.yaml <<'EOF'
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Release.Name }}-webapp
  labels: {app: {{ .Release.Name }}-webapp}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels: {app: {{ .Release.Name }}-webapp}
  template:
    metadata:
      labels: {app: {{ .Release.Name }}-webapp}
    spec:
      containers:
      - name: web
        image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
        env:
        - {name: ENVIRONMENT, value: "{{ .Values.environment }}"}
        ports:
        - {containerPort: {{ .Values.service.targetPort }}}
EOF
cat > /root/helmfile-lab/charts/webapp/templates/service.yaml <<'EOF'
apiVersion: v1
kind: Service
metadata:
  name: {{ .Release.Name }}-webapp
spec:
  type: {{ .Values.service.type }}
  selector: {app: {{ .Release.Name }}-webapp}
  ports:
  - {port: {{ .Values.service.port }}, targetPort: {{ .Values.service.targetPort }}}
EOF`,
		Checks: map[int]string{
			6: kcheck(hdeployed("web-dev")+` && kubectl get deploy web-dev-webapp >/dev/null 2>&1`,
				"web-dev синхронизирован (helmfile sync -e dev), release deployed",
				"cd /root/helmfile-lab && helmfile -f helmfile.yaml.gotmpl -e dev sync"),
			7: kcheck(hdeployed("web-staging")+` && `+jp("get deploy web-staging-webapp", "{.spec.replicas}", "2"),
				"web-staging синхронизирован с 2 репликами",
				"в values/staging.yaml задай replicaCount: 2, затем helmfile -e staging sync"),
			8: kcheck(jp("get deploy web-dev-webapp", "{.spec.replicas}", "2"),
				"web-dev пересобран с replicaCount: 2",
				"поменяй replicaCount на 2 в values/dev.yaml и helmfile -e dev sync"),
			10: kcheck(`! helm status web-dev >/dev/null 2>&1 && ! helm status web-staging >/dev/null 2>&1`,
				"оба окружения удалены через helmfile destroy",
				"helmfile -f helmfile.yaml.gotmpl -e dev destroy; helmfile ... -e staging destroy"),
		},
	},

	// ── Lab 2: Bitnami charts через (локальный) mirror ── the golden runs an offline
	// Bitnami helm mirror on 127.0.0.1:8879 (baked systemd unit serving a vendored
	// nginx chart); the Setup registers it as the `bitnami` repo. The student updates
	// the index, searches, shows, pulls/unpacks the chart and builds a dependency.
	// Checks verify the downloaded/unpacked artefacts (install/show tasks stay manual).
	"ch-helm-lab-repos": {
		Image: sandboxImageK8s,
		Setup: k8sBoot + `
export PATH=$PATH:/usr/local/bin
rm -rf /root/bitnami-charts /root/bitnami-unpacked /root/bitnami-parent
for i in $(seq 1 20); do helm repo add bitnami http://localhost:8879 >/dev/null 2>&1 && break; sleep 1; done`,
		Checks: map[int]string{
			2: kcheck(`[ -f /root/.cache/helm/repository/bitnami-index.yaml ]`,
				"индекс bitnami repo обновлён (bitnami-index.yaml в кэше)",
				"helm repo update bitnami"),
			6: kcheck(`ls /root/bitnami-charts/nginx-*.tgz >/dev/null 2>&1`,
				"chart bitnami/nginx скачан архивом в /root/bitnami-charts/",
				"helm pull bitnami/nginx --destination /root/bitnami-charts"),
			7: kcheck(`[ -f /root/bitnami-unpacked/nginx/Chart.yaml ]`,
				"chart скачан и распакован в /root/bitnami-unpacked/nginx/",
				"helm pull bitnami/nginx --untar --untardir /root/bitnami-unpacked"),
			9: kcheck(`ls /root/bitnami-parent/charts/nginx-*.tgz >/dev/null 2>&1`,
				"dependency bitnami/nginx собран в charts/ parent-чарта",
				"в /root/bitnami-parent Chart.yaml с dependencies -> helm dependency build /root/bitnami-parent"),
		},
	},
}
