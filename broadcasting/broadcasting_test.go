package broadcasting

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Test channel types
func TestChannelTypes(t *testing.T) {
	tests := []struct {
		name     string
		channel  Channel
		expected ChannelType
	}{
		{
			name:     "public channel",
			channel:  PublicChannel("news"),
			expected: Public,
		},
		{
			name:     "private channel",
			channel:  PrivateChannel("chat.1"),
			expected: Private,
		},
		{
			name:     "presence channel",
			channel:  PresenceChannel("room.1"),
			expected: Presence,
		},
		{
			name:     "encrypted private channel",
			channel:  EncryptedPrivateChannel("secure.1"),
			expected: EncryptedPrivate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.channel.Type != tt.expected {
				t.Errorf("expected type %v, got %v", tt.expected, tt.channel.Type)
			}
			if tt.channel.Name == "" {
				t.Error("channel name should not be empty")
			}
		})
	}
}

// Test null broadcaster
func TestNullBroadcaster(t *testing.T) {
	broadcaster := &NullBroadcaster{}

	err := broadcaster.Broadcast(
		[]Channel{PublicChannel("test")},
		"TestEvent",
		map[string]any{"key": "value"},
	)

	if err != nil {
		t.Errorf("null broadcaster should not return error: %v", err)
	}

	result, err := broadcaster.Auth(nil, PrivateChannel("test"))
	if err != nil {
		t.Errorf("null broadcaster auth should not return error: %v", err)
	}
	if result != nil {
		t.Error("null broadcaster auth should return nil")
	}
}

// Test log broadcaster
func TestLogBroadcaster(t *testing.T) {
	var buf bytes.Buffer
	broadcaster := NewLogBroadcaster(&buf)

	channels := []Channel{PublicChannel("test"), PrivateChannel("private")}
	event := "TestEvent"
	data := map[string]any{"message": "hello"}

	err := broadcaster.Broadcast(channels, event, data)
	if err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "[Broadcasting]") {
		t.Error("output should contain [Broadcasting] prefix")
	}
	if !strings.Contains(output, event) {
		t.Errorf("output should contain event name: %s", event)
	}
	if !strings.Contains(output, "test") {
		t.Error("output should contain channel names")
	}

	// Test Auth logging
	buf.Reset()
	_, err = broadcaster.Auth(nil, PrivateChannel("test"))
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	output = buf.String()
	if !strings.Contains(output, "auth") {
		t.Error("auth output should contain 'auth'")
	}
}

// Test manager driver selection
func TestManagerDriverSelection(t *testing.T) {
	manager := NewManager()

	// Test default driver (null)
	driver := manager.Driver("")
	if _, ok := driver.(*NullBroadcaster); !ok {
		t.Error("default driver should be null broadcaster")
	}

	// Register custom driver
	fake := Fake()
	manager.Extend("fake", fake)
	manager.SetDefaultDriver("fake")

	driver = manager.Driver("")
	if driver != fake {
		t.Error("should return fake driver")
	}

	driver = manager.Driver("fake")
	if driver != fake {
		t.Error("should return fake driver by name")
	}

	// Non-existent driver should return null
	driver = manager.Driver("nonexistent")
	if _, ok := driver.(*NullBroadcaster); !ok {
		t.Error("non-existent driver should return null broadcaster")
	}
}

// Test manager broadcast methods
func TestManagerBroadcast(t *testing.T) {
	manager := NewManager()
	fake := Fake()
	manager.Extend("fake", fake)
	manager.SetDefaultDriver("fake")

	channels := []Channel{PublicChannel("test")}
	event := "TestEvent"
	data := map[string]any{"key": "value"}

	err := manager.Broadcast(channels, event, data)
	if err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	if !fake.AssertBroadcast(event) {
		t.Error("event should have been broadcast")
	}
}

// Test ShouldBroadcast interface
type TestEvent struct {
	Message string
}

func (e *TestEvent) BroadcastOn() []Channel {
	return []Channel{PublicChannel("test")}
}

func (e *TestEvent) BroadcastAs() string {
	return "TestEvent"
}

func (e *TestEvent) BroadcastWith() map[string]any {
	return map[string]any{"message": e.Message}
}

