package orm_test

import (
	"fmt"
	"time"

	"github.com/sazzadh88/ignite/orm"
)

// User model with soft deletes.
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

// Post model.
type Post struct {
	orm.Model
	UserID  uint64 `db:"user_id"`
	Title   string `db:"title"`
	Content string `db:"content"`
	Status  string `db:"status"`
}

func (p Post) Table() string {
	return "posts"
}

func (p Post) Fillable() []string {
	return []string{"user_id", "title", "content", "status"}
}

func (p Post) Hidden() []string {
	return []string{}
}

func (p Post) PrimaryKey() string {
	return "id"
}

// ExampleQuery demonstrates basic query building.
func ExampleQuery() {
	query := orm.NewQuery[User]().Table("users")

	// Simple WHERE
	query.Where("active", "=", true)

	// Get SQL
	sql, bindings := query.ToSQL()

	fmt.Println(sql)
	fmt.Println(bindings)
	// Output:
	// SELECT * FROM users WHERE active = ?
	// [true]
}

// ExampleQuery_multipleWhere demonstrates multiple WHERE clauses.
func ExampleQuery_multipleWhere() {
	query := orm.NewQuery[User]().Table("users")

	query.Where("active", "=", true).
		Where("email", "LIKE", "%@example.com")

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users WHERE active = ? AND email LIKE ?
}

// ExampleQuery_whereIn demonstrates WHERE IN clause.
func ExampleQuery_whereIn() {
	query := orm.NewQuery[User]().Table("users")

	query.WhereIn("id", []any{1, 2, 3, 4, 5})

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users WHERE id IN (?, ?, ?, ?, ?)
}

// ExampleQuery_orderBy demonstrates ordering results.
func ExampleQuery_orderBy() {
	query := orm.NewQuery[User]().Table("users")

	query.OrderBy("created_at", "desc").
		OrderBy("name", "asc")

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users ORDER BY created_at DESC, name ASC
}

// ExampleQuery_pagination demonstrates pagination.
func ExampleQuery_pagination() {
	query := orm.NewQuery[User]().Table("users")

	query.Limit(10).Offset(20)

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users LIMIT 10 OFFSET 20
}

// ExampleQuery_complex demonstrates complex query building.
func ExampleQuery_complex() {
	query := orm.NewQuery[Post]().Table("posts")

	query.Select("id", "title", "created_at").
		Where("status", "=", "published").
		WhereNull("deleted_at").
		OrderBy("created_at", "desc").
		Limit(10)

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT id, title, created_at FROM posts WHERE status = ? AND deleted_at IS NULL ORDER BY created_at DESC LIMIT 10
}

// ExampleModelInstance demonstrates dirty tracking.
func ExampleModelInstance() {
	attrs := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
	}

	model := orm.NewModelInstance(&User{}, attrs)

	// Check if dirty
	fmt.Println("Initially dirty:", model.IsDirty())

	// Set a value
	model.Set("name", "Jane Doe")

	fmt.Println("After set dirty:", model.IsDirty())
	fmt.Println("Name dirty:", model.IsDirty("name"))
	fmt.Println("Email dirty:", model.IsDirty("email"))

	// Output:
	// Initially dirty: false
	// After set dirty: true
	// Name dirty: true
	// Email dirty: false
}

// ExampleModelInstance_toMap demonstrates conversion to map.
func ExampleModelInstance_toMap() {
	attrs := map[string]any{
		"id":       1,
		"name":     "John Doe",
		"email":    "john@example.com",
		"password": "secret123",
	}

	model := orm.NewModelInstance(&User{}, attrs)
	result := model.ToMap(&User{})

	// Password is hidden
	fmt.Println("Has password:", result["password"] != nil)
	fmt.Println("Has name:", result["name"] != nil)
	fmt.Println("Has email:", result["email"] != nil)

	// Output:
	// Has password: false
	// Has name: true
	// Has email: true
}

// ExampleJSONCast demonstrates JSON casting.
func ExampleJSONCast() {
	caster := orm.JSONCast{}

	// Get: JSON string to Go value
	input := `{"key": "value", "count": 42}`
	result := caster.Get(input)

	if m, ok := result.(map[string]any); ok {
		fmt.Println("Key:", m["key"])
		fmt.Println("Count:", m["count"])
	}

	// Output:
	// Key: value
	// Count: 42
}

