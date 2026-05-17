package broadcasting

import "net/http"

// NullBroadcaster is a broadcaster that discards all broadcasts.
// Useful for testing or when broadcasting is disabled.
type NullBroadcaster struct{}

// Broadcast discards the broadcast.
func (n *NullBroadcaster) Broadcast(channels []Channel, event string, data map[string]any) error {
	return nil
}

// Auth always returns nil for authorization.
func (n *NullBroadcaster) Auth(request *http.Request, channel Channel) (any, error) {
	return nil, nil
}
