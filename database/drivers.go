package database

// Registered database drivers. Importing these for their side effects
// registers them with database/sql so sql.Open works for any connection
// the user selects via DB_CONNECTION (sqlite by default).
import (
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)
