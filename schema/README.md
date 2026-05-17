# Schema Package

The `schema` package provides a Laravel-inspired Schema Builder and Blueprint API for database schema management in GoFrame.

## Features

- Fluent API for defining table structures
- Support for all common column types
- Index management (regular, unique, fulltext)
- Foreign key constraints with cascading
- Migration management system
- Zero external dependencies
- MySQL syntax by default (extensible to other databases)

## Installation

The schema package is part of GoFrame. Simply import it:

```go
import "github.com/sazzad/goframe/schema"
```

## Quick Start

### Creating a Table

```go
db := &YourDatabaseConnection{}
s := schema.NewSchema(db)

err := s.Create("users", func(b *schema.Blueprint) {
    b.ID()
    b.String("name")
    b.String("email").Unique()
    b.String("password")
    b.RememberToken()
    b.Timestamps()
    b.SoftDeletes()
})
```

### Modifying a Table

```go
err := s.Table("users", func(b *schema.Blueprint) {
    b.String("phone", 20).Nullable()
    b.Date("birth_date").Nullable()
    b.Boolean("is_verified").Default("0")
})
```

### Foreign Keys

```go
err := s.Create("posts", func(b *schema.Blueprint) {
    b.ID()
    b.ForeignId("user_id").Constrained().CascadeOnDelete()
    b.String("title")
    b.Text("content")
    b.Timestamps()
    
    b.Index([]string{"user_id", "created_at"})
})
```

## Column Types

### Numeric Types

```go
b.Integer("count")
b.BigInteger("views")
b.SmallInteger("priority")
b.TinyInteger("status")
b.Float("rating", 3, 1)
b.Decimal("price", 10, 2)
b.Double("latitude")
b.Boolean("is_active")
```

### String Types

```go
b.String("name")                // VARCHAR(255)
b.String("code", 50)            // VARCHAR(50)
b.Text("description")
b.MediumText("content")
b.LongText("article")
b.Char("country_code", 2)
```

### Date and Time Types

```go
b.Date("birth_date")
b.DateTime("published_at")
b.Timestamp("created_at")
b.Time("start_time")
b.Timestamps()                  // created_at + updated_at
```

### Special Types

```go
b.JSON("metadata")
b.JSONB("settings")
b.Enum("status", []string{"active", "inactive"})
b.Binary("data")
b.UUID("uuid")
b.ULID("ulid")
```

### Laravel Conveniences

```go
b.ID()                          // BIGINT UNSIGNED AUTO_INCREMENT
b.Timestamps()                  // created_at + updated_at
b.SoftDeletes()                 // deleted_at (nullable)
b.RememberToken()               // remember_token VARCHAR(100)
b.MorphColumns("commentable")   // commentable_id + commentable_type
```

## Column Modifiers

All column methods return a `*ColumnDefinition` that supports chaining:

```go
b.String("email").
    Nullable().
    Default("user@example.com").
    Unique().
    Comment("User email address").
    After("name")

b.Integer("views").
    Unsigned().
    Default("0").
    Index()

b.String("name").
    Charset("utf8mb4").
    Collation("utf8mb4_unicode_ci")
```

Available modifiers:
- `Nullable()` - Allow NULL values
- `Default(value)` - Set default value
- `Unsigned()` - Unsigned numeric type
- `Unique()` - Add unique constraint
- `Primary()` - Mark as primary key
- `AutoIncrement()` - Auto-increment
- `After(column)` - Position after column (ALTER TABLE)
- `First()` - Position first in table (ALTER TABLE)
- `Comment(text)` - Add column comment
- `Index()` - Add regular index
- `Charset(cs)` - Set character set
- `Collation(co)` - Set collation

## Indexes

### Regular Indexes

```go
b.Index([]string{"email"})
b.Index([]string{"first_name", "last_name"}, "idx_full_name")
```

### Unique Indexes

```go
b.Unique([]string{"email"})
b.Unique([]string{"tenant_id", "slug"}, "uniq_tenant_slug")
```

### Fulltext Indexes

```go
b.FullText([]string{"title", "content"})
```

### Primary Keys

```go
b.Primary([]string{"id"})
b.Primary([]string{"tenant_id", "user_id"}) // Composite key
```

## Foreign Keys

```go
// Simple foreign key
b.ForeignId("user_id").Constrained()

// With explicit table
b.ForeignId("author_id").Constrained("users")

// With cascade on delete
b.ForeignId("user_id").Constrained().CascadeOnDelete()

// Manual foreign key definition
// (handled automatically by ForeignId().Constrained())
```

## Altering Tables

### Adding Columns

