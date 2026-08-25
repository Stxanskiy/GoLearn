package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ── Simulator engine (turn-based business decisions with metric effects) ──

type SimMetric struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Unit   string `json:"unit"`
	Start  int    `json:"start"`
	Higher bool   `json:"higher"` // true: higher is better; false: lower is better
}

type SimChoice struct {
	Text    string         `json:"text"`
	Effects map[string]int `json:"effects"`
	Result  string         `json:"result"`
}

type SimTurn struct {
	Title     string      `json:"title"`
	Situation string      `json:"situation"`
	Choices   []SimChoice `json:"choices"`
}

type Scenario struct {
	Slug    string      `json:"slug"`
	Title   string      `json:"title"`
	Role    string      `json:"role"`
	Icon    string      `json:"icon"`
	Intro   string      `json:"intro"`
	Metrics []SimMetric `json:"metrics"`
	Turns   []SimTurn   `json:"turns"`
}

func scenarios() []Scenario {
	return []Scenario{juniorDevopsScenario(), middleDevopsScenario(), sreScenario()}
}

func baseMetrics(budget, sat, vel, debt int) []SimMetric {
	return []SimMetric{
		{Key: "budget", Label: "Бюджет", Unit: "k₽", Start: budget, Higher: true},
		{Key: "satisfaction", Label: "Удовлетворённость", Unit: "%", Start: sat, Higher: true},
		{Key: "velocity", Label: "Скорость", Unit: "SP", Start: vel, Higher: true},
		{Key: "techdebt", Label: "Техдолг", Unit: "%", Start: debt, Higher: false},
	}
}

func middleDevopsScenario() Scenario {
	return Scenario{
		Slug: "middle-devops", Title: "Middle DevOps: Второй год", Role: "Middle DevOps-инженер", Icon: "📈",
		Intro:   "Ты Middle DevOps в финтех-компании «ФинТехПро»: 50 000 пользователей, платёжная система, строгое SLA и регулятор ЦБ. За год нужно выстроить наблюдаемость, SLO, on-call и зрелые релизы. Нарушение SLA грозит штрафами.",
		Metrics: baseMetrics(4000, 48, 35, 68),
		Turns: []SimTurn{
			{Title: "Наблюдаемость с нуля", Situation: "Есть метрики CPU, но нет понимания пользовательского опыта. Инциденты ловите постфактум.", Choices: []SimChoice{
				{Text: "Внедрить трассировку + RED-метрики (latency/errors/rate) и дашборды", Effects: map[string]int{"budget": -200, "techdebt": -15, "satisfaction": 8}, Result: "Видишь систему глазами пользователя. Инциденты находятся за минуты."},
				{Text: "Добавить ещё инфраструктурных метрик в Zabbix", Effects: map[string]int{"budget": -60, "techdebt": -3}, Result: "Графиков больше, но пользовательский опыт по-прежнему вслепую."},
				{Text: "Отложить — сначала фичи", Effects: map[string]int{"satisfaction": -10, "techdebt": 8}, Result: "Очередной инцидент заметил регулятор. Неприятный разговор."},
			}},
			{Title: "SLO и error budget", Situation: "Бизнес требует 100% аптайма, разработка хочет катить быстрее. Конфликт.", Choices: []SimChoice{
				{Text: "Договориться об SLO 99.9% и error budget, привязать к скорости релизов", Effects: map[string]int{"satisfaction": 10, "techdebt": -10, "velocity": 4}, Result: "Error budget стал общим языком: есть бюджет — катим, кончился — стабилизируем."},
				{Text: "Пообещать 100% аптайма", Effects: map[string]int{"satisfaction": -8, "techdebt": 6}, Result: "Нереалистичное обещание — любой сбой теперь «провал»."},
				{Text: "Не формализовать, решать по ситуации", Effects: map[string]int{"satisfaction": -4}, Result: "Споры про релизы повторяются каждую неделю."},
			}},
			{Title: "On-call процесс", Situation: "Ночью всё чинит один человек по памяти. Выгорание близко.", Choices: []SimChoice{
				{Text: "Ввести ротацию on-call + runbook'и + алерты с приоритетами", Effects: map[string]int{"budget": -80, "satisfaction": 9, "techdebt": -8}, Result: "Дежурства честные, есть инструкции. MTTR падает, люди не выгорают."},
				{Text: "Просто завести чат для алертов", Effects: map[string]int{"satisfaction": -3, "techdebt": 4}, Result: "Алерты тонут в шуме, реагирует всё тот же один человек."},
				{Text: "Оставить как есть", Effects: map[string]int{"satisfaction": -12}, Result: "Ключевой инженер уволился. Знания ушли с ним."},
			}},
			{Title: "Инфраструктура руками", Situation: "Серверы настроены вручную, воспроизвести окружение тяжело.", Choices: []SimChoice{
				{Text: "Перевести инфраструктуру в Terraform (IaC) + ревью изменений", Effects: map[string]int{"budget": -150, "techdebt": -18, "velocity": 3}, Result: "Инфраструктура в git, изменения через PR. Прозрачно и воспроизводимо."},
				{Text: "Написать bash-скрипты для частых операций", Effects: map[string]int{"techdebt": -5}, Result: "Лучше, но дрейф конфигураций остаётся."},
				{Text: "Не трогать — работает же", Effects: map[string]int{"techdebt": 10, "satisfaction": -5}, Result: "Снежинка-сервер упал, восстановление заняло полдня."},
			}},
			{Title: "Релиз платёжной фичи", Situation: "Большое изменение в платежах. Регулятор не прощает простоев.", Choices: []SimChoice{
				{Text: "Blue-green деплой с мгновенным откатом", Effects: map[string]int{"budget": -100, "satisfaction": 10, "techdebt": -4}, Result: "Переключение без даунтайма, проблему откатил мгновенно. SLA соблюдён."},
				{Text: "Catch ночью в окно обслуживания", Effects: map[string]int{"satisfaction": -4, "velocity": -3}, Result: "Сработало, но ночные релизы изматывают команду."},
				{Text: "Прямой деплой в часы нагрузки", Effects: map[string]int{"satisfaction": -15, "techdebt": 8}, Result: "Просадка платежей в пик. Штраф от регулятора."},
			}},
		},
	}
}

