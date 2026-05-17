package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test helper to create a simple middleware that adds a header
func headerMiddleware(key, value string) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		w.Header().Set(key, value)
		next(w, r)
	})
}

// Test helper to create middleware that can short-circuit
func shortCircuitMiddleware(statusCode int) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		w.WriteHeader(statusCode)
		w.Write([]byte("short-circuited"))
		// Don't call next() - pipeline stops here
	})
}

func TestMiddlewareFunc_Handle(t *testing.T) {
	called := false
	mw := MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		called = true
		next(w, r)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if !called {
		t.Error("MiddlewareFunc.Handle should call the underlying function")
	}
}

func TestPipeline_BasicChaining(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	pipeline := NewPipeline()
	handler := pipeline.
		Send(req).
		Through([]Middleware{
			headerMiddleware("X-First", "1"),
			headerMiddleware("X-Second", "2"),
		}).
		Then(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
		})

	handler(rec, req)

	if rec.Header().Get("X-First") != "1" {
		t.Error("First middleware should set X-First header")
	}
	if rec.Header().Get("X-Second") != "2" {
		t.Error("Second middleware should set X-Second header")
	}
	if rec.Body.String() != "success" {
		t.Errorf("Expected 'success', got %s", rec.Body.String())
	}
}

func TestPipeline_ShortCircuit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	finalHandlerCalled := false

	pipeline := NewPipeline()
	handler := pipeline.
		Send(req).
		Through([]Middleware{
			headerMiddleware("X-First", "1"),
			shortCircuitMiddleware(http.StatusUnauthorized),
			headerMiddleware("X-Third", "3"), // Should not execute
		}).
		Then(func(w http.ResponseWriter, r *http.Request) {
			finalHandlerCalled = true
		})

	handler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Body.String() != "short-circuited" {
		t.Errorf("Expected 'short-circuited', got %s", rec.Body.String())
	}
	if rec.Header().Get("X-Third") != "" {
		t.Error("Middleware after short-circuit should not execute")
	}
	if finalHandlerCalled {
		t.Error("Final handler should not be called after short-circuit")
	}
}

func TestPipeline_ThenReturn(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	pipeline := NewPipeline()
	handler := pipeline.
		Send(req).
		Through([]Middleware{
			MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("from middleware"))
			}),
		}).
		ThenReturn()

	handler(rec, req)

	if rec.Body.String() != "from middleware" {
		t.Errorf("Expected 'from middleware', got %s", rec.Body.String())
	}
}

func TestPipeline_EmptyMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	pipeline := NewPipeline()
	handler := pipeline.
		Through([]Middleware{}).
		Then(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("direct"))
		})

	handler(rec, req)

	if rec.Body.String() != "direct" {
		t.Errorf("Expected 'direct', got %s", rec.Body.String())
	}
}

func TestStack_GlobalMiddleware(t *testing.T) {
	stack := NewStack()

	stack.PushGlobal(headerMiddleware("X-A", "a"), 2)
	stack.PushGlobal(headerMiddleware("X-B", "b"), 1)
	stack.PushGlobal(headerMiddleware("X-C", "c"), 0) // Auto-assigned priority 3

	global := stack.Global()

	if len(global) != 3 {
		t.Fatalf("Expected 3 global middleware, got %d", len(global))
	}

	// Test priority ordering
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	pipeline := NewPipeline().Through(global)
	handler := pipeline.Then(func(w http.ResponseWriter, r *http.Request) {})
	handler(rec, req)

	// X-B should be set first (priority 1), then X-A (priority 2), then X-C (priority 3)
	if rec.Header().Get("X-B") != "b" || rec.Header().Get("X-A") != "a" || rec.Header().Get("X-C") != "c" {
		t.Error("Global middleware not executed in priority order")
	}
}

func TestStack_NamedMiddleware(t *testing.T) {
	stack := NewStack()

	stack.Register("auth", headerMiddleware("X-Auth", "true"))
	stack.Register("log", headerMiddleware("X-Log", "true"))

	if stack.Get("auth") == nil {
		t.Error("Should retrieve registered middleware")
	}
	if stack.Get("nonexistent") != nil {
		t.Error("Should return nil for nonexistent middleware")
	}

	middlewares := stack.Resolve([]string{"auth", "log"})
	if len(middlewares) != 2 {
		t.Fatalf("Expected 2 resolved middleware, got %d", len(middlewares))
	}
}

func TestStack_Groups(t *testing.T) {
	stack := NewStack()

	stack.Register("auth", headerMiddleware("X-Auth", "true"))
	stack.Register("csrf", headerMiddleware("X-CSRF", "true"))
	stack.Register("log", headerMiddleware("X-Log", "true"))

	stack.Group("web", []string{"auth", "csrf", "log"})
	stack.Group("api", []string{"auth", "log"})

	webGroup := stack.GetGroup("web")
	if len(webGroup) != 3 {
		t.Fatalf("Expected 3 middleware in web group, got %d", len(webGroup))
	}

	apiGroup := stack.GetGroup("api")
	if len(apiGroup) != 2 {
		t.Fatalf("Expected 2 middleware in api group, got %d", len(apiGroup))
	}

	// Test resolving group by name
	resolved := stack.Resolve([]string{"web"})
	if len(resolved) != 3 {
		t.Fatalf("Expected 3 resolved middleware from group, got %d", len(resolved))
	}
}

