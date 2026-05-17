package events

// Listener defines the interface for event listeners.
type Listener interface {
	// Handle processes an event with the given payload.
	Handle(event string, payload any) error
}

// ResponderListener extends Listener to support returning responses.
type ResponderListener interface {
	Listener
	// HandleWithResponse processes an event and returns a response.
	HandleWithResponse(event string, payload any) any
}

// ListenerFunc is a function adapter that implements the Listener interface.
type ListenerFunc func(event string, payload any) error

// Handle implements the Listener interface for ListenerFunc.
func (f ListenerFunc) Handle(event string, payload any) error {
	return f(event, payload)
}

// ResponderFunc is a function adapter that implements the ResponderListener interface.
type ResponderFunc func(event string, payload any) any

// Handle implements the Listener interface for ResponderFunc.
func (f ResponderFunc) Handle(event string, payload any) error {
	f(event, payload)
	return nil
}

// HandleWithResponse implements the ResponderListener interface for ResponderFunc.
func (f ResponderFunc) HandleWithResponse(event string, payload any) any {
	return f(event, payload)
}

// QueuedListener marks a listener for asynchronous dispatch.
// This is a placeholder for future queue integration.
type QueuedListener struct {
	Listener Listener
}

// Handle implements the Listener interface for QueuedListener.
func (q *QueuedListener) Handle(event string, payload any) error {
	// In the future, this would push to a queue
	// For now, it just calls the wrapped listener
	return q.Listener.Handle(event, payload)
}
