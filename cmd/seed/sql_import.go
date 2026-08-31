package main

import (
	"encoding/json"
	"fmt"
	"html"
	"sort"
	"strings"
)

// ── sql-academy JSON schema ──

type sqlProp struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	IsKey       bool   `json:"isKey"`
	Description string `json:"description"`
}
type sqlTable struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Props       []sqlProp `json:"props"`
}
type sqlDB struct {
	Name        string     `json:"name"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Tables      []sqlTable `json:"tables"`
}
type sqlQuestion struct {
	ID         int      `json:"id"`
	Fields     []string `json:"fields"`
	Difficulty string   `json:"difficulty"`
	Database   sqlDB    `json:"database"`
	Question   string   `json:"question"`
	Title      string   `json:"title"`
}

// sqlAcademyModules builds 3 Database-track modules (by difficulty) from the
// embedded sql-academy export. Each exercise becomes a kind="sql" lesson with an
// interactive in-browser SQLite editor (auto-checks result columns).
func sqlAcademyModules() []M {
	entries, err := contentFS.ReadDir("content/sql_academy/questions")
	if err != nil {
		return nil
	}
	var qs []sqlQuestion
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		data, err := contentFS.ReadFile("content/sql_academy/questions/" + e.Name())
		if err != nil {
			continue
		}
		var q sqlQuestion
		if json.Unmarshal(data, &q) == nil && q.ID > 0 {
			qs = append(qs, q)
		}
	}
	sort.Slice(qs, func(i, j int) bool { return qs[i].ID < qs[j].ID })

	groups := []struct {
		diff, slug, title string
	}{
		{"easy", "sql-easy", "Лёгкие задачи"},
		{"medium", "sql-medium", "Средние задачи"},
		{"hard", "sql-hard", "Сложные задачи"},
	}
	var mods []M
	for _, g := range groups {
		m := M{
			Slug: g.slug, Title: g.title, Track: "database", Category: "Database", Difficulty: g.diff,
			Description: "Практика SQL: реальные задачи со схемами БД и проверкой прямо в браузере.",
		}
		order := 0
		for _, q := range qs {
			if q.difficulty() != g.diff {
				continue
			}
			order++
			m.Lessons = append(m.Lessons, L{
				Slug:    fmt.Sprintf("q-%d", q.ID),
				Title:   sqlTitle(q),
				Content: renderSQL(q),
				Order:   order,
				Track:   "database",
				Kind:    "sql",
			})
		}
		if len(m.Lessons) > 0 {
			mods = append(mods, m)
		}
	}
	return mods
}

func (q sqlQuestion) difficulty() string {
	switch q.Difficulty {
	case "easy", "medium", "hard":
		return q.Difficulty
	default:
		return "medium"
	}
}

func sqlTitle(q sqlQuestion) string {
	if strings.TrimSpace(q.Title) != "" {
		return fmt.Sprintf("№%d. %s", q.ID, q.Title)
	}
	return fmt.Sprintf("Задача №%d", q.ID)
}

// renderSQL builds the lesson HTML: condition + expected columns + schema tables,
// plus a hidden JSON payload (DDL + expected fields) for the in-browser editor.
func renderSQL(q sqlQuestion) string {
	var b strings.Builder
	b.WriteString(`<p>` + html.EscapeString(q.Question) + `</p>`)
	if len(q.Fields) > 0 {
		b.WriteString(`<p><strong>Ожидаемые колонки результата:</strong> `)
		for i, f := range q.Fields {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(`<code>` + html.EscapeString(f) + `</code>`)
		}
		b.WriteString(`</p>`)
	}
	b.WriteString(`<h3>Схема базы данных: ` + html.EscapeString(q.Database.Title) + `</h3>`)

	// Schema + DDL + expected columns travel as a data island on a plain <div>.
	// Unlike <script>, a <div> survives neutralizeActiveHTML, so nothing leaks as
	// raw text. The lesson's front-end reads data-schema to draw the ER diagram
	// AND to seed the in-browser SQLite runner.
	schema := sqlSchemaPayload(q)
	sj, _ := json.Marshal(schema)
	b.WriteString(`<div class="sql-erd" data-schema="` + html.EscapeString(string(sj)) + `"></div>`)
	return b.String()
}

// sqlSchemaPayload is the JSON consumed by the SQL lesson front-end: table
// structure (for the ER diagram) plus the DDL and expected result columns (for
// the in-browser SQLite editor).
func sqlSchemaPayload(q sqlQuestion) map[string]any {
	tables := make([]map[string]any, 0, len(q.Database.Tables))
	for _, t := range q.Database.Tables {
		cols := make([]map[string]any, 0, len(t.Props))
		for _, p := range t.Props {
			cols = append(cols, map[string]any{
				"name":  p.Name,
				"type":  p.Type,
				"isKey": p.IsKey,
				"desc":  p.Description,
			})
		}
		tables = append(tables, map[string]any{"id": t.ID, "cols": cols})
	}
	return map[string]any{
		"title":  q.Database.Title,
		"tables": tables,
		"ddl":    sqlDDL(q.Database),
		"fields": q.Fields,
	}
}

// sqlDDL generates SQLite CREATE TABLE statements from the schema (no data — the
// editor checks result columns, not rows).
func sqlDDL(db sqlDB) string {
	var b strings.Builder
	for _, t := range db.Tables {
		fmt.Fprintf(&b, "CREATE TABLE IF NOT EXISTS \"%s\" (", t.ID)
		var keys []string
		for i, p := range t.Props {
			if i > 0 {
				b.WriteString(", ")
			}
			typ := p.Type
			if typ == "" {
				typ = "TEXT"
			}
			fmt.Fprintf(&b, "\"%s\" %s", p.Name, typ)
			if p.IsKey {
				keys = append(keys, "\""+p.Name+"\"")
			}
		}
		if len(keys) > 0 {
			fmt.Fprintf(&b, ", PRIMARY KEY (%s)", strings.Join(keys, ", "))
		}
		b.WriteString(");\n")
	}
	return b.String()
}
