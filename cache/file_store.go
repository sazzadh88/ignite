package cache

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type fileItem struct {
	Value  any       `json:"value"`
	Expiry time.Time `json:"expiry"`
}

func (i *fileItem) isExpired() bool {
	return !i.Expiry.IsZero() && time.Now().After(i.Expiry)
}

// FileStore is a file-based cache implementation.
type FileStore struct {
	directory string
	mu        sync.RWMutex
}

// NewFileStore creates a new file-based cache store.
func NewFileStore(directory string) (*FileStore, error) {
	// Ensure directory exists
	if err := os.MkdirAll(directory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	s := &FileStore{
		directory: directory,
	}

	// Start background cleanup goroutine
	go s.cleanup()

	return s, nil
}

// Get retrieves a value from the cache.
func (s *FileStore) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.path(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, false
	}

	if item.isExpired() {
		os.Remove(path) // Clean up expired file
		return nil, false
	}

	return item.Value, true
}

// Put stores a value in the cache with a TTL.
func (s *FileStore) Put(key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	item := fileItem{
		Value:  value,
		Expiry: expiry,
	}

	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal cache item: %w", err)
	}

	path := s.path(key)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write cache file: %w", err)
	}

	return nil
}

// Forever stores a value in the cache permanently.
func (s *FileStore) Forever(key string, value any) error {
	return s.Put(key, value, 0)
}

// Forget removes a value from the cache.
func (s *FileStore) Forget(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	path := s.path(key)
	err := os.Remove(path)
	return err == nil
}

// Flush clears all values from the cache.
func (s *FileStore) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.directory)
	if err != nil {
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		path := filepath.Join(s.directory, entry.Name())
		os.Remove(path)
	}

	return nil
}

// Has checks if a key exists in the cache.
func (s *FileStore) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	path := s.path(key)
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	var item fileItem
	if err := json.Unmarshal(data, &item); err != nil {
		return false
	}

	if item.isExpired() {
		os.Remove(path)
		return false
	}

	return true
}

// path generates a file path for the given key.
func (s *FileStore) path(key string) string {
	hash := md5.Sum([]byte(key))
	filename := hex.EncodeToString(hash[:])
	return filepath.Join(s.directory, filename)
}

// cleanup removes expired files in the background.
func (s *FileStore) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		entries, err := os.ReadDir(s.directory)
		if err != nil {
			s.mu.Unlock()
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			path := filepath.Join(s.directory, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				continue
			}

			var item fileItem
			if err := json.Unmarshal(data, &item); err != nil {
				continue
			}

			if item.isExpired() {
				os.Remove(path)
			}
		}
		s.mu.Unlock()
	}
}
