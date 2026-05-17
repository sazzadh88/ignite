package gate

import (
	"net/http"
)

// CanFunc returns a middleware function that authorizes the given ability.
// If the authorization check fails, it returns a 403 Forbidden response.
// The middleware expects the user to be set in the Gate via SetUser or to be retrieved from context.
func CanFunc(gate *Gate, ability string, args ...any) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if gate.Denies(ability, args...) {
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte("This action is unauthorized."))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// Can returns a middleware function that authorizes the given ability using the global Access gate.
func Can(ability string, args ...any) func(http.Handler) http.Handler {
	return CanFunc(Access, ability, args...)
}
