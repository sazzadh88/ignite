package storage

import (
	"path"
)

// PublicDisk extends LocalDisk with public URL generation capabilities.
type PublicDisk struct {
	*LocalDisk
	baseURL string
}

// NewPublicDisk creates a new public disk with the given root and base URL.
func NewPublicDisk(root, baseURL string) *PublicDisk {
	return &PublicDisk{
		LocalDisk: NewLocalDisk(root),
		baseURL:   baseURL,
	}
}

// URL returns the public URL for the given path.
func (d *PublicDisk) URL(filePath string) string {
	return path.Join(d.baseURL, filePath)
}
