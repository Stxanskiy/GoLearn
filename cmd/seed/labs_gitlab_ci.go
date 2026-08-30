package main

// Fixtures + auto-checks for "GitLab CI/CD" (gitlab-ci).
//
// The devops404 original ran these labs against a live GitLab stand: students
// pushed to a real GitLab, watched pipelines run, clicked jobs in the UI, used
// the container registry and a deploy VM. None of that exists in our
// --network none sandbox. So the labs are reframed as *offline authoring* labs:
// the student writes .gitlab-ci.yml in /root/gitlab-ci-lab and each check
// validates the file's structure (grep-based — the base image has no YAML
// parser and no network). Steps that were "push and watch in the UI" become
// "commit the file"; UI/registry/deploy-only steps are reframed to the YAML the
// student would write. Descs override the imported prompts (several of which the
// export also mangled into "$25"-style placeholders).

// ciInit builds the working repository the CI labs assume at /root/gitlab-ci-lab.
const ciInit = `set -e
rm -rf /root/gitlab-ci-lab
git init -q /root/gitlab-ci-lab
git -C /root/gitlab-ci-lab config user.name student
git -C /root/gitlab-ci-lab config user.email student@golearn.local
git -C /root/gitlab-ci-lab symbolic-ref HEAD refs/heads/main
printf '# CI demo project\n' > /root/gitlab-ci-lab/README.md
mkdir -p /root/gitlab-ci-lab/scripts
printf '#!/bin/sh\necho "tests ok"\n' > /root/gitlab-ci-lab/scripts/test.sh
chmod +x /root/gitlab-ci-lab/scripts/test.sh
git -C /root/gitlab-ci-lab add -A
git -C /root/gitlab-ci-lab commit -qm init`

// ci greps the lab's .gitlab-ci.yml (most checks look for keys/jobs in it).
// The pattern is passed with -e so patterns that start with '-' (e.g. a YAML
// list item like "- build-package") are not mistaken for grep options.
func ci(pattern string) string {
	return `grep -Eq -e '` + pattern + `' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null`
}

