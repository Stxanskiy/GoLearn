package repository

import "strings"

// What a draft carries, declared once per table.
//
// These lists existed six times over, written out inside the SQL of Create and
// mergeDraft, and keeping them in step was left to whoever added a column.
// Nobody did: prerequisites never reached the live course, so an author could
// set "finish Linux first", send the draft for review, see it approved — and
// the requirement was gone. vm_cpus and vm_mem_mib were lost the same way.
//
// The failure was silent in the worst direction: a forgotten column means the
// author's work disappears with no error. So the lists are declared here, the
// SQL is built from them, and draft_columns_test.go fails if a column of the
// table appears in neither this list nor the "deliberately not carried" one.
// Adding a column now forces the choice instead of defaulting to losing it.

// draftModuleCols are the course fields an author edits in a draft.
var draftModuleCols = []string{
	"title", "description", "track", "difficulty", "prerequisites",
	"category", "label", "tags", "cover_image", "accent", "est_minutes", "is_trainer",
}

// moduleNotCarried are the course columns a draft deliberately leaves alone,
// each for its own reason.
var moduleNotCarried = map[string]string{
	"id":          "identity",
	"slug":        "the draft's own slug is ~draft-<id>; the live one never changes",
	"order_num":   "position in the catalogue is not edited in a draft",
	"source":      "a draft is always 'admin'; merging must not relabel a seeded course",
	"published":   "publishing is its own act, not a side effect of approving a draft",
	"owner_id":    "ownership is not transferred by editing",
	"draft_of":    "identity: what makes a draft a draft",
	"access_tier": "what the subscription covers is an admin's call, set on the live course",
	"edited_at":   "set by whoever writes the row, not copied",
	"created_at":  "identity",
}

// draftLessonCols are the lesson fields an author edits in a draft.
var draftLessonCols = []string{
	"slug", "title", "content", "order_num", "difficulty", "track", "kind", "format",
	"vm_image", "vm_init", "vm_cpus", "vm_mem_mib", "published",
}

// lessonNotCarried are the lesson columns a draft deliberately leaves alone.
var lessonNotCarried = map[string]string{
	"id":         "identity",
	"module_id":  "identity: which course the row belongs to",
	"origin_id":  "identity: what links a draft row to the live one",
	"source":     "a lesson added in a draft is 'admin'; merging must not relabel a seeded one",
	"edited_at":  "set by whoever writes the row, not copied",
	"created_at": "identity",
}

// prefixed renders a column list with a table alias: "d.title, d.content, …".
func prefixed(alias string, cols []string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = alias + "." + c
	}
	return strings.Join(out, ", ")
}

// assignments renders an UPDATE's SET body: "title = d.title, …".
func assignments(alias string, cols []string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = c + " = " + alias + "." + c
	}
	return strings.Join(out, ", ")
}

func columnList(cols []string) string { return strings.Join(cols, ", ") }
