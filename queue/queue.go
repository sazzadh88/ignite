package queue

import (
	"fmt"
	"sync"
	"time"
)

// Job represents a unit of work that can be queued.
type Job interface {
	// Handle executes the job's logic.
	Handle() error
	// Queue returns the queue name this job should be pushed to.
	Queue() string
	// Tries returns the maximum number of times to attempt this job.
	Tries() int
	// Timeout returns the maximum duration allowed for this job.
	Timeout() time.Duration
}

// Driver defines the interface for queue backends.
type Driver interface {
	// Push adds a job to the specified queue.
	Push(job Job, queue string) error
	// Pop retrieves the next available job from the queue.
	Pop(queue string) (Job, error)
	// Later schedules a job to be pushed after a delay.
	Later(delay time.Duration, job Job, queue string) error
	// Size returns the number of jobs in the queue.
	Size(queue string) int
	// Flush removes all jobs from the queue.
	Flush(queue string) error
}

// Manager manages multiple queue drivers and provides a facade.
type Manager struct {
	mu            sync.RWMutex
	drivers       map[string]Driver
	defaultDriver string
}

// NewManager creates a new queue manager.
func NewManager() *Manager {
	return &Manager{
		drivers: make(map[string]Driver),
	}
}

// Register adds a driver with the given name.
func (m *Manager) Register(name string, driver Driver) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[name] = driver
	if m.defaultDriver == "" {
		m.defaultDriver = name
	}
}

// SetDefault sets the default driver name.
func (m *Manager) SetDefault(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultDriver = name
}

// Connection returns the driver with the given name, or the default driver if name is empty.
func (m *Manager) Connection(name string) Driver {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.defaultDriver
	}

	driver, ok := m.drivers[name]
	if !ok {
		return nil
	}
	return driver
}

// defaultManager is the package-level manager instance.
var defaultManager = NewManager()

// Register adds a driver to the default manager.
func Register(name string, driver Driver) {
	defaultManager.Register(name, driver)
}

// SetDefault sets the default driver for the default manager.
func SetDefault(name string) {
	defaultManager.SetDefault(name)
}

// Connection returns a driver from the default manager.
func Connection(name string) Driver {
	return defaultManager.Connection(name)
}

// Dispatch pushes a job to its designated queue using the default connection.
func Dispatch(job Job) error {
	driver := defaultManager.Connection("")
	if driver == nil {
		return fmt.Errorf("no default queue driver registered")
	}
	return driver.Push(job, job.Queue())
}

// DispatchSync executes a job synchronously using the sync driver.
func DispatchSync(job Job) error {
	sync := &SyncDriver{}
	return sync.Push(job, job.Queue())
}

// DispatchOn pushes a job to a specific queue using the default connection.
func DispatchOn(queue string, job Job) error {
	driver := defaultManager.Connection("")
	if driver == nil {
		return fmt.Errorf("no default queue driver registered")
	}
	return driver.Push(job, queue)
}

// Later schedules a job to be pushed after a delay using the default connection.
func Later(delay time.Duration, job Job) error {
	driver := defaultManager.Connection("")
	if driver == nil {
		return fmt.Errorf("no default queue driver registered")
	}
	return driver.Later(delay, job, job.Queue())
}

// Chain creates a new chained job from the given jobs.
func Chain(jobs []Job) *ChainedJob {
	return &ChainedJob{Jobs: jobs}
}

// NewBatch creates a new batch builder from the given jobs.
func NewBatch(jobs []Job) *BatchBuilder {
	return &BatchBuilder{
		jobs: jobs,
	}
}
