package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.WriteHeader(200)
		w.Write([]byte("success"))
	}))
	defer server.Close()

	client := New()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}

	if resp.Body() != "success" {
		t.Errorf("expected body 'success', got %s", resp.Body())
	}
}

func TestPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), "name") {
			t.Errorf("expected body to contain 'name', got %s", string(body))
		}

		w.WriteHeader(201)
		w.Write([]byte("created"))
	}))
	defer server.Close()

	client := New()
	resp, err := client.Post(server.URL, map[string]any{
		"name": "test",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 201 {
		t.Errorf("expected status 201, got %d", resp.Status())
	}
}

func TestPut(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("expected PUT, got %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New()
	resp, err := client.Put(server.URL, map[string]any{"data": "value"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestPatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New()
	resp, err := client.Patch(server.URL, map[string]any{"data": "value"})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(204)
	}))
	defer server.Close()

	client := New()
	resp, err := client.Delete(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 204 {
		t.Errorf("expected status 204, got %d", resp.Status())
	}
}

func TestJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		json.NewEncoder(w).Encode(map[string]any{
			"message": "hello",
			"code":    42,
		})
	}))
	defer server.Close()

	client := New()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	data, err := resp.JSON()
	if err != nil {
		t.Fatalf("expected no error parsing JSON, got %v", err)
	}

	if data["message"] != "hello" {
		t.Errorf("expected message 'hello', got %v", data["message"])
	}

	if int(data["code"].(float64)) != 42 {
		t.Errorf("expected code 42, got %v", data["code"])
	}
}

func TestJSONInto(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"name": "test",
		})
	}))
	defer server.Close()

	client := New()
	resp, _ := client.Get(server.URL)

	var result struct {
		Name string `json:"name"`
	}
	err := resp.JSONInto(&result)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.Name != "test" {
		t.Errorf("expected name 'test', got %s", result.Name)
	}
}

func TestWithToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-token" {
			t.Errorf("expected 'Bearer test-token', got %s", auth)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().WithToken("test-token")
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestWithBasicAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "testuser" || pass != "testpass" {
			t.Errorf("expected basic auth testuser:testpass, got %s:%s (ok=%v)", user, pass, ok)
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().WithBasicAuth("testuser", "testpass")
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestWithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "value" {
			t.Errorf("expected X-Custom header 'value', got %s", r.Header.Get("X-Custom"))
		}
		if r.Header.Get("X-Another") != "test" {
			t.Errorf("expected X-Another header 'test', got %s", r.Header.Get("X-Another"))
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().WithHeaders(map[string]string{
		"X-Custom":  "value",
		"X-Another": "test",
	})
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestWithHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "value" {
			t.Errorf("expected X-Test header 'value', got %s", r.Header.Get("X-Test"))
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().WithHeader("X-Test", "value")
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().Timeout(50 * time.Millisecond)
	_, err := client.Get(server.URL)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}

func TestRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(500)
		} else {
			w.WriteHeader(200)
			w.Write([]byte("success"))
		}
	}))
	defer server.Close()

	client := New().Retry(3, 10*time.Millisecond)
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error after retry, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}

	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestAcceptJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Accept") != "application/json" {
			t.Errorf("expected Accept header 'application/json', got %s", r.Header.Get("Accept"))
		}
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().AcceptJSON()
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestAsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %s", r.Header.Get("Content-Type"))
		}

		var data map[string]any
		json.NewDecoder(r.Body).Decode(&data)

		if data["name"] != "test" {
			t.Errorf("expected JSON field name='test', got %v", data["name"])
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().AsJSON()
	resp, err := client.Post(server.URL, map[string]any{
		"name": "test",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestAsForm(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contentType := r.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/x-www-form-urlencoded") {
			t.Errorf("expected Content-Type with 'application/x-www-form-urlencoded', got %s", contentType)
		}

		r.ParseForm()
		if r.Form.Get("name") != "test" {
			t.Errorf("expected form field name='test', got %s", r.Form.Get("name"))
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().AsForm()
	resp, err := client.Post(server.URL, map[string]any{
		"name": "test",
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestResponseStatusHelpers(t *testing.T) {
	tests := []struct {
		status      int
		successful  bool
		failed      bool
		serverError bool
		clientError bool
		redirect    bool
		ok          bool
		unauthorized bool
		forbidden   bool
		notFound    bool
	}{
		{200, true, false, false, false, false, true, false, false, false},
		{201, true, false, false, false, false, false, false, false, false},
		{301, false, true, false, false, true, false, false, false, false},
		{400, false, true, false, true, false, false, false, false, false},
		{401, false, true, false, true, false, false, true, false, false},
		{403, false, true, false, true, false, false, false, true, false},
		{404, false, true, false, true, false, false, false, false, true},
		{500, false, true, true, false, false, false, false, false, false},
	}

	for _, tt := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(tt.status)
		}))

		client := New()
		resp, _ := client.Get(server.URL)

		if resp.Successful() != tt.successful {
			t.Errorf("status %d: expected Successful=%v, got %v", tt.status, tt.successful, resp.Successful())
		}
		if resp.Failed() != tt.failed {
			t.Errorf("status %d: expected Failed=%v, got %v", tt.status, tt.failed, resp.Failed())
		}
		if resp.ServerError() != tt.serverError {
			t.Errorf("status %d: expected ServerError=%v, got %v", tt.status, tt.serverError, resp.ServerError())
		}
		if resp.ClientError() != tt.clientError {
			t.Errorf("status %d: expected ClientError=%v, got %v", tt.status, tt.clientError, resp.ClientError())
		}
		if resp.Redirect() != tt.redirect {
			t.Errorf("status %d: expected Redirect=%v, got %v", tt.status, tt.redirect, resp.Redirect())
		}
		if resp.Ok() != tt.ok {
			t.Errorf("status %d: expected Ok=%v, got %v", tt.status, tt.ok, resp.Ok())
		}
		if resp.Unauthorized() != tt.unauthorized {
			t.Errorf("status %d: expected Unauthorized=%v, got %v", tt.status, tt.unauthorized, resp.Unauthorized())
		}
		if resp.Forbidden() != tt.forbidden {
			t.Errorf("status %d: expected Forbidden=%v, got %v", tt.status, tt.forbidden, resp.Forbidden())
		}
		if resp.NotFound() != tt.notFound {
			t.Errorf("status %d: expected NotFound=%v, got %v", tt.status, tt.notFound, resp.NotFound())
		}

		server.Close()
	}
}

func TestFakeClient(t *testing.T) {
	fake := Fake(
		FakeResponse(map[string]string{"message": "first"}, 200),
		FakeResponse(map[string]string{"message": "second"}, 201),
	)

	// First request
	resp1, _ := fake.Get("http://example.com/test1")
	if resp1.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp1.Status())
	}

	data1, _ := resp1.JSON()
	if data1["message"] != "first" {
		t.Errorf("expected message 'first', got %v", data1["message"])
	}

	// Second request
	resp2, _ := fake.Post("http://example.com/test2", map[string]any{"data": "value"})
	if resp2.Status() != 201 {
		t.Errorf("expected status 201, got %d", resp2.Status())
	}

	data2, _ := resp2.JSON()
	if data2["message"] != "second" {
		t.Errorf("expected message 'second', got %v", data2["message"])
	}

	// Assert sent
	if !fake.AssertSentCount(2) {
		t.Error("expected 2 requests to be sent")
	}

	if !fake.AssertSent(func(req *http.Request) bool {
		return req.URL.Path == "/test1"
	}) {
		t.Error("expected request to /test1 to be sent")
	}

	if !fake.AssertNotSent(func(req *http.Request) bool {
		return req.URL.Path == "/nonexistent"
	}) {
		t.Error("expected no request to /nonexistent")
	}
}