```go
s.Table("users", func(b *schema.Blueprint) {
    b.String("phone", 20).Nullable()
    b.Date("birth_date").Nullable().After("email")
})
```

### Modifying Columns

```go
s.Table("users", func(b *schema.Blueprint) {
    b.ModifyColumn("name", "VARCHAR", 150)
})
```

### Renaming Columns

```go
s.Table("users", func(b *schema.Blueprint) {
    b.RenameColumn("old_name", "new_name")
})
```

### Dropping Columns

```go
s.Table("users", func(b *schema.Blueprint) {
    b.DropColumn("deprecated_field")
    b.DropColumn("another_field", "third_field")
})
```

### Dropping Indexes

```go
s.Table("users", func(b *schema.Blueprint) {
    b.DropIndex("idx_email")
    b.DropUnique("uniq_username")
    b.DropForeign("fk_user_id")
    b.DropPrimary()
})
```

## Schema Operations

### Table Management

```go
// Check if table exists
if s.HasTable("users") {
    fmt.Println("Table exists")
}

// Check if column exists
if s.HasColumn("users", "email") {
    fmt.Println("Column exists")
}

// Get column type
dataType := s.GetColumnType("users", "email")

// Rename table
s.Rename("old_users", "users")

// Drop table
s.Drop("users")
s.DropIfExists("temp_table")

// Truncate table
s.Truncate("sessions")
```

## Migrations

### Creating Migrations

```go
type CreateUsersTable struct{}

func (m *CreateUsersTable) Up(s *schema.Schema) error {
    return s.Create("users", func(b *schema.Blueprint) {
        b.ID()
        b.String("name")
        b.String("email").Unique()
        b.Timestamps()
    })
}

func (m *CreateUsersTable) Down(s *schema.Schema) error {
    return s.Drop("users")
}
```

### Running Migrations

```go
migrator := schema.NewMigrator(s)

// Register migrations in chronological order
migrator.Register("2024_01_01_000001_create_users_table", &CreateUsersTable{})
migrator.Register("2024_01_01_000002_create_posts_table", &CreatePostsTable{})

// Run all pending migrations
if err := migrator.Migrate(); err != nil {
    log.Fatal(err)
}

// Rollback last batch
if err := migrator.Rollback(); err != nil {
    log.Fatal(err)
}

// Reset all migrations
if err := migrator.Reset(); err != nil {
    log.Fatal(err)
}

// Drop all tables and re-run migrations
if err := migrator.Fresh(); err != nil {
    log.Fatal(err)
}

// Get migration status
statuses, err := migrator.Status()
for _, status := range statuses {
    fmt.Printf("%s: %s\n", status.Name, status.Status)
}
```

## Connection Interface

The schema package works with any database connection that implements the `Connection` interface:

```go
type Connection interface {
    Exec(query string, args ...interface{}) (int64, error)
    Query(query string, args ...interface{}) ([]map[string]interface{}, error)
}
```

This makes it easy to integrate with your existing database layer or ORM.

## Generated SQL Examples

### CREATE TABLE

```sql
CREATE TABLE `users` (
  `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
  `name` VARCHAR(255) NOT NULL,
  `email` VARCHAR(255) NOT NULL,
  `password` VARCHAR(255) NOT NULL,
  `remember_token` VARCHAR(100) NULL,
  `created_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `deleted_at` TIMESTAMP NULL,
  PRIMARY KEY (`id`),
  UNIQUE INDEX `uniq_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci
```

### ALTER TABLE

```sql
ALTER TABLE `users` ADD COLUMN `phone` VARCHAR(20) NULL AFTER `email`;
ALTER TABLE `users` DROP COLUMN `deprecated_field`;
ALTER TABLE `users` RENAME COLUMN `old_name` TO `new_name`;
ALTER TABLE `users` MODIFY COLUMN `name` VARCHAR(150) NOT NULL
```

## Testing

The package includes comprehensive tests:

```bash
go test ./schema -v
```

Run benchmarks:

```bash
go test ./schema -bench=.
```

## Best Practices

1. **Migration Naming**: Use timestamp prefixes for migrations (e.g., `2024_01_01_000001_create_users_table`)
2. **Rollback Support**: Always implement both `Up()` and `Down()` methods
3. **Foreign Keys**: Define foreign keys after all referenced tables are created
4. **Indexes**: Add indexes on frequently queried columns
5. **Soft Deletes**: Use `SoftDeletes()` for data that should be recoverable
6. **Timestamps**: Always include `Timestamps()` for audit trails

## Examples

See [example_test.go](./example_test.go) for comprehensive usage examples including:
- Basic table creation
- Foreign key relationships
- Polymorphic relationships
- E-commerce schema
- Migration management

## License

Part of the GoFrame project.
