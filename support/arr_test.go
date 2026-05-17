package support

import (
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"user": map[string]any{
			"email": "john@example.com",
			"profile": map[string]any{
				"age": 30,
			},
		},
	}

	tests := []struct {
		key      string
		want     any
		hasDefault bool
		defaultVal any
	}{
		{"name", "John", false, nil},
		{"user.email", "john@example.com", false, nil},
		{"user.profile.age", 30, false, nil},
		{"missing", nil, false, nil},
		{"missing", "default", true, "default"},
		{"user.missing", "default", true, "default"},
	}

	for _, tt := range tests {
		var got any
		if tt.hasDefault {
			got = Arr.Get(data, tt.key, tt.defaultVal)
		} else {
			got = Arr.Get(data, tt.key)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("Get(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestSet(t *testing.T) {
	data := make(map[string]any)
	data = Arr.Set(data, "name", "John")
	data = Arr.Set(data, "user.email", "john@example.com")
	data = Arr.Set(data, "user.profile.age", 30)

	if Arr.Get(data, "name") != "John" {
		t.Errorf("Set name failed")
	}
	if Arr.Get(data, "user.email") != "john@example.com" {
		t.Errorf("Set user.email failed")
	}
	if Arr.Get(data, "user.profile.age") != 30 {
		t.Errorf("Set user.profile.age failed")
	}
}

func TestHas(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"user": map[string]any{
			"email": "john@example.com",
		},
	}

	tests := []struct {
		key  string
		want bool
	}{
		{"name", true},
		{"user.email", true},
		{"missing", false},
		{"user.missing", false},
	}

	for _, tt := range tests {
		if got := Arr.Has(data, tt.key); got != tt.want {
			t.Errorf("Has(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestForget(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"user": map[string]any{
			"email": "john@example.com",
			"age":   30,
		},
	}

	data = Arr.Forget(data, "name")
	if Arr.Has(data, "name") {
		t.Errorf("Forget failed to remove 'name'")
	}

	data = Arr.Forget(data, "user.email")
	if Arr.Has(data, "user.email") {
		t.Errorf("Forget failed to remove 'user.email'")
	}
	if !Arr.Has(data, "user.age") {
		t.Errorf("Forget incorrectly removed 'user.age'")
	}
}

func TestOnly(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	result := Arr.Only(data, []string{"name", "email"})
	if len(result) != 2 {
		t.Errorf("Only returned wrong number of keys: %d, want 2", len(result))
	}
	if result["name"] != "John" || result["email"] != "john@example.com" {
		t.Errorf("Only returned wrong values")
	}
	if _, exists := result["age"]; exists {
		t.Errorf("Only should not include 'age'")
	}
}

func TestExcept(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	result := Arr.Except(data, []string{"age"})
	if len(result) != 2 {
		t.Errorf("Except returned wrong number of keys: %d, want 2", len(result))
	}
	if _, exists := result["age"]; exists {
		t.Errorf("Except should not include 'age'")
	}
	if result["name"] != "John" {
		t.Errorf("Except removed wrong keys")
	}
}

func TestDot(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"user": map[string]any{
			"email": "john@example.com",
			"profile": map[string]any{
				"age": 30,
			},
		},
	}

	result := Arr.Dot(data)
	if result["name"] != "John" {
		t.Errorf("Dot failed for 'name'")
	}
	if result["user.email"] != "john@example.com" {
		t.Errorf("Dot failed for 'user.email'")
	}
	if result["user.profile.age"] != 30 {
		t.Errorf("Dot failed for 'user.profile.age'")
	}
}

func TestUndot(t *testing.T) {
	data := map[string]any{
		"name":             "John",
		"user.email":       "john@example.com",
		"user.profile.age": 30,
	}

	result := Arr.Undot(data)
	if result["name"] != "John" {
		t.Errorf("Undot failed for 'name'")
	}
	if Arr.Get(result, "user.email") != "john@example.com" {
		t.Errorf("Undot failed for 'user.email'")
	}
	if Arr.Get(result, "user.profile.age") != 30 {
		t.Errorf("Undot failed for 'user.profile.age'")
	}
}

func TestFlatten(t *testing.T) {
	data := []any{
		1,
		[]any{2, 3},
		[]any{4, []any{5, 6}},
	}

	result := Arr.Flatten(data)
	expected := []any{1, 2, 3, 4, 5, 6}
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Flatten = %v, want %v", result, expected)
	}

	resultDepth := Arr.Flatten(data, 1)
	expectedDepth := []any{1, 2, 3, 4, []any{5, 6}}
	if !reflect.DeepEqual(resultDepth, expectedDepth) {
		t.Errorf("Flatten with depth = %v, want %v", resultDepth, expectedDepth)
	}
}

func TestKeys(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   30,
	}

	keys := Arr.Keys(data)
	if len(keys) != 3 {
		t.Errorf("Keys returned wrong count: %d, want 3", len(keys))
	}

	expected := []string{"age", "email", "name"}
	if !reflect.DeepEqual(keys, expected) {
		t.Errorf("Keys = %v, want %v", keys, expected)
	}
}

func TestValues(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": 2,
		"c": 3,
	}

	values := Arr.Values(data)
	if len(values) != 3 {
		t.Errorf("Values returned wrong count: %d, want 3", len(values))
	}
}

