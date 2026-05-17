package validation

import "sync"

// RuleFunc defines the signature for custom validation rules.
// It receives the attribute name, value, rule parameters, and full data map.
// It returns true if validation passes, false otherwise.
type RuleFunc func(attribute string, value any, params []string, data map[string]any) bool

type customRule struct {
	fn      RuleFunc
	message string
}

var (
	customRules   = make(map[string]customRule)
	customRulesMu sync.RWMutex
)

// Extend registers a custom validation rule.
// The name is the rule identifier, fn is the validation function, and message is the default error message.
// Use :attribute, :value, and :param placeholders in the message.
func Extend(name string, fn RuleFunc, message string) {
	customRulesMu.Lock()
	defer customRulesMu.Unlock()
	customRules[name] = customRule{
		fn:      fn,
		message: message,
	}
}

// getCustomRule retrieves a custom rule by name.
func getCustomRule(name string) (customRule, bool) {
	customRulesMu.RLock()
	defer customRulesMu.RUnlock()
	rule, ok := customRules[name]
	return rule, ok
}
