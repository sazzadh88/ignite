package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var (
	// ErrPasswordResetFailed is returned when password reset fails.
	ErrPasswordResetFailed = errors.New("password reset failed")
	// ErrInvalidResetToken is returned when reset token is invalid.
	ErrInvalidResetToken = errors.New("invalid reset token")
)

// PasswordResetToken represents a password reset token.
type PasswordResetToken struct {
	Email     string
	Token     string
	CreatedAt time.Time
}

// PasswordResetStorage defines the interface for storing reset tokens.
type PasswordResetStorage interface {
	// Create stores a new reset token.
	Create(email, token string) error
	// Get retrieves a reset token for an email.
	Get(email string) (*PasswordResetToken, error)
	// Delete removes a reset token.
	Delete(email string) error
	// DeleteExpired removes expired tokens.
	DeleteExpired(before time.Time) error
}

// PasswordResetter defines the interface for resetting passwords.
type PasswordResetter interface {
	// ResetPassword resets a user's password.
	ResetPassword(user Authenticatable, password string) error
}

// PasswordBroker handles password reset operations.
type PasswordBroker struct {
	provider  UserProvider
	storage   PasswordResetStorage
	resetter  PasswordResetter
	tokenLife time.Duration
}

// NewPasswordBroker creates a new password broker.
func NewPasswordBroker(
	provider UserProvider,
	storage PasswordResetStorage,
	resetter PasswordResetter,
) *PasswordBroker {
	return &PasswordBroker{
		provider:  provider,
		storage:   storage,
		resetter:  resetter,
		tokenLife: 60 * time.Minute, // Default: 60 minutes
	}
}

// SetTokenLife sets the token lifetime.
func (b *PasswordBroker) SetTokenLife(d time.Duration) {
	b.tokenLife = d
}

// SendResetLink sends a password reset link to the user.
// In a real implementation, this would send an email.
func (b *PasswordBroker) SendResetLink(email string) error {
	// Retrieve user by email
	user, err := b.provider.RetrieveByCredentials(map[string]string{"email": email})
	if err != nil {
		// Don't reveal if user exists
		return nil
	}

	// Create reset token
	token := b.CreateToken(user)

	// Store token
	if err := b.storage.Create(email, token); err != nil {
		return err
	}

	// In production, send email here
	// emailService.Send(email, token)

	return nil
}

// Reset resets a user's password using a token.
func (b *PasswordBroker) Reset(token, email, password string) error {
	// Clean up expired tokens
	b.storage.DeleteExpired(time.Now().Add(-b.tokenLife))

	// Get stored token
	storedToken, err := b.storage.Get(email)
	if err != nil {
		return ErrInvalidResetToken
	}

	// Check if token matches and is not expired
	if storedToken.Token != token {
		return ErrInvalidResetToken
	}

	if time.Since(storedToken.CreatedAt) > b.tokenLife {
		b.storage.Delete(email)
		return ErrInvalidResetToken
	}

	// Retrieve user
	user, err := b.provider.RetrieveByCredentials(map[string]string{"email": email})
	if err != nil {
		return ErrPasswordResetFailed
	}

	// Reset password
	if err := b.resetter.ResetPassword(user, password); err != nil {
		return err
	}

	// Delete used token
	b.storage.Delete(email)

	return nil
}

// CreateToken creates a reset token for a user.
func (b *PasswordBroker) CreateToken(user Authenticatable) string {
	// Generate random token
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// ValidateToken validates a reset token for a user.
// This is a simplified version that should be extended in production.
func (b *PasswordBroker) ValidateToken(user Authenticatable, token string) bool {
	if token == "" {
		return false
	}
	// In production, extract email from user and validate against storage
	// For now, just check token exists
	return true
}

// MemoryPasswordResetStorage is an in-memory storage for reset tokens.
type MemoryPasswordResetStorage struct {
	tokens map[string]*PasswordResetToken
	mu     sync.RWMutex
}

// NewMemoryPasswordResetStorage creates a new in-memory storage.
func NewMemoryPasswordResetStorage() *MemoryPasswordResetStorage {
	return &MemoryPasswordResetStorage{
		tokens: make(map[string]*PasswordResetToken),
	}
}

// Create stores a new reset token.
func (s *MemoryPasswordResetStorage) Create(email, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[email] = &PasswordResetToken{
		Email:     email,
		Token:     token,
		CreatedAt: time.Now(),
	}
	return nil
}

// Get retrieves a reset token for an email.
func (s *MemoryPasswordResetStorage) Get(email string) (*PasswordResetToken, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, exists := s.tokens[email]
	if !exists {
		return nil, ErrInvalidResetToken
	}
	return token, nil
}

// Delete removes a reset token.
func (s *MemoryPasswordResetStorage) Delete(email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tokens, email)
	return nil
}

// DeleteExpired removes expired tokens.
func (s *MemoryPasswordResetStorage) DeleteExpired(before time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for email, token := range s.tokens {
		if token.CreatedAt.Before(before) {
			delete(s.tokens, email)
		}
	}
	return nil
}
