package events

import (
	"sync"
	"testing"
	"time"
)

func TestListenAndDispatch(t *testing.T) {
	dispatcher := NewDispatcher()
	called := false

	listener := ListenerFunc(func(event string, payload any) error {
		called = true
		if event != "test.event" {
			t.Errorf("Expected event 'test.event', got '%s'", event)
		}
		if payload != "test data" {
			t.Errorf("Expected payload 'test data', got '%v'", payload)
		}
		return nil
	})

	dispatcher.Listen("test.event", listener)
	dispatcher.DispatchSync("test.event", "test data")

	if !called {
		t.Error("Listener was not called")
	}
}

func TestMultipleListeners(t *testing.T) {
	dispatcher := NewDispatcher()
	callCount := 0
	mu := sync.Mutex{}

	listener1 := ListenerFunc(func(event string, payload any) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	})

	listener2 := ListenerFunc(func(event string, payload any) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		return nil
	})

	dispatcher.Listen("multi.event", listener1)
	dispatcher.Listen("multi.event", listener2)
	dispatcher.DispatchSync("multi.event", nil)

	if callCount != 2 {
		t.Errorf("Expected 2 listeners to be called, got %d", callCount)
	}
}

func TestWildcardListeners(t *testing.T) {
	dispatcher := NewDispatcher()
	calls := []string{}
	mu := sync.Mutex{}

	listener := ListenerFunc(func(event string, payload any) error {
		mu.Lock()
		calls = append(calls, event)
		mu.Unlock()
		return nil
	})

	dispatcher.Listen("user.*", listener)
	dispatcher.DispatchSync("user.created", nil)
	dispatcher.DispatchSync("user.updated", nil)
	dispatcher.DispatchSync("user.deleted", nil)
	dispatcher.DispatchSync("order.created", nil) // Should not match

	if len(calls) != 3 {
		t.Errorf("Expected 3 calls, got %d", len(calls))
	}

	expected := []string{"user.created", "user.updated", "user.deleted"}
	for i, call := range calls {
		if call != expected[i] {
			t.Errorf("Expected call %d to be '%s', got '%s'", i, expected[i], call)
		}
	}
}

func TestUntilStopsOnResponse(t *testing.T) {
	dispatcher := NewDispatcher()

	listener1 := ResponderFunc(func(event string, payload any) any {
		return nil // No response
	})

	listener2 := ResponderFunc(func(event string, payload any) any {
		return "response from listener2"
	})

	listener3 := ResponderFunc(func(event string, payload any) any {
		t.Error("Listener 3 should not be called")
		return "should not reach here"
	})

	dispatcher.Listen("until.event", listener1)
	dispatcher.Listen("until.event", listener2)
	dispatcher.Listen("until.event", listener3)

	response, ok := dispatcher.Until("until.event", nil)

	if !ok {
		t.Error("Until should return true when a response is received")
	}

	if response != "response from listener2" {
		t.Errorf("Expected 'response from listener2', got '%v'", response)
	}
}

func TestUntilNoResponse(t *testing.T) {
	dispatcher := NewDispatcher()

	listener := ResponderFunc(func(event string, payload any) any {
		return nil
	})

	dispatcher.Listen("no.response", listener)

	response, ok := dispatcher.Until("no.response", nil)

	if ok {
		t.Error("Until should return false when no response is received")
	}

	if response != nil {
		t.Errorf("Expected nil response, got '%v'", response)
	}
}

func TestHasListeners(t *testing.T) {
	dispatcher := NewDispatcher()

	if dispatcher.HasListeners("nonexistent") {
		t.Error("HasListeners should return false for unregistered event")
	}

	listener := ListenerFunc(func(event string, payload any) error {
		return nil
	})

	dispatcher.Listen("has.listeners", listener)

	if !dispatcher.HasListeners("has.listeners") {
		t.Error("HasListeners should return true for registered event")
	}
}

