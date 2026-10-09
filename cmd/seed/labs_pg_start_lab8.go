package main

// Lab 8 of pg-start: the course project, a database for a helpdesk
// (ch-pgs-lab8).
//
// Lesson 18 designs a school library from what the librarian says: entities,
// attributes and types, links, rules, a schema file run with \i, then good and
// bad rows to see each rule work. Here the student does the same alone, for the
// support desk of an online shop of board games. The intro gives the
// requirements in words and the names the checks look for — tables and columns
// — but no types and no rules: choosing them is the project. The setup only
// starts the server: a new one, with no helpdesk database and no pereplet.
//
// The design the requirements describe (never shown to the student):
//
//	clients (id, full_name NOT NULL, email NOT NULL UNIQUE, phone text)
//	agents  (id, full_name NOT NULL, email NOT NULL UNIQUE,
//	         is_active boolean NOT NULL DEFAULT true)
//	tickets (id, client_id NOT NULL → clients, agent_id → agents (may be empty),
//	         subject NOT NULL,
//	         priority NOT NULL DEFAULT 'normal', one of low, normal, high,
//	         status   NOT NULL DEFAULT 'open',   one of open, in_progress, closed,
//	         created_at date NOT NULL DEFAULT current_date, closed_at date,
//	         closed_at not before created_at)
//
// The table checks look at behaviour, never at how a rule is written. A table
// must have exactly the columns named in the intro (any types); then every rule
// is tried with rows in a transaction that is rolled back (glpg state) and the
// SQLSTATE says whether the database did what the requirements ask. So serial
// and identity ids pass alike, a rule may sit on its column or on its own line,
// with or without a name, CHECK … IN as well as a list of ORs, and an ENUM
// type passes for priority and status (it refuses a bad value with 22P02 where
// CHECK gives 23514). Types are tried the same way: a phone like
// +7 912 345-67-89 has to go in (not integer), '17 марта' must not go in as a
// date (not text) and a time must not be kept (date, not timestamp). A default
// is read back after an INSERT that leaves the column out: SELECT 1/(x IS NOT
// NULL)::int fails with 22012 when it stays empty, ln((x = want)::int) with
// 2201E when it is something else — non-constant on purpose, because the
// planner folds a constant 1/0 and fails before reading a row.
//
// Trial rows use ids from 30001 (OVERRIDING SYSTEM VALUE, which a serial
// column ignores and a GENERATED ALWAYS one needs; small enough for a
// smallserial id, which is not taught but is a correct one) and emails at
// helpdesk.test, so they never meet the student's rows and the checks pass in
// any order and before or after the data task. Only the one rule per table
// that tests numbering draws from the student's sequence, and it puts the
// sequence back where it was (pgsLab8KeepSeq): pressing "Проверить" leaves no
// gap, and the student's first client is still number 1. Trial tickets come
// with a trial client (30001) and agent (30002): a ticket with client_id
// 30002 or agent_id 30001 tells a link that points to the wrong table. Every
// trial ticket that is closed has a closing date and no other one has, so a
// student's extra rule tying the two together does not fail them.
//
// Every trial row of a table is one base row — everything the requirements
// tell about a client, an agent or a new ticket — with one thing changed. A
// second client or agent differs from the first in every column, and a second
// ticket in its id and subject (client and agent it shares, as the
// requirements allow), so only the rule being tried can refuse it; a UNIQUE
// nobody asked for — on a name, on a subject — has rules of its own (тёзки
// and repeated subjects go in). The base row is tried first, as a gate: when
// it does not go in, nothing else is tried, because every other rule would
// fail for the same reason and the student would read a list of things that
// are not so ("база принимает двух клиентов с одним email" about a table that
// took no client at all). A closed ticket leads the rules about closing the
// same way: when it does not go in, the status list or the rule on the two
// dates is wrong, and the rest of that group is not tried.
//
// When a check fails, the student reads what the database does now, in the
// words of the requirements — "клиента без email база принимает" — up to five
// rules at a time, then the columns as \d lists them, a table made in the
// database postgres (\i run before \c helpdesk) and an open transaction.
//
// The data task wants exactly the rows it lists, matched by email and subject,
// and ticket links by the email of the row they point to, so the ids the
// student's sequences hand out do not matter. The ticket that came "today"
// takes created_at from the default; a day either way is accepted for a
// student whose clock is not the server's.
//
// The quizzes ask why phone is text, why the link lives in tickets and why
// agent_id may be empty; their answers follow from lessons 13, 14 and 18, not
// from data.
//
// Reference solutions: scripts/labcheck/solutions-pg-lab8.sh.

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"sort"
	"strconv"
	"strings"
)

func init() {
	pgStartLabs["ch-pgs-lab8"] = labSpec{
		Image: sandboxImagePG,
		Setup: pgsLab8Setup,
		Checks: map[int]string{
			1: check(pgDBExists("helpdesk"),
				"в PostgreSQL есть база helpdesk",
				"Сейчас: "+pgVal("postgres", pgsLab8DBs)),
			2: pgsLab8TableCheck("clients", pgsLab8ClientCols, nil, pgsLab8ClientRules,
				"в базе helpdesk есть таблица clients, и она ведёт себя так, как сказано в требованиях о клиентах"),
			3: pgsLab8TableCheck("agents", pgsLab8AgentCols, nil, pgsLab8AgentRules,
				"в базе helpdesk есть таблица agents, и она ведёт себя так, как сказано в требованиях о сотрудниках"),
			4: pgsLab8TableCheck("tickets", pgsLab8TicketCols, pgsLab8People, pgsLab8TicketRules,
				"в базе helpdesk есть таблица tickets, и она ведёт себя так, как сказано в требованиях об обращениях"),
			5: pgCheck(pgsLab8Run(pgsLab8DataScript),
				"в базе helpdesk ровно данные из задания: 2 клиента, 1 сотрудник и 3 обращения"),
		},
	}
}

// pgsLab8Setup starts the server and leaves it as a new one: no database but
// the service ones, and none of the project's tables in postgres, where a
// student who ran the file before \c helpdesk leaves them. A setup run again
// over a used sandbox so starts from the state the intro describes.
var pgsLab8Setup = pgSetup(
	"glpg q postgres " + b64("SELECT datname FROM pg_database WHERE NOT datistemplate AND datname <> 'postgres'") +
		" | while IFS= read -r db; do dropdb --force \"$db\"; done\n" +
		pgRun("postgres", "SET client_min_messages = warning;\nDROP TABLE IF EXISTS tickets, agents, clients CASCADE;\n"))

// pgsLab8DBs says which databases there are, for the first task.
const pgsLab8DBs = "SELECT 'базы helpdesk нет, базы на сервере: ' || string_agg(datname, ', ' ORDER BY datname) " +
	"FROM pg_database WHERE NOT datistemplate"

