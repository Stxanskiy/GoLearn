package simulators

import "testing"

// The scenarios are data, and data rots quietly: a renamed metric key or an
// empty turn list breaks a simulator only when a student opens it. These assert
// the shape the player depends on.
func TestScenariosAreWellFormed(t *testing.T) {
	all := scenarios()
	if len(all) == 0 {
		t.Fatal("no scenarios")
	}
	seen := map[string]bool{}
	for _, s := range all {
		if s.Slug == "" || s.Title == "" {
			t.Errorf("scenario without slug or title: %+v", s)
		}
		if seen[s.Slug] {
			t.Errorf("duplicate slug %q — the seeder upserts by slug, so one would overwrite the other", s.Slug)
		}
		seen[s.Slug] = true

		if len(s.Metrics) == 0 {
			t.Errorf("%s: no metrics", s.Slug)
		}
		keys := map[string]bool{}
		for _, m := range s.Metrics {
			if m.Key == "" || m.Label == "" {
				t.Errorf("%s: metric without key or label", s.Slug)
			}
			keys[m.Key] = true
		}

		if len(s.Turns) == 0 {
			t.Errorf("%s: no turns", s.Slug)
		}
		for i, turn := range s.Turns {
			if len(turn.Choices) == 0 {
				t.Errorf("%s turn %d: no choices — the player would dead-end", s.Slug, i+1)
			}
			for _, c := range turn.Choices {
				if c.Text == "" {
					t.Errorf("%s turn %d: choice without text", s.Slug, i+1)
				}
				// An effect on an unknown metric is silently dropped by the
				// player, so the choice would look consequence-free.
				for key := range c.Effects {
					if !keys[key] {
						t.Errorf("%s turn %d: effect on unknown metric %q", s.Slug, i+1, key)
					}
				}
			}
		}
	}
}
