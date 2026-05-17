package schema

import (
	"fmt"
	"strings"
	"testing"
)

// mockConnection is a mock implementation of Connection for testing.
type mockConnection struct {
	queries []string
	results [][]map[string]interface{}
}

func (m *mockConnection) Exec(query string, args ...interface{}) (int64, error) {
	m.queries = append(m.queries, query)
	return 1, nil
}

func (m *mockConnection) Query(query string, args ...interface{}) ([]map[string]interface{}, error) {
	m.queries = append(m.queries, query)
	if len(m.results) > 0 {
		result := m.results[0]
		m.results = m.results[1:]
		return result, nil
	}
	return []map[string]interface{}{}, nil
}

func TestBlueprintColumns(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Blueprint)
		expected string
	}{
		{
			name: "ID column",
			setup: func(b *Blueprint) {
				b.ID()
			},
			expected: "`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		},
		{
			name: "String column with default length",
			setup: func(b *Blueprint) {
				b.String("name")
			},
			expected: "`name` VARCHAR(255) NOT NULL",
		},
		{
			name: "String column with custom length",
			setup: func(b *Blueprint) {
				b.String("code", 50)
			},
			expected: "`code` VARCHAR(50) NOT NULL",
		},
		{
			name: "Nullable string column",
			setup: func(b *Blueprint) {
				b.String("nickname").Nullable()
			},
			expected: "`nickname` VARCHAR(255) NULL",
		},
		{
			name: "String column with default value",
			setup: func(b *Blueprint) {
				b.String("status").Default("active")
			},
			expected: "`status` VARCHAR(255) NOT NULL DEFAULT 'active'",
		},
		{
			name: "Integer column",
			setup: func(b *Blueprint) {
				b.Integer("age")
			},
			expected: "`age` INT NOT NULL",
		},
		{
			name: "Unsigned integer",
			setup: func(b *Blueprint) {
				b.Integer("count").Unsigned()
			},
			expected: "`count` INT UNSIGNED NOT NULL",
		},
		{
			name: "BigInteger column",
			setup: func(b *Blueprint) {
				b.BigInteger("views")
			},
			expected: "`views` BIGINT NOT NULL",
		},
		{
			name: "SmallInteger column",
			setup: func(b *Blueprint) {
				b.SmallInteger("priority")
			},
			expected: "`priority` SMALLINT NOT NULL",
		},
		{
			name: "TinyInteger column",
			setup: func(b *Blueprint) {
				b.TinyInteger("status_code")
			},
			expected: "`status_code` TINYINT NOT NULL",
		},
		{
			name: "Boolean column",
			setup: func(b *Blueprint) {
				b.Boolean("is_active")
			},
			expected: "`is_active` TINYINT(1) NOT NULL",
		},
		{
			name: "Decimal column",
			setup: func(b *Blueprint) {
				b.Decimal("price", 10, 2)
			},
			expected: "`price` DECIMAL(10,2) NOT NULL",
		},
		{
			name: "Float column",
			setup: func(b *Blueprint) {
				b.Float("rating", 3, 1)
			},
			expected: "`rating` FLOAT(3,1) NOT NULL",
		},
		{
			name: "Double column",
			setup: func(b *Blueprint) {
				b.Double("latitude")
			},
			expected: "`latitude` DOUBLE NOT NULL",
		},
		{
			name: "Text column",
			setup: func(b *Blueprint) {
				b.Text("description")
			},
			expected: "`description` TEXT NOT NULL",
		},
		{
			name: "LongText column",
			setup: func(b *Blueprint) {
				b.LongText("content")
			},
			expected: "`content` LONGTEXT NOT NULL",
		},
		{
			name: "MediumText column",
			setup: func(b *Blueprint) {
				b.MediumText("body")
			},
			expected: "`body` MEDIUMTEXT NOT NULL",
		},
		{
			name: "Date column",
			setup: func(b *Blueprint) {
				b.Date("birth_date")
			},
			expected: "`birth_date` DATE NOT NULL",
		},
		{
			name: "DateTime column",
			setup: func(b *Blueprint) {
				b.DateTime("published_at")
			},
			expected: "`published_at` DATETIME NOT NULL",
		},
		{
			name: "Timestamp column",
			setup: func(b *Blueprint) {
				b.Timestamp("created_at")
			},
			expected: "`created_at` TIMESTAMP NOT NULL",
		},
		{
			name: "Time column",
			setup: func(b *Blueprint) {
				b.Time("start_time")
			},
			expected: "`start_time` TIME NOT NULL",
		},
		{
			name: "JSON column",
			setup: func(b *Blueprint) {
				b.JSON("metadata")
			},
			expected: "`metadata` JSON NOT NULL",
		},
		{
			name: "JSONB column",
			setup: func(b *Blueprint) {
				b.JSONB("settings")
			},
			expected: "`settings` JSON NOT NULL",
		},
		{
			name: "Enum column",
			setup: func(b *Blueprint) {
				b.Enum("status", []string{"pending", "active", "inactive"})
			},
			expected: "`status` ENUM('pending', 'active', 'inactive') NOT NULL",
		},
		{
			name: "Binary column",
			setup: func(b *Blueprint) {
				b.Binary("data")
			},
			expected: "`data` BLOB NOT NULL",
		},
		{
			name: "UUID column",
			setup: func(b *Blueprint) {
				b.UUID("uuid")
			},
			expected: "`uuid` CHAR(36) NOT NULL",
		},
		{
			name: "ULID column",
			setup: func(b *Blueprint) {
				b.ULID("ulid")
			},
			expected: "`ulid` CHAR(26) NOT NULL",
		},
		{
			name: "Column with comment",
			setup: func(b *Blueprint) {
				b.String("email").Comment("User email address")
			},
			expected: "`email` VARCHAR(255) NOT NULL COMMENT 'User email address'",
		},
		{
			name: "Column with charset and collation",
			setup: func(b *Blueprint) {
				b.String("name").Charset("utf8mb4").Collation("utf8mb4_unicode_ci")
			},
			expected: "`name` VARCHAR(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL",
		},
		{
			name: "Unique column",
			setup: func(b *Blueprint) {
				b.String("email").Unique()
			},
			expected: "`email` VARCHAR(255) NOT NULL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := NewBlueprint()
			tt.setup(b)

			if len(b.columns) == 0 {
				t.Fatal("No columns were added to blueprint")
			}

			result := b.columns[0].toSQL()
			if result != tt.expected {
				t.Errorf("\nExpected: %s\nGot:      %s", tt.expected, result)
			}
		})
	}
}