func TestShouldBroadcastInterface(t *testing.T) {
	manager := NewManager()
	fake := Fake()
	manager.Extend("fake", fake)
	manager.SetDefaultDriver("fake")

	event := &TestEvent{Message: "hello"}
	err := manager.Event(event)
	if err != nil {
		t.Fatalf("event broadcast failed: %v", err)
	}

	if !fake.AssertBroadcast("TestEvent") {
		t.Error("TestEvent should have been broadcast")
	}

	if !fake.AssertBroadcastOn(PublicChannel("test"), "TestEvent") {
		t.Error("TestEvent should have been broadcast on test channel")
	}
}

func TestManagerEventNil(t *testing.T) {
	manager := NewManager()
	err := manager.Event(nil)
	if err == nil {
		t.Error("broadcasting nil event should return error")
	}
}

// Test fake broadcaster assertions
func TestFakeBroadcasterAssertions(t *testing.T) {
	fake := Fake()

	// Test nothing broadcast
	if !fake.AssertNothingBroadcast() {
		t.Error("should assert nothing broadcast initially")
	}

	// Broadcast an event
	channels := []Channel{PublicChannel("test"), PrivateChannel("private")}
	err := fake.Broadcast(channels, "TestEvent", map[string]any{"key": "value"})
	if err != nil {
		t.Fatalf("broadcast failed: %v", err)
	}

	// Test assertions
	if !fake.AssertBroadcast("TestEvent") {
		t.Error("should assert TestEvent was broadcast")
	}

	if fake.AssertNotBroadcast("TestEvent") {
		t.Error("should not assert TestEvent was not broadcast")
	}

	if fake.AssertNothingBroadcast() {
		t.Error("should not assert nothing broadcast after broadcasting")
	}

	if !fake.AssertBroadcastOn(PublicChannel("test"), "TestEvent") {
		t.Error("should assert broadcast on public test channel")
	}

	if !fake.AssertBroadcastOn(PrivateChannel("private"), "TestEvent") {
		t.Error("should assert broadcast on private channel")
	}

	if fake.AssertBroadcastOn(PublicChannel("other"), "TestEvent") {
		t.Error("should not assert broadcast on non-existent channel")
	}

	// Test broadcast count
	if err := fake.AssertBroadcastCount(1); err != nil {
		t.Errorf("broadcast count assertion failed: %v", err)
	}

	if fake.BroadcastCount() != 1 {
		t.Errorf("expected 1 broadcast, got %d", fake.BroadcastCount())
	}

	// Test reset
	fake.Reset()
	if !fake.AssertNothingBroadcast() {
		t.Error("should assert nothing broadcast after reset")
	}
}

// Test channel authorization
func TestChannelAuthorization(t *testing.T) {
	// Clear any existing auth registrations for testing
	authMu.Lock()
	channelAuths = []channelAuth{}
	authMu.Unlock()

	// Register a simple authorization callback
	RegisterChannel("chat.{id}", func(user any, params map[string]string) any {
		if user == nil {
			return nil
		}
		chatID := params["id"]
		if chatID == "" {
			return nil
		}
		return map[string]any{"user": user, "chat_id": chatID}
	})

	// Test successful authorization
	user := map[string]any{"id": 1, "name": "John"}
	channel := PrivateChannel("chat.123")

	result, err := AuthorizeChannel(user, channel)
	if err != nil {
		t.Fatalf("authorization failed: %v", err)
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		t.Fatal("result should be a map")
	}

	if resultMap["chat_id"] != "123" {
		t.Errorf("expected chat_id 123, got %v", resultMap["chat_id"])
	}

	// Test authorization denied (nil user)
	_, err = AuthorizeChannel(nil, channel)
	if err == nil {
		t.Error("authorization should fail for nil user")
	}

	// Test no handler found
	_, err = AuthorizeChannel(user, PrivateChannel("unknown.channel"))
	if err == nil {
		t.Error("authorization should fail when no handler found")
	}
}

