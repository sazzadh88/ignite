package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// mockUser is a test implementation of Authenticatable.
type mockUser struct {
	id       int
	email    string
	password string
}

func (u *mockUser) GetAuthIdentifier() any {
	return u.id
}

func (u *mockUser) GetAuthPassword() string {
	return u.password
}

// mockSession is a test implementation of Session.
type mockSession struct {
	data map[string]any
}

func newMockSession() *mockSession {
	return &mockSession{
		data: make(map[string]any),
	}
}

func (s *mockSession) Get(key string) (any, bool) {
	val, exists := s.data[key]
	return val, exists
}

func (s *mockSession) Put(key, value any) {
	s.data[key.(string)] = value
}

func (s *mockSession) Forget(key string) {
	delete(s.data, key)
}

func (s *mockSession) Flush() {
	s.data = make(map[string]any)
}

// mockProvider is a test implementation of UserProvider.
type mockProvider struct {
	users map[int]*mockUser
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		users: map[int]*mockUser{
			1: {id: 1, email: "test@example.com", password: "hashed_password"},
			2: {id: 2, email: "admin@example.com", password: "hashed_admin_pass"},
		},
	}
}

func (p *mockProvider) RetrieveByID(id any) (Authenticatable, error) {
	userID, ok := id.(int)
	if !ok {
		return nil, ErrUserNotFound
	}

	user, exists := p.users[userID]
	if !exists {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (p *mockProvider) RetrieveByCredentials(credentials map[string]string) (Authenticatable, error) {
	email, exists := credentials["email"]
	if !exists {
		return nil, ErrUserNotFound
	}

	for _, user := range p.users {
		if user.email == email {
			return user, nil
		}
	}
	return nil, ErrUserNotFound
}

func (p *mockProvider) ValidateCredentials(user Authenticatable, credentials map[string]string) bool {
	password, exists := credentials["password"]
	if !exists {
		return false
	}

	return user.GetAuthPassword() == password
}

// mockPasswordResetter is a test implementation of PasswordResetter.
type mockPasswordResetter struct {
	resetCalled bool
	lastUser    Authenticatable
	lastPass    string
}

func (r *mockPasswordResetter) ResetPassword(user Authenticatable, password string) error {
	r.resetCalled = true
	r.lastUser = user
	r.lastPass = password
	return nil
}

func TestManager(t *testing.T) {
	manager := NewManager()

	// Test default guard name
	if manager.defaultGuard != "session" {
		t.Errorf("expected default guard to be 'session', got %s", manager.defaultGuard)
	}

	// Test adding guard
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)
	manager.AddGuard("test", guard)

	if !manager.HasGuard("test") {
		t.Error("expected guard 'test' to exist")
	}

	// Test retrieving guard
	retrievedGuard := manager.Guard("test")
	if retrievedGuard == nil {
		t.Error("expected to retrieve guard 'test'")
	}

	// Test setting default guard
	manager.SetDefaultGuard("test")
	if manager.defaultGuard != "test" {
		t.Errorf("expected default guard to be 'test', got %s", manager.defaultGuard)
	}
}

func TestSessionGuard_GuestCheck(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	if !guard.Guest() {
		t.Error("expected guard to be guest when no user")
	}

	if guard.Check() {
		t.Error("expected guard check to be false when no user")
	}

	if guard.User() != nil {
		t.Error("expected user to be nil")
	}

	if guard.ID() != nil {
		t.Error("expected ID to be nil")
	}
}

func TestSessionGuard_Login(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	user := &mockUser{id: 1, email: "test@example.com"}
	guard.Login(user)

	if guard.Guest() {
		t.Error("expected guard to not be guest after login")
	}

	if !guard.Check() {
		t.Error("expected guard check to be true after login")
	}

	if guard.User() != user {
		t.Error("expected user to match logged in user")
	}

	if guard.ID() != 1 {
		t.Errorf("expected ID to be 1, got %v", guard.ID())
	}

	// Check session
	sessionID, exists := session.Get("auth_user_id")
	if !exists || sessionID != 1 {
		t.Error("expected user ID to be stored in session")
	}
}

func TestSessionGuard_LoginUsingID(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	err := guard.LoginUsingID(1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	if guard.ID() != 1 {
		t.Errorf("expected ID to be 1, got %v", guard.ID())
	}

	// Test invalid ID
	err = guard.LoginUsingID(999)
	if err == nil {
		t.Error("expected error for invalid user ID")
	}
}

func TestSessionGuard_Logout(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	user := &mockUser{id: 1, email: "test@example.com"}
	guard.Login(user)
	guard.Logout()

	if !guard.Guest() {
		t.Error("expected guard to be guest after logout")
	}

	if guard.User() != nil {
		t.Error("expected user to be nil after logout")
	}

	// Check session
	_, exists := session.Get("auth_user_id")
	if exists {
		t.Error("expected user ID to be removed from session")
	}
}

func TestSessionGuard_Attempt(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	// Test successful attempt
	success := guard.Attempt(map[string]string{
		"email":    "test@example.com",
		"password": "hashed_password",
	})

	if !success {
		t.Error("expected attempt to succeed with valid credentials")
	}

	if guard.Guest() {
		t.Error("expected user to be logged in after successful attempt")
	}

	// Logout for next test
	guard.Logout()

	// Test failed attempt
	success = guard.Attempt(map[string]string{
		"email":    "test@example.com",
		"password": "wrong_password",
	})

	if success {
		t.Error("expected attempt to fail with invalid credentials")
	}

	if !guard.Guest() {
		t.Error("expected user to remain guest after failed attempt")
	}
}

func TestSessionGuard_Validate(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	// Test valid credentials
	valid := guard.Validate(map[string]string{
		"email":    "test@example.com",
		"password": "hashed_password",
	})

	if !valid {
		t.Error("expected credentials to be valid")
	}

	// Should not log in user
	if !guard.Guest() {
		t.Error("expected user to remain guest after validate")
	}

	// Test invalid credentials
	valid = guard.Validate(map[string]string{
		"email":    "test@example.com",
		"password": "wrong_password",
	})

	if valid {
		t.Error("expected credentials to be invalid")
	}
}

func TestSessionGuard_Once(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	// Test successful once
	success := guard.Once(map[string]string{
		"email":    "test@example.com",
		"password": "hashed_password",
	})

	if !success {
		t.Error("expected once to succeed with valid credentials")
	}

	if guard.Guest() {
		t.Error("expected user to be set after successful once")
	}

	// Check session - should not be persisted
	_, exists := session.Get("auth_user_id")
	if exists {
		t.Error("expected user ID to not be stored in session for once")
	}
}

func TestTokenGuard(t *testing.T) {
	provider := newMockProvider()
	storage := NewMemoryTokenStorage()
	guard := NewTokenGuard(provider, storage)

	user := &mockUser{id: 1, email: "test@example.com"}

	// Create token
	token, err := guard.CreateToken(user, "test-token", []string{"read", "write"})
	if err != nil {
		t.Errorf("expected no error creating token, got %v", err)
	}

	if token.PlainTextToken == "" {
		t.Error("expected plain text token to be generated")
	}

	if token.Name != "test-token" {
		t.Errorf("expected token name to be 'test-token', got %s", token.Name)
	}

	// Set token
	err = guard.SetToken(token.PlainTextToken)
	if err != nil {
		t.Errorf("expected no error setting token, got %v", err)
	}

	if guard.Guest() {
		t.Error("expected user to be authenticated with token")
	}

	if guard.ID() != 1 {
		t.Errorf("expected user ID to be 1, got %v", guard.ID())
	}

	// Test invalid token
	err = guard.SetToken("invalid-token")
	if err == nil {
		t.Error("expected error with invalid token")
	}
}

func TestPasswordBroker(t *testing.T) {
	provider := newMockProvider()
	storage := NewMemoryPasswordResetStorage()
	resetter := &mockPasswordResetter{}
	broker := NewPasswordBroker(provider, storage, resetter)

	// Test sending reset link
	err := broker.SendResetLink("test@example.com")
	if err != nil {
		t.Errorf("expected no error sending reset link, got %v", err)
	}

	// Check token was created
	token, err := storage.Get("test@example.com")
	if err != nil {
		t.Error("expected reset token to be created")
	}

	if token.Email != "test@example.com" {
		t.Errorf("expected email to be 'test@example.com', got %s", token.Email)
	}

	// Test reset
	err = broker.Reset(token.Token, "test@example.com", "new_password")
	if err != nil {
		t.Errorf("expected no error resetting password, got %v", err)
	}

	if !resetter.resetCalled {
		t.Error("expected reset password to be called")
	}

	// Token should be deleted after use
	_, err = storage.Get("test@example.com")
	if err == nil {
		t.Error("expected token to be deleted after use")
	}
}

func TestAuthMiddleware(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	middleware := NewAuthMiddleware(guard)

	// Test unauthenticated request - should redirect
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}

	middleware.Handle(w, req, next)

	if nextCalled {
		t.Error("expected next to not be called for unauthenticated request")
	}

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}

	// Test authenticated request
	user := &mockUser{id: 1, email: "test@example.com"}
	guard.Login(user)

	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.Handle(w, req, next)

	if !nextCalled {
		t.Error("expected next to be called for authenticated request")
	}

	// Test with abort on fail
	guard.Logout()
	middleware.SetAbortOnFail(true)

	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.Handle(w, req, next)

	if nextCalled {
		t.Error("expected next to not be called")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestGuestMiddleware(t *testing.T) {
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	middleware := NewGuestMiddleware(guard)

	// Test guest request - should pass
	req := httptest.NewRequest("GET", "/login", nil)
	w := httptest.NewRecorder()

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}

	middleware.Handle(w, req, next)

	if !nextCalled {
		t.Error("expected next to be called for guest request")
	}

	// Test authenticated request - should redirect
	user := &mockUser{id: 1, email: "test@example.com"}
	guard.Login(user)

	req = httptest.NewRequest("GET", "/login", nil)
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.Handle(w, req, next)

	if nextCalled {
		t.Error("expected next to not be called for authenticated request")
	}

	if w.Code != http.StatusFound {
		t.Errorf("expected status 302, got %d", w.Code)
	}
}

func TestTokenAuthMiddleware(t *testing.T) {
	provider := newMockProvider()
	storage := NewMemoryTokenStorage()
	guard := NewTokenGuard(provider, storage)

	middleware := NewTokenAuthMiddleware(guard)

	// Create a token
	user := &mockUser{id: 1, email: "test@example.com"}
	token, err := guard.CreateToken(user, "test-token", []string{"read"})
	if err != nil {
		t.Fatalf("failed to create token: %v", err)
	}

	// Test with valid token
	req := httptest.NewRequest("GET", "/api/resource", nil)
	req.Header.Set("Authorization", "Bearer "+token.PlainTextToken)
	w := httptest.NewRecorder()

	nextCalled := false
	next := func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	}

	middleware.Handle(w, req, next)

	if !nextCalled {
		t.Error("expected next to be called with valid token")
	}

	// Test without token
	req = httptest.NewRequest("GET", "/api/resource", nil)
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.Handle(w, req, next)

	if nextCalled {
		t.Error("expected next to not be called without token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}

	// Test with invalid token
	req = httptest.NewRequest("GET", "/api/resource", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	nextCalled = false

	middleware.Handle(w, req, next)

	if nextCalled {
		t.Error("expected next to not be called with invalid token")
	}

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

func TestFacadeFunctions(t *testing.T) {
	// Create new manager
	manager := NewManager()
	provider := newMockProvider()
	session := newMockSession()
	guard := NewSessionGuard(provider, session)

	manager.AddGuard("session", guard)
	manager.SetDefaultGuard("session")

	SetManager(manager)

	// Test Guest function
	if !Guest() {
		t.Error("expected Guest() to return true when no user")
	}

	// Test Check function
	if Check() {
		t.Error("expected Check() to return false when no user")
	}

	// Test User function
	if User() != nil {
		t.Error("expected User() to return nil when no user")
	}

	// Test ID function
	if ID() != nil {
		t.Error("expected ID() to return nil when no user")
	}

	// Test Attempt function
	success := Attempt(map[string]string{
		"email":    "test@example.com",
		"password": "hashed_password",
	})

	if !success {
		t.Error("expected Attempt() to succeed with valid credentials")
	}

	// Now should be authenticated
	if Guest() {
		t.Error("expected Guest() to return false after login")
	}

	if !Check() {
		t.Error("expected Check() to return true after login")
	}

	if User() == nil {
		t.Error("expected User() to return user after login")
	}

	if ID() != 1 {
		t.Errorf("expected ID() to return 1, got %v", ID())
	}

	// Test Logout function
	Logout()

	if !Guest() {
		t.Error("expected Guest() to return true after logout")
	}
}
