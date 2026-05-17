# Validation Package

Laravel-inspired request validation for GoFrame.

## Features

- **Zero Dependencies**: Uses only Go standard library
- **40+ Built-in Rules**: Comprehensive validation rules covering most use cases
- **Custom Rules**: Extend with custom validation logic
- **Multiple Rule Syntax**: Pipe-separated rules like Laravel
- **Custom Messages**: Override default error messages
- **Conditional Validation**: Apply rules conditionally with `Sometimes()`
- **FormRequest Interface**: Create self-validating request objects
- **Thread-safe Error Handling**: Concurrent-safe error bag
- **Nested Field Support**: Validate nested data structures

## Quick Start

```go
import "github.com/sazzad/goframe/validation"

// Basic validation
data := map[string]any{
    "email":    "john@example.com",
    "password": "secret123",
    "age":      25,
}

rules := map[string]string{
    "email":    "required|email|max:255",
    "password": "required|min:8",
    "age":      "required|integer|between:18,100",
}

v := validation.Make(data, rules)

if v.Fails() {
    // Get all errors
    errors := v.Errors().All()
    
    // Get first error for a field
    emailError := v.Errors().First("email")
    
    // Convert to JSON
    jsonErrors, _ := v.Errors().ToJSON()
}

if v.Passes() {
    // Get only validated fields
    validated := v.Validated()
}
```

## Built-in Validation Rules

### Basic Type Rules

- `required` - Field must be present and not empty
- `nullable` - Field can be nil
- `string` - Must be a string
- `integer` - Must be an integer
- `numeric` - Must be numeric (int or float)
- `boolean` - Must be boolean
- `array` - Must be slice or array

### Conditional Required Rules

- `required_if:field,value` - Required when other field equals value
- `required_with:field` - Required when other field is present
- `required_without:field` - Required when other field is absent

### Value Constraints

- `in:val1,val2,...` - Must be one of the listed values
- `not_in:val1,val2,...` - Must not be one of the listed values
- `min:n` - Minimum length (string) or value (numeric)
- `max:n` - Maximum length (string) or value (numeric)
- `between:min,max` - Must be between min and max
- `size:n` - Exact length (string) or value (numeric)

### String Format Rules

- `email` - Valid email address
- `url` - Valid URL with scheme and host
- `ip` - Valid IPv4 or IPv6 address
- `uuid` - Valid UUID format
- `alpha` - Only letters (a-zA-Z)
- `alpha_num` - Letters and numbers
- `alpha_dash` - Letters, numbers, dashes, underscores
- `starts_with:val1,val2` - Must start with one of the values
- `ends_with:val1,val2` - Must end with one of the values
- `contains:val` - Must contain substring
- `regex:pattern` - Must match regex pattern
- `not_regex:pattern` - Must not match regex pattern
- `digits:n` - Exactly n digits
- `digits_between:min,max` - Between min and max digits

### Date/Time Rules

- `date` - Valid parseable date
- `after:date_or_field` - Must be after specified date or field value
- `before:date_or_field` - Must be before specified date or field value

Supported date formats:
- RFC3339: `2006-01-02T15:04:05Z07:00`
- Date: `2006-01-02`
- DateTime: `2006-01-02 15:04:05`
- US: `01/02/2006`
- UK: `02/01/2006`

### Field Comparison Rules

- `confirmed` - Field must match `field_confirmation`
- `same:field` - Must equal other field
- `different:field` - Must differ from other field
- `gt:field` - Greater than other field
- `gte:field` - Greater than or equal to other field
- `lt:field` - Less than other field
- `lte:field` - Less than or equal to other field

### Data Format Rules

- `json` - Valid JSON string

### Database Rules (Placeholder)

- `unique:table,column` - Returns true (implement DB check in app)
- `exists:table,column` - Returns true (implement DB check in app)

## Custom Error Messages

```go
data := map[string]any{"email": ""}
rules := map[string]string{"email": "required|email"}

// Custom messages: field.rule => message
messages := map[string]string{
    "email.required": "Please provide your email address",
    "email.email":    "Email format is invalid",
}

v := validation.Make(data, rules, messages)
```

Message placeholders:
- `:attribute` - Field name
- `:value` - Field value
- `:min` - Min parameter
- `:max` - Max parameter
- `:size` - Size parameter
- `:other` - Other field name
- `:values` - Comma-separated list of values

## Custom Attribute Names

```go
v := validation.Make(data, rules)
v.SetAttributeNames(map[string]string{
    "email": "email address",
    "dob":   "date of birth",
})
```

Note: Set attribute names before calling `Make()` to apply them to error messages.

## Conditional Validation

```go
data := map[string]any{
    "role": "admin",
}

rules := map[string]string{
    "role": "required|in:admin,user",
}

v := &validation.Validator{
    Data:           data,
    Rules:          rules,
    CustomMessages: make(map[string]string),
    AttributeNames: make(map[string]string),
    Errors:         validation.NewErrorBag(),
    Conditionals:   make(map[string]validation.conditionalRule),
}

// Only validate admin_key if role is admin
v.Sometimes("admin_key", "required|string", func(data map[string]any) bool {
    return data["role"] == "admin"
})

v.Validate()
```

## Custom Validation Rules

```go
// Register a custom rule
validation.Extend("uppercase", func(attribute string, value any, params []string, data map[string]any) bool {
    str, ok := value.(string)
    if !ok {
        return false
    }
    return str == strings.ToUpper(str)
}, "The :attribute must be uppercase.")

// Use the custom rule
rules := map[string]string{
    "code": "required|uppercase",
}
```

