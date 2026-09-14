package api

import (
	"context"
	"net/http"

	"github.com/backendraz/golearn/internal/api/apigen"
	"github.com/backendraz/golearn/internal/content"
	"github.com/backendraz/golearn/internal/model"
)

// quizQuestions returns the lesson's quiz questions; a lesson without a quiz has none.
func (a *API) quizQuestions(ctx context.Context, lessonID int) ([]model.QuizQuestion, error) {
	_, questions, err := a.Lessons.GetQuiz(ctx, lessonID)
	if isNotFound(err) {
		return nil, nil
	}
	return questions, err
}

func (a *API) answerQuizQuestion(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	var body apigen.AnswerQuizQuestionJSONRequestBody
	if !decodeJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	questions, err := a.quizQuestions(ctx, ref.lesson.ID)
	if err != nil {
		a.internalError(w, "answer: questions", err)
		return
	}
	q, found := findQuestion(questions, body.QuestionID)
	if !found {
		writeError(w, http.StatusNotFound, codeNotFound, "question not found in this lesson")
		return
	}
	if body.SelectedIndex < 0 || body.SelectedIndex >= len(q.Options) {
		writeValidation(w, map[string]string{"selected_index": fieldInvalidFormat})
		return
	}
	stored, created, err := a.QuizAnswers.Record(ctx, userFrom(ctx).ID, q.ID, body.SelectedIndex)
	if err != nil {
		a.internalError(w, "answer: record", err)
		return
	}
	writeJSON(w, http.StatusOK, answerResult(q, stored, !created))
}

func (a *API) resetQuizAttempt(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	if err := a.QuizAnswers.ResetLesson(ctx, userFrom(ctx).ID, ref.lesson.ID); err != nil {
		a.internalError(w, "reset quiz", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (a *API) submitQuiz(w http.ResponseWriter, r *http.Request) {
	ref, ok := a.lessonByID(w, r)
	if !ok {
		return
	}
	var body apigen.SubmitQuizJSONRequestBody
	if !decodeOptionalJSON(w, r, &body) {
		return
	}
	ctx := r.Context()
	uid := userFrom(ctx).ID
	questions, err := a.quizQuestions(ctx, ref.lesson.ID)
	if err != nil {
		a.internalError(w, "submit: questions", err)
		return
	}
	if len(questions) == 0 {
		writeError(w, http.StatusNotFound, codeNotFound, "lesson has no quiz")
		return
	}

	if body.Answers != nil {
		fields := map[string]string{}
		for _, in := range *body.Answers {
			q, found := findQuestion(questions, in.QuestionID)
			if !found || in.SelectedIndex < 0 || in.SelectedIndex >= len(q.Options) {
				fields["answers"] = fieldInvalidFormat
				break
			}
		}
		if len(fields) > 0 {
			writeValidation(w, fields)
			return
		}
		for _, in := range *body.Answers {
			if _, _, err := a.QuizAnswers.Record(ctx, uid, in.QuestionID, in.SelectedIndex); err != nil {
				a.internalError(w, "submit: record", err)
				return
			}
		}
	}

	answers, err := a.QuizAnswers.ForLesson(ctx, uid, ref.lesson.ID)
	if err != nil {
		a.internalError(w, "submit: answers", err)
		return
	}
	out := apigen.QuizResult{Total: len(questions), Results: make([]apigen.QuizQuestionResult, 0, len(questions))}
	for _, q := range questions {
		res := apigen.QuizQuestionResult{
			QuestionID:      q.ID,
			QuestionHTML:    content.Sanitize(q.Question),
			OptionsHTML:     sanitizeAll(q.Options),
			CorrectIndex:    q.CorrectIndex,
			ExplanationHTML: optionalHTML(q.Explanation),
		}
		if sel, ok := answers[q.ID]; ok {
			res.SelectedIndex = &sel
			res.IsCorrect = sel == q.CorrectIndex
		}
		if res.IsCorrect {
			out.Score++
		}
		out.Results = append(out.Results, res)
	}
	out.Percent = out.Score * 100 / out.Total
	if err := a.Progress.SaveQuizResult(ctx, uid, ref.lesson.ID, out.Score, out.Total); err != nil {
		a.internalError(w, "submit: save result", err)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

func findQuestion(questions []model.QuizQuestion, id int) (model.QuizQuestion, bool) {
	for _, q := range questions {
		if q.ID == id {
			return q, true
		}
	}
	return model.QuizQuestion{}, false
}

// answerResult reveals the correct option and explanations for an answered question.
func answerResult(q model.QuizQuestion, selected int, locked bool) apigen.QuizAnswerResult {
	res := apigen.QuizAnswerResult{
		QuestionID:      q.ID,
		SelectedIndex:   selected,
		CorrectIndex:    q.CorrectIndex,
		IsCorrect:       selected == q.CorrectIndex,
		ExplanationHTML: optionalHTML(q.Explanation),
		Locked:          locked,
	}
	if len(q.OptionExpl) > 0 {
		expl := sanitizeAll(q.OptionExpl)
		res.OptionExplanationsHTML = &expl
	}
	return res
}

func optionalHTML(s string) *string {
	if s == "" {
		return nil
	}
	h := content.Sanitize(s)
	return &h
}
