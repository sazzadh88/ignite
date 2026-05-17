package view

import (
	"fmt"
	"regexp"
)

// compiler handles the compilation of Blade-like directives into Go template syntax.
type compiler struct{}

// compile transforms Blade-like syntax into Go template syntax.
func (c *compiler) compile(source string) (string, error) {
	// Strip Blade comments first
	source = c.compileComments(source)

	// Compile directives in order
	source = c.compileExtends(source)
	source = c.compileSections(source)
	source = c.compileYields(source)
	source = c.compileIncludes(source)
	source = c.compileLoops(source)
	source = c.compileConditionals(source)
	source = c.compileAuth(source)
	source = c.compileHelpers(source)
	source = c.compileEchos(source)

	return source, nil
}

// compileComments removes Blade comments {{-- comment --}}
func (c *compiler) compileComments(source string) string {
	re := regexp.MustCompile(`(?s)\{\{--.*?--\}\}`)
	return re.ReplaceAllString(source, "")
}

// compileEchos compiles {{ variable }} and {!! raw !!}
func (c *compiler) compileEchos(source string) string {
	// Raw output {!! expr !!}
	rawRe := regexp.MustCompile(`\{!!\s*(.+?)\s*!!\}`)
	source = rawRe.ReplaceAllString(source, `{{ raw $1 }}`)

	// Escaped output {{ expr }}
	echoRe := regexp.MustCompile(`\{\{\s*(.+?)\s*\}\}`)
	source = echoRe.ReplaceAllString(source, `{{ $1 }}`)

	return source
}

// compileConditionals compiles @if, @elseif, @else, @endif, @unless, @endunless, @isset, @empty
func (c *compiler) compileConditionals(source string) string {
	// @if(condition)
	source = regexp.MustCompile(`@if\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ if $1 }}`)

	// @elseif(condition)
	source = regexp.MustCompile(`@elseif\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ else if $1 }}`)

	// @else
	source = regexp.MustCompile(`@else\b`).ReplaceAllString(source, `{{ else }}`)

	// @endif
	source = regexp.MustCompile(`@endif\b`).ReplaceAllString(source, `{{ end }}`)

	// @unless(condition) -> if not condition
	source = regexp.MustCompile(`@unless\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ if not $1 }}`)

	// @endunless
	source = regexp.MustCompile(`@endunless\b`).ReplaceAllString(source, `{{ end }}`)

	// @isset(var)
	source = regexp.MustCompile(`@isset\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ if $1 }}`)

	// @endisset
	source = regexp.MustCompile(`@endisset\b`).ReplaceAllString(source, `{{ end }}`)

	// @empty(var)
	source = regexp.MustCompile(`@empty\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ if not $1 }}`)

	// @endempty
	source = regexp.MustCompile(`@endempty\b`).ReplaceAllString(source, `{{ end }}`)

	// @switch(var)
	source = regexp.MustCompile(`@switch\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ $__switch := $1 }}`)

	// @case(val)
	source = regexp.MustCompile(`@case\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ if eq $__switch $1 }}`)

	// @default
	source = regexp.MustCompile(`@default\b`).ReplaceAllString(source, `{{ else }}`)

	// @endswitch
	source = regexp.MustCompile(`@endswitch\b`).ReplaceAllString(source, `{{ end }}`)

	return source
}

// compileLoops compiles @for, @foreach, @forelse, @while
func (c *compiler) compileLoops(source string) string {
	// @foreach(items, "item") - simplified syntax
	source = regexp.MustCompile(`@foreach\s*\(\s*(.+?)\s*,\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ range $$$2 := $1 }}`)

	// @endforeach
	source = regexp.MustCompile(`@endforeach\b`).ReplaceAllString(source, `{{ end }}`)

	// @forelse(items, "item")
	source = regexp.MustCompile(`@forelse\s*\(\s*(.+?)\s*,\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ if $1 }}{{ range $$$2 := $1 }}`)

	// @empty
	source = regexp.MustCompile(`@empty\b`).ReplaceAllString(source, `{{ end }}{{ else }}`)

	// @endforelse
	source = regexp.MustCompile(`@endforelse\b`).ReplaceAllString(source, `{{ end }}`)

	// @for(...) - pass through to Go template
	source = regexp.MustCompile(`@for\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ $1 }}`)

	// @endfor
	source = regexp.MustCompile(`@endfor\b`).ReplaceAllString(source, `{{ end }}`)

	// @while(condition)
	source = regexp.MustCompile(`@while\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ range $1 }}`)

	// @endwhile
	source = regexp.MustCompile(`@endwhile\b`).ReplaceAllString(source, `{{ end }}`)

	return source
}

