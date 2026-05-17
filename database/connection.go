package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Connection wraps a database/sql connection with additional features.
type Connection struct {
	db           *sql.DB
	name         string
	tx           *sql.Tx
	savepoints   []string
	listeners    []func(*QueryEvent)
	beforeHooks  []func(string)
	inTransaction bool
}

// newConnection creates a new connection wrapper.
func newConnection(db *sql.DB, name string) *Connection {
	return &Connection{
		db:        db,
		name:      name,
		listeners: make([]func(*QueryEvent), 0),
		beforeHooks: make([]func(string), 0),
		savepoints: make([]string, 0),
	}
}

// Table creates a new query builder for the specified table.
func (c *Connection) Table(name string) *QueryBuilder {
	return newQueryBuilder(c, name)
}

// GetDB returns the underlying *sql.DB.
func (c *Connection) GetDB() *sql.DB {
	return c.db
}

// Select executes a SELECT query and returns all rows as maps.
func (c *Connection) Select(query string, args ...any) ([]map[string]any, error) {
	start := time.Now()

	c.fireBeforeHooks(query)

	var rows *sql.Rows
	var err error

	if c.inTransaction && c.tx != nil {
		rows, err = c.tx.Query(query, args...)
	} else {
		rows, err = c.db.Query(query, args...)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]any, 0)

	for rows.Next() {
		values := make([]any, len(cols))
		valuePtrs := make([]any, len(cols))
		for i := range cols {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}

		row := make(map[string]any)
		for i, col := range cols {
			val := values[i]

			// Convert []byte to string for readability
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}

		results = append(results, row)
	}

	c.fireListeners(&QueryEvent{
		SQL:        query,
		Bindings:   args,
		Time:       time.Since(start),
		Connection: c.name,
	})

	return results, rows.Err()
}

// SelectOne executes a SELECT query and returns the first row.
func (c *Connection) SelectOne(query string, args ...any) (map[string]any, error) {
	results, err := c.Select(query, args...)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	return results[0], nil
}

// Insert executes an INSERT query.
func (c *Connection) Insert(query string, args ...any) (sql.Result, error) {
	start := time.Now()

	c.fireBeforeHooks(query)

	var result sql.Result
	var err error

	if c.inTransaction && c.tx != nil {
		result, err = c.tx.Exec(query, args...)
	} else {
		result, err = c.db.Exec(query, args...)
	}

	c.fireListeners(&QueryEvent{
		SQL:        query,
		Bindings:   args,
		Time:       time.Since(start),
		Connection: c.name,
	})

	return result, err
}

// Update executes an UPDATE query and returns affected rows count.
func (c *Connection) Update(query string, args ...any) (int64, error) {
	start := time.Now()

	c.fireBeforeHooks(query)

	var result sql.Result
	var err error

	if c.inTransaction && c.tx != nil {
		result, err = c.tx.Exec(query, args...)
	} else {
		result, err = c.db.Exec(query, args...)
	}

	if err != nil {
		return 0, err
	}

	c.fireListeners(&QueryEvent{
		SQL:        query,
		Bindings:   args,
		Time:       time.Since(start),
		Connection: c.name,
	})

	return result.RowsAffected()
}

// Delete executes a DELETE query and returns affected rows count.
func (c *Connection) Delete(query string, args ...any) (int64, error) {
	start := time.Now()

	c.fireBeforeHooks(query)

	var result sql.Result
	var err error

	if c.inTransaction && c.tx != nil {
		result, err = c.tx.Exec(query, args...)
	} else {
		result, err = c.db.Exec(query, args...)
	}

	if err != nil {
		return 0, err
	}

	c.fireListeners(&QueryEvent{
		SQL:        query,
		Bindings:   args,
		Time:       time.Since(start),
		Connection: c.name,
	})

	return result.RowsAffected()
}

