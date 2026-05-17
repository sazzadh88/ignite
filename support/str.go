package support

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/mail"
	"net/url"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// StringHelper provides string manipulation utilities.
type StringHelper struct{}

// Str is the global string helper instance.
var Str = StringHelper{}

// After returns the portion of the string after the first occurrence of the search string.
func (s StringHelper) After(subject, search string) string {
	if search == "" {
		return subject
	}
	pos := strings.Index(subject, search)
	if pos == -1 {
		return subject
	}
	return subject[pos+len(search):]
}

// AfterLast returns the portion of the string after the last occurrence of the search string.
func (s StringHelper) AfterLast(subject, search string) string {
	if search == "" {
		return subject
	}
	pos := strings.LastIndex(subject, search)
	if pos == -1 {
		return subject
	}
	return subject[pos+len(search):]
}

// Before returns the portion of the string before the first occurrence of the search string.
func (s StringHelper) Before(subject, search string) string {
	if search == "" {
		return subject
	}
	pos := strings.Index(subject, search)
	if pos == -1 {
		return subject
	}
	return subject[:pos]
}

// BeforeLast returns the portion of the string before the last occurrence of the search string.
func (s StringHelper) BeforeLast(subject, search string) string {
	if search == "" {
		return subject
	}
	pos := strings.LastIndex(subject, search)
	if pos == -1 {
		return subject
	}
	return subject[:pos]
}

// Between returns the portion of the string between two strings.
func (s StringHelper) Between(subject, from, to string) string {
	if from == "" && to == "" {
		return subject
	}
	start := 0
	if from != "" {
		pos := strings.Index(subject, from)
		if pos == -1 {
			return ""
		}
		start = pos + len(from)
	}
	if to == "" {
		return subject[start:]
	}
	end := strings.Index(subject[start:], to)
	if end == -1 {
		return ""
	}
	return subject[start : start+end]
}

// Camel converts a string to camelCase.
func (s StringHelper) Camel(str string) string {
	words := splitWords(str)
	if len(words) == 0 {
		return ""
	}
	result := strings.ToLower(words[0])
	for i := 1; i < len(words); i++ {
		result += ucfirst(words[i])
	}
	return result
}

// Snake converts a string to snake_case.
func (s StringHelper) Snake(str string) string {
	var result strings.Builder
	for i, r := range str {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(rune(str[i-1])) || (i < len(str)-1 && unicode.IsLower(rune(str[i+1])))) {
				result.WriteRune('_')
			}
			result.WriteRune(unicode.ToLower(r))
		} else if r == ' ' || r == '-' {
			result.WriteRune('_')
		} else {
			result.WriteRune(r)
		}
	}
	return strings.Trim(result.String(), "_")
}

// Kebab converts a string to kebab-case.
func (s StringHelper) Kebab(str string) string {
	return strings.ReplaceAll(s.Snake(str), "_", "-")
}

// Studly converts a string to StudlyCase (PascalCase).
func (s StringHelper) Studly(str string) string {
	words := splitWords(str)
	var result strings.Builder
	for _, word := range words {
		result.WriteString(ucfirst(word))
	}
	return result.String()
}

// Title converts a string to Title Case.
func (s StringHelper) Title(str string) string {
	return strings.Title(strings.ToLower(str))
}

// Slug converts a string to a URL-friendly slug.
func (s StringHelper) Slug(str string, sep ...string) string {
	separator := "-"
	if len(sep) > 0 {
		separator = sep[0]
	}
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	slug := reg.ReplaceAllString(strings.ToLower(str), separator)
	return strings.Trim(slug, separator)
}

// Upper converts a string to uppercase.
func (s StringHelper) Upper(str string) string {
	return strings.ToUpper(str)
}

// Lower converts a string to lowercase.
func (s StringHelper) Lower(str string) string {
	return strings.ToLower(str)
}

// Ucfirst converts the first character to uppercase.
func (s StringHelper) Ucfirst(str string) string {
	return ucfirst(str)
}

// Lcfirst converts the first character to lowercase.
func (s StringHelper) Lcfirst(str string) string {
	if str == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(str)
	return string(unicode.ToLower(r)) + str[size:]
}

