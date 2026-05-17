package schema_test

import (
	"fmt"
	"log"

	"github.com/sazzadh88/ignite/schema"
)

// mockDB is a simple mock for demonstration purposes.
type mockDB struct{}

func (m *mockDB) Exec(query string, args ...interface{}) (int64, error) {
	fmt.Printf("Executing: %s\n", query)
	return 1, nil
}

func (m *mockDB) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	fmt.Printf("Querying: %s\n", query)
	return nil, nil
}

// Example demonstrates creating a users table with various column types.
func Example_createTable() {
	// Create schema instance
	db := &mockDB{}
	s := schema.NewSchema(db)

	// Create users table
	err := s.Create("users", func(b *schema.Blueprint) {
		b.ID()
		b.String("name")
		b.String("email").Unique()
		b.String("password")
		b.RememberToken()
		b.Timestamps()
		b.SoftDeletes()
	})

	if err != nil {
		log.Fatal(err)
	}

	// Output will show the CREATE TABLE SQL
}

// Example demonstrates altering an existing table.
func Example_alterTable() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	// Add columns to existing table
	err := s.Table("users", func(b *schema.Blueprint) {
		b.String("phone", 20).Nullable()
		b.Date("birth_date").Nullable()
		b.Boolean("is_verified").Default("0")
	})

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates creating a posts table with foreign keys.
func Example_foreignKeys() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	err := s.Create("posts", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("user_id").Constrained().CascadeOnDelete()
		b.String("title")
		b.Text("content")
		b.Enum("status", []string{"draft", "published", "archived"})
		b.Integer("views").Unsigned().Default("0")
		b.Timestamps()

		// Add indexes
		b.Index([]string{"status"})
		b.Index([]string{"user_id", "created_at"})
		b.FullText([]string{"title", "content"})
	})

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates creating a polymorphic relationship table.
func Example_polymorphicRelations() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	err := s.Create("comments", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("user_id").Constrained().CascadeOnDelete()
		b.MorphColumns("commentable") // Creates commentable_id and commentable_type
		b.Text("content")
		b.Timestamps()

		b.Index([]string{"commentable_id", "commentable_type"})
	})

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates creating a products table with decimal columns.
func Example_decimalColumns() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	err := s.Create("products", func(b *schema.Blueprint) {
		b.ID()
		b.String("name")
		b.String("sku", 50).Unique()
		b.Text("description").Nullable()
		b.Decimal("price", 10, 2)
		b.Decimal("cost", 10, 2).Nullable()
		b.Integer("stock").Unsigned().Default("0")
		b.Boolean("is_active").Default("1")
		b.JSON("metadata").Nullable()
		b.Timestamps()
		b.SoftDeletes()

		b.Index([]string{"sku"})
		b.Index([]string{"is_active", "created_at"})
	})

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates using UUID and ULID columns.
func Example_uuidColumns() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	err := s.Create("sessions", func(b *schema.Blueprint) {
		b.UUID("id").Primary()
		b.ForeignId("user_id").Constrained().CascadeOnDelete()
		b.String("ip_address", 45).Nullable()
		b.Text("user_agent").Nullable()
		b.LongText("payload")
		b.Integer("last_activity")

		b.Index([]string{"user_id"})
		b.Index([]string{"last_activity"})
	})

	if err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates modifying existing columns.
func Example_modifyColumns() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	err := s.Table("users", func(b *schema.Blueprint) {
		// Modify existing column
		b.ModifyColumn("name", "VARCHAR", 100)

		// Rename column
		b.RenameColumn("old_field", "new_field")

		// Drop column
		b.DropColumn("deprecated_field")

		// Drop index
		b.DropIndex("idx_old_index")
	})

	if err != nil {
		log.Fatal(err)
	}
}

// CreateUsersTable is an example migration.
type CreateUsersTable struct{}

func (m *CreateUsersTable) Up(s *schema.Schema) error {
	return s.Create("users", func(b *schema.Blueprint) {
		b.ID()
		b.String("name")
		b.String("email").Unique()
		b.String("password")
		b.RememberToken()
		b.Timestamps()
		b.SoftDeletes()
	})
}

func (m *CreateUsersTable) Down(s *schema.Schema) error {
	return s.Drop("users")
}

// CreatePostsTable is another example migration.
type CreatePostsTable struct{}

func (m *CreatePostsTable) Up(s *schema.Schema) error {
	return s.Create("posts", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("user_id").Constrained().CascadeOnDelete()
		b.String("title")
		b.Text("content")
		b.Enum("status", []string{"draft", "published"})
		b.Timestamps()

		b.Index([]string{"status"})
		b.FullText([]string{"title", "content"})
	})
}

func (m *CreatePostsTable) Down(s *schema.Schema) error {
	return s.Drop("posts")
}

