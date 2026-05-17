package validation

// FormRequest defines an interface for form request validation.
// Implement this interface to create self-validating request objects.
type FormRequest interface {
	// Rules returns the validation rules for the request.
	Rules() map[string]string

	// Authorize determines if the user is authorized to make this request.
	Authorize() bool

	// Messages returns custom error messages for validation rules.
	Messages() map[string]string
}
