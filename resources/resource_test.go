package resources

import (
	"testing"
)

type User struct {
	ID       int
	Name     string
	Email    string
	Password string
	IsAdmin  bool
}

type UserTransformer struct{}

func (t *UserTransformer) ToArray(user User) map[string]any {
	return CleanMap(map[string]any{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"admin": When(user.IsAdmin, true),
	})
}

func TestMakeTransformsSingleResource(t *testing.T) {
	user := User{
		ID:    1,
		Name:  "John Doe",
		Email: "john@example.com",
	}

	result := Make(user, &UserTransformer{})

	if result["id"] != 1 {
		t.Errorf("expected id 1, got %v", result["id"])
	}
	if result["name"] != "John Doe" {
		t.Errorf("expected name John Doe, got %v", result["name"])
	}
	if result["email"] != "john@example.com" {
		t.Errorf("expected email john@example.com, got %v", result["email"])
	}
	if _, exists := result["admin"]; exists {
		t.Error("admin field should not exist when IsAdmin is false")
	}
}

func TestCollectionTransformsList(t *testing.T) {
	users := []User{
		{ID: 1, Name: "John Doe", Email: "john@example.com"},
		{ID: 2, Name: "Jane Smith", Email: "jane@example.com"},
	}

	result := Collection(users, &UserTransformer{})

	if len(result) != 2 {
		t.Errorf("expected 2 items, got %d", len(result))
	}
	if result[0]["id"] != 1 {
		t.Errorf("expected first id 1, got %v", result[0]["id"])
	}
	if result[1]["id"] != 2 {
		t.Errorf("expected second id 2, got %v", result[1]["id"])
	}
}

func TestWhenIncludesExcludesConditionally(t *testing.T) {
	// Test true condition
	result := When(true, "included")
	if result == missing {
		t.Error("When(true) should not return missing sentinel")
	}
	if result != "included" {
		t.Errorf("expected 'included', got %v", result)
	}

	// Test false condition
	result = When(false, "excluded")
	if !IsMissing(result) {
		t.Error("When(false) should return missing sentinel")
	}
}

func TestWhenNotNilBehavior(t *testing.T) {
	// Test non-nil value
	value := "present"
	result := WhenNotNil(value)
	if IsMissing(result) {
		t.Error("WhenNotNil should not return missing for non-nil value")
	}
	if result != "present" {
		t.Errorf("expected 'present', got %v", result)
	}

	// Test nil value
	var nilValue *string
	result = WhenNotNil(nilValue)
	if !IsMissing(result) {
		t.Error("WhenNotNil should return missing for nil value")
	}
}

func TestUnlessBehavior(t *testing.T) {
	// Unless is inverse of When
	result := Unless(false, "included")
	if IsMissing(result) {
		t.Error("Unless(false) should not return missing sentinel")
	}

	result = Unless(true, "excluded")
	if !IsMissing(result) {
		t.Error("Unless(true) should return missing sentinel")
	}
}

func TestMergeWhenMergesConditionally(t *testing.T) {
	// Test true condition
	data := map[string]any{"key": "value"}
	result := MergeWhen(true, data)
	if len(result) != 1 {
		t.Errorf("expected 1 item in result, got %d", len(result))
	}
	if result["key"] != "value" {
		t.Errorf("expected key=value, got %v", result["key"])
	}

	// Test false condition
	result = MergeWhen(false, data)
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d items", len(result))
	}
}

func TestCleanMapRemovesMissing(t *testing.T) {
	data := map[string]any{
		"keep":   "value",
		"remove": missing,
		"also":   "keep",
	}

	result := CleanMap(data)

	if len(result) != 2 {
		t.Errorf("expected 2 items after cleaning, got %d", len(result))
	}
	if _, exists := result["remove"]; exists {
		t.Error("missing sentinel should be removed")
	}
	if result["keep"] != "value" {
		t.Error("non-missing values should be kept")
	}
}

