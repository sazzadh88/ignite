package support

import (
	"crypto/rand"
	"math/big"
	"net/url"
	"sort"
	"strings"
)

// ArrayHelper provides array and map manipulation utilities.
type ArrayHelper struct{}

// Arr is the global array helper instance.
var Arr = ArrayHelper{}

// Get retrieves a value from a map using dot notation.
func (a ArrayHelper) Get(data map[string]any, key string, defaultVal ...any) any {
	if !strings.Contains(key, ".") {
		if val, ok := data[key]; ok {
			return val
		}
		if len(defaultVal) > 0 {
			return defaultVal[0]
		}
		return nil
	}

	keys := strings.Split(key, ".")
	current := any(data)

	for _, k := range keys {
		switch v := current.(type) {
		case map[string]any:
			if val, ok := v[k]; ok {
				current = val
			} else {
				if len(defaultVal) > 0 {
					return defaultVal[0]
				}
				return nil
			}
		default:
			if len(defaultVal) > 0 {
				return defaultVal[0]
			}
			return nil
		}
	}

	return current
}

// Set sets a value in a map using dot notation.
func (a ArrayHelper) Set(data map[string]any, key string, value any) map[string]any {
	if data == nil {
		data = make(map[string]any)
	}

	if !strings.Contains(key, ".") {
		data[key] = value
		return data
	}

	keys := strings.Split(key, ".")
	current := data

	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		if _, ok := current[k]; !ok {
			current[k] = make(map[string]any)
		}
		if nested, ok := current[k].(map[string]any); ok {
			current = nested
		} else {
			current[k] = make(map[string]any)
			current = current[k].(map[string]any)
		}
	}

	current[keys[len(keys)-1]] = value
	return data
}

// Has checks if a key exists in a map using dot notation.
func (a ArrayHelper) Has(data map[string]any, key string) bool {
	if !strings.Contains(key, ".") {
		_, ok := data[key]
		return ok
	}

	keys := strings.Split(key, ".")
	current := any(data)

	for _, k := range keys {
		switch v := current.(type) {
		case map[string]any:
			if val, ok := v[k]; ok {
				current = val
			} else {
				return false
			}
		default:
			return false
		}
	}

	return true
}

// Forget removes a key from a map using dot notation.
func (a ArrayHelper) Forget(data map[string]any, key string) map[string]any {
	if data == nil {
		return data
	}

	if !strings.Contains(key, ".") {
		delete(data, key)
		return data
	}

	keys := strings.Split(key, ".")
	current := data

	for i := 0; i < len(keys)-1; i++ {
		k := keys[i]
		if nested, ok := current[k].(map[string]any); ok {
			current = nested
		} else {
			return data
		}
	}

	delete(current, keys[len(keys)-1])
	return data
}

// Only returns a map with only the specified keys.
func (a ArrayHelper) Only(data map[string]any, keys []string) map[string]any {
	result := make(map[string]any)
	for _, key := range keys {
		if val, ok := data[key]; ok {
			result[key] = val
		}
	}
	return result
}

// Except returns a map with all keys except the specified ones.
func (a ArrayHelper) Except(data map[string]any, keys []string) map[string]any {
	result := make(map[string]any)
	excluded := make(map[string]bool)
	for _, key := range keys {
		excluded[key] = true
	}
	for k, v := range data {
		if !excluded[k] {
			result[k] = v
		}
	}
	return result
}

// Dot flattens a nested map to dot notation.
func (a ArrayHelper) Dot(data map[string]any) map[string]any {
	result := make(map[string]any)
	a.dotRecursive(data, "", result)
	return result
}

func (a ArrayHelper) dotRecursive(data map[string]any, prefix string, result map[string]any) {
	for k, v := range data {
		key := k
		if prefix != "" {
			key = prefix + "." + k
		}
		if nested, ok := v.(map[string]any); ok {
			a.dotRecursive(nested, key, result)
		} else {
			result[key] = v
		}
	}
}

