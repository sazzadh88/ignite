package database

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
)

// Manager manages database connections.
type Manager struct {
	mu          sync.RWMutex
	connections map[string]*Connection
	config      map[string]any
	listeners   []func(*QueryEvent)
	beforeHooks []func(string)
	defaultConn string
}

// NewManager creates a new database manager.
func NewManager(config map[string]any) *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		config:      config,
		listeners:   make([]func(*QueryEvent), 0),
		beforeHooks: make([]func(string), 0),
	}
}

// Connection gets or creates a named connection.
// Config structure expected:
//   map[string]any{
//     "default": "mysql",
//     "connections": map[string]any{
//       "mysql": map[string]any{
//         "driver": "mysql",
//         "dsn": "user:pass@tcp(localhost:3306)/dbname",
//       },
//     },
//   }
func (m *Manager) Connection(name string) (*Connection, error) {
	m.mu.RLock()
	conn, exists := m.connections[name]
	m.mu.RUnlock()

	if exists {
		return conn, nil
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if conn, exists := m.connections[name]; exists {
		return conn, nil
	}

	// Get connection config
	connectionsMap, ok := m.config["connections"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid connections config")
	}

	connConfig, ok := connectionsMap[name].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("connection '%s' not configured", name)
	}

	driver, ok := connConfig["driver"].(string)
	if !ok {
		return nil, fmt.Errorf("driver not specified for connection '%s'", name)
	}

	// Resolve the list of DSNs to try. A precomputed dsn wins. For Postgres
	// we emulate libpq's "prefer": attempt an encrypted connection first and
	// fall back to plaintext only if the server has no SSL — most security
	// that still establishes a connection.
	var dsns []string
	if pre, _ := connConfig["dsn"].(string); pre != "" {
		dsns = []string{pre}
	} else if driver == "postgres" || driver == "pgsql" {
		for _, mode := range postgresSSLAttempts(connConfig) {
			cc := cloneConfig(connConfig)
			cc["sslmode"] = mode
			built, err := buildDSN(driver, cc)
			if err != nil {
				return nil, fmt.Errorf("connection '%s': %w", name, err)
			}
			dsns = append(dsns, built)
		}
	} else {
		built, err := buildDSN(driver, connConfig)
		if err != nil {
			return nil, fmt.Errorf("connection '%s': %w", name, err)
		}
		dsns = []string{built}
	}

	var db *sql.DB
	var lastErr error
	for i, dsn := range dsns {
		db, lastErr = sql.Open(driver, dsn)
		if lastErr == nil {
			if maxOpen, ok := connConfig["max_open_conns"].(int); ok {
				db.SetMaxOpenConns(maxOpen)
			}
			if maxIdle, ok := connConfig["max_idle_conns"].(int); ok {
				db.SetMaxIdleConns(maxIdle)
			}
			lastErr = db.Ping()
			if lastErr == nil {
				break
			}
			db.Close()
		}
		// Only fall back to the next (less secure) DSN when the failure is
		// specifically the server lacking SSL — never mask auth/other errors.
		if i < len(dsns)-1 && strings.Contains(lastErr.Error(), "SSL is not enabled") {
			continue
		}
		return nil, fmt.Errorf("failed to connect '%s': %w", name, lastErr)
	}

	conn = newConnection(db, name)

	// Register global listeners and hooks
	for _, listener := range m.listeners {
		conn.Listen(listener)
	}

	for _, hook := range m.beforeHooks {
		conn.BeforeExecuting(hook)
	}

	m.connections[name] = conn

	return conn, nil
}

// Default returns the default connection.
func (m *Manager) Default() (*Connection, error) {
	defaultName := "default"

	if name, ok := m.config["default"].(string); ok {
		defaultName = name
	}

	if m.defaultConn == "" {
		m.defaultConn = defaultName
	}

	return m.Connection(m.defaultConn)
}

// SetDefaultConnection sets the default connection name.
func (m *Manager) SetDefaultConnection(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultConn = name
}

// Listen registers a global query event listener.
// This listener will be attached to all current and future connections.
func (m *Manager) Listen(callback func(*QueryEvent)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.listeners = append(m.listeners, callback)

	// Register with existing connections
	for _, conn := range m.connections {
		conn.Listen(callback)
	}
}

// BeforeExecuting registers a global pre-query hook.
// This hook will be attached to all current and future connections.
func (m *Manager) BeforeExecuting(callback func(string)) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.beforeHooks = append(m.beforeHooks, callback)

	// Register with existing connections
	for _, conn := range m.connections {
		conn.BeforeExecuting(callback)
	}
}

// Close closes all database connections.
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var lastErr error

	for name, conn := range m.connections {
		if err := conn.db.Close(); err != nil {
			lastErr = fmt.Errorf("failed to close connection '%s': %w", name, err)
		}
	}

	m.connections = make(map[string]*Connection)

	return lastErr
}

// HasConnection checks if a connection exists.
func (m *Manager) HasConnection(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	_, exists := m.connections[name]
	return exists
}

// DisconnectConnection closes a specific connection.
func (m *Manager) DisconnectConnection(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[name]
	if !exists {
		return fmt.Errorf("connection '%s' does not exist", name)
	}

	if err := conn.db.Close(); err != nil {
		return err
	}

	delete(m.connections, name)

	return nil
}

// ReconnectConnection closes and reopens a specific connection.
func (m *Manager) ReconnectConnection(name string) error {
	if err := m.DisconnectConnection(name); err != nil {
		// If connection doesn't exist, that's fine
		if !m.HasConnection(name) {
			_, err := m.Connection(name)
			return err
		}
		return err
	}

	_, err := m.Connection(name)
	return err
}
