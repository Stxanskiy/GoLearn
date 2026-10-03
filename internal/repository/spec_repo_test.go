package repository

import (
	"context"
	"testing"
	"time"

	"github.com/backendraz/golearn/internal/model"
)

// Sections had no test at all, and the INSERT in Upsert listed seven columns
// against nine placeholders — so creating or editing one in the studio failed
// every time. The fakes in the API tests answered nil and hid it, which is the
// argument for running this against a real database.
func TestSpecUpsertCreatesAndUpdates(t *testing.T) {
	pool := billingPool(t)
	ctx := context.Background()
	repo := &SpecRepo{pool: pool}

	slug := "spec-test-" + time.Now().Format("150405.000000000")
	t.Cleanup(func() {
		_, _ = pool.Exec(ctx, `DELETE FROM specializations WHERE slug=$1`, slug)
	})

	first := model.Specialization{
		Slug: slug, Name: "Раздел", Description: "первый вариант",
		OrderNum: 9000, Published: false,
	}
	if err := repo.Upsert(ctx, first); err != nil {
		t.Fatalf("создание раздела: %v", err)
	}

	got, err := repo.Get(ctx, slug)
	if err != nil {
		t.Fatalf("чтение раздела: %v", err)
	}
	if got.Name != "Раздел" || got.Description != "первый вариант" || got.Published {
		t.Fatalf("создан не тот раздел: %+v", got)
	}

	second := first
	second.Name = "Переименованный"
	second.Description = "второй вариант"
	second.Published = true
	if err := repo.Upsert(ctx, second); err != nil {
		t.Fatalf("правка раздела: %v", err)
	}

	got, err = repo.Get(ctx, slug)
	if err != nil {
		t.Fatalf("чтение после правки: %v", err)
	}
	if got.Name != "Переименованный" || got.Description != "второй вариант" || !got.Published {
		t.Fatalf("правка не сохранилась: %+v", got)
	}
}