// Statement executes an arbitrary SQL statement.
func (c *Connection) Statement(query string, args ...any) error {
	start := time.Now()

	c.fireBeforeHooks(query)

	var err error

	if c.inTransaction && c.tx != nil {
		_, err = c.tx.Exec(query, args...)
	} else {
		_, err = c.db.Exec(query, args...)
	}

	c.fireListeners(&QueryEvent{
		SQL:        query,
		Bindings:   args,
		Time:       time.Since(start),
		Connection: c.name,
	})

	return err
}

// Transaction runs a callback within a transaction, with automatic commit/rollback.
// Supports retry attempts on failure.
func (c *Connection) Transaction(fn func(*Connection) error, attempts ...int) error {
	maxAttempts := 1
	if len(attempts) > 0 && attempts[0] > 0 {
		maxAttempts = attempts[0]
	}

	var lastErr error

	for attempt := 0; attempt < maxAttempts; attempt++ {
		txConn, err := c.BeginTransaction()
		if err != nil {
			lastErr = err
			continue
		}

		err = fn(txConn)

		if err != nil {
			if rbErr := txConn.Rollback(); rbErr != nil {
				lastErr = fmt.Errorf("rollback failed: %v (original error: %v)", rbErr, err)
			} else {
				lastErr = err
			}
			continue
		}

		if err := txConn.Commit(); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return lastErr
}

// BeginTransaction starts a new transaction.
func (c *Connection) BeginTransaction() (*Connection, error) {
	if c.inTransaction {
		// Nested transaction - use savepoint
		savepoint := fmt.Sprintf("sp%d", len(c.savepoints)+1)
		if err := c.Statement(fmt.Sprintf("SAVEPOINT %s", savepoint)); err != nil {
			return nil, err
		}
		c.savepoints = append(c.savepoints, savepoint)
		return c, nil
	}

	tx, err := c.db.Begin()
	if err != nil {
		return nil, err
	}

	return &Connection{
		db:           c.db,
		name:         c.name,
		tx:           tx,
		inTransaction: true,
		listeners:    c.listeners,
		beforeHooks:  c.beforeHooks,
		savepoints:   make([]string, 0),
	}, nil
}

// Commit commits the current transaction.
func (c *Connection) Commit() error {
	if !c.inTransaction {
		return errors.New("no transaction in progress")
	}

	if len(c.savepoints) > 0 {
		// Release savepoint for nested transaction
		savepoint := c.savepoints[len(c.savepoints)-1]
		c.savepoints = c.savepoints[:len(c.savepoints)-1]
		return c.Statement(fmt.Sprintf("RELEASE SAVEPOINT %s", savepoint))
	}

	if c.tx == nil {
		return errors.New("no transaction handle")
	}

	err := c.tx.Commit()
	if err == nil {
		c.inTransaction = false
		c.tx = nil
	}
	return err
}

// Rollback rolls back the current transaction.
func (c *Connection) Rollback() error {
	if !c.inTransaction {
		return errors.New("no transaction in progress")
	}

	if len(c.savepoints) > 0 {
		// Rollback to savepoint for nested transaction
		savepoint := c.savepoints[len(c.savepoints)-1]
		c.savepoints = c.savepoints[:len(c.savepoints)-1]
		return c.Statement(fmt.Sprintf("ROLLBACK TO SAVEPOINT %s", savepoint))
	}

	if c.tx == nil {
		return errors.New("no transaction handle")
	}

	err := c.tx.Rollback()
	if err == nil {
		c.inTransaction = false
		c.tx = nil
	}
	return err
}

// Listen registers a query event listener.
func (c *Connection) Listen(callback func(*QueryEvent)) {
	c.listeners = append(c.listeners, callback)
}

// BeforeExecuting registers a pre-query hook.
func (c *Connection) BeforeExecuting(callback func(string)) {
	c.beforeHooks = append(c.beforeHooks, callback)
}

func (c *Connection) fireListeners(event *QueryEvent) {
	for _, listener := range c.listeners {
		listener(event)
	}
}

func (c *Connection) fireBeforeHooks(query string) {
	for _, hook := range c.beforeHooks {
		hook(query)
	}
}