// ExampleBoolCast demonstrates bool casting.
func ExampleBoolCast() {
	caster := orm.BoolCast{}

	// Get: various inputs to bool
	fmt.Println("true:", caster.Get(true))
	fmt.Println("1:", caster.Get(1))
	fmt.Println("'1':", caster.Get("1"))
	fmt.Println("'true':", caster.Get("true"))
	fmt.Println("0:", caster.Get(0))

	// Set: bool to database value
	fmt.Println("true ->", caster.Set(true))
	fmt.Println("false ->", caster.Set(false))

	// Output:
	// true: true
	// 1: true
	// '1': true
	// 'true': true
	// 0: false
	// true -> 1
	// false -> 0
}

// ExampleSoftDeleteModel demonstrates soft deletes.
func ExampleSoftDeleteModel() {
	model := &orm.SoftDeleteModel{}

	fmt.Println("Initially trashed:", model.Trashed())

	// Soft delete
	now := time.Now()
	model.SetDeletedAt(&now)

	fmt.Println("After delete trashed:", model.Trashed())

	// Restore
	model.SetDeletedAt(nil)

	fmt.Println("After restore trashed:", model.Trashed())

	// Output:
	// Initially trashed: false
	// After delete trashed: true
	// After restore trashed: false
}

// ExampleSoftDeleteScope demonstrates soft delete query filtering.
func ExampleSoftDeleteScope() {
	query := orm.NewQuery[User]().Table("users")

	// Apply soft delete scope
	scope := orm.SoftDeleteScope[User]()
	query = scope(query)

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users WHERE deleted_at IS NULL
}

// UserObserver observes User model events.
type UserObserver struct{}

// Creating is called before a user is created.
func (o *UserObserver) Creating(model any) error {
	return nil
}

// ExampleObserver demonstrates model observers.
func ExampleObserver() {
	// Register observer
	orm.Observe(&User{}, &UserObserver{})

	fmt.Println("Observer registered")
	// Output:
	// Observer registered
}

// ExampleScope demonstrates query scopes.
func ExampleScope() {
	// Define a scope
	activeScope := func(q *orm.Query[User]) *orm.Query[User] {
		return q.Where("active", "=", true)
	}

	query := orm.NewQuery[User]().Table("users")
	query = activeScope(query)

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users WHERE active = ?
}

// ExampleWhereActive demonstrates the built-in active scope.
func ExampleWhereActive() {
	query := orm.NewQuery[User]().Table("users")

	scope := orm.WhereActive[User]()
	query = scope(query)

	sql, _ := query.ToSQL()

	fmt.Println(sql)
	// Output:
	// SELECT * FROM users WHERE active = ?
}

// ExampleHasMany demonstrates relationship creation.
func ExampleHasMany() {
	relation := orm.NewHasMany[Post]("user_id", "id")

	fmt.Println("Foreign key:", relation.ForeignKey)
	fmt.Println("Local key:", relation.LocalKey)

	// Output:
	// Foreign key: user_id
	// Local key: id
}

// ExampleBelongsTo demonstrates inverse relationship.
func ExampleBelongsTo() {
	relation := orm.NewBelongsTo[User]("user_id", "id")

	fmt.Println("Foreign key:", relation.ForeignKey)
	fmt.Println("Owner key:", relation.OwnerKey)

	// Output:
	// Foreign key: user_id
	// Owner key: id
}

// ExampleBelongsToMany demonstrates many-to-many relationship.
func ExampleBelongsToMany() {
	relation := orm.NewBelongsToMany[User]("role_user", "user_id", "role_id")

	fmt.Println("Pivot table:", relation.Table)
	fmt.Println("Foreign pivot key:", relation.ForeignPivotKey)
	fmt.Println("Related pivot key:", relation.RelatedPivotKey)

	// Output:
	// Pivot table: role_user
	// Foreign pivot key: user_id
	// Related pivot key: role_id
}

// ExampleCasts demonstrates attribute casting.
func ExampleCasts() {
	casts := orm.Casts{
		"settings": orm.JSONCast{},
		"age":      orm.IntCast{},
		"active":   orm.BoolCast{},
	}

	// Apply casts
	attrs := map[string]any{
		"settings": `{"theme": "dark"}`,
		"age":      "25",
		"active":   1,
	}

	result := orm.ApplyCasts(casts, attrs, true)

	fmt.Printf("Settings type: %T\n", result["settings"])
	fmt.Printf("Age type: %T\n", result["age"])
	fmt.Printf("Active type: %T\n", result["active"])

	// Output:
	// Settings type: map[string]interface {}
	// Age type: int
	// Active type: bool
}
