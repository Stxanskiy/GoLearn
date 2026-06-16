package handler

import (
	"encoding/json"
	"html/template"
	"net/http"
	"strconv"

	"github.com/backendraz/golearn/internal/model"
	"github.com/go-chi/chi/v5"
)

type QuizPageData struct {
	PageTitle     string
	Module        *model.Module
	Lesson        *model.Lesson
	Quiz          *model.Quiz
	Questions     []model.QuizQuestion
	QuestionsJSON template.JS
}

type QuizResultData struct {
	PageTitle string
	Module    *model.Module
	Lesson    *model.Lesson
	Score     int
	Total     int
	Percent   int
	Results   []QuestionResult
}

type QuestionResult struct {
	Question    string
	Options     []string
	Selected    int
	Correct     int
	IsCorrect   bool
	Explanation string
}

func (h *Handler) QuizPage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")
	lessonSlug := chi.URLParam(r, "lessonSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lesson, err := h.lessonRepo.GetBySlug(ctx, mod.ID, lessonSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	quiz, questions, err := h.lessonRepo.GetQuiz(ctx, lesson.ID)
	if err != nil || quiz == nil {
		http.NotFound(w, r)
		return
	}

	// Compact payload for client-side one-at-a-time checking.
	type qpayload struct {
		Q string   `json:"q"`
		O []string `json:"o"`
		C int      `json:"c"`
		E string   `json:"e"`
	}
	payload := make([]qpayload, len(questions))
	for i, q := range questions {
		payload[i] = qpayload{Q: q.Question, O: q.Options, C: q.CorrectIndex, E: q.Explanation}
	}
	qjson, _ := json.Marshal(payload) // json.Marshal escapes <,>,& -> safe inside <script>

	data := QuizPageData{
		PageTitle:     "Квиз: " + lesson.Title,
		Module:        mod,
		Lesson:        lesson,
		Quiz:          quiz,
		Questions:     questions,
		QuestionsJSON: template.JS(qjson),
	}

	h.render(w, "quiz", &data)
}

func (h *Handler) SubmitQuiz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	moduleSlug := chi.URLParam(r, "moduleSlug")
	lessonSlug := chi.URLParam(r, "lessonSlug")

	mod, err := h.moduleRepo.GetBySlug(ctx, moduleSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	lesson, err := h.lessonRepo.GetBySlug(ctx, mod.ID, lessonSlug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	_, questions, err := h.lessonRepo.GetQuiz(ctx, lesson.ID)
	if err != nil {
		http.Error(w, "Quiz not found", 404)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Bad form", 400)
		return
	}

	// Check if JSON submission
	contentType := r.Header.Get("Content-Type")
	var answers map[string]int

	if contentType == "application/json" {
		if err := json.NewDecoder(r.Body).Decode(&answers); err != nil {
			http.Error(w, "Bad JSON", 400)
			return
		}
	} else {
		answers = make(map[string]int)
		for key, values := range r.Form {
			if len(values) > 0 {
				if v, err := strconv.Atoi(values[0]); err == nil {
					answers[key] = v
				}
			}
		}
	}

	score := 0
	var results []QuestionResult

	for i, q := range questions {
		key := "q" + strconv.Itoa(i)
		selected, ok := answers[key]
		if !ok {
			selected = -1
		}
		isCorrect := selected == q.CorrectIndex
		if isCorrect {
			score++
		}
		results = append(results, QuestionResult{
			Question:    q.Question,
			Options:     q.Options,
			Selected:    selected,
			Correct:     q.CorrectIndex,
			IsCorrect:   isCorrect,
			Explanation: q.Explanation,
		})
	}

	total := len(questions)
	_ = h.progressRepo.SaveQuizResult(ctx, lesson.ID, score, total)

	percent := 0
	if total > 0 {
		percent = score * 100 / total
	}

	data := QuizResultData{
		PageTitle: "Quiz Results: " + lesson.Title,
		Module:    mod,
		Lesson:    lesson,
		Score:     score,
		Total:     total,
		Percent:   percent,
		Results:   results,
	}

	h.render(w, "quiz_result", &data)
}
