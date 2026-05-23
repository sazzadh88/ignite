package scaffold

import (
	"embed"
	"strings"
)

//go:embed all:stubs
var stubs embed.FS

// stub reads a stub file and applies placeholder replacements.
func stub(name string, replacements map[string]string) string {
	data, err := stubs.ReadFile("stubs/" + name)
	if err != nil {
		panic("scaffold: missing stub file: " + name)
	}
	result := string(data)
	for k, v := range replacements {
		result = strings.ReplaceAll(result, k, v)
	}
	return result
}