// Undot expands a dot notation map to nested structure.
func (a ArrayHelper) Undot(data map[string]any) map[string]any {
	result := make(map[string]any)
	for key, value := range data {
		a.Set(result, key, value)
	}
	return result
}

// Flatten flattens a multi-dimensional array/map into a single level.
func (a ArrayHelper) Flatten(data any, depth ...int) []any {
	maxDepth := -1
	if len(depth) > 0 {
		maxDepth = depth[0]
	}
	return a.flattenRecursive(data, 0, maxDepth)
}

func (a ArrayHelper) flattenRecursive(data any, currentDepth, maxDepth int) []any {
	result := []any{}

	switch v := data.(type) {
	case []any:
		for _, item := range v {
			if maxDepth == -1 || currentDepth < maxDepth {
				result = append(result, a.flattenRecursive(item, currentDepth+1, maxDepth)...)
			} else {
				result = append(result, item)
			}
		}
	case map[string]any:
		for _, value := range v {
			if maxDepth == -1 || currentDepth < maxDepth {
				result = append(result, a.flattenRecursive(value, currentDepth+1, maxDepth)...)
			} else {
				result = append(result, value)
			}
		}
	default:
		result = append(result, data)
	}

	return result
}

// Keys returns all keys from a map.
func (a ArrayHelper) Keys(data map[string]any) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Values returns all values from a map.
func (a ArrayHelper) Values(data map[string]any) []any {
	values := make([]any, 0, len(data))
	for _, v := range data {
		values = append(values, v)
	}
	return values
}

// Pluck extracts values from a slice of maps by key.
func (a ArrayHelper) Pluck(items []map[string]any, key string) []any {
	result := make([]any, 0, len(items))
	for _, item := range items {
		if val, ok := item[key]; ok {
			result = append(result, val)
		}
	}
	return result
}

// Where filters a slice of maps using a callback function.
func (a ArrayHelper) Where(items []map[string]any, fn func(map[string]any) bool) []map[string]any {
	result := []map[string]any{}
	for _, item := range items {
		if fn(item) {
			result = append(result, item)
		}
	}
	return result
}

// First returns the first element that passes a truth test.
func (a ArrayHelper) First(items []any, fn func(any) bool) any {
	for _, item := range items {
		if fn(item) {
			return item
		}
	}
	return nil
}

// Last returns the last element that passes a truth test.
func (a ArrayHelper) Last(items []any, fn func(any) bool) any {
	for i := len(items) - 1; i >= 0; i-- {
		if fn(items[i]) {
			return items[i]
		}
	}
	return nil
}

// Shuffle randomly shuffles an array.
func (a ArrayHelper) Shuffle(items []any) []any {
	result := make([]any, len(items))
	copy(result, items)

	for i := len(result) - 1; i > 0; i-- {
		j, _ := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		result[i], result[j.Int64()] = result[j.Int64()], result[i]
	}

	return result
}

// Random returns a random element from an array, or n random elements if specified.
func (a ArrayHelper) Random(items []any, n ...int) any {
	if len(items) == 0 {
		return nil
	}

	if len(n) > 0 && n[0] > 1 {
		count := n[0]
		if count > len(items) {
			count = len(items)
		}
		shuffled := a.Shuffle(items)
		return shuffled[:count]
	}

	idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(items))))
	return items[idx.Int64()]
}

// Wrap wraps a value in an array if it's not already an array.
func (a ArrayHelper) Wrap(value any) []any {
	if value == nil {
		return []any{}
	}
	if arr, ok := value.([]any); ok {
		return arr
	}
	return []any{value}
}

// Query builds a URL query string from a map.
func (a ArrayHelper) Query(data map[string]string) string {
	values := url.Values{}
	for k, v := range data {
		values.Set(k, v)
	}
	return values.Encode()
}
