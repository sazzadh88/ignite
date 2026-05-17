package validation_test

import (
	"fmt"
	"strings"

	"github.com/sazzad/ignite/validation"
)

func ExampleMake() {
	data := map[string]any{
		"email": "john@example.com",
		"age":   25,
	}

	rules := map[string]string{
		"email": "required|email",
		"age":   "required|integer|min:18",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("Validation passed")
	}
	// Output: Validation passed
}

func ExampleMake_withErrors() {
	data := map[string]any{
		"email": "invalid-email",
	}

	rules := map[string]string{
		"email": "required|email",
	}

	v := validation.Make(data, rules)

	if v.Fails() {
		fmt.Println("Validation failed:")
		fmt.Printf("  email: %s\n", v.Errors().First("email"))
	}
	// Output:
	// Validation failed:
	//   email: The email must be a valid email address.
}

func ExampleMake_customMessages() {
	data := map[string]any{
		"username": "",
	}

	rules := map[string]string{
		"username": "required|alpha_dash|min:3",
	}

	messages := map[string]string{
		"username.required": "Please choose a username",
		"username.min":      "Username must be at least 3 characters",
	}

	v := validation.Make(data, rules, messages)

	if v.Fails() {
		fmt.Println(v.Errors().First("username"))
	}
	// Output: Please choose a username
}

func ExampleValidator_Validated() {
	data := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"extra": "this should not be included",
	}

	rules := map[string]string{
		"name":  "required|string",
		"email": "required|email",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		validated := v.Validated()
		fmt.Printf("Validated fields: %d\n", len(validated))
		fmt.Printf("Has name: %v\n", validated["name"] != nil)
		fmt.Printf("Has extra: %v\n", validated["extra"] != nil)
	}
	// Output:
	// Validated fields: 2
	// Has name: true
	// Has extra: false
}

func ExampleExtend() {
	// Register a custom validation rule
	validation.Extend("uppercase", func(attribute string, value any, params []string, data map[string]any) bool {
		str, ok := value.(string)
		if !ok {
			return false
		}
		return str == strings.ToUpper(str)
	}, "The :attribute must be uppercase.")

	data := map[string]any{
		"country_code": "US",
	}

	rules := map[string]string{
		"country_code": "required|uppercase",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("Country code is uppercase")
	}
	// Output: Country code is uppercase
}

func ExampleErrorBag_First() {
	data := map[string]any{
		"email": "",
	}

	rules := map[string]string{
		"email": "required|email",
	}

	v := validation.Make(data, rules)

	if v.Fails() {
		// Get the first error message for the email field
		fmt.Println(v.Errors().First("email"))
	}
	// Output: The email field is required.
}

func ExampleErrorBag_All() {
	data := map[string]any{
		"email": "",
		"age":   "",
	}

	rules := map[string]string{
		"email": "required|email",
		"age":   "required|integer",
	}

	v := validation.Make(data, rules)

	if v.Fails() {
		allErrors := v.Errors().All()
		fmt.Printf("Total fields with errors: %d\n", len(allErrors))
		fmt.Printf("Email has errors: %v\n", v.Errors().Has("email"))
		fmt.Printf("Age has errors: %v\n", v.Errors().Has("age"))
	}
	// Output:
	// Total fields with errors: 2
	// Email has errors: true
	// Age has errors: true
}

func ExampleValidator_SetAttributeNames() {
	data := map[string]any{
		"dob": "2000-01-01",
	}

	rules := map[string]string{
		"dob": "required|date",
	}

	v := validation.Make(data, rules)
	v.SetAttributeNames(map[string]string{
		"dob": "date of birth",
	})

	if v.Passes() {
		fmt.Println("Date of birth is valid")
	}
	// Output: Date of birth is valid
}

// CreateUserRequest is an example FormRequest implementation
type CreateUserRequest struct{}

func (r *CreateUserRequest) Rules() map[string]string {
	return map[string]string{
		"email":    "required|email",
		"password": "required|min:8",
	}
}

func (r *CreateUserRequest) Authorize() bool {
	return true
}

func (r *CreateUserRequest) Messages() map[string]string {
	return map[string]string{
		"email.required":    "Email is required",
		"password.required": "Password is required",
	}
}

func ExampleFormRequest() {
	// Use the form request
	req := &CreateUserRequest{}
	data := map[string]any{
		"email":    "user@example.com",
		"password": "securepass123",
	}

	if !req.Authorize() {
		fmt.Println("Unauthorized")
		return
	}

	v := validation.Make(data, req.Rules(), req.Messages())

	if v.Passes() {
		fmt.Println("User registration data is valid")
	}
	// Output: User registration data is valid
}

func Example_multipleRules() {
	data := map[string]any{
		"username": "john_doe",
		"email":    "john@example.com",
		"age":      25,
		"website":  "https://example.com",
	}

	rules := map[string]string{
		"username": "required|string|alpha_dash|min:3|max:20",
		"email":    "required|email|max:255",
		"age":      "required|integer|between:18,100",
		"website":  "nullable|url",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("All validations passed")
		validated := v.Validated()
		fmt.Printf("Validated %d fields\n", len(validated))
	}
	// Output:
	// All validations passed
	// Validated 4 fields
}

func Example_conditionalValidation() {
	// Example showing required_if conditional validation
	data := map[string]any{
		"payment_method": "credit_card",
		"card_number":    "4111111111111111",
	}

	rules := map[string]string{
		"payment_method": "required|in:credit_card,paypal,bank_transfer",
		"card_number":    "required_if:payment_method,credit_card|digits:16",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("Payment information is valid")
	}
	// Output: Payment information is valid
}

func Example_dateValidation() {
	data := map[string]any{
		"start_date": "2024-01-01",
		"end_date":   "2024-12-31",
	}

	rules := map[string]string{
		"start_date": "required|date",
		"end_date":   "required|date|after:start_date",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("Date range is valid")
	}
	// Output: Date range is valid
}

func Example_passwordConfirmation() {
	data := map[string]any{
		"password":              "secret123",
		"password_confirmation": "secret123",
	}

	rules := map[string]string{
		"password": "required|min:8|confirmed",
	}

	v := validation.Make(data, rules)

	if v.Passes() {
		fmt.Println("Password confirmed")
	}
	// Output: Password confirmed
}
