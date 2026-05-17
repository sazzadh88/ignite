package schema

import (
	"fmt"
	"strings"
)

// Blueprint provides a fluent API for defining table schema.
// It collects column definitions, indexes, and constraints that
// can be converted to CREATE TABLE or ALTER TABLE SQL statements.
type Blueprint struct {
	columns         []*ColumnDefinition
	primaryKeys     []string
	indexes         []indexDefinition
	uniqueIndexes   []indexDefinition
	foreignKeys     []foreignKeyDefinition
	dropColumns     []string
	renameColumns   map[string]string
	commands        []command
}

// indexDefinition represents an index on one or more columns.
type indexDefinition struct {
	name    string
	columns []string
}

// foreignKeyDefinition represents a foreign key constraint.
type foreignKeyDefinition struct {
	column         string
	referencedTable string
	referencedColumn string
	onDelete       string
	onUpdate       string
}

// command represents a schema modification command.
type command struct {
	commandType string
	params      map[string]interface{}
}

// NewBlueprint creates a new Blueprint instance.
func NewBlueprint() *Blueprint {
	return &Blueprint{
		columns:       make([]*ColumnDefinition, 0),
		renameColumns: make(map[string]string),
	}
}

// ID creates a BIGINT UNSIGNED AUTO_INCREMENT primary key column named "id".
func (b *Blueprint) ID() *ColumnDefinition {
	col := newColumn("id", "BIGINT")
	col.unsigned = true
	col.autoIncrement = true
	col.primary = true
	b.columns = append(b.columns, col)
	b.primaryKeys = append(b.primaryKeys, "id")
	return col
}

// String creates a VARCHAR column with optional length (default 255).
func (b *Blueprint) String(name string, length ...int) *ColumnDefinition {
	l := 255
	if len(length) > 0 && length[0] > 0 {
		l = length[0]
	}
	col := newColumn(name, "VARCHAR")
	col.length = l
	b.columns = append(b.columns, col)
	return col
}

// Text creates a TEXT column.
func (b *Blueprint) Text(name string) *ColumnDefinition {
	col := newColumn(name, "TEXT")
	b.columns = append(b.columns, col)
	return col
}

// LongText creates a LONGTEXT column.
func (b *Blueprint) LongText(name string) *ColumnDefinition {
	col := newColumn(name, "LONGTEXT")
	b.columns = append(b.columns, col)
	return col
}

// MediumText creates a MEDIUMTEXT column.
func (b *Blueprint) MediumText(name string) *ColumnDefinition {
	col := newColumn(name, "MEDIUMTEXT")
	b.columns = append(b.columns, col)
	return col
}

// Integer creates an INT column.
func (b *Blueprint) Integer(name string) *ColumnDefinition {
	col := newColumn(name, "INT")
	b.columns = append(b.columns, col)
	return col
}

// BigInteger creates a BIGINT column.
func (b *Blueprint) BigInteger(name string) *ColumnDefinition {
	col := newColumn(name, "BIGINT")
	b.columns = append(b.columns, col)
	return col
}

// SmallInteger creates a SMALLINT column.
func (b *Blueprint) SmallInteger(name string) *ColumnDefinition {
	col := newColumn(name, "SMALLINT")
	b.columns = append(b.columns, col)
	return col
}

// TinyInteger creates a TINYINT column.
func (b *Blueprint) TinyInteger(name string) *ColumnDefinition {
	col := newColumn(name, "TINYINT")
	b.columns = append(b.columns, col)
	return col
}

// Float creates a FLOAT column with optional precision and scale.
func (b *Blueprint) Float(name string, precision, scale int) *ColumnDefinition {
	col := newColumn(name, "FLOAT")
	col.precision = precision
	col.scale = scale
	b.columns = append(b.columns, col)
	return col
}

// Decimal creates a DECIMAL column with precision and scale.
func (b *Blueprint) Decimal(name string, precision, scale int) *ColumnDefinition {
	col := newColumn(name, "DECIMAL")
	col.precision = precision
	col.scale = scale
	b.columns = append(b.columns, col)
	return col
}

// Double creates a DOUBLE column.
func (b *Blueprint) Double(name string) *ColumnDefinition {
	col := newColumn(name, "DOUBLE")
	b.columns = append(b.columns, col)
	return col
}