// The columns of each table, in the order the intro lists them.
var (
	pgsLab8ClientCols = []string{"id", "full_name", "email", "phone"}
	pgsLab8AgentCols  = []string{"id", "full_name", "email", "is_active"}
	pgsLab8TicketCols = []string{"id", "client_id", "agent_id", "subject", "priority", "status", "created_at", "closed_at"}
)

// pgsLab8Rule is one thing a table has to do: sql, run in a transaction that
// is rolled back, ends with one of the SQLSTATEs in want. say is what the
// student is told when it does not — what the database does now, in the words
// of the requirements; by overrides it for particular SQLSTATEs. A gate rule
// is a row every other rule builds on: when it fails, the rules after it are
// not tried, because they would fail for the same reason and say something
// untrue about it. group does the same for a few rules about one thing: when
// the lead of a group fails, the other rules of the group are not tried.
type pgsLab8Rule struct {
	sql   string
	want  []string
	say   string
	by    map[string]string
	gate  bool
	group string
	lead  bool
}

// SQLSTATEs a rule may want.
var (
	pgsLab8OK       = []string{"00000"}
	pgsLab8NotNull  = []string{"23502"}
	pgsLab8Unique   = []string{"23505"}
	pgsLab8FK       = []string{"23503"}
	pgsLab8Check    = []string{"23514"}
	pgsLab8BadValue = []string{"23514", "22P02"} // CHECK, or an ENUM type
	pgsLab8BadDate  = []string{"22007"}
)

// pgsLab8Row is an INSERT of the trial row base into table, with columns
// changed: set holds pairs of column and SQL value, and the value "-" leaves
// the column out of the INSERT, so its default applies. OVERRIDING SYSTEM
// VALUE lets an explicit id into an identity column and changes nothing for
// serial.
func pgsLab8Row(table string, base [][2]string, set ...string) string {
	over := map[string]string{}
	for i := 0; i+1 < len(set); i += 2 {
		over[set[i]] = set[i+1]
	}
	var cols, vals []string
	for _, c := range base {
		v := c[1]
		if o, ok := over[c[0]]; ok {
			v = o
		}
		if v == "-" {
			continue
		}
		cols = append(cols, c[0])
		vals = append(vals, v)
	}
	return "INSERT INTO " + table + " (" + strings.Join(cols, ", ") + ") OVERRIDING SYSTEM VALUE VALUES (" + strings.Join(vals, ", ") + ")"
}

// pgsLab8KeepSeq is sql, which leaves id out and so draws from the sequence
// behind table's id, followed by putting that sequence back. A sequence does
// not roll back with the transaction; setval does not either, so it undoes the
// draw. pg_get_serial_sequence finds the sequence of serial and identity alike
// (NULL for an id without one, when there is nothing to put back), and
// pg_sequences says where it stood: last_value is NULL before its first use.
func pgsLab8KeepSeq(table, sql string) string {
	return "CREATE TEMP TABLE gl_seq AS SELECT s.seq, p.last_value, p.start_value " +
		"FROM (SELECT pg_get_serial_sequence('" + table + "', 'id') AS seq) s " +
		"LEFT JOIN pg_sequences p ON format('%I.%I', p.schemaname, p.sequencename) = s.seq; " +
		sql + "; SELECT setval(seq, coalesce(last_value, start_value), last_value IS NOT NULL) FROM gl_seq WHERE seq IS NOT NULL"
}

// pgsLab8ClientBase is the trial client, everything the requirements tell
// about one; pgsLab8Client is it with columns changed (see pgsLab8Row).
var pgsLab8ClientBase = [][2]string{
	{"id", "30001"},
	{"full_name", "'Проверка Один'"},
	{"email", "'check-1@helpdesk.test'"},
	{"phone", "'+7 912 345-67-89'"},
}

func pgsLab8Client(set ...string) string { return pgsLab8Row("clients", pgsLab8ClientBase, set...) }

// pgsLab8Client2 are the changes that make the trial client a second, other
// one: a second row differs in every column a student may have made unique,
// so only the rule being tried can refuse it.
var pgsLab8Client2 = []string{"id", "30002", "full_name", "'Проверка Два'", "email", "'check-2@helpdesk.test'", "phone", "'+7 495 123-45-67'"}

// pgsLab8With is set2 with the pairs of set over it.
func pgsLab8With(set2 []string, set ...string) []string {
	return append(append([]string(nil), set2...), set...)
}

var pgsLab8ClientRules = []pgsLab8Rule{
	{sql: pgsLab8Client(), want: pgsLab8OK, gate: true,
		say: "правильного клиента — с именем, email и телефоном +7 912 345-67-89 — в clients добавить не получается: посмотри на таблицу, \\d clients",
		by: map[string]string{
			"22P02": "правильные значения не записываются в столбцы clients — проверь типы: телефон вида +7 912 345-67-89 должен записываться",
			"22001": "правильные значения не помещаются в столбцы clients — проверь типы: телефон вида +7 912 345-67-89 должен записываться",
			"23514": "правильного клиента — с именем, email и телефоном +7 912 345-67-89 — не пропускают твои правила CHECK: посмотри на них, \\d clients",
		}},
	{sql: pgsLab8KeepSeq("clients", pgsLab8Client("id", "-")+"; "+pgsLab8Client(pgsLab8With(pgsLab8Client2, "id", "-")...)), want: pgsLab8OK,
		say: "двух клиентов без id в clients добавить не получается: посмотри на таблицу, \\d clients",
		by: map[string]string{
			"23502": "клиента без номера id база не принимает — номер должна ставить сама база, как во всех таблицах курса",
			"23505": "двух клиентов, добавленных без id, база не принимает — номер id должна ставить сама база, каждому свой",
		}},
	{sql: pgsLab8Client("phone", "NULL"), want: pgsLab8OK,
		say: "клиента без телефона база не принимает — а телефон по требованиям оставляют не все"},
	{sql: pgsLab8Client("full_name", "NULL"), want: pgsLab8NotNull,
		say: "клиента без имени (пустой full_name) база принимает — а имя клиента по требованиям нужно знать обязательно"},
	{sql: pgsLab8Client("email", "NULL"), want: pgsLab8NotNull,
		say: "клиента без email база принимает — а email клиента по требованиям нужно знать обязательно"},
	{sql: pgsLab8Client() + "; " + pgsLab8Client(pgsLab8With(pgsLab8Client2, "email", "'check-1@helpdesk.test'")...), want: pgsLab8Unique,
		say: "двух клиентов с одинаковым email база принимает — а по требованиям так быть не может"},
	{sql: pgsLab8Client() + "; " + pgsLab8Client(pgsLab8With(pgsLab8Client2, "full_name", "'Проверка Один'")...), want: pgsLab8OK,
		say: "двух клиентов с одинаковым именем и разными email база не принимает — а тёзки бывают: человека узнают по email, а не по имени"},
	{sql: pgsLab8Client() + "; " + pgsLab8Client(pgsLab8With(pgsLab8Client2, "id", "30001")...), want: pgsLab8Unique,
		say: "двух клиентов с одним номером id база принимает — id должен быть первичным ключом"},
	{sql: pgsLab8Client("id", "NULL"), want: pgsLab8NotNull,
		say: "клиента с пустым id база принимает — id должен быть первичным ключом"},
}

