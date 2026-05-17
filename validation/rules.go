package validation

import (
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/url"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// builtInRules maps rule names to their validation functions.
var builtInRules = map[string]func(attribute string, value any, params []string, data map[string]any) bool{
	"required":         validateRequired,
	"required_if":      validateRequiredIf,
	"required_with":    validateRequiredWith,
	"required_without": validateRequiredWithout,
	"nullable":         validateNullable,
	"string":           validateString,
	"integer":          validateInteger,
	"numeric":          validateNumeric,
	"boolean":          validateBoolean,
	"array":            validateArray,
	"in":               validateIn,
	"not_in":           validateNotIn,
	"min":              validateMin,
	"max":              validateMax,
	"between":          validateBetween,
	"size":             validateSize,
	"email":            validateEmail,
	"url":              validateURL,
	"ip":               validateIP,
	"uuid":             validateUUID,
	"regex":            validateRegex,
	"not_regex":        validateNotRegex,
	"alpha":            validateAlpha,
	"alpha_num":        validateAlphaNum,
	"alpha_dash":       validateAlphaDash,
	"starts_with":      validateStartsWith,
	"ends_with":        validateEndsWith,
	"contains":         validateContains,
	"date":             validateDate,
	"after":            validateAfter,
	"before":           validateBefore,
	"confirmed":        validateConfirmed,
	"same":             validateSame,
	"different":        validateDifferent,
	"gt":               validateGt,
	"gte":              validateGte,
	"lt":               validateLt,
	"lte":              validateLte,
	"json":             validateJSON,
	"unique":           validateUnique,
	"exists":           validateExists,
	"digits":           validateDigits,
	"digits_between":   validateDigitsBetween,
}

// defaultMessages provides default error messages for built-in rules.
var defaultMessages = map[string]string{
	"required":         "The :attribute field is required.",
	"required_if":      "The :attribute field is required when :other is :value.",
	"required_with":    "The :attribute field is required when :other is present.",
	"required_without": "The :attribute field is required when :other is not present.",
	"string":           "The :attribute must be a string.",
	"integer":          "The :attribute must be an integer.",
	"numeric":          "The :attribute must be numeric.",
	"boolean":          "The :attribute must be true or false.",
	"array":            "The :attribute must be an array.",
	"in":               "The :attribute must be one of: :values.",
	"not_in":           "The :attribute must not be one of: :values.",
	"min":              "The :attribute must be at least :min.",
	"max":              "The :attribute must not exceed :max.",
	"between":          "The :attribute must be between :min and :max.",
	"size":             "The :attribute must be :size.",
	"email":            "The :attribute must be a valid email address.",
	"url":              "The :attribute must be a valid URL.",
	"ip":               "The :attribute must be a valid IP address.",
	"uuid":             "The :attribute must be a valid UUID.",
	"regex":            "The :attribute format is invalid.",
	"not_regex":        "The :attribute format is invalid.",
	"alpha":            "The :attribute may only contain letters.",
	"alpha_num":        "The :attribute may only contain letters and numbers.",
	"alpha_dash":       "The :attribute may only contain letters, numbers, dashes and underscores.",
	"starts_with":      "The :attribute must start with one of: :values.",
	"ends_with":        "The :attribute must end with one of: :values.",
	"contains":         "The :attribute must contain :value.",
	"date":             "The :attribute must be a valid date.",
	"after":            "The :attribute must be after :date.",
	"before":           "The :attribute must be before :date.",
	"confirmed":        "The :attribute confirmation does not match.",
	"same":             "The :attribute and :other must match.",
	"different":        "The :attribute and :other must be different.",
	"gt":               "The :attribute must be greater than :other.",
	"gte":              "The :attribute must be greater than or equal to :other.",
	"lt":               "The :attribute must be less than :other.",
	"lte":              "The :attribute must be less than or equal to :other.",
	"json":             "The :attribute must be a valid JSON string.",
	"unique":           "The :attribute has already been taken.",
	"exists":           "The selected :attribute is invalid.",
	"digits":           "The :attribute must be :digits digits.",
	"digits_between":   "The :attribute must be between :min and :max digits.",
}

func validateRequired(attribute string, value any, params []string, data map[string]any) bool {
	return !isEmpty(value)
}

func validateRequiredIf(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 2 {
		return true
	}
	otherField := params[0]
	otherValue := params[1]
	if v, ok := data[otherField]; ok && fmt.Sprintf("%v", v) == otherValue {
		return !isEmpty(value)
	}
	return true
}

func validateRequiredWith(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherField := params[0]
	if v, ok := data[otherField]; ok && !isEmpty(v) {
		return !isEmpty(value)
	}
	return true
}

func validateRequiredWithout(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherField := params[0]
	if v, ok := data[otherField]; !ok || isEmpty(v) {
		return !isEmpty(value)
	}
	return true
}

func validateNullable(attribute string, value any, params []string, data map[string]any) bool {
	return true
}

func validateString(attribute string, value any, params []string, data map[string]any) bool {
	_, ok := value.(string)
	return ok
}

func validateInteger(attribute string, value any, params []string, data map[string]any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		f, _ := toFloat64(value)
		return f == float64(int64(f))
	case string:
		_, err := strconv.ParseInt(value.(string), 10, 64)
		return err == nil
	}
	return false
}