func TestPluck(t *testing.T) {
	items := []map[string]any{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
		{"name": "Bob"},
	}

	names := Arr.Pluck(items, "name")
	expected := []any{"John", "Jane", "Bob"}
	if !reflect.DeepEqual(names, expected) {
		t.Errorf("Pluck = %v, want %v", names, expected)
	}

	ages := Arr.Pluck(items, "age")
	expectedAges := []any{30, 25}
	if !reflect.DeepEqual(ages, expectedAges) {
		t.Errorf("Pluck ages = %v, want %v", ages, expectedAges)
	}
}

func TestWhere(t *testing.T) {
	items := []map[string]any{
		{"name": "John", "age": 30},
		{"name": "Jane", "age": 25},
		{"name": "Bob", "age": 35},
	}

	result := Arr.Where(items, func(item map[string]any) bool {
		age, ok := item["age"].(int)
		return ok && age >= 30
	})

	if len(result) != 2 {
		t.Errorf("Where returned wrong count: %d, want 2", len(result))
	}
}

func TestFirst(t *testing.T) {
	items := []any{1, 2, 3, 4, 5}

	result := Arr.First(items, func(item any) bool {
		return item.(int) > 3
	})

	if result != 4 {
		t.Errorf("First = %v, want 4", result)
	}

	result = Arr.First(items, func(item any) bool {
		return item.(int) > 10
	})

	if result != nil {
		t.Errorf("First = %v, want nil", result)
	}
}

func TestLast(t *testing.T) {
	items := []any{1, 2, 3, 4, 5}

	result := Arr.Last(items, func(item any) bool {
		return item.(int) < 4
	})

	if result != 3 {
		t.Errorf("Last = %v, want 3", result)
	}
}

func TestShuffle(t *testing.T) {
	items := []any{1, 2, 3, 4, 5}
	result := Arr.Shuffle(items)

	if len(result) != len(items) {
		t.Errorf("Shuffle changed array length")
	}

	original := make(map[any]bool)
	for _, item := range items {
		original[item] = true
	}

	for _, item := range result {
		if !original[item] {
			t.Errorf("Shuffle introduced new element: %v", item)
		}
	}
}

func TestRandom(t *testing.T) {
	items := []any{1, 2, 3, 4, 5}

	result := Arr.Random(items)
	if result == nil {
		t.Errorf("Random returned nil")
	}

	found := false
	for _, item := range items {
		if item == result {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Random returned element not in array")
	}

	multiResult := Arr.Random(items, 3)
	if multiResult == nil {
		t.Errorf("Random with count returned nil")
	}
	if arr, ok := multiResult.([]any); ok {
		if len(arr) != 3 {
			t.Errorf("Random with count returned wrong length: %d, want 3", len(arr))
		}
	} else {
		t.Errorf("Random with count didn't return array")
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		input any
		want  []any
	}{
		{nil, []any{}},
		{"hello", []any{"hello"}},
		{[]any{1, 2, 3}, []any{1, 2, 3}},
		{42, []any{42}},
	}

	for _, tt := range tests {
		result := Arr.Wrap(tt.input)
		if !reflect.DeepEqual(result, tt.want) {
			t.Errorf("Wrap(%v) = %v, want %v", tt.input, result, tt.want)
		}
	}
}

func TestQuery(t *testing.T) {
	data := map[string]string{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   "30",
	}

	result := Arr.Query(data)
	if result == "" {
		t.Errorf("Query returned empty string")
	}

	if !containsAll(result, []string{"name=John", "email=john", "age=30"}) {
		t.Errorf("Query = %q, missing expected parameters", result)
	}
}

func containsAll(str string, substrs []string) bool {
	for _, substr := range substrs {
		if !containsSubstr(str, substr) {
			return false
		}
	}
	return true
}

func containsSubstr(str, substr string) bool {
	return len(str) >= len(substr) && findSubstr(str, substr)
}

func findSubstr(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestSetNilMap(t *testing.T) {
	var data map[string]any
	data = Arr.Set(data, "name", "John")
	if data["name"] != "John" {
		t.Errorf("Set on nil map failed")
	}
}

func TestForgetNilMap(t *testing.T) {
	var data map[string]any
	result := Arr.Forget(data, "name")
	if result != nil {
		t.Errorf("Forget on nil map should return nil")
	}
}

func TestGetNestedNonMap(t *testing.T) {
	data := map[string]any{
		"name": "John",
		"age":  30,
	}

	result := Arr.Get(data, "age.invalid", "default")
	if result != "default" {
		t.Errorf("Get nested on non-map = %v, want default", result)
	}
}

func TestFlattenMap(t *testing.T) {
	data := map[string]any{
		"a": 1,
		"b": 2,
	}

	result := Arr.Flatten(data)
	if len(result) != 2 {
		t.Errorf("Flatten map = %d elements, want 2", len(result))
	}
}

func TestRandomEmptyArray(t *testing.T) {
	items := []any{}
	result := Arr.Random(items)
	if result != nil {
		t.Errorf("Random on empty array = %v, want nil", result)
	}
}

func TestRandomMoreThanLength(t *testing.T) {
	items := []any{1, 2, 3}
	result := Arr.Random(items, 10)
	if arr, ok := result.([]any); ok {
		if len(arr) != 3 {
			t.Errorf("Random with count > length = %d elements, want 3", len(arr))
		}
	}
}