func TestFakeClientSequence(t *testing.T) {
	fake := Sequence(
		FakeResponse("first", 200),
		FakeResponse("second", 200),
		FakeResponse("third", 200),
	)

	resp1, _ := fake.Get("http://example.com")
	if resp1.Body() != "first" {
		t.Errorf("expected 'first', got %s", resp1.Body())
	}

	resp2, _ := fake.Get("http://example.com")
	if resp2.Body() != "second" {
		t.Errorf("expected 'second', got %s", resp2.Body())
	}

	resp3, _ := fake.Get("http://example.com")
	if resp3.Body() != "third" {
		t.Errorf("expected 'third', got %s", resp3.Body())
	}
}

func TestWithoutRedirecting(t *testing.T) {
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "/final", http.StatusFound)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte("final"))
	}))
	defer redirectServer.Close()

	// With redirect following (default)
	client1 := New()
	resp1, _ := client1.Get(redirectServer.URL + "/redirect")
	if resp1.Status() != 200 {
		t.Errorf("expected status 200 after redirect, got %d", resp1.Status())
	}

	// Without redirect following
	client2 := New().WithoutRedirecting()
	resp2, _ := client2.Get(redirectServer.URL + "/redirect")
	if resp2.Status() != 302 {
		t.Errorf("expected status 302 without following redirect, got %d", resp2.Status())
	}
}

func TestWithCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie1, _ := r.Cookie("session")
		cookie2, _ := r.Cookie("token")

		if cookie1 == nil || cookie1.Value != "abc123" {
			t.Errorf("expected session cookie 'abc123', got %v", cookie1)
		}
		if cookie2 == nil || cookie2.Value != "xyz789" {
			t.Errorf("expected token cookie 'xyz789', got %v", cookie2)
		}

		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New().WithCookies(map[string]string{
		"session": "abc123",
		"token":   "xyz789",
	})
	resp, err := client.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Status() != 200 {
		t.Errorf("expected status 200, got %d", resp.Status())
	}
}

func TestResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Custom", "test-value")
		w.Header().Set("X-Another", "another-value")
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New()
	resp, _ := client.Get(server.URL)

	if resp.Header("X-Custom") != "test-value" {
		t.Errorf("expected X-Custom header 'test-value', got %s", resp.Header("X-Custom"))
	}

	headers := resp.Headers()
	if headers["X-Another"][0] != "another-value" {
		t.Errorf("expected X-Another header 'another-value', got %v", headers["X-Another"])
	}
}

func TestResponseCookies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "session", Value: "abc123"})
		http.SetCookie(w, &http.Cookie{Name: "token", Value: "xyz789"})
		w.WriteHeader(200)
	}))
	defer server.Close()

	client := New()
	resp, _ := client.Get(server.URL)

	cookies := resp.Cookies()
	if len(cookies) != 2 {
		t.Errorf("expected 2 cookies, got %d", len(cookies))
	}

	foundSession := false
	foundToken := false
	for _, cookie := range cookies {
		if cookie.Name == "session" && cookie.Value == "abc123" {
			foundSession = true
		}
		if cookie.Name == "token" && cookie.Value == "xyz789" {
			foundToken = true
		}
	}

	if !foundSession {
		t.Error("expected to find session cookie")
	}
	if !foundToken {
		t.Error("expected to find token cookie")
	}
}

func TestGlobalHttpFacade(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("facade works"))
	}))
	defer server.Close()

	resp, err := Http.Get(server.URL)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if resp.Body() != "facade works" {
		t.Errorf("expected 'facade works', got %s", resp.Body())
	}
}
