package auth

import (
	"errors"
)

var (
	// ErrUserNotFound is returned when a user cannot be found.
	ErrUserNotFound = errors.New("user not found")
)

// UserProvider defines the interface for retrieving users.
// Implementations can fetch users from databases, APIs, etc.
type UserProvider interface {
	// RetrieveByID retrieves a user by their unique identifier.
	RetrieveByID(id any) (Authenticatable, error)

	// RetrieveByCredentials retrieves a user by their credentials (e.g., email).
	RetrieveByCredentials(credentials map[string]string) (Authenticatable, error)

	// ValidateCredentials validates that the user's credentials match.
	ValidateCredentials(user Authenticatable, credentials map[string]string) bool
}

// RepositoryFunc is a function type for retrieving users by ID.
type RepositoryFunc func(id any) (Authenticatable, error)

// CredentialsFunc is a function type for retrieving users by credentials.
type CredentialsFunc func(credentials map[string]string) (Authenticatable, error)

// ValidatorFunc is a function type for validating credentials.
type ValidatorFunc func(user Authenticatable, credentials map[string]string) bool

// CallbackProvider is a simple UserProvider that uses callback functions.
// This is useful for custom implementations without defining a new type.
type CallbackProvider struct {
	retrieveByID          RepositoryFunc
	retrieveByCredentials CredentialsFunc
	validateCredentials   ValidatorFunc
}

// NewCallbackProvider creates a new callback-based user provider.
func NewCallbackProvider(
	retrieveByID RepositoryFunc,
	retrieveByCredentials CredentialsFunc,
	validateCredentials ValidatorFunc,
) *CallbackProvider {
	return &CallbackProvider{
		retrieveByID:          retrieveByID,
		retrieveByCredentials: retrieveByCredentials,
		validateCredentials:   validateCredentials,
	}
}

// RetrieveByID retrieves a user by their unique identifier.
func (p *CallbackProvider) RetrieveByID(id any) (Authenticatable, error) {
	if p.retrieveByID == nil {
		return nil, ErrUserNotFound
	}
	return p.retrieveByID(id)
}

// RetrieveByCredentials retrieves a user by their credentials.
func (p *CallbackProvider) RetrieveByCredentials(credentials map[string]string) (Authenticatable, error) {
	if p.retrieveByCredentials == nil {
		return nil, ErrUserNotFound
	}
	return p.retrieveByCredentials(credentials)
}

// ValidateCredentials validates that the user's credentials match.
func (p *CallbackProvider) ValidateCredentials(user Authenticatable, credentials map[string]string) bool {
	if p.validateCredentials == nil {
		return false
	}
	return p.validateCredentials(user, credentials)
}