func TestHasListenersWildcard(t *testing.T) {
	dispatcher := NewDispatcher()

	listener := ListenerFunc(func(event string, payload any) error {
		return nil
	})

	dispatcher.Listen("user.*", listener)

	if !dispatcher.HasListeners("user.created") {
		t.Error("HasListeners should return true for wildcard match")
	}
}

func TestForget(t *testing.T) {
	dispatcher := NewDispatcher()

	listener := ListenerFunc(func(event string, payload any) error {
		t.Error("Listener should not be called after Forget")
		return nil
	})

	dispatcher.Listen("forget.event", listener)
	dispatcher.Forget("forget.event")

	if dispatcher.HasListeners("forget.event") {
		t.Error("HasListeners should return false after Forget")
	}

	dispatcher.DispatchSync("forget.event", nil)
}

func TestFlush(t *testing.T) {
	dispatcher := NewDispatcher()

	listener := ListenerFunc(func(event string, payload any) error {
		t.Error("Listener should not be called after Flush")
		return nil
	})

	dispatcher.Listen("event1", listener)
	dispatcher.Listen("event2", listener)
	dispatcher.Listen("event3", listener)

	dispatcher.Flush()

	if dispatcher.HasListeners("event1") || dispatcher.HasListeners("event2") || dispatcher.HasListeners("event3") {
		t.Error("No listeners should exist after Flush")
	}
}

func TestSubscriber(t *testing.T) {
	dispatcher := NewDispatcher()
	called := false

	subscriber := &testSubscriber{
		onSubscribe: func(d *Dispatcher) {
			listener := ListenerFunc(func(event string, payload any) error {
				called = true
				return nil
			})
			d.Listen("subscriber.event", listener)
		},
	}

	dispatcher.Subscribe(subscriber)
	dispatcher.DispatchSync("subscriber.event", nil)

	if !called {
		t.Error("Subscriber listener was not called")
	}
}

type testSubscriber struct {
	onSubscribe func(*Dispatcher)
}

func (s *testSubscriber) Subscribe(dispatcher *Dispatcher) {
	s.onSubscribe(dispatcher)
}

func TestAsyncDispatch(t *testing.T) {
	dispatcher := NewDispatcher()
	wg := sync.WaitGroup{}
	wg.Add(1)

	listener := ListenerFunc(func(event string, payload any) error {
		time.Sleep(10 * time.Millisecond)
		wg.Done()
		return nil
	})

	dispatcher.Listen("async.event", listener)
	dispatcher.Dispatch("async.event", nil)

	// Should return immediately
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Error("Async dispatch took too long")
	}
}

func TestFakeDispatcher(t *testing.T) {
	fake := Fake()

	fake.Dispatch("test.event", "payload1")
	fake.Dispatch("test.event", "payload2")
	fake.Dispatch("other.event", "payload3")

	if !fake.AssertDispatched("test.event") {
		t.Error("AssertDispatched should return true for dispatched event")
	}

	if !fake.AssertDispatchedTimes("test.event", 2) {
		t.Error("AssertDispatchedTimes should return true for correct count")
	}

	if !fake.AssertNotDispatched("never.dispatched") {
		t.Error("AssertNotDispatched should return true for non-dispatched event")
	}

	if fake.AssertNothingDispatched() {
		t.Error("AssertNothingDispatched should return false when events were dispatched")
	}
}

func TestFakeDispatcherNothing(t *testing.T) {
	fake := Fake()

	if !fake.AssertNothingDispatched() {
		t.Error("AssertNothingDispatched should return true for fresh fake")
	}

	if !fake.AssertNotDispatched("any.event") {
		t.Error("AssertNotDispatched should return true for any event on fresh fake")
	}

	if !fake.AssertDispatchedTimes("any.event", 0) {
		t.Error("AssertDispatchedTimes should return true for 0 times on fresh fake")
	}
}

func TestQueuedListener(t *testing.T) {
	called := false

	listener := ListenerFunc(func(event string, payload any) error {
		called = true
		return nil
	})

	queued := &QueuedListener{Listener: listener}
	queued.Handle("test.event", nil)

	if !called {
		t.Error("QueuedListener should call wrapped listener")
	}
}
