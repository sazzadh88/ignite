package validation

import (
	"strings"
	"testing"
)

func TestRequired(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		rules   map[string]string
		wantErr bool
	}{
		{
			name:    "required field present",
			data:    map[string]any{"name": "John"},
			rules:   map[string]string{"name": "required"},
			wantErr: false,
		},
		{
			name:    "required field missing",
			data:    map[string]any{},
			rules:   map[string]string{"name": "required"},
			wantErr: true,
		},
		{
			name:    "required field empty string",
			data:    map[string]any{"name": ""},
			rules:   map[string]string{"name": "required"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(tt.data, tt.rules)
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestRequiredIf(t *testing.T) {
	data := map[string]any{
		"role": "admin",
	}
	rules := map[string]string{
		"password": "required_if:role,admin",
	}

	v := Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when required_if condition is met but field is missing")
	}

	data["password"] = "secret"
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when required_if condition is met and field is present")
	}
}

func TestRequiredWith(t *testing.T) {
	data := map[string]any{
		"email": "test@example.com",
	}
	rules := map[string]string{
		"password": "required_with:email",
	}

	v := Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when required_with field is present but target is missing")
	}

	data["password"] = "secret"
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when both fields are present")
	}
}

func TestRequiredWithout(t *testing.T) {
	data := map[string]any{}
	rules := map[string]string{
		"phone": "required_without:email",
	}

	v := Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when required_without field is absent and target is missing")
	}

	data["email"] = "test@example.com"
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when other field is present")
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid string", "hello", false},
		{"integer", 123, true},
		{"boolean", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "string"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestInteger(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"valid int", 123, false},
		{"valid int64", int64(123), false},
		{"string int", "123", false},
		{"float", 123.45, true},
		{"string", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "integer"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestNumeric(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"integer", 123, false},
		{"float", 123.45, false},
		{"string number", "123.45", false},
		{"string", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "numeric"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestBoolean(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"bool true", true, false},
		{"bool false", false, false},
		{"string true", "true", false},
		{"string false", "false", false},
		{"int 1", 1, false},
		{"int 0", 0, false},
		{"string hello", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "boolean"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestArray(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		wantErr bool
	}{
		{"slice", []string{"a", "b"}, false},
		{"array", [3]int{1, 2, 3}, false},
		{"string", "hello", true},
		{"int", 123, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "array"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"valid value", "red", "in:red,green,blue", false},
		{"invalid value", "yellow", "in:red,green,blue", true},
		{"int valid", 2, "in:1,2,3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestNotIn(t *testing.T) {
	v := Make(map[string]any{"color": "red"}, map[string]string{"color": "not_in:red,green"})
	if !v.Fails() {
		t.Error("Expected validation to fail for value in exclusion list")
	}

	v = Make(map[string]any{"color": "blue"}, map[string]string{"color": "not_in:red,green"})
	if v.Fails() {
		t.Error("Expected validation to pass for value not in exclusion list")
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"string valid", "hello", "min:3", false},
		{"string invalid", "hi", "min:3", true},
		{"int valid", 10, "min:5", false},
		{"int invalid", 3, "min:5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"string valid", "hi", "max:5", false},
		{"string invalid", "hello world", "max:5", true},
		{"int valid", 3, "max:5", false},
		{"int invalid", 10, "max:5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestBetween(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"string valid", "hello", "between:3,10", false},
		{"string too short", "hi", "between:3,10", true},
		{"string too long", "hello world", "between:3,10", true},
		{"int valid", 5, "between:3,10", false},
		{"int invalid", 15, "between:3,10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"string valid", "hello", "size:5", false},
		{"string invalid", "hi", "size:5", true},
		{"int valid", 10, "size:10", false},
		{"int invalid", 5, "size:10", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"valid email with name", "John Doe <john@example.com>", false},
		{"invalid email", "notanemail", true},
		{"invalid email missing @", "test.example.com", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"email": tt.value}, map[string]string{"email": "email"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid http", "http://example.com", false},
		{"valid https", "https://example.com/path", false},
		{"invalid no scheme", "example.com", true},
		{"invalid no host", "http://", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"url": tt.value}, map[string]string{"url": "url"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestIP(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid ipv4", "192.168.1.1", false},
		{"valid ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"invalid ip", "999.999.999.999", true},
		{"not an ip", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"ip": tt.value}, map[string]string{"ip": "ip"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestUUID(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid uuid", "not-a-uuid", true},
		{"invalid format", "550e8400-e29b-41d4-a716", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"id": tt.value}, map[string]string{"id": "uuid"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestRegex(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		rules   string
		wantErr bool
	}{
		{"valid pattern", "abc123", "regex:^[a-z0-9]+$", false},
		{"invalid pattern", "ABC123", "regex:^[a-z0-9]+$", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestNotRegex(t *testing.T) {
	v := Make(map[string]any{"field": "ABC123"}, map[string]string{"field": "not_regex:^[a-z0-9]+$"})
	if v.Fails() {
		t.Error("Expected validation to pass when value doesn't match pattern")
	}

	v = Make(map[string]any{"field": "abc123"}, map[string]string{"field": "not_regex:^[a-z0-9]+$"})
	if !v.Fails() {
		t.Error("Expected validation to fail when value matches pattern")
	}
}

func TestAlpha(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid alpha", "abcXYZ", false},
		{"invalid with numbers", "abc123", true},
		{"invalid with spaces", "abc xyz", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "alpha"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestAlphaNum(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid alphanum", "abc123XYZ", false},
		{"invalid with dash", "abc-123", true},
		{"invalid with spaces", "abc 123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "alpha_num"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestAlphaDash(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid with dash", "abc-123", false},
		{"valid with underscore", "abc_123", false},
		{"invalid with space", "abc 123", true},
		{"invalid with dot", "abc.123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": "alpha_dash"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestStartsWith(t *testing.T) {
	v := Make(map[string]any{"name": "hello world"}, map[string]string{"name": "starts_with:hello,hi"})
	if v.Fails() {
		t.Error("Expected validation to pass when value starts with one of the prefixes")
	}

	v = Make(map[string]any{"name": "goodbye world"}, map[string]string{"name": "starts_with:hello,hi"})
	if !v.Fails() {
		t.Error("Expected validation to fail when value doesn't start with any prefix")
	}
}

func TestEndsWith(t *testing.T) {
	v := Make(map[string]any{"file": "document.pdf"}, map[string]string{"file": "ends_with:.pdf,.doc"})
	if v.Fails() {
		t.Error("Expected validation to pass when value ends with one of the suffixes")
	}

	v = Make(map[string]any{"file": "document.txt"}, map[string]string{"file": "ends_with:.pdf,.doc"})
	if !v.Fails() {
		t.Error("Expected validation to fail when value doesn't end with any suffix")
	}
}

func TestContains(t *testing.T) {
	v := Make(map[string]any{"text": "hello world"}, map[string]string{"text": "contains:world"})
	if v.Fails() {
		t.Error("Expected validation to pass when value contains substring")
	}

	v = Make(map[string]any{"text": "hello world"}, map[string]string{"text": "contains:goodbye"})
	if !v.Fails() {
		t.Error("Expected validation to fail when value doesn't contain substring")
	}
}

func TestDate(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid rfc3339", "2023-01-01T00:00:00Z", false},
		{"valid date", "2023-01-01", false},
		{"valid datetime", "2023-01-01 15:04:05", false},
		{"invalid date", "not-a-date", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"date": tt.value}, map[string]string{"date": "date"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestAfter(t *testing.T) {
	data := map[string]any{
		"start_date": "2023-01-01",
		"end_date":   "2023-12-31",
	}
	rules := map[string]string{
		"end_date": "after:start_date",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when date is after other field")
	}

	data["end_date"] = "2022-12-31"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when date is before other field")
	}
}

func TestBefore(t *testing.T) {
	data := map[string]any{
		"start_date": "2023-01-01",
		"end_date":   "2023-12-31",
	}
	rules := map[string]string{
		"start_date": "before:end_date",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when date is before other field")
	}

	data["start_date"] = "2024-01-01"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when date is after other field")
	}
}

func TestConfirmed(t *testing.T) {
	data := map[string]any{
		"password":              "secret123",
		"password_confirmation": "secret123",
	}
	rules := map[string]string{
		"password": "confirmed",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when confirmation matches")
	}

	data["password_confirmation"] = "different"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when confirmation doesn't match")
	}
}

func TestSame(t *testing.T) {
	data := map[string]any{
		"password":         "secret123",
		"password_confirm": "secret123",
	}
	rules := map[string]string{
		"password_confirm": "same:password",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when fields match")
	}

	data["password_confirm"] = "different"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when fields don't match")
	}
}

func TestDifferent(t *testing.T) {
	data := map[string]any{
		"username": "john",
		"password": "secret123",
	}
	rules := map[string]string{
		"password": "different:username",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass when fields are different")
	}

	data["password"] = "john"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail when fields are the same")
	}
}

func TestComparisonRules(t *testing.T) {
	data := map[string]any{
		"min_value": 5,
		"max_value": 10,
	}

	// Test gt
	rules := map[string]string{"max_value": "gt:min_value"}
	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected gt validation to pass")
	}

	// Test gte
	data["max_value"] = 5
	rules = map[string]string{"max_value": "gte:min_value"}
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected gte validation to pass")
	}

	// Test lt - reset max_value
	data["min_value"] = 5
	data["max_value"] = 10
	rules = map[string]string{"min_value": "lt:max_value"}
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected lt validation to pass")
	}

	// Test lte
	data["min_value"] = 5
	data["max_value"] = 5
	rules = map[string]string{"min_value": "lte:max_value"}
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected lte validation to pass")
	}
}

func TestJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid json object", `{"key":"value"}`, false},
		{"valid json array", `[1,2,3]`, false},
		{"invalid json", `{invalid}`, true},
		{"not json", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"data": tt.value}, map[string]string{"data": "json"})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestDigits(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"valid digits", "12345", "digits:5", false},
		{"invalid length", "123", "digits:5", true},
		{"non-digits", "abc12", "digits:5", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestDigitsBetween(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"valid range", "12345", "digits_between:3,6", false},
		{"too short", "12", "digits_between:3,6", true},
		{"too long", "1234567", "digits_between:3,6", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := Make(map[string]any{"field": tt.value}, map[string]string{"field": tt.rules})
			if v.Fails() != tt.wantErr {
				t.Errorf("Fails() = %v, want %v", v.Fails(), tt.wantErr)
			}
		})
	}
}

func TestNullable(t *testing.T) {
	data := map[string]any{
		"optional": nil,
	}
	rules := map[string]string{
		"optional": "nullable|string",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass for nullable field with nil value")
	}

	data["optional"] = "value"
	v = Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass for nullable field with string value")
	}
}

func TestMultipleRules(t *testing.T) {
	data := map[string]any{
		"email": "test@example.com",
	}
	rules := map[string]string{
		"email": "required|string|email|max:255",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Errorf("Expected validation to pass, errors: %v", v.Errors().All())
	}

	data["email"] = "not-an-email"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail for invalid email")
	}
}

func TestCustomMessages(t *testing.T) {
	data := map[string]any{
		"name": "",
	}
	rules := map[string]string{
		"name": "required",
	}
	messages := map[string]string{
		"name.required": "Please provide your name",
	}

	v := Make(data, rules, messages)
	if !v.Fails() {
		t.Fatal("Expected validation to fail")
	}

	if v.Errors().First("name") != "Please provide your name" {
		t.Errorf("Expected custom message, got: %s", v.Errors().First("name"))
	}
}

func TestSetAttributeNames(t *testing.T) {
	data := map[string]any{
		"email": "",
	}
	rules := map[string]string{
		"email": "required",
	}

	// Create validator without running validation first
	v := &Validator{
		data:           data,
		rules:          rules,
		customMessages: make(map[string]string),
		attributeNames: map[string]string{
			"email": "email address",
		},
		errors:       NewErrorBag(),
		conditionals: make(map[string]conditionalRule),
	}
	v.validate()

	if !v.Fails() {
		t.Fatal("Expected validation to fail")
	}

	errorMsg := v.Errors().First("email")
	if errorMsg != "The email address field is required." {
		t.Errorf("Expected custom attribute name in message, got: %s", errorMsg)
	}
}

func TestValidated(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"extra": "should not be included",
	}
	rules := map[string]string{
		"name":  "required",
		"email": "required|email",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Fatal("Expected validation to pass")
	}

	validated := v.Validated()
	if len(validated) != 2 {
		t.Errorf("Expected 2 validated fields, got %d", len(validated))
	}

	if validated["name"] != "John" {
		t.Error("Expected name to be in validated data")
	}

	if validated["email"] != "john@example.com" {
		t.Error("Expected email to be in validated data")
	}

	if _, ok := validated["extra"]; ok {
		t.Error("Expected extra field to not be in validated data")
	}
}

func TestSometimes(t *testing.T) {
	data := map[string]any{
		"is_company": true,
	}
	rules := map[string]string{
		"is_company": "boolean",
	}

	// Create validator and set up Sometimes before validation
	v := &Validator{
		data:           data,
		rules:          rules,
		customMessages: make(map[string]string),
		attributeNames: make(map[string]string),
		errors:         NewErrorBag(),
		conditionals:   make(map[string]conditionalRule),
	}
	v.Sometimes("company_name", "required|string", func(data map[string]any) bool {
		isCompany, ok := data["is_company"].(bool)
		return ok && isCompany
	})
	v.validate()

	if !v.Fails() {
		t.Error("Expected validation to fail when conditional rule is applied and field is missing")
	}

	data["company_name"] = "Acme Inc"
	v = &Validator{
		data:           data,
		rules:          rules,
		customMessages: make(map[string]string),
		attributeNames: make(map[string]string),
		errors:         NewErrorBag(),
		conditionals:   make(map[string]conditionalRule),
	}
	v.Sometimes("company_name", "required|string", func(data map[string]any) bool {
		isCompany, ok := data["is_company"].(bool)
		return ok && isCompany
	})
	v.validate()

	if v.Fails() {
		t.Error("Expected validation to pass when conditional rule is applied and field is present")
	}
}

func TestCustomRuleExtension(t *testing.T) {
	// Register a custom rule that checks if value is uppercase
	Extend("uppercase", func(attribute string, value any, params []string, data map[string]any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return str == strings.ToUpper(str)
	}, "The :attribute must be uppercase.")

	data := map[string]any{
		"code": "HELLO",
	}
	rules := map[string]string{
		"code": "required|uppercase",
	}

	v := Make(data, rules)
	if v.Fails() {
		t.Errorf("Expected validation to pass for uppercase value, errors: %v", v.Errors().All())
	}

	data["code"] = "hello"
	v = Make(data, rules)
	if !v.Fails() {
		t.Error("Expected validation to fail for lowercase value")
	}

	if v.Errors().First("code") != "The code must be uppercase." {
		t.Errorf("Expected custom rule message, got: %s", v.Errors().First("code"))
	}
}

func TestErrorBag(t *testing.T) {
	eb := NewErrorBag()

	if !eb.IsEmpty() {
		t.Error("New ErrorBag should be empty")
	}

	eb.Add("email", "Email is required")
	eb.Add("email", "Email is invalid")
	eb.Add("password", "Password is required")

	if eb.IsEmpty() {
		t.Error("ErrorBag should not be empty after adding errors")
	}

	if eb.Count() != 2 {
		t.Errorf("Expected 2 fields with errors, got %d", eb.Count())
	}

	if !eb.Has("email") {
		t.Error("Expected email field to have errors")
	}

	if eb.Has("nonexistent") {
		t.Error("Expected nonexistent field to not have errors")
	}

	emailErrors := eb.Get("email")
	if len(emailErrors) != 2 {
		t.Errorf("Expected 2 errors for email, got %d", len(emailErrors))
	}

	firstError := eb.First("email")
	if firstError != "Email is required" {
		t.Errorf("Expected first error to be 'Email is required', got: %s", firstError)
	}

	all := eb.All()
	if len(all) != 2 {
		t.Errorf("Expected 2 fields in all errors, got %d", len(all))
	}

	jsonData, err := eb.ToJSON()
	if err != nil {
		t.Errorf("Failed to convert to JSON: %v", err)
	}
	if jsonData == nil {
		t.Error("Expected non-nil JSON data")
	}
}

// LoginRequest is a test implementation of FormRequest
type LoginRequest struct{}

func (r *LoginRequest) Rules() map[string]string {
	return map[string]string{
		"email":    "required|email",
		"password": "required|min:8",
	}
}

func (r *LoginRequest) Authorize() bool {
	return true
}

func (r *LoginRequest) Messages() map[string]string {
	return map[string]string{
		"email.required":    "Email is required",
		"password.required": "Password is required",
	}
}

func TestFormRequestInterface(t *testing.T) {
	// Test that a struct implementing FormRequest can be used
	req := &LoginRequest{}
	data := map[string]any{
		"email":    "test@example.com",
		"password": "secret123",
	}

	v := Make(data, req.Rules(), req.Messages())
	if v.Fails() {
		t.Error("Expected validation to pass for valid LoginRequest")
	}
}

func TestFacade(t *testing.T) {
	data := map[string]any{
		"name": "John",
	}
	rules := map[string]string{
		"name": "required|string",
	}

	v := Validate.Make(data, rules)
	if v.Fails() {
		t.Error("Expected validation to pass using facade")
	}
}