func sreScenario() Scenario {
	return Scenario{
		Slug: "senior-sre", Title: "Senior SRE: платформа", Role: "Senior SRE / Platform Lead", Icon: "🛡️",
		Intro:   "Ты Senior SRE в топ-маркетплейсе «МаркетПульс»: 3 млн пользователей, 150 микросервисов, 200+ инженеров. Задача — подготовить платформу к «Чёрной пятнице» (в прошлом году 4 часа даунтайма и потери 180 млн) и построить Internal Developer Platform: GitOps, DevSecOps, Golden Path.",
		Metrics: baseMetrics(5000, 50, 32, 70),
		Turns: []SimTurn{
			{Title: "Готовность к Чёрной пятнице", Situation: "Прошлый год — 4 часа даунтайма в пик. До распродажи 2 месяца.", Choices: []SimChoice{
				{Text: "Нагрузочное тестирование + capacity planning + автоскейлинг", Effects: map[string]int{"budget": -300, "techdebt": -12, "satisfaction": 10}, Result: "Нашли узкие места заранее, настроили HPA. Чёрная пятница прошла без падений."},
				{Text: "Просто добавить серверов «с запасом»", Effects: map[string]int{"budget": -500, "satisfaction": 2}, Result: "Дорого и неэффективно — часть сервисов всё равно деградировала."},
				{Text: "Понадеяться, что в этот раз пронесёт", Effects: map[string]int{"satisfaction": -15, "techdebt": 10}, Result: "Снова даунтайм в пик. Репутационные и денежные потери."},
			}},
			{Title: "Деплой 150 микросервисов", Situation: "Команды катят кто во что горазд, состояние кластера непредсказуемо.", Choices: []SimChoice{
				{Text: "GitOps: ArgoCD, желаемое состояние в git, авто-синк", Effects: map[string]int{"budget": -150, "techdebt": -18, "velocity": 5}, Result: "Кластер = git. Откаты тривиальны, дрейфа нет, аудит из коробки."},
				{Text: "Стандартизировать helm-чарты, катить вручную", Effects: map[string]int{"techdebt": -6}, Result: "Единообразнее, но релизы всё ещё ручные и разнородные."},
				{Text: "Не вмешиваться в процессы команд", Effects: map[string]int{"techdebt": 12, "satisfaction": -6}, Result: "Очередной сервис уронил соседей из-за кривого деплоя."},
			}},
			{Title: "Безопасность цепочки поставок", Situation: "Секреты в образах, уязвимые зависимости, нет сканирования.", Choices: []SimChoice{
				{Text: "DevSecOps: SAST/контейнер-скан в CI + секреты в Vault", Effects: map[string]int{"budget": -180, "techdebt": -14, "satisfaction": 7}, Result: "Уязвимости ловятся до прода, секреты под контролем. Security-incidents падают."},
				{Text: "Раз в квартал ручной аудит безопасности", Effects: map[string]int{"budget": -80, "techdebt": -4}, Result: "Точечно помогает, но между аудитами дыры копятся."},
				{Text: "Отложить до после распродажи", Effects: map[string]int{"satisfaction": -6, "techdebt": 8}, Result: "Утечка ключа в публичный образ — экстренный разбор."},
			}},
			{Title: "Платформа для команд (IDP)", Situation: "Каждая команда заново изобретает CI, деплой и мониторинг.", Choices: []SimChoice{
				{Text: "Построить Golden Path + self-service Internal Developer Platform", Effects: map[string]int{"budget": -250, "techdebt": -16, "velocity": 8}, Result: "Команды поднимают сервис по шаблону за час. Скорость и единообразие выросли."},
				{Text: "Написать подробную документацию", Effects: map[string]int{"techdebt": -4, "velocity": 1}, Result: "Док читают редко, велосипеды продолжаются."},
				{Text: "Пусть каждая команда сама", Effects: map[string]int{"techdebt": 10, "velocity": -3}, Result: "Зоопарк инструментов растёт, поддержка дорожает."},
			}},
			{Title: "Политики инфраструктуры", Situation: "В прод иногда попадают манифесты без лимитов и с привилегиями.", Choices: []SimChoice{
				{Text: "IaC policy as code (OPA/Gatekeeper) в пайплайне", Effects: map[string]int{"budget": -120, "techdebt": -12, "satisfaction": 6}, Result: "Небезопасные манифесты отклоняются автоматически. Прод стал предсказуемым."},
				{Text: "Code review без автоматических проверок", Effects: map[string]int{"techdebt": -3}, Result: "Человеческий фактор пропускает часть нарушений."},
				{Text: "Доверять командам", Effects: map[string]int{"techdebt": 9, "satisfaction": -4}, Result: "Под без лимитов съел ноду — каскадная деградация."},
			}},
		},
	}
}

