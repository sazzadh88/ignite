package foundation

import (
	"os"
	"path/filepath"
	"testing"
)

// dbEnvKeys are cleared before each test so config.LoadEnv (which only sets a
// var when it is unset) actually populates from the temp .env file.
var dbEnvKeys = []string{
	"APP_NAME", "APP_ENV", "APP_DEBUG", "APP_URL", "APP_PORT", "APP_KEY", "APP_TIMEZONE",
	"DB_CONNECTION", "DB_HOST", "DB_PORT", "DB_DATABASE", "DB_USERNAME", "DB_PASSWORD",
	"CACHE_DRIVER", "SESSION_DRIVER", "SESSION_LIFETIME", "QUEUE_CONNECTION",
	"MAIL_MAILER", "MAIL_HOST", "MAIL_FROM_ADDRESS",
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range dbEnvKeys {
		os.Unsetenv(k)
	}
}

func bootstrapWithEnv(t *testing.T, env string) *Application {
	t.Helper()
	clearEnv(t)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(env), 0644); err != nil {
		t.Fatal(err)
	}
	app := NewApplication(dir)
	app.Bootstrap()
	return app
}

func TestConfigMapsEnvValues(t *testing.T) {
	app := bootstrapWithEnv(t, `APP_NAME=MyBlog
APP_ENV=production
APP_DEBUG=false
APP_PORT=9090
DB_CONNECTION=mysql
DB_HOST=db.internal
DB_PORT=3307
DB_DATABASE=blogdb
DB_USERNAME=admin
DB_PASSWORD=s3cret
CACHE_DRIVER=memory
SESSION_LIFETIME=240
QUEUE_CONNECTION=redis
MAIL_FROM_ADDRESS=team@blog.test
`)
	c := app.Config()

	cases := map[string]any{
		"app.name":                              "MyBlog",
		"app.env":                               "production",
		"app.debug":                             false,
		"app.port":                              9090,
		"database.default":                      "mysql",
		"database.connections.mysql.host":       "db.internal",
		"database.connections.mysql.port":       3307,
		"database.connections.mysql.database":   "blogdb",
		"database.connections.mysql.username":   "admin",
		"database.connections.sqlite.driver":    "sqlite3",
		"cache.default":                         "memory",
		"session.lifetime":                      240,
		"queue.default":                         "redis",
		"mail.from.address":                     "team@blog.test",
	}
	for key, want := range cases {
		got := c.Get(key)
		if got != want {
			t.Errorf("config %q = %#v, want %#v", key, got, want)
		}
	}

	// mysql password is read from the environment too (not in the asserted set above).
	if got := c.GetString("database.connections.mysql.password"); got != "s3cret" {
		t.Errorf("mysql password = %q, want s3cret", got)
	}

	if !app.IsProduction() {
		t.Error("environment should be production")
	}
}

func TestConfigDefaultsWhenEnvAbsent(t *testing.T) {
	app := bootstrapWithEnv(t, "APP_NAME=Bare\n")
	c := app.Config()

	if c.GetString("database.default") != "sqlite" {
		t.Errorf("database.default default = %q, want sqlite", c.GetString("database.default"))
	}
	if c.GetString("database.connections.sqlite.database") != "database/database.sqlite" {
		t.Errorf("sqlite path default wrong: %q", c.GetString("database.connections.sqlite.database"))
	}
	if c.GetString("cache.default") != "file" {
		t.Errorf("cache.default default = %q, want file", c.GetString("cache.default"))
	}
	if c.GetInt("app.port") != 8080 {
		t.Errorf("app.port default = %d, want 8080", c.GetInt("app.port"))
	}
	// MAIL_FROM_NAME falls back to the app name when unset.
	if c.GetString("mail.from.name") != "Bare" {
		t.Errorf("mail.from.name = %q, want Bare", c.GetString("mail.from.name"))
	}
}
