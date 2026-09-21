# Вынос фронтенда в `frontend-tot`: инвентаризация и API

Контракт: [`api/openapi.yaml`](../../api/openapi.yaml) (OpenAPI 3.1, линт: `npx @redocly/cli lint api/openapi.yaml --config api/redocly.yaml`).
Фронт генерирует типы: `npx openapi-typescript api/openapi.yaml -o src/shared/api/schema.d.ts`.

## 1. Принципы API

| Решение | Почему |
|---|---|
| Префикс `/api/v1` | Старые `/api/*` нужны Go-шаблонам, пока идёт постепенная миграция; пути не конфликтуют |
| Один origin (Next.js и Go за одним reverse proxy) | cookie `session` остаётся `HttpOnly; SameSite=Lax`, CORS не нужен, WebSocket и iframe превью работают с проверкой Origin |
| `401` JSON вместо `303 /login` | Редирект решает фронт (middleware Next.js) |
| Ошибки `{error: {code, message, details}}` | Тексты переводит фронт по `code` (i18n), сейчас бэкенд отдаёт русские строки |
| CSRF: mutating-запросы только `application/json` + проверка `Origin`/`Sec-Fetch-Site` | Сейчас защиты нет вообще, только SameSite=Lax |
| `*_html` рендерит бэкенд | goldmark + chroma + `neutralizeActiveHTML` есть только в Go; фронт не дублирует |
| GET без побочных эффектов | Next.js делает prefetch и SSR — старый `GET` урока ставил `in_progress`, теперь `POST /lessons/{id}/visit` |
| `/me` вместо cookie `gl_role`/`gl_user` | Клиентские cookie подделываются и не очищаются при логауте/понижении роли |
| Каталог и страницы курсов открыты анонимам (`/public/*`, `security: []`) | SEO и продажи; контент уроков и прогресс остаются за логином |

## 2. Карта страниц

Режим рендера: **SSG/ISR** — публично и кешируемо; **SSR** — данные пользователя на сервере Next; **CSR** — интерактив в браузере.

| Сейчас (Go) | Next.js route | Рендер | Эндпоинты v1 |
|---|---|---|---|
| `/` (аноним) landing | `/` | ISR | `GET /public/landing` |
| `/` (залогинен) dashboard | `/dashboard` (+ редирект с `/` в middleware) | SSR | `GET /me/dashboard`, `GET /catalog` |
| `/login`, `/register`, `/logout` | `/login`, `/register` | SSR+CSR | `GET /auth/config`, `POST /auth/login\|register\|logout` |
| `/profile` | `/profile` | SSR | `GET /me/profile` |
| `/courses` | `/courses` | SSR | `GET /catalog` |
| `/courses/{track}` | `/courses/s/{spec}` * | SSR, фильтры/поиск/сортировка CSR | `GET /specializations/{slug}` |
| `/roadmap` | `/roadmap` | SSR | `GET /catalog` |
| `/module/{m}` | `/courses/{course}` * | SSR | `GET /courses/{slug}` |
| `/module/{m}/lesson/{l}` | `/courses/{course}/{lesson}` | SSR + CSR (квиз, заметки, sql.js) | `GET /courses/{c}/lessons/{l}`, `POST /lessons/{id}/visit\|complete`, `PUT /lessons/{id}/notes`, `POST\|DELETE /lessons/{id}/quiz/answers` |
| `/module/{m}/lesson/{l}/quiz` (+ POST → `quiz_result`) | `/courses/{course}/{lesson}/quiz` | SSR + CSR | `POST /lessons/{id}/quiz/answers`, `POST /lessons/{id}/quiz/submit` |
| `/module/{m}/lesson/{l}/tasks` | `/courses/{course}/{lesson}/lab` | SSR-оболочка + CSR (xterm, Monaco) | `GET /lessons/{id}/lab`, `GET\|DELETE /lessons/{id}/lab/session`, `POST /lessons/{id}/lab/retry`, WS `/lessons/{id}/lab/terminal`, `GET /lessons/{id}/lab/git-graph`, `fs/entries`, `fs/content`, `preview/{port}/*`, `POST /tasks/{id}/check\|done\|run` |
| `/trainers` | `/trainers` | SSR | `GET /catalog` (`trainers`) |
| `/git-trainer` | `/trainers/git` | CSR | `GET\|DELETE /git-trainer/session`, WS `/git-trainer/terminal`, `GET /git-trainer/git-graph` |
| `/playground` | `/trainers/playground` | CSR | `POST /playground/run` |
| `/simulators`, `/simulator/{slug}` | `/simulators`, `/simulators/{slug}` | SSR + CSR (игра) | `GET /simulators`, `GET /simulators/{slug}` |
| `/admin` | `/admin/courses` | CSR | `GET /admin/courses`, `PUT …/published`, `POST …/move`, `DELETE` |
| `/admin/module/new\|{id}` | `/admin/courses/new\|{id}` | CSR | `POST\|GET\|PUT /admin/courses[/{id}]`, `PUT\|DELETE …/cover` |
| `/admin/lesson/…`, `/admin/question/…`, `/admin/task/…` | `/admin/lessons/{id}` (вопросы и задания — диалоги) | CSR | `/admin/lessons/*`, `/admin/questions/*`, `/admin/tasks/*`, `POST /admin/content/preview` |
| `/admin/import` | `/admin/import` | CSR | `POST /admin/import/preview`, `POST /admin/import`, `GET /admin/courses/{id}/export` |
| `/admin/specs` | `/admin/specializations` | CSR | `/admin/specializations/*` |
| `/admin/sims` | `/admin/simulators` | CSR | `/admin/simulators/*` |
| `/admin/users` | `/admin/users` | CSR | `/admin/users/*` |

