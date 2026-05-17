package http

import (
	"bytes"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"
)

func createTestFileHeader(filename, content string) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, _ := writer.CreateFormFile("file", filename)
	fileWriter.Write([]byte(content))
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)

	return form.File["file"][0]
}

func TestNewUploadedFile(t *testing.T) {
	header := createTestFileHeader("test.txt", "content")
	file := NewUploadedFile(header)

	if file == nil {
		t.Fatal("NewUploadedFile returned nil")
	}

	if file.Header() != header {
		t.Error("Header() did not return original header")
	}
}

func TestGetClientOriginalName(t *testing.T) {
	header := createTestFileHeader("document.pdf", "content")
	file := NewUploadedFile(header)

	if name := file.GetClientOriginalName(); name != "document.pdf" {
		t.Errorf("GetClientOriginalName() = %s; want document.pdf", name)
	}
}

func TestGetClientOriginalExtension(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"document.pdf", "pdf"},
		{"image.jpg", "jpg"},
		{"archive.tar.gz", "gz"},
		{"noextension", ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			header := createTestFileHeader(tt.filename, "content")
			file := NewUploadedFile(header)

			if ext := file.GetClientOriginalExtension(); ext != tt.expected {
				t.Errorf("GetClientOriginalExtension() = %s; want %s", ext, tt.expected)
			}
		})
	}
}

func TestGetMimeType(t *testing.T) {
	// Create a more realistic multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, _ := writer.CreateFormFile("file", "test.jpg")
	fileWriter.Write([]byte("content"))

	// Manually set the Content-Type header
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(32 << 20)
	header := form.File["file"][0]

	// Set a specific MIME type
	header.Header.Set("Content-Type", "image/jpeg")

	file := NewUploadedFile(header)

	if mimeType := file.GetMimeType(); mimeType != "image/jpeg" {
		t.Errorf("GetMimeType() = %s; want image/jpeg", mimeType)
	}
}

func TestGetSize(t *testing.T) {
	content := "This is a test file content"
	header := createTestFileHeader("test.txt", content)
	file := NewUploadedFile(header)

	expectedSize := int64(len(content))
	if size := file.GetSize(); size != expectedSize {
		t.Errorf("GetSize() = %d; want %d", size, expectedSize)
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{"valid file", "content", true},
		{"empty file", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := createTestFileHeader("test.txt", tt.content)
			file := NewUploadedFile(header)

			if valid := file.IsValid(); valid != tt.expected {
				t.Errorf("IsValid() = %v; want %v", valid, tt.expected)
			}
		})
	}
}

func TestIsImage(t *testing.T) {
	tests := []struct {
		mimeType string
		expected bool
	}{
		{"image/jpeg", true},
		{"image/png", true},
		{"image/gif", true},
		{"text/plain", false},
		{"application/pdf", false},
	}

	for _, tt := range tests {
		t.Run(tt.mimeType, func(t *testing.T) {
			header := createTestFileHeader("test", "content")
			header.Header.Set("Content-Type", tt.mimeType)
			file := NewUploadedFile(header)

			if isImage := file.IsImage(); isImage != tt.expected {
				t.Errorf("IsImage() = %v for %s; want %v", isImage, tt.mimeType, tt.expected)
			}
		})
	}
}

func TestStore(t *testing.T) {
	tmpDir := t.TempDir()
	content := "test file content"

	header := createTestFileHeader("uploaded.txt", content)
	file := NewUploadedFile(header)

	path, err := file.Store(tmpDir)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "uploaded.txt")
	if path != expectedPath {
		t.Errorf("Store() path = %s; want %s", path, expectedPath)
	}

	// Verify file exists and has correct content
	storedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read stored file: %v", err)
	}

	if string(storedContent) != content {
		t.Errorf("Stored content = %s; want %s", string(storedContent), content)
	}
}

func TestStoreAs(t *testing.T) {
	tmpDir := t.TempDir()
	content := "test file content"

	header := createTestFileHeader("original.txt", content)
	file := NewUploadedFile(header)

	path, err := file.StoreAs(tmpDir, "custom-name.txt")
	if err != nil {
		t.Fatalf("StoreAs() error: %v", err)
	}

	expectedPath := filepath.Join(tmpDir, "custom-name.txt")
	if path != expectedPath {
		t.Errorf("StoreAs() path = %s; want %s", path, expectedPath)
	}

	// Verify file exists with custom name
	if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
		t.Error("StoreAs() did not create file with custom name")
	}

	// Verify content
	storedContent, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read stored file: %v", err)
	}

	if string(storedContent) != content {
		t.Errorf("Stored content = %s; want %s", string(storedContent), content)
	}
}

func TestStoreCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "uploads", "images")
	content := "test"

	header := createTestFileHeader("test.jpg", content)
	file := NewUploadedFile(header)

	path, err := file.Store(subDir)
	if err != nil {
		t.Fatalf("Store() error: %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Error("Store() did not create directory")
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("Store() did not create file")
	}
}

func TestOpen(t *testing.T) {
	content := "test file content"
	header := createTestFileHeader("test.txt", content)
	file := NewUploadedFile(header)

	reader, err := file.Open()
	if err != nil {
		t.Fatalf("Open() error: %v", err)
	}
	defer reader.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	if buf.String() != content {
		t.Errorf("Open() content = %s; want %s", buf.String(), content)
	}
}
