package collection

// MapCollection is a generic map-based collection.
type MapCollection[K comparable, V any] struct {
	items map[K]V
}

// MakeMap creates a new MapCollection from a map.
func MakeMap[K comparable, V any](items map[K]V) *MapCollection[K, V] {
	copied := make(map[K]V, len(items))
	for k, v := range items {
		copied[k] = v
	}
	return &MapCollection[K, V]{items: copied}
}

// Keys returns all keys as a Collection.
func (m *MapCollection[K, V]) Keys() *Collection[K] {
	keys := make([]K, 0, len(m.items))
	for k := range m.items {
		keys = append(keys, k)
	}
	return &Collection[K]{items: keys}
}

// Values returns all values as a Collection.
func (m *MapCollection[K, V]) Values() *Collection[V] {
	values := make([]V, 0, len(m.items))
	for _, v := range m.items {
		values = append(values, v)
	}
	return &Collection[V]{items: values}
}

// Has checks if a key exists.
func (m *MapCollection[K, V]) Has(key K) bool {
	_, exists := m.items[key]
	return exists
}

// Get retrieves a value by key, returns zero value and false if not found.
func (m *MapCollection[K, V]) Get(key K) (V, bool) {
	val, exists := m.items[key]
	return val, exists
}

// Put sets a key-value pair and returns a new MapCollection.
func (m *MapCollection[K, V]) Put(key K, val V) *MapCollection[K, V] {
	newItems := make(map[K]V, len(m.items)+1)
	for k, v := range m.items {
		newItems[k] = v
	}
	newItems[key] = val
	return &MapCollection[K, V]{items: newItems}
}

// Only returns a new MapCollection with only the specified keys.
func (m *MapCollection[K, V]) Only(keys []K) *MapCollection[K, V] {
	result := make(map[K]V)
	for _, key := range keys {
		if val, exists := m.items[key]; exists {
			result[key] = val
		}
	}
	return &MapCollection[K, V]{items: result}
}

// Except returns a new MapCollection without the specified keys.
func (m *MapCollection[K, V]) Except(keys []K) *MapCollection[K, V] {
	excluded := make(map[K]bool)
	for _, key := range keys {
		excluded[key] = true
	}
	result := make(map[K]V)
	for k, v := range m.items {
		if !excluded[k] {
			result[k] = v
		}
	}
	return &MapCollection[K, V]{items: result}
}

// Filter returns a new MapCollection with entries matching the predicate.
func (m *MapCollection[K, V]) Filter(fn func(K, V) bool) *MapCollection[K, V] {
	result := make(map[K]V)
	for k, v := range m.items {
		if fn(k, v) {
			result[k] = v
		}
	}
	return &MapCollection[K, V]{items: result}
}

// Map applies a transformation function to all values.
func (m *MapCollection[K, V]) Map(fn func(K, V) V) *MapCollection[K, V] {
	result := make(map[K]V, len(m.items))
	for k, v := range m.items {
		result[k] = fn(k, v)
	}
	return &MapCollection[K, V]{items: result}
}

// Merge combines this MapCollection with another.
func (m *MapCollection[K, V]) Merge(other *MapCollection[K, V]) *MapCollection[K, V] {
	result := make(map[K]V, len(m.items)+len(other.items))
	for k, v := range m.items {
		result[k] = v
	}
	for k, v := range other.items {
		result[k] = v
	}
	return &MapCollection[K, V]{items: result}
}

// Count returns the number of items.
func (m *MapCollection[K, V]) Count() int {
	return len(m.items)
}

// IsEmpty returns true if the collection has no items.
func (m *MapCollection[K, V]) IsEmpty() bool {
	return len(m.items) == 0
}

// ToMap returns the underlying map as a copy.
func (m *MapCollection[K, V]) ToMap() map[K]V {
	result := make(map[K]V, len(m.items))
	for k, v := range m.items {
		result[k] = v
	}
	return result
}