// Boolean creates a TINYINT(1) column.
func (b *Blueprint) Boolean(name string) *ColumnDefinition {
	col := newColumn(name, "TINYINT")
	col.length = 1
	b.columns = append(b.columns, col)
	return col
}

// Date creates a DATE column.
func (b *Blueprint) Date(name string) *ColumnDefinition {
	col := newColumn(name, "DATE")
	b.columns = append(b.columns, col)
	return col
}

// DateTime creates a DATETIME column.
func (b *Blueprint) DateTime(name string) *ColumnDefinition {
	col := newColumn(name, "DATETIME")
	b.columns = append(b.columns, col)
	return col
}

// Timestamp creates a TIMESTAMP column.
func (b *Blueprint) Timestamp(name string) *ColumnDefinition {
	col := newColumn(name, "TIMESTAMP")
	b.columns = append(b.columns, col)
	return col
}

// Time creates a TIME column.
func (b *Blueprint) Time(name string) *ColumnDefinition {
	col := newColumn(name, "TIME")
	b.columns = append(b.columns, col)
	return col
}

// Timestamps creates created_at and updated_at TIMESTAMP columns.
func (b *Blueprint) Timestamps() {
	createdAt := b.Timestamp("created_at")
	createdAt.nullable = false
	defaultCreated := "CURRENT_TIMESTAMP"
	createdAt.defaultValue = &defaultCreated

	updatedAt := b.Timestamp("updated_at")
	updatedAt.nullable = false
	defaultUpdated := "CURRENT_TIMESTAMP"
	updatedAt.defaultValue = &defaultUpdated
}

// SoftDeletes creates a nullable deleted_at TIMESTAMP column.
func (b *Blueprint) SoftDeletes() *ColumnDefinition {
	col := b.Timestamp("deleted_at")
	col.nullable = true
	return col
}

// JSON creates a JSON column.
func (b *Blueprint) JSON(name string) *ColumnDefinition {
	col := newColumn(name, "JSON")
	b.columns = append(b.columns, col)
	return col
}

// JSONB creates a JSON column (MySQL doesn't have JSONB, so we use JSON).
func (b *Blueprint) JSONB(name string) *ColumnDefinition {
	return b.JSON(name)
}

// Enum creates an ENUM column with the given values.
func (b *Blueprint) Enum(name string, values []string) *ColumnDefinition {
	col := newColumn(name, "ENUM")
	col.values = values
	b.columns = append(b.columns, col)
	return col
}

// Binary creates a BLOB column.
func (b *Blueprint) Binary(name string) *ColumnDefinition {
	col := newColumn(name, "BLOB")
	b.columns = append(b.columns, col)
	return col
}

// UUID creates a CHAR(36) column for UUID storage.
func (b *Blueprint) UUID(name string) *ColumnDefinition {
	col := newColumn(name, "CHAR")
	col.length = 36
	b.columns = append(b.columns, col)
	return col
}

// ULID creates a CHAR(26) column for ULID storage.
func (b *Blueprint) ULID(name string) *ColumnDefinition {
	col := newColumn(name, "CHAR")
	col.length = 26
	b.columns = append(b.columns, col)
	return col
}

// MorphColumns creates name_id and name_type columns for polymorphic relations.
func (b *Blueprint) MorphColumns(name string) {
	b.BigInteger(name + "_id").Unsigned()
	b.String(name + "_type")
}

// RememberToken creates a VARCHAR(100) nullable column named "remember_token".
func (b *Blueprint) RememberToken() *ColumnDefinition {
	col := b.String("remember_token", 100)
	col.nullable = true
	return col
}

// ForeignId creates an unsigned BIGINT column suitable for foreign keys.
func (b *Blueprint) ForeignId(name string) *ForeignIdBuilder {
	col := b.BigInteger(name)
	col.unsigned = true
	return &ForeignIdBuilder{
		blueprint:  b,
		columnName: name,
	}
}

// ForeignIdBuilder provides methods for defining foreign key constraints.
type ForeignIdBuilder struct {
	blueprint  *Blueprint
	columnName string
}