func validateNumeric(attribute string, value any, params []string, data map[string]any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return true
	case string:
		_, err := strconv.ParseFloat(value.(string), 64)
		return err == nil
	}
	return false
}

func validateBoolean(attribute string, value any, params []string, data map[string]any) bool {
	switch v := value.(type) {
	case bool:
		return true
	case string:
		lower := strings.ToLower(v)
		return lower == "true" || lower == "false" || lower == "1" || lower == "0" || lower == "yes" || lower == "no"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return true
	}
	return false
}

func validateArray(attribute string, value any, params []string, data map[string]any) bool {
	v := reflect.ValueOf(value)
	return v.Kind() == reflect.Slice || v.Kind() == reflect.Array
}

func validateIn(attribute string, value any, params []string, data map[string]any) bool {
	strVal := fmt.Sprintf("%v", value)
	for _, p := range params {
		if strVal == p {
			return true
		}
	}
	return false
}

func validateNotIn(attribute string, value any, params []string, data map[string]any) bool {
	return !validateIn(attribute, value, params, data)
}

func validateMin(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	min, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return float64(len(v)) >= min
	default:
		if num, ok := toFloat64(value); ok {
			return num >= min
		}
	}
	return false
}

func validateMax(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	max, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return float64(len(v)) <= max
	default:
		if num, ok := toFloat64(value); ok {
			return num <= max
		}
	}
	return false
}

func validateBetween(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 2 {
		return true
	}
	min, err1 := strconv.ParseFloat(params[0], 64)
	max, err2 := strconv.ParseFloat(params[1], 64)
	if err1 != nil || err2 != nil {
		return false
	}

	switch v := value.(type) {
	case string:
		length := float64(len(v))
		return length >= min && length <= max
	default:
		if num, ok := toFloat64(value); ok {
			return num >= min && num <= max
		}
	}
	return false
}

func validateSize(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	size, err := strconv.ParseFloat(params[0], 64)
	if err != nil {
		return false
	}

	switch v := value.(type) {
	case string:
		return float64(len(v)) == size
	default:
		if num, ok := toFloat64(value); ok {
			return num == size
		}
	}
	return false
}

func validateEmail(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	_, err := mail.ParseAddress(str)
	return err == nil
}

func validateURL(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func validateIP(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	return net.ParseIP(str) != nil
}

func validateUUID(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(str)
}

func validateRegex(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	str, ok := value.(string)
	if !ok {
		return false
	}
	re, err := regexp.Compile(params[0])
	if err != nil {
		return false
	}
	return re.MatchString(str)
}

func validateNotRegex(attribute string, value any, params []string, data map[string]any) bool {
	return !validateRegex(attribute, value, params, data)
}

func validateAlpha(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	return alphaRegex.MatchString(str)
}

func validateAlphaNum(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	alphaNumRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	return alphaNumRegex.MatchString(str)
}

func validateAlphaDash(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	alphaDashRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return alphaDashRegex.MatchString(str)
}

func validateStartsWith(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	for _, prefix := range params {
		if strings.HasPrefix(str, prefix) {
			return true
		}
	}
	return false
}

func validateEndsWith(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	for _, suffix := range params {
		if strings.HasSuffix(str, suffix) {
			return true
		}
	}
	return false
}

func validateContains(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	str, ok := value.(string)
	if !ok {
		return false
	}
	return strings.Contains(str, params[0])
}

func validateDate(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"01/02/2006",
	}
	for _, format := range formats {
		if _, err := time.Parse(format, str); err == nil {
			return true
		}
	}
	return false
}

