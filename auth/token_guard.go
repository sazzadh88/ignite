package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	// ErrInvalidToken is returned when a token is invalid.
	ErrInvalidToken = errors.New("invalid token")
)

// PersonalAccessToken represents an API token for a user.
type PersonalAccessToken struct {
	ID             any
	Name           string
	Token          string // Hashed token stored in database
	PlainTextToken string // Plain text token shown once to user
	Abilities      []string
	CreatedAt      time.Time
	LastUsedAt     *time.Time
}

// TokenStorage defines the interface for storing and retrieving tokens.
type TokenStorage interface {
	// FindToken retrieves a token by its hashed value.
	FindToken(hashedToken string) (*PersonalAccessToken, error)
	// StoreToken stores a new token.
	StoreToken(token *PersonalAccessToken) error
	// RevokeToken revokes a token.
	RevokeToken(tokenID any) error
	// UpdateLastUsed updates the last used timestamp.
	UpdateLastUsed(tokenID any, t time.Time) error
}

// TokenGuard implements token-based authentication for APIs.
type TokenGuard struct {
	provider UserProvider
	storage  TokenStorage
	user     Authenticatable
	token    string
	mu       sync.RWMutex
}

// NewTokenGuard creates a new token guard.
func NewTokenGuard(provider UserProvider, storage TokenStorage) *TokenGuard {
	return &TokenGuard{
		provider: provider,
		storage:  storage,
	}
}

// User returns the currently authenticated user.
func (g *TokenGuard) User() Authenticatable {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.user
}

// ID returns the identifier for the currently authenticated user.
func (g *TokenGuard) ID() any {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.user == nil {
		return nil
	}
	return g.user.GetAuthIdentifier()
}

// Check returns true if a user is authenticated.
func (g *TokenGuard) Check() bool {
	return g.User() != nil
}

// Guest returns true if no user is authenticated.
func (g *TokenGuard) Guest() bool {
	return !g.Check()
}

// Validate checks if the given credentials are valid without logging in.
func (g *TokenGuard) Validate(credentials map[string]string) bool {
	// Token guard doesn't use credentials validation
	return false
}

// SetToken sets the bearer token for authentication.
func (g *TokenGuard) SetToken(token string) error {
	if token == "" {
		return ErrInvalidToken
	}

	// Hash the token to find it in storage
	hashedToken := hashToken(token)
	tokenData, err := g.storage.FindToken(hashedToken)
	if err != nil {
		return err
	}

	// Retrieve the user
	user, err := g.provider.RetrieveByID(tokenData.ID)
	if err != nil {
		return err
	}

	// Update last used timestamp
	g.storage.UpdateLastUsed(tokenData.ID, time.Now())

	g.mu.Lock()
	g.user = user
	g.token = token
	g.mu.Unlock()

	return nil
}

// CreateToken creates a new personal access token for the user.
func (g *TokenGuard) CreateToken(user Authenticatable, name string, abilities []string) (*PersonalAccessToken, error) {
	// Generate a random token
	plainTextToken, err := generateToken(40)
	if err != nil {
		return nil, err
	}

	// Hash the token for storage
	hashedToken := hashToken(plainTextToken)

	token := &PersonalAccessToken{
		ID:             user.GetAuthIdentifier(),
		Name:           name,
		Token:          hashedToken,
		PlainTextToken: plainTextToken,
		Abilities:      abilities,
		CreatedAt:      time.Now(),
	}

	if err := g.storage.StoreToken(token); err != nil {
		return nil, err
	}

	return token, nil
}

// RevokeToken revokes a token by its ID.
func (g *TokenGuard) RevokeToken(tokenID any) error {
	return g.storage.RevokeToken(tokenID)
}

// CurrentToken returns the current token string.
func (g *TokenGuard) CurrentToken() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.token
}

// generateToken generates a random token of the specified length.
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashToken hashes a token for storage (simple implementation).
// In production, use a proper hashing algorithm like SHA-256.
func hashToken(token string) string {
	// Simple hash for demonstration - in production use crypto/sha256
	return hex.EncodeToString([]byte(token))
}

// MemoryTokenStorage is a simple in-memory token storage for testing.
type MemoryTokenStorage struct {
	tokens map[string]*PersonalAccessToken
	mu     sync.RWMutex
}

// NewMemoryTokenStorage creates a new in-memory token storage.
func NewMemoryTokenStorage() *MemoryTokenStorage {
	return &MemoryTokenStorage{
		tokens: make(map[string]*PersonalAccessToken),
	}
}

// FindToken retrieves a token by its hashed value.
func (s *MemoryTokenStorage) FindToken(hashedToken string) (*PersonalAccessToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[hashedToken]
	if !exists {
		return nil, ErrInvalidToken
	}
	return token, nil
}

// StoreToken stores a new token.
func (s *MemoryTokenStorage) StoreToken(token *PersonalAccessToken) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tokens[token.Token] = token
	return nil
}

// RevokeToken revokes a token.
func (s *MemoryTokenStorage) RevokeToken(tokenID any) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, token := range s.tokens {
		if token.ID == tokenID {
			delete(s.tokens, key)
			return nil
		}
	}
	return ErrInvalidToken
}

// UpdateLastUsed updates the last used timestamp.
func (s *MemoryTokenStorage) UpdateLastUsed(tokenID any, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, token := range s.tokens {
		if token.ID == tokenID {
			token.LastUsedAt = &t
			return nil
		}
	}
	return ErrInvalidToken
}
