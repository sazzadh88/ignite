package database

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// TestQueryBuilder_ToSQL tests SQL generation without actual DB execution
func TestQueryBuilder_ToSQL(t *testing.T) {
	// Create a mock connection (just for query building, won't execute)
	conn := &Connection{name: "test"}

	tests := []struct {
		name     string
		build    func(*QueryBuilder)
		wantSQL  string
		wantArgs []any
	}{
		{
			name: "simple select",
			build: func(qb *QueryBuilder) {
				qb.Select("id", "name")
			},
			wantSQL:  "SELECT id, name FROM users",
			wantArgs: []any{},
		},
		{
			name: "select all",
			build: func(qb *QueryBuilder) {},
			wantSQL:  "SELECT * FROM users",
			wantArgs: []any{},
		},
		{
			name: "where clause",
			build: func(qb *QueryBuilder) {
				qb.Where("age", ">", 18)
			},
			wantSQL:  "SELECT * FROM users WHERE age > ?",
			wantArgs: []any{18},
		},
		{
			name: "multiple where clauses",
			build: func(qb *QueryBuilder) {
				qb.Where("age", ">", 18).Where("active", "=", true)
			},
			wantSQL:  "SELECT * FROM users WHERE age > ? AND active = ?",
			wantArgs: []any{18, true},
		},
		{
			name: "or where",
			build: func(qb *QueryBuilder) {
				qb.Where("age", ">", 18).OrWhere("role", "=", "admin")
			},
			wantSQL:  "SELECT * FROM users WHERE age > ? OR role = ?",
			wantArgs: []any{18, "admin"},
		},
		{
			name: "where in",
			build: func(qb *QueryBuilder) {
				qb.WhereIn("status", []any{"active", "pending"})
			},
			wantSQL:  "SELECT * FROM users WHERE status IN (?, ?)",
			wantArgs: []any{"active", "pending"},
		},
		{
			name: "where not in",
			build: func(qb *QueryBuilder) {
				qb.WhereNotIn("role", []any{"guest", "banned"})
			},
			wantSQL:  "SELECT * FROM users WHERE role NOT IN (?, ?)",
			wantArgs: []any{"guest", "banned"},
		},
		{
			name: "where null",
			build: func(qb *QueryBuilder) {
				qb.WhereNull("deleted_at")
			},
			wantSQL:  "SELECT * FROM users WHERE deleted_at IS NULL",
			wantArgs: []any{},
		},
		{
			name: "where not null",
			build: func(qb *QueryBuilder) {
				qb.WhereNotNull("email")
			},
			wantSQL:  "SELECT * FROM users WHERE email IS NOT NULL",
			wantArgs: []any{},
		},
		{
			name: "where between",
			build: func(qb *QueryBuilder) {
				qb.WhereBetween("age", 18, 65)
			},
			wantSQL:  "SELECT * FROM users WHERE age BETWEEN ? AND ?",
			wantArgs: []any{18, 65},
		},
		{
			name: "inner join",
			build: func(qb *QueryBuilder) {
				qb.Join("orders", "users.id", "=", "orders.user_id")
			},
			wantSQL:  "SELECT * FROM users INNER JOIN orders ON users.id = orders.user_id",
			wantArgs: []any{},
		},
		{
			name: "left join",
			build: func(qb *QueryBuilder) {
				qb.LeftJoin("profiles", "users.id", "=", "profiles.user_id")
			},
			wantSQL:  "SELECT * FROM users LEFT JOIN profiles ON users.id = profiles.user_id",
			wantArgs: []any{},
		},
		{
			name: "cross join",
			build: func(qb *QueryBuilder) {
				qb.CrossJoin("roles")
			},
			wantSQL:  "SELECT * FROM users CROSS JOIN roles",
			wantArgs: []any{},
		},
		{
			name: "group by",
			build: func(qb *QueryBuilder) {
				qb.Select("role", "COUNT(*) as count").GroupBy("role")
			},
			wantSQL:  "SELECT role, COUNT(*) as count FROM users GROUP BY role",
			wantArgs: []any{},
		},
		{
			name: "having",
			build: func(qb *QueryBuilder) {
				qb.Select("role", "COUNT(*) as count").
					GroupBy("role").
					Having("count", ">", 5)
			},
			wantSQL:  "SELECT role, COUNT(*) as count FROM users GROUP BY role HAVING count > ?",
			wantArgs: []any{5},
		},
		{
			name: "order by",
			build: func(qb *QueryBuilder) {
				qb.OrderBy("name", "ASC")
			},
			wantSQL:  "SELECT * FROM users ORDER BY name ASC",
			wantArgs: []any{},
		},
		{
			name: "order by desc",
			build: func(qb *QueryBuilder) {
				qb.OrderByDesc("created_at")
			},
			wantSQL:  "SELECT * FROM users ORDER BY created_at DESC",
			wantArgs: []any{},
		},
		{
			name: "order by raw",
			build: func(qb *QueryBuilder) {
				qb.OrderByRaw("RAND()")
			},
			wantSQL:  "SELECT * FROM users ORDER BY RAND()",
			wantArgs: []any{},
		},
		{
			name: "limit",
			build: func(qb *QueryBuilder) {
				qb.Limit(10)
			},
			wantSQL:  "SELECT * FROM users LIMIT 10",
			wantArgs: []any{},
		},
		{
			name: "limit and offset",
			build: func(qb *QueryBuilder) {
				qb.Limit(10).Offset(20)
			},
			wantSQL:  "SELECT * FROM users LIMIT 10 OFFSET 20",
			wantArgs: []any{},
		},
		{
			name: "lock for update",
			build: func(qb *QueryBuilder) {
				qb.Where("id", "=", 1).LockForUpdate()
			},
			wantSQL:  "SELECT * FROM users WHERE id = ? FOR UPDATE",
			wantArgs: []any{1},
		},
		{
			name: "shared lock",
			build: func(qb *QueryBuilder) {
				qb.Where("id", "=", 1).SharedLock()
			},
			wantSQL:  "SELECT * FROM users WHERE id = ? LOCK IN SHARE MODE",
			wantArgs: []any{1},
		},
		{
			name: "complex query",
			build: func(qb *QueryBuilder) {
				qb.Select("users.id", "users.name", "COUNT(orders.id) as order_count").
					LeftJoin("orders", "users.id", "=", "orders.user_id").
					Where("users.active", "=", true).
					WhereNotNull("users.email").
					GroupBy("users.id", "users.name").
					Having("order_count", ">", 5).
					OrderBy("order_count", "DESC").
					Limit(10)
			},
			wantSQL:  "SELECT users.id, users.name, COUNT(orders.id) as order_count FROM users LEFT JOIN orders ON users.id = orders.user_id WHERE users.active = ? AND users.email IS NOT NULL GROUP BY users.id, users.name HAVING order_count > ? ORDER BY order_count DESC LIMIT 10",
			wantArgs: []any{true, 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			qb := newQueryBuilder(conn, "users")
			tt.build(qb)

			gotSQL, gotArgs := qb.ToSQL()

			if gotSQL != tt.wantSQL {
				t.Errorf("\nwant SQL: %s\ngot  SQL: %s", tt.wantSQL, gotSQL)
			}

			if len(gotArgs) != len(tt.wantArgs) {
				t.Errorf("want %d args, got %d args", len(tt.wantArgs), len(gotArgs))
			}

			for i := range tt.wantArgs {
				if i >= len(gotArgs) {
					break
				}
				if gotArgs[i] != tt.wantArgs[i] {
					t.Errorf("arg[%d]: want %v, got %v", i, tt.wantArgs[i], gotArgs[i])
				}
			}
		})
	}
}

