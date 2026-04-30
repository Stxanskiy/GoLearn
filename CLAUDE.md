# GoLearn — Learning Management System

## What This Is
A local LMS (Learning Management System) for learning Go backend development + DevOps.
This is the **platform** project. The **learning** project is WatchTogether (../WatchTogether/).

## Stack
- **Backend:** Go 1.22+, chi router, pgx (PostgreSQL)
- **Frontend:** Server-side HTML templates (Go html/template)
- **Database:** PostgreSQL 16 (via docker-compose, port 5433)
- **No external CSS frameworks** — custom CSS in templates

## How to Run
```bash
# Start database
docker compose up -d

# Seed course content
go run cmd/seed/main.go

# Start server
go run cmd/server/main.go
# Open http://localhost:8080
```

## Project Structure
```
cmd/server/     — HTTP server entry point
cmd/seed/       — Database seeder with all course content
internal/
  config/       — Environment config
  handler/      — HTTP handlers (dashboard, lesson, quiz, tasks, progress)
  model/        — Data models
  repository/   — PostgreSQL queries
  templates/    — HTML templates (layouts/, pages/, components/)
  static/       — CSS/JS/images
migrations/     — SQL migrations
```

## Learning Context
- Student builds WatchTogether through lesson tasks
- Each lesson has: theory (HTML content), quiz (multiple choice), practical tasks
- Progress is tracked in the database
- Brain/memory for this project: ~/.claude/projects/-Users-backendraz-GolandProjects-GoLearn/memory/

## Rules
- This is a LEARNING project — explain things thoroughly
- When student asks about a topic, check which lesson covers it
- Update learning_stage.md in brain after each completed lesson
- Language: Russian for explanations, English for code
