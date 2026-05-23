package database

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

// mysqlStrictSQLMode is the strict sql_mode Laravel applies when
// 'strict' => true: it makes MySQL reject invalid/truncated data instead
// of silently coercing it.
const mysqlStrictSQLMode = "ONLY_FULL_GROUP_BY,STRICT_TRANS_TABLES," +
	"NO_ZERO_IN_DATE,NO_ZERO_DATE,ERROR_FOR_DIVISION_BY_ZERO,NO_ENGINE_SUBSTITUTION"

// mysqlTLS maps the unified sslmode to go-sql-driver's tls option.
// "prefer" (default) uses TLS when the server supports it and falls back to
// plaintext otherwise — most security that still connects.
func mysqlTLS(sslmode string) string {
	switch sslmode {
	case "", "prefer":
		return "preferred"
	case "require":
		return "skip-verify"
	case "verify-ca", "verify-full":
		return "true"
	case "disable":
		return "false"
	default:
		return "preferred"
	}
}

// buildDSN assembles a driver-specific DSN from the connection's config
// fields, mirroring Laravel's per-driver connectors. Config no longer stores
// a precomputed dsn; it is derived here from host/port/database/etc.
func buildDSN(driver string, c map[string]any) (string, error) {
	switch driver {
	case "sqlite3", "sqlite":
		path := cfgStr(c, "database")
		if path == "" {
			return "", errors.New("sqlite: database path not configured")
		}
		return path + "?_foreign_keys=on", nil

	case "mysql":
		params := "parseTime=true&charset=utf8mb4&loc=Local"
		if cfgBool(c, "strict", true) {
			params += "&sql_mode=" + url.QueryEscape("'"+mysqlStrictSQLMode+"'")
		}
		// A configured unix socket takes precedence over host:port — some
		// MySQL servers (e.g. socket-only installs) don't listen on TCP.
		// TLS is not applied over a local unix socket.
		if socket := cfgStr(c, "unix_socket"); socket != "" {
			return fmt.Sprintf("%s:%s@unix(%s)/%s?%s",
				cfgStr(c, "username"), cfgStr(c, "password"),
				socket, cfgStr(c, "database"), params), nil
		}
		params += "&tls=" + mysqlTLS(cfgStr(c, "sslmode"))
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			cfgStr(c, "username"), cfgStr(c, "password"),
			cfgStr(c, "host"), cfgInt(c, "port", 3306), cfgStr(c, "database"), params,
		), nil

	case "postgres", "pgsql":
		sslmode := cfgStr(c, "sslmode")
		if sslmode == "" {
			sslmode = "disable"
		}
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfgStr(c, "host"), cfgInt(c, "port", 5432),
			cfgStr(c, "username"), cfgStr(c, "password"),
			cfgStr(c, "database"), sslmode,
		)
		if sp := cfgStr(c, "search_path"); sp != "" {
			dsn += " search_path=" + sp
		}
		if cs := cfgStr(c, "charset"); cs != "" {
			dsn += " client_encoding=" + cs
		}
		return dsn, nil

	default:
		return "", fmt.Errorf("unsupported database driver %q", driver)
	}
}

// postgresSSLAttempts returns the ordered sslmode values to try. lib/pq does
// not implement libpq's "prefer"/"allow", so we emulate them: "prefer"
// (the secure default) tries an encrypted connection first and only falls
// back to plaintext if the server has no SSL. An explicit mode is honored
// as-is (require/verify-ca/verify-full/disable).
func postgresSSLAttempts(c map[string]any) []string {
	switch cfgStr(c, "sslmode") {
	case "", "prefer":
		return []string{"require", "disable"}
	case "allow":
		return []string{"disable", "require"}
	default:
		return []string{cfgStr(c, "sslmode")}
	}
}

// cloneConfig makes a shallow copy so per-attempt sslmode changes don't
// mutate the shared connection config.
func cloneConfig(c map[string]any) map[string]any {
	out := make(map[string]any, len(c))
	for k, v := range c {
		out[k] = v
	}
	return out
}

func cfgBool(c map[string]any, key string, def bool) bool {
	v, ok := c[key]
	if !ok {
		return def
	}
	switch b := v.(type) {
	case bool:
		return b
	case string:
		return b == "true" || b == "1" || b == "yes"
	}
	return def
}

func cfgStr(c map[string]any, key string) string {
	v, ok := c[key]
	if !ok {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

func cfgInt(c map[string]any, key string, def int) int {
	v, ok := c[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	case string:
		if i, err := strconv.Atoi(n); err == nil {
			return i
		}
	}
	return def
}
