package session

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func TestSessionPutGet(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key", "value")
	if got := sess.GetString("key"); got != "value" {
		t.Errorf("Get() = %v, want %v", got, "value")
	}
}

func TestSessionPutMany(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	data := map[string]any{
		"key1": "value1",
		"key2": 42,
	}
	sess.PutMany(data)

	if got := sess.GetString("key1"); got != "value1" {
		t.Errorf("GetString(key1) = %v, want %v", got, "value1")
	}
	if got := sess.GetInt("key2"); got != 42 {
		t.Errorf("GetInt(key2) = %v, want %v", got, 42)
	}
}

func TestSessionHasMissingExists(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key", "value")
	sess.Put("nil-key", nil)

	if !sess.Has("key") {
		t.Error("Has(key) = false, want true")
	}
	if sess.Missing("key") {
		t.Error("Missing(key) = true, want false")
	}
	if !sess.Exists("key") {
		t.Error("Exists(key) = false, want true")
	}

	if sess.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}
	if !sess.Missing("missing") {
		t.Error("Missing(missing) = false, want true")
	}

	if sess.Exists("nil-key") {
		t.Error("Exists(nil-key) = true, want false")
	}
}

func TestSessionPull(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key", "value")
	if got := sess.Pull("key"); got != "value" {
		t.Errorf("Pull() = %v, want %v", got, "value")
	}

	// Key should be removed
	if sess.Has("key") {
		t.Error("Has(key) = true after Pull, want false")
	}
}

func TestSessionFlash(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	// Set flash data
	sess.Flash("message", "Hello")
	if err := sess.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Start new session with same ID
	sess2 := NewSession(store, "test-id")
	if err := sess2.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Flash data should be available
	if got := sess2.GetString("message"); got != "Hello" {
		t.Errorf("GetString(message) = %v, want %v", got, "Hello")
	}

	// Save and start again
	if err := sess2.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	sess3 := NewSession(store, "test-id")
	if err := sess3.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	// Flash data should be gone
	if got := sess3.Get("message"); got != nil {
		t.Errorf("Get(message) = %v, want nil", got)
	}
}

func TestSessionReflash(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	// Set flash data
	sess.Flash("message", "Hello")
	if err := sess.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Start new session and reflash
	sess2 := NewSession(store, "test-id")
	if err := sess2.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	sess2.Reflash()
	if err := sess2.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Start again - flash data should still be available
	sess3 := NewSession(store, "test-id")
	if err := sess3.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if got := sess3.GetString("message"); got != "Hello" {
		t.Errorf("GetString(message) = %v, want %v", got, "Hello")
	}
}

func TestSessionKeepFlash(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	// Set multiple flash values
	sess.Flash("keep", "value1")
	sess.Flash("discard", "value2")
	if err := sess.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Start new session and keep only one flash
	sess2 := NewSession(store, "test-id")
	if err := sess2.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	sess2.KeepFlash("keep")
	if err := sess2.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Start again
	sess3 := NewSession(store, "test-id")
	if err := sess3.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if got := sess3.GetString("keep"); got != "value1" {
		t.Errorf("GetString(keep) = %v, want %v", got, "value1")
	}
	if got := sess3.Get("discard"); got != nil {
		t.Errorf("Get(discard) = %v, want nil", got)
	}
}

func TestSessionNow(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	// Set data with Now
	sess.Now("temp", "value")

	// Should be available immediately
	if got := sess.GetString("temp"); got != "value" {
		t.Errorf("GetString(temp) = %v, want %v", got, "value")
	}

	// Should not persist after save
	if err := sess.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	sess2 := NewSession(store, "test-id")
	if err := sess2.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if got := sess2.Get("temp"); got != nil {
		t.Errorf("Get(temp) = %v, want nil", got)
	}
}

func TestSessionRegenerate(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	originalID := sess.GetID()
	sess.Put("key", "value")

	newID := sess.Regenerate()

	if newID == originalID {
		t.Error("Regenerate() returned same ID")
	}
	if sess.GetID() != newID {
		t.Error("GetID() does not match regenerated ID")
	}

	// Data should be preserved
	if got := sess.GetString("key"); got != "value" {
		t.Errorf("GetString(key) = %v, want %v", got, "value")
	}
}

func TestSessionInvalidate(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	originalID := sess.GetID()
	sess.Put("key", "value")

	newID := sess.Invalidate()

	if newID == originalID {
		t.Error("Invalidate() returned same ID")
	}

	// Data should be cleared
	if got := sess.Get("key"); got != nil {
		t.Errorf("Get(key) = %v, want nil", got)
	}
}

func TestSessionFlush(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key1", "value1")
	sess.Put("key2", "value2")

	sess.Flush()

	if sess.Has("key1") || sess.Has("key2") {
		t.Error("Flush() did not clear all data")
	}
}

func TestSessionIncrement(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Increment("counter")
	if got := sess.GetInt("counter"); got != 1 {
		t.Errorf("GetInt(counter) = %v, want 1", got)
	}

	sess.Increment("counter", 5)
	if got := sess.GetInt("counter"); got != 6 {
		t.Errorf("GetInt(counter) = %v, want 6", got)
	}
}

