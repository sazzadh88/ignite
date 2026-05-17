package view

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Engine is the view template engine that manages template compilation and rendering.
type Engine struct {
	paths       []string
	cache       map[string]*Template
	cacheMutex  sync.RWMutex
	sharedData  map[string]any
	dataMutex   sync.RWMutex
	composers   map[string][]func(map[string]any)
	composerMux sync.RWMutex
}

// View is the default package-level Engine instance.
var View = NewEngine("resources/views")

// NewEngine creates a new view engine with the specified view path.
func NewEngine(viewPath string) *Engine {
	return &Engine{
		paths:      []string{viewPath},
		cache:      make(map[string]*Template),
		sharedData: make(map[string]any),
		composers:  make(map[string][]func(map[string]any)),
	}
}

// AddPath adds an additional view search path.
func (e *Engine) AddPath(path string) {
	e.paths = append(e.paths, path)
}

// Share adds shared data that will be available to all views.
func (e *Engine) Share(key string, value any) {
	e.dataMutex.Lock()
	defer e.dataMutex.Unlock()
	e.sharedData[key] = value
}

// Composer registers a view composer callback that runs before the view is rendered.
// The composer receives the view data and can modify it.
func (e *Engine) Composer(name string, fn func(data map[string]any)) {
	e.composerMux.Lock()
	defer e.composerMux.Unlock()
	e.composers[name] = append(e.composers[name], fn)
}

// executeComposers runs all registered composers for a view.
func (e *Engine) executeComposers(name string, data map[string]any) {
	e.composerMux.RLock()
	composers := e.composers[name]
	e.composerMux.RUnlock()

	for _, composer := range composers {
		composer(data)
	}
}

// Exists checks if a view template exists.
func (e *Engine) Exists(name string) bool {
	_, err := e.findView(name)
	return err == nil
}

// Render renders a view template with the given data and returns the output as a string.
func (e *Engine) Render(name string, data map[string]any) (string, error) {
	if data == nil {
		data = make(map[string]any)
	}

	// Check cache first
	e.cacheMutex.RLock()
	tmpl, cached := e.cache[name]
	e.cacheMutex.RUnlock()

	if cached {
		return tmpl.Execute(data)
	}

	// Compile and cache the template
	tmpl, err := e.compile(name)
	if err != nil {
		return "", err
	}

	e.cacheMutex.Lock()
	e.cache[name] = tmpl
	e.cacheMutex.Unlock()

	return tmpl.Execute(data)
}

// compile compiles a view template and all its dependencies.
func (e *Engine) compile(name string) (*Template, error) {
	viewPath, err := e.findView(name)
	if err != nil {
		return nil, err
	}

	// Read the view source
	source, err := os.ReadFile(viewPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read view %s: %w", name, err)
	}

	// Parse directives to check for extends/includes
	comp := &compiler{}
	compiled, err := comp.compile(string(source))
	if err != nil {
		return nil, fmt.Errorf("failed to compile view %s: %w", name, err)
	}

	// Check if this view extends a layout
	layoutName, hasExtends := extractExtends(compiled)

	// Create a new template with helper functions
	tmpl := template.New(name).Funcs(templateFuncs())

	// If it extends a layout, compile the layout first
	if hasExtends {
		if err := e.compileLayout(tmpl, layoutName); err != nil {
			return nil, err
		}
	}

	// Compile any includes
	includes := extractIncludes(compiled)
	missingIncludes := make(map[string]bool)
	for _, includeName := range includes {
		if !e.Exists(includeName) {
			missingIncludes[includeName] = true
			continue
		}
		if err := e.compileInclude(tmpl, includeName); err != nil {
			return nil, err
		}
	}

	// Remove markers for missing includeIf templates
	for name := range missingIncludes {
		compiled = removeMissingInclude(compiled, name)
	}

	// Resolve include markers
	compiled = resolveIncludes(compiled)

	// Remove extends marker
	compiled = removeExtendsMark(compiled)

	// Parse directly - sections will define themselves
	// Only wrap in define if there are no sections
	if !strings.Contains(compiled, "{{ define") {
		compiled = fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, name, compiled)
	} else {
		// If sections exist, wrap everything but keep sections as siblings
		// We need a root define for the main template
		compiled = fmt.Sprintf(`{{ define "%s" }}{{ end }}`, name) + compiled
	}

	// Parse the compiled template
	if _, err := tmpl.Parse(compiled); err != nil {
		return nil, fmt.Errorf("failed to parse compiled view %s: %w", name, err)
	}

	return &Template{
		name:       name,
		tmpl:       tmpl,
		layoutName: layoutName,
		engine:     e,
	}, nil
}

