package schema

import "testing"

type fakeMigration struct{ name string }

func (fakeMigration) Up(*Schema) error   { return nil }
func (fakeMigration) Down(*Schema) error { return nil }

func TestRegistryOrdersByName(t *testing.T) {
	// Reset package state for a deterministic test.
	registeredMigrations = map[string]Migration{}
	registeredOrder = nil

	RegisterMigration("2026_05_18_120000_create_posts_table", fakeMigration{"posts"})
	RegisterMigration("2026_05_17_090000_create_users_table", fakeMigration{"users"})
	RegisterMigration("2026_05_18_010000_create_tags_table", fakeMigration{"tags"})

	names, all := RegisteredMigrations()
	want := []string{
		"2026_05_17_090000_create_users_table",
		"2026_05_18_010000_create_tags_table",
		"2026_05_18_120000_create_posts_table",
	}
	for i, n := range want {
		if names[i] != n {
			t.Errorf("order[%d] = %q, want %q", i, names[i], n)
		}
	}
	if len(all) != 3 {
		t.Errorf("expected 3 registered, got %d", len(all))
	}
}
