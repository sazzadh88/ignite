package broadcasting

import (
	"fmt"
	"net/http"
	"sync"
)

// Broadcaster defines the interface for broadcasting events to channels.
type Broadcaster interface {
	// Broadcast sends an event with data to the specified channels.
	Broadcast(channels []Channel, event string, data map[string]any) error
	// Auth authorizes a request for a channel and returns user data.
	Auth(request *http.Request, channel Channel) (any, error)
}

// Manager manages multiple broadcaster drivers.
type Manager struct {
	mu            sync.RWMutex
	drivers       map[string]Broadcaster
	defaultDriver string
}

// NewManager creates a new broadcaster manager.
func NewManager() *Manager {
	return &Manager{
		drivers:       make(map[string]Broadcaster),
		defaultDriver: "null",
	}
}

// Driver returns the broadcaster for the specified driver name.
// If name is empty, returns the default driver.
// Returns a null broadcaster if the driver is not found.
func (m *Manager) Driver(name string) Broadcaster {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.defaultDriver
	}

	if driver, ok := m.drivers[name]; ok {
		return driver
	}

	return &NullBroadcaster{}
}

// Extend registers a new broadcaster driver.
func (m *Manager) Extend(name string, driver Broadcaster) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[name] = driver
}

// SetDefaultDriver sets the default broadcaster driver.
func (m *Manager) SetDefaultDriver(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDriver = name
}

// Broadcast broadcasts using the default driver.
func (m *Manager) Broadcast(channels []Channel, event string, data map[string]any) error {
	return m.Driver("").Broadcast(channels, event, data)
}

// Event broadcasts a ShouldBroadcast event using the default driver.
func (m *Manager) Event(event ShouldBroadcast) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}
	return m.Driver("").Broadcast(event.BroadcastOn(), event.BroadcastAs(), event.BroadcastWith())
}

// Broadcast is the package-level default broadcaster manager.
var Broadcast = NewManager()