// TestQueryBuilder_InsertSQL tests INSERT SQL generation
func TestQueryBuilder_InsertSQL(t *testing.T) {
	// We can't easily test the Insert method without a real DB,
	// but we can verify the SQL logic by inspecting the method's behavior.
	// For now, we'll test with SQLite in memory.

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Create test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	conn := newConnection(db, "test")

	// Test single insert
	qb := conn.Table("users")
	id, err := qb.Insert(map[string]any{
		"name": "John",
		"age":  30,
	})

	if err != nil {
		t.Errorf("Insert failed: %v", err)
	}

	if id == 0 {
		t.Error("Expected non-zero insert ID")
	}

	// Verify insert
	row, err := conn.SelectOne("SELECT name, age FROM users WHERE id = ?", id)
	if err != nil {
		t.Errorf("Failed to select inserted row: %v", err)
	}

	if row["name"] != "John" {
		t.Errorf("Expected name 'John', got %v", row["name"])
	}
}

// TestQueryBuilder_UpdateSQL tests UPDATE SQL generation
func TestQueryBuilder_UpdateSQL(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (name, age) VALUES ('John', 30)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	conn := newConnection(db, "test")

	// Test update
	qb := conn.Table("users")
	affected, err := qb.Where("name", "=", "John").Update(map[string]any{
		"age": 31,
	})

	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// Verify update
	row, err := conn.SelectOne("SELECT age FROM users WHERE name = ?", "John")
	if err != nil {
		t.Errorf("Failed to select updated row: %v", err)
	}

	// SQLite returns int64 for integers
	age, ok := row["age"].(int64)
	if !ok {
		t.Errorf("age is not int64, got %T", row["age"])
	}

	if age != 31 {
		t.Errorf("Expected age 31, got %d", age)
	}
}

