package storage

import (
	"os"
	"path/filepath"
	"time"
)

// LocalDisk implements the Disk interface for local filesystem operations.
type LocalDisk struct {
	root string
}

// NewLocalDisk creates a new local disk with the given root directory.
func NewLocalDisk(root string) *LocalDisk {
	return &LocalDisk{root: root}
}

// Put writes content to the specified path.
func (d *LocalDisk) Put(path string, content []byte) error {
	fullPath := d.Path(path)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, content, 0644)
}

// Get retrieves the content of a file.
func (d *LocalDisk) Get(path string) ([]byte, error) {
	return os.ReadFile(d.Path(path))
}

// Exists checks if a file exists.
func (d *LocalDisk) Exists(path string) bool {
	_, err := os.Stat(d.Path(path))
	return err == nil
}

// Missing checks if a file does not exist.
func (d *LocalDisk) Missing(path string) bool {
	return !d.Exists(path)
}

// Delete removes a file.
func (d *LocalDisk) Delete(path string) error {
	return os.Remove(d.Path(path))
}

// DeleteMany removes multiple files.
func (d *LocalDisk) DeleteMany(paths []string) error {
	for _, path := range paths {
		if err := d.Delete(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Copy copies a file from one path to another.
func (d *LocalDisk) Copy(from, to string) error {
	content, err := d.Get(from)
	if err != nil {
		return err
	}
	return d.Put(to, content)
}

// Move moves a file from one path to another.
func (d *LocalDisk) Move(from, to string) error {
	fromPath := d.Path(from)
	toPath := d.Path(to)
	dir := filepath.Dir(toPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.Rename(fromPath, toPath)
}

// Size returns the size of a file in bytes.
func (d *LocalDisk) Size(path string) (int64, error) {
	info, err := os.Stat(d.Path(path))
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// LastModified returns the last modification time.
func (d *LocalDisk) LastModified(path string) (time.Time, error) {
	info, err := os.Stat(d.Path(path))
	if err != nil {
		return time.Time{}, err
	}
	return info.ModTime(), nil
}

// Files lists all files in a directory (non-recursive).
func (d *LocalDisk) Files(dir string) ([]string, error) {
	fullPath := d.Path(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files, nil
}

// AllFiles lists all files recursively.
func (d *LocalDisk) AllFiles(dir string) ([]string, error) {
	var files []string
	fullPath := d.Path(dir)

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.Join(dir, relPath))
		}
		return nil
	})

	return files, err
}

// Directories lists all subdirectories (non-recursive).
func (d *LocalDisk) Directories(dir string) ([]string, error) {
	fullPath := d.Path(dir)
	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, err
	}

	var dirs []string
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(dir, entry.Name()))
		}
	}
	return dirs, nil
}

// AllDirectories lists all subdirectories recursively.
func (d *LocalDisk) AllDirectories(dir string) ([]string, error) {
	var dirs []string
	fullPath := d.Path(dir)

	err := filepath.Walk(fullPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != fullPath {
			relPath, err := filepath.Rel(fullPath, path)
			if err != nil {
				return err
			}
			dirs = append(dirs, filepath.Join(dir, relPath))
		}
		return nil
	})

	return dirs, err
}

// MakeDirectory creates a directory.
func (d *LocalDisk) MakeDirectory(path string) error {
	return os.MkdirAll(d.Path(path), 0755)
}

// DeleteDirectory removes a directory.
func (d *LocalDisk) DeleteDirectory(path string) error {
	return os.RemoveAll(d.Path(path))
}

// URL returns an empty string for local disks (no public URL).
func (d *LocalDisk) URL(path string) string {
	return ""
}

// Path returns the full filesystem path.
func (d *LocalDisk) Path(path string) string {
	return filepath.Join(d.root, path)
}

// Prepend adds data to the beginning of a file.
func (d *LocalDisk) Prepend(path string, data []byte) error {
	var existing []byte
	if d.Exists(path) {
		var err error
		existing, err = d.Get(path)
		if err != nil {
			return err
		}
	}
	return d.Put(path, append(data, existing...))
}

// Append adds data to the end of a file.
func (d *LocalDisk) Append(path string, data []byte) error {
	fullPath := d.Path(path)
	f, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		f, err = os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}
