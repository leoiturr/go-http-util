package handlers

import (
	"html/template"
	"strings"
)

// TemplateFuncs is the single source of truth for the custom template helpers.
//
// It lives in the handlers package rather than in main so that both the server
// and the handler tests build templates from the same map. Duplicating the map
// per test file previously meant a new helper broke every test that loaded the
// templates, because a missing function panics at LoadHTMLGlob time.
var TemplateFuncs = template.FuncMap{
	"add": func(a, b int) int {
		return a + b
	},
	"replace": func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	},
	"upper": strings.ToUpper,
}
