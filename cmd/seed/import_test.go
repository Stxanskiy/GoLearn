package main

import "testing"

func TestCleanContent(t *testing.T) {
	in := `<h1>Linux</h1><p>Текст урока.</p>` +
		"\nf:[\"$\",\"div\",null,{\"className\":\"x\"}]\n28:I[65346,[\"423\"]]\n29:T31bcf,data:image/webp;base64,AAAA"
	got := cleanContent(in)
	if got != `<h1>Linux</h1><p>Текст урока.</p>` {
		t.Errorf("cleanContent did not strip RSC junk:\n%q", got)
	}
	// content without junk is preserved
	clean := `<p>Просто текст</p>`
	if cleanContent(clean) != clean {
		t.Errorf("cleanContent altered clean content: %q", cleanContent(clean))
	}
}

func TestSlugify(t *testing.T) {
	cases := map[string]string{
		"ch_lnav_1_intro": "ch-lnav-1-intro",
		"Some Title":       "some-title",
		"a/b_c":            "a-b-c",
	}
	for in, want := range cases {
		if got := slugify(in, 1); got != want {
			t.Errorf("slugify(%q)=%q want %q", in, got, want)
		}
	}
	if got := slugify("...", 7); got != "ch-7" {
		t.Errorf("slugify empty fallback=%q want ch-7", got)
	}
}

func TestSQLDDL(t *testing.T) {
	db := sqlDB{Tables: []sqlTable{{ID: "users", Props: []sqlProp{
		{Name: "id", Type: "INT", IsKey: true}, {Name: "email", Type: "VARCHAR"},
	}}}}
	ddl := sqlDDL(db)
	want := "CREATE TABLE IF NOT EXISTS \"users\" (\"id\" INT, \"email\" VARCHAR, PRIMARY KEY (\"id\"));\n"
	if ddl != want {
		t.Errorf("sqlDDL=\n%q\nwant\n%q", ddl, want)
	}
}