func TestBlueprintTimestamps(t *testing.T) {
	b := NewBlueprint()
	b.Timestamps()

	if len(b.columns) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(b.columns))
	}

	createdAt := b.columns[0].toSQL()
	if !strings.Contains(createdAt, "created_at") {
		t.Errorf("Expected created_at column, got: %s", createdAt)
	}

	updatedAt := b.columns[1].toSQL()
	if !strings.Contains(updatedAt, "updated_at") {
		t.Errorf("Expected updated_at column, got: %s", updatedAt)
	}
}

func TestBlueprintSoftDeletes(t *testing.T) {
	b := NewBlueprint()
	b.SoftDeletes()

	if len(b.columns) != 1 {
		t.Fatalf("Expected 1 column, got %d", len(b.columns))
	}

	deletedAt := b.columns[0].toSQL()
	if !strings.Contains(deletedAt, "deleted_at") || !strings.Contains(deletedAt, "NULL") {
		t.Errorf("Expected nullable deleted_at column, got: %s", deletedAt)
	}
}

func TestBlueprintMorphColumns(t *testing.T) {
	b := NewBlueprint()
	b.MorphColumns("commentable")

	if len(b.columns) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(b.columns))
	}

	idCol := b.columns[0].toSQL()
	if !strings.Contains(idCol, "commentable_id") {
		t.Errorf("Expected commentable_id column, got: %s", idCol)
	}

	typeCol := b.columns[1].toSQL()
	if !strings.Contains(typeCol, "commentable_type") {
		t.Errorf("Expected commentable_type column, got: %s", typeCol)
	}
}

func TestBlueprintRememberToken(t *testing.T) {
	b := NewBlueprint()
	b.RememberToken()

	if len(b.columns) != 1 {
		t.Fatalf("Expected 1 column, got %d", len(b.columns))
	}

	token := b.columns[0].toSQL()
	if !strings.Contains(token, "remember_token") || !strings.Contains(token, "VARCHAR(100)") {
		t.Errorf("Expected VARCHAR(100) remember_token column, got: %s", token)
	}
}

func TestBlueprintCreateSQL(t *testing.T) {
	b := NewBlueprint()
	b.ID()
	b.String("name")
	b.String("email").Unique()
	b.Timestamps()

	sql := b.ToCreateSQL("users")

	expectedParts := []string{
		"CREATE TABLE `users`",
		"`id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT",
		"`name` VARCHAR(255) NOT NULL",
		"`email` VARCHAR(255) NOT NULL",
		"PRIMARY KEY (`id`)",
		"ENGINE=InnoDB",
	}

	for _, part := range expectedParts {
		if !strings.Contains(sql, part) {
			t.Errorf("Expected SQL to contain: %s\nGot: %s", part, sql)
		}
	}
}

