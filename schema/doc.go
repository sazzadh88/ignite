/*
Package schema provides a Laravel-inspired Schema Builder and Blueprint API
for database schema management in Ignite.

The schema package offers a fluent, expressive API for creating and modifying
database tables without writing raw SQL. It supports all common column types,
indexes, foreign keys, and provides a complete migration management system.

# Basic Usage

Create a new table:

	db := &YourDatabaseConnection{}
	s := schema.NewSchema(db)

	err := s.Create("users", func(b *schema.Blueprint) {
	    b.ID()
	    b.String("name")
	    b.String("email").Unique()
	    b.String("password")
	    b.Timestamps()
	})

Modify an existing table:

	err := s.Table("users", func(b *schema.Blueprint) {
	    b.String("phone", 20).Nullable()
	    b.Boolean("is_verified").Default("0")
	})

# Column Types

The Blueprint provides methods for all common column types:

Numeric: Integer, BigInteger, SmallInteger, TinyInteger, Float, Decimal, Double, Boolean

String: String (VARCHAR), Text, MediumText, LongText

Date/Time: Date, DateTime, Timestamp, Time

Special: JSON, JSONB, Enum, Binary, UUID, ULID

Each column method returns a ColumnDefinition that supports chaining modifiers
like Nullable(), Default(), Unique(), Comment(), etc.

# Indexes and Constraints

Add indexes:

	b.Index([]string{"email"})
	b.Unique([]string{"username"})
	b.FullText([]string{"title", "content"})

Add foreign keys:

	b.ForeignId("user_id").Constrained().CascadeOnDelete()

# Migrations

Create migrations by implementing the Migration interface:

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

Run migrations:

	migrator := schema.NewMigrator(s)
	migrator.Register("2024_01_01_create_users_table", &CreateUsersTable{})
	migrator.Migrate()

# Connection Interface

The schema package works with any database connection implementing:

	type Connection interface {
	    Exec(query string, args ...interface{}) (int64, error)
	    Query(query string, args ...interface{}) ([]map[string]interface{}, error)
	}

This makes it easy to integrate with your existing database layer or ORM.

# Zero Dependencies

The schema package has zero external dependencies and generates clean,
standard SQL that works with MySQL by default. The design is extensible
to support other database dialects in the future.
*/
package schema
