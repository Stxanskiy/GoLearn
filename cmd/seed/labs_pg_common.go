package main

// Shared building blocks for the PostgreSQL course labs.
//
// Every lab setup installs glpg (pgcourse/glpg.sh) and starts the server with
// it, and every check is a one-line call into it. Checks have to stay on one
// line: the labcheck harness flattens newlines to spaces before it runs them.
// SQL is passed base64-encoded for the same reason — a reference query full of
// quotes cannot break the shell word it travels in.

import (
	_ "embed"
	"encoding/base64"
	"strings"
)

//go:embed pgcourse/glpg.sh
var glpgScript string

// pgInstallGlpg writes the helper into the sandbox.
var pgInstallGlpg = "echo " + b64(glpgScript) + " | base64 -d > /usr/local/bin/glpg && chmod +x /usr/local/bin/glpg\n"

func b64(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }

// pgSetup is the setup of a lab that needs a running server: install the
// helper, start PostgreSQL (root gets a superuser role, as the course explains
// in its first lab), then run the lab's own preparation.
func pgSetup(extra string) string {
	return "set -e\n" + pgInstallGlpg + "glpg start\n" + strings.TrimSpace(extra) + "\n"
}

// pgRun is a setup step: run SQL in db, failing the setup on the first error.
func pgRun(db, sql string) string {
	return "glpg run " + db + " " + b64(sql) + "\n"
}

// pgFresh drops db if it exists and creates it empty, owned by root.
func pgFresh(db string) string {
	return "dropdb --if-exists --force " + db + "\ncreatedb " + db + "\n"
}

// pgVal is a shell substitution printing a query's value (tuples only, '|'
// between columns). Use it inside conditions and "Сейчас:" diagnostics.
func pgVal(db, sql string) string {
	return `$(glpg q ` + db + ` ` + b64(sql) + `)`
}

// pgIs is a condition: the query's output equals want exactly.
func pgIs(db, sql, want string) string {
	return `[ "` + pgVal(db, sql) + `" = "` + escapeQuotes(want) + `" ]`
}

// pgTrue is a condition: the query returns the boolean true.
func pgTrue(db, sql string) string { return pgIs(db, sql, "t") }

// pgDBExists is a condition: database name exists.
func pgDBExists(name string) string {
	return pgIs("postgres", "SELECT count(*) FROM pg_database WHERE datname = '"+name+"'", "1")
}

// pgRejects is a condition: running sql (in a transaction that is rolled back)
// fails with the given SQLSTATE. This checks what a constraint does, not what
// it is called, so any correct way of writing it passes.
//
//	23502 not_null_violation   23503 foreign_key_violation
//	23505 unique_violation     23514 check_violation
func pgRejects(db, sql, sqlstate string) string {
	return `[ "$(glpg state ` + db + ` ` + b64(sql) + `)" = "` + sqlstate + `" ]`
}

// pgAccepts is a condition: sql runs without error (and is rolled back).
func pgAccepts(db, sql string) string { return pgRejects(db, sql, "00000") }

// pgCheck wraps a glpg comparison that prints its own diagnostic: on success
// the student sees good, on failure good plus what differs. The diagnostic
// never contains the expected rows, only counts and shapes.
func pgCheck(cmd, good string) string {
	return `if d=$(` + cmd + ` 2>&1); then ` + ok(good) +
		`; else echo "✗ не выполнено: ` + escapeQuotes(good) + ` — $d"; exit 1; fi`
}

// pgAnswer checks a query the student saved to file against ref on each of
// dbs (comma-separated). ordered: the rows must also come in ref's order.
func pgAnswer(file, dbs, ref string, ordered bool, good string) string {
	mode := "set"
	if ordered {
		mode = "ordered"
	}
	return pgCheck("glpg answer "+file+" "+dbs+" "+b64(ref)+" "+mode, good)
}

// pgOutfile checks a file the student produced (with psql -At, \o, \copy …)
// against ref's unaligned, tuples-only output.
func pgOutfile(file, db, ref, good string) string {
	return pgCheck("glpg outfile "+file+" "+db+" "+b64(ref), good)
}

// pgColType is the SQL for "type of table.column", as format_type prints it:
// integer, bigint, numeric(8,2), text, boolean, date, timestamp with time zone.
func pgColType(table, column string) string {
	return "SELECT format_type(a.atttypid, a.atttypmod) FROM pg_attribute a " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attname = '" + column + "' AND NOT a.attisdropped"
}

// pgHasCol is a condition: table.column exists with the given type.
func pgHasCol(db, table, column, typ string) string {
	return pgIs(db, pgColType(table, column), typ)
}

// pgIdentityPK is a condition: table.id is an identity column and the primary key.
func pgIdentityPK(db, table string) string {
	return pgTrue(db, "SELECT EXISTS (SELECT 1 FROM pg_attribute a JOIN pg_constraint c "+
		"ON c.conrelid = a.attrelid AND c.contype = 'p' AND c.conkey = ARRAY[a.attnum] "+
		"WHERE a.attrelid = to_regclass('"+table+"') AND a.attname = 'id' AND a.attidentity IN ('a','d'))")
}
