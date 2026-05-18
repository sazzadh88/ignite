package foundation

import (
	"github.com/sazzadh88/ignite/config"
)

// loadConfiguration maps the .env environment into the config repository,
// mirroring Laravel's default config/*.php files: each section is a single
// nested map. Application code reads configuration through app.Config() with
// dot notation, never directly from the environment.
func (app *Application) loadConfiguration() {
	c := app.config

	c.Set("app", map[string]any{
		"name":     config.Env("APP_NAME", "Ignite"),
		"env":      app.environment,
		"debug":    config.EnvBool("APP_DEBUG", true),
		"url":      config.Env("APP_URL", "http://localhost"),
		"port":     config.EnvInt("APP_PORT", 8080),
		"key":      config.Env("APP_KEY", ""),
		"timezone": config.Env("APP_TIMEZONE", "UTC"),
	})

	c.Set("database", map[string]any{
		"default": config.Env("DB_CONNECTION", "sqlite"),
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver":   "sqlite3",
				"database": config.Env("DB_DATABASE", "database/database.sqlite"),
			},
			"mysql": map[string]any{
				"driver":   "mysql",
				"host":     config.Env("DB_HOST", "127.0.0.1"),
				"port":     config.EnvInt("DB_PORT", 3306),
				"database": config.Env("DB_DATABASE", "ignite"),
				"username": config.Env("DB_USERNAME", "root"),
				"password": config.Env("DB_PASSWORD", ""),
			},
		},
	})

	c.Set("cache", map[string]any{
		"default": config.Env("CACHE_DRIVER", "file"),
		"prefix":  config.Env("CACHE_PREFIX", "ignite_cache"),
	})

	c.Set("session", map[string]any{
		"driver":   config.Env("SESSION_DRIVER", "file"),
		"lifetime": config.EnvInt("SESSION_LIFETIME", 120),
	})

	c.Set("queue", map[string]any{
		"default": config.Env("QUEUE_CONNECTION", "sync"),
	})

	c.Set("mail", map[string]any{
		"mailer":     config.Env("MAIL_MAILER", "smtp"),
		"host":       config.Env("MAIL_HOST", "127.0.0.1"),
		"port":       config.EnvInt("MAIL_PORT", 1025),
		"username":   config.Env("MAIL_USERNAME", ""),
		"password":   config.Env("MAIL_PASSWORD", ""),
		"encryption": config.Env("MAIL_ENCRYPTION", ""),
		"from": map[string]any{
			"address": config.Env("MAIL_FROM_ADDRESS", "hello@example.com"),
			"name":    config.Env("MAIL_FROM_NAME", config.Env("APP_NAME", "Ignite")),
		},
	})
}
