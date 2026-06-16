package handler

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// ── Simulator engine (turn-based business decisions with metric effects) ──

type SimMetric struct {
	Key     string `json:"key"`
	Label   string `json:"label"`
	Unit    string `json:"unit"`
	Start   int    `json:"start"`
	Higher  bool   `json:"higher"` // true: higher is better; false: lower is better
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
	return []Scenario{juniorDevopsScenario()}
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
