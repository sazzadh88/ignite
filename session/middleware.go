package session

import (
	"context"
	"net/http"
)

type contextKey string

const sessionKey contextKey = "session"

// StartSession is middleware that manages session lifecycle.
func StartSession(manager *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start session
			session, err := manager.Start(r)
			if err != nil {
				http.Error(w, "Failed to start session", http.StatusInternalServerError)
				return
			}

			// Add session to request context
			ctx := context.WithValue(r.Context(), sessionKey, session)
			r = r.WithContext(ctx)

			// Wrap response writer to save session after response
			sw := &sessionWriter{
				ResponseWriter: w,
				session:        session,
				manager:        manager,
			}

			// Call next handler
			next.ServeHTTP(sw, r)

			// Save session if not already saved
			if !sw.saved {
				_ = manager.Save(session, w)
			}
		})
	}
}

// sessionWriter wraps http.ResponseWriter to save session after writing.
type sessionWriter struct {
	http.ResponseWriter
	session *Session
	manager *Manager
	saved   bool
}

// WriteHeader saves the session before writing headers.
func (s *sessionWriter) WriteHeader(statusCode int) {
	if !s.saved {
		_ = s.manager.Save(s.session, s.ResponseWriter)
		s.saved = true
	}
	s.ResponseWriter.WriteHeader(statusCode)
}

// Write saves the session before writing data.
func (s *sessionWriter) Write(b []byte) (int, error) {
	if !s.saved {
		_ = s.manager.Save(s.session, s.ResponseWriter)
		s.saved = true
	}
	return s.ResponseWriter.Write(b)
}

// FromContext retrieves the session from the request context.
func FromContext(ctx context.Context) *Session {
	if sess, ok := ctx.Value(sessionKey).(*Session); ok {
		return sess
	}
	return nil
}

// FromRequest retrieves the session from the request.
func FromRequest(r *http.Request) *Session {
	return FromContext(r.Context())
}
