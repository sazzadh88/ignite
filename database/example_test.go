package database_test

import (
	"fmt"
	"log"

	"github.com/sazzad/ignite/database"
	_ "github.com/mattn/go-sqlite3"
)

// Example demonstrates basic usage of the database package.
func Example() {
	// Create database manager with configuration
	config := map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver": "sqlite3",
				"dsn":    ":memory:",
			},
		},
	}

	manager := database.NewManager(config)
	defer manager.Close()

	// Get default connection
	conn, err := manager.Default()
	if err != nil {
		log.Fatal(err)
	}

	// Create a table
	err = conn.Statement(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT UNIQUE,
			age INTEGER
		)
	`)
	if err != nil {
		log.Fatal(err)
	}

	// Insert a user
	qb := conn.Table("users")
	userID, err := qb.Insert(map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Inserted user with ID: %d\n", userID)

	// Query users
	users, err := conn.Table("users").
		Where("age", ">", 25).
		OrderBy("name", "ASC").
		Get()
	if err != nil {
		log.Fatal(err)
	}

	for _, user := range users {
		fmt.Printf("User: %s (%s)\n", user["name"], user["email"])
	}

	// Update user
	affected, err := conn.Table("users").
		Where("email", "=", "john@example.com").
		Update(map[string]any{
			"age": 31,
		})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Updated %d rows\n", affected)

	// Use transaction
	err = conn.Transaction(func(tx *database.Connection) error {
		_, err := tx.Table("users").Insert(map[string]any{
			"name":  "Jane Doe",
			"email": "jane@example.com",
			"age":   28,
		})
		if err != nil {
			return err
		}

		_, err = tx.Table("users").Insert(map[string]any{
			"name":  "Bob Smith",
			"email": "bob@example.com",
			"age":   35,
		})
		return err
	})
	if err != nil {
		log.Fatal(err)
	}

	// Count users
	count, err := conn.Table("users").Count()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total users: %d\n", count)

	// Output:
	// Inserted user with ID: 1
	// User: John Doe (john@example.com)
	// Updated 1 rows
	// Total users: 3
}

// ExampleQueryBuilder_joins demonstrates JOIN operations.
func ExampleQueryBuilder_joins() {
	config := map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver": "sqlite3",
				"dsn":    ":memory:",
			},
		},
	}

	manager := database.NewManager(config)
	defer manager.Close()

	conn, err := manager.Default()
	if err != nil {
		log.Fatal(err)
	}

	// Create tables
	conn.Statement(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)`)
	conn.Statement(`CREATE TABLE orders (id INTEGER PRIMARY KEY, user_id INTEGER, total REAL)`)

	// Insert data
	conn.Table("users").Insert(map[string]any{"name": "John"})
	conn.Table("orders").Insert(map[string]any{"user_id": 1, "total": 99.99})

	// Query with join
	results, err := conn.Table("users").
		Select("users.name", "orders.total").
		LeftJoin("orders", "users.id", "=", "orders.user_id").
		Get()
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range results {
		fmt.Printf("%s ordered $%.2f\n", row["name"], row["total"])
	}

	// Output:
	// John ordered $99.99
}

// ExampleQueryBuilder_aggregates demonstrates aggregate functions.
func ExampleQueryBuilder_aggregates() {
	config := map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver": "sqlite3",
				"dsn":    ":memory:",
			},
		},
	}

	manager := database.NewManager(config)
	defer manager.Close()

	conn, err := manager.Default()
	if err != nil {
		log.Fatal(err)
	}

	conn.Statement(`CREATE TABLE orders (id INTEGER PRIMARY KEY, total REAL)`)
	conn.Table("orders").InsertMany([]map[string]any{
		{"total": 10.50},
		{"total": 25.00},
		{"total": 15.75},
	})

	// Get sum
	sum, err := conn.Table("orders").Sum("total")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Total revenue: $%.2f\n", sum)

	// Get average
	avg, err := conn.Table("orders").Avg("total")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Average order: $%.2f\n", avg)

	// Output:
	// Total revenue: $51.25
	// Average order: $17.08
}

// ExampleManager_Listen demonstrates query event listening.
func ExampleManager_Listen() {
	config := map[string]any{
		"default": "sqlite",
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver": "sqlite3",
				"dsn":    ":memory:",
			},
		},
	}

	manager := database.NewManager(config)
	defer manager.Close()

	// Register query listener
	manager.Listen(func(event *database.QueryEvent) {
		fmt.Printf("Query executed in %v: %s\n", event.Time, event.SQL)
	})

	conn, _ := manager.Default()
	conn.Statement(`CREATE TABLE test (id INTEGER)`)

	// This will trigger the listener
	conn.Select("SELECT * FROM test")

	// Output will vary due to timing, but format will be:
	// Query executed in 123µs: CREATE TABLE test (id INTEGER)
	// Query executed in 456µs: SELECT * FROM test
}
