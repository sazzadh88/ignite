package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Repository holds all configuration values with dot-notation access.
type Repository struct {
	mu    sync.RWMutex
	items map[string]any
}

// New creates a new config repository.
func New() *Repository {
	return &Repository{
		items: make(map[string]any),
	}
}

// LoadEnv loads a .env file into environment variables.
func LoadEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)

		// Expand ${VAR} references
		value = os.Expand(value, os.Getenv)

		// Only set if not already in environment (env overrides .env file)
		if os.Getenv(key) == "" {
			os.Setenv(key, value)
		}
	}
	return nil
}

// Set sets a config value using dot notation.
func (r *Repository) Set(key string, value any) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.set(key, value)
}

func (r *Repository) set(key string, value any) {
	parts := strings.Split(key, ".")
	if len(parts) == 1 {
		r.items[key] = value
		return
	}

	current := r.items
	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
			return
		}
		if next, ok := current[part]; ok {
			if m, ok := next.(map[string]any); ok {
				current = m
			} else {
				m := make(map[string]any)
				current[part] = m
				current = m
			}
		} else {
			m := make(map[string]any)
			current[part] = m
			current = m
		}
	}
}

// Get retrieves a config value using dot notation.
func (r *Repository) Get(key string) any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.get(key)
}

func (r *Repository) get(key string) any {
	parts := strings.Split(key, ".")
	var current any = r.items

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]any:
			val, ok := v[part]
			if !ok {
				return nil
			}
			current = val
		default:
			return nil
		}
	}
	return current
}

// GetString retrieves a string value with a default.
func (r *Repository) GetString(key string, defaultVal ...string) string {
	val := r.Get(key)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	default:
		return fmt.Sprintf("%v", v)
	}
}

// GetInt retrieves an int value with a default.
func (r *Repository) GetInt(key string, defaultVal ...int) int {
	val := r.Get(key)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return 0
		}
		return i
	default:
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
}

// GetBool retrieves a bool value with a default.
func (r *Repository) GetBool(key string, defaultVal ...bool) bool {
	val := r.Get(key)
	if val == nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	switch v := val.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1" || v == "yes"
	case int:
		return v != 0
	default:
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
}

// GetSlice retrieves a slice value.
func (r *Repository) GetSlice(key string) []any {
	val := r.Get(key)
	if val == nil {
		return nil
	}
	if s, ok := val.([]any); ok {
		return s
	}
	return nil
}

// Has checks if a key exists.
func (r *Repository) Has(key string) bool {
	return r.Get(key) != nil
}

// All returns all config items.
func (r *Repository) All() map[string]any {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.items
}

// Env retrieves an environment variable with optional default.
func Env(key string, defaultVal ...string) string {
	val := os.Getenv(key)
	if val == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return val
}

// EnvInt retrieves an environment variable as int.
func EnvInt(key string, defaultVal ...int) int {
	val := os.Getenv(key)
	if val == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return 0
	}
	return i
}

// EnvBool retrieves an environment variable as bool.
func EnvBool(key string, defaultVal ...bool) bool {
	val := os.Getenv(key)
	if val == "" {
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return false
	}
	return val == "true" || val == "1" || val == "yes"
}
