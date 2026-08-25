package main

import "fmt"

// The curriculum is the single source of truth for the order courses are meant
// to be taken in. Each specialization gets its own 100-wide band, so adding a
// course to one path never renumbers another, and the catalog, the roadmap and
// the "next course" link all read the same sequence.
//
// Within a path every course builds on the one before it: terminal skills
// before Linux administration, Linux before containers, containers before
// orchestration.
var curriculum = map[string][]string{
	"devops": {
		"express-devops",       // что такое DevOps и зачем всё дальше
		"linux-terminal-start", // первые команды в реальном терминале
		"linux-start",          // файлы, права, процессы
		"linux-core",           // сеть, systemd, диски, логи
		"linux-advanced",       // диагностика и продвинутая эксплуатация
		"git-basics",           // версионирование — нужно всему, что дальше
		"devops-foundations",   // культура, CI/CD, метрики, надёжность
		"docker-basics",        // контейнеры и образы
		"docker-compose",       // многоконтейнерные окружения
		"k8s-intro",            // оркестрация: основы
		"k8s-ckad",             // оркестрация: продвинутая практика
		"helm",                 // пакетирование и деплой в кластер
	},
	"golang": {
		"basics", "collections", "functions", "pointers", "structs",
		"interfaces", "errors", "project-cli", "generics", "packages",
		"context", "git", "files-json", "http", "database", "architecture",
		"testing", "project-api", "auth", "concurrency", "advanced",
		"go-internals",
	},
	"database": {
		"sql-express", // быстрый старт по SQL
		"sql-easy",    // затем практика по возрастанию сложности
		"sql-medium",
		"sql-hard",
	},
	"security": {
		"security-offense",
		"security-defense",
	},
	"gym": {
		"gym-linux-start",
		"gym-linux-troubleshoot",
		"gym-git",
	},
}

// specBand is the order_num range reserved for each specialization; the bands
// also fix the order the specializations themselves appear in.
var specBand = map[string]int{
	"devops":   100,
	"golang":   200,
	"database": 300,
	"security": 400,
	"gym":      500,
}

// specForTrack maps a module track onto its catalog specialization.
func specForTrack(track string) string {
	switch track {
	case "devops":
		return "devops"
	case "database":
		return "database"
	case "gym":
		return "gym"
	case "security", "security-offense", "security-defense":
		return "security"
	default:
		return "golang"
	}
}

// assignOrder numbers every module by its position in the curriculum. A module
// that is not listed keeps a stable slot at the end of its band, so new content
// shows up last instead of silently jumping to the front.
func assignOrder(mods []M) error {
	pos := make(map[string]map[string]int, len(curriculum))
	for spec, slugs := range curriculum {
		m := make(map[string]int, len(slugs))
		for i, slug := range slugs {
			if _, dup := m[slug]; dup {
				return fmt.Errorf("curriculum: duplicate slug %q in %s", slug, spec)
			}
			m[slug] = i + 1
		}
		pos[spec] = m
	}

	unlisted := make(map[string]int)
	for i := range mods {
		spec := specForTrack(mods[i].Track)
		band, ok := specBand[spec]
		if !ok {
			band = 900
		}
		if p, ok := pos[spec][mods[i].Slug]; ok {
			mods[i].Order = band + p
			continue
		}
		unlisted[spec]++
		mods[i].Order = band + 50 + unlisted[spec]
	}
	return nil
}
