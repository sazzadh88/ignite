package schema

import (
	"fmt"
	"strings"
)

// Dialect compiles Blueprint/Schema operations into driver-specific SQL.
// It is the analog of Laravel's schema grammars (MySqlGrammar,
// SQLiteGrammar, PostgresGrammar). MySQL is the default and preserves the
// original behaviour; sqlite/postgres add their own DDL.
type Dialect interface {
	Name() string
	// Placeholder returns the bind placeholder for the i-th (1-based)
	// parameter: "?" for mysql/sqlite, "$i" for postgres.
	Placeholder(i int) string
	// CompileCreate returns one or more statements to create a table
	// (e.g. a CREATE TABLE plus any standalone CREATE INDEX statements).
	CompileCreate(table string, b *Blueprint) []string
	// CompileAlter returns statements for ALTER TABLE operations.
	CompileAlter(table string, b *Blueprint) []string
	CompileDrop(table string) string
	CompileDropIfExists(table string) string
	CompileRename(from, to string) string
	CompileTruncate(table string) string
	// CompileTableExists returns a query and args; a non-empty result set
	// means the table exists.
	CompileTableExists(table string) (string, []any)
	// CompileColumnExists returns a query and args for column existence.
	CompileColumnExists(table, column string) (string, []any)
	// CompileTables returns a query listing user tables and the result
	// column that holds the table name.
	CompileTables() (query string, nameColumn string)
}

// DialectFor maps a Go sql driver name to its dialect. Unknown drivers
// fall back to MySQL to preserve historical behaviour.
func DialectFor(driver string) Dialect {
	switch driver {
	case "sqlite3", "sqlite":
		return sqliteDialect{}
	case "postgres", "pgx", "pgsql":
		return postgresDialect{}
	default:
		return mysqlDialect{}
	}
}

// ---------- MySQL (default, original behaviour) ----------

type mysqlDialect struct{}

func (mysqlDialect) Name() string            { return "mysql" }
func (mysqlDialect) Placeholder(int) string  { return "?" }

func (mysqlDialect) CompileCreate(table string, b *Blueprint) []string {
	return []string{b.ToCreateSQL(table)}
}

func (mysqlDialect) CompileAlter(table string, b *Blueprint) []string {
	sql := b.ToAlterSQL(table)
	if sql == "" {
		return nil
	}
	return splitStatements(sql)
}

func (mysqlDialect) CompileDrop(table string) string {
	return fmt.Sprintf("DROP TABLE `%s`", table)
}
func (mysqlDialect) CompileDropIfExists(table string) string {
	return fmt.Sprintf("DROP TABLE IF EXISTS `%s`", table)
}
func (mysqlDialect) CompileRename(from, to string) string {
	return fmt.Sprintf("RENAME TABLE `%s` TO `%s`", from, to)
}
func (mysqlDialect) CompileTruncate(table string) string {
	return fmt.Sprintf("TRUNCATE TABLE `%s`", table)
}
func (mysqlDialect) CompileTableExists(table string) (string, []any) {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = ?", []any{table}
}
func (mysqlDialect) CompileColumnExists(table, column string) (string, []any) {
	return "SELECT column_name FROM information_schema.columns WHERE table_schema = DATABASE() AND table_name = ? AND column_name = ?", []any{table, column}
}
func (mysqlDialect) CompileTables() (string, string) {
	return "SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE()", "table_name"
}

// ---------- SQLite ----------

type sqliteDialect struct{}

func (sqliteDialect) Name() string           { return "sqlite" }
func (sqliteDialect) Placeholder(int) string { return "?" }

func (sqliteDialect) CompileCreate(table string, b *Blueprint) []string {
	var cols []string
	autoPK := ""
	for _, c := range b.columns {
		if c.autoIncrement {
			autoPK = c.name
			cols = append(cols, fmt.Sprintf(`"%s" INTEGER PRIMARY KEY AUTOINCREMENT`, c.name))
			continue
		}
		cols = append(cols, sqliteColumn(c))
	}

	// Composite / explicit primary key (skip when handled inline above).
	if len(b.primaryKeys) > 0 && !(len(b.primaryKeys) == 1 && b.primaryKeys[0] == autoPK) {
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", quoteList(b.primaryKeys, `"`)))
	}

	for _, fk := range b.foreignKeys {
		s := fmt.Sprintf(`FOREIGN KEY ("%s") REFERENCES "%s"("%s")`, fk.column, fk.referencedTable, fk.referencedColumn)
		if fk.onDelete != "" {
			s += " ON DELETE " + fk.onDelete
		}
		if fk.onUpdate != "" {
			s += " ON UPDATE " + fk.onUpdate
		}
		cols = append(cols, s)
	}

	stmts := []string{fmt.Sprintf("CREATE TABLE \"%s\" (\n  %s\n)", table, strings.Join(cols, ",\n  "))}
	stmts = append(stmts, sqliteIndexes(table, b)...)
	return stmts
}

func (sqliteDialect) CompileAlter(table string, b *Blueprint) []string {
	var stmts []string
	for _, c := range b.columns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN %s`, table, sqliteColumn(c)))
	}
	for _, name := range b.dropColumns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN "%s"`, table, name))
	}
	for from, to := range b.renameColumns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" RENAME COLUMN "%s" TO "%s"`, table, from, to))
	}
	stmts = append(stmts, sqliteIndexes(table, b)...)
	return stmts
}

