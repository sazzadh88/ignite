package logging

import (
	"os"
	"sync"
)

// Writer is a thread-safe file writer for log output.
type Writer struct {
	mu   sync.Mutex
	file *os.File
	path string
}

// NewWriter creates a new writer that writes to the specified file path.
// The file is opened in append mode and created if it doesn't exist.
func NewWriter(path string) (*Writer, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Writer{
		file: file,
		path: path,
	}, nil
}

// Write writes data to the file in a thread-safe manner.
func (w *Writer) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Write(data)
}

// Close closes the underlying file handle.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// Path returns the file path being written to.
func (w *Writer) Path() string {
	return w.path
}