func TestSessionDecrement(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("counter", 10)
	sess.Decrement("counter")
	if got := sess.GetInt("counter"); got != 9 {
		t.Errorf("GetInt(counter) = %v, want 9", got)
	}

	sess.Decrement("counter", 5)
	if got := sess.GetInt("counter"); got != 4 {
		t.Errorf("GetInt(counter) = %v, want 4", got)
	}
}

func TestSessionPush(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Push("items", "item1")
	sess.Push("items", "item2")

	items := sess.Get("items")
	if arr, ok := items.([]any); !ok || len(arr) != 2 {
		t.Errorf("Get(items) = %v, want []any with 2 elements", items)
	}
}

func TestSessionToken(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	if err := sess.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	token := sess.Token()
	if token == "" {
		t.Error("Token() returned empty string")
	}

	// Token should persist
	if err := sess.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	sess2 := NewSession(store, "test-id")
	if err := sess2.Start(); err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if sess2.Token() != token {
		t.Error("Token did not persist")
	}
}

func TestSessionAll(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key1", "value1")
	sess.Put("key2", 42)

	all := sess.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d items, want 2", len(all))
	}
	if all["key1"] != "value1" || all["key2"] != 42 {
		t.Error("All() returned incorrect data")
	}
}

func TestSessionOnly(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key1", "value1")
	sess.Put("key2", "value2")
	sess.Put("key3", "value3")

	only := sess.Only([]string{"key1", "key3"})
	if len(only) != 2 {
		t.Errorf("Only() returned %d items, want 2", len(only))
	}
	if only["key1"] != "value1" || only["key3"] != "value3" {
		t.Error("Only() returned incorrect data")
	}
}

func TestSessionForget(t *testing.T) {
	store := NewMemoryStore()
	sess := NewSession(store, "test-id")

	sess.Put("key1", "value1")
	sess.Put("key2", "value2")

	sess.Forget("key1")

	if sess.Has("key1") {
		t.Error("Forget() did not remove key1")
	}
	if !sess.Has("key2") {
		t.Error("Forget() removed key2")
	}
}

func TestFileStoreReadWrite(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	data := map[string]any{
		"key": "value",
	}

	if err := store.Write("test-id", data, time.Hour); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	read, err := store.Read("test-id")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if !reflect.DeepEqual(read, data) {
		t.Errorf("Read() = %v, want %v", read, data)
	}
}

func TestFileStoreDestroy(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	data := map[string]any{"key": "value"}
	if err := store.Write("test-id", data, time.Hour); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if err := store.Destroy("test-id"); err != nil {
		t.Fatalf("Destroy() error = %v", err)
	}

	// File should be gone
	filePath := filepath.Join(tmpDir, "sess_test-id")
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("Destroy() did not remove file")
	}
}

func TestFileStoreGC(t *testing.T) {
	tmpDir := t.TempDir()
	store, err := NewFileStore(tmpDir)
	if err != nil {
		t.Fatalf("NewFileStore() error = %v", err)
	}

	// Write expired session
	data := map[string]any{"key": "value"}
	if err := store.Write("old-session", data, -time.Hour); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Write valid session
	if err := store.Write("new-session", data, time.Hour); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	// Run GC
	if err := store.GC(30 * time.Minute); err != nil {
		t.Fatalf("GC() error = %v", err)
	}

	// Old session should be gone
	if read, err := store.Read("old-session"); err != nil {
		t.Fatalf("Read() error = %v", err)
	} else if len(read) != 0 {
		t.Error("GC() did not remove expired session")
	}

	// New session should remain
	if read, err := store.Read("new-session"); err != nil {
		t.Fatalf("Read() error = %v", err)
	} else if read["key"] != "value" {
		t.Error("GC() removed valid session")
	}
}

func TestMemoryStoreOperations(t *testing.T) {
	store := NewMemoryStore()

	data := map[string]any{
		"key": "value",
	}

	if err := store.Write("test-id", data, time.Hour); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	read, err := store.Read("test-id")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if !reflect.DeepEqual(read, data) {
		t.Errorf("Read() = %v, want %v", read, data)
	}

	if err := store.Destroy("test-id"); err != nil {
		t.Fatalf("Destroy() error = %v", err)
	}

	read, err = store.Read("test-id")
	if err != nil {
		t.Fatalf("Read() error = %v", err)
	}

	if len(read) != 0 {
		t.Error("Destroy() did not remove data")
	}
}

func TestManagerStart(t *testing.T) {
	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	session, err := manager.Start(req)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if session == nil {
		t.Error("Start() returned nil session")
	}
}

func TestManagerSave(t *testing.T) {
	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	req := httptest.NewRequest("GET", "/", nil)
	session, err := manager.Start(req)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	session.Put("key", "value")

	w := httptest.NewRecorder()
	if err := manager.Save(session, w); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Check cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("Save() did not set cookie")
	}
}

func TestStartSessionMiddleware(t *testing.T) {
	config := DefaultConfig()
	manager, err := NewManager(config)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	handler := StartSession(manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := FromRequest(r)
		if sess == nil {
			t.Error("FromRequest() returned nil")
			return
		}

		sess.Put("key", "value")
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Check cookie was set
	resp := w.Result()
	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("Middleware did not set cookie")
	}
}