func TestPaginatedResponseStructure(t *testing.T) {
	users := []User{
		{ID: 1, Name: "User 1", Email: "user1@example.com"},
		{ID: 2, Name: "User 2", Email: "user2@example.com"},
	}

	result := Paginate(users, &UserTransformer{}, 50, 10, 2)

	// Check data
	data, ok := result["data"].([]map[string]any)
	if !ok {
		t.Error("expected data to be []map[string]any")
	}
	if len(data) != 2 {
		t.Errorf("expected 2 items in data, got %d", len(data))
	}

	// Check meta
	meta, ok := result["meta"].(map[string]any)
	if !ok {
		t.Error("expected meta to be map[string]any")
	}
	if meta["total"] != 50 {
		t.Errorf("expected total 50, got %v", meta["total"])
	}
	if meta["per_page"] != 10 {
		t.Errorf("expected per_page 10, got %v", meta["per_page"])
	}
	if meta["current_page"] != 2 {
		t.Errorf("expected current_page 2, got %v", meta["current_page"])
	}
	if meta["last_page"] != 5 {
		t.Errorf("expected last_page 5, got %v", meta["last_page"])
	}

	// Check links
	links, ok := result["links"].(map[string]any)
	if !ok {
		t.Error("expected links to be map[string]any")
	}
	if links["first"] != "?page=1" {
		t.Errorf("expected first link ?page=1, got %v", links["first"])
	}
	if links["last"] != "?page=5" {
		t.Errorf("expected last link ?page=5, got %v", links["last"])
	}
	if links["prev"] != "?page=1" {
		t.Errorf("expected prev link ?page=1, got %v", links["prev"])
	}
	if links["next"] != "?page=3" {
		t.Errorf("expected next link ?page=3, got %v", links["next"])
	}
}

func TestPaginationFirstPage(t *testing.T) {
	users := []User{{ID: 1, Name: "User 1", Email: "user1@example.com"}}
	result := Paginate(users, &UserTransformer{}, 50, 10, 1)

	links := result["links"].(map[string]any)
	if links["prev"] != nil {
		t.Error("first page should have nil prev link")
	}
	if links["next"] != "?page=2" {
		t.Errorf("expected next link ?page=2, got %v", links["next"])
	}
}

func TestPaginationLastPage(t *testing.T) {
	users := []User{{ID: 1, Name: "User 1", Email: "user1@example.com"}}
	result := Paginate(users, &UserTransformer{}, 50, 10, 5)

	links := result["links"].(map[string]any)
	if links["next"] != nil {
		t.Error("last page should have nil next link")
	}
	if links["prev"] != "?page=4" {
		t.Errorf("expected prev link ?page=4, got %v", links["prev"])
	}
}

func TestConditionalAttributeInResource(t *testing.T) {
	// Test with admin user
	adminUser := User{ID: 1, Name: "Admin", Email: "admin@example.com", IsAdmin: true}
	adminResult := Make(adminUser, &UserTransformer{})

	if adminResult["admin"] != true {
		t.Error("admin field should be present and true for admin user")
	}

	// Test with regular user
	regularUser := User{ID: 2, Name: "User", Email: "user@example.com", IsAdmin: false}
	regularResult := Make(regularUser, &UserTransformer{})

	if _, exists := regularResult["admin"]; exists {
		t.Error("admin field should not exist for regular user")
	}
}

func TestCalculateLastPage(t *testing.T) {
	tests := []struct {
		total    int
		perPage  int
		expected int
	}{
		{50, 10, 5},
		{51, 10, 6},
		{49, 10, 5},
		{10, 10, 1},
		{0, 10, 1},
		{5, 10, 1},
		{100, 0, 1}, // Edge case: avoid division by zero
	}

	for _, tt := range tests {
		result := calculateLastPage(tt.total, tt.perPage)
		if result != tt.expected {
			t.Errorf("calculateLastPage(%d, %d) = %d, expected %d", tt.total, tt.perPage, result, tt.expected)
		}
	}
}
