package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewResponse(t *testing.T) {
	resp := NewResponse()

	if resp == nil {
		t.Fatal("NewResponse returned nil")
	}

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("Default status = %d; want %d", resp.GetStatusCode(), http.StatusOK)
	}
}

func TestMake(t *testing.T) {
	body := []byte("Hello, World!")
	resp := Make(body, http.StatusCreated)

	if resp.GetStatusCode() != http.StatusCreated {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusCreated)
	}

	if string(resp.GetBody()) != "Hello, World!" {
		t.Errorf("Body = %s; want Hello, World!", string(resp.GetBody()))
	}
}

func TestJSON(t *testing.T) {
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
		"age":   25,
	}

	resp := JSON(data, http.StatusOK)

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusOK)
	}

	if resp.headers.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %s; want application/json", resp.headers.Get("Content-Type"))
	}

	var decoded map[string]any
	if err := json.Unmarshal(resp.GetBody(), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON: %v", err)
	}

	if decoded["name"] != "John" {
		t.Errorf("JSON name = %v; want John", decoded["name"])
	}
}

func TestJSONWithCustomStatus(t *testing.T) {
	data := map[string]string{"error": "Not found"}
	resp := JSON(data, http.StatusNotFound)

	if resp.GetStatusCode() != http.StatusNotFound {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusNotFound)
	}
}

func TestRedirect(t *testing.T) {
	resp := Redirect("/dashboard", http.StatusSeeOther)

	if resp.GetStatusCode() != http.StatusSeeOther {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusSeeOther)
	}

	if location := resp.headers.Get("Location"); location != "/dashboard" {
		t.Errorf("Location = %s; want /dashboard", location)
	}
}

func TestRedirectDefaultStatus(t *testing.T) {
	resp := Redirect("/home")

	if resp.GetStatusCode() != http.StatusFound {
		t.Errorf("Status = %d; want %d (default)", resp.GetStatusCode(), http.StatusFound)
	}
}

func TestNoContent(t *testing.T) {
	resp := NoContent()

	if resp.GetStatusCode() != http.StatusNoContent {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusNoContent)
	}

	if resp.GetBody() != nil {
		t.Error("Body should be nil for NoContent")
	}
}

func TestResponseHeader(t *testing.T) {
	resp := NewResponse().
		Header("X-Custom-Header", "CustomValue").
		Header("X-Another-Header", "AnotherValue")

	if resp.headers.Get("X-Custom-Header") != "CustomValue" {
		t.Error("Header X-Custom-Header not set correctly")
	}

	if resp.headers.Get("X-Another-Header") != "AnotherValue" {
		t.Error("Header X-Another-Header not set correctly")
	}
}

func TestResponseCookie(t *testing.T) {
	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
		Path:  "/",
	}

	resp := NewResponse().Cookie(cookie)

	if len(resp.cookies) != 1 {
		t.Errorf("Cookies count = %d; want 1", len(resp.cookies))
	}

	if resp.cookies[0].Name != "session" {
		t.Errorf("Cookie name = %s; want session", resp.cookies[0].Name)
	}

	if resp.cookies[0].Value != "abc123" {
		t.Errorf("Cookie value = %s; want abc123", resp.cookies[0].Value)
	}
}

func TestResponseStatus(t *testing.T) {
	resp := NewResponse().Status(http.StatusAccepted)

	if resp.GetStatusCode() != http.StatusAccepted {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusAccepted)
	}
}

func TestResponseChaining(t *testing.T) {
	resp := NewResponse().
		Status(http.StatusCreated).
		Header("X-Custom", "Value").
		SetBody([]byte("Created"))

	if resp.GetStatusCode() != http.StatusCreated {
		t.Error("Chaining failed for Status")
	}

	if resp.headers.Get("X-Custom") != "Value" {
		t.Error("Chaining failed for Header")
	}

	if string(resp.GetBody()) != "Created" {
		t.Error("Chaining failed for SetBody")
	}
}

