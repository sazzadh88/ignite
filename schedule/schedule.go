package schedule

import (
	"sync"
	"time"
)

// Schedule manages a collection of scheduled events.
type Schedule struct {
	mu     sync.RWMutex
	events []*Event
	mutex  Mutex
}

// NewSchedule creates a new schedule manager.
func NewSchedule() *Schedule {
	return &Schedule{
		events: make([]*Event, 0),
		mutex:  NewMemoryMutex(),
	}
}

// SetMutex sets the mutex implementation for preventing overlapping events.
func (s *Schedule) SetMutex(mutex Mutex) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mutex = mutex
}

// Call schedules a callback function.
func (s *Schedule) Call(fn func()) *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &Event{
		callback: fn,
		mutex:    s.mutex,
	}
	s.events = append(s.events, event)
	return event
}

// Command schedules a shell command.
func (s *Schedule) Command(cmd string) *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &Event{
		command: cmd,
		mutex:   s.mutex,
	}
	s.events = append(s.events, event)
	return event
}

// Job schedules a queue job.
func (s *Schedule) Job(job any) *Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := &Event{
		job:   job,
		mutex: s.mutex,
	}
	s.events = append(s.events, event)
	return event
}

// DueEvents returns all events that are due to run at the given time.
func (s *Schedule) DueEvents(now time.Time) []*Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Truncate to the minute for cron matching
	now = now.Truncate(time.Minute)

	var due []*Event
	for _, event := range s.events {
		if event.IsDue(now) {
			due = append(due, event)
		}
	}

	return due
}

// Run executes all events that are currently due.
// This should be called every minute by a scheduler daemon (e.g., ignite schedule:run).
func (s *Schedule) Run() {
	now := time.Now().Truncate(time.Minute)
	events := s.DueEvents(now)

	for _, event := range events {
		if event.runInBg {
			go event.run()
		} else {
			event.run()
		}
	}
}

// Events returns all registered events.
func (s *Schedule) Events() []*Event {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	events := make([]*Event, len(s.events))
	copy(events, s.events)
	return events
}

// Clear removes all scheduled events.
func (s *Schedule) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.events = make([]*Event, 0)
}