// compileLayout compiles a layout template.
func (e *Engine) compileLayout(tmpl *template.Template, layoutName string) error {
	layoutPath, err := e.findView(layoutName)
	if err != nil {
		return fmt.Errorf("layout not found: %s", layoutName)
	}

	source, err := os.ReadFile(layoutPath)
	if err != nil {
		return fmt.Errorf("failed to read layout %s: %w", layoutName, err)
	}

	comp := &compiler{}
	compiled, err := comp.compile(string(source))
	if err != nil {
		return fmt.Errorf("failed to compile layout %s: %w", layoutName, err)
	}

	// Check if layout has includes
	includes := extractIncludes(compiled)
	missingIncludes := make(map[string]bool)
	for _, includeName := range includes {
		if !e.Exists(includeName) {
			missingIncludes[includeName] = true
			continue
		}
		if err := e.compileInclude(tmpl, includeName); err != nil {
			return err
		}
	}

	// Remove markers for missing includes
	for name := range missingIncludes {
		compiled = removeMissingInclude(compiled, name)
	}

	// Resolve includes
	compiled = resolveIncludes(compiled)

	// Parse directly without double wrapping if sections exist
	if !strings.Contains(compiled, "{{ define") {
		compiled = fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, layoutName, compiled)
	} else {
		compiled = fmt.Sprintf(`{{ define "%s" }}{{ end }}`, layoutName) + compiled
	}

	if _, err := tmpl.Parse(compiled); err != nil {
		return fmt.Errorf("failed to parse layout %s: %w", layoutName, err)
	}

	return nil
}

// compileInclude compiles an included partial template.
func (e *Engine) compileInclude(tmpl *template.Template, includeName string) error {
	includePath, err := e.findView(includeName)
	if err != nil {
		return fmt.Errorf("include not found: %s", includeName)
	}

	source, err := os.ReadFile(includePath)
	if err != nil {
		return fmt.Errorf("failed to read include %s: %w", includeName, err)
	}

	comp := &compiler{}
	compiled, err := comp.compile(string(source))
	if err != nil {
		return fmt.Errorf("failed to compile include %s: %w", includeName, err)
	}

	// Recursively compile nested includes
	includes := extractIncludes(compiled)
	missingIncludes := make(map[string]bool)
	for _, nestedInclude := range includes {
		if !e.Exists(nestedInclude) {
			missingIncludes[nestedInclude] = true
			continue
		}
		if err := e.compileInclude(tmpl, nestedInclude); err != nil {
			return err
		}
	}

	// Remove markers for missing includes
	for name := range missingIncludes {
		compiled = removeMissingInclude(compiled, name)
	}

	// Resolve includes
	compiled = resolveIncludes(compiled)

	// Wrap in define block
	finalSource := fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, includeName, compiled)

	if _, err := tmpl.Parse(finalSource); err != nil {
		return fmt.Errorf("failed to parse include %s: %w", includeName, err)
	}

	return nil
}

// findView searches for a view file in the configured paths.
func (e *Engine) findView(name string) (string, error) {
	// Normalize the name: replace dots with path separators and add .ignite.html extension
	fileName := strings.ReplaceAll(name, ".", string(filepath.Separator)) + ".ignite.html"

	for _, basePath := range e.paths {
		fullPath := filepath.Join(basePath, fileName)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("view not found: %s", name)
}
