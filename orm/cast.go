package orm

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// Caster defines the interface for attribute casting.
type Caster interface {
	// Get transforms a raw database value into a Go value.
	Get(value any) any

	// Set transforms a Go value into a database value.
	Set(value any) any
}

// Casts is a map of field names to their casters.
type Casts map[string]Caster

// JSONCast casts values to/from JSON.
type JSONCast struct{}

// Get decodes JSON into a map or slice.
func (c JSONCast) Get(value any) any {
	if value == nil {
		return nil
	}

	var result any
	switch v := value.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &result); err != nil {
			return nil
		}
	case []byte:
		if err := json.Unmarshal(v, &result); err != nil {
			return nil
		}
	default:
		return value
	}

	return result
}

// Set encodes a value to JSON.
func (c JSONCast) Set(value any) any {
	if value == nil {
		return nil
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return nil
	}

	return string(bytes)
}

// BoolCast casts values to/from bool.
type BoolCast struct{}

// Get converts a database value to bool.
func (c BoolCast) Get(value any) any {
	if value == nil {
		return false
	}

	switch v := value.(type) {
	case bool:
		return v
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", v) != "0"
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v) != "0"
	case string:
		return v == "1" || v == "true" || v == "yes" || v == "on"
	default:
		return false
	}
}

// Set converts a bool to a database value.
func (c BoolCast) Set(value any) any {
	if b, ok := value.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	return 0
}

// IntCast casts values to/from int.
type IntCast struct{}

// Get converts a database value to int.
func (c IntCast) Get(value any) any {
	if value == nil {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	default:
		return 0
	}
}

// Set converts an int to a database value.
func (c IntCast) Set(value any) any {
	if i, ok := value.(int); ok {
		return i
	}
	return 0
}

// FloatCast casts values to/from float64.
type FloatCast struct{}

// Get converts a database value to float64.
func (c FloatCast) Get(value any) any {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	default:
		return 0.0
	}
}

// Set converts a float64 to a database value.
func (c FloatCast) Set(value any) any {
	if f, ok := value.(float64); ok {
		return f
	}
	return 0.0
}

// StringCast casts values to/from string.
type StringCast struct{}

// Get converts a database value to string.
func (c StringCast) Get(value any) any {
	if value == nil {
		return ""
	}

	return fmt.Sprintf("%v", value)
}

// Set converts a string to a database value.
func (c StringCast) Set(value any) any {
	if s, ok := value.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", value)
}

// DateCast casts values to/from time.Time.
type DateCast struct {
	Format string
}

// Get converts a database value to time.Time.
func (c DateCast) Get(value any) any {
	if value == nil {
		return time.Time{}
	}

	switch v := value.(type) {
	case time.Time:
		return v
	case string:
		format := c.Format
		if format == "" {
			format = time.RFC3339
		}
		t, _ := time.Parse(format, v)
		return t
	default:
		return time.Time{}
	}
}

// Set converts a time.Time to a database value.
func (c DateCast) Set(value any) any {
	if t, ok := value.(time.Time); ok {
		format := c.Format
		if format == "" {
			format = time.RFC3339
		}
		return t.Format(format)
	}
	return nil
}

// ArrayCast casts values to/from slices.
type ArrayCast struct {
	Separator string
}

// Get converts a database value to a slice.
func (c ArrayCast) Get(value any) any {
	if value == nil {
		return []string{}
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			return []string{}
		}
		sep := c.Separator
		if sep == "" {
			sep = ","
		}
		// Try JSON first
		var arr []string
		if err := json.Unmarshal([]byte(v), &arr); err == nil {
			return arr
		}
		// Fall back to separator
		return splitString(v, sep)
	case []byte:
		var arr []string
		if err := json.Unmarshal(v, &arr); err == nil {
			return arr
		}
		return []string{}
	case []any:
		result := make([]string, len(v))
		for i, item := range v {
			result[i] = fmt.Sprintf("%v", item)
		}
		return result
	default:
		return []string{}
	}
}

// Set converts a slice to a database value.
func (c ArrayCast) Set(value any) any {
	if value == nil {
		return "[]"
	}

	bytes, err := json.Marshal(value)
	if err != nil {
		return "[]"
	}

	return string(bytes)
}

// splitString splits a string by a separator.
func splitString(s string, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])

	return result
}

// CastAttribute applies the appropriate caster to a value.
func CastAttribute(casts Casts, key string, value any, get bool) any {
	if caster, ok := casts[key]; ok {
		if get {
			return caster.Get(value)
		}
		return caster.Set(value)
	}
	return value
}

// ApplyCasts applies all casters to a map of attributes.
func ApplyCasts(casts Casts, attributes map[string]any, get bool) map[string]any {
	result := make(map[string]any)
	for k, v := range attributes {
		result[k] = CastAttribute(casts, k, v, get)
	}
	return result
}
