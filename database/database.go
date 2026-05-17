package database

import (
	"database/sql"
	"errors"
	"fmt"
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

	dsn, ok := connConfig["dsn"].(string)
	if !ok {
		return nil, fmt.Errorf("dsn not specified for connection '%s'", name)
	}

	// Open database connection
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open connection '%s': %w", name, err)
	}

	// Set connection pool settings if provided
	if maxOpen, ok := connConfig["max_open_conns"].(int); ok {
		db.SetMaxOpenConns(maxOpen)
	}

	if maxIdle, ok := connConfig["max_idle_conns"].(int); ok {
		db.SetMaxIdleConns(maxIdle)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping connection '%s': %w", name, err)
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
