// Package resources provides JSON API resource transformers for converting
// domain models to API responses, similar to Laravel's API Resources.
package resources

// Resource wraps a single data item for transformation to JSON API format.
type Resource[T any] struct {
	Data T
}

// ResourceCollection wraps a collection of items with optional metadata.
type ResourceCollection[T any] struct {
	Data []T
	Meta map[string]any
}

// Transformer defines the interface for converting domain models to API representation.
type Transformer[T any] interface {
	ToArray(data T) map[string]any
}

// Make transforms a single resource using the provided transformer.
func Make[T any](data T, transformer Transformer[T]) map[string]any {
	return transformer.ToArray(data)
}

// Collection transforms a collection of resources using the provided transformer.
func Collection[T any](data []T, transformer Transformer[T]) []map[string]any {
	result := make([]map[string]any, len(data))
	for i, item := range data {
		result[i] = transformer.ToArray(item)
	}
	return result
}