// Constrained adds a foreign key constraint referencing the primary key
// of the table with the same name as the column (without _id suffix).
func (f *ForeignIdBuilder) Constrained(table ...string) *ForeignIdBuilder {
	tableName := strings.TrimSuffix(f.columnName, "_id")
	if len(table) > 0 {
		tableName = table[0]
	}

	f.blueprint.foreignKeys = append(f.blueprint.foreignKeys, foreignKeyDefinition{
		column:           f.columnName,
		referencedTable:  tableName,
		referencedColumn: "id",
	})
	return f
}

// CascadeOnDelete sets CASCADE on delete for the foreign key.
func (f *ForeignIdBuilder) CascadeOnDelete() *ForeignIdBuilder {
	if len(f.blueprint.foreignKeys) > 0 {
		lastFK := &f.blueprint.foreignKeys[len(f.blueprint.foreignKeys)-1]
		lastFK.onDelete = "CASCADE"
	}
	return f
}

// Index creates a regular index on the given columns.
func (b *Blueprint) Index(columns []string, name ...string) {
	indexName := ""
	if len(name) > 0 {
		indexName = name[0]
	} else {
		indexName = "idx_" + strings.Join(columns, "_")
	}
	b.indexes = append(b.indexes, indexDefinition{
		name:    indexName,
		columns: columns,
	})
}

// Unique creates a unique index on the given columns.
func (b *Blueprint) Unique(columns []string, name ...string) {
	indexName := ""
	if len(name) > 0 {
		indexName = name[0]
	} else {
		indexName = "uniq_" + strings.Join(columns, "_")
	}
	b.uniqueIndexes = append(b.uniqueIndexes, indexDefinition{
		name:    indexName,
		columns: columns,
	})
}

// Primary sets the primary key for the table.
func (b *Blueprint) Primary(columns []string) {
	b.primaryKeys = columns
}

// FullText creates a fulltext index on the given columns.
func (b *Blueprint) FullText(columns []string, name ...string) {
	indexName := ""
	if len(name) > 0 {
		indexName = name[0]
	} else {
		indexName = "ft_" + strings.Join(columns, "_")
	}
	b.commands = append(b.commands, command{
		commandType: "fulltext",
		params: map[string]interface{}{
			"name":    indexName,
			"columns": columns,
		},
	})
}

// DropIndex drops an index by name.
func (b *Blueprint) DropIndex(name string) {
	b.commands = append(b.commands, command{
		commandType: "dropIndex",
		params:      map[string]interface{}{"name": name},
	})
}

// DropUnique drops a unique index by name.
func (b *Blueprint) DropUnique(name string) {
	b.commands = append(b.commands, command{
		commandType: "dropUnique",
		params:      map[string]interface{}{"name": name},
	})
}

// DropForeign drops a foreign key constraint by name.
func (b *Blueprint) DropForeign(name string) {
	b.commands = append(b.commands, command{
		commandType: "dropForeign",
		params:      map[string]interface{}{"name": name},
	})
}

// DropPrimary drops the primary key constraint.
func (b *Blueprint) DropPrimary() {
	b.commands = append(b.commands, command{
		commandType: "dropPrimary",
		params:      make(map[string]interface{}),
	})
}

// DropColumn marks columns for deletion (ALTER TABLE operations).
func (b *Blueprint) DropColumn(names ...string) {
	b.dropColumns = append(b.dropColumns, names...)
}

// RenameColumn marks a column for renaming (ALTER TABLE operations).
func (b *Blueprint) RenameColumn(from, to string) {
	b.renameColumns[from] = to
}

// ModifyColumn modifies an existing column's type and length.
func (b *Blueprint) ModifyColumn(name, typ string, length ...int) *ColumnDefinition {
	col := newColumn(name, typ)
	if len(length) > 0 {
		col.length = length[0]
	}
	b.commands = append(b.commands, command{
		commandType: "modify",
		params:      map[string]interface{}{"column": col},
	})
	return col
}

