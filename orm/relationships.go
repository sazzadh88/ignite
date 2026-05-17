package orm

// HasOne represents a one-to-one relationship where the foreign key is on the related model.
type HasOne[T any] struct {
	ForeignKey string
	LocalKey   string
	query      *Query[T]
}

// NewHasOne creates a new HasOne relationship.
func NewHasOne[T any](foreignKey, localKey string) *HasOne[T] {
	return &HasOne[T]{
		ForeignKey: foreignKey,
		LocalKey:   localKey,
		query:      NewQuery[T](),
	}
}

// Get retrieves the related model.
func (h *HasOne[T]) Get() (*T, error) {
	return h.query.First()
}

// Query returns the underlying query builder for customization.
func (h *HasOne[T]) Query() *Query[T] {
	return h.query
}

// HasMany represents a one-to-many relationship where the foreign key is on the related model.
type HasMany[T any] struct {
	ForeignKey string
	LocalKey   string
	query      *Query[T]
}

// NewHasMany creates a new HasMany relationship.
func NewHasMany[T any](foreignKey, localKey string) *HasMany[T] {
	return &HasMany[T]{
		ForeignKey: foreignKey,
		LocalKey:   localKey,
		query:      NewQuery[T](),
	}
}

// Get retrieves all related models.
func (h *HasMany[T]) Get() ([]T, error) {
	return h.query.Get()
}

// Query returns the underlying query builder for customization.
func (h *HasMany[T]) Query() *Query[T] {
	return h.query
}

// Create creates a new related model.
func (h *HasMany[T]) Create(data map[string]any) (*T, error) {
	// TODO: Set foreign key value
	return h.query.Create(data)
}

// BelongsTo represents an inverse one-to-one or many-to-one relationship.
type BelongsTo[T any] struct {
	ForeignKey string
	OwnerKey   string
	query      *Query[T]
}

// NewBelongsTo creates a new BelongsTo relationship.
func NewBelongsTo[T any](foreignKey, ownerKey string) *BelongsTo[T] {
	return &BelongsTo[T]{
		ForeignKey: foreignKey,
		OwnerKey:   ownerKey,
		query:      NewQuery[T](),
	}
}

// Get retrieves the related model.
func (b *BelongsTo[T]) Get() (*T, error) {
	return b.query.First()
}

// Query returns the underlying query builder for customization.
func (b *BelongsTo[T]) Query() *Query[T] {
	return b.query
}

// Associate associates the model with the given parent.
func (b *BelongsTo[T]) Associate(parent *T) error {
	// TODO: Set foreign key value
	return nil
}

// Dissociate removes the association.
func (b *BelongsTo[T]) Dissociate() error {
	// TODO: Set foreign key to null
	return nil
}

// BelongsToMany represents a many-to-many relationship.
type BelongsToMany[T any] struct {
	Table            string
	ForeignPivotKey  string
	RelatedPivotKey  string
	ParentKey        string
	RelatedKey       string
	query            *Query[T]
	pivotColumns     []string
	withPivot        []string
	withTimestamps   bool
}

// NewBelongsToMany creates a new BelongsToMany relationship.
func NewBelongsToMany[T any](table, foreignPivotKey, relatedPivotKey string) *BelongsToMany[T] {
	return &BelongsToMany[T]{
		Table:           table,
		ForeignPivotKey: foreignPivotKey,
		RelatedPivotKey: relatedPivotKey,
		query:           NewQuery[T](),
		pivotColumns:    make([]string, 0),
		withPivot:       make([]string, 0),
	}
}

// Get retrieves all related models.
func (b *BelongsToMany[T]) Get() ([]T, error) {
	return b.query.Get()
}

// Query returns the underlying query builder for customization.
func (b *BelongsToMany[T]) Query() *Query[T] {
	return b.query
}

// WithPivot specifies additional pivot columns to retrieve.
func (b *BelongsToMany[T]) WithPivot(columns ...string) *BelongsToMany[T] {
	b.withPivot = append(b.withPivot, columns...)
	return b
}

