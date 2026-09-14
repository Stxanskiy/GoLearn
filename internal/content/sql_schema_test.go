package content

import (
	"html"
	"testing"
)

func TestExtractSQLSchema(t *testing.T) {
	payload := `{"title":"Shop","tables":[{"id":"users","cols":[{"name":"id","type":"INTEGER","isKey":true,"desc":"PK"}]}],"ddl":"CREATE TABLE users (id INTEGER);","fields":["id"]}`
	raw := `<p>Q</p><div class="sql-erd" data-schema="` + html.EscapeString(payload) + `"></div>`
	s, ok := ExtractSQLSchema(raw)
	if !ok || s.Title != "Shop" || len(s.Tables) != 1 || !s.Tables[0].Cols[0].IsKey || s.Fields[0] != "id" {
		t.Fatalf("schema = %+v ok=%v", s, ok)
	}
	if _, ok := ExtractSQLSchema(`<p>plain</p>`); ok {
		t.Error("no island must return false")
	}
}
