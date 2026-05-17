package broadcasting

import (
	"fmt"
	"net/http"
	"sync"
)

// FakeBroadcaster is a fake broadcaster for testing.
// It records all broadcast calls for assertions.
type FakeBroadcaster struct {
	mu         sync.RWMutex
	broadcasts []broadcastRecord
}

type broadcastRecord struct {
	channels []Channel
	event    string
	data     map[string]any
}

// Fake creates a new fake broadcaster for testing.
func Fake() *FakeBroadcaster {
	return &FakeBroadcaster{
		broadcasts: make([]broadcastRecord, 0),
	}
}

// Broadcast records the broadcast call.
func (f *FakeBroadcaster) Broadcast(channels []Channel, event string, data map[string]any) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.broadcasts = append(f.broadcasts, broadcastRecord{
		channels: channels,
		event:    event,
		data:     data,
	})

	return nil
}

// Auth returns nil for testing.
func (f *FakeBroadcaster) Auth(request *http.Request, channel Channel) (any, error) {
	return nil, nil
}

// AssertBroadcast asserts that an event was broadcast.
func (f *FakeBroadcaster) AssertBroadcast(event string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, record := range f.broadcasts {
		if record.event == event {
			return true
		}
	}

	return false
}

// AssertBroadcastOn asserts that an event was broadcast on a specific channel.
func (f *FakeBroadcaster) AssertBroadcastOn(channel Channel, event string) bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	for _, record := range f.broadcasts {
		if record.event == event {
			for _, ch := range record.channels {
				if ch.Name == channel.Name && ch.Type == channel.Type {
					return true
				}
			}
		}
	}

	return false
}

// AssertNotBroadcast asserts that an event was not broadcast.
func (f *FakeBroadcaster) AssertNotBroadcast(event string) bool {
	return !f.AssertBroadcast(event)
}

// AssertNothingBroadcast asserts that no broadcasts were made.
func (f *FakeBroadcaster) AssertNothingBroadcast() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.broadcasts) == 0
}

// GetBroadcasts returns all recorded broadcasts for custom assertions.
func (f *FakeBroadcaster) GetBroadcasts() []broadcastRecord {
	f.mu.RLock()
	defer f.mu.RUnlock()

	records := make([]broadcastRecord, len(f.broadcasts))
	copy(records, f.broadcasts)
	return records
}

// Reset clears all recorded broadcasts.
func (f *FakeBroadcaster) Reset() {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.broadcasts = make([]broadcastRecord, 0)
}

// BroadcastCount returns the number of broadcasts made.
func (f *FakeBroadcaster) BroadcastCount() int {
	f.mu.RLock()
	defer f.mu.RUnlock()

	return len(f.broadcasts)
}

// AssertBroadcastCount asserts the exact number of broadcasts made.
func (f *FakeBroadcaster) AssertBroadcastCount(count int) error {
	actual := f.BroadcastCount()
	if actual != count {
		return fmt.Errorf("expected %d broadcasts, got %d", count, actual)
	}
	return nil
}
