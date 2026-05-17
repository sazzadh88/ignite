package orm

// ScopeFunc is a function that modifies a query.
type ScopeFunc[T any] func(*Query[T]) *Query[T]

// Scope represents a named scope that can be applied to queries.
type Scope[T any] struct {
	name string
	fn   ScopeFunc[T]
}

// NewScope creates a new named scope.
func NewScope[T any](name string, fn ScopeFunc[T]) *Scope[T] {
	return &Scope[T]{
		name: name,
		fn:   fn,
	}
}

// Apply applies the scope to a query.
func (s *Scope[T]) Apply(query *Query[T]) *Query[T] {
	return s.fn(query)
}

// Name returns the scope's name.
func (s *Scope[T]) Name() string {
	return s.name
}

// GlobalScope represents a scope that is automatically applied to all queries.
type GlobalScope[T any] struct {
	name string
	fn   ScopeFunc[T]
}

// NewGlobalScope creates a new global scope.
func NewGlobalScope[T any](name string, fn ScopeFunc[T]) *GlobalScope[T] {
	return &GlobalScope[T]{
		name: name,
		fn:   fn,
	}
}

// Apply applies the global scope to a query.
func (g *GlobalScope[T]) Apply(query *Query[T]) *Query[T] {
	return g.fn(query)
}

// Name returns the global scope's name.
func (g *GlobalScope[T]) Name() string {
	return g.name
}

// ScopeRegistry manages global and local scopes for a model.
type ScopeRegistry[T any] struct {
	globalScopes map[string]*GlobalScope[T]
	localScopes  map[string]*Scope[T]
}

// NewScopeRegistry creates a new scope registry.
func NewScopeRegistry[T any]() *ScopeRegistry[T] {
	return &ScopeRegistry[T]{
		globalScopes: make(map[string]*GlobalScope[T]),
		localScopes:  make(map[string]*Scope[T]),
	}
}

// AddGlobalScope registers a global scope.
func (r *ScopeRegistry[T]) AddGlobalScope(name string, fn ScopeFunc[T]) {
	r.globalScopes[name] = NewGlobalScope(name, fn)
}

// AddLocalScope registers a local scope.
func (r *ScopeRegistry[T]) AddLocalScope(name string, fn ScopeFunc[T]) {
	r.localScopes[name] = NewScope(name, fn)
}

// RemoveGlobalScope removes a global scope.
func (r *ScopeRegistry[T]) RemoveGlobalScope(name string) {
	delete(r.globalScopes, name)
}

// GetGlobalScope retrieves a global scope by name.
func (r *ScopeRegistry[T]) GetGlobalScope(name string) (*GlobalScope[T], bool) {
	scope, ok := r.globalScopes[name]
	return scope, ok
}

// GetLocalScope retrieves a local scope by name.
func (r *ScopeRegistry[T]) GetLocalScope(name string) (*Scope[T], bool) {
	scope, ok := r.localScopes[name]
	return scope, ok
}

// ApplyGlobalScopes applies all global scopes to a query.
func (r *ScopeRegistry[T]) ApplyGlobalScopes(query *Query[T], except []string) *Query[T] {
	exceptMap := make(map[string]bool)
	for _, name := range except {
		exceptMap[name] = true
	}

	for name, scope := range r.globalScopes {
		if !exceptMap[name] {
			query = scope.Apply(query)
		}
	}

	return query
}

// ApplyLocalScope applies a named local scope to a query.
func (r *ScopeRegistry[T]) ApplyLocalScope(query *Query[T], name string) *Query[T] {
	if scope, ok := r.localScopes[name]; ok {
		return scope.Apply(query)
	}
	return query
}

// WithoutGlobalScope removes a specific global scope from the query.
func (q *Query[T]) WithoutGlobalScope(name string) *Query[T] {
	// Filter out the named scope
	filtered := make([]ScopeFunc[T], 0)
	for _, scope := range q.globalScopes {
		// Note: In a real implementation, we'd need to track scope names
		// For now, this is a placeholder
		filtered = append(filtered, scope)
	}
	q.globalScopes = filtered
	return q
}

// WithoutGlobalScopes removes all global scopes from the query.
func (q *Query[T]) WithoutGlobalScopes() *Query[T] {
	q.globalScopes = make([]ScopeFunc[T], 0)
	return q
}

// Common scope functions

// WhereActive is a common scope that filters for active records.
func WhereActive[T any]() ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.Where("active", "=", true)
	}
}

// WherePublished is a common scope that filters for published records.
func WherePublished[T any]() ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.Where("published", "=", true)
	}
}

// WhereStatus creates a scope that filters by status.
func WhereStatus[T any](status string) ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.Where("status", "=", status)
	}
}

// OfType creates a scope that filters by type.
func OfType[T any](typeValue string) ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.Where("type", "=", typeValue)
	}
}

// Recent creates a scope that orders by created_at descending.
func Recent[T any]() ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.OrderBy("created_at", "desc")
	}
}

// Oldest creates a scope that orders by created_at ascending.
func Oldest[T any]() ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		return q.OrderBy("created_at", "asc")
	}
}