// Length returns the length of the string in runes.
func (s StringHelper) Length(str string) int {
	return utf8.RuneCountInString(str)
}

// Limit truncates a string to a given length and appends an ending.
func (s StringHelper) Limit(str string, limit int, end ...string) string {
	ending := "..."
	if len(end) > 0 {
		ending = end[0]
	}
	if utf8.RuneCountInString(str) <= limit {
		return str
	}
	runes := []rune(str)
	return string(runes[:limit]) + ending
}

// Words truncates a string to a given number of words.
func (s StringHelper) Words(str string, words int, end ...string) string {
	ending := "..."
	if len(end) > 0 {
		ending = end[0]
	}
	wordList := strings.Fields(str)
	if len(wordList) <= words {
		return str
	}
	return strings.Join(wordList[:words], " ") + ending
}

// Contains determines if a string contains a given substring.
func (s StringHelper) Contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// ContainsAll determines if a string contains all given substrings.
func (s StringHelper) ContainsAll(haystack string, needles []string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}

// StartsWith determines if a string starts with a given prefix.
func (s StringHelper) StartsWith(str, prefix string) bool {
	return strings.HasPrefix(str, prefix)
}

// EndsWith determines if a string ends with a given suffix.
func (s StringHelper) EndsWith(str, suffix string) bool {
	return strings.HasSuffix(str, suffix)
}

// Is determines if a string matches a wildcard pattern.
func (s StringHelper) Is(pattern, value string) bool {
	if pattern == value {
		return true
	}
	pattern = regexp.QuoteMeta(pattern)
	pattern = strings.ReplaceAll(pattern, `\*`, ".*")
	pattern = "^" + pattern + "$"
	matched, _ := regexp.MatchString(pattern, value)
	return matched
}

// IsJSON determines if a string is valid JSON.
func (s StringHelper) IsJSON(str string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(str), &js) == nil
}

// IsUUID determines if a string is a valid UUID.
func (s StringHelper) IsUUID(str string) bool {
	uuidRegex := regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	return uuidRegex.MatchString(str)
}

// IsEmail determines if a string is a valid email address.
func (s StringHelper) IsEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}

// IsURL determines if a string is a valid URL.
func (s StringHelper) IsURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// Pad pads a string to a certain length with another string.
func (s StringHelper) Pad(str string, length int, pad string, padType ...string) string {
	strLen := utf8.RuneCountInString(str)
	if strLen >= length || pad == "" {
		return str
	}
	padLen := length - strLen
	pType := "both"
	if len(padType) > 0 {
		pType = padType[0]
	}
	padStr := strings.Repeat(pad, (padLen/utf8.RuneCountInString(pad))+1)
	switch pType {
	case "left":
		return string([]rune(padStr)[:padLen]) + str
	case "right":
		return str + string([]rune(padStr)[:padLen])
	default: // both
		left := padLen / 2
		right := padLen - left
		return string([]rune(padStr)[:left]) + str + string([]rune(padStr)[:right])
	}
}

// PadLeft pads a string to the left.
func (s StringHelper) PadLeft(str, pad string, length int) string {
	return s.Pad(str, length, pad, "left")
}

// PadRight pads a string to the right.
func (s StringHelper) PadRight(str, pad string, length int) string {
	return s.Pad(str, length, pad, "right")
}

// Repeat repeats a string n times.
func (s StringHelper) Repeat(str string, times int) string {
	return strings.Repeat(str, times)
}

// Replace replaces all occurrences of search with replace.
func (s StringHelper) Replace(search, replace, subject string) string {
	return strings.ReplaceAll(subject, search, replace)
}

// ReplaceFirst replaces the first occurrence of search with replace.
func (s StringHelper) ReplaceFirst(search, replace, subject string) string {
	return strings.Replace(subject, search, replace, 1)
}

// ReplaceLast replaces the last occurrence of search with replace.
func (s StringHelper) ReplaceLast(search, replace, subject string) string {
	pos := strings.LastIndex(subject, search)
	if pos == -1 {
		return subject
	}
	return subject[:pos] + replace + subject[pos+len(search):]
}

// Remove removes all occurrences of search from subject.
func (s StringHelper) Remove(search, subject string) string {
	return strings.ReplaceAll(subject, search, "")
}