func validateAfter(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	t1, ok := parseDateTime(value)
	if !ok {
		return false
	}

	var t2 time.Time
	if fieldValue, exists := data[params[0]]; exists {
		t2, ok = parseDateTime(fieldValue)
		if !ok {
			return false
		}
	} else {
		t2, ok = parseDateTime(params[0])
		if !ok {
			return false
		}
	}

	return t1.After(t2)
}

func validateBefore(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	t1, ok := parseDateTime(value)
	if !ok {
		return false
	}

	var t2 time.Time
	if fieldValue, exists := data[params[0]]; exists {
		t2, ok = parseDateTime(fieldValue)
		if !ok {
			return false
		}
	} else {
		t2, ok = parseDateTime(params[0])
		if !ok {
			return false
		}
	}

	return t1.Before(t2)
}

func validateConfirmed(attribute string, value any, params []string, data map[string]any) bool {
	confirmField := attribute + "_confirmation"
	confirmValue, ok := data[confirmField]
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", confirmValue)
}

func validateSame(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return false
	}
	return fmt.Sprintf("%v", value) == fmt.Sprintf("%v", otherValue)
}

func validateDifferent(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return true
	}
	return fmt.Sprintf("%v", value) != fmt.Sprintf("%v", otherValue)
}

func validateGt(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return false
	}
	v1, ok1 := toFloat64(value)
	v2, ok2 := toFloat64(otherValue)
	if !ok1 || !ok2 {
		return false
	}
	return v1 > v2
}

func validateGte(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return false
	}
	v1, ok1 := toFloat64(value)
	v2, ok2 := toFloat64(otherValue)
	if !ok1 || !ok2 {
		return false
	}
	return v1 >= v2
}

func validateLt(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return false
	}
	v1, ok1 := toFloat64(value)
	v2, ok2 := toFloat64(otherValue)
	if !ok1 || !ok2 {
		return false
	}
	return v1 < v2
}

func validateLte(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	otherValue, ok := data[params[0]]
	if !ok {
		return false
	}
	v1, ok1 := toFloat64(value)
	v2, ok2 := toFloat64(otherValue)
	if !ok1 || !ok2 {
		return false
	}
	return v1 <= v2
}

func validateJSON(attribute string, value any, params []string, data map[string]any) bool {
	str, ok := value.(string)
	if !ok {
		return false
	}
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

func validateUnique(attribute string, value any, params []string, data map[string]any) bool {
	// Placeholder: actual database check should be implemented by the application
	return true
}

func validateExists(attribute string, value any, params []string, data map[string]any) bool {
	// Placeholder: actual database check should be implemented by the application
	return true
}

func validateDigits(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 1 {
		return true
	}
	digits, err := strconv.Atoi(params[0])
	if err != nil {
		return false
	}
	str := fmt.Sprintf("%v", value)
	digitsRegex := regexp.MustCompile(`^\d+$`)
	return digitsRegex.MatchString(str) && len(str) == digits
}

func validateDigitsBetween(attribute string, value any, params []string, data map[string]any) bool {
	if len(params) < 2 {
		return true
	}
	min, err1 := strconv.Atoi(params[0])
	max, err2 := strconv.Atoi(params[1])
	if err1 != nil || err2 != nil {
		return false
	}
	str := fmt.Sprintf("%v", value)
	digitsRegex := regexp.MustCompile(`^\d+$`)
	return digitsRegex.MatchString(str) && len(str) >= min && len(str) <= max
}

// Helper functions

func isEmpty(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Slice, reflect.Map, reflect.Array:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	}
	return false
}

func toFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case string:
		f, err := strconv.ParseFloat(v, 64)
		return f, err == nil
	}
	return 0, false
}

func parseDateTime(value any) (time.Time, bool) {
	str, ok := value.(string)
	if !ok {
		return time.Time{}, false
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02 15:04:05",
		"02/01/2006",
		"01/02/2006",
	}
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
