package main

// Lab fixtures for "PostgreSQL: старт с нуля" (pg-start).
//
// The course runs on golearn/sandbox-pg. Every lab but the first starts with
// pgSetup: the server is up and root is a superuser, which the first lab
// teaches the student to arrange by hand. Each lab rebuilds the database the
// way the previous labs leave it, so a student who skipped one, or broke the
// database, still starts from the state the tasks describe.
//
// Reference solutions: scripts/labcheck/solutions-pg.sh.
var pgStartLabs = map[string]labSpec{}
