package logging

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Entry represents a single log entry with all its metadata.
type Entry struct {
	// Level is the severity level of the log entry.
	Level Level
	// Message is the log message.
	Message string
	// Context contains additional structured data for the log entry.
	Context map[string]any
	// Timestamp is when the log entry was created.
	Timestamp time.Time
	// Channel is the name of the logging channel.
	Channel string
}

// Formatter defines the interface for formatting log entries.
type Formatter interface {
	// Format converts a log entry into a string representation.
	Format(entry *Entry) string
}

// LineFormatter formats log entries as single-line text with context.
// Format: [2024-01-01 12:00:00] channel.LEVEL: message {"key": "val"}
type LineFormatter struct {
	// DateFormat specifies the time format. Defaults to "2006-01-02 15:04:05".
	DateFormat string
}

// NewLineFormatter creates a new line formatter with default settings.
func NewLineFormatter() *LineFormatter {
	return &LineFormatter{
		DateFormat: "2006-01-02 15:04:05",
	}
}

// Format converts a log entry into a formatted line string.
func (f *LineFormatter) Format(entry *Entry) string {
	var sb strings.Builder

	// Timestamp
	sb.WriteString("[")
	sb.WriteString(entry.Timestamp.Format(f.DateFormat))
	sb.WriteString("] ")

	// Channel and level
	sb.WriteString(entry.Channel)
	sb.WriteString(".")
	sb.WriteString(entry.Level.String())
	sb.WriteString(": ")

	// Message
	sb.WriteString(entry.Message)

	// Context (if present)
	if len(entry.Context) > 0 {
		contextJSON, _ := json.Marshal(entry.Context)
		sb.WriteString(" ")
		sb.WriteString(string(contextJSON))
	}

	sb.WriteString("\n")
	return sb.String()
}

// JSONFormatter formats log entries as JSON objects.
type JSONFormatter struct{}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter() *JSONFormatter {
	return &JSONFormatter{}
}

// Format converts a log entry into a JSON string.
func (f *JSONFormatter) Format(entry *Entry) string {
	data := map[string]any{
		"timestamp": entry.Timestamp.Format(time.RFC3339),
		"channel":   entry.Channel,
		"level":     entry.Level.String(),
		"message":   entry.Message,
	}

	if len(entry.Context) > 0 {
		data["context"] = entry.Context
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		// Fallback to simple format if JSON encoding fails
		return fmt.Sprintf(`{"timestamp":"%s","channel":"%s","level":"%s","message":"%s","error":"json_encode_failed"}%s`,
			entry.Timestamp.Format(time.RFC3339),
			entry.Channel,
			entry.Level.String(),
			entry.Message,
			"\n")
	}

	return string(jsonData) + "\n"
}
