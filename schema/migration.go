package schema

import (
	"fmt"
	"sort"
	"time"
)

// Migration defines the interface for database migrations.
// Each migration must implement Up (apply changes) and Down (revert changes).
type Migration interface {
	// Up applies the migration changes to the database.
	Up(schema *Schema) error

	// Down reverts the migration changes.
	Down(schema *Schema) error
}

// MigrationRecord represents a migration that has been executed.
type MigrationRecord struct {
	ID        int64
	Migration string
	Batch     int
	RanAt     time.Time
}

// MigrationStatus represents the status of a migration.
type MigrationStatus struct {
	Name   string
	Batch  int
	RanAt  time.Time
	Status string // "pending" or "ran"
}

// Migrator manages database migrations.
type Migrator struct {
	schema     *Schema
	migrations map[string]Migration
	order      []string
}

// NewMigrator creates a new Migrator instance.
func NewMigrator(schema *Schema) *Migrator {
	return &Migrator{
		schema:     schema,
		migrations: make(map[string]Migration),
		order:      make([]string, 0),
	}
}

// Register registers a migration with the given name.
// Migrations should be registered in chronological order.
func (m *Migrator) Register(name string, migration Migration) {
	m.migrations[name] = migration
	m.order = append(m.order, name)
}

// createMigrationsTable creates the migrations tracking table if it doesn't exist.
func (m *Migrator) createMigrationsTable() error {
	return m.schema.Create("migrations", func(b *Blueprint) {
		b.ID()
		b.String("migration")
		b.Integer("batch")
		b.Timestamp("ran_at").Default("CURRENT_TIMESTAMP")
	})
}

// ensureMigrationsTable ensures the migrations table exists.
func (m *Migrator) ensureMigrationsTable() error {
	if !m.schema.HasTable("migrations") {
		return m.createMigrationsTable()
	}
	return nil
}

// getRanMigrations retrieves all migrations that have been executed.
func (m *Migrator) getRanMigrations() ([]MigrationRecord, error) {
	results, err := m.schema.connection.Query("SELECT id, migration, batch, ran_at FROM migrations ORDER BY batch, id")
	if err != nil {
		return nil, err
	}

	records := make([]MigrationRecord, len(results))
	for i, row := range results {
		records[i] = MigrationRecord{
			ID:        row["id"].(int64),
			Migration: row["migration"].(string),
			Batch:     int(row["batch"].(int64)),
			RanAt:     row["ran_at"].(time.Time),
		}
	}

	return records, nil
}

// getLastBatch returns the last batch number.
func (m *Migrator) getLastBatch() (int, error) {
	results, err := m.schema.connection.Query("SELECT MAX(batch) as max_batch FROM migrations")
	if err != nil {
		return 0, err
	}

	if len(results) == 0 || results[0]["max_batch"] == nil {
		return 0, nil
	}

	maxBatch, ok := results[0]["max_batch"].(int64)
	if !ok {
		return 0, nil
	}

	return int(maxBatch), nil
}

// Migrate runs all pending migrations.
func (m *Migrator) Migrate() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	ranMigrations, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Create a map of ran migrations for quick lookup
	ranMap := make(map[string]bool)
	for _, record := range ranMigrations {
		ranMap[record.Migration] = true
	}

	// Get the next batch number
	lastBatch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}
	nextBatch := lastBatch + 1

	// Run pending migrations
	for _, name := range m.order {
		if ranMap[name] {
			continue
		}

		migration := m.migrations[name]
		if err := migration.Up(m.schema); err != nil {
			return fmt.Errorf("migration %s failed: %w", name, err)
		}

		// Record the migration
		_, err := m.schema.connection.Exec(
			"INSERT INTO migrations (migration, batch) VALUES (?, ?)",
			name, nextBatch,
		)
		if err != nil {
			return fmt.Errorf("failed to record migration %s: %w", name, err)
		}
	}

	return nil
}