func TestBlueprintIndexes(t *testing.T) {
	b := NewBlueprint()
	b.String("email")
	b.Index([]string{"email"})

	sql := b.ToCreateSQL("users")

	if !strings.Contains(sql, "INDEX `idx_email` (`email`)") {
		t.Errorf("Expected index definition in SQL, got: %s", sql)
	}
}

func TestBlueprintUniqueIndexes(t *testing.T) {
	b := NewBlueprint()
	b.String("email")
	b.Unique([]string{"email"})

	sql := b.ToCreateSQL("users")

	if !strings.Contains(sql, "UNIQUE INDEX `uniq_email` (`email`)") {
		t.Errorf("Expected unique index definition in SQL, got: %s", sql)
	}
}

func TestBlueprintCompositeIndexes(t *testing.T) {
	b := NewBlueprint()
	b.String("first_name")
	b.String("last_name")
	b.Index([]string{"first_name", "last_name"}, "idx_full_name")

	sql := b.ToCreateSQL("users")

	if !strings.Contains(sql, "INDEX `idx_full_name` (`first_name`, `last_name`)") {
		t.Errorf("Expected composite index definition in SQL, got: %s", sql)
	}
}

func TestBlueprintForeignKeys(t *testing.T) {
	b := NewBlueprint()
	b.ForeignId("user_id").Constrained().CascadeOnDelete()

	sql := b.ToCreateSQL("posts")

	expectedParts := []string{
		"FOREIGN KEY (`user_id`)",
		"REFERENCES `user`(`id`)",
		"ON DELETE CASCADE",
	}

	for _, part := range expectedParts {
		if !strings.Contains(sql, part) {
			t.Errorf("Expected SQL to contain: %s\nGot: %s", part, sql)
		}
	}
}

func TestBlueprintFullText(t *testing.T) {
	b := NewBlueprint()
	b.Text("title")
	b.Text("content")
	b.FullText([]string{"title", "content"})

	sql := b.ToCreateSQL("posts")

	if !strings.Contains(sql, "FULLTEXT INDEX `ft_title_content` (`title`, `content`)") {
		t.Errorf("Expected fulltext index definition in SQL, got: %s", sql)
	}
}

func TestBlueprintAlterSQL(t *testing.T) {
	b := NewBlueprint()
	b.String("new_column")
	b.DropColumn("old_column")
	b.RenameColumn("from_col", "to_col")

	sql := b.ToAlterSQL("users")

	expectedParts := []string{
		"ALTER TABLE `users` ADD COLUMN `new_column` VARCHAR(255) NOT NULL",
		"ALTER TABLE `users` DROP COLUMN `old_column`",
		"ALTER TABLE `users` RENAME COLUMN `from_col` TO `to_col`",
	}

	for _, part := range expectedParts {
		if !strings.Contains(sql, part) {
			t.Errorf("Expected SQL to contain: %s\nGot: %s", part, sql)
		}
	}
}

func TestBlueprintModifyColumn(t *testing.T) {
	b := NewBlueprint()
	b.ModifyColumn("name", "VARCHAR", 100)

	sql := b.ToAlterSQL("users")

	if !strings.Contains(sql, "ALTER TABLE `users` MODIFY COLUMN `name` VARCHAR(100) NOT NULL") {
		t.Errorf("Expected modify column statement, got: %s", sql)
	}
}

func TestSchemaCreate(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.Create("users", func(b *Blueprint) {
		b.ID()
		b.String("name")
		b.String("email")
		b.Timestamps()
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(mock.queries))
	}

	query := mock.queries[0]
	if !strings.Contains(query, "CREATE TABLE `users`") {
		t.Errorf("Expected CREATE TABLE statement, got: %s", query)
	}
}

func TestSchemaTable(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.Table("users", func(b *Blueprint) {
		b.String("new_field")
	})

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(mock.queries))
	}

	query := mock.queries[0]
	if !strings.Contains(query, "ALTER TABLE `users`") {
		t.Errorf("Expected ALTER TABLE statement, got: %s", query)
	}
}

