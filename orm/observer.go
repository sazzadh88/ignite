package orm

import (
	"reflect"
)

// Observer defines optional lifecycle hooks for models.
// Implement only the methods you need.
type Observer interface{}

// CreatingObserver is called before a model is created.
type CreatingObserver interface {
	Creating(model any) error
}

// CreatedObserver is called after a model is created.
type CreatedObserver interface {
	Created(model any) error
}

// UpdatingObserver is called before a model is updated.
type UpdatingObserver interface {
	Updating(model any) error
}

// UpdatedObserver is called after a model is updated.
type UpdatedObserver interface {
	Updated(model any) error
}

// SavingObserver is called before a model is saved (created or updated).
type SavingObserver interface {
	Saving(model any) error
}

// SavedObserver is called after a model is saved (created or updated).
type SavedObserver interface {
	Saved(model any) error
}

// DeletingObserver is called before a model is deleted.
type DeletingObserver interface {
	Deleting(model any) error
}

// DeletedObserver is called after a model is deleted.
type DeletedObserver interface {
	Deleted(model any) error
}

// RestoringObserver is called before a soft-deleted model is restored.
type RestoringObserver interface {
	Restoring(model any) error
}

// RestoredObserver is called after a soft-deleted model is restored.
type RestoredObserver interface {
	Restored(model any) error
}

// ForceDeletingObserver is called before a model is force deleted.
type ForceDeletingObserver interface {
	ForceDeleting(model any) error
}

// ForceDeletedObserver is called after a model is force deleted.
type ForceDeletedObserver interface {
	ForceDeleted(model any) error
}

// RetrievedObserver is called after a model is retrieved from the database.
type RetrievedObserver interface {
	Retrieved(model any) error
}

// ObserverRegistry manages observers for models.
type ObserverRegistry struct {
	observers map[string][]Observer
}

var globalObserverRegistry = NewObserverRegistry()

// NewObserverRegistry creates a new observer registry.
func NewObserverRegistry() *ObserverRegistry {
	return &ObserverRegistry{
		observers: make(map[string][]Observer),
	}
}

// Observe registers an observer for a model type.
func Observe(model any, observer Observer) {
	globalObserverRegistry.Register(model, observer)
}

// Register registers an observer for a model type.
func (r *ObserverRegistry) Register(model any, observer Observer) {
	modelType := reflect.TypeOf(model).String()
	r.observers[modelType] = append(r.observers[modelType], observer)
}

// GetObservers returns all observers for a model type.
func (r *ObserverRegistry) GetObservers(model any) []Observer {
	modelType := reflect.TypeOf(model).String()
	return r.observers[modelType]
}

// FireEvent fires an event for a model.
func (r *ObserverRegistry) FireEvent(event string, model any) error {
	observers := r.GetObservers(model)

	for _, observer := range observers {
		var err error

		switch event {
		case "creating":
			if o, ok := observer.(CreatingObserver); ok {
				err = o.Creating(model)
			}
		case "created":
			if o, ok := observer.(CreatedObserver); ok {
				err = o.Created(model)
			}
		case "updating":
			if o, ok := observer.(UpdatingObserver); ok {
				err = o.Updating(model)
			}
		case "updated":
			if o, ok := observer.(UpdatedObserver); ok {
				err = o.Updated(model)
			}
		case "saving":
			if o, ok := observer.(SavingObserver); ok {
				err = o.Saving(model)
			}
		case "saved":
			if o, ok := observer.(SavedObserver); ok {
				err = o.Saved(model)
			}
		case "deleting":
			if o, ok := observer.(DeletingObserver); ok {
				err = o.Deleting(model)
			}
		case "deleted":
			if o, ok := observer.(DeletedObserver); ok {
				err = o.Deleted(model)
			}
		case "restoring":
			if o, ok := observer.(RestoringObserver); ok {
				err = o.Restoring(model)
			}
		case "restored":
			if o, ok := observer.(RestoredObserver); ok {
				err = o.Restored(model)
			}
		case "force_deleting":
			if o, ok := observer.(ForceDeletingObserver); ok {
				err = o.ForceDeleting(model)
			}
		case "force_deleted":
			if o, ok := observer.(ForceDeletedObserver); ok {
				err = o.ForceDeleted(model)
			}
		case "retrieved":
			if o, ok := observer.(RetrievedObserver); ok {
				err = o.Retrieved(model)
			}
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// FireCreating fires the creating event.
func FireCreating(model any) error {
	return globalObserverRegistry.FireEvent("creating", model)
}

// FireCreated fires the created event.
func FireCreated(model any) error {
	return globalObserverRegistry.FireEvent("created", model)
}

// FireUpdating fires the updating event.
func FireUpdating(model any) error {
	return globalObserverRegistry.FireEvent("updating", model)
}

// FireUpdated fires the updated event.
func FireUpdated(model any) error {
	return globalObserverRegistry.FireEvent("updated", model)
}

// FireSaving fires the saving event.
func FireSaving(model any) error {
	return globalObserverRegistry.FireEvent("saving", model)
}

// FireSaved fires the saved event.
func FireSaved(model any) error {
	return globalObserverRegistry.FireEvent("saved", model)
}

// FireDeleting fires the deleting event.
func FireDeleting(model any) error {
	return globalObserverRegistry.FireEvent("deleting", model)
}

// FireDeleted fires the deleted event.
func FireDeleted(model any) error {
	return globalObserverRegistry.FireEvent("deleted", model)
}

// FireRestoring fires the restoring event.
func FireRestoring(model any) error {
	return globalObserverRegistry.FireEvent("restoring", model)
}

// FireRestored fires the restored event.
func FireRestored(model any) error {
	return globalObserverRegistry.FireEvent("restored", model)
}

// FireForceDeleting fires the force_deleting event.
func FireForceDeleting(model any) error {
	return globalObserverRegistry.FireEvent("force_deleting", model)
}

// FireForceDeleted fires the force_deleted event.
func FireForceDeleted(model any) error {
	return globalObserverRegistry.FireEvent("force_deleted", model)
}

// FireRetrieved fires the retrieved event.
func FireRetrieved(model any) error {
	return globalObserverRegistry.FireEvent("retrieved", model)
}
