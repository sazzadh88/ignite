package database

import "database/sql"

// Query and Exec let *Connection structurally satisfy the schema package's
// Connection interface (schema.NewSchemaWithDriver), so migrations run
// through the same connection/transaction machinery as the rest of the app.

// Query runs a SELECT and returns rows as maps (alias of Select).
func (c *Connection) Query(query string, args ...any) ([]map[string]any, error) {
	return c.Select(query, args...)
}

// Exec runs a statement and returns the number of affected rows.
func (c *Connection) Exec(query string, args ...any) (int64, error) {
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

	affected, _ := result.RowsAffected()
	return affected, nil
}