// Rollback rolls back the last batch of migrations.
func (m *Migrator) Rollback() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	lastBatch, err := m.getLastBatch()
	if err != nil {
		return fmt.Errorf("failed to get last batch: %w", err)
	}

	if lastBatch == 0 {
		return nil // Nothing to rollback
	}

	// Get migrations from the last batch
	results, err := m.schema.connection.Query(
		"SELECT migration FROM migrations WHERE batch = ? ORDER BY id DESC",
		lastBatch,
	)
	if err != nil {
		return fmt.Errorf("failed to get migrations for batch %d: %w", lastBatch, err)
	}

	// Rollback each migration in reverse order
	for _, row := range results {
		name := row["migration"].(string)
		migration, exists := m.migrations[name]
		if !exists {
			return fmt.Errorf("migration %s not found", name)
		}

		if err := migration.Down(m.schema); err != nil {
			return fmt.Errorf("rollback of migration %s failed: %w", name, err)
		}

		// Remove the migration record
		_, err := m.schema.connection.Exec(
			"DELETE FROM migrations WHERE migration = ?",
			name,
		)
		if err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", name, err)
		}
	}

	return nil
}

// Fresh drops all tables and re-runs all migrations.
func (m *Migrator) Fresh() error {
	// Get all tables
	results, err := m.schema.connection.Query(
		"SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()",
	)
	if err != nil {
		return fmt.Errorf("failed to get tables: %w", err)
	}

	// Drop all tables
	for _, row := range results {
		tableName := row["table_name"].(string)
		if err := m.schema.Drop(tableName); err != nil {
			return fmt.Errorf("failed to drop table %s: %w", tableName, err)
		}
	}

	// Re-run all migrations
	return m.Migrate()
}

// Status returns the status of all migrations.
func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := m.ensureMigrationsTable(); err != nil {
		return nil, fmt.Errorf("failed to create migrations table: %w", err)
	}

	ranMigrations, err := m.getRanMigrations()
	if err != nil {
		return nil, fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Create a map of ran migrations for quick lookup
	ranMap := make(map[string]MigrationRecord)
	for _, record := range ranMigrations {
		ranMap[record.Migration] = record
	}

	// Build status list
	statuses := make([]MigrationStatus, 0, len(m.order))
	for _, name := range m.order {
		if record, ran := ranMap[name]; ran {
			statuses = append(statuses, MigrationStatus{
				Name:   name,
				Batch:  record.Batch,
				RanAt:  record.RanAt,
				Status: "ran",
			})
		} else {
			statuses = append(statuses, MigrationStatus{
				Name:   name,
				Status: "pending",
			})
		}
	}

	return statuses, nil
}

// Reset rolls back all migrations.
func (m *Migrator) Reset() error {
	if err := m.ensureMigrationsTable(); err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Get all ran migrations
	ranMigrations, err := m.getRanMigrations()
	if err != nil {
		return fmt.Errorf("failed to get ran migrations: %w", err)
	}

	// Sort by batch descending, then ID descending
	sort.Slice(ranMigrations, func(i, j int) bool {
		if ranMigrations[i].Batch == ranMigrations[j].Batch {
			return ranMigrations[i].ID > ranMigrations[j].ID
		}
		return ranMigrations[i].Batch > ranMigrations[j].Batch
	})

	// Rollback each migration
	for _, record := range ranMigrations {
		migration, exists := m.migrations[record.Migration]
		if !exists {
			return fmt.Errorf("migration %s not found", record.Migration)
		}

		if err := migration.Down(m.schema); err != nil {
			return fmt.Errorf("rollback of migration %s failed: %w", record.Migration, err)
		}

		// Remove the migration record
		_, err := m.schema.connection.Exec(
			"DELETE FROM migrations WHERE id = ?",
			record.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to remove migration record %s: %w", record.Migration, err)
		}
	}

	return nil
}
