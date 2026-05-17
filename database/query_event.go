package database

import "time"

// QueryEvent represents a database query execution event.
type QueryEvent struct {
	// SQL is the raw SQL query string
	SQL string
	// Bindings are the parameter values bound to the query
	Bindings []any
	// Time is the duration the query took to execute
	Time time.Duration
	// Connection is the name of the connection that executed the query
	Connection string
}
