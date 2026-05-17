package orm

import (
	"encoding/json"
	"time"
)

// Model represents the base model with common fields for all database models.
// It provides timestamps and soft delete support.
type Model struct {
	ID        uint64     `db:"id"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

// Modeler defines the interface that all models must implement.
// It provides metadata about the model's table structure and behavior.
type Modeler interface {
	// Table returns the database table name for this model.
	Table() string

	// Fillable returns the list of mass-assignable fields.
	Fillable() []string

	// Hidden returns the list of fields that should be hidden from JSON output.
	Hidden() []string

	// PrimaryKey returns the name of the primary key column.
	PrimaryKey() string
}

// ModelInstance wraps a model with runtime state tracking.
type ModelInstance struct {
	model      any
	original   map[string]any
	attributes map[string]any
	dirty      map[string]any
	exists     bool
}

// NewModelInstance creates a new model instance with dirty tracking.
func NewModelInstance(model any, attributes map[string]any) *ModelInstance {
	m := &ModelInstance{
		model:      model,
		attributes: make(map[string]any),
		original:   make(map[string]any),
		dirty:      make(map[string]any),
		exists:     false,
	}

	for k, v := range attributes {
		m.attributes[k] = v
		m.original[k] = v
	}

	return m
}

// Set sets an attribute value and marks it as dirty.
func (m *ModelInstance) Set(key string, value any) {
	m.attributes[key] = value

	// Mark as dirty if value differs from original
	if origVal, exists := m.original[key]; !exists || origVal != value {
		m.dirty[key] = value
	}
}

// Get retrieves an attribute value.
func (m *ModelInstance) Get(key string) (any, bool) {
	val, ok := m.attributes[key]
	return val, ok
}

// IsDirty checks if any of the specified fields have been modified.
// If no fields are specified, returns true if any field is dirty.
func (m *ModelInstance) IsDirty(fields ...string) bool {
	if len(fields) == 0 {
		return len(m.dirty) > 0
	}

	for _, field := range fields {
		if _, exists := m.dirty[field]; exists {
			return true
		}
	}

	return false
}

// GetDirty returns a map of all modified fields and their new values.
func (m *ModelInstance) GetDirty() map[string]any {
	result := make(map[string]any)
	for k, v := range m.dirty {
		result[k] = v
	}
	return result
}

// SyncOriginal synchronizes the original state with current attributes.
// This is typically called after saving to the database.
func (m *ModelInstance) SyncOriginal() {
	m.original = make(map[string]any)
	for k, v := range m.attributes {
		m.original[k] = v
	}
	m.dirty = make(map[string]any)
}

// ToMap converts the model instance to a map, respecting hidden fields.
func (m *ModelInstance) ToMap(modeler Modeler) map[string]any {
	hidden := make(map[string]bool)
	for _, field := range modeler.Hidden() {
		hidden[field] = true
	}

	result := make(map[string]any)
	for k, v := range m.attributes {
		if !hidden[k] {
			result[k] = v
		}
	}

	return result
}

// ToJSON converts the model instance to JSON, respecting hidden fields.
func (m *ModelInstance) ToJSON(modeler Modeler) ([]byte, error) {
	return json.Marshal(m.ToMap(modeler))
}

// Save persists the model to the database.
// It will perform an insert if the model is new, or an update if it exists.
func (m *ModelInstance) Save(modeler Modeler) error {
	now := time.Now()

	if !m.exists {
		// Insert
		m.Set("created_at", now)
		m.Set("updated_at", now)
		// TODO: Execute insert query
		m.exists = true
		m.SyncOriginal()
		return nil
	}

	// Update
	if !m.IsDirty() {
		return nil // Nothing to update
	}

	m.Set("updated_at", now)
	// TODO: Execute update query
	m.SyncOriginal()
	return nil
}

// Delete soft deletes the model by setting deleted_at timestamp.
func (m *ModelInstance) Delete() error {
	if _, ok := m.model.(SoftDeletes); ok {
		now := time.Now()
		m.Set("deleted_at", &now)
		m.Set("updated_at", now)
		// TODO: Execute update query
		m.SyncOriginal()
		return nil
	}

	// If not soft deletable, perform hard delete
	return m.ForceDelete()
}

// ForceDelete permanently removes the model from the database.
func (m *ModelInstance) ForceDelete() error {
	// TODO: Execute delete query
	m.exists = false
	return nil
}

// Restore restores a soft-deleted model.
func (m *ModelInstance) Restore() error {
	if _, ok := m.model.(SoftDeletes); ok {
		m.Set("deleted_at", nil)
		m.Set("updated_at", time.Now())
		// TODO: Execute update query
		m.SyncOriginal()
		return nil
	}

	return nil
}

// Touch updates the updated_at timestamp without modifying other fields.
func (m *ModelInstance) Touch() error {
	m.Set("updated_at", time.Now())
	// TODO: Execute update query
	m.SyncOriginal()
	return nil
}

// Fresh retrieves a fresh copy of the model from the database.
func (m *ModelInstance) Fresh(modeler Modeler) error {
	// TODO: Execute select query
	return nil
}