// ToCreateSQL generates a CREATE TABLE SQL statement.
func (b *Blueprint) ToCreateSQL(table string) string {
	var parts []string

	// Column definitions
	for _, col := range b.columns {
		parts = append(parts, col.toSQL())
	}

	// Primary key
	if len(b.primaryKeys) > 0 {
		pkCols := make([]string, len(b.primaryKeys))
		for i, pk := range b.primaryKeys {
			pkCols[i] = fmt.Sprintf("`%s`", pk)
		}
		parts = append(parts, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkCols, ", ")))
	}

	// Indexes
	for _, idx := range b.indexes {
		idxCols := make([]string, len(idx.columns))
		for i, col := range idx.columns {
			idxCols[i] = fmt.Sprintf("`%s`", col)
		}
		parts = append(parts, fmt.Sprintf("INDEX `%s` (%s)", idx.name, strings.Join(idxCols, ", ")))
	}

	// Unique indexes
	for _, idx := range b.uniqueIndexes {
		idxCols := make([]string, len(idx.columns))
		for i, col := range idx.columns {
			idxCols[i] = fmt.Sprintf("`%s`", col)
		}
		parts = append(parts, fmt.Sprintf("UNIQUE INDEX `%s` (%s)", idx.name, strings.Join(idxCols, ", ")))
	}

	// Foreign keys
	for _, fk := range b.foreignKeys {
		fkSQL := fmt.Sprintf("FOREIGN KEY (`%s`) REFERENCES `%s`(`%s`)",
			fk.column, fk.referencedTable, fk.referencedColumn)
		if fk.onDelete != "" {
			fkSQL += fmt.Sprintf(" ON DELETE %s", fk.onDelete)
		}
		if fk.onUpdate != "" {
			fkSQL += fmt.Sprintf(" ON UPDATE %s", fk.onUpdate)
		}
		parts = append(parts, fkSQL)
	}

	// Fulltext indexes
	for _, cmd := range b.commands {
		if cmd.commandType == "fulltext" {
			columns := cmd.params["columns"].([]string)
			name := cmd.params["name"].(string)
			ftCols := make([]string, len(columns))
			for i, col := range columns {
				ftCols[i] = fmt.Sprintf("`%s`", col)
			}
			parts = append(parts, fmt.Sprintf("FULLTEXT INDEX `%s` (%s)", name, strings.Join(ftCols, ", ")))
		}
	}

	sql := fmt.Sprintf("CREATE TABLE `%s` (\n  %s\n) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci",
		table, strings.Join(parts, ",\n  "))

	return sql
}

// ToAlterSQL generates ALTER TABLE SQL statements.
func (b *Blueprint) ToAlterSQL(table string) string {
	var statements []string

	// Add columns
	for _, col := range b.columns {
		stmt := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s", table, col.toSQL())
		if col.after != "" {
			stmt += fmt.Sprintf(" AFTER `%s`", col.after)
		} else if col.first {
			stmt += " FIRST"
		}
		statements = append(statements, stmt)
	}

	// Drop columns
	for _, colName := range b.dropColumns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` DROP COLUMN `%s`", table, colName))
	}

	// Rename columns
	for from, to := range b.renameColumns {
		statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` RENAME COLUMN `%s` TO `%s`", table, from, to))
	}

	// Modify columns
	for _, cmd := range b.commands {
		if cmd.commandType == "modify" {
			col := cmd.params["column"].(*ColumnDefinition)
			statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` MODIFY COLUMN %s", table, col.toSQL()))
		} else if cmd.commandType == "dropIndex" {
			statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", table, cmd.params["name"].(string)))
		} else if cmd.commandType == "dropUnique" {
			statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` DROP INDEX `%s`", table, cmd.params["name"].(string)))
		} else if cmd.commandType == "dropForeign" {
			statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` DROP FOREIGN KEY `%s`", table, cmd.params["name"].(string)))
		} else if cmd.commandType == "dropPrimary" {
			statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` DROP PRIMARY KEY", table))
		}
	}

	// Add indexes
	for _, idx := range b.indexes {
		idxCols := make([]string, len(idx.columns))
		for i, col := range idx.columns {
			idxCols[i] = fmt.Sprintf("`%s`", col)
		}
		statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` ADD INDEX `%s` (%s)", table, idx.name, strings.Join(idxCols, ", ")))
	}

	// Add unique indexes
	for _, idx := range b.uniqueIndexes {
		idxCols := make([]string, len(idx.columns))
		for i, col := range idx.columns {
			idxCols[i] = fmt.Sprintf("`%s`", col)
		}
		statements = append(statements, fmt.Sprintf("ALTER TABLE `%s` ADD UNIQUE INDEX `%s` (%s)", table, idx.name, strings.Join(idxCols, ", ")))
	}

	return strings.Join(statements, ";\n")
}
