package validation

import (
	"fmt"
	"strings"
)

// Validator validates data against rules.
type Validator struct {
	data           map[string]any
	rules          map[string]string
	customMessages map[string]string
	attributeNames map[string]string
	errors         *ErrorBag
	conditionals   map[string]conditionalRule
}

type conditionalRule struct {
	rules     string
	condition func(map[string]any) bool
}

// Make creates a new Validator instance.
// data is the data to validate, rules are validation rules, and messages are optional custom error messages.
func Make(data map[string]any, rules map[string]string, messages ...map[string]string) *Validator {
	v := &Validator{
		data:           data,
		rules:          rules,
		customMessages: make(map[string]string),
		attributeNames: make(map[string]string),
		errors:         NewErrorBag(),
		conditionals:   make(map[string]conditionalRule),
	}

	if len(messages) > 0 {
		v.customMessages = messages[0]
	}

	v.validate()
	return v
}

// Fails returns true if validation failed.
func (v *Validator) Fails() bool {
	return !v.errors.IsEmpty()
}

// Passes returns true if validation passed.
func (v *Validator) Passes() bool {
	return v.errors.IsEmpty()
}

// Errors returns the error bag containing all validation errors.
func (v *Validator) Errors() *ErrorBag {
	return v.errors
}

// Validated returns a map of only the validated fields.
func (v *Validator) Validated() map[string]any {
	validated := make(map[string]any)
	for field := range v.rules {
		if value, ok := v.data[field]; ok {
			validated[field] = value
		}
	}
	return validated
}

// SetAttributeNames sets custom display names for attributes in error messages.
func (v *Validator) SetAttributeNames(names map[string]string) {
	v.attributeNames = names
}

// Sometimes adds a conditional validation rule.
// The rule is only applied if the condition function returns true.
func (v *Validator) Sometimes(field, rules string, condition func(map[string]any) bool) {
	v.conditionals[field] = conditionalRule{
		rules:     rules,
		condition: condition,
	}
}

// validate performs the validation.
func (v *Validator) validate() {
	// Apply regular rules
	for field, ruleString := range v.rules {
		v.validateField(field, ruleString)
	}

	// Apply conditional rules
	for field, conditional := range v.conditionals {
		if conditional.condition(v.data) {
			v.validateField(field, conditional.rules)
		}
	}
}

// validateField validates a single field against its rules.
func (v *Validator) validateField(field, ruleString string) {
	rules := parseRules(ruleString)
	value, exists := v.data[field]

	// Handle nullable
	nullable := false
	for _, rule := range rules {
		if rule.name == "nullable" {
			nullable = true
			break
		}
	}

	if !exists || (nullable && value == nil) {
		// Skip validation for nullable fields if value is nil
		if nullable {
			return
		}
	}

	for _, rule := range rules {
		if rule.name == "nullable" {
			continue
		}

		if !v.validateRule(field, value, rule) {
			message := v.buildErrorMessage(field, rule)
			v.errors.Add(field, message)
		}
	}
}

// validateRule validates a single rule.
func (v *Validator) validateRule(field string, value any, rule parsedRule) bool {
	// Check built-in rules
	if fn, ok := builtInRules[rule.name]; ok {
		return fn(field, value, rule.params, v.data)
	}

	// Check custom rules
	if customRule, ok := getCustomRule(rule.name); ok {
		return customRule.fn(field, value, rule.params, v.data)
	}

	// Unknown rule, skip validation
	return true
}

// buildErrorMessage constructs the error message for a failed validation.
func (v *Validator) buildErrorMessage(field string, rule parsedRule) string {
	// Priority: custom message > custom rule message > default message
	key := field + "." + rule.name
	if msg, ok := v.customMessages[key]; ok {
		return v.formatMessage(msg, field, rule)
	}

	if customRule, ok := getCustomRule(rule.name); ok {
		return v.formatMessage(customRule.message, field, rule)
	}

	if msg, ok := defaultMessages[rule.name]; ok {
		return v.formatMessage(msg, field, rule)
	}

	return fmt.Sprintf("Validation failed for %s", field)
}

// formatMessage replaces placeholders in error messages.
func (v *Validator) formatMessage(message, field string, rule parsedRule) string {
	// Get display name for attribute
	displayName := field
	if name, ok := v.attributeNames[field]; ok {
		displayName = name
	}

	message = strings.ReplaceAll(message, ":attribute", displayName)
	message = strings.ReplaceAll(message, ":field", displayName)

	// Replace common placeholders
	if len(rule.params) > 0 {
		message = strings.ReplaceAll(message, ":min", rule.params[0])
		message = strings.ReplaceAll(message, ":max", rule.params[0])
		message = strings.ReplaceAll(message, ":size", rule.params[0])
		message = strings.ReplaceAll(message, ":value", rule.params[0])
		message = strings.ReplaceAll(message, ":other", rule.params[0])
		message = strings.ReplaceAll(message, ":date", rule.params[0])
		message = strings.ReplaceAll(message, ":digits", rule.params[0])
		message = strings.ReplaceAll(message, ":values", strings.Join(rule.params, ", "))
	}

	if len(rule.params) > 1 {
		message = strings.ReplaceAll(message, ":max", rule.params[1])
	}

	// Get value for certain placeholders
	if value, ok := v.data[field]; ok {
		message = strings.ReplaceAll(message, ":value", fmt.Sprintf("%v", value))
	}

	return message
}

type parsedRule struct {
	name   string
	params []string
}

// parseRules parses a pipe-separated rule string into individual rules.
func parseRules(ruleString string) []parsedRule {
	parts := strings.Split(ruleString, "|")
	rules := make([]parsedRule, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Split rule name and parameters
		colonIdx := strings.Index(part, ":")
		if colonIdx == -1 {
			rules = append(rules, parsedRule{name: part})
			continue
		}

		name := part[:colonIdx]
		paramString := part[colonIdx+1:]
		params := strings.Split(paramString, ",")

		// Trim params
		for i := range params {
			params[i] = strings.TrimSpace(params[i])
		}

		rules = append(rules, parsedRule{
			name:   name,
			params: params,
		})
	}

	return rules
}

// Factory provides a facade for creating validators.
type Factory struct{}

// Make creates a new validator instance.
func (f *Factory) Make(data map[string]any, rules map[string]string, messages ...map[string]string) *Validator {
	return Make(data, rules, messages...)
}

// Validate is the package-level facade for the validator factory.
var Validate = &Factory{}