func TestSchemaDrop(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.Drop("users")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if len(mock.queries) != 1 {
		t.Fatalf("Expected 1 query, got %d", len(mock.queries))
	}

	query := mock.queries[0]
	expected := "DROP TABLE `users`"
	if query != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, query)
	}
}

func TestSchemaDropIfExists(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.DropIfExists("users")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	query := mock.queries[0]
	expected := "DROP TABLE IF EXISTS `users`"
	if query != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, query)
	}
}

func TestSchemaRename(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.Rename("users", "customers")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	query := mock.queries[0]
	expected := "RENAME TABLE `users` TO `customers`"
	if query != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, query)
	}
}

func TestSchemaTruncate(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)

	err := schema.Truncate("users")

	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	query := mock.queries[0]
	expected := "TRUNCATE TABLE `users`"
	if query != expected {
		t.Errorf("Expected: %s\nGot: %s", expected, query)
	}
}

func TestSchemaHasTable(t *testing.T) {
	mock := &mockConnection{
		results: [][]map[string]interface{}{
			{{"count": int64(1)}},
		},
	}
	schema := NewSchema(mock)

	exists := schema.HasTable("users")

	if !exists {
		t.Error("Expected table to exist")
	}
}

func TestSchemaHasColumn(t *testing.T) {
	mock := &mockConnection{
		results: [][]map[string]interface{}{
			{{"count": int64(1)}},
		},
	}
	schema := NewSchema(mock)

	exists := schema.HasColumn("users", "email")

	if !exists {
		t.Error("Expected column to exist")
	}
}

func TestSchemaGetColumnType(t *testing.T) {
	mock := &mockConnection{
		results: [][]map[string]interface{}{
			{{"data_type": "varchar"}},
		},
	}
	schema := NewSchema(mock)

	dataType := schema.GetColumnType("users", "email")

	if dataType != "varchar" {
		t.Errorf("Expected varchar, got: %s", dataType)
	}
}

// Example migration for testing
type CreateUsersTable struct{}

func (m *CreateUsersTable) Up(schema *Schema) error {
	return schema.Create("users", func(b *Blueprint) {
		b.ID()
		b.String("name")
		b.String("email")
		b.Timestamps()
	})
}

func (m *CreateUsersTable) Down(schema *Schema) error {
	return schema.Drop("users")
}

func TestMigratorRegister(t *testing.T) {
	mock := &mockConnection{}
	schema := NewSchema(mock)
	migrator := NewMigrator(schema)

	migrator.Register("create_users_table", &CreateUsersTable{})

	if len(migrator.migrations) != 1 {
		t.Errorf("Expected 1 migration, got %d", len(migrator.migrations))
	}

	if len(migrator.order) != 1 {
		t.Errorf("Expected 1 migration in order, got %d", len(migrator.order))
	}
}

func TestColumnDefinitionChaining(t *testing.T) {
	b := NewBlueprint()
	col := b.String("email").
		Nullable().
		Default("test@example.com").
		Unique().
		Comment("User email address")

	sql := col.toSQL()

	expectedParts := []string{
		"`email`",
		"VARCHAR(255)",
		"NULL",
		"DEFAULT 'test@example.com'",
		"COMMENT 'User email address'",
	}

	for _, part := range expectedParts {
		if !strings.Contains(sql, part) {
			t.Errorf("Expected SQL to contain: %s\nGot: %s", part, sql)
		}
	}
}

// Benchmark tests
func BenchmarkBlueprintToCreateSQL(b *testing.B) {
	blueprint := NewBlueprint()
	blueprint.ID()
	blueprint.String("name")
	blueprint.String("email")
	blueprint.Timestamps()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = blueprint.ToCreateSQL("users")
	}
}

func BenchmarkColumnToSQL(b *testing.B) {
	col := newColumn("email", "VARCHAR")
	col.length = 255
	col.nullable = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = col.toSQL()
	}
}

// Example usage tests
func ExampleBlueprint_ToCreateSQL() {
	b := NewBlueprint()
	b.ID()
	b.String("name")
	b.String("email").Unique()
	b.Timestamps()

	sql := b.ToCreateSQL("users")
	fmt.Println(strings.Contains(sql, "CREATE TABLE `users`"))
	// Output: true
}

func ExampleBlueprint_Timestamps() {
	b := NewBlueprint()
	b.Timestamps()

	sql := b.ToCreateSQL("posts")
	containsCreatedAt := strings.Contains(sql, "created_at")
	containsUpdatedAt := strings.Contains(sql, "updated_at")
	fmt.Printf("%t %t", containsCreatedAt, containsUpdatedAt)
	// Output: true true
}
