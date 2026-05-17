package logging

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestLogLevels(t *testing.T) {
	levels := []struct {
		level    Level
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{NOTICE, "NOTICE"},
		{WARNING, "WARNING"},
		{ERROR, "ERROR"},
		{CRITICAL, "CRITICAL"},
		{ALERT, "ALERT"},
		{EMERGENCY, "EMERGENCY"},
	}

	for _, tc := range levels {
		if tc.level.String() != tc.expected {
			t.Errorf("Level.String() = %q, want %q", tc.level.String(), tc.expected)
		}
	}
}

func TestLevelFiltering(t *testing.T) {
	tmpFile := "/tmp/test-level-filter.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("test", tmpFile, INFO, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	// These should be logged (>= INFO)
	channel.Info("info message")
	channel.Notice("notice message")
	channel.Error("error message")

	// This should NOT be logged (< INFO)
	channel.Debug("debug message")

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !strings.Contains(logContent, "info message") {
		t.Error("Expected info message to be logged")
	}

	if !strings.Contains(logContent, "notice message") {
		t.Error("Expected notice message to be logged")
	}

	if !strings.Contains(logContent, "error message") {
		t.Error("Expected error message to be logged")
	}

	if strings.Contains(logContent, "debug message") {
		t.Error("Debug message should not be logged when level is INFO")
	}
}

func TestFileChannelWritesToFile(t *testing.T) {
	tmpFile := "/tmp/test-file-channel.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("test", tmpFile, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	testMessage := "test log message"
	channel.Info(testMessage)

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), testMessage) {
		t.Errorf("Log file does not contain expected message: %q", testMessage)
	}

	if !strings.Contains(string(content), "test.INFO") {
		t.Error("Log file does not contain expected channel and level")
	}
}

func TestDailyChannelUsesDateInFilename(t *testing.T) {
	basePath := "/tmp/test-daily.log"
	today := time.Now().Format("2006-01-02")
	expectedPath := basePath + "-" + today
	defer os.Remove(expectedPath)

	channel := NewDailyChannel("daily", basePath, DEBUG, NewLineFormatter())
	defer channel.Close()

	testMessage := "daily log message"
	channel.Info(testMessage)

	content, err := os.ReadFile(expectedPath)
	if err != nil {
		t.Fatalf("Failed to read daily log file %q: %v", expectedPath, err)
	}

	if !strings.Contains(string(content), testMessage) {
		t.Errorf("Daily log file does not contain expected message: %q", testMessage)
	}
}

func TestStackChannelWritesToAll(t *testing.T) {
	tmpFile1 := "/tmp/test-stack-1.log"
	tmpFile2 := "/tmp/test-stack-2.log"
	defer os.Remove(tmpFile1)
	defer os.Remove(tmpFile2)

	channel1, err := NewFileChannel("channel1", tmpFile1, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel 1: %v", err)
	}
	defer channel1.Close()

	channel2, err := NewFileChannel("channel2", tmpFile2, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel 2: %v", err)
	}
	defer channel2.Close()

	stack := NewStackChannel([]Logger{channel1, channel2})

	testMessage := "stack message"
	stack.Info(testMessage)

	// Check first file
	content1, err := os.ReadFile(tmpFile1)
	if err != nil {
		t.Fatalf("Failed to read log file 1: %v", err)
	}

	if !strings.Contains(string(content1), testMessage) {
		t.Error("Stack did not write to first channel")
	}

	// Check second file
	content2, err := os.ReadFile(tmpFile2)
	if err != nil {
		t.Fatalf("Failed to read log file 2: %v", err)
	}

	if !strings.Contains(string(content2), testMessage) {
		t.Error("Stack did not write to second channel")
	}
}

func TestNullChannelDiscards(t *testing.T) {
	channel := NewNullChannel("null")

	// Should not panic or produce any output
	channel.Debug("debug")
	channel.Info("info")
	channel.Warning("warning")
	channel.Error("error")
	channel.Emergency("emergency")
}

func TestContextMerging(t *testing.T) {
	tmpFile := "/tmp/test-context.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("test", tmpFile, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	manager := NewManager(nil)
	manager.RegisterChannel("default", channel)

	contextLogger := manager.WithContext(map[string]any{
		"user_id": 123,
		"session": "abc",
	})

	contextLogger.Info("user action", map[string]any{
		"action": "login",
	})

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	if !strings.Contains(logContent, "user_id") {
		t.Error("Context does not contain user_id")
	}

	if !strings.Contains(logContent, "123") {
		t.Error("Context does not contain user_id value")
	}

	if !strings.Contains(logContent, "session") {
		t.Error("Context does not contain session")
	}

	if !strings.Contains(logContent, "action") {
		t.Error("Context does not contain action")
	}

	if !strings.Contains(logContent, "login") {
		t.Error("Context does not contain action value")
	}
}

func TestJSONFormatterOutput(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     ERROR,
		Message:   "test error",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Channel:   "test",
		Context: map[string]any{
			"key": "value",
		},
	}

	formatted := formatter.Format(entry)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(formatted), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON output: %v", err)
	}

	if decoded["level"] != "ERROR" {
		t.Errorf("JSON level = %q, want %q", decoded["level"], "ERROR")
	}

	if decoded["message"] != "test error" {
		t.Errorf("JSON message = %q, want %q", decoded["message"], "test error")
	}

	if decoded["channel"] != "test" {
		t.Errorf("JSON channel = %q, want %q", decoded["channel"], "test")
	}

	context, ok := decoded["context"].(map[string]any)
	if !ok {
		t.Fatal("JSON context is not a map")
	}

	if context["key"] != "value" {
		t.Errorf("JSON context key = %q, want %q", context["key"], "value")
	}
}

