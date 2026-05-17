package orm

import (
	"time"
)

// SoftDeletes is a marker interface for models that support soft deletion.
// Models implementing this interface will have deleted_at tracked automatically.
type SoftDeletes interface {
	// GetDeletedAt returns the deleted_at timestamp.
	GetDeletedAt() *time.Time

	// SetDeletedAt sets the deleted_at timestamp.
	SetDeletedAt(t *time.Time)
}

// SoftDeleteModel can be embedded in models to add soft delete support.
type SoftDeleteModel struct {
	DeletedAt *time.Time `db:"deleted_at"`
}

// GetDeletedAt returns the deleted_at timestamp.
func (m *SoftDeleteModel) GetDeletedAt() *time.Time {
	return m.DeletedAt
}

// SetDeletedAt sets the deleted_at timestamp.
func (m *SoftDeleteModel) SetDeletedAt(t *time.Time) {
	m.DeletedAt = t
}

// Trashed checks if the model is soft deleted.
func (m *SoftDeleteModel) Trashed() bool {
	return m.DeletedAt != nil
}

// SoftDeleteScope is a global scope that excludes soft-deleted records.
func SoftDeleteScope[T any]() ScopeFunc[T] {
	return func(q *Query[T]) *Query[T] {
		if !q.withTrashed && !q.onlyTrashed {
			return q.WhereNull("deleted_at")
		}
		if q.onlyTrashed {
			return q.WhereNotNull("deleted_at")
		}
		return q
	}
}

// ApplySoftDeleteScope applies soft delete filtering to a query.
func ApplySoftDeleteScope[T any](q *Query[T]) *Query[T] {
	// Check if the model implements SoftDeletes
	var model T
	if _, ok := any(model).(SoftDeletes); ok {
		return SoftDeleteScope[T]()(q)
	}
	return q
}

// RestoreModel restores a soft-deleted model instance.
func RestoreModel(m *ModelInstance) error {
	if err := FireRestoring(m.model); err != nil {
		return err
	}

	m.Set("deleted_at", nil)
	m.Set("updated_at", time.Now())

	if err := FireRestored(m.model); err != nil {
		return err
	}

	return nil
}

// TrashedModels returns whether a model instance is trashed.
func TrashedModels(m *ModelInstance) bool {
	if deletedAt, ok := m.Get("deleted_at"); ok {
		return deletedAt != nil
	}
	return false
}

// SoftDeleteHelper provides helper methods for soft delete operations.
type SoftDeleteHelper struct{}

// IsSoftDeletable checks if a model implements SoftDeletes.
func (h SoftDeleteHelper) IsSoftDeletable(model any) bool {
	_, ok := model.(SoftDeletes)
	return ok
}

// PerformSoftDelete performs a soft delete on a model.
func (h SoftDeleteHelper) PerformSoftDelete(model SoftDeletes) error {
	now := time.Now()
	model.SetDeletedAt(&now)
	return nil
}

// PerformRestore restores a soft-deleted model.
func (h SoftDeleteHelper) PerformRestore(model SoftDeletes) error {
	model.SetDeletedAt(nil)
	return nil
}

// PerformForceDelete performs a hard delete on a model.
func (h SoftDeleteHelper) PerformForceDelete(model any) error {
	// This would execute the actual DELETE query
	return nil
}

// NewSoftDeleteHelper creates a new soft delete helper.
func NewSoftDeleteHelper() *SoftDeleteHelper {
	return &SoftDeleteHelper{}
}
