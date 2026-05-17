package gate

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Gate manages authorization logic through ability definitions and policies.
// It supports before/after hooks, policy resolution, and per-user authorization checks.
type Gate struct {
	abilities map[string]func(user any, args ...any) bool
	policies  map[reflect.Type]any
	before    []func(user any, ability string) *bool
	after     []func(user any, ability string, result bool) *bool
	user      any
	mu        sync.RWMutex
}

// NewGate creates a new Gate instance.
func NewGate() *Gate {
	return &Gate{
		abilities: make(map[string]func(user any, args ...any) bool),
		policies:  make(map[reflect.Type]any),
	}
}

// Define registers an ability with a callback function.
// The callback receives the current user and optional arguments.
func (g *Gate) Define(ability string, callback func(user any, args ...any) bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.abilities[ability] = callback
}

// Has checks if an ability is defined.
func (g *Gate) Has(ability string) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	_, exists := g.abilities[ability]
	return exists
}

// Allows checks if the current user is authorized for the given ability.
// Returns true if authorized, false otherwise.
func (g *Gate) Allows(ability string, args ...any) bool {
	return g.check(g.user, ability, args...)
}

// Denies checks if the current user is not authorized for the given ability.
// Returns true if denied, false otherwise.
func (g *Gate) Denies(ability string, args ...any) bool {
	return !g.Allows(ability, args...)
}

// Authorize checks if the current user is authorized for the given ability.
// Returns an error with 403 status if denied.
func (g *Gate) Authorize(ability string, args ...any) error {
	if g.Denies(ability, args...) {
		return errors.New("this action is unauthorized")
	}
	return nil
}

// Any checks if the current user is authorized for any of the given abilities.
func (g *Gate) Any(abilities []string, args ...any) bool {
	for _, ability := range abilities {
		if g.Allows(ability, args...) {
			return true
		}
	}
	return false
}

// None checks if the current user is not authorized for any of the given abilities.
func (g *Gate) None(abilities []string, args ...any) bool {
	return !g.Any(abilities, args...)
}

// Check checks if the current user is authorized for all of the given abilities.
func (g *Gate) Check(abilities []string, args ...any) bool {
	for _, ability := range abilities {
		if g.Denies(ability, args...) {
			return false
		}
	}
	return true
}

// Before registers a callback to run before all authorization checks.
// If the callback returns a non-nil bool, that result is used and no further checks are performed.
// This is useful for granting access to super admins or similar scenarios.
func (g *Gate) Before(callback func(user any, ability string) *bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.before = append(g.before, callback)
}

// After registers a callback to run after all authorization checks.
// If the callback returns a non-nil bool, that result overrides the original result.
// This is useful for logging or applying additional rules.
func (g *Gate) After(callback func(user any, ability string, result bool) *bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.after = append(g.after, callback)
}

// ForUser returns a UserGate that performs authorization checks against a specific user.
func (g *Gate) ForUser(user any) *UserGate {
	return &UserGate{
		gate: g,
		user: user,
	}
}

// SetUser sets the current user for authorization checks.
func (g *Gate) SetUser(user any) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.user = user
}

// check performs the actual authorization check.
func (g *Gate) check(user any, ability string, args ...any) bool {
	// Run before hooks
	g.mu.RLock()
	beforeHooks := g.before
	g.mu.RUnlock()

	for _, hook := range beforeHooks {
		if result := hook(user, ability); result != nil {
			return *result
		}
	}

	// Check if there's a policy for the first argument
	if len(args) > 0 && args[0] != nil {
		modelType := reflect.TypeOf(args[0])
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}

		g.mu.RLock()
		policy, hasPolicy := g.policies[modelType]
		g.mu.RUnlock()

		if hasPolicy {
			if result := g.callPolicy(policy, ability, user, args...); result != nil {
				return g.runAfterHooks(user, ability, *result)
			}
		}
	}

	// Check ability
	g.mu.RLock()
	callback, exists := g.abilities[ability]
	g.mu.RUnlock()

	if !exists {
		return false
	}

	result := callback(user, args...)
	return g.runAfterHooks(user, ability, result)
}

// runAfterHooks runs after hooks and returns the final result.
func (g *Gate) runAfterHooks(user any, ability string, result bool) bool {
	g.mu.RLock()
	afterHooks := g.after
	g.mu.RUnlock()

	finalResult := result
	for _, hook := range afterHooks {
		if override := hook(user, ability, finalResult); override != nil {
			finalResult = *override
		}
	}
	return finalResult
}

// callPolicy attempts to call a policy method via reflection.
func (g *Gate) callPolicy(policy any, ability string, user any, args ...any) *bool {
	policyValue := reflect.ValueOf(policy)

	// Try common policy method names
	methodNames := []string{
		capitalize(ability),
	}

	for _, methodName := range methodNames {
		method := policyValue.MethodByName(methodName)
		if !method.IsValid() {
			continue
		}

		// Build arguments: user + args
		methodArgs := []reflect.Value{reflect.ValueOf(user)}
		for _, arg := range args {
			methodArgs = append(methodArgs, reflect.ValueOf(arg))
		}

		// Call the method
		results := method.Call(methodArgs)
		if len(results) > 0 && results[0].Kind() == reflect.Bool {
			result := results[0].Bool()
			return &result
		}
	}

	return nil
}

// UserGate provides authorization checks against a specific user.
type UserGate struct {
	gate *Gate
	user any
}

// Allows checks if the user is authorized for the given ability.
func (ug *UserGate) Allows(ability string, args ...any) bool {
	return ug.gate.check(ug.user, ability, args...)
}

// Denies checks if the user is not authorized for the given ability.
func (ug *UserGate) Denies(ability string, args ...any) bool {
	return !ug.Allows(ability, args...)
}

// Authorize checks if the user is authorized for the given ability.
// Returns an error if denied.
func (ug *UserGate) Authorize(ability string, args ...any) error {
	if ug.Denies(ability, args...) {
		return errors.New("this action is unauthorized")
	}
	return nil
}

// Any checks if the user is authorized for any of the given abilities.
func (ug *UserGate) Any(abilities []string, args ...any) bool {
	for _, ability := range abilities {
		if ug.Allows(ability, args...) {
			return true
		}
	}
	return false
}

// None checks if the user is not authorized for any of the given abilities.
func (ug *UserGate) None(abilities []string, args ...any) bool {
	return !ug.Any(abilities, args...)
}

// Check checks if the user is authorized for all of the given abilities.
func (ug *UserGate) Check(abilities []string, args ...any) bool {
	for _, ability := range abilities {
		if ug.Denies(ability, args...) {
			return false
		}
	}
	return true
}

// capitalize converts the first character of a string to uppercase.
func capitalize(s string) string {
	if s == "" {
		return ""
	}
	return fmt.Sprintf("%c%s", s[0]-32, s[1:])
}

// Access is the default Gate instance for package-level operations.
var Access = NewGate()
