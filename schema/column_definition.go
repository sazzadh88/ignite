// Package schema provides database schema building functionality for Ignite.
// It implements a Laravel-inspired Schema Builder and Blueprint API with
// support for table creation, modification, and migration management.
package schema

import (
	"fmt"
	"strings"
)

// ColumnDefinition represents a single column in a table schema.
// It holds the column's name, type, and all modifiers that can be
// chained to customize the column's behavior.
type ColumnDefinition struct {
	name          string
	columnType    string
	length        int
	precision     int
	scale         int
	nullable      bool
	defaultValue  *string
	unsigned      bool
	unique        bool
	primary       bool
	autoIncrement bool
	after         string
	first         bool
	comment       string
	index         bool
	charset       string
	collation     string
	values        []string // For enum type
}

// newColumn creates a new ColumnDefinition with the given name and type.
func newColumn(name, columnType string) *ColumnDefinition {
	return &ColumnDefinition{
		name:       name,
		columnType: columnType,
	}
}

// Nullable marks the column as allowing NULL values.
func (c *ColumnDefinition) Nullable() *ColumnDefinition {
	c.nullable = true
	return c
}

// Default sets the default value for the column.
func (c *ColumnDefinition) Default(value string) *ColumnDefinition {
	c.defaultValue = &value
	return c
}

// Unsigned marks the column as unsigned (numeric types only).
func (c *ColumnDefinition) Unsigned() *ColumnDefinition {
	c.unsigned = true
	return c
}

// Unique adds a unique index to the column.
func (c *ColumnDefinition) Unique() *ColumnDefinition {
	c.unique = true
	return c
}

// Primary marks the column as the primary key.
func (c *ColumnDefinition) Primary() *ColumnDefinition {
	c.primary = true
	return c
}

// After positions the column after the specified column (ALTER TABLE operations).
func (c *ColumnDefinition) After(column string) *ColumnDefinition {
	c.after = column
	return c
}

// First positions the column first in the table (ALTER TABLE operations).
func (c *ColumnDefinition) First() *ColumnDefinition {
	c.first = true
	return c
}

// Comment sets a comment for the column.
func (c *ColumnDefinition) Comment(text string) *ColumnDefinition {
	c.comment = text
	return c
}

// Index adds a regular index to the column.
func (c *ColumnDefinition) Index() *ColumnDefinition {
	c.index = true
	return c
}

// AutoIncrement marks the column as auto-incrementing.
func (c *ColumnDefinition) AutoIncrement() *ColumnDefinition {
	c.autoIncrement = true
	return c
}

// Charset sets the character set for the column.
func (c *ColumnDefinition) Charset(cs string) *ColumnDefinition {
	c.charset = cs
	return c
}

// Collation sets the collation for the column.
func (c *ColumnDefinition) Collation(co string) *ColumnDefinition {
	c.collation = co
	return c
}

// toSQL generates the SQL definition for this column.
func (c *ColumnDefinition) toSQL() string {
	var parts []string

	// Column name
	parts = append(parts, fmt.Sprintf("`%s`", c.name))

	// Column type with length/precision
	typeStr := strings.ToUpper(c.columnType)
	if c.length > 0 {
		typeStr = fmt.Sprintf("%s(%d)", typeStr, c.length)
	} else if c.precision > 0 {
		if c.scale > 0 {
			typeStr = fmt.Sprintf("%s(%d,%d)", typeStr, c.precision, c.scale)
		} else {
			typeStr = fmt.Sprintf("%s(%d)", typeStr, c.precision)
		}
	} else if len(c.values) > 0 {
		// For ENUM type
		quotedValues := make([]string, len(c.values))
		for i, v := range c.values {
			quotedValues[i] = fmt.Sprintf("'%s'", strings.ReplaceAll(v, "'", "''"))
		}
		typeStr = fmt.Sprintf("%s(%s)", typeStr, strings.Join(quotedValues, ", "))
	}

	if c.unsigned {
		typeStr += " UNSIGNED"
	}

	parts = append(parts, typeStr)

	// Character set and collation
	if c.charset != "" {
		parts = append(parts, fmt.Sprintf("CHARACTER SET %s", c.charset))
	}
	if c.collation != "" {
		parts = append(parts, fmt.Sprintf("COLLATE %s", c.collation))
	}

	// Nullable
	if c.nullable {
		parts = append(parts, "NULL")
	} else {
		parts = append(parts, "NOT NULL")
	}

	// Default value
	if c.defaultValue != nil {
		if *c.defaultValue == "CURRENT_TIMESTAMP" || *c.defaultValue == "NULL" {
			parts = append(parts, fmt.Sprintf("DEFAULT %s", *c.defaultValue))
		} else {
			parts = append(parts, fmt.Sprintf("DEFAULT '%s'", strings.ReplaceAll(*c.defaultValue, "'", "''")))
		}
	}

	// Auto increment
	if c.autoIncrement {
		parts = append(parts, "AUTO_INCREMENT")
	}

	// Comment
	if c.comment != "" {
		parts = append(parts, fmt.Sprintf("COMMENT '%s'", strings.ReplaceAll(c.comment, "'", "''")))
	}

	return strings.Join(parts, " ")
}
