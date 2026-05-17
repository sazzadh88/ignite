// Package database provides a Laravel-inspired database layer for Ignite.
//
// The package offers connection management, query building, and transaction support
// using only the standard library's database/sql package with zero external dependencies
// (except for database drivers).
//
// # Basic Usage
//
// Create a database manager with configuration:
//
//	config := map[string]any{
//	    "default": "mysql",
//	    "connections": map[string]any{
//	        "mysql": map[string]any{
//	            "driver": "mysql",
//	            "dsn": "user:pass@tcp(localhost:3306)/dbname",
//	            "max_open_conns": 10,
//	            "max_idle_conns": 5,
//	        },
//	    },
//	}
//	manager := database.NewManager(config)
//	defer manager.Close()
//
// # Query Builder
//
// The query builder provides a fluent interface for building SQL queries:
//
//	conn, _ := manager.Default()
//
//	// SELECT queries
//	users, _ := conn.Table("users").
//	    Where("age", ">", 18).
//	    Where("active", "=", true).
//	    OrderBy("name", "ASC").
//	    Limit(10).
//	    Get()
//
//	// First row
//	user, _ := conn.Table("users").
//	    Where("email", "=", "john@example.com").
//	    First()
//
//	// Aggregates
//	count, _ := conn.Table("users").Count()
//	total, _ := conn.Table("orders").Sum("amount")
//	average, _ := conn.Table("orders").Avg("amount")
//
// # INSERT, UPDATE, DELETE
//
//	// Insert
//	id, _ := conn.Table("users").Insert(map[string]any{
//	    "name": "John Doe",
//	    "email": "john@example.com",
//	})
//
//	// Bulk insert
//	conn.Table("users").InsertMany([]map[string]any{
//	    {"name": "Alice", "email": "alice@example.com"},
//	    {"name": "Bob", "email": "bob@example.com"},
//	})
//
//	// Update
//	affected, _ := conn.Table("users").
//	    Where("id", "=", 1).
//	    Update(map[string]any{"age": 31})
//
//	// Delete
//	affected, _ := conn.Table("users").
//	    Where("active", "=", false).
//	    Delete()
//
//	// Increment/Decrement
//	conn.Table("posts").
//	    Where("id", "=", 1).
//	    Increment("views", 1)
//
// # Transactions
//
// Automatic transaction management with commit/rollback:
//
//	err := conn.Transaction(func(tx *database.Connection) error {
//	    _, err := tx.Table("users").Insert(map[string]any{
//	        "name": "Jane",
//	        "email": "jane@example.com",
//	    })
//	    if err != nil {
//	        return err // Will trigger rollback
//	    }
//
//	    _, err = tx.Table("profiles").Insert(map[string]any{
//	        "user_id": 1,
//	        "bio": "...",
//	    })
//	    return err // Success will commit
//	})
//
// Manual transaction control:
//
//	tx, _ := conn.BeginTransaction()
//	tx.Table("users").Insert(...)
//	tx.Commit() // or tx.Rollback()
//
// # Nested Transactions
//
// The package supports nested transactions using savepoints:
//
//	conn.Transaction(func(tx1 *database.Connection) error {
//	    tx1.Table("users").Insert(...)
//
//	    return tx1.Transaction(func(tx2 *database.Connection) error {
//	        // Nested transaction uses savepoint
//	        return tx2.Table("profiles").Insert(...)
//	    })
//	})
//
// # Query Events
//
// Listen to query execution events:
//
//	manager.Listen(func(event *database.QueryEvent) {
//	    log.Printf("[%s] %s (%v)", event.Connection, event.SQL, event.Time)
//	})
//
// Pre-query hooks:
//
//	manager.BeforeExecuting(func(sql string) {
//	    log.Printf("About to execute: %s", sql)
//	})
//
// # Advanced WHERE Clauses
//
//	// Multiple conditions
//	conn.Table("users").
//	    Where("age", ">", 18).
//	    OrWhere("role", "=", "admin")
//
//	// IN clause
//	conn.Table("users").
//	    WhereIn("status", []any{"active", "pending"})
//
//	// NULL checks
//	conn.Table("users").
//	    WhereNull("deleted_at")
//
//	// BETWEEN
//	conn.Table("users").
//	    WhereBetween("age", 18, 65)
//
//	// EXISTS subquery
//	conn.Table("users").WhereExists(func(qb *database.QueryBuilder) {
//	    qb.Table("orders").
//	        Where("orders.user_id", "=", "users.id")
//	})
//
// # JOINs
//
//	conn.Table("users").
//	    Select("users.name", "orders.total").
//	    LeftJoin("orders", "users.id", "=", "orders.user_id").
//	    Get()
//
// # Pagination
//
//	paginator, _ := conn.Table("users").
//	    OrderBy("created_at", "DESC").
//	    Paginate(15, 1) // 15 per page, page 1
//
//	fmt.Printf("Page %d of %d\n", paginator.CurrentPage, paginator.LastPage)
//	fmt.Printf("Total: %d\n", paginator.Total)
//
//	for _, item := range paginator.Items {
//	    // Process items
//	}
//
// # Chunking
//
// Process large result sets in chunks:
//
//	conn.Table("users").Chunk(100, func(users []map[string]any) bool {
//	    for _, user := range users {
//	        // Process user
//	    }
//	    return true // Continue to next chunk
//	})
//
// # Raw Queries
//
// Execute raw SQL when needed:
//
//	rows, _ := conn.Select("SELECT * FROM users WHERE age > ?", 18)
//	row, _ := conn.SelectOne("SELECT * FROM users WHERE id = ?", 1)
//	affected, _ := conn.Update("UPDATE users SET active = ? WHERE id = ?", true, 1)
//	affected, _ := conn.Delete("DELETE FROM users WHERE id = ?", 1)
//	conn.Statement("CREATE INDEX idx_email ON users(email)")
//
// # SQL Generation
//
// Get the SQL without executing:
//
//	sql, args := conn.Table("users").
//	    Where("age", ">", 18).
//	    OrderBy("name", "ASC").
//	    ToSQL()
//	fmt.Printf("SQL: %s\nArgs: %v\n", sql, args)
package database
