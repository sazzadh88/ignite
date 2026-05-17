package events

import "sync"

// FakeDispatcher is a test double for the Dispatcher that records dispatched events.
type FakeDispatcher struct {
	mu         sync.RWMutex
	dispatched map[string][]any
}

// Fake creates a new fake dispatcher for testing.
func Fake() *FakeDispatcher {
	return &FakeDispatcher{
		dispatched: make(map[string][]any),
	}
}

// Dispatch records an event dispatch without actually calling listeners.
func (f *FakeDispatcher) Dispatch(event string, payload any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.dispatched[event] = append(f.dispatched[event], payload)
	return nil
}

// DispatchSync records an event dispatch synchronously.
func (f *FakeDispatcher) DispatchSync(event string, payload any) error {
	return f.Dispatch(event, payload)
}

// AssertDispatched checks if an event was dispatched at least once.
func (f *FakeDispatcher) AssertDispatched(event string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payloads, ok := f.dispatched[event]
	return ok && len(payloads) > 0
}

// AssertDispatchedTimes checks if an event was dispatched exactly n times.
func (f *FakeDispatcher) AssertDispatchedTimes(event string, times int) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payloads, ok := f.dispatched[event]
	if !ok {
		return times == 0
	}

	return len(payloads) == times
}

// AssertNotDispatched checks if an event was never dispatched.
func (f *FakeDispatcher) AssertNotDispatched(event string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	payloads, ok := f.dispatched[event]
	return !ok || len(payloads) == 0
}

// AssertNothingDispatched checks if no events were dispatched.
func (f *FakeDispatcher) AssertNothingDispatched() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, payloads := range f.dispatched {
		if len(payloads) > 0 {
			return false
		}
	}

	return len(f.dispatched) == 0
}
