package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func setupTestDir(t *testing.T) (string, func()) {
	t.Helper()
	dir := t.TempDir()

	viewsDir := filepath.Join(dir, "views")
	if err := os.MkdirAll(viewsDir, 0755); err != nil {
		t.Fatalf("failed to create views directory: %v", err)
	}

	cleanup := func() {
		os.RemoveAll(dir)
	}

	return viewsDir, cleanup
}

func writeView(t *testing.T, dir, name, content string) {
	t.Helper()
	filePath := filepath.Join(dir, strings.ReplaceAll(name, ".", string(filepath.Separator))+".ignite.html")

	dirPath := filepath.Dir(filePath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("failed to create directory %s: %v", dirPath, err)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write view %s: %v", name, err)
	}
}

func TestBasicVariableRendering(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "hello", `Hello {{ .Name }}!`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("hello", map[string]any{
		"Name": "World",
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	expected := "Hello World!"
	if !strings.Contains(output, expected) {
		t.Errorf("expected output to contain %q, got %q", expected, output)
	}
}

func TestRawOutput(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "raw", `{!! .HTML !!}`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("raw", map[string]any{
		"HTML": "<strong>Bold</strong>",
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "<strong>Bold</strong>") {
		t.Errorf("expected raw HTML output, got %q", output)
	}
}

func TestIfDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "conditional", `@if(.Show)Visible@endif`)

	engine := NewEngine(viewsDir)

	// Test true condition
	output, err := engine.Render("conditional", map[string]any{
		"Show": true,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Visible") {
		t.Errorf("expected 'Visible' in output, got %q", output)
	}

	// Test false condition
	output, err = engine.Render("conditional", map[string]any{
		"Show": false,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if strings.Contains(output, "Visible") {
		t.Errorf("expected no 'Visible' in output, got %q", output)
	}
}

func TestIfElseDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "ifelse", `@if(.LoggedIn)Welcome@else Login@endif`)

	engine := NewEngine(viewsDir)

	output, err := engine.Render("ifelse", map[string]any{
		"LoggedIn": true,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Welcome") {
		t.Errorf("expected 'Welcome' in output, got %q", output)
	}

	output, err = engine.Render("ifelse", map[string]any{
		"LoggedIn": false,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Login") {
		t.Errorf("expected 'Login' in output, got %q", output)
	}
}

func TestForeachDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "list", `@foreach(.Items, "item"){{ $item }}@endforeach`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("list", map[string]any{
		"Items": []string{"A", "B", "C"},
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "A") || !strings.Contains(output, "B") || !strings.Contains(output, "C") {
		t.Errorf("expected items A, B, C in output, got %q", output)
	}
}

func TestForelseWithEmptyCollection(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "forelse", `@forelse(.Items, "item"){{ $item }}@empty No items @endforelse`)

	engine := NewEngine(viewsDir)

	// Test with items
	output, err := engine.Render("forelse", map[string]any{
		"Items": []string{"Item1"},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Item1") {
		t.Errorf("expected 'Item1' in output, got %q", output)
	}

	// Test with empty collection
	output, err = engine.Render("forelse", map[string]any{
		"Items": []string{},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "No items") {
		t.Errorf("expected 'No items' in output, got %q", output)
	}
}

func TestLayoutInheritance(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create layout
	writeView(t, viewsDir, "layout", `<html>@yield("content")</html>`)

	// Create child view
	writeView(t, viewsDir, "page", `@extends("layout")
@section("content")
<body>Page content</body>
@endsection`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("page", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "<html>") || !strings.Contains(output, "Page content") || !strings.Contains(output, "</html>") {
		t.Errorf("expected layout with content, got %q", output)
	}
}

func TestIncludePartial(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "partial", `<div>Partial content</div>`)
	writeView(t, viewsDir, "main", `<main>@include("partial")</main>`)

	engine := NewEngine(viewsDir)

	// Debug: Check if partial exists
	if !engine.Exists("partial") {
		t.Fatal("partial view does not exist")
	}

	output, err := engine.Render("main", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "Partial content") {
		t.Errorf("expected partial content in output, got %q", output)
	}
}

func TestCSRFDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "form", `<form>@csrf</form>`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("form", map[string]any{
		"CSRFToken": "test-token-123",
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, `name="_token"`) || !strings.Contains(output, "test-token-123") {
		t.Errorf("expected CSRF input field, got %q", output)
	}
}

func TestMethodDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "form", `<form>@method("PUT")</form>`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("form", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, `name="_method"`) || !strings.Contains(output, "PUT") {
		t.Errorf("expected method input field, got %q", output)
	}
}

func TestJSONDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "json", `<script>var data = @json(.Data);</script>`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("json", map[string]any{
		"Data": map[string]any{"key": "value"},
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("expected JSON output, got %q", output)
	}
}

func TestCommentsStripped(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "comments", `Before{{-- This is a comment --}}After`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("comments", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if strings.Contains(output, "This is a comment") {
		t.Errorf("expected comment to be stripped, got %q", output)
	}

	if !strings.Contains(output, "Before") || !strings.Contains(output, "After") {
		t.Errorf("expected 'Before' and 'After' in output, got %q", output)
	}
}

func TestSharedData(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "shared", `App: {{ .AppName }}`)

	engine := NewEngine(viewsDir)
	engine.Share("AppName", "Ignite")

	output, err := engine.Render("shared", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "Ignite") {
		t.Errorf("expected shared data in output, got %q", output)
	}
}

func TestViewComposer(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "composed", `Count: {{ .Count }}`)

	engine := NewEngine(viewsDir)

	composerCalled := false
	engine.Composer("composed", func(data map[string]any) {
		composerCalled = true
		data["Count"] = 42
	})

	output, err := engine.Render("composed", map[string]any{})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !composerCalled {
		t.Error("expected composer to be called")
	}

	if !strings.Contains(output, "42") {
		t.Errorf("expected composer data in output, got %q", output)
	}
}

func TestExists(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "exists", `Content`)

	engine := NewEngine(viewsDir)

	if !engine.Exists("exists") {
		t.Error("expected view to exist")
	}

	if engine.Exists("nonexistent") {
		t.Error("expected view to not exist")
	}
}

func TestAddPath(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create a second views directory
	altDir := filepath.Join(filepath.Dir(viewsDir), "alt_views")
	if err := os.MkdirAll(altDir, 0755); err != nil {
		t.Fatalf("failed to create alt views directory: %v", err)
	}

	writeView(t, altDir, "alt", `Alt content`)

	engine := NewEngine(viewsDir)
	engine.AddPath(altDir)

	output, err := engine.Render("alt", nil)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "Alt content") {
		t.Errorf("expected alt view content, got %q", output)
	}
}

func TestUnlessDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "unless", `@unless(.IsAdmin)Not admin@endunless`)

	engine := NewEngine(viewsDir)

	output, err := engine.Render("unless", map[string]any{
		"IsAdmin": false,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Not admin") {
		t.Errorf("expected 'Not admin' in output, got %q", output)
	}

	output, err = engine.Render("unless", map[string]any{
		"IsAdmin": true,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if strings.Contains(output, "Not admin") {
		t.Errorf("expected no 'Not admin' in output, got %q", output)
	}
}

func TestHelperFunctions(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "helpers", `{{ upper .Text }}`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("helpers", map[string]any{
		"Text": "hello",
	})

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "HELLO") {
		t.Errorf("expected uppercase output, got %q", output)
	}
}

func TestAuthDirective(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "auth", `@auth Authenticated @endauth @guest Guest @endguest`)

	engine := NewEngine(viewsDir)

	output, err := engine.Render("auth", map[string]any{
		"Auth": true,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Authenticated") {
		t.Errorf("expected 'Authenticated' in output, got %q", output)
	}

	output, err = engine.Render("auth", map[string]any{
		"Auth": false,
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	if !strings.Contains(output, "Guest") {
		t.Errorf("expected 'Guest' in output, got %q", output)
	}
}

func TestComplexLayoutWithMultipleSections(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	// Create layout with multiple yields
	writeView(t, viewsDir, "app", `<html>
<head>@yield("title")</head>
<body>@yield("content")</body>
</html>`)

	// Create child view with multiple sections
	writeView(t, viewsDir, "home", `@extends("app")
@section("title")
<title>Home Page</title>
@endsection
@section("content")
<main>Welcome</main>
@endsection`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("home", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "<title>Home Page</title>") {
		t.Errorf("expected title section, got %q", output)
	}
	if !strings.Contains(output, "<main>Welcome</main>") {
		t.Errorf("expected content section, got %q", output)
	}
}

func TestNestedIncludes(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "inner", `<span>Inner</span>`)
	writeView(t, viewsDir, "outer", `<div>@include("inner")</div>`)
	writeView(t, viewsDir, "main", `<main>@include("outer")</main>`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("main", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "<span>Inner</span>") {
		t.Errorf("expected nested include content, got %q", output)
	}
}

func TestIncludeIf(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "optional", `Optional`)
	writeView(t, viewsDir, "main", `Main @includeIf("optional")`)

	engine := NewEngine(viewsDir)
	output, err := engine.Render("main", nil)

	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(output, "Optional") {
		t.Errorf("expected optional include, got %q", output)
	}

	// Test with non-existent include (should not error)
	writeView(t, viewsDir, "main2", `Main @includeIf("nonexistent")`)
	output, err = engine.Render("main2", nil)

	if err != nil {
		t.Fatalf("render with missing includeIf should not fail: %v", err)
	}
}

func TestCaching(t *testing.T) {
	viewsDir, cleanup := setupTestDir(t)
	defer cleanup()

	writeView(t, viewsDir, "cached", `Content`)

	engine := NewEngine(viewsDir)

	// First render
	_, err := engine.Render("cached", nil)
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	// Check cache
	engine.cacheMutex.RLock()
	_, cached := engine.cache["cached"]
	engine.cacheMutex.RUnlock()

	if !cached {
		t.Error("expected view to be cached")
	}

	// Second render should use cache
	_, err = engine.Render("cached", nil)
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}
}