// pgsLab8AgentBase is the trial agent, at work; pgsLab8Agent is it with
// columns changed, pgsLab8Agent2 makes it a second, other one.
var pgsLab8AgentBase = [][2]string{
	{"id", "30001"},
	{"full_name", "'Проверка Один'"},
	{"email", "'check-1@helpdesk.test'"},
	{"is_active", "true"},
}

func pgsLab8Agent(set ...string) string { return pgsLab8Row("agents", pgsLab8AgentBase, set...) }

var pgsLab8Agent2 = []string{"id", "30002", "full_name", "'Проверка Два'", "email", "'check-2@helpdesk.test'"}

var pgsLab8AgentRules = []pgsLab8Rule{
	{sql: pgsLab8Agent(), want: pgsLab8OK, gate: true,
		say: "правильного сотрудника — с именем, email и отметкой true в is_active — в agents добавить не получается: посмотри на таблицу, \\d agents",
		by: map[string]string{
			"42804": "в is_active не записывается значение true — подумай, какой тип хранит «да» или «нет»",
			"22P02": "в is_active не записывается значение true — подумай, какой тип хранит «да» или «нет»",
			"23514": "правильного сотрудника — с именем, email и отметкой true в is_active — не пропускают твои правила CHECK: посмотри на них, \\d agents",
		}},
	{sql: pgsLab8KeepSeq("agents", pgsLab8Agent("id", "-")+"; "+pgsLab8Agent(pgsLab8With(pgsLab8Agent2, "id", "-")...)), want: pgsLab8OK,
		say: "двух сотрудников без id в agents добавить не получается: посмотри на таблицу, \\d agents",
		by: map[string]string{
			"23502": "сотрудника без номера id база не принимает — номер должна ставить сама база, как во всех таблицах курса",
			"23505": "двух сотрудников, добавленных без id, база не принимает — номер id должна ставить сама база, каждому свой",
		}},
	{sql: pgsLab8Agent("is_active", "false"), want: pgsLab8OK,
		say: "уволенного сотрудника (is_active — false) база не принимает — а по требованиям уволенных не удаляют, а отмечают"},
	{sql: pgsLab8Agent("is_active", "-") +
		"; SELECT 1/(is_active IS NOT NULL)::int FROM agents WHERE id = 30001" +
		"; SELECT ln((is_active IS TRUE)::int) FROM agents WHERE id = 30001",
		want: pgsLab8OK,
		say:  "сотрудника, у которого не указано is_active, база не отмечает работающим — а по требованиям новый сотрудник работает",
		by: map[string]string{
			"23502": "сотрудника без is_active база не принимает — а по требованиям база сама отмечает нового сотрудника работающим",
			"22012": "у сотрудника, добавленного без is_active, отметка остаётся пустой — а по требованиям база сама отмечает нового сотрудника работающим",
			"2201E": "сотрудник, добавленный без is_active, получается неработающим — а по требованиям новый сотрудник работает",
			"42804": "в is_active хранится не «да» или «нет» — подумай, какой тип для этого нужен",
		}},
	{sql: pgsLab8Agent("is_active", "NULL"), want: pgsLab8NotNull,
		say: "сотрудника с пустой отметкой is_active база принимает — а работает сотрудник или нет, по требованиям известно всегда"},
	{sql: pgsLab8Agent("full_name", "NULL"), want: pgsLab8NotNull,
		say: "сотрудника без имени (пустой full_name) база принимает — а имя сотрудника по требованиям нужно знать обязательно"},
	{sql: pgsLab8Agent("email", "NULL"), want: pgsLab8NotNull,
		say: "сотрудника без email база принимает — а рабочий email по требованиям нужно знать обязательно"},
	{sql: pgsLab8Agent() + "; " + pgsLab8Agent(pgsLab8With(pgsLab8Agent2, "email", "'check-1@helpdesk.test'")...), want: pgsLab8Unique,
		say: "двух сотрудников с одинаковым email база принимает — а по требованиям email у каждого свой"},
	{sql: pgsLab8Agent() + "; " + pgsLab8Agent(pgsLab8With(pgsLab8Agent2, "full_name", "'Проверка Один'")...), want: pgsLab8OK,
		say: "двух сотрудников с одинаковым именем и разными email база не принимает — а тёзки бывают: свой у каждого email, а не имя"},
	{sql: pgsLab8Agent() + "; " + pgsLab8Agent(pgsLab8With(pgsLab8Agent2, "id", "30001")...), want: pgsLab8Unique,
		say: "двух сотрудников с одним номером id база принимает — id должен быть первичным ключом"},
	{sql: pgsLab8Agent("id", "NULL"), want: pgsLab8NotNull,
		say: "сотрудника с пустым id база принимает — id должен быть первичным ключом"},
}

// pgsLab8Prep is a row a table check adds before each of its rules, and what
// the student is told when it does not go in.
type pgsLab8Prep struct{ sql, say string }

// pgsLab8People are the trial client 30001 and agent 30002 every trial
// ticket refers to: the smallest rows the requirements allow, a client with
// no phone and an agent at work. When one does not go in, its table is not
// done yet, and the tickets check says which instead of trying its rules.
var pgsLab8People = []pgsLab8Prep{
	{pgsLab8Client("full_name", "'Проверка Клиент'", "email", "'check-client@helpdesk.test'", "phone", "-"),
		"пробного клиента — с именем и email, без телефона — в clients добавить не получается: сначала сдай задание «Таблица клиентов»"},
	{pgsLab8Agent("id", "30002", "full_name", "'Проверка Сотрудник'", "email", "'check-agent@helpdesk.test'"),
		"пробного сотрудника — с именем, email и отметкой, что он работает, — в agents добавить не получается: сначала сдай задание «Таблица сотрудников»"},
}

// pgsLab8TicketBase is the trial ticket: new, from the trial client, assigned
// to the trial agent, not closed yet. Columns in the order of the INSERT.
var pgsLab8TicketBase = [][2]string{
	{"id", "30001"},
	{"client_id", "30001"},
	{"agent_id", "30002"},
	{"subject", "'Проверка'"},
	{"priority", "'normal'"},
	{"status", "'open'"},
	{"created_at", "'2026-10-01'"},
	{"closed_at", "NULL"},
}

