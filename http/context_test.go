package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewContext(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)

	if ctx == nil {
		t.Fatal("NewContext returned nil")
	}

	if ctx.Request == nil {
		t.Error("Context.Request is nil")
	}

	if ctx.Writer == nil {
		t.Error("Context.Writer is nil")
	}
}

func TestContextParam(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/users/123", nil)

	ctx := NewContext(w, r)
	ctx.SetParam("id", "123")

	if id := ctx.Param("id"); id != "123" {
		t.Errorf("Param(id) = %s; want 123", id)
	}
}

func TestContextJSON(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.JSON(http.StatusOK, map[string]string{"message": "success"})

	result := w.Result()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusOK)
	}

	if result.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}
}

func TestContextString(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.String(http.StatusOK, "Hello, World!")

	result := w.Result()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusOK)
	}

	if w.Body.String() != "Hello, World!" {
		t.Errorf("Body = %s; want Hello, World!", w.Body.String())
	}
}

func TestContextHTML(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.HTML(http.StatusOK, "<h1>Hello</h1>")

	result := w.Result()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusOK)
	}

	if result.Header.Get("Content-Type") != "text/html; charset=utf-8" {
		t.Error("Content-Type should be text/html")
	}

	if w.Body.String() != "<h1>Hello</h1>" {
		t.Error("Body content mismatch")
	}
}

func TestContextRedirect(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Redirect("/dashboard")

	result := w.Result()

	if result.StatusCode != http.StatusFound {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusFound)
	}

	if location := result.Header.Get("Location"); location != "/dashboard" {
		t.Errorf("Location = %s; want /dashboard", location)
	}
}

func TestContextRedirectWithCustomStatus(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Redirect("/home", http.StatusPermanentRedirect)

	result := w.Result()

	if result.StatusCode != http.StatusPermanentRedirect {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusPermanentRedirect)
	}
}

func TestContextBack(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	r.Header.Set("Referer", "/previous-page")

	ctx := NewContext(w, r)
	ctx.Back()

	result := w.Result()

	if location := result.Header.Get("Location"); location != "/previous-page" {
		t.Errorf("Location = %s; want /previous-page", location)
	}
}

func TestContextBackWithFallback(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	// No Referer set

	ctx := NewContext(w, r)
	ctx.Back("/fallback")

	result := w.Result()

	if location := result.Header.Get("Location"); location != "/fallback" {
		t.Errorf("Location = %s; want /fallback", location)
	}
}

func TestContextBackDefaultFallback(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)
	// No Referer set

	ctx := NewContext(w, r)
	ctx.Back()

	result := w.Result()

	if location := result.Header.Get("Location"); location != "/" {
		t.Errorf("Location = %s; want / (default)", location)
	}
}

func TestContextSetGet(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Set("user_id", 123)
	ctx.Set("username", "john")

	val, exists := ctx.Get("user_id")
	if !exists {
		t.Error("Get(user_id) should exist")
	}

	if val != 123 {
		t.Errorf("Get(user_id) = %v; want 123", val)
	}

	username, exists := ctx.Get("username")
	if !exists {
		t.Error("Get(username) should exist")
	}

	if username != "john" {
		t.Errorf("Get(username) = %v; want john", username)
	}

	_, exists = ctx.Get("missing")
	if exists {
		t.Error("Get(missing) should not exist")
	}
}

func TestContextMustGet(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Set("key", "value")

	val := ctx.MustGet("key")
	if val != "value" {
		t.Errorf("MustGet(key) = %v; want value", val)
	}
}

func TestContextMustGetPanics(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGet should panic for missing key")
		}
	}()

	ctx.MustGet("missing")
}

func TestContextHeader(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Header("X-Custom-Header", "CustomValue")

	if w.Header().Get("X-Custom-Header") != "CustomValue" {
		t.Error("Header not set correctly")
	}
}

func TestContextCookie(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
	}
	ctx.Cookie(cookie)

	result := w.Result()
	cookies := result.Cookies()

	if len(cookies) != 1 {
		t.Fatalf("Cookies count = %d; want 1", len(cookies))
	}

	if cookies[0].Name != "session" {
		t.Error("Cookie not set correctly")
	}
}

func TestContextError(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.Error(http.StatusBadRequest, "Invalid input")

	result := w.Result()

	if result.StatusCode != http.StatusBadRequest {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusBadRequest)
	}

	if result.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json for Error()")
	}
}

func TestContextNoContent(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/test", nil)

	ctx := NewContext(w, r)
	ctx.NoContent()

	result := w.Result()

	if result.StatusCode != http.StatusNoContent {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusNoContent)
	}
}

func TestContextSetContentType(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	ctx.SetContentType("application/xml")

	if w.Header().Get("Content-Type") != "application/xml" {
		t.Error("SetContentType did not set Content-Type header")
	}
}

func TestContextSend(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)
	response := JSON(map[string]string{"status": "ok"}, http.StatusOK)
	ctx.Send(response)

	result := w.Result()

	if result.StatusCode != http.StatusOK {
		t.Errorf("Status = %d; want %d", result.StatusCode, http.StatusOK)
	}

	if result.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}
}

func TestContextAbort(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/test", nil)

	ctx := NewContext(w, r)

	if ctx.IsAborted() {
		t.Error("Context should not be aborted initially")
	}

	ctx.Abort()

	if !ctx.IsAborted() {
		t.Error("Context should be aborted after calling Abort()")
	}
}