// compileExtends extracts @extends directive
func (c *compiler) compileExtends(source string) string {
	// We'll use a marker for later processing
	source = regexp.MustCompile(`@extends\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `__BLADE_EXTENDS__${1}__`)
	return source
}

// compileSections compiles @section and @endsection
func (c *compiler) compileSections(source string) string {
	// @section("name")
	source = regexp.MustCompile(`@section\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ define "$1" }}`)

	// @endsection
	source = regexp.MustCompile(`@endsection\b`).ReplaceAllString(source, `{{ end }}`)

	return source
}

// compileYields compiles @yield
func (c *compiler) compileYields(source string) string {
	// @yield("name", "default")
	source = regexp.MustCompile(`@yield\s*\(\s*"(.+?)"\s*,\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ template "$1" . }}`)

	// @yield("name")
	source = regexp.MustCompile(`@yield\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ template "$1" . }}`)

	return source
}

// compileIncludes compiles @include, @includeIf, @includeWhen
func (c *compiler) compileIncludes(source string) string {
	// @include("partial", data) - simplified: just pass current context
	source = regexp.MustCompile(`@include\s*\(\s*"(.+?)"\s*,\s*.+?\)`).ReplaceAllString(source, `__BLADE_INCLUDE__${1}__`)

	// @include("partial")
	source = regexp.MustCompile(`@include\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `__BLADE_INCLUDE__${1}__`)

	// @includeIf("partial")
	source = regexp.MustCompile(`@includeIf\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `__BLADE_INCLUDEIF__${1}__`)

	// @includeWhen(condition, "view")
	source = regexp.MustCompile(`@includeWhen\s*\(\s*(.+?)\s*,\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ if ${1} }}__BLADE_INCLUDE__${2}__{{ end }}`)

	return source
}

// compileAuth compiles @auth, @guest
func (c *compiler) compileAuth(source string) string {
	// @auth
	source = regexp.MustCompile(`@auth\b`).ReplaceAllString(source, `{{ if .Auth }}`)

	// @endauth
	source = regexp.MustCompile(`@endauth\b`).ReplaceAllString(source, `{{ end }}`)

	// @guest
	source = regexp.MustCompile(`@guest\b`).ReplaceAllString(source, `{{ if not .Auth }}`)

	// @endguest
	source = regexp.MustCompile(`@endguest\b`).ReplaceAllString(source, `{{ end }}`)

	return source
}

// compileHelpers compiles @csrf, @method, @json, @push, @stack, @once, @error
func (c *compiler) compileHelpers(source string) string {
	// @csrf
	source = regexp.MustCompile(`@csrf\b`).ReplaceAllString(source, `<input type="hidden" name="_token" value="{{ .CSRFToken }}">`)

	// @method("PUT")
	source = regexp.MustCompile(`@method\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `<input type="hidden" name="_method" value="$1">`)

	// @json(data)
	source = regexp.MustCompile(`@json\s*\(\s*(.+?)\s*\)`).ReplaceAllString(source, `{{ json $1 }}`)

	// @push("stack") - simplified: define template
	source = regexp.MustCompile(`@push\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ define "stack_$1" }}`)

	// @endpush
	source = regexp.MustCompile(`@endpush\b`).ReplaceAllString(source, `{{ end }}`)

	// @stack("name")
	source = regexp.MustCompile(`@stack\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ template "stack_$1" . }}`)

	// @once
	source = regexp.MustCompile(`@once\b`).ReplaceAllString(source, `{{ if not .__once }}{{ $.__once = true }}`)

	// @endonce
	source = regexp.MustCompile(`@endonce\b`).ReplaceAllString(source, `{{ end }}`)

	// @error("field")
	source = regexp.MustCompile(`@error\s*\(\s*"(.+?)"\s*\)`).ReplaceAllString(source, `{{ if index .Errors "$1" }}`)

	// @enderror
	source = regexp.MustCompile(`@enderror\b`).ReplaceAllString(source, `{{ end }}`)

	return source
}

// extractExtends extracts the extends directive from compiled source
func extractExtends(source string) (layout string, hasExtends bool) {
	re := regexp.MustCompile(`__BLADE_EXTENDS__(.+?)__`)
	matches := re.FindStringSubmatch(source)
	if len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// extractIncludes extracts include directives and returns their names
func extractIncludes(source string) []string {
	includes := []string{}

	// Regular includes
	re := regexp.MustCompile(`__BLADE_INCLUDE__(.+?)__`)
	matches := re.FindAllStringSubmatch(source, -1)
	for _, match := range matches {
		if len(match) > 1 {
			includes = append(includes, match[1])
		}
	}

	// Conditional includes
	reIf := regexp.MustCompile(`__BLADE_INCLUDEIF__(.+?)__`)
	matchesIf := reIf.FindAllStringSubmatch(source, -1)
	for _, match := range matchesIf {
		if len(match) > 1 {
			includes = append(includes, match[1])
		}
	}

	return includes
}

// resolveIncludes replaces include markers with template invocations
func resolveIncludes(source string) string {
	// Regular includes - use capturing group in template invocation
	source = regexp.MustCompile(`__BLADE_INCLUDE__(.+?)__`).ReplaceAllString(source, `{{ template "$1" . }}`)

	// Conditional includes
	source = regexp.MustCompile(`__BLADE_INCLUDEIF__(.+?)__`).ReplaceAllString(source, `{{ template "$1" . }}`)

	return source
}

// removeExtendsMark removes the extends marker from source
func removeExtendsMark(source string) string {
	return regexp.MustCompile(`__BLADE_EXTENDS__.+?__`).ReplaceAllString(source, "")
}

// removeMissingInclude removes include/includeIf markers for a specific template name
func removeMissingInclude(source, name string) string {
	// Remove both regular and conditional include markers
	source = regexp.MustCompile(`__BLADE_INCLUDE__`+regexp.QuoteMeta(name)+`__`).ReplaceAllString(source, "")
	source = regexp.MustCompile(`__BLADE_INCLUDEIF__`+regexp.QuoteMeta(name)+`__`).ReplaceAllString(source, "")
	return source
}

// compileStandalone compiles a standalone template (no layout inheritance)
func compileStandalone(name, source string) (string, error) {
	comp := &compiler{}
	compiled, err := comp.compile(source)
	if err != nil {
		return "", err
	}

	// Wrap in a define block with the template name
	return fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, name, compiled), nil
}

// compileWithLayout compiles a template that extends a layout
func compileWithLayout(name, source, layoutName string) (string, error) {
	comp := &compiler{}
	compiled, err := comp.compile(source)
	if err != nil {
		return "", err
	}

	// Remove the extends marker
	compiled = removeExtendsMark(compiled)

	// Wrap in define block
	return fmt.Sprintf(`{{ define "%s" }}%s{{ end }}`, name, compiled), nil
}