\* URL меняются → 301-редиректы со старых путей в `next.config` (`/module/:m` → `/courses/:m`, `…/tasks` → `…/lab`).
Нужно решить: `/courses/{spec}` и `/courses/{course}` конфликтуют, отсюда `/courses/s/{spec}` (см. §6).

## 3. Что переезжает с бэкенда на фронт

Чистая презентация — в бэкенде больше не нужна:

- `plural`, `pct`, `chipCat` (цвет чипа категории), `categoryIcon` (эмодзи), `humanDuration` (`~N ч M мин`) → `Intl.PluralRules` / i18n / `shared/lib`.
- `Initial` (буква аватара), `Joined` (формат даты), `RoleLabel` → фронт по `Me`.
- Heatmap: сетка недель, уровни 0–4, подписи месяцев (`buildHeatmap`) — фронт строит из `activity`.
- Категории-фильтры раздела, поиск, сортировка курсов (`gl_course_sort`), вкладки роадмапа (`gl_roadmap_path`), тема и rail (`gl-theme`, `gl-rail`) — уже клиентские.
- Раскладка git-графа по дорожкам (SVG), парсинг ref-ов остаётся на бэке (`GitGraph.refs[].type`).
- Логика симулятора (метрики, вердикт) — уже целиком клиентская.
- `label` отдаётся кодом `start|practice|challenge`, перевод на фронте.

Остаётся в API (бизнес-логика): `categorize()` (fallback-категория), `specForTrack()`, фильтры `published`, расчёт статусов уроков/курсов, `Continue`, соседи prev/next, генерация SVG-обложек, рендер контента, валидация импорта.

## 4. Найденные проблемы

### Безопасность — контракт закрывает, бэкенд надо доработать

| # | Проблема | Сейчас | В v1 |
|---|---|---|---|
| S1 | Правильные ответы квиза уходят в браузер до ответа (`c`, `oe`, `e` в JSON страницы) | lesson/quiz/tasks | `QuizQuestionPublic` без ответов; ответ раскрывается после `POST …/quiz/answers`, первый ответ фиксируется. **Нужна таблица попыток** |
| S2 | `POST /api/shell/{task}/done` засчитывает любую задачу, включая автопроверяемые — обход проверки | shell.go | `409 task_auto_checked` для задач с check/tests |
| S3 | `POST /api/progress/{id}/status` ставит `completed` любому уроку | progress.go | `POST /lessons/{id}/complete` только для theory/sql/sim |
| S4 | `POST /api/run/{task}` возвращает `expected` всех тест-кейсов | run.go | `test_results` без `expected` |
| S5 | Черновики модулей/уроков/разделов открываются по прямому URL | GetBySlug без фильтра | `404` для студентов |
| S6 | `POST /api/shell/{task}/exec` — произвольная команда для любой задачи; нужен только для git-графа | shell.go | Эндпоинта нет; `GET …/git-graph` |
| S7 | Нет CSRF | всё | Origin + JSON-only |
| S8 | Open redirect через `back` в `/admin/module/{id}/publish` | admin.go | Нет редиректов в API |
| S9 | Удалить или понизить можно последнего админа; смена пароля не отзывает сессии | admin_users.go | `409 last_admin`; `PUT …/password` отзывает сессии |
| S10 | HTML квиза (`question`, `options`, `explanation`) не проходит `neutralizeActiveHTML`, вставляется через `innerHTML`/`safeHTML` | quiz | Все `*_html` через общий санитайзер |
| S11 | SVG обложки раздела: `Icon` не экранируется | covers.go | Экранировать |
| S12 | Нет rate limit на run/exec/check/terminal | — | `429` в контракте для playground; лимиты на sandbox-эндпоинты |

### Потеря данных в админке — в v1 формы шлют полные объекты

- Сохранение урока затирает `vm_init` (нет поля в форме).
- Сохранение вопроса затирает `option_explanations` (пишется `null`).
- Сохранение задания затирает `glossary` и `test_cases`.
- Сохранение курса затирает `accent`; если `track` не совпадает ни с одним разделом, select молча меняет его на первый.
- Импорт удаляет submissions студентов при замене задач — в `ImportPreview` добавлен `lost_submissions`.
- Удаление раздела оставляет курсы-сироты → `409 specialization_not_empty`.