## FormRequest Interface

Create self-validating request objects:

```go
type CreateUserRequest struct {
    Email    string
    Password string
    Age      int
}

func (r *CreateUserRequest) Rules() map[string]string {
    return map[string]string{
        "email":    "required|email|max:255",
        "password": "required|min:8",
        "age":      "required|integer|between:18,100",
    }
}

func (r *CreateUserRequest) Authorize() bool {
    // Check if user is authorized to make this request
    return true
}

func (r *CreateUserRequest) Messages() map[string]string {
    return map[string]string{
        "email.required": "Email is required",
        "email.email":    "Please provide a valid email",
    }
}

// Usage
req := &CreateUserRequest{}
data := extractDataFromHTTPRequest() // Your data extraction logic

v := validation.Make(data, req.Rules(), req.Messages())
if !req.Authorize() {
    // Handle authorization failure
}
```

## Error Bag API

```go
errors := v.Errors()

// Check if field has errors
if errors.Has("email") {
    // ...
}

// Get all errors for a field
emailErrors := errors.Get("email")

// Get first error for a field
firstError := errors.First("email")

// Get all errors
all := errors.All() // map[string][]string

// Convert to JSON
jsonBytes, err := errors.ToJSON()

// Check if empty
if errors.IsEmpty() {
    // No errors
}

// Count fields with errors
count := errors.Count()
```

## Package Facade

```go
// Use the package-level facade
v := validation.Validate.Make(data, rules)
```

## Examples

### User Registration

```go
data := map[string]any{
    "email":                "john@example.com",
    "password":             "secret123",
    "password_confirmation": "secret123",
    "age":                  25,
    "terms":                true,
}

rules := map[string]string{
    "email":    "required|email|max:255",
    "password": "required|min:8|confirmed",
    "age":      "required|integer|between:18,100",
    "terms":    "required|boolean",
}

messages := map[string]string{
    "email.required":    "Email is required",
    "password.min":      "Password must be at least 8 characters",
    "password.confirmed": "Passwords do not match",
    "terms.required":    "You must accept the terms",
}

v := validation.Make(data, rules, messages)

if v.Passes() {
    validated := v.Validated()
    // Create user with validated data
} else {
    // Return errors to user
    errors := v.Errors().ToMap()
}
```

### API Request Validation

```go
func CreateProduct(w http.ResponseWriter, r *http.Request) {
    var requestData map[string]any
    json.NewDecoder(r.Body).Decode(&requestData)
    
    rules := map[string]string{
        "name":        "required|string|max:100",
        "price":       "required|numeric|min:0",
        "sku":         "required|alpha_dash|unique:products,sku",
        "category":    "required|in:electronics,clothing,food",
        "launch_date": "nullable|date|after:2024-01-01",
    }
    
    v := validation.Make(requestData, rules)
    
    if v.Fails() {
        w.WriteHeader(http.StatusUnprocessableEntity)
        json.NewEncoder(w).Encode(map[string]any{
            "errors": v.Errors().All(),
        })
        return
    }
    
    validated := v.Validated()
    // Create product with validated data
}
```

### Conditional Business Logic

```go
data := map[string]any{
    "shipping_method": "express",
    "insurance":       true,
}

rules := map[string]string{
    "shipping_method": "required|in:standard,express,overnight",
}

v := &validation.Validator{
    Data:           data,
    Rules:          rules,
    CustomMessages: make(map[string]string),
    AttributeNames: make(map[string]string),
    Errors:         validation.NewErrorBag(),
    Conditionals:   make(map[string]validation.conditionalRule),
}

// Require tracking number for express/overnight
v.Sometimes("tracking_number", "required|alpha_num", func(data map[string]any) bool {
    method := data["shipping_method"]
    return method == "express" || method == "overnight"
})

// Require insurance details if insurance is selected
v.Sometimes("insurance_provider", "required|string", func(data map[string]any) bool {
    insurance, ok := data["insurance"].(bool)
    return ok && insurance
})

v.Validate()
```

## Thread Safety

The ErrorBag is thread-safe and can be used concurrently. All read and write operations are protected by RWMutex.

## Performance

Validation is fast and efficient:
- Rules are parsed once during validation setup
- Built-in rules use optimized standard library functions
- No reflection except for type checking

## Design Philosophy

This package follows Laravel's validation API for familiarity while maintaining Go idioms:

- Zero external dependencies
- Explicit error handling
- Type safety where possible
- Extensibility through interfaces
- Thread-safe by default

## Extending with Database Rules

The `unique` and `exists` rules are placeholders. Implement them in your application:

```go
validation.Extend("unique", func(attribute string, value any, params []string, data map[string]any) bool {
    if len(params) < 2 {
        return true
    }
    table := params[0]
    column := params[1]
    
    // Query your database
    var count int
    err := db.QueryRow(
        "SELECT COUNT(*) FROM "+table+" WHERE "+column+" = ?", 
        value,
    ).Scan(&count)
    
    return err == nil && count == 0
}, "The :attribute has already been taken.")
```

## Testing

Run tests:

```bash
go test ./validation
```

All 40+ validation rules are covered with comprehensive tests including:
- Individual rule validation
- Multiple rules combined
- Custom messages
- Custom rules
- Error bag operations
- FormRequest interface
- Conditional validation

## Contributing

When adding new rules:
1. Add rule function to `builtInRules` map in `rules.go`
2. Add default message to `defaultMessages` map
3. Add comprehensive tests in `validation_test.go`
4. Update this README

## License

Part of the GoFrame project.
