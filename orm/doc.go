// Package orm provides a Laravel Eloquent-inspired ORM for Go.
//
// The orm package offers a fluent, expressive API for interacting with databases
// using Go's type system and generics. It provides a model-based approach similar
// to Laravel's Eloquent ORM.
//
// # Features
//
// - Generic query builder with type safety
// - Model lifecycle hooks and observers
// - Soft deletes
// - Relationships (HasOne, HasMany, BelongsTo, BelongsToMany, etc.)
// - Query scopes (global and local)
// - Attribute casting
// - Dirty tracking
// - Mass assignment protection
// - Zero external dependencies
//
// # Basic Usage
//
// Define a model:
//
//	type User struct {
//	    orm.Model
//	    Name  string `db:"name"`
//	    Email string `db:"email"`
//	}
//
//	func (u User) Table() string {
//	    return "users"
//	}
//
//	func (u User) Fillable() []string {
//	    return []string{"name", "email"}
//	}
//
//	func (u User) Hidden() []string {
//	    return []string{"password"}
//	}
//
//	func (u User) PrimaryKey() string {
//	    return "id"
//	}
//
// Query users:
//
//	query := orm.NewQuery[User]().Table("users")
//	users, err := query.Where("active", "=", true).Get()
//
// # Query Building
//
// The query builder provides a fluent interface for constructing SQL queries:
//
//	query := orm.NewQuery[User]().Table("users")
//
//	// Basic WHERE
//	query.Where("name", "=", "John")
//
//	// OR WHERE
//	query.OrWhere("name", "=", "Jane")
//
//	// WHERE IN
//	query.WhereIn("id", []any{1, 2, 3})
//
//	// WHERE NULL / NOT NULL
//	query.WhereNull("deleted_at")
//	query.WhereNotNull("email")
//
//	// ORDER BY
//	query.OrderBy("created_at", "desc")
//
//	// LIMIT and OFFSET
//	query.Limit(10).Offset(20)
//
// # Relationships
//
// Define relationships between models:
//
//	type User struct {
//	    orm.Model
//	    Posts *orm.HasMany[Post]
//	}
//
//	type Post struct {
//	    orm.Model
//	    User *orm.BelongsTo[User]
//	}
//
// # Soft Deletes
//
// Enable soft deletes on a model:
//
//	type User struct {
//	    orm.Model
//	    orm.SoftDeleteModel
//	    Name string `db:"name"`
//	}
//
// Soft deleted records are automatically excluded from queries:
//
//	users, _ := query.Get() // Excludes soft deleted
//	users, _ := query.WithTrashed().Get() // Includes soft deleted
//	users, _ := query.OnlyTrashed().Get() // Only soft deleted
//
// # Attribute Casting
//
// Cast attributes to specific types:
//
//	casts := orm.Casts{
//	    "settings": orm.JSONCast{},
//	    "age":      orm.IntCast{},
//	    "active":   orm.BoolCast{},
//	}
//
// # Observers
//
// Listen to model lifecycle events:
//
//	type UserObserver struct{}
//
//	func (o *UserObserver) Creating(model any) error {
//	    // Called before a user is created
//	    return nil
//	}
//
//	func (o *UserObserver) Created(model any) error {
//	    // Called after a user is created
//	    return nil
//	}
//
//	orm.Observe(&User{}, &UserObserver{})
//
// # Scopes
//
// Define reusable query constraints:
//
//	activeScope := func(q *orm.Query[User]) *orm.Query[User] {
//	    return q.Where("active", "=", true)
//	}
//
//	users, _ := query.Scope(activeScope).Get()
//
// # Pagination
//
// Paginate query results:
//
//	paginator, err := query.Paginate(15, 1)
//	// Returns: Data, Total, PerPage, Page, LastPage, etc.
//
// # Design Principles
//
// - Zero external dependencies: Uses only Go standard library
// - Type safety: Leverages Go generics for compile-time type checking
// - Fluent API: Chainable methods for expressive queries
// - Laravel-inspired: Familiar API for Laravel developers
// - Testable: Query builder generates SQL without database execution
//
package orm
