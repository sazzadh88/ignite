// Package storage provides a filesystem abstraction layer for Ignite.
// It supports multiple disk drivers with a unified interface.
package storage

import (
	"time"
)

// Disk defines the interface for filesystem operations.
// All implementations must provide these methods for file and directory manipulation.
type Disk interface {
	// Put writes content to the specified path, creating directories if needed.
	Put(path string, content []byte) error

	// Get retrieves the content of a file at the given path.
	Get(path string) ([]byte, error)

	// Exists checks if a file exists at the given path.
	Exists(path string) bool

	// Missing checks if a file does not exist at the given path.
	Missing(path string) bool

	// Delete removes the file at the given path.
	Delete(path string) error

	// DeleteMany removes multiple files by their paths.
	DeleteMany(paths []string) error

	// Copy copies a file from one path to another.
	Copy(from, to string) error

	// Move moves a file from one path to another.
	Move(from, to string) error

	// Size returns the size of the file in bytes.
	Size(path string) (int64, error)

	// LastModified returns the last modification time of the file.
	LastModified(path string) (time.Time, error)

	// Files lists all files in the given directory (non-recursive).
	Files(dir string) ([]string, error)

	// AllFiles lists all files in the given directory recursively.
	AllFiles(dir string) ([]string, error)

	// Directories lists all subdirectories in the given directory (non-recursive).
	Directories(dir string) ([]string, error)

	// AllDirectories lists all subdirectories in the given directory recursively.
	AllDirectories(dir string) ([]string, error)

	// MakeDirectory creates a directory at the given path.
	MakeDirectory(path string) error

	// DeleteDirectory removes a directory at the given path.
	DeleteDirectory(path string) error

	// URL returns the public URL for the given path.
	URL(path string) string

	// Path returns the full filesystem path for the given path.
	Path(path string) string

	// Prepend adds data to the beginning of a file.
	Prepend(path string, data []byte) error

	// Append adds data to the end of a file.
	Append(path string, data []byte) error
}

// Manager manages multiple disk instances with different configurations.
type Manager struct {
	disks  map[string]Disk
	config map[string]any
}

// NewManager creates a new storage manager with the given configuration.
// The config map should contain disk configurations keyed by disk name.
func NewManager(config map[string]any) *Manager {
	return &Manager{
		disks:  make(map[string]Disk),
		config: config,
	}
}

// DiskInstance returns the disk instance for the given name.
// If the disk doesn't exist, it creates a new one based on the configuration.
func (m *Manager) DiskInstance(name string) Disk {
	if disk, ok := m.disks[name]; ok {
		return disk
	}

	// Create disk based on config
	diskConfig, ok := m.config[name].(map[string]any)
	if !ok {
		return nil
	}

	driver, ok := diskConfig["driver"].(string)
	if !ok {
		return nil
	}

	var disk Disk
	switch driver {
	case "local":
		root, _ := diskConfig["root"].(string)
		disk = NewLocalDisk(root)
	case "public":
		root, _ := diskConfig["root"].(string)
		url, _ := diskConfig["url"].(string)
		disk = NewPublicDisk(root, url)
	default:
		return nil
	}

	m.disks[name] = disk
	return disk
}

// Storage is the package-level facade for the storage manager.
var Storage *Manager
