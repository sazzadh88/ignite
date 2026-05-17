package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileStore implements session storage using the filesystem.
type FileStore struct {
	path string
	mu   sync.RWMutex
}

// NewFileStore creates a new file-based session store.
func NewFileStore(path string) (*FileStore, error) {
	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	return &FileStore{
		path: path,
	}, nil
}

// Read retrieves session data from a file.
func (f *FileStore) Read(id string) (map[string]any, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	filePath := f.getFilePath(id)

	// Check if file exists and is not expired
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("failed to stat session file: %w", err)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	var wrapper struct {
		Data      map[string]any `json:"data"`
		ExpiresAt int64          `json:"expires_at"`
	}

	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	// Check if expired
	if wrapper.ExpiresAt > 0 && time.Now().Unix() > wrapper.ExpiresAt {
		_ = os.Remove(filePath)
		return make(map[string]any), nil
	}

	// Update access time
	_ = os.Chtimes(filePath, time.Now(), info.ModTime())

	return wrapper.Data, nil
}

// Write persists session data to a file.
func (f *FileStore) Write(id string, data map[string]any, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := f.getFilePath(id)

	wrapper := struct {
		Data      map[string]any `json:"data"`
		ExpiresAt int64          `json:"expires_at"`
	}{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// Destroy removes the session file.
func (f *FileStore) Destroy(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	filePath := f.getFilePath(id)
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file: %w", err)
	}

	return nil
}

// GC performs garbage collection on expired session files.
func (f *FileStore) GC(maxLifetime time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	entries, err := os.ReadDir(f.path)
	if err != nil {
		return fmt.Errorf("failed to read session directory: %w", err)
	}

	cutoff := time.Now().Add(-maxLifetime)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filePath := filepath.Join(f.path, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		// Remove files that haven't been accessed since cutoff
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(filePath)
		}
	}

	return nil
}

// getFilePath returns the full path for a session file.
func (f *FileStore) getFilePath(id string) string {
	return filepath.Join(f.path, fmt.Sprintf("sess_%s", id))
}