// TestQueryBuilder_DeleteSQL tests DELETE SQL generation
func TestQueryBuilder_DeleteSQL(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Create and populate test table
	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, age INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO users (name, age) VALUES ('John', 30), ('Jane', 25)`)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}

	conn := newConnection(db, "test")

	// Test delete
	qb := conn.Table("users")
	affected, err := qb.Where("name", "=", "John").Delete()

	if err != nil {
		t.Errorf("Delete failed: %v", err)
	}

	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// Verify delete
	rows, err := conn.Select("SELECT * FROM users")
	if err != nil {
		t.Errorf("Failed to select after delete: %v", err)
	}

	if len(rows) != 1 {
		t.Errorf("Expected 1 remaining row, got %d", len(rows))
	}

	if rows[0]["name"] != "Jane" {
		t.Errorf("Expected remaining user 'Jane', got %v", rows[0]["name"])
	}
}

// TestConnection_Transaction tests transaction handling
func TestConnection_Transaction(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	conn := newConnection(db, "test")

	// Test successful transaction
	err = conn.Transaction(func(tx *Connection) error {
		_, err := tx.Insert("INSERT INTO users (name) VALUES (?)", "Alice")
		if err != nil {
			return err
		}

		_, err = tx.Insert("INSERT INTO users (name) VALUES (?)", "Bob")
		return err
	})

	if err != nil {
		t.Errorf("Transaction failed: %v", err)
	}

	// Verify both inserts succeeded
	rows, err := conn.Select("SELECT name FROM users ORDER BY id")
	if err != nil {
		t.Errorf("Failed to select: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows, got %d", len(rows))
	}

	// Test rollback on error
	err = conn.Transaction(func(tx *Connection) error {
		_, err := tx.Insert("INSERT INTO users (name) VALUES (?)", "Charlie")
		if err != nil {
			return err
		}

		return sql.ErrTxDone // Force error to trigger rollback
	})

	if err == nil {
		t.Error("Expected transaction to fail")
	}

	// Verify rollback
	rows, err = conn.Select("SELECT name FROM users ORDER BY id")
	if err != nil {
		t.Errorf("Failed to select: %v", err)
	}

	if len(rows) != 2 {
		t.Errorf("Expected 2 rows after rollback, got %d", len(rows))
	}
}

// TestManager_Connection tests connection management
func TestManager_Connection(t *testing.T) {
	config := map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver": "sqlite3",
				"dsn":    ":memory:",
			},
		},
	}

	manager := NewManager(config)
	defer manager.Close()

	// Test getting connection
	conn, err := manager.Connection("sqlite")
	if err != nil {
		t.Fatalf("Failed to get connection: %v", err)
	}

	if conn == nil {
		t.Error("Connection is nil")
	}

	// Test getting same connection again (should return cached)
	conn2, err := manager.Connection("sqlite")
	if err != nil {
		t.Fatalf("Failed to get connection second time: %v", err)
	}

	if conn != conn2 {
		t.Error("Expected same connection instance")
	}

	// Test default connection
	defaultConn, err := manager.Default()
	if err != nil {
		t.Fatalf("Failed to get default connection: %v", err)
	}

	if defaultConn != conn {
		t.Error("Expected default connection to match 'sqlite'")
	}
}

// TestManager_Listeners tests query event listeners
func TestManager_Listeners(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	conn := newConnection(db, "test")

	var capturedEvent *QueryEvent

	// Register listener
	conn.Listen(func(event *QueryEvent) {
		capturedEvent = event
	})

	// Execute query
	_, err = conn.Select("SELECT * FROM users")
	if err != nil {
		t.Fatalf("Select failed: %v", err)
	}

	// Verify listener was called
	if capturedEvent == nil {
		t.Error("Listener was not called")
	}

	if capturedEvent.SQL != "SELECT * FROM users" {
		t.Errorf("Expected SQL 'SELECT * FROM users', got %s", capturedEvent.SQL)
	}

	if capturedEvent.Connection != "test" {
		t.Errorf("Expected connection 'test', got %s", capturedEvent.Connection)
	}
}

// TestQueryBuilder_Increment tests increment operation
func TestQueryBuilder_Increment(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE counters (id INTEGER PRIMARY KEY, count INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO counters (count) VALUES (10)`)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	conn := newConnection(db, "test")

	// Test increment
	qb := conn.Table("counters")
	affected, err := qb.Where("id", "=", 1).Increment("count", 5)

	if err != nil {
		t.Errorf("Increment failed: %v", err)
	}

	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// Verify
	row, err := conn.SelectOne("SELECT count FROM counters WHERE id = 1")
	if err != nil {
		t.Errorf("Failed to select: %v", err)
	}

	count, ok := row["count"].(int64)
	if !ok {
		t.Errorf("count is not int64")
	}

	if count != 15 {
		t.Errorf("Expected count 15, got %d", count)
	}
}

// TestQueryBuilder_Decrement tests decrement operation
func TestQueryBuilder_Decrement(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE counters (id INTEGER PRIMARY KEY, count INTEGER)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	_, err = db.Exec(`INSERT INTO counters (count) VALUES (20)`)
	if err != nil {
		t.Fatalf("Failed to insert: %v", err)
	}

	conn := newConnection(db, "test")

	// Test decrement
	qb := conn.Table("counters")
	affected, err := qb.Where("id", "=", 1).Decrement("count", 5)

	if err != nil {
		t.Errorf("Decrement failed: %v", err)
	}

	if affected != 1 {
		t.Errorf("Expected 1 affected row, got %d", affected)
	}

	// Verify
	row, err := conn.SelectOne("SELECT count FROM counters WHERE id = 1")
	if err != nil {
		t.Errorf("Failed to select: %v", err)
	}

	count, ok := row["count"].(int64)
	if !ok {
		t.Errorf("count is not int64")
	}

	if count != 15 {
		t.Errorf("Expected count 15, got %d", count)
	}
}
