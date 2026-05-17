package orm

import (
	"encoding/json"
	"testing"
	"time"
)

// TestModel is a sample model for testing.
type TestModel struct {
	Model
	Name   string `db:"name"`
	Email  string `db:"email"`
	Age    int    `db:"age"`
	Active bool   `db:"active"`
}

func (t TestModel) Table() string {
	return "test_models"
}

func (t TestModel) Fillable() []string {
	return []string{"name", "email", "age", "active"}
}

func (t TestModel) Hidden() []string {
	return []string{"email"}
}

func (t TestModel) PrimaryKey() string {
	return "id"
}

// TestQueryBasicWhere tests basic WHERE clause generation.
func TestQueryBasicWhere(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Where("name", "=", "John")

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE name = ?"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 1 || bindings[0] != "John" {
		t.Errorf("Expected bindings: [John], got: %v", bindings)
	}
}

// TestQueryMultipleWhere tests multiple WHERE clauses.
func TestQueryMultipleWhere(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Where("name", "=", "John").Where("age", ">", 25)

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE name = ? AND age > ?"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 2 {
		t.Errorf("Expected 2 bindings, got: %d", len(bindings))
	}
}

// TestQueryOrWhere tests OR WHERE clauses.
func TestQueryOrWhere(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Where("name", "=", "John").OrWhere("name", "=", "Jane")

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE name = ? OR name = ?"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 2 {
		t.Errorf("Expected 2 bindings, got: %d", len(bindings))
	}
}

// TestQueryWhereIn tests WHERE IN clause.
func TestQueryWhereIn(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.WhereIn("id", []any{1, 2, 3})

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE id IN (?, ?, ?)"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 3 {
		t.Errorf("Expected 3 bindings, got: %d", len(bindings))
	}
}

// TestQueryWhereNull tests WHERE NULL clause.
func TestQueryWhereNull(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.WhereNull("deleted_at")

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE deleted_at IS NULL"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 0 {
		t.Errorf("Expected 0 bindings, got: %d", len(bindings))
	}
}

// TestQueryWhereNotNull tests WHERE NOT NULL clause.
func TestQueryWhereNotNull(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.WhereNotNull("email")

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE email IS NOT NULL"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 0 {
		t.Errorf("Expected 0 bindings, got: %d", len(bindings))
	}
}

// TestQueryOrderBy tests ORDER BY clause.
func TestQueryOrderBy(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.OrderBy("name", "asc").OrderBy("age", "desc")

	sql, _ := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models ORDER BY name ASC, age DESC"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}
}

// TestQueryLimit tests LIMIT clause.
func TestQueryLimit(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Limit(10)

	sql, _ := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models LIMIT 10"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}
}

// TestQueryOffset tests OFFSET clause.
func TestQueryOffset(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Limit(10).Offset(20)

	sql, _ := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models LIMIT 10 OFFSET 20"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}
}

// TestQuerySelect tests SELECT columns.
func TestQuerySelect(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Select("id", "name", "email")

	sql, _ := q.ToSQL()

	expectedSQL := "SELECT id, name, email FROM test_models"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}
}

// TestQueryComplex tests complex query building.
func TestQueryComplex(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")
	q.Select("id", "name").
		Where("active", "=", true).
		WhereIn("age", []any{25, 30, 35}).
		OrderBy("created_at", "desc").
		Limit(5).
		Offset(10)

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT id, name FROM test_models WHERE active = ? AND age IN (?, ?, ?) ORDER BY created_at DESC LIMIT 5 OFFSET 10"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 4 {
		t.Errorf("Expected 4 bindings, got: %d", len(bindings))
	}
}

// TestModelDirtyTracking tests dirty tracking on models.
func TestModelDirtyTracking(t *testing.T) {
	attrs := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	m := NewModelInstance(&TestModel{}, attrs)

	// Initially not dirty
	if m.IsDirty() {
		t.Error("Model should not be dirty initially")
	}

	// Set a value
	m.Set("name", "Jane")

	// Should be dirty now
	if !m.IsDirty() {
		t.Error("Model should be dirty after setting a value")
	}

	if !m.IsDirty("name") {
		t.Error("Model should be dirty for 'name' field")
	}

	if m.IsDirty("email") {
		t.Error("Model should not be dirty for 'email' field")
	}

	// Get dirty fields
	dirty := m.GetDirty()
	if len(dirty) != 1 {
		t.Errorf("Expected 1 dirty field, got: %d", len(dirty))
	}

	if dirty["name"] != "Jane" {
		t.Errorf("Expected dirty name to be 'Jane', got: %v", dirty["name"])
	}

	// Sync original
	m.SyncOriginal()

	if m.IsDirty() {
		t.Error("Model should not be dirty after sync")
	}
}

