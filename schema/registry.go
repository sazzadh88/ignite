package schema

import "sort"

// Go cannot discover migration files at runtime the way Laravel scans a
// directory. Instead each generated migration registers itself here from an
// init() function; the migrate command then loads them in name order
// (filenames are timestamp-prefixed, so name order is chronological).

var (
	registeredMigrations = map[string]Migration{}
	registeredOrder      []string
)

// RegisterMigration records a migration under a unique name (its filename
// without extension, e.g. "2026_05_18_120000_create_posts_table").
func RegisterMigration(name string, m Migration) {
	if _, exists := registeredMigrations[name]; !exists {
		registeredOrder = append(registeredOrder, name)
	}
	registeredMigrations[name] = m
}

// RegisteredMigrations returns all registered migrations sorted by name.
func RegisteredMigrations() ([]string, map[string]Migration) {
	names := append([]string(nil), registeredOrder...)
	sort.Strings(names)
	return names, registeredMigrations
}

// LoadRegistered registers every globally-registered migration with this
// Migrator, in chronological (name) order.
func (m *Migrator) LoadRegistered() {
	names, all := RegisteredMigrations()
	for _, name := range names {
		m.Register(name, all[name])
	}
}
