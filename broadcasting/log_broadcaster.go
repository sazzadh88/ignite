package broadcasting

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// LogBroadcaster is a broadcaster that writes events to an io.Writer.
// Useful for development and testing.
type LogBroadcaster struct {
	mu     sync.Mutex
	writer io.Writer
}

// NewLogBroadcaster creates a new log broadcaster writing to the specified writer.
// If writer is nil, defaults to os.Stdout.
func NewLogBroadcaster(writer io.Writer) *LogBroadcaster {
	if writer == nil {
		writer = os.Stdout
	}
	return &LogBroadcaster{writer: writer}
}

// Broadcast writes the broadcast event to the log writer.
func (l *LogBroadcaster) Broadcast(channels []Channel, event string, data map[string]any) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	channelNames := make([]string, len(channels))
	for i, ch := range channels {
		channelNames[i] = ch.Name
	}

	logEntry := map[string]any{
		"event":    event,
		"channels": channelNames,
		"data":     data,
	}

	jsonData, err := json.Marshal(logEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	_, err = fmt.Fprintf(l.writer, "[Broadcasting] %s\n", jsonData)
	return err
}

// Auth logs the authorization attempt and returns nil.
func (l *LogBroadcaster) Auth(request *http.Request, channel Channel) (any, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	logEntry := map[string]any{
		"action":  "auth",
		"channel": channel.Name,
		"type":    channel.Type,
	}

	jsonData, _ := json.Marshal(logEntry)
	fmt.Fprintf(l.writer, "[Broadcasting] %s\n", jsonData)

	return nil, nil
}