// WithTimestamps indicates the pivot table has created_at and updated_at columns.
func (b *BelongsToMany[T]) WithTimestamps() *BelongsToMany[T] {
	b.withTimestamps = true
	return b
}

// Attach attaches related models to the parent.
func (b *BelongsToMany[T]) Attach(ids []any, attributes map[string]any) error {
	// TODO: Insert pivot records
	return nil
}

// Detach detaches related models from the parent.
func (b *BelongsToMany[T]) Detach(ids []any) error {
	// TODO: Delete pivot records
	return nil
}

// Sync synchronizes the relationship with the given IDs.
func (b *BelongsToMany[T]) Sync(ids []any) error {
	// TODO: Sync pivot records
	return nil
}

// Toggle toggles the attachment of the given IDs.
func (b *BelongsToMany[T]) Toggle(ids []any) error {
	// TODO: Toggle pivot records
	return nil
}

// HasManyThrough represents a has-many-through relationship.
type HasManyThrough[T any] struct {
	Through   any
	FirstKey  string
	SecondKey string
	LocalKey  string
	query     *Query[T]
}

// NewHasManyThrough creates a new HasManyThrough relationship.
func NewHasManyThrough[T any](through any, firstKey, secondKey string) *HasManyThrough[T] {
	return &HasManyThrough[T]{
		Through:   through,
		FirstKey:  firstKey,
		SecondKey: secondKey,
		query:     NewQuery[T](),
	}
}

// Get retrieves all related models through the intermediate model.
func (h *HasManyThrough[T]) Get() ([]T, error) {
	return h.query.Get()
}

// Query returns the underlying query builder for customization.
func (h *HasManyThrough[T]) Query() *Query[T] {
	return h.query
}

// MorphTo represents a polymorphic belongs-to relationship.
type MorphTo struct {
	MorphType string
	MorphID   string
	models    map[string]any
}

// NewMorphTo creates a new MorphTo relationship.
func NewMorphTo(morphType, morphID string) *MorphTo {
	return &MorphTo{
		MorphType: morphType,
		MorphID:   morphID,
		models:    make(map[string]any),
	}
}

// Get retrieves the morphable model.
func (m *MorphTo) Get() (any, error) {
	// TODO: Query the appropriate model based on morph type
	return nil, nil
}

// MorphMany represents a polymorphic one-to-many relationship.
type MorphMany[T any] struct {
	Name      string
	MorphType string
	MorphID   string
	query     *Query[T]
}

// NewMorphMany creates a new MorphMany relationship.
func NewMorphMany[T any](name string) *MorphMany[T] {
	return &MorphMany[T]{
		Name:      name,
		MorphType: name + "_type",
		MorphID:   name + "_id",
		query:     NewQuery[T](),
	}
}

// Get retrieves all related models.
func (m *MorphMany[T]) Get() ([]T, error) {
	return m.query.Get()
}

// Query returns the underlying query builder for customization.
func (m *MorphMany[T]) Query() *Query[T] {
	return m.query
}

// Create creates a new related model.
func (m *MorphMany[T]) Create(data map[string]any) (*T, error) {
	// TODO: Set morph type and ID
	return m.query.Create(data)
}

// MorphOne represents a polymorphic one-to-one relationship.
type MorphOne[T any] struct {
	Name      string
	MorphType string
	MorphID   string
	query     *Query[T]
}

// NewMorphOne creates a new MorphOne relationship.
func NewMorphOne[T any](name string) *MorphOne[T] {
	return &MorphOne[T]{
		Name:      name,
		MorphType: name + "_type",
		MorphID:   name + "_id",
		query:     NewQuery[T](),
	}
}

// Get retrieves the related model.
func (m *MorphOne[T]) Get() (*T, error) {
	return m.query.First()
}

// Query returns the underlying query builder for customization.
func (m *MorphOne[T]) Query() *Query[T] {
	return m.query
}