func TestLineFormatterOutput(t *testing.T) {
	formatter := NewLineFormatter()

	entry := &Entry{
		Level:     WARNING,
		Message:   "test warning",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Channel:   "test",
		Context: map[string]any{
			"key": "value",
		},
	}

	formatted := formatter.Format(entry)

	if !strings.Contains(formatted, "[2024-01-01 12:00:00]") {
		t.Error("Line format does not contain expected timestamp")
	}

	if !strings.Contains(formatted, "test.WARNING") {
		t.Error("Line format does not contain expected channel.level")
	}

	if !strings.Contains(formatted, "test warning") {
		t.Error("Line format does not contain expected message")
	}

	if !strings.Contains(formatted, `"key":"value"`) {
		t.Error("Line format does not contain expected context")
	}
}

func TestManagerChannelAccess(t *testing.T) {
	manager := NewManager(nil)

	tmpFile := "/tmp/test-manager.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("custom", tmpFile, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	manager.RegisterChannel("custom", channel)

	logger := manager.Channel("custom")
	logger.Info("test message")

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Error("Manager channel did not write to file")
	}
}

func TestManagerStack(t *testing.T) {
	manager := NewManager(nil)

	tmpFile1 := "/tmp/test-manager-stack-1.log"
	tmpFile2 := "/tmp/test-manager-stack-2.log"
	defer os.Remove(tmpFile1)
	defer os.Remove(tmpFile2)

	channel1, err := NewFileChannel("ch1", tmpFile1, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel 1: %v", err)
	}
	defer channel1.Close()

	channel2, err := NewFileChannel("ch2", tmpFile2, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel 2: %v", err)
	}
	defer channel2.Close()

	manager.RegisterChannel("ch1", channel1)
	manager.RegisterChannel("ch2", channel2)

	stackLogger := manager.Stack([]string{"ch1", "ch2"})
	stackLogger.Info("stack test")

	// Verify both files contain the message
	content1, _ := os.ReadFile(tmpFile1)
	content2, _ := os.ReadFile(tmpFile2)

	if !strings.Contains(string(content1), "stack test") {
		t.Error("Stack logger did not write to channel 1")
	}

	if !strings.Contains(string(content2), "stack test") {
		t.Error("Stack logger did not write to channel 2")
	}
}

func TestAllLogLevelsWrite(t *testing.T) {
	tmpFile := "/tmp/test-all-levels.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("test", tmpFile, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	channel.Emergency("emergency msg")
	channel.Alert("alert msg")
	channel.Critical("critical msg")
	channel.Error("error msg")
	channel.Warning("warning msg")
	channel.Notice("notice msg")
	channel.Info("info msg")
	channel.Debug("debug msg")

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)

	expectedMessages := []string{
		"emergency msg",
		"alert msg",
		"critical msg",
		"error msg",
		"warning msg",
		"notice msg",
		"info msg",
		"debug msg",
	}

	for _, msg := range expectedMessages {
		if !strings.Contains(logContent, msg) {
			t.Errorf("Log file does not contain %q", msg)
		}
	}
}

func TestManagerDefaultChannel(t *testing.T) {
	manager := NewManager(nil)

	tmpFile := "/tmp/test-manager-default.log"
	defer os.Remove(tmpFile)

	channel, err := NewFileChannel("default", tmpFile, DEBUG, NewLineFormatter())
	if err != nil {
		t.Fatalf("Failed to create file channel: %v", err)
	}
	defer channel.Close()

	manager.RegisterChannel("default", channel)

	// Use manager directly (should use default channel)
	manager.Info("default channel message")

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "default channel message") {
		t.Error("Manager default channel did not write message")
	}
}

func TestWriterThreadSafety(t *testing.T) {
	tmpFile := "/tmp/test-writer-threadsafe.log"
	defer os.Remove(tmpFile)

	writer, err := NewWriter(tmpFile)
	if err != nil {
		t.Fatalf("Failed to create writer: %v", err)
	}
	defer writer.Close()

	// Write from multiple goroutines
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 10; j++ {
				writer.Write([]byte("test\n"))
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	content, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Should have 100 lines (10 goroutines * 10 writes)
	lines := strings.Count(string(content), "test\n")
	if lines != 100 {
		t.Errorf("Expected 100 lines, got %d", lines)
	}
}

func TestJSONFormatterWithoutContext(t *testing.T) {
	formatter := NewJSONFormatter()

	entry := &Entry{
		Level:     INFO,
		Message:   "test",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Channel:   "test",
		Context:   map[string]any{},
	}

	formatted := formatter.Format(entry)

	var decoded map[string]any
	if err := json.Unmarshal([]byte(formatted), &decoded); err != nil {
		t.Fatalf("Failed to decode JSON output: %v", err)
	}

	// Context should not be present when empty
	if _, exists := decoded["context"]; exists {
		t.Error("JSON should not include empty context")
	}
}

func TestLineFormatterWithoutContext(t *testing.T) {
	formatter := NewLineFormatter()

	entry := &Entry{
		Level:     INFO,
		Message:   "test",
		Timestamp: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
		Channel:   "test",
		Context:   map[string]any{},
	}

	formatted := formatter.Format(entry)

	// Should not contain empty JSON object
	if strings.Contains(formatted, "{}") {
		t.Error("Line format should not include empty context object")
	}

	// Should still have proper format
	if !strings.Contains(formatted, "[2024-01-01 12:00:00] test.INFO: test") {
		t.Error("Line format is incorrect")
	}
}
