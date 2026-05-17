// Package events provides an event dispatching system for Ignite.
// It supports synchronous and asynchronous event handling with wildcard matching.
package events

import (
	"strings"
	"sync"
)

// Dispatcher manages event listeners and dispatches events to them.
type Dispatcher struct {
	mu        sync.RWMutex
	listeners map[string][]Listener
}

// NewDispatcher creates a new event dispatcher.
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		listeners: make(map[string][]Listener),
	}
}

// Listen registers a listener for the specified event.
// Supports wildcard patterns using "*" (e.g., "user.*" matches "user.created", "user.updated").
func (d *Dispatcher) Listen(event string, listener Listener) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.listeners[event] = append(d.listeners[event], listener)
}

// Dispatch fires an event asynchronously to all matching listeners.
// Returns nil immediately and runs listeners in goroutines.
func (d *Dispatcher) Dispatch(event string, payload any) error {
	listeners := d.getMatchingListeners(event)

	for _, listener := range listeners {
		go func(l Listener) {
			l.Handle(event, payload)
		}(listener)
	}

	return nil
}

// DispatchSync fires an event synchronously to all matching listeners.
// Waits for all listeners to complete before returning.
func (d *Dispatcher) DispatchSync(event string, payload any) error {
	listeners := d.getMatchingListeners(event)

	for _, listener := range listeners {
		if err := listener.Handle(event, payload); err != nil {
			return err
		}
	}

	return nil
}

// Until dispatches an event until a listener returns a non-nil response.
// Returns the response and true if a listener returned a value, or nil and false otherwise.
func (d *Dispatcher) Until(event string, payload any) (any, bool) {
	listeners := d.getMatchingListeners(event)

	for _, listener := range listeners {
		if responder, ok := listener.(ResponderListener); ok {
			if response := responder.HandleWithResponse(event, payload); response != nil {
				return response, true
			}
		}
	}

	return nil, false
}

// HasListeners checks if any listeners are registered for the given event.
// Supports wildcard matching.
func (d *Dispatcher) HasListeners(event string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for registeredEvent := range d.listeners {
		if d.matchesPattern(event, registeredEvent) || d.matchesPattern(registeredEvent, event) {
			if len(d.listeners[registeredEvent]) > 0 {
				return true
			}
		}
	}

	return false
}

// Forget removes all listeners for the specified event.
func (d *Dispatcher) Forget(event string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	delete(d.listeners, event)
}

// Flush removes all registered listeners.
func (d *Dispatcher) Flush() {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.listeners = make(map[string][]Listener)
}

// Subscribe registers an event subscriber with the dispatcher.
func (d *Dispatcher) Subscribe(subscriber Subscriber) {
	subscriber.Subscribe(d)
}

// getMatchingListeners returns all listeners that match the given event.
func (d *Dispatcher) getMatchingListeners(event string) []Listener {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var matched []Listener

	for registeredEvent, listeners := range d.listeners {
		if d.matchesPattern(event, registeredEvent) {
			matched = append(matched, listeners...)
		}
	}

	return matched
}

// matchesPattern checks if an event matches a pattern with wildcards.
func (d *Dispatcher) matchesPattern(event, pattern string) bool {
	if pattern == event {
		return true
	}

	if !strings.Contains(pattern, "*") {
		return false
	}

	// Convert wildcard pattern to match
	parts := strings.Split(pattern, "*")
	if len(parts) != 2 {
		return false // Only support single wildcard for now
	}

	prefix := parts[0]
	suffix := parts[1]

	return strings.HasPrefix(event, prefix) && strings.HasSuffix(event, suffix)
}

// Event is the package-level facade for the event dispatcher.
var Event = NewDispatcher()
