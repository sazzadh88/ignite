package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTestDisk(t *testing.T) (*LocalDisk, string) {
	tmpDir, err := os.MkdirTemp("", "storage_test_")
	if err != nil {
		t.Fatal(err)
	}
	return NewLocalDisk(tmpDir), tmpDir
}

func TestPutAndGet(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("test content")
	err := disk.Put("test.txt", content)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	retrieved, err := disk.Get("test.txt")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrieved) != string(content) {
		t.Errorf("Expected %q, got %q", content, retrieved)
	}
}

func TestExistsAndMissing(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	if !disk.Missing("nonexistent.txt") {
		t.Error("Missing should return true for nonexistent file")
	}

	disk.Put("exists.txt", []byte("data"))

	if !disk.Exists("exists.txt") {
		t.Error("Exists should return true for existing file")
	}

	if disk.Missing("exists.txt") {
		t.Error("Missing should return false for existing file")
	}
}

func TestDelete(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("delete.txt", []byte("data"))
	if !disk.Exists("delete.txt") {
		t.Fatal("File should exist before delete")
	}

	err := disk.Delete("delete.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	if disk.Exists("delete.txt") {
		t.Error("File should not exist after delete")
	}
}

func TestDeleteMany(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("file1.txt", []byte("data1"))
	disk.Put("file2.txt", []byte("data2"))
	disk.Put("file3.txt", []byte("data3"))

	err := disk.DeleteMany([]string{"file1.txt", "file2.txt"})
	if err != nil {
		t.Fatalf("DeleteMany failed: %v", err)
	}

	if disk.Exists("file1.txt") || disk.Exists("file2.txt") {
		t.Error("Files should not exist after DeleteMany")
	}

	if !disk.Exists("file3.txt") {
		t.Error("file3.txt should still exist")
	}
}

func TestCopy(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("copy test")
	disk.Put("original.txt", content)

	err := disk.Copy("original.txt", "copied.txt")
	if err != nil {
		t.Fatalf("Copy failed: %v", err)
	}

	if !disk.Exists("original.txt") {
		t.Error("Original should still exist")
	}

	copied, err := disk.Get("copied.txt")
	if err != nil {
		t.Fatalf("Get copied file failed: %v", err)
	}

	if string(copied) != string(content) {
		t.Error("Copied content doesn't match")
	}
}

func TestMove(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("move test")
	disk.Put("source.txt", content)

	err := disk.Move("source.txt", "destination.txt")
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if disk.Exists("source.txt") {
		t.Error("Source should not exist after move")
	}

	moved, err := disk.Get("destination.txt")
	if err != nil {
		t.Fatalf("Get moved file failed: %v", err)
	}

	if string(moved) != string(content) {
		t.Error("Moved content doesn't match")
	}
}

func TestSize(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	content := []byte("size test content")
	disk.Put("size.txt", content)

	size, err := disk.Size("size.txt")
	if err != nil {
		t.Fatalf("Size failed: %v", err)
	}

	if size != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), size)
	}
}

func TestLastModified(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	before := time.Now()
	disk.Put("modified.txt", []byte("data"))
	after := time.Now()

	modTime, err := disk.LastModified("modified.txt")
	if err != nil {
		t.Fatalf("LastModified failed: %v", err)
	}

	if modTime.Before(before) || modTime.After(after) {
		t.Errorf("ModTime %v not between %v and %v", modTime, before, after)
	}
}

func TestFiles(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("files/file1.txt", []byte("data"))
	disk.Put("files/file2.txt", []byte("data"))
	disk.MakeDirectory("files/subdir")

	files, err := disk.Files("files")
	if err != nil {
		t.Fatalf("Files failed: %v", err)
	}

	if len(files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(files))
	}
}

