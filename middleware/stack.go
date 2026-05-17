package middleware

import (
	"sort"
	"sync"
)

// Stack manages middleware registration and organization.
// It supports global middleware, named middleware, and middleware groups
// with priority-based ordering.
type Stack struct {
	mu sync.RWMutex

	// Global middleware applied to all routes
	global []prioritizedMiddleware

	// Named middleware registered by key
	named map[string]Middleware

	// Middleware groups (e.g., "web", "api")
	groups map[string][]string

	// Priority counter for maintaining order
	priorityCounter int
}

// prioritizedMiddleware wraps middleware with a priority value for ordering.
type prioritizedMiddleware struct {
	middleware Middleware
	priority   int
}

// NewStack creates a new middleware stack.
func NewStack() *Stack {
	return &Stack{
		named:  make(map[string]Middleware),
		groups: make(map[string][]string),
		global: make([]prioritizedMiddleware, 0),
	}
}

// PushGlobal adds middleware to the global stack with the given priority.
// Lower priority values execute first. If priority is 0, it's assigned
// automatically based on insertion order.
func (s *Stack) PushGlobal(middleware Middleware, priority int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if priority == 0 {
		s.priorityCounter++
		priority = s.priorityCounter
	}

	s.global = append(s.global, prioritizedMiddleware{
		middleware: middleware,
		priority:   priority,
	})

	// Sort global middleware by priority
	sort.Slice(s.global, func(i, j int) bool {
		return s.global[i].priority < s.global[j].priority
	})
}

// Register adds a named middleware that can be referenced by key.
func (s *Stack) Register(name string, middleware Middleware) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.named[name] = middleware
}

// Get retrieves a named middleware by key.
// Returns nil if the middleware is not found.
func (s *Stack) Get(name string) Middleware {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.named[name]
}

// Group defines a middleware group with the given name and middleware keys.
// Groups allow multiple middleware to be applied together using a single name.
func (s *Stack) Group(name string, middlewareNames []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groups[name] = middlewareNames
}

// GetGroup retrieves middleware in a group by name.
// Returns an empty slice if the group is not found.
func (s *Stack) GetGroup(name string) []Middleware {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names, exists := s.groups[name]
	if !exists {
		return []Middleware{}
	}

	middlewares := make([]Middleware, 0, len(names))
	for _, name := range names {
		if m, exists := s.named[name]; exists {
			middlewares = append(middlewares, m)
		}
	}

	return middlewares
}

// Global returns all global middleware in priority order.
func (s *Stack) Global() []Middleware {
	s.mu.RLock()
	defer s.mu.RUnlock()

	middlewares := make([]Middleware, len(s.global))
	for i, pm := range s.global {
		middlewares[i] = pm.middleware
	}
	return middlewares
}

// Resolve converts a list of middleware names into their corresponding
// Middleware instances. Unknown names are skipped.
func (s *Stack) Resolve(names []string) []Middleware {
	s.mu.RLock()
	defer s.mu.RUnlock()

	middlewares := make([]Middleware, 0, len(names))
	for _, name := range names {
		// Check if it's a group first
		if group, exists := s.groups[name]; exists {
			for _, gName := range group {
				if m, exists := s.named[gName]; exists {
					middlewares = append(middlewares, m)
				}
			}
			continue
		}

		// Otherwise treat as named middleware
		if m, exists := s.named[name]; exists {
			middlewares = append(middlewares, m)
		}
	}

	return middlewares
}