// Test auth handler
func TestAuthHandler(t *testing.T) {
	// Clear and register auth callback
	authMu.Lock()
	channelAuths = []channelAuth{}
	authMu.Unlock()

	RegisterChannel("private-chat.{id}", func(user any, params map[string]string) any {
		if user == nil {
			return nil
		}
		return map[string]any{"user_id": 1}
	})

	userExtractor := func(r *http.Request) any {
		return map[string]any{"id": 1}
	}

	handler := AuthHandler(userExtractor)

	// Test successful authorization
	req := httptest.NewRequest("POST", "/broadcasting/auth", nil)
	req.Form = map[string][]string{
		"channel_name": {"private-chat.1"},
	}
	w := httptest.NewRecorder()

	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var response map[string]any
	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if auth, ok := response["auth"].(bool); !ok || !auth {
		t.Error("response should contain auth:true")
	}

	// Test missing channel_name
	req = httptest.NewRequest("POST", "/broadcasting/auth", nil)
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	// Test unauthenticated user
	unauthExtractor := func(r *http.Request) any {
		return nil
	}
	handler = AuthHandler(unauthExtractor)

	req = httptest.NewRequest("POST", "/broadcasting/auth", nil)
	req.Form = map[string][]string{
		"channel_name": {"private-chat.1"},
	}
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Test public channel (should fail)
	handler = AuthHandler(userExtractor)
	req = httptest.NewRequest("POST", "/broadcasting/auth", nil)
	req.Form = map[string][]string{
		"channel_name": {"public-channel"},
	}
	w = httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for public channel, got %d", w.Code)
	}
}

// Test pattern to regex conversion
func TestPatternToRegex(t *testing.T) {
	tests := []struct {
		pattern string
		input   string
		matches bool
		params  map[string]string
	}{
		{
			pattern: "chat.{id}",
			input:   "chat.123",
			matches: true,
			params:  map[string]string{"id": "123"},
		},
		{
			pattern: "user.{userId}.notifications",
			input:   "user.456.notifications",
			matches: true,
			params:  map[string]string{"userId": "456"},
		},
		{
			pattern: "room.{roomId}",
			input:   "room.abc",
			matches: true,
			params:  map[string]string{"roomId": "abc"},
		},
		{
			pattern: "chat.{id}",
			input:   "chat.123.messages",
			matches: false,
			params:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			regex := patternToRegex(tt.pattern)
			matches := regex.FindStringSubmatch(tt.input)

			if tt.matches && matches == nil {
				t.Errorf("pattern %s should match %s", tt.pattern, tt.input)
				return
			}

			if !tt.matches && matches != nil {
				t.Errorf("pattern %s should not match %s", tt.pattern, tt.input)
				return
			}

			if tt.matches && tt.params != nil {
				params := extractParams(regex, matches)
				for key, expected := range tt.params {
					if params[key] != expected {
						t.Errorf("expected param %s=%s, got %s", key, expected, params[key])
					}
				}
			}
		})
	}
}

// Test parse channel
func TestParseChannel(t *testing.T) {
	tests := []struct {
		name         string
		channelName  string
		expectedType ChannelType
	}{
		{
			name:         "public channel",
			channelName:  "news",
			expectedType: Public,
		},
		{
			name:         "private channel",
			channelName:  "private-chat.1",
			expectedType: Private,
		},
		{
			name:         "presence channel",
			channelName:  "presence-room.1",
			expectedType: Presence,
		},
		{
			name:         "encrypted private channel",
			channelName:  "private-encrypted-secure.1",
			expectedType: EncryptedPrivate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			channel := parseChannel(tt.channelName)
			if channel.Type != tt.expectedType {
				t.Errorf("expected type %v, got %v", tt.expectedType, channel.Type)
			}
			if channel.Name != tt.channelName {
				t.Errorf("expected name %s, got %s", tt.channelName, channel.Name)
			}
		})
	}
}

// Test package-level Broadcast variable
func TestPackageLevelBroadcast(t *testing.T) {
	if Broadcast == nil {
		t.Fatal("package-level Broadcast should be initialized")
	}

	// Should have null broadcaster as default
	driver := Broadcast.Driver("")
	if _, ok := driver.(*NullBroadcaster); !ok {
		t.Error("default driver should be null broadcaster")
	}
}