// pgsLab8Ticket is the trial ticket with columns changed (see pgsLab8Row).
func pgsLab8Ticket(set ...string) string { return pgsLab8Row("tickets", pgsLab8TicketBase, set...) }

// pgsLab8Default is the rule "a ticket that leaves col out gets want": it is
// not refused (23502), col is not empty (22012) and holds want (2201E).
func pgsLab8Default(col, value, read string, by map[string]string) pgsLab8Rule {
	return pgsLab8Rule{
		sql: pgsLab8Ticket(col, "-") +
			"; SELECT 1/(" + col + " IS NOT NULL)::int FROM tickets WHERE id = 30001" +
			"; SELECT ln((" + read + " = " + value + ")::int) FROM tickets WHERE id = 30001",
		want: pgsLab8OK, say: by["2201E"], by: by,
	}
}

var pgsLab8TicketRules = []pgsLab8Rule{
	// What has to go in. The trial ticket's only empty column is closed_at.
	{sql: pgsLab8Ticket(), want: pgsLab8OK, gate: true,
		say: "правильное новое обращение — open, normal, ещё не закрытое — в tickets добавить не получается: посмотри на таблицу, \\d tickets",
		by: map[string]string{
			"23502": "обращение без даты закрытия (пустой closed_at) база не принимает — а пока обращение не закрыто, этой даты по требованиям нет",
			"23503": "обращение от клиента из clients сотруднику из agents база не принимает — проверь, куда ведут ссылки: client_id — на clients (id), agent_id — на agents (id)",
			"23514": "правильное новое обращение — open, normal, ещё не закрытое — не проходит твои правила CHECK: проверь списки значений priority и status",
			"22P02": "правильные значения не записываются в столбцы tickets — проверь их типы",
			"22007": "правильные значения не записываются в столбцы tickets — проверь их типы",
			"42804": "правильные значения не записываются в столбцы tickets — проверь их типы",
		}},
	{sql: pgsLab8Ticket() + "; " + pgsLab8Ticket("id", "30002", "subject", "'Проверка 2'"), want: pgsLab8OK, gate: true,
		say: "второе обращение от того же клиента тому же сотруднику база не принимает — а по требованиям клиент пишет сколько угодно обращений, и сотрудник разбирает не одно"},
	{sql: pgsLab8Ticket() + "; " + pgsLab8Ticket("id", "30002"), want: pgsLab8OK,
		say: "два обращения с одинаковой темой база не принимает — а темы повторяются: «Не доставили заказ» пишут многие"},
	{sql: pgsLab8KeepSeq("tickets", pgsLab8Ticket("id", "-")+"; "+pgsLab8Ticket("id", "-", "subject", "'Проверка 2'")), want: pgsLab8OK,
		say: "два обращения без id в tickets добавить не получается: посмотри на таблицу, \\d tickets",
		by: map[string]string{
			"23502": "обращение без номера id база не принимает — номер должна ставить сама база, как во всех таблицах курса",
			"23505": "двух обращений, добавленных без id, база не принимает — номер id должна ставить сама база, каждому свой",
		}},
	{sql: pgsLab8Ticket("agent_id", "NULL"), want: pgsLab8OK,
		say: "новое обращение без сотрудника (пустой agent_id) база не принимает — а по требованиям обращение приходит ни на кого не назначенным"},
	// Who it is from and who handles it.
	{sql: pgsLab8Ticket("client_id", "-"), want: pgsLab8NotNull,
		say: "обращение без клиента (client_id не указан) база принимает — а по требованиям обращение всегда от клиента",
		by: map[string]string{
			"23503": "если client_id не указать, база подставляет номер клиента сама — а клиента выбирает человек: ссылка — обычное целое число, без serial",
		}},
	{sql: pgsLab8Ticket("client_id", "30002"), want: pgsLab8FK,
		say: "обращение от клиента, которого нет в clients, база принимает — client_id должен ссылаться на clients (id)"},
	{sql: pgsLab8Ticket("agent_id", "30001"), want: pgsLab8FK,
		say: "обращение сотруднику, которого нет в agents, база принимает — agent_id должен ссылаться на agents (id)"},
	{sql: pgsLab8Ticket("subject", "NULL"), want: pgsLab8NotNull,
		say: "обращение без темы (пустой subject) база принимает — а тема у обращения по требованиям есть всегда"},
	// Priority.
	{sql: pgsLab8Ticket("priority", "'low'") + "; " + pgsLab8Ticket("id", "30002", "subject", "'Проверка 2'", "priority", "'normal'") + "; " +
		pgsLab8Ticket("id", "30003", "subject", "'Проверка 3'", "priority", "'high'"), want: pgsLab8OK,
		say: "не все важности low, normal и high база принимает — проверь список допустимых значений priority"},
	{sql: pgsLab8Ticket("priority", "'urgent'"), want: pgsLab8BadValue,
		say: "важность 'urgent' база принимает — а по требованиям бывает только low, normal или high"},
	{sql: pgsLab8Ticket("priority", "'High'"), want: pgsLab8BadValue,
		say: "важность 'High' с большой буквы база принимает — а допустимы только low, normal и high, именно так, маленькими буквами"},
	{sql: pgsLab8Ticket("priority", "NULL"), want: pgsLab8NotNull,
		say: "обращение с пустой важностью (priority) база принимает — а важность у обращения по требованиям есть всегда"},
	pgsLab8Default("priority", "'normal'", "priority::text", map[string]string{
		"23502": "обращение, у которого не указана важность, база не принимает — а по требованиям такое обращение получает важность normal",
		"22012": "у обращения, добавленного без важности, она остаётся пустой — а по требованиям она становится normal",
		"2201E": "у обращения, добавленного без важности, она становится не normal — а по требованиям должна быть normal",
	}),
	// Status.
	{sql: pgsLab8Ticket("status", "'open'") + "; " + pgsLab8Ticket("id", "30002", "subject", "'Проверка 2'", "status", "'in_progress'"), want: pgsLab8OK,
		say: "не все состояния open, in_progress и closed база принимает — проверь список допустимых значений status"},
	// A closed ticket is the lead of the rules about closing: when it does not
	// go in, the status list or the rule on the two dates is wrong, and the
	// other rules of the group would only repeat that in other words.
	{sql: pgsLab8Ticket("status", "'closed'", "closed_at", "'2026-10-03'"), want: pgsLab8OK, group: "closed", lead: true,
		say: "закрытое обращение — состояние closed, пришло 2026-10-01, закрыто 2026-10-03 — база не принимает: проверь список значений status и правило про две даты"},
	{sql: pgsLab8Ticket("status", "'done'"), want: pgsLab8BadValue,
		say: "состояние 'done' база принимает — а по требованиям бывает только open, in_progress или closed"},
	{sql: pgsLab8Ticket("status", "'Open'"), want: pgsLab8BadValue,
		say: "состояние 'Open' с большой буквы база принимает — а допустимы только open, in_progress и closed, именно так, маленькими буквами"},
	{sql: pgsLab8Ticket("status", "NULL"), want: pgsLab8NotNull,
		say: "обращение с пустым состоянием (status) база принимает — а состояние у обращения по требованиям есть всегда"},
	pgsLab8Default("status", "'open'", "status::text", map[string]string{
		"23502": "обращение, у которого не указано состояние, база не принимает — а по требованиям новое обращение получает состояние open",
		"22012": "у обращения, добавленного без состояния, оно остаётся пустым — а по требованиям новое обращение — open",
		"2201E": "у обращения, добавленного без состояния, оно становится не open — а по требованиям новое обращение — open",
	}),
	// The date it came.
	{sql: pgsLab8Ticket("created_at", "'17 марта'"), want: pgsLab8BadDate,
		say: "в created_at можно записать текст «17 марта» — а там должна храниться дата"},
	{sql: pgsLab8Ticket("created_at", "'2026-10-01 15:30'") +
		"; SELECT ln((created_at::text = '2026-10-01')::int) FROM tickets WHERE id = 30001", want: pgsLab8OK,
		say: "в created_at сохраняется не только дата, но и время — а по требованиям это дата"},
	{sql: pgsLab8Ticket("created_at", "NULL"), want: pgsLab8NotNull,
		say: "обращение с пустой датой created_at база принимает — а дата, когда обращение пришло, по требованиям есть всегда"},
	pgsLab8Default("created_at", "current_date", "created_at::date", map[string]string{
		"23502": "обращение, у которого не указана дата created_at, база не принимает — а по требованиям её ставит сама база: сегодняшнюю",
		"22012": "у обращения, добавленного без created_at, дата остаётся пустой — а по требованиям база ставит сегодняшнюю",
		"2201E": "у обращения, добавленного без created_at, дата получается не сегодняшняя — а по требованиям база ставит сегодняшнюю",
	}),
	// The date it was closed.
	{sql: pgsLab8Ticket("status", "'closed'", "closed_at", "'17 марта'"), want: pgsLab8BadDate, group: "closed",
		say: "в closed_at можно записать текст «17 марта» — а там должна храниться дата"},
	{sql: pgsLab8Ticket("status", "'closed'", "closed_at", "'2026-10-02 15:30'") +
		"; SELECT ln((closed_at::text = '2026-10-02')::int) FROM tickets WHERE id = 30001", want: pgsLab8OK, group: "closed",
		say: "в closed_at сохраняется не только дата, но и время — а по требованиям это дата"},
	{sql: pgsLab8Ticket("status", "'closed'", "created_at", "'2026-10-05'", "closed_at", "'2026-10-01'"), want: pgsLab8Check,
		say: "обращение, закрытое раньше, чем оно пришло, база принимает — а по требованиям так нельзя. Это правило про два столбца сразу"},
	{sql: pgsLab8Ticket("status", "'closed'", "created_at", "'2026-10-05'", "closed_at", "'2026-10-05'"), want: pgsLab8OK, group: "closed",
		say: "обращение, закрытое в тот же день, когда оно пришло, база не принимает — а по требованиям так можно"},
	// The number.
	{sql: pgsLab8Ticket() + "; " + pgsLab8Ticket("subject", "'Проверка 2'"), want: pgsLab8Unique,
		say: "два обращения с одним номером id база принимает — id должен быть первичным ключом"},
	{sql: pgsLab8Ticket("id", "NULL"), want: pgsLab8NotNull,
		say: "обращение с пустым id база принимает — id должен быть первичным ключом"},
}

