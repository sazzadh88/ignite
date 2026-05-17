// Package collection provides a generic, Laravel-inspired collection type for Go.
// All operations are immutable and return new collections.
package collection

import (
	"encoding/json"
	"math/rand"
	"time"
)

// Collection is a generic container for a slice of items.
type Collection[T any] struct {
	items []T
}

// Make creates a new Collection from a slice.
func Make[T any](items []T) *Collection[T] {
	copied := make([]T, len(items))
	copy(copied, items)
	return &Collection[T]{items: copied}
}

// Range creates a Collection of integers from start to end (exclusive).
func Range(start, end int) *Collection[int] {
	if start >= end {
		return &Collection[int]{items: []int{}}
	}
	items := make([]int, end-start)
	for i := range items {
		items[i] = start + i
	}
	return &Collection[int]{items: items}
}

// All returns all items as a slice.
func (c *Collection[T]) All() []T {
	result := make([]T, len(c.items))
	copy(result, c.items)
	return result
}

// Count returns the number of items.
func (c *Collection[T]) Count() int {
	return len(c.items)
}

// IsEmpty returns true if the collection has no items.
func (c *Collection[T]) IsEmpty() bool {
	return len(c.items) == 0
}

// IsNotEmpty returns true if the collection has items.
func (c *Collection[T]) IsNotEmpty() bool {
	return len(c.items) > 0
}

// First returns the first item, optionally matching a predicate.
func (c *Collection[T]) First(fn ...func(T) bool) (T, bool) {
	var zero T
	if len(fn) == 0 {
		if len(c.items) == 0 {
			return zero, false
		}
		return c.items[0], true
	}
	for _, item := range c.items {
		if fn[0](item) {
			return item, true
		}
	}
	return zero, false
}

// Last returns the last item, optionally matching a predicate.
func (c *Collection[T]) Last(fn ...func(T) bool) (T, bool) {
	var zero T
	if len(fn) == 0 {
		if len(c.items) == 0 {
			return zero, false
		}
		return c.items[len(c.items)-1], true
	}
	for i := len(c.items) - 1; i >= 0; i-- {
		if fn[0](c.items[i]) {
			return c.items[i], true
		}
	}
	return zero, false
}

// Filter returns a new collection with items matching the predicate.
func (c *Collection[T]) Filter(fn func(T) bool) *Collection[T] {
	result := []T{}
	for _, item := range c.items {
		if fn(item) {
			result = append(result, item)
		}
	}
	return &Collection[T]{items: result}
}

// Reject returns a new collection with items not matching the predicate.
func (c *Collection[T]) Reject(fn func(T) bool) *Collection[T] {
	result := []T{}
	for _, item := range c.items {
		if !fn(item) {
			result = append(result, item)
		}
	}
	return &Collection[T]{items: result}
}

// Map applies a transformation function to each item.
func (c *Collection[T]) Map(fn func(T) T) *Collection[T] {
	result := make([]T, len(c.items))
	for i, item := range c.items {
		result[i] = fn(item)
	}
	return &Collection[T]{items: result}
}

// Each iterates over all items with index.
func (c *Collection[T]) Each(fn func(T, int)) {
	for i, item := range c.items {
		fn(item, i)
	}
}

// Contains returns true if any item matches the predicate.
func (c *Collection[T]) Contains(fn func(T) bool) bool {
	for _, item := range c.items {
		if fn(item) {
			return true
		}
	}
	return false
}

// Every returns true if all items match the predicate.
func (c *Collection[T]) Every(fn func(T) bool) bool {
	for _, item := range c.items {
		if !fn(item) {
			return false
		}
	}
	return true
}

// Some returns true if at least one item matches the predicate.
func (c *Collection[T]) Some(fn func(T) bool) bool {
	return c.Contains(fn)
}

// Reduce applies an accumulator function to all items.
func (c *Collection[T]) Reduce(fn func(carry any, item T) any, initial any) any {
	carry := initial
	for _, item := range c.items {
		carry = fn(carry, item)
	}
	return carry
}

// Chunk splits the collection into chunks of the given size.
func (c *Collection[T]) Chunk(size int) [][]T {
	if size <= 0 {
		return [][]T{}
	}
	result := [][]T{}
	for i := 0; i < len(c.items); i += size {
		end := i + size
		if end > len(c.items) {
			end = len(c.items)
		}
		result = append(result, c.items[i:end])
	}
	return result
}

// Flatten flattens a collection one level deep.
func (c *Collection[T]) Flatten() *Collection[T] {
	result := []T{}
	for _, item := range c.items {
		result = append(result, item)
	}
	return &Collection[T]{items: result}
}

// Unique returns a collection with duplicate items removed.
// Optional key extractor function can be provided.
func (c *Collection[T]) Unique(fn ...func(T) any) *Collection[T] {
	if len(c.items) == 0 {
		return &Collection[T]{items: []T{}}
	}

	seen := make(map[any]bool)
	result := []T{}

	if len(fn) == 0 {
		// No key function, use item itself as key (requires comparable)
		// We'll use a simple approach: compare by checking all previous items
		for _, item := range c.items {
			found := false
			for _, r := range result {
				// This is a workaround since we can't directly compare generics
				// In practice, this works for basic types
				if any(item) == any(r) {
					found = true
					break
				}
			}
			if !found {
				result = append(result, item)
			}
		}
		return &Collection[T]{items: result}
	}

	keyFn := fn[0]
	for _, item := range c.items {
		key := keyFn(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return &Collection[T]{items: result}
}

// Reverse returns a collection with items in reverse order.
func (c *Collection[T]) Reverse() *Collection[T] {
	result := make([]T, len(c.items))
	for i, item := range c.items {
		result[len(c.items)-1-i] = item
	}
	return &Collection[T]{items: result}
}

// Sort returns a sorted collection using the provided comparison function.
func (c *Collection[T]) Sort(fn func(a, b T) bool) *Collection[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)
	// Simple bubble sort for simplicity (no external deps)
	for i := 0; i < len(result); i++ {
		for j := i + 1; j < len(result); j++ {
			if !fn(result[i], result[j]) {
				result[i], result[j] = result[j], result[i]
			}
		}
	}
	return &Collection[T]{items: result}
}

// Shuffle returns a collection with items in random order.
func (c *Collection[T]) Shuffle() *Collection[T] {
	result := make([]T, len(c.items))
	copy(result, c.items)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})
	return &Collection[T]{items: result}
}

