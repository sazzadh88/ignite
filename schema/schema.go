package schema

import (
	"fmt"
)

// Connection is an interface that represents a database connection.
// This allows Schema to work with different database implementations
// without depending on a specific ORM or database driver.
type Connection interface {
	// Exec executes a SQL statement and returns the number of affected rows.
	Exec(query string, args ...interface{}) (int64, error)

	// Query executes a SQL query and returns results as a slice of maps.
	Query(query string, args ...interface{}) ([]map[string]interface{}, error)
}

// Schema provides methods for creating and modifying database tables.
// It uses a fluent Blueprint API to define table structure and generates
// the appropriate SQL statements for the database.
type Schema struct {
	connection Connection
}

// NewSchema creates a new Schema instance with the given database connection.
func NewSchema(conn Connection) *Schema {
	return &Schema{
		connection: conn,
	}
}

// Create creates a new table with the given name using the Blueprint callback.
// The callback receives a Blueprint instance to define columns, indexes, and constraints.
func (s *Schema) Create(table string, fn func(*Blueprint)) error {
	blueprint := NewBlueprint()
	fn(blueprint)

	sql := blueprint.ToCreateSQL(table)
	_, err := s.connection.Exec(sql)
	return err
}

// Table modifies an existing table using the Blueprint callback.
// The callback receives a Blueprint instance to add, modify, or drop columns and indexes.
func (s *Schema) Table(table string, fn func(*Blueprint)) error {
	blueprint := NewBlueprint()
	fn(blueprint)

	sql := blueprint.ToAlterSQL(table)
	if sql == "" {
		return nil
	}

	_, err := s.connection.Exec(sql)
	return err
}

// Drop drops the specified table.
func (s *Schema) Drop(table string) error {
	sql := fmt.Sprintf("DROP TABLE `%s`", table)
	_, err := s.connection.Exec(sql)
	return err
}

// DropIfExists drops the specified table if it exists.
func (s *Schema) DropIfExists(table string) error {
	sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)
	_, err := s.connection.Exec(sql)
	return err
}

// Rename renames a table from the old name to the new name.
func (s *Schema) Rename(from, to string) error {
	sql := fmt.Sprintf("RENAME TABLE `%s` TO `%s`", from, to)
	_, err := s.connection.Exec(sql)
	return err
}

// HasTable checks if the specified table exists in the database.
func (s *Schema) HasTable(table string) bool {
	sql := "SELECT COUNT(*) as count FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?"
	results, err := s.connection.Query(sql, table)
	if err != nil || len(results) == 0 {
		return false
	}

	count, ok := results[0]["count"].(int64)
	return ok && count > 0
}

// HasColumn checks if the specified column exists in the table.
func (s *Schema) HasColumn(table, column string) bool {
	sql := "SELECT COUNT(*) as count FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	results, err := s.connection.Query(sql, table, column)
	if err != nil || len(results) == 0 {
		return false
	}

	count, ok := results[0]["count"].(int64)
	return ok && count > 0
}

// GetColumnType returns the data type of the specified column in the table.
func (s *Schema) GetColumnType(table, column string) string {
	sql := "SELECT data_type FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?"
	results, err := s.connection.Query(sql, table, column)
	if err != nil || len(results) == 0 {
		return ""
	}

	dataType, ok := results[0]["data_type"].(string)
	if !ok {
		return ""
	}
	return dataType
}

// Truncate removes all rows from the specified table.
func (s *Schema) Truncate(table string) error {
	sql := fmt.Sprintf("TRUNCATE TABLE `%s`", table)
	_, err := s.connection.Exec(sql)
	return err
}
