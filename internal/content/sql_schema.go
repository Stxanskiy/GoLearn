package content

import (
	"encoding/json"
	"html"
	"regexp"
)

// SQLSchema is the practice database embedded in SQL lessons.
type SQLSchema struct {
	Title  string     `json:"title"`
	Tables []SQLTable `json:"tables"`
	DDL    string     `json:"ddl"`
	Fields []string   `json:"fields"` // expected result columns
}

type SQLTable struct {
	ID   string      `json:"id"`
	Cols []SQLColumn `json:"cols"`
}

type SQLColumn struct {
	Name  string `json:"name"`
	Type  string `json:"type"`
	IsKey bool   `json:"isKey"`
	Desc  string `json:"desc"`
}

var sqlSchemaRe = regexp.MustCompile(`<div class="sql-erd" data-schema="([^"]*)"`)

// ExtractSQLSchema reads the `.sql-erd[data-schema]` island from raw lesson content.
func ExtractSQLSchema(raw string) (*SQLSchema, bool) {
	m := sqlSchemaRe.FindStringSubmatch(raw)
	if m == nil {
		return nil, false
	}
	var s SQLSchema
	if err := json.Unmarshal([]byte(html.UnescapeString(m[1])), &s); err != nil {
		return nil, false
	}
	return &s, true
}