// pgsLab8MaxSaid is how many broken rules one diagnostic names; the rest are
// counted. A table with no rules at all breaks a dozen, and a list that long
// is not read.
const pgsLab8MaxSaid = 5

// pgsLab8Run is a command running a generated bash script in the sandbox. The
// script travels gzipped and base64-encoded, so it may quote as it likes and
// still sit in a one-line check. Inside it SQL is plain text in single quotes,
// handed to glpg by two functions that encode it there. The tickets check tries
// some thirty rules, each an INSERT of the whole trial row: as plain base64 it
// came to 40 KB, and the runner encodes a check twice more on its way into a
// VM, which brings one command-line argument close to the kernel's 128 KB
// limit. gzip, an essential package of Ubuntu, packs the repeated rows into a
// fifth of that; Go writes no time into the header, so the seed is stable.
func pgsLab8Run(script string) string {
	var z bytes.Buffer
	w, _ := gzip.NewWriterLevel(&z, gzip.BestCompression)
	w.Write([]byte(pgsLab8Funcs + script))
	w.Close()
	return "echo " + base64.StdEncoding.EncodeToString(z.Bytes()) + " | base64 -d | gzip -dc | bash 2>/dev/null"
}

// pgsLab8Funcs: q DB SQL prints a query's value (glpg q), try SQL prints the
// SQLSTATE that SQL ends with in helpdesk, rolled back (glpg state).
const pgsLab8Funcs = `q() { glpg q "$1" "$(printf '%s' "$2" | base64 -w0)"; }
try() { glpg state helpdesk "$(printf '%s' "$1" | base64 -w0)"; }
`

