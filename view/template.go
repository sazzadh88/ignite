package view

import (
	"bytes"
	"html/template"
)

// Template represents a compiled view template.
type Template struct {
	name       string
	tmpl       *template.Template
	layoutName string
	engine     *Engine
}

// Execute renders the template with the given data and returns the output as a string.
func (t *Template) Execute(data map[string]any) (string, error) {
	// Merge shared data with view-specific data
	mergedData := make(map[string]any)
	for k, v := range t.engine.sharedData {
		mergedData[k] = v
	}
	for k, v := range data {
		mergedData[k] = v
	}

	// Execute view composers
	t.engine.executeComposers(t.name, mergedData)

	var buf bytes.Buffer

	// If this template extends a layout, execute the layout
	if t.layoutName != "" {
		// Execute the child template first to define sections
		if err := t.tmpl.ExecuteTemplate(&buf, t.name, mergedData); err != nil {
			return "", err
		}
		// Clear buffer as we just needed to define the sections
		buf.Reset()

		// Now execute the layout
		if err := t.tmpl.ExecuteTemplate(&buf, t.layoutName, mergedData); err != nil {
			return "", err
		}
	} else {
		// Execute the template directly
		if err := t.tmpl.ExecuteTemplate(&buf, t.name, mergedData); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
