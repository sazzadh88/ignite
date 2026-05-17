package view

import (
	"strings"
	"testing"
)

func TestCompilerInclude(t *testing.T) {
	comp := &compiler{}
	source := `<main>@include("partial")</main>`
	compiled, err := comp.compile(source)
	if err != nil {
		t.Fatalf("compile failed: %v", err)
	}

	if !strings.Contains(compiled, "__BLADE_INCLUDE__partial__") {
		t.Errorf("expected marker with name, got: %s", compiled)
	}
}

func TestResolveIncludes(t *testing.T) {
	source := `<main>__BLADE_INCLUDE__partial__</main>`
	resolved := resolveIncludes(source)

	if !strings.Contains(resolved, `{{ template "partial" . }}`) {
		t.Errorf("expected template invocation, got: %s", resolved)
	}
}