var gitlabCILabs = map[string]labSpec{
	// ── Lab 1: первый pipeline ──
	"ch-gitlab-ci-lab1": {
		Setup: ciInit,
		Descs: map[int]string{
			1: `<p>В этой песочнице нет живого GitLab — мы учимся <b>писать</b> <code>.gitlab-ci.yml</code>, а проверка смотрит на структуру файла.</p>
<p>В корне репозитория <code>/root/gitlab-ci-lab</code> создай файл <code>.gitlab-ci.yml</code> с одной job, у которой есть <code>script:</code>. Минимум:</p>
<pre>hello:
  script:
    - echo "Hello, CI"</pre>`,
			2: `<p>В настоящем GitLab pipeline запускается после <code>git push</code>. Здесь сети нет, поэтому «запуск» мы имитируем коммитом: закоммить свой <code>.gitlab-ci.yml</code>.</p>
<pre>cd /root/gitlab-ci-lab
git add .gitlab-ci.yml
git commit -m "Add pipeline"</pre>`,
			3: `<p>Добавь вторую job <code>shell-basics</code> с несколькими bash-командами в <code>script</code>:</p>
<pre>shell-basics:
  script:
    - echo "step 1"
    - date
    - ls -la</pre>`,
			4: `<p>Теперь перепиши <code>script</code> в <code>shell-basics</code> на <b>многострочный блок</b> через <code>- |</code>. В YAML это значит «один shell-блок, записанный в несколько строк»:</p>
<pre>shell-basics:
  script:
    - |
      echo "начало"
      for i in 1 2 3; do echo "шаг $i"; done
      echo "конец"</pre>`,
		},
		Checks: map[int]string{
			1: check(`[ -f /root/gitlab-ci-lab/.gitlab-ci.yml ] && `+ci(`^[[:space:]]*script:`),
				".gitlab-ci.yml создан и содержит job со script",
				"создай /root/gitlab-ci-lab/.gitlab-ci.yml с job, у которой есть блок script:"),
			2: check(`git -C /root/gitlab-ci-lab log --oneline -- .gitlab-ci.yml 2>/dev/null | grep -q .`,
				".gitlab-ci.yml закоммичен",
				"cd /root/gitlab-ci-lab && git add .gitlab-ci.yml && git commit -m 'Add pipeline'"),
			3: check(ci(`^shell-basics:`),
				"job shell-basics добавлена",
				"добавь в .gitlab-ci.yml job 'shell-basics:' со script из нескольких команд"),
			4: check(`awk '/^shell-basics:/{f=1} f&&/script:/{s=1} s&&/^[[:space:]]*-[[:space:]]*\|/{print;exit}' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null | grep -q .`,
				"многострочный блок '- |' есть в script",
				"внутри script у shell-basics используй '- |' и напиши несколько строк под ним"),
		},
	},

	// ── Lab 2: stages, variables, artifacts, needs ──
	"ch-gitlab-ci-lab2": {
		Setup: ciInit + `
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
stages:
  - build
  - test
  - validate

variables:
  APP_NAME: ci-demo

build-package:
  stage: build
  script:
    - echo "building"

test-package:
  stage: test
  script:
    - echo "testing"

lint:
  stage: validate
  script:
    - echo "linting"
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline"`,
		Descs: map[int]string{
			1: `<p>Открой <code>/root/gitlab-ci-lab/.gitlab-ci.yml</code>. В блок <code>variables</code> добавь <code>APP_ENV: lab</code> и <code>PACKAGE_NAME: ci-demo-package</code>.</p>`,
			2: `<p>У job <code>build-package</code> сохрани артефакты сборки: добавь <code>artifacts</code> с <code>paths: - dist/</code> и <code>expire_in: 1 hour</code>.</p><pre>build-package:
  stage: build
  script:
    - echo "building"
  artifacts:
    paths:
      - dist/
    expire_in: 1 hour</pre>`,
			3: `<p>Пусть <code>test-package</code> явно зависит от <code>build-package</code> — добавь <code>needs</code>. Так GitLab знает, что тесту нужен результат сборки.</p><pre>test-package:
  stage: test
  needs:
    - build-package
  script:
    - echo "testing"</pre>`,
			4: `<p>Добавь job <code>test-scripts</code> в stage <code>validate</code>, запускающий <code>./scripts/test.sh</code> — он пойдёт параллельно с <code>lint</code>.</p><pre>test-scripts:
  stage: validate
  script:
    - ./scripts/test.sh</pre>`,
			5: `<p>(В оригинале это делалось в UI GitLab — здесь мы объявляем переменную в YAML.) Добавь в глобальный <code>variables</code> ещё <code>CI_DEBUG: "true"</code>.</p>`,
			6: `<p>Artifacts сохраняют файлы между стадиями. Добавь артефакт и у <code>test-package</code>: <code>artifacts: paths: - test-results/</code>.</p>`,
		},
		Checks: map[int]string{
			1: check(ci(`APP_ENV:[[:space:]]*lab`)+` && `+ci(`PACKAGE_NAME:[[:space:]]*ci-demo-package`),
				"переменные APP_ENV и PACKAGE_NAME добавлены",
				"в блок variables впиши APP_ENV: lab и PACKAGE_NAME: ci-demo-package"),
			2: check(ci(`artifacts:`)+` && `+ci(`dist/`)+` && `+ci(`expire_in:[[:space:]]*1 hour`),
				"artifacts.paths: dist/ и expire_in: 1 hour заданы",
				"у build-package добавь artifacts: paths: [dist/], expire_in: 1 hour"),
			3: check(ci(`needs:`)+` && `+ci(`-[[:space:]]*build-package`),
				"явная зависимость needs: build-package добавлена",
				"у test-package добавь needs: [build-package]"),
			4: check(ci(`^test-scripts:`)+` && `+ci(`scripts/test\.sh`),
				"job test-scripts запускает ./scripts/test.sh",
				"добавь job test-scripts (stage: validate) со script: - ./scripts/test.sh"),
			5: check(ci(`CI_DEBUG:`),
				"переменная CI_DEBUG объявлена",
				"в variables впиши CI_DEBUG: \"true\""),
			6: check(ci(`test-results`),
				"артефакт test-results/ добавлен",
				"у test-package добавь artifacts: paths: - test-results/"),
		},
	},

	// ── Lab 3: rules, manual, tag pipelines ──
	"ch-gitlab-ci-lab3": {
		Setup: ciInit + `
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
stages:
  - test
  - deploy
  - release

unit:
  stage: test
  script:
    - echo "tests"
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline"`,
		Descs: map[int]string{
			1: `<p>(В оригинале job запускался кнопкой в UI GitLab.) Добавь job <code>manual-deploy</code> в stage <code>deploy</code> с <code>when: manual</code> — так он не стартует сам, а ждёт ручного запуска.</p><pre>manual-deploy:
  stage: deploy
  when: manual
  script:
    - echo "deploying"</pre>`,
			2: `<p>Теги в Git запускают отдельный tag-pipeline. Создай тег <code>v0.1.0</code> в репозитории:</p><pre>cd /root/gitlab-ci-lab
git tag v0.1.0</pre>`,
			3: `<p>Добавь job <code>release-note</code> в stage <code>release</code>, который печатает тег.</p><pre>release-note:
  stage: release
  script:
    - echo "release $CI_COMMIT_TAG"</pre>`,
			4: `<p>Ограничь <code>release-note</code> так, чтобы он запускался <b>только по тегу</b> — через <code>rules</code> с условием <code>$CI_COMMIT_TAG</code>.</p><pre>  rules:
    - if: '$CI_COMMIT_TAG'</pre>`,
		},
		Checks: map[int]string{
			1: check(ci(`^manual-deploy:`)+` && `+ci(`when:[[:space:]]*manual`),
				"job manual-deploy с when: manual добавлен",
				"добавь manual-deploy (stage: deploy) с when: manual"),
			2: check(`git -C /root/gitlab-ci-lab tag | grep -qx 'v0.1.0'`,
				"тег v0.1.0 создан",
				"cd /root/gitlab-ci-lab && git tag v0.1.0"),
			3: check(ci(`^release-note:`),
				"job release-note добавлен",
				"добавь job release-note в stage release"),
			4: check(`awk '/^release-note:/{f=1} f&&/rules:/{r=1} r&&/CI_COMMIT_TAG/{print;exit}' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null | grep -q .`,
				"release-note ограничен правилом if: $CI_COMMIT_TAG",
				"в release-note добавь rules: - if: '$CI_COMMIT_TAG'"),
		},
	},

	// ── Lab 5: inputs, variables, scoped/environment variables ──
	"ch-gitlab-ci-lab5": {
		Setup: ciInit + `
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
stages:
  - test
  - deploy

unit:
  stage: test
  script:
    - echo "unit"
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline"`,
		Descs: map[int]string{
			1: `<p>(Оригинал просил push в GitLab — здесь просто закоммить изменение.) Добавь строку в <code>README.md</code> и закоммить её.</p><pre>cd /root/gitlab-ci-lab
echo "change" >> README.md
git commit -am "trigger"</pre>`,
			2: `<p>Добавь job <code>full-tests</code>, который запускается только когда переменная <code>RUN_FULL_TESTS</code> равна <code>"true"</code> (через <code>rules</code>).</p><pre>full-tests:
  stage: test
  rules:
    - if: '$RUN_FULL_TESTS == "true"'
  script:
    - echo "full tests"</pre>`,
			3: `<p>(inputs задаются в UI — здесь объявим их в YAML.) Добавь в начало файла блок <code>spec</code> с input <code>deploy_target</code>:</p><pre>spec:
  inputs:
    deploy_target:
      default: stage
---</pre>`,
			4: `<p>Используй input в job: добавь job <code>deploy</code> в stage <code>deploy</code>, который печатает <code>$[[ inputs.deploy_target ]]</code>.</p><pre>deploy:
  stage: deploy
  script:
    - echo "deploy to $[[ inputs.deploy_target ]]"</pre>`,
			5: `<p>Объяви разные переменные для окружений: добавь глобальный <code>variables</code> с <code>DEPLOY_STAGE_URL</code> и <code>DEPLOY_PROD_URL</code>.</p>`,
			6: `<p>Задай scoped-переменную для job <code>deploy</code>: добавь ему <code>variables</code> с <code>DEPLOY_NOTE: "stage deploy"</code>.</p>`,
			7: `<p>Добавь <code>environment</code> в job <code>deploy</code>, чтобы GitLab показывал его в разделе Environments.</p><pre>  environment:
    name: production</pre>`,
			8: `<p>Убедись, что в pipeline есть хотя бы два окружения по смыслу: добавь второй job <code>deploy-stage</code> с <code>environment: name: staging</code>.</p>`,
		},
		Checks: map[int]string{
			1: check(`git -C /root/gitlab-ci-lab log --oneline | wc -l | awk '$1>1{exit 0} {exit 1}'`,
				"есть новый коммит поверх стартового",
				"cd /root/gitlab-ci-lab && echo change >> README.md && git commit -am trigger"),
			2: check(ci(`^full-tests:`)+` && `+ci(`RUN_FULL_TESTS`),
				"job full-tests с правилом на RUN_FULL_TESTS добавлен",
				"добавь full-tests с rules: - if: '$RUN_FULL_TESTS == \"true\"'"),
			3: check(ci(`inputs:`)+` && `+ci(`deploy_target`),
				"spec.inputs.deploy_target объявлен",
				"добавь в начало файла блок spec: inputs: deploy_target: default: stage, потом ---"),
			4: check(ci(`inputs.deploy_target`),
				"input deploy_target используется в job",
				"в job deploy используй $[[ inputs.deploy_target ]]"),
			5: check(ci(`DEPLOY_STAGE_URL`)+` && `+ci(`DEPLOY_PROD_URL`),
				"переменные окружений объявлены",
				"в variables добавь DEPLOY_STAGE_URL и DEPLOY_PROD_URL"),
			6: check(ci(`DEPLOY_NOTE`),
				"scoped-переменная DEPLOY_NOTE добавлена",
				"у job deploy добавь variables: DEPLOY_NOTE: \"stage deploy\""),
			7: check(ci(`environment:`)+` && `+ci(`name:[[:space:]]*production`),
				"environment: production задан",
				"у job deploy добавь environment: name: production"),
			8: check(ci(`name:[[:space:]]*staging`),
				"второе окружение staging добавлено",
				"добавь job deploy-stage с environment: name: staging"),
		},
	},

	// ── Lab 6: include, extends, hidden jobs, parallel matrix ──
	"ch-gitlab-ci-lab6": {
		Setup: ciInit + `
mkdir -p /root/gitlab-ci-lab/ci
cat > /root/gitlab-ci-lab/ci/templates.yml <<'YML'
.alpine-base:
  image: alpine:3.20

.test-template:
  script:
    - echo "template test"
YML
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
include:
  - local: ci/templates.yml

stages:
  - test

unit:
  extends: .test-template
  stage: test
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline with templates"`,
		Descs: map[int]string{
			1: `<p>Создай файл <code>ci/security.yml</code> с job <code>secret-scan</code> и подключи его вторым local include в <code>.gitlab-ci.yml</code>.</p><pre># ci/security.yml
secret-scan:
  stage: test
  script:
    - echo "scanning secrets"</pre><p>А в include добавь строку <code>- local: ci/security.yml</code>.</p>`,
			2: `<p>(Оригинал просил push — здесь закоммить.) Закоммить <code>.gitlab-ci.yml</code> и <code>ci/security.yml</code>.</p><pre>cd /root/gitlab-ci-lab
git add .gitlab-ci.yml ci/security.yml
git commit -m "Add reusable security include"</pre>`,
			3: `<p>Добавь свой скрытый job-шаблон <code>.lint-template</code> (имя с точки — GitLab его не запускает напрямую) и job, наследующий его через <code>extends</code>.</p><pre>.lint-template:
  script:
    - echo "lint"

lint:
  extends: .lint-template
  stage: test</pre>`,
			4: `<p>Добавь job <code>compat-test</code>, который запускается на двух версиях Alpine через <code>parallel: matrix</code>.</p><pre>compat-test:
  stage: test
  image: alpine:$ALPINE_VERSION
  parallel:
    matrix:
      - ALPINE_VERSION: ["3.19", "3.20"]
  script:
    - cat /etc/alpine-release</pre>`,
		},
		Checks: map[int]string{
			1: check(`[ -f /root/gitlab-ci-lab/ci/security.yml ] && grep -q 'secret-scan' /root/gitlab-ci-lab/ci/security.yml && `+ci(`ci/security\.yml`),
				"ci/security.yml создан и подключён вторым include",
				"создай ci/security.yml с job secret-scan и добавь '- local: ci/security.yml' в include"),
			2: check(`git -C /root/gitlab-ci-lab log --oneline | wc -l | awk '$1>1{exit 0} {exit 1}'`,
				"изменения закоммичены",
				"cd /root/gitlab-ci-lab && git add -A && git commit -m 'Add reusable security include'"),
			3: check(ci(`^\.[a-z-]+-template:`)+` && `+ci(`extends:`),
				"скрытый шаблон и extends добавлены",
				"добавь .lint-template (имя с точки) и job с extends: .lint-template"),
			4: check(ci(`parallel:`)+` && `+ci(`matrix:`),
				"parallel: matrix добавлен",
				"добавь compat-test с parallel: matrix и ALPINE_VERSION: [\"3.19\", \"3.20\"]"),
		},
	},

	// ── Lab 7: reports, retry/timeout, resource_group, allow_failure ──
	"ch-gitlab-ci-lab7": {
		Setup: ciInit + `
mkdir -p /root/gitlab-ci-lab/scripts
cat > /root/gitlab-ci-lab/scripts/write-junit.sh <<'SH'
#!/bin/sh
echo '<testsuite tests="1"><testcase name="ok"/></testsuite>' > report.xml
SH
chmod +x /root/gitlab-ci-lab/scripts/write-junit.sh
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
stages:
  - test
  - deploy

download-deps:
  stage: test
  script:
    - echo "downloading"
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline"`,
		Descs: map[int]string{
			1: `<p>Добавь job <code>unit-report</code>, который запускает <code>./scripts/write-junit.sh</code> и сохраняет JUnit-отчёт через <code>artifacts: reports: junit</code>.</p><pre>unit-report:
  stage: test
  script:
    - ./scripts/write-junit.sh
  artifacts:
    reports:
      junit: report.xml</pre>`,
			2: `<p>(Оригинал просил push — здесь закоммить.) Закоммить изменения.</p><pre>cd /root/gitlab-ci-lab
git commit -am "Add junit report"</pre>`,
			3: `<p>У job <code>download-deps</code> добавь <code>retry: 1</code> и <code>timeout: 5 minutes</code> — уместно для временных сбоев.</p>`,
			4: `<p>Добавь manual job <code>deploy-production</code> с <code>environment: production</code> и <code>resource_group: production</code> (чтобы два деплоя не шли одновременно).</p><pre>deploy-production:
  stage: deploy
  when: manual
  environment:
    name: production
  resource_group: production
  script:
    - echo "deploy"</pre>`,
			5: `<p>Добавь job <code>optional-lint</code> <b>без</b> <code>allow_failure</code>, который специально падает (<code>exit 1</code>) — чтобы увидеть обычное поведение: упавший job валит pipeline.</p><pre>optional-lint:
  stage: test
  script:
    - exit 1</pre>`,
			6: `<p>Теперь разреши <code>optional-lint</code> падать, не роняя pipeline — добавь ему <code>allow_failure: true</code>.</p>`,
		},
		Checks: map[int]string{
			1: check(ci(`^unit-report:`)+` && `+ci(`reports:`)+` && `+ci(`junit:`),
				"job unit-report с JUnit-отчётом добавлен",
				"добавь unit-report со script: ./scripts/write-junit.sh и artifacts: reports: junit: report.xml"),
			2: check(`git -C /root/gitlab-ci-lab log --oneline | wc -l | awk '$1>1{exit 0} {exit 1}'`,
				"изменения закоммичены",
				"cd /root/gitlab-ci-lab && git commit -am 'Add junit report'"),
			3: check(ci(`retry:[[:space:]]*1`)+` && `+ci(`timeout:[[:space:]]*5 minutes`),
				"retry: 1 и timeout: 5 minutes заданы",
				"у download-deps добавь retry: 1 и timeout: 5 minutes"),
			4: check(ci(`^deploy-production:`)+` && `+ci(`resource_group:[[:space:]]*production`),
				"deploy-production с resource_group: production добавлен",
				"добавь manual job deploy-production с environment: production и resource_group: production"),
			5: check(ci(`^optional-lint:`),
				"job optional-lint добавлен",
				"добавь optional-lint со script: - exit 1"),
			6: check(`awk '/^optional-lint:/{f=1} f&&/allow_failure:[[:space:]]*true/{print;exit}' /root/gitlab-ci-lab/.gitlab-ci.yml 2>/dev/null | grep -q .`,
				"optional-lint помечен allow_failure: true",
				"у optional-lint добавь allow_failure: true"),
		},
	},

	// ── Lab 4: image build, registry, deploy (reframed offline) ──
	"ch-gitlab-ci-lab4": {
		Setup: ciInit + `
printf 'FROM alpine:3.20\nCMD ["echo","hi"]\n' > /root/gitlab-ci-lab/Dockerfile
printf 'services:\n  app:\n    image: demo\n' > /root/gitlab-ci-lab/docker-compose.yml
cat > /root/gitlab-ci-lab/.gitlab-ci.yml <<'YML'
stages:
  - build
  - deploy

placeholder:
  stage: build
  script:
    - echo "todo"
YML
git -C /root/gitlab-ci-lab add -A && git -C /root/gitlab-ci-lab commit -qm "starter pipeline"`,
		Descs: map[int]string{
			1: `<p>(В оригинале образ уходил в живой GitLab Registry — здесь мы пишем YAML, который это делает.) Добавь job <code>build-image</code> в stage <code>build</code>, который собирает и пушит образ в <code>$CI_REGISTRY_IMAGE</code>.</p><pre>build-image:
  stage: build
  image: docker:27
  services:
    - docker:27-dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:latest .
    - docker push $CI_REGISTRY_IMAGE:latest</pre>`,
			2: `<p>Добавь job <code>deploy-compose</code> в stage <code>deploy</code>, который поднимает сервис через docker compose.</p><pre>deploy-compose:
  stage: deploy
  script:
    - docker compose up -d</pre>`,
			3: `<p>Дай <code>deploy-compose</code> окружение, чтобы GitLab отслеживал деплой: добавь <code>environment: name: production</code>.</p>`,
			4: `<p>Добавь сборку образа с release-тегом: тег образа — <code>$CI_COMMIT_TAG</code>. Добавь <code>rules</code>, чтобы это работало только по тегу.</p><pre>    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG .
  rules:
    - if: '$CI_COMMIT_TAG'</pre>`,
			5: `<p>Для обычного (не тегового) образа добавь в тег номер pipeline — <code>$CI_PIPELINE_ID</code> — чтобы каждый билд был уникальным.</p><pre>    - docker build -t $CI_REGISTRY_IMAGE:$CI_PIPELINE_ID .</pre>`,
		},
		Checks: map[int]string{
			1: check(ci(`^build-image:`)+` && `+ci(`docker build`)+` && `+ci(`CI_REGISTRY_IMAGE`),
				"build-image собирает и пушит образ в registry",
				"добавь build-image со script: docker build/push -t $CI_REGISTRY_IMAGE:latest ."),
			2: check(ci(`^deploy-compose:`)+` && `+ci(`docker compose`),
				"deploy-compose поднимает сервис через compose",
				"добавь deploy-compose со script: docker compose up -d"),
			3: check(ci(`environment:`)+` && `+ci(`name:[[:space:]]*production`),
				"environment: production у деплоя задан",
				"у deploy-compose добавь environment: name: production"),
			4: check(ci(`CI_COMMIT_TAG`),
				"сборка с release-тегом $CI_COMMIT_TAG добавлена",
				"добавь docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_TAG . и rules: - if: '$CI_COMMIT_TAG'"),
			5: check(ci(`CI_PIPELINE_ID`),
				"в тег обычного образа добавлен $CI_PIPELINE_ID",
				"добавь docker build -t $CI_REGISTRY_IMAGE:$CI_PIPELINE_ID ."),
		},
	},
}