// Slice returns a subset of items from offset with given length.
func (c *Collection[T]) Slice(offset, length int) *Collection[T] {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(c.items) {
		return &Collection[T]{items: []T{}}
	}
	end := offset + length
	if end > len(c.items) {
		end = len(c.items)
	}
	result := make([]T, end-offset)
	copy(result, c.items[offset:end])
	return &Collection[T]{items: result}
}

// Take returns the first n items.
func (c *Collection[T]) Take(n int) *Collection[T] {
	if n < 0 {
		return c.Slice(len(c.items)+n, -n)
	}
	return c.Slice(0, n)
}

// Skip returns all items except the first n.
func (c *Collection[T]) Skip(n int) *Collection[T] {
	if n < 0 {
		n = 0
	}
	if n >= len(c.items) {
		return &Collection[T]{items: []T{}}
	}
	result := make([]T, len(c.items)-n)
	copy(result, c.items[n:])
	return &Collection[T]{items: result}
}

// Pop removes and returns the last item.
func (c *Collection[T]) Pop() (T, *Collection[T]) {
	var zero T
	if len(c.items) == 0 {
		return zero, c
	}
	last := c.items[len(c.items)-1]
	result := make([]T, len(c.items)-1)
	copy(result, c.items[:len(c.items)-1])
	return last, &Collection[T]{items: result}
}

// Push appends items to the end.
func (c *Collection[T]) Push(items ...T) *Collection[T] {
	result := make([]T, len(c.items)+len(items))
	copy(result, c.items)
	copy(result[len(c.items):], items)
	return &Collection[T]{items: result}
}

// Prepend adds items to the beginning.
func (c *Collection[T]) Prepend(items ...T) *Collection[T] {
	result := make([]T, len(items)+len(c.items))
	copy(result, items)
	copy(result[len(items):], c.items)
	return &Collection[T]{items: result}
}

// Concat merges another collection.
func (c *Collection[T]) Concat(other *Collection[T]) *Collection[T] {
	return c.Push(other.items...)
}

// Diff returns items in this collection not in the other.
func (c *Collection[T]) Diff(other *Collection[T], eq func(T, T) bool) *Collection[T] {
	result := []T{}
	for _, item := range c.items {
		found := false
		for _, otherItem := range other.items {
			if eq(item, otherItem) {
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}
	return &Collection[T]{items: result}
}

// Intersect returns items present in both collections.
func (c *Collection[T]) Intersect(other *Collection[T], eq func(T, T) bool) *Collection[T] {
	result := []T{}
	for _, item := range c.items {
		for _, otherItem := range other.items {
			if eq(item, otherItem) {
				result = append(result, item)
				break
			}
		}
	}
	return &Collection[T]{items: result}
}

// Nth returns every nth item, starting at offset.
func (c *Collection[T]) Nth(step int, offset ...int) *Collection[T] {
	if step <= 0 {
		return &Collection[T]{items: []T{}}
	}
	start := 0
	if len(offset) > 0 {
		start = offset[0]
	}
	result := []T{}
	for i := start; i < len(c.items); i += step {
		result = append(result, c.items[i])
	}
	return &Collection[T]{items: result}
}

// Partition splits the collection into two based on predicate.
func (c *Collection[T]) Partition(fn func(T) bool) (*Collection[T], *Collection[T]) {
	truthy := []T{}
	falsy := []T{}
	for _, item := range c.items {
		if fn(item) {
			truthy = append(truthy, item)
		} else {
			falsy = append(falsy, item)
		}
	}
	return &Collection[T]{items: truthy}, &Collection[T]{items: falsy}
}

// Tap executes a callback with the collection and returns it.
func (c *Collection[T]) Tap(fn func(*Collection[T])) *Collection[T] {
	fn(c)
	return c
}

// Pipe passes the collection to a callback and returns the result.
func (c *Collection[T]) Pipe(fn func(*Collection[T]) any) any {
	return fn(c)
}

// When executes a callback if the condition is true.
func (c *Collection[T]) When(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	if condition {
		return fn(c)
	}
	return c
}

// Unless executes a callback if the condition is false.
func (c *Collection[T]) Unless(condition bool, fn func(*Collection[T]) *Collection[T]) *Collection[T] {
	if !condition {
		return fn(c)
	}
	return c
}

// ForPage returns items for the given page and page size.
func (c *Collection[T]) ForPage(page, perPage int) *Collection[T] {
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * perPage
	return c.Slice(offset, perPage)
}

// Random returns n random items from the collection.
func (c *Collection[T]) Random(n ...int) *Collection[T] {
	count := 1
	if len(n) > 0 {
		count = n[0]
	}
	if count <= 0 || len(c.items) == 0 {
		return &Collection[T]{items: []T{}}
	}
	if count >= len(c.items) {
		return c.Shuffle()
	}
	shuffled := c.Shuffle()
	return shuffled.Take(count)
}

// ToJSON serializes the collection to JSON.
func (c *Collection[T]) ToJSON() ([]byte, error) {
	return json.Marshal(c.items)
}
