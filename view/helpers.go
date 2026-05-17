package view

import (
	"encoding/json"
	"html/template"
	"strings"
	"time"
)

// templateFuncs returns the default template helper functions available in all views.
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"upper":    strings.ToUpper,
		"lower":    strings.ToLower,
		"title":    strings.Title,
		"truncate": truncate,
		"nl2br":    nl2br,
		"raw":      raw,
		"json":     jsonEncode,
		"date":     formatDate,
		"default":  defaultValue,
	}
}

// truncate truncates a string to the specified length and appends "..." if truncated.
func truncate(s string, length int) string {
	if len(s) <= length {
		return s
	}
	if length <= 3 {
		return s[:length]
	}
	return s[:length-3] + "..."
}

// nl2br converts newlines to <br> tags.
func nl2br(s string) template.HTML {
	return template.HTML(strings.ReplaceAll(template.HTMLEscapeString(s), "\n", "<br>"))
}

// raw returns unescaped HTML (use with caution).
func raw(s string) template.HTML {
	return template.HTML(s)
}

// jsonEncode encodes data as JSON.
func jsonEncode(v any) (template.JS, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return template.JS(b), nil
}

// formatDate formats a time value with the given layout.
// If the value is not a time.Time, returns empty string.
func formatDate(layout string, t any) string {
	switch v := t.(type) {
	case time.Time:
		return v.Format(layout)
	case *time.Time:
		if v != nil {
			return v.Format(layout)
		}
	}
	return ""
}

// defaultValue returns the value if non-nil, otherwise returns the default.
func defaultValue(defaultVal, val any) any {
	if val == nil {
		return defaultVal
	}
	return val
}