// pgsLab8Sh is s quoted for bash.
func pgsLab8Sh(s string) string { return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'" }

// pgsLab8Q is a bash substitution printing a query's value in db.
func pgsLab8Q(db, sql string) string { return "$(q " + db + " " + pgsLab8Sh(sql) + ")" }

// pgsLab8NoDB is the bash that stops a check when there is no helpdesk.
var pgsLab8NoDB = "if [ \"" + pgsLab8Q("postgres", "SELECT count(*) FROM pg_database WHERE datname = 'helpdesk'") + "\" != 1 ]; then\n" +
	"\tprintf '%s' " + pgsLab8Sh("Сейчас: базы helpdesk нет — сначала создай её (первое задание)") + "\n"

// pgsLab8Tables is SQL for the tables in helpdesk, as \dt lists them.
const pgsLab8Tables = "coalesce((SELECT string_agg(tablename, ', ' ORDER BY tablename) FROM pg_tables " +
	"WHERE schemaname = 'public'), 'ни одной')"

// pgsLab8Shape is SQL true when table's columns are exactly cols, any types.
func pgsLab8Shape(table string, cols []string) string {
	want := append([]string(nil), cols...)
	sort.Strings(want) // byte order, as attname sorts
	return "SELECT coalesce((SELECT string_agg(attname, ',' ORDER BY attname) FROM pg_attribute " +
		"WHERE attrelid = to_regclass('" + table + "') AND attnum > 0 AND NOT attisdropped), '') = '" +
		strings.Join(want, ",") + "'"
}

// pgsLab8Columns is SQL describing table for a "Сейчас:" line: that it is
// missing, or its columns as \d lists them and which are missing or extra.
func pgsLab8Columns(table string, cols []string) string {
	vals := make([]string, len(cols))
	for i, c := range cols {
		vals[i] = "('" + c + "', " + strconv.Itoa(i+1) + ")"
	}
	w := "(VALUES " + strings.Join(vals, ", ") + ") w(name, n)"
	h := "(SELECT a.attname::text AS name, format_type(a.atttypid, a.atttypmod) AS typ, a.attnum AS n FROM pg_attribute a " +
		"WHERE a.attrelid = to_regclass('" + table + "') AND a.attnum > 0 AND NOT a.attisdropped) h"
	return "SELECT CASE WHEN to_regclass('" + table + "') IS NULL " +
		"THEN 'в базе helpdesk нет таблицы " + table + ", таблицы в базе: ' || " + pgsLab8Tables + " " +
		"ELSE concat_ws('; ', 'столбцы " + table + ": ' || " +
		"coalesce((SELECT string_agg(h.name || ' ' || h.typ, ', ' ORDER BY h.n) FROM " + h + "), 'ни одного'), " +
		"(SELECT 'не хватает столбцов: ' || string_agg(w.name, ', ' ORDER BY w.n) FROM " + w + " WHERE w.name NOT IN (SELECT h.name FROM " + h + ")), " +
		"(SELECT 'лишние столбцы: ' || string_agg(h.name, ', ' ORDER BY h.n) FROM " + h + " WHERE h.name NOT IN (SELECT w.name FROM " + w + "))" +
		") END"
}

// pgsLab8Busy is what a check says when it waited for a table in vain (55P03,
// glpg's lock_timeout): the student's psql holds it in an open transaction —
// say, after BEGIN and an ALTER TABLE — and pgsLab8Tx tells how to end it.
const pgsLab8Busy = "проверка не дождалась таблиц: их держит незакрытая транзакция"

// pgsLab8Tx is SQL text: a note when one of the student's psql sessions sits
// in an open transaction — a table or rows made inside one are invisible to
// the check until COMMIT — or an empty string.
const pgsLab8Tx = "SELECT CASE " +
	"WHEN bool_or(state = 'idle in transaction (aborted)') " +
	"THEN '; в твоей транзакции была ошибка (приглашение вида helpdesk=!#): набери ROLLBACK; и повтори команды' " +
	"WHEN bool_or(state = 'idle in transaction') " +
	"THEN '; у тебя открыта транзакция (приглашение вида helpdesk=*#): проверка не видит её изменений — сохрани их командой COMMIT;' " +
	"ELSE '' END FROM pg_stat_activity WHERE backend_type = 'client backend' AND pid <> pg_backend_pid()"

// pgsLab8Elsewhere is the bash that notes table made in the database postgres
// while helpdesk has none: the file was run before \c helpdesk.
func pgsLab8Elsewhere(table, helpdeskUp string) string {
	note := "; таблица " + table + " нашлась в базе postgres — ты выполнил файл не в той базе: переключись в helpdesk (\\c helpdesk) и выполни его там"
	return "if " + helpdeskUp + "[ \"" + pgsLab8Q("postgres", "SELECT to_regclass('"+table+"') IS NOT NULL") + "\" = t ]; then\n" +
		"\tprintf '%s' " + pgsLab8Sh(note) + "\n" +
		"fi\n"
}

// pgsLab8TableCheck is the check of a table the student designs: its columns
// are exactly cols and it does what rules say, the rows of prep (the trial
// client and agent of the tickets) added first in each rule's transaction. The
// script it runs is silent on success; otherwise it prints "Сейчас: …" with
// the columns and, when they are right, up to pgsLab8MaxSaid broken rules.
func pgsLab8TableCheck(table string, cols []string, prep []pgsLab8Prep, rules []pgsLab8Rule, good string) string {
	return pgCheck(pgsLab8Run(pgsLab8TableScript(table, cols, prep, rules)), good)
}

func pgsLab8TableScript(table string, cols []string, prep []pgsLab8Prep, rules []pgsLab8Rule) string {
	busy := "55P03) stop=1; add " + pgsLab8Sh(pgsLab8Busy) + " ;;\n"
	pre := ""
	for _, p := range prep {
		pre += p.sql + "; "
	}
	var b strings.Builder
	b.WriteString("n=0 msg= stop= P=" + pgsLab8Sh(pre) + "\n")
	b.WriteString("add() { [ \"$n\" -lt " + strconv.Itoa(pgsLab8MaxSaid) + " ] && msg=\"$msg; $1\"; n=$((n + 1)); }\n")
	b.WriteString(pgsLab8NoDB)
	b.WriteString(pgsLab8Elsewhere(table, ""))
	b.WriteString("\texit 1\nfi\n")
	b.WriteString("s=" + pgsLab8Q("helpdesk", pgsLab8Columns(table, cols)) + " || " +
		"{ printf '%s' " + pgsLab8Sh("Сейчас: не получается подключиться к базе helpdesk") + "; exit 1; }\n")
	b.WriteString("if [ \"" + pgsLab8Q("helpdesk", pgsLab8Shape(table, cols)) + "\" != t ]; then\n\tn=1\n")
	b.WriteString("else\n")
	// Each trial row of prep on top of the ones before it: the first that
	// does not go in names the table to finish first, and no rule is tried.
	pre = ""
	for _, p := range prep {
		pre += p.sql + "; "
		b.WriteString("\t[ -z \"$stop\" ] && case $(try " + pgsLab8Sh(pre) + ") in\n")
		b.WriteString("\t\t00000) ;;\n\t\t" + busy)
		b.WriteString("\t\t*) stop=1; add " + pgsLab8Sh(p.say) + " ;;\n\t\tesac\n")
	}
	for _, r := range rules {
		failed, guard := "add ", ""
		if r.gate {
			failed = "stop=1; add "
		}
		if r.group != "" {
			guard = "[ -z \"$no_" + r.group + "\" ] && "
			if r.lead {
				failed = "no_" + r.group + "=1; add "
			}
		}
		b.WriteString("\t[ -z \"$stop\" ] && " + guard + "st=$(try \"$P\"" + pgsLab8Sh(r.sql) + ") && case $st in\n")
		b.WriteString("\t\t" + strings.Join(r.want, "|") + ") ;;\n")
		b.WriteString("\t\t" + busy)
		states := make([]string, 0, len(r.by))
		for st := range r.by {
			states = append(states, st)
		}
		sort.Strings(states)
		for _, st := range states {
			b.WriteString("\t\t" + st + ") " + failed + pgsLab8Sh(r.by[st]) + " ;;\n")
		}
		b.WriteString("\t\t*) " + failed + pgsLab8Sh(r.say) + " ;;\n\t\tesac\n")
	}
	b.WriteString("fi\n")
	b.WriteString("[ \"$n\" = 0 ] && exit 0\n")
	b.WriteString("[ \"$n\" -gt " + strconv.Itoa(pgsLab8MaxSaid) + " ] && msg=\"$msg; и ещё нарушений: $((n - " + strconv.Itoa(pgsLab8MaxSaid) + "))\"\n")
	b.WriteString("printf 'Сейчас: %s%s' \"$s\" \"$msg\"\n")
	b.WriteString(pgsLab8Elsewhere(table, "[ \""+pgsLab8Q("helpdesk", "SELECT to_regclass('"+table+"') IS NULL")+"\" = t ] && "))
	b.WriteString("q helpdesk " + pgsLab8Sh(pgsLab8Tx) + "\n")
	b.WriteString("exit 1\n")
	return b.String()
}

// The rows of the data task. phone, agent and closed "" are empty (NULL);
// created "" is the day the row goes in, which the default gives.
type pgsLab8Person struct{ name, email, phone string }

type pgsLab8TicketRow struct {
	client, agent, subject, priority, status, created, closed string
}

var (
	pgsLab8Clients = []pgsLab8Person{
		{"Мария Волкова", "maria.volkova@example.com", "+7 912 345-67-89"},
		{"Денис Орлов", "denis.orlov@example.com", ""},
	}
	pgsLab8Agents = []pgsLab8Person{
		{"Елена Морозова", "morozova@support.example.com", ""},
	}
	pgsLab8Tickets = []pgsLab8TicketRow{
		{"maria.volkova@example.com", "morozova@support.example.com", "Не доставили заказ", "high", "closed", "2026-10-01", "2026-10-03"},
		{"denis.orlov@example.com", "morozova@support.example.com", "В коробке не хватает фишек", "normal", "in_progress", "2026-10-05", ""},
		{"maria.volkova@example.com", "", "Как вернуть игру", "low", "open", "", ""},
	}
)

// pgsLab8Lit is s as an SQL literal, "" as NULL.
func pgsLab8Lit(s string) string {
	if s == "" {
		return "NULL"
	}
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// pgsLab8Emails is the SQL list of the emails of people.
func pgsLab8Emails(people []pgsLab8Person) string {
	l := make([]string, len(people))
	for i, p := range people {
		l[i] = pgsLab8Lit(p.email)
	}
	return "(" + strings.Join(l, ", ") + ")"
}

// pgsLab8PersonConds are the SQL conditions on a row of clients (phone) or
// agents (is_active true) that make it person p, by column.
func pgsLab8PersonConds(table string, p pgsLab8Person) [][2]string {
	c := [][2]string{{"full_name", "full_name = " + pgsLab8Lit(p.name)}}
	if table == "clients" {
		c = append(c, [2]string{"phone", "phone::text IS NOT DISTINCT FROM " + pgsLab8Lit(p.phone)})
	} else {
		c = append(c, [2]string{"is_active", "is_active::text = 'true'"})
	}
	return c
}

// pgsLab8TicketConds are the SQL conditions on t (a ticket), c (its client) and
// a (its agent) that make it ticket r, by what the student is told differs.
func pgsLab8TicketConds(r pgsLab8TicketRow) [][2]string {
	agent := "a.email = " + pgsLab8Lit(r.agent)
	if r.agent == "" {
		agent = "t.agent_id IS NULL"
	}
	created := "t.created_at::date = " + pgsLab8Lit(r.created)
	if r.created == "" {
		created = "t.created_at::date BETWEEN current_date - 1 AND current_date + 1"
	}
	closed := "t.closed_at::date = " + pgsLab8Lit(r.closed)
	if r.closed == "" {
		closed = "t.closed_at IS NULL"
	}
	return [][2]string{
		{"клиент (client_id)", "c.email = " + pgsLab8Lit(r.client)},
		{"сотрудник (agent_id)", agent},
		{"priority", "t.priority::text = " + pgsLab8Lit(r.priority)},
		{"status", "t.status::text = " + pgsLab8Lit(r.status)},
		{"created_at", created},
		{"closed_at", closed},
	}
}

// pgsLab8TicketFrom is the FROM of a ticket with its client and agent.
const pgsLab8TicketFrom = "tickets t LEFT JOIN clients c ON c.id = t.client_id LEFT JOIN agents a ON a.id = t.agent_id"

func pgsLab8And(conds [][2]string) string {
	l := make([]string, len(conds))
	for i, c := range conds {
		l[i] = c[1]
	}
	return strings.Join(l, " AND ")
}

// pgsLab8DataOK is SQL true when the three tables hold exactly the rows of the
// data task.
var pgsLab8DataOK = func() string {
	conds := []string{
		"(SELECT count(*) FROM clients) = " + strconv.Itoa(len(pgsLab8Clients)),
		"(SELECT count(*) FROM agents) = " + strconv.Itoa(len(pgsLab8Agents)),
		"(SELECT count(*) FROM tickets) = " + strconv.Itoa(len(pgsLab8Tickets)),
	}
	for _, p := range pgsLab8Clients {
		conds = append(conds, "EXISTS (SELECT 1 FROM clients WHERE email = "+pgsLab8Lit(p.email)+" AND "+pgsLab8And(pgsLab8PersonConds("clients", p))+")")
	}
	for _, p := range pgsLab8Agents {
		conds = append(conds, "EXISTS (SELECT 1 FROM agents WHERE email = "+pgsLab8Lit(p.email)+" AND "+pgsLab8And(pgsLab8PersonConds("agents", p))+")")
	}
	for _, r := range pgsLab8Tickets {
		conds = append(conds,
			"(SELECT count(*) FROM tickets WHERE subject = "+pgsLab8Lit(r.subject)+") = 1",
			"EXISTS (SELECT 1 FROM "+pgsLab8TicketFrom+" WHERE t.subject = "+pgsLab8Lit(r.subject)+" AND "+pgsLab8And(pgsLab8TicketConds(r))+")")
	}
	return "SELECT " + strings.Join(conds, " AND ")
}()

// pgsLab8Words names the rows of clients or agents in a diagnostic: how many
// the task has, one that is missing or differs, the extra ones.
type pgsLab8Words struct{ many, one, extra string }

// pgsLab8PeopleState says what table (clients or agents) holds compared with
// the task, or NULL when it matches: empty, how many rows, who is missing,
// whose values differ (by column) and who is extra.
func pgsLab8PeopleState(table string, w pgsLab8Words, people []pgsLab8Person) string {
	parts := []string{"CASE WHEN (SELECT count(*) FROM " + table + ") <> " + strconv.Itoa(len(people)) +
		" THEN 'в " + table + " строк: ' || (SELECT count(*) FROM " + table + ") || ', а " + w.many + " в задании " + strconv.Itoa(len(people)) + "' END"}
	for _, p := range people {
		diff := []string{}
		for _, c := range pgsLab8PersonConds(table, p) {
			diff = append(diff, "CASE WHEN NOT EXISTS (SELECT 1 FROM "+table+" WHERE email = "+pgsLab8Lit(p.email)+" AND "+c[1]+") THEN '"+c[0]+"' END")
		}
		parts = append(parts, "CASE WHEN NOT EXISTS (SELECT 1 FROM "+table+" WHERE email = "+pgsLab8Lit(p.email)+") "+
			"THEN 'нет "+w.one+" "+p.name+" ("+p.email+")' "+
			"ELSE 'у "+w.one+" "+p.email+" не совпадает: ' || nullif(concat_ws(', ', "+strings.Join(diff, ", ")+"), '') END")
	}
	parts = append(parts, "(SELECT 'лишние "+w.extra+": ' || string_agg(coalesce(email, 'без email'), ', ' ORDER BY email) FROM "+table+
		" WHERE email IS NULL OR email NOT IN "+pgsLab8Emails(people)+")")
	return "CASE WHEN NOT EXISTS (SELECT 1 FROM " + table + ") THEN 'таблица " + table + " пустая' " +
		"ELSE nullif(concat_ws('; ', " + strings.Join(parts, ", ") + "), '') END"
}

// pgsLab8TicketsState is pgsLab8PeopleState for tickets, which are told apart
// by subject.
var pgsLab8TicketsState = func() string {
	subjects := make([]string, len(pgsLab8Tickets))
	parts := []string{"CASE WHEN (SELECT count(*) FROM tickets) <> " + strconv.Itoa(len(pgsLab8Tickets)) +
		" THEN 'в tickets строк: ' || (SELECT count(*) FROM tickets) || ', а обращений в задании " + strconv.Itoa(len(pgsLab8Tickets)) + "' END"}
	for i, r := range pgsLab8Tickets {
		s := pgsLab8Lit(r.subject)
		subjects[i] = s
		diff := []string{}
		for _, c := range pgsLab8TicketConds(r) {
			diff = append(diff, "CASE WHEN NOT EXISTS (SELECT 1 FROM "+pgsLab8TicketFrom+" WHERE t.subject = "+s+" AND "+c[1]+") THEN '"+c[0]+"' END")
		}
		parts = append(parts, "CASE WHEN NOT EXISTS (SELECT 1 FROM tickets WHERE subject = "+s+") "+
			"THEN 'нет обращения «"+r.subject+"»' "+
			"WHEN (SELECT count(*) FROM tickets WHERE subject = "+s+") > 1 "+
			"THEN 'обращение «"+r.subject+"» записано несколько раз: ' || (SELECT count(*) FROM tickets WHERE subject = "+s+") "+
			"ELSE 'у обращения «"+r.subject+"» не совпадает: ' || nullif(concat_ws(', ', "+strings.Join(diff, ", ")+"), '') END")
	}
	parts = append(parts, "(SELECT 'лишние обращения: ' || string_agg('«' || coalesce(subject, 'без темы') || '»', ', ' ORDER BY id) FROM tickets"+
		" WHERE subject IS NULL OR subject NOT IN ("+strings.Join(subjects, ", ")+"))")
	return "CASE WHEN NOT EXISTS (SELECT 1 FROM tickets) THEN 'таблица tickets пустая' " +
		"ELSE nullif(concat_ws('; ', " + strings.Join(parts, ", ") + "), '') END"
}()

// pgsLab8DataState says what differs from the data task, table by table.
var pgsLab8DataState = "SELECT coalesce(nullif(concat_ws('; ', " +
	pgsLab8PeopleState("clients", pgsLab8Words{"клиентов", "клиента", "клиенты"}, pgsLab8Clients) + ", " +
	pgsLab8PeopleState("agents", pgsLab8Words{"сотрудников", "сотрудника", "сотрудники"}, pgsLab8Agents) + ", " +
	pgsLab8TicketsState + "), ''), 'строки не совпадают с заданием')"

// pgsLab8DataScript is the bash of the data check: silent with exit 0 when
// the rows are those of the task, otherwise "Сейчас: …" and exit 1.
var pgsLab8DataScript = pgsLab8NoDB + "\texit 1\nfi\n" +
	"m=" + pgsLab8Q("helpdesk", "SELECT concat_ws(', ', "+
	"CASE WHEN to_regclass('clients') IS NULL THEN 'clients' END, "+
	"CASE WHEN to_regclass('agents') IS NULL THEN 'agents' END, "+
	"CASE WHEN to_regclass('tickets') IS NULL THEN 'tickets' END)") + " || " +
	"{ printf '%s' " + pgsLab8Sh("Сейчас: не получается подключиться к базе helpdesk") + "; exit 1; }\n" +
	"if [ -n \"$m\" ]; then\n" +
	"\tprintf 'Сейчас: в базе helpdesk нет таблиц: %s — сначала сделай задания про таблицы' \"$m\"\n" +
	"\tq helpdesk " + pgsLab8Sh(pgsLab8Tx) + "\n" +
	"\texit 1\n" +
	"fi\n" +
	"if [ \"$(try 'SELECT FROM clients, agents, tickets LIMIT 1')\" = 55P03 ]; then\n" +
	"\tprintf '%s' " + pgsLab8Sh("Сейчас: "+pgsLab8Busy) + "\n" +
	"\tq helpdesk " + pgsLab8Sh(pgsLab8Tx) + "\n" +
	"\texit 1\n" +
	"fi\n" +
	"[ \"" + pgsLab8Q("helpdesk", pgsLab8DataOK) + "\" = t ] && exit 0\n" +
	"s=" + pgsLab8Q("helpdesk", pgsLab8DataState) + " || " +
	"s=" + pgsLab8Sh("не получилось сравнить строки с заданием — проверь, что таблицы сделаны по заданиям про таблицы") + "\n" +
	"printf 'Сейчас: %s' \"$s\"\n" +
	"q helpdesk " + pgsLab8Sh(pgsLab8Tx) + "\n" +
	"exit 1\n"