func TestResponseSend(t *testing.T) {
	resp := JSON(map[string]string{"message": "success"}, http.StatusOK).
		Header("X-Custom", "Test")

	w := httptest.NewRecorder()
	resp.Send(w)

	result := w.Result()

	if result.StatusCode != http.StatusOK {
		t.Errorf("HTTP status = %d; want %d", result.StatusCode, http.StatusOK)
	}

	if result.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type header not set")
	}

	if result.Header.Get("X-Custom") != "Test" {
		t.Error("Custom header not set")
	}

	var decoded map[string]string
	if err := json.NewDecoder(result.Body).Decode(&decoded); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if decoded["message"] != "success" {
		t.Error("Response body not correct")
	}
}

func TestResponseSendWithCookies(t *testing.T) {
	cookie := &http.Cookie{
		Name:  "token",
		Value: "xyz789",
	}

	resp := NewResponse().Cookie(cookie)
	w := httptest.NewRecorder()
	resp.Send(w)

	result := w.Result()
	cookies := result.Cookies()

	if len(cookies) != 1 {
		t.Fatalf("Cookies count = %d; want 1", len(cookies))
	}

	if cookies[0].Name != "token" {
		t.Error("Cookie not set correctly")
	}
}

func TestDownload(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	content := []byte("test file content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	resp := Download(tmpFile)

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusOK)
	}

	if string(resp.GetBody()) != "test file content" {
		t.Errorf("Body = %s; want test file content", string(resp.GetBody()))
	}

	contentDisposition := resp.headers.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=\"test.txt\"" {
		t.Errorf("Content-Disposition = %s; want attachment with filename", contentDisposition)
	}

	if resp.headers.Get("Content-Type") != "application/octet-stream" {
		t.Error("Content-Type should be application/octet-stream")
	}
}

func TestDownloadWithCustomFilename(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "original.txt")
	content := []byte("test content")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	resp := Download(tmpFile, "custom-name.txt")

	contentDisposition := resp.headers.Get("Content-Disposition")
	if contentDisposition != "attachment; filename=\"custom-name.txt\"" {
		t.Errorf("Content-Disposition = %s; want custom filename", contentDisposition)
	}
}

func TestDownloadNonExistent(t *testing.T) {
	resp := Download("/non/existent/file.txt")

	if resp.GetStatusCode() != http.StatusNotFound {
		t.Errorf("Status = %d; want %d for missing file", resp.GetStatusCode(), http.StatusNotFound)
	}
}

func TestFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.html")
	content := []byte("<html><body>Hello</body></html>")

	if err := os.WriteFile(tmpFile, content, 0644); err != nil {
		t.Fatal(err)
	}

	resp := File(tmpFile)

	if resp.GetStatusCode() != http.StatusOK {
		t.Errorf("Status = %d; want %d", resp.GetStatusCode(), http.StatusOK)
	}

	if string(resp.GetBody()) != string(content) {
		t.Error("File content mismatch")
	}

	// Should have Content-Type set (detected)
	if resp.headers.Get("Content-Type") == "" {
		t.Error("Content-Type should be set")
	}
}

func TestFileNonExistent(t *testing.T) {
	resp := File("/non/existent/file.html")

	if resp.GetStatusCode() != http.StatusNotFound {
		t.Errorf("Status = %d; want %d for missing file", resp.GetStatusCode(), http.StatusNotFound)
	}
}

func TestSetBody(t *testing.T) {
	resp := NewResponse()
	body := []byte("Updated body")
	resp.SetBody(body)

	if string(resp.GetBody()) != "Updated body" {
		t.Errorf("Body = %s; want Updated body", string(resp.GetBody()))
	}
}

func TestGetHeaders(t *testing.T) {
	resp := NewResponse().
		Header("X-First", "Value1").
		Header("X-Second", "Value2")

	headers := resp.GetHeaders()

	if headers.Get("X-First") != "Value1" {
		t.Error("GetHeaders() did not return correct headers")
	}

	if headers.Get("X-Second") != "Value2" {
		t.Error("GetHeaders() did not return correct headers")
	}
}