### Баги и расхождения

- **Статус урока считается тремя способами:** raw `progress.status` (dashboard, courses, roadmap); с учётом `quiz_score` / прохождения лабы (страница курса и урока); у роадмапа `in_progress` только при `completed > 0`. В v1 одна функция на бэке для всех ответов.
- `Overview.Simulators`/`Trainers` всегда 0, `SimulatorsTot` = 3 захардкожено, прохождение симулятора не сохраняется.
- Статистика лендинга и `ArticlesTotal` учитывают неопубликованное.
- Терминал: нажатия клавиш не продлевают idle TTL → лабу, где работают только в терминале, прибивает через 30 мин. В контракте: keystrokes продлевают TTL + ping каждые 30 с.
- Таймер сессии чисто клиентский и сбрасывается при перезагрузке → `SandboxSession.expires_at`.
- `fs/read` падает на файлах > ~48 КБ (обрезка до base64-декода); превью — на ответах > ~48 КБ. В v1 `fs/content` отдаёт сырые байты.
- Превью не пробрасывает `Location` на 3xx.
- Открытие git-тренажёра убивает VM открытой лабы (одна VM на пользователя) — поведение оставлено, но задокументировано.
- `Home` не чистит невалидную сессию и не делает ADMIN_EMAILS-промоут.
- Понижение email из `ADMIN_EMAILS` не держится → `409 role_pinned_by_env`.

## 5. Реализация на бэкенде (порядок)

1. **Каркас v1:** `internal/api` (роутер `/api/v1`), `writeError` с кодами, `RequireUser` с 401, CSRF-middleware, `GET /me`, `auth/*`. Сюда же — генерация Go-типов из спеки (`oapi-codegen`, режим `chi-server` strict) и тест, что роутер совпадает со спекой.
2. **Витрина:** `public/landing`, `catalog`, `specializations/{slug}`, `courses/{slug}`, обложки. Унифицировать расчёт статусов (§4).
3. **Уроки и квиз:** миграция `quiz_attempt_answers`, `lessons/*`, `quiz/*`.
4. **Лабы:** `lab`, `session`, WS-терминал с новым префиксом, `git-graph`, `fs/*` без base64-ограничения, `tasks/*` с правилами S2/S4.
4. **Админка:** CRUD JSON, cover upload, импорт с `lost_submissions`, пользователи с защитой последнего админа.
6. Удалить `html/template`, `internal/templates`, `internal/static` (кроме того, что переедет во фронт), старые `/api/*` — после переноса всех страниц.

Фронт стартует параллельно с шага 1: лендинг и логин можно делать сразу, остальное проксируется в Go через `rewrites` fallback.

## 6. Открытые вопросы

1. **Владение контентом в админке.** Сейчас любой админ правит и удаляет чужие курсы/черновики, `owner_id` только скрывает чужие черновики в списке. Оставить «все админы равны» или ввести роли (`admin` / `author`)? Для коммерческого продукта с внешними авторами — второе.
2. **Структура URL каталога:** `/courses/s/{spec}` vs `/tracks/{spec}` + `/courses/{course}`. Предлагаю второе — чище для SEO.
3. **Go-задачи и playground.** Go-курсы удалены (миграция `013`), но `kind=go`, `/run` и playground живы. Переносить playground или удалить? В контракте `tasks/{id}/run` помечен `deprecated`.
4. **Сброс квиза.** «Заново» стирает зафиксированные ответы → можно подсмотреть ответы и пересдать. Ограничить число попыток / считать первый результат?
5. **Сохранять прохождение симуляторов** (сейчас нигде не пишется) — нужен ли `POST /simulators/{slug}/runs`?
6. **Оплата/подписки** — на профиле статичный блок «Бесплатный доступ». Если монетизация планируется, её лучше заложить в модель доступа (`GET /me` → `entitlements`) до переноса страниц курсов.

## 7. Статика и зависимости фронта

| Сейчас | В `frontend-tot` |
|---|---|
| `internal/static/vendor/xterm` (+ fit addon) | `@xterm/xterm`, `@xterm/addon-fit` из npm, `dynamic(…, { ssr: false })` |
| `internal/static/vendor/monaco` (AMD, ~50 МБ) | `@monaco-editor/react`, ленивая загрузка |
| `internal/static/vendor/sqljs` | `sql.js` из npm, `sql-wasm.wasm` в `public/` |
| Google Fonts (Inter, JetBrains Mono) | `next/font` (self-host, без внешнего CSS) |
| Phosphor icons с unpkg | `@phosphor-icons/react` |
| `app.css` (334 строки, свои токены) | Перенести токены в дизайн-систему фронта |

CSP фронта: добавить `wasm-unsafe-eval` (sql.js) и `worker-src blob:` (Monaco) — сейчас их нет, и в строгих браузерах это ломается.
