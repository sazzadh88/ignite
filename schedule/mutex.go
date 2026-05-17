package schedule

import (
	"os"
	"path/filepath"
	"sync"
)

// Mutex provides an interface for preventing overlapping scheduled tasks.
type Mutex interface {
	// Create attempts to create a mutex with the given name.
	// Returns true if the mutex was created, false if it already exists.
	Create(name string) bool

	// Exists checks if a mutex with the given name exists.
	Exists(name string) bool

	// Forget removes the mutex with the given name.
	Forget(name string)
}

// FileMutex is a file-based mutex implementation.
type FileMutex struct {
	dir string
}

// NewFileMutex creates a new file-based mutex.
// Locks are stored in the specified directory.
func NewFileMutex(dir string) *FileMutex {
	if dir == "" {
		dir = os.TempDir()
	}
	return &FileMutex{dir: dir}
}

// Create attempts to create a mutex file.
func (m *FileMutex) Create(name string) bool {
	path := filepath.Join(m.dir, "schedule-"+name+".lock")

	// Check if file exists
	if _, err := os.Stat(path); err == nil {
		return false
	}

	// Create the lock file
	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		return false
	}
	f.Close()

	return true
}

// Exists checks if a mutex file exists.
func (m *FileMutex) Exists(name string) bool {
	path := filepath.Join(m.dir, "schedule-"+name+".lock")
	_, err := os.Stat(path)
	return err == nil
}

// Forget removes a mutex file.
func (m *FileMutex) Forget(name string) {
	path := filepath.Join(m.dir, "schedule-"+name+".lock")
	os.Remove(path)
}

// MemoryMutex is an in-memory mutex implementation.
type MemoryMutex struct {
	mu    sync.RWMutex
	locks map[string]bool
}

// NewMemoryMutex creates a new in-memory mutex.
func NewMemoryMutex() *MemoryMutex {
	return &MemoryMutex{
		locks: make(map[string]bool),
	}
}

// Create attempts to create an in-memory mutex.
func (m *MemoryMutex) Create(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.locks[name] {
		return false
	}

	m.locks[name] = true
	return true
}

// Exists checks if an in-memory mutex exists.
func (m *MemoryMutex) Exists(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.locks[name]
}

// Forget removes an in-memory mutex.
func (m *MemoryMutex) Forget(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.locks, name)
}
