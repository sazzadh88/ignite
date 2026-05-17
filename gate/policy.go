package gate

import "reflect"

// Policy is a marker interface for policy types.
// Policies group authorization logic for a particular model.
// Policy methods should follow the naming convention: ViewAny, View, Create, Update, Delete, Restore, ForceDelete.
// Each method receives the user and optionally the model instance(s).
type Policy interface{}

// Register registers a policy for a given model type.
// The policy will be automatically invoked when authorization checks
// involve the model type as the first argument.
func (g *Gate) Register(model any, policy Policy) {
	modelType := reflect.TypeOf(model)
	if modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	g.mu.Lock()
	defer g.mu.Unlock()
	g.policies[modelType] = policy
}

// RegisterPolicy is a convenience function that registers a policy on the global Access gate.
func RegisterPolicy(model any, policy Policy) {
	Access.Register(model, policy)
}
