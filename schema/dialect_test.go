package schema

import (
	"strings"
	"testing"
)

func sampleBlueprint() *Blueprint {
	b := NewBlueprint()
	b.ID()
	b.String("title")
	b.Text("body")
	b.Timestamps()
	b.Index([]string{"title"})
	return b
}

func TestSQLiteCreateGrammar(t *testing.T) {
	stmts := sqliteDialect{}.CompileCreate("posts", sampleBlueprint())
	create := stmts[0]

	if !strings.Contains(create, `CREATE TABLE "posts"`) {
		t.Errorf("missing quoted table: %s", create)
	}
	if !strings.Contains(create, `"id" INTEGER PRIMARY KEY AUTOINCREMENT`) {
		t.Errorf("sqlite id should be INTEGER PK AUTOINCREMENT: %s", create)
	}
	if strings.Contains(create, "ENGINE=") || strings.Contains(create, "AUTO_INCREMENT") || strings.Contains(create, "`") {
		t.Errorf("sqlite must not emit MySQL syntax: %s", create)
	}
	// Index is a separate statement, not inline.
	if len(stmts) < 2 || !strings.Contains(stmts[1], `CREATE INDEX IF NOT EXISTS "idx_title" ON "posts"`) {
		t.Errorf("expected separate CREATE INDEX, got %v", stmts)
	}
}

func TestPostgresCreateGrammar(t *testing.T) {
	stmts := postgresDialect{}.CompileCreate("posts", sampleBlueprint())
	create := stmts[0]

	if !strings.Contains(create, `"id" BIGSERIAL PRIMARY KEY`) {
		t.Errorf("postgres id should be BIGSERIAL PK: %s", create)
	}
	if strings.Contains(create, "ENGINE=") || strings.Contains(create, "`") || strings.Contains(create, "AUTO_INCREMENT") {
		t.Errorf("postgres must not emit MySQL syntax: %s", create)
	}
}

func TestMySQLGrammarUnchanged(t *testing.T) {
	stmts := mysqlDialect{}.CompileCreate("posts", sampleBlueprint())
	if len(stmts) != 1 || !strings.Contains(stmts[0], "ENGINE=InnoDB") || !strings.Contains(stmts[0], "AUTO_INCREMENT") {
		t.Errorf("mysql grammar should keep original behaviour: %v", stmts)
	}
}

func TestDialectFor(t *testing.T) {
	if DialectFor("sqlite3").Name() != "sqlite" {
		t.Error("sqlite3 -> sqlite")
	}
	if DialectFor("postgres").Name() != "postgres" {
		t.Error("postgres -> postgres")
	}
	if DialectFor("mysql").Name() != "mysql" {
		t.Error("mysql -> mysql")
	}
	if DialectFor("unknown").Name() != "mysql" {
		t.Error("unknown should fall back to mysql")
	}
}

func TestTableExistsPlaceholders(t *testing.T) {
	if q, _ := (sqliteDialect{}).CompileTableExists("t"); !strings.Contains(q, "sqlite_master") {
		t.Errorf("sqlite table-exists should use sqlite_master: %s", q)
	}
	if q, _ := (postgresDialect{}).CompileTableExists("t"); !strings.Contains(q, "pg_tables") || !strings.Contains(q, "$1") {
		t.Errorf("postgres table-exists should use pg_tables and $1: %s", q)
	}
}
