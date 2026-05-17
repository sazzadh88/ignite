# GoFrame ORM

A Laravel Eloquent-inspired ORM package for Go, featuring type-safe queries with generics, relationships, soft deletes, observers, and more.

## Features

- **Type-Safe Queries**: Generic query builder with compile-time type checking
- **Zero Dependencies**: Uses only Go standard library
- **Laravel-Inspired API**: Familiar interface for Laravel developers
- **Relationships**: HasOne, HasMany, BelongsTo, BelongsToMany, HasManyThrough, Polymorphic
- **Soft Deletes**: Automatic filtering of soft-deleted records
- **Query Scopes**: Global and local scopes for reusable query constraints
- **Attribute Casting**: Automatic type conversion (JSON, Bool, Int, Float, Date, Array)
- **Model Observers**: Lifecycle hooks (Creating, Created, Updating, Updated, etc.)
- **Dirty Tracking**: Track which fields have been modified
- **Mass Assignment Protection**: Fillable and hidden fields
- **Pagination**: Built-in pagination support
- **SQL Generation**: Testable query builder that generates SQL without database execution

## Installation

```bash
go get github.com/sazzad/goframe/orm
```

## Quick Start

### Define a Model

```go
type User struct {
    orm.Model
    orm.SoftDeleteModel
    Name     string `db:"name"`
    Email    string `db:"email"`
    Password string `db:"password"`
    Active   bool   `db:"active"`
}

func (u User) Table() string {
    return "users"
}

func (u User) Fillable() []string {
    return []string{"name", "email", "password"}
}

func (u User) Hidden() []string {
    return []string{"password"}
}

func (u User) PrimaryKey() string {
    return "id"
}
```

### Query Building

```go
// Create a query
query := orm.NewQuery[User]().Table("users")

// Simple WHERE
query.Where("active", "=", true)

// Multiple conditions
query.Where("active", "=", true).
    Where("email", "LIKE", "%@example.com")

// WHERE IN
query.WhereIn("id", []any{1, 2, 3, 4, 5})

// NULL checks
query.WhereNull("deleted_at")
query.WhereNotNull("email")

// Ordering
query.OrderBy("created_at", "desc")

// Pagination
query.Limit(10).Offset(20)

// Generate SQL (for testing)
sql, bindings := query.ToSQL()
// SELECT * FROM users WHERE active = ? LIMIT 10 OFFSET 20
```

### Relationships

```go
// HasMany
type User struct {
    orm.Model
    Posts *orm.HasMany[Post]
}

// BelongsTo
type Post struct {
    orm.Model
    User *orm.BelongsTo[User]
}

// BelongsToMany
type User struct {
    orm.Model
    Roles *orm.BelongsToMany[Role]
}

// Create relationships
posts := orm.NewHasMany[Post]("user_id", "id")
user := orm.NewBelongsTo[User]("user_id", "id")
roles := orm.NewBelongsToMany[Role]("user_role", "user_id", "role_id")
```

### Soft Deletes

```go
type User struct {
    orm.Model
    orm.SoftDeleteModel  // Add soft delete support
    Name string `db:"name"`
}

// Queries automatically exclude soft-deleted records
users, _ := query.Get()

// Include soft-deleted records
users, _ := query.WithTrashed().Get()

// Only soft-deleted records
users, _ := query.OnlyTrashed().Get()

// Check if trashed
model := &orm.SoftDeleteModel{}
if model.Trashed() {
    // Record is soft-deleted
}
```

### Attribute Casting

```go
casts := orm.Casts{
    "settings": orm.JSONCast{},
    "age":      orm.IntCast{},
    "active":   orm.BoolCast{},
    "price":    orm.FloatCast{},
    "tags":     orm.ArrayCast{},
    "birthday": orm.DateCast{Format: "2006-01-02"},
}

// Apply casts when retrieving from database
attrs := orm.ApplyCasts(casts, attributes, true)

// Apply casts when saving to database
attrs := orm.ApplyCasts(casts, attributes, false)
```

