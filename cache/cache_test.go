package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMemoryStore_PutGet(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Put and Get
	err := repo.Put("key1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	val := repo.Get("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestMemoryStore_Has(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Has
	if repo.Has("missing") {
		t.Error("expected Has to return false for missing key")
	}

	repo.Put("key1", "value1", 1*time.Minute)
	if !repo.Has("key1") {
		t.Error("expected Has to return true for existing key")
	}
}

func TestMemoryStore_Missing(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Missing
	if !repo.Missing("missing") {
		t.Error("expected Missing to return true for missing key")
	}

	repo.Put("key1", "value1", 1*time.Minute)
	if repo.Missing("key1") {
		t.Error("expected Missing to return false for existing key")
	}
}

func TestMemoryStore_Forget(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Forget
	repo.Put("key1", "value1", 1*time.Minute)
	if !repo.Forget("key1") {
		t.Error("expected Forget to return true")
	}

	if repo.Has("key1") {
		t.Error("expected key to be forgotten")
	}
}

func TestMemoryStore_Flush(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Flush
	repo.Put("key1", "value1", 1*time.Minute)
	repo.Put("key2", "value2", 1*time.Minute)

	err := repo.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if repo.Has("key1") || repo.Has("key2") {
		t.Error("expected all keys to be flushed")
	}
}

func TestMemoryStore_TTLExpiration(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test TTL expiration
	repo.Put("key1", "value1", 100*time.Millisecond)

	if !repo.Has("key1") {
		t.Error("expected key to exist immediately after Put")
	}

	time.Sleep(150 * time.Millisecond)

	if repo.Has("key1") {
		t.Error("expected key to be expired")
	}
}

func TestMemoryStore_Forever(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Forever
	err := repo.Forever("key1", "value1")
	if err != nil {
		t.Fatalf("Forever failed: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	if !repo.Has("key1") {
		t.Error("expected key to persist forever")
	}
}

func TestRepository_Increment(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Increment
	val, err := repo.Increment("counter")
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	val, err = repo.Increment("counter", 5)
	if err != nil {
		t.Fatalf("Increment failed: %v", err)
	}
	if val != 6 {
		t.Errorf("expected 6, got %d", val)
	}
}

func TestRepository_Decrement(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Decrement
	repo.Put("counter", 10, 1*time.Minute)

	val, err := repo.Decrement("counter")
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}
	if val != 9 {
		t.Errorf("expected 9, got %d", val)
	}

	val, err = repo.Decrement("counter", 3)
	if err != nil {
		t.Fatalf("Decrement failed: %v", err)
	}
	if val != 6 {
		t.Errorf("expected 6, got %d", val)
	}
}

func TestRepository_Remember(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	callCount := 0
	fn := func() any {
		callCount++
		return "computed"
	}

	// First call should execute fn
	val := repo.Remember("key1", 1*time.Minute, fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to be called once, got %d", callCount)
	}

	// Second call should return cached value
	val = repo.Remember("key1", 1*time.Minute, fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to not be called again, got %d", callCount)
	}
}

func TestRepository_RememberForever(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	callCount := 0
	fn := func() any {
		callCount++
		return "computed"
	}

	// First call should execute fn
	val := repo.RememberForever("key1", fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to be called once, got %d", callCount)
	}

	// Second call should return cached value
	val = repo.RememberForever("key1", fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to not be called again, got %d", callCount)
	}
}

func TestRepository_Add(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Add should succeed for new key
	ok := repo.Add("key1", "value1", 1*time.Minute)
	if !ok {
		t.Error("expected Add to succeed for new key")
	}

	// Add should fail for existing key
	ok = repo.Add("key1", "value2", 1*time.Minute)
	if ok {
		t.Error("expected Add to fail for existing key")
	}

	// Value should remain unchanged
	val := repo.Get("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestRepository_Pull(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Pull
	repo.Put("key1", "value1", 1*time.Minute)

	val := repo.Pull("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	if repo.Has("key1") {
		t.Error("expected key to be deleted after Pull")
	}
}

func TestRepository_GetString(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test GetString
	repo.Put("key1", "value1", 1*time.Minute)

	val := repo.GetString("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}

	val = repo.GetString("missing", "default")
	if val != "default" {
		t.Errorf("expected default, got %v", val)
	}
}

func TestRepository_GetInt(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test GetInt
	repo.Put("key1", 42, 1*time.Minute)

	val := repo.GetInt("key1")
	if val != 42 {
		t.Errorf("expected 42, got %v", val)
	}

	val = repo.GetInt("missing", 99)
	if val != 99 {
		t.Errorf("expected 99, got %v", val)
	}
}

func TestRepository_Many(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test Many
	repo.Put("key1", "value1", 1*time.Minute)
	repo.Put("key2", "value2", 1*time.Minute)

	values := repo.Many([]string{"key1", "key2", "key3"})
	if len(values) != 2 {
		t.Errorf("expected 2 values, got %d", len(values))
	}
	if values["key1"] != "value1" {
		t.Errorf("expected value1, got %v", values["key1"])
	}
	if values["key2"] != "value2" {
		t.Errorf("expected value2, got %v", values["key2"])
	}
}

func TestRepository_PutMany(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	// Test PutMany
	values := map[string]any{
		"key1": "value1",
		"key2": "value2",
	}

	err := repo.PutMany(values, 1*time.Minute)
	if err != nil {
		t.Fatalf("PutMany failed: %v", err)
	}

	if !repo.Has("key1") || !repo.Has("key2") {
		t.Error("expected both keys to exist")
	}
}

func TestFileStore_PutGet(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "cache-test")
	defer os.RemoveAll(dir)

	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	repo := NewRepository(store)

	// Test Put and Get
	err = repo.Put("key1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	val := repo.Get("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestFileStore_TTLExpiration(t *testing.T) {
	dir := filepath.Join(os.TempDir(), "cache-test-ttl")
	defer os.RemoveAll(dir)

	store, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	repo := NewRepository(store)

	// Test TTL expiration
	repo.Put("key1", "value1", 100*time.Millisecond)

	if !repo.Has("key1") {
		t.Error("expected key to exist immediately after Put")
	}

	time.Sleep(150 * time.Millisecond)

	if repo.Has("key1") {
		t.Error("expected key to be expired")
	}
}

func TestNullStore_AlwaysMisses(t *testing.T) {
	store := NewNullStore()
	repo := NewRepository(store)

	// Test that NullStore always misses
	repo.Put("key1", "value1", 1*time.Minute)

	if repo.Has("key1") {
		t.Error("expected NullStore to always return false for Has")
	}

	val := repo.Get("key1")
	if val != nil {
		t.Error("expected NullStore to always return nil for Get")
	}
}

func TestLock_Get(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	lock := repo.Lock("resource", 1*time.Minute)

	executed := false
	ok := lock.Get(func() {
		executed = true
	})

	if !ok {
		t.Error("expected lock to be acquired")
	}
	if !executed {
		t.Error("expected callback to be executed")
	}
}

func TestLock_Block(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	lock1 := repo.Lock("resource", 1*time.Minute)
	lock2 := repo.Lock("resource", 1*time.Minute)

	// Acquire first lock
	executed1 := false
	go func() {
		lock1.Get(func() {
			time.Sleep(200 * time.Millisecond)
			executed1 = true
		})
	}()

	time.Sleep(50 * time.Millisecond)

	// Second lock should block
	executed2 := false
	ok := lock2.Block(500*time.Millisecond, func() {
		executed2 = true
	})

	if !ok {
		t.Error("expected lock to be acquired after waiting")
	}
	if !executed1 || !executed2 {
		t.Error("expected both callbacks to be executed")
	}
}

func TestLock_Release(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	lock := repo.Lock("resource", 1*time.Minute)

	// Acquire lock
	if !lock.acquire() {
		t.Fatal("expected lock to be acquired")
	}

	// Release lock
	if !lock.Release() {
		t.Error("expected Release to succeed")
	}

	// Lock should be available again
	if !lock.acquire() {
		t.Error("expected lock to be available after release")
	}
}

func TestTaggedCache_PutGet(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	tagged := repo.Tags([]string{"users", "posts"})

	// Test Put and Get
	err := tagged.Put("key1", "value1", 1*time.Minute)
	if err != nil {
		t.Fatalf("Put failed: %v", err)
	}

	val := tagged.Get("key1")
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}

func TestTaggedCache_Flush(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	tagged := repo.Tags([]string{"users"})

	// Store multiple values
	tagged.Put("key1", "value1", 1*time.Minute)
	tagged.Put("key2", "value2", 1*time.Minute)

	// Flush by tag
	err := tagged.Flush()
	if err != nil {
		t.Fatalf("Flush failed: %v", err)
	}

	if tagged.Has("key1") || tagged.Has("key2") {
		t.Error("expected all tagged keys to be flushed")
	}
}

func TestTaggedCache_Remember(t *testing.T) {
	store := NewMemoryStore()
	repo := NewRepository(store)

	tagged := repo.Tags([]string{"users"})

	callCount := 0
	fn := func() any {
		callCount++
		return "computed"
	}

	// First call should execute fn
	val := tagged.Remember("key1", 1*time.Minute, fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to be called once, got %d", callCount)
	}

	// Second call should return cached value
	val = tagged.Remember("key1", 1*time.Minute, fn)
	if val != "computed" {
		t.Errorf("expected computed, got %v", val)
	}
	if callCount != 1 {
		t.Errorf("expected fn to not be called again, got %d", callCount)
	}
}

func TestManager_Store(t *testing.T) {
	manager := NewManager()

	// Register a file store
	dir := filepath.Join(os.TempDir(), "cache-manager-test")
	defer os.RemoveAll(dir)

	fileStore, err := NewFileStore(dir)
	if err != nil {
		t.Fatalf("NewFileStore failed: %v", err)
	}
	manager.Register("file", fileStore)

	// Get file store
	repo := manager.Store("file")
	repo.Put("key1", "value1", 1*time.Minute)

	if !repo.Has("key1") {
		t.Error("expected key to exist in file store")
	}

	// Get default store
	defaultRepo := manager.Store("")
	defaultRepo.Put("key2", "value2", 1*time.Minute)

	if !defaultRepo.Has("key2") {
		t.Error("expected key to exist in default store")
	}
}

func TestManager_SetDefault(t *testing.T) {
	manager := NewManager()

	// Register null store
	manager.Register("null", NewNullStore())

	// Set null as default
	err := manager.SetDefault("null")
	if err != nil {
		t.Fatalf("SetDefault failed: %v", err)
	}

	// Default store should be null
	repo := manager.Store("")
	repo.Put("key1", "value1", 1*time.Minute)

	if repo.Has("key1") {
		t.Error("expected null store to always miss")
	}
}