func TestStack_ResolveWithGroups(t *testing.T) {
	stack := NewStack()

	stack.Register("auth", headerMiddleware("X-Auth", "true"))
	stack.Register("log", headerMiddleware("X-Log", "true"))
	stack.Register("custom", headerMiddleware("X-Custom", "true"))

	stack.Group("web", []string{"auth", "log"})

	// Resolve mix of group and named middleware
	resolved := stack.Resolve([]string{"web", "custom"})
	if len(resolved) != 3 {
		t.Fatalf("Expected 3 resolved middleware, got %d", len(resolved))
	}
}

func TestTrimStrings(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test?name=++John++&email=+test@test.com+", nil)
	rec := httptest.NewRecorder()

	mw := TrimStrings()
	mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") != "John" {
			t.Errorf("Expected 'John', got '%s'", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("email") != "test@test.com" {
			t.Errorf("Expected 'test@test.com', got '%s'", r.URL.Query().Get("email"))
		}
	})
}

func TestCORS(t *testing.T) {
	config := DefaultCORSConfig()
	mw := CORS(config)

	t.Run("preflight", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
			t.Error("Next should not be called for OPTIONS request")
		})

		if rec.Code != http.StatusNoContent {
			t.Errorf("Expected status %d, got %d", http.StatusNoContent, rec.Code)
		}
		if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
			t.Error("Should set CORS headers")
		}
	})

	t.Run("regular request", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("Origin", "http://example.com")
		rec := httptest.NewRecorder()

		nextCalled := false
		mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
			nextCalled = true
		})

		if !nextCalled {
			t.Error("Next should be called for non-OPTIONS request")
		}
		if rec.Header().Get("Access-Control-Allow-Origin") == "" {
			t.Error("Should set CORS headers")
		}
	})
}

func TestRecovery(t *testing.T) {
	mw := Recovery()

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Internal Server Error") {
		t.Error("Should return error message")
	}
}

func TestLogger(t *testing.T) {
	var logOutput string
	config := LoggerConfig{
		Output: func(format string, v ...interface{}) {
			logOutput = fmt.Sprintf(format, v...)
		},
	}

	mw := Logger(config)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if logOutput == "" {
		t.Error("Logger should produce output")
	}
	if !strings.Contains(logOutput, "GET") {
		t.Error("Log should contain HTTP method")
	}
	if !strings.Contains(logOutput, "/test") {
		t.Error("Log should contain path")
	}
}

func TestLogger_SkipPaths(t *testing.T) {
	var logOutput string
	config := LoggerConfig{
		SkipPaths: []string{"/health"},
		Output: func(format string, v ...interface{}) {
			logOutput = fmt.Sprintf(format, v...)
		},
	}

	mw := Logger(config)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	if logOutput != "" {
		t.Error("Logger should skip configured paths")
	}
}

func TestRequestID(t *testing.T) {
	mw := RequestID()

	t.Run("generates new ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Request-ID") == "" {
				t.Error("Should set X-Request-ID in request")
			}
		})

		if rec.Header().Get("X-Request-ID") == "" {
			t.Error("Should set X-Request-ID in response")
		}
	})

	t.Run("preserves existing ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Request-ID", "existing-id")
		rec := httptest.NewRecorder()

		mw.Handle(rec, req, func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("X-Request-ID") != "existing-id" {
				t.Error("Should preserve existing X-Request-ID")
			}
		})
	})
}

func TestTerminable(t *testing.T) {
	terminateCalled := false

	type testTerminable struct {
		MiddlewareFunc
	}

	tm := &testTerminable{
		MiddlewareFunc: func(w http.ResponseWriter, r *http.Request, next Next) {
			next(w, r)
		},
	}

	// Implement Terminate to make it terminable
	terminateFunc := func(r *http.Request, w http.ResponseWriter) {
		terminateCalled = true
	}

	// Manual implementation of Terminable interface
	var _ Terminable = terminableWrapper{tm, terminateFunc}

	middlewares := []Middleware{terminableWrapper{tm, terminateFunc}}

	handler := WrapTerminable(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}, middlewares)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)

	if !terminateCalled {
		t.Error("Terminate should be called after handler completes")
	}
}

// Helper type to satisfy Terminable interface in tests
type terminableWrapper struct {
	Middleware
	terminate func(*http.Request, http.ResponseWriter)
}

func (t terminableWrapper) Terminate(r *http.Request, w http.ResponseWriter) {
	t.terminate(r, w)
}

func TestResponseRecorder(t *testing.T) {
	rec := httptest.NewRecorder()
	recorder := newResponseRecorder(rec)

	// Test WriteHeader
	recorder.WriteHeader(http.StatusCreated)
	if recorder.statusCode != http.StatusCreated {
		t.Errorf("Expected status code %d, got %d", http.StatusCreated, recorder.statusCode)
	}

	// Test Write
	recorder.Write([]byte("test"))
	if !recorder.written {
		t.Error("Should mark as written after Write")
	}
}

func TestResponseRecorder_DefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	recorder := newResponseRecorder(rec)

	recorder.Write([]byte("test"))
	if recorder.statusCode != http.StatusOK {
		t.Errorf("Expected default status %d, got %d", http.StatusOK, recorder.statusCode)
	}
}