func findScenario(slug string) *Scenario {
	for _, s := range scenarios() {
		if s.Slug == slug {
			return &s
		}
	}
	return nil
}

func (h *Handler) SimulatorsPage(w http.ResponseWriter, r *http.Request) {
	h.render(w, "simulators", &struct {
		PageTitle string
		Scenarios []Scenario
	}{PageTitle: "Симуляторы — TOT", Scenarios: scenarios()})
}

func (h *Handler) SimulatorPage(w http.ResponseWriter, r *http.Request) {
	s := findScenario(chi.URLParam(r, "slug"))
	if s == nil {
		http.NotFound(w, r)
		return
	}
	data, _ := json.Marshal(s)
	h.render(w, "simulator", &struct {
		PageTitle    string
		Scenario     *Scenario
		ScenarioJSON template.JS
	}{PageTitle: s.Title + " — TOT", Scenario: s, ScenarioJSON: template.JS(data)})
}

// juniorDevopsScenario is authored from the devops404 intro (gameplay is not in
// the static export, so the turns/effects are designed here).
func juniorDevopsScenario() Scenario {
	return Scenario{
		Slug:  "junior-devops",
		Title: "Junior DevOps: Первый год",
		Role:  "Junior DevOps-инженер",
		Icon:  "🚀",
		Intro: "Тебя взяли Junior DevOps в продуктовую компанию «БыстроДев». Команда: 5 разработчиков, Senior DevOps Никита и PM Катя. Никита уходит через 3 месяца — нужно перенять знания и за год закрыть очевидные проблемы: ручные релизы, разные окружения, знания в одной голове, долгая обратная связь, инциденты и безопасность.",
		Metrics: []SimMetric{
			{Key: "budget", Label: "Бюджет", Unit: "k₽", Start: 2000, Higher: true},
			{Key: "satisfaction", Label: "Удовлетворённость", Unit: "%", Start: 65, Higher: true},
			{Key: "velocity", Label: "Скорость", Unit: "SP", Start: 25, Higher: true},
			{Key: "techdebt", Label: "Техдолг", Unit: "%", Start: 75, Higher: false},
		},
		Turns: []SimTurn{
			{
				Title:     "Релиз по инструкции на две страницы",
				Situation: "Деплой делается вручную по длинному README: скопировать файлы, перезапустить сервис, проверить. Раз в неделю кто-нибудь ошибается.",
				Choices: []SimChoice{
					{Text: "Настроить CI/CD пайплайн (GitHub Actions)", Effects: map[string]int{"budget": -150, "techdebt": -15, "velocity": 5}, Result: "Сборка, тесты и деплой автоматизированы. Релизы стали предсказуемыми."},
					{Text: "Оставить как есть — некогда", Effects: map[string]int{"satisfaction": -10, "techdebt": 5}, Result: "Очередная ошибка на релизе. Команда устала от ручной рутины."},
					{Text: "Нанять отдельного релиз-инженера", Effects: map[string]int{"budget": -400, "satisfaction": 3}, Result: "Дорого и не решает корень проблемы — процесс всё ещё ручной."},
				},
			},
			{
				Title:     "«У меня на ноуте работает»",
				Situation: "Баги воспроизводятся только в проде: версии библиотек и ОС у всех разные.",
				Choices: []SimChoice{
					{Text: "Завернуть приложение в Docker, одинаковые образы везде", Effects: map[string]int{"budget": -80, "techdebt": -20, "velocity": 4}, Result: "Окружения стали идентичными. «Работает у меня» больше не аргумент."},
					{Text: "Написать инструкцию по настройке окружения", Effects: map[string]int{"techdebt": -3, "satisfaction": -3}, Result: "Инструкция устаревает через месяц. Дрейф окружений остался."},
					{Text: "Игнорировать — это проблема разработчиков", Effects: map[string]int{"satisfaction": -12, "techdebt": 8}, Result: "Время на отладку «фантомных» багов растёт."},
				},
			},
			{
				Title:     "Все знания — в голове Никиты",
				Situation: "Никита уходит через 3 месяца. Конфиги серверов, доступы и «как чинить» — только у него.",
				Choices: []SimChoice{
					{Text: "Сесть рядом и составить runbook'и по всем системам", Effects: map[string]int{"satisfaction": 8, "techdebt": -18, "velocity": -3}, Result: "Документация и runbook'и сняли зависимость от одного человека."},
					{Text: "Записать пару видео-созвонов «на всякий»", Effects: map[string]int{"techdebt": -6}, Result: "Кое-что зафиксировано, но без структуры искать тяжело."},
					{Text: "Разберусь сам, когда понадобится", Effects: map[string]int{"satisfaction": -8, "techdebt": 12}, Result: "После ухода Никиты простой инцидент превращается в квест."},
				},
			},
			{
				Title:     "Узнаём о падении от клиентов",
				Situation: "Прод падает — и первыми сообщают пользователи в поддержку. Метрик и алертов нет.",
				Choices: []SimChoice{
					{Text: "Поднять Prometheus + Grafana + алерты", Effects: map[string]int{"budget": -120, "techdebt": -15, "satisfaction": 7}, Result: "Теперь о проблемах узнаёшь раньше клиентов. MTTR падает."},
					{Text: "Грепать логи вручную при жалобах", Effects: map[string]int{"satisfaction": -6, "techdebt": 5}, Result: "Реакция медленная, инциденты тянутся часами."},
					{Text: "Купить дорогой APM-сервис без настройки", Effects: map[string]int{"budget": -300, "techdebt": -5}, Result: "Платишь много, но без настроенных дашбордов толку мало."},
				},
			},
			{
				Title:     "Секреты в репозитории",
				Situation: "Аудит показал: пароли БД и API-ключи закоммичены в git в открытом виде.",
				Choices: []SimChoice{
					{Text: "Вынести секреты в Vault / Secrets, почистить историю", Effects: map[string]int{"budget": -90, "techdebt": -12, "satisfaction": 5}, Result: "Секреты под контролем, ротация настроена. Риск утечки снят."},
					{Text: "Перенести в .env и добавить в .gitignore", Effects: map[string]int{"techdebt": -4}, Result: "Лучше, чем было, но старые ключи всё ещё в истории git."},
					{Text: "Отложить — пока не утекло", Effects: map[string]int{"satisfaction": -5, "techdebt": 10}, Result: "Рискованно: один публичный коммит — и инцидент безопасности."},
				},
			},
			{
				Title:     "Первый самостоятельный релиз",
				Situation: "Никита ушёл. Большая фича готова, бизнес ждёт. Как катишь?",
				Choices: []SimChoice{
					{Text: "Canary: 5% трафика, смотрю метрики, потом 100%", Effects: map[string]int{"satisfaction": 10, "techdebt": -5, "velocity": 3}, Result: "Проблему заметил на 5% и откатил без боли. Релиз прошёл гладко."},
					{Text: "Прямой деплой в пятницу вечером", Effects: map[string]int{"satisfaction": -15, "techdebt": 8}, Result: "Конечно, что-то сломалось. Выходные спасали прод."},
					{Text: "Отложить релиз на месяц «для надёжности»", Effects: map[string]int{"satisfaction": -5, "velocity": -5}, Result: "Бизнес недоволен задержкой, а страх релизов только вырос."},
				},
			},
		},
	}
}
