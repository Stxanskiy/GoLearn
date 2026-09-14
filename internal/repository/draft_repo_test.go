package repository

import (
	"slices"
	"testing"
)

func TestCompareItems(t *testing.T) {
	row := func(id, origin, parent int, title string) contentRow {
		return contentRow{id: id, origin: origin, parent: parent, fields: map[string]any{"title": title}}
	}
	live := []contentRow{row(1, 0, 10, "a"), row(2, 0, 10, "b"), row(3, 0, 10, "c"), row(4, 0, 11, "other lesson")}
	draft := []contentRow{row(101, 1, 20, "a"), row(102, 2, 20, "b2"), row(103, 0, 20, "new"), row(104, 4, 21, "other lesson")}
	if got := compareItems(live, draft, 10, 20); got != (ItemChanges{Added: 1, Removed: 1, Modified: 1}) {
		t.Errorf("changes = %+v", got)
	}
	if got := compareItems(live, nil, 10, 0); got != (ItemChanges{Removed: 3}) {
		t.Errorf("removed lesson = %+v", got)
	}
	if got := compareItems(live, draft, 0, 20); got != (ItemChanges{Added: 3}) {
		t.Errorf("added lesson = %+v", got)
	}
	if got := changedFields(map[string]any{"a": 1.0, "b": []any{"x"}}, map[string]any{"a": 1.0, "b": []any{"y"}, "c": "z"}); !slices.Equal(got, []string{"b", "c"}) {
		t.Errorf("fields = %v", got)
	}
}