func TestAllFiles(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("allfiles/file1.txt", []byte("data"))
	disk.Put("allfiles/subdir/file2.txt", []byte("data"))
	disk.Put("allfiles/subdir/nested/file3.txt", []byte("data"))

	files, err := disk.AllFiles("allfiles")
	if err != nil {
		t.Fatalf("AllFiles failed: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("Expected 3 files, got %d", len(files))
	}
}

func TestDirectories(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.MakeDirectory("dirs/sub1")
	disk.MakeDirectory("dirs/sub2")
	disk.MakeDirectory("dirs/sub1/nested")
	disk.Put("dirs/file.txt", []byte("data"))

	dirs, err := disk.Directories("dirs")
	if err != nil {
		t.Fatalf("Directories failed: %v", err)
	}

	if len(dirs) != 2 {
		t.Errorf("Expected 2 directories, got %d", len(dirs))
	}
}

func TestAllDirectories(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.MakeDirectory("alldirs/sub1")
	disk.MakeDirectory("alldirs/sub2")
	disk.MakeDirectory("alldirs/sub1/nested")

	dirs, err := disk.AllDirectories("alldirs")
	if err != nil {
		t.Fatalf("AllDirectories failed: %v", err)
	}

	if len(dirs) != 3 {
		t.Errorf("Expected 3 directories, got %d", len(dirs))
	}
}

func TestMakeDirectory(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	err := disk.MakeDirectory("newdir/nested/deep")
	if err != nil {
		t.Fatalf("MakeDirectory failed: %v", err)
	}

	fullPath := disk.Path("newdir/nested/deep")
	info, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Path is not a directory")
	}
}

func TestDeleteDirectory(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.MakeDirectory("deldir/sub")
	disk.Put("deldir/sub/file.txt", []byte("data"))

	err := disk.DeleteDirectory("deldir")
	if err != nil {
		t.Fatalf("DeleteDirectory failed: %v", err)
	}

	if disk.Exists("deldir") {
		t.Error("Directory should not exist after delete")
	}
}

func TestAppend(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("append.txt", []byte("first"))
	err := disk.Append("append.txt", []byte(" second"))
	if err != nil {
		t.Fatalf("Append failed: %v", err)
	}

	content, _ := disk.Get("append.txt")
	expected := "first second"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestPrepend(t *testing.T) {
	disk, tmpDir := setupTestDisk(t)
	defer os.RemoveAll(tmpDir)

	disk.Put("prepend.txt", []byte("second"))
	err := disk.Prepend("prepend.txt", []byte("first "))
	if err != nil {
		t.Fatalf("Prepend failed: %v", err)
	}

	content, _ := disk.Get("prepend.txt")
	expected := "first second"
	if string(content) != expected {
		t.Errorf("Expected %q, got %q", expected, string(content))
	}
}

func TestPublicDiskURL(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "public_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	disk := NewPublicDisk(tmpDir, "/storage")
	url := disk.URL("images/photo.jpg")
	expected := "/storage/images/photo.jpg"

	if url != expected {
		t.Errorf("Expected URL %q, got %q", expected, url)
	}
}

func TestManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "manager_test_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	config := map[string]any{
		"local": map[string]any{
			"driver": "local",
			"root":   filepath.Join(tmpDir, "local"),
		},
		"public": map[string]any{
			"driver": "public",
			"root":   filepath.Join(tmpDir, "public"),
			"url":    "/storage",
		},
	}

	manager := NewManager(config)

	localDisk := manager.DiskInstance("local")
	if localDisk == nil {
		t.Fatal("Local disk should not be nil")
	}

	err = localDisk.Put("test.txt", []byte("local data"))
	if err != nil {
		t.Fatalf("Put to local disk failed: %v", err)
	}

	publicDisk := manager.DiskInstance("public")
	if publicDisk == nil {
		t.Fatal("Public disk should not be nil")
	}

	url := publicDisk.URL("image.jpg")
	if url != "/storage/image.jpg" {
		t.Errorf("Expected URL /storage/image.jpg, got %s", url)
	}

	// Test caching - should return same instance
	localDisk2 := manager.DiskInstance("local")
	if localDisk != localDisk2 {
		t.Error("DiskInstance should cache instances")
	}
}