// Example demonstrates using the migrator to manage migrations.
func Example_migrations() {
	db := &mockDB{}
	s := schema.NewSchema(db)
	migrator := schema.NewMigrator(s)

	// Register migrations in order
	migrator.Register("2024_01_01_000001_create_users_table", &CreateUsersTable{})
	migrator.Register("2024_01_01_000002_create_posts_table", &CreatePostsTable{})

	// Run all pending migrations
	if err := migrator.Migrate(); err != nil {
		log.Fatal(err)
	}

	// Get migration status
	statuses, err := migrator.Status()
	if err != nil {
		log.Fatal(err)
	}

	for _, status := range statuses {
		fmt.Printf("%s: %s\n", status.Name, status.Status)
	}
}

// Example demonstrates rolling back migrations.
func Example_rollback() {
	db := &mockDB{}
	s := schema.NewSchema(db)
	migrator := schema.NewMigrator(s)

	migrator.Register("2024_01_01_000001_create_users_table", &CreateUsersTable{})
	migrator.Register("2024_01_01_000002_create_posts_table", &CreatePostsTable{})

	// Rollback the last batch of migrations
	if err := migrator.Rollback(); err != nil {
		log.Fatal(err)
	}

	// Reset all migrations (rollback everything)
	if err := migrator.Reset(); err != nil {
		log.Fatal(err)
	}

	// Drop all tables and re-run all migrations
	if err := migrator.Fresh(); err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates checking table and column existence.
func Example_schemaInspection() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	// Check if table exists
	if s.HasTable("users") {
		fmt.Println("Users table exists")
	}

	// Check if column exists
	if s.HasColumn("users", "email") {
		fmt.Println("Email column exists")
	}

	// Get column type
	columnType := s.GetColumnType("users", "email")
	fmt.Printf("Email column type: %s\n", columnType)

	// Truncate table
	if err := s.Truncate("sessions"); err != nil {
		log.Fatal(err)
	}

	// Rename table
	if err := s.Rename("old_users", "users"); err != nil {
		log.Fatal(err)
	}

	// Drop table
	if err := s.DropIfExists("temp_table"); err != nil {
		log.Fatal(err)
	}
}

// Example demonstrates creating a complex e-commerce schema.
func Example_ecommerceSchema() {
	db := &mockDB{}
	s := schema.NewSchema(db)

	// Categories table
	s.Create("categories", func(b *schema.Blueprint) {
		b.ID()
		b.String("name")
		b.String("slug").Unique()
		b.Text("description").Nullable()
		b.BigInteger("parent_id").Unsigned().Nullable()
		b.Integer("sort_order").Default("0")
		b.Timestamps()

		b.Index([]string{"parent_id"})
		b.Index([]string{"slug"})
	})

	// Products table
	s.Create("products", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("category_id").Constrained()
		b.String("name")
		b.String("slug").Unique()
		b.Text("description").Nullable()
		b.LongText("specifications").Nullable()
		b.Decimal("price", 10, 2)
		b.Decimal("sale_price", 10, 2).Nullable()
		b.Integer("stock").Unsigned().Default("0")
		b.String("sku", 50).Unique()
		b.Boolean("is_featured").Default("0")
		b.Boolean("is_active").Default("1")
		b.JSON("images").Nullable()
		b.JSON("attributes").Nullable()
		b.Timestamps()
		b.SoftDeletes()

		b.Index([]string{"category_id"})
		b.Index([]string{"slug"})
		b.Index([]string{"sku"})
		b.Index([]string{"is_featured", "is_active"})
		b.FullText([]string{"name", "description"})
	})

	// Orders table
	s.Create("orders", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("user_id").Constrained()
		b.String("order_number", 50).Unique()
		b.Enum("status", []string{"pending", "processing", "completed", "cancelled"})
		b.Decimal("subtotal", 10, 2)
		b.Decimal("tax", 10, 2)
		b.Decimal("shipping", 10, 2)
		b.Decimal("total", 10, 2)
		b.Text("shipping_address")
		b.Text("billing_address")
		b.Text("notes").Nullable()
		b.Timestamps()

		b.Index([]string{"user_id"})
		b.Index([]string{"order_number"})
		b.Index([]string{"status", "created_at"})
	})

	// Order items table
	s.Create("order_items", func(b *schema.Blueprint) {
		b.ID()
		b.ForeignId("order_id").Constrained().CascadeOnDelete()
		b.ForeignId("product_id").Constrained()
		b.Integer("quantity").Unsigned()
		b.Decimal("price", 10, 2)
		b.Decimal("total", 10, 2)
		b.Timestamps()

		b.Index([]string{"order_id"})
		b.Index([]string{"product_id"})
	})
}
