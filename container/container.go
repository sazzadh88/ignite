package container

import (
	"fmt"
	"reflect"
	"sync"
)

// Container is a dependency injection container inspired by Laravel's IoC Container.
type Container struct {
	mu          sync.RWMutex
	bindings    map[string]*binding
	instances   map[string]any
	aliases     map[string]string
	tags        map[string][]string
	resolving   map[string][]ResolveCallback
	afterResolv map[string][]ResolveCallback
	contextual  []contextualBinding
	resolved    map[string]bool
}

type binding struct {
	concrete  any
	shared    bool
	factory   func(*Container) any
}

type contextualBinding struct {
	concrete string
	needs    string
	give     any
}

// ResolveCallback is called when a binding is resolved.
type ResolveCallback func(any)

// New creates a new Container instance.
func New() *Container {
	return &Container{
		bindings:    make(map[string]*binding),
		instances:   make(map[string]any),
		aliases:     make(map[string]string),
		tags:        make(map[string][]string),
		resolving:   make(map[string][]ResolveCallback),
		afterResolv: make(map[string][]ResolveCallback),
		resolved:    make(map[string]bool),
	}
}

// Bind registers a binding in the container.
func (c *Container) Bind(abstract string, concrete any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := &binding{shared: false}
	switch fn := concrete.(type) {
	case func(*Container) any:
		b.factory = fn
	default:
		b.concrete = concrete
	}
	c.bindings[abstract] = b
	delete(c.instances, abstract)
}

// Singleton registers a shared binding.
func (c *Container) Singleton(abstract string, concrete any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	b := &binding{shared: true}
	switch fn := concrete.(type) {
	case func(*Container) any:
		b.factory = fn
	default:
		b.concrete = concrete
	}
	c.bindings[abstract] = b
}

// Instance registers an existing instance in the container.
func (c *Container) Instance(abstract string, instance any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.instances[abstract] = instance
}

// Make resolves a binding from the container.
func (c *Container) Make(abstract string) any {
	c.mu.RLock()
	if alias, ok := c.aliases[abstract]; ok {
		abstract = alias
	}

	if instance, ok := c.instances[abstract]; ok {
		c.mu.RUnlock()
		return instance
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	b, ok := c.bindings[abstract]
	if !ok {
		panic(fmt.Sprintf("container: no binding found for '%s'", abstract))
	}

	var resolved any
	if b.factory != nil {
		resolved = b.factory(c)
	} else {
		resolved = b.concrete
	}

	if b.shared {
		c.instances[abstract] = resolved
	}

	c.resolved[abstract] = true

	// Fire resolving callbacks
	if callbacks, ok := c.resolving[abstract]; ok {
		for _, cb := range callbacks {
			cb(resolved)
		}
	}
	if callbacks, ok := c.afterResolv[abstract]; ok {
		for _, cb := range callbacks {
			cb(resolved)
		}
	}

	return resolved
}

// MakeInto resolves and casts to the target type.
func MakeInto[T any](c *Container, abstract string) T {
	result := c.Make(abstract)
	typed, ok := result.(T)
	if !ok {
		panic(fmt.Sprintf("container: cannot cast '%s' to %s", abstract, reflect.TypeOf((*T)(nil)).Elem().String()))
	}
	return typed
}

// Call invokes a function with auto-resolved dependencies from the container.
func (c *Container) Call(fn any, params ...any) []any {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic("container: Call() requires a function")
	}

	args := make([]reflect.Value, fnType.NumIn())
	paramIdx := 0

	for i := 0; i < fnType.NumIn(); i++ {
		argType := fnType.In(i)

		// Check explicit params first
		if paramIdx < len(params) {
			args[i] = reflect.ValueOf(params[paramIdx])
			paramIdx++
			continue
		}

		// Try to resolve from container by type name
		typeName := argType.String()
		if argType.Kind() == reflect.Ptr {
			typeName = argType.Elem().String()
		}

		c.mu.RLock()
		_, bound := c.bindings[typeName]
		_, instanced := c.instances[typeName]
		c.mu.RUnlock()

		if bound || instanced {
			args[i] = reflect.ValueOf(c.Make(typeName))
		} else {
			args[i] = reflect.Zero(argType)
		}
	}

	results := reflect.ValueOf(fn).Call(args)
	out := make([]any, len(results))
	for i, r := range results {
		out[i] = r.Interface()
	}
	return out
}

// Extend decorates a resolved value.
func (c *Container) Extend(abstract string, fn func(any, *Container) any) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if instance, ok := c.instances[abstract]; ok {
		c.instances[abstract] = fn(instance, c)
		return
	}

	original := c.bindings[abstract]
	if original == nil {
		panic(fmt.Sprintf("container: cannot extend unbound '%s'", abstract))
	}

	prevFactory := original.factory
	prevConcrete := original.concrete

	original.factory = func(cont *Container) any {
		var resolved any
		if prevFactory != nil {
			resolved = prevFactory(cont)
		} else {
			resolved = prevConcrete
		}
		return fn(resolved, cont)
	}
}

// Bound returns true if a binding exists.
func (c *Container) Bound(abstract string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, b := c.bindings[abstract]
	_, i := c.instances[abstract]
	return b || i
}

// Resolved returns true if a binding has been resolved at least once.
func (c *Container) Resolved(abstract string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.resolved[abstract]
}

// Flush resets all bindings and instances.
func (c *Container) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings = make(map[string]*binding)
	c.instances = make(map[string]any)
	c.aliases = make(map[string]string)
	c.tags = make(map[string][]string)
	c.resolving = make(map[string][]ResolveCallback)
	c.afterResolv = make(map[string][]ResolveCallback)
	c.resolved = make(map[string]bool)
	c.contextual = nil
}

// Alias creates an alias for a binding.
func (c *Container) Alias(abstract, alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = abstract
}

// Tag assigns a tag to a group of abstracts.
func (c *Container) Tag(abstracts []string, tag string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tags[tag] = append(c.tags[tag], abstracts...)
}

// Tagged resolves all bindings for a tag.
func (c *Container) Tagged(tag string) []any {
	c.mu.RLock()
	abstracts, ok := c.tags[tag]
	c.mu.RUnlock()
	if !ok {
		return nil
	}

	results := make([]any, 0, len(abstracts))
	for _, abstract := range abstracts {
		results = append(results, c.Make(abstract))
	}
	return results
}

// Resolving registers a callback to be called when resolving.
func (c *Container) Resolving(abstract string, callback ResolveCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.resolving[abstract] = append(c.resolving[abstract], callback)
}

// AfterResolving registers a callback to be called after resolving.
func (c *Container) AfterResolving(abstract string, callback ResolveCallback) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.afterResolv[abstract] = append(c.afterResolv[abstract], callback)
}

// ContextualBuilder enables contextual binding.
type ContextualBuilder struct {
	container *Container
	concrete  string
	needs     string
}

// When starts a contextual binding definition.
func (c *Container) When(concrete string) *ContextualBuilder {
	return &ContextualBuilder{container: c, concrete: concrete}
}

// Needs specifies what the contextual binding needs.
func (cb *ContextualBuilder) Needs(abstract string) *ContextualBuilder {
	cb.needs = abstract
	return cb
}

// Give provides the contextual implementation.
func (cb *ContextualBuilder) Give(concrete any) {
	cb.container.mu.Lock()
	defer cb.container.mu.Unlock()
	cb.container.contextual = append(cb.container.contextual, contextualBinding{
		concrete: cb.concrete,
		needs:    cb.needs,
		give:     concrete,
	})
}