// TestModelToMap tests model to map conversion.
func TestModelToMap(t *testing.T) {
	attrs := map[string]any{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	m := NewModelInstance(&TestModel{}, attrs)
	result := m.ToMap(&TestModel{})

	// email should be hidden
	if _, exists := result["email"]; exists {
		t.Error("Email should be hidden")
	}

	if result["name"] != "John" {
		t.Error("Name should be present")
	}

	if result["age"] != 30 {
		t.Error("Age should be present")
	}
}

// TestModelToJSON tests model to JSON conversion.
func TestModelToJSON(t *testing.T) {
	attrs := map[string]any{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	m := NewModelInstance(&TestModel{}, attrs)
	jsonBytes, err := m.ToJSON(&TestModel{})

	if err != nil {
		t.Errorf("Error converting to JSON: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Errorf("Error parsing JSON: %v", err)
	}

	// email should be hidden
	if _, exists := result["email"]; exists {
		t.Error("Email should be hidden in JSON")
	}

	if result["name"] != "John" {
		t.Error("Name should be present in JSON")
	}
}

// TestCastJSON tests JSON casting.
func TestCastJSON(t *testing.T) {
	caster := JSONCast{}

	// Test Get
	input := `{"key": "value"}`
	result := caster.Get(input)

	if m, ok := result.(map[string]any); ok {
		if m["key"] != "value" {
			t.Errorf("Expected key=value, got: %v", m["key"])
		}
	} else {
		t.Error("Expected map[string]any")
	}

	// Test Set
	data := map[string]any{"key": "value"}
	output := caster.Set(data)

	if s, ok := output.(string); ok {
		var parsed map[string]any
		if err := json.Unmarshal([]byte(s), &parsed); err != nil {
			t.Errorf("Error parsing JSON: %v", err)
		}
		if parsed["key"] != "value" {
			t.Error("Expected key=value in JSON")
		}
	} else {
		t.Error("Expected string output")
	}
}

// TestCastBool tests bool casting.
func TestCastBool(t *testing.T) {
	caster := BoolCast{}

	// Test Get
	tests := []struct {
		input    any
		expected bool
	}{
		{true, true},
		{false, false},
		{1, true},
		{0, false},
		{"1", true},
		{"0", false},
		{"true", true},
		{"false", false},
	}

	for _, test := range tests {
		result := caster.Get(test.input)
		if result != test.expected {
			t.Errorf("Expected %v for input %v, got: %v", test.expected, test.input, result)
		}
	}

	// Test Set
	if caster.Set(true) != 1 {
		t.Error("Expected 1 for true")
	}

	if caster.Set(false) != 0 {
		t.Error("Expected 0 for false")
	}
}

// TestCastInt tests int casting.
func TestCastInt(t *testing.T) {
	caster := IntCast{}

	// Test Get
	if caster.Get(42) != 42 {
		t.Error("Expected 42")
	}

	if caster.Get(int64(42)) != 42 {
		t.Error("Expected 42 from int64")
	}

	if caster.Get("42") != 42 {
		t.Error("Expected 42 from string")
	}

	// Test Set
	if caster.Set(42) != 42 {
		t.Error("Expected 42 from Set")
	}
}

// TestCastFloat tests float casting.
func TestCastFloat(t *testing.T) {
	caster := FloatCast{}

	// Test Get
	if caster.Get(42.5) != 42.5 {
		t.Error("Expected 42.5")
	}

	if caster.Get(42) != 42.0 {
		t.Error("Expected 42.0 from int")
	}

	// Test Set
	if caster.Set(42.5) != 42.5 {
		t.Error("Expected 42.5 from Set")
	}
}

// TestCastArray tests array casting.
func TestCastArray(t *testing.T) {
	caster := ArrayCast{}

	// Test Get from JSON
	input := `["a", "b", "c"]`
	result := caster.Get(input)

	if arr, ok := result.([]string); ok {
		if len(arr) != 3 {
			t.Errorf("Expected 3 elements, got: %d", len(arr))
		}
		if arr[0] != "a" || arr[1] != "b" || arr[2] != "c" {
			t.Error("Unexpected array values")
		}
	} else {
		t.Error("Expected []string")
	}

	// Test Set
	data := []string{"a", "b", "c"}
	output := caster.Set(data)

	if s, ok := output.(string); ok {
		var parsed []string
		if err := json.Unmarshal([]byte(s), &parsed); err != nil {
			t.Errorf("Error parsing JSON: %v", err)
		}
		if len(parsed) != 3 {
			t.Error("Expected 3 elements in JSON")
		}
	} else {
		t.Error("Expected string output")
	}
}

// TestSoftDelete tests soft delete functionality.
func TestSoftDelete(t *testing.T) {
	model := &SoftDeleteModel{}

	// Initially not trashed
	if model.Trashed() {
		t.Error("Model should not be trashed initially")
	}

	// Soft delete
	now := time.Now()
	model.SetDeletedAt(&now)

	if !model.Trashed() {
		t.Error("Model should be trashed after soft delete")
	}

	if model.GetDeletedAt() == nil {
		t.Error("DeletedAt should not be nil")
	}

	// Restore
	model.SetDeletedAt(nil)

	if model.Trashed() {
		t.Error("Model should not be trashed after restore")
	}
}

// TestSoftDeleteScope tests soft delete scope application.
func TestSoftDeleteScope(t *testing.T) {
	q := NewQuery[TestModel]().Table("test_models")

	// Apply soft delete scope
	scope := SoftDeleteScope[TestModel]()
	q = scope(q)

	sql, _ := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE deleted_at IS NULL"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}
}

// TestObserverRegistration tests observer registration.
func TestObserverRegistration(t *testing.T) {
	registry := NewObserverRegistry()

	observer := &TestObserver{}
	model := &TestModel{}

	registry.Register(model, observer)

	observers := registry.GetObservers(model)
	if len(observers) != 1 {
		t.Errorf("Expected 1 observer, got: %d", len(observers))
	}
}

// TestObserver is a test observer.
type TestObserver struct {
	CreatingCalled bool
	CreatedCalled  bool
	UpdatingCalled bool
	UpdatedCalled  bool
}

func (o *TestObserver) Creating(model any) error {
	o.CreatingCalled = true
	return nil
}

func (o *TestObserver) Created(model any) error {
	o.CreatedCalled = true
	return nil
}

func (o *TestObserver) Updating(model any) error {
	o.UpdatingCalled = true
	return nil
}

func (o *TestObserver) Updated(model any) error {
	o.UpdatedCalled = true
	return nil
}

// TestObserverFiring tests observer event firing.
func TestObserverFiring(t *testing.T) {
	registry := NewObserverRegistry()

	observer := &TestObserver{}
	model := &TestModel{}

	registry.Register(model, observer)

	// Fire creating event
	if err := registry.FireEvent("creating", model); err != nil {
		t.Errorf("Error firing creating event: %v", err)
	}

	if !observer.CreatingCalled {
		t.Error("Creating method should have been called")
	}

	// Fire created event
	if err := registry.FireEvent("created", model); err != nil {
		t.Errorf("Error firing created event: %v", err)
	}

	if !observer.CreatedCalled {
		t.Error("Created method should have been called")
	}
}

// TestScopeApplication tests scope application to queries.
func TestScopeApplication(t *testing.T) {
	activeScope := func(q *Query[TestModel]) *Query[TestModel] {
		return q.Where("active", "=", true)
	}

	q := NewQuery[TestModel]().Table("test_models")
	q = activeScope(q)

	sql, bindings := q.ToSQL()

	expectedSQL := "SELECT * FROM test_models WHERE active = ?"
	if sql != expectedSQL {
		t.Errorf("Expected SQL: %s, got: %s", expectedSQL, sql)
	}

	if len(bindings) != 1 || bindings[0] != true {
		t.Errorf("Expected bindings: [true], got: %v", bindings)
	}
}

// TestRelationshipCreation tests relationship struct creation.
func TestRelationshipCreation(t *testing.T) {
	// HasOne
	hasOne := NewHasOne[TestModel]("user_id", "id")
	if hasOne.ForeignKey != "user_id" {
		t.Error("HasOne ForeignKey not set correctly")
	}
	if hasOne.LocalKey != "id" {
		t.Error("HasOne LocalKey not set correctly")
	}

	// HasMany
	hasMany := NewHasMany[TestModel]("user_id", "id")
	if hasMany.ForeignKey != "user_id" {
		t.Error("HasMany ForeignKey not set correctly")
	}

	// BelongsTo
	belongsTo := NewBelongsTo[TestModel]("user_id", "id")
	if belongsTo.ForeignKey != "user_id" {
		t.Error("BelongsTo ForeignKey not set correctly")
	}

	// BelongsToMany
	belongsToMany := NewBelongsToMany[TestModel]("user_role", "user_id", "role_id")
	if belongsToMany.Table != "user_role" {
		t.Error("BelongsToMany Table not set correctly")
	}
}