func (sqliteDialect) CompileDrop(table string) string {
	return fmt.Sprintf(`DROP TABLE "%s"`, table)
}
func (sqliteDialect) CompileDropIfExists(table string) string {
	return fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, table)
}
func (sqliteDialect) CompileRename(from, to string) string {
	return fmt.Sprintf(`ALTER TABLE "%s" RENAME TO "%s"`, from, to)
}
func (sqliteDialect) CompileTruncate(table string) string {
	return fmt.Sprintf(`DELETE FROM "%s"`, table)
}
func (sqliteDialect) CompileTableExists(table string) (string, []any) {
	return "SELECT name FROM sqlite_master WHERE type='table' AND name = ?", []any{table}
}
func (sqliteDialect) CompileColumnExists(table, column string) (string, []any) {
	// pragma_table_info is the portable way to introspect columns in sqlite.
	return "SELECT name FROM pragma_table_info(?) WHERE name = ?", []any{table, column}
}
func (sqliteDialect) CompileTables() (string, string) {
	return "SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'", "name"
}

func sqliteIndexes(table string, b *Blueprint) []string {
	var out []string
	for _, idx := range b.indexes {
		out = append(out, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "%s" ON "%s" (%s)`,
			idx.name, table, quoteList(idx.columns, `"`)))
	}
	for _, idx := range b.uniqueIndexes {
		out = append(out, fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS "%s" ON "%s" (%s)`,
			idx.name, table, quoteList(idx.columns, `"`)))
	}
	return out
}

func sqliteColumn(c *ColumnDefinition) string {
	parts := []string{fmt.Sprintf(`"%s"`, c.name), sqliteType(c)}
	if c.nullable {
		parts = append(parts, "NULL")
	} else {
		parts = append(parts, "NOT NULL")
	}
	if c.defaultValue != nil {
		parts = append(parts, "DEFAULT "+formatDefault(*c.defaultValue))
	}
	if c.unique {
		parts = append(parts, "UNIQUE")
	}
	return strings.Join(parts, " ")
}

func sqliteType(c *ColumnDefinition) string {
	switch strings.ToUpper(c.columnType) {
	case "INT", "INTEGER", "BIGINT", "SMALLINT", "TINYINT":
		return "INTEGER"
	case "FLOAT", "DOUBLE":
		return "REAL"
	case "DECIMAL":
		return "NUMERIC"
	case "BOOLEAN":
		return "INTEGER"
	case "DATE", "DATETIME", "TIMESTAMP", "TIME":
		return "DATETIME"
	case "BLOB", "BINARY":
		return "BLOB"
	case "VARCHAR", "CHAR":
		if c.length > 0 {
			return fmt.Sprintf("VARCHAR(%d)", c.length)
		}
		return "TEXT"
	default: // TEXT, LONGTEXT, MEDIUMTEXT, JSON, ENUM, UUID, ULID...
		return "TEXT"
	}
}

// ---------- PostgreSQL ----------

type postgresDialect struct{}

func (postgresDialect) Name() string             { return "postgres" }
func (postgresDialect) Placeholder(i int) string { return fmt.Sprintf("$%d", i) }

func (postgresDialect) CompileCreate(table string, b *Blueprint) []string {
	var cols []string
	autoPK := ""
	for _, c := range b.columns {
		if c.autoIncrement {
			autoPK = c.name
			cols = append(cols, fmt.Sprintf(`"%s" BIGSERIAL PRIMARY KEY`, c.name))
			continue
		}
		cols = append(cols, postgresColumn(c))
	}
	if len(b.primaryKeys) > 0 && !(len(b.primaryKeys) == 1 && b.primaryKeys[0] == autoPK) {
		cols = append(cols, fmt.Sprintf("PRIMARY KEY (%s)", quoteList(b.primaryKeys, `"`)))
	}
	for _, fk := range b.foreignKeys {
		s := fmt.Sprintf(`FOREIGN KEY ("%s") REFERENCES "%s"("%s")`, fk.column, fk.referencedTable, fk.referencedColumn)
		if fk.onDelete != "" {
			s += " ON DELETE " + fk.onDelete
		}
		if fk.onUpdate != "" {
			s += " ON UPDATE " + fk.onUpdate
		}
		cols = append(cols, s)
	}
	stmts := []string{fmt.Sprintf("CREATE TABLE \"%s\" (\n  %s\n)", table, strings.Join(cols, ",\n  "))}
	for _, idx := range b.indexes {
		stmts = append(stmts, fmt.Sprintf(`CREATE INDEX IF NOT EXISTS "%s" ON "%s" (%s)`,
			idx.name, table, quoteList(idx.columns, `"`)))
	}
	for _, idx := range b.uniqueIndexes {
		stmts = append(stmts, fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS "%s" ON "%s" (%s)`,
			idx.name, table, quoteList(idx.columns, `"`)))
	}
	return stmts
}

func (postgresDialect) CompileAlter(table string, b *Blueprint) []string {
	var stmts []string
	for _, c := range b.columns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" ADD COLUMN %s`, table, postgresColumn(c)))
	}
	for _, name := range b.dropColumns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" DROP COLUMN "%s"`, table, name))
	}
	for from, to := range b.renameColumns {
		stmts = append(stmts, fmt.Sprintf(`ALTER TABLE "%s" RENAME COLUMN "%s" TO "%s"`, table, from, to))
	}
	return stmts
}

func (postgresDialect) CompileDrop(table string) string {
	return fmt.Sprintf(`DROP TABLE "%s"`, table)
}
func (postgresDialect) CompileDropIfExists(table string) string {
	return fmt.Sprintf(`DROP TABLE IF EXISTS "%s"`, table)
}
func (postgresDialect) CompileRename(from, to string) string {
	return fmt.Sprintf(`ALTER TABLE "%s" RENAME TO "%s"`, from, to)
}
func (postgresDialect) CompileTruncate(table string) string {
	return fmt.Sprintf(`TRUNCATE TABLE "%s" RESTART IDENTITY CASCADE`, table)
}
func (postgresDialect) CompileTableExists(table string) (string, []any) {
	return "SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename = $1", []any{table}
}
func (postgresDialect) CompileColumnExists(table, column string) (string, []any) {
	return "SELECT column_name FROM information_schema.columns WHERE table_name = $1 AND column_name = $2", []any{table, column}
}
func (postgresDialect) CompileTables() (string, string) {
	return "SELECT tablename FROM pg_tables WHERE schemaname = 'public'", "tablename"
}

func postgresColumn(c *ColumnDefinition) string {
	parts := []string{fmt.Sprintf(`"%s"`, c.name), postgresType(c)}
	if c.nullable {
		parts = append(parts, "NULL")
	} else {
		parts = append(parts, "NOT NULL")
	}
	if c.defaultValue != nil {
		parts = append(parts, "DEFAULT "+formatDefault(*c.defaultValue))
	}
	if c.unique {
		parts = append(parts, "UNIQUE")
	}
	return strings.Join(parts, " ")
}

func postgresType(c *ColumnDefinition) string {
	switch strings.ToUpper(c.columnType) {
	case "INT", "INTEGER":
		return "INTEGER"
	case "BIGINT":
		return "BIGINT"
	case "SMALLINT", "TINYINT":
		return "SMALLINT"
	case "BOOLEAN":
		return "BOOLEAN"
	case "FLOAT":
		return "REAL"
	case "DOUBLE":
		return "DOUBLE PRECISION"
	case "DECIMAL":
		if c.precision > 0 {
			if c.scale > 0 {
				return fmt.Sprintf("NUMERIC(%d,%d)", c.precision, c.scale)
			}
			return fmt.Sprintf("NUMERIC(%d)", c.precision)
		}
		return "NUMERIC"
	case "DATE":
		return "DATE"
	case "TIME":
		return "TIME"
	case "DATETIME", "TIMESTAMP":
		return "TIMESTAMP"
	case "BLOB", "BINARY":
		return "BYTEA"
	case "JSON", "JSONB":
		return "JSONB"
	case "VARCHAR", "CHAR":
		if c.length > 0 {
			return fmt.Sprintf("%s(%d)", strings.ToUpper(c.columnType), c.length)
		}
		return "TEXT"
	default: // TEXT, LONGTEXT, MEDIUMTEXT, ENUM, UUID, ULID...
		return "TEXT"
	}
}

// ---------- shared helpers ----------

func quoteList(cols []string, q string) string {
	out := make([]string, len(cols))
	for i, c := range cols {
		out[i] = q + c + q
	}
	return strings.Join(out, ", ")
}

func formatDefault(v string) string {
	if v == "CURRENT_TIMESTAMP" || v == "NULL" || v == "TRUE" || v == "FALSE" {
		return v
	}
	return "'" + strings.ReplaceAll(v, "'", "''") + "'"
}

func splitStatements(sql string) []string {
	parts := strings.Split(sql, ";\n")
	var out []string
	for _, p := range parts {
		if s := strings.TrimSpace(p); s != "" {
			out = append(out, s)
		}
	}
	return out
}
