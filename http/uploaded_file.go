package http

import (
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// UploadedFile represents an uploaded file.
type UploadedFile struct {
	header *multipart.FileHeader
}

// NewUploadedFile creates a new UploadedFile from multipart.FileHeader.
func NewUploadedFile(header *multipart.FileHeader) *UploadedFile {
	return &UploadedFile{header: header}
}

// GetClientOriginalName returns the original filename.
func (f *UploadedFile) GetClientOriginalName() string {
	return f.header.Filename
}

// GetClientOriginalExtension returns the file extension without the dot.
func (f *UploadedFile) GetClientOriginalExtension() string {
	ext := filepath.Ext(f.header.Filename)
	return strings.TrimPrefix(ext, ".")
}

// GetMimeType returns the MIME type of the file.
func (f *UploadedFile) GetMimeType() string {
	if len(f.header.Header["Content-Type"]) > 0 {
		return f.header.Header["Content-Type"][0]
	}
	return "application/octet-stream"
}

// GetSize returns the file size in bytes.
func (f *UploadedFile) GetSize() int64 {
	return f.header.Size
}

// IsValid checks if the uploaded file is valid.
func (f *UploadedFile) IsValid() bool {
	return f.header != nil && f.header.Size > 0
}

// IsImage checks if the file is an image based on MIME type.
func (f *UploadedFile) IsImage() bool {
	mimeType := f.GetMimeType()
	return strings.HasPrefix(mimeType, "image/")
}

// Store saves the file to the given directory.
// Returns the stored path relative to the storage root.
func (f *UploadedFile) Store(directory string) (string, error) {
	return f.StoreAs(directory, f.header.Filename)
}

// StoreAs saves the file with a custom name.
// Returns the stored path relative to the storage root.
func (f *UploadedFile) StoreAs(directory, name string) (string, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(directory, 0755); err != nil {
		return "", err
	}

	// Construct full path
	fullPath := filepath.Join(directory, name)

	// Open the uploaded file
	src, err := f.header.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	return fullPath, nil
}

// Open opens the uploaded file for reading.
func (f *UploadedFile) Open() (multipart.File, error) {
	return f.header.Open()
}

// Header returns the underlying multipart.FileHeader.
func (f *UploadedFile) Header() *multipart.FileHeader {
	return f.header
}
