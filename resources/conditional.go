package resources

import "reflect"

// sentinel is a private type used to represent missing values.
type sentinel struct{}

var missing = sentinel{}

// When returns the value if the condition is true, otherwise returns a sentinel missing value.
// Use this for conditional inclusion of attributes in resource transformations.
func When(condition bool, value any) any {
	if condition {
		return value
	}
	return missing
}

// Unless returns the value if the condition is false, otherwise returns a sentinel missing value.
// It's the inverse of When.
func Unless(condition bool, value any) any {
	return When(!condition, value)
}

// WhenNotNil returns the value if it's not nil, otherwise returns a sentinel missing value.
func WhenNotNil(value any) any {
	if value == nil {
		return missing
	}

	// Use reflection to check for typed nil pointers, slices, maps, channels, and functions
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map, reflect.Chan, reflect.Func, reflect.Interface:
		if v.IsNil() {
			return missing
		}
	}

	return value
}

// MergeWhen conditionally merges a map into another map based on the condition.
// If the condition is true, it returns the data map, otherwise returns an empty map.
func MergeWhen(condition bool, data map[string]any) map[string]any {
	if condition {
		return data
	}
	return map[string]any{}
}

// IsMissing checks if a value is the sentinel missing value.
func IsMissing(value any) bool {
	_, ok := value.(sentinel)
	return ok
}

// CleanMap removes all missing sentinel values from a map.
// This is useful after building a resource with conditional fields.
func CleanMap(data map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		if !IsMissing(value) {
			result[key] = value
		}
	}
	return result
}
