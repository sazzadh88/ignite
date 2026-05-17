// Package view provides a Blade-inspired template engine for Ignite.
//
// The view package implements a powerful template engine with Laravel Blade-like
// syntax for Go applications. It compiles templates with directives into Go's
// html/template format and provides features like layout inheritance, partials,
// and view composers.
//
// # Basic Usage
//
// Create a new engine and render a template:
//
//	engine := view.NewEngine("resources/views")
//	output, err := engine.Render("welcome", map[string]any{
//	    "Name": "World",
//	})
//
// # Template Syntax
//
// Variables (HTML-escaped):
//
//	{{ .Name }}
//
// Raw output (unescaped):
//
//	{!! .HTML !!}
//
// Conditionals:
//
//	@if(.LoggedIn)
//	    Welcome back!
//	@else
//	    Please login
//	@endif
//
//	@unless(.IsGuest)
//	    You have access
//	@endunless
//
// Loops:
//
//	@foreach(.Users, "user")
//	    {{ $user.Name }}
//	@endforeach
//
//	@forelse(.Items, "item")
//	    {{ $item }}
//	@empty
//	    No items found
//	@endforelse
//
// Layout Inheritance:
//
//	@extends("layouts.app")
//
//	@section("content")
//	    <h1>Page Content</h1>
//	@endsection
//
// In layout file:
//
//	<html>
//	    @yield("content")
//	</html>
//
// Including Partials:
//
//	@include("partials.header")
//	@includeIf("partials.sidebar")
//
// Form Helpers:
//
//	<form>
//	    @csrf
//	    @method("PUT")
//	</form>
//
// JSON:
//
//	<script>
//	    var data = @json(.Data);
//	</script>
//
// Comments (stripped from output):
//
//	{{-- This will not appear in output --}}
//
// # Shared Data
//
// Share data across all views:
//
//	engine.Share("AppName", "Ignite")
//	engine.Share("Version", "1.0")
//
// # View Composers
//
// Register callbacks that run before a view is rendered:
//
//	engine.Composer("dashboard", func(data map[string]any) {
//	    data["Stats"] = fetchStats()
//	})
//
// # Helper Functions
//
// Templates have access to these helper functions:
//
//	{{ upper .Text }}       - Convert to uppercase
//	{{ lower .Text }}       - Convert to lowercase
//	{{ title .Text }}       - Title case
//	{{ truncate .Text 50 }} - Truncate string
//	{{ nl2br .Text }}       - Convert newlines to <br>
//	{{ json .Data }}        - JSON encode
//	{{ date "2006-01-02" .Time }} - Format date
//
// # File Organization
//
// Views are stored in files with .ignite.html extension.
// View names use dot notation that maps to directory structure:
//
//	"welcome"           -> welcome.ignite.html
//	"users.profile"     -> users/profile.ignite.html
//	"layouts.app"       -> layouts/app.ignite.html
package view
