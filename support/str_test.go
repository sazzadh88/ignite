package support

import (
	"strings"
	"testing"
)

func TestAfter(t *testing.T) {
	tests := []struct {
		subject, search, want string
	}{
		{"hello world", "hello ", "world"},
		{"hello world", "o ", "world"},
		{"hello world", "xyz", "hello world"},
		{"hello world", "", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.After(tt.subject, tt.search); got != tt.want {
			t.Errorf("After(%q, %q) = %q, want %q", tt.subject, tt.search, got, tt.want)
		}
	}
}

func TestAfterLast(t *testing.T) {
	tests := []struct {
		subject, search, want string
	}{
		{"hello world world", "world ", "world"},
		{"hello world world", "o ", "world world"},
		{"hello world", "xyz", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.AfterLast(tt.subject, tt.search); got != tt.want {
			t.Errorf("AfterLast(%q, %q) = %q, want %q", tt.subject, tt.search, got, tt.want)
		}
	}
}

func TestBefore(t *testing.T) {
	tests := []struct {
		subject, search, want string
	}{
		{"hello world", " world", "hello"},
		{"hello world", "o", "hell"},
		{"hello world", "xyz", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.Before(tt.subject, tt.search); got != tt.want {
			t.Errorf("Before(%q, %q) = %q, want %q", tt.subject, tt.search, got, tt.want)
		}
	}
}

func TestBeforeLast(t *testing.T) {
	tests := []struct {
		subject, search, want string
	}{
		{"hello world world", " world", "hello world"},
		{"hello world", "o", "hello w"},
		{"hello world", "xyz", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.BeforeLast(tt.subject, tt.search); got != tt.want {
			t.Errorf("BeforeLast(%q, %q) = %q, want %q", tt.subject, tt.search, got, tt.want)
		}
	}
}

func TestBetween(t *testing.T) {
	tests := []struct {
		subject, from, to, want string
	}{
		{"hello [world] test", "[", "]", "world"},
		{"hello world test", "hello ", " test", "world"},
		{"hello world", "xyz", "abc", ""},
	}
	for _, tt := range tests {
		if got := Str.Between(tt.subject, tt.from, tt.to); got != tt.want {
			t.Errorf("Between(%q, %q, %q) = %q, want %q", tt.subject, tt.from, tt.to, got, tt.want)
		}
	}
}

func TestCamel(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"foo_bar", "fooBar"},
		{"FooBar", "foobar"},
		{"foo-bar", "fooBar"},
		{"foo bar", "fooBar"},
	}
	for _, tt := range tests {
		if got := Str.Camel(tt.input); got != tt.want {
			t.Errorf("Camel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSnake(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"FooBar", "foo_bar"},
		{"fooBar", "foo_bar"},
		{"foo-bar", "foo_bar"},
		{"foo bar", "foo_bar"},
		{"XMLHttpRequest", "xml_http_request"},
	}
	for _, tt := range tests {
		if got := Str.Snake(tt.input); got != tt.want {
			t.Errorf("Snake(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestKebab(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"FooBar", "foo-bar"},
		{"fooBar", "foo-bar"},
		{"foo_bar", "foo-bar"},
	}
	for _, tt := range tests {
		if got := Str.Kebab(tt.input); got != tt.want {
			t.Errorf("Kebab(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestStudly(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"foo_bar", "FooBar"},
		{"foo-bar", "FooBar"},
		{"foo bar", "FooBar"},
		{"fooBar", "Foobar"},
	}
	for _, tt := range tests {
		if got := Str.Studly(tt.input); got != tt.want {
			t.Errorf("Studly(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSlug(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Hello World", "hello-world"},
		{"hello_world", "hello-world"},
		{"Hello  World!", "hello-world"},
	}
	for _, tt := range tests {
		if got := Str.Slug(tt.input); got != tt.want {
			t.Errorf("Slug(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}

	if got := Str.Slug("Hello World", "_"); got != "hello_world" {
		t.Errorf("Slug with custom separator = %q, want %q", got, "hello_world")
	}
}

func TestUpper(t *testing.T) {
	if got := Str.Upper("hello"); got != "HELLO" {
		t.Errorf("Upper(hello) = %q, want HELLO", got)
	}
}

func TestLower(t *testing.T) {
	if got := Str.Lower("HELLO"); got != "hello" {
		t.Errorf("Lower(HELLO) = %q, want hello", got)
	}
}

func TestUcfirst(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"hello", "Hello"},
		{"Hello", "Hello"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Str.Ucfirst(tt.input); got != tt.want {
			t.Errorf("Ucfirst(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLcfirst(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"Hello", "hello"},
		{"hello", "hello"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := Str.Lcfirst(tt.input); got != tt.want {
			t.Errorf("Lcfirst(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"hello", 5},
		{"hello world", 11},
		{"", 0},
		{"日本語", 3},
	}
	for _, tt := range tests {
		if got := Str.Length(tt.input); got != tt.want {
			t.Errorf("Length(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestLimit(t *testing.T) {
	tests := []struct {
		input    string
		limit    int
		want     string
		hasEnding bool
		ending    string
	}{
		{"hello world", 5, "hello...", false, ""},
		{"hello world", 20, "hello world", false, ""},
		{"hello world", 5, "hello***", true, "***"},
	}
	for _, tt := range tests {
		var got string
		if tt.hasEnding {
			got = Str.Limit(tt.input, tt.limit, tt.ending)
		} else {
			got = Str.Limit(tt.input, tt.limit)
		}
		if got != tt.want {
			t.Errorf("Limit(%q, %d) = %q, want %q", tt.input, tt.limit, got, tt.want)
		}
	}
}

func TestWords(t *testing.T) {
	tests := []struct {
		input string
		words int
		want  string
	}{
		{"hello world test", 2, "hello world..."},
		{"hello world", 5, "hello world"},
		{"one two three", 1, "one..."},
	}
	for _, tt := range tests {
		if got := Str.Words(tt.input, tt.words); got != tt.want {
			t.Errorf("Words(%q, %d) = %q, want %q", tt.input, tt.words, got, tt.want)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		haystack, needle string
		want             bool
	}{
		{"hello world", "world", true},
		{"hello world", "xyz", false},
	}
	for _, tt := range tests {
		if got := Str.Contains(tt.haystack, tt.needle); got != tt.want {
			t.Errorf("Contains(%q, %q) = %v, want %v", tt.haystack, tt.needle, got, tt.want)
		}
	}
}

func TestContainsAll(t *testing.T) {
	tests := []struct {
		haystack string
		needles  []string
		want     bool
	}{
		{"hello world", []string{"hello", "world"}, true},
		{"hello world", []string{"hello", "xyz"}, false},
	}
	for _, tt := range tests {
		if got := Str.ContainsAll(tt.haystack, tt.needles); got != tt.want {
			t.Errorf("ContainsAll(%q, %v) = %v, want %v", tt.haystack, tt.needles, got, tt.want)
		}
	}
}

func TestStartsWith(t *testing.T) {
	tests := []struct {
		str, prefix string
		want        bool
	}{
		{"hello world", "hello", true},
		{"hello world", "world", false},
	}
	for _, tt := range tests {
		if got := Str.StartsWith(tt.str, tt.prefix); got != tt.want {
			t.Errorf("StartsWith(%q, %q) = %v, want %v", tt.str, tt.prefix, got, tt.want)
		}
	}
}

func TestEndsWith(t *testing.T) {
	tests := []struct {
		str, suffix string
		want        bool
	}{
		{"hello world", "world", true},
		{"hello world", "hello", false},
	}
	for _, tt := range tests {
		if got := Str.EndsWith(tt.str, tt.suffix); got != tt.want {
			t.Errorf("EndsWith(%q, %q) = %v, want %v", tt.str, tt.suffix, got, tt.want)
		}
	}
}

func TestIs(t *testing.T) {
	tests := []struct {
		pattern, value string
		want           bool
	}{
		{"foo*", "foobar", true},
		{"foo*", "barfoo", false},
		{"*bar", "foobar", true},
		{"foo", "foo", true},
	}
	for _, tt := range tests {
		if got := Str.Is(tt.pattern, tt.value); got != tt.want {
			t.Errorf("Is(%q, %q) = %v, want %v", tt.pattern, tt.value, got, tt.want)
		}
	}
}

func TestIsJSON(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{`{"name":"test"}`, true},
		{`[1,2,3]`, true},
		{`"hello"`, true},
		{`not json`, false},
	}
	for _, tt := range tests {
		if got := Str.IsJSON(tt.input); got != tt.want {
			t.Errorf("IsJSON(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsUUID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"not-a-uuid", false},
	}
	for _, tt := range tests {
		if got := Str.IsUUID(tt.input); got != tt.want {
			t.Errorf("IsUUID(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsEmail(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"test@example.com", true},
		{"not-an-email", false},
	}
	for _, tt := range tests {
		if got := Str.IsEmail(tt.input); got != tt.want {
			t.Errorf("IsEmail(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsURL(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"https://example.com", true},
		{"http://localhost:8080", true},
		{"not-a-url", false},
	}
	for _, tt := range tests {
		if got := Str.IsURL(tt.input); got != tt.want {
			t.Errorf("IsURL(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestPad(t *testing.T) {
	tests := []struct {
		str, pad, padType, want string
		length                  int
	}{
		{"hello", "-", "both", "--hello---", 10},
		{"hello", "-", "left", "-----hello", 10},
		{"hello", "-", "right", "hello-----", 10},
	}
	for _, tt := range tests {
		if got := Str.Pad(tt.str, tt.length, tt.pad, tt.padType); got != tt.want {
			t.Errorf("Pad(%q, %d, %q, %q) = %q, want %q", tt.str, tt.length, tt.pad, tt.padType, got, tt.want)
		}
	}
}

func TestPadLeft(t *testing.T) {
	if got := Str.PadLeft("5", "0", 3); got != "005" {
		t.Errorf("PadLeft(5, 0, 3) = %q, want 005", got)
	}
}

func TestPadRight(t *testing.T) {
	if got := Str.PadRight("5", "0", 3); got != "500" {
		t.Errorf("PadRight(5, 0, 3) = %q, want 500", got)
	}
}

func TestRepeat(t *testing.T) {
	if got := Str.Repeat("ab", 3); got != "ababab" {
		t.Errorf("Repeat(ab, 3) = %q, want ababab", got)
	}
}

func TestReplace(t *testing.T) {
	if got := Str.Replace("o", "0", "hello world"); got != "hell0 w0rld" {
		t.Errorf("Replace = %q, want hell0 w0rld", got)
	}
}

func TestReplaceFirst(t *testing.T) {
	if got := Str.ReplaceFirst("o", "0", "hello world"); got != "hell0 world" {
		t.Errorf("ReplaceFirst = %q, want hell0 world", got)
	}
}

func TestReplaceLast(t *testing.T) {
	if got := Str.ReplaceLast("o", "0", "hello world"); got != "hello w0rld" {
		t.Errorf("ReplaceLast = %q, want hello w0rld", got)
	}
}

func TestRemove(t *testing.T) {
	if got := Str.Remove("l", "hello"); got != "heo" {
		t.Errorf("Remove = %q, want heo", got)
	}
}

func TestReverse(t *testing.T) {
	if got := Str.Reverse("hello"); got != "olleh" {
		t.Errorf("Reverse(hello) = %q, want olleh", got)
	}
}

func TestStart(t *testing.T) {
	tests := []struct {
		str, prefix, want string
	}{
		{"world", "hello ", "hello world"},
		{"hello world", "hello ", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.Start(tt.str, tt.prefix); got != tt.want {
			t.Errorf("Start(%q, %q) = %q, want %q", tt.str, tt.prefix, got, tt.want)
		}
	}
}

func TestFinish(t *testing.T) {
	tests := []struct {
		str, suffix, want string
	}{
		{"hello", " world", "hello world"},
		{"hello world", " world", "hello world"},
	}
	for _, tt := range tests {
		if got := Str.Finish(tt.str, tt.suffix); got != tt.want {
			t.Errorf("Finish(%q, %q) = %q, want %q", tt.str, tt.suffix, got, tt.want)
		}
	}
}

func TestStrWrap(t *testing.T) {
	tests := []struct {
		str, before, after, want string
		hasAfter                 bool
	}{
		{"hello", "'", "'", "'hello'", true},
		{"hello", "(", ")", "(hello)", true},
		{"hello", "*", "", "*hello*", false},
	}
	for _, tt := range tests {
		var got string
		if tt.hasAfter {
			got = Str.Wrap(tt.str, tt.before, tt.after)
		} else {
			got = Str.Wrap(tt.str, tt.before)
		}
		if got != tt.want {
			t.Errorf("Wrap(%q, %q, %q) = %q, want %q", tt.str, tt.before, tt.after, got, tt.want)
		}
	}
}

func TestSubstr(t *testing.T) {
	tests := []struct {
		str    string
		start  int
		length int
		want   string
	}{
		{"hello world", 0, 5, "hello"},
		{"hello world", 6, 5, "world"},
		{"hello world", -5, 5, "world"},
		{"hello world", 0, -1, "hello world"},
	}
	for _, tt := range tests {
		if got := Str.Substr(tt.str, tt.start, tt.length); got != tt.want {
			t.Errorf("Substr(%q, %d, %d) = %q, want %q", tt.str, tt.start, tt.length, got, tt.want)
		}
	}
}

func TestSubstrCount(t *testing.T) {
	if got := Str.SubstrCount("hello world", "l"); got != 3 {
		t.Errorf("SubstrCount = %d, want 3", got)
	}
}

func TestTrim(t *testing.T) {
	if got := Str.Trim("  hello  "); got != "hello" {
		t.Errorf("Trim = %q, want hello", got)
	}
}

func TestLTrim(t *testing.T) {
	if got := Str.LTrim("  hello  "); got != "hello  " {
		t.Errorf("LTrim = %q, want 'hello  '", got)
	}
}

func TestRTrim(t *testing.T) {
	if got := Str.RTrim("  hello  "); got != "  hello" {
		t.Errorf("RTrim = %q, want '  hello'", got)
	}
}

func TestSquish(t *testing.T) {
	if got := Str.Squish("hello    world  \n  test"); got != "hello world test" {
		t.Errorf("Squish = %q, want 'hello world test'", got)
	}
}

func TestWordCount(t *testing.T) {
	if got := Str.WordCount("hello world test"); got != 3 {
		t.Errorf("WordCount = %d, want 3", got)
	}
}

func TestUUID(t *testing.T) {
	uuid := Str.UUID()
	if !Str.IsUUID(uuid) {
		t.Errorf("UUID() generated invalid UUID: %s", uuid)
	}
}

func TestStrRandom(t *testing.T) {
	random := Str.Random(10)
	if len(random) != 10 {
		t.Errorf("Random(10) length = %d, want 10", len(random))
	}
}

func TestFluent(t *testing.T) {
	result := Str.Of("  HELLO WORLD  ").Trim().Lower().Value()
	if result != "hello world" {
		t.Errorf("Fluent chain = %q, want 'hello world'", result)
	}

	result = Str.Of("hello_world").Camel().Value()
	if result != "helloWorld" {
		t.Errorf("Fluent Camel = %q, want 'helloWorld'", result)
	}

	result = Str.Of("HelloWorld").Snake().Value()
	if result != "hello_world" {
		t.Errorf("Fluent Snake = %q, want 'hello_world'", result)
	}

	result = Str.Of("Hello World").Slug().Value()
	if result != "hello-world" {
		t.Errorf("Fluent Slug = %q, want 'hello-world'", result)
	}
}

func TestTitle(t *testing.T) {
	if got := Str.Title("hello world"); !strings.Contains(got, "Hello") {
		t.Errorf("Title(hello world) = %q, want string containing 'Hello'", got)
	}
}
