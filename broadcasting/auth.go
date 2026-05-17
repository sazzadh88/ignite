package broadcasting

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// AuthCallback is a function that authorizes a user for a channel.
// It receives the authenticated user and channel parameters extracted from the channel name.
// It should return user data for presence channels, or any value for private channels.
// Return nil to deny authorization.
type AuthCallback func(user any, params map[string]string) any

type channelAuth struct {
	pattern  *regexp.Regexp
	callback AuthCallback
}

var (
	authMu       sync.RWMutex
	channelAuths []channelAuth
)

// RegisterChannel registers an authorization callback for a channel pattern.
// The pattern can contain named parameters like "chat.{id}" or "user.{userId}".
// The callback receives the authenticated user and extracted parameters.
func RegisterChannel(pattern string, callback AuthCallback) {
	authMu.Lock()
	defer authMu.Unlock()

	regex := patternToRegex(pattern)
	channelAuths = append(channelAuths, channelAuth{
		pattern:  regex,
		callback: callback,
	})
}

// AuthorizeChannel authorizes a user for a channel using registered callbacks.
// Returns the result from the callback (user data for presence, any value for private).
func AuthorizeChannel(user any, channel Channel) (any, error) {
	authMu.RLock()
	defer authMu.RUnlock()

	for _, auth := range channelAuths {
		if matches := auth.pattern.FindStringSubmatch(channel.Name); matches != nil {
			params := extractParams(auth.pattern, matches)
			result := auth.callback(user, params)
			if result == nil {
				return nil, fmt.Errorf("authorization denied for channel: %s", channel.Name)
			}
			return result, nil
		}
	}

	return nil, fmt.Errorf("no authorization handler found for channel: %s", channel.Name)
}

// AuthHandler is an HTTP handler for channel authorization.
// It expects the authenticated user to be in the request context.
func AuthHandler(userExtractor func(*http.Request) any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		channelName := r.FormValue("channel_name")
		if channelName == "" {
			http.Error(w, "channel_name is required", http.StatusBadRequest)
			return
		}

		channel := parseChannel(channelName)
		if channel.Type == Public {
			http.Error(w, "public channels do not require authorization", http.StatusBadRequest)
			return
		}

		user := userExtractor(r)
		if user == nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}

		result, err := AuthorizeChannel(user, channel)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		response := map[string]any{
			"auth":         true,
			"channel_data": result,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}
}

// parseChannel determines the channel type from its name.
func parseChannel(name string) Channel {
	switch {
	case strings.HasPrefix(name, "private-encrypted-"):
		return EncryptedPrivateChannel(name)
	case strings.HasPrefix(name, "private-"):
		return PrivateChannel(name)
	case strings.HasPrefix(name, "presence-"):
		return PresenceChannel(name)
	default:
		return PublicChannel(name)
	}
}

// patternToRegex converts a channel pattern to a regular expression.
// Converts "chat.{id}" to "^chat\.([^.]+)$"
func patternToRegex(pattern string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(pattern)
	regex := regexp.MustCompile(`\\{([^}]+)\\}`).ReplaceAllString(escaped, `(?P<$1>[^.]+)`)
	return regexp.MustCompile("^" + regex + "$")
}

// extractParams extracts named parameters from regex matches.
func extractParams(regex *regexp.Regexp, matches []string) map[string]string {
	params := make(map[string]string)
	names := regex.SubexpNames()

	for i, name := range names {
		if i > 0 && i < len(matches) && name != "" {
			params[name] = matches[i]
		}
	}

	return params
}