// Reverse reverses a string.
func (s StringHelper) Reverse(str string) string {
	runes := []rune(str)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// Start ensures a string starts with a given prefix.
func (s StringHelper) Start(str, prefix string) string {
	if strings.HasPrefix(str, prefix) {
		return str
	}
	return prefix + str
}

// Finish ensures a string ends with a given suffix.
func (s StringHelper) Finish(str, suffix string) string {
	if strings.HasSuffix(str, suffix) {
		return str
	}
	return str + suffix
}

// Wrap wraps a string with the given strings.
func (s StringHelper) Wrap(str, before string, after ...string) string {
	aft := before
	if len(after) > 0 {
		aft = after[0]
	}
	return before + str + aft
}

// Substr returns a substring starting at start with length.
func (s StringHelper) Substr(str string, start, length int) string {
	runes := []rune(str)
	runeLen := len(runes)
	if start < 0 {
		start = runeLen + start
	}
	if start < 0 {
		start = 0
	}
	if start >= runeLen {
		return ""
	}
	end := start + length
	if length < 0 || end > runeLen {
		end = runeLen
	}
	return string(runes[start:end])
}

// SubstrCount counts the number of substring occurrences.
func (s StringHelper) SubstrCount(str, substr string) int {
	return strings.Count(str, substr)
}

// Trim removes whitespace from both ends.
func (s StringHelper) Trim(str string) string {
	return strings.TrimSpace(str)
}

// LTrim removes whitespace from the left end.
func (s StringHelper) LTrim(str string) string {
	return strings.TrimLeftFunc(str, unicode.IsSpace)
}

// RTrim removes whitespace from the right end.
func (s StringHelper) RTrim(str string) string {
	return strings.TrimRightFunc(str, unicode.IsSpace)
}

// Squish collapses all whitespace sequences into single spaces.
func (s StringHelper) Squish(str string) string {
	return strings.Join(strings.Fields(str), " ")
}

// WordCount returns the number of words in the string.
func (s StringHelper) WordCount(str string) int {
	return len(strings.Fields(str))
}

// UUID generates a new UUID v4 string.
func (s StringHelper) UUID() string {
	uuid := make([]byte, 16)
	if _, err := rand.Read(uuid); err != nil {
		return ""
	}
	// Set version (4) and variant (RFC 4122)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80

	buf := make([]byte, 36)
	hex.Encode(buf[0:8], uuid[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], uuid[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], uuid[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], uuid[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], uuid[10:])

	return string(buf)
}

// Random generates a random alphanumeric string of the given length.
func (s StringHelper) Random(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = chars[b%byte(len(chars))]
	}
	return string(bytes)
}

// Fluent represents a fluent string builder.
type Fluent struct {
	value string
}

// Of creates a new fluent string builder.
func (s StringHelper) Of(str string) *Fluent {
	return &Fluent{value: str}
}

// Trim removes whitespace from both ends.
func (f *Fluent) Trim() *Fluent {
	f.value = Str.Trim(f.value)
	return f
}

// Lower converts to lowercase.
func (f *Fluent) Lower() *Fluent {
	f.value = Str.Lower(f.value)
	return f
}

// Upper converts to uppercase.
func (f *Fluent) Upper() *Fluent {
	f.value = Str.Upper(f.value)
	return f
}

// Slug converts to a URL slug.
func (f *Fluent) Slug(sep ...string) *Fluent {
	f.value = Str.Slug(f.value, sep...)
	return f
}

// Camel converts to camelCase.
func (f *Fluent) Camel() *Fluent {
	f.value = Str.Camel(f.value)
	return f
}

// Snake converts to snake_case.
func (f *Fluent) Snake() *Fluent {
	f.value = Str.Snake(f.value)
	return f
}

// Value returns the final string value.
func (f *Fluent) Value() string {
	return f.value
}

// Helper functions

func splitWords(str string) []string {
	return strings.FieldsFunc(str, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

func ucfirst(str string) string {
	if str == "" {
		return ""
	}
	r, size := utf8.DecodeRuneInString(str)
	return string(unicode.ToUpper(r)) + strings.ToLower(str[size:])
}