### Model Observers

```go
type UserObserver struct{}

func (o *UserObserver) Creating(model any) error {
    // Called before creating a user
    fmt.Println("Creating user...")
    return nil
}

func (o *UserObserver) Created(model any) error {
    // Called after creating a user
    fmt.Println("User created!")
    return nil
}

// Register observer
orm.Observe(&User{}, &UserObserver{})
```

Available observer methods:
- `Creating`, `Created`
- `Updating`, `Updated`
- `Saving`, `Saved`
- `Deleting`, `Deleted`
- `Restoring`, `Restored`
- `ForceDeleting`, `ForceDeleted`
- `Retrieved`

### Query Scopes

```go
// Local scope
activeScope := func(q *orm.Query[User]) *orm.Query[User] {
    return q.Where("active", "=", true)
}

query.Scope(activeScope)

// Built-in scopes
query.Scope(orm.WhereActive[User]())
query.Scope(orm.WherePublished[User]())
query.Scope(orm.Recent[User]())
query.Scope(orm.Oldest[User]())
```

### Dirty Tracking

```go
attrs := map[string]any{
    "name":  "John",
    "email": "john@example.com",
}

model := orm.NewModelInstance(&User{}, attrs)

// Modify a field
model.Set("name", "Jane")

// Check if dirty
if model.IsDirty() {
    fmt.Println("Model has unsaved changes")
}

// Check specific field
if model.IsDirty("name") {
    fmt.Println("Name has changed")
}

// Get all dirty fields
dirty := model.GetDirty()
// map[string]any{"name": "Jane"}

// Sync changes
model.SyncOriginal()
```

### Pagination

```go
paginator, err := query.Paginate(15, 1)

// Paginator contains:
// - Data: []T
// - Total: int64
// - PerPage: int
// - Page: int
// - LastPage: int
// - From: int
// - To: int
// - HasMore: bool
```

### Chunking

```go
// Process records in chunks
query.Chunk(100, func(users []User) bool {
    for _, user := range users {
        // Process each user
    }
    
    // Return true to continue, false to stop
    return true
})
```

## Testing

The query builder generates SQL without executing it, making it easy to test:

```go
func TestUserQuery(t *testing.T) {
    query := orm.NewQuery[User]().Table("users")
    query.Where("active", "=", true).OrderBy("name", "asc")
    
    sql, bindings := query.ToSQL()
    
    expected := "SELECT * FROM users WHERE active = ? ORDER BY name ASC"
    if sql != expected {
        t.Errorf("Expected: %s, Got: %s", expected, sql)
    }
}
```

Run tests:

```bash
go test ./orm/... -v -cover
```

## Design Principles

1. **Zero Dependencies**: Uses only Go standard library
2. **Type Safety**: Leverages Go generics for compile-time type checking
3. **Fluent API**: Chainable methods for expressive queries
4. **Laravel-Inspired**: Familiar API for Laravel developers
5. **Testable**: Query builder generates SQL without database execution
6. **Clean Code**: Functions under 20 lines, SOLID principles

## Architecture

### Core Components

- **model.go**: Base Model struct, Modeler interface, dirty tracking
- **query.go**: Generic query builder with WHERE, ORDER, LIMIT, etc.
- **relationships.go**: HasOne, HasMany, BelongsTo, BelongsToMany, etc.
- **scope.go**: Global and local scopes for reusable constraints
- **observer.go**: Model lifecycle hooks and event system
- **cast.go**: Attribute casting system with built-in casters
- **soft_delete.go**: Soft delete trait and scope

## Examples

See `example_test.go` for comprehensive examples of all features.

## License

MIT License

## Contributing

Contributions are welcome! Please ensure:
- All tests pass
- Code follows Go conventions
- Functions are under 20 lines
- Public APIs are documented with GoDoc
