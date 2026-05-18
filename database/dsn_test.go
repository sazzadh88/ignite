package database

import (
	"strings"
	"testing"
)

func TestBuildDSNSQLite(t *testing.T) {
	dsn, err := buildDSN("sqlite3", map[string]any{"database": "database/database.sqlite"})
	if err != nil {
		t.Fatal(err)
	}
	if dsn != "database/database.sqlite?_foreign_keys=on" {
		t.Errorf("sqlite dsn = %q", dsn)
	}
	if _, err := buildDSN("sqlite3", map[string]any{}); err == nil {
		t.Error("missing sqlite path should error")
	}
}

func TestBuildDSNMySQLTCP(t *testing.T) {
	dsn, err := buildDSN("mysql", map[string]any{
		"host": "db.host", "port": 3307, "database": "blog",
		"username": "admin", "password": "s3cret",
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, frag := range []string{
		"admin:s3cret@tcp(db.host:3307)/blog?",
		"parseTime=true", "charset=utf8mb4", "loc=Local",
		"tls=preferred",                   // default sslmode "prefer"
		"sql_mode=%27ONLY_FULL_GROUP_BY",  // strict default true (url-escaped)
	} {
		if !strings.Contains(dsn, frag) {
			t.Errorf("mysql tcp dsn missing %q\n  got: %s", frag, dsn)
		}
	}
}

func TestBuildDSNMySQLSocketStrictOffNoTLS(t *testing.T) {
	dsn, err := buildDSN("mysql", map[string]any{
		"database": "blog", "username": "root", "password": "password",
		"unix_socket": "/tmp/mysql.sock", "strict": false,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(dsn, "root:password@unix(/tmp/mysql.sock)/blog?") {
		t.Errorf("socket dsn wrong: %s", dsn)
	}
	if strings.Contains(dsn, "tls=") {
		t.Errorf("unix socket must not set tls: %s", dsn)
	}
	if strings.Contains(dsn, "sql_mode=") {
		t.Errorf("strict=false must not set sql_mode: %s", dsn)
	}
}

func TestMySQLTLSMapping(t *testing.T) {
	cases := map[string]string{
		"": "preferred", "prefer": "preferred", "require": "skip-verify",
		"verify-ca": "true", "verify-full": "true", "disable": "false",
	}
	for mode, want := range cases {
		if got := mysqlTLS(mode); got != want {
			t.Errorf("mysqlTLS(%q) = %q, want %q", mode, got, want)
		}
	}
}

func TestBuildDSNPostgres(t *testing.T) {
	dsn, err := buildDSN("postgres", map[string]any{
		"host": "pg.host", "port": 5433, "database": "blog",
		"username": "admin", "password": "s3cret", "sslmode": "require",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "host=pg.host port=5433 user=admin password=s3cret dbname=blog sslmode=require"
	if dsn != want {
		t.Errorf("postgres dsn = %q, want %q", dsn, want)
	}

	// sslmode defaults to disable when absent.
	dsn, _ = buildDSN("postgres", map[string]any{
		"host": "h", "port": 5432, "database": "d", "username": "u", "password": "",
	})
	if want := "host=h port=5432 user=u password= dbname=d sslmode=disable"; dsn != want {
		t.Errorf("postgres default sslmode dsn = %q, want %q", dsn, want)
	}
}

func TestBuildDSNPostgresWithSearchPathAndCharset(t *testing.T) {
	dsn, err := buildDSN("postgres", map[string]any{
		"host": "127.0.0.1", "port": 5432, "database": "blog",
		"username": "leggit", "password": "pw", "sslmode": "disable",
		"search_path": "public", "charset": "utf8",
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "host=127.0.0.1 port=5432 user=leggit password=pw dbname=blog sslmode=disable search_path=public client_encoding=utf8"
	if dsn != want {
		t.Errorf("postgres dsn = %q, want %q", dsn, want)
	}
}

func TestPostgresSSLAttempts(t *testing.T) {
	cases := []struct {
		mode string
		want []string
	}{
		{"", []string{"require", "disable"}},        // default = prefer
		{"prefer", []string{"require", "disable"}},   // try SSL, fall back
		{"allow", []string{"disable", "require"}},    // try plaintext, then SSL
		{"require", []string{"require"}},             // explicit, no fallback
		{"verify-full", []string{"verify-full"}},     // explicit, no fallback
		{"disable", []string{"disable"}},             // explicit plaintext
	}
	for _, c := range cases {
		got := postgresSSLAttempts(map[string]any{"sslmode": c.mode})
		if len(got) != len(c.want) {
			t.Errorf("sslmode %q: got %v, want %v", c.mode, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("sslmode %q: got %v, want %v", c.mode, got, c.want)
				break
			}
		}
	}
}

func TestBuildDSNUnsupported(t *testing.T) {
	if _, err := buildDSN("oracle", map[string]any{}); err == nil {
		t.Error("unsupported driver should error")
	}
}

func TestCfgIntCoercions(t *testing.T) {
	c := map[string]any{"a": 5, "b": int64(6), "c": 7.0, "d": "8", "e": "x"}
	if cfgInt(c, "a", 0) != 5 || cfgInt(c, "b", 0) != 6 || cfgInt(c, "c", 0) != 7 || cfgInt(c, "d", 0) != 8 {
		t.Error("cfgInt coercion failed")
	}
	if cfgInt(c, "e", 99) != 99 || cfgInt(c, "missing", 42) != 42 {
		t.Error("cfgInt fallback failed")
	}
}
